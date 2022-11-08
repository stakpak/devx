package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	configDir string
)

func init() {
	rootCmd.PersistentFlags().StringVarP(&configDir, "project", "p", ".", "project config dir")
	rootCmd.AddCommand(
		doCmd,
		projectCmd,
	)

	projectCmd.AddCommand(
		initCmd,
		updateCmd,
	)
}

var rootCmd = &cobra.Command{
	Use:   "guku",
	Short: "guku DevX cloud native self-service magic",
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
