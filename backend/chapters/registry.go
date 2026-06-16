// Package chapters owns the chapter registry: a list of chapters, each
// backed by embedded files under content/. The registry is the single
// source of truth for what chapters exist; the api package reads it to
// serve templates, hints, and (Phase 3) verifier policies.
package chapters

import (
	"fmt"
	"io/fs"
	"regexp"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

// todoPattern matches a line whose comment marker IS a TODO (so prose
// mentions like "fill in the // TODO" don't false-positive). Allows
// leading whitespace, then `//` followed by optional spaces and `TODO`.
var todoPattern = regexp.MustCompile(`^\s*//\s*TODO\b`)

// Hint is one bilingual hint entry from a chapter's hints.yaml.
type Hint struct {
	Line int    `yaml:"line"`
	Hint Locale `yaml:"hint"`
}

// Locale is a bilingual string. The handler selects zh or en based on
// the Accept-Language header.
type Locale struct {
	Zh string `yaml:"zh"`
	En string `yaml:"en"`
}

// HintsFile is the top-level shape of hints.yaml.
type HintsFile struct {
	Todos []Hint `yaml:"todos"`
}

// Chapter describes one learning chapter. Static metadata is hardcoded
// in registry.go; dynamic content (template code, hints, tests, solution)
// is loaded from the embedded filesystem lazily.
type Chapter struct {
	ID          string
	Title       Locale
	Description Locale
	Ordinal     int

	// AllowImports is the verifier whitelist (Phase 3). Empty means
	// "default safe stdlib subset"; populated adds extras like "net/http".
	AllowImports []string

	// AllowListen permits net.Listen / http.ListenAndServe calls in
	// the user's code. Default false; urlcheck keeps it false too
	// (client-only).
	AllowListen bool

	// contentDir is the path under the embedded FS, e.g. "calc".
	contentDir string
}

// registry is the static list. Order matters — ordinals drive unlock gating.
var registry = []Chapter{
	{
		ID:          "calc",
		Title:       Locale{Zh: "命令行计算器", En: "CLI Calculator"},
		Description: Locale{
			Zh: "解析命令行参数，实现一个支持 +、-、*、/ 的简单计算器。",
			En: "Parse command-line arguments and build a simple calculator supporting +, -, *, /.",
		},
		Ordinal:    1,
		contentDir: "calc",
	},
	{
		ID:          "urlcheck",
		Title:       Locale{Zh: "并发 URL 检查器", En: "Concurrent URL Checker"},
		Description: Locale{
			Zh: "用 goroutine、channel 和 sync.WaitGroup 并发探测多个 URL 的状态码。",
			En: "Probe the status code of many URLs concurrently using goroutines, channels, and sync.WaitGroup.",
		},
		Ordinal:      2,
		AllowImports: []string{"net/http", "sync", "time"},
		contentDir:   "urlcheck",
	},
	// Chapters 3–11 are drawn from the Go knowledge points used in AiDeptus,
	// an LLM API gateway (../AiDeptus). Each chapter maps to a real pattern
	// from that codebase: error classification → interface/strategy →
	// concurrency primitives → rate limiting → circuit breaking → HTTP retry
	// → SSE streaming.
	{
		ID:          "errs",
		Title:       Locale{Zh: "错误处理", En: "Error Handling"},
		Description: Locale{
			Zh: "用哨兵错误、errors.Is/As 和 %w 包裹分类错误链——网关据此映射 HTTP 状态码。",
			En: "Classify a wrapped error chain with sentinel errors, errors.Is/As, and %w — how a gateway maps errors to HTTP status.",
		},
		Ordinal:    3,
		contentDir: "errs",
	},
	{
		ID:          "lb",
		Title:       Locale{Zh: "接口与策略模式", En: "Interfaces & Strategy"},
		Description: Locale{
			Zh: "定义 Selector 接口，实现轮询与加权轮询两种负载均衡策略。",
			En: "Define a Selector interface and implement round-robin and weighted load-balancing strategies.",
		},
		Ordinal:    4,
		contentDir: "lb",
	},
	{
		ID:          "pool",
		Title:       Locale{Zh: "并发求和", En: "Concurrent Sum"},
		Description: Locale{
			Zh: "用 goroutine、sync.WaitGroup 和 sync.Mutex 把切片分给多个 worker 并发求和。",
			En: "Split a slice across goroutines and merge the partial sums with sync.WaitGroup and sync.Mutex.",
		},
		Ordinal:    5,
		contentDir: "pool",
	},
	{
		ID:          "chan",
		Title:       Locale{Zh: "channel 与 select", En: "Channels & select"},
		Description: Locale{
			Zh: "用缓冲 channel 传值，select 同时监听数据与 ctx.Done()，超时提前返回。",
			En: "Pipe values through a buffered channel, select on both data and ctx.Done(), and bail out on timeout.",
		},
		Ordinal:    6,
		contentDir: "chan",
	},
	{
		ID:          "ctx",
		Title:       Locale{Zh: "context 超时取消", En: "context Cancellation"},
		Description: Locale{
			Zh: "用 context.WithTimeout 派生超时上下文，让慢任务响应 ctx.Done() 并返回 ctx.Err()。",
			En: "Derive a timeout context with context.WithTimeout and let slow tasks honor ctx.Done() and report ctx.Err().",
		},
		Ordinal:    7,
		contentDir: "ctx",
	},
	{
		ID:          "bucket",
		Title:       Locale{Zh: "令牌桶限流", En: "Token Bucket Limiter"},
		Description: Locale{
			Zh: "手写令牌桶：按速率懒补充令牌、容量封顶，Allow 判断是否放行。",
			En: "Implement a token bucket by hand: lazy refill at a fixed rate, capacity cap, Allow decides admission.",
		},
		Ordinal:    8,
		contentDir: "bucket",
	},
	{
		ID:          "breaker",
		Title:       Locale{Zh: "熔断器状态机", En: "Circuit Breaker"},
		Description: Locale{
			Zh: "实现 Closed→Open→HalfOpen 熔断状态机，用注入时间驱动确定性转移。",
			En: "Implement a Closed→Open→HalfOpen circuit breaker, driven by injected time for deterministic transitions.",
		},
		Ordinal:    9,
		contentDir: "breaker",
	},
	{
		ID:          "retry",
		Title:       Locale{Zh: "HTTP 重试与退避", En: "HTTP Retry & Backoff"},
		Description: Locale{
			Zh: "对不稳定的上游做指数退避重试，退避期间 context 取消能立即中断。",
			En: "Retry a flaky upstream with exponential backoff that bails out the moment context is cancelled.",
		},
		Ordinal:      10,
		AllowImports: []string{"net/http"},
		contentDir:   "retry",
	},
	{
		ID:          "sse",
		Title:       Locale{Zh: "SSE 流式转发", En: "SSE Streaming"},
		Description: Locale{
			Zh: "用 http.Flusher 逐行转发上游 SSE 流并即时 flush——LLM 网关的看家本领。",
			En: "Forward an upstream SSE stream line-by-line with http.Flusher — the heart of an LLM gateway.",
		},
		Ordinal:      11,
		AllowImports: []string{"net/http"},
		contentDir:   "sse",
	},
}

// List returns all chapters sorted by Ordinal. The returned slice is a
// copy so callers can't mutate the package-level registry.
func List() []Chapter {
	out := make([]Chapter, len(registry))
	copy(out, registry)
	sort.Slice(out, func(i, j int) bool { return out[i].Ordinal < out[j].Ordinal })
	return out
}

// Get returns the chapter with the given ID, or ok=false if missing.
func Get(id string) (Chapter, bool) {
	for _, c := range registry {
		if c.ID == id {
			return c, true
		}
	}
	return Chapter{}, false
}

// TemplateCode returns the contents of `<contentDir>/template.txt`.
// This is the skeleton shown to the learner. We use .txt because go:embed
// explicitly excludes files ending in .go — the template is Go source the
// user reads and edits, but it ships as a string resource, not compiled.
func (c Chapter) TemplateCode() (string, error) {
	b, err := fs.ReadFile(contentFS, c.contentDir+"/template.txt")
	if err != nil {
		return "", fmt.Errorf("read template for %s: %w", c.ID, err)
	}
	return string(b), nil
}

// SolutionCode returns the contents of `<contentDir>/solution/main.txt` —
// the reference solution shown on demand in the chapter detail view. Like
// the template it ships as .txt to dodge go:embed's .go exclusion.
func (c Chapter) SolutionCode() (string, error) {
	b, err := fs.ReadFile(contentFS, c.contentDir+"/solution/main.txt")
	if err != nil {
		return "", fmt.Errorf("read solution for %s: %w", c.ID, err)
	}
	return string(b), nil
}

// Hints loads and parses `<contentDir>/hints.yaml`.
func (c Chapter) Hints() (HintsFile, error) {
	b, err := fs.ReadFile(contentFS, c.contentDir+"/hints.yaml")
	if err != nil {
		return HintsFile{}, fmt.Errorf("read hints for %s: %w", c.ID, err)
	}
	var hf HintsFile
	if err := yaml.Unmarshal(b, &hf); err != nil {
		return HintsFile{}, fmt.Errorf("parse hints.yaml for %s: %w", c.ID, err)
	}
	return hf, nil
}

// HintForLine returns the bilingual hint at the given 1-based line, or
// ok=false if no hint is defined for that line.
func (c Chapter) HintForLine(line int) (Hint, bool) {
	hf, err := c.Hints()
	if err != nil {
		return Hint{}, false
	}
	for _, h := range hf.Todos {
		if h.Line == line {
			return h, true
		}
	}
	return Hint{}, false
}

// TestFiles returns the chapter's test files keyed by their on-disk name.
// Files in `<contentDir>/tests/` ship as `<name>_test.go.txt` — the `.go.txt`
// suffix works around go:embed skipping `.go` files. Stripping the trailing
// `.txt` yields `<name>_test.go`, which `go test` discovers normally.
//
// Example: content/calc/tests/calculator_test.go.txt
//        → {"calculator_test.go": <text>}
func (c Chapter) TestFiles() (map[string]string, error) {
	testDir := c.contentDir + "/tests"
	entries, err := fs.ReadDir(contentFS, testDir)
	if err != nil {
		return nil, fmt.Errorf("read tests dir for %s: %w", c.ID, err)
	}
	out := map[string]string{}
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if !strings.HasSuffix(name, ".txt") {
			continue
		}
		b, err := fs.ReadFile(contentFS, testDir+"/"+name)
		if err != nil {
			return nil, fmt.Errorf("read test %s: %w", name, err)
		}
		// Strip the trailing .txt so the on-disk filename ends in _test.go.
		goName := strings.TrimSuffix(name, ".txt")
		out[goName] = string(b)
	}
	return out, nil
}

// TemplateTodos scans the template for lines containing `TODO` and
// returns them as Todo entries with their hint text (locale already
// resolved by the caller's preference). Used by GET /template.
func (c Chapter) TemplateTodos(preferZh bool) ([]TodoEntry, error) {
	code, err := c.TemplateCode()
	if err != nil {
		return nil, err
	}
	hf, _ := c.Hints() // best-effort; missing file = no hints attached

	byLine := map[int]Hint{}
	for _, h := range hf.Todos {
		byLine[h.Line] = h
	}

	var out []TodoEntry
	for i, raw := range strings.Split(code, "\n") {
		lineNo := i + 1
		// Require the line's comment marker to BE a TODO. Prose mentions
		// like "fill in the // TODO" don't match because they have text
		// between line start and `// TODO`.
		if !todoPattern.MatchString(raw) {
			continue
		}
		h, ok := byLine[lineNo]
		text := ""
		if ok {
			if preferZh {
				text = h.Hint.Zh
			} else {
				text = h.Hint.En
			}
		}
		out = append(out, TodoEntry{Line: lineNo, Hint: text})
	}
	return out, nil
}

// TodoEntry is a flattened (line, resolved-hint-text) pair.
type TodoEntry struct {
	Line int
	Hint string
}
