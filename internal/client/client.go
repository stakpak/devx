package client

import (
	"context"
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
	ctx := context.Background()
	ctx = context.WithValue(ctx, utils.ConfigDirKey, configDir)
	ctx = context.WithValue(ctx, utils.DryRunKey, dryRun)

	stack, builder, err := buildStack(ctx, environment, configDir, stackPath, buildersPath)
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

	targetCtx := context.Background()
	targetCtx = context.WithValue(targetCtx, utils.ConfigDirKey, targetDir)
	targetCtx = context.WithValue(targetCtx, utils.DryRunKey, true)
	targetStack, _, err := buildStack(targetCtx, environment, targetDir, stackPath, buildersPath)
	if err != nil {
		return err
	}

	log.Info("\n📍 Processing current stack")
	currentCtx := context.Background()
	currentCtx = context.WithValue(currentCtx, utils.ConfigDirKey, configDir)
	currentCtx = context.WithValue(currentCtx, utils.DryRunKey, true)
	currentStack, _, err := buildStack(currentCtx, environment, configDir, stackPath, buildersPath)
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
			log.Infof("\t%s %s: %v", remColor.Sprintf("-"), tv.Path, tv.Value)
			ti++
			continue
		}
		if ti == len(targetValues) {
			cv := currentValues[ci]
			log.Infof("\t%s %s: %v", addColor.Sprintf("+"), cv.Path, cv.Value)
			ci++
			continue
		}

		cv := currentValues[ci]
		tv := targetValues[ti]
		switch strings.Compare(cv.Path, tv.Path) {
		case 0:
			if strings.Compare(fmt.Sprint(cv.Value), fmt.Sprint(tv.Value)) != 0 {
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

func dumpexpr(v cue.Value, level int) {
	op, values := v.Expr()

	switch op {
	case cue.NoOp:
		//fmt.Printf("%s└ %v\n", strings.Repeat(" ", level), v)
	case cue.SelectorOp:
		fmt.Printf("%s└─ %v\n", strings.Repeat(" ", level), cue.Dereference(v).Path())
	case cue.AndOp, cue.OrOp, cue.InterpolationOp:
		lastpath := ""

		for _, value := range values {
			if value.Path().String() != lastpath {
				fmt.Printf("%s└─ %v\n", strings.Repeat(" ", level), value.Path())
				lastpath = value.Path().String()
			}

			dumpexpr(value, level+2)
		}
	default:
		fmt.Printf("%s└─ %v\n", strings.Repeat(" ", level), op)
	}
}

func Lint(environment string, configDir string, stackPath string, buildersPath string) error {
	ctx := context.Background()
	stack, _, err := buildStack(ctx, environment, configDir, stackPath, buildersPath)
	if err != nil {
		return err
	}

	leaves := utils.GetLeaves(stack.GetComponents(), false)

	for _, leave := range leaves {
		if strings.Contains(leave.Path, "$resources") {
			// resource
			fmt.Printf("[R] %s\n", leave.Path)
			dumpexpr(leave.Value, 1)
		} else {
			// component
			fmt.Printf("[C] %s\n", leave.Path)
			dumpexpr(leave.Value, 1)
		}
	}

	return nil

	resources := []string{}
	components := []string{}

	for _, leave := range leaves {
		if !strings.Contains(leave.Path, "echo") {
			continue
		}

		if strings.Contains(leave.Path, "$resources") {
			refs := utils.GetReferences(leave.Value)

			for _, ref := range refs {
				if !strings.Contains(ref.Path().String(), "$resources") {
					rpath := cue.MakePath(ref.Path().Selectors()[2:]...).String()
					resources = append(resources, rpath)
					//fmt.Printf("[R] %s -> %s\n", leave.Path, rpath)
				}
			}
		} else {
			components = append(components, leave.Path)
		}
	}

	for _, resource := range resources {
		fmt.Printf("[R] %s\n", resource)
	}

	for _, component := range components {
		fmt.Printf("[C] %s\n", component)
	}

	/*
		for _, leave := range leaves {
			if strings.Contains(leave.Path, "$resources") {
				refs := utils.GetReferences(leave.Value)

				for _, ref := range refs {
					if !strings.Contains(ref, "$resources") {
						fmt.Printf("%s -> %s\n", leave.Path, ref)
					}
				}
			}
		}

		return nil

		components := map[string]cue.Value{}
		resources := []utils.Leaf{}

		for _, leave := range leaves {
			path := leave.Path

			if strings.Contains(path, "$resources") {
				resources = append(resources, utils.Leaf{Path: path, Value: leave.Value})
			} else {
				components[path] = leave.Value
			}

		}

		refs := []string{}

		for _, resource := range resources {
			refs = append(refs, utils.GetReferences(resource.Value)...)
		}

		for _, ref := range refs {
			//fmt.Printf("\n---\n%#v\n", refs)
			fmt.Println(ref)
		}
	*/

	return nil
}

func buildStack(ctx context.Context, environment string, configDir string, stackPath string, buildersPath string) (*stack.Stack, *stackbuilder.StackBuilder, error) {
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
		return nil, nil, fmt.Errorf("environment %s was not found", environment)
	}

	stack, err := stack.NewStack(value.LookupPath(cue.ParsePath(stackPath)))
	if err != nil {
		return nil, nil, err
	}

	err = builder.TransformStack(ctx, stack)
	if err != nil {
		return nil, nil, err
	}

	return stack, builder, nil
}
