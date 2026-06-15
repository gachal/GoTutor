import { app, BrowserWindow } from 'electron'
import { join } from 'path'
import { existsSync } from 'fs'
import { detectGo } from './goDetector'
import { spawnBackend, BackendHandle } from './backend'

// Fixed port — the frontend hardcodes http://localhost:8081/api in
// src/api/client.ts. No port-file discovery needed.
const BACKEND_PORT = 8081

let backend: BackendHandle | null = null
let mainWindow: BrowserWindow | null = null

const DEV_SERVER_URL = process.env.GOTUTOR_DEV_URL || 'http://localhost:5173'

async function bootstrap() {
  const go = detectGo()
  if (!go.found) {
    console.warn('[gotutor] Go toolchain not found on PATH')
  } else {
    console.log(`[gotutor] Go: ${go.version} @ ${go.path}`)
  }

  backend = spawnBackend()
  backend.port = BACKEND_PORT
  backend.process.on('exit', (code, signal) => {
    console.error(`[gotutor] backend exited code=${code} signal=${signal}`)
    if (mainWindow && !mainWindow.isDestroyed()) {
      mainWindow.webContents.send('backend:exited', { code, signal })
    }
  })

  // Give the backend a moment to bind before opening the window so the
  // first /api/chapters fetch doesn't get ECONNREFUSED. 500ms is plenty
  // for a Gin server cold-start; if it's not up by then the frontend
  // shows its "backend down" state and the user can retry.
  await new Promise((r) => setTimeout(r, 500))

  createWindow()
}

function createWindow() {
  mainWindow = new BrowserWindow({
    width: 1280,
    height: 800,
    minWidth: 960,
    minHeight: 600,
    backgroundColor: '#1a1a1f',
    title: 'GoTutor',
    webPreferences: {
      preload: join(__dirname, 'preload.js'),
      contextIsolation: true,
      nodeIntegration: false,
      sandbox: true,
    },
  })

  if (process.env.GOTUTOR_DEV === '1' || process.env.GOTUTOR_DEV_URL) {
    mainWindow.loadURL(DEV_SERVER_URL)
    mainWindow.webContents.openDevTools({ mode: 'detach' })
  } else {
    // Packaged mode: frontend ships via extraResources at
    // <resourcesPath>/frontend/dist/index.html. Dev mode uses Vite URL.
    const candidates = [
      join(process.resourcesPath, 'frontend', 'dist', 'index.html'),
      join(__dirname, '..', '..', 'frontend', 'dist', 'index.html'),
      join(app.getAppPath(), 'frontend', 'dist', 'index.html'),
    ]
    const distIndex = candidates.find((p) => existsSync(p))
    if (!distIndex) {
      console.error('[gotutor] frontend index.html not found in any of:', candidates)
    } else {
      mainWindow.loadFile(distIndex).catch((e) => {
        console.error(`[gotutor] failed to load ${distIndex}:`, e)
      })
    }
  }

  mainWindow.on('closed', () => {
    mainWindow = null
  })
}

app.whenReady().then(() => {
  bootstrap().catch((e) => console.error('[gotutor] bootstrap failed:', e))
})

app.on('activate', () => {
  if (BrowserWindow.getAllWindows().length === 0 && backend) {
    createWindow()
  }
})

app.on('before-quit', async (event) => {
  if (backend) {
    event.preventDefault()
    await backend.stop()
    backend = null
    app.quit()
  }
})

app.on('window-all-closed', () => {
  if (backend) {
    backend.stop().finally(() => app.quit())
  } else {
    app.quit()
  }
})
