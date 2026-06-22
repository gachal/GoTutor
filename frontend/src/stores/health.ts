import { defineStore } from 'pinia'
import { ref } from 'vue'
import { api } from '../api/client'

// HealthStatus drives the boot gate in App.vue:
//   loading      — first probe in flight, no answer yet
//   ready        — /api/health returned with goFound=true → main app
//   goMissing    — backend up but Go toolchain absent → install screen
//   backendDown  — /api/health failed (ECONNREFUSED, timeout, 5xx) → retry screen
export type HealthStatus = 'loading' | 'ready' | 'goMissing' | 'backendDown'

// Poll cadence: every 2s until ready. Fast enough to catch a backend
// cold-starting in ~100ms without making the user wait the old 500ms
// hard delay; loose enough not to hammer the server. We stop polling
// once status flips to ready — health changes mid-session are rare and
// the user would notice when they submit anyway.
const POLL_INTERVAL_MS = 2_000

export const useHealthStore = defineStore('health', () => {
  const status = ref<HealthStatus>('loading')
  const goVersion = ref('')
  let pollHandle: number | null = null

  async function check() {
    try {
      const r = await api.health()
      goVersion.value = r.goVersion
      const next: HealthStatus = r.goFound ? 'ready' : 'goMissing'
      status.value = next
      if (next === 'ready') stopPolling()
    } catch {
      // Backend not up yet (cold start) or crashed mid-session — either
      // way the user should see the retry screen. Polling continues so
      // we auto-recover when the backend comes back.
      status.value = 'backendDown'
    }
  }

  function startPolling() {
    void check()
    if (pollHandle !== null) return
    pollHandle = window.setInterval(() => { void check() }, POLL_INTERVAL_MS)
  }

  function stopPolling() {
    if (pollHandle !== null) {
      clearInterval(pollHandle)
      pollHandle = null
    }
  }

  return { status, goVersion, check, startPolling, stopPolling }
})
