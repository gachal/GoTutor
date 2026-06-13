<script setup lang="ts">
// Monaco-based editor with Go syntax highlighting, // TODO glyph decorations,
// theme sync, and submit guard. Lazy-loaded by ChapterView so the chapter
// list view doesn't pay Monaco's ~3MB cost on first paint.
import { computed, onBeforeUnmount, ref, shallowRef, watch } from 'vue'
import { VueMonacoEditor } from '@guolao/vue-monaco-editor'
import type { editor } from 'monaco-editor'
import type * as Monaco from 'monaco-editor'
import editorWorker from 'monaco-editor/esm/vs/editor/editor.worker?worker'
import type { Todo } from '../api/types'
import { useThemeStore } from '../stores/theme'

// Go has no built-in Monaco language worker — only the basic editor
// worker is needed for syntax highlighting + diffing. We deliberately
// skip the json/css/html/ts workers (~8MB combined) since GoTutor never
// edits those languages.
;(self as unknown as { MonacoEnvironment: unknown }).MonacoEnvironment = {
  getWorker() {
    return new editorWorker()
  },
}

const props = defineProps<{
  modelValue: string
  todos: Todo[]
  submitting: boolean
}>()

const emit = defineEmits<{
  (e: 'update:modelValue', value: string): void
  (e: 'submit'): void
}>()

const theme = useThemeStore()
const editorRef = shallowRef<editor.IStandaloneCodeEditor>()
const monacoRef = shallowRef<typeof Monaco>()
const decorations = ref<string[]>([])

const value = computed({
  get: () => props.modelValue,
  set: (v: string) => emit('update:modelValue', v),
})

// Custom themes — kept minimal so the editor reads tokens.css vars through
// the chrome. We define light + dark and switch via the theme store.
function defineThemes(monaco: typeof Monaco) {
  monaco.editor.defineTheme('gotutor-light', {
    base: 'vs',
    inherit: true,
    rules: [{ token: 'comment', foreground: '888888', fontStyle: 'italic' }],
    colors: {},
  })
  monaco.editor.defineTheme('gotutor-dark', {
    base: 'vs-dark',
    inherit: true,
    rules: [{ token: 'comment', foreground: '708090', fontStyle: 'italic' }],
    colors: {},
  })
}

function applyTheme() {
  if (!monacoRef.value) return
  monacoRef.value.editor.setTheme(theme.isDark ? 'gotutor-dark' : 'gotutor-light')
}

// Re-apply TODO decorations whenever the model or todo list changes.
// Decorations are line highlights + a glyph margin marker; hovering the
// marker shows the hint text.
function applyDecorations() {
  const ed = editorRef.value
  const monaco = monacoRef.value
  if (!ed || !monaco) return
  const model = ed.getModel()
  if (!model) return

  const newDecos = props.todos.map((t) => ({
    range: new monaco.Range(t.line, 1, t.line, 1),
    options: {
      isWholeLine: true,
      className: 'todo-line',
      glyphMarginClassName: 'todo-glyph',
      glyphMarginHoverMessage: { value: t.hint || 'TODO' },
      stickiness: monaco.editor.TrackedRangeStickiness.NeverGrowsWhenTypingAtEdges,
    },
  }))
  decorations.value = ed.deltaDecorations(decorations.value, newDecos)
}

function handleMount(ed: editor.IStandaloneCodeEditor, monaco: typeof Monaco) {
  editorRef.value = ed
  monacoRef.value = monaco
  defineThemes(monaco)
  applyTheme()
  applyDecorations()

  // ⌘/Ctrl+Enter to submit.
  ed.addCommand(monaco.KeyMod.CtrlCmd | monaco.KeyCode.Enter, () => emit('submit'))
}

watch(() => props.todos, applyDecorations, { deep: true })
watch(() => props.modelValue, applyDecorations)
watch(() => theme.isDark, applyTheme)

onBeforeUnmount(() => {
  editorRef.value?.dispose()
})
</script>

<template>
  <div class="editor-shell">
    <div class="toolbar">
      <button
        type="button"
        class="submit-btn"
        :disabled="submitting"
        @click="emit('submit')"
      >
        <span v-if="submitting" class="spinner" aria-hidden="true">···</span>
        <template v-else>▶</template>
        <span class="submit-label">{{ submitting ? '' : '' }}</span>
      </button>
      <span class="hint">{{ todos.length }} TODOs · ⌘/Ctrl+Enter</span>
    </div>
    <div class="editor-body">
      <VueMonacoEditor
        :value="value"
        theme="vs"
        language="go"
        :options="{
          minimap: { enabled: false },
          fontSize: 13,
          fontLigatures: true,
          scrollBeyondLastLine: false,
          automaticLayout: true,
          tabSize: 4,
          glyphMargin: true,
          renderLineHighlight: 'all',
        }"
        @mount="handleMount"
        @update:model-value="(v: string | undefined) => emit('update:modelValue', v ?? '')"
      />
    </div>
  </div>
</template>

<style>
/* Global — Monaco manages its own DOM, so these aren't scoped. */
.monaco-editor .todo-line {
  background: color-mix(in oklch, var(--glyph-todo) 18%, transparent);
}
.monaco-editor .todo-glyph {
  background: var(--glyph-todo);
  border-radius: 50%;
  margin-left: 6px;
  width: 8px !important;
  height: 8px !important;
}
.monaco-editor .todo-glyph::before {
  margin: 0;
}
</style>

<style scoped>
.editor-shell {
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
  display: inline-flex;
  align-items: center;
  gap: var(--space-1);
  background: var(--accent);
  color: var(--accent-fg);
  border: 0;
  border-radius: var(--radius);
  padding: var(--space-1) var(--space-3);
  font-size: var(--text-sm);
  font-weight: 600;
  min-width: 60px;
  justify-content: center;
}
.submit-btn:disabled { opacity: 0.6; cursor: wait; }
.spinner { font-family: var(--font-mono); animation: pulse 1s infinite; }
@keyframes pulse { 0%,100% { opacity: 0.4 } 50% { opacity: 1 } }
.hint {
  font-size: var(--text-xs);
  color: var(--fg-subtle);
}
.editor-body {
  height: 100%;
  min-height: 360px;
  text-align: left;
}
</style>
