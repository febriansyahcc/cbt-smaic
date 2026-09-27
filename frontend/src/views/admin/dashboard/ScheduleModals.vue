<template>
  <!-- MODAL: BUAT / EDIT JADWAL UJIAN -->
  <div v-if="showScheduleModal" class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-slate-900/60 backdrop-blur-xs">
    <div class="bg-white rounded-3xl max-w-lg w-full shadow-2xl overflow-hidden flex flex-col max-h-[90vh]">
      <!-- Header: Title only -->
      <div class="flex items-center justify-between px-6 py-4 border-b border-slate-100 shrink-0 bg-white">
        <h3 class="text-base font-bold text-slate-900">
          {{ isEditSchedule ? 'Edit Data Jadwal Ujian' : 'Buat Jadwal Ujian Baru' }}
        </h3>
        <button @click="showScheduleModal = false" class="p-1.5 rounded-xl text-slate-400 hover:text-slate-600 hover:bg-slate-100 transition cursor-pointer">
          ✕
        </button>
      </div>

      <form @submit.prevent="submitScheduleForm" class="flex flex-col flex-1 overflow-hidden min-h-0">
        <!-- Scrollable Body Content -->
        <div class="p-6 space-y-4 text-xs overflow-y-auto flex-1">
          <!-- 1. Alokasi Kelas & Mata Pelajaran (Paling Atas) -->
          <div>
            <div class="flex items-center justify-between mb-1">
              <label class="block font-bold text-slate-700">Pilih Kelas & Mata Pelajaran:</label>
              <span
                v-if="classSubjects.length > 0 && availableClassSubjects.length > 0"
                class="text-[10px] text-indigo-600 font-bold bg-indigo-50 px-2 py-0.5 rounded-full"
              >
                Tersedia: {{ availableClassSubjects.length }} belum dijadwalkan
              </span>
            </div>

            <!-- JIKA ADA ALOKASI YANG BELUM DIJADWALKAN -->
            <select
              v-if="availableClassSubjects.length > 0"
              v-model="selectedScheduleClassSubjectId"
              @change="onClassSubjectChange"
              required
              class="w-full px-3.5 py-2.5 pr-8 rounded-xl border border-slate-300 font-medium focus:ring-2 focus:ring-indigo-500 text-xs bg-white text-slate-800 truncate"
              style="text-overflow: ellipsis; overflow: hidden; white-space: nowrap;"
            >
              <option value="" disabled>-- Pilih Rombel & Mata Pelajaran --</option>
              <optgroup
                v-for="grade in availableClassSubjectGrades"
                :key="grade"
                :label="'📁 Tingkat ' + grade"
              >
                <option
                  v-for="cs in groupedAvailableClassSubjects[grade]"
                  :key="cs.id"
                  :value="cs.id"
                >
                  {{ cs.class_room?.name }} • {{ cs.subject?.name }} — {{ cs.teacher?.full_name || 'Guru Pengampu' }}
                </option>
              </optgroup>
            </select>

            <!-- JIKA SEMUA SUDAH DIJADWALKAN PADA EVENT INI -->
            <div
              v-else-if="classSubjects.length > 0"
              class="p-3.5 bg-emerald-50 border border-emerald-200 rounded-2xl text-xs text-emerald-800 flex items-start space-x-2"
            >
              <span class="text-sm">✓</span>
              <div>
                <div class="font-bold">Semua alokasi telah dijadwalkan</div>
                <div class="text-[11px] text-emerald-700 mt-0.5">
                  Seluruh kelas dan mata pelajaran dari Master Kelas Mapel sudah memiliki sesi jadwal ujian pada event ini.
                </div>
              </div>
            </div>

            <!-- FALLBACK JIKA MASTER KELAS MAPEL MASIH KOSONG SAMA SEKALI -->
            <div
              v-else
              class="p-3.5 bg-amber-50 border border-amber-200 rounded-2xl text-xs text-amber-800 space-y-2"
            >
              <div class="flex items-center space-x-2">
                <span class="text-sm">⚠️</span>
                <span class="font-bold">Master Kelas Mapel Masih Kosong</span>
              </div>
              <p class="text-[11px] text-amber-700 leading-relaxed">
                Belum ada pemetaan mata pelajaran untuk rombel di menu <strong>Master > Kelas Mapel</strong>. Silakan alokasikan mapel ke kelas terlebih dahulu agar jadwal ujian terstruktur.
              </p>
              <button
                type="button"
                @click="showScheduleModal = false; switchTab('class-subjects')"
                class="px-3 py-1.5 bg-amber-600 hover:bg-amber-700 text-white rounded-xl font-bold text-xs transition cursor-pointer"
              >
                Buka Menu Kelas Mapel →
              </button>
            </div>
          </div>

          <!-- 2. Judul Sesi (Mengikuti Pilihan Kelas & Mapel di Atas) -->
          <div>
            <label class="block font-bold text-slate-700 mb-1">Judul Ujian / Sesi:</label>
            <input
              v-model="scheduleForm.title"
              type="text"
              required
              placeholder="Contoh: Penilaian Harian / Ujian Matematika Wajib"
              class="w-full px-3.5 py-2.5 rounded-xl border border-slate-300 focus:ring-2 focus:ring-indigo-500 text-xs"
            />
          </div>

          <!-- Waktu: Tanggal, Jam Mulai, Jam Selesai -->
          <div class="bg-slate-50 p-3.5 rounded-2xl border border-slate-200/80 space-y-2">
            <div class="font-bold text-[11px] text-slate-700 flex items-center gap-1.5">
              <span>📅</span>
              <span>Waktu Pelaksanaan Asesmen</span>
            </div>
            <div class="grid grid-cols-1 sm:grid-cols-3 gap-2.5">
              <div>
                <label class="block text-[11px] text-slate-600 mb-1">Tanggal Ujian:</label>
                <input
                  v-model="scheduleForm.exam_date"
                  type="date"
                  required
                  class="w-full px-3 py-2 rounded-xl border border-slate-300 bg-white"
                />
              </div>
              <div>
                <label class="block text-[11px] text-slate-600 mb-1">Jam Mulai (WIB):</label>
                <input
                  v-model="scheduleForm.start_time"
                  type="time"
                  required
                  class="w-full px-3 py-2 rounded-xl border border-slate-300 bg-white"
                />
              </div>
              <div>
                <label class="block text-[11px] text-slate-600 mb-1">Jam Selesai (WIB):</label>
                <input
                  v-model="scheduleForm.end_time"
                  type="time"
                  required
                  class="w-full px-3 py-2 rounded-xl border border-slate-300 bg-white"
                />
              </div>
            </div>
          </div>

          <!-- Token, Durasi, Batas Tab -->
          <div class="grid grid-cols-3 gap-3">
            <div>
              <label class="block font-bold text-slate-700 mb-1">Token Sesi:</label>
              <input
                v-model="scheduleForm.exam_token"
                type="text"
                required
                placeholder="CBT2026"
                class="w-full uppercase font-mono px-3 py-2 rounded-xl border border-slate-300 text-center font-bold"
              />
            </div>
            <div>
              <label class="block font-bold text-slate-700 mb-1">Durasi (Menit):</label>
              <input
                v-model.number="scheduleForm.duration_minutes"
                type="number"
                min="10"
                max="300"
                required
                class="w-full px-3 py-2 rounded-xl border border-slate-300 text-center font-bold"
              />
            </div>
            <div>
              <label class="block font-bold text-slate-700 mb-1">Max Pelanggaran:</label>
              <input
                v-model.number="scheduleForm.max_violations"
                type="number"
                min="1"
                max="20"
                required
                class="w-full px-3 py-2 rounded-xl border border-slate-300 text-center font-bold"
              />
            </div>
          </div>
          <p class="-mt-1 text-[11px] text-slate-500">
            <template v-if="scheduleFormSessionPeers.length">
              Token dipakai bersama {{ scheduleFormSessionPeers.length }} jadwal lain pada sesi {{ scheduleForm.start_time }}
              ({{ scheduleFormSessionPeers.map(p => p.class_room?.name).filter(Boolean).join(', ') }}). Mengubah token di sini ikut mengubah token jadwal tersebut.
            </template>
            <template v-else>Semua jadwal pada tanggal dan jam mulai yang sama otomatis memakai token ini.</template>
          </p>

          <!-- Pengawas (opsional, hanya pemegang schedules:manage) -->
          <div v-if="authStore.hasPermission('schedules:manage')">
            <div class="flex items-center justify-between gap-2 mb-1.5">
              <label class="block font-bold text-slate-700">Pengawas <span class="font-normal text-slate-400">(opsional)</span>:</label>
              <button
                type="button"
                @click="showFormProctorPicker = true"
                class="px-3 h-8 inline-flex items-center justify-center gap-1.5 rounded-lg border text-xs font-semibold whitespace-nowrap shrink-0 transition active:scale-95 cursor-pointer bg-slate-50 hover:bg-slate-100 text-slate-700 border-slate-200"
              >
                <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m5.618-4.016A11.955 11.955 0 0112 2.944a11.955 11.955 0 01-8.618 3.04A12.02 12.02 0 003 9c0 5.591 3.824 10.29 9 11.622 5.176-1.332 9-6.03 9-11.622 0-1.042-.133-2.052-.382-3.016z" />
                </svg>
                <span>{{ scheduleForm.proctors.length > 0 ? 'Ubah Pengawas' : 'Pilih Pengawas' }}</span>
              </button>
            </div>
            <div v-if="scheduleForm.proctors.length > 0" class="flex flex-wrap gap-1.5">
              <span
                v-for="p in scheduleForm.proctors"
                :key="p.id"
                class="inline-flex items-center gap-1 pl-2.5 pr-1 py-1 bg-indigo-50 text-indigo-700 border border-indigo-100 rounded-lg text-[11px] font-semibold"
              >
                <span class="truncate max-w-[180px]">{{ p.full_name || p.username }}</span>
                <button
                  type="button"
                  @click="removeFormProctor(p.id)"
                  class="p-0.5 rounded-md text-indigo-400 hover:text-indigo-700 hover:bg-indigo-100 transition cursor-pointer"
                  :title="`Hapus ${p.full_name || p.username}`"
                >
                  <svg class="w-3 h-3" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
                  </svg>
                </button>
              </span>
            </div>
            <p v-else class="text-[11px] text-slate-400">Belum ada pengawas dipilih. Dapat diatur nanti dari tabel jadwal.</p>
          </div>

          <!-- Randomize Options -->
          <div class="grid grid-cols-2 gap-3 pt-1">
            <label class="flex items-center space-x-2 p-2.5 rounded-xl border border-slate-200 bg-slate-50 cursor-pointer hover:bg-slate-100/60">
              <input v-model="scheduleForm.randomize_questions" type="checkbox" class="w-4 h-4 text-indigo-600 rounded" />
              <span class="text-xs font-semibold text-slate-700">Acak Urutan Soal</span>
            </label>
            <label class="flex items-center space-x-2 p-2.5 rounded-xl border border-slate-200 bg-slate-50 cursor-pointer hover:bg-slate-100/60">
              <input v-model="scheduleForm.randomize_options" type="checkbox" class="w-4 h-4 text-indigo-600 rounded" />
              <span class="text-xs font-semibold text-slate-700">Acak Opsi Pilihan</span>
            </label>
          </div>
        </div>

        <!-- Pinned Footer -->
        <div class="px-6 py-3 bg-slate-50 border-t border-slate-100 flex items-center justify-end space-x-2 shrink-0">
          <button
            type="button"
            @click="showScheduleModal = false"
            class="px-3.5 py-1.5 bg-slate-100 hover:bg-slate-200 rounded-xl text-slate-600 font-semibold text-xs transition cursor-pointer"
          >
            Batal
          </button>
          <button
            type="submit"
            :disabled="!isEditSchedule && availableClassSubjects.length === 0"
            class="px-4 py-1.5 bg-indigo-600 hover:bg-indigo-700 disabled:opacity-50 disabled:cursor-not-allowed text-white rounded-xl font-bold text-xs shadow-xs transition cursor-pointer"
          >
            {{ isEditSchedule ? 'Simpan Perubahan' : 'Terbitkan Jadwal' }}
          </button>
        </div>
      </form>
    </div>
  </div>

  <!-- MODAL: ATUR PENGAWAS (dari tabel jadwal, disimpan langsung) -->
  <ScheduleProctorModal
    v-model="showProctorAssignModal"
    :schedule="scheduleForProctors"
    @saved="loadSchedules"
  />

  <!-- MODAL: PILIH PENGAWAS DI FORM JADWAL (mode lokal, disimpan bersama jadwal) -->
  <ScheduleProctorModal
    v-model="showFormProctorPicker"
    local
    :schedule="{ title: scheduleForm.title }"
    :initial-selected="scheduleForm.proctors"
    @apply="applyFormProctors"
  />

  <!-- MODAL: DETAIL RINCIAN JADWAL UJIAN -->
  <div v-if="showScheduleDetailModal && selectedScheduleDetail" class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-slate-900/60 backdrop-blur-xs">
    <div class="bg-white rounded-3xl max-w-lg w-full shadow-2xl overflow-hidden flex flex-col max-h-[90vh]">
      <!-- Header -->
      <div class="flex items-center justify-between px-6 py-4 border-b border-slate-100 shrink-0 bg-white">
        <h3 class="text-base font-bold text-slate-900">Rincian Informasi Jadwal</h3>
        <button @click="showScheduleDetailModal = false" class="text-slate-400 hover:text-slate-600 p-1.5 rounded-xl hover:bg-slate-100 transition cursor-pointer">
          ✕
        </button>
      </div>

      <!-- Tab Navigation -->
      <div class="flex gap-1 px-4 pt-3 pb-0 shrink-0 border-b border-slate-100 bg-white">
        <button
          type="button"
          @click="scheduleDetailTab = 'info'"
          :class="[
            'px-3 py-2 text-xs font-semibold rounded-t-xl transition cursor-pointer border-b-2 -mb-px',
            scheduleDetailTab === 'info'
              ? 'text-indigo-700 border-indigo-600 bg-indigo-50/60'
              : 'text-slate-500 border-transparent hover:text-slate-700 hover:bg-slate-50'
          ]"
        >Informasi</button>
        <button
          v-if="canGradeEssay(selectedScheduleDetail)"
          type="button"
          @click="switchToEssayTab()"
          :class="[
            'px-3 py-2 text-xs font-semibold rounded-t-xl transition cursor-pointer border-b-2 -mb-px flex items-center gap-1.5',
            scheduleDetailTab === 'essay'
              ? 'text-indigo-700 border-indigo-600 bg-indigo-50/60'
              : 'text-slate-500 border-transparent hover:text-slate-700 hover:bg-slate-50'
          ]"
        >
          <span>Koreksi Essay</span>
          <span
            v-if="!essayLoading && essayQuestions.length > 0"
            :class="[
              'inline-flex items-center justify-center min-w-[18px] h-[18px] px-1 rounded-full text-[10px] font-bold',
              essayTotalPending === 0 ? 'bg-emerald-100 text-emerald-700' : 'bg-amber-100 text-amber-700'
            ]"
          >{{ essayTotalPending === 0 ? 'Selesai' : essayTotalPending }}</span>
        </button>
      </div>

      <!-- Tab: Informasi -->
      <div v-if="scheduleDetailTab === 'info'" class="p-6 space-y-3.5 text-xs overflow-y-auto flex-1">
        <!-- Sesi & Kelas Banner -->
        <div class="p-3.5 bg-slate-50 rounded-2xl border border-slate-200/80 space-y-1">
          <div class="text-[10px] uppercase tracking-wider text-slate-400 font-bold">Judul Sesi Ujian</div>
          <div class="text-sm font-bold text-slate-900 leading-snug">{{ selectedScheduleDetail.title }}</div>
          <div class="flex items-center gap-2 pt-1 flex-wrap">
            <span class="px-2 py-0.5 bg-white border border-slate-200 text-slate-700 font-bold rounded-md text-[11px]">
              Kelas: {{ selectedScheduleDetail.class_room?.name }}
            </span>
            <span class="px-2 py-0.5 bg-indigo-50 border border-indigo-100 text-indigo-700 font-bold rounded-md text-[11px]">
              {{ selectedScheduleDetail.subject?.name || selectedScheduleDetail.bank?.subject?.name || 'Mata Pelajaran' }}
            </span>
          </div>
        </div>

        <!-- Parameter Teknis Grid -->
        <div class="grid grid-cols-2 gap-2.5">
          <div class="p-3 bg-white border border-slate-200 rounded-2xl">
            <span class="text-[10px] text-slate-400 font-bold uppercase block">Durasi Ujian</span>
            <span class="text-base font-black text-slate-800 font-mono">{{ selectedScheduleDetail.duration_minutes }}</span>
            <span class="text-xs text-slate-500 font-medium ml-1">Menit</span>
          </div>
          <div class="p-3 bg-white border border-slate-200 rounded-2xl">
            <span class="text-[10px] text-slate-400 font-bold uppercase block">Batas Toleransi Tab</span>
            <span class="text-base font-black text-slate-800 font-mono">{{ selectedScheduleDetail.max_violations }}x</span>
            <span class="text-xs text-slate-500 font-medium ml-1">Peringatan</span>
          </div>
          <div class="p-3 bg-white border border-slate-200 rounded-2xl">
            <span class="text-[10px] text-slate-400 font-bold uppercase block">Acak Urutan Soal</span>
            <span :class="['text-xs font-bold mt-1 inline-flex items-center gap-1', selectedScheduleDetail.randomize_questions ? 'text-emerald-700' : 'text-slate-500']">
              {{ selectedScheduleDetail.randomize_questions ? 'Aktif (Diacak per Siswa)' : 'Tidak Diacak' }}
            </span>
          </div>
          <div class="p-3 bg-white border border-slate-200 rounded-2xl">
            <span class="text-[10px] text-slate-400 font-bold uppercase block">Acak Opsi Pilihan</span>
            <span :class="['text-xs font-bold mt-1 inline-flex items-center gap-1', selectedScheduleDetail.randomize_options ? 'text-emerald-700' : 'text-slate-500']">
              {{ selectedScheduleDetail.randomize_options ? 'Aktif (Opsi A-E Diacak)' : 'Tidak Diacak' }}
            </span>
          </div>
        </div>

        <!-- Waktu & Token -->
        <div class="p-3.5 bg-indigo-50/50 rounded-2xl border border-indigo-100 grid grid-cols-1 sm:grid-cols-2 gap-3">
          <div>
            <span class="text-[10px] text-indigo-700 font-bold uppercase block">Waktu Pelaksanaan</span>
            <span class="text-xs font-semibold text-slate-800 mt-0.5 block">
              {{ formatScheduleTimeRange(selectedScheduleDetail.start_time, selectedScheduleDetail.end_time) }}
            </span>
          </div>
          <div v-if="selectedScheduleDetail.exam_token">
            <span class="text-[10px] text-indigo-700 font-bold uppercase block">Token Ujian Masuk</span>
            <span class="font-mono font-black text-indigo-700 text-sm tracking-wider mt-0.5 block">
              {{ selectedScheduleDetail.exam_token }}
            </span>
          </div>
        </div>

        <!-- Bank Soal Status -->
        <div class="p-3 bg-white border border-slate-200 rounded-2xl flex items-center justify-between">
          <div>
            <span class="text-[10px] text-slate-400 font-bold uppercase block">Paket Naskah Soal</span>
            <span class="text-xs font-bold text-slate-800">
              {{ selectedScheduleDetail.bank?.title || 'Belum Ada Bank Soal Tertaut' }}
            </span>
          </div>
          <span :class="['px-2.5 py-1 rounded-full text-[10px] font-bold', selectedScheduleDetail.bank?.is_locked ? 'bg-emerald-100 text-emerald-800' : 'bg-yellow-100 text-yellow-800']">
            {{ selectedScheduleDetail.bank?.is_locked ? 'Siap Ujian' : (selectedScheduleDetail.bank ? 'Draft Soal' : 'Kosong') }}
          </span>
        </div>
      </div>

      <!-- Tab: Koreksi Essay -->
      <div v-else-if="scheduleDetailTab === 'essay'" class="overflow-y-auto flex-1 p-4 space-y-4 text-xs">
        <!-- Loading -->
        <div v-if="essayLoading" class="flex flex-col items-center justify-center py-12 gap-3 text-slate-400">
          <svg class="w-6 h-6 animate-spin" fill="none" viewBox="0 0 24 24">
            <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
            <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z"></path>
          </svg>
          <span class="text-xs font-medium">Memuat data jawaban essay...</span>
        </div>

        <!-- Empty state -->
        <div v-else-if="essayQuestions.length === 0" class="flex flex-col items-center justify-center py-12 gap-2 text-slate-400">
          <svg class="w-10 h-10 text-slate-300" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
          </svg>
          <p class="text-xs font-medium text-slate-500 text-center">Jadwal ini tidak memiliki soal essay atau isian singkat.</p>
        </div>

        <!-- Question cards -->
        <div v-else class="space-y-5">
          <div
            v-for="q in essayQuestions"
            :key="q.question_id"
            class="border border-slate-200 rounded-2xl overflow-hidden"
          >
            <!-- Question header -->
            <div class="flex items-center justify-between px-4 py-3 bg-slate-50 border-b border-slate-200">
              <div class="flex items-center gap-2 flex-wrap">
                <span class="font-black text-slate-800 text-[13px]">No. {{ q.question_number }}</span>
                <span class="px-2 py-0.5 bg-white border border-slate-200 text-slate-600 rounded-md text-[10px] font-bold uppercase">{{ q.question_type }}</span>
                <span class="text-slate-500 font-medium">Bobot {{ q.score_weight }}</span>
              </div>
              <span :class="[
                'px-2 py-0.5 rounded-full text-[10px] font-bold',
                q.graded_count >= q.total_count ? 'bg-emerald-100 text-emerald-700' : 'bg-amber-100 text-amber-700'
              ]">
                {{ q.graded_count }} / {{ q.total_count }} dinilai
              </span>
            </div>

            <!-- Question text -->
            <div class="px-4 py-3 bg-white border-b border-slate-100">
              <RichContentRenderer :content="q.content_html" custom-class="text-slate-700 font-medium leading-relaxed" />
            </div>

            <!-- Answers -->
            <div class="divide-y divide-slate-100">
              <div
                v-for="a in q.answers"
                :key="a.answer_id"
                class="px-4 py-3 space-y-2.5"
                :class="a.is_graded ? 'bg-emerald-50/40' : 'bg-white'"
              >
                <!-- Student info -->
                <div class="flex items-center justify-between">
                  <div class="flex items-center gap-2">
                    <span class="font-bold text-slate-800">{{ a.student_name }}</span>
                    <span class="text-slate-400 font-mono text-[11px]">{{ a.student_nis }}</span>
                  </div>
                  <span v-if="a.is_graded" class="flex items-center gap-1 text-emerald-600 font-bold text-[11px]">
                    <svg class="w-3.5 h-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2.5" d="M5 13l4 4L19 7" />
                    </svg>
                    Sudah dinilai
                  </span>
                </div>

                <!-- Answer text -->
                <div class="bg-slate-50 border border-slate-200 rounded-xl px-3 py-2.5">
                  <div class="text-[10px] text-slate-400 font-bold uppercase mb-1">Jawaban Siswa:</div>
                  <div class="text-slate-700 leading-relaxed whitespace-pre-wrap">{{ a.answer_text || '(Tidak ada jawaban)' }}</div>
                </div>

                <!-- Grading inputs -->
                <div class="flex items-start gap-2 flex-wrap">
                  <div class="flex items-center gap-1.5">
                    <label :for="`score-${a.answer_id}`" class="text-slate-600 font-semibold whitespace-nowrap">Nilai:</label>
                    <input
                      :id="`score-${a.answer_id}`"
                      type="number"
                      min="0"
                      :max="q.score_weight"
                      step="0.5"
                      :value="essayDraft[a.answer_id]?.score_awarded ?? ''"
                      @input="e => { if (!essayDraft[a.answer_id]) essayDraft[a.answer_id] = { score_awarded: null, teacher_comment: '' }; essayDraft[a.answer_id].score_awarded = e.target.value === '' ? null : parseFloat(e.target.value) }"
                      class="w-20 px-2 py-1.5 border border-slate-300 rounded-xl font-mono text-center focus:outline-none focus:ring-2 focus:ring-indigo-400 text-xs"
                      :placeholder="`/ ${q.score_weight}`"
                    />
                  </div>
                  <div class="flex-1 flex items-center gap-1.5 min-w-[140px]">
                    <label :for="`comment-${a.answer_id}`" class="text-slate-600 font-semibold whitespace-nowrap">Komentar:</label>
                    <input
                      :id="`comment-${a.answer_id}`"
                      type="text"
                      :value="essayDraft[a.answer_id]?.teacher_comment ?? ''"
                      @input="e => { if (!essayDraft[a.answer_id]) essayDraft[a.answer_id] = { score_awarded: null, teacher_comment: '' }; essayDraft[a.answer_id].teacher_comment = e.target.value }"
                      class="flex-1 px-2 py-1.5 border border-slate-300 rounded-xl focus:outline-none focus:ring-2 focus:ring-indigo-400 text-xs"
                      placeholder="Opsional..."
                    />
                  </div>
                </div>
              </div>
            </div>

            <!-- Save button per question -->
            <div class="px-4 py-3 bg-slate-50 border-t border-slate-100 flex justify-end">
              <button
                type="button"
                @click="saveEssayQuestion(q)"
                :disabled="essaySaving[q.question_id]"
                class="px-4 py-2 bg-indigo-600 hover:bg-indigo-700 disabled:opacity-60 text-white font-bold text-xs rounded-xl shadow-xs transition active:scale-95 cursor-pointer flex items-center gap-1.5"
              >
                <svg v-if="essaySaving[q.question_id]" class="w-3.5 h-3.5 animate-spin" fill="none" viewBox="0 0 24 24">
                  <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
                  <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z"></path>
                </svg>
                <span>{{ essaySaving[q.question_id] ? 'Menyimpan...' : `Simpan Penilaian Soal No. ${q.question_number}` }}</span>
              </button>
            </div>
          </div>
        </div>
      </div>

      <!-- Actions in Modal (Pinned Footer - Only Edit and Tutup) -->
      <div class="px-6 py-3.5 bg-slate-50 border-t border-slate-100 flex items-center justify-end gap-2 shrink-0">
        <button
          v-if="canManageSchedules && scheduleDetailTab === 'info'"
          type="button"
          @click="openEditSchedule(selectedScheduleDetail); showScheduleDetailModal = false"
          class="px-4 py-2 bg-indigo-600 hover:bg-indigo-700 text-white font-bold text-xs rounded-xl shadow-xs transition active:scale-95 cursor-pointer flex items-center gap-1.5"
        >
          <svg class="w-3.5 h-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z" />
          </svg>
          <span>Edit Sesi</span>
        </button>
        <button
          type="button"
          @click="showScheduleDetailModal = false"
          class="px-4 py-2 bg-slate-200 hover:bg-slate-300 text-slate-700 font-bold text-xs rounded-xl transition cursor-pointer"
        >
          Tutup
        </button>
      </div>
    </div>
  </div>

  <!-- MODAL: TAUTKAN BANK SOAL KE JADWAL -->
  <div v-if="showLinkBankModal" class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-slate-900/60 backdrop-blur-xs">
    <div class="bg-white rounded-3xl max-w-md w-full p-6 shadow-2xl space-y-4">
      <div class="flex items-center justify-between border-b border-slate-100 pb-3">
        <div>
          <h3 class="text-base font-bold text-slate-900">Tautkan Naskah Bank Soal</h3>
          <p class="text-xs text-slate-500 mt-0.5">
            Hubungkan bank naskah soal dengan sesi jadwal ujian ini.
          </p>
        </div>
        <button @click="showLinkBankModal = false" class="text-slate-400 hover:text-slate-600 p-1">✕</button>
      </div>

      <div v-if="selectedScheduleForLink" class="bg-indigo-50/50 p-3.5 rounded-2xl border border-indigo-100 text-xs space-y-1.5">
        <div class="font-bold text-indigo-950">{{ selectedScheduleForLink.title }}</div>
        <div class="text-[11px] text-slate-600 flex flex-wrap gap-x-3">
          <span>Kelas: <strong>{{ selectedScheduleForLink.class_room?.name }}</strong></span>
          <span>Mapel: <strong>{{ selectedScheduleForLink.subject?.name || selectedScheduleForLink.bank?.subject?.name || '-' }}</strong></span>
        </div>
      </div>

      <form @submit.prevent="submitLinkBank" class="space-y-4 text-xs">
        <div>
          <label class="block font-bold text-slate-700 mb-1.5">Pilih Bank Soal:</label>
          <select
            v-model="selectedBankForLink"
            class="w-full px-3 py-2 rounded-xl border border-slate-300 font-medium focus:ring-2 focus:ring-indigo-500"
          >
            <option value="">-- Jangan Tautkan Soal (Kosongkan) --</option>
            <option v-for="b in availableBanksForSelectedSchedule" :key="b.id" :value="b.id">
              [{{ b.is_locked ? 'Terkunci' : 'Draft' }}] {{ b.title }} ({{ b.total_questions }} Soal)
            </option>
          </select>
        </div>

        <div class="bg-amber-50 border border-amber-200 rounded-xl p-3 text-[11px] text-amber-900 flex items-start gap-2">
          <span class="text-sm shrink-0 leading-none">⚠️</span>
          <span>
            Pastikan status bank soal telah <strong>Terkunci</strong> pada modul <strong>Kesiapan Soal</strong> agar siswa dapat mengakses dan memulai ujian.
          </span>
        </div>

        <div class="pt-2 flex justify-end space-x-2">
          <button
            type="button"
            @click="showLinkBankModal = false"
            class="px-4 py-2 bg-slate-100 hover:bg-slate-200 rounded-xl text-slate-700 font-semibold transition"
          >
            Batal
          </button>
          <button
            type="submit"
            class="px-5 py-2 bg-indigo-600 hover:bg-indigo-700 text-white rounded-xl font-bold shadow-xs transition"
          >
            Simpan Tautan
          </button>
        </div>
      </form>
    </div>
  </div>
</template>

<script setup>
import RichContentRenderer from '../../../components/common/RichContentRenderer.vue'
import ScheduleProctorModal from '@/components/admin/ScheduleProctorModal.vue'
import { useDashboard } from './context'

const {
  applyFormProctors,
  authStore,
  availableBanksForSelectedSchedule,
  availableClassSubjectGrades,
  availableClassSubjects,
  canManageSchedules,
  classSubjects,
  essayDraft,
  essayLoading,
  essayQuestions,
  essaySaving,
  essayTotalPending,
  formatScheduleTimeRange,
  groupedAvailableClassSubjects,
  isEditSchedule,
  loadSchedules,
  onClassSubjectChange,
  openEditSchedule,
  removeFormProctor,
  saveEssayQuestion,
  scheduleDetailTab,
  scheduleForProctors,
  scheduleForm,
  scheduleFormSessionPeers,
  selectedBankForLink,
  selectedScheduleClassSubjectId,
  selectedScheduleDetail,
  selectedScheduleForLink,
  showFormProctorPicker,
  showLinkBankModal,
  showProctorAssignModal,
  showScheduleDetailModal,
  showScheduleModal,
  submitLinkBank,
  submitScheduleForm,
  switchTab,
  switchToEssayTab,
  canGradeEssay,
} = useDashboard()
</script>
