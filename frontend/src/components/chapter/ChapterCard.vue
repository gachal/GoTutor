<script setup lang="ts">
import { computed } from 'vue'
import { useRunsStore } from '../../stores/runs'
import type { Chapter } from '../../api/types'

const props = defineProps<{ chapter: Chapter }>()
const runs = useRunsStore()

// Card status is inferred from completion + practice count:
//   completed   — chapter.completed === true (passed at least once)
//   in_progress — user has run it but never passed
//   not_started — no runs yet
// "completed" wins over "in_progress" once the user passes.
const status = computed<'completed' | 'in_progress' | 'not_started'>(() => {
  if (props.chapter.completed) return 'completed'
  if (runs.get(props.chapter.id) > 0) return 'in_progress'
  return 'not_started'
})

const practicedCount = computed(() => runs.get(props.chapter.id))

// Binary progress bar fill — 100% when completed, 0% otherwise. Phase 6
// may add partial progress (e.g. based on TODOs filled); for now this
// gives visual rhythm without false precision.
const progressPct = computed(() => (props.chapter.completed ? 100 : 0))
</script>

<template>
  <router-link :to="`/chapter/${chapter.id}`" class="card">
    <div class="meta-row">
      <span class="ord" :aria-label="`Chapter ${chapter.ordinal}`">{{ chapter.ordinal }}</span>
      <span class="dot" aria-hidden="true">·</span>
      <span class="difficulty">{{ $t(`difficulty.${chapter.difficulty}`) }}</span>
      <span class="dot" aria-hidden="true">·</span>
      <span class="minutes">{{ $t('estimated_time', { minutes: chapter.estimatedMinutes }) }}</span>
      <span class="spacer" />
      <span
        v-if="status === 'completed'"
        class="status status-completed"
        :aria-label="$t('card.status.completed')"
      >✓ {{ $t('card.status.completed') }}</span>
      <span
        v-else-if="status === 'in_progress'"
        class="status status-in-progress"
        :aria-label="$t('card.status.in_progress')"
      >{{ $t('card.status.in_progress') }}</span>
    </div>

    <h3 class="title">{{ chapter.title }}</h3>
    <p class="desc">{{ chapter.description }}</p>

    <div class="progress-track" aria-hidden="true">
      <div class="progress-fill" :style="{ width: `${progressPct}%` }" />
    </div>

    <p v-if="practicedCount > 0" class="practiced">{{ $t('card.practiced', { count: practicedCount }) }}</p>
  </router-link>
</template>

<style scoped>
.card {
  display: flex;
  flex-direction: column;
  gap: var(--space-2);
  padding: var(--space-4);
  background: var(--surface);
  border: 1px solid var(--border);
  border-radius: var(--radius-lg);
  text-decoration: none;
  color: var(--fg);
  transition: background 0.15s, border-color 0.15s;
}
.card:hover {
  background: var(--surface-2);
  border-color: color-mix(in oklch, var(--accent) 30%, var(--border));
}
.meta-row {
  display: flex;
  align-items: center;
  gap: var(--space-2);
  font-size: var(--text-xs);
  color: var(--fg-muted);
  flex-wrap: wrap;
}
.ord {
  display: inline-grid;
  place-items: center;
  min-width: 22px;
  height: 22px;
  border-radius: 50%;
  background: var(--surface-2);
  color: var(--fg);
  font-weight: 600;
  padding: 0 4px;
}
.dot { color: var(--fg-subtle); }
.difficulty {
  text-transform: lowercase;
  letter-spacing: 0.02em;
}
.minutes { font-variant-numeric: tabular-nums; }
.spacer { flex: 1; }
.status {
  font-size: var(--text-xs);
  font-weight: 600;
  padding: 2px 8px;
  border-radius: 999px;
}
.status-completed {
  background: color-mix(in oklch, var(--success) 18%, transparent);
  color: var(--success);
}
.status-in-progress {
  background: var(--surface-2);
  color: var(--fg-muted);
}
.title {
  font-size: 16px;
  font-weight: 600;
  margin: 0;
}
.desc {
  font-size: var(--text-sm);
  color: var(--fg-muted);
  margin: 0;
  line-height: 1.5;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}
.progress-track {
  height: 4px;
  background: var(--surface-2);
  border-radius: 999px;
  overflow: hidden;
  margin-top: var(--space-1);
}
.progress-fill {
  height: 100%;
  background: var(--accent);
  border-radius: inherit;
  transition: width 0.3s var(--ease-out-expo, ease-out);
}
.practiced {
  font-size: var(--text-xs);
  color: var(--fg-subtle);
  margin: 0;
}
</style>
