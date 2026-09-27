import { computed } from 'vue'

// Angka ringkasan untuk kartu statistik di berbagai tab.
export function useDashboardMetrics(ctx) {
  const {
    classSubjects,
    classes,
    currentEventSchedules,
    currentScopedEventId,
    events,
    readinessData,
    selectedExamEvent,
    students,
    teachers,
  } = ctx

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

  const lockedStudentsCount = computed(() => students.value.filter(s => s.has_active_session).length)
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

  return {
    activeEventsCount,
    totalSchedulesCount,
    activeSchedulesCount,
    scheduledClassesCount,
    selectedEventFilterName,
    lockedStudentsCount,
    maleStudentsCount,
    femaleStudentsCount,
    guruCount,
    adminStaffCount,
    questionMakersCount,
    subjectsWithBanksCount,
    allocatedClassesCount,
    allocatedTeachersCount,
    getSubjectBankCount,
    getSubjectClassCount,
    gradeCounts,
    lockedBanksCount,
    draftBanksCount,
    totalQuestionsCount,
  }
}
