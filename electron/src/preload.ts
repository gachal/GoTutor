import { contextBridge, ipcRenderer } from 'electron'

// Phase 9 v1 — frontend talks HTTP to the backend; IPC only for diagnostics.
// Phase 2 added getGoInfo / getLogPath so the Go-missing screen and the
// backend-down screen can show platform-aware install hints and the log
// file location. Both are lazy + memoized: the first caller triggers the
// IPC round-trip, subsequent callers get the cached promise.
contextBridge.exposeInMainWorld('gotutor', {
  version: process.env.npm_package_version || '0.0.0',
  platform: process.platform,
  isElectron: true,
  getGoInfo: (() => {
    let cached: Promise<{ found: boolean; version: string; path: string }> | null = null
    return () => {
      if (!cached) cached = ipcRenderer.invoke('get-go-info')
      return cached
    }
  })(),
  getLogPath: (() => {
    let cached: Promise<string> | null = null
    return () => {
      if (!cached) cached = ipcRenderer.invoke('get-log-path')
      return cached
    }
  })(),
})
