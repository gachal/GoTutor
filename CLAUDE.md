# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working
with code in this repository.

## Status: v1 shipped

GoTutor v1 is a working cross-platform desktop app for learning Go via
`// TODO` mini-projects. Backend, frontend, and Electron shell are all
wired end-to-end; both shipped chapters (calc, urlcheck) play through
the full Monaco + sandboxed-go-test flow.

Read the docs first when picking up context:
- [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md) — system layers + submit flow
- [docs/SECURITY.md](docs/SECURITY.md) — sandbox threat model + future hardening
- [docs/ADDING_A_CHAPTER.md](docs/ADDING_A_CHAPTER.md) — how to add Chapter 3+

## Repo layout

```
backend/      Go module — Gin + SQLite + sandboxed verifier
  internal/
    api/      HTTP handlers (chapters, submit, reset)
    config/   flag/env-driven Config
    db/       SQLite open + embedded migrations
    server/   Server lifecycle, routes, logger
    verifier/ policy.go, astcheck.go, sandbox.go, exec.go, pathguard.go
  chapters/
    registry.go       Chapter struct + List/Get
    chapters.go       //go:embed all:content
    content/<id>/     template.txt, hints.yaml, tests/*.go.txt, solution/main.txt

frontend/     Vue 3 SPA — Vite + Pinia + vue-i18n + Monaco
  src/
    api/      Axios wrapper (Accept-Language interceptor)
    components/CodeEditor.vue — Monaco + TODO decorations
    i18n/     locales/{zh-CN,en}.yaml loaded at build time
    stores/   theme, locale, chapters (Pinia)
    views/    ChapterListView, ChapterView

electron/     Electron main process
  src/
    main.ts          app lifecycle + window
    backend.ts       spawn Go binary sidecar
    portDiscovery.ts poll ~/.gotutor/backend.port
    goDetector.ts    detect `go` on PATH at boot
    preload.ts       exposes window.gotutor diagnostics
  electron-builder.yml

docs/         ARCHITECTURE.md, SECURITY.md, ADDING_A_CHAPTER.md
```

## Build / run / test commands

```bash
make                                  # list all targets
make backend-dev                      # Go sidecar on :8081
make backend-build                    # cross-compile all 4 targets
make backend-test                     # go test ./... in backend/
make frontend-install && make frontend-dev    # Vite on :5173
make frontend-build                   # production SPA build → frontend/dist/
make electron-install && make electron-build # TypeScript → electron/dist/
make package-darwin                   # release/GoTutor-<ver>-arm64.dmg
make package-linux                    # release/GoTutor-<ver>.AppImage + .deb
make package-win                      # release/GoTutor-Setup-<ver>-x64.exe
```

## Conventions

- **Embed over disk**: chapter assets ship inside the binary via
  `//go:embed`. Templates live as `.txt` (go:embed skips `.go`); tests
  as `<name>_test.go.txt` (stripped to `<name>_test.go` at submit time).
- **Pure-Go SQLite** (`modernc.org/sqlite`) — never switch to a CGO
  driver; cross-compilation in Phase 10 depends on it.
- **Bilingual everything**: chapter `Locale{Zh, En}`, hints.yaml `zh`/`en`
  fields, frontend locales/zh-CN.yaml + en.yaml. Add new strings to both.
- **TODO detection**: `^\s*//\s*TODO\b` regex in `chapters/registry.go`.
  Any line whose comment literally starts with `// TODO` becomes a TODO
  marker. Keep hints.yaml line numbers in sync after edits.
- **Sandbox timeouts**: handler 15s, `go test -timeout` 10s,
  `cmd.WaitDelay` 2s. Don't widen without updating docs/SECURITY.md.
- **Vite `base: './'`**: load-bearing for Electron's `file://` loader.
  Don't change.
- **Hash history in router**: required for `file://` — HTML5 pushState
  can't navigate file:// paths.
- **No `.go` files in `backend/chapters/content/`**: go:embed will
  skip them. Always use `.txt` extensions for embedded Go source.

## Known gaps (post-v1)

- **Unsigned macOS builds** — Gatekeeper bypass documented in README.
- **Windows-specific RLIMIT_* enforcement** missing — see SECURITY.md.
- **No CI release workflow yet** — `.github/workflows/ci.yml` exists
  but release.yml is TODO.
- **Welcome overlay / first-run tutorial** planned but not shipped.
