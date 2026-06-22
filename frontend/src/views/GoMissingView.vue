<script setup lang="ts">
import { computed, ref } from 'vue'
import { useHealthStore } from '../stores/health'

const health = useHealthStore()
const rechecking = ref(false)

// Platform is exposed by the Electron preload. Default to 'mac' outside
// Electron (dev/Vite) since that's the only platform where packaging
// currently ships.
type Platform = 'mac' | 'windows' | 'linux'
const platform = computed<Platform>(() => {
  const p = (window as unknown as { gotutor?: { platform?: string } }).gotutor?.platform
  if (p === 'win32') return 'windows'
  if (p === 'linux') return 'linux'
  return 'mac'
})

const whyOpen = ref(false)

async function recheck() {
  if (rechecking.value) return
  rechecking.value = true
  await health.check()
  // status flip is reactive; rechecking flag returns regardless of result
  rechecking.value = false
}
</script>

<template>
  <div class="go-missing">
    <h2 class="headline">{{ $t('go_missing.title') }}</h2>
    <p class="body">{{ $t('go_missing.body_1') }}</p>
    <p class="body">{{ $t('go_missing.body_2') }}</p>

    <!-- macOS -->
    <div v-if="platform === 'mac'" class="install-block">
      <h3 class="install-label">{{ $t('go_missing.install_mac.label') }}</h3>
      <p class="step">{{ $t('go_missing.install_mac.step_1') }}</p>
      <pre class="cmd"><code>brew install go</code></pre>
      <p class="step">{{ $t('go_missing.install_mac.step_2') }}</p>
      <a class="download-link" href="https://go.dev/dl/" target="_blank" rel="noopener">
        {{ $t('go_missing.install_mac.download') }} ↗
      </a>
    </div>

    <!-- Windows -->
    <div v-else-if="platform === 'windows'" class="install-block">
      <h3 class="install-label">{{ $t('go_missing.install_windows.label') }}</h3>
      <p class="step">{{ $t('go_missing.install_windows.step_1') }}</p>
      <a class="download-link" href="https://go.dev/dl/" target="_blank" rel="noopener">
        {{ $t('go_missing.install_windows.download') }} ↗
      </a>
      <p class="step">{{ $t('go_missing.install_windows.step_2') }}</p>
    </div>

    <!-- Linux -->
    <div v-else class="install-block">
      <h3 class="install-label">{{ $t('go_missing.install_linux.label') }}</h3>
      <p class="step">{{ $t('go_missing.install_linux.step_1') }}</p>
      <a class="download-link" href="https://go.dev/dl/" target="_blank" rel="noopener">
        {{ $t('go_missing.install_linux.download') }} ↗
      </a>
      <p class="step">{{ $t('go_missing.install_linux.step_2') }}</p>
      <pre class="cmd"><code>tar -C /usr/local -xzf go*.tar.gz
export PATH=$PATH:/usr/local/go/bin</code></pre>
    </div>

    <button type="button" class="recheck-btn" :disabled="rechecking" @click="recheck">
      {{ rechecking ? $t('go_missing.rechecking') : $t('go_missing.recheck') }}
    </button>

    <button type="button" class="why-link" @click="whyOpen = !whyOpen">
      {{ $t('go_missing.why_go') }}
    </button>
    <p v-if="whyOpen" class="why-body">{{ $t('go_missing.why_go_body') }}</p>
  </div>
</template>

<style scoped>
.go-missing {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: var(--space-3);
  max-width: 560px;
  text-align: center;
  padding: var(--space-2);
}
.headline {
  font-size: 24px;
  font-weight: 700;
  color: var(--fg);
  margin: 0 0 var(--space-1);
}
.body {
  font-size: var(--text-sm);
  color: var(--fg-muted);
  line-height: 1.6;
  margin: 0;
}
.install-block {
  width: 100%;
  margin-top: var(--space-4);
  padding: var(--space-4);
  background: var(--surface);
  border: 1px solid var(--border);
  border-radius: var(--radius-lg);
  text-align: left;
  display: flex;
  flex-direction: column;
  gap: var(--space-2);
}
.install-label {
  font-size: 12px;
  text-transform: uppercase;
  letter-spacing: 0.06em;
  color: var(--fg-subtle);
  margin: 0;
}
.step {
  font-size: var(--text-sm);
  color: var(--fg);
  margin: 0;
}
.cmd {
  background: var(--surface-2);
  border-radius: var(--radius);
  padding: var(--space-2) var(--space-3);
  font-family: var(--font-mono);
  font-size: var(--text-sm);
  color: var(--fg);
  margin: 0;
  overflow-x: auto;
}
.download-link {
  color: var(--accent);
  font-size: var(--text-sm);
  font-weight: 600;
  text-decoration: none;
}
.download-link:hover { text-decoration: underline; }
.recheck-btn {
  margin-top: var(--space-4);
  background: var(--accent);
  color: var(--accent-fg);
  border: 0;
  border-radius: var(--radius);
  padding: var(--space-2) var(--space-5);
  font-size: var(--text-sm);
  font-weight: 600;
  cursor: pointer;
}
.recheck-btn:hover:not(:disabled) { filter: brightness(1.05); }
.recheck-btn:disabled { opacity: 0.6; cursor: progress; }
.why-link {
  background: transparent;
  border: 0;
  color: var(--fg-muted);
  font-size: var(--text-sm);
  cursor: pointer;
  text-decoration: underline;
  padding: 0;
  margin-top: var(--space-2);
}
.why-body {
  font-size: var(--text-sm);
  color: var(--fg-muted);
  line-height: 1.6;
  margin: var(--space-2) 0 0;
  max-width: 480px;
}
</style>
