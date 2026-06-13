package verifier

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
)

// Sandbox is an isolated working directory where the learner's code is
// materialized as main.go alongside the chapter's test files, then
// `go test ./...` is invoked with cwd=sandbox.Dir().
//
// Lifecycle:
//
//	sb := NewSandbox(chapterID)
//	defer sb.Cleanup()
//	sb.WriteUser(userCode)              // writes <dir>/main.go
//	sb.WriteTests(testFiles)            // writes <dir>/<name>_test.go
//	sb.WriteGoMod()                     // writes <dir>/go.mod
//	// run `go test` with cwd=sb.Dir()
type Sandbox struct {
	dir string
}

// NewSandbox creates a fresh tempdir under os.TempDir()/gotutor-<chapter>-<id>/.
// Caller MUST defer sb.Cleanup().
func NewSandbox(chapterID string) (*Sandbox, error) {
	uuid, err := randomID()
	if err != nil {
		return nil, fmt.Errorf("gen id: %w", err)
	}
	dir, err := os.MkdirTemp("", "gotutor-"+chapterID+"-"+uuid+"-*")
	if err != nil {
		return nil, fmt.Errorf("mkdir temp: %w", err)
	}
	return &Sandbox{dir: dir}, nil
}

// Dir returns the absolute path to the sandbox root. Pass this as cwd
// to exec.Command when invoking go test.
func (s *Sandbox) Dir() string { return s.dir }

// WriteFile writes content under the sandbox root at the relative path.
// Used internally; tests and main.go go through WriteUser/WriteTests
// for clarity.
func (s *Sandbox) WriteFile(relpath, content string) error {
	full := filepath.Join(s.dir, relpath)
	if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
		return fmt.Errorf("mkdir %s: %w", filepath.Dir(relpath), err)
	}
	if err := os.WriteFile(full, []byte(content), 0o644); err != nil {
		return fmt.Errorf("write %s: %w", relpath, err)
	}
	return nil
}

// WriteUser writes the learner's code as main.go at the sandbox root.
func (s *Sandbox) WriteUser(userCode string) error {
	return s.WriteFile("main.go", userCode)
}

// WriteTests writes each (filename, content) pair to the sandbox root.
// Test files come from the chapter's embedded tests/ directory.
func (s *Sandbox) WriteTests(files map[string]string) error {
	for name, content := range files {
		if err := s.WriteFile(name, content); err != nil {
			return err
		}
	}
	return nil
}

// WriteGoMod writes a minimal go.mod with module name "gotutoruser"
// and the current Go major version. Tests are stdlib-only so no deps.
func (s *Sandbox) WriteGoMod(goVersion string) error {
	content := fmt.Sprintf("module gotutoruser\n\ngo %s\n", goVersion)
	return s.WriteFile("go.mod", content)
}

// Cleanup removes the entire tempdir. Safe to call multiple times.
func (s *Sandbox) Cleanup() {
	if s.dir != "" {
		_ = os.RemoveAll(s.dir)
		s.dir = ""
	}
}

// randomID returns 8 hex chars from crypto/rand. Used in the tempdir
// name to keep concurrent submissions from colliding.
func randomID() (string, error) {
	b := make([]byte, 4)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
