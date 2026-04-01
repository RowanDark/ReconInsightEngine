package ingest_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/rowandark/reconinsightengine/internal/ingest"
)

// writeTemp writes content to a temporary file and returns its path.
func writeTemp(t *testing.T, content string) string {
	t.Helper()
	f, err := os.CreateTemp(t.TempDir(), "ingest-*.txt")
	if err != nil {
		t.Fatalf("create temp file: %v", err)
	}
	if _, err := f.WriteString(content); err != nil {
		t.Fatalf("write temp file: %v", err)
	}
	if err := f.Close(); err != nil {
		t.Fatalf("close temp file: %v", err)
	}
	return f.Name()
}

func TestLoad_URLsAndHosts(t *testing.T) {
	path := writeTemp(t, `
# comment line
https://example.com/login
http://example.com/api/v1/users
example.com
192.168.1.1
`)
	input, err := ingest.Load("test", path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got, want := len(input.URLs), 2; got != want {
		t.Errorf("URLs: got %d, want %d", got, want)
	}
	if got, want := len(input.Hosts), 2; got != want {
		t.Errorf("Hosts: got %d, want %d", got, want)
	}
}

func TestLoad_SourceLabel(t *testing.T) {
	path := writeTemp(t, "https://example.com/admin\n")
	input, err := ingest.Load("ffuf", path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(input.URLs) == 0 {
		t.Fatal("expected at least one URL")
	}
	if got, want := input.URLs[0].Source.Tool, "ffuf"; got != want {
		t.Errorf("Source.Tool: got %q, want %q", got, want)
	}
	if got := input.URLs[0].Source.File; got != path {
		t.Errorf("Source.File: got %q, want %q", got, path)
	}
}

func TestLoad_SkipsCommentsAndBlanks(t *testing.T) {
	path := writeTemp(t, `
# this is a comment

# another comment
https://example.com/upload
`)
	input, err := ingest.Load("test", path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got, want := len(input.URLs), 1; got != want {
		t.Errorf("URLs: got %d, want %d", got, want)
	}
	if len(input.Hosts) != 0 {
		t.Errorf("expected no hosts, got %d", len(input.Hosts))
	}
}

func TestLoad_EmptyFile(t *testing.T) {
	path := writeTemp(t, "")
	input, err := ingest.Load("test", path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(input.URLs) != 0 || len(input.Hosts) != 0 {
		t.Errorf("expected empty input, got %+v", input)
	}
}

func TestLoad_FileNotFound(t *testing.T) {
	_, err := ingest.Load("test", filepath.Join(t.TempDir(), "nonexistent.txt"))
	if err == nil {
		t.Error("expected error for missing file, got nil")
	}
}

func TestLoad_HttpAndHttpsAreBothURLs(t *testing.T) {
	path := writeTemp(t, "http://plain.example.com/page\nhttps://secure.example.com/page\n")
	input, err := ingest.Load("test", path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got, want := len(input.URLs), 2; got != want {
		t.Errorf("URLs: got %d, want %d", got, want)
	}
	if len(input.Hosts) != 0 {
		t.Errorf("expected no hosts, got %d", len(input.Hosts))
	}
}

func TestLoad_TimestampSet(t *testing.T) {
	path := writeTemp(t, "https://example.com/\n")
	input, err := ingest.Load("test", path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if input.URLs[0].Source.Timestamp.IsZero() {
		t.Error("expected non-zero timestamp on source")
	}
}
