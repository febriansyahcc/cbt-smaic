<template>
  <div v-if="examStore.isDrawerOpen" class="fixed inset-0 z-50 overflow-hidden">
    <!-- Backdrop -->
    <div 
      @click="examStore.isDrawerOpen = false" 
      class="fixed inset-0 bg-slate-900/60 backdrop-blur-xs transition-opacity"
    ></div>

    <!-- Bottom Sheet Container (Max 80vh height on mobile) -->
    <div class="fixed inset-x-0 bottom-0 max-h-[85vh] bg-white rounded-t-3xl shadow-2xl flex flex-col z-10 transition-transform">
      <!-- Drawer Handle & Header -->
      <div class="p-4 border-b border-slate-200 flex flex-col items-center">
        <div class="w-12 h-1.5 bg-slate-300 rounded-full mb-3"></div>
        <div class="w-full flex items-center justify-between">
          <div>
            <h2 class="text-sm font-bold text-slate-900">Kisi Nomor Soal</h2>
            <p class="text-xs text-slate-500">Pilih nomor untuk melompat langsung</p>
          </div>
          <button 
            @click="examStore.isDrawerOpen = false"
            class="p-1.5 rounded-full hover:bg-slate-100 text-slate-400 hover:text-slate-700"
          >
            <svg class="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
            </svg>
          </button>
        </div>

        <!-- Legend -->
        <div class="flex flex-wrap items-center justify-start gap-3 w-full mt-3 text-[11px] font-medium text-slate-600 bg-slate-50 p-2.5 rounded-xl border border-slate-100">
          <div class="flex items-center space-x-1.5">
            <span class="w-3 h-3 rounded-full bg-emerald-500"></span>
            <span>Terjawab ({{ examStore.answeredCount }})</span>
          </div>
          <div class="flex items-center space-x-1.5">
            <span class="w-3 h-3 rounded-full bg-amber-500"></span>
            <span>Ragu-ragu ({{ examStore.doubtfulCount }})</span>
          </div>
          <div class="flex items-center space-x-1.5">
            <span class="w-3 h-3 rounded-full bg-slate-200 border border-slate-300"></span>
            <span>Belum Dijawab ({{ examStore.unansweredCount }})</span>
          </div>
        </div>
      </div>

      <!-- Question Grid Area -->
      <div class="flex-1 overflow-y-auto p-4 sm:p-6">
        <div class="grid grid-cols-5 sm:grid-cols-8 md:grid-cols-10 gap-2.5 sm:gap-3">
          <button
            v-for="(q, idx) in examStore.questions"
            :key="q.id"
            @click="examStore.goToQuestion(idx)"
            :class="[
              'h-12 rounded-xl flex flex-col items-center justify-center font-bold text-xs transition-all relative shadow-xs active:scale-90',
              getItemClass(q, idx)
            ]"
          >
            <span>{{ idx + 1 }}</span>
            <span v-if="getAnswerKey(q)" class="text-[9px] uppercase opacity-90 -mt-0.5">
              {{ getAnswerKey(q) }}
            </span>
            <!-- Active indicator dot -->
            <span 
              v-if="examStore.currentIndex === idx"
              class="absolute -top-1 -right-1 w-2.5 h-2.5 rounded-full bg-indigo-600 ring-2 ring-white"
            ></span>
          </button>
        </div>
      </div>

      <!-- Bottom Close Button -->
      <div class="p-3 border-t border-slate-200 bg-slate-50">
        <button
          @click="examStore.isDrawerOpen = false"
          class="w-full py-2.5 bg-slate-200 hover:bg-slate-300 text-slate-700 font-semibold text-xs rounded-xl transition"
        >
          Tutup Kisi Soal
        </button>
      </div>
    </div>
  </div>
</template>

<script setup>
import { useExamStore } from '../../stores/exam'

const examStore = useExamStore()

const getAnswerKey = (q) => {
  const ans = examStore.localAnswers[q.id]
  if (q.type === 'SHORT_ANSWER' || q.type === 'ESSAY') {
    return (ans?.answerText && ans.answerText.trim() !== '') ? '✓' : ''
  }
  return ans?.selectedOption || ''
}

const getItemClass = (q, idx) => {
  const ans = examStore.localAnswers[q.id]
  const isCurrent = examStore.currentIndex === idx
  const activeRing = isCurrent ? 'ring-2 ring-indigo-600 ring-offset-2' : ''

  if (ans?.isDoubtful) {
    return `bg-amber-500 text-white shadow-amber-200 ${activeRing}`
  }
  const isAnswered = (ans?.selectedOption && ans.selectedOption !== '') || (ans?.answerText && ans.answerText.trim() !== '')
  if (isAnswered) {
    return `bg-emerald-600 text-white shadow-emerald-200 ${activeRing}`
  }
  return `bg-slate-100 text-slate-700 border border-slate-300 hover:bg-slate-200 ${activeRing}`
}
</script>
