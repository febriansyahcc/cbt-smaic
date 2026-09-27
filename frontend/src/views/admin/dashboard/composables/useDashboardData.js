import { ref, computed } from 'vue'
import api from '../../../../services/api'

// Data bersama dasbor (event, jadwal, siswa, guru, kelas, mapel) beserta cakupan event terpilih dan fungsi pemuatnya.
export function useDashboardData(ctx) {
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

  const stats = ref({})
  const schedules = ref([])

  const students = ref([])
  const teachers = ref([])
  const classes = ref([])
  const readinessData = ref({})

  // Subject & ClassSubject States
  const subjects = ref([])

  const classSubjects = ref([])

  const loadSubjects = async () => {
    try {
      const res = await api.get('/admin/subjects')
      subjects.value = res.data.data || []
    } catch (e) {
      console.error('Failed to load subjects', e)
    }
  }

  const loadClassSubjects = async () => {
    try {
      const res = await api.get('/admin/class-subjects')
      classSubjects.value = res.data.data || []
    } catch (e) {
      console.error('Failed to load class subjects', e)
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

  return {
    events,
    selectedEventId,
    selectedEventFilter,
    isEventDropdownOpen,
    selectedExamEvent,
    activeExamEvent,
    currentScopedEventId,
    currentEventSchedules,
    stats,
    schedules,
    students,
    teachers,
    classes,
    readinessData,
    subjects,
    classSubjects,
    loadSubjects,
    loadClassSubjects,
    loadEvents,
    loadSchedules,
  }
}
