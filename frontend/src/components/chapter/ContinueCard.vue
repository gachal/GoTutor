<script setup lang="ts">
import { computed } from 'vue'
import { useRouter } from 'vue-router'
import type { Chapter, ProgressResponse } from '../../api/types'

const props = defineProps<{
  progress: ProgressResponse
  chapters: Chapter[]
}>()

const router = useRouter()

// Resolve lastChapterId → chapter title. If the user has completed every
// chapter, surface "all done" instead.
const lastChapter = computed<Chapter | undefined>(() => {
  if (!props.progress.lastChapterId) return undefined
  return props.chapters.find((c) => c.id === props.progress.lastChapterId)
})

const allDone = computed(() => props.progress.completedChapters === props.progress.totalChapters)

function continueLast() {
  if (lastChapter.value) {
    router.push(`/chapter/${lastChapter.value.id}`)
  }
}
</script>

<template>
  <section v-if="allDone" class="all-done">
    <h2>{{ $t('all_done.title') }}</h2>
    <p>{{ $t('all_done.body') }}</p>
  </section>

  <section v-else-if="lastChapter" class="continue-card">
    <div class="left">
      <p class="label">{{ $t('continue.label') }}</p>
      <p class="title">{{ $t('continue.subtitle', { title: lastChapter.title }) }}</p>
    </div>
    <button type="button" class="cta" @click="continueLast">
      {{ $t('continue.cta') }}
    </button>
  </section>
</template>

<style scoped>
.continue-card {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: var(--space-4);
  padding: var(--space-4) var(--space-5);
  background: color-mix(in oklch, var(--accent) 12%, var(--surface));
  border: 1px solid color-mix(in oklch, var(--accent) 30%, var(--border));
  border-radius: var(--radius-lg);
}
.left {
  display: flex;
  flex-direction: column;
  gap: 2px;
  min-width: 0;
}
.label {
  font-size: var(--text-xs);
  text-transform: uppercase;
  letter-spacing: 0.06em;
  color: var(--accent);
  margin: 0;
  font-weight: 600;
}
.title {
  font-size: 16px;
  color: var(--fg);
  margin: 0;
  font-weight: 500;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.cta {
  flex-shrink: 0;
  background: var(--accent);
  color: var(--accent-fg);
  border: 0;
  border-radius: var(--radius);
  padding: var(--space-2) var(--space-4);
  font-size: var(--text-sm);
  font-weight: 600;
  cursor: pointer;
}
.cta:hover { filter: brightness(1.05); }

.all-done {
  padding: var(--space-5);
  background: var(--surface);
  border: 1px solid var(--border);
  border-radius: var(--radius-lg);
  text-align: center;
  display: flex;
  flex-direction: column;
  gap: var(--space-2);
}
.all-done h2 {
  font-size: 18px;
  font-weight: 700;
  color: var(--fg);
  margin: 0;
}
.all-done p {
  font-size: var(--text-sm);
  color: var(--fg-muted);
  margin: 0;
  line-height: 1.6;
}
</style>
