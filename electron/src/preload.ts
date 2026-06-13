import { contextBridge } from 'electron'

// Phase 9 v1 — frontend talks HTTP to the backend; no IPC needed.
// We expose a tiny diagnostics surface for the renderer to read the
// app version and platform (used by the error page if the backend dies).
contextBridge.exposeInMainWorld('gotutor', {
  version: process.env.npm_package_version || '0.0.0',
  platform: process.platform,
  isElectron: true,
})
