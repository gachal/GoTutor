<script setup lang="ts">
import { useHealthStore } from '../stores/health'
import GoMissingView from '../views/GoMissingView.vue'

const health = useHealthStore()
</script>

<template>
  <section class="boot-gate" aria-live="polite">
    <!-- Loading: first probe in flight. Branded spinner; no jargon. -->
    <div v-if="health.status === 'loading'" class="state loading">
      <div class="spinner" aria-hidden="true" />
      <p class="headline">{{ $t('boot.loading') }}</p>
    </div>

    <!-- Go toolchain missing: full install instructions + recheck. -->
    <GoMissingView v-else-if="health.status === 'goMissing'" />

    <!-- Backend not responding: short explanation + retry. -->
    <div v-else-if="health.status === 'backendDown'" class="state error">
      <h2 class="headline">{{ $t('backend_down.title') }}</h2>
      <p class="body">{{ $t('backend_down.body') }}</p>
      <button type="button" class="primary-btn" @click="health.check()">
        {{ $t('backend_down.retry') }}
      </button>
    </div>
  </section>
</template>

<style scoped>
.boot-gate {
  height: 100%;
  display: grid;
  place-items: center;
  padding: var(--space-6);
}
.state {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: var(--space-3);
  max-width: 520px;
  text-align: center;
}
.headline {
  font-size: 22px;
  font-weight: 700;
  color: var(--fg);
  margin: 0;
}
.body {
  font-size: var(--text-sm);
  color: var(--fg-muted);
  margin: 0;
  line-height: 1.6;
}
.loading .spinner {
  width: 28px;
  height: 28px;
  border-radius: 50%;
  border: 2.5px solid var(--surface-2);
  border-top-color: var(--accent);
  animation: spin 0.8s linear infinite;
  margin-bottom: var(--space-2);
}
@keyframes spin {
  to { transform: rotate(360deg); }
}
.primary-btn {
  background: var(--accent);
  color: var(--accent-fg);
  border: 0;
  border-radius: var(--radius);
  padding: var(--space-2) var(--space-4);
  font-size: var(--text-sm);
  font-weight: 600;
  cursor: pointer;
}
.primary-btn:hover { filter: brightness(1.05); }
</style>
