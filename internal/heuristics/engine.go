// Package heuristics provides first-pass heuristics for detecting potentially
// interesting endpoints and parameters in normalized reconnaissance data.
//
// Run accepts a PipelineOutput from the normalize package and returns a
// deduplicated slice of Insights. Each insight carries a human-readable title,
// category, placeholder score, confidence value, explanation, and evidence
// linking the finding back to the matched URL and pattern.
//
// Design constraints:
//   - No network calls.
//   - No regular expressions — plain substring and equality checks only.
//   - No ML/AI logic.
package heuristics

import "github.com/rowandark/reconinsightengine/internal/normalize"

// Run analyzes the provided normalized pipeline output and returns a
// deduplicated slice of Insights ordered by first detection.
func Run(output normalize.PipelineOutput) []Insight {
	var raw []Insight
	raw = append(raw, applyPathHeuristics(output.URLs)...)
	raw = append(raw, applyParamHeuristics(output.URLs, output.Parameters)...)
	return deduplicate(raw)
}

// deduplicate removes insights with identical (URL, Pattern) evidence pairs,
// preserving first-seen ordering.
func deduplicate(insights []Insight) []Insight {
	seen := make(map[string]bool, len(insights))
	out := make([]Insight, 0, len(insights))
	for _, ins := range insights {
		key := ins.Evidence.URL + "\x00" + ins.Evidence.Pattern
		if !seen[key] {
			seen[key] = true
			out = append(out, ins)
		}
	}
	return out
}
