package heuristics

// Category classifies the type of heuristic finding.
type Category string

const (
	CategoryAuth         Category = "auth"
	CategoryAPI          Category = "api"
	CategoryAdmin        Category = "admin"
	CategoryFileHandling Category = "file_handling"
	CategoryRedirect     Category = "redirect"
)

// Evidence captures where and how the heuristic matched.
type Evidence struct {
	URL     string `json:"url"`
	Pattern string `json:"pattern"`
}

// Insight represents a single heuristic finding on normalized reconnaissance data.
type Insight struct {
	Title       string   `json:"title"`
	Category    Category `json:"category"`
	Score       int      `json:"score"`
	Confidence  float64  `json:"confidence"`
	Explanation string   `json:"explanation"`
	Evidence    Evidence `json:"evidence"`
}
