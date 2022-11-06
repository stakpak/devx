package main

import (
	"github.com/spf13/cobra"
)

var projectCmd = &cobra.Command{
	Use:   "project",
	Short: "Manage a DevX project",
	RunE: func(cmd *cobra.Command, args []string) error {
		// if err := client.Run(args[0]); err != nil {
		// 	return err
		// }
		return nil
	},
}

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a project",
	RunE: func(cmd *cobra.Command, args []string) error {
		// if err := client.Run(args[0]); err != nil {
		// 	return err
		// }
		return nil
	},
}

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update/Install project dependencies",
	RunE: func(cmd *cobra.Command, args []string) error {
		// if err := client.Run(args[0]); err != nil {
		// 	return err
		// }
		return nil
	},
}
