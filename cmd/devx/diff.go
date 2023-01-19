package main

import (
	"fmt"

	"cuelang.org/go/cue/errors"
	"devopzilla.com/guku/internal/client"
	"github.com/spf13/cobra"
)

var diffCmd = &cobra.Command{
	Use:   "diff [target revision] [environment]",
	Short: "Diff the current stack with that @ target (e.g. HEAD, commit, tag).",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := client.Diff(args[0], args[1], configDir, stackPath, buildersPath, strict); err != nil {
			return fmt.Errorf(errors.Details(err, nil))
		}
		return nil
	},
}
