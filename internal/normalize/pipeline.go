// Package normalize provides a normalization pipeline for reconnaissance data.
//
// The pipeline accepts raw hosts, URLs, and parameters from multiple ingestion
// sources and produces a deduplicated, consistently structured output while
// preserving full source attribution on every record.
//
// Normalization rules:
//   - Hostnames are lowercased and whitespace-trimmed.
//   - URL schemes and hosts are lowercased; paths are preserved as-is.
//   - Query parameters are extracted from URLs and keyed by (name, url_ctx).
//   - Deduplication is strict equality — no fuzzy matching is performed.
//   - Output order is first-seen for deterministic results given the same input.
package normalize

// Run executes the normalization pipeline over the provided PipelineInput and
// returns a PipelineOutput with normalized, deduplicated hosts, URLs, and
// parameters.
//
// Processing steps:
//  1. Normalize and deduplicate hosts.
//  2. Normalize and deduplicate URLs.
//  3. Extract query parameters from all raw URLs, merge with any explicitly
//     provided RawParameters, then deduplicate by (name, url_ctx).
//  4. Attach URL-scoped parameters to their parent NormalizedURL.
//  5. Parameters whose URLCTX does not match any NormalizedURL are placed in
//     PipelineOutput.Parameters (standalone / orphaned parameters).
func Run(input PipelineInput) PipelineOutput {
	// Step 1: hosts.
	hosts := deduplicateHosts(input.Hosts)

	// Step 2: URLs.
	urls := deduplicateURLs(input.URLs)

	// Step 3: collect all raw parameters.
	// URL-extracted params come first so that explicitly provided params with
	// the same key override them (first-seen wins in deduplicateParams).
	// Reverse the order here: explicit inputs take precedence over extracted.
	allRaw := make([]RawParameter, 0, len(input.Parameters))
	allRaw = append(allRaw, input.Parameters...)
	for _, ru := range input.URLs {
		allRaw = append(allRaw, parseParamsFromURL(ru.Value, ru.Source)...)
	}
	params := deduplicateParams(allRaw)

	// Build a lookup from normalized URL key → parameters.
	paramsByURL := make(map[string][]NormalizedParameter, len(params))
	for _, p := range params {
		if p.URLCTX != "" {
			paramsByURL[p.URLCTX] = append(paramsByURL[p.URLCTX], p)
		}
	}

	// Step 4: attach parameters to their parent NormalizedURL.
	urlKeySet := make(map[string]bool, len(urls))
	for i, u := range urls {
		urls[i].Parameters = paramsByURL[u.Raw]
		urlKeySet[u.Raw] = true
	}

	// Step 5: collect standalone parameters (URLCTX empty or no matching URL).
	var standalone []NormalizedParameter
	for _, p := range params {
		if !urlKeySet[p.URLCTX] {
			standalone = append(standalone, p)
		}
	}

	return PipelineOutput{
		Hosts:      hosts,
		URLs:       urls,
		Parameters: standalone,
	}
}
