# GoTutor Architecture

> English | [中文](ARCHITECTURE-zh.md)

GoTutor is a **muscle-memory gym for AI-era developers** — it teaches
Go by having learners fill `// TODO` gaps in real mini-projects, verified
by **actually compiling and running the learner's Go code** with
`go test` (never string matching). Fifteen chapters ship in three tracks
(fundamentals / concurrency / gateway patterns). This document traces a
submission end-to-end.

```
┌────────────────────────────────────────────────────────────────┐
│ Electron main process (electron/src/main.ts)                   │
│  ┌─────────────────────────┐    ┌───────────────────────────┐  │
│  │ BrowserWindow           │    │ spawn Go binary sidecar   │  │
│  │  ┌───────────────────┐  │    │  (gotutor-backend)       │  │
│  │  │ Vue 3 SPA         │  │    │                           │  │
│  │  │  ┌─────────────┐  │  │    │ Gin :8081-8090           │  │
│  │  │  │ Monaco      │──┼──┼────┼──► HTTP /api/*           │  │
│  │  │  │ Editor      │  │  │    │                           │  │
│  │  │  └─────────────┘  │  │    │ SQLite (progress.db)     │  │
│  │  └───────────────────┘  │    │ Verifier (tempdir+go test)│  │
│  └─────────────────────────┘    └───────────────────────────┘  │
└────────────────────────────────────────────────────────────────┘
                       ↑ go build/test on host toolchain
```

## Layers

### Electron (`electron/src/`)
- `main.ts`: app lifecycle. Boot detects Go on PATH (`goDetector.ts`),
  spawns the backend (`backend.ts`), awaits port-file discovery
  (`portDiscovery.ts`), then opens the BrowserWindow.
- `backend.ts`: resolves the Go binary (dev: pre-built
  `backend/bin/<os>-<arch>/gotutor-backend`; prod:
  `process.resourcesPath/backend/`). Pipes stdout/stderr to
  `app.getPath('logs')/backend.log`. SIGTERM → 3s grace → SIGKILL.
- Window loads `http://localhost:5173` in dev (Vite), or
  `file://...frontend/dist/index.html` in prod. Vite's `base: './'`
  is load-bearing for the file:// case.

### Frontend (`frontend/src/`)
- Vue 3 SPA, Pinia stores, vue-i18n (zh-CN + en), vue-router with hash
  history (Electron-friendly).
- `App.vue`: sidebar with theme/locale toggles + chapter list + `?`
  help button (reopens the welcome overlay). Wraps `<RouterView>` in
  a `<BootGate>` driven by the health store.
- **Boot flow** (`stores/health.ts`): polls `/api/health` every 2s
  until backend is up. While loading/down, `<BootGate>` shows a branded
  spinner / Go-missing screen / backend-down retry screen instead of
  the app. `WelcomeView` overlay shows once per locale-storage slot
  (`stores/onboarding.ts`); dismissable via Skip or CTA.
- `views/ChapterListView.vue`: orchestrator that renders
  `OverallProgress` (top progress bar), `ContinueCard` ("continue
  where you left off" hero from `/api/progress.lastChapterId`), then
  one `TrackSection` per track (fixed order: fundamentals → concurrency
  → gateway). Each section is a grid of `ChapterCard` components
  showing ordinal, difficulty, estimated minutes, completion status,
  and practice count (`stores/runs.ts`, localStorage).
- `views/ChapterView.vue`: lazy-loads Monaco via
  `defineAsyncComponent` so the list view doesn't pay Monaco's ~3MB
  cost. Phase 4 additions: `HintsPanel` (per-TODO hint cards with
  jump-to-line buttons), friendly submit/run copy, removed duration
  noise, first-time-pass celebration, failure coaching heuristic.
- `components/CodeEditor.vue`: Monaco integration. Custom
  `gotutor-light` / `gotutor-dark` themes synced to the theme store.
  TODO decorations: glyph margin dot + line tint + hover hint.
  ⌘/Ctrl+Enter submits. Exposes `jumpTo(line)` via `defineExpose` so
  `HintsPanel` can drive the cursor.
- `api/client.ts`: Axios with `Accept-Language` interceptor so the
  backend resolves locale on every request.

### Backend (`backend/`)
- **HTTP layer** (`internal/server/`): Gin. Routes wired in
  `routes.go` with closures over `s.DB()`. Layered timeouts via
  `context.WithTimeout`. Graceful shutdown on SIGINT/SIGTERM.
- **Persistence** (`internal/db/`): SQLite via `modernc.org/sqlite`
  (pure-Go, no CGO — critical for cross-compilation in Phase 10).
  Migrations embedded via `//go:embed *.sql`, applied idempotently on
  every connect.
- **Chapter registry** (`chapters/`): static list of chapters with
  metadata; dynamic content (template, hints, tests, solution) loaded
  from `//go:embed all:content`. Test files ship as
  `<name>_test.go.txt` (go:embed skips `.go` files); stripped to
  `<name>_test.go` at submit time.
- **Verifier** (`internal/verifier/`): the heart of the app.
  - `policy.go`: per-chapter Policy (allowed imports, timeouts,
    output cap).
  - `astcheck.go`: `go/parser` based import scan + listen-call
    detection. Runs before compile, fails fast.
  - `sandbox.go`: tempdir lifecycle. Writes user code + tests + go.mod.
  - `exec.go`: `exec.CommandContext` with `cmd.Cancel=SIGKILL` +
    `WaitDelay`. Output captured into a 64 KiB-capped buffer.
  - `verifier.go`: top-level orchestration + NumCPU-sized concurrency
    semaphore.
  - `pathguard.go`: symlink-escape defense.

## Data flow: a submit

1. Learner types code in Monaco, clicks Submit.
2. Frontend calls `POST /api/chapters/calc/submit` with `{userCode}`.
3. `api.HandleSubmit` (in `internal/api/submit.go`):
   - Looks up chapter, binds + 256 KiB body-caps the request.
   - Calls `verifier.Verify(ctx, chapter, userCode, goBin)` with a
     15 s outer ctx.
4. `verifier.Verify`:
   - `ASTCheck`: parse + scan imports. Reject on banned import → return
     early without spawning `go test`.
   - Acquire concurrency slot.
   - `ch.TestFiles()` reads embedded `<chapter>/tests/*_test.go.txt`.
   - `NewSandbox`: tempdir under `os.TempDir()/gotutor-<chapter>-<uuid>/`.
   - Write `main.go` (user code), `<name>_test.go` (chapter tests),
     `go.mod` (`module gotutoruser`, `go 1.26`).
   - `RunGoTest`: `go test -v -timeout 10s ./...` with cwd=sandbox.
     Output streamed to cappedBuffer.
   - defer sandbox.Cleanup().
5. `HandleSubmit` upserts `progress` (completed_at on pass, attempts++
   on fail), returns `SubmitResult`.
6. Frontend flips the chapter's `completed` flag optimistically; sidebar
   updates reactively.

## Storage

`progress` table (SQLite, WAL mode):
- `chapter_id TEXT PRIMARY KEY`
- `completed_at INTEGER NULL` (unix seconds)
- `attempts INTEGER NOT NULL DEFAULT 0`
- `last_output TEXT`

Plus `chapters` (metadata, currently populated from registry not DB) and
`settings` (key/value, unused in v1).

## Cross-compilation

`make backend-build-<os>-<arch>` produces backend binaries for
mac-arm64, mac-x64, linux-x64, win-x64. Pure-Go SQLite makes this
trivial — no CGO, no cross-compiler toolchain. `electron-builder.yml`
bundles each binary into `Resources/backend/` via `extraResources`
with `${os}-${arch}` substitution.

## What's NOT in v2

- 15 chapters across 3 tracks (fundamentals / concurrency / gateway).
  Adding more: see [ADDING_A_CHAPTER.md](./ADDING_A_CHAPTER.md).
- **Practice count is localStorage-only** (`stores/runs.ts`) — does not
  sync across machines. The existing `progress` table records
  completion in SQLite, so cross-reinstall completion state survives
  but "练过 N 次" doesn't.
- **"Last visited chapter" is derived from `progress.completed_at`** —
  a true visits table (`chapter_visits`) is deferred to a future
  release. The ContinueCard uses the most recently *passed* chapter,
  which is the right UX 95% of the time but lags if the user is
  mid-attempt on a new chapter.
- **Hint coaching is single-level** — each TODO has one hint text. A
  `why` field (concept-level coaching layered on top of the tactical
  hint) is planned but not yet authored across all 15 chapters.
- Unsigned macOS builds (Gatekeeper bypass documented in README).
- Windows-specific resource limits (RLIMIT is Linux/BSD only). See
  [SECURITY.md](./SECURITY.md).
- Auto-update (Phase 13+, not in current plan).
- Multi-user progress (single-user model in v2).
