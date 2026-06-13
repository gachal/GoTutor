import { spawnSync } from 'child_process'

// detectGo returns true iff `go` resolves on PATH and `go version`
// succeeds. We call this on app boot; the renderer shows an install-Go
// screen when false (the backend's verifier can't run without Go).
export interface GoInfo {
  found: boolean
  version: string
  path: string
}

export function detectGo(): GoInfo {
  try {
    const result = spawnSync('go', ['version'], {
      encoding: 'utf8',
      windowsHide: true,
    })
    if (result.status === 0) {
      return {
        found: true,
        version: (result.stdout || '').trim(),
        path: whichGo(),
      }
    }
  } catch {
    // fallthrough to not-found
  }
  return { found: false, version: '', path: '' }
}

// whichGo shells `which go` (unix) or `where go` (windows) to record the
// resolved path for diagnostics. Best-effort — empty on failure.
function whichGo(): string {
  const cmd = process.platform === 'win32' ? 'where' : 'which'
  try {
    const result = spawnSync(cmd, ['go'], { encoding: 'utf8', windowsHide: true })
    if (result.status === 0) {
      return (result.stdout || '').split(/\r?\n/)[0].trim()
    }
  } catch {
    // ignore
  }
  return ''
}
