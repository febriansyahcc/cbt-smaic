<template>
  <footer class="bg-white border-t border-slate-200 shadow-lg sticky bottom-0 z-20 px-3 py-2.5 sm:px-6">
    <div class="max-w-4xl mx-auto flex items-center justify-between gap-2">
      <!-- Button: Sebelumnya -->
      <button
        @click="examStore.prevQuestion"
        :disabled="examStore.isFirstQuestion"
        class="flex-1 sm:flex-none sm:px-4 py-2.5 rounded-xl border border-slate-300 font-semibold text-xs sm:text-sm text-slate-700 bg-white hover:bg-slate-50 active:scale-95 disabled:opacity-40 disabled:cursor-not-allowed flex items-center justify-center space-x-1 transition shadow-xs"
      >
        <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 19l-7-7 7-7" />
        </svg>
        <span>Sebelumnya</span>
      </button>

      <!-- Button: Ragu-ragu Toggle -->
      <button
        @click="examStore.toggleDoubtful"
        :class="[
          'flex-1 sm:flex-none sm:px-4 py-2.5 rounded-xl border font-semibold text-xs sm:text-sm transition flex items-center justify-center space-x-1 shadow-xs active:scale-95',
          isDoubtful
            ? 'bg-amber-500 border-amber-600 text-white shadow-amber-200'
            : 'bg-amber-50 border-amber-300 text-amber-800 hover:bg-amber-100'
        ]"
      >
        <input 
          type="checkbox" 
          :checked="isDoubtful" 
          readonly 
          class="w-3.5 h-3.5 rounded text-amber-600 pointer-events-none accent-amber-600"
        />
        <span>Ragu-ragu</span>
      </button>

      <!-- Button: Daftar Soal Sheet Trigger -->
      <button
        @click="examStore.isDrawerOpen = true"
        class="flex-1 sm:flex-none sm:px-4 py-2.5 rounded-xl border border-indigo-200 font-semibold text-xs sm:text-sm text-indigo-700 bg-indigo-50 hover:bg-indigo-100 active:scale-95 flex items-center justify-center space-x-1.5 transition shadow-xs"
      >
        <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 6h16M4 10h16M4 14h16M4 18h16" />
        </svg>
        <span class="hidden sm:inline">Daftar Soal</span>
        <span class="sm:hidden">Soal ▦</span>
      </button>

      <!-- Button: Selanjutnya / Selesai -->
      <button
        v-if="!examStore.isLastQuestion"
        @click="examStore.nextQuestion"
        class="flex-1 sm:flex-none sm:px-4 py-2.5 rounded-xl bg-indigo-600 hover:bg-indigo-700 active:scale-95 text-white font-semibold text-xs sm:text-sm flex items-center justify-center space-x-1 transition shadow-sm"
      >
        <span>Berikutnya</span>
        <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" />
        </svg>
      </button>
      <button
        v-else
        @click="examStore.isSubmitModalOpen = true"
        class="flex-1 sm:flex-none sm:px-5 py-2.5 rounded-xl bg-emerald-600 hover:bg-emerald-700 active:scale-95 text-white font-semibold text-xs sm:text-sm flex items-center justify-center space-x-1 transition shadow-sm"
      >
        <span>Kumpulkan</span>
        <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" />
        </svg>
      </button>
    </div>
  </footer>
</template>

<script setup>
import { computed } from 'vue'
import { useExamStore } from '../../stores/exam'

const examStore = useExamStore()

const isDoubtful = computed(() => {
  const q = examStore.currentQuestion
  if (!q) return false
  return !!examStore.localAnswers[q.id]?.isDoubtful
})
</script>
