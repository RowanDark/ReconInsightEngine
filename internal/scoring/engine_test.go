package scoring_test

import (
	"testing"

	"github.com/rowandark/reconinsightengine/internal/heuristics"
	"github.com/rowandark/reconinsightengine/internal/scoring"
)

// insight builds a minimal Insight for test use.
func insight(url, pattern string, cat heuristics.Category) heuristics.Insight {
	return heuristics.Insight{
		Title:    "test",
		Category: cat,
		Score:    60,
		Evidence: heuristics.Evidence{URL: url, Pattern: pattern},
	}
}

// --- single-category scores --------------------------------------------------

func TestSingleAdminEndpointScore(t *testing.T) {
	ins := []heuristics.Insight{insight("https://example.com/admin", "/admin", heuristics.CategoryAdmin)}
	got := scoring.Score(ins)
	if len(got) != 1 {
		t.Fatalf("expected 1 ScoredInsight, got %d", len(got))
	}
	if got[0].Score != 30 {
		t.Errorf("admin score: want 30, got %d", got[0].Score)
	}
}

func TestSingleAuthEndpointScore(t *testing.T) {
	ins := []heuristics.Insight{insight("https://example.com/login", "/login", heuristics.CategoryAuth)}
	got := scoring.Score(ins)
	if got[0].Score != 25 {
		t.Errorf("auth score: want 25, got %d", got[0].Score)
	}
}

func TestSingleAPIEndpointScore(t *testing.T) {
	ins := []heuristics.Insight{insight("https://example.com/api/v1", "/api/", heuristics.CategoryAPI)}
	got := scoring.Score(ins)
	if got[0].Score != 20 {
		t.Errorf("API score: want 20, got %d", got[0].Score)
	}
}

func TestSingleUploadEndpointScore(t *testing.T) {
	ins := []heuristics.Insight{insight("https://example.com/upload", "/upload", heuristics.CategoryFileHandling)}
	got := scoring.Score(ins)
	if got[0].Score != 25 {
		t.Errorf("upload score: want 25, got %d", got[0].Score)
	}
}

func TestSingleRedirectParamScore(t *testing.T) {
	ins := []heuristics.Insight{insight("https://example.com/page", "redirect", heuristics.CategoryRedirect)}
	got := scoring.Score(ins)
	if got[0].Score != 20 {
		t.Errorf("redirect score: want 20, got %d", got[0].Score)
	}
}

// --- stacking ----------------------------------------------------------------

func TestAdminPlusUploadStacks(t *testing.T) {
	// /admin/upload triggers both admin (+30) and upload (+25) → 55
	ins := []heuristics.Insight{
		insight("https://example.com/admin/upload", "/admin", heuristics.CategoryAdmin),
		insight("https://example.com/admin/upload", "/upload", heuristics.CategoryFileHandling),
	}
	got := scoring.Score(ins)
	if len(got) != 1 {
		t.Fatalf("expected 1 ScoredInsight for same URL, got %d", len(got))
	}
	if got[0].Score != 55 {
		t.Errorf("admin+upload score: want 55, got %d", got[0].Score)
	}
	if len(got[0].Reasons) != 2 {
		t.Errorf("expected 2 reasons, got %d", len(got[0].Reasons))
	}
}

func TestAllCategoriesStack(t *testing.T) {
	// admin(30) + auth(25) + file(25) + api(20) + redirect(20) = 120, capped to 100
	url := "https://example.com/admin/upload"
	ins := []heuristics.Insight{
		insight(url, "/admin", heuristics.CategoryAdmin),
		insight(url, "/login", heuristics.CategoryAuth),
		insight(url, "/upload", heuristics.CategoryFileHandling),
		insight(url, "/api/", heuristics.CategoryAPI),
		insight(url, "redirect", heuristics.CategoryRedirect),
	}
	got := scoring.Score(ins)
	if got[0].Score != 100 {
		t.Errorf("all-category score: want 100 (capped), got %d", got[0].Score)
	}
}

func TestSameCategoryCountedOnce(t *testing.T) {
	// Two API insights on same URL (/api/ and /v1/) → only counted once → 20
	url := "https://example.com/api/v1/items"
	ins := []heuristics.Insight{
		insight(url, "/api/", heuristics.CategoryAPI),
		insight(url, "/v1/", heuristics.CategoryAPI),
	}
	got := scoring.Score(ins)
	if got[0].Score != 20 {
		t.Errorf("duplicate-category score: want 20, got %d", got[0].Score)
	}
	if len(got[0].Reasons) != 1 {
		t.Errorf("expected 1 reason for single category, got %d: %v", len(got[0].Reasons), got[0].Reasons)
	}
}

// --- sorting -----------------------------------------------------------------

func TestSortedByDescendingScore(t *testing.T) {
	ins := []heuristics.Insight{
		insight("https://example.com/api/users", "/api/", heuristics.CategoryAPI),          // 20
		insight("https://example.com/admin", "/admin", heuristics.CategoryAdmin),            // 30
		insight("https://example.com/login", "/login", heuristics.CategoryAuth),             // 25
	}
	got := scoring.Score(ins)
	if len(got) != 3 {
		t.Fatalf("expected 3 ScoredInsights, got %d", len(got))
	}
	if got[0].Score != 30 {
		t.Errorf("first result score: want 30, got %d", got[0].Score)
	}
	if got[1].Score != 25 {
		t.Errorf("second result score: want 25, got %d", got[1].Score)
	}
	if got[2].Score != 20 {
		t.Errorf("third result score: want 20, got %d", got[2].Score)
	}
}

func TestHigherRiskRanksFirst(t *testing.T) {
	// admin+upload (55) should rank above auth-only (25)
	ins := []heuristics.Insight{
		insight("https://example.com/login", "/login", heuristics.CategoryAuth),
		insight("https://example.com/admin/upload", "/admin", heuristics.CategoryAdmin),
		insight("https://example.com/admin/upload", "/upload", heuristics.CategoryFileHandling),
	}
	got := scoring.Score(ins)
	if got[0].URL != "https://example.com/admin/upload" {
		t.Errorf("highest-risk URL should be first, got %q", got[0].URL)
	}
	if got[0].Score <= got[1].Score {
		t.Errorf("first score (%d) should be greater than second (%d)", got[0].Score, got[1].Score)
	}
}

// --- reasons -----------------------------------------------------------------

func TestReasonsAreExplainable(t *testing.T) {
	ins := []heuristics.Insight{
		insight("https://example.com/admin", "/admin", heuristics.CategoryAdmin),
	}
	got := scoring.Score(ins)
	if len(got[0].Reasons) == 0 {
		t.Fatal("expected at least one reason")
	}
	// Reason must contain the label and weight
	reason := got[0].Reasons[0]
	if reason != "admin endpoint (+30)" {
		t.Errorf("unexpected reason format: %q", reason)
	}
}

func TestStackingReasonsListsAllFactors(t *testing.T) {
	url := "https://example.com/admin/upload"
	ins := []heuristics.Insight{
		insight(url, "/admin", heuristics.CategoryAdmin),
		insight(url, "/upload", heuristics.CategoryFileHandling),
	}
	got := scoring.Score(ins)
	found := map[string]bool{}
	for _, r := range got[0].Reasons {
		found[r] = true
	}
	if !found["admin endpoint (+30)"] {
		t.Error("expected 'admin endpoint (+30)' in reasons")
	}
	if !found["upload endpoint (+25)"] {
		t.Error("expected 'upload endpoint (+25)' in reasons")
	}
}

// --- edge cases --------------------------------------------------------------

func TestEmptyInputReturnsEmpty(t *testing.T) {
	got := scoring.Score(nil)
	if len(got) != 0 {
		t.Errorf("expected empty result for nil input, got %d", len(got))
	}
}

func TestInsightsAttachedToScoredInsight(t *testing.T) {
	ins := []heuristics.Insight{
		insight("https://example.com/admin", "/admin", heuristics.CategoryAdmin),
	}
	got := scoring.Score(ins)
	if len(got[0].Insights) != 1 {
		t.Errorf("expected 1 underlying insight, got %d", len(got[0].Insights))
	}
}

func TestURLFieldMatchesEvidence(t *testing.T) {
	ins := []heuristics.Insight{
		insight("https://example.com/login", "/login", heuristics.CategoryAuth),
	}
	got := scoring.Score(ins)
	if got[0].URL != "https://example.com/login" {
		t.Errorf("URL mismatch: want %q, got %q", "https://example.com/login", got[0].URL)
	}
}
