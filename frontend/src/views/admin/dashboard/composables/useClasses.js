import { ref, computed, watch } from 'vue'
import api from '../../../../services/api'

// Data master kelas/rombel: filter, form, dan impor Excel.
export function useClasses(ctx) {
  const {
    classes,
    showAlertModal,
    showConfirmModal,
    showToast,
  } = ctx

  // Didefinisikan composable yang dipanggil kemudian; diteruskan saat dipanggil.
  const loadAllData = (...args) => ctx.loadAllData(...args)

  const showClassModal = ref(false)

  // Import states — Kelas
  const showImportClassModal = ref(false)
  const importClassFile = ref(null)
  const isImportingClass = ref(false)
  const importClassResult = ref(null)
  const newClass = ref({ name: '', grade: 'XII', major: 'MIPA' })

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

  return {
    showClassModal,
    showImportClassModal,
    importClassFile,
    isImportingClass,
    importClassResult,
    newClass,
    classSearchQuery,
    selectedClassGrade,
    classSortKey,
    classSortOrder,
    classCurrentPage,
    classPerPage,
    showEditClassModal,
    editClassForm,
    toggleClassSort,
    filteredClasses,
    totalClassPages,
    paginatedClasses,
    displayedClassPages,
    setClassPage,
    downloadClassTemplate,
    submitImportClasses,
    openCreateClassModal,
    openEditClass,
    createClass,
    updateClass,
    deleteClass,
  }
}
