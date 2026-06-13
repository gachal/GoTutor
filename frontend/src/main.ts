import { createApp } from 'vue'
import { createPinia } from 'pinia'
import App from './App.vue'
import { router } from './router'
import { i18n, setI18nLocale } from './i18n'
import { useLocaleStore } from './stores/locale'

const app = createApp(App)
const pinia = createPinia()

app.use(pinia)
app.use(router)
app.use(i18n)

// Sync the persisted locale into both vue-i18n and the API client
// before mounting so the first render uses the right language.
const locale = useLocaleStore(pinia)
setI18nLocale(locale.locale)

app.mount('#app')
