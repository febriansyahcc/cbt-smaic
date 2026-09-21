<template>
  <header class="bg-white border-b border-slate-200 shadow-sm sticky top-0 z-30 px-3 py-2 sm:px-6">
    <div class="max-w-5xl mx-auto flex items-center justify-between">
      <!-- Left: Subject & Info -->
      <div class="flex items-center space-x-2">
        <img
          src="/logo-smic.png"
          alt="SMAS Islamic Centre Demak"
          class="w-8 h-8 object-contain drop-shadow-xs shrink-0"
        />
        <div class="leading-tight">
          <h1 class="text-xs sm:text-sm font-semibold text-slate-900 truncate max-w-[130px] sm:max-w-xs">
            {{ examStore.subjectName }}
          </h1>
          <p class="text-[10px] sm:text-xs text-slate-500 truncate max-w-[130px] sm:max-w-xs">
            {{ examStore.scheduleTitle }}
          </p>
        </div>
      </div>

      <!-- Center: Monospace Authoritative Countdown Timer -->
      <div 
        class="flex items-center space-x-1.5 px-3 py-1 rounded-full text-xs sm:text-sm font-bold shadow-inner"
        :class="timerClass"
      >
        <svg class="w-4 h-4 animate-pulse" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
        </svg>
        <span class="font-mono tabular-timer tracking-wider">
          {{ examStore.formattedTimer }}
        </span>
      </div>

      <!-- Right: Sync Status & Submit Button -->
      <div class="flex items-center space-x-2">
        <!-- Network & Sync Indicator -->
        <div 
          class="hidden sm:flex items-center space-x-1 text-[11px] font-medium px-2 py-1 rounded-full"
          :class="syncBadgeClass"
        >
          <span class="w-2 h-2 rounded-full" :class="syncDotClass"></span>
          <span>{{ syncStatusText }}</span>
        </div>

        <button
          @click="examStore.isSubmitModalOpen = true"
          class="px-2.5 sm:px-3 py-1.5 bg-emerald-600 hover:bg-emerald-700 active:scale-95 text-white font-medium text-xs rounded-lg shadow transition flex items-center space-x-1"
        >
          <svg class="w-3.5 h-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" />
          </svg>
          <span>Selesai</span>
        </button>
      </div>
    </div>
  </header>
</template>

<script setup>
import { computed } from 'vue'
import { useExamStore } from '../../stores/exam'

const examStore = useExamStore()

const isUrgentTime = computed(() => examStore.remainingSeconds < 300) // less than 5 min

const timerClass = computed(() => {
  if (isUrgentTime.value) {
    return 'bg-red-100 text-red-700 border border-red-200'
  }
  return 'bg-slate-100 text-indigo-900 border border-slate-200'
})

const syncBadgeClass = computed(() => {
  if (!examStore.isOnline) return 'bg-red-50 text-red-700 border border-red-200'
  if (examStore.pendingCount > 0 || examStore.isSyncing) return 'bg-amber-50 text-amber-700 border border-amber-200'
  return 'bg-emerald-50 text-emerald-700 border border-emerald-200'
})

const syncDotClass = computed(() => {
  if (!examStore.isOnline) return 'bg-red-500'
  if (examStore.pendingCount > 0 || examStore.isSyncing) return 'bg-amber-500 animate-pulse'
  return 'bg-emerald-500'
})

const syncStatusText = computed(() => {
  if (!examStore.isOnline) return 'Offline (Tersimpan Lokal)'
  if (examStore.isSyncing) return 'Menyinkronkan...'
  if (examStore.pendingCount > 0) return `${examStore.pendingCount} Menunggu Sync`
  return 'Tersimpan Online'
})
</script>
