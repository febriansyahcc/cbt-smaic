<template>
  <Transition
    enter-active-class="transition-opacity duration-200 ease-out"
    leave-active-class="transition-opacity duration-150 ease-in"
    enter-from-class="opacity-0"
    leave-to-class="opacity-0"
  >
    <div
      v-if="modelValue"
      class="fixed inset-0 z-50 flex items-start justify-center p-4 bg-slate-900/60 backdrop-blur-sm overflow-y-auto"
      @click.self="$emit('update:modelValue', false)"
    >
      <Transition
        enter-active-class="transition-all duration-200 ease-out"
        leave-active-class="transition-all duration-150 ease-in"
        enter-from-class="opacity-0 scale-95 translate-y-2"
        leave-to-class="opacity-0 scale-95 translate-y-2"
        appear
      >
        <div
          v-if="modelValue"
          class="bg-white rounded-3xl w-full max-w-2xl shadow-2xl border border-slate-100 flex flex-col max-h-[90vh] my-4"
          @click.stop
        >
          <!-- Header -->
          <div class="flex items-start justify-between p-5 border-b border-slate-100 shrink-0">
            <div class="min-w-0 flex-1 pr-3">
              <div class="flex items-center gap-2 flex-wrap mb-0.5">
                <h3 class="text-sm font-black text-slate-900 truncate">{{ sessionData?.student_name || 'Detail Jawaban Siswa' }}</h3>
                <span
                  v-if="sessionData?.total_score != null"
                  class="px-2.5 py-0.5 bg-indigo-50 text-indigo-700 border border-indigo-100 text-[11px] font-bold rounded-full shrink-0"
                >
                  Nilai: {{ formatScore(sessionData.total_score) }}
                </span>
              </div>
              <p class="text-[11px] text-slate-500 font-mono">
                NIS: {{ sessionData?.student_nis || '-' }}
                <span v-if="sessionData?.schedule_title" class="font-sans ml-2 text-slate-400">
                  — {{ sessionData.schedule_title }}
                </span>
              </p>
            </div>
            <button
              type="button"
              @click="$emit('update:modelValue', false)"
              class="p-1.5 text-slate-400 hover:text-slate-600 hover:bg-slate-100 rounded-xl transition cursor-pointer shrink-0"
              title="Tutup"
            >
              <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
              </svg>
            </button>
          </div>

          <!-- Scrollable Body -->
          <div class="overflow-y-auto flex-1 p-5">
            <!-- Loading State -->
            <div v-if="loading" class="py-16 flex flex-col items-center justify-center gap-3 text-slate-400">
              <svg class="w-8 h-8 animate-spin text-indigo-500" fill="none" viewBox="0 0 24 24">
                <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" />
                <path
                  class="opacity-75"
                  fill="currentColor"
                  d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
                />
              </svg>
              <span class="text-xs font-medium">Memuat jawaban siswa...</span>
            </div>

            <!-- Questions List -->
            <div v-else-if="sessionData?.questions?.length" class="space-y-4">
              <div
                v-for="q in sessionData.questions"
                :key="q.question_number"
                class="border border-slate-200 rounded-2xl overflow-hidden"
              >
                <!-- Question Meta Header -->
                <div class="px-4 py-2.5 bg-slate-50 border-b border-slate-200 flex items-center justify-between gap-3">
                  <div class="flex items-center gap-2 text-xs flex-wrap">
                    <span class="font-black text-slate-900">No. {{ q.question_number }}</span>
                    <span class="px-2 py-0.5 bg-white border border-slate-200 text-slate-600 font-semibold rounded-md text-[10px]">
                      {{ formatQuestionType(q.type) }}
                    </span>
                    <span class="text-slate-500 text-[11px]">Bobot: {{ q.score_weight }}</span>
                  </div>
                  <div class="flex items-center gap-2 shrink-0">
                    <!-- Correct/Wrong indicator for MC -->
                    <span
                      v-if="q.type === 'MULTIPLE_CHOICE' && q.is_correct === true"
                      class="flex items-center gap-1 text-emerald-600 font-bold text-[11px]"
                    >
                      <svg class="w-3.5 h-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2.5" d="M5 13l4 4L19 7" />
                      </svg>
                      Benar
                    </span>
                    <span
                      v-else-if="q.type === 'MULTIPLE_CHOICE' && q.is_correct === false && !q.selected_option"
                      class="text-[11px] font-bold text-slate-500"
                    >Tidak Menjawab</span>
                    <span
                      v-else-if="q.type === 'MULTIPLE_CHOICE' && q.is_correct === false"
                      class="flex items-center gap-1 text-rose-600 font-bold text-[11px]"
                    >
                      <svg class="w-3.5 h-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2.5" d="M6 18L18 6M6 6l12 12" />
                      </svg>
                      Salah
                    </span>
                    <!-- Score awarded -->
                    <span
                      v-if="q.score_awarded != null"
                      class="text-[11px] font-bold text-indigo-700 bg-indigo-50 border border-indigo-100 px-2 py-0.5 rounded-full"
                    >
                      +{{ formatScore(q.score_awarded) }}
                    </span>
                  </div>
                </div>

                <!-- Question Content Body -->
                <div class="px-4 py-3">
                  <!-- Question text / HTML -->
                  <div class="text-xs text-slate-800 mb-3 leading-relaxed">
                    <RichContentRenderer :content="q.content_html" />
                  </div>

                  <!-- MULTIPLE_CHOICE: Options List -->
                  <div v-if="q.type === 'MULTIPLE_CHOICE' && q.options?.length" class="space-y-1.5">
                    <div
                      v-for="opt in q.options"
                      :key="opt.key"
                      :class="['flex items-start gap-2.5 px-3 py-2.5 rounded-xl border text-xs', getOptionClass(q, opt)]"
                    >
                      <span class="font-black shrink-0 mt-px w-4 text-center">{{ opt.key }}</span>
                      <div class="flex-1 min-w-0 leading-relaxed">
                        <span v-if="!isHtmlContent(opt.text)">{{ opt.text }}</span>
                        <div v-else v-html="opt.text"></div>
                        <img
                          v-if="opt.image_url"
                          :src="opt.image_url"
                          class="mt-1.5 max-h-24 rounded-lg object-contain"
                          loading="lazy"
                        />
                      </div>
                      <!-- Icon / label on right -->
                      <div class="shrink-0 mt-px">
                        <svg
                          v-if="q.selected_option === opt.key && q.is_correct === true"
                          class="w-4 h-4 text-emerald-600"
                          fill="none"
                          viewBox="0 0 24 24"
                          stroke="currentColor"
                        >
                          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2.5" d="M5 13l4 4L19 7" />
                        </svg>
                        <svg
                          v-else-if="q.selected_option === opt.key && q.is_correct === false"
                          class="w-4 h-4 text-rose-600"
                          fill="none"
                          viewBox="0 0 24 24"
                          stroke="currentColor"
                        >
                          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2.5" d="M6 18L18 6M6 6l12 12" />
                        </svg>
                        <span
                          v-else-if="q.correct_key === opt.key && q.selected_option !== opt.key && q.is_correct !== null"
                          class="text-[9px] font-black text-emerald-700 bg-emerald-50 border border-emerald-200 px-1.5 py-0.5 rounded-md whitespace-nowrap"
                        >
                          Kunci
                        </span>
                      </div>
                    </div>
                  </div>

                  <!-- ESSAY / SHORT_ANSWER: Answer Text -->
                  <div v-else-if="q.type === 'ESSAY' || q.type === 'SHORT_ANSWER'" class="space-y-2">
                    <div
                      v-if="q.answer_text"
                      class="bg-slate-50 border border-slate-200 rounded-xl px-3.5 py-3 text-xs text-slate-700 leading-relaxed whitespace-pre-wrap"
                    >{{ q.answer_text }}</div>
                    <div
                      v-else
                      class="bg-slate-50 border border-slate-200 rounded-xl px-3.5 py-3 text-xs text-slate-400 italic"
                    >
                      Tidak ada jawaban
                    </div>
                    <div v-if="q.is_graded" class="flex items-center gap-1.5 text-xs text-indigo-700 font-semibold">
                      <svg class="w-3.5 h-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2" />
                      </svg>
                      Nilai: {{ q.score_awarded != null ? formatScore(q.score_awarded) : '?' }} / {{ q.score_weight }}
                    </div>
                    <div v-else-if="q.answer_text" class="text-[11px] text-amber-600 font-medium">
                      Belum dinilai
                    </div>
                  </div>
                </div>
              </div>
            </div>

            <!-- Empty / No Data -->
            <div v-else class="py-16 text-center text-slate-400">
              <p class="text-xs font-bold text-slate-700">Tidak ada data jawaban</p>
              <p class="text-[11px] mt-1">Data belum tersedia atau terjadi kesalahan saat memuat.</p>
            </div>
          </div>

          <!-- Footer -->
          <div class="px-5 py-3.5 border-t border-slate-100 flex items-center justify-end shrink-0">
            <button
              type="button"
              @click="$emit('update:modelValue', false)"
              class="px-4 py-2 bg-slate-100 hover:bg-slate-200 active:scale-95 text-slate-700 font-bold text-xs rounded-xl transition cursor-pointer"
            >
              Tutup
            </button>
          </div>
        </div>
      </Transition>
    </div>
  </Transition>
</template>

<script setup>
import RichContentRenderer from '../common/RichContentRenderer.vue'

defineProps({
  modelValue: {
    type: Boolean,
    default: false
  },
  sessionData: {
    type: Object,
    default: null
  },
  loading: {
    type: Boolean,
    default: false
  }
})

defineEmits(['update:modelValue'])

const formatScore = (val) => {
  if (val == null) return '-'
  return typeof val === 'number' ? val.toFixed(1) : val
}

const formatQuestionType = (type) => {
  switch (type) {
    case 'MULTIPLE_CHOICE': return 'Pilihan Ganda'
    case 'ESSAY': return 'Esai'
    case 'SHORT_ANSWER': return 'Isian Singkat'
    default: return type || '-'
  }
}

const isHtmlContent = (text) => {
  if (!text) return false
  return /<[a-z][\s\S]*>/i.test(text)
}

/**
 * Returns Tailwind classes for each option row based on correctness state.
 */
const getOptionClass = (q, opt) => {
  const key = opt.key
  const selected = q.selected_option
  const correct = q.correct_key
  const isCorrect = q.is_correct

  // Not yet submitted (is_correct === null/undefined): highlight selected with indigo only
  if (isCorrect === null || isCorrect === undefined) {
    if (selected === key) return 'border-indigo-400 bg-indigo-50 text-indigo-900'
    return 'border-slate-200 bg-white text-slate-700'
  }

  // Submitted: apply result colours
  if (selected === key && isCorrect === true) {
    return 'border-emerald-500 bg-emerald-50 text-emerald-900'
  }
  // Jika siswa tidak menjawab (selected kosong), jangan highlight merah — hanya tampilkan kunci
  if (selected === key && isCorrect === false && selected !== '') {
    return 'border-rose-500 bg-rose-50 text-rose-900'
  }
  if (correct === key && selected !== key) {
    return 'border-emerald-300 border-dashed bg-emerald-50/40 text-emerald-900'
  }
  return 'border-slate-200 bg-white text-slate-700'
}
</script>
