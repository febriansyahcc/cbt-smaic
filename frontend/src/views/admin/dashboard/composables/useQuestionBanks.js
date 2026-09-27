import { computed, ref } from 'vue'
import api from '../../../../services/api'

// Bank soal: cakupan kelas, filter, buat/ubah/kunci/hapus, unggah Excel, dan cetak naskah.
export function useQuestionBanks(ctx) {
  const {
    authStore,
    classes,
    currentEventSchedules,
    extractExportErrorMessage,
    readinessData,
    scheduleClassIdOf,
    scheduleGradeOf,
    scheduleSubjectOf,
    selectedExamEvent,
    showAlertModal,
    showConfirmModal,
    showToast,
    subjects,
  } = ctx

  // Didefinisikan composable yang dipanggil kemudian; diteruskan saat dipanggil.
  const loadAllData = (...args) => ctx.loadAllData(...args)

  const canLockQuestionBanks = computed(() => authStore.hasPermission('questions:lock'))
  const canPrintQuestionBanks = computed(() => authStore.hasPermission('questions:print'))

  // Cetak naskah soal & kunci jawaban (hanya bank yang sudah terkunci).
  const showQuestionPrint = ref(false)
  const questionPrintBank = ref(null)
  const openQuestionPrint = (bank) => {
    questionPrintBank.value = bank
    showQuestionPrint.value = true
  }

  // Bank Soal States & Logic (Unified Bank Soal & Kesiapan)
  const questionBankFile = ref(null)
  const selectedBankUploadId = ref('')
  const isUploadingQuestionBank = ref(false)
  const questionBankUploadStatus = ref(null)
  const showUploadBankModal = ref(false)
  const showCreateBankModal = ref(false)
  const isCreatingBank = ref(false)
  const isEditBank = ref(false)
  const editingBankId = ref(null)
  const questionBankSearchQuery = ref('')
  const questionBankStatusFilter = ref('all')
  const newBankForm = ref({
    scope: 'grade', // 'grade' = per angkatan (default), 'classes' = kelas tertentu
    subject_id: '',
    grade: '',
    class_ids: [],
    title: ''
  })

  const bankScopeOptions = [
    { value: 'grade', label: 'Per Angkatan' },
    { value: 'classes', label: 'Pilih Kelas' }
  ]

  const bankGradeOrder = { X: 1, XI: 2, XII: 3 }

  // Kunci opsi dropdown = mapel + tingkat, karena Informatika XI dan XII butuh naskah berbeda.
  const bankOptionKey = (subjectId, grade) => `${subjectId}|${grade || ''}`

  const newBankOptionKey = computed({
    get: () => newBankForm.value.subject_id ? bankOptionKey(newBankForm.value.subject_id, newBankForm.value.grade) : '',
    set: (key) => {
      const [subjectId, grade = ''] = (key || '').split('|')
      newBankForm.value.subject_id = subjectId || ''
      newBankForm.value.grade = grade
    }
  })

  const bankScopeLabel = (b) => {
    if (b.classes?.length) return b.classes.map(c => c.name).sort((x, y) => x.localeCompare(y, undefined, { numeric: true })).join(', ')
    return b.grade ? `Kelas ${b.grade}` : ''
  }

  // Kelas yang sudah tercakup bank soal lain, per mapel: subject_id -> Set(class_id)
  const bankCoveredClassIds = computed(() => {
    const scheds = currentEventSchedules.value || []
    const covered = new Map()
    ;(readinessData.value.question_banks || [])
      .filter(b => isEditBank.value ? b.id !== editingBankId.value : true)
      .forEach(b => {
        const subId = b.subject_id || b.subject?.id
        if (!subId) return
        if (!covered.has(subId)) covered.set(subId, new Set())
        const set = covered.get(subId)
        if (b.classes?.length) {
          b.classes.forEach(c => set.add(c.id))
        } else if (b.grade) {
          classes.value.filter(c => c.grade === b.grade).forEach(c => set.add(c.id))
        } else {
          // Bank lama tanpa cakupan: anggap mencakup kelas jadwal yang sudah ditautkan kepadanya.
          scheds.filter(s => s.bank_id === b.id).forEach(s => set.add(scheduleClassIdOf(s)))
        }
      })
    return covered
  })

  const availableBankSubjects = computed(() => {
    const scheds = currentEventSchedules.value || []
    const covered = bankCoveredClassIds.value

    const map = new Map()
    const addOption = (sub, grade) => {
      const key = bankOptionKey(sub.id, grade)
      if (!map.has(key)) map.set(key, { key, id: sub.id, name: sub.name, code: sub.code, grade })
    }

    if (scheds.length > 0) {
      // Opsi angkatan muncul selama masih ada kelas terjadwal di angkatan itu yang belum punya bank soal.
      scheds.forEach(s => {
        const sub = scheduleSubjectOf(s)
        if (sub && !covered.get(sub.id)?.has(scheduleClassIdOf(s))) addOption(sub, scheduleGradeOf(s))
      })
    } else {
      (subjects.value || []).filter(sub => !covered.has(sub.id)).forEach(sub => addOption(sub, ''))
    }

    // Saat edit, pastikan pilihan bank yang sedang disunting tetap ada
    if (isEditBank.value && editingBankId.value) {
      const currentBank = (readinessData.value.question_banks || []).find(b => b.id === editingBankId.value)
      if (currentBank?.subject && !currentBank.classes?.length) addOption(currentBank.subject, currentBank.grade || '')
    }

    return Array.from(map.values())
      .sort((a, b) => a.name.localeCompare(b.name) || (bankGradeOrder[a.grade] || 9) - (bankGradeOrder[b.grade] || 9))
  })

  // Mode "Pilih Kelas": mapel dari jadwal event (atau semua mapel bila belum ada jadwal)
  const bankCustomSubjects = computed(() => {
    const map = new Map()
    ;(currentEventSchedules.value || []).forEach(s => {
      const sub = scheduleSubjectOf(s)
      if (sub) map.set(sub.id, sub)
    })
    const currentBank = isEditBank.value && (readinessData.value.question_banks || []).find(b => b.id === editingBankId.value)
    if (currentBank?.subject) map.set(currentBank.subject.id, currentBank.subject)
    const list = map.size > 0 ? Array.from(map.values()) : (subjects.value || [])
    return [...list].sort((a, b) => a.name.localeCompare(b.name))
  })

  const bankCustomClassOptions = computed(() => {
    const subId = newBankForm.value.subject_id
    if (!subId) return []
    const ids = new Set(
      (currentEventSchedules.value || [])
        .filter(s => scheduleSubjectOf(s)?.id === subId)
        .map(scheduleClassIdOf)
    )
    // Pertahankan kelas yang sudah dipilih (mis. saat edit) meski tidak ada di jadwal
    newBankForm.value.class_ids.forEach(id => ids.add(id))
    const list = ids.size > 0 ? classes.value.filter(c => ids.has(c.id)) : classes.value
    const covered = bankCoveredClassIds.value.get(subId) || new Set()
    return list
      .map(c => ({ id: c.id, name: c.name, grade: c.grade, covered: covered.has(c.id) && !newBankForm.value.class_ids.includes(c.id) }))
      .sort((a, b) => (bankGradeOrder[a.grade] || 9) - (bankGradeOrder[b.grade] || 9) || a.name.localeCompare(b.name, undefined, { numeric: true }))
  })

  const bankFormLocked = computed(() =>
    !isEditBank.value && newBankForm.value.scope === 'grade' && availableBankSubjects.value.length === 0
  )

  const bankEventPrefix = () => selectedExamEvent.value?.title ? `${selectedExamEvent.value.title} - ` : ''

  const bankTitleFor = (opt) => `${bankEventPrefix()}${opt.name}${opt.grade ? ` Kelas ${opt.grade}` : ''}`

  const refreshCustomBankTitle = () => {
    const sub = bankCustomSubjects.value.find(s => s.id === newBankForm.value.subject_id)
    if (!sub) return
    const names = bankCustomClassOptions.value.filter(c => newBankForm.value.class_ids.includes(c.id)).map(c => c.name)
    newBankForm.value.title = `${bankEventPrefix()}${sub.name}${names.length ? ` ${names.join(', ')}` : ''}`
  }

  const onBankSubjectChange = () => {
    const selOpt = availableBankSubjects.value.find(o => o.key === newBankOptionKey.value)
    if (selOpt) newBankForm.value.title = bankTitleFor(selOpt)
  }

  const onBankCustomSubjectChange = () => {
    newBankForm.value.class_ids = []
    refreshCustomBankTitle()
  }

  const toggleBankClass = (id) => {
    const ids = newBankForm.value.class_ids
    newBankForm.value.class_ids = ids.includes(id) ? ids.filter(x => x !== id) : [...ids, id]
    refreshCustomBankTitle()
  }

  const setBankScope = (scope) => {
    if (newBankForm.value.scope === scope) return
    newBankForm.value.scope = scope
    newBankForm.value.class_ids = []
    if (scope === 'grade') {
      const opt = availableBankSubjects.value.find(o => o.id === newBankForm.value.subject_id) || availableBankSubjects.value[0]
      newBankForm.value.subject_id = opt?.id || ''
      newBankForm.value.grade = opt?.grade || ''
      if (opt) newBankForm.value.title = bankTitleFor(opt)
    } else {
      newBankForm.value.grade = ''
      refreshCustomBankTitle()
    }
  }

  const questionBankSortKey = ref('title')
  const questionBankSortOrder = ref('asc')
  const questionBankCurrentPage = ref(1)
  const questionBankPerPage = ref(10)

  const toggleQuestionBankSort = (key) => {
    if (questionBankSortKey.value === key) {
      questionBankSortOrder.value = questionBankSortOrder.value === 'asc' ? 'desc' : 'asc'
    } else {
      questionBankSortKey.value = key
      questionBankSortOrder.value = 'asc'
    }
  }

  const filteredQuestionBanks = computed(() => {
    let list = readinessData.value.question_banks || []
    if (questionBankStatusFilter.value === 'locked') {
      list = list.filter(b => b.is_locked)
    } else if (questionBankStatusFilter.value === 'draft') {
      list = list.filter(b => !b.is_locked)
    }

    if (questionBankSearchQuery.value.trim()) {
      const q = questionBankSearchQuery.value.toLowerCase().trim()
      list = list.filter(b => {
        const title = (b.title || '').toLowerCase()
        const subjectName = (b.subject?.name || '').toLowerCase()
        const subjectCode = (b.subject?.code || '').toLowerCase()
        const creatorName = (b.created_by?.full_name || '').toLowerCase()
        return title.includes(q) || subjectName.includes(q) || subjectCode.includes(q) || creatorName.includes(q)
      })
    }

    // Sort
    if (questionBankSortKey.value) {
      list = [...list].sort((a, b) => {
        let result = 0
        if (questionBankSortKey.value === 'title') {
          const titleA = (a.title || '').trim().toLowerCase()
          const titleB = (b.title || '').trim().toLowerCase()
          result = titleA.localeCompare(titleB, undefined, { numeric: true, sensitivity: 'base' })
          if (result === 0) {
            const subA = (a.subject?.name || '').trim().toLowerCase()
            const subB = (b.subject?.name || '').trim().toLowerCase()
            result = subA.localeCompare(subB, undefined, { numeric: true, sensitivity: 'base' })
          }
        } else if (questionBankSortKey.value === 'creator') {
          const nameA = (a.created_by?.full_name || '').trim().toLowerCase()
          const nameB = (b.created_by?.full_name || '').trim().toLowerCase()
          result = nameA.localeCompare(nameB, undefined, { numeric: true, sensitivity: 'base' })
        } else if (questionBankSortKey.value === 'questions') {
          const countA = a.total_questions || 0
          const countB = b.total_questions || 0
          result = countA - countB
        } else if (questionBankSortKey.value === 'status') {
          const lockA = a.is_locked ? 1 : 0
          const lockB = b.is_locked ? 1 : 0
          result = lockA - lockB
        }
        return questionBankSortOrder.value === 'asc' ? result : -result
      })
    }

    return list
  })

  const totalQuestionBankPages = computed(() => {
    return Math.max(1, Math.ceil(filteredQuestionBanks.value.length / questionBankPerPage.value))
  })

  const paginatedQuestionBanks = computed(() => {
    const start = (questionBankCurrentPage.value - 1) * questionBankPerPage.value
    return filteredQuestionBanks.value.slice(start, start + questionBankPerPage.value)
  })

  const displayedQuestionBankPages = computed(() => {
    const total = totalQuestionBankPages.value
    const current = questionBankCurrentPage.value
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

  const setQuestionBankPage = (p) => {
    if (p === '...' || p < 1 || p > totalQuestionBankPages.value) return
    questionBankCurrentPage.value = p
  }

  const downloadQuestionBankTemplate = async () => {
    try {
      const res = await api.get('/admin/template/question-bank.xlsx', {
        responseType: 'blob'
      })
      const url = window.URL.createObjectURL(new Blob([res.data], {
        type: 'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet'
      }))
      const link = document.createElement('a')
      link.href = url
      link.setAttribute('download', 'Template_Bank_Soal_CBT.xlsx')
      document.body.appendChild(link)
      link.click()
      link.remove()
      window.URL.revokeObjectURL(url)
      showToast('Format template Excel berhasil diunduh!', 'success')
    } catch (err) {
      await showAlertModal({
        title: 'Gagal Mengunduh Template',
        message: await extractExportErrorMessage(err, 'Gagal mengunduh template'),
        type: 'danger'
      })
    }
  }

  const handleQuestionBankFileChange = (e) => {
    questionBankFile.value = e.target.files[0]
  }

  const openUploadBankModal = (bank = null) => {
    if (bank) {
      selectedBankUploadId.value = bank.id
    } else if (!selectedBankUploadId.value && (readinessData.value.question_banks || []).length > 0) {
      selectedBankUploadId.value = readinessData.value.question_banks[0].id
    }
    questionBankFile.value = null
    questionBankUploadStatus.value = null
    showUploadBankModal.value = true
  }

  const openCreateBankModal = () => {
    isEditBank.value = false
    editingBankId.value = null
    const defaultOpt = availableBankSubjects.value[0]
    newBankForm.value = {
      scope: 'grade',
      subject_id: defaultOpt?.id || '',
      grade: defaultOpt?.grade || '',
      class_ids: [],
      title: defaultOpt ? bankTitleFor(defaultOpt) : ''
    }
    showCreateBankModal.value = true
  }

  const openEditBankModal = (bank) => {
    isEditBank.value = true
    editingBankId.value = bank.id
    newBankForm.value = {
      scope: bank.classes?.length ? 'classes' : 'grade',
      subject_id: bank.subject_id || bank.subject?.id || '',
      grade: bank.grade || '',
      class_ids: (bank.classes || []).map(c => c.id),
      title: bank.title || ''
    }
    showCreateBankModal.value = true
  }

  const submitCreateBank = async () => {
    if (!newBankForm.value.subject_id || !newBankForm.value.title.trim()) {
      await showAlertModal({
        title: 'Form Belum Lengkap',
        message: 'Mata pelajaran dan judul bank soal wajib diisi.',
        type: 'warning'
      })
      return
    }
    const isCustomScope = newBankForm.value.scope === 'classes'
    if (isCustomScope && newBankForm.value.class_ids.length === 0) {
      await showAlertModal({
        title: 'Kelas Belum Dipilih',
        message: 'Pilih minimal satu kelas untuk bank soal dengan cakupan kelas tertentu.',
        type: 'warning'
      })
      return
    }
    const scopePayload = {
      grade: isCustomScope ? '' : newBankForm.value.grade,
      class_ids: isCustomScope ? newBankForm.value.class_ids : []
    }

    isCreatingBank.value = true
    try {
      if (isEditBank.value && editingBankId.value) {
        await api.put(`/admin/question-banks/${editingBankId.value}`, {
          subject_id: newBankForm.value.subject_id,
          ...scopePayload,
          title: newBankForm.value.title.trim()
        })
        showToast('Bank soal berhasil diperbarui!', 'success')
      } else {
        const res = await api.post('/admin/question-banks', {
          subject_id: newBankForm.value.subject_id,
          ...scopePayload,
          title: newBankForm.value.title.trim()
        })
        showToast('Bank soal baru berhasil dibuat!', 'success')
        if (res.data.data?.id) {
          selectedBankUploadId.value = res.data.data.id
        }
      }
      showCreateBankModal.value = false
      await loadAllData()
    } catch (err) {
      await showAlertModal({
        title: isEditBank.value ? 'Gagal Memperbarui Bank Soal' : 'Gagal Membuat Bank Soal',
        message: err.response?.data?.message || 'Terjadi kesalahan saat menyimpan bank soal.',
        type: 'danger'
      })
    } finally {
      isCreatingBank.value = false
    }
  }

  const deleteQuestionBank = async (bank) => {
    const confirmed = await showConfirmModal({
      title: 'Hapus Bank Soal',
      message: `Apakah Anda yakin ingin menghapus bank soal "${bank.title}" (${bank.subject?.name || ''}) beserta seluruh butir soalnya?`,
      type: 'danger',
      confirmText: 'Hapus Bank Soal',
      cancelText: 'Batal'
    })
    if (!confirmed) return

    try {
      await api.delete(`/admin/question-banks/${bank.id}`)
      showToast('Bank soal berhasil dihapus!', 'success')
      await loadAllData()
    } catch (err) {
      await showAlertModal({
        title: 'Gagal Menghapus Bank Soal',
        message: err.response?.data?.message || 'Terjadi kesalahan saat menghapus bank soal.',
        type: 'danger'
      })
    }
  }

  const toggleBankLockWithConfirm = async (bank) => {
    const willLock = !bank.is_locked
    const actionTitle = willLock ? 'Kunci Naskah Bank Soal' : 'Buka Kunci Naskah Bank Soal'
    const actionMessage = willLock
      ? `Kunci naskah "${bank.title}"? Setelah dikunci, naskah dinyatakan Siap Ujian dan butir soal siap disajikan kepada peserta ujian.`
      : `Buka kunci naskah "${bank.title}"? Status naskah akan kembali menjadi Draft Terbuka.`

    const confirmed = await showConfirmModal({
      title: actionTitle,
      message: actionMessage,
      type: willLock ? 'confirm' : 'warning',
      confirmText: willLock ? 'Kunci Naskah' : 'Buka Kunci',
      cancelText: 'Batal'
    })
    if (!confirmed) return

    try {
      await api.post(`/admin/question-banks/${bank.id}/toggle-lock`)
      showToast(willLock ? 'Naskah soal berhasil dikunci (Terkunci)!' : 'Kunci naskah soal berhasil dibuka (Draft)!', 'success')
      await loadAllData()
    } catch (err) {
      await showAlertModal({
        title: 'Gagal Mengubah Status Kunci',
        message: err.response?.data?.message || 'Terjadi kesalahan saat mengubah status penguncian bank soal.',
        type: 'danger'
      })
    }
  }

  const submitQuestionBankUpload = async () => {
    if (!selectedBankUploadId.value) {
      await showAlertModal({
        title: 'Pilih Bank Soal',
        message: 'Pilih bank soal tujuan terlebih dahulu.',
        type: 'warning'
      })
      return
    }
    if (!questionBankFile.value) {
      await showAlertModal({
        title: 'Pilih File Soal',
        message: 'Pilih file Excel (.xlsx) terlebih dahulu.',
        type: 'warning'
      })
      return
    }
    isUploadingQuestionBank.value = true
    questionBankUploadStatus.value = null
    const formData = new FormData()
    formData.append('bank_id', selectedBankUploadId.value)
    formData.append('file', questionBankFile.value)
    try {
      const res = await api.post('/admin/questions/import', formData, {
        headers: { 'Content-Type': 'multipart/form-data' }
      })
      questionBankUploadStatus.value = {
        success: true,
        message: res.data.message || 'Soal berhasil diimpor ke bank soal!'
      }
      showToast('Soal berhasil diimpor ke bank soal!', 'success')
      questionBankFile.value = null
      showUploadBankModal.value = false
      await loadAllData()
    } catch (err) {
      questionBankUploadStatus.value = {
        success: false,
        message: err.response?.data?.message || 'Gagal mengimpor file Excel bank soal.'
      }
      await showAlertModal({
        title: 'Gagal Impor Soal',
        message: err.response?.data?.message || 'Gagal mengimpor file Excel bank soal.',
        type: 'danger'
      })
    } finally {
      isUploadingQuestionBank.value = false
    }
  }

  const toggleBankLock = async (b) => {
    try {
      const res = await api.post(`/admin/question-banks/${b.id}/toggle-lock`)
      b.is_locked = res.data.is_locked
      showToast(b.is_locked ? 'Bank Soal dikunci (Terkunci)' : 'Bank Soal dibuka (Draft)', 'info')
    } catch (e) {
      await showAlertModal({
        title: 'Gagal Mengubah Status Bank Soal',
        message: 'Terjadi kesalahan saat mengubah status penguncian bank soal.',
        type: 'danger'
      })
    }
  }

  return {
    canLockQuestionBanks,
    canPrintQuestionBanks,
    showQuestionPrint,
    questionPrintBank,
    openQuestionPrint,
    selectedBankUploadId,
    isUploadingQuestionBank,
    questionBankUploadStatus,
    showUploadBankModal,
    showCreateBankModal,
    isCreatingBank,
    isEditBank,
    questionBankSearchQuery,
    questionBankStatusFilter,
    newBankForm,
    bankScopeOptions,
    newBankOptionKey,
    bankScopeLabel,
    availableBankSubjects,
    bankCustomSubjects,
    bankCustomClassOptions,
    bankFormLocked,
    onBankSubjectChange,
    onBankCustomSubjectChange,
    toggleBankClass,
    setBankScope,
    questionBankSortKey,
    questionBankSortOrder,
    questionBankCurrentPage,
    questionBankPerPage,
    toggleQuestionBankSort,
    filteredQuestionBanks,
    totalQuestionBankPages,
    paginatedQuestionBanks,
    displayedQuestionBankPages,
    setQuestionBankPage,
    downloadQuestionBankTemplate,
    handleQuestionBankFileChange,
    openUploadBankModal,
    openCreateBankModal,
    openEditBankModal,
    submitCreateBank,
    deleteQuestionBank,
    toggleBankLockWithConfirm,
    submitQuestionBankUpload,
  }
}
