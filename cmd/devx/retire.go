package main

import (
	"github.com/spf13/cobra"
	"github.com/stakpak/devx/pkg/client"
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
