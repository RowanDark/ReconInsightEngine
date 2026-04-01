package normalize

import "strings"

// normalizeHost lowercases and trims whitespace from a raw hostname or IP address.
func normalizeHost(raw string) string {
	return strings.ToLower(strings.TrimSpace(raw))
}

// deduplicateHosts collapses RawHosts with the same normalized value into a
// single NormalizedHost, merging all source attributions. Output order
// matches first-seen order for determinism.
func deduplicateHosts(raws []RawHost) []NormalizedHost {
	type entry struct {
		sources []Source
	}
	seen := make(map[string]*entry)
	order := make([]string, 0, len(raws))

	for _, r := range raws {
		key := normalizeHost(r.Value)
		if key == "" {
			continue
		}
		if _, ok := seen[key]; !ok {
			seen[key] = &entry{}
			order = append(order, key)
		}
		seen[key].sources = append(seen[key].sources, r.Source)
	}

	result := make([]NormalizedHost, 0, len(order))
	for _, key := range order {
		e := seen[key]
		result = append(result, NormalizedHost{
			Value:   key,
			Sources: e.sources,
		})
	}
	return result
}
