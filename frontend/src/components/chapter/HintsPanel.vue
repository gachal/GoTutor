<script setup lang="ts">
import type { Todo } from '../../api/types'

defineProps<{
  todos: Todo[]
}>()

const emit = defineEmits<{
  (e: 'jump', line: number): void
}>()
</script>

<template>
  <aside class="hints-panel" role="region" :aria-label="$t('hints.title')">
    <header class="panel-header">
      <h2>{{ $t('hints.title') }}</h2>
      <p class="subtitle">{{ $t('hints.subtitle') }}</p>
    </header>

    <ol class="hint-list">
      <li v-for="(t, i) in todos" :key="t.line" class="hint-card">
        <div class="hint-head">
          <span class="hint-num">TODO {{ i + 1 }}</span>
          <button type="button" class="jump-btn" @click="emit('jump', t.line)">
            {{ $t('hints.jump') }}
          </button>
        </div>
        <p v-if="t.hint" class="hint-body">{{ t.hint }}</p>
        <p v-else class="hint-body muted">{{ $t('hints.no_hint') }}</p>
      </li>
    </ol>
  </aside>
</template>

<style scoped>
.hints-panel {
  flex-shrink: 0;
  width: min(420px, 38vw);
  min-height: 0;
  display: flex;
  flex-direction: column;
  border: 1px solid var(--border);
  border-radius: var(--radius-lg);
  background: var(--surface);
  overflow: hidden;
}
.panel-header {
  padding: var(--space-3) var(--space-4);
  border-bottom: 1px solid var(--border);
}
.panel-header h2 {
  font-size: 16px;
  font-weight: 700;
  margin: 0;
}
.subtitle {
  font-size: var(--text-xs);
  color: var(--fg-muted);
  margin-top: var(--space-1);
}
.hint-list {
  list-style: none;
  margin: 0;
  padding: var(--space-3);
  display: flex;
  flex-direction: column;
  gap: var(--space-3);
  overflow: auto;
  flex: 1;
  min-height: 0;
}
.hint-card {
  padding: var(--space-3);
  background: var(--surface-2);
  border-radius: var(--radius);
  display: flex;
  flex-direction: column;
  gap: var(--space-2);
}
.hint-head {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
.hint-num {
  font-size: var(--text-xs);
  text-transform: uppercase;
  letter-spacing: 0.06em;
  color: var(--fg-subtle);
  font-weight: 600;
}
.jump-btn {
  background: transparent;
  border: 1px solid var(--border);
  color: var(--fg-muted);
  font-size: var(--text-xs);
  padding: 2px 8px;
  border-radius: var(--radius);
  cursor: pointer;
}
.jump-btn:hover {
  background: var(--surface);
  color: var(--fg);
  border-color: var(--accent);
}
.hint-body {
  font-size: var(--text-sm);
  color: var(--fg);
  line-height: 1.6;
  margin: 0;
}
.hint-body.muted { color: var(--fg-subtle); }
</style>
