# Packaging GoTutor for macOS

End-to-end guide for building a runnable `.dmg` from source.

## One-command build (recommended)

```bash
make package-darwin
```

This target chains:
1. **Backend cross-compile** — `make backend-build-darwin-arm64` writes
   `backend/bin/mac-arm64/gotutor-backend` (~27 MB).
2. **Frontend build** — `pnpm --dir frontend build` writes
   `frontend/dist/` (initial bundle ~76 KB gzip + 30 KB Monaco chunk).
3. **Electron TS build** — `pnpm --dir electron build` writes
   `electron/dist/`.
4. **electron-builder** — bundles the above three plus the Electron
   runtime into a `.app`, then DMGs it.

Output: **`release/GoTutor-<version>-arm64.dmg`**

## Prerequisites

One-time installs (frontend + electron deps):

```bash
make frontend-install    # Vue/Vite/Monaco etc.
make electron-install    # Electron + electron-builder (~200 MB, 5-25 min)
```

Toolchain on the build machine:

```bash
go version       # 1.26+
node --version   # 22+
pnpm --version   # 8+
```

End users who install the `.dmg` only need **Go 1.26+** on PATH — the
backend verifier invokes `go test`. The app shows an install-Go screen
on first launch if `go` is missing.

## Intel Mac

For x64 builds (older Intel Macs):

```bash
make package-darwin-x64
```

Output: `release/GoTutor-<version>-x64.dmg`

## Universal binary?

Not in v1 — we ship per-arch dmgs. Universal via `lipo` is a future
option (`electron-builder` supports it with `--arch universal`).

## Verify the build

```bash
ls -lh release/
# Expected:
# GoTutor-0.1.0-arm64.dmg
# GoTutor-0.1.0-arm64-mac.zip
```

Double-click the `.dmg`, drag GoTutor to Applications.

## Gatekeeper (v1 is unsigned)

First launch shows *"GoTutor cannot be opened because it is from an
unidentified developer."* Workarounds:

**Option A — Finder right-click** (recommended):
1. Locate GoTutor.app in Finder.
2. **Right-click** → **Open**.
3. Confirm in the dialog.

**Option B — command line**:
```bash
xattr -cr /Applications/GoTutor.app
open /Applications/GoTutor.app
```

Future v2 will sign with an Apple Developer ID. See CLAUDE.md →
"Known gaps".

## Dev mode (no packaging)

For local testing without producing a `.dmg`:

```bash
# Terminal 1 — backend
make backend-dev                  # serves :8081

# Terminal 2 — frontend
make frontend-dev                 # Vite on :5173, proxies /api → :8081

# Terminal 3 — Electron shell
GOTUTOR_DEV=1 make electron-dev   # opens the desktop window
```

Hot-reload on every save; no rebuild needed.

## Cross-compiling to other OSes

Same Makefile, different targets:

```bash
make package-linux                # release/GoTutor-<ver>.AppImage + .deb
make package-win                  # release/GoTutor-Setup-<ver>-x64.exe
```

Cross-compile of the Go backend is trivial thanks to pure-Go SQLite
(`modernc.org/sqlite`) — no CGO, no per-OS C toolchain. electron-builder
downloads the Electron binary for the target OS automatically.

## Troubleshooting

- **`tsc: command not found`** — `make electron-install` didn't finish.
  Re-run; if it hangs, retry with `pnpm install --dir electron`.
- **electron-builder fails with `extraResources` error** — the backend
  binary for your target arch doesn't exist. Run
  `make backend-build-darwin-arm64` (or matching target) first.
- **App opens but shows "Cannot reach backend"** — backend crashed on
  boot. Check `~/Library/Logs/gotutor/backend.log` (macOS path; the
  Electron main process writes there via `app.getPath('logs')`).
- **App shows install-Go screen** — `go` isn't on PATH for GUI apps.
  GUI launchers don't inherit your shell PATH. Install Go via the
  official `.pkg` from go.dev (puts it in `/usr/local/go/bin` which the
  app can see), or symlink: `sudo ln -s $(which go) /usr/local/bin/go`.

## File layout of a packaged app

```
GoTutor.app/
├── Contents/
│   ├── Info.plist
│   ├── MacOS/
│   │   └── GoTutor              # Electron launcher binary
│   ├── Resources/
│   │   ├── backend/
│   │   │   └── gotutor-backend  # Go sidecar (cross-compiled)
│   │   ├── frontend/dist/       # Vue SPA
│   │   ├── electron/dist/       # compiled TS main + preload
│   │   └── app.asar             # Electron shell code
│   └── Frameworks/              # Chromium, Node, etc.
└── ...
```

At runtime, `electron/src/backend.ts` resolves the sidecar via
`process.resourcesPath/backend/gotutor-backend`.
