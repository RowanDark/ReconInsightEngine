package scoring

import "github.com/rowandark/reconinsightengine/internal/heuristics"

// ScoredInsight groups all heuristic findings for a single URL and attaches a
// combined weighted score with an explainable list of contributing factors.
type ScoredInsight struct {
	URL      string               `json:"url"`
	Score    int                  `json:"score"`    // 0–100, capped
	Reasons  []string             `json:"reasons"`  // e.g. ["admin endpoint (+30)", "upload endpoint (+25)"]
	Insights []heuristics.Insight `json:"insights"` // underlying raw findings
}
