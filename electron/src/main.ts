import { app, BrowserWindow } from 'electron'
import { join } from 'path'
import { detectGo } from './goDetector'
import { spawnBackend, BackendHandle } from './backend'
import { waitForPortFile, defaultPortFile } from './portDiscovery'

// Global so we can stop() it on before-quit even if createWindow
// captured the handle in a closure.
let backend: BackendHandle | null = null
let mainWindow: BrowserWindow | null = null

const DEV_SERVER_URL = process.env.GOTUTOR_DEV_URL || 'http://localhost:5173'

async function bootstrap() {
  const go = detectGo()
  if (!go.found) {
    // We still launch — the renderer shows an install-Go screen and
    // polls /api/health (which will report goFound=false).
    console.warn('[gotutor] Go toolchain not found on PATH')
  } else {
    console.log(`[gotutor] Go: ${go.version} @ ${go.path}`)
  }

  backend = spawnBackend()
  backend.process.on('exit', (code, signal) => {
    console.error(`[gotutor] backend exited code=${code} signal=${signal}`)
    if (mainWindow && !mainWindow.isDestroyed()) {
      mainWindow.webContents.send('backend:exited', { code, signal })
    }
  })

  let port: number
  try {
    port = await waitForPortFile(defaultPortFile())
    backend.port = port
    console.log(`[gotutor] backend listening on :${port}`)
  } catch (e) {
    console.error(`[gotutor] ${e instanceof Error ? e.message : String(e)}`)
    port = 0
  }

  createWindow(port)
}

function createWindow(backendPort: number) {
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

  mainWindow.webContents.on('did-finish-load', () => {
    mainWindow?.webContents.executeJavaScript(
      `window.__GOTUTOR_BACKEND_PORT__ = ${backendPort};`,
      true,
    ).catch(() => { /* ignore */ })
  })

  if (process.env.GOTUTOR_DEV === '1' || process.env.GOTUTOR_DEV_URL) {
    mainWindow.loadURL(DEV_SERVER_URL)
    mainWindow.webContents.openDevTools({ mode: 'detach' })
  } else {
    const distIndex = join(__dirname, '..', '..', 'frontend', 'dist', 'index.html')
    mainWindow.loadFile(distIndex).catch((e) => {
      console.error(`[gotutor] failed to load ${distIndex}:`, e)
    })
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
    createWindow(backend.port)
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
