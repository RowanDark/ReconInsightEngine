// Package scoring assigns weighted scores to heuristic insights and sorts
// results by descending risk.
//
// Base weights per category:
//
//	admin endpoint      → +30
//	auth endpoint       → +25
//	upload endpoint     → +25
//	API endpoint        → +20
//	redirect parameter  → +20
//
// When multiple categories match the same URL their weights stack. The final
// score is capped at 100. Each category is counted once per URL regardless of
// how many individual patterns matched within that category.
//
// Design constraints:
//   - No config-driven weights.
//   - No hidden modifiers.
//   - Scoring logic is transparent and small.
package scoring

import (
	"fmt"
	"sort"

	"github.com/rowandark/reconinsightengine/internal/heuristics"
)

// categoryWeight maps each insight category to its base score contribution.
var categoryWeight = map[heuristics.Category]int{
	heuristics.CategoryAdmin:        30,
	heuristics.CategoryAuth:         25,
	heuristics.CategoryFileHandling: 25,
	heuristics.CategoryAPI:          20,
	heuristics.CategoryRedirect:     20,
}

// categoryLabel maps each category to a human-readable reason string.
var categoryLabel = map[heuristics.Category]string{
	heuristics.CategoryAdmin:        "admin endpoint",
	heuristics.CategoryAuth:         "auth endpoint",
	heuristics.CategoryFileHandling: "upload endpoint",
	heuristics.CategoryAPI:          "API endpoint",
	heuristics.CategoryRedirect:     "redirect parameter",
}

// Score groups insights by URL, computes a stacked weighted score for each URL,
// and returns the results sorted by descending score.
//
// Each category contributes its weight at most once per URL, even if multiple
// patterns within that category matched.
func Score(insights []heuristics.Insight) []ScoredInsight {
	// Preserve first-seen URL order for deterministic output before sorting.
	urlOrder := make([]string, 0)
	byURL := make(map[string][]heuristics.Insight)

	for _, ins := range insights {
		url := ins.Evidence.URL
		if _, exists := byURL[url]; !exists {
			urlOrder = append(urlOrder, url)
		}
		byURL[url] = append(byURL[url], ins)
	}

	result := make([]ScoredInsight, 0, len(urlOrder))

	for _, url := range urlOrder {
		group := byURL[url]

		seenCat := make(map[heuristics.Category]bool)
		total := 0
		reasons := make([]string, 0)

		for _, ins := range group {
			cat := ins.Category
			if seenCat[cat] {
				continue
			}
			seenCat[cat] = true
			w := categoryWeight[cat]
			total += w
			reasons = append(reasons, fmt.Sprintf("%s (+%d)", categoryLabel[cat], w))
		}

		if total > 100 {
			total = 100
		}

		result = append(result, ScoredInsight{
			URL:      url,
			Score:    total,
			Reasons:  reasons,
			Insights: group,
		})
	}

	sort.SliceStable(result, func(i, j int) bool {
		return result[i].Score > result[j].Score
	})

	return result
}
