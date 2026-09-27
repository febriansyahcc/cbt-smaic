import { ref } from 'vue'
import api from '../../../../services/api'

// Pengelola butir soal: tambah/ubah/hapus soal, gambar, rumus, dan urutan seret-lepas.
export function useQuestionEditor(ctx) {
  const {
    showAlertModal,
    showConfirmModal,
    showToast,
  } = ctx

  // Didefinisikan composable yang dipanggil kemudian; diteruskan saat dipanggil.
  const loadAllData = (...args) => ctx.loadAllData(...args)

  // ---------------- QUESTION VIEWER & EDITOR ----------------
  const showBankQuestionsModal = ref(false)
  const showSimulatorModal = ref(false)
  const viewingBank = ref(null)
  const bankQuestions = ref([])
  const isLoadingBankQuestions = ref(false)
  const editingQuestionId = ref(null)
  const isSavingQuestion = ref(false)
  const isUploadingQuestionImage = ref(false)
  const uploadingOptionKey = ref(null)
  const draggedQuestionIdx = ref(null)
  const dragOverQuestionIdx = ref(null)
  const isReorderingQuestions = ref(false)
  const reorderStatusMessage = ref('')

  const previewBankDirectly = async (bank) => {
    viewingBank.value = bank
    await fetchBankQuestions(bank.id)
    showSimulatorModal.value = true
  }

  const handleQuestionImageUpload = async (event) => {
    const file = event.target?.files?.[0]
    if (!file) return
    await uploadAndInsertQuestionImage(file)
    if (event.target) event.target.value = ''
  }

  const uploadAndInsertQuestionImage = async (file) => {
    if (!file.type.startsWith('image/')) {
      showToast('File harus berupa gambar (JPG, PNG, WEBP, GIF, SVG)', 'warning')
      return
    }
    if (file.size > 5 * 1024 * 1024) {
      showToast('Ukuran gambar maksimal 5 MB', 'warning')
      return
    }

    isUploadingQuestionImage.value = true
    const formData = new FormData()
    formData.append('image', file)

    try {
      const res = await api.post('/admin/upload-image', formData, {
        headers: { 'Content-Type': 'multipart/form-data' }
      })
      const imgUrl = res.data.url
      const imgTag = `\n<img src="${imgUrl}" alt="Gambar Soal" />\n`
      questionForm.value.content_html = (questionForm.value.content_html || '') + imgTag
      showToast('Gambar pertanyaan berhasil diunggah!', 'success')
    } catch (err) {
      showToast('Gagal mengunggah gambar: ' + (err.response?.data?.message || err.message), 'danger')
    } finally {
      isUploadingQuestionImage.value = false
    }
  }

  const handleQuestionPaste = async (event) => {
    const items = (event.clipboardData || event.originalEvent?.clipboardData)?.items
    if (!items) return

    for (const item of items) {
      if (item.type.indexOf('image') !== -1) {
        event.preventDefault()
        const file = item.getAsFile()
        if (file) {
          await uploadAndInsertQuestionImage(file)
        }
        break
      }
    }
  }

  const handleQuestionDrop = async (event) => {
    const files = event.dataTransfer?.files
    if (files && files.length > 0) {
      const file = files[0]
      if (file.type.startsWith('image/')) {
        await uploadAndInsertQuestionImage(file)
      }
    }
  }

  const handleOptionImageUpload = async (event, optKey) => {
    const file = event.target?.files?.[0]
    if (!file) return
    if (!file.type.startsWith('image/')) {
      showToast('File harus berupa gambar (JPG, PNG, WEBP, GIF, SVG)', 'warning')
      return
    }
    if (file.size > 5 * 1024 * 1024) {
      showToast('Ukuran gambar maksimal 5 MB', 'warning')
      return
    }

    uploadingOptionKey.value = optKey
    const formData = new FormData()
    formData.append('image', file)

    try {
      const res = await api.post('/admin/upload-image', formData, {
        headers: { 'Content-Type': 'multipart/form-data' }
      })
      const opt = questionForm.value.options.find(o => o.key === optKey)
      if (opt) {
        opt.image_url = res.data.url
      }
      showToast(`Gambar pilihan ${optKey} berhasil diunggah!`, 'success')
    } catch (err) {
      showToast('Gagal mengunggah gambar opsi: ' + (err.response?.data?.message || err.message), 'danger')
    } finally {
      uploadingOptionKey.value = null
      if (event.target) event.target.value = ''
    }
  }

  const removeOptionImage = (optKey) => {
    const opt = questionForm.value.options.find(o => o.key === optKey)
    if (opt) {
      opt.image_url = ''
    }
  }

  const insertMathSnippet = (snippet) => {
    questionForm.value.content_html = (questionForm.value.content_html || '') + ' ' + snippet + ' '
  }

  const onQuestionDragStart = (event, index) => {
    draggedQuestionIdx.value = index
    if (event.dataTransfer) {
      event.dataTransfer.effectAllowed = 'move'
      event.dataTransfer.setData('text/plain', index.toString())
    }
  }

  const onQuestionDragOver = (event, index) => {
    event.preventDefault()
    if (event.dataTransfer) {
      event.dataTransfer.dropEffect = 'move'
    }
    if (dragOverQuestionIdx.value !== index) {
      dragOverQuestionIdx.value = index
    }
  }

  const onQuestionDragLeave = (event, index) => {
    if (dragOverQuestionIdx.value === index) {
      dragOverQuestionIdx.value = null
    }
  }

  const onQuestionDrop = async (event, targetIndex) => {
    event.preventDefault()
    const fromIndex = draggedQuestionIdx.value
    dragOverQuestionIdx.value = null
    draggedQuestionIdx.value = null

    if (fromIndex === null || fromIndex === targetIndex || fromIndex === undefined || targetIndex === undefined) {
      return
    }

    // Local reorder
    const movedItem = bankQuestions.value.splice(fromIndex, 1)[0]
    bankQuestions.value.splice(targetIndex, 0, movedItem)

    // Update question_number locally
    bankQuestions.value.forEach((q, idx) => {
      q.question_number = idx + 1
    })

    // Sync to backend
    const bankId = viewingBank.value?.id
    if (!bankId) return

    isReorderingQuestions.value = true
    reorderStatusMessage.value = 'Menyimpan urutan...'
    try {
      const questionIds = bankQuestions.value.map(q => q.id)
      await api.put(`/admin/question-banks/${bankId}/reorder`, {
        question_ids: questionIds
      })
      reorderStatusMessage.value = '✓ Urutan tersimpan'
      showToast('Urutan nomor soal berhasil diperbarui!', 'success')
      setTimeout(() => {
        if (reorderStatusMessage.value === '✓ Urutan tersimpan') {
          reorderStatusMessage.value = ''
        }
      }, 2500)
    } catch (err) {
      reorderStatusMessage.value = 'Gagal menyimpan urutan'
      showToast('Gagal menyimpan urutan soal: ' + (err.response?.data?.message || err.message), 'danger')
    } finally {
      isReorderingQuestions.value = false
    }
  }

  const onQuestionDragEnd = () => {
    draggedQuestionIdx.value = null
    dragOverQuestionIdx.value = null
  }

  const questionForm = ref({
    question_number: 1,
    type: 'MULTIPLE_CHOICE',
    content_html: '',
    options: [
      { key: 'A', text: '', image_url: '' },
      { key: 'B', text: '', image_url: '' },
      { key: 'C', text: '', image_url: '' },
      { key: 'D', text: '', image_url: '' },
      { key: 'E', text: '', image_url: '' }
    ],
    correct_key: 'A',
    rubric_guide: '',
    score_weight: 1.0
  })

  const getQuestionTypeLabel = (type) => {
    switch (type) {
      case 'SHORT_ANSWER':
        return 'Jawaban Singkat'
      case 'ESSAY':
        return 'Essay / Uraian'
      default:
        return 'Pilihan Ganda'
    }
  }

  const getQuestionTypeBadgeClass = (type) => {
    switch (type) {
      case 'SHORT_ANSWER':
        return 'bg-emerald-50 border-emerald-200 text-emerald-700'
      case 'ESSAY':
        return 'bg-purple-50 border-purple-200 text-purple-700'
      default:
        return 'bg-blue-50 border-blue-200 text-blue-700'
    }
  }

  const openBankQuestionsModal = async (bank) => {
    viewingBank.value = bank
    editingQuestionId.value = null
    showBankQuestionsModal.value = true
    await fetchBankQuestions(bank.id)
  }

  const fetchBankQuestions = async (bankId = null) => {
    const id = bankId || viewingBank.value?.id
    if (!id) return
    isLoadingBankQuestions.value = true
    try {
      const res = await api.get(`/admin/question-banks/${id}/questions`)
      bankQuestions.value = res.data.data || []
      if (res.data.bank) {
        viewingBank.value = res.data.bank
      }
    } catch (err) {
      showToast('Gagal memuat daftar butir soal: ' + (err.response?.data?.message || err.message), 'danger')
    } finally {
      isLoadingBankQuestions.value = false
    }
  }

  const startAddQuestion = () => {
    editingQuestionId.value = 'new'
    questionForm.value = {
      question_number: (bankQuestions.value.length || 0) + 1,
      type: 'MULTIPLE_CHOICE',
      content_html: '',
      options: [
        { key: 'A', text: '', image_url: '' },
        { key: 'B', text: '', image_url: '' },
        { key: 'C', text: '', image_url: '' },
        { key: 'D', text: '', image_url: '' },
        { key: 'E', text: '', image_url: '' }
      ],
      correct_key: 'A',
      rubric_guide: '',
      score_weight: 1.0
    }
  }

  const startEditQuestion = (q) => {
    editingQuestionId.value = q.id
    let parsedOpts = q.options || []
    if (typeof parsedOpts === 'string') {
      try {
        parsedOpts = JSON.parse(parsedOpts)
      } catch {
        parsedOpts = []
      }
    }
    const opts = ['A', 'B', 'C', 'D', 'E'].map(k => {
      const existing = parsedOpts.find(o => o.key === k)
      return {
        key: k,
        text: existing ? existing.text : '',
        image_url: existing ? (existing.image_url || '') : ''
      }
    })
    const qType = q.type || 'MULTIPLE_CHOICE'
    questionForm.value = {
      question_number: q.question_number,
      type: qType,
      content_html: q.content_html,
      options: opts,
      correct_key: q.correct_key || (qType === 'SHORT_ANSWER' ? '' : 'A'),
      rubric_guide: q.rubric_guide || '',
      score_weight: q.score_weight || 1.0
    }
  }

  const cancelEditQuestion = () => {
    editingQuestionId.value = null
  }

  const saveQuestion = async () => {
    if (!questionForm.value.content_html.trim()) {
      await showAlertModal({
        title: 'Form Belum Lengkap',
        message: 'Konten / naskah soal wajib diisi.',
        type: 'warning'
      })
      return
    }

    const qType = questionForm.value.type || 'MULTIPLE_CHOICE'

    if (qType === 'MULTIPLE_CHOICE') {
      const filledOptions = questionForm.value.options.filter(o => (o.text && o.text.trim()) || (o.image_url && o.image_url.trim()))
      if (filledOptions.length < 2) {
        await showAlertModal({
          title: 'Pilihan Jawaban Kurang',
          message: 'Harap isi minimal 2 pilihan jawaban (misal A dan B teks atau gambar).',
          type: 'warning'
        })
        return
      }
    } else if (qType === 'SHORT_ANSWER') {
      if (!questionForm.value.correct_key.trim()) {
        await showAlertModal({
          title: 'Kunci Jawaban Kosong',
          message: 'Kunci jawaban singkat wajib diisi.',
          type: 'warning'
        })
        return
      }
    }

    isSavingQuestion.value = true
    try {
      const payload = {
        type: qType,
        content_html: questionForm.value.content_html.trim(),
        options: qType === 'MULTIPLE_CHOICE'
          ? questionForm.value.options.filter(o => (o.text && o.text.trim()) || (o.image_url && o.image_url.trim()))
          : [],
        correct_key: questionForm.value.correct_key.trim(),
        rubric_guide: questionForm.value.rubric_guide.trim(),
        score_weight: Number(questionForm.value.score_weight) || 1.0
      }

      if (editingQuestionId.value === 'new') {
        await api.post(`/admin/question-banks/${viewingBank.value.id}/questions`, payload)
        showToast('Butir soal baru berhasil ditambahkan!', 'success')
      } else {
        await api.put(`/admin/questions/${editingQuestionId.value}`, payload)
        showToast('Butir soal berhasil diperbarui!', 'success')
      }

      editingQuestionId.value = null
      await fetchBankQuestions()
      await loadAllData()
    } catch (err) {
      await showAlertModal({
        title: 'Gagal Menyimpan Soal',
        message: err.response?.data?.message || 'Terjadi kesalahan saat menyimpan soal.',
        type: 'danger'
      })
    } finally {
      isSavingQuestion.value = false
    }
  }

  const deleteQuestionItem = async (q) => {
    const confirmed = await showConfirmModal({
      title: 'Hapus Butir Soal',
      message: `Apakah Anda yakin ingin menghapus Soal #${q.question_number}? Nomor urut soal lainnya akan disesuaikan otomatis.`,
      type: 'danger',
      confirmText: 'Hapus Soal',
      cancelText: 'Batal'
    })
    if (!confirmed) return

    try {
      await api.delete(`/admin/questions/${q.id}`)
      showToast(`Soal #${q.question_number} berhasil dihapus!`, 'success')
      await fetchBankQuestions()
      await loadAllData()
    } catch (err) {
      await showAlertModal({
        title: 'Gagal Menghapus Soal',
        message: err.response?.data?.message || 'Terjadi kesalahan saat menghapus soal.',
        type: 'danger'
      })
    }
  }

  return {
    showBankQuestionsModal,
    showSimulatorModal,
    viewingBank,
    bankQuestions,
    isLoadingBankQuestions,
    editingQuestionId,
    isSavingQuestion,
    isUploadingQuestionImage,
    uploadingOptionKey,
    draggedQuestionIdx,
    dragOverQuestionIdx,
    isReorderingQuestions,
    reorderStatusMessage,
    previewBankDirectly,
    handleQuestionImageUpload,
    handleQuestionPaste,
    handleQuestionDrop,
    handleOptionImageUpload,
    removeOptionImage,
    insertMathSnippet,
    onQuestionDragStart,
    onQuestionDragOver,
    onQuestionDragLeave,
    onQuestionDrop,
    onQuestionDragEnd,
    questionForm,
    getQuestionTypeLabel,
    getQuestionTypeBadgeClass,
    openBankQuestionsModal,
    startAddQuestion,
    startEditQuestion,
    cancelEditQuestion,
    saveQuestion,
    deleteQuestionItem,
  }
}
