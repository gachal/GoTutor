package verifier

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// assertWithinSandbox guards against symlink-based escapes. After every
// file write inside the sandbox, we EvalSymlinks the resulting absolute
// path and confirm it still resolves inside the sandbox root.
//
// Without this, malicious user code (or a buggy test helper) could
// trick a future WriteFile call into following a symlink out of the
// tempdir. Phase 3 doesn't write user-controlled paths directly, but
// defense-in-depth: any new write site should call this.
func assertWithinSandbox(sandboxDir, relpath string) error {
	full := filepath.Join(sandboxDir, relpath)
	resolved, err := filepath.EvalSymlinks(full)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("pathguard: eval symlinks %s: %w", full, err)
	}
	if resolved == "" {
		resolved = full
	}

	cleanRoot, err := filepath.EvalSymlinks(sandboxDir)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("pathguard: eval root %s: %w", sandboxDir, err)
	}
	if cleanRoot == "" {
		cleanRoot = sandboxDir
	}

	absResolved, _ := filepath.Abs(resolved)
	absRoot, _ := filepath.Abs(cleanRoot)

	if !strings.HasPrefix(absResolved+string(filepath.Separator), absRoot+string(filepath.Separator)) &&
		absResolved != absRoot {
		return fmt.Errorf("pathguard: %q resolves outside sandbox %q", relpath, sandboxDir)
	}
	return nil
}
