<script setup lang="ts">
import { ref, watch, watchEffect, onMounted, onBeforeUnmount, defineAsyncComponent } from 'vue'
import { useChaptersStore } from '../stores/chapters'
import { useLocaleStore } from '../stores/locale'
import type { SubmitResult } from '../api/types'

// Lazy-load Monaco (Phase 6 component) so the chapter list view doesn't
// pay its ~3MB cost on first paint.
const CodeEditor = defineAsyncComponent(() => import('../components/CodeEditor.vue'))

const props = defineProps<{ id: string }>()
const chapters = useChaptersStore()
const locale = useLocaleStore()

const userCode = ref('')
const submitting = ref(false)
const result = ref<SubmitResult | null>(null)

// Recommended-answer drawer state. The solution is fetched lazily the first
// time the user opens the drawer (not on chapter load) to avoid an extra
// round-trip for learners who never look at it. The drawer sits beside the
// editor so the learner can copy directly into their code.
const solutionOpen = ref(false)
const solutionCode = ref('')
const solutionLoading = ref(false)
const solutionError = ref<string | null>(null)
const copied = ref(false)

watchEffect(async () => {
  await chapters.fetchTemplate(props.id)
  if (chapters.current) userCode.value = chapters.current.code
})

// When the language changes, re-fetch the template so the TODO hints refresh
// in the new locale — but DON'T clobber code the user already typed: only
// reset userCode if it still equals the previous (un-edited) template.
watch(() => locale.locale, async () => {
  const oldTemplate = chapters.current?.code
  await chapters.fetchTemplate(props.id)
  const newTemplate = chapters.current?.code
  if (userCode.value === oldTemplate) {
    userCode.value = newTemplate ?? userCode.value
  }
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

async function openSolution() {
  // Avoid refetching if the drawer was already populated for this chapter.
  if (!solutionCode.value && !solutionError.value) {
    solutionLoading.value = true
    solutionError.value = null
    try {
      solutionCode.value = await chapters.fetchSolution(props.id)
    } catch (e: unknown) {
      solutionError.value = e instanceof Error ? e.message : String(e)
    } finally {
      solutionLoading.value = false
    }
  }
  solutionOpen.value = true
}

function closeSolution() {
  solutionOpen.value = false
}

async function copySolution() {
  try {
    await navigator.clipboard.writeText(solutionCode.value)
    copied.value = true
    setTimeout(() => (copied.value = false), 1500)
  } catch {
    // clipboard may be unavailable (e.g. insecure context); ignore silently
  }
}

function onKeydown(e: KeyboardEvent) {
  if (e.key === 'Escape' && solutionOpen.value) closeSolution()
}

onMounted(() => window.addEventListener('keydown', onKeydown))
onBeforeUnmount(() => window.removeEventListener('keydown', onKeydown))

// Reset cached solution when navigating between chapters so we don't show
// chapter N's answer under chapter M's heading.
watchEffect(() => {
  void props.id
  solutionCode.value = ''
  solutionError.value = null
  solutionOpen.value = false
})
</script>

<template>
  <section class="chapter">
    <header class="chapter-header">
      <div class="title-row">
        <div>
          <h1>{{ chapters.findInList(id)?.title ?? id }}</h1>
          <p class="desc">{{ chapters.findInList(id)?.description }}</p>
        </div>
        <button type="button" class="answer-btn" @click="openSolution">
          {{ $t('solution.open') }}
        </button>
      </div>
    </header>

    <div class="main-row">
      <div class="editor-wrap">
        <CodeEditor
          v-model="userCode"
          :todos="chapters.current?.todos ?? []"
          :submitting="submitting"
          @submit="onSubmit"
        />
      </div>

      <aside v-if="solutionOpen" class="answer-panel" role="region" :aria-label="$t('solution.title')">
        <header class="panel-header">
          <div>
            <h2>{{ $t('solution.title') }}</h2>
            <p class="panel-sub">{{ $t('solution.subtitle') }}</p>
          </div>
          <button type="button" class="panel-close" :aria-label="$t('solution.close')" @click="closeSolution">✕</button>
        </header>
        <div class="panel-body">
          <p v-if="solutionLoading" class="panel-empty">{{ $t('solution.loading') }}</p>
          <p v-else-if="solutionError" class="panel-empty error">{{ $t('solution.load_error') }}: {{ solutionError }}</p>
          <pre v-else class="panel-pre">{{ solutionCode }}</pre>
        </div>
        <footer v-if="!solutionLoading && !solutionError" class="panel-footer">
          <button type="button" class="copy-btn" @click="copySolution">
            {{ copied ? $t('solution.copied') : $t('solution.copy') }}
          </button>
        </footer>
      </aside>
    </div>

    <aside class="output" aria-live="polite">
      <h2>{{ $t('output.title') }}</h2>
      <div v-if="!result" class="empty">{{ $t('output.empty') }}</div>
      <template v-else>
        <div :class="['status', result.passed ? 'pass' : 'fail']">
          <span class="status-icon">{{ result.passed ? '✅' : '❌' }}</span>
          <span class="status-text">
            {{ result.passed ? $t('output.passed') : $t('output.failed') }}
          </span>
          <span v-if="result.passed && result.nextChapterUnlocked" class="status-extra">
            · {{ $t('output.unlocked_next') }}
          </span>
          <span v-if="result.durationMs > 0" class="status-extra">
            · {{ $t('output.duration', { ms: result.durationMs }) }}
          </span>
        </div>
        <pre v-if="result.output" :class="['output-pre', { pass: result.passed, fail: !result.passed }]">{{ result.output }}</pre>
      </template>
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
.title-row {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: var(--space-3);
}
.desc {
  color: var(--fg-muted);
  font-size: var(--text-sm);
  margin-top: var(--space-1);
}
.answer-btn {
  flex-shrink: 0;
  background: transparent;
  color: var(--accent);
  border: 1px solid var(--accent);
  border-radius: var(--radius);
  padding: var(--space-1) var(--space-3);
  font-size: var(--text-sm);
  font-weight: 600;
  cursor: pointer;
}
.answer-btn:hover { background: color-mix(in oklch, var(--accent) 12%, transparent); }

/* Editor + answer drawer side by side. The drawer is optional; when closed
   the editor takes the full width. min-width:0 lets the editor shrink;
   min-height:0 lets the whole row stay within its grid cell (the .chapter
   grid's 1fr row) instead of being blown up by Monaco's intrinsic height —
   without it the row grows to fit content and the page scrolls forever. */
.main-row {
  display: flex;
  gap: var(--space-3);
  min-height: 0;
}
.editor-wrap {
  flex: 1;
  min-width: 0;
  min-height: 0;
  border: 1px solid var(--border);
  border-radius: var(--radius-lg);
  overflow: hidden;
  background: var(--surface);
}
.answer-panel {
  flex-shrink: 0;
  width: min(520px, 42vw);
  min-height: 0;
  display: flex;
  flex-direction: column;
  border: 1px solid var(--border);
  border-radius: var(--radius-lg);
  background: var(--surface);
  overflow: hidden;
}
.panel-header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: var(--space-3);
  padding: var(--space-3) var(--space-4);
  border-bottom: 1px solid var(--border);
}
.panel-header h2 { font-size: 16px; font-weight: 700; }
.panel-sub {
  font-size: var(--text-sm);
  color: var(--fg-muted);
  margin-top: var(--space-1);
}
.panel-close {
  flex-shrink: 0;
  background: transparent;
  border: 0;
  color: var(--fg-muted);
  font-size: 16px;
  cursor: pointer;
  padding: var(--space-1);
  line-height: 1;
  border-radius: var(--radius);
}
.panel-close:hover { background: var(--surface-2); color: var(--fg); }
.panel-body {
  flex: 1;
  min-height: 0;
  padding: var(--space-4);
  overflow: auto;
}
.panel-empty { color: var(--fg-subtle); font-size: var(--text-sm); }
.panel-empty.error { color: var(--danger); }
.panel-pre {
  font-family: var(--font-mono);
  font-size: var(--text-sm);
  white-space: pre-wrap;
  color: var(--fg);
  margin: 0;
}
.panel-footer {
  border-top: 1px solid var(--border);
  padding: var(--space-2) var(--space-4);
  display: flex;
  justify-content: flex-end;
}
.copy-btn {
  background: var(--accent);
  color: var(--accent-fg);
  border: 0;
  border-radius: var(--radius);
  padding: var(--space-1) var(--space-3);
  font-size: var(--text-sm);
  font-weight: 600;
  cursor: pointer;
}
.copy-btn:hover { filter: brightness(1.05); }

.output {
  border: 1px solid var(--border);
  border-radius: var(--radius-lg);
  background: var(--surface);
  padding: var(--space-3);
  max-height: 240px;
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
.status {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: var(--space-1);
  font-size: var(--text-sm);
  font-weight: 600;
  margin-bottom: var(--space-2);
}
.status.pass { color: var(--success); }
.status.fail { color: var(--danger); }
.status-icon { font-size: 15px; }
.status-extra {
  font-weight: 400;
  color: var(--fg-muted);
}
.output-pre {
  font-family: var(--font-mono);
  font-size: var(--text-sm);
  white-space: pre-wrap;
  color: var(--fg);
  margin: 0;
}
.output-pre.pass { color: var(--success); }
.output-pre.fail { color: var(--danger); }
</style>
