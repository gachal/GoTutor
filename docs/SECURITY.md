# GoTutor Security Model

GoTutor runs user-submitted Go code via `go test`. That code is
untrusted by definition — a learner might paste anything. This document
explains what v1 defends against, what it doesn't, and the future
hardening path.

## v1 sandbox

Every `POST /api/chapters/:id/submit` does the following before
returning a verdict:

1. **Body cap.** Request body limited to 256 KiB. Prevents giant
   payloads from OOMing the handler before verification.
2. **AST import check.** `internal/verifier/astcheck.go` parses the
   user's source with `go/parser` and rejects any import not in the
   chapter's whitelist. Defaults (in `policy.go`):
   - **Allowed**: `fmt`, `os`, `strconv`, `strings`, `math`, `time`,
     `errors`, `log`, `sort`, `path`, `path/filepath`, `io`, `io/fs`,
     `bytes`, `unicode`, `regexp`, `context`, `sync`.
   - **Denied always** (`DeniedStdlibImports`): `os/exec`, `unsafe`,
     `syscall`, `reflect`, `net`, `C`, `cgo`. These win regardless of
     chapter policy.
   - Chapters extend the whitelist via `Chapter.AllowImports`
     (urlcheck opts into `net/http`, `sync`, `time`).
   - `net.Listen` and `http.ListenAndServe` call sites are also
     rejected unless `AllowListen` is true (it never is in v1).
3. **Tempdir isolation.** Each submission runs in
   `os.TempDir()/gotutor-<chapter>-<uuid>/` with its own `go.mod`
   (`module gotutoruser`). No shared state across submissions.
4. **Layered timeouts.**
   - Handler outer ctx: 15 s.
   - `go test -timeout`: 10 s.
   - `cmd.WaitDelay`: 2 s (reaps orphan children of user code).
   - On timeout, `cmd.Cancel` fires `SIGKILL` so user code can't
     escape via signal handlers.
5. **Output cap.** Combined stdout+stderr truncated at 64 KiB with a
   `[output truncated at 65536 bytes]` marker. Prevents log-bomb DoS.
6. **Path guard.** `pathguard.go` resolves symlinks before any write
   and asserts the resolved path stays under the sandbox root.
7. **Concurrency limit.** A `chan struct{}` sized to `NumCPU` bounds
   parallel compilations so a flood of submits can't peg every core.

## What v1 defends against

- **Process spawning**: `os/exec` denied at AST, so the learner can't
  shell out to `rm`, `curl`, etc.
- **Unsafe memory tricks**: `unsafe` and `syscall` denied.
- **Reflection-based bypasses**: `reflect` denied so AST can't be
  fooled by reflective import lookups.
- **Network listeners**: `net.Listen` / `http.ListenAndServe` blocked.
  Outbound HTTP allowed only for `urlcheck` (`net/http` client).
- **Fork bombs**: process count is bounded by `NumCPU` semaphore +
  10 s test timeout.
- **Infinite loops**: `go test -timeout 10s` + outer 15 s ctx.
- **Log-bomb DoS**: 64 KiB output cap.

## What v1 does NOT defend against

These are documented residual risks; a determined attacker can likely
achieve them on a v1 install:

- **Disk-fill via os.WriteFile in a loop.** 10 s timeout caps duration
  but not total bytes. Mitigation: future `RLIMIT_FSIZE`.
- **CPU burn.** 10 s timeout bounds but a tight loop on all cores
  still pegs the machine. Mitigation: future `RLIMIT_CPU` + cgroups.
- **`net/http` client abuse** (urlcheck chapter only). A learner could
  DoS a third-party URL from the learner's own machine. Acceptable in
  v1's "single trusted user" model; not acceptable for multi-tenant.
- **Windows-specific resource limits.** `RLIMIT_*` is Linux/BSD only.
  v1 Windows relies on timeouts + AST alone.

## Future hardening path

When GoTutor ships to a multi-tenant context, layer these in:

1. **`RLIMIT_FSIZE`** (10 MiB) on Linux/macOS — bounds disk writes.
2. **`RLIMIT_NPROC`** (1) on Linux — anti fork-bomb.
3. **`RLIMIT_CPU`** (10 s) on Linux/macOS — kernel-level CPU cap.
4. **Docker backend**. `verifier.backend = "docker"` config switch
   runs `go test` inside `--network=none --memory=256m --pids-limit=64
   --read-only --tmpfs /tmp`. The gold-standard sandbox. Requires
   Docker Desktop on macOS/Windows.
5. **Network namespacing** on Linux (`unshare -n`) for true network
   isolation without Docker.
6. **Windows Job Objects** for CPU/memory/pids limits.

## Reporting a security issue

Email security@teahouse.dev with details. Please don't open a public
issue for security-sensitive bugs.
