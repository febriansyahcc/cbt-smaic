import { computed, ref } from 'vue'

// Tata letak, label portal, izin tampilan menu, dialog konfirmasi/peringatan, dan toast.
export function useDashboardUi(ctx) {
  const {
    authStore,
  } = ctx

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

  const formatDate = (d) => {
    if (!d) return '-'
    const date = new Date(d)
    return date.toLocaleDateString('id-ID', { day: 'numeric', month: 'short', year: 'numeric' })
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

  return {
    portalLabel,
    portalRoleLabel,
    activeTab,
    isSidebarCollapsed,
    isMobileSidebarOpen,
    currentSectionMeta,
    dialogState,
    showConfirmModal,
    showAlertModal,
    handleDialogConfirm,
    handleDialogCancel,
    toastMessage,
    toastType,
    showToast,
    formatDate,
    canManageSchedules,
    canManageMaster,
    canManageEvents,
    canUploadQuestionBank,
    canManageLoginSessions,
    canImportStudents,
    canReadPeople,
    showWelcomeCard,
    showQuickActions,
    showEventReadiness,
    extractExportErrorMessage,
    handleLogout,
  }
}
