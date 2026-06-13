import { defineStore } from 'pinia'
import { ref, computed, watch } from 'vue'
import { usePreferredDark } from '@vueuse/core'

export type ThemeMode = 'light' | 'dark' | 'system'

const STORAGE_KEY = 'gotutor.theme'

function loadStored(): ThemeMode {
  const v = localStorage.getItem(STORAGE_KEY)
  if (v === 'light' || v === 'dark' || v === 'system') return v
  return 'system'
}

export const useThemeStore = defineStore('theme', () => {
  const mode = ref<ThemeMode>(loadStored())
  const prefersDark = usePreferredDark()

  const isDark = computed(() => {
    if (mode.value === 'system') return prefersDark.value
    return mode.value === 'dark'
  })

  function apply() {
    document.documentElement.classList.toggle('dark', isDark.value)
  }

  function setMode(next: ThemeMode) {
    mode.value = next
    localStorage.setItem(STORAGE_KEY, next)
  }

  watch(isDark, apply, { immediate: true })

  return { mode, isDark, setMode }
})
