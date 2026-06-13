<script setup lang="ts">
import { useChaptersStore } from '../stores/chapters'

const chapters = useChaptersStore()
</script>

<template>
  <section class="hero">
    <h1>{{ $t('app.name') }}</h1>
    <p class="lede">{{ $t('app.tagline') }}</p>

    <p v-if="chapters.loading">{{ $t('common.retry') }}…</p>
    <p v-else-if="chapters.error" class="error">{{ $t('errors.backend_down') }}</p>
    <ul v-else class="chapter-grid">
      <li
        v-for="ch in chapters.list"
        :key="ch.id"
        :class="['tile', { locked: !ch.unlocked }]"
      >
        <router-link v-if="ch.unlocked" :to="`/chapter/${ch.id}`" class="tile-link">
          <span class="tile-ord">{{ ch.ordinal }}</span>
          <span class="tile-title">{{ ch.title }}</span>
          <span class="tile-desc">{{ ch.description }}</span>
        </router-link>
        <div v-else class="tile-link disabled">
          <span class="tile-ord">🔒</span>
          <span class="tile-title">{{ ch.title }}</span>
        </div>
      </li>
    </ul>
  </section>
</template>

<style scoped>
.hero {
  padding: var(--space-6);
  max-width: 720px;
}
.hero h1 { font-size: 36px; font-weight: 700; }
.lede {
  font-size: 16px;
  color: var(--fg-muted);
  margin-top: var(--space-2);
  margin-bottom: var(--space-5);
}
.chapter-grid {
  list-style: none;
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(220px, 1fr));
  gap: var(--space-3);
}
.tile {
  border: 1px solid var(--border);
  border-radius: var(--radius-lg);
  background: var(--surface);
  overflow: hidden;
}
.tile-link {
  display: flex;
  flex-direction: column;
  padding: var(--space-4);
  text-decoration: none;
  color: var(--fg);
  height: 100%;
  gap: var(--space-2);
}
.tile-link:hover { background: var(--surface-2); }
.tile-link.disabled { opacity: 0.5; cursor: not-allowed; }
.tile-ord {
  font-size: var(--text-xs);
  color: var(--fg-subtle);
  text-transform: uppercase;
}
.tile-title { font-size: 16px; font-weight: 600; }
.tile-desc {
  font-size: var(--text-sm);
  color: var(--fg-muted);
}
.error { color: var(--danger); }
</style>
