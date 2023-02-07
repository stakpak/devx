package main

import (
	"devopzilla.com/guku/internal/workflows"
	"github.com/spf13/cobra"
)

var runFlags workflows.RunFlags

var runCmd = &cobra.Command{
	Use:   "run [environment] [task...]",
	Short: "Run taskfile.dev tasks",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := workflows.Run(configDir, buildersPath, runFlags, args[0], args[1:]); err != nil {
			return err
		}
		return nil
	},
}
