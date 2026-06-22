// Mirrors backend/internal/api/types.go. Update both when the contract
// changes — Phase 5 has no auto-generation; future CI can enforce this.

export type Track = 'fundamentals' | 'concurrency' | 'gateway'
export type Difficulty = 'beginner' | 'intermediate' | 'advanced'

export interface Chapter {
  id: string
  title: string
  description: string
  ordinal: number
  track: Track
  difficulty: Difficulty
  estimatedMinutes: number
  prerequisites: string[]
  completed: boolean
  unlocked: boolean
}

export interface TrackProgress {
  track: Track
  totalChapters: number
  completedChapters: number
}

export interface ProgressResponse {
  totalChapters: number
  completedChapters: number
  percent: number
  lastChapterId: string
  byTrack: TrackProgress[]
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

export interface SolutionResponse {
  code: string
}

export interface HealthResponse {
  ok: boolean
  port: number
  goFound: boolean
  goVersion: string
}
