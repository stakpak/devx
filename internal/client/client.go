package client

import (
	"fmt"
	"os"
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

func Run(environment string, configDir string, stackPath string, buildersPath string, dryRun bool) error {
	stack, builder, err := buildStack(environment, configDir, stackPath, buildersPath)
	if err != nil {
		return err
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

func Diff(target string, environment string, configDir string, stackPath string, buildersPath string) error {
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

	targetStack, _, err := buildStack(environment, targetDir, stackPath, buildersPath)
	if err != nil {
		return err
	}

	log.Info("\n📍 Processing current stack")
	currentStack, _, err := buildStack(environment, configDir, stackPath, buildersPath)
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
			log.Info(remColor.Sprintf("\t%s %s: %s", "-", tv.Path, tv.Value))
			ti++
			continue
		}
		if ti == len(targetValues) {
			cv := currentValues[ci]
			log.Info(addColor.Sprintf("\t%s %s: %s", "+", cv.Path, cv.Value))
			ci++
			continue
		}

		cv := currentValues[ci]
		tv := targetValues[ti]
		switch strings.Compare(cv.Path, tv.Path) {
		case 0:
			if strings.Compare(cv.Value, tv.Value) != 0 {
				log.Info(updColor.Sprintf("\t%s %s: %s -> %s", "~", cv.Path, tv.Value, cv.Value))
			}
			ci++
			ti++
		case -1:
			log.Info(addColor.Sprintf("\t%s %s: %s", "+", cv.Path, cv.Value))
			ci++
		case 1:
			log.Info(remColor.Sprintf("\t%s %s: %s", "-", tv.Path, tv.Value))
			ti++
		}
	}

	return nil
}

func buildStack(environment string, configDir string, stackPath string, buildersPath string) (*stack.Stack, *stackbuilder.StackBuilder, error) {
	log.Infof("🏗️  Loading stack...")
	overlays, err := utils.GetOverlays(configDir)
	if err != nil {
		return nil, nil, err
	}
	value := utils.LoadProject(configDir, &overlays)

	log.Info("👀 Validating stack...")
	err = project.ValidateProject(value, stackPath)
	if err != nil {
		return nil, nil, err
	}

	builders, err := stackbuilder.NewEnvironments(value.LookupPath(cue.ParsePath(buildersPath)))
	if err != nil {
		return nil, nil, err
	}

	builder, ok := builders[environment]
	if !ok {
		return nil, nil, fmt.Errorf("Environment %s was not found", environment)
	}

	stack, err := stack.NewStack(value.LookupPath(cue.ParsePath(stackPath)))
	if err != nil {
		return nil, nil, err
	}

	err = builder.TransformStack(stack)
	if err != nil {
		return nil, nil, err
	}

	return stack, builder, nil
}
