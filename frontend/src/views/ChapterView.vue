<script setup lang="ts">
import { ref, watchEffect, defineAsyncComponent } from 'vue'
import { useChaptersStore } from '../stores/chapters'
import type { SubmitResult } from '../api/types'

// Lazy-load Monaco (Phase 6 component) so the chapter list view doesn't
// pay its ~3MB cost on first paint.
const CodeEditor = defineAsyncComponent(() => import('../components/CodeEditor.vue'))

const props = defineProps<{ id: string }>()
const chapters = useChaptersStore()

const userCode = ref('')
const submitting = ref(false)
const result = ref<SubmitResult | null>(null)

watchEffect(async () => {
  await chapters.fetchTemplate(props.id)
  if (chapters.current) userCode.value = chapters.current.code
})

async function onSubmit() {
  if (submitting.value) return
  submitting.value = true
  result.value = null
  try {
    const r = await chapters.submit(props.id, userCode.value)
    result.value = r
    if (r.passed) chapters.applyPass(props.id)
  } catch (e: unknown) {
    result.value = {
      passed: false,
      output: e instanceof Error ? e.message : String(e),
      durationMs: 0,
      nextChapterUnlocked: false,
    }
  } finally {
    submitting.value = false
  }
}
</script>

<template>
  <section class="chapter">
    <header class="chapter-header">
      <h1>{{ chapters.findInList(id)?.title ?? id }}</h1>
      <p class="desc">{{ chapters.findInList(id)?.description }}</p>
    </header>

    <div class="editor-wrap">
      <CodeEditor
        v-model="userCode"
        :todos="chapters.current?.todos ?? []"
        :submitting="submitting"
        @submit="onSubmit"
      />
    </div>

    <aside class="output" aria-live="polite">
      <h2>{{ $t('output.title') }}</h2>
      <div v-if="!result" class="empty">{{ $t('output.empty') }}</div>
      <pre v-else :class="['output-pre', { pass: result.passed, fail: !result.passed }]">{{ result.output }}</pre>
    </aside>
  </section>
</template>

<style scoped>
.chapter {
  display: grid;
  grid-template-rows: auto 1fr auto;
  height: 100%;
  padding: var(--space-4);
  gap: var(--space-3);
}
.chapter-header h1 { font-size: 20px; font-weight: 700; }
.desc {
  color: var(--fg-muted);
  font-size: var(--text-sm);
  margin-top: var(--space-1);
}
.editor-wrap {
  min-height: 360px;
  border: 1px solid var(--border);
  border-radius: var(--radius-lg);
  overflow: hidden;
  background: var(--surface);
}
.output {
  border: 1px solid var(--border);
  border-radius: var(--radius-lg);
  background: var(--surface);
  padding: var(--space-3);
  max-height: 200px;
  overflow: auto;
}
.output h2 {
  font-size: var(--text-xs);
  text-transform: uppercase;
  letter-spacing: 0.06em;
  color: var(--fg-subtle);
  margin-bottom: var(--space-2);
}
.empty { color: var(--fg-subtle); font-size: var(--text-sm); }
.output-pre {
  font-family: var(--font-mono);
  font-size: var(--text-sm);
  white-space: pre-wrap;
  color: var(--fg);
}
.output-pre.pass { color: var(--success); }
.output-pre.fail { color: var(--danger); }
</style>
