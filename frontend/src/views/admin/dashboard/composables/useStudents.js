import { ref, computed, watch } from 'vue'
import api from '../../../../services/api'

// Data master siswa: filter, urutan, halaman, form, impor Excel, dan reset sesi login.
export function useStudents(ctx) {
  const {
    classes,
    extractExportErrorMessage,
    showAlertModal,
    showConfirmModal,
    showToast,
    students,
  } = ctx

  // Didefinisikan composable yang dipanggil kemudian; diteruskan saat dipanggil.
  const loadAllData = (...args) => ctx.loadAllData(...args)

  const showStudentModal = ref(false)
  const showImportModal = ref(false)
  const importExcelFile = ref(null)
  const isImporting = ref(false)

  // ---------------- MASTER DATA STATES & COMPUTEDS ----------------

  // Form states for user/class
  const newStudent = ref({ full_name: '', username: '', password: '', nis: '', nisn: '', class_id: '', gender: 'L' })

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
      list = list.filter(st => st.has_active_session)
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
          const lockA = a.has_active_session ? 1 : 0
          const lockB = b.has_active_session ? 1 : 0
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

  const resetStudentSession = async (st) => {
    const confirmed = await showConfirmModal({
      title: 'Akhiri Sesi Login',
      message: `Akhiri sesi login ${st.user?.full_name}? Perangkat yang sedang dipakai akan keluar otomatis dan siswa perlu login ulang.`,
      type: 'warning',
      confirmText: 'Akhiri Sesi',
      cancelText: 'Batal'
    })
    if (!confirmed) return

    try {
      await api.post(`/admin/users/${st.user_id}/reset-session`)
      showToast(`Sesi login ${st.user?.full_name} berhasil diakhiri`, 'success')
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

  const resetUserSession = async (u) => {
    const confirmed = await showConfirmModal({
      title: 'Akhiri Sesi Login',
      message: `Akhiri sesi login ${u.username}? Perangkat yang sedang dipakai akan keluar otomatis dan siswa perlu login ulang.`,
      type: 'warning',
      confirmText: 'Akhiri Sesi',
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

  return {
    showStudentModal,
    showImportModal,
    isImporting,
    newStudent,
    studentSearchQuery,
    selectedStudentGrade,
    studentSortKey,
    studentSortOrder,
    studentCurrentPage,
    studentPerPage,
    showStudentDetailModal,
    selectedStudentDetail,
    showEditStudentModal,
    editStudentForm,
    availableStudentGrades,
    toggleStudentSort,
    resetStudentFilters,
    filteredStudents,
    totalStudentPages,
    paginatedStudents,
    displayedStudentPages,
    setStudentPage,
    resetStudentSession,
    resetUserSession,
    downloadStudentTemplate,
    handleExcelFileSelect,
    submitImportExcel,
    openCreateStudentModal,
    openEditStudent,
    openStudentDetail,
    createStudent,
    updateStudent,
    deleteStudent,
  }
}
