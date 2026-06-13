<script setup lang="ts">
// Phase 5 STUB — Phase 6 replaces with Monaco integration.
// Provides a functional textarea so the chapter view works end-to-end
// before the polished editor lands.
import { computed } from 'vue'
import type { Todo } from '../api/types'

const props = defineProps<{
  modelValue: string
  todos: Todo[]
  submitting: boolean
}>()

const emit = defineEmits<{
  (e: 'update:modelValue', value: string): void
  (e: 'submit'): void
}>()

const value = computed({
  get: () => props.modelValue,
  set: (v: string) => emit('update:modelValue', v),
})

function onKeydown(e: KeyboardEvent) {
  if ((e.metaKey || e.ctrlKey) && e.key === 'Enter') {
    e.preventDefault()
    emit('submit')
  }
}
</script>

<template>
  <div class="stub-editor">
    <div class="toolbar">
      <button
        type="button"
        class="submit-btn"
        :disabled="submitting"
        @click="emit('submit')"
      >{{ submitting ? '…' : '▶' }}</button>
      <span class="hint">{{ todos.length }} TODOs · ⌘/Ctrl+Enter</span>
    </div>
    <textarea
      v-model="value"
      class="textarea"
      spellcheck="false"
      @keydown="onKeydown"
    />
  </div>
</template>

<style scoped>
.stub-editor {
  display: grid;
  grid-template-rows: auto 1fr;
  height: 100%;
  background: var(--surface);
}
.toolbar {
  display: flex;
  align-items: center;
  gap: var(--space-3);
  padding: var(--space-2) var(--space-3);
  border-bottom: 1px solid var(--border);
  background: var(--surface-2);
}
.submit-btn {
  background: var(--accent);
  color: var(--accent-fg);
  border: 0;
  border-radius: var(--radius);
  padding: var(--space-1) var(--space-3);
  font-size: var(--text-sm);
  font-weight: 600;
}
.submit-btn:disabled { opacity: 0.5; cursor: wait; }
.hint {
  font-size: var(--text-xs);
  color: var(--fg-subtle);
}
.textarea {
  width: 100%;
  height: 100%;
  border: 0;
  outline: 0;
  resize: none;
  padding: var(--space-3);
  font-family: var(--font-mono);
  font-size: var(--text-sm);
  background: var(--bg);
  color: var(--fg);
  line-height: 1.5;
  tab-size: 4;
}
</style>
