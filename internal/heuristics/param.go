package heuristics

import (
	"strings"

	"github.com/rowandark/reconinsightengine/internal/normalize"
)

type paramRule struct {
	name        string
	title       string
	score       int
	confidence  float64
	explanation string
}

var paramRules = []paramRule{
	{"redirect", "Open redirect parameter detected", 85, 0.85,
		"The parameter 'redirect' is commonly used to control post-action navigation and may be exploitable for open redirects."},
	{"url", "URL parameter detected", 80, 0.8,
		"The parameter 'url' may control server-side or client-side navigation and is a common target for open redirect or SSRF attacks."},
	{"next", "Next-redirect parameter detected", 80, 0.8,
		"The parameter 'next' is commonly used for post-login redirects and may be exploitable for open redirects."},
	{"return", "Return parameter detected", 80, 0.8,
		"The parameter 'return' is commonly used for redirect flows and may be exploitable for open redirects."},
	{"dest", "Destination parameter detected", 80, 0.8,
		"The parameter 'dest' may control navigation destination and is a common open redirect target."},
	{"callback", "Callback parameter detected", 75, 0.75,
		"The parameter 'callback' may influence JSONP callbacks or redirect targets and warrants further inspection."},
}

func applyParamHeuristics(urls []normalize.NormalizedURL, standalone []normalize.NormalizedParameter) []Insight {
	var insights []Insight
	for _, u := range urls {
		for _, p := range u.Parameters {
			insights = append(insights, matchParam(p.Name, u.Raw)...)
		}
	}
	for _, p := range standalone {
		insights = append(insights, matchParam(p.Name, p.URLCTX)...)
	}
	return insights
}

func matchParam(name, urlCtx string) []Insight {
	lower := strings.ToLower(name)
	var insights []Insight
	for _, rule := range paramRules {
		if lower == rule.name {
			insights = append(insights, Insight{
				Title:       rule.title,
				Category:    CategoryRedirect,
				Score:       rule.score,
				Confidence:  rule.confidence,
				Explanation: rule.explanation,
				Evidence:    Evidence{URL: urlCtx, Pattern: rule.name},
			})
		}
	}
	return insights
}
