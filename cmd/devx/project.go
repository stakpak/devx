package main

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"cuelang.org/go/cue/errors"
	"github.com/stakpak/devx/pkg/catalog"
	"github.com/stakpak/devx/pkg/policy"
	"github.com/stakpak/devx/pkg/project"
)

var projectCmd = &cobra.Command{
	Use:   "project",
	Short: "Manage a DevX project",
}

var initCmd = &cobra.Command{
	Use:   "init [module name]",
	Short: "Initialize a project",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return errors.New("1 argument required: init [module name]")
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := project.Init(context.TODO(), configDir, args[0]); err != nil {
			return err
		}
		return nil
	},
}

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update/Install project dependencies",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := project.Update(configDir, server); err != nil {
			return err
		}
		return nil
	},
}

var validateCmd = &cobra.Command{
	Use:     "validate",
	Aliases: []string{"v"},
	Short:   "Validate configurations",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := project.Validate(configDir, stackPath, buildersPath, noStrict); err != nil {
			return fmt.Errorf(errors.Details(err, nil))
		}
		return nil
	},
}

var discoverCmd = &cobra.Command{
	Use:     "discover",
	Aliases: []string{"d"},
	Short:   "Discover traits",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := project.Discover(configDir, showDefs, showTransformers); err != nil {
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

var publishCmd = &cobra.Command{
	Use:   "publish",
	Short: "Publish this project",
}

var publishStackCmd = &cobra.Command{
	Use:   "stack",
	Short: "Publish this stack",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := project.Publish(configDir, stackPath, buildersPath, server); err != nil {
			return err
		}
		return nil
	},
}

var publishPolicyCmd = &cobra.Command{
	Use:   "policy",
	Short: "Publish global policies in this project",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := policy.Publish(configDir, server); err != nil {
			return err
		}
		return nil
	},
}

var publishCatalogCmd = &cobra.Command{
	Use:   "catalog",
	Short: "Publish catalog components in this project",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := catalog.Publish(gitDir, configDir, server); err != nil {
			return err
		}
		return nil
	},
}

var publishModuleCmd = &cobra.Command{
	Use:   "mod",
	Short: "Publish this module",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := catalog.PublishModule(gitDir, configDir, server, tags); err != nil {
			return err
		}
		return nil
	},
}

var importCmd = &cobra.Command{
	Use:   "import [<git repo>@<git revision>]",
	Short: "Import a dependency",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := project.Import(args[0], configDir, server); err != nil {
			return err
		}
		return nil
	},
}
