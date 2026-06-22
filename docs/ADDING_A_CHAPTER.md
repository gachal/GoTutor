# Adding a Chapter

> English | [中文](ADDING_A_CHAPTER-zh.md)

A chapter is a single learning unit with:
- a **template** — Go skeleton shown to the learner, with `// TODO` markers
- **hints** — bilingual (zh + en) text for each TODO line
- **tests** — `go test` cases the learner's code must pass
- **solution** — reference solution, surfaced to the learner on demand via
  the `GET /api/chapters/:id/solution` endpoint (shown in the chapter
  detail view's "参考答案 / Reference solution" drawer)

Every chapter also carries metadata that drives the home screen:
**Track** (which section it appears in), **Difficulty** (beginner /
intermediate / advanced — informational only, doesn't gate access),
**EstimatedMinutes** (shown on the card), and **Prerequisites**
(informational list of chapter IDs the learner should have completed
first). All four live on the `Chapter` struct in `registry.go`.

This document walks through adding a `time` package exercise.

## 1. Pick an ID and concept

Two-letter IDs work well. Pick a Go concept the chapter teaches:
- `time` — `time.Now`, `time.Sleep`, `time.Duration`
- `sort` — `sort.Slice`, custom `Less`
- `io` — `io.Reader` / `io.Writer` composition

## 2. Create the content directory

```
backend/chapters/content/<id>/
├── template.txt              # skeleton with // TODO
├── hints.yaml                # bilingual hints keyed by line
├── solution/main.txt         # reference solution
└── tests/<name>_test.go.txt  # test files (suffix MUST be .go.txt)
```

`go:embed` skips `.go` files — that's why templates and tests ship as
`.txt`. The `.go.txt` suffix on tests strips to `_test.go` at submit
time so `go test` discovers them.

## 3. Write the template

Open with a one-line `// Chapter N: <title>` header, then `package main`
and the function stubs. Mark TODO gaps:

```go
// Chapter 3: Time Wrapper — initial skeleton.
package main

import (
	"fmt"
	"time"
)

// FormatDuration returns a human string for a Duration, e.g.
// "1h2m3s" → "1 hour 2 minutes 3 seconds".
func FormatDuration(d time.Duration) string {
	// TODO 1: convert d to hours, minutes, seconds.
	return ""
}

func main() {
	fmt.Println(FormatDuration(1 * time.Hour))
}
```

## 4. Write the hints

`hints.yaml` keys hint text by the 1-based line number of the matching
`// TODO`:

```yaml
todos:
  - line: 12
    hint:
      zh: 用 d.Hours(), d.Minutes() 等方法提取各部分。
      en: Use d.Hours(), d.Minutes() etc. to extract each part.
```

Use `grep -n '// TODO' template.txt` to confirm line numbers; the
backend's TODO scan uses `^\s*//\s*TODO\b` so any line whose comment
marker literally starts with `// TODO` will be flagged.

## 5. Write the tests

Tests must be deterministic and fast (<10s — the verifier's timeout).
For chapters that need servers or external state, use `httptest` (see
`urlcheck/tests/urlcheck_test.go.txt`). For pure functions, table-driven
tests work best.

```go
// tests/time_test.go.txt
package main

import (
	"testing"
	"time"
)

func TestFormatDuration(t *testing.T) {
	cases := []struct {
		in   time.Duration
		want string
	}{
		{1 * time.Hour, "1 hour"},
		{1*time.Hour + 30*time.Minute, "1 hour 30 minutes"},
	}
	for _, tc := range cases {
		got := FormatDuration(tc.in)
		if got != tc.want {
			t.Errorf("FormatDuration(%v) = %q, want %q", tc.in, got, tc.want)
		}
	}
}
```

## 6. Write the solution

Drop your reference solution in `solution/main.txt`. It is served to
learners on demand via `GET /api/chapters/:id/solution` (the "Recommended
answer" modal in the chapter detail view), and CI can compile it + run the
tests to verify the chapter is solvable.

## 7. Register the chapter

Edit `backend/chapters/registry.go`. Append to the `registry` slice:

```go
{
	ID:    "time",
	Title: Locale{Zh: "时间格式化", En: "Time Formatting"},
	Description: Locale{
		Zh: "用 time 包将 Duration 转成人类可读的字符串。",
		En: "Convert a time.Duration into a human-readable string.",
	},
	Ordinal:          16, // next free slot — current registry uses 1–15
	Track:            TrackFundamentals, // or TrackConcurrency / TrackGateway
	Difficulty:       DifficultyBeginner, // beginner | intermediate | advanced
	EstimatedMinutes: 12,
	Prerequisites:    []string{"calc"}, // chapter IDs, informational only
	// Optional: extend the verifier whitelist beyond the safe stdlib
	// baseline. time/fmt/etc. are already allowed; you only need this
	// for things like net/http.
	// AllowImports: []string{"net/http"},
	contentDir: "time",
},
```

All chapters are unlocked from the start — learners can explore any chapter
freely (see `HandleListChapters` in `internal/api/chapters.go`, where
`Unlocked` is always `true`). `Ordinal` controls display order within a
track. `Track` controls which home-screen section the card lands in
(fundamentals / concurrency / gateway). `Difficulty` and
`EstimatedMinutes` are informational card decorations.
`Prerequisites` is shown as a soft hint but never gates access.
`Completed` reflects whether the user has ever passed.

## 8. Test end-to-end

```bash
# Rebuild backend (embeds new content/)
make backend-dev

# In another terminal — verify the chapter appears in the list:
curl localhost:8081/api/chapters

# Fetch the template:
curl localhost:8081/api/chapters/time/template

# Test hint lookup:
curl 'localhost:8081/api/chapters/time/hint?line=12'

# Submit the reference solution:
SOLUTION=$(cat backend/chapters/content/time/solution/main.txt | \
           python3 -c "import sys,json; print(json.dumps({'userCode': sys.stdin.read()}))")
curl -X POST localhost:8081/api/chapters/time/submit \
  -H 'Content-Type: application/json' -d "$SOLUTION"
```

The response should have `passed: true` and the chapter's `completed` flag
should flip (all chapters are already unlocked from the start).

## 9. Verify the AST policy

The verifier's default whitelist is in
`backend/internal/verifier/policy.go` (`SafeStdlibImports` +
`DeniedStdlibImports`). If your chapter needs an import not on either
list, add it to the chapter's `AllowImports`. Test rejection:

```bash
# Submit code with a banned import — should be rejected at AST.
BAD=$(python3 -c "import json; print(json.dumps({'userCode': 'package main\nimport \"os/exec\"\nfunc main() {}\n'}))")
curl -X POST localhost:8081/api/chapters/time/submit \
  -H 'Content-Type: application/json' -d "$BAD"
# Expected: {passed: false, output: 'import "os/exec" is not allowed...'}
```

## 10. Document

Add the chapter to the README's chapter list. Done.
