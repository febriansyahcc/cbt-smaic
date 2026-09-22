<template>
  <div class="min-h-full bg-surface-50 flex flex-col justify-center items-center p-4">
    <div class="max-w-md w-full bg-white rounded-3xl p-6 sm:p-8 border border-slate-200 shadow-xl text-center space-y-5">
      <!-- School Header -->
      <div class="flex items-center justify-center space-x-2 pb-2 border-b border-slate-100">
        <img src="/logo-smic.png" alt="Logo SMAS Islamic Centre Demak" class="w-8 h-8 object-contain shrink-0" />
        <span class="text-xs font-bold text-slate-700 tracking-tight">SMAS ISLAMIC CENTRE DEMAK</span>
      </div>

      <!-- Success Icon -->
      <div class="w-20 h-20 mx-auto rounded-full bg-emerald-100 text-emerald-600 flex items-center justify-center shadow-lg shadow-emerald-100">
        <svg class="w-10 h-10" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2.5" d="M5 13l4 4L19 7" />
        </svg>
      </div>

      <div>
        <h2 class="text-2xl font-black text-slate-900 tracking-tight">Ujian Selesai!</h2>
        <p class="text-xs text-slate-500 mt-1">
          Seluruh lembar jawaban Anda telah berhasil tersimpan dan terverifikasi oleh server CBT.
        </p>
      </div>

      <!-- Student & Exam Details -->
      <div class="bg-slate-50 rounded-2xl p-4 border border-slate-100 text-left text-xs space-y-2 text-slate-600">
        <div class="flex justify-between">
          <span class="text-slate-400">Nama Siswa:</span>
          <span class="font-bold text-slate-800">{{ authStore.user?.full_name }}</span>
        </div>
        <div class="flex justify-between">
          <span class="text-slate-400">Kelas:</span>
          <span class="font-semibold text-slate-800">{{ authStore.profile?.class_room?.name || 'XII MIPA 1' }}</span>
        </div>
        <div class="flex justify-between">
          <span class="text-slate-400">Waktu Pengumpulan:</span>
          <span class="font-mono text-slate-800">{{ currentTime }}</span>
        </div>
      </div>

      <div class="pt-2 flex flex-col space-y-2">
        <button
          @click="handleBackToHome"
          class="w-full py-3 bg-indigo-600 hover:bg-indigo-700 active:scale-95 text-white font-bold text-xs rounded-xl shadow-md transition"
        >
          Kembali ke Portal Siswa
        </button>
        <button
          @click="handleLogout"
          class="w-full py-2.5 bg-slate-100 hover:bg-slate-200 text-slate-700 font-semibold text-xs rounded-xl transition"
        >
          Keluar dari Akun (Logout)
        </button>
      </div>
    </div>
  </div>
</template>

<script setup>
import { computed } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '../../stores/auth'

const router = useRouter()
const authStore = useAuthStore()

const currentTime = computed(() => {
  return new Date().toLocaleTimeString('id-ID', { hour: '2-digit', minute: '2-digit', second: '2-digit' }) + ' WIB'
})

const handleBackToHome = () => {
  router.push('/student')
}

const handleLogout = async () => {
  await authStore.logout()
  router.push('/login')
}
</script>
