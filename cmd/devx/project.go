package main

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"devopzilla.com/guku/internal/project"
	"devopzilla.com/guku/pkg/cuemods"
)

var projectCmd = &cobra.Command{
	Use:   "project",
	Short: "Manage a DevX project",
}

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a project",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := cuemods.Init(context.TODO(), configDir, ""); err != nil {
			return err
		}
		return nil
	},
}

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update/Install project dependencies",
	RunE: func(cmd *cobra.Command, args []string) error {
		cueModPath, cueModExists := cuemods.GetCueModParent(configDir)
		if !cueModExists {
			return fmt.Errorf("guku DevX project not found. Run `guku project init`")
		}

		return cuemods.InstallCore(cueModPath)
	},
}

var validateCmd = &cobra.Command{
	Use:     "validate",
	Aliases: []string{"v"},
	Short:   "Validate configurations",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := project.Validate(configDir); err != nil {
			return err
		}
		return nil
	},
}

var discoverCmd = &cobra.Command{
	Use:     "discover",
	Aliases: []string{"d"},
	Short:   "Discover traits",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := project.Discover(configDir, showTraitDef); err != nil {
			return err
		}
		return nil
	},
}

var genCmd = &cobra.Command{
	Use:   "gen",
	Short: "Generate bare config file",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := project.Generate(configDir); err != nil {
			return err
		}
		return nil
	},
}
