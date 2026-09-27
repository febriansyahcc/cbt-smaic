import { ref, computed, watch } from 'vue'
import api from '../../../../services/api'

// Jadwal & token ujian: filter, aksi massal, form jadwal, pengawas, tautan bank soal, koreksi esai, dan dokumen cetak.
export function useSchedules(ctx) {
  const {
    authStore,
    canManageSchedules,
    classSubjects,
    classes,
    currentEventSchedules,
    events,
    loadSchedules,
    readinessData,
    schedules,
    selectedExamEvent,
    showAlertModal,
    showConfirmModal,
    showToast,
    students,
    subjects,
  } = ctx

  // Didefinisikan composable yang dipanggil kemudian; diteruskan saat dipanggil.
  const loadAllData = (...args) => ctx.loadAllData(...args)

  const copyTokenToClipboard = (token) => {
    if (!token) return
    if (navigator.clipboard && navigator.clipboard.writeText) {
      navigator.clipboard.writeText(token).then(() => {
        showToast(`Token ${token} berhasil disalin ke clipboard!`, 'success')
      }).catch(() => {
        showToast(`Token: ${token}`, 'info')
      })
    } else {
      showToast(`Token: ${token}`, 'info')
    }
  }

  // Live Status Helper
  const getScheduleLiveStatus = (sch) => {
    if (!sch || !sch.start_time || !sch.end_time) return 'upcoming'
    const now = Date.now()
    const start = new Date(sch.start_time).getTime()
    const end = new Date(sch.end_time).getTime()
    if (now >= start && now <= end) return 'live'
    if (now < start) return 'upcoming'
    return 'finished'
  }

  // Schedule Filtering, Sorting & Detail States
  const scheduleSearchQuery = ref('')
  const selectedScheduleGrade = ref('all')
  const selectedScheduleTimeFilter = ref('all')
  const scheduleSortKey = ref('time')
  const scheduleSortOrder = ref('asc')
  const showScheduleDetailModal = ref(false)
  const selectedScheduleDetail = ref(null)

  // Essay Correction State (for Schedule Detail Modal)
  const scheduleDetailTab = ref('info')
  const essayQuestions = ref([])
  const essayLoading = ref(false)
  const essayDraft = ref({})
  const essaySaving = ref({})

  // Bulk Selection States & Logic
  const selectedScheduleIds = ref([])

  const toggleScheduleSort = (key) => {
    if (scheduleSortKey.value === key) {
      scheduleSortOrder.value = scheduleSortOrder.value === 'asc' ? 'desc' : 'asc'
    } else {
      scheduleSortKey.value = key
      scheduleSortOrder.value = 'asc'
    }
  }

  const openScheduleDetail = (sch) => {
    selectedScheduleDetail.value = sch
    showScheduleDetailModal.value = true
    scheduleDetailTab.value = 'info'
    essayQuestions.value = []
    essayDraft.value = {}
    essaySaving.value = {}
    makeupStudentsList.value = []
  }

  const fetchEssayAnswers = async (scheduleId) => {
    if (essayLoading.value) return
    essayLoading.value = true
    try {
      const res = await api.get(`/api/v1/admin/schedules/${scheduleId}/essay-answers`)
      essayQuestions.value = res.data.data || []
      // Initialize draft from existing data
      const draft = {}
      for (const q of essayQuestions.value) {
        for (const a of q.answers || []) {
          draft[a.answer_id] = {
            score_awarded: a.score_awarded !== null && a.score_awarded !== undefined ? a.score_awarded : null,
            teacher_comment: a.teacher_comment || ''
          }
        }
      }
      essayDraft.value = draft
    } catch (e) {
      showToast('Gagal memuat data jawaban essay.', 'error')
    } finally {
      essayLoading.value = false
    }
  }

  // Koreksi essay hanya untuk administrator dan guru pengampu kelas + mapel jadwal (sama dengan aturan server).
  const canGradeEssay = (sch) => {
    const u = authStore.user
    if (u?.role === 'ADMIN' || (u?.permissions || []).includes('*')) return true
    return Array.isArray(sch?.relations) && sch.relations.includes('mengampu')
  }

  const switchToEssayTab = () => {
    scheduleDetailTab.value = 'essay'
    if (essayQuestions.value.length === 0 && !essayLoading.value && selectedScheduleDetail.value) {
      fetchEssayAnswers(selectedScheduleDetail.value.id)
    }
  }

  const essayTotalPending = computed(() => {
    return essayQuestions.value.reduce((sum, q) => sum + (q.total_count - q.graded_count), 0)
  })

  const saveEssayQuestion = async (question) => {
    const qid = question.question_id
    // Collect answers for this question
    const payload = (question.answers || []).map(a => {
      const d = essayDraft.value[a.answer_id] || {}
      return {
        answer_id: a.answer_id,
        score_awarded: d.score_awarded !== null && d.score_awarded !== undefined ? Number(d.score_awarded) : null,
        teacher_comment: d.teacher_comment || ''
      }
    })

    // Validate scores
    for (const p of payload) {
      if (p.score_awarded !== null) {
        if (p.score_awarded < 0) {
          showToast(`Nilai tidak boleh negatif (Soal No. ${question.question_number}).`, 'error')
          return
        }
        if (p.score_awarded > question.score_weight) {
          showToast(`Nilai melebihi bobot soal ${question.score_weight} (Soal No. ${question.question_number}).`, 'error')
          return
        }
      }
    }

    essaySaving.value = { ...essaySaving.value, [qid]: true }
    try {
      await api.patch(`/api/v1/admin/schedules/${selectedScheduleDetail.value.id}/essay-answers`, payload)
      showToast(`Penilaian soal No. ${question.question_number} berhasil disimpan.`, 'success')
      // Refresh essay data
      await fetchEssayAnswers(selectedScheduleDetail.value.id)
    } catch (e) {
      showToast('Gagal menyimpan penilaian.', 'error')
    } finally {
      essaySaving.value = { ...essaySaving.value, [qid]: false }
    }
  }

  const availableScheduleGrades = computed(() => {
    const grades = classes.value.map(c => c.grade).filter(Boolean)
    const unique = [...new Set(grades)]
    const order = ['X', 'XI', 'XII', '7', '8', '9', 'VII', 'VIII', 'IX', '10', '11', '12']
    return unique.sort((a, b) => {
      const idxA = order.indexOf(a)
      const idxB = order.indexOf(b)
      if (idxA !== -1 && idxB !== -1) return idxA - idxB
      return a.localeCompare(b)
    })
  })

  const filteredSchedules = computed(() => {
    let list = currentEventSchedules.value

    // Filter by Grade
    if (selectedScheduleGrade.value !== 'all') {
      list = list.filter(sch => {
        const g = sch.class_room?.grade || (classes.value.find(c => c.id === (sch.class_room_id || sch.class_id))?.grade)
        return g === selectedScheduleGrade.value
      })
    }

    // Filter by Time (Hari Ini / Mendatang / Semua)
    if (selectedScheduleTimeFilter.value === 'today') {
      const todayStr = new Date().toDateString()
      list = list.filter(sch => {
        if (!sch.start_time) return false
        return new Date(sch.start_time).toDateString() === todayStr
      })
    } else if (selectedScheduleTimeFilter.value === 'upcoming') {
      const now = Date.now()
      list = list.filter(sch => {
        if (!sch.start_time) return false
        return new Date(sch.start_time).getTime() > now
      })
    }

    // Filter by Search Query
    if (scheduleSearchQuery.value.trim()) {
      const q = scheduleSearchQuery.value.toLowerCase().trim()
      list = list.filter(sch => {
        const titleMatch = sch.title && sch.title.toLowerCase().includes(q)
        const subMatch = (sch.subject?.name && sch.subject.name.toLowerCase().includes(q)) ||
                         (sch.bank?.subject?.name && sch.bank.subject.name.toLowerCase().includes(q))
        const classMatch = sch.class_room?.name && sch.class_room.name.toLowerCase().includes(q)
        const tokenMatch = sch.exam_token && sch.exam_token.toLowerCase().includes(q)
        return titleMatch || subMatch || classMatch || tokenMatch
      })
    }

    // Sort
    if (scheduleSortKey.value) {
      list = [...list].sort((a, b) => {
        let result = 0
        if (scheduleSortKey.value === 'title') {
          const valA = (a.title || '').trim().toLowerCase()
          const valB = (b.title || '').trim().toLowerCase()
          result = valA.localeCompare(valB, undefined, { numeric: true, sensitivity: 'base' })
        } else if (scheduleSortKey.value === 'time') {
          const timeA = a.start_time ? new Date(a.start_time).getTime() : 0
          const timeB = b.start_time ? new Date(b.start_time).getTime() : 0
          result = timeA - timeB
        } else if (scheduleSortKey.value === 'class') {
          const gradeA = a.class_room?.grade || (classes.value.find(c => c.id === (a.class_room_id || a.class_id))?.grade) || ''
          const gradeB = b.class_room?.grade || (classes.value.find(c => c.id === (b.class_room_id || b.class_id))?.grade) || ''
          const nameA = (a.class_room?.name || '').trim().toLowerCase()
          const nameB = (b.class_room?.name || '').trim().toLowerCase()
          if (gradeA !== gradeB) {
            const order = ['X', 'XI', 'XII', '7', '8', '9', 'VII', 'VIII', 'IX', '10', '11', '12']
            const idxA = order.indexOf(gradeA)
            const idxB = order.indexOf(gradeB)
            if (idxA !== -1 && idxB !== -1) {
              result = idxA - idxB
            } else {
              result = gradeA.localeCompare(gradeB)
            }
          }
          if (result === 0) {
            result = nameA.localeCompare(nameB, undefined, { numeric: true, sensitivity: 'base' })
          }
        } else if (scheduleSortKey.value === 'subject') {
          const nameA = (a.subject?.name || a.bank?.subject?.name || '').trim().toLowerCase()
          const nameB = (b.subject?.name || b.bank?.subject?.name || '').trim().toLowerCase()
          result = nameA.localeCompare(nameB, undefined, { numeric: true, sensitivity: 'base' })
        } else if (scheduleSortKey.value === 'status') {
          const actA = a.is_active ? 1 : 0
          const actB = b.is_active ? 1 : 0
          if (actA !== actB) {
            result = actA - actB
          } else {
            const tokA = (a.exam_token || '').trim().toLowerCase()
            const tokB = (b.exam_token || '').trim().toLowerCase()
            result = tokA.localeCompare(tokB)
          }
        }
        return scheduleSortOrder.value === 'asc' ? result : -result
      })
    }

    return list
  })

  // Schedule Pagination States & Logic
  const scheduleCurrentPage = ref(1)
  const schedulePerPage = ref(10)

  const totalSchedulePages = computed(() => {
    return Math.max(1, Math.ceil(filteredSchedules.value.length / schedulePerPage.value))
  })

  const paginatedSchedules = computed(() => {
    const start = (scheduleCurrentPage.value - 1) * schedulePerPage.value
    return filteredSchedules.value.slice(start, start + schedulePerPage.value)
  })

  const displayedSchedulePages = computed(() => {
    const total = totalSchedulePages.value
    const current = scheduleCurrentPage.value
    if (total <= 7) {
      return Array.from({ length: total }, (_, i) => i + 1)
    }
    if (current <= 4) {
      return [1, 2, 3, 4, 5, '...', total]
    } else if (current >= total - 3) {
      return [1, '...', total - 4, total - 3, total - 2, total - 1, total]
    } else {
      return [1, '...', current - 1, current, current + 1, '...', total]
    }
  })

  const setSchedulePage = (p) => {
    if (p === '...' || p < 1 || p > totalSchedulePages.value) return
    scheduleCurrentPage.value = p
  }

  const isAllPaginatedSelected = computed(() => {
    if (!paginatedSchedules.value.length) return false
    return paginatedSchedules.value.every(s => selectedScheduleIds.value.includes(s.id))
  })

  const toggleSelectAllPaginatedSchedules = () => {
    if (!canManageSchedules.value) return
    const pageIds = paginatedSchedules.value.map(s => s.id)
    if (isAllPaginatedSelected.value) {
      selectedScheduleIds.value = selectedScheduleIds.value.filter(id => !pageIds.includes(id))
    } else {
      selectedScheduleIds.value = [...new Set([...selectedScheduleIds.value, ...pageIds])]
    }
  }

  // Bulk Actions
  const bulkActivateSchedules = async () => {
    if (!canManageSchedules.value || !selectedScheduleIds.value.length) return
    const toActivate = schedules.value.filter(s => selectedScheduleIds.value.includes(s.id))
    const invalid = toActivate.filter(s => !s.bank_id || (s.bank && !s.bank.is_locked))
    if (invalid.length > 0) {
      await showAlertModal({
        title: 'Validasi Bank Soal',
        message: `Ada ${invalid.length} sesi jadwal terpilih yang belum memiliki Bank Soal yang siap (terkunci).\n\nSilakan pastikan semua jadwal yang dipilih telah ditautkan dengan bank soal yang berstatus Siap/Terkunci.`
      })
      return
    }
    const confirmed = await showConfirmModal({
      title: 'Aktivasi Massal',
      message: `Aktifkan ${selectedScheduleIds.value.length} sesi jadwal ujian terpilih?`,
      confirmText: 'Aktifkan',
      cancelText: 'Batal'
    })
    if (!confirmed) return

    try {
      for (const sch of toActivate) {
        if (!sch.is_active) {
          const res = await api.post(`/admin/schedules/${sch.id}/toggle`)
          sch.is_active = res.data.is_active
        }
      }
      showToast(`${selectedScheduleIds.value.length} sesi jadwal berhasil diaktifkan!`, 'success')
    } catch (e) {
      showToast('Terjadi kesalahan saat mengaktifkan jadwal secara massal', 'error')
    }
  }

  const bulkDeactivateSchedules = async () => {
    if (!canManageSchedules.value || !selectedScheduleIds.value.length) return
    const confirmed = await showConfirmModal({
      title: 'Nonaktifkan Massal',
      message: `Nonaktifkan ${selectedScheduleIds.value.length} sesi jadwal ujian terpilih?`,
      confirmText: 'Nonaktifkan',
      cancelText: 'Batal'
    })
    if (!confirmed) return

    try {
      const toDeactivate = schedules.value.filter(s => selectedScheduleIds.value.includes(s.id))
      for (const sch of toDeactivate) {
        if (sch.is_active) {
          const res = await api.post(`/admin/schedules/${sch.id}/toggle`)
          sch.is_active = res.data.is_active
        }
      }
      showToast(`${selectedScheduleIds.value.length} sesi jadwal dinonaktifkan!`, 'success')
    } catch (e) {
      showToast('Terjadi kesalahan saat menonaktifkan jadwal secara massal', 'error')
    }
  }

  const bulkRegenerateTokens = async () => {
    if (!canManageSchedules.value || !selectedScheduleIds.value.length) return
    const confirmed = await showConfirmModal({
      title: 'Samakan Token',
      message: `${selectedScheduleIds.value.length} jadwal terpilih akan mendapat satu token baru yang sama. Jadwal lain dengan tanggal dan jam mulai yang sama ikut memakai token ini.`,
      confirmText: 'Buat Token',
      cancelText: 'Batal'
    })
    if (!confirmed) return

    try {
      const res = await api.post('/admin/schedules/regenerate-tokens', { schedule_ids: selectedScheduleIds.value })
      await loadSchedules()
      showToast(res.data?.message || 'Token baru berhasil dibuat', 'success')
    } catch (e) {
      showToast(e.response?.data?.message || 'Gagal mengacak token massal', 'error')
    }
  }

  // Print Document States & Methods
  const showPrintModal = ref(false)
  const printDocType = ref('attendance')
  const printSchedule = ref(null)

  const openPrintModal = (sch, docType = 'attendance') => {
    printSchedule.value = sch
    printDocType.value = docType
    showPrintModal.value = true
  }

  const printStudentsList = computed(() => {
    if (!printSchedule.value) return []
    const targetClassId = printSchedule.value.class_room_id || printSchedule.value.class_id
    return students.value
      .filter(s => (s.class_room_id === targetClassId || s.class_id === targetClassId))
      .sort((a, b) => (a.user?.full_name || a.name || '').localeCompare(b.user?.full_name || b.name || ''))
  })

  const triggerPrintDocument = () => {
    window.print()
  }

  const formatScheduleDateFull = (dt) => {
    if (!dt) return '-'
    return new Date(dt).toLocaleDateString('id-ID', {
      weekday: 'long',
      day: 'numeric',
      month: 'long',
      year: 'numeric'
    })
  }

  const formatScheduleTimeOnly = (dt) => {
    if (!dt) return '-'
    return new Date(dt).toLocaleTimeString('id-ID', {
      hour: '2-digit',
      minute: '2-digit'
    })
  }

  const getDayName = (dt) => {
    if (!dt) return '..................'
    return new Date(dt).toLocaleDateString('id-ID', { weekday: 'long' })
  }

  watch([scheduleSearchQuery, selectedScheduleGrade, selectedScheduleTimeFilter, schedulePerPage, scheduleSortKey, scheduleSortOrder], () => {
    scheduleCurrentPage.value = 1
  })

  watch(totalSchedulePages, (newTotal) => {
    if (scheduleCurrentPage.value > newTotal) {
      scheduleCurrentPage.value = Math.max(1, newTotal)
    }
  })

  watch(
    () => scheduleForm.value.is_makeup,
    (val) => {
      if (val) {
        loadStudentsForClass(scheduleForm.value.class_id)
      } else {
        allStudentsForClass.value = []
      }
    }
  )

  // Modals
  const showScheduleModal = ref(false)

  // Schedule Form & Link Bank States
  const isEditSchedule = ref(false)
  const scheduleForm = ref({
    id: null,
    event_id: '',
    title: '',
    subject_id: '',
    bank_id: '',
    class_id: '',
    exam_date: '',
    start_time: '07:30',
    end_time: '09:00',
    exam_token: 'CBT2026',
    duration_minutes: 90,
    max_violations: 3,
    randomize_questions: true,
    randomize_options: true,
    is_active: true,
    proctors: [],
    is_makeup: false,
    makeup_students: [],
  })

  // Penugasan pengawas: modal dari tabel jadwal dan pemilih di dalam form jadwal
  const showProctorAssignModal = ref(false)
  const scheduleForProctors = ref(null)
  const showFormProctorPicker = ref(false)
  // Daftar id pengawas saat form dibuka, dipakai untuk mendeteksi perubahan saat menyimpan
  const scheduleFormInitialProctorIds = ref([])

  // Ujian susulan: pemilih siswa di form dan daftar siswa di detail modal
  const showMakeupStudentPicker = ref(false)
  const allStudentsForClass = ref([])
  const loadingMakeupStudents = ref(false)
  const makeupStudentsList = ref([])
  const loadingMakeupStudentsList = ref(false)

  const selectedMakeupStudentIDs = computed(() =>
    (scheduleForm.value.makeup_students || []).map(s => s.student_id)
  )

  const loadStudentsForClass = async (classId) => {
    if (!classId) { allStudentsForClass.value = []; return }
    loadingMakeupStudents.value = true
    try {
      const res = await api.get('/admin/students', { params: { class_id: classId } })
      allStudentsForClass.value = res.data?.data || []
    } catch (e) {
      allStudentsForClass.value = []
    } finally {
      loadingMakeupStudents.value = false
    }
  }

  const loadMakeupStudents = async (scheduleId) => {
    if (!scheduleId) { makeupStudentsList.value = []; return }
    loadingMakeupStudentsList.value = true
    try {
      const res = await api.get(`/admin/schedules/${scheduleId}/makeup-students`)
      makeupStudentsList.value = res.data?.data || []
    } catch (e) {
      makeupStudentsList.value = []
    } finally {
      loadingMakeupStudentsList.value = false
    }
  }

  const toggleMakeupStudent = (student) => {
    if (!scheduleForm.value.makeup_students) scheduleForm.value.makeup_students = []
    const idx = scheduleForm.value.makeup_students.findIndex(s => s.student_id === student.id)
    if (idx >= 0) {
      scheduleForm.value.makeup_students.splice(idx, 1)
    } else {
      scheduleForm.value.makeup_students.push({
        student_id: student.id,
        full_name: student.user?.full_name || student.full_name,
        nis: student.nis,
        class_name: student.class_room?.name || ''
      })
    }
  }

  const isStudentSelected = (studentId) => {
    return (scheduleForm.value.makeup_students || []).some(s => s.student_id === studentId)
  }

  const openProctorAssign = (sch) => {
    scheduleForProctors.value = sch
    showProctorAssignModal.value = true
  }

  const formatProctorNames = (list) => {
    const names = (list || []).map(p => p.full_name).filter(Boolean)
    if (names.length <= 2) return names.join(', ')
    return `${names.slice(0, 2).join(', ')} +${names.length - 2}`
  }

  const applyFormProctors = (users) => {
    scheduleForm.value.proctors = users || []
  }

  const removeFormProctor = (id) => {
    scheduleForm.value.proctors = scheduleForm.value.proctors.filter(p => p.id !== id)
  }

  const sameIdSet = (a, b) => {
    if (a.length !== b.length) return false
    const setB = new Set(b)
    return a.every(id => setB.has(id))
  }

  const showLinkBankModal = ref(false)
  const selectedScheduleForLink = ref(null)
  const selectedBankForLink = ref('')

  const availableBanksForScheduleForm = computed(() => {
    const allBanks = readinessData.value.question_banks || []
    if (!scheduleForm.value.subject_id) return allBanks
    const matching = allBanks.filter(b => b.subject_id === scheduleForm.value.subject_id)
    return matching.length > 0 ? matching : allBanks
  })

  const availableBanksForSelectedSchedule = computed(() => {
    const allBanks = readinessData.value.question_banks || []
    if (!selectedScheduleForLink.value) return allBanks
    const targetSubId = selectedScheduleForLink.value.subject_id || selectedScheduleForLink.value.bank?.subject_id
    if (!targetSubId) return allBanks
    const targetGrade = scheduleGradeOf(selectedScheduleForLink.value)
    const targetClassId = scheduleClassIdOf(selectedScheduleForLink.value)
    // Bank tanpa cakupan (data lama) tetap ditampilkan agar bisa ditautkan.
    const matching = allBanks.filter(b => {
      if (b.subject_id !== targetSubId) return false
      if (b.classes?.length) return b.classes.some(c => c.id === targetClassId)
      return !b.grade || !targetGrade || b.grade === targetGrade
    })
    return matching.length > 0 ? matching : allBanks
  })

  const formatScheduleTimeRange = (startTime, endTime) => {
    if (!startTime) return '-'
    const st = new Date(startTime)
    const dateStr = st.toLocaleDateString('id-ID', { weekday: 'short', day: 'numeric', month: 'short', year: 'numeric' })
    const pad = (n) => String(n).padStart(2, '0')
    const timeStartStr = `${pad(st.getHours())}:${pad(st.getMinutes())}`
    if (!endTime) return `${dateStr} • ${timeStartStr} WIB`
    const et = new Date(endTime)
    const timeEndStr = `${pad(et.getHours())}:${pad(et.getMinutes())}`
    return `${dateStr} • ${timeStartStr} - ${timeEndStr} WIB`
  }

  const selectedScheduleClassSubjectId = ref('')

  const alreadyScheduledKeys = computed(() => {
    const currentEventId = selectedExamEvent.value?.id || events.value.find(e => e.is_active)?.id || null
    const set = new Set()
    schedules.value.forEach(s => {
      if (!currentEventId || s.event_id === currentEventId) {
        const classId = s.class_room_id || s.class_id
        const subId = s.subject_id || s.bank?.subject_id
        if (classId && subId) {
          if (isEditSchedule.value && scheduleForm.value.id === s.id) {
            return
          }
          set.add(`${classId}_${subId}`)
        }
      }
    })
    return set
  })

  const availableClassSubjects = computed(() => {
    return classSubjects.value.filter(cs => {
      const key = `${cs.class_id}_${cs.subject_id}`
      if (isEditSchedule.value && selectedScheduleClassSubjectId.value === cs.id) return true
      return !alreadyScheduledKeys.value.has(key)
    })
  })

  const availableClassSubjectGrades = computed(() => {
    const grades = availableClassSubjects.value.map(cs => cs.class_room?.grade || 'Lainnya')
    const unique = [...new Set(grades)]
    const order = ['X', 'XI', 'XII', '7', '8', '9', 'VII', 'VIII', 'IX', '10', '11', '12']
    return unique.sort((a, b) => {
      const idxA = order.indexOf(a)
      const idxB = order.indexOf(b)
      if (idxA !== -1 && idxB !== -1) return idxA - idxB
      if (idxA !== -1) return -1
      if (idxB !== -1) return 1
      return String(a).localeCompare(String(b))
    })
  })

  const groupedAvailableClassSubjects = computed(() => {
    const grouped = {}
    availableClassSubjects.value.forEach(cs => {
      const grade = cs.class_room?.grade || 'Lainnya'
      if (!grouped[grade]) grouped[grade] = []
      grouped[grade].push(cs)
    })
    return grouped
  })

  const onClassSubjectChange = () => {
    const cs = classSubjects.value.find(c => c.id === selectedScheduleClassSubjectId.value)
    if (!cs) return
    scheduleForm.value.class_id = cs.class_id
    scheduleForm.value.subject_id = cs.subject_id
    if (!scheduleForm.value.title || scheduleForm.value.title.startsWith('Ujian ') || scheduleForm.value.title.startsWith('Penilaian ')) {
      scheduleForm.value.title = `Ujian ${cs.subject?.name || ''} - ${cs.class_room?.name || ''}`
    }
  }

  const openCreateSchedule = (defaultSubjectId = null) => {
    isEditSchedule.value = false
    const now = new Date()
    const pad = (n) => String(n).padStart(2, '0')
    const todayStr = `${now.getFullYear()}-${pad(now.getMonth() + 1)}-${pad(now.getDate())}`
    const activeEv = selectedExamEvent.value || events.value.find(e => e.is_active) || events.value[0]

    // Pick first available class-subject
    const firstCS = availableClassSubjects.value[0]
    selectedScheduleClassSubjectId.value = firstCS ? firstCS.id : ''
    const subId = firstCS ? firstCS.subject_id : (defaultSubjectId || subjects.value[0]?.id || '')
    const clsId = firstCS ? firstCS.class_id : (classes.value[0]?.id || '')
    const defaultTitle = firstCS ? `Ujian ${firstCS.subject?.name || ''} - ${firstCS.class_room?.name || ''}` : ''

    scheduleForm.value = {
      id: null,
      event_id: activeEv ? activeEv.id : '',
      title: defaultTitle,
      subject_id: subId,
      bank_id: '',
      class_id: clsId,
      exam_date: todayStr,
      start_time: '07:30',
      end_time: '09:00',
      exam_token: 'CBT' + Math.floor(1000 + Math.random() * 9000),
      duration_minutes: 90,
      max_violations: 3,
      randomize_questions: true,
      randomize_options: true,
      is_active: true,
      proctors: [],
      is_makeup: false,
      makeup_students: [],
    }
    scheduleFormInitialProctorIds.value = []
    allStudentsForClass.value = []
    showScheduleModal.value = true
  }

  const openEditSchedule = async (sch) => {
    isEditSchedule.value = true
    const st = sch.start_time ? new Date(sch.start_time) : new Date()
    const et = sch.end_time ? new Date(sch.end_time) : new Date(st.getTime() + (sch.duration_minutes || 90) * 60000)

    const pad = (n) => String(n).padStart(2, '0')
    const dateStr = `${st.getFullYear()}-${pad(st.getMonth() + 1)}-${pad(st.getDate())}`
    const startTimeStr = `${pad(st.getHours())}:${pad(st.getMinutes())}`
    const endTimeStr = `${pad(et.getHours())}:${pad(et.getMinutes())}`

    const targetClassId = sch.class_room_id || sch.class_id
    const targetSubId = sch.subject_id || sch.bank?.subject_id
    const matchedCS = classSubjects.value.find(cs => cs.class_id === targetClassId && cs.subject_id === targetSubId)
    selectedScheduleClassSubjectId.value = matchedCS ? matchedCS.id : ''

    scheduleForm.value = {
      id: sch.id,
      event_id: sch.event_id || '',
      title: sch.title || '',
      subject_id: targetSubId || (subjects.value[0]?.id || ''),
      bank_id: sch.bank_id || '',
      class_id: targetClassId || (classes.value[0]?.id || ''),
      exam_date: dateStr,
      start_time: startTimeStr,
      end_time: endTimeStr,
      exam_token: sch.exam_token || 'CBT2026',
      duration_minutes: sch.duration_minutes || 90,
      max_violations: sch.max_violations ?? 3,
      randomize_questions: sch.randomize_questions ?? true,
      randomize_options: sch.randomize_options ?? true,
      is_active: sch.is_active ?? true,
      proctors: (sch.proctors || []).map(p => ({ id: p.id, full_name: p.full_name })),
      is_makeup: sch.is_makeup ?? false,
      makeup_students: [],
    }
    scheduleFormInitialProctorIds.value = (sch.proctors || []).map(p => p.id)
    allStudentsForClass.value = []

    // Jika jadwal susulan, muat daftar siswa peserta dan semua siswa kelas secara paralel
    if (sch.is_makeup) {
      try {
        const [makeupRes] = await Promise.all([
          api.get(`/admin/schedules/${sch.id}/makeup-students`),
          loadStudentsForClass(targetClassId),
        ])
        scheduleForm.value.makeup_students = makeupRes.data?.data || []
      } catch (e) {
        // silent – form tetap terbuka
      }
    }

    showScheduleModal.value = true
  }

  // Jadwal lain pada sesi waktu yang sama (event + tanggal + jam mulai) berbagi satu token.
  const localDateTimeParts = (iso) => {
    const d = new Date(iso)
    const pad = (n) => String(n).padStart(2, '0')
    return { date: `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())}`, time: `${pad(d.getHours())}:${pad(d.getMinutes())}` }
  }
  const scheduleFormSessionPeers = computed(() => {
    const f = scheduleForm.value
    if (!f?.exam_date || !f?.start_time) return []
    return schedules.value.filter((sch) => {
      if (sch.id === f.id || !sch.start_time) return false
      if ((sch.event_id || '') !== (f.event_id || '')) return false
      const { date, time } = localDateTimeParts(sch.start_time)
      return date === f.exam_date && time === f.start_time
    })
  })
  watch(scheduleFormSessionPeers, (peers) => {
    // Jadwal baru langsung menampilkan token sesi yang sudah ada (server juga menerapkannya).
    if (!isEditSchedule.value && peers.length && peers[0].exam_token) {
      scheduleForm.value.exam_token = peers[0].exam_token
    }
  })

  const submitScheduleForm = async () => {
    try {
      if (!scheduleForm.value.title || !scheduleForm.value.class_id || !scheduleForm.value.subject_id) {
        await showAlertModal({
          title: 'Form Belum Lengkap',
          message: 'Judul sesi, kelas, dan mata pelajaran wajib diisi.',
          type: 'warning'
        })
        return
      }

      const activeEv = selectedExamEvent.value || events.value.find(e => e.is_active) || events.value[0]
      const eventId = scheduleForm.value.event_id || (activeEv ? activeEv.id : null)

      const payload = {
        event_id: eventId,
        title: scheduleForm.value.title,
        subject_id: scheduleForm.value.subject_id,
        bank_id: scheduleForm.value.bank_id || null,
        class_id: scheduleForm.value.class_id,
        exam_date: scheduleForm.value.exam_date,
        start_time: scheduleForm.value.start_time,
        end_time: scheduleForm.value.end_time,
        exam_token: scheduleForm.value.exam_token.toUpperCase(),
        duration_minutes: scheduleForm.value.duration_minutes,
        max_violations: scheduleForm.value.max_violations,
        randomize_questions: scheduleForm.value.randomize_questions,
        randomize_options: scheduleForm.value.randomize_options,
        is_active: scheduleForm.value.is_active,
        is_makeup: scheduleForm.value.is_makeup,
      }

      let savedScheduleId = scheduleForm.value.id
      if (isEditSchedule.value) {
        await api.put(`/admin/schedules/${scheduleForm.value.id}`, payload)
        showToast('Jadwal ujian berhasil diperbarui!', 'success')
      } else {
        const createRes = await api.post('/admin/schedules', payload)
        savedScheduleId = createRes.data?.data?.id || null
        showToast('Jadwal ujian berhasil diterbitkan!', 'success')
      }

      // Penugasan pengawas dikirim terpisah setelah jadwal tersimpan; kegagalannya tidak membatalkan jadwal.
      let proctorWarning = ''
      if (authStore.hasPermission('schedules:manage')) {
        const proctorIds = scheduleForm.value.proctors.map(p => p.id)
        const needSync = isEditSchedule.value
          ? !sameIdSet(proctorIds, scheduleFormInitialProctorIds.value)
          : proctorIds.length > 0
        if (needSync) {
          if (!savedScheduleId) {
            proctorWarning = 'ID jadwal tidak ditemukan pada respons server.'
          } else {
            try {
              await api.put(`/admin/schedules/${savedScheduleId}/proctors`, { user_ids: proctorIds })
            } catch (pe) {
              proctorWarning = pe.response?.data?.message || 'Terjadi kesalahan saat menyimpan pengawas.'
            }
          }
        }
      }

      // Sinkronisasi daftar siswa susulan (jika is_makeup aktif)
      if (scheduleForm.value.is_makeup && savedScheduleId) {
        try {
          await api.post(`/admin/schedules/${savedScheduleId}/makeup-students`, {
            student_ids: selectedMakeupStudentIDs.value
          })
        } catch (me) {
          // Kegagalan sinkronisasi whitelist tidak membatalkan jadwal yang sudah tersimpan
        }
      }

      showScheduleModal.value = false
      await loadSchedules()
      await loadAllData()

      if (proctorWarning) {
        await showAlertModal({
          title: 'Pengawas Belum Tersimpan',
          message: `Jadwal sudah tersimpan, tetapi penugasan pengawas gagal: ${proctorWarning} Atur ulang pengawas lewat tombol "Atur Pengawas" pada baris jadwal.`,
          type: 'warning'
        })
      }
    } catch (e) {
      await showAlertModal({
        title: 'Gagal Menyimpan Jadwal',
        message: e.response?.data?.message || 'Terjadi kesalahan saat menyimpan data jadwal ujian.',
        type: 'danger'
      })
    }
  }

  const deleteSchedule = async (sch) => {
    const confirmed = await showConfirmModal({
      title: 'Hapus Sesi Ujian',
      message: `Apakah Anda yakin ingin menghapus sesi ujian "${sch.title}" untuk kelas ${sch.class_room?.name || ''}?`,
      type: 'danger',
      confirmText: 'Hapus Sesi',
      cancelText: 'Batal'
    })
    if (!confirmed) return

    try {
      await api.delete(`/admin/schedules/${sch.id}`)
      await loadSchedules()
      await loadAllData()
      showToast('Jadwal ujian berhasil dihapus', 'success')
    } catch (e) {
      await showAlertModal({
        title: 'Gagal Menghapus Jadwal',
        message: e.response?.data?.message || 'Terjadi kesalahan saat menghapus data jadwal ujian.',
        type: 'danger'
      })
    }
  }

  const openLinkBankModal = (sch) => {
    selectedScheduleForLink.value = sch
    selectedBankForLink.value = sch.bank_id || ''
    showLinkBankModal.value = true
  }

  const submitLinkBank = async () => {
    if (!selectedScheduleForLink.value) return
    try {
      await api.post(`/admin/schedules/${selectedScheduleForLink.value.id}/link-bank`, {
        bank_id: selectedBankForLink.value || null
      })
      showLinkBankModal.value = false
      await loadSchedules()
      await loadAllData()
      showToast('Naskah bank soal berhasil ditautkan ke jadwal!', 'success')
    } catch (e) {
      await showAlertModal({
        title: 'Gagal Menautkan Soal',
        message: e.response?.data?.message || 'Terjadi kesalahan saat menautkan bank soal.',
        type: 'danger'
      })
    }
  }

  const scheduleClassIdOf = (s) => s.class_room?.id || s.class_room_id || s.class_id
  const scheduleGradeOf = (s) =>
    s.class_room?.grade || classes.value.find(c => c.id === scheduleClassIdOf(s))?.grade || ''

  const scheduleSubjectOf = (s) => {
    const subId = s.subject_id || s.subject?.id || s.bank?.subject_id || s.bank?.subject?.id
    if (subId) return s.subject || s.bank?.subject || subjects.value.find(sub => sub.id === subId) || null
    if (s.title) return subjects.value.find(sub => s.title.toLowerCase().includes(sub.name.toLowerCase())) || null
    return null
  }

  const toggleScheduleStatus = async (sch) => {
    if (!sch.is_active) {
      if (!sch.bank_id) {
        await showAlertModal({
          title: 'Naskah Soal Belum Ditautkan',
          message: 'Jadwal belum dapat diaktifkan karena belum ada Naskah Bank Soal yang ditautkan.\n\nSilakan klik "Tautkan Soal" terlebih dahulu.',
          type: 'warning'
        })
        return
      }
      if (sch.bank && !sch.bank.is_locked) {
        await showAlertModal({
          title: 'Bank Soal Belum Siap',
          message: 'Bank Soal masih berstatus Draft (Belum Dikunci).\n\nSilakan kunci/finalisasi bank soal pada menu Kesiapan Soal sebelum mengaktifkan sesi ujian.',
          type: 'warning'
        })
        return
      }
    }
    try {
      const res = await api.post(`/admin/schedules/${sch.id}/toggle`)
      sch.is_active = res.data.is_active
      showToast(sch.is_active ? `Sesi ${sch.title} berhasil diaktifkan!` : `Sesi ${sch.title} dinonaktifkan.`, 'success')
    } catch (e) {
      await showAlertModal({
        title: 'Gagal Mengubah Status',
        message: 'Terjadi kesalahan saat mengubah status aktif jadwal sesi ujian.',
        type: 'danger'
      })
    }
  }

  return {
    copyTokenToClipboard,
    scheduleSearchQuery,
    selectedScheduleGrade,
    selectedScheduleTimeFilter,
    scheduleSortKey,
    scheduleSortOrder,
    showScheduleDetailModal,
    selectedScheduleDetail,
    scheduleDetailTab,
    essayQuestions,
    essayLoading,
    essayDraft,
    essaySaving,
    selectedScheduleIds,
    toggleScheduleSort,
    openScheduleDetail,
    canGradeEssay,
    switchToEssayTab,
    essayTotalPending,
    saveEssayQuestion,
    availableScheduleGrades,
    filteredSchedules,
    scheduleCurrentPage,
    schedulePerPage,
    totalSchedulePages,
    paginatedSchedules,
    displayedSchedulePages,
    setSchedulePage,
    isAllPaginatedSelected,
    toggleSelectAllPaginatedSchedules,
    bulkActivateSchedules,
    bulkDeactivateSchedules,
    bulkRegenerateTokens,
    showPrintModal,
    printDocType,
    printSchedule,
    openPrintModal,
    printStudentsList,
    triggerPrintDocument,
    formatScheduleDateFull,
    formatScheduleTimeOnly,
    getDayName,
    showScheduleModal,
    isEditSchedule,
    scheduleForm,
    showProctorAssignModal,
    scheduleForProctors,
    showFormProctorPicker,
    openProctorAssign,
    formatProctorNames,
    applyFormProctors,
    removeFormProctor,
    showLinkBankModal,
    selectedScheduleForLink,
    selectedBankForLink,
    availableBanksForSelectedSchedule,
    formatScheduleTimeRange,
    selectedScheduleClassSubjectId,
    availableClassSubjects,
    availableClassSubjectGrades,
    groupedAvailableClassSubjects,
    onClassSubjectChange,
    openCreateSchedule,
    openEditSchedule,
    scheduleFormSessionPeers,
    submitScheduleForm,
    deleteSchedule,
    openLinkBankModal,
    submitLinkBank,
    scheduleClassIdOf,
    scheduleGradeOf,
    scheduleSubjectOf,
    toggleScheduleStatus,
    showMakeupStudentPicker,
    allStudentsForClass,
    loadingMakeupStudents,
    makeupStudentsList,
    loadingMakeupStudentsList,
    selectedMakeupStudentIDs,
    loadStudentsForClass,
    loadMakeupStudents,
    toggleMakeupStudent,
    isStudentSelected,
  }
}
