package client

import (
	"context"
	"fmt"
	"os"
	"path"
	"strings"

	"cuelang.org/go/cue"
	"github.com/fatih/color"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	log "github.com/sirupsen/logrus"

	"devopzilla.com/guku/internal/drivers"
	"devopzilla.com/guku/internal/project"
	"devopzilla.com/guku/internal/stack"
	"devopzilla.com/guku/internal/stackbuilder"
	"devopzilla.com/guku/internal/utils"
)

func Run(environment string, configDir string, stackPath string, buildersPath string, dryRun bool, telemetry string, strict bool) error {
	ctx := context.Background()
	ctx = context.WithValue(ctx, utils.ConfigDirKey, configDir)
	ctx = context.WithValue(ctx, utils.DryRunKey, dryRun)

	stack, builder, err := buildStack(ctx, environment, configDir, stackPath, buildersPath, strict)
	if err != nil {
		return err
	}

	if telemetry != "" {
		buildId, err := stack.SendBuild(configDir, telemetry, environment)
		if err != nil {
			return err
		}
		log.Infof("\nCreated build at %s/builds/%s", telemetry, buildId)
		log.Info("To reserve build resources run:")
		log.Infof("devx reserve %s --telemetry %s\n", buildId, telemetry)
	}

	if dryRun {
		log.Info(stack.GetComponents())
		return nil
	}

	for id, driver := range drivers.NewDriversMap(environment, builder.DriverConfig) {
		if err := driver.ApplyAll(stack); err != nil {
			return fmt.Errorf("error running %s driver: %s", id, err)
		}
	}

	return nil
}

func Diff(target string, environment string, configDir string, stackPath string, buildersPath string, strict bool) error {
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

	err = project.Update(targetDir)
	if err != nil {
		return err
	}

	targetCtx := context.Background()
	targetCtx = context.WithValue(targetCtx, utils.ConfigDirKey, targetDir)
	targetCtx = context.WithValue(targetCtx, utils.DryRunKey, true)
	targetStack, _, err := buildStack(targetCtx, environment, targetDir, stackPath, buildersPath, strict)
	if err != nil {
		return err
	}

	log.Info("\n📍 Processing current stack")
	currentCtx := context.Background()
	currentCtx = context.WithValue(currentCtx, utils.ConfigDirKey, configDir)
	currentCtx = context.WithValue(currentCtx, utils.DryRunKey, true)
	currentStack, _, err := buildStack(currentCtx, environment, configDir, stackPath, buildersPath, strict)
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

func buildStack(ctx context.Context, environment string, configDir string, stackPath string, buildersPath string, strict bool) (*stack.Stack, *stackbuilder.StackBuilder, error) {
	log.Infof("🏗️  Loading stack...")
	overlays, err := utils.GetOverlays(configDir)
	if err != nil {
		return nil, nil, err
	}
	value, stackId, depIds := utils.LoadProject(configDir, &overlays)

	log.Info("👀 Validating stack...")
	err = project.ValidateProject(value, stackPath, buildersPath, strict)
	if err != nil {
		return nil, nil, err
	}

	builders, err := stackbuilder.NewEnvironments(value.LookupPath(cue.ParsePath(buildersPath)))
	if err != nil {
		return nil, nil, err
	}

	builder, ok := builders[environment]
	if !ok {
		return nil, nil, fmt.Errorf("environment %s was not found", environment)
	}

	stack, err := stack.NewStack(value.LookupPath(cue.ParsePath(stackPath)), stackId, depIds)
	if err != nil {
		return nil, nil, err
	}

	err = builder.TransformStack(ctx, stack)
	if err != nil {
		return nil, nil, err
	}

	return stack, builder, nil
}

func Reserve(buildId string, telemetry string, dryRun bool) error {
	if telemetry == "" {
		return fmt.Errorf("telemtry endpoint is required to reserve build resources")
	}

	reserveData := struct {
		DryRun bool `json:"dryRun,omitempty"`
	}{
		DryRun: dryRun,
	}

	apiPath := path.Join("builds", buildId, "reserve")
	data, err := utils.SendTelemtry(telemetry, apiPath, reserveData)
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
