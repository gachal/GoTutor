import { defineStore } from 'pinia'
import { ref } from 'vue'

// localStorage-backed practice count per chapter. UX decoration, not
// crash-safe data — losing it just resets the "练过 N 次" badge. Backend
// persistence would be over-engineering: the existing `progress` table
// already records completion; runs-per-chapter is a vibe metric.
const KEY = 'gotutor.runs'

function load(): Record<string, number> {
  try {
    const raw = localStorage.getItem(KEY)
    return raw ? JSON.parse(raw) as Record<string, number> : {}
  } catch {
    return {}
  }
}

export const useRunsStore = defineStore('runs', () => {
  const counts = ref<Record<string, number>>(load())

  function increment(chapterId: string) {
    counts.value = {
      ...counts.value,
      [chapterId]: (counts.value[chapterId] ?? 0) + 1,
    }
    try {
      localStorage.setItem(KEY, JSON.stringify(counts.value))
    } catch {
      // localStorage full or disabled — in-session counts still work
    }
  }

  function get(chapterId: string): number {
    return counts.value[chapterId] ?? 0
  }

  return { counts, increment, get }
})
