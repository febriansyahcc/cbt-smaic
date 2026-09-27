import { landingTab } from '../../../../utils/access'
import { ref, computed } from 'vue'
import api from '../../../../services/api'

// Event ujian: masuk/keluar cakupan event, pencarian, kartu peserta, serta tambah/ubah/hapus event.
export function useEvents(ctx) {
  const {
    activeTab,
    authStore,
    events,
    isEventDropdownOpen,
    loadEvents,
    loadSchedules,
    selectedEventFilter,
    selectedEventId,
    showAlertModal,
    showConfirmModal,
    showToast,
  } = ctx

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

  const closeEventMenuOnClickOutside = () => {
    if (activeEventMenuId.value !== null) {
      activeEventMenuId.value = null
    }
  }

  return {
    enterEvent,
    exitEventScope,
    eventSearchQuery,
    eventStatusFilter,
    activeEventMenuId,
    showCardPrint,
    cardPrintEvent,
    openCardPrint,
    toggleEventMenu,
    showEventModal,
    isEditEvent,
    eventForm,
    filteredEvents,
    openCreateEvent,
    openEditEvent,
    submitEventForm,
    toggleEventActive,
    deleteEvent,
    closeEventMenuOnClickOutside,
  }
}
