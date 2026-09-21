<template>
  <div class="min-h-full bg-surface-50 flex flex-col">
    <!-- Student Header -->
    <header class="bg-white border-b border-slate-200 px-4 py-3 shadow-xs">
      <div class="max-w-4xl mx-auto flex items-center justify-between">
        <div class="flex items-center space-x-3">
          <img src="/logo-smic.png" alt="Logo SMAS Islamic Centre Demak" class="w-10 h-10 object-contain drop-shadow-xs shrink-0" />
          <div>
            <h1 class="text-sm font-bold text-slate-900 leading-tight">SMAS Islamic Centre Demak</h1>
            <p class="text-xs text-slate-500">Portal Siswa • Ujian Berbasis Komputer & HP</p>
          </div>
        </div>

        <button
          @click="handleLogout"
          class="px-3 py-1.5 border border-slate-300 hover:bg-slate-50 text-slate-700 text-xs font-semibold rounded-lg transition"
        >
          Keluar
        </button>
      </div>
    </header>

    <!-- Main Content Area -->
    <main class="flex-1 max-w-4xl mx-auto w-full p-4 sm:p-6 space-y-6">
      <!-- Student Profile Card -->
      <div class="bg-white rounded-3xl p-5 sm:p-6 border border-slate-200 shadow-sm flex flex-col sm:flex-row sm:items-center justify-between gap-4">
        <div class="flex items-center space-x-4">
          <div class="w-14 h-14 rounded-2xl bg-indigo-50 border border-indigo-200 text-indigo-700 flex items-center justify-center font-bold text-xl">
            {{ userInitials }}
          </div>
          <div>
            <h2 class="text-base sm:text-lg font-bold text-slate-900">{{ authStore.user?.full_name }}</h2>
            <div class="flex flex-wrap gap-2 mt-1 text-xs text-slate-600">
              <span class="px-2 py-0.5 bg-slate-100 rounded-md font-medium">NIS: {{ authStore.profile?.nis || '-' }}</span>
              <span class="px-2 py-0.5 bg-slate-100 rounded-md font-medium">NISN: {{ authStore.profile?.nisn || '-' }}</span>
              <span class="px-2 py-0.5 bg-indigo-50 text-indigo-700 font-bold rounded-md">Kelas: {{ authStore.profile?.class_room?.name || 'XII' }}</span>
            </div>
          </div>
        </div>

        <div class="text-xs text-slate-400 bg-slate-50 p-3 rounded-2xl border border-slate-100">
          <p class="font-semibold text-slate-700">Aturan Sesi Aktif:</p>
          <p class="text-[11px] mt-0.5">Satu akun hanya dapat aktif pada 1 perangkat HP/Laptop secara bersamaan.</p>
        </div>
      </div>

      <!-- Active Schedules List -->
      <div class="space-y-3">
        <h3 class="text-sm font-bold text-slate-900 flex items-center space-x-1.5">
          <svg class="w-4 h-4 text-indigo-600" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2" />
          </svg>
          <span>Jadwal Ujian Tersedia Hari Ini</span>
        </h3>

        <div v-if="loading" class="text-center py-10">
          <div class="animate-spin w-8 h-8 mx-auto border-3 border-indigo-600 border-t-transparent rounded-full mb-2"></div>
          <p class="text-xs text-slate-500 font-medium">Memuat jadwal ujian...</p>
        </div>

        <div v-else-if="schedules.length === 0" class="bg-white rounded-3xl p-8 border border-slate-200 text-center space-y-2">
          <div class="w-12 h-12 mx-auto rounded-full bg-slate-100 text-slate-400 flex items-center justify-center">
            <svg class="w-6 h-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M20 13V6a2 2 0 00-2-2H6a2 2 0 00-2 2v7m16 0v5a2 2 0 01-2 2H6a2 2 0 01-2-2v-5m16 0h-2.586a1 1 0 00-.707.293l-2.414 2.414a1 1 0 01-.707.293h-3.172a1 1 0 01-.707-.293l-2.414-2.414A1 1 0 006.586 13H4" />
            </svg>
          </div>
          <p class="text-sm font-bold text-slate-800">Tidak Ada Jadwal Ujian Aktif</p>
          <p class="text-xs text-slate-500">Saat ini tidak ada sesi ujian yang dijadwalkan untuk kelas Anda.</p>
        </div>

        <div v-else class="grid grid-cols-1 gap-4">
          <div 
            v-for="item in schedules" 
            :key="item.id"
            class="bg-white rounded-3xl p-5 sm:p-6 border-2 border-indigo-100 shadow-md shadow-indigo-50/50 hover:border-indigo-300 transition space-y-4"
          >
            <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-2">
              <div>
                <div class="flex flex-wrap items-center gap-1.5 mb-1">
                  <span class="px-2.5 py-0.5 bg-indigo-50 text-indigo-700 text-[11px] font-bold rounded-full">
                    {{ item.bank?.subject?.name || 'Mata Pelajaran' }}
                  </span>
                  <span v-if="item.event" class="px-2.5 py-0.5 bg-purple-50 text-purple-700 text-[11px] font-bold rounded-full">
                    {{ item.event.code }} - {{ item.event.title }}
                  </span>
                </div>
                <h4 class="text-base font-bold text-slate-900">{{ item.title }}</h4>
              </div>

              <div class="flex items-center space-x-3 text-xs text-slate-500 font-medium">
                <div class="flex items-center space-x-1">
                  <svg class="w-4 h-4 text-slate-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
                  </svg>
                  <span>{{ item.duration_minutes }} Menit</span>
                </div>
                <div class="flex items-center space-x-1">
                  <svg class="w-4 h-4 text-slate-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 7V3m8 4V3m-9 8h10M5 21h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z" />
                  </svg>
                  <span>{{ formatDate(item.start_time) }}</span>
                </div>
              </div>
            </div>

            <!-- Token Entry & Start -->
            <div class="pt-3 border-t border-slate-100 flex flex-col sm:flex-row items-stretch sm:items-center justify-between gap-3">
              <div class="flex items-center space-x-2 flex-1 max-w-sm">
                <label class="text-xs font-bold text-slate-700 flex-shrink-0">Token Ujian:</label>
                <input
                  v-model="examTokens[item.id]"
                  type="text"
                  placeholder="Masukkan token dari pengawas"
                  class="w-full uppercase font-mono px-3 py-2 rounded-xl border border-slate-300 text-xs tracking-wider focus:ring-2 focus:ring-indigo-600 focus:border-indigo-600"
                />
              </div>

              <button
                @click="handleStartExam(item)"
                :disabled="startingId === item.id"
                class="px-6 py-2.5 bg-indigo-600 hover:bg-indigo-700 active:scale-95 disabled:opacity-50 text-white font-bold text-xs rounded-xl shadow-md transition flex items-center justify-center space-x-1.5"
              >
                <svg v-if="startingId === item.id" class="animate-spin w-4 h-4" fill="none" viewBox="0 0 24 24">
                  <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
                  <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8v8H4z"></path>
                </svg>
                <span>{{ startingId === item.id ? 'Memuat Soal...' : 'Mulai Kerjakan Ujian' }}</span>
              </button>
            </div>
          </div>
        </div>
      </div>
    </main>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '../../stores/auth'
import { useExamStore } from '../../stores/exam'
import api from '../../services/api'
import { useDialog } from '@/composables/useDialog'

const router = useRouter()
const authStore = useAuthStore()
const examStore = useExamStore()
const { alert: showAlertModal } = useDialog()

const schedules = ref([])
const loading = ref(true)
const startingId = ref(null)
const examTokens = ref({})

const userInitials = computed(() => {
  const name = authStore.user?.full_name || 'S'
  return name.split(' ').map(n => n[0]).slice(0, 2).join('').toUpperCase()
})

const formatDate = (dateStr) => {
  if (!dateStr) return '-'
  const d = new Date(dateStr)
  return d.toLocaleDateString('id-ID', { day: 'numeric', month: 'short', hour: '2-digit', minute: '2-digit' })
}

const loadSchedules = async () => {
  loading.value = true
  try {
    const res = await api.get('/student/schedules')
    schedules.value = res.data.data || []
  } catch (err) {
    console.error('Failed to load schedules', err)
  } finally {
    loading.value = false
  }
}

const handleStartExam = async (schedule) => {
  const token = examTokens.value[schedule.id]
  if (!token) {
    await showAlertModal({
      title: 'Token Belum Diisi',
      message: 'Silakan masukkan token ruang ujian.',
      type: 'warning'
    })
    return
  }

  startingId.value = schedule.id
  try {
    // Save active schedule id & token
    localStorage.setItem('cbt_active_schedule_id', schedule.id)
    localStorage.setItem('cbt_active_token', token)

    // Start exam: creates session, shuffles with seeded PRNG, returns stripped questions
    await examStore.startExam(schedule.id, token)

    // Request fullscreen on user gesture
    try {
      if (document.documentElement.requestFullscreen) {
        await document.documentElement.requestFullscreen()
      }
    } catch (e) {
      // Fullscreen optional if browser blocks
    }

    router.push('/exam')
  } catch (err) {
    await showAlertModal({
      title: 'Gagal Memulai Ujian',
      message: err.response?.data?.message || 'Gagal memulai ujian. Periksa token Anda.',
      type: 'danger'
    })
  } finally {
    startingId.value = null
  }
}

const handleLogout = async () => {
  await authStore.logout()
  router.push('/login')
}

onMounted(() => {
  loadSchedules()
})
</script>
