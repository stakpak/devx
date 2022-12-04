package client

import (
	"fmt"

	"cuelang.org/go/cue"
	"devopzilla.com/guku/internal/drivers"
	"devopzilla.com/guku/internal/project"
	"devopzilla.com/guku/internal/stack"
	"devopzilla.com/guku/internal/stackbuilder"
	"devopzilla.com/guku/internal/utils"
)

func Run(environment string, configDir string, stackPath string, buildersPath string, dryRun bool) error {
	fmt.Printf("🏗️  Loading stack...\n")
	value := utils.LoadProject(configDir)
	fmt.Printf("👀 Validating stack...\n")

	err := project.ValidateProject(value, stackPath)
	if err != nil {
		return err
	}

	builders, err := stackbuilder.NewEnvironments(value.LookupPath(cue.ParsePath(buildersPath)))
	if err != nil {
		return err
	}

	builder, ok := builders[environment]
	if !ok {
		return fmt.Errorf("Environment %s was not found", environment)
	}

	stack, err := stack.NewStack(value.LookupPath(cue.ParsePath(stackPath)))
	if err != nil {
		return err
	}

	err = builder.TransformStack(stack)
	if err != nil {
		return err
	}

	if dryRun {
		fmt.Println(stack.GetComponents())
		return nil
	}

	for id, driver := range drivers.NewDriversMap(environment, builder.DriverConfig) {
		if err := driver.ApplyAll(stack); err != nil {
			return fmt.Errorf("error running %s driver: %s", id, err)
		}
	}

	return nil
}
