package client

import (
	"context"
	"fmt"
	"os"
	"path"
	"strings"

	"cuelang.org/go/cue"
	"cuelang.org/go/cue/errors"
	"cuelang.org/go/cue/format"
	"github.com/fatih/color"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	log "github.com/sirupsen/logrus"

	"github.com/devopzilla/guku-devx/pkg/auth"
	"github.com/devopzilla/guku-devx/pkg/drivers"
	"github.com/devopzilla/guku-devx/pkg/project"
	"github.com/devopzilla/guku-devx/pkg/stack"
	"github.com/devopzilla/guku-devx/pkg/stackbuilder"
	"github.com/devopzilla/guku-devx/pkg/utils"
)

func Run(environment string, configDir string, stackPath string, buildersPath string, reserve bool, dryRun bool, server auth.ServerConfig, noStrict bool, stdout bool) error {
	ctx := context.Background()
	ctx = context.WithValue(ctx, utils.ConfigDirKey, configDir)
	ctx = context.WithValue(ctx, utils.DryRunKey, dryRun)

	if err := project.Update(configDir, server); err != nil {
		return err
	}

	stack, builder, err := buildStack(ctx, environment, configDir, stackPath, buildersPath, noStrict)
	if err != nil {
		if auth.IsLoggedIn(server) {
			details := errors.Details(err, nil)
			if buildId, err := stack.SendBuild(configDir, server, environment, &details); err != nil {
				log.Error("failed to save build data: ", err.Error())
			} else {
				log.Infof("\nSaved failed build at %s/builds/%s\n", server.Endpoint, buildId)
			}
		}
		return err
	}

	if dryRun {
		log.Info(stack.GetComponents())
		return nil
	}

	for id, driver := range drivers.NewDriversMap(environment, builder.DriverConfig) {
		if err := driver.ApplyAll(stack, stdout); err != nil {
			newErr := fmt.Errorf("error running %s driver: %s", id, errors.Details(err, nil))
			if auth.IsLoggedIn(server) {
				details := newErr.Error()
				if buildId, err := stack.SendBuild(configDir, server, environment, &details); err != nil {
					log.Error("failed to save build data: ", err.Error())
				} else {
					log.Infof("\nSaved failed build at %s/builds/%s", server.Endpoint, buildId)
				}
			}
			return newErr
		}
	}

	if auth.IsLoggedIn(server) {
		log.Info("📤 Analyzing & uploading build data...")
		buildId, err := stack.SendBuild(configDir, server, environment, nil)
		if err != nil {
			return err
		}
		log.Infof("\nCreated build at %s/builds/%s", server.Endpoint, buildId)

		if reserve {
			err := Reserve(buildId, server, dryRun)
			if err != nil {
				return err
			}
		} else {
			log.Info("To reserve build resources run:")
			log.Infof("devx reserve %s\n", buildId)
		}
	}

	return nil
}

func Diff(target string, environment string, configDir string, stackPath string, buildersPath string, server auth.ServerConfig, noStrict bool) error {
	log.Infof("📍 Processing target stack @ %s", target)
	targetDir, err := os.MkdirTemp("", "devx-target-*")
	if err != nil {
		return err
	}
	defer os.RemoveAll(targetDir)

	repo, err := git.PlainClone(targetDir, false, &git.CloneOptions{
		URL: configDir,
	})
	if err != nil {
		return err
	}

	w, err := repo.Worktree()
	if err != nil {
		return err
	}

	hash, err := repo.ResolveRevision(plumbing.Revision(target))
	if err != nil {
		return err
	}

	err = w.Checkout(&git.CheckoutOptions{
		Hash: *hash,
	})
	if err != nil {
		return err
	}

	err = project.Update(targetDir, server)
	if err != nil {
		return err
	}

	targetCtx := context.Background()
	targetCtx = context.WithValue(targetCtx, utils.ConfigDirKey, targetDir)
	targetCtx = context.WithValue(targetCtx, utils.DryRunKey, true)
	targetStack, _, err := buildStack(targetCtx, environment, targetDir, stackPath, buildersPath, noStrict)
	if err != nil {
		return err
	}

	log.Info("\n📍 Processing current stack")
	currentCtx := context.Background()
	currentCtx = context.WithValue(currentCtx, utils.ConfigDirKey, configDir)
	currentCtx = context.WithValue(currentCtx, utils.DryRunKey, true)
	currentStack, _, err := buildStack(currentCtx, environment, configDir, stackPath, buildersPath, noStrict)
	if err != nil {
		return err
	}

	currentValues := utils.GetLeaves(currentStack.GetComponents(), false)
	targetValues := utils.GetLeaves(targetStack.GetComponents(), false)

	remColor := color.New(color.FgRed)
	addColor := color.New(color.FgGreen)
	updColor := color.New(color.FgYellow)
	log.Info("\n🔬 Diff")
	ci, ti := 0, 0
	for ci < len(currentValues) || ti < len(targetValues) {
		if ci == len(currentValues) {
			tv := targetValues[ti]
			log.Infof("\t%s %s: %s", remColor.Sprintf("-"), tv.Path, tv.Value)
			ti++
			continue
		}
		if ti == len(targetValues) {
			cv := currentValues[ci]
			log.Infof("\t%s %s: %s", addColor.Sprintf("+"), cv.Path, cv.Value)
			ci++
			continue
		}

		cv := currentValues[ci]
		tv := targetValues[ti]
		switch strings.Compare(cv.Path, tv.Path) {
		case 0:
			if strings.Compare(cv.Value, tv.Value) != 0 {
				log.Infof("\t%s %s: %s -> %s", updColor.Sprintf("~"), cv.Path, tv.Value, cv.Value)
			}
			ci++
			ti++
		case -1:
			log.Infof("\t%s %s: %s", addColor.Sprintf("+"), cv.Path, cv.Value)
			ci++
		case 1:
			log.Infof("\t%s %s: %s", remColor.Sprintf("-"), tv.Path, tv.Value)
			ti++
		}
	}

	return nil
}

func buildStack(ctx context.Context, environment string, configDir string, stackPath string, buildersPath string, noStrict bool) (*stack.Stack, *stackbuilder.StackBuilder, error) {
	log.Infof("🏗️  Loading stack...")

	overlays, err := utils.GetOverlays(configDir)
	if err != nil {
		return nil, nil, err
	}
	value, stackId, depIds := utils.LoadProject(configDir, &overlays)

	buildSource, err := format.Node(value.Syntax(), format.Simplify())
	if err != nil {
		log.Fatal(err)
	}

	emptyStack := stack.Stack{
		ID:          stackId,
		DepIDs:      depIds,
		BuildSource: string(buildSource),
	}
	emptyStack.AddComponents(value.Context().CompileString("{}"))

	log.Info("👀 Validating stack...")
	err = project.ValidateProject(value, stackPath, buildersPath, noStrict)
	if err != nil {
		return &emptyStack, nil, err
	}

	builders, err := stackbuilder.NewEnvironments(value.LookupPath(cue.ParsePath(buildersPath)))
	if err != nil {
		return &emptyStack, nil, err
	}

	builder, ok := builders[environment]
	if !ok {
		return &emptyStack, nil, fmt.Errorf("environment %s was not found", environment)
	}

	stack, err := stack.NewStack(value.LookupPath(cue.ParsePath(stackPath)), stackId, depIds)
	if err != nil {
		return &emptyStack, nil, err
	}
	stack.BuildSource = emptyStack.BuildSource

	err = builder.TransformStack(ctx, stack)
	if err != nil {
		return stack, nil, err
	}

	return stack, builder, nil
}

func Reserve(buildId string, server auth.ServerConfig, dryRun bool) error {
	if !auth.IsLoggedIn(server) {
		return fmt.Errorf("must be logged in to be able to reserve resources")
	}

	reserveData := struct {
		DryRun bool `json:"dryRun,omitempty"`
	}{
		DryRun: dryRun,
	}

	apiPath := path.Join("builds", buildId, "reserve")
	data, err := utils.SendData(server, apiPath, reserveData)
	if err != nil {
		log.Debug(string(data))
		return err
	}

	if dryRun {
		log.Infof("Looks good, you can reserve this build!")
		return nil
	}

	log.Infof("Reserved build with id %s", buildId)

	return nil
}

func Retire(buildId string, server auth.ServerConfig) error {
	if !auth.IsLoggedIn(server) {
		return fmt.Errorf("must be logged in to be able to retire resources")
	}

	apiPath := path.Join("builds", buildId, "retire")
	data, err := utils.SendData(server, apiPath, map[string]interface{}{})
	if err != nil {
		log.Debug(string(data))
		return err
	}

	log.Infof("Retired build with id %s", buildId)

	return nil
}
