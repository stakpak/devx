package main

import (
	"fmt"

	"cuelang.org/go/cue/errors"
	"github.com/spf13/cobra"
	"github.com/stakpak/devx/pkg/xray"
)

var xrayCmd = &cobra.Command{
	Use:     "ray",
	Short:   "Let devx analyze your source code and auto-generate a stack",
	Aliases: []string{"x", "detect"},
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := xray.Run(configDir, server); err != nil {
			return fmt.Errorf(errors.Details(err, nil))
		}
		return nil
	},
}
