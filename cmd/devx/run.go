package main

import (
	"devopzilla.com/guku/internal/taskfile"
	"github.com/spf13/cobra"
)

var runFlags taskfile.RunFlags

var runCmd = &cobra.Command{
	Use:   "run [environment] [task...]",
	Short: "Run taskfile.dev tasks",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		doubleDashPos := cmd.ArgsLenAtDash()
		if err := taskfile.Run(configDir, buildersPath, runFlags, args[0], doubleDashPos, args[1:]); err != nil {
			return err
		}
		return nil
	},
}
