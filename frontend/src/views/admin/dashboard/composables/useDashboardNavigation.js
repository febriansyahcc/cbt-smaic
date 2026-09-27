import { landingTab } from '../../../../utils/access'
import api from '../../../../services/api'

// Perpindahan tab dan pemuatan ulang seluruh data; memakai hampir semua domain sehingga dipanggil terakhir.
export function useDashboardNavigation(ctx) {
  const {
    activeProctorScheduleId,
    activeTab,
    authStore,
    canReadPeople,
    classSubjects,
    classes,
    currentScopedEventId,
    events,
    isMobileSidebarOpen,
    loadClassSubjects,
    loadEvents,
    loadPermissionCatalog,
    loadProctorSchedules,
    loadSchedules,
    loadSubjects,
    newStudent,
    readinessData,
    schedules,
    selectedBankUploadId,
    selectedEventFilter,
    selectedEventId,
    stats,
    students,
    subjects,
    teachers,
  } = ctx

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

  return {
    switchTab,
    loadAllData,
  }
}
