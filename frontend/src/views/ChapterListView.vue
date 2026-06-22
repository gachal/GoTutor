<script setup lang="ts">
import { computed, onMounted } from 'vue'
import { useChaptersStore } from '../stores/chapters'
import OverallProgress from '../components/chapter/OverallProgress.vue'
import ContinueCard from '../components/chapter/ContinueCard.vue'
import TrackSection from '../components/chapter/TrackSection.vue'
import type { Track } from '../api/types'

const chapters = useChaptersStore()

// Track display order is a frontend concern — the API returns track
// *identity* per chapter, but the *order* we render the sections is a
// presentation choice. Fixed list keeps fundamentals → concurrency →
// gateway regardless of registry iteration order.
const TRACK_ORDER: Track[] = ['fundamentals', 'concurrency', 'gateway']

// Partition chapters by track, preserving per-track ordering (which
// already comes from the registry sorted by Ordinal). Sections with
// zero chapters (none in current content, but defensive) are skipped.
const tracks = computed(() =>
  TRACK_ORDER.map((trackId) => ({
    trackId,
    chapters: chapters.list.filter((c) => c.track === trackId),
  })).filter((t) => t.chapters.length > 0)
)

onMounted(() => {
  // fetchList runs once via App.vue's health watch; fetchProgress is the
  // Phase 3 addition. Cheap call, fine to fire on every list-view mount.
  chapters.fetchProgress()
})
</script>

<template>
  <section class="home">
    <header class="hero">
      <h1>{{ $t('app.name') }}</h1>
      <p class="lede">{{ $t('app.tagline') }}</p>
    </header>

    <p v-if="chapters.loading" class="muted">{{ $t('common.retry') }}…</p>
    <p v-else-if="chapters.error" class="error">{{ $t('errors.backend_down') }}</p>

    <template v-else>
      <OverallProgress :progress="chapters.progress" />

      <ContinueCard
        v-if="chapters.progress"
        :progress="chapters.progress"
        :chapters="chapters.list"
      />

      <div class="track-list">
        <TrackSection
          v-for="t in tracks"
          :key="t.trackId"
          :track-id="t.trackId"
          :chapters="t.chapters"
        />
      </div>
    </template>
  </section>
</template>

<style scoped>
.home {
  padding: var(--space-6);
  max-width: 1080px;
  display: flex;
  flex-direction: column;
  gap: var(--space-5);
}
.hero h1 {
  font-size: 36px;
  font-weight: 700;
  margin: 0;
}
.lede {
  font-size: 16px;
  color: var(--fg-muted);
  margin: var(--space-1) 0 0;
}
.muted { color: var(--fg-subtle); }
.error { color: var(--danger); }
.track-list {
  display: flex;
  flex-direction: column;
  gap: var(--space-6);
  margin-top: var(--space-2);
}
</style>
