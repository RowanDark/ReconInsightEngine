package normalize

import "time"

// Source tracks where a data point originated.
type Source struct {
	Tool      string    `json:"tool"`
	File      string    `json:"file"`
	Timestamp time.Time `json:"timestamp"`
}

// RawHost is an unnormalized hostname or IP address from ingestion.
type RawHost struct {
	Value  string
	Source Source
}

// RawURL is an unnormalized URL string from ingestion.
type RawURL struct {
	Value  string
	Source Source
}

// RawParameter is an unnormalized query or form parameter from ingestion.
type RawParameter struct {
	Name   string
	Value  string
	URLCTX string // raw URL context; may be empty for standalone parameters
	Source Source
}

// NormalizedHost is a deduplicated, lowercased hostname or IP address.
type NormalizedHost struct {
	Value   string   `json:"value"`
	Sources []Source `json:"sources"`
}

// NormalizedParameter is a deduplicated parameter scoped to a URL context.
type NormalizedParameter struct {
	Name    string   `json:"name"`
	Value   string   `json:"value"`
	URLCTX  string   `json:"url_ctx"` // normalized URL (scheme+host+path)
	Sources []Source `json:"sources"`
}

// NormalizedURL is a deduplicated URL with parsed components and attached parameters.
type NormalizedURL struct {
	Raw        string                `json:"raw"`        // scheme+host+path (no query string)
	Scheme     string                `json:"scheme"`
	Host       string                `json:"host"`
	Path       string                `json:"path"`
	Parameters []NormalizedParameter `json:"parameters"` // parameters scoped to this URL
	Sources    []Source              `json:"sources"`
}

// PipelineInput holds raw data to be normalized.
type PipelineInput struct {
	Hosts      []RawHost
	URLs       []RawURL
	Parameters []RawParameter
}

// PipelineOutput holds the normalized, deduplicated results.
// Parameters contains only standalone entries whose URLCTX does not
// match any NormalizedURL; URL-scoped parameters live on NormalizedURL.Parameters.
type PipelineOutput struct {
	Hosts      []NormalizedHost
	URLs       []NormalizedURL
	Parameters []NormalizedParameter
}
