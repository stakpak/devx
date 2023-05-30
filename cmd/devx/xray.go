package main

import (
	"fmt"

	"cuelang.org/go/cue/errors"
	"github.com/devopzilla/guku-devx/pkg/xray"
	"github.com/spf13/cobra"
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
