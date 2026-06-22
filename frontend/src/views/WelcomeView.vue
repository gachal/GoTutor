<script setup lang="ts">
import { useRouter } from 'vue-router'
import { useOnboardingStore } from '../stores/onboarding'

const onboarding = useOnboardingStore()
const router = useRouter()

// CTA: dismiss + jump straight to the first chapter. Skipping keeps
// the user on the chapter list — the overlay just gets out of the way.
function start() {
  onboarding.dismiss()
  router.push('/chapter/calc')
}
function skip() {
  onboarding.dismiss()
}
</script>

<template>
  <div v-if="onboarding.showWelcome" class="welcome-overlay" role="dialog" aria-modal="true">
    <div class="welcome-modal">
      <header class="modal-header">
        <h2>{{ $t('welcome.title') }}</h2>
        <p class="subtitle">{{ $t('welcome.subtitle') }}</p>
      </header>

      <p class="body">{{ $t('welcome.body') }}</p>

      <ol class="steps">
        <li>{{ $t('welcome.step_1') }}</li>
        <li>{{ $t('welcome.step_2') }}</li>
        <li>{{ $t('welcome.step_3') }}</li>
      </ol>

      <footer class="actions">
        <button type="button" class="primary" @click="start">{{ $t('welcome.cta') }}</button>
        <button type="button" class="secondary" @click="skip">{{ $t('welcome.skip') }}</button>
      </footer>
    </div>
  </div>
</template>

<style scoped>
.welcome-overlay {
  position: fixed;
  inset: 0;
  background: color-mix(in oklch, var(--bg) 75%, transparent);
  backdrop-filter: blur(4px);
  display: grid;
  place-items: center;
  z-index: 100;
  padding: var(--space-5);
}
.welcome-modal {
  background: var(--surface);
  border: 1px solid var(--border);
  border-radius: var(--radius-lg);
  max-width: 560px;
  width: 100%;
  padding: var(--space-6);
  display: flex;
  flex-direction: column;
  gap: var(--space-4);
  box-shadow: 0 12px 40px rgba(0, 0, 0, 0.25);
}
.modal-header h2 {
  font-size: 24px;
  font-weight: 700;
  color: var(--fg);
  margin: 0;
}
.subtitle {
  font-size: var(--text-sm);
  color: var(--accent);
  margin-top: var(--space-1);
  font-weight: 600;
}
.body {
  font-size: var(--text-sm);
  color: var(--fg);
  line-height: 1.7;
  white-space: pre-line;
  margin: 0;
}
.steps {
  display: flex;
  flex-direction: column;
  gap: var(--space-2);
  padding-left: var(--space-5);
  margin: 0;
}
.steps li {
  font-size: var(--text-sm);
  color: var(--fg-muted);
  line-height: 1.6;
}
.actions {
  display: flex;
  gap: var(--space-2);
  justify-content: flex-end;
  margin-top: var(--space-2);
}
.primary {
  background: var(--accent);
  color: var(--accent-fg);
  border: 0;
  border-radius: var(--radius);
  padding: var(--space-2) var(--space-4);
  font-size: var(--text-sm);
  font-weight: 600;
  cursor: pointer;
}
.primary:hover { filter: brightness(1.05); }
.secondary {
  background: transparent;
  color: var(--fg-muted);
  border: 1px solid var(--border);
  border-radius: var(--radius);
  padding: var(--space-2) var(--space-4);
  font-size: var(--text-sm);
  cursor: pointer;
}
.secondary:hover { background: var(--surface-2); }
</style>
