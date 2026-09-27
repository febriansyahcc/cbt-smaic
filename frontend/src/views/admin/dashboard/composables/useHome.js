import { ref, computed, watch } from 'vue'
import api from '../../../../services/api'

// Beranda: kesiapan event dan sesi login siswa (pengelola), ringkasan tugas guru/pengawas.
export function useHome(ctx) {
  const {
    activeExamEvent,
    activeTab,
    authStore,
    canLockQuestionBanks,
    canManageLoginSessions,
    currentEventSchedules,
    currentScopedEventId,
    enterEvent,
    formatScheduleTimeRange,
    loadProctorSchedules,
    proctorSchedules,
    readinessData,
    schedules,
    selectedExamEvent,
    showWelcomeCard,
    toggleBankLockWithConfirm,
  } = ctx

  // Didefinisikan composable yang dipanggil kemudian; diteruskan saat dipanggil.
  const switchTab = (...args) => ctx.switchTab(...args)

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

  // ---- Beranda guru / pengawas: ringkasan tugas milik pengguna yang sedang login ----
  // Sumber data: jadwal (GET /admin/schedules, sudah dibatasi server), bank soal buatan sendiri
  // (GET /admin/readiness-matrix), dan jadwal pengawasan (GET /proctor/schedules).
  const canShowTeacherQuestionSection = computed(() => authStore.canAccessTab('questions'))
  const canShowTeacherProctorSection = computed(() => authStore.canAccessTab('proctor'))

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

  return {
    eventStatusLabel,
    eventStatusBadgeClass,
    eventActionLabel,
    eventActionBtnClass,
    eventActionClick,
    activeLoginSessions,
    loadLoginSessions,
    canShowTeacherQuestionSection,
    canShowTeacherProctorSection,
    teacherHomeIsInactive,
    teacherHomeRoleLabel,
    teacherHomeSchedules,
    teacherHomeTodayScheduleCount,
    teacherHomeUpcomingList,
    teacherHomeSchedulesWithoutBank,
    teacherHomeBanks,
    teacherHomeLockedBankCount,
    teacherHomeDraftBankCount,
    teacherHomeUnfinishedBanks,
    teacherHomeActionItems,
    teacherHomeVisibleActionItems,
    runTeacherHomeAction,
    teacherHomeScheduleBankStatus,
    teacherHomeProctorTasks,
    teacherHomeProctorTodayCount,
    teacherHomeProctorUpcomingList,
    teacherHomeGridClass,
    teacherHomeProctorCardSpanClass,
    loadTeacherHomeProctorTasks,
  }
}
