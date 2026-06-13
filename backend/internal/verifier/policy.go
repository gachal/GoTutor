// Package verifier compiles and runs the learner's Go code in a sandboxed
// tempdir, then reports a deterministic pass/fail with captured output.
//
// v1 sandbox (per the plan):
//   - tempdir per request under os.TempDir()/gotutor-<chapter>-<uuid>/
//   - layered timeouts (handler 15s, go test 10s, cmd.WaitDelay 2s)
//   - AST-based stdlib whitelist (rejects os/exec, unsafe, syscall)
//   - 64KB output cap to prevent log-bomb DoS
//   - SIGKILL on timeout via cmd.Cancel + WaitDelay
//
// Phase 11 will harden with RLIMIT_FSIZE/NPROC and optional Docker.
package verifier

import (
	"time"

	"gotutor/backend/chapters"
)

// Policy is the per-chapter ruleset the verifier enforces. Default is
// safe stdlib; chapters can opt in to extras (net/http for urlcheck).
type Policy struct {
	// AllowedImports is the import path whitelist. Empty defaults to
	// SafeStdlibImports. Chapters add their extras (e.g. "net/http").
	AllowedImports []string

	// AllowListen permits net.Listen / http.ListenAndServe call sites.
	// Default false; even urlcheck keeps it false (client-only).
	AllowListen bool

	// OuterTimeout caps the whole verifier call (default 15s).
	OuterTimeout time.Duration
	// GoTestTimeout is passed to `go test -timeout` (default 10s).
	GoTestTimeout time.Duration
	// WaitDelay is the grace period before SIGKILL after context cancel
	// (default 2s). Reaps orphan child processes.
	WaitDelay time.Duration

	// OutputCap is the max combined stdout+stderr bytes captured
	// (default 64 KiB). Beyond this the output is truncated with a marker.
	OutputCap int
}

// DefaultPolicy returns the safe baseline. Chapters extend it via FromChapter.
func DefaultPolicy() Policy {
	return Policy{
		AllowedImports: nil,
		AllowListen:    false,
		OuterTimeout:   15 * time.Second,
		GoTestTimeout:  10 * time.Second,
		WaitDelay:      2 * time.Second,
		OutputCap:      64 * 1024,
	}
}

// SafeStdlibImports is the default whitelist. Stdlib minus the packages
// that let user code spawn processes, escape the type system, or do raw
// syscalls. Chapters may add to this via Chapter.AllowImports.
//
// Note: this list is intentionally conservative. urlcheck opts into
// net/http + sync + time; nothing opts into os/exec or unsafe.
var SafeStdlibImports = map[string]bool{
	"fmt":          true,
	"os":           true,
	"strconv":      true,
	"strings":      true,
	"math":         true,
	"time":         true,
	"errors":       true,
	"log":          true,
	"sort":         true,
	"path":         true,
	"path/filepath": true,
	"io":           true,
	"io/fs":        true,
	"bytes":        true,
	"unicode":      true,
	"regexp":       true,
	"context":      true,
	"sync":         true,
}

// DeniedStdlibImports is the explicit blocklist — redundant with the
// whitelist but defense-in-depth: even if a future maintainer adds a
// chapter that opens up "os" broadly, these stay blocked.
var DeniedStdlibImports = map[string]bool{
	"os/exec": true,
	"unsafe":  true,
	"syscall": true,
	"reflect": true,
	"net":     true,
	"C":       true,
	"cgo":     true,
}

// FromChapter derives a Policy from a chapters.Chapter. The chapter's
// AllowImports extends the safe baseline; AllowListen is honored.
func FromChapter(ch chapters.Chapter) Policy {
	p := DefaultPolicy()
	for _, imp := range ch.AllowImports {
		p.AllowedImports = append(p.AllowedImports, imp)
	}
	p.AllowListen = ch.AllowListen
	return p
}

// IsImportAllowed reports whether the given import path is permitted
// under this policy. DeniedStdlibImports always wins.
func (p Policy) IsImportAllowed(imp string) bool {
	if DeniedStdlibImports[imp] {
		return false
	}
	if SafeStdlibImports[imp] {
		return true
	}
	for _, a := range p.AllowedImports {
		if a == imp {
			return true
		}
	}
	return false
}
