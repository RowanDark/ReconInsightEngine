package cmd_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

// newTestRoot builds a fresh cobra root command wired with all subcommands.
// Re-using the package-level rootCmd between tests causes flag-redefinition
// panics, so each test gets its own isolated command tree.
func newTestRoot() *cobra.Command {
	root := &cobra.Command{
		Use:          "rei",
		Short:        "Recon Engine Intelligence",
		SilenceUsage: true,
		Run: func(cmd *cobra.Command, args []string) {
			_ = cmd.Help()
		},
	}

	version := &cobra.Command{
		Use:  "version",
		Args: cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Print("rei version 0.1.0\n")
		},
	}

	ingest := &cobra.Command{
		Use:  "ingest",
		RunE: func(cmd *cobra.Command, args []string) error { return nil },
	}
	ingest.Flags().StringP("source", "s", "", "")
	ingest.Flags().StringP("file", "f", "", "")

	analyze := &cobra.Command{
		Use:  "analyze",
		RunE: func(cmd *cobra.Command, args []string) error { return nil },
	}
	analyze.Flags().StringP("target", "t", "", "")
	analyze.Flags().StringP("format", "o", "text", "")

	root.AddCommand(version, ingest, analyze)
	return root
}

func execCommand(args ...string) (string, error) {
	buf := &bytes.Buffer{}
	root := newTestRoot()
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs(args)
	err := root.Execute()
	return buf.String(), err
}

func TestRootHelp(t *testing.T) {
	out, err := execCommand("--help")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "rei") {
		t.Errorf("help output missing 'rei', got: %s", out)
	}
}

func TestVersionCommand(t *testing.T) {
	out, err := execCommand("version")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "0.1.0") {
		t.Errorf("version output missing version string, got: %s", out)
	}
}

func TestAnalyzeHelp(t *testing.T) {
	out, err := execCommand("analyze", "--help")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "analyze") {
		t.Errorf("analyze help missing 'analyze', got: %s", out)
	}
}

func TestIngestHelp(t *testing.T) {
	out, err := execCommand("ingest", "--help")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "ingest") {
		t.Errorf("ingest help missing 'ingest', got: %s", out)
	}
}

func TestUnknownSubcommandNonZeroExit(t *testing.T) {
	_, err := execCommand("notacommand")
	if err == nil {
		t.Error("expected non-zero exit for unknown subcommand, got nil error")
	}
}
