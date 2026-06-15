// prebuild.js — asserts the Go backend binary for the current target
// exists before electron-builder runs. electron-builder silently copies
// nothing when extraResources.from is empty, which produces a .app that
// crashes on launch with ENOENT for Resources/backend/gotutor-backend.
//
// This script fails fast with a clear error so the package flow stops
// at the right step.
//
// Usage: invoked from `pnpm package` (see electron/package.json). Reads
// the --mac/--win/--linux and --arm64/--x64 flags from argv; defaults
// to the host platform/arch.

const fs = require('fs')
const path = require('path')

function parseArgs(argv) {
  let osFlag = null
  let archFlag = null
  for (const a of argv.slice(2)) {
    if (a === '--mac' || a === '--win' || a === '--linux') osFlag = a.slice(2)
    else if (a === '--arm64' || a === '--x64') archFlag = a.slice(2)
  }
  if (!osFlag) {
    osFlag = process.platform === 'darwin' ? 'mac' : process.platform === 'win32' ? 'win' : 'linux'
  }
  if (!archFlag) {
    archFlag = process.arch === 'arm64' ? 'arm64' : 'x64'
  }
  return { os: osFlag, arch: archFlag }
}

function main() {
  const { os: osName, arch } = parseArgs(process.argv)
  const dirName = `${osName}-${arch}`
  const binName = osName === 'win' ? 'gotutor-backend.exe' : 'gotutor-backend'

  // electron/scripts/ → electron/ → GoTutor/
  const repoRoot = path.resolve(__dirname, '..', '..')
  const binPath = path.join(repoRoot, 'backend', 'bin', dirName, binName)
  const frontendIndex = path.join(repoRoot, 'frontend', 'dist', 'index.html')

  const errors = []
  if (!fs.existsSync(binPath)) {
    errors.push(
      `backend binary not found at: ${binPath}`,
      `  fix: run \`make backend-build-${goosTarget(osName)}-${goarchTarget(arch)}\``,
    )
  } else {
    const stat = fs.statSync(binPath)
    console.log(`[prebuild] OK: ${path.relative(repoRoot, binPath)} (${(stat.size / 1024 / 1024).toFixed(1)} MB)`)
  }

  if (!fs.existsSync(frontendIndex)) {
    errors.push(
      `frontend build not found at: ${frontendIndex}`,
      `  fix: run \`make frontend-build\``,
    )
  } else {
    console.log(`[prebuild] OK: ${path.relative(repoRoot, frontendIndex)}`)
  }

  if (errors.length > 0) {
    console.error('[prebuild] FAILED:')
    for (const e of errors) console.error('  ' + e)
    console.error('[prebuild] Or run `make package-' + packageTarget(osName) + '` which builds everything.')
    process.exit(1)
  }
}

function goosTarget(osName) {
  if (osName === 'mac') return 'darwin'
  if (osName === 'win') return 'windows'
  return 'linux'
}

function goarchTarget(arch) {
  if (arch === 'x64') return 'amd64'
  return 'arm64'
}

function packageTarget(osName) {
  if (osName === 'mac') return 'darwin'
  if (osName === 'win') return 'win'
  return 'linux'
}

main()
