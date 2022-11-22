package client

import (
	"fmt"
	"path"

	"cuelang.org/go/cue"
	"devopzilla.com/guku/internal/drivers"
	"devopzilla.com/guku/internal/project"
	"devopzilla.com/guku/internal/stack"
	"devopzilla.com/guku/internal/stackbuilder"
	"devopzilla.com/guku/internal/utils"
)

func Run(environment string, configDir string, stackPath string, buildersPath string) error {
	value := utils.LoadProject(configDir)
	fmt.Printf("👀 Validating stack...\n")

	err := project.ValidateProject(value, stackPath)
	if err != nil {
		return err
	}

	fmt.Printf("🏭 Transforming stack for the \"%s\" environment...\n", environment)

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

	compose := drivers.ComposeDriver{
		Path: path.Join("build", environment, "compose"),
	}
	compose.ApplyAll(stack)

	terraform := drivers.TerraformDriver{
		Path: path.Join("build", environment, "terraform"),
	}
	terraform.ApplyAll(stack)

	kubernetes := drivers.KubernetesDriver{
		Path: path.Join("build", environment, "kubernetes"),
	}
	kubernetes.ApplyAll(stack)

	return nil
}
