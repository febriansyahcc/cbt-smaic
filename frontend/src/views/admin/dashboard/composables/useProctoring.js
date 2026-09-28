import { ref, computed } from 'vue'
import api from '../../../../services/api'
import { useAuthStore } from '../../../../stores/auth'

// Pengawasan live: jadwal pengawasan, status peserta, serta ekspor nilai (Excel) dan berita acara (PDF).
export function useProctoring(ctx) {
  const {
    extractExportErrorMessage,
    showAlertModal,
  } = ctx
  const authStore = useAuthStore()

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
      const proctorName = authStore.user?.full_name || ''
      const res = await api.get(`/proctor/reports/pdf/${activeProctorScheduleId.value}`, {
        responseType: 'blob',
        params: { proctor: proctorName },
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

  return {
    proctorSchedules,
    activeProctorScheduleId,
    proctorData,
    isExportingNilai,
    showProctorPrint,
    isExportingPDF,
    loadProctorSchedules,
    exportNilaiExcel,
    exportBeritaAcaraPDF,
  }
}
