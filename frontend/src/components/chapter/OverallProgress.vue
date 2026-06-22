<script setup lang="ts">
import { computed } from 'vue'
import type { ProgressResponse } from '../../api/types'

const props = defineProps<{
  progress: ProgressResponse | null
}>()

const pct = computed(() => props.progress?.percent ?? 0)
const done = computed(() => props.progress?.completedChapters ?? 0)
const total = computed(() => props.progress?.totalChapters ?? 0)
</script>

<template>
  <section v-if="progress" class="overall">
    <div class="row">
      <span class="label">{{ $t('overall.progress', { done, total }) }}</span>
      <span class="pct">{{ pct }}%</span>
    </div>
    <div class="bar" aria-hidden="true">
      <div class="fill" :style="{ width: `${pct}%` }" />
    </div>
  </section>
</template>

<style scoped>
.overall {
  display: flex;
  flex-direction: column;
  gap: var(--space-2);
}
.row {
  display: flex;
  justify-content: space-between;
  align-items: baseline;
  font-size: var(--text-sm);
}
.label {
  color: var(--fg-muted);
}
.pct {
  color: var(--fg);
  font-weight: 600;
  font-variant-numeric: tabular-nums;
}
.bar {
  height: 6px;
  background: var(--surface-2);
  border-radius: 999px;
  overflow: hidden;
}
.fill {
  height: 100%;
  background: var(--accent);
  border-radius: inherit;
  transition: width 0.3s var(--ease-out-expo, ease-out);
}
</style>
