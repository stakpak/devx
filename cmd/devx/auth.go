package main

import (
	"github.com/devopzilla/guku-devx/pkg/auth"
	"github.com/spf13/cobra"
)

var loginCmd = &cobra.Command{
	Use:     "auth",
	Short:   "Authenticate to DevX Server",
	Aliases: []string{"login"},
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := auth.Login(server); err != nil {
			return err
		}
		return nil
	},
}

var clearCmd = &cobra.Command{
	Use:   "clear",
	Short: "Clear cached credentials",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := auth.Clear(server); err != nil {
			return err
		}
		return nil
	},
}

var infoCmd = &cobra.Command{
	Use:   "info",
	Short: "Display session information",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := auth.Info(server); err != nil {
			return err
		}
		return nil
	},
}
