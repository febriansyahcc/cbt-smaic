import { ref, computed, watch } from 'vue'
import api from '../../../../services/api'

// Backup Data (khusus administrator): daftar arsip, buat backup, unduh, dan hapus.
export function useBackups(ctx) {
  const {
    activeTab,
    extractExportErrorMessage,
    showAlertModal,
    showConfirmModal,
    showToast,
  } = ctx

  const backupStatus = ref({ items: [], running: false })
  const isLoadingBackups = ref(false)
  const downloadingBackup = ref('')
  let pollTimer = null

  const backupItems = computed(() => backupStatus.value.items || [])
  const isBackupRunning = computed(() => !!backupStatus.value.running)
  const latestBackup = computed(() => backupItems.value[0] || null)

  const stopPolling = () => {
    if (pollTimer) clearTimeout(pollTimer)
    pollTimer = null
  }

  const loadBackups = async () => {
    stopPolling()
    isLoadingBackups.value = true
    try {
      const res = await api.get('/admin/backups')
      backupStatus.value = res.data.data || { items: [], running: false }
    } catch (err) {
      console.error('Gagal memuat daftar backup', err)
    } finally {
      isLoadingBackups.value = false
    }
    // Selama backup berjalan, periksa ulang setiap 3 detik selama tab masih dibuka.
    if (backupStatus.value.running && activeTab.value === 'backup') {
      pollTimer = setTimeout(loadBackups, 3000)
    }
  }

  const createBackup = async () => {
    if (isBackupRunning.value) return
    try {
      await api.post('/admin/backups')
      backupStatus.value = { ...backupStatus.value, running: true, last_error: '' }
      showToast('Backup sedang dibuat', 'success')
      loadBackups()
    } catch (err) {
      await showAlertModal({
        title: 'Gagal Membuat Backup',
        message: err?.response?.data?.message || 'Backup tidak dapat dimulai',
        type: 'danger'
      })
    }
  }

  const downloadBackup = async (item) => {
    if (downloadingBackup.value) return
    downloadingBackup.value = item.name
    try {
      const res = await api.get(`/admin/backups/${encodeURIComponent(item.name)}/download`, {
        responseType: 'blob'
      })
      const url = window.URL.createObjectURL(new Blob([res.data], { type: 'application/gzip' }))
      const link = document.createElement('a')
      link.href = url
      link.setAttribute('download', item.name)
      document.body.appendChild(link)
      link.click()
      link.remove()
      window.URL.revokeObjectURL(url)
    } catch (err) {
      await showAlertModal({
        title: 'Gagal Mengunduh Backup',
        message: await extractExportErrorMessage(err, 'Berkas backup tidak dapat diunduh'),
        type: 'danger'
      })
    } finally {
      downloadingBackup.value = ''
    }
  }

  const deleteBackup = async (item) => {
    const ok = await showConfirmModal({
      title: 'Hapus Backup',
      message: `Backup ${formatBackupDate(item.created_at)} akan dihapus permanen dari server. Pastikan Anda sudah menyimpan salinannya bila masih diperlukan.`,
      type: 'danger',
      confirmText: 'Hapus'
    })
    if (!ok) return
    try {
      await api.delete(`/admin/backups/${encodeURIComponent(item.name)}`)
      showToast('Backup dihapus', 'success')
      loadBackups()
    } catch (err) {
      await showAlertModal({
        title: 'Gagal Menghapus Backup',
        message: err?.response?.data?.message || 'Backup tidak dapat dihapus',
        type: 'danger'
      })
    }
  }

  const formatBackupSize = (bytes) => {
    if (!bytes && bytes !== 0) return '-'
    const units = ['B', 'KB', 'MB', 'GB']
    let size = bytes
    let i = 0
    while (size >= 1024 && i < units.length - 1) {
      size /= 1024
      i++
    }
    return `${size.toLocaleString('id-ID', { maximumFractionDigits: i === 0 ? 0 : 1 })} ${units[i]}`
  }

  const formatBackupDate = (d) => {
    if (!d) return '-'
    return new Date(d).toLocaleString('id-ID', {
      day: 'numeric', month: 'short', year: 'numeric', hour: '2-digit', minute: '2-digit'
    })
  }

  // Hentikan polling saat pengguna pindah dari tab backup.
  watch(activeTab, (tab) => {
    if (tab !== 'backup') stopPolling()
  })

  return {
    backupStatus,
    backupItems,
    isBackupRunning,
    isLoadingBackups,
    latestBackup,
    downloadingBackup,
    loadBackups,
    createBackup,
    downloadBackup,
    deleteBackup,
    formatBackupSize,
    formatBackupDate,
  }
}
