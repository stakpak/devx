package main

import (
	"github.com/spf13/cobra"
	"github.com/stakpak/devx/pkg/taskfile"
)

var runFlags taskfile.RunFlags

var runCmd = &cobra.Command{
	Use:   "run [environment] [task...]",
	Short: "Run taskfile.dev tasks",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		doubleDashPos := cmd.ArgsLenAtDash()
		if err := taskfile.Run(configDir, buildersPath, server, runFlags, args[0], doubleDashPos, args[1:]); err != nil {
			return err
		}
		return nil
	},
}
