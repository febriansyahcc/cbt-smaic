const STORAGE_PREFIX = 'cbt_offline_'

export const localExamStorage = {
  saveAnswer(sessionId, questionId, selectedOption, answerText, isDoubtful) {
    const key = `${STORAGE_PREFIX}answers_${sessionId}`
    const queueKey = `${STORAGE_PREFIX}queue_${sessionId}`

    // 1. Update full answers map
    let currentAnswers = {}
    try {
      const raw = localStorage.getItem(key)
      if (raw) currentAnswers = JSON.parse(raw)
    } catch (e) {
      currentAnswers = {}
    }

    currentAnswers[questionId] = {
      selectedOption: selectedOption || '',
      answerText: answerText || '',
      isDoubtful: !!isDoubtful,
      updatedAt: Date.now(),
    }
    localStorage.setItem(key, JSON.stringify(currentAnswers))

    // 2. Add to sync queue
    let queue = {}
    try {
      const qRaw = localStorage.getItem(queueKey)
      if (qRaw) queue = JSON.parse(qRaw)
    } catch (e) {
      queue = {}
    }

    queue[questionId] = {
      question_id: questionId,
      selected_option: selectedOption || '',
      answer_text: answerText || '',
      is_doubtful: !!isDoubtful,
    }
    localStorage.setItem(queueKey, JSON.stringify(queue))
  },

  getAllAnswers(sessionId) {
    const key = `${STORAGE_PREFIX}answers_${sessionId}`
    try {
      const raw = localStorage.getItem(key)
      return raw ? JSON.parse(raw) : {}
    } catch (e) {
      return {}
    }
  },

  getPendingQueue(sessionId) {
    const queueKey = `${STORAGE_PREFIX}queue_${sessionId}`
    try {
      const raw = localStorage.getItem(queueKey)
      if (!raw) return []
      const obj = JSON.parse(raw)
      return Object.values(obj)
    } catch (e) {
      return []
    }
  },

  clearPendingItems(sessionId, questionIds) {
    const queueKey = `${STORAGE_PREFIX}queue_${sessionId}`
    try {
      const raw = localStorage.getItem(queueKey)
      if (!raw) return
      const obj = JSON.parse(raw)
      questionIds.forEach((id) => {
        delete obj[id]
      })
      localStorage.setItem(queueKey, JSON.stringify(obj))
    } catch (e) {
      console.error('Failed to clear pending items', e)
    }
  },

  clearSession(sessionId) {
    localStorage.removeItem(`${STORAGE_PREFIX}answers_${sessionId}`)
    localStorage.removeItem(`${STORAGE_PREFIX}queue_${sessionId}`)
    localStorage.removeItem(`${STORAGE_PREFIX}payload_${sessionId}`)
  },

  cacheExamPayload(sessionId, payload) {
    localStorage.setItem(`${STORAGE_PREFIX}payload_${sessionId}`, JSON.stringify(payload))
  },

  getCachedExamPayload(sessionId) {
    try {
      const raw = localStorage.getItem(`${STORAGE_PREFIX}payload_${sessionId}`)
      return raw ? JSON.parse(raw) : null
    } catch (e) {
      return null
    }
  }
}
