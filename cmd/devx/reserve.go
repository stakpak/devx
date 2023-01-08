package main

import (
	"devopzilla.com/guku/internal/client"
	"github.com/spf13/cobra"
)

var reserveCmd = &cobra.Command{
	Use:   "reserve [build id]",
	Short: "Reserve build resources",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := client.Reserve(args[0], telemetry); err != nil {
			return err
		}
		return nil
	},
}
