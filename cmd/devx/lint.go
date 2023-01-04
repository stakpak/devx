package main

import (
	"fmt"

	"cuelang.org/go/cue/errors"
	"devopzilla.com/guku/internal/client"
	"github.com/spf13/cobra"
)

var lintCmd = &cobra.Command{
	Use:   "lint [environment]",
	Short: "lint project",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := client.Lint(args[0], configDir, stackPath, buildersPath); err != nil {
			return fmt.Errorf(errors.Details(err, nil))
		}
		return nil
	},
}
