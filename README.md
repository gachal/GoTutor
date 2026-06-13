# GoTutor

An interactive desktop app for learning Go by filling in `// TODO` gaps in real mini-projects. The app actually compiles and runs your code with `go test` — no string matching.

> **Status**: pre-development. See the phased plan in `~/.claude/plans/claude-code-go-robust-lagoon.md` for the implementation roadmap.

## Stack

- **Frontend**: Vue 3 + Vite + Electron + Monaco Editor + Pinia + vue-i18n
- **Backend**: Go + Gin HTTP sidecar
- **Storage**: SQLite (`modernc.org/sqlite`, pure-Go — no CGO)
- **Verifier**: sandboxed `go test` in an isolated tempdir with timeouts + AST-based stdlib whitelist

## Prerequisites

- Go 1.26+
- Node.js 22+
- pnpm 8+

## Development

```bash
make            # list all targets
make backend-dev
make frontend-dev
make electron-dev
```

Full documentation lands in Phase 12.
