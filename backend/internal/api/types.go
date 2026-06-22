// Package api defines the request/response DTOs for the GoTutor HTTP API.
// These types are the contract the frontend consumes; changes here must
// be mirrored in frontend/src/api/types.ts (Phase 5).
package api

// Chapter is one row in the chapter list returned by GET /api/chapters.
// ID is stable across releases; Title/Description are display-only.
type Chapter struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Ordinal     int    `json:"ordinal"`
	// Track groups this chapter in the UI ("fundamentals" / "concurrency"
	// / "gateway"). Drives the sectioned chapter list. The frontend owns
	// display order; this field carries identity only.
	Track string `json:"track"`
	// Difficulty is "beginner" | "intermediate" | "advanced". Shown on
	// the chapter card. Informational — does NOT gate access.
	Difficulty string `json:"difficulty"`
	// EstimatedMinutes is the approximate time-to-pass, shown on the card.
	EstimatedMinutes int `json:"estimatedMinutes"`
	// Prerequisites lists chapter IDs the learner should have completed
	// first. Informational — does NOT gate access.
	Prerequisites []string `json:"prerequisites"`
	// Completed is true iff the user has ever passed this chapter's verifier.
	Completed bool `json:"completed"`
	// Unlocked is true iff the user is allowed to attempt this chapter
	// (chapter 1 always is; chapter N requires chapter N-1 completed).
	Unlocked bool `json:"unlocked"`
}

// TrackProgress is one track's contribution to ProgressResponse.ByTrack.
type TrackProgress struct {
	Track             string `json:"track"`
	TotalChapters     int    `json:"totalChapters"`
	CompletedChapters int    `json:"completedChapters"`
}

// ProgressResponse is the body of GET /api/progress. Powers the overall
// progress bar, the "continue where you left off" hero, and per-track
// completion counts in the chapter list. LastChapterID is the most
// recently passed chapter (Phase 1 fallback for "last visited" — a
// visits table lands later).
type ProgressResponse struct {
	TotalChapters     int             `json:"totalChapters"`
	CompletedChapters int             `json:"completedChapters"`
	Percent           int             `json:"percent"`
	LastChapterID     string          `json:"lastChapterId"`
	ByTrack           []TrackProgress `json:"byTrack"`
}

// Todo marks one // TODO line in a chapter template. Line is 1-based.
// Hint is the user-facing hint text in the current locale (already
// resolved by the backend based on Accept-Language).
type Todo struct {
	Line int    `json:"line"`
	Hint string `json:"hint"`
}

// Template is the body of GET /api/chapters/:id/template. Code is the
// raw Go source the editor loads as its initial contents.
type Template struct {
	Code  string `json:"code"`
	Todos []Todo `json:"todos"`
}

// SubmitRequest is the body of POST /api/chapters/:id/submit. UserCode
// is the full text of the user's main.go (including their TODO fill-ins).
type SubmitRequest struct {
	UserCode string `json:"userCode"`
}

// SubmitResult is the response from POST /api/chapters/:id/submit.
// Output contains combined stdout+stderr from `go test`, capped at 64KB.
type SubmitResult struct {
	Passed     bool   `json:"passed"`
	Output     string `json:"output"`
	DurationMs int64  `json:"durationMs"`
	// NextChapterUnlocked is true iff this pass newly unlocked the next
	// chapter (false if it was already completed, or if this is the last).
	NextChapterUnlocked bool `json:"nextChapterUnlocked"`
}

// HintResponse is the body of GET /api/chapters/:id/hint?line=N.
type HintResponse struct {
	Text string `json:"text"`
}

// SolutionResponse is the body of GET /api/chapters/:id/solution. Code is
// the reference solution's raw Go source, shown on demand in the chapter
// detail view's answer modal.
type SolutionResponse struct {
	Code string `json:"code"`
}
