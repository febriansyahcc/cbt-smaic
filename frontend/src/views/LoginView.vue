<template>
  <div class="min-h-full flex flex-col justify-center py-12 px-4 sm:px-6 lg:px-8 bg-surface-50">
    <div class="sm:mx-auto sm:w-full sm:max-w-md text-center">
      <!-- School / SMAS Islamic Centre Demak Logo -->
      <img
        src="/logo-smic.png"
        alt="Logo SMAS Islamic Centre Demak"
        class="mx-auto w-20 h-20 object-contain drop-shadow-md hover:scale-105 transition duration-300"
      />
      <h2 class="mt-3 text-xl sm:text-2xl font-black text-slate-900 tracking-tight">
        SMAS ISLAMIC CENTRE DEMAK
      </h2>
      <p class="mt-1 text-xs text-slate-500 font-medium">
        Sistem Ujian Sekolah Terintegrasi • CBT & Anti-Kecurangan
      </p>
    </div>

    <div class="mt-6 sm:mx-auto sm:w-full sm:max-w-md">
      <div class="bg-white py-8 px-6 shadow-xl shadow-slate-200/50 rounded-3xl border border-slate-200">
        <!-- Error Alert -->
        <div v-if="authStore.error" class="mb-4 p-3 rounded-xl bg-red-50 border border-red-200 text-xs font-semibold text-red-700 flex items-center space-x-2">
          <svg class="w-4 h-4 text-red-500 flex-shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
          </svg>
          <span>{{ authStore.error }}</span>
        </div>

        <form @submit.prevent="handleLogin" class="space-y-4">
          <div>
            <label class="block text-xs font-bold text-slate-700 mb-1">
              Nomor Induk Siswa (NIS) / Username
            </label>
            <input
              v-model="username"
              type="text"
              required
              autocomplete="username"
              placeholder="NIS / Username"
              class="w-full px-3.5 py-2.5 rounded-xl border border-slate-300 text-sm focus:ring-2 focus:ring-indigo-600 focus:border-indigo-600 transition"
            />
          </div>

          <div>
            <label class="block text-xs font-bold text-slate-700 mb-1">
              Kata Sandi (Password)
            </label>
            <input
              v-model="password"
              type="password"
              required
              autocomplete="current-password"
              placeholder="Masukkan password akun Anda"
              class="w-full px-3.5 py-2.5 rounded-xl border border-slate-300 text-sm focus:ring-2 focus:ring-indigo-600 focus:border-indigo-600 transition"
            />
          </div>

          <button
            type="submit"
            :disabled="authStore.loading"
            class="w-full py-3 bg-indigo-600 hover:bg-indigo-700 active:scale-98 disabled:opacity-50 text-white font-bold text-sm rounded-xl shadow-md shadow-indigo-200 transition flex items-center justify-center space-x-2"
          >
            <svg v-if="authStore.loading" class="animate-spin w-4 h-4" fill="none" viewBox="0 0 24 24">
              <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
              <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8v8H4z"></path>
            </svg>
            <span>{{ authStore.loading ? 'Memvalidasi...' : 'Masuk ke Sistem Ujian' }}</span>
          </button>
        </form>
      </div>

      <!-- School & Foundation Footer -->
      <div class="mt-6 text-center text-xs text-slate-500 font-medium space-y-0.5">
        <p class="font-semibold text-slate-700">Yayasan Islamic Centre Sultan Fatah Demak</p>
        <p class="text-[11px] text-slate-400">© 2026 SMAS Islamic Centre Demak. All rights reserved.</p>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '../stores/auth'
import { homePathFor } from '../utils/access'

const router = useRouter()
const authStore = useAuthStore()

const username = ref('')
const password = ref('')

onMounted(() => {
  const loginErr = sessionStorage.getItem('cbt_login_error')
  if (loginErr) {
    authStore.error = loginErr
    sessionStorage.removeItem('cbt_login_error')
  }
})

const handleLogin = async () => {
  try {
    const data = await authStore.login(username.value, password.value)
    router.push(homePathFor(data.user))
  } catch (err) {
    // Handled in store
  }
}
</script>
