package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var ingestCmd = &cobra.Command{
	Use:   "ingest",
	Short: "Ingest reconnaissance data from a source",
	Long: `Ingest normalizes raw reconnaissance data from various tools and
stores it for later analysis.

Examples:
  rei ingest --source nmap --file scan.xml
  rei ingest --source masscan --file results.json`,
	// Placeholder: real ingestion logic is added in a future milestone.
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("ingest: not yet implemented")
		return nil
	},
}

func init() {
	ingestCmd.Flags().StringP("source", "s", "", "Data source type (e.g. nmap, masscan)")
	ingestCmd.Flags().StringP("file", "f", "", "Path to the input file")
	rootCmd.AddCommand(ingestCmd)
}
