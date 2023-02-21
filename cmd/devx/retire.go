package main

import (
	"devopzilla.com/guku-devx/pkg/client"
	"github.com/spf13/cobra"
)

var retireCmd = &cobra.Command{
	Use:   "retire [build id]",
	Short: "Retire build resources",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := client.Retire(args[0], server); err != nil {
			return err
		}
		return nil
	},
}
