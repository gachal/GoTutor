package verifier

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os/exec"
	"runtime"
	"syscall"
	"time"
)

// ExecResult holds the captured output and timing of a go test run.
type ExecResult struct {
	Passed     bool
	Output     string // combined stdout+stderr, capped at policy.OutputCap
	DurationMs int64
	// TimedOut is true iff the test exceeded GoTestTimeout and was killed.
	TimedOut bool
}

// RunGoTest invokes `go test -v -timeout <policy.GoTestTimeout> ./...`
// in the sandbox dir. The outer context caps the whole call; if it
// cancels, cmd.Cancel fires SIGKILL and cmd.WaitDelay reaps orphans.
//
// Output is captured via a LimitReader-style growable buffer so the
// captured text never exceeds OutputCap+marker bytes, even for very
// noisy programs.
func RunGoTest(ctx context.Context, dir string, goBin string, policy Policy) (ExecResult, error) {
	if goBin == "" {
		goBin = "go"
	}

	// Per-test timeout context. Inherits parent ctx cancellation too.
	testCtx, cancel := context.WithTimeout(ctx, policy.GoTestTimeout+policy.WaitDelay)
	defer cancel()

	args := []string{"test", "-v", fmt.Sprintf("-timeout=%s", policy.GoTestTimeout), "./..."}
	cmd := exec.CommandContext(testCtx, goBin, args...)
	cmd.Dir = dir

	// Combined output buffer with hard cap.
	var buf cappedBuffer
	buf.cap = policy.OutputCap
	cmd.Stdout = &buf
	cmd.Stderr = &buf

	// SIGKILL on context cancel. WaitDelay ensures orphan children
	// (e.g. a subprocess spawned by user code) get reaped.
	cmd.Cancel = func() error {
		if cmd.Process == nil {
			return nil
		}
		return syscall.Kill(cmd.Process.Pid, syscall.SIGKILL)
	}
	cmd.WaitDelay = policy.WaitDelay

	start := time.Now()
	err := cmd.Run()
	elapsed := time.Since(start)

	result := ExecResult{
		Passed:     err == nil,
		Output:     buf.String(),
		DurationMs: elapsed.Milliseconds(),
		TimedOut:   testCtx.Err() == context.DeadlineExceeded,
	}

	if result.TimedOut {
		marker := fmt.Sprintf("[verifier: killed after %s timeout]\n", policy.GoTestTimeout)
		result.Output = marker + result.Output
		result.Passed = false
	}

	return result, nil
}

// cappedBuffer is a bytes.Buffer that stops appending after reaching cap
// bytes, then appends a truncation marker exactly once. This bounds
// memory and prevents log-bomb DoS from user programs with infinite loops.
type cappedBuffer struct {
	bytes.Buffer
	cap       int
	truncated bool
}

func (b *cappedBuffer) Write(p []byte) (int, error) {
	if b.cap <= 0 {
		return b.Buffer.Write(p)
	}
	remaining := b.cap - b.Buffer.Len()
	if remaining <= 0 {
		if !b.truncated {
			b.truncated = true
			_, _ = b.Buffer.WriteString("\n[output truncated at " + fmt.Sprintf("%d", b.cap) + " bytes]\n")
		}
		return len(p), nil
	}
	if len(p) <= remaining {
		return b.Buffer.Write(p)
	}
	n, _ := b.Buffer.Write(p[:remaining])
	if !b.truncated {
		b.truncated = true
		_, _ = b.Buffer.WriteString("\n[output truncated at " + fmt.Sprintf("%d", b.cap) + " bytes]\n")
	}
	return n, nil
}

var _ io.Writer = (*cappedBuffer)(nil)

// goBinPath resolves the go binary path. Empty arg falls back to PATH
// lookup of "go". Used by verifier.go.
func goBinPath(cfg string) string {
	if cfg != "" {
		return cfg
	}
	if path, err := exec.LookPath("go"); err == nil {
		return path
	}
	return "go"
}

// numCPU returns runtime.NumCPU — used by the Phase 3 concurrency
// semaphore. Wrapped so tests can stub it if needed.
var numCPU = runtime.NumCPU
