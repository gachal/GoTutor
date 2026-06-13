import { ChildProcess, spawn } from 'child_process'
import { createWriteStream, WriteStream, existsSync } from 'fs'
import { join } from 'path'
import { app } from 'electron'
import { defaultPortFile } from './portDiscovery'

// BackendHandle owns the spawned Go binary's lifecycle. main.ts creates
// one on boot, awaits its port via portDiscovery, and calls stop() on
// before-quit. Crash detection flows through the child's `exit` event.
export interface BackendHandle {
  process: ChildProcess
  port: number
  portFile: string
  logFile: string
  stop: () => Promise<void>
}

// resolveBackendBinary finds the Go binary for the current platform:
//   dev (GOTUTOR_DEV=1) — pre-built backend/bin/<os>-<arch>/gotutor-backend
//                         if present; otherwise `go run ./backend`.
//   prod                — process.resourcesPath/backend/<name>
//
// We avoid `go run` in dev when possible because it recompiles on every
// launch (~1-2s). `make backend-build` produces the fast-path binary.
export function resolveBackendBinary(): string {
  if (process.env.GOTUTOR_DEV === '1') {
    const local = devBinaryPath()
    if (local) return local
    return 'go'
  }
  const name = process.platform === 'win32' ? 'gotutor-backend.exe' : 'gotutor-backend'
  return join(process.resourcesPath, 'backend', name)
}

// devBinaryPath returns the path to backend/bin/<os>-<arch>/<binary> if
// it exists, otherwise empty. Computed relative to the electron/dist/
// output dir so it works from `pnpm dev` launched in electron/.
function devBinaryPath(): string {
  const repoRoot = join(__dirname, '..', '..')
  const plat = `${process.platform}-${process.arch}`
  const name = process.platform === 'win32' ? 'gotutor-backend.exe' : 'gotutor-backend'
  const candidate = join(repoRoot, 'backend', 'bin', plat, name)
  return existsSync(candidate) ? candidate : ''
}

// spawnBackend launches the Go binary with sensible defaults and returns
// a handle. The caller MUST call handle.stop() on shutdown.
export function spawnBackend(): BackendHandle {
  const portFile = defaultPortFile()
  const logFile = join(app.getPath('logs'), 'backend.log')

  const logStream: WriteStream = createWriteStream(logFile, { flags: 'a' })
  logStream.write(`\n--- ${new Date().toISOString()} ---\n`)

  let proc: ChildProcess
  const bin = resolveBackendBinary()
  if (bin === 'go') {
    const repoRoot = join(__dirname, '..', '..')
    proc = spawn('go', ['run', './backend', '-portfile', portFile], {
      cwd: repoRoot,
      stdio: ['ignore', 'pipe', 'pipe'],
    })
  } else {
    proc = spawn(bin, ['-portfile', portFile], {
      stdio: ['ignore', 'pipe', 'pipe'],
    })
  }

  proc.stdout?.pipe(logStream)
  proc.stderr?.pipe(logStream)

  // stop(): SIGTERM → 3s grace → SIGKILL. Resolves once the process is gone.
  const stop = async () => {
    if (proc.killed || proc.exitCode !== null) return
    return new Promise<void>((resolve) => {
      proc.once('exit', () => resolve())
      try { proc.kill('SIGTERM') } catch { /* already gone */ resolve(); return }
      const force = setTimeout(() => {
        try { proc.kill('SIGKILL') } catch { /* ignore */ }
        resolve()
      }, 3_000)
      proc.once('exit', () => clearTimeout(force))
    })
  }

  return { process: proc, port: 0, portFile, logFile, stop }
}
