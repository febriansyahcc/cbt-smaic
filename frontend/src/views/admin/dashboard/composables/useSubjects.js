import { ref, computed, watch } from 'vue'
import api from '../../../../services/api'

// Mata pelajaran dan alokasi kelas-mapel: filter, form, dan impor Excel.
export function useSubjects(ctx) {
  const {
    activeExamEvent,
    classSubjects,
    classes,
    loadClassSubjects,
    loadSubjects,
    showAlertModal,
    showConfirmModal,
    showToast,
    subjects,
    teachers,
  } = ctx

  // Didefinisikan composable yang dipanggil kemudian; diteruskan saat dipanggil.
  const loadAllData = (...args) => ctx.loadAllData(...args)

  const showSubjectModal = ref(false)
  const isEditSubject = ref(false)
  const subjectForm = ref({ id: null, code: '', name: '' })
  const showClassSubjectModal = ref(false)
  const classSubjectForm = ref({
    id: null,
    class_id: '',
    subject_id: '',
    teacher_id: '',
    academic_year: '2026/2027',
  })
  const filterClassSubjectClass = ref('')

  // Import states — Mata Pelajaran
  const showImportSubjectModal = ref(false)
  const importSubjectFile = ref(null)
  const isImportingSubject = ref(false)
  const importSubjectResult = ref(null)

  // 4. Mata Pelajaran Filtering, Sorting & Pagination
  const subjectSearchQuery = ref('')
  const subjectSortKey = ref('name')
  const subjectSortOrder = ref('asc')
  const subjectCurrentPage = ref(1)
  const subjectPerPage = ref(10)

  const toggleSubjectSort = (key) => {
    if (subjectSortKey.value === key) {
      subjectSortOrder.value = subjectSortOrder.value === 'asc' ? 'desc' : 'asc'
    } else {
      subjectSortKey.value = key
      subjectSortOrder.value = 'asc'
    }
  }

  const filteredSubjects = computed(() => {
    let list = subjects.value || []
    if (subjectSearchQuery.value.trim()) {
      const q = subjectSearchQuery.value.toLowerCase().trim()
      list = list.filter(s => {
        const codeMatch = s.code && s.code.toLowerCase().includes(q)
        const nameMatch = s.name && s.name.toLowerCase().includes(q)
        return codeMatch || nameMatch
      })
    }
    if (subjectSortKey.value) {
      list = [...list].sort((a, b) => {
        let result = 0
        if (subjectSortKey.value === 'code') {
          const valA = (a.code || '').trim().toLowerCase()
          const valB = (b.code || '').trim().toLowerCase()
          result = valA.localeCompare(valB, undefined, { numeric: true, sensitivity: 'base' })
        } else if (subjectSortKey.value === 'name') {
          const valA = (a.name || '').trim().toLowerCase()
          const valB = (b.name || '').trim().toLowerCase()
          result = valA.localeCompare(valB, undefined, { numeric: true, sensitivity: 'base' })
        }
        return subjectSortOrder.value === 'asc' ? result : -result
      })
    }
    return list
  })

  const totalSubjectPages = computed(() => Math.ceil(filteredSubjects.value.length / subjectPerPage.value) || 1)
  const paginatedSubjects = computed(() => {
    const start = (subjectCurrentPage.value - 1) * subjectPerPage.value
    return filteredSubjects.value.slice(start, start + subjectPerPage.value)
  })

  const displayedSubjectPages = computed(() => {
    const total = totalSubjectPages.value
    const current = subjectCurrentPage.value
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

  const setSubjectPage = (p) => {
    if (p === '...' || p < 1 || p > totalSubjectPages.value) return
    subjectCurrentPage.value = p
  }

  watch([subjectSearchQuery, subjectPerPage, subjectSortKey, subjectSortOrder], () => {
    subjectCurrentPage.value = 1
  })

  const openCreateSubject = () => {
    isEditSubject.value = false
    subjectForm.value = { id: null, code: '', name: '' }
    showSubjectModal.value = true
  }

  const openEditSubject = (s) => {
    isEditSubject.value = true
    subjectForm.value = { id: s.id, code: s.code, name: s.name }
    showSubjectModal.value = true
  }

  const submitSubjectForm = async () => {
    if (!subjectForm.value.code || !subjectForm.value.name) {
      await showAlertModal({
        title: 'Form Belum Lengkap',
        message: 'Kode dan Nama Mata Pelajaran wajib diisi.',
        type: 'warning'
      })
      return
    }
    try {
      if (isEditSubject.value) {
        await api.put(`/admin/subjects/${subjectForm.value.id}`, subjectForm.value)
        showToast('Mata pelajaran berhasil diperbarui!', 'success')
      } else {
        await api.post('/admin/subjects', subjectForm.value)
        showToast('Mata pelajaran berhasil ditambahkan!', 'success')
      }
      showSubjectModal.value = false
      await loadSubjects()
      await loadAllData()
    } catch (e) {
      await showAlertModal({
        title: 'Gagal Menyimpan Mapel',
        message: e.response?.data?.message || 'Terjadi kesalahan saat menyimpan mata pelajaran.',
        type: 'danger'
      })
    }
  }

  const deleteSubject = async (s) => {
    const confirmed = await showConfirmModal({
      title: 'Hapus Mata Pelajaran',
      message: `Yakin ingin menghapus mata pelajaran ${s.name} (${s.code})?`,
      type: 'danger',
      confirmText: 'Hapus',
      cancelText: 'Batal'
    })
    if (!confirmed) return

    try {
      await api.delete(`/admin/subjects/${s.id}`)
      await loadSubjects()
      await loadAllData()
      showToast('Mata pelajaran berhasil dihapus', 'success')
    } catch (e) {
      await showAlertModal({
        title: 'Gagal Menghapus Mapel',
        message: e.response?.data?.message || 'Terjadi kesalahan saat menghapus mata pelajaran.',
        type: 'danger'
      })
    }
  }

  // Class Subjects Search, Filter, Sort, Pagination
  const classSubjectSearchQuery = ref('')
  const selectedClassSubjectGrade = ref('all')
  const selectedClassSubjectClassId = ref('')
  const classSubjectSortKey = ref('class')
  const classSubjectSortOrder = ref('asc')
  const classSubjectCurrentPage = ref(1)
  const classSubjectPerPage = ref(10)

  const toggleClassSubjectSort = (key) => {
    if (classSubjectSortKey.value === key) {
      classSubjectSortOrder.value = classSubjectSortOrder.value === 'asc' ? 'desc' : 'asc'
    } else {
      classSubjectSortKey.value = key
      classSubjectSortOrder.value = 'asc'
    }
  }

  const filteredClassSubjects = computed(() => {
    let list = classSubjects.value || []

    // Filter by Grade
    if (selectedClassSubjectGrade.value !== 'all') {
      list = list.filter(cs => {
        const cls = cs.class_room || classes.value.find(c => c.id === cs.class_id)
        return cls && cls.grade === selectedClassSubjectGrade.value
      })
    }

    // Filter by Specific Class ID
    if (selectedClassSubjectClassId.value) {
      list = list.filter(cs => cs.class_id === selectedClassSubjectClassId.value || cs.class_room?.id === selectedClassSubjectClassId.value)
    }

    // Filter by Search Query
    if (classSubjectSearchQuery.value.trim()) {
      const q = classSubjectSearchQuery.value.toLowerCase().trim()
      list = list.filter(cs => {
        const clsName = (cs.class_room?.name || '').toLowerCase()
        const clsGrade = (cs.class_room?.grade || '').toLowerCase()
        const subName = (cs.subject?.name || '').toLowerCase()
        const subCode = (cs.subject?.code || '').toLowerCase()
        const teacherName = (cs.teacher?.full_name || '').toLowerCase()
        const teacherUser = (cs.teacher?.username || '').toLowerCase()
        const academicYear = (cs.academic_year || '').toLowerCase()
        return clsName.includes(q) || clsGrade.includes(q) || subName.includes(q) || subCode.includes(q) || teacherName.includes(q) || teacherUser.includes(q) || academicYear.includes(q)
      })
    }

    // Sorting
    if (classSubjectSortKey.value) {
      list = [...list].sort((a, b) => {
        let result = 0
        if (classSubjectSortKey.value === 'class') {
          const nameA = (a.class_room?.name || '').trim().toLowerCase()
          const nameB = (b.class_room?.name || '').trim().toLowerCase()
          result = nameA.localeCompare(nameB, undefined, { numeric: true, sensitivity: 'base' })
        } else if (classSubjectSortKey.value === 'subject') {
          const nameA = (a.subject?.name || '').trim().toLowerCase()
          const nameB = (b.subject?.name || '').trim().toLowerCase()
          result = nameA.localeCompare(nameB, undefined, { numeric: true, sensitivity: 'base' })
        } else if (classSubjectSortKey.value === 'teacher') {
          const nameA = (a.teacher?.full_name || '').trim().toLowerCase()
          const nameB = (b.teacher?.full_name || '').trim().toLowerCase()
          result = nameA.localeCompare(nameB, undefined, { numeric: true, sensitivity: 'base' })
        } else if (classSubjectSortKey.value === 'year') {
          const valA = (a.academic_year || '').trim().toLowerCase()
          const valB = (b.academic_year || '').trim().toLowerCase()
          result = valA.localeCompare(valB, undefined, { numeric: true, sensitivity: 'base' })
        }
        return classSubjectSortOrder.value === 'asc' ? result : -result
      })
    }

    return list
  })

  const totalClassSubjectPages = computed(() => {
    return Math.max(1, Math.ceil(filteredClassSubjects.value.length / classSubjectPerPage.value))
  })

  const paginatedClassSubjects = computed(() => {
    const start = (classSubjectCurrentPage.value - 1) * classSubjectPerPage.value
    return filteredClassSubjects.value.slice(start, start + classSubjectPerPage.value)
  })

  const displayedClassSubjectPages = computed(() => {
    const total = totalClassSubjectPages.value
    const current = classSubjectCurrentPage.value
    if (total <= 7) {
      return Array.from({ length: total }, (_, i) => i + 1)
    }
    if (current <= 4) {
      return [1, 2, 3, 4, 5, '...', total]
    } else if (current >= total - 3) {
      return [1, '...', total - 4, total - 3, total - 2, total - 1, total]
    } else {
      return [1, '...', current - 1, current, current + 1, '...', total]
    }
  })

  const setClassSubjectPage = (p) => {
    if (p === '...' || p < 1 || p > totalClassSubjectPages.value) return
    classSubjectCurrentPage.value = p
  }

  const isEditClassSubject = ref(false)

  const openCreateClassSubject = () => {
    isEditClassSubject.value = false
    classSubjectForm.value = {
      id: null,
      class_id: classes.value[0]?.id || '',
      subject_id: subjects.value[0]?.id || '',
      teacher_id: teachers.value[0]?.id || '',
      academic_year: activeExamEvent.value?.academic_year || '2026/2027',
    }
    showClassSubjectModal.value = true
  }

  const openEditClassSubject = (cs) => {
    isEditClassSubject.value = true
    classSubjectForm.value = {
      id: cs.id,
      class_id: cs.class_id || cs.class_room?.id || '',
      subject_id: cs.subject_id || cs.subject?.id || '',
      teacher_id: cs.teacher_id || cs.teacher?.id || '',
      academic_year: cs.academic_year || activeExamEvent.value?.academic_year || '2026/2027',
    }
    showClassSubjectModal.value = true
  }

  const submitClassSubjectForm = async () => {
    if (!classSubjectForm.value.class_id || !classSubjectForm.value.subject_id || !classSubjectForm.value.teacher_id) {
      await showAlertModal({
        title: 'Form Belum Lengkap',
        message: 'Kelas, Mata Pelajaran, dan Guru Pengampu wajib dipilih.',
        type: 'warning'
      })
      return
    }
    try {
      if (isEditClassSubject.value && classSubjectForm.value.id) {
        await api.put(`/admin/class-subjects/${classSubjectForm.value.id}`, classSubjectForm.value)
        showToast('Alokasi kelas mapel berhasil diperbarui!', 'success')
      } else {
        await api.post('/admin/class-subjects', classSubjectForm.value)
        showToast('Alokasi kelas mapel berhasil disimpan!', 'success')
      }
      showClassSubjectModal.value = false
      await loadClassSubjects()
      await loadAllData()
    } catch (e) {
      await showAlertModal({
        title: isEditClassSubject.value ? 'Gagal Memperbarui Alokasi' : 'Gagal Menyimpan Alokasi',
        message: e.response?.data?.message || 'Terjadi kesalahan saat menyimpan alokasi kelas mapel.',
        type: 'danger'
      })
    }
  }

  const deleteClassSubject = async (cs) => {
    const confirmed = await showConfirmModal({
      title: 'Hapus Alokasi Mapel',
      message: `Hapus alokasi ${cs.subject?.name} untuk kelas ${cs.class_room?.name}?`,
      type: 'danger',
      confirmText: 'Hapus Alokasi',
      cancelText: 'Batal'
    })
    if (!confirmed) return

    try {
      await api.delete(`/admin/class-subjects/${cs.id}`)
      await loadClassSubjects()
      await loadAllData()
      showToast('Alokasi kelas mapel berhasil dihapus', 'success')
    } catch (e) {
      await showAlertModal({
        title: 'Gagal Menghapus Alokasi',
        message: e.response?.data?.message || 'Terjadi kesalahan saat menghapus alokasi.',
        type: 'danger'
      })
    }
  }

  const downloadSubjectTemplate = async () => {
    try {
      const res = await api.get('/admin/subjects/template', { responseType: 'blob' })
      const url = window.URL.createObjectURL(new Blob([res.data], {
        type: 'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet'
      }))
      const a = document.createElement('a')
      a.href = url
      a.download = 'Template_Import_Mapel_CBT.xlsx'
      document.body.appendChild(a)
      a.click()
      a.remove()
      window.URL.revokeObjectURL(url)
      showToast('Template mata pelajaran berhasil diunduh!', 'success')
    } catch (err) {
      await showAlertModal({ title: 'Gagal Mengunduh Template', message: 'Gagal mengunduh template mata pelajaran.', type: 'danger' })
    }
  }

  const submitImportSubjects = async () => {
    if (!importSubjectFile.value) return
    isImportingSubject.value = true
    importSubjectResult.value = null
    try {
      const fd = new FormData()
      fd.append('file', importSubjectFile.value)
      const res = await api.post('/admin/subjects/import', fd, {
        headers: { 'Content-Type': 'multipart/form-data' },
      })
      importSubjectResult.value = { success: true, message: res.data.message || 'Impor mata pelajaran berhasil!' }
      await loadAllData()
    } catch (e) {
      importSubjectResult.value = { success: false, message: e.response?.data?.message || 'Gagal mengunggah file' }
    } finally {
      isImportingSubject.value = false
    }
  }

  return {
    showSubjectModal,
    isEditSubject,
    subjectForm,
    showClassSubjectModal,
    classSubjectForm,
    showImportSubjectModal,
    importSubjectFile,
    isImportingSubject,
    importSubjectResult,
    subjectSearchQuery,
    subjectSortKey,
    subjectSortOrder,
    subjectCurrentPage,
    subjectPerPage,
    toggleSubjectSort,
    filteredSubjects,
    totalSubjectPages,
    paginatedSubjects,
    displayedSubjectPages,
    setSubjectPage,
    openCreateSubject,
    openEditSubject,
    submitSubjectForm,
    deleteSubject,
    classSubjectSearchQuery,
    selectedClassSubjectGrade,
    selectedClassSubjectClassId,
    classSubjectSortKey,
    classSubjectSortOrder,
    classSubjectCurrentPage,
    classSubjectPerPage,
    toggleClassSubjectSort,
    filteredClassSubjects,
    totalClassSubjectPages,
    paginatedClassSubjects,
    displayedClassSubjectPages,
    setClassSubjectPage,
    isEditClassSubject,
    openCreateClassSubject,
    openEditClassSubject,
    submitClassSubjectForm,
    deleteClassSubject,
    downloadSubjectTemplate,
    submitImportSubjects,
  }
}
