<template>
  <div v-if="isOpen" class="fixed inset-0 z-50 flex items-center justify-center bg-slate-950/80 backdrop-blur-sm select-none p-2 sm:p-4 overflow-hidden">
    <!-- Simulator Outer Container -->
    <div
      :class="[
        'bg-slate-900 rounded-3xl shadow-2xl border border-slate-700 flex flex-col transition-all duration-300 overflow-hidden',
        viewMode === 'mobile' ? 'w-[420px] max-w-full h-[95vh] max-h-[860px]' : 'w-full max-w-6xl h-[92vh]'
      ]"
    >
      <!-- Simulator Control Bar (Top) -->
      <div class="px-4 py-3 bg-slate-850 border-b border-slate-700 flex items-center justify-between shrink-0 text-white">
        <div class="flex items-center space-x-2.5 truncate">
          <span class="w-2.5 h-2.5 rounded-full bg-emerald-400 animate-pulse"></span>
          <div class="truncate">
            <h3 class="text-xs sm:text-sm font-bold truncate">
              Simulator Tampilan Siswa
            </h3>
            <p class="text-[10px] text-slate-400 truncate">
              {{ bank?.title || 'Bank Soal' }} ({{ questions.length }} Butir Soal)
            </p>
          </div>
        </div>

        <!-- View Mode Switcher & Close -->
        <div class="flex items-center space-x-2 shrink-0">
          <!-- Switcher Buttons -->
          <div class="bg-slate-800 p-0.5 rounded-xl border border-slate-700 flex items-center space-x-0.5 text-xs font-semibold">
            <button
              type="button"
              @click="viewMode = 'mobile'"
              :class="[
                'px-2.5 py-1 rounded-lg transition flex items-center space-x-1.5 cursor-pointer',
                viewMode === 'mobile' ? 'bg-indigo-600 text-white shadow-xs' : 'text-slate-400 hover:text-white'
              ]"
              title="Mode Smartphone Mobile (BYOD)"
            >
              <svg class="w-3.5 h-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 18h.01M8 21h8a2 2 0 002-2V5a2 2 0 00-2-2H8a2 2 0 00-2 2v14a2 2 0 002 2z" />
              </svg>
              <span class="hidden sm:inline text-[11px]">Smartphone</span>
            </button>
            <button
              type="button"
              @click="viewMode = 'desktop'"
              :class="[
                'px-2.5 py-1 rounded-lg transition flex items-center space-x-1.5 cursor-pointer',
                viewMode === 'desktop' ? 'bg-indigo-600 text-white shadow-xs' : 'text-slate-400 hover:text-white'
              ]"
              title="Mode Layar Komputer Desktop"
            >
              <svg class="w-3.5 h-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9.75 17L9 20l-1 1h8l-1-1-.75-3M3 13h18M5 17h14a2 2 0 002-2V5a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z" />
              </svg>
              <span class="hidden sm:inline text-[11px]">Desktop</span>
            </button>
          </div>

          <!-- Reset Mock Answers -->
          <button
            type="button"
            @click="resetSimulatorState"
            class="p-1.5 rounded-xl bg-slate-800 text-slate-400 hover:text-amber-300 hover:bg-slate-700 transition cursor-pointer"
            title="Reset Jawaban Simulasi"
          >
            <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
            </svg>
          </button>

          <!-- Close Button -->
          <button
            type="button"
            @click="$emit('close')"
            class="p-1.5 rounded-xl bg-slate-800 text-slate-400 hover:text-white hover:bg-slate-700 transition cursor-pointer"
            title="Tutup Simulator"
          >
            <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
            </svg>
          </button>
        </div>
      </div>

      <!-- No Questions Warning -->
      <div v-if="questions.length === 0" class="flex-1 flex flex-col items-center justify-center p-8 text-center text-slate-400 bg-slate-900">
        <div class="w-12 h-12 rounded-2xl bg-slate-800 flex items-center justify-center text-2xl mb-3">📭</div>
        <h4 class="text-sm font-bold text-slate-200">Belum Ada Soal</h4>
        <p class="text-xs text-slate-400 mt-1 max-w-xs">Tambahkan atau impor soal ke bank soal ini terlebih dahulu untuk melihat pratinjau simulator siswa.</p>
      </div>

      <!-- ============================================================== -->
      <!-- VIEW 1: SMARTPHONE MOBILE SIMULATOR (3-ZONE MOBILE VIEWPORT) -->
      <!-- ============================================================== -->
      <div
        v-else-if="viewMode === 'mobile'"
        class="flex-1 bg-slate-950 p-2 sm:p-4 flex items-center justify-center overflow-hidden"
      >
        <!-- Phone Device Frame -->
        <div class="w-full max-w-[375px] h-full max-h-[780px] bg-white rounded-[38px] shadow-2xl border-[6px] border-slate-800 flex flex-col overflow-hidden relative">
          <!-- Smartphone Notch / Dynamic Bar -->
          <div class="h-6 bg-slate-900 text-white flex items-center justify-between px-6 text-[10px] font-mono shrink-0 select-none">
            <span>09:41</span>
            <div class="w-16 h-3 bg-black rounded-full"></div>
            <div class="flex items-center space-x-1">
              <span>📶</span>
              <span>🔋 100%</span>
            </div>
          </div>

          <!-- Zone 1: Fixed Topbar -->
          <header class="bg-indigo-600 text-white px-3 py-2 flex items-center justify-between shrink-0 shadow-sm">
            <div class="flex items-center space-x-2 overflow-hidden">
              <img src="/logo-smic.png" alt="Logo" class="w-6 h-6 object-contain shrink-0" />
              <div class="truncate">
                <div class="text-[11px] font-bold leading-tight truncate">CBT SIMULATOR</div>
                <div class="text-[9px] text-indigo-200 truncate">SMAS Islamic Centre</div>
              </div>
            </div>

            <!-- Timer Mockup -->
            <div class="flex items-center space-x-1 bg-indigo-700/80 px-2 py-1 rounded-lg font-mono text-xs font-bold text-amber-300 border border-indigo-500/50 tabular-timer">
              <svg class="w-3 h-3 text-amber-300" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
              </svg>
              <span>{{ timerString }}</span>
            </div>
          </header>

          <!-- Zone 2: Scrollable Main Question Area -->
          <main class="flex-1 overflow-y-auto p-3 space-y-3 bg-slate-50/60">
            <!-- Question Meta Header -->
            <div class="bg-white rounded-2xl p-3 border border-slate-200 shadow-2xs space-y-2">
              <div class="flex items-center justify-between flex-wrap gap-1 border-b border-slate-100 pb-2">
                <div class="flex items-center space-x-1.5">
                  <span class="px-2 py-0.5 bg-indigo-50 border border-indigo-200 text-indigo-700 font-bold text-[11px] rounded-lg">
                    Soal No. {{ currentIndex + 1 }}
                  </span>
                  <span class="px-2 py-0.5 rounded-lg text-[10px] font-bold border" :class="getQuestionTypeBadgeClass(currentQuestion.type)">
                    {{ getQuestionTypeLabel(currentQuestion.type) }}
                  </span>
                </div>

                <!-- Font Size Controls -->
                <div class="flex items-center space-x-0.5 bg-slate-100 rounded-lg p-0.5 text-[10px] font-bold text-slate-600">
                  <button @click="fontSize = 'text-xs'" :class="fontSize === 'text-xs' ? 'bg-indigo-600 text-white' : 'hover:bg-slate-200'" class="px-1.5 py-0.5 rounded transition">A-</button>
                  <button @click="fontSize = 'text-sm'" :class="fontSize === 'text-sm' ? 'bg-indigo-600 text-white' : 'hover:bg-slate-200'" class="px-1.5 py-0.5 rounded transition">A</button>
                  <button @click="fontSize = 'text-base'" :class="fontSize === 'text-base' ? 'bg-indigo-600 text-white' : 'hover:bg-slate-200'" class="px-1.5 py-0.5 rounded transition">A+</button>
                </div>
              </div>

              <!-- Doubtful status badge -->
              <div v-if="isCurrentDoubtful" class="inline-flex items-center space-x-1 px-2 py-0.5 bg-amber-100 border border-amber-300 text-amber-800 text-[10px] font-bold rounded-full">
                <span>★ Ragu-ragu</span>
              </div>

              <!-- Question Content HTML (KaTeX + Arabic + Korean + Image) -->
              <div :class="[fontSize, 'text-slate-900 leading-relaxed pt-1']">
                <RichContentRenderer :content="currentQuestion.content_html" />
              </div>
            </div>

            <!-- Choice Options / Inputs -->
            <!-- 1. MULTIPLE CHOICE -->
            <div v-if="!currentQuestion.type || currentQuestion.type === 'MULTIPLE_CHOICE'" class="space-y-2">
              <div
                v-for="opt in currentQuestionOptions"
                :key="opt.key"
                @click="selectOption(opt.key)"
                :class="[
                  'cursor-pointer rounded-2xl border-2 p-3 transition-all flex flex-col space-y-2 active:scale-[0.99]',
                  currentSelectedOption === opt.key
                    ? 'border-indigo-600 bg-indigo-50/80 shadow-xs'
                    : 'border-slate-200 bg-white hover:border-slate-300'
                ]"
              >
                <div class="flex items-start space-x-2.5">
                  <div
                    :class="[
                      'w-7 h-7 rounded-full flex-shrink-0 flex items-center justify-center font-bold text-xs transition-colors mt-0.5',
                      currentSelectedOption === opt.key ? 'bg-indigo-600 text-white shadow-xs' : 'bg-slate-100 text-slate-700 border border-slate-300'
                    ]"
                  >
                    {{ opt.key }}
                  </div>
                  <div class="flex-1 pt-0.5">
                    <RichContentRenderer :content="opt.text" :custom-class="fontSize" />
                  </div>
                </div>

                <!-- Option Image if present -->
                <div v-if="opt.image_url" class="pl-9">
                  <img :src="opt.image_url" alt="Opsi Jawaban" class="max-h-40 rounded-xl border border-slate-200 object-contain shadow-2xs" />
                </div>
              </div>
            </div>

            <!-- 2. SHORT ANSWER -->
            <div v-else-if="currentQuestion.type === 'SHORT_ANSWER'" class="bg-white p-3.5 rounded-2xl border border-slate-200 shadow-2xs space-y-2">
              <label class="block text-[11px] font-bold text-slate-700">Jawaban Singkat Siswa:</label>
              <input
                type="text"
                v-model="currentAnswerText"
                placeholder="Ketik jawaban singkat di sini..."
                :class="[fontSize, 'w-full px-3 py-2 rounded-xl border border-slate-300 focus:outline-none focus:ring-2 focus:ring-indigo-500 font-medium']"
              />
            </div>

            <!-- 3. ESSAY -->
            <div v-else-if="currentQuestion.type === 'ESSAY'" class="bg-white p-3.5 rounded-2xl border border-slate-200 shadow-2xs space-y-2">
              <label class="block text-[11px] font-bold text-slate-700">Uraian / Essay Siswa:</label>
              <textarea
                v-model="currentAnswerText"
                rows="5"
                placeholder="Ketik uraian jawaban lengkap di sini..."
                :class="[fontSize, 'w-full px-3 py-2 rounded-xl border border-slate-300 focus:outline-none focus:ring-2 focus:ring-indigo-500 font-medium']"
              ></textarea>
            </div>
          </main>

          <!-- Zone 3: Fixed Bottom Navigation Bar -->
          <footer class="bg-white border-t border-slate-200 px-3 py-2 flex items-center justify-between shrink-0 shadow-lg">
            <button
              type="button"
              @click="prevQuestion"
              :disabled="currentIndex === 0"
              class="px-2.5 py-1.5 bg-slate-100 hover:bg-slate-200 disabled:opacity-30 rounded-xl text-xs font-bold text-slate-700 transition flex items-center space-x-1 cursor-pointer"
            >
              <span>←</span>
              <span class="text-[11px]">Sebelum</span>
            </button>

            <!-- Doubtful Toggle Button -->
            <button
              type="button"
              @click="toggleDoubtful"
              :class="[
                'px-2.5 py-1.5 rounded-xl text-[11px] font-bold transition flex items-center space-x-1 cursor-pointer border',
                isCurrentDoubtful
                  ? 'bg-amber-400 text-amber-950 border-amber-500 shadow-xs'
                  : 'bg-amber-50 text-amber-800 border-amber-200 hover:bg-amber-100'
              ]"
            >
              <span>★</span>
              <span>Ragu</span>
            </button>

            <!-- Drawer Toggle Button -->
            <button
              type="button"
              @click="isMobileDrawerOpen = !isMobileDrawerOpen"
              class="px-2.5 py-1.5 bg-indigo-50 hover:bg-indigo-100 text-indigo-700 border border-indigo-200 rounded-xl text-[11px] font-bold transition cursor-pointer flex items-center space-x-1"
            >
              <span>📋</span>
              <span>Daftar Soal</span>
            </button>

            <button
              type="button"
              @click="nextQuestion"
              :disabled="currentIndex === questions.length - 1"
              class="px-2.5 py-1.5 bg-indigo-600 hover:bg-indigo-700 disabled:opacity-30 text-white rounded-xl text-xs font-bold transition flex items-center space-x-1 cursor-pointer"
            >
              <span class="text-[11px]">Berikut</span>
              <span>→</span>
            </button>
          </footer>

          <!-- Mobile Drawer Bottom Sheet -->
          <transition
            enter-active-class="transition duration-200 ease-out transform"
            enter-from-class="translate-y-full"
            enter-to-class="translate-y-0"
            leave-active-class="transition duration-150 ease-in transform"
            leave-from-class="translate-y-0"
            leave-to-class="translate-y-full"
          >
            <div
              v-if="isMobileDrawerOpen"
              class="absolute inset-x-0 bottom-0 bg-white rounded-t-3xl shadow-2xl border-t border-slate-200 max-h-[70%] flex flex-col z-20"
            >
              <div class="p-3 border-b border-slate-100 flex items-center justify-between">
                <div class="font-bold text-xs text-slate-800">Daftar Nomor Soal</div>
                <button @click="isMobileDrawerOpen = false" class="text-xs font-bold text-slate-500 hover:text-slate-800">✕ Tutup</button>
              </div>
              <div class="p-4 overflow-y-auto grid grid-cols-5 gap-2">
                <button
                  v-for="(q, idx) in questions"
                  :key="q.id || idx"
                  @click="jumpToQuestion(idx); isMobileDrawerOpen = false"
                  :class="[
                    'h-10 rounded-xl text-xs font-mono font-bold flex flex-col items-center justify-center transition-all border',
                    currentIndex === idx ? 'ring-2 ring-indigo-500 ring-offset-1' : '',
                    getQuestionStatusClass(idx)
                  ]"
                >
                  <span>{{ idx + 1 }}</span>
                  <span v-if="answers[getQuestionId(idx)]?.selectedOption" class="text-[9px] font-sans">
                    {{ answers[getQuestionId(idx)].selectedOption }}
                  </span>
                </button>
              </div>
            </div>
          </transition>
        </div>
      </div>

      <!-- ============================================================== -->
      <!-- VIEW 2: DESKTOP SIMULATOR (FULL SCREEN 2-COLUMN VIEWPORT) -->
      <!-- ============================================================== -->
      <div v-else class="flex-1 flex flex-col overflow-hidden bg-slate-100">
        <!-- Desktop Exam Topbar Mockup -->
        <header class="h-14 bg-indigo-700 text-white px-6 flex items-center justify-between shrink-0 shadow-sm">
          <div class="flex items-center space-x-3">
            <img src="/logo-smic.png" alt="Logo" class="w-8 h-8 object-contain" />
            <div>
              <div class="text-xs font-bold tracking-wide">SMAS ISLAMIC CENTRE DEMAK</div>
              <div class="text-[10px] text-indigo-200 font-semibold">{{ bank?.title || 'Simulasi Ujian Siswa' }}</div>
            </div>
          </div>

          <div class="flex items-center space-x-4">
            <!-- Simulated Student Name -->
            <div class="text-right hidden sm:block">
              <div class="text-xs font-bold">Ahmad Fulan (Simulasi Siswa)</div>
              <div class="text-[10px] text-indigo-200">NIS: 2026001 • Kelas: X-A</div>
            </div>

            <!-- Timer Countdown -->
            <div class="flex items-center space-x-2 bg-indigo-800/80 px-3 py-1.5 rounded-xl font-mono text-sm font-bold text-amber-300 border border-indigo-600 tabular-timer shadow-xs">
              <svg class="w-4 h-4 text-amber-300 animate-spin" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
              </svg>
              <span>{{ timerString }}</span>
            </div>
          </div>
        </header>

        <!-- Main Desktop Content (2 Columns) -->
        <div class="flex-1 flex overflow-hidden p-4 sm:p-6 gap-6">
          <!-- Left Column: Question Card & Answers (Scrollable) -->
          <div class="flex-1 bg-white rounded-3xl border border-slate-200 shadow-sm flex flex-col overflow-hidden">
            <!-- Header Bar -->
            <div class="px-6 py-3.5 bg-slate-50/80 border-b border-slate-100 flex items-center justify-between shrink-0">
              <div class="flex items-center space-x-2.5">
                <span class="px-3 py-1 bg-indigo-50 border border-indigo-200 text-indigo-700 font-bold text-xs rounded-xl">
                  Soal No. {{ currentIndex + 1 }}
                </span>
                <span class="px-2.5 py-1 rounded-xl text-xs font-bold border" :class="getQuestionTypeBadgeClass(currentQuestion.type)">
                  {{ getQuestionTypeLabel(currentQuestion.type) }}
                </span>
                <span class="text-xs text-slate-500 font-medium">
                  dari {{ questions.length }} Butir Soal
                </span>
                <span v-if="isCurrentDoubtful" class="px-2.5 py-0.5 bg-amber-100 border border-amber-300 text-amber-800 text-xs font-bold rounded-full">
                  ★ Ragu-ragu
                </span>
              </div>

              <!-- Font Size Selector -->
              <div class="flex items-center space-x-1 bg-white border border-slate-200 rounded-xl p-0.5 text-xs font-semibold text-slate-600 shadow-2xs">
                <button @click="fontSize = 'text-sm'" :class="fontSize === 'text-sm' ? 'bg-indigo-600 text-white' : 'hover:bg-slate-100'" class="px-2 py-1 rounded-lg transition">A-</button>
                <button @click="fontSize = 'text-base'" :class="fontSize === 'text-base' ? 'bg-indigo-600 text-white' : 'hover:bg-slate-100'" class="px-2 py-1 rounded-lg transition">A</button>
                <button @click="fontSize = 'text-lg'" :class="fontSize === 'text-lg' ? 'bg-indigo-600 text-white' : 'hover:bg-slate-100'" class="px-2 py-1 rounded-lg transition">A+</button>
              </div>
            </div>

            <!-- Question Scrollable Body -->
            <div class="flex-1 overflow-y-auto p-6 space-y-6">
              <!-- Question Text with KaTeX & Typography -->
              <div :class="[fontSize, 'text-slate-900 leading-relaxed font-sans']">
                <RichContentRenderer :content="currentQuestion.content_html" />
              </div>

              <!-- Options / Answer Input -->
              <!-- 1. MULTIPLE CHOICE -->
              <div v-if="!currentQuestion.type || currentQuestion.type === 'MULTIPLE_CHOICE'" class="space-y-3 pt-2">
                <div
                  v-for="opt in currentQuestionOptions"
                  :key="opt.key"
                  @click="selectOption(opt.key)"
                  :class="[
                    'cursor-pointer rounded-2xl border-2 p-4 transition-all flex flex-col space-y-3 active:scale-[0.99]',
                    currentSelectedOption === opt.key
                      ? 'border-indigo-600 bg-indigo-50/70 shadow-xs'
                      : 'border-slate-200 bg-white hover:border-slate-300 hover:bg-slate-50/50'
                  ]"
                >
                  <div class="flex items-start space-x-3.5">
                    <div
                      :class="[
                        'w-8 h-8 rounded-full flex-shrink-0 flex items-center justify-center font-bold text-sm transition-colors mt-0.5',
                        currentSelectedOption === opt.key ? 'bg-indigo-600 text-white shadow-sm' : 'bg-slate-100 text-slate-700 border border-slate-300'
                      ]"
                    >
                      {{ opt.key }}
                    </div>
                    <div class="flex-1 pt-0.5">
                      <RichContentRenderer :content="opt.text" :custom-class="fontSize" />
                    </div>
                  </div>

                  <!-- Image in option -->
                  <div v-if="opt.image_url" class="pl-11">
                    <img :src="opt.image_url" alt="Pilihan Jawaban" class="max-h-48 rounded-xl border border-slate-200 object-contain shadow-xs" />
                  </div>
                </div>
              </div>

              <!-- 2. SHORT ANSWER -->
              <div v-else-if="currentQuestion.type === 'SHORT_ANSWER'" class="p-5 bg-slate-50 rounded-2xl border border-slate-200 space-y-3">
                <label class="block text-xs font-bold text-slate-700">Jawaban Singkat Siswa:</label>
                <input
                  type="text"
                  v-model="currentAnswerText"
                  placeholder="Ketikkan jawaban singkat Anda di sini..."
                  :class="[fontSize, 'w-full px-4 py-3 rounded-xl border border-slate-300 focus:outline-none focus:ring-2 focus:ring-indigo-500 font-medium bg-white shadow-2xs']"
                />
              </div>

              <!-- 3. ESSAY -->
              <div v-else-if="currentQuestion.type === 'ESSAY'" class="p-5 bg-slate-50 rounded-2xl border border-slate-200 space-y-3">
                <label class="block text-xs font-bold text-slate-700">Uraian / Jawaban Lengkap Siswa:</label>
                <textarea
                  v-model="currentAnswerText"
                  rows="6"
                  placeholder="Tuliskan uraian atau penjelasan lengkap jawaban Anda di sini..."
                  :class="[fontSize, 'w-full p-4 rounded-xl border border-slate-300 focus:outline-none focus:ring-2 focus:ring-indigo-500 font-medium bg-white shadow-2xs leading-relaxed']"
                ></textarea>
              </div>
            </div>

            <!-- Bottom Navigation Bar -->
            <div class="px-6 py-4 bg-slate-50/80 border-t border-slate-100 flex items-center justify-between shrink-0">
              <button
                type="button"
                @click="prevQuestion"
                :disabled="currentIndex === 0"
                class="px-4 py-2.5 bg-slate-200 hover:bg-slate-300 disabled:opacity-40 rounded-xl text-xs font-bold text-slate-700 transition flex items-center space-x-1.5 cursor-pointer"
              >
                <span>←</span>
                <span>Soal Sebelumnya</span>
              </button>

              <button
                type="button"
                @click="toggleDoubtful"
                :class="[
                  'px-4 py-2.5 rounded-xl text-xs font-bold transition flex items-center space-x-1.5 cursor-pointer border',
                  isCurrentDoubtful
                    ? 'bg-amber-400 text-amber-950 border-amber-500 shadow-xs'
                    : 'bg-amber-50 text-amber-800 border-amber-200 hover:bg-amber-100'
                ]"
              >
                <span>★</span>
                <span>Tandai Ragu-ragu</span>
              </button>

              <button
                type="button"
                @click="nextQuestion"
                :disabled="currentIndex === questions.length - 1"
                class="px-5 py-2.5 bg-indigo-600 hover:bg-indigo-700 disabled:opacity-40 text-white rounded-xl text-xs font-bold transition flex items-center space-x-1.5 shadow-xs cursor-pointer"
              >
                <span>Soal Selanjutnya</span>
                <span>→</span>
              </button>
            </div>
          </div>

          <!-- Right Column: Number Grid Sidebar (1-N) -->
          <div class="w-80 bg-white rounded-3xl border border-slate-200 shadow-sm flex flex-col overflow-hidden shrink-0">
            <div class="p-4 border-b border-slate-100 bg-slate-50/70">
              <h4 class="font-bold text-xs text-slate-800">Navigasi Nomor Soal</h4>
              <p class="text-[10px] text-slate-500 mt-0.5">Klik nomor untuk melompat langsung ke soal</p>
            </div>

            <div class="flex-1 overflow-y-auto p-4 grid grid-cols-4 gap-2.5 content-start">
              <button
                v-for="(q, idx) in questions"
                :key="q.id || idx"
                @click="jumpToQuestion(idx)"
                :class="[
                  'h-12 rounded-2xl text-xs font-mono font-bold flex flex-col items-center justify-center transition-all border cursor-pointer',
                  currentIndex === idx ? 'ring-2 ring-indigo-600 ring-offset-2 scale-105' : 'hover:scale-102',
                  getQuestionStatusClass(idx)
                ]"
              >
                <span>{{ idx + 1 }}</span>
                <span v-if="answers[getQuestionId(idx)]?.selectedOption" class="text-[10px] font-sans">
                  {{ answers[getQuestionId(idx)].selectedOption }}
                </span>
              </button>
            </div>

            <!-- Legend Status Colors -->
            <div class="p-4 border-t border-slate-100 bg-slate-50/50 space-y-1.5 text-[11px]">
              <div class="flex items-center space-x-2">
                <span class="w-3.5 h-3.5 rounded-lg bg-emerald-500 text-white flex items-center justify-center text-[9px] font-bold">✓</span>
                <span class="text-slate-600">Sudah Dijawab (Hijau)</span>
              </div>
              <div class="flex items-center space-x-2">
                <span class="w-3.5 h-3.5 rounded-lg bg-amber-400 text-amber-950 flex items-center justify-center text-[9px] font-bold">★</span>
                <span class="text-slate-600">Ragu-ragu (Kuning)</span>
              </div>
              <div class="flex items-center space-x-2">
                <span class="w-3.5 h-3.5 rounded-lg bg-slate-100 border border-slate-300"></span>
                <span class="text-slate-600">Belum Dijawab (Abu-abu)</span>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, watch } from 'vue'
import RichContentRenderer from '../common/RichContentRenderer.vue'

const props = defineProps({
  isOpen: {
    type: Boolean,
    default: false
  },
  bank: {
    type: Object,
    default: () => ({})
  },
  questions: {
    type: Array,
    default: () => []
  }
})

defineEmits(['close'])

const viewMode = ref('mobile') // 'mobile' | 'desktop'
const currentIndex = ref(0)
const fontSize = ref('text-base')
const isMobileDrawerOpen = ref(false)
const timerString = ref('01:29:45')

// Local reactive mock answers: { [questionId]: { selectedOption: 'A', answerText: '', isDoubtful: false } }
const answers = ref({})

const resetSimulatorState = () => {
  currentIndex.value = 0
  answers.value = {}
  isMobileDrawerOpen.value = false
}

watch(() => props.isOpen, (newVal) => {
  if (newVal) {
    resetSimulatorState()
  }
})

const getQuestionId = (idx) => {
  const q = props.questions[idx]
  return q ? (q.id || `q_${idx}`) : `q_${idx}`
}

const currentQuestion = computed(() => {
  return props.questions[currentIndex.value] || {}
})

const currentQuestionId = computed(() => {
  return getQuestionId(currentIndex.value)
})

const currentQuestionOptions = computed(() => {
  const q = currentQuestion.value
  if (!q.options) return []
  if (typeof q.options === 'string') {
    try {
      return JSON.parse(q.options)
    } catch {
      return []
    }
  }
  return q.options
})

const currentSelectedOption = computed(() => {
  return answers.value[currentQuestionId.value]?.selectedOption || ''
})

const currentAnswerText = computed({
  get: () => answers.value[currentQuestionId.value]?.answerText || '',
  set: (val) => {
    const qId = currentQuestionId.value
    if (!answers.value[qId]) {
      answers.value[qId] = { selectedOption: '', answerText: val, isDoubtful: false }
    } else {
      answers.value[qId].answerText = val
    }
  }
})

const isCurrentDoubtful = computed(() => {
  return !!answers.value[currentQuestionId.value]?.isDoubtful
})

const selectOption = (key) => {
  const qId = currentQuestionId.value
  if (!answers.value[qId]) {
    answers.value[qId] = { selectedOption: key, answerText: '', isDoubtful: false }
  } else {
    answers.value[qId].selectedOption = key
  }
}

const toggleDoubtful = () => {
  const qId = currentQuestionId.value
  if (!answers.value[qId]) {
    answers.value[qId] = { selectedOption: '', answerText: '', isDoubtful: true }
  } else {
    answers.value[qId].isDoubtful = !answers.value[qId].isDoubtful
  }
}

const prevQuestion = () => {
  if (currentIndex.value > 0) {
    currentIndex.value--
  }
}

const nextQuestion = () => {
  if (currentIndex.value < props.questions.length - 1) {
    currentIndex.value++
  }
}

const jumpToQuestion = (idx) => {
  if (idx >= 0 && idx < props.questions.length) {
    currentIndex.value = idx
  }
}

const getQuestionStatusClass = (idx) => {
  const qId = getQuestionId(idx)
  const a = answers.value[qId]
  if (a?.isDoubtful) {
    return 'bg-amber-400 text-amber-950 border-amber-500 shadow-2xs font-bold'
  }
  if (a?.selectedOption || (a?.answerText && a.answerText.trim().length > 0)) {
    return 'bg-emerald-500 text-white border-emerald-600 shadow-2xs font-bold'
  }
  return 'bg-slate-100 text-slate-700 border-slate-300 hover:bg-slate-200'
}

const getQuestionTypeLabel = (type) => {
  switch (type) {
    case 'SHORT_ANSWER':
      return 'Jawaban Singkat'
    case 'ESSAY':
      return 'Essay / Uraian'
    default:
      return 'Pilihan Ganda'
  }
}

const getQuestionTypeBadgeClass = (type) => {
  switch (type) {
    case 'SHORT_ANSWER':
      return 'bg-emerald-50 border-emerald-200 text-emerald-700'
    case 'ESSAY':
      return 'bg-purple-50 border-purple-200 text-purple-700'
    default:
      return 'bg-blue-50 border-blue-200 text-blue-700'
  }
}
</script>
