<template>
  <div>
    <!-- 1. VIOLATION WARNING MODAL (Offense 1 & 2) -->
    <div 
      v-if="warningMessage && !examStore.isBlocked" 
      class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-slate-900/80 backdrop-blur-xs animate-fade-in"
    >
      <div class="bg-white rounded-2xl max-w-md w-full p-6 shadow-2xl border-2 border-red-500 text-center">
        <div class="w-14 h-14 mx-auto rounded-full bg-red-100 text-red-600 flex items-center justify-center mb-4">
          <svg class="w-8 h-8 animate-bounce" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
          </svg>
        </div>

        <h3 class="text-lg font-black text-red-600 uppercase tracking-wide">
          Peringatan Pelanggaran Sistem!
        </h3>
        <p class="text-xs text-slate-500 mt-1 font-medium">
          {{ warningMessage }}
        </p>

        <div class="my-4 p-3 bg-red-50 rounded-xl border border-red-200">
          <p class="text-xs font-bold text-red-800">
            Pelanggaran: {{ examStore.violationCount }} dari {{ examStore.maxViolations }} kali batas maksimal.
          </p>
          <p class="text-[11px] text-red-700 mt-0.5">
            Jika batas terlampaui, ujian Anda akan <span class="font-bold underline">terkunci otomatis</span> dan harus dibuka oleh pengawas.
          </p>
        </div>

        <button
          @click="$emit('dismiss-warning')"
          class="w-full py-2.5 bg-red-600 hover:bg-red-700 active:scale-95 text-white font-bold text-xs rounded-xl shadow transition"
        >
          Saya Paham & Kembali Mengerjakan
        </button>
      </div>
    </div>

    <!-- 2. TOTAL LOCKOUT SCREEN (When Blocked) -->
    <div 
      v-if="examStore.isBlocked" 
      class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-red-950 text-white select-none"
    >
      <div class="max-w-md w-full text-center space-y-4 p-6 bg-slate-900/90 rounded-3xl border border-red-500/50 shadow-2xl">
        <div class="w-20 h-20 mx-auto rounded-full bg-red-600/20 border-2 border-red-500 flex items-center justify-center">
          <svg class="w-10 h-10 text-red-500 animate-pulse" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z" />
          </svg>
        </div>

        <div>
          <h2 class="text-2xl font-black tracking-tight text-red-400">UJIAN TERKUNCI</h2>
          <p class="text-xs text-slate-300 mt-2 leading-relaxed">
            Sesi ujian Anda telah terkunci otomatis karena terdeteksi meninggalkan antarmuka ujian melebihi batas yang diizinkan ({{ examStore.maxViolations }}x).
          </p>
        </div>

        <div class="p-3 bg-red-900/40 border border-red-700/50 rounded-2xl text-left text-xs text-red-200">
          <p class="font-semibold text-white mb-1">Langkah Penyelesaian:</p>
          <ol class="list-decimal list-inside space-y-1 text-[11px]">
            <li>Angkat tangan dan panggil Pengawas Ruang / Proktor.</li>
            <li>Pengawas akan memverifikasi dan membuka kunci sesi Anda via Live Proctoring Dashboard.</li>
            <li>Tekan tombol di bawah setelah pengawas memberikan konfirmasi.</li>
          </ol>
        </div>

        <button
          @click="checkUnlockStatus"
          :disabled="isChecking"
          class="w-full py-3 bg-indigo-600 hover:bg-indigo-700 active:scale-95 disabled:opacity-50 text-white font-bold text-xs rounded-xl shadow-lg transition flex items-center justify-center space-x-2"
        >
          <svg v-if="isChecking" class="animate-spin w-4 h-4" fill="none" viewBox="0 0 24 24">
            <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
            <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8v8H4z"></path>
          </svg>
          <span>{{ isChecking ? 'Memeriksa Otorisasi Pengawas...' : 'Periksa Status Pembukaan Kunci' }}</span>
        </button>
      </div>
    </div>

    <!-- 3. DOUBLE-CONFIRMATION SUBMIT MODAL -->
    <div 
      v-if="examStore.isSubmitModalOpen" 
      class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-slate-900/70 backdrop-blur-xs"
    >
      <div class="bg-white rounded-3xl max-w-lg w-full p-6 shadow-2xl border border-slate-200">
        <div class="flex items-center justify-between pb-3 border-b border-slate-200">
          <h3 class="text-base font-bold text-slate-900">Konfirmasi Pengumpulan Ujian</h3>
          <button 
            @click="examStore.isSubmitModalOpen = false"
            class="text-slate-400 hover:text-slate-600 p-1"
          >
            <svg class="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
            </svg>
          </button>
        </div>

        <!-- Breakdown Statistics -->
        <div class="grid grid-cols-3 gap-2.5 my-4">
          <div class="p-3 bg-emerald-50 border border-emerald-200 rounded-2xl text-center">
            <span class="text-xl font-black text-emerald-700">{{ examStore.answeredCount }}</span>
            <p class="text-[11px] font-semibold text-emerald-800 mt-0.5">Terjawab</p>
          </div>
          <div class="p-3 bg-amber-50 border border-amber-200 rounded-2xl text-center">
            <span class="text-xl font-black text-amber-700">{{ examStore.doubtfulCount }}</span>
            <p class="text-[11px] font-semibold text-amber-800 mt-0.5">Ragu-ragu</p>
          </div>
          <div class="p-3 bg-slate-50 border border-slate-200 rounded-2xl text-center">
            <span class="text-xl font-black text-slate-700">{{ examStore.unansweredCount }}</span>
            <p class="text-[11px] font-semibold text-slate-800 mt-0.5">Belum Terisi</p>
          </div>
        </div>

        <!-- Warnings if any incomplete -->
        <div 
          v-if="examStore.unansweredCount > 0 || examStore.doubtfulCount > 0"
          class="p-3 bg-amber-50 border border-amber-200 rounded-xl text-xs text-amber-800 flex items-start space-x-2 mb-4"
        >
          <svg class="w-4 h-4 text-amber-600 flex-shrink-0 mt-0.5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
          </svg>
          <div>
            <p class="font-bold">Peringatan:</p>
            <p class="text-[11px]">
              Masih terdapat <strong>{{ examStore.unansweredCount }} soal belum dijawab</strong> dan <strong>{{ examStore.doubtfulCount }} soal ragu-ragu</strong>.
            </p>
          </div>
        </div>

        <!-- Agreement Checkbox -->
        <label class="flex items-start space-x-2.5 p-3 rounded-xl border border-slate-200 bg-slate-50/70 cursor-pointer mb-5 text-xs text-slate-700">
          <input 
            type="checkbox" 
            v-model="confirmed"
            class="mt-0.5 w-4 h-4 rounded text-indigo-600 focus:ring-indigo-500"
          />
          <span class="leading-relaxed select-none">
            Saya telah memeriksa jawaban saya dengan sungguh-sungguh dan yakin untuk mengakhiri sesi ujian ini.
          </span>
        </label>

        <!-- Actions -->
        <div class="flex items-center space-x-2">
          <button
            @click="examStore.isSubmitModalOpen = false"
            class="flex-1 py-2.5 rounded-xl border border-slate-300 font-semibold text-xs text-slate-700 hover:bg-slate-100 transition"
          >
            Kembali Periksa
          </button>
          <button
            @click="handleSubmit"
            :disabled="!confirmed || isSubmitting"
            class="flex-1 py-2.5 rounded-xl bg-emerald-600 hover:bg-emerald-700 active:scale-95 disabled:opacity-40 disabled:cursor-not-allowed text-white font-bold text-xs shadow-md transition flex items-center justify-center space-x-1"
          >
            <svg v-if="isSubmitting" class="animate-spin w-4 h-4" fill="none" viewBox="0 0 24 24">
              <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
              <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8v8H4z"></path>
            </svg>
            <span>{{ isSubmitting ? 'Mengumpulkan...' : 'Kumpulkan Jawaban' }}</span>
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { useExamStore } from '../../stores/exam'
import api from '../../services/api'
import { useDialog } from '@/composables/useDialog'

const props = defineProps({
  warningMessage: {
    type: String,
    default: '',
  }
})

defineEmits(['dismiss-warning'])

const examStore = useExamStore()
const router = useRouter()
const { alert: showAlertModal, toast: showToast } = useDialog()

const confirmed = ref(false)
const isSubmitting = ref(false)
const isChecking = ref(false)

const handleSubmit = async () => {
  if (!confirmed.value || isSubmitting.value) return
  isSubmitting.value = true
  try {
    await examStore.submitExam()
    examStore.isSubmitModalOpen = false
    router.push('/exam-finished')
  } catch (err) {
    await showAlertModal({
      title: 'Gagal Mengumpulkan Ujian',
      message: err.response?.data?.message || 'Gagal mengumpulkan ujian. Silakan coba lagi.',
      type: 'danger'
    })
  } finally {
    isSubmitting.value = false
  }
}

const checkUnlockStatus = async () => {
  const token = localStorage.getItem('cbt_active_token')
  const scheduleId = localStorage.getItem('cbt_active_schedule_id')
  if (!token || !scheduleId) {
    router.push('/student')
    return
  }

  isChecking.value = true
  try {
    await examStore.startExam(scheduleId, token)
    if (!examStore.isBlocked) {
      showToast('Kunci telah dibuka oleh pengawas. Anda dapat melanjutkan ujian!', 'success')
    } else {
      await showAlertModal({
        title: 'Status Masih Terkunci',
        message: 'Sesi Anda masih berstatus terkunci. Mohon tunggu konfirmasi pengawas.',
        type: 'warning'
      })
    }
  } catch (e) {
    //
  } finally {
    isChecking.value = false
  }
}
</script>
