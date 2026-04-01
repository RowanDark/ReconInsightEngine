package heuristics

import (
	"strings"

	"github.com/rowandark/reconinsightengine/internal/normalize"
)

type pathRule struct {
	pattern     string
	category    Category
	title       string
	score       int
	confidence  float64
	explanation string
}

var pathRules = []pathRule{
	// Auth
	{"/login", CategoryAuth, "Login endpoint detected", 70, 0.8,
		"The path contains /login, a common authentication entry point that may be vulnerable to credential attacks."},
	{"/auth", CategoryAuth, "Auth endpoint detected", 70, 0.8,
		"The path contains /auth, likely an authentication or OAuth handler worth inspecting."},
	{"/signin", CategoryAuth, "Sign-in endpoint detected", 70, 0.8,
		"The path contains /signin, a common authentication entry point."},
	{"/reset", CategoryAuth, "Password reset endpoint detected", 75, 0.8,
		"The path contains /reset, typically a password-reset flow that may expose account enumeration or token weaknesses."},
	// API
	{"/api/", CategoryAPI, "API endpoint detected", 60, 0.7,
		"The path contains /api/, indicating a programmatic interface that may expose additional functionality."},
	{"/v1/", CategoryAPI, "Versioned API endpoint detected", 60, 0.7,
		"The path contains /v1/, suggesting a versioned API that may have older, less-hardened endpoints."},
	{"/graphql", CategoryAPI, "GraphQL endpoint detected", 65, 0.75,
		"The path contains /graphql, a query interface that may allow schema introspection or over-fetching."},
	// Admin
	{"/admin", CategoryAdmin, "Admin interface detected", 85, 0.85,
		"The path contains /admin, an administrative interface that is a high-value target if accessible."},
	{"/internal", CategoryAdmin, "Internal endpoint detected", 80, 0.8,
		"The path contains /internal, suggesting functionality intended only for internal use."},
	// File handling
	{"/upload", CategoryFileHandling, "File upload endpoint detected", 80, 0.8,
		"The path contains /upload, a file-upload handler that may accept dangerous file types."},
	{"/import", CategoryFileHandling, "Import endpoint detected", 75, 0.75,
		"The path contains /import, a data-import handler that may be vulnerable to malicious input files."},
	{"/export", CategoryFileHandling, "Export endpoint detected", 65, 0.7,
		"The path contains /export, a data-export handler that may leak sensitive records."},
}

func applyPathHeuristics(urls []normalize.NormalizedURL) []Insight {
	var insights []Insight
	for _, u := range urls {
		lowerPath := strings.ToLower(u.Path)
		for _, rule := range pathRules {
			if strings.Contains(lowerPath, rule.pattern) {
				insights = append(insights, Insight{
					Title:       rule.title,
					Category:    rule.category,
					Score:       rule.score,
					Confidence:  rule.confidence,
					Explanation: rule.explanation,
					Evidence:    Evidence{URL: u.Raw, Pattern: rule.pattern},
				})
			}
		}
	}
	return insights
}
