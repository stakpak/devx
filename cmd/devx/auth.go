package main

import (
	"github.com/devopzilla/guku-devx/pkg/auth"
	"github.com/spf13/cobra"
)

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Login to DevX Server",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := auth.Login(server); err != nil {
			return err
		}
		return nil
	},
}
