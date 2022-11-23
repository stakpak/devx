package main

import (
	"devopzilla.com/guku/internal/client"
	"github.com/spf13/cobra"
)

var doCmd = &cobra.Command{
	Use:   "do [environment]",
	Short: "do DevX magic for the specified environment",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := client.Run(args[0], configDir, stackPath, buildersPath); err != nil {
			return err
		}
		return nil
	},
}
