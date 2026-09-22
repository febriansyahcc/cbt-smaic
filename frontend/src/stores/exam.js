import { defineStore } from 'pinia'
import api from '../services/api'
import { localExamStorage } from '../services/storage'
import router from '../router/index.js'

// Referensi handler disimpan di level modul agar bisa di-removeEventListener dengan tepat.
// Pinia store tidak punya lifecycle, jadi cleanup dilakukan dari ExamView.
let _onlineHandler = null
let _offlineHandler = null

export const useExamStore = defineStore('exam', {
  state: () => ({
    sessionId: null,
    scheduleTitle: '',
    subjectName: '',
    durationMinutes: 90,
    serverDeadline: null,
    remainingSeconds: 0,
    maxViolations: 3,
    violationCount: 0,
    isBlocked: false,
    questions: [],
    currentIndex: 0,
    localAnswers: {}, // { [qId]: { selectedOption, isDoubtful } }
    isOnline: navigator.onLine,
    isSyncing: false,
    pendingCount: 0,
    isDrawerOpen: false,
    isSubmitModalOpen: false,
    timerHandle: null,
    syncTimerHandle: null,
    scoreResult: null,
    submittedAt: null,
  }),
  getters: {
    currentQuestion: (state) => state.questions[state.currentIndex] || null,
    totalQuestions: (state) => state.questions.length,
    formattedTimer: (state) => {
      const s = Math.max(0, state.remainingSeconds)
      const hours = Math.floor(s / 3600).toString().padStart(2, '0')
      const minutes = Math.floor((s % 3600) / 60).toString().padStart(2, '0')
      const seconds = (s % 60).toString().padStart(2, '0')
      return `${hours}:${minutes}:${seconds}`
    },
    isLastQuestion: (state) => state.currentIndex === state.questions.length - 1,
    isFirstQuestion: (state) => state.currentIndex === 0,
    answeredCount: (state) => {
      return Object.values(state.localAnswers).filter(
        (a) => (a.selectedOption && a.selectedOption !== '') || (a.answerText && a.answerText.trim() !== '')
      ).length
    },
    doubtfulCount: (state) => {
      return Object.values(state.localAnswers).filter((a) => a.isDoubtful).length
    },
    unansweredCount: (state) => {
      const answered = Object.values(state.localAnswers).filter(
        (a) => (a.selectedOption && a.selectedOption !== '') || (a.answerText && a.answerText.trim() !== '')
      ).length
      return Math.max(0, state.questions.length - answered)
    },
  },
  actions: {
    async startExam(scheduleId, token) {
      const res = await api.post('/student/exams/start', {
        schedule_id: scheduleId,
        token: token,
      })
      const payload = res.data.data
      this.sessionId = payload.session_id
      this.scheduleTitle = payload.schedule_title
      this.subjectName = payload.subject_name
      this.durationMinutes = payload.duration_minutes
      this.serverDeadline = new Date(payload.server_deadline)
      this.remainingSeconds = payload.remaining_seconds
      this.maxViolations = payload.max_violations
      this.violationCount = payload.current_violations || 0
      this.isBlocked = this.violationCount >= this.maxViolations
      this.questions = payload.questions || []
      this.currentIndex = 0

      // Cache payload for offline recoverability
      localExamStorage.cacheExamPayload(this.sessionId, payload)

      // Initialize answers from server saved + local storage
      const serverSaved = payload.saved_answers || {}
      const localSaved = localExamStorage.getAllAnswers(this.sessionId)

      // Merge: local saved takes precedence if newer
      this.localAnswers = { ...serverSaved, ...localSaved }

      this.updatePendingCount()
      this.startCountdown()

      return payload
    },

    startCountdown() {
      if (this.timerHandle) clearInterval(this.timerHandle)
      this.timerHandle = setInterval(() => {
        if (this.remainingSeconds > 0) {
          this.remainingSeconds--
        } else {
          clearInterval(this.timerHandle)
          this.autoSubmitOnTimeout()
        }
      }, 1000)
    },

    listenNetwork() {
      if (_onlineHandler) return // sudah terdaftar, jangan duplikat
      _onlineHandler = () => { this.isOnline = true; this.flushSyncQueue() }
      _offlineHandler = () => { this.isOnline = false }
      window.addEventListener('online', _onlineHandler)
      window.addEventListener('offline', _offlineHandler)
    },

    cleanupNetwork() {
      if (_onlineHandler) window.removeEventListener('online', _onlineHandler)
      if (_offlineHandler) window.removeEventListener('offline', _offlineHandler)
      _onlineHandler = null
      _offlineHandler = null
    },

    selectOption(optionKey) {
      if (!this.currentQuestion) return
      const qId = this.currentQuestion.id
      const existing = this.localAnswers[qId] || { isDoubtful: false, answerText: '' }

      // If clicking same option, allow unselect or keep selected
      const newOption = existing.selectedOption === optionKey ? '' : optionKey

      this.localAnswers[qId] = {
        ...existing,
        selectedOption: newOption,
      }

      // 1. Save immediately to LocalStorage
      localExamStorage.saveAnswer(this.sessionId, qId, newOption, existing.answerText || '', existing.isDoubtful)
      this.updatePendingCount()

      // 2. Schedule debounced sync (300ms)
      this.debounceSync()
    },

    setAnswerText(text) {
      if (!this.currentQuestion) return
      const qId = this.currentQuestion.id
      const existing = this.localAnswers[qId] || { isDoubtful: false, selectedOption: '' }

      this.localAnswers[qId] = {
        ...existing,
        answerText: text,
      }

      localExamStorage.saveAnswer(this.sessionId, qId, existing.selectedOption || '', text, existing.isDoubtful)
      this.updatePendingCount()
      this.debounceSync()
    },

    toggleDoubtful() {
      if (!this.currentQuestion) return
      const qId = this.currentQuestion.id
      const existing = this.localAnswers[qId] || { selectedOption: '', answerText: '' }

      const newDoubtful = !existing.isDoubtful
      this.localAnswers[qId] = {
        ...existing,
        isDoubtful: newDoubtful,
      }

      localExamStorage.saveAnswer(this.sessionId, qId, existing.selectedOption || '', existing.answerText || '', newDoubtful)
      this.updatePendingCount()
      this.debounceSync()
    },

    nextQuestion() {
      if (this.currentIndex < this.questions.length - 1) {
        this.currentIndex++
      }
    },

    prevQuestion() {
      if (this.currentIndex > 0) {
        this.currentIndex--
      }
    },

    goToQuestion(idx) {
      if (idx >= 0 && idx < this.questions.length) {
        this.currentIndex = idx
        this.isDrawerOpen = false
      }
    },

    debounceSync() {
      if (this.syncTimerHandle) clearTimeout(this.syncTimerHandle)
      this.syncTimerHandle = setTimeout(() => {
        this.flushSyncQueue()
      }, 300)
    },

    async flushSyncQueue() {
      if (!this.sessionId || !navigator.onLine) return
      const pendingItems = localExamStorage.getPendingQueue(this.sessionId)
      if (pendingItems.length === 0) {
        this.pendingCount = 0
        return
      }

      this.isSyncing = true
      try {
        await api.post('/student/exams/sync', {
          session_id: this.sessionId,
          answers: pendingItems,
        })
        const syncedIds = pendingItems.map((p) => p.question_id)
        localExamStorage.clearPendingItems(this.sessionId, syncedIds)
      } catch (err) {
        console.warn('Sync failed, will retry later:', err)
      } finally {
        this.isSyncing = false
        this.updatePendingCount()
      }
    },

    updatePendingCount() {
      if (!this.sessionId) return
      const pending = localExamStorage.getPendingQueue(this.sessionId)
      this.pendingCount = pending.length
    },

    async reportViolation(eventType, details) {
      if (this.isBlocked || !this.sessionId) return
      try {
        const res = await api.post('/student/exams/violation', {
          session_id: this.sessionId,
          event_type: eventType,
          details: details || '',
        })
        this.violationCount = res.data.violation_count
        if (res.data.is_blocked) {
          this.isBlocked = true
        }
      } catch (e) {
        this.violationCount++
        if (this.violationCount >= this.maxViolations) {
          this.isBlocked = true
        }
      }
    },

    async submitExam() {
      if (this.timerHandle) clearInterval(this.timerHandle)
      // Flush pending queue first
      await this.flushSyncQueue()

      // Catat waktu tepat sebelum request — mendekati submitted_at di server
      const submittedAt = new Date()

      const res = await api.post('/student/exams/submit', {
        session_id: this.sessionId,
      })

      this.scoreResult = res.data.total_score
      this.submittedAt = submittedAt
      localExamStorage.clearSession(this.sessionId)
      return this.scoreResult
    },

    autoSubmitOnTimeout() {
      this.submitExam()
        .finally(() => {
          router.push('/exam-finished')
        })
    }
  }
})
