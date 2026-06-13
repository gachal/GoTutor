import { defineStore } from 'pinia'
import { ref, watch } from 'vue'
import { setLocale as setApiClientLocale } from '../api/client'

export type Locale = 'zh-CN' | 'en'

const STORAGE_KEY = 'gotutor.locale'
const VALID: Locale[] = ['zh-CN', 'en']

function loadStored(): Locale {
  const v = localStorage.getItem(STORAGE_KEY)
  if (v && VALID.includes(v as Locale)) return v as Locale
  return navigator.language.toLowerCase().startsWith('zh') ? 'zh-CN' : 'en'
}

export const useLocaleStore = defineStore('locale', () => {
  const locale = ref<Locale>(loadStored())

  function set(next: Locale) {
    locale.value = next
    localStorage.setItem(STORAGE_KEY, next)
  }

  watch(locale, (l) => setApiClientLocale(l), { immediate: true })

  return { locale, set }
})
