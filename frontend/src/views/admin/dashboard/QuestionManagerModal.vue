<template>
  <!-- MODAL: KELOLA & LIHAT BUTIR SOAL -->
  <div v-if="showBankQuestionsModal && viewingBank" class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-slate-900/60 backdrop-blur-xs select-none">
    <div class="bg-white rounded-3xl max-w-4xl w-full max-h-[90vh] flex flex-col shadow-2xl overflow-hidden">
      <!-- Modal Header -->
      <div class="px-6 py-4 border-b border-slate-100 flex items-center justify-between shrink-0 bg-slate-50/70">
        <div class="min-w-0 flex-1 pr-3">
          <h3 class="text-base font-bold text-slate-900 truncate" :title="viewingBank.title">{{ viewingBank.title }}</h3>
          <div class="flex items-center gap-x-2 gap-y-1 flex-wrap mt-1.5 text-xs text-slate-500">
            <span :class="['px-2.5 py-0.5 rounded-full text-[10px] font-bold inline-block shrink-0', viewingBank.is_locked ? 'bg-emerald-100 text-emerald-800' : 'bg-amber-100 text-amber-800']">
              {{ viewingBank.is_locked ? 'Terkunci' : 'Draft' }}
            </span>
            <span class="text-slate-300 hidden sm:inline">|</span>
            <span>
              Mata Pelajaran: <span class="font-semibold text-slate-700">{{ viewingBank.subject?.name || '-' }}</span>
              <span class="text-slate-400 font-mono text-[10px] ml-1">({{ viewingBank.subject?.code || '-' }})</span>
            </span>
            <span class="text-slate-300 hidden sm:inline">|</span>
            <span>Penyusun: <span class="font-semibold text-slate-700">{{ viewingBank.created_by?.full_name || 'Guru' }}</span></span>
          </div>
        </div>
        <div class="flex items-center gap-2 shrink-0">
          <button
            v-if="editingQuestionId === null"
            @click="startAddQuestion"
            class="px-3 py-1.5 bg-indigo-600 hover:bg-indigo-700 text-white font-bold text-xs rounded-xl shadow-xs transition flex items-center gap-1 cursor-pointer"
          >
            <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
            </svg>
            <span>Tambah Soal</span>
          </button>
          <button
            @click="showBankQuestionsModal = false; editingQuestionId = null"
            class="p-2 text-slate-400 hover:text-slate-600 hover:bg-slate-200/60 rounded-xl transition cursor-pointer"
          >
            ✕
          </button>
        </div>
      </div>

      <!-- Modal Body (Scrollable) -->
      <div class="p-6 overflow-y-auto flex-1 space-y-4 bg-slate-50/40">
        <!-- Loading Spinner -->
        <div v-if="isLoadingBankQuestions" class="py-16 text-center text-slate-400 space-y-2">
          <div class="w-8 h-8 border-4 border-indigo-600 border-t-transparent rounded-full animate-spin mx-auto"></div>
          <p class="text-xs font-semibold text-slate-500">Memuat butir naskah soal...</p>
        </div>

        <!-- Question Form (Add / Edit) -->
        <div v-else-if="editingQuestionId !== null" class="bg-white p-5 rounded-3xl border border-indigo-100 shadow-sm space-y-4">
          <div class="flex items-center justify-between border-b border-slate-100 pb-3">
            <h4 class="font-bold text-sm text-slate-900">
              {{ editingQuestionId === 'new' ? '➕ Tambah Butir Soal Baru' : '✏️ Edit Butir Soal #' + questionForm.question_number }}
            </h4>
            <span class="text-xs font-semibold px-2.5 py-1 rounded-lg border" :class="getQuestionTypeBadgeClass(questionForm.type)">
              {{ getQuestionTypeLabel(questionForm.type) }}
            </span>
          </div>

          <form @submit.prevent="saveQuestion" class="space-y-4 text-xs">
            <div class="grid grid-cols-1 sm:grid-cols-3 gap-3">
              <div>
                <label class="block font-bold text-slate-700 mb-1">Nomor Urut Soal:</label>
                <input
                  v-model.number="questionForm.question_number"
                  type="number"
                  min="1"
                  required
                  class="w-full px-3 py-2 rounded-xl border border-slate-300 font-mono focus:outline-none focus:ring-2 focus:ring-indigo-500 bg-slate-50/50"
                />
              </div>
              <div>
                <label class="block font-bold text-slate-700 mb-1">Tipe / Model Soal:</label>
                <select
                  v-model="questionForm.type"
                  required
                  class="w-full px-3 py-2 rounded-xl border border-slate-300 font-bold text-indigo-700 focus:outline-none focus:ring-2 focus:ring-indigo-500 bg-indigo-50/40 cursor-pointer"
                >
                  <option value="MULTIPLE_CHOICE">Pilihan Ganda (PG)</option>
                  <option value="SHORT_ANSWER">Jawaban Singkat / Isian</option>
                  <option value="ESSAY">Essay / Uraian</option>
                </select>
              </div>
              <div>
                <label class="block font-bold text-slate-700 mb-1">Bobot Nilai (Score Weight):</label>
                <input
                  v-model.number="questionForm.score_weight"
                  type="number"
                  step="0.1"
                  min="0.1"
                  required
                  class="w-full px-3 py-2 rounded-xl border border-slate-300 font-mono focus:outline-none focus:ring-2 focus:ring-indigo-500 bg-slate-50/50"
                />
              </div>
            </div>

            <!-- Question Content with Toolbar & Paste/Drop Listener -->
            <div class="space-y-2">
              <div class="flex items-center justify-between flex-wrap gap-2">
                <label class="block font-bold text-slate-700">Konten Pertanyaan (HTML, KaTeX Math, Arab, Korea):</label>

                <!-- Toolbar Helper Buttons -->
                <div class="flex items-center flex-wrap gap-1.5">
                  <!-- Image Upload Trigger Button -->
                  <label class="px-2.5 py-1 bg-indigo-50 hover:bg-indigo-100 text-indigo-700 border border-indigo-200 rounded-lg text-[11px] font-bold cursor-pointer transition flex items-center gap-1">
                    <svg v-if="isUploadingQuestionImage" class="animate-spin w-3 h-3 text-indigo-600" fill="none" viewBox="0 0 24 24">
                      <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
                      <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8v8H4z"></path>
                    </svg>
                    <span v-else>📷</span>
                    <span>{{ isUploadingQuestionImage ? 'Mengunggah...' : 'Unggah Gambar' }}</span>
                    <input
                      type="file"
                      accept="image/*"
                      @change="handleQuestionImageUpload"
                      class="hidden"
                      :disabled="isUploadingQuestionImage"
                    />
                  </label>

                  <!-- Math Snippet Shortcuts -->
                  <button
                    type="button"
                    @click="insertMathSnippet('$\\frac{a}{b}$')"
                    class="px-2 py-1 bg-slate-100 hover:bg-slate-200 text-slate-700 rounded-lg text-[11px] font-mono font-bold transition"
                    title="Sisipkan Rumus Pecahan LaTeX"
                  >
                    \frac{a}{b}
                  </button>
                  <button
                    type="button"
                    @click="insertMathSnippet('$\\sqrt{x}$')"
                    class="px-2 py-1 bg-slate-100 hover:bg-slate-200 text-slate-700 rounded-lg text-[11px] font-mono font-bold transition"
                    title="Sisipkan Rumus Akar LaTeX"
                  >
                    \sqrt{x}
                  </button>
                  <button
                    type="button"
                    @click="insertMathSnippet('$$\\int_{0}^{1} f(x) dx$$')"
                    class="px-2 py-1 bg-slate-100 hover:bg-slate-200 text-slate-700 rounded-lg text-[11px] font-mono font-bold transition"
                    title="Sisipkan Rumus Integral Display"
                  >
                    \int
                  </button>
                </div>
              </div>

              <div class="relative">
                <textarea
                  v-model="questionForm.content_html"
                  @paste="handleQuestionPaste"
                  @dragover.prevent
                  @drop.prevent="handleQuestionDrop"
                  rows="4"
                  required
                  placeholder="Tuliskan pertanyaan soal di sini... (Dukungan: Ctrl+V tempel screenshot gambar, Drag-and-drop file gambar, rumus KaTeX $x^2$, teks Arab & Hangul Korea)"
                  class="w-full px-3 py-2.5 rounded-2xl border border-slate-300 focus:outline-none focus:ring-2 focus:ring-indigo-500 font-sans"
                ></textarea>
              </div>
              <div class="flex items-center justify-between text-[11px] text-slate-500">
                <span>💡 <em>Tips:</em> Anda dapat langsung menekan <strong>Ctrl+V</strong> untuk menempel screenshot gambar di kolom ini.</span>
              </div>

              <!-- Live Preview of Question Content -->
              <div v-if="questionForm.content_html && questionForm.content_html.trim()" class="p-3 bg-slate-50 rounded-2xl border border-slate-200 space-y-1">
                <div class="text-[10px] font-bold uppercase tracking-wider text-slate-400">Pratinjau Tampilan Konten:</div>
                <div class="text-xs text-slate-900 bg-white p-3 rounded-xl border border-slate-100">
                  <RichContentRenderer :content="questionForm.content_html" />
                </div>
              </div>
            </div>

            <!-- TYPE 1: MULTIPLE CHOICE (Options A - E) -->
            <div v-if="questionForm.type === 'MULTIPLE_CHOICE'" class="space-y-2.5">
              <div class="flex items-center justify-between">
                <label class="block font-bold text-slate-700">Pilihan Jawaban (A - E):</label>
                <span class="text-[11px] text-slate-500">Pilih radio button di kiri sebagai kunci jawaban</span>
              </div>
              <div
                v-for="(opt, oIdx) in questionForm.options"
                :key="opt.key"
                class="flex flex-col space-y-2 p-2.5 rounded-2xl border transition"
                :class="questionForm.correct_key === opt.key ? 'border-emerald-300 bg-emerald-50/40' : 'border-slate-200 bg-white'"
              >
                <div class="flex items-center gap-2.5">
                  <label class="flex items-center gap-1.5 cursor-pointer shrink-0">
                    <input
                      type="radio"
                      :value="opt.key"
                      v-model="questionForm.correct_key"
                      name="correct_answer"
                      class="text-emerald-600 focus:ring-emerald-500"
                    />
                    <span
                      class="w-6 h-6 rounded-lg flex items-center justify-center font-bold text-xs"
                      :class="questionForm.correct_key === opt.key ? 'bg-emerald-600 text-white' : 'bg-slate-100 text-slate-700'"
                    >
                      {{ opt.key }}
                    </span>
                  </label>
                  <input
                    v-model="opt.text"
                    type="text"
                    :placeholder="'Teks jawaban pilihan ' + opt.key"
                    class="flex-1 px-3 py-1.5 rounded-xl border border-slate-200 text-xs focus:outline-none focus:ring-2 focus:ring-indigo-500 bg-white"
                  />

                  <!-- Upload Image for this Choice Button -->
                  <label class="px-2 py-1 bg-slate-100 hover:bg-slate-200 text-slate-600 border border-slate-200 rounded-lg text-[11px] font-bold cursor-pointer transition shrink-0 flex items-center gap-1" :title="'Unggah gambar untuk opsi ' + opt.key">
                    <svg v-if="uploadingOptionKey === opt.key" class="animate-spin w-3 h-3 text-indigo-600" fill="none" viewBox="0 0 24 24">
                      <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
                      <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8v8H4z"></path>
                    </svg>
                    <span v-else>🖼️</span>
                    <span>{{ uploadingOptionKey === opt.key ? '...' : 'Gambar' }}</span>
                    <input
                      type="file"
                      accept="image/*"
                      @change="handleOptionImageUpload($event, opt.key)"
                      class="hidden"
                      :disabled="uploadingOptionKey === opt.key"
                    />
                  </label>
                </div>

                <!-- Choice Option Image Preview Thumbnail if attached -->
                <div v-if="opt.image_url" class="pl-8 flex items-center gap-2">
                  <div class="relative group inline-block">
                    <img :src="opt.image_url" alt="Opsi" class="h-16 rounded-xl border border-slate-200 object-contain bg-white p-1" />
                    <button
                      type="button"
                      @click="removeOptionImage(opt.key)"
                      class="absolute -top-1.5 -right-1.5 w-5 h-5 rounded-full bg-rose-600 text-white flex items-center justify-center text-[10px] font-bold shadow-xs hover:bg-rose-700 transition"
                      title="Hapus Gambar Opsi"
                    >
                      ✕
                    </button>
                  </div>
                  <span class="text-[11px] text-slate-500 font-mono truncate max-w-xs">{{ opt.image_url }}</span>
                </div>
              </div>
            </div>

            <!-- TYPE 2: SHORT ANSWER -->
            <div v-else-if="questionForm.type === 'SHORT_ANSWER'" class="space-y-2 bg-emerald-50/40 p-4 rounded-2xl border border-emerald-200">
              <label class="block font-bold text-emerald-900">Kunci Jawaban Singkat & Alternatif:</label>
              <input
                v-model="questionForm.correct_key"
                type="text"
                required
                placeholder="Contoh: Nusantara|IKN|Ibu Kota Nusantara"
                class="w-full px-3 py-2 rounded-xl border border-emerald-300 text-xs focus:outline-none focus:ring-2 focus:ring-emerald-500 bg-white font-medium text-slate-900"
              />
              <p class="text-[11px] text-emerald-700 leading-relaxed">
                💡 <strong>Tips:</strong> Gunakan tanda pipa <code>|</code> untuk memisahkan variasi jawaban yang diperbolehkan. Sistem akan memeriksa jawaban siswa secara toleran (tidak sensitif huruf besar/kecil).
              </p>
            </div>

            <!-- TYPE 3: ESSAY -->
            <div v-else-if="questionForm.type === 'ESSAY'" class="space-y-2 bg-purple-50/40 p-4 rounded-2xl border border-purple-200">
              <label class="block font-bold text-purple-900">Pedoman Penilaian / Rubrik Guru (Opsional):</label>
              <textarea
                v-model="questionForm.rubric_guide"
                rows="3"
                placeholder="Tuliskan kata kunci, poin esensial, atau pedoman penilaian untuk memudahkan guru saat memeriksa jawaban siswa..."
                class="w-full px-3 py-2 rounded-xl border border-purple-300 text-xs focus:outline-none focus:ring-2 focus:ring-purple-500 bg-white font-sans text-slate-800"
              ></textarea>
              <p class="text-[11px] text-purple-700">
                ℹ️ Soal Essay akan dikerjakan siswa dalam kolom teks bebas dan dinilai secara manual oleh guru melalui rekap nilai.
              </p>
            </div>

            <div class="pt-2 flex justify-end gap-2 border-t border-slate-100">
              <button
                type="button"
                @click="cancelEditQuestion"
                class="px-4 py-2 bg-slate-100 hover:bg-slate-200 rounded-xl text-slate-700 font-semibold cursor-pointer"
                :disabled="isSavingQuestion"
              >
                Batal
              </button>
              <button
                type="submit"
                class="px-5 py-2 bg-indigo-600 hover:bg-indigo-700 text-white rounded-xl font-bold shadow-xs cursor-pointer flex items-center gap-1"
                :disabled="isSavingQuestion"
              >
                <span>{{ isSavingQuestion ? 'Menyimpan...' : 'Simpan Butir Soal' }}</span>
              </button>
            </div>
          </form>
        </div>

        <!-- Question List -->
        <div v-else class="space-y-3">
          <div v-if="bankQuestions.length === 0" class="py-12 text-center text-slate-400 bg-white rounded-3xl border border-slate-200 p-6 space-y-3">
            <div class="w-12 h-12 rounded-2xl bg-slate-100 text-slate-400 flex items-center justify-center text-xl mx-auto">
              📭
            </div>
            <div class="space-y-1">
              <div class="font-bold text-slate-700 text-sm">Belum Ada Butir Soal</div>
              <p class="text-xs text-slate-500 max-w-sm mx-auto">
                Bank soal ini masih kosong. Silakan tambahkan butir soal secara manual atau unggah menggunakan berkas spreadsheet Excel.
              </p>
            </div>
            <div class="flex items-center justify-center gap-2 pt-2">
              <button
                v-if="!viewingBank.is_locked"
                @click="startAddQuestion"
                class="px-4 py-2 bg-indigo-600 hover:bg-indigo-700 text-white font-bold text-xs rounded-xl shadow-xs transition cursor-pointer"
              >
                + Tambah Soal Manual
              </button>
              <button
                @click="showBankQuestionsModal = false; openUploadBankModal(viewingBank)"
                class="px-4 py-2 bg-emerald-600 hover:bg-emerald-700 text-white font-bold text-xs rounded-xl shadow-xs transition cursor-pointer"
              >
                📤 Unggah Excel
              </button>
            </div>
          </div>

          <!-- Reorder Info Bar -->
          <div v-if="bankQuestions.length > 1" class="flex items-center justify-between gap-3 px-4 py-2.5 bg-indigo-50/70 border border-indigo-100/80 rounded-2xl text-xs text-indigo-950">
            <div class="flex items-center gap-2">
              <span class="text-sm">↕️</span>
              <span class="font-medium text-[11px] sm:text-xs">Geser (drag & drop) kartu soal untuk mengatur ulang nomor urut soal secara interaktif.</span>
            </div>
            <div v-if="isReorderingQuestions || reorderStatusMessage" class="flex items-center gap-1.5 shrink-0 font-bold text-[10px] sm:text-[11px] px-2.5 py-1 bg-white rounded-xl shadow-2xs border border-indigo-100">
              <div v-if="isReorderingQuestions" class="w-3 h-3 border-2 border-indigo-600 border-t-transparent rounded-full animate-spin"></div>
              <span :class="reorderStatusMessage.includes('Gagal') ? 'text-rose-600' : 'text-indigo-700'">{{ reorderStatusMessage }}</span>
            </div>
          </div>

          <!-- List of Question Cards with Drag & Drop -->
          <transition-group name="card-list" tag="div" class="space-y-3">
            <div
              v-for="(q, qIdx) in bankQuestions"
              :key="q.id"
              draggable="true"
              @dragstart="onQuestionDragStart($event, qIdx)"
              @dragover="onQuestionDragOver($event, qIdx)"
              @dragleave="onQuestionDragLeave($event, qIdx)"
              @drop="onQuestionDrop($event, qIdx)"
              @dragend="onQuestionDragEnd"
              class="bg-white p-5 rounded-3xl border transition-all duration-200 space-y-3 relative group"
              :class="[
                draggedQuestionIdx === qIdx ? 'opacity-40 scale-[0.98] border-dashed border-indigo-400 bg-indigo-50/30 shadow-inner' : 'hover:border-slate-300 shadow-xs',
                dragOverQuestionIdx === qIdx && draggedQuestionIdx !== qIdx ? 'border-indigo-500 ring-2 ring-indigo-200 bg-indigo-50/20' : 'border-slate-200'
              ]"
            >
              <!-- Card Header -->
              <div class="flex items-center justify-between border-b border-slate-100 pb-2.5">
                <div class="flex items-center gap-2 flex-wrap">
                  <!-- Drag Grip Handle -->
                  <div
                    class="p-1 -ml-1 text-slate-400 hover:text-indigo-600 hover:bg-indigo-50 rounded-lg cursor-grab active:cursor-grabbing transition"
                    title="Klik & tahan untuk menggeser posisi soal"
                  >
                    <svg class="w-4 h-4" viewBox="0 0 24 24" fill="currentColor">
                      <circle cx="8" cy="6" r="1.5" />
                      <circle cx="16" cy="6" r="1.5" />
                      <circle cx="8" cy="12" r="1.5" />
                      <circle cx="16" cy="12" r="1.5" />
                      <circle cx="8" cy="18" r="1.5" />
                      <circle cx="16" cy="18" r="1.5" />
                    </svg>
                  </div>

                  <!-- Reactive Dynamic Number Badge -->
                  <span class="px-2.5 py-1 rounded-xl bg-indigo-50 text-indigo-700 font-mono font-bold text-xs border border-indigo-100 shadow-2xs">
                    #{{ qIdx + 1 }}
                  </span>
                  <!-- Type Badge -->
                  <span
                    class="px-2.5 py-0.5 rounded-full text-[10px] font-bold border"
                    :class="getQuestionTypeBadgeClass(q.type)"
                  >
                    {{ getQuestionTypeLabel(q.type) }}
                  </span>
                  <span class="px-2 py-0.5 rounded-full text-[10px] font-semibold bg-slate-100 text-slate-600">
                    Bobot: {{ q.score_weight || 1.0 }}
                  </span>
                </div>
                <div class="flex items-center gap-1.5">
                  <button
                    @click="startEditQuestion(q)"
                    class="p-1.5 bg-slate-100 hover:bg-slate-200 text-slate-600 rounded-full transition cursor-pointer flex items-center justify-center"
                    title="Edit Soal & Jawaban"
                  >
                    <svg class="w-3.5 h-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z" />
                    </svg>
                  </button>
                  <button
                    @click="deleteQuestionItem(q)"
                    class="p-1.5 bg-rose-50 hover:bg-rose-100 text-rose-600 rounded-full transition cursor-pointer flex items-center justify-center"
                    title="Hapus Soal"
                  >
                    <svg class="w-3.5 h-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
                    </svg>
                  </button>
                </div>
              </div>

              <!-- Question Content with KaTeX & Typography -->
              <div class="text-xs text-slate-800 leading-relaxed font-sans select-text">
                <RichContentRenderer :content="q.content_html" />
              </div>

              <!-- 1. MULTIPLE CHOICE OPTIONS GRID -->
              <div v-if="!q.type || q.type === 'MULTIPLE_CHOICE'" class="grid grid-cols-1 sm:grid-cols-2 gap-2 pt-1">
                <div
                  v-for="opt in (q.options || [])"
                  :key="opt.key"
                  class="flex flex-col space-y-1.5 p-2 rounded-xl text-xs transition"
                  :class="q.correct_key === opt.key ? 'bg-emerald-50 border border-emerald-300 text-emerald-900 font-semibold' : 'bg-slate-50 text-slate-700 border border-slate-100'"
                >
                  <div class="flex items-start gap-2">
                    <span
                      class="w-5 h-5 rounded-md flex items-center justify-center text-[10px] font-bold shrink-0 mt-0.5"
                      :class="q.correct_key === opt.key ? 'bg-emerald-600 text-white' : 'bg-slate-200 text-slate-600'"
                    >
                      {{ opt.key }}
                    </span>
                    <div class="flex-1 break-words select-text">
                      <RichContentRenderer :content="opt.text" />
                    </div>
                    <span v-if="q.correct_key === opt.key" class="text-[10px] font-bold text-emerald-700 shrink-0">✓ Kunci</span>
                  </div>

                  <!-- Option Image if present -->
                  <div v-if="opt.image_url" class="pl-7">
                    <img :src="opt.image_url" alt="Opsi" class="max-h-24 rounded-lg border border-slate-200 object-contain shadow-2xs" />
                  </div>
                </div>
              </div>

              <!-- 2. SHORT ANSWER KEY DISPLAY -->
              <div v-else-if="q.type === 'SHORT_ANSWER'" class="p-3 bg-emerald-50/60 rounded-2xl border border-emerald-200 text-xs space-y-1">
                <div class="font-bold text-emerald-900 flex items-center gap-1.5">
                  <span class="w-2 h-2 rounded-full bg-emerald-500"></span>
                  <span>Kunci Jawaban Singkat:</span>
                </div>
                <div class="font-mono text-emerald-800 bg-white/80 p-2 rounded-xl border border-emerald-100 select-text">
                  {{ q.correct_key || '-' }}
                </div>
              </div>

              <!-- 3. ESSAY RUBRIC DISPLAY -->
              <div v-else-if="q.type === 'ESSAY'" class="p-3 bg-purple-50/60 rounded-2xl border border-purple-200 text-xs space-y-1">
                <div class="font-bold text-purple-900 flex items-center gap-1.5">
                  <span class="w-2 h-2 rounded-full bg-purple-500"></span>
                  <span>Pedoman Rubrik Penilaian:</span>
                </div>
                <p class="text-purple-800 leading-relaxed bg-white/80 p-2 rounded-xl border border-purple-100 font-sans select-text">
                  {{ q.rubric_guide || '(Tidak ada pedoman rubrik khusus)' }}
                </p>
              </div>
            </div>
          </transition-group>
        </div>
      </div>

      <!-- Modal Footer -->
      <div class="px-6 py-3 border-t border-slate-100 bg-slate-50/70 flex items-center justify-between shrink-0">
        <span class="text-xs text-slate-500 font-medium">
          Total: <strong class="text-slate-800">{{ bankQuestions.length }}</strong> butir soal
        </span>
        <button
          type="button"
          @click="showBankQuestionsModal = false; editingQuestionId = null"
          class="px-4 py-2 bg-slate-200 hover:bg-slate-300 rounded-xl text-slate-700 font-bold text-xs transition cursor-pointer"
        >
          Tutup
        </button>
      </div>
    </div>
  </div>
</template>

<script setup>
import RichContentRenderer from '../../../components/common/RichContentRenderer.vue'
import { useDashboard } from './context'

const {
  bankQuestions,
  cancelEditQuestion,
  deleteQuestionItem,
  dragOverQuestionIdx,
  draggedQuestionIdx,
  editingQuestionId,
  getQuestionTypeBadgeClass,
  getQuestionTypeLabel,
  handleOptionImageUpload,
  handleQuestionDrop,
  handleQuestionImageUpload,
  handleQuestionPaste,
  insertMathSnippet,
  isLoadingBankQuestions,
  isReorderingQuestions,
  isSavingQuestion,
  isUploadingQuestionImage,
  onQuestionDragEnd,
  onQuestionDragLeave,
  onQuestionDragOver,
  onQuestionDragStart,
  onQuestionDrop,
  openUploadBankModal,
  questionForm,
  removeOptionImage,
  reorderStatusMessage,
  saveQuestion,
  showBankQuestionsModal,
  startAddQuestion,
  startEditQuestion,
  uploadingOptionKey,
  viewingBank,
} = useDashboard()
</script>

<style scoped>
.card-list-move {
  transition: transform 0.25s cubic-bezier(0.2, 0, 0, 1);
}
</style>
