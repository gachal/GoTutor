package verifier

import (
	"context"
	"runtime"
	"strings"
	"sync"

	"gotutor/backend/chapters"
)

// Result is what the verifier hands back. The api/submit handler copies
// Passed/Output/DurationMs into its JSON response.
type Result struct {
	Passed     bool
	Output     string
	DurationMs int64
	// RejectedAtAST is true iff ASTCheck refused the code before compile.
	// Lets the UI distinguish "your code has a banned import" from
	// "your code didn't compile".
	RejectedAtAST bool
}

// concurrencySem bounds parallel `go test` invocations to NumCPU so a
// flood of submissions can't peg every core. Sized once at init.
//
//nolint:gochecknoglobals // deliberate package-level semaphore
var concurrencySem = make(chan struct{}, max1(numCPU()))

// Verify is the top-level verifier entry. Steps:
//  1. Derive Policy from chapter.
//  2. ASTCheck userCode (fail fast on banned imports).
//  3. Acquire concurrency slot.
//  4. Create sandbox, write main.go + tests + go.mod.
//  5. Run `go test -v -timeout ./...` with cmd.Cancel = SIGKILL.
//  6. Cleanup and return.
//
// Outer ctx should carry a deadline (handler enforces 15s); this fn
// adds an inner per-test timeout via Policy.GoTestTimeout.
func Verify(ctx context.Context, ch chapters.Chapter, userCode string, goBin string) Result {
	policy := FromChapter(ch)

	if err := ASTCheck(userCode, policy); err != nil {
		return Result{
			Passed:        false,
			Output:        err.Error(),
			RejectedAtAST: true,
		}
	}

	// Acquire concurrency slot. On ctx cancel, return promptly.
	select {
	case concurrencySem <- struct{}{}:
		defer func() { <-concurrencySem }()
	case <-ctx.Done():
		return Result{Passed: false, Output: "verifier: cancelled before run: " + ctx.Err().Error()}
	}

	tests, err := ch.TestFiles()
	if err != nil {
		return Result{Passed: false, Output: "verifier: load chapter tests: " + err.Error()}
	}

	sb, err := NewSandbox(ch.ID)
	if err != nil {
		return Result{Passed: false, Output: "verifier: sandbox: " + err.Error()}
	}
	defer sb.Cleanup()

	if err := sb.WriteUser(userCode); err != nil {
		return Result{Passed: false, Output: err.Error()}
	}
	if err := sb.WriteTests(tests); err != nil {
		return Result{Passed: false, Output: err.Error()}
	}
	if err := sb.WriteGoMod(currentGoMajorMinor()); err != nil {
		return Result{Passed: false, Output: err.Error()}
	}

	bin := goBinPath(goBin)
	exec, _ := RunGoTest(ctx, sb.Dir(), bin, policy)
	return Result{
		Passed:     exec.Passed,
		Output:     sanitizeOutput(exec.Output),
		DurationMs: exec.DurationMs,
	}
}

// sanitizeOutput drops the "FAIL\tgotutoruser [build failed]" trailer
// `go test -v` emits — the build error above it is what the user needs.
func sanitizeOutput(s string) string {
	lines := strings.Split(s, "\n")
	out := make([]string, 0, len(lines))
	for _, ln := range lines {
		if strings.HasPrefix(ln, "FAIL\tgotutoruser") {
			continue
		}
		out = append(out, ln)
	}
	return strings.TrimSpace(strings.Join(out, "\n"))
}

// currentGoMajorMinor returns "1.26" for go1.26.1. We pin the sandbox's
// go.mod to the running toolchain's major.minor so user code matches
// the stdlib the verifier builds against.
func currentGoMajorMinor() string {
	v := strings.TrimPrefix(runtime.Version(), "go")
	if i := strings.Index(v, "."); i >= 0 {
		if j := strings.Index(v[i+1:], "."); j >= 0 {
			return v[:i+1+j]
		}
	}
	return v
}

// max1 returns n or 1 if n <= 0. Avoids a zero-sized channel blocking forever.
func max1(n int) int {
	if n < 1 {
		return 1
	}
	return n
}

// once reserved for future first-use init (e.g. seeding, schema warm-up).
var _ = sync.Once{}
