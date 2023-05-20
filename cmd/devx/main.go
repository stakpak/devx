package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/devopzilla/guku-devx/pkg/auth"
	log "github.com/sirupsen/logrus"
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
	noStrict         bool
	verbosity        string
	stdout           bool
	reserve          bool
	tags             []string
)
var server = auth.ServerConfig{}

var version = "DEV"
var commit = "X"

type Version struct {
	Version string `json:"version"`
	Commit  string `json:"commit"`
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&verbosity, "verbosity", "v", "info", "log verbosity *info | debug | error")
	rootCmd.PersistentFlags().BoolVarP(&server.Disable, "offline", "D", false, "disable sending telemetry to the Hub")
	rootCmd.PersistentFlags().StringVarP(&server.Endpoint, "server", "e", auth.DEVX_CLOUD_ENDPOINT, "server endpoint")
	rootCmd.PersistentFlags().StringVarP(&server.Tenant, "tenant", "n", "", "server tenant")
	rootCmd.PersistentFlags().StringVarP(&configDir, "project", "p", ".", "project config dir")
	rootCmd.PersistentFlags().StringVarP(&gitDir, "git", "g", ".", "project git dir")
	rootCmd.PersistentFlags().StringVarP(&stackPath, "stack", "s", "stack", "stack field name in config file")
	rootCmd.PersistentFlags().StringVarP(&buildersPath, "builders", "b", "builders", "builders field name in config file")
	rootCmd.PersistentFlags().BoolVar(&noColor, "no-color", false, "disable colors")
	rootCmd.PersistentFlags().BoolVarP(&noStrict, "no-strict", "S", false, "ignore traits not fulfilled by a builder")
	buildCmd.PersistentFlags().BoolVarP(&reserve, "reserve", "r", false, "reserve build resources")
	buildCmd.PersistentFlags().BoolVarP(&dryRun, "dry-run", "d", false, "output the entire stack after transformation without applying drivers")
	buildCmd.PersistentFlags().BoolVarP(&stdout, "stdout", "o", false, "output result to stdout")
	discoverCmd.PersistentFlags().BoolVarP(&showDefs, "definitions", "d", false, "show definitions")
	discoverCmd.PersistentFlags().BoolVarP(&showTransformers, "transformers", "t", false, "show transformers")
	reserveCmd.PersistentFlags().BoolVarP(&dryRun, "dry-run", "d", false, "attempt reserving stack resources")

	runCmd.PersistentFlags().BoolVar(&runFlags.Verbose, "verbose", false, "enables verbose mode")
	runCmd.PersistentFlags().BoolVar(&runFlags.Parallel, "parallel", false, "executes tasks provided on command line in parallel")
	runCmd.PersistentFlags().BoolVar(&runFlags.List, "list", false, "lists tasks with description of current Taskfile")
	runCmd.PersistentFlags().BoolVar(&runFlags.ListAll, "list-all", false, "lists tasks with or without a description")
	runCmd.PersistentFlags().BoolVar(&runFlags.ListJson, "json", false, "formats task list as json")
	runCmd.PersistentFlags().BoolVar(&runFlags.Status, "status", false, "exits with non-zero exit code if any of the given tasks is not up-to-date")
	runCmd.PersistentFlags().BoolVar(&runFlags.Force, "force", false, "forces execution even when the task is up-to-date")
	runCmd.PersistentFlags().BoolVar(&runFlags.Watch, "watch", false, "enables watch of the given task")
	runCmd.PersistentFlags().BoolVar(&runFlags.Dry, "dry", false, "compiles and prints tasks in the order that they would be run, without executing them")
	runCmd.PersistentFlags().BoolVar(&runFlags.Summary, "summary", false, "show summary about a task")
	runCmd.PersistentFlags().BoolVar(&runFlags.ExitCode, "exit-code", false, "pass-through the exit code of the task command")
	runCmd.PersistentFlags().BoolVar(&runFlags.Color, "color", true, "colored output. Enabled by default. Set flag to false to disable")
	runCmd.PersistentFlags().DurationVar(&runFlags.Interval, "interval", 0, "interval to watch for changes")

	publishModuleCmd.PersistentFlags().StringArrayVarP(&tags, "tag", "t", []string{}, "tags to publish the module with")

	rootCmd.AddCommand(
		buildCmd,
		projectCmd,
		versionCmd,
		diffCmd,
		reserveCmd,
		runCmd,
		loginCmd,
		retireCmd,
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
		publishModuleCmd,
	)

	loginCmd.AddCommand(
		clearCmd,
		infoCmd,
	)
}

var rootCmd = &cobra.Command{
	Use:              "devx",
	Short:            "guku DevX cloud native self-service magic",
	SilenceUsage:     true,
	PersistentPreRun: preRun,
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

func preRun(cmd *cobra.Command, args []string) {
	setupLogging(cmd, args)

	resp, err := http.Get("https://api.github.com/repos/devopzilla/guku-devx/releases?per_page=1")
	if err == nil {
		releases := []struct {
			TagName string `json:"tag_name"`
		}{}

		json.NewDecoder(resp.Body).Decode(&releases)

		if len(releases) > 0 {
			latestVersion := strings.TrimPrefix(releases[0].TagName, "v")
			if latestVersion != version && version != "DEV" {
				log.Infof("A newer version of DevX \"v%s\" is available, please upgrade!\n", latestVersion)
			}
		}

	}
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
