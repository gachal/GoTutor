import { defineStore } from 'pinia'
import { ref } from 'vue'

// Welcome overlay persistence. localStorage is the right store:
// hasSeenWelcome is a UX flag, not crash-safe data — losing it just means
// the user sees the welcome modal again, which is a delightful surprise
// rather than a data-loss event. No backend involvement needed.
const SEEN_KEY = 'gotutor.hasSeenWelcome'

export const useOnboardingStore = defineStore('onboarding', () => {
  const hasSeen = ref(localStorage.getItem(SEEN_KEY) === '1')
  const showWelcome = ref(!hasSeen.value)

  function dismiss() {
    hasSeen.value = true
    localStorage.setItem(SEEN_KEY, '1')
    showWelcome.value = false
  }

  function reopen() {
    showWelcome.value = true
  }

  return { showWelcome, dismiss, reopen }
})
