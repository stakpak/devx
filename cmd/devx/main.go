package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	gitDir           string
	configDir        string
	stackPath        string
	buildersPath     string
	showDefs         bool
	showTransformers bool
	dryRun           bool
	noColor          bool
	telemetry        string
	strict           bool
	verbosity        string
	stdout           bool
)

var version = "DEV"
var commit = "X"

type Version struct {
	Version string `json:"version"`
	Commit  string `json:"commit"`
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&verbosity, "verbosity", "v", "info", "log verbosity *info | debug | error")
	rootCmd.PersistentFlags().StringVarP(&telemetry, "telemetry", "T", "", "telemetry endpoint")
	rootCmd.PersistentFlags().StringVarP(&configDir, "project", "p", ".", "project config dir")
	rootCmd.PersistentFlags().StringVarP(&gitDir, "git", "g", ".", "project git dir")
	rootCmd.PersistentFlags().StringVarP(&stackPath, "stack", "s", "stack", "stack field name in config file")
	rootCmd.PersistentFlags().StringVarP(&buildersPath, "builders", "b", "builders", "builders field name in config file")
	rootCmd.PersistentFlags().BoolVar(&noColor, "no-color", false, "disable colors")
	rootCmd.PersistentFlags().BoolVarP(&strict, "strict", "S", false, "make sure all traits are fulfilled by at least one flow")
	buildCmd.PersistentFlags().BoolVarP(&dryRun, "dry-run", "d", false, "output the entire stack after transformation without applying drivers")
	buildCmd.PersistentFlags().BoolVarP(&stdout, "stdout", "o", false, "output result to stdout")
	discoverCmd.PersistentFlags().BoolVarP(&showDefs, "definitions", "d", false, "show definitions")
	discoverCmd.PersistentFlags().BoolVarP(&showTransformers, "transformers", "t", false, "show transformers")
	reserveCmd.PersistentFlags().BoolVarP(&dryRun, "dry-run", "d", false, "attempt reserving stack resources")

	rootCmd.AddCommand(
		buildCmd,
		projectCmd,
		versionCmd,
		diffCmd,
		reserveCmd,
	)

	projectCmd.AddCommand(
		initCmd,
		updateCmd,
		validateCmd,
		discoverCmd,
		genCmd,
		publishCmd,
		importCmd,
	)

	publishCmd.AddCommand(
		publishPolicyCmd,
		publishStackCmd,
		publishCatalogCmd,
	)
}

var rootCmd = &cobra.Command{
	Use:              "devx",
	Short:            "guku DevX cloud native self-service magic",
	SilenceUsage:     true,
	PersistentPreRun: setupLogging,
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
