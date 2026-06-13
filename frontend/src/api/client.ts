import axios, { type AxiosInstance } from 'axios'
import type {
  Chapter,
  Template,
  SubmitResult,
  HintResponse,
  SubmitRequest,
  HealthResponse,
} from './types'

// Locale is module-level so the locale store can set it without
// recreating the axios instance. Default 'en'; updated on app boot.
let currentLocale: 'zh-CN' | 'en' = 'en'

export function setLocale(loc: 'zh-CN' | 'en') {
  currentLocale = loc
}

// 30s timeout bounds hung submit requests (verifier caps at 15s; this
// gives ~15s of network/queue slack).
const instance: AxiosInstance = axios.create({
  baseURL: '/api',
  timeout: 30_000,
  headers: { 'Content-Type': 'application/json' },
})

instance.interceptors.request.use((config) => {
  config.headers['Accept-Language'] = currentLocale
  return config
})

export const api = {
  health(): Promise<HealthResponse> {
    return instance.get('/health').then((r) => r.data)
  },

  chapters: {
    list(): Promise<Chapter[]> {
      return instance.get('/chapters').then((r) => r.data)
    },
    template(id: string): Promise<Template> {
      return instance.get(`/chapters/${encodeURIComponent(id)}/template`).then((r) => r.data)
    },
    hint(id: string, line: number): Promise<HintResponse> {
      return instance
        .get(`/chapters/${encodeURIComponent(id)}/hint`, { params: { line } })
        .then((r) => r.data)
    },
    submit(id: string, body: SubmitRequest): Promise<SubmitResult> {
      return instance.post(`/chapters/${encodeURIComponent(id)}/submit`, body).then((r) => r.data)
    },
    reset(): Promise<void> {
      return instance.post('/reset').then(() => undefined)
    },
  },
}
