package main

import (
	"fmt"

	"cuelang.org/go/cue/errors"
	"github.com/spf13/cobra"
	"github.com/stakpak/devx/pkg/client"
)

var diffCmd = &cobra.Command{
	Use:   "diff [environment] [target git revision]",
	Short: "Diff the current stack with that @ target (e.g. HEAD, commit, tag).",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := client.Diff(args[1], args[0], configDir, stackPath, buildersPath, server, noStrict); err != nil {
			return fmt.Errorf(errors.Details(err, nil))
		}
		return nil
	},
}
