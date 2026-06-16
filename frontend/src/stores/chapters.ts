import { defineStore } from 'pinia'
import { ref } from 'vue'
import { api } from '../api/client'
import type { Chapter, Template, SubmitResult } from '../api/types'

export const useChaptersStore = defineStore('chapters', () => {
  const list = ref<Chapter[]>([])
  const current = ref<Template | null>(null)
  const loading = ref(false)
  const error = ref<string | null>(null)

  async function fetchList() {
    loading.value = true
    error.value = null
    try {
      list.value = await api.chapters.list()
    } catch (e: unknown) {
      error.value = e instanceof Error ? e.message : String(e)
    } finally {
      loading.value = false
    }
  }

  async function fetchTemplate(id: string) {
    loading.value = true
    error.value = null
    try {
      current.value = await api.chapters.template(id)
    } catch (e: unknown) {
      error.value = e instanceof Error ? e.message : String(e)
      current.value = null
    } finally {
      loading.value = false
    }
  }

  async function submit(id: string, userCode: string): Promise<SubmitResult> {
    return api.chapters.submit(id, { userCode })
  }

  async function fetchSolution(id: string): Promise<string> {
    const r = await api.chapters.solution(id)
    return r.code
  }

  async function reset() {
    await api.chapters.reset()
    await fetchList()
  }

  function findInList(id: string): Chapter | undefined {
    return list.value.find((c) => c.id === id)
  }

  // Optimistic UI: flip the local chapter's completed flag so the ✓ badge
  // shows immediately. fetchList() later reconciles if needed. All chapters
  // are unlocked up front, so there's no next-chapter unlocking to do.
  function applyPass(id: string) {
    const idx = list.value.findIndex((c) => c.id === id)
    if (idx < 0) return
    list.value[idx] = { ...list.value[idx], completed: true }
  }

  return { list, current, loading, error, fetchList, fetchTemplate, submit, fetchSolution, reset, findInList, applyPass }
})
