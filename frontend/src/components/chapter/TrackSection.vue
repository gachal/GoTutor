<script setup lang="ts">
import ChapterCard from './ChapterCard.vue'
import type { Chapter } from '../../api/types'

defineProps<{
  trackId: string
  chapters: Chapter[]
}>()
</script>

<template>
  <section class="track-section">
    <header class="track-header">
      <h2 class="track-title">{{ $t(`tracks.${trackId}.title`) }}</h2>
      <p class="track-subtitle">{{ $t(`tracks.${trackId}.subtitle`) }}</p>
    </header>
    <div class="card-grid">
      <ChapterCard
        v-for="ch in chapters"
        :key="ch.id"
        :chapter="ch"
      />
    </div>
  </section>
</template>

<style scoped>
.track-section {
  display: flex;
  flex-direction: column;
  gap: var(--space-3);
}
.track-header {
  display: flex;
  flex-direction: column;
  gap: 2px;
}
.track-title {
  font-size: 18px;
  font-weight: 700;
  color: var(--fg);
  margin: 0;
}
.track-subtitle {
  font-size: var(--text-sm);
  color: var(--fg-muted);
  margin: 0;
}
.card-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
  gap: var(--space-3);
}
</style>
