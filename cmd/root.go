package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "rei",
	Short: "Recon Engine Intelligence - a reconnaissance data analysis tool",
	Long: `rei (Recon Engine Intelligence) aggregates, normalizes, and analyzes
reconnaissance data from multiple sources to produce actionable intelligence.

Examples:
  rei version            Print version information
  rei ingest --help      Learn how to ingest recon data
  rei analyze --help     Learn how to analyze ingested data`,
	// No-op run so that 'rei' alone prints help rather than erroring.
	Run: func(cmd *cobra.Command, args []string) {
		_ = cmd.Help()
	},
}

// Execute runs the root command and exits with a non-zero code on failure.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	// Silence the default "unknown command" suggestion to keep output clean;
	// cobra already prints a usage message and returns a non-zero exit code.
	rootCmd.SilenceErrors = true
}
