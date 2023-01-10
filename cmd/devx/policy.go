package main

import (
	"devopzilla.com/guku/internal/policy"
	"github.com/spf13/cobra"
)

var policyCmd = &cobra.Command{
	Use:   "policy",
	Short: "Manage global resource policies",
}

var policyPublishCmd = &cobra.Command{
	Use:   "publish",
	Short: "Publish global policies in this project",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := policy.Publish(configDir, telemetry); err != nil {
			return err
		}
		return nil
	},
}
