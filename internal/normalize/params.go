package normalize

import (
	"net/url"
	"sort"
	"strings"
)

// parseParamsFromURL extracts query parameters from rawURL and returns them as
// RawParameters. The URLCTX on each parameter is set to the normalized URL key
// (scheme+host+path, no query string). Parameter names are iterated in sorted
// order so extraction is deterministic regardless of map iteration order.
func parseParamsFromURL(rawURL string, source Source) []RawParameter {
	trimmed := strings.TrimSpace(rawURL)
	parsed, err := url.Parse(trimmed)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return nil
	}

	scheme := strings.ToLower(parsed.Scheme)
	host := strings.ToLower(parsed.Host)
	ctx := scheme + "://" + host + parsed.Path

	query := parsed.Query()
	if len(query) == 0 {
		return nil
	}

	// Sort names for deterministic output.
	names := make([]string, 0, len(query))
	for name := range query {
		names = append(names, name)
	}
	sort.Strings(names)

	var params []RawParameter
	for _, name := range names {
		for _, value := range query[name] {
			params = append(params, RawParameter{
				Name:   name,
				Value:  value,
				URLCTX: ctx,
				Source: source,
			})
		}
	}
	return params
}

// deduplicateParams collapses RawParameters with the same (name, URLCTX) key
// into a single NormalizedParameter, keeping the first observed value and
// merging all source attributions. Output order matches first-seen order for
// determinism.
func deduplicateParams(raws []RawParameter) []NormalizedParameter {
	type key struct {
		name   string
		urlctx string
	}
	type entry struct {
		value   string
		sources []Source
	}
	seen := make(map[key]*entry)
	order := make([]key, 0, len(raws))

	for _, r := range raws {
		k := key{name: r.Name, urlctx: r.URLCTX}
		if _, ok := seen[k]; !ok {
			seen[k] = &entry{value: r.Value}
			order = append(order, k)
		}
		seen[k].sources = append(seen[k].sources, r.Source)
	}

	result := make([]NormalizedParameter, 0, len(order))
	for _, k := range order {
		e := seen[k]
		result = append(result, NormalizedParameter{
			Name:    k.name,
			Value:   e.value,
			URLCTX:  k.urlctx,
			Sources: e.sources,
		})
	}
	return result
}
