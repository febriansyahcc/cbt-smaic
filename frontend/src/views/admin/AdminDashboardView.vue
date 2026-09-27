<template>
  <div class="h-screen flex bg-slate-50 overflow-hidden font-sans">
    <DashboardSidebar />

    <!-- Right Area: Top Navbar + Main Content -->
    <div class="flex-1 flex flex-col min-w-0 overflow-hidden">
      <DashboardTopbar />

      <!-- Main Scrollable Area: satu tab tampil sesuai activeTab (v-if di tiap komponen) -->
      <main class="flex-1 overflow-y-auto p-4 sm:p-6 space-y-6">
        <HomeTab />
        <SchedulesTab />
        <EventsTab />
        <SubjectsTab />
        <ClassSubjectsTab />
        <StudentsTab />
        <TeachersTab />
        <ClassesTab />
        <ProctorTab />
        <QuestionBanksTab />
      </main>
    </div>

    <!-- Modal: urutan dipertahankan agar tumpukan (z-order) antarmodal tetap sama -->
    <ScheduleModals />
    <EventFormModal />
    <StudentModals />
    <TeacherModals />
    <ClassModals />
    <SubjectModals />
    <MasterImportModals />
    <ExamDocumentModals />
    <QuestionBankModals />
    <QuestionManagerModal />
    <DashboardDialogs />

    <!-- STUDENT EXAM SIMULATOR MODAL -->
    <StudentExamSimulatorModal
      :is-open="showSimulatorModal"
      :bank="viewingBank"
      :questions="bankQuestions"
      @close="showSimulatorModal = false"
    />
  </div>
</template>

<script setup>
import { ref, computed, watch, onMounted, onUnmounted, provide } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '../../stores/auth'
import api from '../../services/api'
import StudentExamSimulatorModal from '../../components/admin/StudentExamSimulatorModal.vue'
import DashboardSidebar from './dashboard/DashboardSidebar.vue'
import DashboardTopbar from './dashboard/DashboardTopbar.vue'
import HomeTab from './dashboard/HomeTab.vue'
import SchedulesTab from './dashboard/SchedulesTab.vue'
import EventsTab from './dashboard/EventsTab.vue'
import SubjectsTab from './dashboard/SubjectsTab.vue'
import ClassSubjectsTab from './dashboard/ClassSubjectsTab.vue'
import StudentsTab from './dashboard/StudentsTab.vue'
import TeachersTab from './dashboard/TeachersTab.vue'
import ClassesTab from './dashboard/ClassesTab.vue'
import ProctorTab from './dashboard/ProctorTab.vue'
import QuestionBanksTab from './dashboard/QuestionBanksTab.vue'
import ScheduleModals from './dashboard/ScheduleModals.vue'
import EventFormModal from './dashboard/EventFormModal.vue'
import StudentModals from './dashboard/StudentModals.vue'
import TeacherModals from './dashboard/TeacherModals.vue'
import ClassModals from './dashboard/ClassModals.vue'
import SubjectModals from './dashboard/SubjectModals.vue'
import MasterImportModals from './dashboard/MasterImportModals.vue'
import ExamDocumentModals from './dashboard/ExamDocumentModals.vue'
import QuestionBankModals from './dashboard/QuestionBankModals.vue'
import QuestionManagerModal from './dashboard/QuestionManagerModal.vue'
import DashboardDialogs from './dashboard/DashboardDialogs.vue'
import { DASHBOARD_CONTEXT } from './dashboard/context'
import { landingTab } from '../../utils/access'
import {
  applyImplications,
  effectivePermissions,
  templateByKey,
  templateKeyFor,
  templateLabelFor,
} from '../../utils/staffAccess'

const router = useRouter()
const authStore = useAuthStore()

// Label portal mengikuti izin, bukan role
const portalLabel = computed(() =>
  authStore.hasPermission('master:manage') ? 'Portal Kurikulum & Admin' : 'Portal Guru / Pengawas'
)
const portalRoleLabel = computed(() =>
  authStore.hasPermission('master:manage') ? 'Kurikulum & CBT' : 'Guru & Pengawas'
)

// Sidebar & Layout State
const activeTab = ref('dashboard')
const isSidebarCollapsed = ref(false)
const isMobileSidebarOpen = ref(false)

// Dynamic Header Title & Tag based on current menu
const currentSectionMeta = computed(() => {
  switch (activeTab.value) {
    case 'dashboard':
      return {
        title: 'Beranda',
        tag: 'Ringkasan',
        description: 'Pusat kendali dan monitoring menyeluruh aktivitas ujian sekolah'
      }
    case 'events':
      return {
        title: 'Event Ujian',
        tag: 'Periode Ujian',
        description: 'Kelola kalender pekan ujian (PSAT, ASAT, PTS, PAT, Gladi Bersih, US)'
      }
    case 'schedules':
      return {
        title: 'Jadwal Ujian',
        tag: 'Pelaksanaan',
        description: 'Pengaturan sesi ruang, alokasi rombel, token sesi, dan durasi ujian'
      }
    case 'questions':
      return {
        title: 'Bank Soal & Kesiapan',
        tag: 'Naskah & Kesiapan',
        description: 'Pengelolaan paket naskah butir soal, impor file Excel, dan finalisasi status siap ujian'
      }
    case 'proctor':
      return {
        title: 'Live Proctoring',
        tag: 'Pengawasan',
        description: 'Pemantauan real-time integritas pengerjaan siswa, pelanggaran layar, dan berita acara ujian'
      }
    case 'students':
      return {
        title: 'Data Siswa',
        tag: 'Master Data',
        description: 'Daftar data siswa, nomor induk (NIS/NISN), alokasi kelas, dan reset sesi perangkat'
      }
    case 'teachers':
      return {
        title: 'Guru dan Staf',
        tag: 'Master Data',
        description: 'Daftar akun pendidik pengampu mata pelajaran, penyusun naskah soal, dan staf pengelola CBT'
      }
    case 'classes':
      return {
        title: 'Data Kelas',
        tag: 'Master Data',
        description: 'Struktur ruang kelas, tingkatan (grade), dan konsentrasi keahlian/jurusan'
      }
    case 'subjects':
      return {
        title: 'Mata Pelajaran',
        tag: 'Master Data',
        description: 'Daftar mata pelajaran kurikulum dan kode unik bidang studi asesmen'
      }
    case 'class-subjects':
      return {
        title: 'Kelas Mapel',
        tag: 'Master Data',
        description: 'Alokasi mata pelajaran per rombel dan penetapan guru pengampu'
      }
    default:
      return {
        title: 'Portal CBT',
        tag: 'Kurikulum',
        description: 'Pusat kontrol dan tata kelola ujian sekolah'
      }
  }
})

// Event Management & Scoping States
const events = ref([])
const selectedEventId = ref(sessionStorage.getItem('cbt_selected_event_id') || '')
const selectedEventFilter = ref(sessionStorage.getItem('cbt_selected_event_id') || '')
const isEventDropdownOpen = ref(false)

const selectedExamEvent = computed(() => {
  const id = selectedEventId.value || selectedEventFilter.value
  if (!id) return null
  return events.value.find(e => e.id === id) || null
})

const activeExamEvent = computed(() => {
  return events.value.find(e => e.is_active) || null
})

const currentScopedEventId = computed(() => {
  return selectedExamEvent.value?.id || selectedEventId.value || selectedEventFilter.value || null
})

const currentEventSchedules = computed(() => {
  if (!currentScopedEventId.value) return schedules.value
  return schedules.value.filter(s => s.event_id === currentScopedEventId.value)
})

// Contextual Metric Computeds
const eventsCount = computed(() => events.value.length)
const activeEventsCount = computed(() => events.value.filter(e => e.is_active).length)
const totalSchedulesCount = computed(() => currentEventSchedules.value.length)
const activeSchedulesCount = computed(() => currentEventSchedules.value.filter(s => s.is_active).length)
const scheduledClassesCount = computed(() => new Set(currentEventSchedules.value.map(s => s.class_room_id || s.class_id)).size)
const selectedEventFilterName = computed(() => {
  const ev = selectedExamEvent.value || (currentScopedEventId.value ? events.value.find(e => e.id === currentScopedEventId.value) : null)
  return ev ? ev.code : 'Semua Event'
})

const lockedStudentsCount = computed(() => students.value.filter(s => s.user?.session_token).length)
const maleStudentsCount = computed(() => students.value.filter(s => s.gender === 'L' || s.user?.gender === 'L').length)
const femaleStudentsCount = computed(() => students.value.filter(s => s.gender === 'P' || s.user?.gender === 'P').length)
// Guru dan Pengawas = semua akun non-Administrator (termasuk Kustom), sehingga jumlahnya konsisten dengan Total
const guruCount = computed(() => teachers.value.filter(t => t.role !== 'ADMIN').length)
const adminStaffCount = computed(() => teachers.value.filter(t => t.role === 'ADMIN').length)

const questionMakersCount = computed(() => {
  const banks = readinessData.value.question_banks || []
  return new Set(banks.map(b => b.created_by_id)).size
})

const subjectsWithBanksCount = computed(() => {
  const banks = readinessData.value.question_banks || []
  return new Set(banks.map(b => b.subject_id)).size
})

const allocatedClassesCount = computed(() => new Set(classSubjects.value.map(cs => cs.class_id)).size)
const allocatedTeachersCount = computed(() => new Set(classSubjects.value.map(cs => cs.teacher_id)).size)

const getSubjectBankCount = (subId) => {
  const banks = readinessData.value.question_banks || []
  return banks.filter(b => b.subject_id === subId).length
}

const getSubjectClassCount = (subId) => {
  return classSubjects.value.filter(cs => cs.subject_id === subId).length
}

const gradeCounts = computed(() => {
  const counts = { 'X': 0, 'XI': 0, 'XII': 0 }
  classes.value.forEach(c => {
    if (counts[c.grade] !== undefined) {
      counts[c.grade]++
    }
  })
  return counts
})

const lockedBanksCount = computed(() => (readinessData.value.question_banks || []).filter(b => b.is_locked).length)
const draftBanksCount = computed(() => (readinessData.value.question_banks || []).filter(b => !b.is_locked).length)
const totalQuestionsCount = computed(() => (readinessData.value.question_banks || []).reduce((acc, b) => acc + (b.total_questions || 0), 0))

// Cloudflare-style Event Context Switcher Actions (Two-Tier Model)

const enterEvent = (ev) => {
  selectedEventId.value = ev.id
  selectedEventFilter.value = ev.id
  sessionStorage.setItem('cbt_selected_event_id', ev.id)
  isEventDropdownOpen.value = false
  activeTab.value = authStore.canAccessTab('schedules') ? 'schedules' : landingTab(authStore.user)
  loadSchedules()
}

const selectEvent = (ev) => {
  selectedEventId.value = ev.id
  selectedEventFilter.value = ev.id
  sessionStorage.setItem('cbt_selected_event_id', ev.id)
  isEventDropdownOpen.value = false
  loadSchedules()
}

const exitEventScope = () => {
  selectedEventId.value = ''
  selectedEventFilter.value = ''
  sessionStorage.removeItem('cbt_selected_event_id')
  isEventDropdownOpen.value = false
  activeTab.value = 'events'
}

const stats = ref({})
const schedules = ref([])

// Global Reactive Dialog State & Methods (Replaces native browser alert & confirm)
const dialogState = ref({
  isOpen: false,
  type: 'confirm', // 'confirm' | 'danger' | 'alert' | 'warning'
  title: '',
  message: '',
  confirmText: 'OK',
  cancelText: '',
  resolve: null
})

const showConfirmModal = ({ title = 'Konfirmasi', message = '', type = 'confirm', confirmText = 'Lanjutkan', cancelText = 'Batal' }) => {
  return new Promise((resolve) => {
    dialogState.value = {
      isOpen: true,
      type,
      title,
      message,
      confirmText,
      cancelText,
      resolve
    }
  })
}

const showAlertModal = ({ title = 'Pemberitahuan', message = '', type = 'alert', confirmText = 'Mengerti' }) => {
  return new Promise((resolve) => {
    dialogState.value = {
      isOpen: true,
      type,
      title,
      message,
      confirmText,
      cancelText: '',
      resolve
    }
  })
}

const handleDialogConfirm = () => {
  if (dialogState.value.resolve) dialogState.value.resolve(true)
  dialogState.value.isOpen = false
}

const handleDialogCancel = () => {
  if (dialogState.value.resolve) dialogState.value.resolve(false)
  dialogState.value.isOpen = false
}

// Toast Notification State
const toastMessage = ref('')
const toastType = ref('success')
let toastTimeout = null
const showToast = (msg, type = 'success') => {
  toastMessage.value = msg
  toastType.value = type
  if (toastTimeout) clearTimeout(toastTimeout)
  toastTimeout = setTimeout(() => {
    toastMessage.value = ''
  }, 3000)
}

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
const selectedScheduleTimeFilter = ref('all') // 'all' | 'today' | 'upcoming'
const scheduleSortKey = ref('time')
const scheduleSortOrder = ref('asc')
const showScheduleDetailModal = ref(false)
const selectedScheduleDetail = ref(null)

// Essay Correction State (for Schedule Detail Modal)
const scheduleDetailTab = ref('info') // 'info' | 'essay'
const essayQuestions = ref([])
const essayLoading = ref(false)
const essayDraft = ref({}) // keyed by answer_id: { score_awarded: number|null, teacher_comment: string }
const essaySaving = ref({}) // keyed by question_id: boolean

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
const printDocType = ref('attendance') // 'attendance' | 'report'
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

const students = ref([])
const teachers = ref([])
const classes = ref([])
const readinessData = ref({})

// Subject & ClassSubject States
const subjects = ref([])
const showSubjectModal = ref(false)
const isEditSubject = ref(false)
const subjectForm = ref({ id: null, code: '', name: '' })

const classSubjects = ref([])
const showClassSubjectModal = ref(false)
const classSubjectForm = ref({
  id: null,
  class_id: '',
  subject_id: '',
  teacher_id: '',
  academic_year: '2026/2027',
})
const filterClassSubjectClass = ref('')

// Event Management States
const eventSearchQuery = ref('')
const eventStatusFilter = ref('all')
const activeEventMenuId = ref(null)
const showCardPrint = ref(false)
const cardPrintEvent = ref(null)
const openCardPrint = (ev) => {
  cardPrintEvent.value = ev
  showCardPrint.value = true
}

const toggleEventMenu = (id) => {
  activeEventMenuId.value = activeEventMenuId.value === id ? null : id
}
const showEventModal = ref(false)
const isEditEvent = ref(false)
const eventForm = ref({
  id: null,
  title: '',
  code: '',
  academic_year: '2026/2027',
  semester: 'GANJIL',
  start_date: '',
  end_date: '',
  is_active: true,
  description: '',
})

const filteredEvents = computed(() => {
  let list = events.value
  if (eventStatusFilter.value === 'active') {
    list = list.filter(e => e.is_active)
  } else if (eventStatusFilter.value === 'archived') {
    list = list.filter(e => !e.is_active)
  }
  if (eventSearchQuery.value.trim()) {
    const q = eventSearchQuery.value.toLowerCase().trim()
    list = list.filter(e =>
      (e.title && e.title.toLowerCase().includes(q)) ||
      (e.code && e.code.toLowerCase().includes(q)) ||
      (e.academic_year && e.academic_year.toLowerCase().includes(q))
    )
  }
  return list
})

// Modals
const showScheduleModal = ref(false)
const showStudentModal = ref(false)
const showTeacherModal = ref(false)
const showClassModal = ref(false)
const showImportModal = ref(false)
const importExcelFile = ref(null)
const isImporting = ref(false)

// Import states — Kelas
const showImportClassModal = ref(false)
const importClassFile = ref(null)
const isImportingClass = ref(false)
const importClassResult = ref(null)

// Import states — Mata Pelajaran
const showImportSubjectModal = ref(false)
const importSubjectFile = ref(null)
const isImportingSubject = ref(false)
const importSubjectResult = ref(null)

// Import states — Guru
const showImportTeacherModal = ref(false)
const importTeacherFile = ref(null)
const isImportingTeacher = ref(false)
const importTeacherResult = ref(null)

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
})

// Penugasan pengawas: modal dari tabel jadwal dan pemilih di dalam form jadwal
const showProctorAssignModal = ref(false)
const scheduleForProctors = ref(null)
const showFormProctorPicker = ref(false)
// Daftar id pengawas saat form dibuka, dipakai untuk mendeteksi perubahan saat menyimpan
const scheduleFormInitialProctorIds = ref([])

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
  }
  scheduleFormInitialProctorIds.value = []
  showScheduleModal.value = true
}

const openEditSchedule = (sch) => {
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
  }
  scheduleFormInitialProctorIds.value = (sch.proctors || []).map(p => p.id)
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

// ---------------- MASTER DATA STATES & COMPUTEDS ----------------

// Form states for user/class
const newStudent = ref({ full_name: '', username: '', password: '', nis: '', nisn: '', class_id: '', gender: 'L' })
const newTeacher = ref({ full_name: '', username: '', password: '' })
const newTeacherAccess = ref({ role: 'GURU', permissions: [] })
const newTeacherAccessValid = ref(true)

// Hanya grantor (ADMIN atau pemegang izin "*") yang boleh mengatur izin, username, dan password akun staf.
// Server tetap menjadi penjaga sebenarnya (403); ini hanya untuk kerapian UI.
const isGrantor = computed(() => authStore.role === 'ADMIN' || authStore.permissions.includes('*'))
const newTeacherCanSave = computed(() => !isGrantor.value || newTeacherAccessValid.value)
const newClass = ref({ name: '', grade: 'XII', major: 'MIPA' })

// 1. Data Siswa Filtering, Sorting & Pagination
const studentSearchQuery = ref('')
const selectedStudentGrade = ref('all')
const selectedStudentSessionFilter = ref('all')
const studentSortKey = ref('name')
const studentSortOrder = ref('asc')
const studentCurrentPage = ref(1)
const studentPerPage = ref(10)
const showStudentDetailModal = ref(false)
const selectedStudentDetail = ref(null)
const showEditStudentModal = ref(false)
const editStudentForm = ref({ id: null, user_id: null, full_name: '', username: '', password: '', nis: '', nisn: '', class_id: '', gender: 'L' })

const availableStudentGrades = computed(() => {
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

const toggleStudentSort = (key) => {
  if (studentSortKey.value === key) {
    studentSortOrder.value = studentSortOrder.value === 'asc' ? 'desc' : 'asc'
  } else {
    studentSortKey.value = key
    studentSortOrder.value = 'asc'
  }
}

const resetStudentFilters = () => {
  studentSearchQuery.value = ''
  selectedStudentGrade.value = 'all'
  selectedStudentSessionFilter.value = 'all'
  studentSortKey.value = 'name'
  studentSortOrder.value = 'asc'
  studentCurrentPage.value = 1
}

const filteredStudents = computed(() => {
  let list = students.value || []

  // Filter by Grade
  if (selectedStudentGrade.value !== 'all') {
    list = list.filter(st => {
      const g = st.class_room?.grade || (classes.value.find(c => c.id === (st.class_room_id || st.class_id))?.grade)
      return g === selectedStudentGrade.value
    })
  }

  // Filter by Session
  if (selectedStudentSessionFilter.value === 'locked') {
    list = list.filter(st => st.user?.session_token)
  }

  // Filter by Search Query
  if (studentSearchQuery.value.trim()) {
    const q = studentSearchQuery.value.toLowerCase().trim()
    list = list.filter(st => {
      const nameMatch = st.user?.full_name && st.user.full_name.toLowerCase().includes(q)
      const userMatch = st.user?.username && st.user.username.toLowerCase().includes(q)
      const nisMatch = st.nis && st.nis.toLowerCase().includes(q)
      const nisnMatch = st.nisn && st.nisn.toLowerCase().includes(q)
      const classMatch = st.class_room?.name && st.class_room.name.toLowerCase().includes(q)
      return nameMatch || userMatch || nisMatch || nisnMatch || classMatch
    })
  }

  // Sort
  if (studentSortKey.value) {
    list = [...list].sort((a, b) => {
      let result = 0
      if (studentSortKey.value === 'name') {
        const valA = (a.user?.full_name || '').trim().toLowerCase()
        const valB = (b.user?.full_name || '').trim().toLowerCase()
        result = valA.localeCompare(valB, undefined, { numeric: true, sensitivity: 'base' })
      } else if (studentSortKey.value === 'username') {
        const valA = (a.user?.username || '').trim().toLowerCase()
        const valB = (b.user?.username || '').trim().toLowerCase()
        result = valA.localeCompare(valB, undefined, { numeric: true, sensitivity: 'base' })
      } else if (studentSortKey.value === 'nis') {
        const valA = (a.nis || '').trim().toLowerCase()
        const valB = (b.nis || '').trim().toLowerCase()
        result = valA.localeCompare(valB, undefined, { numeric: true, sensitivity: 'base' })
      } else if (studentSortKey.value === 'class') {
        const valA = (a.class_room?.name || '').trim().toLowerCase()
        const valB = (b.class_room?.name || '').trim().toLowerCase()
        result = valA.localeCompare(valB, undefined, { numeric: true, sensitivity: 'base' })
      } else if (studentSortKey.value === 'session') {
        const lockA = a.user?.session_token ? 1 : 0
        const lockB = b.user?.session_token ? 1 : 0
        result = lockA - lockB
      }
      return studentSortOrder.value === 'asc' ? result : -result
    })
  }

  return list
})

const totalStudentPages = computed(() => Math.ceil(filteredStudents.value.length / studentPerPage.value) || 1)
const paginatedStudents = computed(() => {
  const start = (studentCurrentPage.value - 1) * studentPerPage.value
  return filteredStudents.value.slice(start, start + studentPerPage.value)
})

const displayedStudentPages = computed(() => {
  const total = totalStudentPages.value
  const current = studentCurrentPage.value
  if (total <= 7) {
    return Array.from({ length: total }, (_, i) => i + 1)
  }
  const pages = []
  pages.push(1)
  if (current > 3) pages.push('...')
  const start = Math.max(2, current - 1)
  const end = Math.min(total - 1, current + 1)
  for (let i = start; i <= end; i++) {
    pages.push(i)
  }
  if (current < total - 2) pages.push('...')
  pages.push(total)
  return pages
})

const setStudentPage = (p) => {
  if (p === '...' || p < 1 || p > totalStudentPages.value) return
  studentCurrentPage.value = p
}

watch([studentSearchQuery, selectedStudentGrade, selectedStudentSessionFilter, studentPerPage, studentSortKey, studentSortOrder], () => {
  studentCurrentPage.value = 1
})

// 2. Guru dan Staf Filtering, Sorting & Pagination
const teacherSearchQuery = ref('')
const selectedTeacherRole = ref('all')
const teacherSortKey = ref('name')
const teacherSortOrder = ref('asc')
const teacherCurrentPage = ref(1)
const teacherPerPage = ref(10)
const showEditTeacherModal = ref(false)
const editTeacherForm = ref({ id: null, full_name: '', username: '', password: '' })
const editTeacherAccess = ref({ role: 'GURU', permissions: [] })
const editTeacherAccessValid = ref(true)
const editTeacherCanSave = computed(() => !isGrantor.value || editTeacherAccessValid.value)

// Katalog izin (GET /admin/permission-catalog): dimuat sekali, dipakai editor akses dan label peran turunan.
const permissionCatalog = ref(null)
let permissionCatalogRequest = null
const loadPermissionCatalog = () => {
  if (permissionCatalog.value) return Promise.resolve(permissionCatalog.value)
  if (!permissionCatalogRequest) {
    permissionCatalogRequest = api.get('/admin/permission-catalog')
      .then((res) => {
        const d = res.data?.data || {}
        permissionCatalog.value = {
          groups: d.groups || [],
          templates: d.templates || [],
          implies: d.implies || {},
        }
        return permissionCatalog.value
      })
      .finally(() => { permissionCatalogRequest = null })
  }
  return permissionCatalogRequest
}

const ensurePermissionCatalog = async () => {
  try {
    return await loadPermissionCatalog()
  } catch (e) {
    await showAlertModal({
      title: 'Gagal Memuat Daftar Izin',
      message: e.response?.data?.message || 'Daftar izin akses tidak dapat dimuat. Coba lagi beberapa saat.',
      type: 'danger'
    })
    return null
  }
}

// Peran turunan akun staf: 'admin' | 'guru' | 'pengawas' | 'custom'
const teacherKey = (t) => templateKeyFor(t, permissionCatalog.value?.templates, permissionCatalog.value?.implies)
const teacherRoleLabel = (t) => templateLabelFor(t, permissionCatalog.value?.templates, permissionCatalog.value?.implies)
const TEACHER_BADGE_CLASSES = {
  admin: 'bg-purple-100 text-purple-800',
  guru: 'bg-indigo-100 text-indigo-800',
  pengawas: 'bg-emerald-100 text-emerald-800',
  custom: 'bg-amber-100 text-amber-800',
  unknown: 'bg-slate-100 text-slate-500',
}
const teacherBadgeClass = (t) => TEACHER_BADGE_CLASSES[teacherKey(t)]

const TEACHER_ROLE_CHIPS = [
  { key: 'admin', label: 'Administrator' },
  { key: 'guru', label: 'Guru' },
  { key: 'pengawas', label: 'Pengawas' },
  { key: 'custom', label: 'Kustom' },
]
const teacherRoleChips = computed(() => {
  const list = teachers.value || []
  const counts = { admin: 0, guru: 0, pengawas: 0, custom: 0 }
  for (const t of list) {
    const key = teacherKey(t)
    if (key in counts) counts[key]++ // 'unknown' (katalog belum termuat) tidak dihitung ke kategori mana pun
  }
  const chips = [{ key: 'all', label: `Semua (${list.length})` }]
  for (const c of TEACHER_ROLE_CHIPS) {
    if (counts[c.key] > 0 || selectedTeacherRole.value === c.key) {
      chips.push({ key: c.key, label: `${c.label} (${counts[c.key]})` })
    }
  }
  return chips
})

// Nilai awal editor: template Guru (izin terisi dari template).
const defaultStaffAccess = (catalog) => {
  const guru = templateByKey(catalog.templates, 'guru')
  return { role: 'GURU', permissions: applyImplications(guru?.permissions, catalog.implies) }
}

// Nilai awal editor dari akun: izin kosong pada akun GURU = izin template Guru.
// Kunci yang tidak ada di katalog (usang atau tak dikenal, mis. "*") disaring lebih dulu.
const staffAccessFromAccount = (account, catalog) => {
  if (account.role === 'ADMIN') return { role: 'ADMIN', permissions: ['*'] }
  const known = new Set(catalog.groups.flatMap((g) => (g.items || []).map((i) => i.key)))
  const effective = effectivePermissions(account, catalog.templates).filter((p) => known.has(p))
  return { role: 'GURU', permissions: applyImplications(effective, catalog.implies) }
}

// Payload sesuai kontrak. Grantor: ADMIN tanpa permissions, GURU dengan permissions eksplisit.
// Bukan grantor: tidak boleh mengirim permissions, role dipaksa GURU.
const buildStaffCreatePayload = (form, access) => {
  const payload = { full_name: form.full_name, username: form.username, role: isGrantor.value ? access.role : 'GURU' }
  if (form.password) payload.password = form.password
  if (isGrantor.value && access.role === 'GURU') payload.permissions = [...access.permissions]
  return payload
}

// Bukan grantor: hanya full_name yang boleh diubah (tanpa role, permissions, username, password).
const buildStaffUpdatePayload = (form, access) => {
  if (!isGrantor.value) return { full_name: form.full_name }
  return buildStaffCreatePayload(form, access)
}

const toggleTeacherSort = (key) => {
  if (teacherSortKey.value === key) {
    teacherSortOrder.value = teacherSortOrder.value === 'asc' ? 'desc' : 'asc'
  } else {
    teacherSortKey.value = key
    teacherSortOrder.value = 'asc'
  }
}

const filteredTeachers = computed(() => {
  let list = teachers.value || []
  if (selectedTeacherRole.value !== 'all') {
    list = list.filter(t => teacherKey(t) === selectedTeacherRole.value)
  }
  if (teacherSearchQuery.value.trim()) {
    const q = teacherSearchQuery.value.toLowerCase().trim()
    list = list.filter(t => {
      const nameMatch = t.full_name && t.full_name.toLowerCase().includes(q)
      const userMatch = t.username && t.username.toLowerCase().includes(q)
      return nameMatch || userMatch
    })
  }
  if (teacherSortKey.value) {
    list = [...list].sort((a, b) => {
      let result = 0
      if (teacherSortKey.value === 'name') {
        const valA = (a.full_name || '').trim().toLowerCase()
        const valB = (b.full_name || '').trim().toLowerCase()
        result = valA.localeCompare(valB, undefined, { numeric: true, sensitivity: 'base' })
      } else if (teacherSortKey.value === 'username') {
        const valA = (a.username || '').trim().toLowerCase()
        const valB = (b.username || '').trim().toLowerCase()
        result = valA.localeCompare(valB, undefined, { numeric: true, sensitivity: 'base' })
      } else if (teacherSortKey.value === 'role') {
        const valA = teacherRoleLabel(a).toLowerCase()
        const valB = teacherRoleLabel(b).toLowerCase()
        result = valA.localeCompare(valB, undefined, { numeric: true, sensitivity: 'base' })
      }
      return teacherSortOrder.value === 'asc' ? result : -result
    })
  }
  return list
})

const totalTeacherPages = computed(() => Math.ceil(filteredTeachers.value.length / teacherPerPage.value) || 1)
const paginatedTeachers = computed(() => {
  const start = (teacherCurrentPage.value - 1) * teacherPerPage.value
  return filteredTeachers.value.slice(start, start + teacherPerPage.value)
})

const displayedTeacherPages = computed(() => {
  const total = totalTeacherPages.value
  const current = teacherCurrentPage.value
  if (total <= 7) return Array.from({ length: total }, (_, i) => i + 1)
  const pages = [1]
  if (current > 3) pages.push('...')
  const start = Math.max(2, current - 1)
  const end = Math.min(total - 1, current + 1)
  for (let i = start; i <= end; i++) pages.push(i)
  if (current < total - 2) pages.push('...')
  pages.push(total)
  return pages
})

const setTeacherPage = (p) => {
  if (p === '...' || p < 1 || p > totalTeacherPages.value) return
  teacherCurrentPage.value = p
}

watch([teacherSearchQuery, selectedTeacherRole, teacherPerPage, teacherSortKey, teacherSortOrder], () => {
  teacherCurrentPage.value = 1
})

// 3. Data Kelas Filtering, Sorting & Pagination
const classSearchQuery = ref('')
const selectedClassGrade = ref('all')
const classSortKey = ref('name')
const classSortOrder = ref('asc')
const classCurrentPage = ref(1)
const classPerPage = ref(10)
const showEditClassModal = ref(false)
const editClassForm = ref({ id: null, name: '', grade: 'XII', major: 'MIPA' })

const toggleClassSort = (key) => {
  if (classSortKey.value === key) {
    classSortOrder.value = classSortOrder.value === 'asc' ? 'desc' : 'asc'
  } else {
    classSortKey.value = key
    classSortOrder.value = 'asc'
  }
}

const filteredClasses = computed(() => {
  let list = classes.value || []
  if (selectedClassGrade.value !== 'all') {
    list = list.filter(c => c.grade === selectedClassGrade.value)
  }
  if (classSearchQuery.value.trim()) {
    const q = classSearchQuery.value.toLowerCase().trim()
    list = list.filter(c => {
      const nameMatch = c.name && c.name.toLowerCase().includes(q)
      const gradeMatch = c.grade && c.grade.toLowerCase().includes(q)
      const majorMatch = c.major && c.major.toLowerCase().includes(q)
      return nameMatch || gradeMatch || majorMatch
    })
  }
  if (classSortKey.value) {
    list = [...list].sort((a, b) => {
      let result = 0
      if (classSortKey.value === 'name') {
        const valA = (a.name || '').trim().toLowerCase()
        const valB = (b.name || '').trim().toLowerCase()
        result = valA.localeCompare(valB, undefined, { numeric: true, sensitivity: 'base' })
      } else if (classSortKey.value === 'grade') {
        const order = ['X', 'XI', 'XII', '7', '8', '9', 'VII', 'VIII', 'IX', '10', '11', '12']
        const idxA = order.indexOf(a.grade)
        const idxB = order.indexOf(b.grade)
        if (idxA !== -1 && idxB !== -1) result = idxA - idxB
        else result = (a.grade || '').localeCompare(b.grade || '')
      } else if (classSortKey.value === 'major') {
        const valA = (a.major || '').trim().toLowerCase()
        const valB = (b.major || '').trim().toLowerCase()
        result = valA.localeCompare(valB, undefined, { numeric: true, sensitivity: 'base' })
      }
      return classSortOrder.value === 'asc' ? result : -result
    })
  }
  return list
})

const totalClassPages = computed(() => Math.ceil(filteredClasses.value.length / classPerPage.value) || 1)
const paginatedClasses = computed(() => {
  const start = (classCurrentPage.value - 1) * classPerPage.value
  return filteredClasses.value.slice(start, start + classPerPage.value)
})

const displayedClassPages = computed(() => {
  const total = totalClassPages.value
  const current = classCurrentPage.value
  if (total <= 7) return Array.from({ length: total }, (_, i) => i + 1)
  const pages = [1]
  if (current > 3) pages.push('...')
  const start = Math.max(2, current - 1)
  const end = Math.min(total - 1, current + 1)
  for (let i = start; i <= end; i++) pages.push(i)
  if (current < total - 2) pages.push('...')
  pages.push(total)
  return pages
})

const setClassPage = (p) => {
  if (p === '...' || p < 1 || p > totalClassPages.value) return
  classCurrentPage.value = p
}

watch([classSearchQuery, selectedClassGrade, classPerPage, classSortKey, classSortOrder], () => {
  classCurrentPage.value = 1
})

// 4. Mata Pelajaran Filtering, Sorting & Pagination
const subjectSearchQuery = ref('')
const subjectSortKey = ref('name')
const subjectSortOrder = ref('asc')
const subjectCurrentPage = ref(1)
const subjectPerPage = ref(10)

const toggleSubjectSort = (key) => {
  if (subjectSortKey.value === key) {
    subjectSortOrder.value = subjectSortOrder.value === 'asc' ? 'desc' : 'asc'
  } else {
    subjectSortKey.value = key
    subjectSortOrder.value = 'asc'
  }
}

const filteredSubjects = computed(() => {
  let list = subjects.value || []
  if (subjectSearchQuery.value.trim()) {
    const q = subjectSearchQuery.value.toLowerCase().trim()
    list = list.filter(s => {
      const codeMatch = s.code && s.code.toLowerCase().includes(q)
      const nameMatch = s.name && s.name.toLowerCase().includes(q)
      return codeMatch || nameMatch
    })
  }
  if (subjectSortKey.value) {
    list = [...list].sort((a, b) => {
      let result = 0
      if (subjectSortKey.value === 'code') {
        const valA = (a.code || '').trim().toLowerCase()
        const valB = (b.code || '').trim().toLowerCase()
        result = valA.localeCompare(valB, undefined, { numeric: true, sensitivity: 'base' })
      } else if (subjectSortKey.value === 'name') {
        const valA = (a.name || '').trim().toLowerCase()
        const valB = (b.name || '').trim().toLowerCase()
        result = valA.localeCompare(valB, undefined, { numeric: true, sensitivity: 'base' })
      }
      return subjectSortOrder.value === 'asc' ? result : -result
    })
  }
  return list
})

const totalSubjectPages = computed(() => Math.ceil(filteredSubjects.value.length / subjectPerPage.value) || 1)
const paginatedSubjects = computed(() => {
  const start = (subjectCurrentPage.value - 1) * subjectPerPage.value
  return filteredSubjects.value.slice(start, start + subjectPerPage.value)
})

const displayedSubjectPages = computed(() => {
  const total = totalSubjectPages.value
  const current = subjectCurrentPage.value
  if (total <= 7) return Array.from({ length: total }, (_, i) => i + 1)
  const pages = [1]
  if (current > 3) pages.push('...')
  const start = Math.max(2, current - 1)
  const end = Math.min(total - 1, current + 1)
  for (let i = start; i <= end; i++) pages.push(i)
  if (current < total - 2) pages.push('...')
  pages.push(total)
  return pages
})

const setSubjectPage = (p) => {
  if (p === '...' || p < 1 || p > totalSubjectPages.value) return
  subjectCurrentPage.value = p
}

watch([subjectSearchQuery, subjectPerPage, subjectSortKey, subjectSortOrder], () => {
  subjectCurrentPage.value = 1
})

const formatDate = (d) => {
  if (!d) return '-'
  const date = new Date(d)
  return date.toLocaleDateString('id-ID', { day: 'numeric', month: 'short', year: 'numeric' })
}

const EVENT_SCOPED_TABS = ['schedules', 'questions', 'proctor']

const switchTab = (tab, targetScheduleId = null) => {
  // Menu yang tidak diizinkan dialihkan ke menu pertama yang boleh diakses
  if (!authStore.canAccessTab(tab)) {
    tab = landingTab(authStore.user)
  }
  // Menu pelaksanaan ujian baru dapat dibuka setelah event dipilih
  if (EVENT_SCOPED_TABS.includes(tab) && !currentScopedEventId.value) {
    tab = 'events'
  }
  activeTab.value = tab
  isMobileSidebarOpen.value = false

  if (tab === 'events') {
    loadEvents()
  } else if (tab === 'schedules') {
    loadSchedules()
  } else if (tab === 'students') {
    api.get('/admin/students').then(res => students.value = res.data.data || [])
  } else if (tab === 'teachers') {
    api.get('/admin/teachers').then(res => teachers.value = res.data.data || [])
    // Katalog izin dipakai untuk label peran turunan dan editor akses
    loadPermissionCatalog().catch(() => {})
  } else if (tab === 'classes') {
    api.get('/admin/classes').then(res => classes.value = res.data.data || [])
  } else if (tab === 'subjects') {
    loadSubjects()
  } else if (tab === 'class-subjects') {
    loadClassSubjects()
  } else if (tab === 'proctor') {
    if (targetScheduleId) {
      activeProctorScheduleId.value = targetScheduleId
    }
    // Polling data live dilakukan oleh LiveProctorControl (mengirim data lewat @data-refreshed).
    loadProctorSchedules()
  } else if (tab === 'questions') {
    loadAllData()
    if (readinessData.value.question_banks?.length > 0 && !selectedBankUploadId.value) {
      selectedBankUploadId.value = readinessData.value.question_banks[0].id
    }
  }
}


// Izin kelola jadwal (hanya kerapian UI; server tetap penjaga sebenarnya)
const canManageSchedules = computed(() => authStore.hasPermission('schedules:manage'))

// ---- Beranda: tampilan menurut izin (hanya kerapian UI; server tetap penjaga sebenarnya) ----
const canManageMaster = computed(() => authStore.hasPermission('master:manage'))
const canManageEvents = computed(() => authStore.hasPermission('events:manage'))
const canUploadQuestionBank = computed(() => authStore.hasPermission('questions:upload'))
// Daftar akun siswa (GET /admin/users) dan reset sesi (POST /admin/users/:id/reset-session) butuh izin ini
const canManageLoginSessions = computed(() => authStore.hasAnyPermission(['users:manage', 'master:manage']))
const canImportStudents = canManageLoginSessions
// Daftar siswa dan akun staf (GET /admin/students, /admin/teachers) hanya diberikan server kepada
// pengelola master, akun, atau jadwal; tanpa itu students kosong dan daftar hadir cetak tidak berguna.
const canReadPeople = computed(() => authStore.hasAnyPermission(['master:manage', 'users:manage', 'schedules:manage']))

// Staf tanpa izin kelola apa pun (guru/pengawas biasa) mendapat kartu sambutan sederhana.
// Dalam mode ini kartu Pintasan Cepat disembunyikan karena satu-satunya tile yang mungkin
// (Unggah Bank Soal) sudah tersedia lewat tombol "Bank Soal" pada kartu sambutan.
const showWelcomeCard = computed(() =>
  !(canManageMaster.value || canManageSchedules.value || canManageLoginSessions.value || canManageEvents.value)
)
const showQuickActions = computed(() =>
  !showWelcomeCard.value &&
  (canManageSchedules.value || canImportStudents.value || canUploadQuestionBank.value || canManageEvents.value)
)

const showEventReadiness = ref(true)

const eventReadinessStatus = (event) => {
  if (!event.total_schedules || event.total_schedules === 0) return 'none'
  const isLoaded =
    selectedExamEvent.value?.id === event.id ||
    (!selectedExamEvent.value && activeExamEvent.value?.id === event.id)
  if (isLoaded) {
    const allLinkedAndLocked =
      schedules.value.length > 0 &&
      schedules.value.every((s) => s.bank_id && s.bank && s.bank.is_locked)
    if (allLinkedAndLocked) return 'ready'
    return 'needs_action'
  }
  return 'running'
}

const eventStatusLabel = (event) => {
  const s = eventReadinessStatus(event)
  if (s === 'ready') return 'Siap'
  if (s === 'needs_action') return 'Perlu tindakan'
  if (s === 'running') return 'Berjalan'
  return 'Belum mulai'
}

const eventStatusBadgeClass = (event) => {
  const s = eventReadinessStatus(event)
  if (s === 'ready') return 'bg-emerald-100 text-emerald-700'
  if (s === 'needs_action') return 'bg-amber-100 text-amber-700'
  if (s === 'running') return 'bg-indigo-100 text-indigo-700'
  return 'bg-slate-100 text-slate-500'
}

const eventActionLabel = (event) => {
  const s = eventReadinessStatus(event)
  if (s === 'ready') return 'Live proctor →'
  if (s === 'needs_action') return 'Kunci soal →'
  if (s === 'running') return 'Kelola →'
  return 'Atur event →'
}

const eventActionBtnClass = (event) => {
  const s = eventReadinessStatus(event)
  if (s === 'ready') return 'bg-emerald-50 text-emerald-700 border border-emerald-200 hover:bg-emerald-100'
  if (s === 'needs_action') return 'bg-indigo-600 text-white hover:bg-indigo-700'
  return 'bg-slate-100 text-slate-700 hover:bg-slate-200 border border-slate-200'
}

const eventActionClick = (event) => {
  const s = eventReadinessStatus(event)
  enterEvent(event)
  if (s === 'ready') switchTab('proctor')
  else if (s === 'needs_action') switchTab('questions')
  else if (s === 'running') switchTab('schedules')
  else switchTab('events')
}

// Sesi login siswa aktif. Server menandai akun dengan has_active_session (token sesi terisi);
// token itu sendiri tidak pernah dikirim ke klien (json:"-"), jadi datanya diambil dari daftar akun.
const loginSessionAccounts = ref([])
const activeLoginSessions = computed(() => loginSessionAccounts.value.filter(u => u.has_active_session))
const loadLoginSessions = async () => {
  if (!canManageLoginSessions.value) {
    loginSessionAccounts.value = []
    return
  }
  try {
    const res = await api.get('/admin/users', { params: { role: 'SISWA' } })
    loginSessionAccounts.value = res.data.data || []
  } catch (e) {
    console.error('Failed to load student login sessions', e)
  }
}
// Segarkan setiap kali Beranda dibuka (pemuatan awal dilakukan di onMounted)
watch(
  () => activeTab.value === 'dashboard' && canManageLoginSessions.value,
  (visible) => { if (visible) loadLoginSessions() }
)

// Live Proctoring States
const proctorSchedules = ref([])
const activeProctorScheduleId = ref(null)
const proctorData = ref(null)
const isExportingNilai = ref(false)
const showProctorPrint = ref(false)
const isExportingPDF = ref(false)

const inProgressCount = computed(() => {
  if (!proctorData.value?.students) return 0
  return proctorData.value.students.filter(s => s.status === 'IN_PROGRESS').length
})

const getProctorStatusBadge = (status) => {
  switch (status) {
    case 'SUBMITTED':
      return 'bg-emerald-100 text-emerald-800'
    case 'IN_PROGRESS':
      return 'bg-blue-100 text-blue-800 animate-pulse'
    case 'BLOCKED':
      return 'bg-rose-100 text-rose-800 border border-rose-300'
    default:
      return 'bg-slate-100 text-slate-600'
  }
}

const getProctorStatusLabel = (status) => {
  switch (status) {
    case 'SUBMITTED':
      return 'Selesai'
    case 'IN_PROGRESS':
      return 'Mengerjakan'
    case 'BLOCKED':
      return 'Terkunci'
    default:
      return 'Belum Mulai'
  }
}

const loadProctorSchedules = async () => {
  try {
    const res = await api.get('/proctor/schedules')
    proctorSchedules.value = res.data.data || []
    if (proctorSchedules.value.length > 0 && !activeProctorScheduleId.value) {
      // Utamakan jadwal yang dapat dikendalikan pengguna (can_control tidak ada = dianggap true)
      const firstControllable = proctorSchedules.value.find(s => s.can_control !== false)
      activeProctorScheduleId.value = (firstControllable || proctorSchedules.value[0]).id
    }
  } catch (err) {
    console.error('Failed to load proctor schedules', err)
  }
}

// ---- Beranda guru / pengawas: ringkasan tugas milik pengguna yang sedang login ----
// Sumber data: jadwal (GET /admin/schedules, sudah dibatasi server), bank soal buatan sendiri
// (GET /admin/readiness-matrix), dan jadwal pengawasan (GET /proctor/schedules).
const canShowTeacherQuestionSection = computed(() => authStore.canAccessTab('questions'))
const canShowTeacherProctorSection = computed(() => authStore.canAccessTab('proctor'))
const canLockQuestionBanks = computed(() => authStore.hasPermission('questions:lock'))
const canPrintQuestionBanks = computed(() => authStore.hasPermission('questions:print'))

// Cetak naskah soal & kunci jawaban (hanya bank yang sudah terkunci).
const showQuestionPrint = ref(false)
const questionPrintBank = ref(null)
const openQuestionPrint = (bank) => {
  questionPrintBank.value = bank
  showQuestionPrint.value = true
}

const isTodayDate = (value) => !!value && new Date(value).toDateString() === new Date().toDateString()
const scheduleStartTimestamp = (sch) => {
  const t = sch?.start_time ? new Date(sch.start_time).getTime() : 0
  return Number.isNaN(t) ? 0 : t
}
const scheduleEndTimestamp = (sch) => {
  const t = new Date(sch?.end_time || sch?.start_time || 0).getTime()
  return Number.isNaN(t) ? 0 : t
}
const isScheduleNotFinished = (sch) => scheduleEndTimestamp(sch) >= Date.now()
const sortByScheduleStart = (list) => [...list].sort((a, b) => scheduleStartTimestamp(a) - scheduleStartTimestamp(b))

// Relasi pengguna dengan jadwal (GET /admin/schedules -> relations: "mengampu" | "bank_saya" | "pengawas").
// Backend lama belum mengirim properti relations sama sekali; dalam kasus itu semua jadwal yang diterima
// dianggap milik pengguna (perilaku lama) agar Beranda tidak menampilkan angka nol yang keliru.
const teacherHomeLegacyRelations = computed(() =>
  !schedules.value.some(s => s && Object.prototype.hasOwnProperty.call(s, 'relations'))
)
const teacherHomeRelationsOf = (sch) => (Array.isArray(sch?.relations) ? sch.relations : [])
const teacherHomeIsMine = (sch) => teacherHomeLegacyRelations.value || teacherHomeRelationsOf(sch).length > 0
const teacherHomeIsTeaching = (sch) => teacherHomeLegacyRelations.value || teacherHomeRelationsOf(sch).includes('mengampu')
const teacherHomeIsInactive = (sch) => sch?.is_active === false
// Peran sebagai teks netral; kosong pada backend lama (peran tidak diketahui)
const teacherHomeRoleLabel = (sch) => {
  const rel = teacherHomeRelationsOf(sch)
  const roles = []
  if (rel.includes('mengampu')) roles.push('Pengampu')
  if (rel.includes('pengawas')) roles.push('Pengawas')
  if (roles.length > 0) return roles.join(' · ')
  return rel.includes('bank_saya') ? 'Bank saya' : ''
}
// Jadwal yang menjadi tugas membuat bank soal: diampu pengguna, tanpa bank, aktif
const teacherHomeNeedsBank = (sch) =>
  teacherHomeIsTeaching(sch) && (!sch.bank_id || !sch.bank) && !teacherHomeIsInactive(sch)

// Jadwal milik pengguna (sudah difilter event bila ada event terpilih)
const teacherHomeSchedules = computed(() => currentEventSchedules.value.filter(teacherHomeIsMine))
const teacherHomeTodayScheduleCount = computed(() =>
  teacherHomeSchedules.value.filter(s => isTodayDate(s.start_time)).length
)
const teacherHomeUpcomingSchedules = computed(() =>
  sortByScheduleStart(teacherHomeSchedules.value.filter(isScheduleNotFinished))
)
const teacherHomeUpcomingList = computed(() => teacherHomeUpcomingSchedules.value.slice(0, 5))
const teacherHomeSchedulesWithoutBank = computed(() =>
  teacherHomeUpcomingSchedules.value.filter(teacherHomeNeedsBank)
)

// Bank soal buatan pengguna (server sudah membatasi; filter created_by_id sebagai pengaman bila pengguna berizin baca semua)
const teacherHomeBanks = computed(() => {
  const banks = readinessData.value.question_banks || []
  const myId = authStore.user?.id
  return myId ? banks.filter(b => !b.created_by_id || b.created_by_id === myId) : banks
})
const teacherHomeLockedBankCount = computed(() => teacherHomeBanks.value.filter(b => b.is_locked).length)
const teacherHomeDraftBankCount = computed(() => teacherHomeBanks.value.length - teacherHomeLockedBankCount.value)
const teacherHomeUnfinishedBanks = computed(() =>
  teacherHomeBanks.value.filter(b => !(b.total_questions > 0) || !b.is_locked)
)

// Daftar "Perlu Tindakan": jadwal tanpa bank soal + bank soal kosong/belum dikunci,
// diurutkan menurut jadwal terdekat yang memakainya (tanpa jadwal terkait di urutan akhir).
const TEACHER_HOME_ACTION_LIMIT = 6
const teacherHomeActionItems = computed(() => {
  if (!canShowTeacherQuestionSection.value) return []
  const upcoming = teacherHomeUpcomingSchedules.value
  const items = []

  teacherHomeSchedulesWithoutBank.value.forEach(sch => {
    items.push({
      key: `schedule-${sch.id}`,
      kind: 'no-bank',
      title: sch.title,
      label: 'Belum ada bank soal',
      detail: `${sch.class_room?.name || '-'} · ${formatScheduleTimeRange(sch.start_time, sch.end_time)}`,
      actionLabel: 'Buat Bank Soal',
      sortTime: scheduleStartTimestamp(sch)
    })
  })

  teacherHomeUnfinishedBanks.value.forEach(bank => {
    const linkedStarts = upcoming.filter(s => s.bank_id === bank.id).map(scheduleStartTimestamp)
    const isEmpty = !(bank.total_questions > 0)
    items.push({
      key: `bank-${bank.id}`,
      kind: isEmpty ? 'empty-bank' : 'unlocked-bank',
      bank,
      title: bank.title,
      label: isEmpty ? 'Belum ada butir soal' : 'Belum dikunci',
      detail: [bank.subject?.name, isEmpty ? '' : `${bank.total_questions} soal`].filter(Boolean).join(' · '),
      actionLabel: isEmpty ? 'Isi Soal' : (canLockQuestionBanks.value ? 'Kunci' : 'Lihat'),
      sortTime: linkedStarts.length > 0 ? Math.min(...linkedStarts) : Number.MAX_SAFE_INTEGER
    })
  })

  return items.sort((a, b) => a.sortTime - b.sortTime)
})
const teacherHomeVisibleActionItems = computed(() => teacherHomeActionItems.value.slice(0, TEACHER_HOME_ACTION_LIMIT))

const runTeacherHomeAction = (item) => {
  if (item.kind === 'unlocked-bank' && canLockQuestionBanks.value) {
    toggleBankLockWithConfirm(item.bank)
    return
  }
  switchTab('questions')
}

const teacherHomeScheduleBankStatus = (sch) => {
  if (!sch.bank_id || !sch.bank) {
    // Amber hanya bila memang tugas pengguna; jadwal pengawas saja / nonaktif tampil netral
    return { label: 'Belum ada bank soal', textClass: teacherHomeNeedsBank(sch) ? 'text-amber-600' : 'text-slate-500' }
  }
  if (sch.bank.is_locked) return { label: 'Siap', textClass: 'text-emerald-600' }
  return { label: 'Draf', textClass: 'text-slate-500' }
}

// Tugas mengawasi: hanya jadwal yang boleh dikendalikan (can_control), dibatasi event terpilih
const teacherHomeProctorTasks = computed(() => {
  if (!canShowTeacherProctorSection.value) return []
  const evId = currentScopedEventId.value
  return proctorSchedules.value.filter(s => s.can_control === true && (!evId || s.event_id === evId))
})
const teacherHomeProctorTodayCount = computed(() =>
  teacherHomeProctorTasks.value.filter(s => isTodayDate(s.start_time)).length
)
const teacherHomeProctorUpcomingList = computed(() =>
  sortByScheduleStart(teacherHomeProctorTasks.value.filter(isScheduleNotFinished)).slice(0, 5)
)

// Tata letak kartu angka: 1, 2, 4, atau 5 kartu tergantung izin
const teacherHomeStatCount = computed(() =>
  1 + (canShowTeacherQuestionSection.value ? 3 : 0) + (canShowTeacherProctorSection.value ? 1 : 0)
)
const teacherHomeGridClass = computed(() => {
  if (teacherHomeStatCount.value === 5) return 'grid-cols-2 xl:grid-cols-5'
  if (teacherHomeStatCount.value === 4) return 'grid-cols-2 sm:grid-cols-4'
  return 'grid-cols-2'
})
const teacherHomeProctorCardSpanClass = computed(() =>
  teacherHomeStatCount.value === 5 ? 'col-span-2 xl:col-span-1' : ''
)

// Jadwal pengawasan hanya dimuat untuk guru/pengawas yang membuka Beranda
const loadTeacherHomeProctorTasks = () => {
  if (showWelcomeCard.value && canShowTeacherProctorSection.value) loadProctorSchedules()
}
watch(
  () => activeTab.value === 'dashboard' && showWelcomeCard.value && canShowTeacherProctorSection.value,
  (visible) => { if (visible) loadProctorSchedules() }
)

// Galat dari request berjenis blob datang sebagai Blob; baca isinya dengan aman.
const extractExportErrorMessage = async (err, fallback) => {
  try {
    const data = err?.response?.data
    if (data instanceof Blob) {
      const parsed = JSON.parse(await data.text())
      return parsed?.message || parsed?.error || fallback
    }
    return data?.message || fallback
  } catch {
    return fallback
  }
}

const exportNilaiExcel = async () => {
  if (!activeProctorScheduleId.value || isExportingNilai.value) return
  isExportingNilai.value = true
  try {
    const res = await api.get(`/proctor/reports/excel/${activeProctorScheduleId.value}`, {
      responseType: 'blob'
    })
    const url = window.URL.createObjectURL(new Blob([res.data]))
    const link = document.createElement('a')
    link.href = url
    const className = proctorData.value?.class_name || 'Kelas'
    link.setAttribute('download', `Rekap_Nilai_${className}.xlsx`)
    document.body.appendChild(link)
    link.click()
    link.remove()
    window.URL.revokeObjectURL(url)
  } catch (err) {
    await showAlertModal({
      title: 'Gagal Mengunduh Laporan',
      message: await extractExportErrorMessage(err, 'Gagal mengunduh laporan'),
      type: 'danger'
    })
  } finally {
    isExportingNilai.value = false
  }
}

const exportBeritaAcaraPDF = async () => {
  if (!activeProctorScheduleId.value || isExportingPDF.value) return
  isExportingPDF.value = true
  try {
    const res = await api.get(`/proctor/reports/pdf/${activeProctorScheduleId.value}`, {
      responseType: 'blob'
    })
    const url = window.URL.createObjectURL(new Blob([res.data], { type: 'application/pdf' }))
    const link = document.createElement('a')
    link.href = url
    const className = proctorData.value?.class_name || 'Kelas'
    link.setAttribute('download', `Berita_Acara_${className}.pdf`)
    document.body.appendChild(link)
    link.click()
    link.remove()
    window.URL.revokeObjectURL(url)
  } catch (err) {
    await showAlertModal({
      title: 'Gagal Mengunduh Laporan',
      message: await extractExportErrorMessage(err, 'Gagal mengunduh laporan'),
      type: 'danger'
    })
  } finally {
    isExportingPDF.value = false
  }
}

// Bank Soal States & Logic (Unified Bank Soal & Kesiapan)
const questionBankFile = ref(null)
const selectedBankUploadId = ref('')
const isUploadingQuestionBank = ref(false)
const questionBankUploadStatus = ref(null)
const showUploadBankModal = ref(false)
const showCreateBankModal = ref(false)
const isCreatingBank = ref(false)
const isEditBank = ref(false)
const editingBankId = ref(null)
const questionBankSearchQuery = ref('')
const questionBankStatusFilter = ref('all') // 'all', 'locked', 'draft'
const newBankForm = ref({
  scope: 'grade', // 'grade' = per angkatan (default), 'classes' = kelas tertentu
  subject_id: '',
  grade: '',
  class_ids: [],
  title: ''
})

const bankScopeOptions = [
  { value: 'grade', label: 'Per Angkatan' },
  { value: 'classes', label: 'Pilih Kelas' }
]

const bankGradeOrder = { X: 1, XI: 2, XII: 3 }

// Kunci opsi dropdown = mapel + tingkat, karena Informatika XI dan XII butuh naskah berbeda.
const bankOptionKey = (subjectId, grade) => `${subjectId}|${grade || ''}`

const newBankOptionKey = computed({
  get: () => newBankForm.value.subject_id ? bankOptionKey(newBankForm.value.subject_id, newBankForm.value.grade) : '',
  set: (key) => {
    const [subjectId, grade = ''] = (key || '').split('|')
    newBankForm.value.subject_id = subjectId || ''
    newBankForm.value.grade = grade
  }
})

const scheduleClassIdOf = (s) => s.class_room?.id || s.class_room_id || s.class_id
const scheduleGradeOf = (s) =>
  s.class_room?.grade || classes.value.find(c => c.id === scheduleClassIdOf(s))?.grade || ''

const scheduleSubjectOf = (s) => {
  const subId = s.subject_id || s.subject?.id || s.bank?.subject_id || s.bank?.subject?.id
  if (subId) return s.subject || s.bank?.subject || subjects.value.find(sub => sub.id === subId) || null
  if (s.title) return subjects.value.find(sub => s.title.toLowerCase().includes(sub.name.toLowerCase())) || null
  return null
}

const bankScopeLabel = (b) => {
  if (b.classes?.length) return b.classes.map(c => c.name).sort((x, y) => x.localeCompare(y, undefined, { numeric: true })).join(', ')
  return b.grade ? `Kelas ${b.grade}` : ''
}

// Kelas yang sudah tercakup bank soal lain, per mapel: subject_id -> Set(class_id)
const bankCoveredClassIds = computed(() => {
  const scheds = currentEventSchedules.value || []
  const covered = new Map()
  ;(readinessData.value.question_banks || [])
    .filter(b => isEditBank.value ? b.id !== editingBankId.value : true)
    .forEach(b => {
      const subId = b.subject_id || b.subject?.id
      if (!subId) return
      if (!covered.has(subId)) covered.set(subId, new Set())
      const set = covered.get(subId)
      if (b.classes?.length) {
        b.classes.forEach(c => set.add(c.id))
      } else if (b.grade) {
        classes.value.filter(c => c.grade === b.grade).forEach(c => set.add(c.id))
      } else {
        // Bank lama tanpa cakupan: anggap mencakup kelas jadwal yang sudah ditautkan kepadanya.
        scheds.filter(s => s.bank_id === b.id).forEach(s => set.add(scheduleClassIdOf(s)))
      }
    })
  return covered
})

const availableBankSubjects = computed(() => {
  const scheds = currentEventSchedules.value || []
  const covered = bankCoveredClassIds.value

  const map = new Map()
  const addOption = (sub, grade) => {
    const key = bankOptionKey(sub.id, grade)
    if (!map.has(key)) map.set(key, { key, id: sub.id, name: sub.name, code: sub.code, grade })
  }

  if (scheds.length > 0) {
    // Opsi angkatan muncul selama masih ada kelas terjadwal di angkatan itu yang belum punya bank soal.
    scheds.forEach(s => {
      const sub = scheduleSubjectOf(s)
      if (sub && !covered.get(sub.id)?.has(scheduleClassIdOf(s))) addOption(sub, scheduleGradeOf(s))
    })
  } else {
    (subjects.value || []).filter(sub => !covered.has(sub.id)).forEach(sub => addOption(sub, ''))
  }

  // Saat edit, pastikan pilihan bank yang sedang disunting tetap ada
  if (isEditBank.value && editingBankId.value) {
    const currentBank = (readinessData.value.question_banks || []).find(b => b.id === editingBankId.value)
    if (currentBank?.subject && !currentBank.classes?.length) addOption(currentBank.subject, currentBank.grade || '')
  }

  return Array.from(map.values())
    .sort((a, b) => a.name.localeCompare(b.name) || (bankGradeOrder[a.grade] || 9) - (bankGradeOrder[b.grade] || 9))
})

// Mode "Pilih Kelas": mapel dari jadwal event (atau semua mapel bila belum ada jadwal)
const bankCustomSubjects = computed(() => {
  const map = new Map()
  ;(currentEventSchedules.value || []).forEach(s => {
    const sub = scheduleSubjectOf(s)
    if (sub) map.set(sub.id, sub)
  })
  const currentBank = isEditBank.value && (readinessData.value.question_banks || []).find(b => b.id === editingBankId.value)
  if (currentBank?.subject) map.set(currentBank.subject.id, currentBank.subject)
  const list = map.size > 0 ? Array.from(map.values()) : (subjects.value || [])
  return [...list].sort((a, b) => a.name.localeCompare(b.name))
})

const bankCustomClassOptions = computed(() => {
  const subId = newBankForm.value.subject_id
  if (!subId) return []
  const ids = new Set(
    (currentEventSchedules.value || [])
      .filter(s => scheduleSubjectOf(s)?.id === subId)
      .map(scheduleClassIdOf)
  )
  // Pertahankan kelas yang sudah dipilih (mis. saat edit) meski tidak ada di jadwal
  newBankForm.value.class_ids.forEach(id => ids.add(id))
  const list = ids.size > 0 ? classes.value.filter(c => ids.has(c.id)) : classes.value
  const covered = bankCoveredClassIds.value.get(subId) || new Set()
  return list
    .map(c => ({ id: c.id, name: c.name, grade: c.grade, covered: covered.has(c.id) && !newBankForm.value.class_ids.includes(c.id) }))
    .sort((a, b) => (bankGradeOrder[a.grade] || 9) - (bankGradeOrder[b.grade] || 9) || a.name.localeCompare(b.name, undefined, { numeric: true }))
})

const bankFormLocked = computed(() =>
  !isEditBank.value && newBankForm.value.scope === 'grade' && availableBankSubjects.value.length === 0
)

const bankEventPrefix = () => selectedExamEvent.value?.title ? `${selectedExamEvent.value.title} - ` : ''

const bankTitleFor = (opt) => `${bankEventPrefix()}${opt.name}${opt.grade ? ` Kelas ${opt.grade}` : ''}`

const refreshCustomBankTitle = () => {
  const sub = bankCustomSubjects.value.find(s => s.id === newBankForm.value.subject_id)
  if (!sub) return
  const names = bankCustomClassOptions.value.filter(c => newBankForm.value.class_ids.includes(c.id)).map(c => c.name)
  newBankForm.value.title = `${bankEventPrefix()}${sub.name}${names.length ? ` ${names.join(', ')}` : ''}`
}

const onBankSubjectChange = () => {
  const selOpt = availableBankSubjects.value.find(o => o.key === newBankOptionKey.value)
  if (selOpt) newBankForm.value.title = bankTitleFor(selOpt)
}

const onBankCustomSubjectChange = () => {
  newBankForm.value.class_ids = []
  refreshCustomBankTitle()
}

const toggleBankClass = (id) => {
  const ids = newBankForm.value.class_ids
  newBankForm.value.class_ids = ids.includes(id) ? ids.filter(x => x !== id) : [...ids, id]
  refreshCustomBankTitle()
}

const setBankScope = (scope) => {
  if (newBankForm.value.scope === scope) return
  newBankForm.value.scope = scope
  newBankForm.value.class_ids = []
  if (scope === 'grade') {
    const opt = availableBankSubjects.value.find(o => o.id === newBankForm.value.subject_id) || availableBankSubjects.value[0]
    newBankForm.value.subject_id = opt?.id || ''
    newBankForm.value.grade = opt?.grade || ''
    if (opt) newBankForm.value.title = bankTitleFor(opt)
  } else {
    newBankForm.value.grade = ''
    refreshCustomBankTitle()
  }
}

const questionBankSortKey = ref('title') // 'title', 'creator', 'questions', 'status'
const questionBankSortOrder = ref('asc') // 'asc' | 'desc'
const questionBankCurrentPage = ref(1)
const questionBankPerPage = ref(10)

const toggleQuestionBankSort = (key) => {
  if (questionBankSortKey.value === key) {
    questionBankSortOrder.value = questionBankSortOrder.value === 'asc' ? 'desc' : 'asc'
  } else {
    questionBankSortKey.value = key
    questionBankSortOrder.value = 'asc'
  }
}

const filteredQuestionBanks = computed(() => {
  let list = readinessData.value.question_banks || []
  if (questionBankStatusFilter.value === 'locked') {
    list = list.filter(b => b.is_locked)
  } else if (questionBankStatusFilter.value === 'draft') {
    list = list.filter(b => !b.is_locked)
  }

  if (questionBankSearchQuery.value.trim()) {
    const q = questionBankSearchQuery.value.toLowerCase().trim()
    list = list.filter(b => {
      const title = (b.title || '').toLowerCase()
      const subjectName = (b.subject?.name || '').toLowerCase()
      const subjectCode = (b.subject?.code || '').toLowerCase()
      const creatorName = (b.created_by?.full_name || '').toLowerCase()
      return title.includes(q) || subjectName.includes(q) || subjectCode.includes(q) || creatorName.includes(q)
    })
  }

  // Sort
  if (questionBankSortKey.value) {
    list = [...list].sort((a, b) => {
      let result = 0
      if (questionBankSortKey.value === 'title') {
        const titleA = (a.title || '').trim().toLowerCase()
        const titleB = (b.title || '').trim().toLowerCase()
        result = titleA.localeCompare(titleB, undefined, { numeric: true, sensitivity: 'base' })
        if (result === 0) {
          const subA = (a.subject?.name || '').trim().toLowerCase()
          const subB = (b.subject?.name || '').trim().toLowerCase()
          result = subA.localeCompare(subB, undefined, { numeric: true, sensitivity: 'base' })
        }
      } else if (questionBankSortKey.value === 'creator') {
        const nameA = (a.created_by?.full_name || '').trim().toLowerCase()
        const nameB = (b.created_by?.full_name || '').trim().toLowerCase()
        result = nameA.localeCompare(nameB, undefined, { numeric: true, sensitivity: 'base' })
      } else if (questionBankSortKey.value === 'questions') {
        const countA = a.total_questions || 0
        const countB = b.total_questions || 0
        result = countA - countB
      } else if (questionBankSortKey.value === 'status') {
        const lockA = a.is_locked ? 1 : 0
        const lockB = b.is_locked ? 1 : 0
        result = lockA - lockB
      }
      return questionBankSortOrder.value === 'asc' ? result : -result
    })
  }

  return list
})

const totalQuestionBankPages = computed(() => {
  return Math.max(1, Math.ceil(filteredQuestionBanks.value.length / questionBankPerPage.value))
})

const paginatedQuestionBanks = computed(() => {
  const start = (questionBankCurrentPage.value - 1) * questionBankPerPage.value
  return filteredQuestionBanks.value.slice(start, start + questionBankPerPage.value)
})

const displayedQuestionBankPages = computed(() => {
  const total = totalQuestionBankPages.value
  const current = questionBankCurrentPage.value
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

const setQuestionBankPage = (p) => {
  if (p === '...' || p < 1 || p > totalQuestionBankPages.value) return
  questionBankCurrentPage.value = p
}

const downloadQuestionBankTemplate = async () => {
  try {
    const res = await api.get('/admin/template/question-bank.xlsx', {
      responseType: 'blob'
    })
    const url = window.URL.createObjectURL(new Blob([res.data], {
      type: 'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet'
    }))
    const link = document.createElement('a')
    link.href = url
    link.setAttribute('download', 'Template_Bank_Soal_CBT.xlsx')
    document.body.appendChild(link)
    link.click()
    link.remove()
    window.URL.revokeObjectURL(url)
    showToast('Format template Excel berhasil diunduh!', 'success')
  } catch (err) {
    await showAlertModal({
      title: 'Gagal Mengunduh Template',
      message: await extractExportErrorMessage(err, 'Gagal mengunduh template'),
      type: 'danger'
    })
  }
}

const handleQuestionBankFileChange = (e) => {
  questionBankFile.value = e.target.files[0]
}

const openUploadBankModal = (bank = null) => {
  if (bank) {
    selectedBankUploadId.value = bank.id
  } else if (!selectedBankUploadId.value && (readinessData.value.question_banks || []).length > 0) {
    selectedBankUploadId.value = readinessData.value.question_banks[0].id
  }
  questionBankFile.value = null
  questionBankUploadStatus.value = null
  showUploadBankModal.value = true
}

const openCreateBankModal = () => {
  isEditBank.value = false
  editingBankId.value = null
  const defaultOpt = availableBankSubjects.value[0]
  newBankForm.value = {
    scope: 'grade',
    subject_id: defaultOpt?.id || '',
    grade: defaultOpt?.grade || '',
    class_ids: [],
    title: defaultOpt ? bankTitleFor(defaultOpt) : ''
  }
  showCreateBankModal.value = true
}

const openEditBankModal = (bank) => {
  isEditBank.value = true
  editingBankId.value = bank.id
  newBankForm.value = {
    scope: bank.classes?.length ? 'classes' : 'grade',
    subject_id: bank.subject_id || bank.subject?.id || '',
    grade: bank.grade || '',
    class_ids: (bank.classes || []).map(c => c.id),
    title: bank.title || ''
  }
  showCreateBankModal.value = true
}

const submitCreateBank = async () => {
  if (!newBankForm.value.subject_id || !newBankForm.value.title.trim()) {
    await showAlertModal({
      title: 'Form Belum Lengkap',
      message: 'Mata pelajaran dan judul bank soal wajib diisi.',
      type: 'warning'
    })
    return
  }
  const isCustomScope = newBankForm.value.scope === 'classes'
  if (isCustomScope && newBankForm.value.class_ids.length === 0) {
    await showAlertModal({
      title: 'Kelas Belum Dipilih',
      message: 'Pilih minimal satu kelas untuk bank soal dengan cakupan kelas tertentu.',
      type: 'warning'
    })
    return
  }
  const scopePayload = {
    grade: isCustomScope ? '' : newBankForm.value.grade,
    class_ids: isCustomScope ? newBankForm.value.class_ids : []
  }

  isCreatingBank.value = true
  try {
    if (isEditBank.value && editingBankId.value) {
      await api.put(`/admin/question-banks/${editingBankId.value}`, {
        subject_id: newBankForm.value.subject_id,
        ...scopePayload,
        title: newBankForm.value.title.trim()
      })
      showToast('Bank soal berhasil diperbarui!', 'success')
    } else {
      const res = await api.post('/admin/question-banks', {
        subject_id: newBankForm.value.subject_id,
        ...scopePayload,
        title: newBankForm.value.title.trim()
      })
      showToast('Bank soal baru berhasil dibuat!', 'success')
      if (res.data.data?.id) {
        selectedBankUploadId.value = res.data.data.id
      }
    }
    showCreateBankModal.value = false
    await loadAllData()
  } catch (err) {
    await showAlertModal({
      title: isEditBank.value ? 'Gagal Memperbarui Bank Soal' : 'Gagal Membuat Bank Soal',
      message: err.response?.data?.message || 'Terjadi kesalahan saat menyimpan bank soal.',
      type: 'danger'
    })
  } finally {
    isCreatingBank.value = false
  }
}

const deleteQuestionBank = async (bank) => {
  const confirmed = await showConfirmModal({
    title: 'Hapus Bank Soal',
    message: `Apakah Anda yakin ingin menghapus bank soal "${bank.title}" (${bank.subject?.name || ''}) beserta seluruh butir soalnya?`,
    type: 'danger',
    confirmText: 'Hapus Bank Soal',
    cancelText: 'Batal'
  })
  if (!confirmed) return

  try {
    await api.delete(`/admin/question-banks/${bank.id}`)
    showToast('Bank soal berhasil dihapus!', 'success')
    await loadAllData()
  } catch (err) {
    await showAlertModal({
      title: 'Gagal Menghapus Bank Soal',
      message: err.response?.data?.message || 'Terjadi kesalahan saat menghapus bank soal.',
      type: 'danger'
    })
  }
}

const toggleBankLockWithConfirm = async (bank) => {
  const willLock = !bank.is_locked
  const actionTitle = willLock ? 'Kunci Naskah Bank Soal' : 'Buka Kunci Naskah Bank Soal'
  const actionMessage = willLock
    ? `Kunci naskah "${bank.title}"? Setelah dikunci, naskah dinyatakan Siap Ujian dan butir soal siap disajikan kepada peserta ujian.`
    : `Buka kunci naskah "${bank.title}"? Status naskah akan kembali menjadi Draft Terbuka.`

  const confirmed = await showConfirmModal({
    title: actionTitle,
    message: actionMessage,
    type: willLock ? 'confirm' : 'warning',
    confirmText: willLock ? 'Kunci Naskah' : 'Buka Kunci',
    cancelText: 'Batal'
  })
  if (!confirmed) return

  try {
    await api.post(`/admin/question-banks/${bank.id}/toggle-lock`)
    showToast(willLock ? 'Naskah soal berhasil dikunci (Terkunci)!' : 'Kunci naskah soal berhasil dibuka (Draft)!', 'success')
    await loadAllData()
  } catch (err) {
    await showAlertModal({
      title: 'Gagal Mengubah Status Kunci',
      message: err.response?.data?.message || 'Terjadi kesalahan saat mengubah status penguncian bank soal.',
      type: 'danger'
    })
  }
}

const submitQuestionBankUpload = async () => {
  if (!selectedBankUploadId.value) {
    await showAlertModal({
      title: 'Pilih Bank Soal',
      message: 'Pilih bank soal tujuan terlebih dahulu.',
      type: 'warning'
    })
    return
  }
  if (!questionBankFile.value) {
    await showAlertModal({
      title: 'Pilih File Soal',
      message: 'Pilih file Excel (.xlsx) terlebih dahulu.',
      type: 'warning'
    })
    return
  }
  isUploadingQuestionBank.value = true
  questionBankUploadStatus.value = null
  const formData = new FormData()
  formData.append('bank_id', selectedBankUploadId.value)
  formData.append('file', questionBankFile.value)
  try {
    const res = await api.post('/admin/questions/import', formData, {
      headers: { 'Content-Type': 'multipart/form-data' }
    })
    questionBankUploadStatus.value = {
      success: true,
      message: res.data.message || 'Soal berhasil diimpor ke bank soal!'
    }
    showToast('Soal berhasil diimpor ke bank soal!', 'success')
    questionBankFile.value = null
    showUploadBankModal.value = false
    await loadAllData()
  } catch (err) {
    questionBankUploadStatus.value = {
      success: false,
      message: err.response?.data?.message || 'Gagal mengimpor file Excel bank soal.'
    }
    await showAlertModal({
      title: 'Gagal Impor Soal',
      message: err.response?.data?.message || 'Gagal mengimpor file Excel bank soal.',
      type: 'danger'
    })
  } finally {
    isUploadingQuestionBank.value = false
  }
}

// ---------------- QUESTION VIEWER & EDITOR ----------------
const showBankQuestionsModal = ref(false)
const showSimulatorModal = ref(false)
const viewingBank = ref(null)
const bankQuestions = ref([])
const isLoadingBankQuestions = ref(false)
const editingQuestionId = ref(null) // null = none, 'new' = add new question, uuid = edit question
const isSavingQuestion = ref(false)
const isUploadingQuestionImage = ref(false)
const uploadingOptionKey = ref(null)
const draggedQuestionIdx = ref(null)
const dragOverQuestionIdx = ref(null)
const isReorderingQuestions = ref(false)
const reorderStatusMessage = ref('')

const previewBankDirectly = async (bank) => {
  viewingBank.value = bank
  await fetchBankQuestions(bank.id)
  showSimulatorModal.value = true
}

const handleQuestionImageUpload = async (event) => {
  const file = event.target?.files?.[0]
  if (!file) return
  await uploadAndInsertQuestionImage(file)
  if (event.target) event.target.value = ''
}

const uploadAndInsertQuestionImage = async (file) => {
  if (!file.type.startsWith('image/')) {
    showToast('File harus berupa gambar (JPG, PNG, WEBP, GIF, SVG)', 'warning')
    return
  }
  if (file.size > 5 * 1024 * 1024) {
    showToast('Ukuran gambar maksimal 5 MB', 'warning')
    return
  }

  isUploadingQuestionImage.value = true
  const formData = new FormData()
  formData.append('image', file)

  try {
    const res = await api.post('/admin/upload-image', formData, {
      headers: { 'Content-Type': 'multipart/form-data' }
    })
    const imgUrl = res.data.url
    const imgTag = `\n<img src="${imgUrl}" alt="Gambar Soal" />\n`
    questionForm.value.content_html = (questionForm.value.content_html || '') + imgTag
    showToast('Gambar pertanyaan berhasil diunggah!', 'success')
  } catch (err) {
    showToast('Gagal mengunggah gambar: ' + (err.response?.data?.message || err.message), 'danger')
  } finally {
    isUploadingQuestionImage.value = false
  }
}

const handleQuestionPaste = async (event) => {
  const items = (event.clipboardData || event.originalEvent?.clipboardData)?.items
  if (!items) return

  for (const item of items) {
    if (item.type.indexOf('image') !== -1) {
      event.preventDefault()
      const file = item.getAsFile()
      if (file) {
        await uploadAndInsertQuestionImage(file)
      }
      break
    }
  }
}

const handleQuestionDrop = async (event) => {
  const files = event.dataTransfer?.files
  if (files && files.length > 0) {
    const file = files[0]
    if (file.type.startsWith('image/')) {
      await uploadAndInsertQuestionImage(file)
    }
  }
}

const handleOptionImageUpload = async (event, optKey) => {
  const file = event.target?.files?.[0]
  if (!file) return
  if (!file.type.startsWith('image/')) {
    showToast('File harus berupa gambar (JPG, PNG, WEBP, GIF, SVG)', 'warning')
    return
  }
  if (file.size > 5 * 1024 * 1024) {
    showToast('Ukuran gambar maksimal 5 MB', 'warning')
    return
  }

  uploadingOptionKey.value = optKey
  const formData = new FormData()
  formData.append('image', file)

  try {
    const res = await api.post('/admin/upload-image', formData, {
      headers: { 'Content-Type': 'multipart/form-data' }
    })
    const opt = questionForm.value.options.find(o => o.key === optKey)
    if (opt) {
      opt.image_url = res.data.url
    }
    showToast(`Gambar pilihan ${optKey} berhasil diunggah!`, 'success')
  } catch (err) {
    showToast('Gagal mengunggah gambar opsi: ' + (err.response?.data?.message || err.message), 'danger')
  } finally {
    uploadingOptionKey.value = null
    if (event.target) event.target.value = ''
  }
}

const removeOptionImage = (optKey) => {
  const opt = questionForm.value.options.find(o => o.key === optKey)
  if (opt) {
    opt.image_url = ''
  }
}

const insertMathSnippet = (snippet) => {
  questionForm.value.content_html = (questionForm.value.content_html || '') + ' ' + snippet + ' '
}

const onQuestionDragStart = (event, index) => {
  draggedQuestionIdx.value = index
  if (event.dataTransfer) {
    event.dataTransfer.effectAllowed = 'move'
    event.dataTransfer.setData('text/plain', index.toString())
  }
}

const onQuestionDragOver = (event, index) => {
  event.preventDefault()
  if (event.dataTransfer) {
    event.dataTransfer.dropEffect = 'move'
  }
  if (dragOverQuestionIdx.value !== index) {
    dragOverQuestionIdx.value = index
  }
}

const onQuestionDragLeave = (event, index) => {
  if (dragOverQuestionIdx.value === index) {
    dragOverQuestionIdx.value = null
  }
}

const onQuestionDrop = async (event, targetIndex) => {
  event.preventDefault()
  const fromIndex = draggedQuestionIdx.value
  dragOverQuestionIdx.value = null
  draggedQuestionIdx.value = null

  if (fromIndex === null || fromIndex === targetIndex || fromIndex === undefined || targetIndex === undefined) {
    return
  }

  // Local reorder
  const movedItem = bankQuestions.value.splice(fromIndex, 1)[0]
  bankQuestions.value.splice(targetIndex, 0, movedItem)

  // Update question_number locally
  bankQuestions.value.forEach((q, idx) => {
    q.question_number = idx + 1
  })

  // Sync to backend
  const bankId = viewingBank.value?.id
  if (!bankId) return

  isReorderingQuestions.value = true
  reorderStatusMessage.value = 'Menyimpan urutan...'
  try {
    const questionIds = bankQuestions.value.map(q => q.id)
    await api.put(`/admin/question-banks/${bankId}/reorder`, {
      question_ids: questionIds
    })
    reorderStatusMessage.value = '✓ Urutan tersimpan'
    showToast('Urutan nomor soal berhasil diperbarui!', 'success')
    setTimeout(() => {
      if (reorderStatusMessage.value === '✓ Urutan tersimpan') {
        reorderStatusMessage.value = ''
      }
    }, 2500)
  } catch (err) {
    reorderStatusMessage.value = 'Gagal menyimpan urutan'
    showToast('Gagal menyimpan urutan soal: ' + (err.response?.data?.message || err.message), 'danger')
  } finally {
    isReorderingQuestions.value = false
  }
}

const onQuestionDragEnd = () => {
  draggedQuestionIdx.value = null
  dragOverQuestionIdx.value = null
}

const questionForm = ref({
  question_number: 1,
  type: 'MULTIPLE_CHOICE',
  content_html: '',
  options: [
    { key: 'A', text: '', image_url: '' },
    { key: 'B', text: '', image_url: '' },
    { key: 'C', text: '', image_url: '' },
    { key: 'D', text: '', image_url: '' },
    { key: 'E', text: '', image_url: '' }
  ],
  correct_key: 'A',
  rubric_guide: '',
  score_weight: 1.0
})

const getQuestionTypeLabel = (type) => {
  switch (type) {
    case 'SHORT_ANSWER':
      return 'Jawaban Singkat'
    case 'ESSAY':
      return 'Essay / Uraian'
    default:
      return 'Pilihan Ganda'
  }
}

const getQuestionTypeBadgeClass = (type) => {
  switch (type) {
    case 'SHORT_ANSWER':
      return 'bg-emerald-50 border-emerald-200 text-emerald-700'
    case 'ESSAY':
      return 'bg-purple-50 border-purple-200 text-purple-700'
    default:
      return 'bg-blue-50 border-blue-200 text-blue-700'
  }
}

const openBankQuestionsModal = async (bank) => {
  viewingBank.value = bank
  editingQuestionId.value = null
  showBankQuestionsModal.value = true
  await fetchBankQuestions(bank.id)
}

const fetchBankQuestions = async (bankId = null) => {
  const id = bankId || viewingBank.value?.id
  if (!id) return
  isLoadingBankQuestions.value = true
  try {
    const res = await api.get(`/admin/question-banks/${id}/questions`)
    bankQuestions.value = res.data.data || []
    if (res.data.bank) {
      viewingBank.value = res.data.bank
    }
  } catch (err) {
    showToast('Gagal memuat daftar butir soal: ' + (err.response?.data?.message || err.message), 'danger')
  } finally {
    isLoadingBankQuestions.value = false
  }
}

const startAddQuestion = () => {
  editingQuestionId.value = 'new'
  questionForm.value = {
    question_number: (bankQuestions.value.length || 0) + 1,
    type: 'MULTIPLE_CHOICE',
    content_html: '',
    options: [
      { key: 'A', text: '', image_url: '' },
      { key: 'B', text: '', image_url: '' },
      { key: 'C', text: '', image_url: '' },
      { key: 'D', text: '', image_url: '' },
      { key: 'E', text: '', image_url: '' }
    ],
    correct_key: 'A',
    rubric_guide: '',
    score_weight: 1.0
  }
}

const startEditQuestion = (q) => {
  editingQuestionId.value = q.id
  let parsedOpts = q.options || []
  if (typeof parsedOpts === 'string') {
    try {
      parsedOpts = JSON.parse(parsedOpts)
    } catch {
      parsedOpts = []
    }
  }
  const opts = ['A', 'B', 'C', 'D', 'E'].map(k => {
    const existing = parsedOpts.find(o => o.key === k)
    return {
      key: k,
      text: existing ? existing.text : '',
      image_url: existing ? (existing.image_url || '') : ''
    }
  })
  const qType = q.type || 'MULTIPLE_CHOICE'
  questionForm.value = {
    question_number: q.question_number,
    type: qType,
    content_html: q.content_html,
    options: opts,
    correct_key: q.correct_key || (qType === 'SHORT_ANSWER' ? '' : 'A'),
    rubric_guide: q.rubric_guide || '',
    score_weight: q.score_weight || 1.0
  }
}

const cancelEditQuestion = () => {
  editingQuestionId.value = null
}

const saveQuestion = async () => {
  if (!questionForm.value.content_html.trim()) {
    await showAlertModal({
      title: 'Form Belum Lengkap',
      message: 'Konten / naskah soal wajib diisi.',
      type: 'warning'
    })
    return
  }

  const qType = questionForm.value.type || 'MULTIPLE_CHOICE'

  if (qType === 'MULTIPLE_CHOICE') {
    const filledOptions = questionForm.value.options.filter(o => (o.text && o.text.trim()) || (o.image_url && o.image_url.trim()))
    if (filledOptions.length < 2) {
      await showAlertModal({
        title: 'Pilihan Jawaban Kurang',
        message: 'Harap isi minimal 2 pilihan jawaban (misal A dan B teks atau gambar).',
        type: 'warning'
      })
      return
    }
  } else if (qType === 'SHORT_ANSWER') {
    if (!questionForm.value.correct_key.trim()) {
      await showAlertModal({
        title: 'Kunci Jawaban Kosong',
        message: 'Kunci jawaban singkat wajib diisi.',
        type: 'warning'
      })
      return
    }
  }

  isSavingQuestion.value = true
  try {
    const payload = {
      type: qType,
      content_html: questionForm.value.content_html.trim(),
      options: qType === 'MULTIPLE_CHOICE'
        ? questionForm.value.options.filter(o => (o.text && o.text.trim()) || (o.image_url && o.image_url.trim()))
        : [],
      correct_key: questionForm.value.correct_key.trim(),
      rubric_guide: questionForm.value.rubric_guide.trim(),
      score_weight: Number(questionForm.value.score_weight) || 1.0
    }

    if (editingQuestionId.value === 'new') {
      await api.post(`/admin/question-banks/${viewingBank.value.id}/questions`, payload)
      showToast('Butir soal baru berhasil ditambahkan!', 'success')
    } else {
      await api.put(`/admin/questions/${editingQuestionId.value}`, payload)
      showToast('Butir soal berhasil diperbarui!', 'success')
    }

    editingQuestionId.value = null
    await fetchBankQuestions()
    await loadAllData()
  } catch (err) {
    await showAlertModal({
      title: 'Gagal Menyimpan Soal',
      message: err.response?.data?.message || 'Terjadi kesalahan saat menyimpan soal.',
      type: 'danger'
    })
  } finally {
    isSavingQuestion.value = false
  }
}

const deleteQuestionItem = async (q) => {
  const confirmed = await showConfirmModal({
    title: 'Hapus Butir Soal',
    message: `Apakah Anda yakin ingin menghapus Soal #${q.question_number}? Nomor urut soal lainnya akan disesuaikan otomatis.`,
    type: 'danger',
    confirmText: 'Hapus Soal',
    cancelText: 'Batal'
  })
  if (!confirmed) return

  try {
    await api.delete(`/admin/questions/${q.id}`)
    showToast(`Soal #${q.question_number} berhasil dihapus!`, 'success')
    await fetchBankQuestions()
    await loadAllData()
  } catch (err) {
    await showAlertModal({
      title: 'Gagal Menghapus Soal',
      message: err.response?.data?.message || 'Terjadi kesalahan saat menghapus soal.',
      type: 'danger'
    })
  }
}



const loadAllData = async () => {
  try {
    const evId = currentScopedEventId.value
    const schedParams = evId ? { event_id: evId } : {}
    // Daftar siswa dan akun staf hanya diberikan server kepada pengelola master, akun, atau jadwal
    const emptyList = { data: { data: [] } }
    const [statsRes, schedRes, stRes, tcRes, clRes, rdnRes, evRes, subRes, csRes] = await Promise.all([
      api.get('/admin/dashboard-stats'),
      api.get('/admin/schedules', { params: schedParams }),
      canReadPeople.value ? api.get('/admin/students') : emptyList,
      canReadPeople.value ? api.get('/admin/teachers') : emptyList,
      api.get('/admin/classes'),
      api.get('/admin/readiness-matrix'),
      api.get('/admin/events'),
      api.get('/admin/subjects'),
      api.get('/admin/class-subjects'),
    ])
    stats.value = statsRes.data.data || {}
    schedules.value = schedRes.data.data || []
    students.value = stRes.data.data || []
    teachers.value = tcRes.data.data || []
    classes.value = clRes.data.data || []
    readinessData.value = rdnRes.data.data || {}
    events.value = evRes.data.data || []
    subjects.value = subRes.data.data || []
    classSubjects.value = csRes.data.data || []

    if (selectedEventId.value) {
      const exists = events.value.some(e => e.id === selectedEventId.value)
      if (!exists) {
        selectedEventId.value = ''
        selectedEventFilter.value = ''
        sessionStorage.removeItem('cbt_selected_event_id')
      }
    }

    if (classes.value.length > 0 && !newStudent.value.class_id) {
      newStudent.value.class_id = classes.value[0].id
    }
  } catch (e) {
    console.error('Failed to load admin dashboard data', e)
  }
}

const resetStudentSession = async (st) => {
  const confirmed = await showConfirmModal({
    title: 'Lepas Kunci Sesi Login',
    message: `Lepas kunci sesi login untuk ${st.user?.full_name}? Siswa dapat segera login kembali di perangkat baru.`,
    type: 'warning',
    confirmText: 'Lepas Kunci',
    cancelText: 'Batal'
  })
  if (!confirmed) return

  try {
    await api.post(`/admin/users/${st.user_id}/reset-session`)
    showToast(`✓ Kunci sesi login HP ${st.user?.full_name} berhasil dilepas!`, 'success')
    const res = await api.get('/admin/students')
    students.value = res.data.data || []
  } catch (e) {
    await showAlertModal({
      title: 'Gagal Melepas Kunci Sesi',
      message: e.response?.data?.message || 'Terjadi kesalahan saat melepas kunci sesi.',
      type: 'danger'
    })
  }
}

const loadSubjects = async () => {
  try {
    const res = await api.get('/admin/subjects')
    subjects.value = res.data.data || []
  } catch (e) {
    console.error('Failed to load subjects', e)
  }
}

const openCreateSubject = () => {
  isEditSubject.value = false
  subjectForm.value = { id: null, code: '', name: '' }
  showSubjectModal.value = true
}

const openEditSubject = (s) => {
  isEditSubject.value = true
  subjectForm.value = { id: s.id, code: s.code, name: s.name }
  showSubjectModal.value = true
}

const submitSubjectForm = async () => {
  if (!subjectForm.value.code || !subjectForm.value.name) {
    await showAlertModal({
      title: 'Form Belum Lengkap',
      message: 'Kode dan Nama Mata Pelajaran wajib diisi.',
      type: 'warning'
    })
    return
  }
  try {
    if (isEditSubject.value) {
      await api.put(`/admin/subjects/${subjectForm.value.id}`, subjectForm.value)
      showToast('Mata pelajaran berhasil diperbarui!', 'success')
    } else {
      await api.post('/admin/subjects', subjectForm.value)
      showToast('Mata pelajaran berhasil ditambahkan!', 'success')
    }
    showSubjectModal.value = false
    await loadSubjects()
    await loadAllData()
  } catch (e) {
    await showAlertModal({
      title: 'Gagal Menyimpan Mapel',
      message: e.response?.data?.message || 'Terjadi kesalahan saat menyimpan mata pelajaran.',
      type: 'danger'
    })
  }
}

const deleteSubject = async (s) => {
  const confirmed = await showConfirmModal({
    title: 'Hapus Mata Pelajaran',
    message: `Yakin ingin menghapus mata pelajaran ${s.name} (${s.code})?`,
    type: 'danger',
    confirmText: 'Hapus',
    cancelText: 'Batal'
  })
  if (!confirmed) return

  try {
    await api.delete(`/admin/subjects/${s.id}`)
    await loadSubjects()
    await loadAllData()
    showToast('Mata pelajaran berhasil dihapus', 'success')
  } catch (e) {
    await showAlertModal({
      title: 'Gagal Menghapus Mapel',
      message: e.response?.data?.message || 'Terjadi kesalahan saat menghapus mata pelajaran.',
      type: 'danger'
    })
  }
}

// Class Subjects Search, Filter, Sort, Pagination
const classSubjectSearchQuery = ref('')
const selectedClassSubjectGrade = ref('all')
const selectedClassSubjectClassId = ref('')
const classSubjectSortKey = ref('class')
const classSubjectSortOrder = ref('asc')
const classSubjectCurrentPage = ref(1)
const classSubjectPerPage = ref(10)

const toggleClassSubjectSort = (key) => {
  if (classSubjectSortKey.value === key) {
    classSubjectSortOrder.value = classSubjectSortOrder.value === 'asc' ? 'desc' : 'asc'
  } else {
    classSubjectSortKey.value = key
    classSubjectSortOrder.value = 'asc'
  }
}

const filteredClassSubjects = computed(() => {
  let list = classSubjects.value || []

  // Filter by Grade
  if (selectedClassSubjectGrade.value !== 'all') {
    list = list.filter(cs => {
      const cls = cs.class_room || classes.value.find(c => c.id === cs.class_id)
      return cls && cls.grade === selectedClassSubjectGrade.value
    })
  }

  // Filter by Specific Class ID
  if (selectedClassSubjectClassId.value) {
    list = list.filter(cs => cs.class_id === selectedClassSubjectClassId.value || cs.class_room?.id === selectedClassSubjectClassId.value)
  }

  // Filter by Search Query
  if (classSubjectSearchQuery.value.trim()) {
    const q = classSubjectSearchQuery.value.toLowerCase().trim()
    list = list.filter(cs => {
      const clsName = (cs.class_room?.name || '').toLowerCase()
      const clsGrade = (cs.class_room?.grade || '').toLowerCase()
      const subName = (cs.subject?.name || '').toLowerCase()
      const subCode = (cs.subject?.code || '').toLowerCase()
      const teacherName = (cs.teacher?.full_name || '').toLowerCase()
      const teacherUser = (cs.teacher?.username || '').toLowerCase()
      const academicYear = (cs.academic_year || '').toLowerCase()
      return clsName.includes(q) || clsGrade.includes(q) || subName.includes(q) || subCode.includes(q) || teacherName.includes(q) || teacherUser.includes(q) || academicYear.includes(q)
    })
  }

  // Sorting
  if (classSubjectSortKey.value) {
    list = [...list].sort((a, b) => {
      let result = 0
      if (classSubjectSortKey.value === 'class') {
        const nameA = (a.class_room?.name || '').trim().toLowerCase()
        const nameB = (b.class_room?.name || '').trim().toLowerCase()
        result = nameA.localeCompare(nameB, undefined, { numeric: true, sensitivity: 'base' })
      } else if (classSubjectSortKey.value === 'subject') {
        const nameA = (a.subject?.name || '').trim().toLowerCase()
        const nameB = (b.subject?.name || '').trim().toLowerCase()
        result = nameA.localeCompare(nameB, undefined, { numeric: true, sensitivity: 'base' })
      } else if (classSubjectSortKey.value === 'teacher') {
        const nameA = (a.teacher?.full_name || '').trim().toLowerCase()
        const nameB = (b.teacher?.full_name || '').trim().toLowerCase()
        result = nameA.localeCompare(nameB, undefined, { numeric: true, sensitivity: 'base' })
      } else if (classSubjectSortKey.value === 'year') {
        const valA = (a.academic_year || '').trim().toLowerCase()
        const valB = (b.academic_year || '').trim().toLowerCase()
        result = valA.localeCompare(valB, undefined, { numeric: true, sensitivity: 'base' })
      }
      return classSubjectSortOrder.value === 'asc' ? result : -result
    })
  }

  return list
})

const totalClassSubjectPages = computed(() => {
  return Math.max(1, Math.ceil(filteredClassSubjects.value.length / classSubjectPerPage.value))
})

const paginatedClassSubjects = computed(() => {
  const start = (classSubjectCurrentPage.value - 1) * classSubjectPerPage.value
  return filteredClassSubjects.value.slice(start, start + classSubjectPerPage.value)
})

const displayedClassSubjectPages = computed(() => {
  const total = totalClassSubjectPages.value
  const current = classSubjectCurrentPage.value
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

const setClassSubjectPage = (p) => {
  if (p === '...' || p < 1 || p > totalClassSubjectPages.value) return
  classSubjectCurrentPage.value = p
}

const loadClassSubjects = async () => {
  try {
    const res = await api.get('/admin/class-subjects')
    classSubjects.value = res.data.data || []
  } catch (e) {
    console.error('Failed to load class subjects', e)
  }
}

const isEditClassSubject = ref(false)

const openCreateClassSubject = () => {
  isEditClassSubject.value = false
  classSubjectForm.value = {
    id: null,
    class_id: classes.value[0]?.id || '',
    subject_id: subjects.value[0]?.id || '',
    teacher_id: teachers.value[0]?.id || '',
    academic_year: activeExamEvent.value?.academic_year || '2026/2027',
  }
  showClassSubjectModal.value = true
}

const openEditClassSubject = (cs) => {
  isEditClassSubject.value = true
  classSubjectForm.value = {
    id: cs.id,
    class_id: cs.class_id || cs.class_room?.id || '',
    subject_id: cs.subject_id || cs.subject?.id || '',
    teacher_id: cs.teacher_id || cs.teacher?.id || '',
    academic_year: cs.academic_year || activeExamEvent.value?.academic_year || '2026/2027',
  }
  showClassSubjectModal.value = true
}

const submitClassSubjectForm = async () => {
  if (!classSubjectForm.value.class_id || !classSubjectForm.value.subject_id || !classSubjectForm.value.teacher_id) {
    await showAlertModal({
      title: 'Form Belum Lengkap',
      message: 'Kelas, Mata Pelajaran, dan Guru Pengampu wajib dipilih.',
      type: 'warning'
    })
    return
  }
  try {
    if (isEditClassSubject.value && classSubjectForm.value.id) {
      await api.put(`/admin/class-subjects/${classSubjectForm.value.id}`, classSubjectForm.value)
      showToast('Alokasi kelas mapel berhasil diperbarui!', 'success')
    } else {
      await api.post('/admin/class-subjects', classSubjectForm.value)
      showToast('Alokasi kelas mapel berhasil disimpan!', 'success')
    }
    showClassSubjectModal.value = false
    await loadClassSubjects()
    await loadAllData()
  } catch (e) {
    await showAlertModal({
      title: isEditClassSubject.value ? 'Gagal Memperbarui Alokasi' : 'Gagal Menyimpan Alokasi',
      message: e.response?.data?.message || 'Terjadi kesalahan saat menyimpan alokasi kelas mapel.',
      type: 'danger'
    })
  }
}

const deleteClassSubject = async (cs) => {
  const confirmed = await showConfirmModal({
    title: 'Hapus Alokasi Mapel',
    message: `Hapus alokasi ${cs.subject?.name} untuk kelas ${cs.class_room?.name}?`,
    type: 'danger',
    confirmText: 'Hapus Alokasi',
    cancelText: 'Batal'
  })
  if (!confirmed) return

  try {
    await api.delete(`/admin/class-subjects/${cs.id}`)
    await loadClassSubjects()
    await loadAllData()
    showToast('Alokasi kelas mapel berhasil dihapus', 'success')
  } catch (e) {
    await showAlertModal({
      title: 'Gagal Menghapus Alokasi',
      message: e.response?.data?.message || 'Terjadi kesalahan saat menghapus alokasi.',
      type: 'danger'
    })
  }
}

const loadEvents = async () => {
  try {
    const res = await api.get('/admin/events')
    events.value = res.data.data || []
    if (selectedEventId.value) {
      const exists = events.value.some(e => e.id === selectedEventId.value)
      if (!exists) {
        selectedEventId.value = ''
        selectedEventFilter.value = ''
        sessionStorage.removeItem('cbt_selected_event_id')
      }
    }
  } catch (e) {
    console.error('Failed to load events', e)
  }
}

const loadSchedules = async () => {
  try {
    const params = {}
    const evId = currentScopedEventId.value
    if (evId) {
      params.event_id = evId
    }
    const res = await api.get('/admin/schedules', { params })
    schedules.value = res.data.data || []
  } catch (e) {
    console.error('Failed to load schedules', e)
  }
}

const filterByEventAndOpenSchedules = (ev) => {
  enterEvent(ev)
}

const openCreateEvent = () => {
  isEditEvent.value = false
  const now = new Date()
  const nextMonth = new Date(now.getTime() + 14 * 24 * 60 * 60 * 1000)
  eventForm.value = {
    id: null,
    title: '',
    code: '',
    academic_year: '2026/2027',
    semester: 'GANJIL',
    start_date: now.toISOString().split('T')[0],
    end_date: nextMonth.toISOString().split('T')[0],
    is_active: false,
    description: '',
  }
  showEventModal.value = true
}

const openEditEvent = (ev) => {
  isEditEvent.value = true
  eventForm.value = {
    id: ev.id,
    title: ev.title,
    code: ev.code,
    academic_year: ev.academic_year,
    semester: ev.semester,
    start_date: ev.start_date ? ev.start_date.split('T')[0] : '',
    end_date: ev.end_date ? ev.end_date.split('T')[0] : '',
    is_active: ev.is_active,
    description: ev.description || '',
  }
  showEventModal.value = true
}

const submitEventForm = async () => {
  try {
    if (isEditEvent.value) {
      await api.put(`/admin/events/${eventForm.value.id}`, eventForm.value)
      showToast('✓ Event berhasil diperbarui!', 'success')
    } else {
      await api.post('/admin/events', eventForm.value)
      showToast('✓ Event ujian baru berhasil dibuat!', 'success')
    }
    showEventModal.value = false
    await loadEvents()
    await loadSchedules()
  } catch (e) {
    await showAlertModal({
      title: 'Gagal Menyimpan Event',
      message: e.response?.data?.message || 'Terjadi kesalahan saat menyimpan data event.',
      type: 'danger'
    })
  }
}

const toggleEventActive = async (ev) => {
  try {
    const res = await api.post(`/admin/events/${ev.id}/toggle-active`)
    ev.is_active = res.data.is_active
    await loadEvents()
    await loadSchedules()
    showToast(ev.is_active ? `Event ${ev.title} diaktifkan!` : `Event ${ev.title} dinonaktifkan.`, 'info')
  } catch (e) {
    await showAlertModal({
      title: 'Gagal Mengubah Status',
      message: e.response?.data?.message || 'Gagal mengubah status event.',
      type: 'danger'
    })
  }
}

const deleteEvent = async (ev) => {
  const confirmed = await showConfirmModal({
    title: 'Hapus Event Ujian',
    message: `Hapus event "${ev.title}"?\n\nPerhatian: Event dengan sesi jadwal ujian aktif tidak dapat dihapus.`,
    type: 'danger',
    confirmText: 'Hapus Event',
    cancelText: 'Batal'
  })
  if (!confirmed) return

  try {
    await api.delete(`/admin/events/${ev.id}`)
    showToast('Event berhasil dihapus', 'success')
    await loadEvents()
  } catch (e) {
    await showAlertModal({
      title: 'Gagal Menghapus Event',
      message: e.response?.data?.message || 'Terjadi kesalahan saat menghapus event.',
      type: 'danger'
    })
  }
}

const resetUserSession = async (u) => {
  const confirmed = await showConfirmModal({
    title: 'Lepas Kunci Sesi Login',
    message: `Lepas kunci sesi login untuk pengguna ${u.username}? Siswa akan dapat login kembali di perangkat baru.`,
    type: 'warning',
    confirmText: 'Lepas Kunci',
    cancelText: 'Batal'
  })
  if (!confirmed) return

  try {
    await api.post(`/admin/users/${u.id}/reset-session`)
    u.has_active_session = false
    showToast(`Sesi login untuk ${u.username} berhasil direset`, 'success')
  } catch (e) {
    await showAlertModal({
      title: 'Gagal Mereset Sesi',
      message: e.response?.data?.message || 'Terjadi kesalahan saat mereset sesi.',
      type: 'danger'
    })
  }
}

const downloadStudentTemplate = async () => {
  try {
    const res = await api.get('/admin/users/template', {
      responseType: 'blob'
    })
    const url = window.URL.createObjectURL(new Blob([res.data], {
      type: 'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet'
    }))
    const link = document.createElement('a')
    link.href = url
    link.setAttribute('download', 'Template_Import_Siswa.xlsx')
    document.body.appendChild(link)
    link.click()
    link.remove()
    window.URL.revokeObjectURL(url)
    showToast('Format template siswa berhasil diunduh!', 'success')
  } catch (err) {
    await showAlertModal({
      title: 'Gagal Mengunduh Template',
      message: await extractExportErrorMessage(err, 'Gagal mengunduh template'),
      type: 'danger'
    })
  }
}

const handleExcelFileSelect = (e) => {
  importExcelFile.value = e.target.files[0]
}

const submitImportExcel = async () => {
  if (!importExcelFile.value) {
    await showAlertModal({
      title: 'Pilih File Excel',
      message: 'Pilih file Excel terlebih dahulu!',
      type: 'warning'
    })
    return
  }
  const formData = new FormData()
  formData.append('file', importExcelFile.value)
  isImporting.value = true
  try {
    const res = await api.post('/admin/users/import-excel', formData, {
      headers: { 'Content-Type': 'multipart/form-data' },
    })
    showToast(res.data.message || 'Import berhasil!', 'success')
    showImportModal.value = false
    importExcelFile.value = null
    await loadAllData()
  } catch (e) {
    await showAlertModal({
      title: 'Gagal Impor Excel',
      message: e.response?.data?.message || 'Terjadi kesalahan saat mengimpor file Excel.',
      type: 'danger'
    })
  } finally {
    isImporting.value = false
  }
}

const downloadClassTemplate = async () => {
  try {
    const res = await api.get('/admin/classes/template', { responseType: 'blob' })
    const url = window.URL.createObjectURL(new Blob([res.data], {
      type: 'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet'
    }))
    const a = document.createElement('a')
    a.href = url
    a.download = 'Template_Import_Kelas_CBT.xlsx'
    document.body.appendChild(a)
    a.click()
    a.remove()
    window.URL.revokeObjectURL(url)
    showToast('Template kelas berhasil diunduh!', 'success')
  } catch (err) {
    await showAlertModal({ title: 'Gagal Mengunduh Template', message: 'Gagal mengunduh template kelas.', type: 'danger' })
  }
}

const submitImportClasses = async () => {
  if (!importClassFile.value) return
  isImportingClass.value = true
  importClassResult.value = null
  try {
    const fd = new FormData()
    fd.append('file', importClassFile.value)
    const res = await api.post('/admin/classes/import', fd, {
      headers: { 'Content-Type': 'multipart/form-data' },
    })
    importClassResult.value = { success: true, message: res.data.message || 'Impor kelas berhasil!' }
    await loadAllData()
  } catch (e) {
    importClassResult.value = { success: false, message: e.response?.data?.message || 'Gagal mengunggah file' }
  } finally {
    isImportingClass.value = false
  }
}

const downloadSubjectTemplate = async () => {
  try {
    const res = await api.get('/admin/subjects/template', { responseType: 'blob' })
    const url = window.URL.createObjectURL(new Blob([res.data], {
      type: 'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet'
    }))
    const a = document.createElement('a')
    a.href = url
    a.download = 'Template_Import_Mapel_CBT.xlsx'
    document.body.appendChild(a)
    a.click()
    a.remove()
    window.URL.revokeObjectURL(url)
    showToast('Template mata pelajaran berhasil diunduh!', 'success')
  } catch (err) {
    await showAlertModal({ title: 'Gagal Mengunduh Template', message: 'Gagal mengunduh template mata pelajaran.', type: 'danger' })
  }
}

const submitImportSubjects = async () => {
  if (!importSubjectFile.value) return
  isImportingSubject.value = true
  importSubjectResult.value = null
  try {
    const fd = new FormData()
    fd.append('file', importSubjectFile.value)
    const res = await api.post('/admin/subjects/import', fd, {
      headers: { 'Content-Type': 'multipart/form-data' },
    })
    importSubjectResult.value = { success: true, message: res.data.message || 'Impor mata pelajaran berhasil!' }
    await loadAllData()
  } catch (e) {
    importSubjectResult.value = { success: false, message: e.response?.data?.message || 'Gagal mengunggah file' }
  } finally {
    isImportingSubject.value = false
  }
}

const downloadTeacherTemplate = async () => {
  try {
    const res = await api.get('/admin/teachers/template', { responseType: 'blob' })
    const url = window.URL.createObjectURL(new Blob([res.data], {
      type: 'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet'
    }))
    const a = document.createElement('a')
    a.href = url
    a.download = 'Template_Import_Guru_CBT.xlsx'
    document.body.appendChild(a)
    a.click()
    a.remove()
    window.URL.revokeObjectURL(url)
    showToast('Template guru berhasil diunduh!', 'success')
  } catch (err) {
    await showAlertModal({ title: 'Gagal Mengunduh Template', message: 'Gagal mengunduh template guru.', type: 'danger' })
  }
}

const submitImportTeachers = async () => {
  if (!importTeacherFile.value) return
  isImportingTeacher.value = true
  importTeacherResult.value = null
  try {
    const fd = new FormData()
    fd.append('file', importTeacherFile.value)
    const res = await api.post('/admin/teachers/import', fd, {
      headers: { 'Content-Type': 'multipart/form-data' },
    })
    importTeacherResult.value = { success: true, message: res.data.message || 'Impor guru berhasil!' }
    await loadAllData()
  } catch (e) {
    importTeacherResult.value = { success: false, message: e.response?.data?.message || 'Gagal mengunggah file' }
  } finally {
    isImportingTeacher.value = false
  }
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

const toggleBankLock = async (b) => {
  try {
    const res = await api.post(`/admin/question-banks/${b.id}/toggle-lock`)
    b.is_locked = res.data.is_locked
    showToast(b.is_locked ? 'Bank Soal dikunci (Terkunci)' : 'Bank Soal dibuka (Draft)', 'info')
  } catch (e) {
    await showAlertModal({
      title: 'Gagal Mengubah Status Bank Soal',
      message: 'Terjadi kesalahan saat mengubah status penguncian bank soal.',
      type: 'danger'
    })
  }
}

// --- Student Master Data Methods ---
const openCreateStudentModal = () => {
  newStudent.value = {
    full_name: '',
    username: '',
    password: '',
    nis: '',
    nisn: '',
    class_id: classes.value[0]?.id || '',
    gender: 'L'
  }
  showStudentModal.value = true
}

const openEditStudent = (st) => {
  editStudentForm.value = {
    id: st.id,
    user_id: st.user_id || st.user?.id,
    full_name: st.user?.full_name || '',
    username: st.user?.username || '',
    password: '',
    nis: st.nis || '',
    nisn: st.nisn || '',
    class_id: st.class_room_id || st.class_id || classes.value[0]?.id || '',
    gender: st.gender || 'L'
  }
  showEditStudentModal.value = true
}

const openStudentDetail = (st) => {
  selectedStudentDetail.value = st
  showStudentDetailModal.value = true
}

const createStudent = async () => {
  try {
    await api.post('/admin/students', newStudent.value)
    showStudentModal.value = false
    newStudent.value = { full_name: '', username: '', password: '', nis: '', nisn: '', class_id: classes.value[0]?.id || '', gender: 'L' }
    await loadAllData()
    showToast('Siswa berhasil ditambahkan!', 'success')
  } catch (e) {
    await showAlertModal({
      title: 'Gagal Menambah Siswa',
      message: e.response?.data?.message || 'Terjadi kesalahan saat menambah data siswa.',
      type: 'danger'
    })
  }
}

const updateStudent = async () => {
  try {
    const targetId = editStudentForm.value.id || editStudentForm.value.user_id
    await api.put(`/admin/students/${targetId}`, editStudentForm.value)
    showEditStudentModal.value = false
    await loadAllData()
    showToast('Data siswa berhasil diperbarui!', 'success')
  } catch (e) {
    await showAlertModal({
      title: 'Gagal Memperbarui Siswa',
      message: e.response?.data?.message || 'Terjadi kesalahan saat memperbarui data siswa.',
      type: 'danger'
    })
  }
}

const deleteStudent = async (st) => {
  const confirmed = await showConfirmModal({
    title: 'Hapus Akun Siswa',
    message: `Yakin ingin menghapus akun siswa ${st.user?.full_name || ''} (${st.user?.username || ''})?\n\nTindakan ini akan menghapus profil siswa beserta riwayat nilai terkait.`,
    type: 'danger',
    confirmText: 'Hapus Siswa',
    cancelText: 'Batal'
  })
  if (!confirmed) return

  try {
    const targetId = st.id || st.user_id
    await api.delete(`/admin/students/${targetId}`)
    await loadAllData()
    showToast('Siswa berhasil dihapus', 'success')
  } catch (e) {
    await showAlertModal({
      title: 'Gagal Menghapus Siswa',
      message: e.response?.data?.message || 'Terjadi kesalahan saat menghapus data siswa.',
      type: 'danger'
    })
  }
}

// --- Teacher Master Data Methods ---
const openCreateTeacherModal = async () => {
  const catalog = await ensurePermissionCatalog()
  if (!catalog) return
  newTeacher.value = { full_name: '', username: '', password: '' }
  newTeacherAccess.value = defaultStaffAccess(catalog)
  newTeacherAccessValid.value = true
  showTeacherModal.value = true
}

const openEditTeacher = async (tc) => {
  const catalog = await ensurePermissionCatalog()
  if (!catalog) return
  editTeacherForm.value = {
    id: tc.id,
    full_name: tc.full_name || '',
    username: tc.username || '',
    password: ''
  }
  editTeacherAccess.value = staffAccessFromAccount(tc, catalog)
  editTeacherAccessValid.value = true
  showEditTeacherModal.value = true
}

const createTeacher = async () => {
  if (!newTeacherCanSave.value) return
  try {
    await api.post('/admin/teachers', buildStaffCreatePayload(newTeacher.value, newTeacherAccess.value))
    showTeacherModal.value = false
    newTeacher.value = { full_name: '', username: '', password: '' }
    await loadAllData()
    showToast('Akun guru / staf berhasil ditambahkan!', 'success')
  } catch (e) {
    await showAlertModal({
      title: 'Gagal Menambah Guru / Staf',
      message: e.response?.data?.message || 'Terjadi kesalahan saat menambah guru / staf.',
      type: 'danger'
    })
  }
}

const updateTeacher = async () => {
  if (!editTeacherCanSave.value) return
  try {
    await api.put(`/admin/teachers/${editTeacherForm.value.id}`, buildStaffUpdatePayload(editTeacherForm.value, editTeacherAccess.value))
    showEditTeacherModal.value = false
    await loadAllData()
    showToast('Data guru / staf berhasil diperbarui!', 'success')
  } catch (e) {
    await showAlertModal({
      title: 'Gagal Memperbarui Guru / Staf',
      message: e.response?.data?.message || 'Terjadi kesalahan saat memperbarui data guru / staf.',
      type: 'danger'
    })
  }
}

const deleteTeacher = async (tc) => {
  const confirmed = await showConfirmModal({
    title: 'Hapus Guru / Staf',
    message: `Yakin ingin menghapus akun ${tc.full_name} (${tc.username})?`,
    type: 'danger',
    confirmText: 'Hapus Akun',
    cancelText: 'Batal'
  })
  if (!confirmed) return

  try {
    await api.delete(`/admin/teachers/${tc.id}`)
    await loadAllData()
    showToast('Akun guru / staf berhasil dihapus', 'success')
  } catch (e) {
    await showAlertModal({
      title: 'Gagal Menghapus Guru / Staf',
      message: e.response?.data?.message || 'Terjadi kesalahan saat menghapus akun guru / staf.',
      type: 'danger'
    })
  }
}

// --- Class Master Data Methods ---
const openCreateClassModal = () => {
  newClass.value = { name: '', grade: 'XII', major: 'MIPA' }
  showClassModal.value = true
}

const openEditClass = (c) => {
  editClassForm.value = {
    id: c.id,
    name: c.name || '',
    grade: c.grade || 'XII',
    major: c.major || ''
  }
  showEditClassModal.value = true
}

const createClass = async () => {
  try {
    await api.post('/admin/classes', newClass.value)
    showClassModal.value = false
    newClass.value = { name: '', grade: 'XII', major: 'MIPA' }
    await loadAllData()
    showToast('Kelas berhasil ditambahkan!', 'success')
  } catch (e) {
    await showAlertModal({
      title: 'Gagal Menambah Kelas',
      message: e.response?.data?.message || 'Terjadi kesalahan saat menambah data kelas.',
      type: 'danger'
    })
  }
}

const updateClass = async () => {
  try {
    await api.put(`/admin/classes/${editClassForm.value.id}`, editClassForm.value)
    showEditClassModal.value = false
    await loadAllData()
    showToast('Data kelas berhasil diperbarui!', 'success')
  } catch (e) {
    await showAlertModal({
      title: 'Gagal Memperbarui Kelas',
      message: e.response?.data?.message || 'Terjadi kesalahan saat memperbarui data kelas.',
      type: 'danger'
    })
  }
}

const deleteClass = async (c) => {
  const confirmed = await showConfirmModal({
    title: 'Hapus Kelas',
    message: `Yakin ingin menghapus rombel kelas "${c.name}"?\n\nPerhatian: Siswa dan alokasi mapel yang terhubung ke kelas ini dapat terdampak.`,
    type: 'danger',
    confirmText: 'Hapus Kelas',
    cancelText: 'Batal'
  })
  if (!confirmed) return

  try {
    await api.delete(`/admin/classes/${c.id}`)
    await loadAllData()
    showToast('Kelas berhasil dihapus', 'success')
  } catch (e) {
    await showAlertModal({
      title: 'Gagal Menghapus Kelas',
      message: e.response?.data?.message || 'Terjadi kesalahan saat menghapus data kelas.',
      type: 'danger'
    })
  }
}

const handleLogout = async () => {
  try {
    await authStore.logout()
  } catch (e) {
    console.error('Logout error', e)
  }
  localStorage.clear()
  sessionStorage.clear()
  window.location.href = '/login'
}


const closeEventMenuOnClickOutside = () => {
  if (activeEventMenuId.value !== null) {
    activeEventMenuId.value = null
  }
}

onUnmounted(() => {
  window.removeEventListener('click', closeEventMenuOnClickOutside)
})

onMounted(() => {
  window.addEventListener('click', closeEventMenuOnClickOutside)
  const queryTab = router.currentRoute.value.query.tab
  const initialTab = queryTab || landingTab(authStore.user)
  if (initialTab !== 'dashboard') {
    switchTab(initialTab)
  } else {
    loadLoginSessions()
    loadTeacherHomeProctorTasks()
  }
  loadAllData()
})
// Bagikan state dan aksi ke komponen tab/modal di ./dashboard (lihat dashboard/context.js).
provide(DASHBOARD_CONTEXT, {
  activeEventMenuId,
  activeEventsCount,
  activeExamEvent,
  activeLoginSessions,
  activeProctorScheduleId,
  activeSchedulesCount,
  activeTab,
  adminStaffCount,
  allocatedClassesCount,
  allocatedTeachersCount,
  applyFormProctors,
  authStore,
  availableBankSubjects,
  availableBanksForSelectedSchedule,
  availableClassSubjectGrades,
  availableClassSubjects,
  availableScheduleGrades,
  availableStudentGrades,
  bankCustomClassOptions,
  bankCustomSubjects,
  bankFormLocked,
  bankQuestions,
  bankScopeLabel,
  bankScopeOptions,
  bulkActivateSchedules,
  bulkDeactivateSchedules,
  bulkRegenerateTokens,
  canImportStudents,
  canManageEvents,
  canManageLoginSessions,
  canManageMaster,
  canManageSchedules,
  canPrintQuestionBanks,
  canReadPeople,
  canShowTeacherProctorSection,
  canShowTeacherQuestionSection,
  canUploadQuestionBank,
  cancelEditQuestion,
  cardPrintEvent,
  classCurrentPage,
  classPerPage,
  classSearchQuery,
  classSortKey,
  classSortOrder,
  classSubjectCurrentPage,
  classSubjectForm,
  classSubjectPerPage,
  classSubjectSearchQuery,
  classSubjectSortKey,
  classSubjectSortOrder,
  classSubjects,
  classes,
  copyTokenToClipboard,
  createClass,
  createStudent,
  createTeacher,
  currentEventSchedules,
  currentSectionMeta,
  deleteClass,
  deleteClassSubject,
  deleteEvent,
  deleteQuestionBank,
  deleteQuestionItem,
  deleteSchedule,
  deleteStudent,
  deleteSubject,
  deleteTeacher,
  dialogState,
  displayedClassPages,
  displayedClassSubjectPages,
  displayedQuestionBankPages,
  displayedSchedulePages,
  displayedStudentPages,
  displayedSubjectPages,
  displayedTeacherPages,
  downloadClassTemplate,
  downloadQuestionBankTemplate,
  downloadStudentTemplate,
  downloadSubjectTemplate,
  downloadTeacherTemplate,
  draftBanksCount,
  dragOverQuestionIdx,
  draggedQuestionIdx,
  editClassForm,
  editStudentForm,
  editTeacherAccess,
  editTeacherAccessValid,
  editTeacherCanSave,
  editTeacherForm,
  editingQuestionId,
  enterEvent,
  essayDraft,
  essayLoading,
  essayQuestions,
  essaySaving,
  essayTotalPending,
  eventActionBtnClass,
  eventActionClick,
  eventActionLabel,
  eventForm,
  eventSearchQuery,
  eventStatusBadgeClass,
  eventStatusFilter,
  eventStatusLabel,
  events,
  exitEventScope,
  exportBeritaAcaraPDF,
  exportNilaiExcel,
  femaleStudentsCount,
  filteredClassSubjects,
  filteredClasses,
  filteredEvents,
  filteredQuestionBanks,
  filteredSchedules,
  filteredStudents,
  filteredSubjects,
  filteredTeachers,
  formatDate,
  formatProctorNames,
  formatScheduleDateFull,
  formatScheduleTimeOnly,
  formatScheduleTimeRange,
  getDayName,
  getQuestionTypeBadgeClass,
  getQuestionTypeLabel,
  getSubjectBankCount,
  getSubjectClassCount,
  gradeCounts,
  groupedAvailableClassSubjects,
  guruCount,
  handleDialogCancel,
  handleDialogConfirm,
  handleExcelFileSelect,
  handleLogout,
  handleOptionImageUpload,
  handleQuestionBankFileChange,
  handleQuestionDrop,
  handleQuestionImageUpload,
  handleQuestionPaste,
  importClassFile,
  importClassResult,
  importSubjectFile,
  importSubjectResult,
  importTeacherFile,
  importTeacherResult,
  insertMathSnippet,
  isAllPaginatedSelected,
  isCreatingBank,
  isEditBank,
  isEditClassSubject,
  isEditEvent,
  isEditSchedule,
  isEditSubject,
  isExportingNilai,
  isExportingPDF,
  isGrantor,
  isImporting,
  isImportingClass,
  isImportingSubject,
  isImportingTeacher,
  isLoadingBankQuestions,
  isMobileSidebarOpen,
  isReorderingQuestions,
  isSavingQuestion,
  isSidebarCollapsed,
  isUploadingQuestionBank,
  isUploadingQuestionImage,
  loadSchedules,
  lockedBanksCount,
  lockedStudentsCount,
  maleStudentsCount,
  newBankForm,
  newBankOptionKey,
  newClass,
  newStudent,
  newTeacher,
  newTeacherAccess,
  newTeacherAccessValid,
  newTeacherCanSave,
  onBankCustomSubjectChange,
  onBankSubjectChange,
  onClassSubjectChange,
  onQuestionDragEnd,
  onQuestionDragLeave,
  onQuestionDragOver,
  onQuestionDragStart,
  onQuestionDrop,
  openBankQuestionsModal,
  openCardPrint,
  openCreateBankModal,
  openCreateClassModal,
  openCreateClassSubject,
  openCreateEvent,
  openCreateSchedule,
  openCreateStudentModal,
  openCreateSubject,
  openCreateTeacherModal,
  openEditBankModal,
  openEditClass,
  openEditClassSubject,
  openEditEvent,
  openEditSchedule,
  openEditStudent,
  openEditSubject,
  openEditTeacher,
  openLinkBankModal,
  openPrintModal,
  openProctorAssign,
  openQuestionPrint,
  openScheduleDetail,
  openStudentDetail,
  openUploadBankModal,
  paginatedClassSubjects,
  paginatedClasses,
  paginatedQuestionBanks,
  paginatedSchedules,
  paginatedStudents,
  paginatedSubjects,
  paginatedTeachers,
  permissionCatalog,
  portalLabel,
  portalRoleLabel,
  previewBankDirectly,
  printDocType,
  printSchedule,
  printStudentsList,
  proctorData,
  proctorSchedules,
  questionBankCurrentPage,
  questionBankPerPage,
  questionBankSearchQuery,
  questionBankSortKey,
  questionBankSortOrder,
  questionBankStatusFilter,
  questionBankUploadStatus,
  questionForm,
  questionMakersCount,
  questionPrintBank,
  readinessData,
  removeFormProctor,
  removeOptionImage,
  reorderStatusMessage,
  resetStudentFilters,
  resetStudentSession,
  resetUserSession,
  runTeacherHomeAction,
  saveEssayQuestion,
  saveQuestion,
  scheduleCurrentPage,
  scheduleDetailTab,
  scheduleForProctors,
  scheduleForm,
  scheduleFormSessionPeers,
  schedulePerPage,
  scheduleSearchQuery,
  scheduleSortKey,
  scheduleSortOrder,
  scheduledClassesCount,
  selectedBankForLink,
  selectedBankUploadId,
  selectedClassGrade,
  selectedClassSubjectClassId,
  selectedClassSubjectGrade,
  selectedEventFilterName,
  selectedExamEvent,
  selectedScheduleClassSubjectId,
  selectedScheduleDetail,
  selectedScheduleForLink,
  selectedScheduleGrade,
  selectedScheduleIds,
  selectedScheduleTimeFilter,
  selectedStudentDetail,
  selectedStudentGrade,
  selectedTeacherRole,
  setBankScope,
  setClassPage,
  setClassSubjectPage,
  setQuestionBankPage,
  setSchedulePage,
  setStudentPage,
  setSubjectPage,
  setTeacherPage,
  showBankQuestionsModal,
  showCardPrint,
  showClassModal,
  showClassSubjectModal,
  showConfirmModal,
  showCreateBankModal,
  showEditClassModal,
  showEditStudentModal,
  showEditTeacherModal,
  showEventModal,
  showEventReadiness,
  showFormProctorPicker,
  showImportClassModal,
  showImportModal,
  showImportSubjectModal,
  showImportTeacherModal,
  showLinkBankModal,
  showPrintModal,
  showProctorAssignModal,
  showProctorPrint,
  showQuestionPrint,
  showQuickActions,
  showScheduleDetailModal,
  showScheduleModal,
  showStudentDetailModal,
  showStudentModal,
  showSubjectModal,
  showTeacherModal,
  showUploadBankModal,
  showWelcomeCard,
  startAddQuestion,
  startEditQuestion,
  stats,
  studentCurrentPage,
  studentPerPage,
  studentSearchQuery,
  studentSortKey,
  studentSortOrder,
  students,
  subjectCurrentPage,
  subjectForm,
  subjectPerPage,
  subjectSearchQuery,
  subjectSortKey,
  subjectSortOrder,
  subjects,
  subjectsWithBanksCount,
  submitClassSubjectForm,
  submitCreateBank,
  submitEventForm,
  submitImportClasses,
  submitImportExcel,
  submitImportSubjects,
  submitImportTeachers,
  submitLinkBank,
  submitQuestionBankUpload,
  submitScheduleForm,
  submitSubjectForm,
  switchTab,
  switchToEssayTab,
  teacherBadgeClass,
  teacherCurrentPage,
  teacherHomeActionItems,
  teacherHomeBanks,
  teacherHomeDraftBankCount,
  teacherHomeGridClass,
  teacherHomeIsInactive,
  teacherHomeLockedBankCount,
  teacherHomeProctorCardSpanClass,
  teacherHomeProctorTasks,
  teacherHomeProctorTodayCount,
  teacherHomeProctorUpcomingList,
  teacherHomeRoleLabel,
  teacherHomeScheduleBankStatus,
  teacherHomeSchedules,
  teacherHomeSchedulesWithoutBank,
  teacherHomeTodayScheduleCount,
  teacherHomeUnfinishedBanks,
  teacherHomeUpcomingList,
  teacherHomeVisibleActionItems,
  teacherPerPage,
  teacherRoleChips,
  teacherRoleLabel,
  teacherSearchQuery,
  teacherSortKey,
  teacherSortOrder,
  teachers,
  toastMessage,
  toastType,
  toggleBankClass,
  toggleBankLockWithConfirm,
  toggleClassSort,
  toggleClassSubjectSort,
  toggleEventActive,
  toggleEventMenu,
  toggleQuestionBankSort,
  toggleScheduleSort,
  toggleScheduleStatus,
  toggleSelectAllPaginatedSchedules,
  toggleStudentSort,
  toggleSubjectSort,
  toggleTeacherSort,
  totalClassPages,
  totalClassSubjectPages,
  totalQuestionBankPages,
  totalQuestionsCount,
  totalSchedulePages,
  totalSchedulesCount,
  totalStudentPages,
  totalSubjectPages,
  totalTeacherPages,
  triggerPrintDocument,
  updateClass,
  updateStudent,
  updateTeacher,
  uploadingOptionKey,
  viewingBank,
})
</script>

