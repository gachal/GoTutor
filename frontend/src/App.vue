<script setup lang="ts">
import { onMounted, watch } from 'vue'
import { useThemeStore } from './stores/theme'
import { useLocaleStore } from './stores/locale'
import { useChaptersStore } from './stores/chapters'
import { useHealthStore } from './stores/health'
import { useOnboardingStore } from './stores/onboarding'
import { setI18nLocale } from './i18n'
import BootGate from './components/BootGate.vue'
import WelcomeView from './views/WelcomeView.vue'

const theme = useThemeStore()
const locale = useLocaleStore()
const chapters = useChaptersStore()
const health = useHealthStore()
const onboarding = useOnboardingStore()

// Locale switch: refresh static UI strings, point API at the new locale
// (Accept-Language), AND re-fetch the chapter list so titles/descriptions
// (which the backend localizes server-side) refresh. Detail-page template
// re-fetch is handled per-view in ChapterView.vue.
//
// The fetchList() call is gated on health.status === 'ready' so a locale
// toggle during the boot gate doesn't trip ECONNREFUSED.
watch(() => locale.locale, (l) => {
  setI18nLocale(l)
  if (health.status === 'ready') chapters.fetchList()
}, { immediate: true })

// Boot the health poller immediately — it drives the boot gate. When the
// gate flips to ready, fetch the chapter list. This is the only path
// that touches /api/chapters before the user sees the home screen, so
// ECONNREFUSED on a cold backend is impossible (the gate holds until
// /api/health succeeds).
onMounted(() => {
  health.startPolling()
})

watch(() => health.status, (s) => {
  if (s === 'ready' && chapters.list.length === 0) {
    chapters.fetchList()
  }
})
</script>

<template>
  <div class="app-shell">
    <aside class="sidebar">
      <header class="app-header">
        <div class="app-header-row">
          <div class="app-header-text">
            <h1 class="app-name">{{ $t('app.name') }}</h1>
            <p class="app-tagline">{{ $t('app.tagline') }}</p>
          </div>
          <button
            type="button"
            class="help-btn"
            :title="$t('welcome.reopen_label')"
            :aria-label="$t('welcome.reopen_label')"
            @click="onboarding.reopen"
          >?</button>
        </div>
      </header>

      <div class="toggles">
        <div class="toggle-group">
          <span class="toggle-label">{{ $t('sidebar.theme.label') }}</span>
          <div class="segmented">
            <button
              v-for="m in (['system', 'light', 'dark'] as const)"
              :key="m"
              :class="['seg', { active: theme.mode === m }]"
              type="button"
              @click="theme.setMode(m)"
            >{{ $t(`sidebar.theme.${m}`) }}</button>
          </div>
        </div>

        <div class="toggle-group">
          <span class="toggle-label">{{ $t('sidebar.locale.label') }}</span>
          <div class="segmented">
            <button
              v-for="l in (['zh-CN', 'en'] as const)"
              :key="l"
              :class="['seg', { active: locale.locale === l }]"
              type="button"
              @click="locale.set(l)"
            >{{ l === 'zh-CN' ? '中文' : 'EN' }}</button>
          </div>
        </div>
      </div>

      <nav class="chapter-nav" role="listbox" :aria-label="$t('sidebar.title')">
        <h2 class="nav-title">{{ $t('sidebar.title') }}</h2>
        <ul>
          <li
            v-for="ch in chapters.list"
            :key="ch.id"
            :class="['chapter-item', { completed: ch.completed }]"
          >
            <router-link :to="`/chapter/${ch.id}`" class="chapter-link">
              <span class="ord">{{ ch.ordinal }}</span>
              <span class="title">{{ ch.title }}</span>
              <span v-if="ch.completed" class="badge" aria-label="completed">✓</span>
            </router-link>
          </li>
        </ul>
      </nav>
    </aside>

    <main class="content">
      <BootGate v-if="health.status !== 'ready'" />
      <RouterView v-else />
    </main>

    <!-- Welcome overlay sits at the shell root so it can cover the whole
         window with position:fixed, regardless of which route is active. -->
    <WelcomeView v-if="health.status === 'ready'" />
  </div>
</template>

<style scoped>
.app-shell {
  display: grid;
  grid-template-columns: var(--sidebar-w) 1fr;
  height: 100%;
}

.sidebar {
  display: flex;
  flex-direction: column;
  border-right: 1px solid var(--border);
  background: var(--surface);
  padding: var(--space-5) var(--space-4);
  gap: var(--space-5);
  overflow-y: auto;
}

.app-header-row {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: var(--space-2);
}

.app-name {
  font-size: 22px;
  font-weight: 700;
  color: var(--fg);
}

.app-tagline {
  font-size: var(--text-sm);
  color: var(--fg-muted);
  margin-top: var(--space-1);
}

.help-btn {
  flex-shrink: 0;
  width: 24px;
  height: 24px;
  border-radius: 50%;
  border: 1px solid var(--border);
  background: transparent;
  color: var(--fg-muted);
  font-size: 13px;
  font-weight: 600;
  cursor: pointer;
  display: grid;
  place-items: center;
  margin-top: 2px;
}
.help-btn:hover {
  background: var(--surface-2);
  color: var(--fg);
}

.toggles {
  display: flex;
  flex-direction: column;
  gap: var(--space-3);
}

.toggle-group {
  display: flex;
  flex-direction: column;
  gap: var(--space-1);
}

.toggle-label {
  font-size: var(--text-xs);
  color: var(--fg-subtle);
  text-transform: uppercase;
  letter-spacing: 0.04em;
}

.segmented {
  display: inline-flex;
  border: 1px solid var(--border);
  border-radius: var(--radius);
  overflow: hidden;
}

.seg {
  background: transparent;
  color: var(--fg-muted);
  border: 0;
  padding: var(--space-1) var(--space-3);
  font-size: var(--text-sm);
  border-right: 1px solid var(--border);
}
.seg:last-child { border-right: 0; }
.seg:hover { background: var(--surface-2); }
.seg.active {
  background: var(--accent);
  color: var(--accent-fg);
}

.chapter-nav {
  flex: 1;
  min-height: 0;
}

.nav-title {
  font-size: var(--text-xs);
  text-transform: uppercase;
  letter-spacing: 0.06em;
  color: var(--fg-subtle);
  margin-bottom: var(--space-2);
}

.chapter-nav ul {
  list-style: none;
  display: flex;
  flex-direction: column;
  gap: var(--space-1);
}

.chapter-item {
  border-radius: var(--radius);
}

.chapter-link {
  display: grid;
  grid-template-columns: auto 1fr auto;
  align-items: center;
  gap: var(--space-3);
  padding: var(--space-2) var(--space-3);
  text-decoration: none;
  color: var(--fg);
  border-radius: var(--radius);
}
.chapter-link:hover { background: var(--surface-2); }
.router-link-active.chapter-link {
  background: var(--surface-2);
  color: var(--accent);
}

.ord {
  display: inline-grid;
  place-items: center;
  width: 24px;
  height: 24px;
  border-radius: 50%;
  background: var(--surface-2);
  font-size: var(--text-xs);
  color: var(--fg-muted);
  flex-shrink: 0;
}

.title {
  font-size: var(--text-sm);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.badge {
  color: var(--success);
  font-weight: 700;
}

.content {
  height: 100%;
  overflow: auto;
  background: var(--bg);
}
</style>
