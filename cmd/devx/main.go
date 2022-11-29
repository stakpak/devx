package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	configDir        string
	stackPath        string
	buildersPath     string
	showDefs         bool
	showTransformers bool
	dryRun           bool
)

var version = "DEV"
var commit = "X"

type Version struct {
	Version string `json:"version"`
	Commit  string `json:"commit"`
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&configDir, "project", "p", ".", "project config dir")
	buildCmd.PersistentFlags().StringVarP(&stackPath, "stack", "s", "stack", "stack field name in config file")
	buildCmd.PersistentFlags().StringVarP(&buildersPath, "builders", "b", "builders", "builders field name in config file")
	buildCmd.PersistentFlags().BoolVarP(&dryRun, "dry-run", "d", false, "output the entire stack after transformation without applying drivers")
	discoverCmd.PersistentFlags().BoolVarP(&showDefs, "definitions", "d", false, "show definitions")
	discoverCmd.PersistentFlags().BoolVarP(&showTransformers, "transformers", "t", false, "show transformers")

	rootCmd.AddCommand(
		buildCmd,
		projectCmd,
		versionCmd,
	)

	projectCmd.AddCommand(
		initCmd,
		updateCmd,
		validateCmd,
		discoverCmd,
		genCmd,
	)
}

var rootCmd = &cobra.Command{
	Use:          "devx",
	Short:        "guku DevX cloud native self-service magic",
	SilenceUsage: true,
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of guku DevX",
	RunE: func(cmd *cobra.Command, args []string) error {
		encoded, err := json.Marshal(Version{
			Version: version,
			Commit:  commit,
		})
		if err != nil {
			return err
		}
		fmt.Println(string(encoded))
		return nil
	},
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
