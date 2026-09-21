<template>
  <div v-if="question" class="max-w-4xl mx-auto w-full px-3 py-4 sm:px-6">
    <div class="bg-white rounded-2xl shadow-sm border border-slate-200 overflow-hidden">
      <!-- Question Header -->
      <div class="bg-slate-50/80 px-4 py-3 border-b border-slate-200 flex flex-wrap items-center justify-between gap-2">
        <div class="flex items-center flex-wrap gap-2">
          <span class="px-2.5 py-1 bg-indigo-50 border border-indigo-200 text-indigo-700 font-bold text-xs rounded-md">
            Soal No. {{ examStore.currentIndex + 1 }}
          </span>
          <!-- Type Badge -->
          <span
            class="px-2.5 py-1 rounded-md text-xs font-bold border"
            :class="getTypeBadgeClass(question.type)"
          >
            {{ getTypeLabel(question.type) }}
          </span>
          <span class="text-xs text-slate-500 font-medium">
            dari {{ examStore.totalQuestions }} Soal
          </span>
          <span 
            v-if="isDoubtful" 
            class="px-2 py-0.5 bg-amber-100 border border-amber-300 text-amber-800 text-[11px] font-semibold rounded-full flex items-center space-x-1"
          >
            <span>★ Ragu-ragu</span>
          </span>
        </div>

        <!-- Font Size Controls -->
        <div class="flex items-center space-x-1 bg-white border border-slate-200 rounded-lg p-0.5 text-xs font-semibold text-slate-600 shadow-xs">
          <button 
            @click="fontSize = 'text-sm'" 
            :class="fontSize === 'text-sm' ? 'bg-indigo-600 text-white' : 'hover:bg-slate-100'"
            class="px-2 py-1 rounded transition cursor-pointer"
            title="Font Kecil"
          >
            A-
          </button>
          <button 
            @click="fontSize = 'text-base'" 
            :class="fontSize === 'text-base' ? 'bg-indigo-600 text-white' : 'hover:bg-slate-100'"
            class="px-2 py-1 rounded transition cursor-pointer"
            title="Font Normal"
          >
            A
          </button>
          <button 
            @click="fontSize = 'text-lg'" 
            :class="fontSize === 'text-lg' ? 'bg-indigo-600 text-white' : 'hover:bg-slate-100'"
            class="px-2 py-1 rounded transition cursor-pointer"
            title="Font Besar"
          >
            A+
          </button>
        </div>
      </div>

      <!-- Question Text Content -->
      <div class="p-4 sm:p-6 border-b border-slate-100">
        <div :class="[fontSize, 'leading-relaxed text-slate-800 break-words font-medium']">
          <RichContentRenderer :content="question.content_html" />
        </div>
      </div>

      <!-- 1. MULTIPLE CHOICE OPTIONS -->
      <div v-if="!question.type || question.type === 'MULTIPLE_CHOICE'" class="p-4 sm:p-6 bg-slate-50/40 space-y-3">
        <div 
          v-for="opt in question.options" 
          :key="opt.key"
          @click="examStore.selectOption(opt.key)"
          :class="[
            'cursor-pointer select-none rounded-2xl border-2 p-3 sm:p-4 transition-all duration-150 flex flex-col space-y-2 active:scale-[0.99]',
            selectedKey === opt.key 
              ? 'border-indigo-600 bg-indigo-50/70 shadow-sm' 
              : 'border-slate-200 bg-white hover:border-slate-300 hover:bg-slate-50/50'
          ]"
        >
          <div class="flex items-start space-x-3">
            <!-- Radio Key Badge -->
            <div 
              :class="[
                'w-8 h-8 rounded-full flex-shrink-0 flex items-center justify-center font-bold text-sm transition-colors mt-0.5',
                selectedKey === opt.key 
                  ? 'bg-indigo-600 text-white shadow-sm' 
                  : 'bg-slate-100 text-slate-600 border border-slate-300'
              ]"
            >
              {{ opt.key }}
            </div>

            <!-- Option Text -->
            <div class="flex-1 pt-0.5">
              <RichContentRenderer :content="opt.text" :custom-class="fontSize" />
            </div>
          </div>

          <!-- Option Image if present -->
          <div v-if="opt.image_url" class="pl-11">
            <img :src="opt.image_url" alt="Pilihan Jawaban" class="max-h-48 rounded-xl border border-slate-200 object-contain shadow-xs" />
          </div>
        </div>
      </div>

      <!-- 2. SHORT ANSWER INPUT -->
      <div v-else-if="question.type === 'SHORT_ANSWER'" class="p-4 sm:p-6 bg-slate-50/40 space-y-3">
        <label class="block text-xs font-bold text-slate-700">Jawaban Singkat:</label>
        <div class="relative">
          <input
            :value="currentAnswerText"
            @input="onTextInput($event.target.value)"
            type="text"
            placeholder="Ketik jawaban singkat Anda di sini..."
            :class="[fontSize, 'w-full px-4 py-3 rounded-xl border-2 border-slate-300 focus:border-indigo-600 focus:bg-white focus:outline-none transition bg-white font-medium text-slate-900 shadow-xs']"
          />
        </div>
        <div class="flex items-center justify-between text-xs text-slate-500 pt-1">
          <span>* Jawaban disimpan secara otomatis</span>
          <span>{{ (currentAnswerText || '').length }} Karakter</span>
        </div>
      </div>

      <!-- 3. ESSAY / URAIAN TEXTAREA -->
      <div v-else-if="question.type === 'ESSAY'" class="p-4 sm:p-6 bg-slate-50/40 space-y-3">
        <label class="block text-xs font-bold text-slate-700">Jawaban Uraian / Essay:</label>
        <div class="relative">
          <textarea
            :value="currentAnswerText"
            @input="onTextInput($event.target.value)"
            rows="6"
            placeholder="Tuliskan uraian atau penjelasan lengkap jawaban Anda di sini..."
            :class="[fontSize, 'w-full p-4 rounded-xl border-2 border-slate-300 focus:border-indigo-600 focus:bg-white focus:outline-none transition bg-white font-medium text-slate-900 shadow-xs leading-relaxed']"
          ></textarea>
        </div>
        <div class="flex items-center justify-between text-xs text-slate-500 pt-1">
          <span>* Jawaban disimpan secara otomatis</span>
          <span>{{ (currentAnswerText || '').length }} Karakter • {{ getWordCount(currentAnswerText) }} Kata</span>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed } from 'vue'
import { useExamStore } from '../../stores/exam'
import RichContentRenderer from '../common/RichContentRenderer.vue'

const examStore = useExamStore()
const fontSize = ref('text-base')

const question = computed(() => examStore.currentQuestion)

const selectedKey = computed(() => {
  if (!question.value) return ''
  return examStore.localAnswers[question.value.id]?.selectedOption || ''
})

const currentAnswerText = computed(() => {
  if (!question.value) return ''
  return examStore.localAnswers[question.value.id]?.answerText || ''
})

const isDoubtful = computed(() => {
  if (!question.value) return false
  return !!examStore.localAnswers[question.value.id]?.isDoubtful
})

const onTextInput = (val) => {
  examStore.setAnswerText(val)
}

const getTypeLabel = (type) => {
  switch (type) {
    case 'SHORT_ANSWER':
      return 'Jawaban Singkat'
    case 'ESSAY':
      return 'Essay / Uraian'
    default:
      return 'Pilihan Ganda'
  }
}

const getTypeBadgeClass = (type) => {
  switch (type) {
    case 'SHORT_ANSWER':
      return 'bg-emerald-50 border-emerald-200 text-emerald-700'
    case 'ESSAY':
      return 'bg-purple-50 border-purple-200 text-purple-700'
    default:
      return 'bg-blue-50 border-blue-200 text-blue-700'
  }
}

const getWordCount = (text) => {
  if (!text || !text.trim()) return 0
  return text.trim().split(/\s+/).length
}
</script>
