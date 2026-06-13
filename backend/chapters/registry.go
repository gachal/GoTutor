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
