<template>
  <!-- TAB: BANK SOAL & KESIAPAN (UNIFIED) -->
  <div v-if="activeTab === 'questions'" class="space-y-4">
    <!-- 4 Contextual Metric Cards -->
    <div class="grid grid-cols-2 sm:grid-cols-4 gap-3">
      <div class="bg-white rounded-3xl p-4 border border-slate-200 shadow-xs">
        <span class="text-[11px] font-semibold text-slate-500">Total Bank Soal</span>
        <div class="text-2xl font-black text-slate-900 mt-1">
          {{ (readinessData.question_banks || []).length }} <span class="text-xs font-medium text-slate-400">Paket</span>
        </div>
      </div>
      <div class="bg-white rounded-3xl p-4 border border-blue-200 shadow-xs bg-blue-50/20">
        <span class="text-[11px] font-semibold text-blue-700">Total Butir Soal</span>
        <div class="text-2xl font-black text-blue-700 mt-1">
          {{ totalQuestionsCount }} <span class="text-xs font-medium text-blue-400">Soal</span>
        </div>
      </div>
      <div class="bg-white rounded-3xl p-4 border border-emerald-200 shadow-xs bg-emerald-50/20">
        <span class="text-[11px] font-semibold text-emerald-700">Terkunci</span>
        <div class="text-2xl font-black text-emerald-700 mt-1">
          {{ lockedBanksCount }} <span class="text-xs font-medium text-emerald-400">Paket</span>
        </div>
      </div>
      <div class="bg-white rounded-3xl p-4 border border-amber-200 shadow-xs bg-amber-50/20">
        <span class="text-[11px] font-semibold text-amber-700">Draft</span>
        <div class="text-2xl font-black text-amber-700 mt-1">
          {{ draftBanksCount }} <span class="text-xs font-medium text-amber-400">Paket</span>
        </div>
      </div>
    </div>

    <!-- Unified Action Toolbar (Search, Filter Pills & Action Buttons) -->
    <div class="flex flex-col lg:flex-row items-stretch lg:items-center justify-between gap-3 bg-white p-4 rounded-3xl border border-slate-200 shadow-xs">
      <!-- Search & Filter Pills -->
      <div class="flex flex-col sm:flex-row items-stretch sm:items-center gap-2.5 flex-1">
        <div class="relative w-full sm:w-72">
          <input
            v-model="questionBankSearchQuery"
            type="text"
            placeholder="Cari mapel, judul bank soal..."
            class="w-full pl-9 pr-3 py-2 bg-slate-50 border border-slate-200 rounded-xl text-xs focus:ring-2 focus:ring-indigo-500 focus:bg-white transition"
          />
          <svg class="w-4 h-4 text-slate-400 absolute left-3 top-2.5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
          </svg>
        </div>

        <!-- Status Filter Pills -->
        <div class="flex items-center gap-1.5 overflow-x-auto select-none">
          <button
            type="button"
            @click="questionBankStatusFilter = 'all'; questionBankCurrentPage = 1"
            :class="[
              'px-3 py-1.5 rounded-xl font-bold text-xs transition cursor-pointer shrink-0',
              questionBankStatusFilter === 'all' ? 'bg-indigo-600 text-white shadow-xs' : 'bg-slate-100 text-slate-600 hover:bg-slate-200'
            ]"
          >
            Semua ({{ (readinessData.question_banks || []).length }})
          </button>
          <button
            type="button"
            @click="questionBankStatusFilter = 'locked'; questionBankCurrentPage = 1"
            :class="[
              'px-3 py-1.5 rounded-xl font-bold text-xs transition cursor-pointer shrink-0',
              questionBankStatusFilter === 'locked' ? 'bg-indigo-600 text-white shadow-xs' : 'bg-slate-100 text-slate-600 hover:bg-slate-200'
            ]"
          >
            Terkunci ({{ lockedBanksCount }})
          </button>
          <button
            type="button"
            @click="questionBankStatusFilter = 'draft'; questionBankCurrentPage = 1"
            :class="[
              'px-3 py-1.5 rounded-xl font-bold text-xs transition cursor-pointer shrink-0',
              questionBankStatusFilter === 'draft' ? 'bg-indigo-600 text-white shadow-xs' : 'bg-slate-100 text-slate-600 hover:bg-slate-200'
            ]"
          >
            Draft ({{ draftBanksCount }})
          </button>
        </div>
      </div>

      <!-- Action Buttons -->
      <div class="flex items-center flex-wrap gap-2 self-start lg:self-auto shrink-0">
        <button
          type="button"
          @click="downloadQuestionBankTemplate"
          class="px-3 py-2 bg-slate-100 hover:bg-slate-200 text-slate-700 font-bold text-xs rounded-xl shadow-xs transition flex items-center space-x-1 cursor-pointer"
          title="Unduh format spreadsheet untuk butir soal"
        >
          <svg class="w-4 h-4 text-emerald-600" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 10v6m0 0l-3-3m3 3l3-3m2 8H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
          </svg>
          <span>Format Excel</span>
        </button>
        <button
          type="button"
          @click="openCreateBankModal"
          class="px-3.5 py-2 bg-indigo-600 hover:bg-indigo-700 active:scale-95 text-white font-bold text-xs rounded-xl shadow-xs transition flex items-center space-x-1 cursor-pointer"
        >
          <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
          </svg>
          <span>Buat Bank Soal</span>
        </button>
      </div>
    </div>

    <!-- Unified Question Banks Table -->
    <div class="bg-white rounded-3xl border border-slate-200 shadow-sm overflow-hidden">
      <div class="overflow-x-auto">
        <table class="w-full text-left text-xs">
          <thead>
            <tr class="bg-slate-50 border-b border-slate-200 text-slate-600 font-bold uppercase text-[10px] select-none whitespace-nowrap">
              <!-- Col 1: Mapel & Judul Naskah (Sortable) -->
              <th class="py-3 px-4">
                <button
                  type="button"
                  @click="toggleQuestionBankSort('title')"
                  class="group inline-flex items-center gap-1 cursor-pointer transition hover:text-indigo-600 font-bold uppercase whitespace-nowrap"
                  :class="questionBankSortKey === 'title' ? 'text-indigo-600' : 'text-slate-600'"
                  title="Urutkan berdasarkan Judul & Mata Pelajaran"
                >
                  <span>Mata Pelajaran & Judul Naskah</span>
                  <svg v-if="questionBankSortKey === 'title' && questionBankSortOrder === 'asc'" class="w-3 h-3 text-indigo-600 shrink-0" viewBox="0 0 20 20" fill="currentColor"><path fill-rule="evenodd" d="M5.293 9.707a1 1 0 010-1.414l4-4a1 1 0 011.414 0l4 4a1 1 0 01-1.414 1.414L11 7.414V15a1 1 0 11-2 0V7.414L6.707 9.707a1 1 0 01-1.414 0z" clip-rule="evenodd" /></svg>
                  <svg v-else-if="questionBankSortKey === 'title' && questionBankSortOrder === 'desc'" class="w-3 h-3 text-indigo-600 shrink-0" viewBox="0 0 20 20" fill="currentColor"><path fill-rule="evenodd" d="M14.707 10.293a1 1 0 010 1.414l-4 4a1 1 0 01-1.414 0l-4-4a1 1 0 111.414-1.414L9 12.586V5a1 1 0 012 0v7.586l2.293-2.293a1 1 0 011.414 0z" clip-rule="evenodd" /></svg>
                  <svg v-else class="w-3 h-3 text-slate-300 group-hover:text-slate-400 shrink-0" viewBox="0 0 20 20" fill="currentColor"><path d="M5 8l5-5 5 5H5zm10 4l-5 5-5-5h10z" /></svg>
                </button>
              </th>

              <!-- Col 2: Penyusun / Guru (Sortable) -->
              <th class="py-3 px-4">
                <button
                  type="button"
                  @click="toggleQuestionBankSort('creator')"
                  class="group inline-flex items-center gap-1 cursor-pointer transition hover:text-indigo-600 font-bold uppercase whitespace-nowrap"
                  :class="questionBankSortKey === 'creator' ? 'text-indigo-600' : 'text-slate-600'"
                  title="Urutkan berdasarkan Penyusun / Guru"
                >
                  <span>Penyusun</span>
                  <svg v-if="questionBankSortKey === 'creator' && questionBankSortOrder === 'asc'" class="w-3 h-3 text-indigo-600 shrink-0" viewBox="0 0 20 20" fill="currentColor"><path fill-rule="evenodd" d="M5.293 9.707a1 1 0 010-1.414l4-4a1 1 0 011.414 0l4 4a1 1 0 01-1.414 1.414L11 7.414V15a1 1 0 11-2 0V7.414L6.707 9.707a1 1 0 01-1.414 0z" clip-rule="evenodd" /></svg>
                  <svg v-else-if="questionBankSortKey === 'creator' && questionBankSortOrder === 'desc'" class="w-3 h-3 text-indigo-600 shrink-0" viewBox="0 0 20 20" fill="currentColor"><path fill-rule="evenodd" d="M14.707 10.293a1 1 0 010 1.414l-4 4a1 1 0 01-1.414 0l-4-4a1 1 0 111.414-1.414L9 12.586V5a1 1 0 012 0v7.586l2.293-2.293a1 1 0 011.414 0z" clip-rule="evenodd" /></svg>
                  <svg v-else class="w-3 h-3 text-slate-300 group-hover:text-slate-400 shrink-0" viewBox="0 0 20 20" fill="currentColor"><path d="M5 8l5-5 5 5H5zm10 4l-5 5-5-5h10z" /></svg>
                </button>
              </th>

              <!-- Col 3: Jumlah Soal (Sortable) -->
              <th class="py-3 px-4 text-center">
                <button
                  type="button"
                  @click="toggleQuestionBankSort('questions')"
                  class="group inline-flex items-center gap-1 cursor-pointer transition hover:text-indigo-600 font-bold uppercase mx-auto whitespace-nowrap"
                  :class="questionBankSortKey === 'questions' ? 'text-indigo-600' : 'text-slate-600'"
                  title="Urutkan berdasarkan Jumlah Soal"
                >
                  <span>Jumlah Soal</span>
                  <svg v-if="questionBankSortKey === 'questions' && questionBankSortOrder === 'asc'" class="w-3 h-3 text-indigo-600 shrink-0" viewBox="0 0 20 20" fill="currentColor"><path fill-rule="evenodd" d="M5.293 9.707a1 1 0 010-1.414l4-4a1 1 0 011.414 0l4 4a1 1 0 01-1.414 1.414L11 7.414V15a1 1 0 11-2 0V7.414L6.707 9.707a1 1 0 01-1.414 0z" clip-rule="evenodd" /></svg>
                  <svg v-else-if="questionBankSortKey === 'questions' && questionBankSortOrder === 'desc'" class="w-3 h-3 text-indigo-600 shrink-0" viewBox="0 0 20 20" fill="currentColor"><path fill-rule="evenodd" d="M14.707 10.293a1 1 0 010 1.414l-4 4a1 1 0 01-1.414 0l-4-4a1 1 0 111.414-1.414L9 12.586V5a1 1 0 012 0v7.586l2.293-2.293a1 1 0 011.414 0z" clip-rule="evenodd" /></svg>
                  <svg v-else class="w-3 h-3 text-slate-300 group-hover:text-slate-400 shrink-0" viewBox="0 0 20 20" fill="currentColor"><path d="M5 8l5-5 5 5H5zm10 4l-5 5-5-5h10z" /></svg>
                </button>
              </th>

              <!-- Col 4: Status (Sortable) -->
              <th class="py-3 px-4 text-center">
                <button
                  type="button"
                  @click="toggleQuestionBankSort('status')"
                  class="group inline-flex items-center gap-1 cursor-pointer transition hover:text-indigo-600 font-bold uppercase mx-auto whitespace-nowrap"
                  :class="questionBankSortKey === 'status' ? 'text-indigo-600' : 'text-slate-600'"
                  title="Urutkan berdasarkan Status"
                >
                  <span>Status</span>
                  <svg v-if="questionBankSortKey === 'status' && questionBankSortOrder === 'asc'" class="w-3 h-3 text-indigo-600 shrink-0" viewBox="0 0 20 20" fill="currentColor"><path fill-rule="evenodd" d="M5.293 9.707a1 1 0 010-1.414l4-4a1 1 0 011.414 0l4 4a1 1 0 01-1.414 1.414L11 7.414V15a1 1 0 11-2 0V7.414L6.707 9.707a1 1 0 01-1.414 0z" clip-rule="evenodd" /></svg>
                  <svg v-else-if="questionBankSortKey === 'status' && questionBankSortOrder === 'desc'" class="w-3 h-3 text-indigo-600 shrink-0" viewBox="0 0 20 20" fill="currentColor"><path fill-rule="evenodd" d="M14.707 10.293a1 1 0 010 1.414l-4 4a1 1 0 01-1.414 0l-4-4a1 1 0 111.414-1.414L9 12.586V5a1 1 0 012 0v7.586l2.293-2.293a1 1 0 011.414 0z" clip-rule="evenodd" /></svg>
                  <svg v-else class="w-3 h-3 text-slate-300 group-hover:text-slate-400 shrink-0" viewBox="0 0 20 20" fill="currentColor"><path d="M5 8l5-5 5 5H5zm10 4l-5 5-5-5h10z" /></svg>
                </button>
              </th>

              <!-- Col 5: Aksi Tindakan -->
              <th class="py-3 px-4 text-right whitespace-nowrap">Aksi</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-slate-100">
            <tr v-if="filteredQuestionBanks.length === 0">
              <td colspan="5" class="py-8 text-center text-slate-400">
                <div v-if="(readinessData.question_banks || []).length === 0" class="space-y-2">
                  <div>Belum ada paket bank soal terdaftar.</div>
                  <button
                    type="button"
                    @click="openCreateBankModal"
                    class="px-3.5 py-1.5 bg-indigo-600 hover:bg-indigo-700 text-white font-bold text-xs rounded-xl shadow-xs transition cursor-pointer"
                  >
                    + Buat Bank Soal Sekarang
                  </button>
                </div>
                <div v-else class="space-y-1.5">
                  <div class="font-bold text-slate-700 text-xs">Tidak ada bank soal yang sesuai dengan filter</div>
                  <div class="text-[11px] text-slate-400">Coba sesuaikan kata kunci pencarian atau filter status naskah.</div>
                  <button
                    type="button"
                    @click="questionBankSearchQuery = ''; questionBankStatusFilter = 'all'; questionBankSortKey = 'title'; questionBankSortOrder = 'asc'; questionBankCurrentPage = 1"
                    class="px-3 py-1 bg-slate-100 hover:bg-slate-200 text-slate-700 font-bold text-xs rounded-xl transition cursor-pointer mt-1"
                  >
                    Reset Filter
                  </button>
                </div>
              </td>
            </tr>
            <tr v-for="b in paginatedQuestionBanks" :key="b.id" class="hover:bg-slate-50/60 transition">
              <!-- Col 1: Mapel & Judul Naskah -->
              <td class="py-3.5 px-4">
                <div class="flex items-center gap-2 flex-wrap">
                  <span class="font-bold text-slate-900 text-xs">{{ b.title }}</span>
                  <span class="px-2 py-0.5 bg-indigo-50 text-indigo-700 text-[10px] font-semibold rounded-md border border-indigo-100">
                    {{ b.subject?.name || '-' }}{{ bankScopeLabel(b) ? ` · ${bankScopeLabel(b)}` : '' }}
                  </span>
                </div>
                <div class="text-[11px] text-slate-400 mt-0.5 font-mono">
                  Kode Mapel: {{ b.subject?.code || '-' }}
                </div>
              </td>

              <!-- Col 2: Penyusun -->
              <td class="py-3.5 px-4 text-slate-700 font-medium">
                {{ b.created_by?.full_name || 'Guru Pengampu' }}
              </td>

              <!-- Col 3: Jumlah Soal -->
              <td class="py-3.5 px-4 text-center">
                <button
                  type="button"
                  @click="openBankQuestionsModal(b)"
                  class="font-mono font-bold text-xs hover:underline cursor-pointer inline-flex items-center gap-1 group"
                  :class="b.total_questions > 0 ? 'text-indigo-600 hover:text-indigo-800' : 'text-amber-600'"
                  title="Klik untuk melihat daftar soal"
                >
                  <span>{{ b.total_questions }} Soal</span>
                  <svg class="w-3 h-3 opacity-60 group-hover:opacity-100 transition" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z" />
                  </svg>
                </button>
              </td>

              <!-- Col 4: Status Kunci -->
              <td class="py-3.5 px-4 text-center">
                <span :class="['px-2.5 py-1 rounded-full text-[10px] font-bold inline-block', b.is_locked ? 'bg-emerald-100 text-emerald-800' : 'bg-amber-100 text-amber-800']">
                  {{ b.is_locked ? 'Terkunci' : 'Draft' }}
                </span>
              </td>

              <!-- Col 5: Aksi Tindakan -->
              <td class="py-3.5 px-4 text-right">
                <div class="flex items-center justify-end gap-1.5">
                  <!-- Aksi bank soal: tinggi, bentuk, dan ketebalan huruf seragam -->
                  <!-- Pratinjau (simulator tampilan siswa) -->
                  <button
                    type="button"
                    @click="previewBankDirectly(b)"
                    class="px-3 h-8 inline-flex items-center justify-center gap-1.5 rounded-lg border text-xs font-semibold whitespace-nowrap shrink-0 transition active:scale-95 cursor-pointer bg-indigo-50 hover:bg-indigo-100 text-indigo-700 border-indigo-200"
                    title="Buka Simulator Pratinjau Siswa"
                  >
                    <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z" />
                    </svg>
                    <span>Pratinjau</span>
                  </button>

                  <!-- Cetak naskah & kunci (hanya bank terkunci) -->
                  <button
                    v-if="canPrintQuestionBanks && b.is_locked"
                    type="button"
                    @click="openQuestionPrint(b)"
                    class="px-3 h-8 inline-flex items-center justify-center gap-1.5 rounded-lg border text-xs font-semibold whitespace-nowrap shrink-0 transition active:scale-95 cursor-pointer bg-slate-50 hover:bg-slate-100 text-slate-700 border-slate-200"
                    title="Cetak naskah soal dan kunci jawaban"
                  >
                    <svg class="w-4 h-4 text-slate-500" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17 17h2a2 2 0 002-2v-4a2 2 0 00-2-2H5a2 2 0 00-2 2v4a2 2 0 002 2h2m2 4h6a2 2 0 002-2v-4a2 2 0 00-2-2H9a2 2 0 00-2 2v4a2 2 0 002 2zm8-12V5a2 2 0 00-2-2H9a2 2 0 00-2 2v4h10z" />
                    </svg>
                    <span>Cetak</span>
                  </button>

                  <!-- Impor / Unggah Soal -->
                  <button
                    type="button"
                    @click="openUploadBankModal(b)"
                    class="px-3 h-8 inline-flex items-center justify-center gap-1.5 rounded-lg border text-xs font-semibold whitespace-nowrap shrink-0 transition active:scale-95 cursor-pointer bg-slate-50 hover:bg-slate-100 text-slate-700 border-slate-200"
                    title="Unggah Soal via Excel"
                  >
                    <svg class="w-4 h-4 text-emerald-600" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-8l-4-4m0 0L8 8m4-4v12" />
                    </svg>
                    <span>Unggah Soal</span>
                  </button>

                  <!-- Kunci / Buka Kunci -->
                  <button
                    type="button"
                    @click="toggleBankLockWithConfirm(b)"
                    :class="[
                      'px-3 h-8 inline-flex items-center justify-center gap-1.5 rounded-lg border text-xs font-semibold whitespace-nowrap shrink-0 transition active:scale-95 cursor-pointer',
                      b.is_locked ? 'bg-amber-50 text-amber-800 hover:bg-amber-100 border-amber-200' : 'bg-emerald-600 text-white hover:bg-emerald-700 border-emerald-600'
                    ]"
                    :title="b.is_locked ? 'Buka kunci naskah agar dapat diubah' : 'Kunci naskah agar tidak dapat diubah'"
                  >
                    <svg v-if="b.is_locked" class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 11V7a4 4 0 118 0m-4 8v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2z" />
                    </svg>
                    <svg v-else class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z" />
                    </svg>
                    <span>{{ b.is_locked ? 'Buka Kunci' : 'Kunci Naskah' }}</span>
                  </button>

                  <!-- Edit Bank Soal -->
                  <button
                    type="button"
                    @click="openEditBankModal(b)"
                    class="w-8 h-8 inline-flex items-center justify-center gap-1.5 rounded-lg border text-xs font-semibold whitespace-nowrap shrink-0 transition active:scale-95 cursor-pointer bg-slate-50 hover:bg-slate-100 text-slate-600 border-slate-200"
                    title="Edit Bank Soal"
                  >
                    <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z" />
                    </svg>
                  </button>

                  <!-- Hapus Bank Soal -->
                  <button
                    type="button"
                    @click="deleteQuestionBank(b)"
                    class="w-8 h-8 inline-flex items-center justify-center gap-1.5 rounded-lg border text-xs font-semibold whitespace-nowrap shrink-0 transition active:scale-95 cursor-pointer bg-rose-50 hover:bg-rose-100 text-rose-600 border-rose-200"
                    title="Hapus Bank Soal"
                  >
                    <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
                    </svg>
                  </button>
                </div>
              </td>
            </tr>
          </tbody>
        </table>
      </div>

      <!-- Pagination Controls Footer for Question Banks -->
      <div v-if="filteredQuestionBanks.length > 0" class="border-t border-slate-200/80 bg-slate-50/60 px-4 py-3 flex flex-col sm:flex-row items-center justify-between gap-3 text-xs text-slate-600 select-none">
        <!-- Left: Info & Per-Page Selector -->
        <div class="flex items-center gap-3 flex-wrap justify-center sm:justify-start">
          <span>
            Menampilkan
            <span class="font-bold text-slate-800">{{ (questionBankCurrentPage - 1) * questionBankPerPage + 1 }}</span>
            –
            <span class="font-bold text-slate-800">{{ Math.min(questionBankCurrentPage * questionBankPerPage, filteredQuestionBanks.length) }}</span>
            dari
            <span class="font-bold text-slate-800">{{ filteredQuestionBanks.length }}</span> paket
          </span>

          <div class="flex items-center gap-1.5 pl-2.5 border-l border-slate-200">
            <span class="text-slate-400 text-[11px]">Tampilkan:</span>
            <select
              v-model.number="questionBankPerPage"
              class="bg-white border border-slate-200 rounded-lg px-2 py-1 text-xs font-semibold text-slate-700 focus:outline-none focus:ring-1 focus:ring-indigo-500 cursor-pointer shadow-2xs"
            >
              <option :value="5">5 / hal</option>
              <option :value="10">10 / hal</option>
              <option :value="25">25 / hal</option>
              <option :value="50">50 / hal</option>
            </select>
          </div>
        </div>

        <!-- Right: Page Navigation -->
        <div v-if="totalQuestionBankPages > 1" class="flex items-center gap-1">
          <!-- Previous Button -->
          <button
            type="button"
            @click="setQuestionBankPage(questionBankCurrentPage - 1)"
            :disabled="questionBankCurrentPage === 1"
            class="px-2.5 py-1 rounded-lg border border-slate-200 bg-white hover:bg-slate-100 disabled:opacity-40 disabled:cursor-not-allowed transition font-medium text-xs flex items-center gap-1 cursor-pointer shadow-2xs"
            title="Halaman Sebelumnya"
          >
            <svg class="w-3.5 h-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 19l-7-7 7-7" />
            </svg>
            <span class="hidden sm:inline">Sebelumnya</span>
          </button>

          <!-- Page Numbers -->
          <div class="flex items-center gap-1">
            <button
              v-for="(p, idx) in displayedQuestionBankPages"
              :key="idx"
              type="button"
              @click="setQuestionBankPage(p)"
              :disabled="p === '...'"
              :class="[
                'min-w-[28px] h-7 px-2 rounded-lg text-xs font-bold transition flex items-center justify-center',
                p === '...' ? 'cursor-default text-slate-400' :
                p === questionBankCurrentPage ? 'bg-indigo-600 text-white shadow-xs' : 'bg-white border border-slate-200 text-slate-700 hover:bg-slate-100 cursor-pointer shadow-2xs'
              ]"
            >
              {{ p }}
            </button>
          </div>

          <!-- Next Button -->
          <button
            type="button"
            @click="setQuestionBankPage(questionBankCurrentPage + 1)"
            :disabled="questionBankCurrentPage === totalQuestionBankPages"
            class="px-2.5 py-1 rounded-lg border border-slate-200 bg-white hover:bg-slate-100 disabled:opacity-40 disabled:cursor-not-allowed transition font-medium text-xs flex items-center gap-1 cursor-pointer shadow-2xs"
            title="Halaman Berikutnya"
          >
            <span class="hidden sm:inline">Berikutnya</span>
            <svg class="w-3.5 h-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" />
            </svg>
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { useDashboard } from './context'

const {
  activeTab,
  bankScopeLabel,
  canPrintQuestionBanks,
  deleteQuestionBank,
  displayedQuestionBankPages,
  downloadQuestionBankTemplate,
  draftBanksCount,
  filteredQuestionBanks,
  lockedBanksCount,
  openBankQuestionsModal,
  openCreateBankModal,
  openEditBankModal,
  openQuestionPrint,
  openUploadBankModal,
  paginatedQuestionBanks,
  previewBankDirectly,
  questionBankCurrentPage,
  questionBankPerPage,
  questionBankSearchQuery,
  questionBankSortKey,
  questionBankSortOrder,
  questionBankStatusFilter,
  readinessData,
  setQuestionBankPage,
  toggleBankLockWithConfirm,
  toggleQuestionBankSort,
  totalQuestionBankPages,
  totalQuestionsCount,
} = useDashboard()
</script>
