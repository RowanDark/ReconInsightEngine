package normalize

import (
	"testing"
	"time"
)

// ts is a fixed timestamp used across tests for deterministic source values.
var ts = time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

func src(tool, file string) Source {
	return Source{Tool: tool, File: file, Timestamp: ts}
}

// ── normalizeHost ─────────────────────────────────────────────────────────────

func TestNormalizeHost(t *testing.T) {
	cases := []struct {
		in   string
		want string
	}{
		{"example.com", "example.com"},
		{"EXAMPLE.COM", "example.com"},
		{"  Example.Com  ", "example.com"},
		{"192.168.1.1", "192.168.1.1"},
		{"", ""},
	}
	for _, tc := range cases {
		got := normalizeHost(tc.in)
		if got != tc.want {
			t.Errorf("normalizeHost(%q) = %q; want %q", tc.in, got, tc.want)
		}
	}
}

// ── deduplicateHosts ──────────────────────────────────────────────────────────

func TestDeduplicateHosts_Empty(t *testing.T) {
	got := deduplicateHosts(nil)
	if len(got) != 0 {
		t.Fatalf("expected empty result, got %v", got)
	}
}

func TestDeduplicateHosts_EmptyValueSkipped(t *testing.T) {
	got := deduplicateHosts([]RawHost{{Value: "", Source: src("nmap", "a.xml")}})
	if len(got) != 0 {
		t.Fatalf("expected empty result for blank host, got %v", got)
	}
}

func TestDeduplicateHosts_SingleHost(t *testing.T) {
	in := []RawHost{{Value: "Example.COM", Source: src("nmap", "a.xml")}}
	got := deduplicateHosts(in)
	if len(got) != 1 {
		t.Fatalf("expected 1 host, got %d", len(got))
	}
	if got[0].Value != "example.com" {
		t.Errorf("Value = %q; want %q", got[0].Value, "example.com")
	}
	if len(got[0].Sources) != 1 {
		t.Errorf("Sources len = %d; want 1", len(got[0].Sources))
	}
}

func TestDeduplicateHosts_DuplicatesMergedSources(t *testing.T) {
	s1 := src("nmap", "a.xml")
	s2 := src("masscan", "b.json")
	in := []RawHost{
		{Value: "EXAMPLE.COM", Source: s1},
		{Value: "example.com", Source: s2},
	}
	got := deduplicateHosts(in)
	if len(got) != 1 {
		t.Fatalf("expected 1 host after dedup, got %d", len(got))
	}
	if len(got[0].Sources) != 2 {
		t.Errorf("Sources len = %d; want 2", len(got[0].Sources))
	}
	if got[0].Sources[0] != s1 || got[0].Sources[1] != s2 {
		t.Errorf("source order not preserved: %v", got[0].Sources)
	}
}

func TestDeduplicateHosts_MultipleDistinct(t *testing.T) {
	in := []RawHost{
		{Value: "alpha.com", Source: src("nmap", "a.xml")},
		{Value: "beta.com", Source: src("nmap", "a.xml")},
		{Value: "ALPHA.COM", Source: src("masscan", "b.json")},
	}
	got := deduplicateHosts(in)
	if len(got) != 2 {
		t.Fatalf("expected 2 distinct hosts, got %d", len(got))
	}
	if got[0].Value != "alpha.com" || got[1].Value != "beta.com" {
		t.Errorf("unexpected order: %v", got)
	}
	if len(got[0].Sources) != 2 {
		t.Errorf("alpha.com should have 2 sources, got %d", len(got[0].Sources))
	}
}

// ── normalizeURL ──────────────────────────────────────────────────────────────

func TestNormalizeURL_Basic(t *testing.T) {
	norm, err := normalizeURL("HTTP://EXAMPLE.COM/Path")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if norm.Scheme != "http" {
		t.Errorf("Scheme = %q; want %q", norm.Scheme, "http")
	}
	if norm.Host != "example.com" {
		t.Errorf("Host = %q; want %q", norm.Host, "example.com")
	}
	if norm.Path != "/Path" {
		t.Errorf("Path = %q; want %q (path must not be altered)", norm.Path, "/Path")
	}
	if norm.Raw != "http://example.com/Path" {
		t.Errorf("Raw = %q; want %q", norm.Raw, "http://example.com/Path")
	}
}

func TestNormalizeURL_QueryStrippedFromRaw(t *testing.T) {
	norm, err := normalizeURL("https://example.com/search?q=test&page=2")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Raw must not contain the query string.
	if norm.Raw != "https://example.com/search" {
		t.Errorf("Raw = %q; want query stripped", norm.Raw)
	}
}

func TestNormalizeURL_EmptyError(t *testing.T) {
	_, err := normalizeURL("")
	if err == nil {
		t.Fatal("expected error for empty URL")
	}
}

func TestNormalizeURL_MissingSchemeError(t *testing.T) {
	_, err := normalizeURL("example.com/path")
	if err == nil {
		t.Fatal("expected error for URL without scheme")
	}
}

// ── deduplicateURLs ───────────────────────────────────────────────────────────

func TestDeduplicateURLs_DuplicatesCaseInsensitive(t *testing.T) {
	s1 := src("spider", "run1.txt")
	s2 := src("spider", "run2.txt")
	in := []RawURL{
		{Value: "HTTP://EXAMPLE.COM/page", Source: s1},
		{Value: "http://example.com/page", Source: s2},
	}
	got := deduplicateURLs(in)
	if len(got) != 1 {
		t.Fatalf("expected 1 URL after dedup, got %d", len(got))
	}
	if len(got[0].Sources) != 2 {
		t.Errorf("Sources len = %d; want 2", len(got[0].Sources))
	}
}

func TestDeduplicateURLs_PathCaseSensitive(t *testing.T) {
	// /Page and /page are different URLs (path is case-sensitive).
	in := []RawURL{
		{Value: "http://example.com/Page", Source: src("s", "f")},
		{Value: "http://example.com/page", Source: src("s", "f")},
	}
	got := deduplicateURLs(in)
	if len(got) != 2 {
		t.Fatalf("expected 2 distinct URLs (path is case-sensitive), got %d", len(got))
	}
}

func TestDeduplicateURLs_QueryStringDoesNotAffectKey(t *testing.T) {
	// Same path, different query strings → same normalized URL key.
	in := []RawURL{
		{Value: "http://example.com/page?a=1", Source: src("s", "f1")},
		{Value: "http://example.com/page?b=2", Source: src("s", "f2")},
	}
	got := deduplicateURLs(in)
	if len(got) != 1 {
		t.Fatalf("expected 1 URL (query does not affect key), got %d", len(got))
	}
	if len(got[0].Sources) != 2 {
		t.Errorf("Sources len = %d; want 2", len(got[0].Sources))
	}
}

func TestDeduplicateURLs_InvalidURLSkipped(t *testing.T) {
	in := []RawURL{
		{Value: "://bad", Source: src("s", "f")},
		{Value: "http://example.com/ok", Source: src("s", "f")},
	}
	got := deduplicateURLs(in)
	if len(got) != 1 {
		t.Fatalf("expected 1 valid URL, got %d", len(got))
	}
}

// ── parseParamsFromURL ────────────────────────────────────────────────────────

func TestParseParamsFromURL_Basic(t *testing.T) {
	s := src("spider", "f.txt")
	got := parseParamsFromURL("http://example.com/search?q=hello&page=2", s)
	if len(got) != 2 {
		t.Fatalf("expected 2 params, got %d", len(got))
	}
	// Sorted by name: page, q.
	if got[0].Name != "page" || got[0].Value != "2" {
		t.Errorf("param[0] = {%q,%q}; want {page,2}", got[0].Name, got[0].Value)
	}
	if got[1].Name != "q" || got[1].Value != "hello" {
		t.Errorf("param[1] = {%q,%q}; want {q,hello}", got[1].Name, got[1].Value)
	}
	for _, p := range got {
		if p.URLCTX != "http://example.com/search" {
			t.Errorf("URLCTX = %q; want %q", p.URLCTX, "http://example.com/search")
		}
		if p.Source != s {
			t.Errorf("Source not preserved")
		}
	}
}

func TestParseParamsFromURL_NoParams(t *testing.T) {
	got := parseParamsFromURL("http://example.com/page", src("s", "f"))
	if len(got) != 0 {
		t.Fatalf("expected no params, got %d", len(got))
	}
}

func TestParseParamsFromURL_Deterministic(t *testing.T) {
	// Run twice; order must be identical.
	url := "http://example.com/?z=1&a=2&m=3"
	s := src("s", "f")
	first := parseParamsFromURL(url, s)
	second := parseParamsFromURL(url, s)
	if len(first) != len(second) {
		t.Fatal("different lengths on repeated call")
	}
	for i := range first {
		if first[i].Name != second[i].Name {
			t.Errorf("position %d: %q != %q", i, first[i].Name, second[i].Name)
		}
	}
}

// ── deduplicateParams ─────────────────────────────────────────────────────────

func TestDeduplicateParams_MergesSourcesKeepsFirstValue(t *testing.T) {
	s1 := src("s1", "f1")
	s2 := src("s2", "f2")
	in := []RawParameter{
		{Name: "q", Value: "hello", URLCTX: "http://example.com/search", Source: s1},
		{Name: "q", Value: "world", URLCTX: "http://example.com/search", Source: s2},
	}
	got := deduplicateParams(in)
	if len(got) != 1 {
		t.Fatalf("expected 1 param after dedup, got %d", len(got))
	}
	if got[0].Value != "hello" {
		t.Errorf("Value = %q; want first-seen %q", got[0].Value, "hello")
	}
	if len(got[0].Sources) != 2 {
		t.Errorf("Sources len = %d; want 2", len(got[0].Sources))
	}
}

func TestDeduplicateParams_SameNameDifferentURLCtxAreDistinct(t *testing.T) {
	in := []RawParameter{
		{Name: "id", Value: "1", URLCTX: "http://example.com/a", Source: src("s", "f")},
		{Name: "id", Value: "2", URLCTX: "http://example.com/b", Source: src("s", "f")},
	}
	got := deduplicateParams(in)
	if len(got) != 2 {
		t.Fatalf("expected 2 params (different url_ctx), got %d", len(got))
	}
}

func TestDeduplicateParams_StandaloneEmptyURLCTX(t *testing.T) {
	in := []RawParameter{
		{Name: "token", Value: "abc", URLCTX: "", Source: src("manual", "notes.txt")},
		{Name: "token", Value: "xyz", URLCTX: "", Source: src("manual", "notes2.txt")},
	}
	got := deduplicateParams(in)
	if len(got) != 1 {
		t.Fatalf("expected 1 standalone param after dedup, got %d", len(got))
	}
	if got[0].Value != "abc" {
		t.Errorf("Value = %q; want %q", got[0].Value, "abc")
	}
}

// ── Run (full pipeline) ───────────────────────────────────────────────────────

func TestRun_EmptyInput(t *testing.T) {
	out := Run(PipelineInput{})
	if len(out.Hosts) != 0 || len(out.URLs) != 0 || len(out.Parameters) != 0 {
		t.Errorf("expected all-empty output for empty input, got %+v", out)
	}
}

func TestRun_HostDeduplication(t *testing.T) {
	out := Run(PipelineInput{
		Hosts: []RawHost{
			{Value: "EXAMPLE.COM", Source: src("nmap", "a.xml")},
			{Value: "example.com", Source: src("masscan", "b.json")},
			{Value: "other.com", Source: src("nmap", "a.xml")},
		},
	})
	if len(out.Hosts) != 2 {
		t.Fatalf("expected 2 distinct hosts, got %d", len(out.Hosts))
	}
	if out.Hosts[0].Value != "example.com" {
		t.Errorf("first host = %q; want %q", out.Hosts[0].Value, "example.com")
	}
	if len(out.Hosts[0].Sources) != 2 {
		t.Errorf("example.com Sources len = %d; want 2", len(out.Hosts[0].Sources))
	}
}

func TestRun_URLNormalizationAndParamAttachment(t *testing.T) {
	s := src("spider", "crawl.txt")
	out := Run(PipelineInput{
		URLs: []RawURL{
			{Value: "HTTP://Example.COM/Search?q=test&page=1", Source: s},
		},
	})
	if len(out.URLs) != 1 {
		t.Fatalf("expected 1 URL, got %d", len(out.URLs))
	}
	u := out.URLs[0]
	if u.Scheme != "http" || u.Host != "example.com" || u.Path != "/Search" {
		t.Errorf("unexpected URL components: scheme=%q host=%q path=%q", u.Scheme, u.Host, u.Path)
	}
	if u.Raw != "http://example.com/Search" {
		t.Errorf("Raw = %q; query string must be stripped", u.Raw)
	}
	if len(u.Parameters) != 2 {
		t.Fatalf("expected 2 parameters on URL, got %d", len(u.Parameters))
	}
	// Parameters are sorted by name: page, q.
	if u.Parameters[0].Name != "page" || u.Parameters[0].Value != "1" {
		t.Errorf("param[0] = {%q,%q}; want {page,1}", u.Parameters[0].Name, u.Parameters[0].Value)
	}
	if u.Parameters[1].Name != "q" || u.Parameters[1].Value != "test" {
		t.Errorf("param[1] = {%q,%q}; want {q,test}", u.Parameters[1].Name, u.Parameters[1].Value)
	}
	// URL-scoped params must NOT appear in top-level Parameters.
	if len(out.Parameters) != 0 {
		t.Errorf("URL-scoped params leaked into top-level Parameters: %v", out.Parameters)
	}
}

func TestRun_StandaloneParametersPreserved(t *testing.T) {
	out := Run(PipelineInput{
		Parameters: []RawParameter{
			{Name: "api_key", Value: "secret", URLCTX: "", Source: src("manual", "notes.txt")},
		},
	})
	if len(out.Parameters) != 1 {
		t.Fatalf("expected 1 standalone param, got %d", len(out.Parameters))
	}
	if out.Parameters[0].Name != "api_key" {
		t.Errorf("Name = %q; want %q", out.Parameters[0].Name, "api_key")
	}
}

func TestRun_ExplicitParamTakesPrecedenceOverExtracted(t *testing.T) {
	// Explicitly provided param for same (name, url_ctx) wins over extracted.
	explicit := src("manual", "notes.txt")
	urlSrc := src("spider", "crawl.txt")
	out := Run(PipelineInput{
		URLs: []RawURL{
			{Value: "http://example.com/page?q=extracted", Source: urlSrc},
		},
		Parameters: []RawParameter{
			{Name: "q", Value: "explicit", URLCTX: "http://example.com/page", Source: explicit},
		},
	})
	if len(out.URLs) != 1 {
		t.Fatalf("expected 1 URL")
	}
	params := out.URLs[0].Parameters
	if len(params) != 1 {
		t.Fatalf("expected 1 param on URL, got %d", len(params))
	}
	if params[0].Value != "explicit" {
		t.Errorf("Value = %q; want %q (explicit should take precedence)", params[0].Value, "explicit")
	}
	if len(params[0].Sources) != 2 {
		t.Errorf("Sources len = %d; want 2 (both sources merged)", len(params[0].Sources))
	}
}

func TestRun_SourceAttributionIntact(t *testing.T) {
	s1 := src("nmap", "scan1.xml")
	s2 := src("nmap", "scan2.xml")
	out := Run(PipelineInput{
		Hosts: []RawHost{
			{Value: "example.com", Source: s1},
			{Value: "EXAMPLE.COM", Source: s2},
		},
	})
	if len(out.Hosts) != 1 {
		t.Fatalf("expected 1 host")
	}
	if out.Hosts[0].Sources[0] != s1 || out.Hosts[0].Sources[1] != s2 {
		t.Errorf("source attribution not intact: %v", out.Hosts[0].Sources)
	}
}

func TestRun_Deterministic(t *testing.T) {
	// Running the same input twice must produce identical output.
	input := PipelineInput{
		Hosts: []RawHost{
			{Value: "beta.com", Source: src("s", "f")},
			{Value: "alpha.com", Source: src("s", "f")},
			{Value: "BETA.COM", Source: src("s2", "f2")},
		},
		URLs: []RawURL{
			{Value: "http://beta.com/?z=1&a=2", Source: src("s", "f")},
			{Value: "http://alpha.com/page", Source: src("s", "f")},
		},
	}
	out1 := Run(input)
	out2 := Run(input)

	if len(out1.Hosts) != len(out2.Hosts) {
		t.Fatal("hosts length differs across runs")
	}
	for i := range out1.Hosts {
		if out1.Hosts[i].Value != out2.Hosts[i].Value {
			t.Errorf("host[%d] differs: %q vs %q", i, out1.Hosts[i].Value, out2.Hosts[i].Value)
		}
	}
	if len(out1.URLs) != len(out2.URLs) {
		t.Fatal("URLs length differs across runs")
	}
	for i := range out1.URLs {
		if out1.URLs[i].Raw != out2.URLs[i].Raw {
			t.Errorf("url[%d] differs: %q vs %q", i, out1.URLs[i].Raw, out2.URLs[i].Raw)
		}
		if len(out1.URLs[i].Parameters) != len(out2.URLs[i].Parameters) {
			t.Errorf("url[%d] param count differs", i)
		}
	}
}

func TestRun_MultiSourceURLDedup(t *testing.T) {
	s1 := src("spider", "run1.txt")
	s2 := src("burp", "run2.xml")
	out := Run(PipelineInput{
		URLs: []RawURL{
			{Value: "https://example.com/api", Source: s1},
			{Value: "HTTPS://EXAMPLE.COM/api", Source: s2},
		},
	})
	if len(out.URLs) != 1 {
		t.Fatalf("expected 1 URL after dedup, got %d", len(out.URLs))
	}
	if len(out.URLs[0].Sources) != 2 {
		t.Errorf("Sources len = %d; want 2", len(out.URLs[0].Sources))
	}
	if out.URLs[0].Sources[0] != s1 || out.URLs[0].Sources[1] != s2 {
		t.Errorf("source order not preserved: %v", out.URLs[0].Sources)
	}
}
