import { existsSync, readFileSync, unlinkSync } from 'fs'
import { join } from 'path'
import { homedir } from 'os'

// Default portfile path — must match backend/main.go's default
// (`~/.gotutor/backend.port`). Override via the function arg if the
// backend was spawned with a custom -portfile.
export function defaultPortFile(): string {
  return join(homedir(), '.gotutor', 'backend.port')
}

// waitForPortFile polls path every 50ms for up to 10s. Resolves with the
// integer port read from the file. Rejects on timeout — the caller
// should treat that as "backend failed to boot" and show an error.
export function waitForPortFile(path: string, timeoutMs = 10_000): Promise<number> {
  const start = Date.now()
  return new Promise<number>((resolve, reject) => {
    const tick = () => {
      if (existsSync(path)) {
        try {
          const text = readFileSync(path, 'utf8').trim()
          const port = parseInt(text, 10)
          if (Number.isInteger(port) && port > 0 && port < 65536) {
            try { unlinkSync(path) } catch { /* ignore */ }
            resolve(port)
            return
          }
        } catch {
          // read/race — fall through to retry
        }
      }
      if (Date.now() - start > timeoutMs) {
        reject(new Error(`backend did not write port file within ${timeoutMs}ms`))
        return
      }
      setTimeout(tick, 50)
    }
    tick()
  })
}
