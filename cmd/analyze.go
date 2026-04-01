package cmd

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/rowandark/reconinsightengine/internal/heuristics"
	"github.com/rowandark/reconinsightengine/internal/ingest"
	"github.com/rowandark/reconinsightengine/internal/normalize"
	"github.com/rowandark/reconinsightengine/internal/scoring"
	"github.com/spf13/cobra"
)

var analyzeCmd = &cobra.Command{
	Use:   "analyze",
	Short: "Run the full analysis pipeline on a recon input file",
	Long: `Analyze executes the full pipeline: ingest → normalize → heuristics → scoring.

Input is a plain-text file with one URL or hostname per line.
Lines beginning with '#' are treated as comments and ignored.

Examples:
  rei analyze --file urls.txt
  rei analyze --file urls.txt --format json
  rei analyze --file urls.txt --source ffuf --debug`,
	RunE: runAnalyze,
}

func init() {
	analyzeCmd.Flags().StringP("file", "f", "", "Path to input file (required)")
	analyzeCmd.Flags().StringP("source", "s", "manual", "Source label for the input data (e.g. nmap, ffuf)")
	analyzeCmd.Flags().StringP("format", "o", "text", "Output format: text, json")
	analyzeCmd.Flags().Bool("debug", false, "Log stage transitions to stderr")
	_ = analyzeCmd.MarkFlagRequired("file")
	rootCmd.AddCommand(analyzeCmd)
}

func runAnalyze(cmd *cobra.Command, _ []string) error {
	filePath, _ := cmd.Flags().GetString("file")
	source, _ := cmd.Flags().GetString("source")
	format, _ := cmd.Flags().GetString("format")
	debug, _ := cmd.Flags().GetBool("debug")

	logger := log.New(os.Stderr, "[rei] ", 0)
	logStage := func(stage, msg string) {
		if debug {
			logger.Printf("%-12s %s", stage, msg)
		}
	}

	// Stage 1: ingest — load raw records from the input file.
	logStage("ingest", fmt.Sprintf("loading %q (source=%s)", filePath, source))
	input, err := ingest.Load(source, filePath)
	if err != nil {
		return fmt.Errorf("ingest stage: %w", err)
	}
	logStage("ingest", fmt.Sprintf("%d URLs, %d hosts loaded", len(input.URLs), len(input.Hosts)))

	// Stage 2: normalize — deduplicate and structure the raw records.
	logStage("normalize", "deduplicating and normalizing records")
	normalized := normalize.Run(input)
	logStage("normalize", fmt.Sprintf("%d URLs, %d hosts after deduplication", len(normalized.URLs), len(normalized.Hosts)))

	// Stage 3: heuristics — match patterns to surface interesting endpoints.
	logStage("heuristics", "running path and parameter pattern rules")
	insights := heuristics.Run(normalized)
	logStage("heuristics", fmt.Sprintf("%d insights found", len(insights)))

	// Stage 4: scoring — weight and rank insights by risk.
	logStage("scoring", "computing weighted scores")
	scored := scoring.Score(insights)
	logStage("scoring", fmt.Sprintf("%d scored URLs", len(scored)))

	// Output results in the requested format.
	switch format {
	case "json":
		return outputJSON(cmd, scored)
	default:
		outputText(cmd, scored)
		return nil
	}
}

func outputJSON(cmd *cobra.Command, scored []scoring.ScoredInsight) error {
	enc := json.NewEncoder(cmd.OutOrStdout())
	enc.SetIndent("", "  ")
	return enc.Encode(scored)
}

func outputText(cmd *cobra.Command, scored []scoring.ScoredInsight) {
	out := cmd.OutOrStdout()
	if len(scored) == 0 {
		fmt.Fprintln(out, "No findings.")
		return
	}
	for _, s := range scored {
		fmt.Fprintf(out, "score:%3d  %s\n", s.Score, s.URL)
		for _, r := range s.Reasons {
			fmt.Fprintf(out, "           - %s\n", r)
		}
	}
}
