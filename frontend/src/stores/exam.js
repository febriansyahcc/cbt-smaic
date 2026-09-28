import { defineStore } from 'pinia'
import api from '../services/api'
import { localExamStorage } from '../services/storage'
import router from '../router/index.js'

// Referensi handler disimpan di level modul agar bisa di-removeEventListener dengan tepat.
// Pinia store tidak punya lifecycle, jadi cleanup dilakukan dari ExamView.
let _onlineHandler = null
let _offlineHandler = null

const HEARTBEAT_MS = 60 * 1000

export const useExamStore = defineStore('exam', {
  state: () => ({
    sessionId: null,
    scheduleTitle: '',
    subjectName: '',
    durationMinutes: 90,
    serverDeadline: null,
    clockOffsetMs: 0, // jam server - jam perangkat
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
    heartbeatHandle: null,
    syncError: '', // pesan penolakan sync dari server; kosong bila sync terakhir berhasil
    isAutoSubmitting: false,
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
      this.applyServerClock(payload.server_time, payload.server_deadline)
      this.syncError = ''
      this.isAutoSubmitting = false
      this.maxViolations = payload.max_violations
      this.violationCount = payload.current_violations || 0
      // SEMENTARA DINONAKTIFKAN — aktifkan kembali setelah sistem stabil
      this.isBlocked = false
      this.questions = payload.questions || []
      this.currentIndex = 0

      // Cache payload for offline recoverability
      localExamStorage.cacheExamPayload(this.sessionId, payload)

      // Jawaban server memakai snake_case; state lokal memakai camelCase.
      const serverSaved = {}
      for (const [qId, a] of Object.entries(payload.saved_answers || {})) {
        serverSaved[qId] = {
          selectedOption: a.selected_option || '',
          answerText: a.answer_text || '',
          isDoubtful: !!a.is_doubtful,
        }
      }

      // Jawaban lokal hanya menang bila belum tersinkron (masih di antrean). Selebihnya server
      // adalah sumber kebenaran, mis. setelah siswa pindah perangkat atau cache dibersihkan.
      const localSaved = localExamStorage.getAllAnswers(this.sessionId)
      const pendingIds = new Set(localExamStorage.getPendingQueue(this.sessionId).map((p) => p.question_id))
      const merged = { ...serverSaved }
      for (const [qId, a] of Object.entries(localSaved)) {
        if (pendingIds.has(qId) || !merged[qId]) merged[qId] = a
      }
      this.localAnswers = merged

      this.updatePendingCount()
      this.startCountdown()
      this.startHeartbeat()

      return payload
    },

    // Simpan deadline server beserta selisih jam perangkat, lalu hitung ulang sisa waktu.
    applyServerClock(serverTime, serverDeadline) {
      if (!serverDeadline) return
      if (serverTime) this.clockOffsetMs = new Date(serverTime).getTime() - Date.now()
      this.serverDeadline = new Date(serverDeadline)
      this.recomputeRemaining()
    },

    recomputeRemaining() {
      if (!this.serverDeadline) return
      const serverNow = Date.now() + this.clockOffsetMs
      this.remainingSeconds = Math.max(0, Math.ceil((this.serverDeadline.getTime() - serverNow) / 1000))
    },

    // Sisa waktu selalu dihitung dari deadline server, bukan dikurangi per detik: interval yang
    // melambat saat layar HP mati atau tab di latar belakang tidak membuat timer tertinggal.
    startCountdown() {
      if (this.timerHandle) clearInterval(this.timerHandle)
      this.recomputeRemaining()
      this.timerHandle = setInterval(() => {
        this.recomputeRemaining()
        if (this.remainingSeconds <= 0) {
          clearInterval(this.timerHandle)
          this.timerHandle = null
          this.autoSubmitOnTimeout()
        }
      }, 1000)
    },

    // Heartbeat berkala mengambil deadline terbaru (tambahan waktu pengawas) dan mengirim antrean.
    startHeartbeat() {
      this.stopHeartbeat()
      this.heartbeatHandle = setInterval(() => this.flushSyncQueue({ force: true }), HEARTBEAT_MS)
    },

    stopHeartbeat() {
      if (this.heartbeatHandle) clearInterval(this.heartbeatHandle)
      this.heartbeatHandle = null
    },

    listenNetwork() {
      if (_onlineHandler) return // sudah terdaftar, jangan duplikat
      _onlineHandler = () => { this.isOnline = true; this.flushSyncQueue() }
      _offlineHandler = () => { this.isOnline = false }
      window.addEventListener('online', _onlineHandler)
      window.addEventListener('offline', _offlineHandler)
    },

    cleanupNetwork() {
      this.stopHeartbeat()
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

    // force: kirim walau antrean kosong (heartbeat / cek deadline sebelum auto-submit).
    async flushSyncQueue({ force = false } = {}) {
      if (!this.sessionId || !navigator.onLine) return
      const pendingItems = localExamStorage.getPendingQueue(this.sessionId)
      if (pendingItems.length === 0) {
        this.pendingCount = 0
        if (!force) return
      }

      this.isSyncing = true
      try {
        const res = await api.post('/student/exams/sync', {
          session_id: this.sessionId,
          answers: pendingItems,
        })
        // Hanya hapus queue jika server mengonfirmasi minimal satu jawaban tersimpan.
        // Jika synced = 0 dengan item non-kosong, berarti semua item ditolak (bank berubah
        // atau masalah lain) — pertahankan queue agar coba lagi di heartbeat berikutnya.
        if (pendingItems.length === 0 || res.data.synced > 0) {
          const syncedIds = pendingItems.map((p) => p.question_id)
          localExamStorage.clearPendingItems(this.sessionId, syncedIds)
          this.syncError = ''
        } else {
          this.syncError = 'Sinkronisasi jawaban tertunda, menunggu konfirmasi server'
        }
        this.applyServerClock(res.data.server_time, res.data.server_deadline)
      } catch (err) {
        // Tanpa respons = jaringan putus; antrean tetap disimpan dan dicoba lagi. Dengan respons =
        // server menolak (waktu habis, sesi terkunci/dikumpulkan) dan siswa perlu tahu.
        if (err.response) {
          this.syncError = err.response.data?.message || 'Jawaban ditolak server'
        }
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
      // SEMENTARA DINONAKTIFKAN — aktifkan kembali setelah sistem stabil
      return
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
      this.stopHeartbeat()
      // Flush pending queue first
      await this.flushSyncQueue()

      // Kumpulkan sisa pending yang mungkin belum tersinkron (flush bisa gagal karena jaringan
      // atau grace deadline habis). Dikirim bersama request submit agar backend simpan atomik
      // sebelum penilaian — pengaman terakhir agar jawaban tidak hilang.
      const remainingPending = localExamStorage.getPendingQueue(this.sessionId)

      // Catat waktu tepat sebelum request — mendekati submitted_at di server
      const submittedAt = new Date()

      const res = await api.post('/student/exams/submit', {
        session_id: this.sessionId,
        answers: remainingPending,
      })

      this.scoreResult = res.data.total_score
      this.submittedAt = submittedAt
      localExamStorage.clearSession(this.sessionId)
      return this.scoreResult
    },

    // Sebelum mengumpulkan, tanya deadline terbaru ke server: bila pengawas baru menambah waktu,
    // timer dilanjutkan alih-alih mengumpulkan terlalu cepat.
    async autoSubmitOnTimeout() {
      // Sesi terkunci hanya bisa dikumpulkan pengawas; tetap di layar kunci.
      if (this.isAutoSubmitting || this.isBlocked) return
      this.isAutoSubmitting = true
      await this.flushSyncQueue({ force: true })
      if (this.remainingSeconds > 0) {
        this.isAutoSubmitting = false
        this.startCountdown()
        return
      }
      try {
        await this.submitExam()
      } catch (e) {
        console.warn('Auto-submit gagal:', e)
      } finally {
        router.push('/exam-finished')
      }
    }
  }
})
