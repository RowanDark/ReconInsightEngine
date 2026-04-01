// Package ingest loads raw reconnaissance data from input files and converts it
// into a normalize.PipelineInput ready for the normalization stage.
//
// Supported format: plain text, one entry per line.
//   - Lines beginning with '#' or empty lines are silently skipped.
//   - Lines starting with "http://" or "https://" are treated as URLs.
//   - All other non-empty lines are treated as hostnames or IP addresses.
//
// The source label and file path are recorded on every emitted record so that
// downstream stages can trace each finding back to its origin.
package ingest

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/rowandark/reconinsightengine/internal/normalize"
)

// Load reads the file at filePath line by line and returns a PipelineInput.
// source is a human-readable label for the tool or method that produced the
// file (e.g. "nmap", "manual", "ffuf").
func Load(source, filePath string) (normalize.PipelineInput, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return normalize.PipelineInput{}, fmt.Errorf("ingest: open %q: %w", filePath, err)
	}
	defer f.Close()

	src := normalize.Source{
		Tool:      source,
		File:      filePath,
		Timestamp: time.Now().UTC(),
	}

	var input normalize.PipelineInput
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if strings.HasPrefix(line, "http://") || strings.HasPrefix(line, "https://") {
			input.URLs = append(input.URLs, normalize.RawURL{Value: line, Source: src})
		} else {
			input.Hosts = append(input.Hosts, normalize.RawHost{Value: line, Source: src})
		}
	}
	if err := scanner.Err(); err != nil {
		return normalize.PipelineInput{}, fmt.Errorf("ingest: read %q: %w", filePath, err)
	}
	return input, nil
}
