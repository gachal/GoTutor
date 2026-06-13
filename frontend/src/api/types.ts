// Mirrors backend/internal/api/types.go. Update both when the contract
// changes — Phase 5 has no auto-generation; future CI can enforce this.

export interface Chapter {
  id: string
  title: string
  description: string
  ordinal: number
  completed: boolean
  unlocked: boolean
}

export interface Todo {
  line: number
  hint: string
}

export interface Template {
  code: string
  todos: Todo[]
}

export interface SubmitRequest {
  userCode: string
}

export interface SubmitResult {
  passed: boolean
  output: string
  durationMs: number
  nextChapterUnlocked: boolean
}

export interface HintResponse {
  text: string
}

export interface HealthResponse {
  ok: boolean
  port: number
  goFound: boolean
  goVersion: string
}
