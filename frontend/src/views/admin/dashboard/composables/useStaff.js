import { ref, computed, watch } from 'vue'
import api from '../../../../services/api'
import { templateKeyFor, templateLabelFor, templateByKey, applyImplications, effectivePermissions } from '../../../../utils/staffAccess'

// Data guru & staf: filter, katalog izin, editor akses, form, dan impor Excel.
export function useStaff(ctx) {
  const {
    authStore,
    showAlertModal,
    showConfirmModal,
    showToast,
    teachers,
  } = ctx

  // Didefinisikan composable yang dipanggil kemudian; diteruskan saat dipanggil.
  const loadAllData = (...args) => ctx.loadAllData(...args)

  const showTeacherModal = ref(false)

  // Import states — Guru
  const showImportTeacherModal = ref(false)
  const importTeacherFile = ref(null)
  const isImportingTeacher = ref(false)
  const importTeacherResult = ref(null)
  const newTeacher = ref({ full_name: '', username: '', password: '' })
  const newTeacherAccess = ref({ role: 'GURU', permissions: [] })
  const newTeacherAccessValid = ref(true)

  // Hanya grantor (ADMIN atau pemegang izin "*") yang boleh mengatur izin, username, dan password akun staf.
  // Server tetap menjadi penjaga sebenarnya (403); ini hanya untuk kerapian UI.
  const isGrantor = computed(() => authStore.role === 'ADMIN' || authStore.permissions.includes('*'))
  const newTeacherCanSave = computed(() => !isGrantor.value || newTeacherAccessValid.value)

  // 2. Guru dan Staf Filtering, Sorting & Pagination
  const teacherSearchQuery = ref('')
  const selectedTeacherRole = ref('all')
  const teacherSortKey = ref('name')
  const teacherSortOrder = ref('asc')
  const teacherCurrentPage = ref(1)
  const teacherPerPage = ref(10)
  const showEditTeacherModal = ref(false)
  const editTeacherForm = ref({ id: null, full_name: '', username: '', password: '' })
  const editTeacherAccess = ref({ role: 'GURU', permissions: [] })
  const editTeacherAccessValid = ref(true)
  const editTeacherCanSave = computed(() => !isGrantor.value || editTeacherAccessValid.value)

  // Katalog izin (GET /admin/permission-catalog): dimuat sekali, dipakai editor akses dan label peran turunan.
  const permissionCatalog = ref(null)
  let permissionCatalogRequest = null
  const loadPermissionCatalog = () => {
    if (permissionCatalog.value) return Promise.resolve(permissionCatalog.value)
    if (!permissionCatalogRequest) {
      permissionCatalogRequest = api.get('/admin/permission-catalog')
        .then((res) => {
          const d = res.data?.data || {}
          permissionCatalog.value = {
            groups: d.groups || [],
            templates: d.templates || [],
            implies: d.implies || {},
          }
          return permissionCatalog.value
        })
        .finally(() => { permissionCatalogRequest = null })
    }
    return permissionCatalogRequest
  }

  const ensurePermissionCatalog = async () => {
    try {
      return await loadPermissionCatalog()
    } catch (e) {
      await showAlertModal({
        title: 'Gagal Memuat Daftar Izin',
        message: e.response?.data?.message || 'Daftar izin akses tidak dapat dimuat. Coba lagi beberapa saat.',
        type: 'danger'
      })
      return null
    }
  }

  // Peran turunan akun staf: 'admin' | 'guru' | 'pengawas' | 'custom'
  const teacherKey = (t) => templateKeyFor(t, permissionCatalog.value?.templates, permissionCatalog.value?.implies)
  const teacherRoleLabel = (t) => templateLabelFor(t, permissionCatalog.value?.templates, permissionCatalog.value?.implies)
  const TEACHER_BADGE_CLASSES = {
    admin: 'bg-purple-100 text-purple-800',
    guru: 'bg-indigo-100 text-indigo-800',
    pengawas: 'bg-emerald-100 text-emerald-800',
    custom: 'bg-amber-100 text-amber-800',
    unknown: 'bg-slate-100 text-slate-500',
  }
  const teacherBadgeClass = (t) => TEACHER_BADGE_CLASSES[teacherKey(t)]

  const TEACHER_ROLE_CHIPS = [
    { key: 'admin', label: 'Administrator' },
    { key: 'guru', label: 'Guru' },
    { key: 'pengawas', label: 'Pengawas' },
    { key: 'custom', label: 'Kustom' },
  ]
  const teacherRoleChips = computed(() => {
    const list = teachers.value || []
    const counts = { admin: 0, guru: 0, pengawas: 0, custom: 0 }
    for (const t of list) {
      const key = teacherKey(t)
      if (key in counts) counts[key]++ // 'unknown' (katalog belum termuat) tidak dihitung ke kategori mana pun
    }
    const chips = [{ key: 'all', label: `Semua (${list.length})` }]
    for (const c of TEACHER_ROLE_CHIPS) {
      if (counts[c.key] > 0 || selectedTeacherRole.value === c.key) {
        chips.push({ key: c.key, label: `${c.label} (${counts[c.key]})` })
      }
    }
    return chips
  })

  // Nilai awal editor: template Guru (izin terisi dari template).
  const defaultStaffAccess = (catalog) => {
    const guru = templateByKey(catalog.templates, 'guru')
    return { role: 'GURU', permissions: applyImplications(guru?.permissions, catalog.implies) }
  }

  // Nilai awal editor dari akun: izin kosong pada akun GURU = izin template Guru.
  // Kunci yang tidak ada di katalog (usang atau tak dikenal, mis. "*") disaring lebih dulu.
  const staffAccessFromAccount = (account, catalog) => {
    if (account.role === 'ADMIN') return { role: 'ADMIN', permissions: ['*'] }
    const known = new Set(catalog.groups.flatMap((g) => (g.items || []).map((i) => i.key)))
    const effective = effectivePermissions(account, catalog.templates).filter((p) => known.has(p))
    return { role: 'GURU', permissions: applyImplications(effective, catalog.implies) }
  }

  // Payload sesuai kontrak. Grantor: ADMIN tanpa permissions, GURU dengan permissions eksplisit.
  // Bukan grantor: tidak boleh mengirim permissions, role dipaksa GURU.
  const buildStaffCreatePayload = (form, access) => {
    const payload = { full_name: form.full_name, username: form.username, role: isGrantor.value ? access.role : 'GURU' }
    if (form.password) payload.password = form.password
    if (isGrantor.value && access.role === 'GURU') payload.permissions = [...access.permissions]
    return payload
  }

  // Bukan grantor: hanya full_name yang boleh diubah (tanpa role, permissions, username, password).
  const buildStaffUpdatePayload = (form, access) => {
    if (!isGrantor.value) return { full_name: form.full_name }
    return buildStaffCreatePayload(form, access)
  }

  const toggleTeacherSort = (key) => {
    if (teacherSortKey.value === key) {
      teacherSortOrder.value = teacherSortOrder.value === 'asc' ? 'desc' : 'asc'
    } else {
      teacherSortKey.value = key
      teacherSortOrder.value = 'asc'
    }
  }

  const filteredTeachers = computed(() => {
    let list = teachers.value || []
    if (selectedTeacherRole.value !== 'all') {
      list = list.filter(t => teacherKey(t) === selectedTeacherRole.value)
    }
    if (teacherSearchQuery.value.trim()) {
      const q = teacherSearchQuery.value.toLowerCase().trim()
      list = list.filter(t => {
        const nameMatch = t.full_name && t.full_name.toLowerCase().includes(q)
        const userMatch = t.username && t.username.toLowerCase().includes(q)
        return nameMatch || userMatch
      })
    }
    if (teacherSortKey.value) {
      list = [...list].sort((a, b) => {
        let result = 0
        if (teacherSortKey.value === 'name') {
          const valA = (a.full_name || '').trim().toLowerCase()
          const valB = (b.full_name || '').trim().toLowerCase()
          result = valA.localeCompare(valB, undefined, { numeric: true, sensitivity: 'base' })
        } else if (teacherSortKey.value === 'username') {
          const valA = (a.username || '').trim().toLowerCase()
          const valB = (b.username || '').trim().toLowerCase()
          result = valA.localeCompare(valB, undefined, { numeric: true, sensitivity: 'base' })
        } else if (teacherSortKey.value === 'role') {
          const valA = teacherRoleLabel(a).toLowerCase()
          const valB = teacherRoleLabel(b).toLowerCase()
          result = valA.localeCompare(valB, undefined, { numeric: true, sensitivity: 'base' })
        }
        return teacherSortOrder.value === 'asc' ? result : -result
      })
    }
    return list
  })

  const totalTeacherPages = computed(() => Math.ceil(filteredTeachers.value.length / teacherPerPage.value) || 1)
  const paginatedTeachers = computed(() => {
    const start = (teacherCurrentPage.value - 1) * teacherPerPage.value
    return filteredTeachers.value.slice(start, start + teacherPerPage.value)
  })

  const displayedTeacherPages = computed(() => {
    const total = totalTeacherPages.value
    const current = teacherCurrentPage.value
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

  const setTeacherPage = (p) => {
    if (p === '...' || p < 1 || p > totalTeacherPages.value) return
    teacherCurrentPage.value = p
  }

  watch([teacherSearchQuery, selectedTeacherRole, teacherPerPage, teacherSortKey, teacherSortOrder], () => {
    teacherCurrentPage.value = 1
  })

  const downloadTeacherTemplate = async () => {
    try {
      const res = await api.get('/admin/teachers/template', { responseType: 'blob' })
      const url = window.URL.createObjectURL(new Blob([res.data], {
        type: 'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet'
      }))
      const a = document.createElement('a')
      a.href = url
      a.download = 'Template_Import_Guru_CBT.xlsx'
      document.body.appendChild(a)
      a.click()
      a.remove()
      window.URL.revokeObjectURL(url)
      showToast('Template guru berhasil diunduh!', 'success')
    } catch (err) {
      await showAlertModal({ title: 'Gagal Mengunduh Template', message: 'Gagal mengunduh template guru.', type: 'danger' })
    }
  }

  const submitImportTeachers = async () => {
    if (!importTeacherFile.value) return
    isImportingTeacher.value = true
    importTeacherResult.value = null
    try {
      const fd = new FormData()
      fd.append('file', importTeacherFile.value)
      const res = await api.post('/admin/teachers/import', fd, {
        headers: { 'Content-Type': 'multipart/form-data' },
      })
      importTeacherResult.value = { success: true, message: res.data.message || 'Impor guru berhasil!' }
      await loadAllData()
    } catch (e) {
      importTeacherResult.value = { success: false, message: e.response?.data?.message || 'Gagal mengunggah file' }
    } finally {
      isImportingTeacher.value = false
    }
  }

  // --- Teacher Master Data Methods ---
  const openCreateTeacherModal = async () => {
    const catalog = await ensurePermissionCatalog()
    if (!catalog) return
    newTeacher.value = { full_name: '', username: '', password: '' }
    newTeacherAccess.value = defaultStaffAccess(catalog)
    newTeacherAccessValid.value = true
    showTeacherModal.value = true
  }

  const openEditTeacher = async (tc) => {
    const catalog = await ensurePermissionCatalog()
    if (!catalog) return
    editTeacherForm.value = {
      id: tc.id,
      full_name: tc.full_name || '',
      username: tc.username || '',
      password: ''
    }
    editTeacherAccess.value = staffAccessFromAccount(tc, catalog)
    editTeacherAccessValid.value = true
    showEditTeacherModal.value = true
  }

  const createTeacher = async () => {
    if (!newTeacherCanSave.value) return
    try {
      await api.post('/admin/teachers', buildStaffCreatePayload(newTeacher.value, newTeacherAccess.value))
      showTeacherModal.value = false
      newTeacher.value = { full_name: '', username: '', password: '' }
      await loadAllData()
      showToast('Akun guru / staf berhasil ditambahkan!', 'success')
    } catch (e) {
      await showAlertModal({
        title: 'Gagal Menambah Guru / Staf',
        message: e.response?.data?.message || 'Terjadi kesalahan saat menambah guru / staf.',
        type: 'danger'
      })
    }
  }

  const updateTeacher = async () => {
    if (!editTeacherCanSave.value) return
    try {
      await api.put(`/admin/teachers/${editTeacherForm.value.id}`, buildStaffUpdatePayload(editTeacherForm.value, editTeacherAccess.value))
      showEditTeacherModal.value = false
      await loadAllData()
      showToast('Data guru / staf berhasil diperbarui!', 'success')
    } catch (e) {
      await showAlertModal({
        title: 'Gagal Memperbarui Guru / Staf',
        message: e.response?.data?.message || 'Terjadi kesalahan saat memperbarui data guru / staf.',
        type: 'danger'
      })
    }
  }

  const deleteTeacher = async (tc) => {
    const confirmed = await showConfirmModal({
      title: 'Hapus Guru / Staf',
      message: `Yakin ingin menghapus akun ${tc.full_name} (${tc.username})?`,
      type: 'danger',
      confirmText: 'Hapus Akun',
      cancelText: 'Batal'
    })
    if (!confirmed) return

    try {
      await api.delete(`/admin/teachers/${tc.id}`)
      await loadAllData()
      showToast('Akun guru / staf berhasil dihapus', 'success')
    } catch (e) {
      await showAlertModal({
        title: 'Gagal Menghapus Guru / Staf',
        message: e.response?.data?.message || 'Terjadi kesalahan saat menghapus akun guru / staf.',
        type: 'danger'
      })
    }
  }

  return {
    showTeacherModal,
    showImportTeacherModal,
    importTeacherFile,
    isImportingTeacher,
    importTeacherResult,
    newTeacher,
    newTeacherAccess,
    newTeacherAccessValid,
    isGrantor,
    newTeacherCanSave,
    teacherSearchQuery,
    selectedTeacherRole,
    teacherSortKey,
    teacherSortOrder,
    teacherCurrentPage,
    teacherPerPage,
    showEditTeacherModal,
    editTeacherForm,
    editTeacherAccess,
    editTeacherAccessValid,
    editTeacherCanSave,
    permissionCatalog,
    loadPermissionCatalog,
    teacherRoleLabel,
    teacherBadgeClass,
    teacherRoleChips,
    toggleTeacherSort,
    filteredTeachers,
    totalTeacherPages,
    paginatedTeachers,
    displayedTeacherPages,
    setTeacherPage,
    downloadTeacherTemplate,
    submitImportTeachers,
    openCreateTeacherModal,
    openEditTeacher,
    createTeacher,
    updateTeacher,
    deleteTeacher,
  }
}
