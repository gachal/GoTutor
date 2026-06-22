# GoTutor

> English | [中文](README-zh.md)

A **muscle-memory gym for AI-era developers**. You already write code —
maybe even let AI write a lot of it. GoTutor makes your hands type out
the patterns real Go projects use, filling `// TODO` gaps in genuine
mini-projects. The app compiles and runs your code with real `go test`
— no string matching, no fake REPL. The drills build the muscle memory
that copy-paste from AI erodes.

## Status

Ships **fifteen chapters** end-to-end, organized into three tracks:

- **Fundamentals (5)** — `calc` (CLI calculator), `structs` (structs &
  methods), `slice` (Filter/Map/Unique/Chunk), `mapjson` (JSON
  round-trip), `http` (Handler basics). Builds the syntax muscle memory
  every later chapter assumes.
- **Concurrency (1)** — `urlcheck` (goroutines, channels, WaitGroup,
  `net/http` client).
- **Gateway Patterns (9)** — drawn from the Go patterns used in a real
  LLM API gateway (AiDeptus): error handling, interfaces & strategy,
  concurrent sum, channels & select, context cancellation, token-bucket
  rate limiting, circuit breaker, HTTP retry & backoff, SSE streaming.

All play through the full Monaco editor + sandboxed `go test` flow.

The home screen shows your overall progress, a "continue where you left
off" entry, and a per-chapter card with difficulty + estimated minutes +
practice count. A first-run welcome overlay explains the practice model;
an install-Go screen appears if the toolchain is missing.

## Stack

| Layer       | Choice                                              |
|-------------|-----------------------------------------------------|
| Frontend    | Vue 3 + Vite + Electron + Monaco + Pinia + vue-i18n |
| Backend     | Go + Gin HTTP sidecar                               |
| Storage     | SQLite (`modernc.org/sqlite`, pure-Go — no CGO)     |
| Verifier    | sandboxed `go test` in tempdir + AST stdlib whitelist |
| Packaging   | electron-builder; Go binary bundled as `extraResources` |

UI is **bilingual** (zh-CN + en) with a toggle in the sidebar; both
light and dark themes follow system or manual selection.

## Prerequisites

- **Go 1.26+** on PATH — the verifier invokes it. App shows an
  install-Go screen if missing.
- **Node.js 22+** and **pnpm 8+** for frontend/electron dev.

## Dev setup

```bash
git clone <repo> GoTutor && cd GoTutor

# 1. Backend (terminal 1)
make backend-dev                          # serves :8081

# 2. Frontend (terminal 2)
make frontend-install                     # one-time
make frontend-dev                         # serves :5173, proxies /api → :8081

# 3. Electron shell (terminal 3, optional — wraps 1 + 2)
make electron-install                     # one-time, ~5-25 min for Electron download
GOTUTOR_DEV=1 make electron-dev           # opens the desktop window
```

For a quick smoke test without Electron, open `http://localhost:5173`
in a browser after steps 1+2.

## Build installers

```bash
# Cross-compile Go backend + build frontend + bundle Electron
make package-darwin        # → release/GoTutor-<ver>-arm64.dmg
make package-linux         # → release/GoTutor-<ver>.AppImage + .deb
make package-win           # → release/GoTutor-Setup-<ver>-x64.exe
```

Pure-Go SQLite means no CGO hell; the same Go source cross-compiles
to all four targets without a C toolchain per OS.

> **macOS Gatekeeper**: v1 is unsigned. To run the dmg, right-click the
> app → Open → confirm. Future releases will sign with an Apple
> Developer ID.

## API

| Endpoint                              | Method | Body                          | Returns                              |
|---------------------------------------|--------|-------------------------------|--------------------------------------|
| `/api/health`                         | GET    | —                             | `{ok, port, goFound, goVersion}`     |
| `/api/chapters`                       | GET    | —                             | `Chapter[]` with track/difficulty/estimatedMinutes/prerequisites |
| `/api/progress`                       | GET    | —                             | `{totalChapters, completedChapters, percent, lastChapterId, byTrack}` |
| `/api/chapters/:id/template`          | GET    | —                             | `{code, todos: [{line, hint}]}`      |
| `/api/chapters/:id/hint?line=N`       | GET    | —                             | `{text}`                             |
| `/api/chapters/:id/submit`            | POST   | `{userCode: string}`          | `{passed, output, durationMs, ...}`  |
| `/api/reset`                          | POST   | —                             | 204                                  |

`Accept-Language` toggles zh-CN vs en on every locale-sensitive endpoint.

## Documentation

- [Packaging](docs/PACKAGING.md) — build `.dmg` / `.exe` / `.AppImage`
  installers, Gatekeeper bypass, troubleshooting.
- [Architecture](docs/ARCHITECTURE.md) — system layers, end-to-end
  submit flow, storage layout.
- [Security model](docs/SECURITY.md) — what the sandbox defends against,
  residual risks, future hardening (RLIMIT, Docker).
- [Adding a chapter](docs/ADDING_A_CHAPTER.md) — write your own
  exercises (15 chapters across 3 tracks already ship).

## License

MIT, Copyright © 2026 TeaHouse. See [LICENSE](LICENSE).
