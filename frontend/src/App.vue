<script setup lang="ts">
import { onMounted, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { useThemeStore } from './stores/theme'
import { useLocaleStore } from './stores/locale'
import { useChaptersStore } from './stores/chapters'
import { setI18nLocale } from './i18n'

const { t } = useI18n()
const theme = useThemeStore()
const locale = useLocaleStore()
const chapters = useChaptersStore()

watch(() => locale.locale, (l) => setI18nLocale(l), { immediate: true })

onMounted(() => {
  chapters.fetchList()
})

async function confirmReset() {
  if (!window.confirm(t('sidebar.reset.confirm'))) return
  await chapters.reset()
}
</script>

<template>
  <div class="app-shell">
    <aside class="sidebar">
      <header class="app-header">
        <h1 class="app-name">{{ $t('app.name') }}</h1>
        <p class="app-tagline">{{ $t('app.tagline') }}</p>
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
            :class="['chapter-item', { locked: !ch.unlocked, completed: ch.completed }]"
          >
            <router-link
              v-if="ch.unlocked"
              :to="`/chapter/${ch.id}`"
              class="chapter-link"
            >
              <span class="ord">{{ ch.ordinal }}</span>
              <span class="title">{{ ch.title }}</span>
              <span v-if="ch.completed" class="badge" aria-label="completed">✓</span>
            </router-link>
            <div v-else class="chapter-link disabled" :title="$t('chapter.locked')">
              <span class="ord">🔒</span>
              <span class="title">{{ ch.title }}</span>
            </div>
          </li>
        </ul>
      </nav>

      <footer class="sidebar-footer">
        <button
          class="reset-btn"
          type="button"
          @click="confirmReset"
        >{{ $t('sidebar.reset.label') }}</button>
      </footer>
    </aside>

    <main class="content">
      <RouterView />
    </main>
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

.chapter-link.disabled {
  opacity: 0.5;
  cursor: not-allowed;
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

.sidebar-footer {
  border-top: 1px solid var(--border);
  padding-top: var(--space-3);
}

.reset-btn {
  background: transparent;
  color: var(--fg-muted);
  border: 1px solid var(--border);
  border-radius: var(--radius);
  padding: var(--space-2) var(--space-3);
  width: 100%;
  font-size: var(--text-sm);
}
.reset-btn:hover {
  background: var(--surface-2);
  color: var(--danger);
}

.content {
  height: 100%;
  overflow: auto;
  background: var(--bg);
}
</style>
