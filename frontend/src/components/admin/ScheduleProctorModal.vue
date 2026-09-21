<template>
  <Teleport to="body">
    <div v-if="rendered" class="fixed inset-0 z-[60] flex items-center justify-center p-4">
      <!-- Backdrop: hanya fade opacity -->
      <Transition
        appear
        enter-active-class="transition-opacity duration-200 ease-out"
        enter-from-class="opacity-0"
        enter-to-class="opacity-100"
        leave-active-class="transition-opacity duration-150 ease-in"
        leave-from-class="opacity-100"
        leave-to-class="opacity-0"
      >
        <div v-if="modelValue" class="fixed inset-0 bg-slate-900/60 backdrop-blur-sm" @click="close"></div>
      </Transition>

      <!-- Kartu modal: scale + translate halus -->
      <Transition
        appear
        enter-active-class="transition duration-200 ease-out transform"
        enter-from-class="opacity-0 scale-95 translate-y-2"
        enter-to-class="opacity-100 scale-100 translate-y-0"
        leave-active-class="transition duration-150 ease-in transform"
        leave-from-class="opacity-100 scale-100 translate-y-0"
        leave-to-class="opacity-0 scale-95 translate-y-2"
        @after-leave="rendered = false"
      >
        <div
          v-if="modelValue"
          class="relative z-10 bg-white rounded-3xl max-w-lg w-full shadow-2xl border border-slate-100 flex flex-col max-h-[90vh] overflow-hidden"
          role="dialog"
          aria-modal="true"
          @click.stop
        >
          <!-- Header -->
          <div class="flex items-start justify-between gap-3 px-6 py-4 border-b border-slate-100 shrink-0">
            <div class="min-w-0">
              <h3 class="text-base font-bold text-slate-900">{{ local ? 'Pilih Pengawas' : 'Atur Pengawas' }}</h3>
              <p v-if="schedule?.title" class="text-[11px] text-slate-500 mt-0.5 truncate">
                {{ schedule.title }}<template v-if="schedule.class_room?.name"> &middot; {{ schedule.class_room.name }}</template>
              </p>
            </div>
            <button
              type="button"
              @click="close"
              class="p-1.5 rounded-xl text-slate-400 hover:text-slate-600 hover:bg-slate-100 transition active:scale-95 cursor-pointer shrink-0"
              title="Tutup"
            >
              <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
              </svg>
            </button>
          </div>

          <!-- Body -->
          <div class="px-6 py-4 space-y-3 text-xs overflow-y-auto flex-1 min-h-0">
            <!-- Memuat -->
            <div v-if="isLoading" class="py-10 flex flex-col items-center gap-2 text-slate-400">
              <svg class="w-5 h-5 animate-spin" fill="none" viewBox="0 0 24 24">
                <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
                <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8v4a4 4 0 00-4 4H4z"></path>
              </svg>
              <span>Memuat daftar staf...</span>
            </div>

            <!-- Galat memuat -->
            <div v-else-if="loadError" class="py-8 text-center space-y-2">
              <p class="font-bold text-slate-700">Daftar staf tidak dapat dimuat</p>
              <p class="text-[11px] text-slate-500">{{ loadError }}</p>
              <button
                type="button"
                @click="loadData"
                class="px-3 h-8 inline-flex items-center justify-center rounded-lg border text-xs font-semibold whitespace-nowrap transition active:scale-95 cursor-pointer bg-indigo-50 hover:bg-indigo-100 text-indigo-700 border-indigo-200"
              >
                Coba Lagi
              </button>
            </div>

            <template v-else>
              <!-- Pencarian -->
              <div class="relative">
                <svg class="w-4 h-4 text-slate-400 absolute left-3 top-1/2 -translate-y-1/2 pointer-events-none" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
                </svg>
                <input
                  v-model="searchQuery"
                  type="text"
                  placeholder="Cari nama atau username staf..."
                  class="w-full h-9 pl-9 pr-9 bg-white border border-slate-200 rounded-xl text-xs font-medium text-slate-800 placeholder:text-slate-400 focus:border-indigo-500 focus:ring-4 focus:ring-indigo-100 focus:outline-none transition"
                />
                <button
                  v-if="searchQuery"
                  type="button"
                  @click="searchQuery = ''"
                  class="absolute right-2 top-1/2 -translate-y-1/2 p-1 rounded-lg text-slate-400 hover:text-slate-600 hover:bg-slate-100 cursor-pointer transition"
                  title="Hapus pencarian"
                >
                  <svg class="w-3.5 h-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
                  </svg>
                </button>
              </div>

              <!-- Chip pengawas terpilih -->
              <div>
                <div class="font-bold text-slate-700 mb-1.5">Pengawas terpilih ({{ selectedIds.length }})</div>
                <div v-if="selectedUsers.length > 0" class="flex flex-wrap gap-1.5">
                  <span
                    v-for="u in selectedUsers"
                    :key="u.id"
                    class="inline-flex items-center gap-1 pl-2.5 pr-1 py-1 bg-indigo-50 text-indigo-700 border border-indigo-100 rounded-lg text-[11px] font-semibold"
                  >
                    <span class="truncate max-w-[180px]">{{ u.full_name || u.username }}</span>
                    <button
                      type="button"
                      @click="removeSelected(u.id)"
                      class="p-0.5 rounded-md text-indigo-400 hover:text-indigo-700 hover:bg-indigo-100 transition cursor-pointer"
                      :title="`Hapus ${u.full_name || u.username}`"
                    >
                      <svg class="w-3 h-3" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
                      </svg>
                    </button>
                  </span>
                </div>
                <p v-else class="text-[11px] text-slate-400">Belum ada pengawas dipilih.</p>
              </div>

              <!-- Daftar staf -->
              <div class="border border-slate-200 rounded-2xl overflow-hidden">
                <div v-if="candidates.length === 0" class="py-8 text-center text-slate-400">
                  <div class="font-bold text-slate-700">Belum ada staf yang dapat ditugaskan</div>
                  <div class="text-[11px] mt-0.5">Tambahkan akun guru atau staf aktif terlebih dahulu.</div>
                </div>
                <div v-else-if="filteredCandidates.length === 0" class="py-8 text-center text-slate-400">
                  <div class="font-bold text-slate-700">Tidak ada staf yang cocok</div>
                  <div class="text-[11px] mt-0.5">Coba kata kunci lain.</div>
                </div>
                <ul v-else class="max-h-64 overflow-y-auto divide-y divide-slate-100">
                  <li v-for="u in filteredCandidates" :key="u.id">
                    <label
                      class="flex items-center gap-3 px-3.5 py-2.5 cursor-pointer hover:bg-slate-50 transition"
                      :class="isSelected(u.id) ? 'bg-indigo-50/40' : ''"
                    >
                      <input
                        type="checkbox"
                        :checked="isSelected(u.id)"
                        @change="toggleSelected(u.id)"
                        class="w-4 h-4 rounded text-indigo-600 border-slate-300 focus:ring-indigo-500 cursor-pointer shrink-0"
                      />
                      <span class="min-w-0 flex-1">
                        <span class="block font-semibold text-slate-800 truncate">{{ u.full_name }}</span>
                        <span class="block text-[11px] text-slate-400 font-mono truncate">{{ u.username }}<template v-if="roleLabel(u.role)"> &middot; {{ roleLabel(u.role) }}</template></span>
                      </span>
                    </label>
                  </li>
                </ul>
              </div>
            </template>
          </div>

          <!-- Footer -->
          <div class="px-6 py-3 bg-slate-50 border-t border-slate-100 flex items-center justify-end gap-2 shrink-0">
            <button
              type="button"
              @click="close"
              :disabled="isSaving"
              class="px-4 py-2 bg-white hover:bg-slate-100 border border-slate-200 rounded-xl text-slate-600 font-semibold text-xs transition active:scale-95 cursor-pointer disabled:opacity-50 disabled:cursor-not-allowed"
            >
              Batal
            </button>
            <button
              type="button"
              @click="submit"
              :disabled="isSaving || isLoading || !!loadError"
              class="px-5 py-2 bg-indigo-600 hover:bg-indigo-700 text-white rounded-xl font-bold text-xs shadow-sm transition active:scale-95 cursor-pointer inline-flex items-center gap-1.5 disabled:opacity-50 disabled:cursor-not-allowed disabled:active:scale-100"
            >
              <svg v-if="isSaving" class="w-3.5 h-3.5 animate-spin" fill="none" viewBox="0 0 24 24">
                <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
                <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8v4a4 4 0 00-4 4H4z"></path>
              </svg>
              <span>{{ isSaving ? 'Menyimpan...' : (local ? 'Terapkan' : 'Simpan') }}</span>
            </button>
          </div>
        </div>
      </Transition>
    </div>
  </Teleport>
</template>

<script setup>
import { ref, computed, watch, onBeforeUnmount } from 'vue'
import api from '@/services/api'
import { useDialog } from '@/composables/useDialog'

// Modal "Atur Pengawas" untuk satu jadwal ujian.
// Mode normal: memuat pengawas saat ini dan menyimpan lewat PUT /admin/schedules/:id/proctors.
// Mode lokal (local): dipakai form Buat/Edit Jadwal; tidak menyimpan ke server, hanya
// mengembalikan pilihan lewat event "apply" agar disimpan setelah jadwal berhasil tersimpan.
const props = defineProps({
  modelValue: { type: Boolean, default: false },
  schedule: { type: Object, default: null },
  local: { type: Boolean, default: false },
  initialSelected: { type: Array, default: () => [] },
})
const emit = defineEmits(['update:modelValue', 'saved', 'apply'])

const { alert: showAlertModal, toast: showToast } = useDialog()

const rendered = ref(false)
const isLoading = ref(false)
const isSaving = ref(false)
const loadError = ref('')
const searchQuery = ref('')
const candidates = ref([])
const selectedIds = ref([])
// Pengguna yang dikenal dari daftar pengawas saat ini (bisa saja tidak lagi ada di kandidat).
const knownUsers = ref({})

let loadToken = 0

const userMap = computed(() => {
  const map = { ...knownUsers.value }
  for (const c of candidates.value) map[c.id] = c
  return map
})

const selectedUsers = computed(() =>
  selectedIds.value.map((id) => userMap.value[id] || { id, full_name: 'Staf tidak dikenal', username: '' })
)

const filteredCandidates = computed(() => {
  const q = searchQuery.value.trim().toLowerCase()
  if (!q) return candidates.value
  return candidates.value.filter(
    (u) => (u.full_name || '').toLowerCase().includes(q) || (u.username || '').toLowerCase().includes(q)
  )
})

const isSelected = (id) => selectedIds.value.includes(id)

const toggleSelected = (id) => {
  if (isSelected(id)) selectedIds.value = selectedIds.value.filter((x) => x !== id)
  else selectedIds.value = [...selectedIds.value, id]
}

const removeSelected = (id) => {
  selectedIds.value = selectedIds.value.filter((x) => x !== id)
}

const roleLabel = (role) => {
  switch (role) {
    case 'ADMIN':
      return 'Admin'
    case 'GURU':
      return 'Guru'
    default:
      return role || ''
  }
}

const errorMessage = (err, fallback) => err?.response?.data?.message || fallback

const loadData = async () => {
  const token = ++loadToken
  isLoading.value = true
  loadError.value = ''
  searchQuery.value = ''
  candidates.value = []
  selectedIds.value = []
  knownUsers.value = {}

  try {
    // Pengawas saat ini: dari objek jadwal / pilihan form, dengan cadangan GET bila kosong.
    let current = props.local ? props.initialSelected : props.schedule?.proctors
    const needFallback = !props.local && props.schedule?.id && (!Array.isArray(current) || current.length === 0)

    const [candRes, fallbackRes] = await Promise.all([
      api.get('/admin/proctor-candidates'),
      needFallback ? api.get(`/admin/schedules/${props.schedule.id}/proctors`) : Promise.resolve(null),
    ])
    if (token !== loadToken) return

    if (fallbackRes) current = fallbackRes.data?.data || []
    current = Array.isArray(current) ? current : []

    candidates.value = candRes.data?.data || []
    const known = {}
    for (const p of current) known[p.id] = p
    knownUsers.value = known
    selectedIds.value = current.map((p) => p.id)
  } catch (err) {
    if (token !== loadToken) return
    loadError.value = errorMessage(err, 'Terjadi kesalahan saat memuat data. Periksa koneksi lalu coba lagi.')
  } finally {
    if (token === loadToken) isLoading.value = false
  }
}

const close = () => {
  if (isSaving.value) return
  emit('update:modelValue', false)
}

const submit = async () => {
  if (isSaving.value || isLoading.value || loadError.value) return

  if (props.local) {
    emit('apply', selectedUsers.value.map((u) => ({
      id: u.id,
      full_name: u.full_name,
      username: u.username,
      role: u.role,
    })))
    emit('update:modelValue', false)
    return
  }

  if (!props.schedule?.id) return
  isSaving.value = true
  try {
    const res = await api.put(`/admin/schedules/${props.schedule.id}/proctors`, {
      user_ids: [...selectedIds.value],
    })
    showToast(res.data?.message || 'Pengawas jadwal berhasil disimpan.', 'success')
    emit('saved')
    emit('update:modelValue', false)
  } catch (err) {
    await showAlertModal({
      title: 'Gagal Menyimpan Pengawas',
      message: errorMessage(err, 'Terjadi kesalahan saat menyimpan pengawas jadwal.'),
      type: 'danger',
    })
  } finally {
    isSaving.value = false
  }
}

const onKeydown = (e) => {
  if (e.key === 'Escape') close()
}

watch(
  () => props.modelValue,
  (open) => {
    if (open) {
      rendered.value = true
      window.addEventListener('keydown', onKeydown)
      loadData()
    } else {
      loadToken++
      window.removeEventListener('keydown', onKeydown)
    }
  },
  { immediate: true }
)

onBeforeUnmount(() => window.removeEventListener('keydown', onKeydown))
</script>
