package normalize

import (
	"fmt"
	"net/url"
	"strings"
)

// normalizeURL parses rawURL and returns a NormalizedURL with the scheme and
// host lowercased. The path is preserved exactly as parsed to avoid altering
// semantic URL meaning. The Raw field holds scheme+host+path with no query
// string; query parameters are handled separately by the params pipeline.
func normalizeURL(raw string) (NormalizedURL, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return NormalizedURL{}, fmt.Errorf("empty URL")
	}

	parsed, err := url.Parse(trimmed)
	if err != nil {
		return NormalizedURL{}, fmt.Errorf("parse URL %q: %w", trimmed, err)
	}
	if parsed.Scheme == "" || parsed.Host == "" {
		return NormalizedURL{}, fmt.Errorf("URL %q missing scheme or host", trimmed)
	}

	scheme := strings.ToLower(parsed.Scheme)
	host := strings.ToLower(parsed.Host)
	path := parsed.Path // paths are case-sensitive; do not alter

	return NormalizedURL{
		Raw:    scheme + "://" + host + path,
		Scheme: scheme,
		Host:   host,
		Path:   path,
	}, nil
}

// deduplicateURLs collapses RawURLs with the same normalized key (scheme+host+path)
// into a single NormalizedURL, merging source attributions. Output order matches
// first-seen order for determinism.
func deduplicateURLs(raws []RawURL) []NormalizedURL {
	type entry struct {
		norm    NormalizedURL
		sources []Source
	}
	seen := make(map[string]*entry)
	order := make([]string, 0, len(raws))

	for _, r := range raws {
		norm, err := normalizeURL(r.Value)
		if err != nil {
			continue
		}
		key := norm.Raw
		if _, ok := seen[key]; !ok {
			seen[key] = &entry{norm: norm}
			order = append(order, key)
		}
		seen[key].sources = append(seen[key].sources, r.Source)
	}

	result := make([]NormalizedURL, 0, len(order))
	for _, key := range order {
		e := seen[key]
		e.norm.Sources = e.sources
		result = append(result, e.norm)
	}
	return result
}
