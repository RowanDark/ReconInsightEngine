package heuristics_test

import (
	"testing"

	"github.com/rowandark/reconinsightengine/internal/heuristics"
	"github.com/rowandark/reconinsightengine/internal/normalize"
)

// --- helpers -----------------------------------------------------------------

func makeURL(raw, path string) normalize.NormalizedURL {
	return normalize.NormalizedURL{
		Raw:    raw,
		Scheme: "https",
		Host:   "example.com",
		Path:   path,
	}
}

func makeURLWithParam(raw, path, paramName string) normalize.NormalizedURL {
	u := makeURL(raw, path)
	u.Parameters = []normalize.NormalizedParameter{
		{Name: paramName, Value: "x", URLCTX: raw},
	}
	return u
}

func findByPattern(insights []heuristics.Insight, pattern string, cat heuristics.Category) bool {
	for _, ins := range insights {
		if ins.Evidence.Pattern == pattern && ins.Category == cat {
			return true
		}
	}
	return false
}

// --- path heuristics ---------------------------------------------------------

func TestPathHeuristicsAuthPatterns(t *testing.T) {
	cases := []struct {
		path    string
		pattern string
	}{
		{"/login", "/login"},
		{"/auth/token", "/auth"},
		{"/user/signin", "/signin"},
		{"/password/reset", "/reset"},
	}
	for _, tc := range cases {
		t.Run(tc.path, func(t *testing.T) {
			out := heuristics.Run(normalize.PipelineOutput{
				URLs: []normalize.NormalizedURL{makeURL("https://example.com"+tc.path, tc.path)},
			})
			if !findByPattern(out, tc.pattern, heuristics.CategoryAuth) {
				t.Errorf("expected auth insight for pattern %q on path %q, got none", tc.pattern, tc.path)
			}
		})
	}
}

func TestPathHeuristicsAPIPatterns(t *testing.T) {
	cases := []struct {
		path    string
		pattern string
	}{
		{"/api/users", "/api/"},
		{"/v1/orders", "/v1/"},
		{"/graphql", "/graphql"},
	}
	for _, tc := range cases {
		t.Run(tc.path, func(t *testing.T) {
			out := heuristics.Run(normalize.PipelineOutput{
				URLs: []normalize.NormalizedURL{makeURL("https://example.com"+tc.path, tc.path)},
			})
			if !findByPattern(out, tc.pattern, heuristics.CategoryAPI) {
				t.Errorf("expected API insight for pattern %q on path %q, got none", tc.pattern, tc.path)
			}
		})
	}
}

func TestPathHeuristicsAdminPatterns(t *testing.T) {
	cases := []struct {
		path    string
		pattern string
	}{
		{"/admin/users", "/admin"},
		{"/internal/status", "/internal"},
	}
	for _, tc := range cases {
		t.Run(tc.path, func(t *testing.T) {
			out := heuristics.Run(normalize.PipelineOutput{
				URLs: []normalize.NormalizedURL{makeURL("https://example.com"+tc.path, tc.path)},
			})
			if !findByPattern(out, tc.pattern, heuristics.CategoryAdmin) {
				t.Errorf("expected admin insight for pattern %q on path %q, got none", tc.pattern, tc.path)
			}
		})
	}
}

func TestPathHeuristicsFilePatterns(t *testing.T) {
	cases := []struct {
		path    string
		pattern string
	}{
		{"/upload/avatar", "/upload"},
		{"/data/import", "/import"},
		{"/report/export", "/export"},
	}
	for _, tc := range cases {
		t.Run(tc.path, func(t *testing.T) {
			out := heuristics.Run(normalize.PipelineOutput{
				URLs: []normalize.NormalizedURL{makeURL("https://example.com"+tc.path, tc.path)},
			})
			if !findByPattern(out, tc.pattern, heuristics.CategoryFileHandling) {
				t.Errorf("expected file-handling insight for pattern %q on path %q, got none", tc.pattern, tc.path)
			}
		})
	}
}

func TestPathHeuristicsCaseInsensitive(t *testing.T) {
	u := normalize.NormalizedURL{
		Raw:    "https://example.com/Admin/Dashboard",
		Scheme: "https",
		Host:   "example.com",
		Path:   "/Admin/Dashboard",
	}
	out := heuristics.Run(normalize.PipelineOutput{URLs: []normalize.NormalizedURL{u}})
	if !findByPattern(out, "/admin", heuristics.CategoryAdmin) {
		t.Error("expected admin insight for mixed-case path /Admin/Dashboard, got none")
	}
}

func TestNoMatchOnUnrelatedPaths(t *testing.T) {
	u := makeURL("https://example.com/about", "/about")
	out := heuristics.Run(normalize.PipelineOutput{URLs: []normalize.NormalizedURL{u}})
	if len(out) != 0 {
		t.Errorf("expected no insights for /about, got %d", len(out))
	}
}

// --- parameter heuristics ----------------------------------------------------

func TestParamHeuristicsAllRedirectParams(t *testing.T) {
	redirectParams := []string{"redirect", "url", "next", "return", "dest", "callback"}
	for _, name := range redirectParams {
		t.Run(name, func(t *testing.T) {
			out := heuristics.Run(normalize.PipelineOutput{
				URLs: []normalize.NormalizedURL{makeURLWithParam("https://example.com/page", "/page", name)},
			})
			if !findByPattern(out, name, heuristics.CategoryRedirect) {
				t.Errorf("expected redirect insight for param %q, got none", name)
			}
		})
	}
}

func TestParamHeuristicsStandaloneParameters(t *testing.T) {
	out := heuristics.Run(normalize.PipelineOutput{
		Parameters: []normalize.NormalizedParameter{
			{Name: "redirect", Value: "/home", URLCTX: ""},
		},
	})
	if !findByPattern(out, "redirect", heuristics.CategoryRedirect) {
		t.Error("expected redirect insight for standalone parameter, got none")
	}
}

func TestParamHeuristicsCaseInsensitive(t *testing.T) {
	out := heuristics.Run(normalize.PipelineOutput{
		URLs: []normalize.NormalizedURL{makeURLWithParam("https://example.com/page", "/page", "REDIRECT")},
	})
	if !findByPattern(out, "redirect", heuristics.CategoryRedirect) {
		t.Error("expected redirect insight for uppercase param REDIRECT, got none")
	}
}

func TestNoMatchOnUnrelatedParams(t *testing.T) {
	out := heuristics.Run(normalize.PipelineOutput{
		URLs: []normalize.NormalizedURL{makeURLWithParam("https://example.com/page", "/page", "page")},
	})
	if len(out) != 0 {
		t.Errorf("expected no insights for param 'page', got %d", len(out))
	}
}

// --- deduplication -----------------------------------------------------------

func TestDeduplicationIdenticalURLs(t *testing.T) {
	u := makeURL("https://example.com/admin", "/admin")
	out := heuristics.Run(normalize.PipelineOutput{
		URLs: []normalize.NormalizedURL{u, u},
	})
	count := 0
	for _, ins := range out {
		if ins.Evidence.Pattern == "/admin" && ins.Evidence.URL == "https://example.com/admin" {
			count++
		}
	}
	if count != 1 {
		t.Errorf("expected 1 insight for duplicate URL, got %d", count)
	}
}

// --- insight field completeness ----------------------------------------------

func TestInsightFieldsArePopulated(t *testing.T) {
	out := heuristics.Run(normalize.PipelineOutput{
		URLs: []normalize.NormalizedURL{makeURL("https://example.com/login", "/login")},
	})
	if len(out) == 0 {
		t.Fatal("expected at least one insight for /login")
	}
	ins := out[0]
	if ins.Title == "" {
		t.Error("Insight.Title must not be empty")
	}
	if ins.Category == "" {
		t.Error("Insight.Category must not be empty")
	}
	if ins.Score == 0 {
		t.Error("Insight.Score must not be zero")
	}
	if ins.Confidence == 0 {
		t.Error("Insight.Confidence must not be zero")
	}
	if ins.Explanation == "" {
		t.Error("Insight.Explanation must not be empty")
	}
	if ins.Evidence.URL == "" {
		t.Error("Insight.Evidence.URL must not be empty")
	}
	if ins.Evidence.Pattern == "" {
		t.Error("Insight.Evidence.Pattern must not be empty")
	}
}

// --- edge cases --------------------------------------------------------------

func TestEmptyInputProducesNoInsights(t *testing.T) {
	out := heuristics.Run(normalize.PipelineOutput{})
	if len(out) != 0 {
		t.Errorf("expected no insights for empty input, got %d", len(out))
	}
}

func TestMultiplePatternsOnSingleURL(t *testing.T) {
	// /api/v1/ should trigger both /api/ and /v1/ rules.
	u := makeURL("https://example.com/api/v1/items", "/api/v1/items")
	out := heuristics.Run(normalize.PipelineOutput{URLs: []normalize.NormalizedURL{u}})
	hasAPI := findByPattern(out, "/api/", heuristics.CategoryAPI)
	hasV1 := findByPattern(out, "/v1/", heuristics.CategoryAPI)
	if !hasAPI {
		t.Error("expected /api/ insight on path /api/v1/items")
	}
	if !hasV1 {
		t.Error("expected /v1/ insight on path /api/v1/items")
	}
}
