import { createI18n } from 'vue-i18n'
import { parse } from 'yaml'
import enText from './locales/en.yaml?raw'
import zhText from './locales/zh-CN.yaml?raw'
import type { Locale } from '../stores/locale'

const messages = {
  en: parse(enText),
  'zh-CN': parse(zhText),
} as const

// vue-i18n "legacy: false" enables the Composition API ($t as a function
// via useI18n()). This is the Vue 3 recommended mode.
export const i18n = createI18n({
  legacy: false,
  locale: 'en',
  fallbackLocale: 'en',
  messages,
})

export function setI18nLocale(loc: Locale) {
  i18n.global.locale.value = loc
}

export type { Locale }
