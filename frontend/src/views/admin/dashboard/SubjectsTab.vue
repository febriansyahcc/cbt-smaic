<template>
  <!-- TAB: MATA PELAJARAN -->
  <div v-if="activeTab === 'subjects'" class="space-y-4">
    <!-- Contextual Cards: Subjects -->
    <div class="grid grid-cols-2 sm:grid-cols-4 gap-3">
      <div class="bg-white rounded-3xl p-4 border border-slate-200 shadow-xs">
        <span class="text-[11px] font-semibold text-slate-500">Total Mata Pelajaran</span>
        <div class="text-2xl font-black text-slate-900 mt-1">{{ subjects.length }} <span class="text-xs font-medium text-slate-400">Mapel</span></div>
      </div>
      <div class="bg-white rounded-3xl p-4 border border-cyan-200 shadow-xs bg-cyan-50/20">
        <span class="text-[11px] font-semibold text-cyan-700">Bank Soal Terdaftar</span>
        <div class="text-2xl font-black text-cyan-700 mt-1">{{ subjectsWithBanksCount }} <span class="text-xs font-medium text-cyan-400">Mapel</span></div>
      </div>
      <div class="bg-white rounded-3xl p-4 border border-emerald-200 shadow-xs bg-emerald-50/20">
        <span class="text-[11px] font-semibold text-emerald-700">Alokasi Kelas Mapel</span>
        <div class="text-2xl font-black text-emerald-700 mt-1">{{ classSubjects.length }} <span class="text-xs font-medium text-emerald-400">Alokasi</span></div>
      </div>
      <div class="bg-white rounded-3xl p-4 border border-slate-200 shadow-xs">
        <span class="text-[11px] font-semibold text-slate-500">Kurikulum Aktif</span>
        <div class="text-2xl font-black text-indigo-600 mt-1">2026/2027</div>
      </div>
    </div>

    <!-- Subject Action & Filter Toolbar -->
    <div class="space-y-3">
      <div class="flex flex-col md:flex-row md:items-center justify-between gap-3 bg-white p-3.5 rounded-3xl border border-slate-200 shadow-xs">
        <!-- Search Box -->
        <div class="relative min-w-[200px] max-w-xs flex-1">
          <span class="absolute inset-y-0 left-0 flex items-center pl-3 pointer-events-none text-slate-400">
            <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
            </svg>
          </span>
          <input
            v-model="subjectSearchQuery"
            type="text"
            placeholder="Cari kode atau nama mapel..."
            class="w-full pl-9 pr-8 py-2 bg-slate-50 border border-slate-200 rounded-2xl text-xs text-slate-800 placeholder-slate-400 focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:bg-white transition"
          />
          <button
            v-if="subjectSearchQuery"
            @click="subjectSearchQuery = ''"
            class="absolute inset-y-0 right-0 pr-2.5 flex items-center text-slate-400 hover:text-slate-600 cursor-pointer"
          >
            ✕
          </button>
        </div>

        <!-- Action Buttons -->
        <div class="flex items-center justify-end gap-2">
          <button
            v-if="canManageMaster"
            @click="showImportSubjectModal = true"
            class="px-3.5 py-2 bg-emerald-600 hover:bg-emerald-700 active:scale-95 text-white font-bold text-xs rounded-xl shadow-xs transition flex items-center space-x-1 cursor-pointer shrink-0"
          >
            <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-8l-4-4m0 0L8 8m4-4v12" />
            </svg>
            <span>Impor Excel</span>
          </button>
          <button
            @click="openCreateSubject"
            class="px-3.5 py-2 bg-indigo-600 hover:bg-indigo-700 active:scale-95 text-white font-bold text-xs rounded-xl shadow-xs transition flex items-center space-x-1 cursor-pointer shrink-0"
          >
            <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
            </svg>
            <span>Tambah Mapel</span>
          </button>
        </div>
      </div>
    </div>

    <!-- Subjects Table -->
    <div class="bg-white rounded-3xl border border-slate-200 shadow-sm overflow-hidden">
      <div class="overflow-x-auto">
        <table class="w-full text-left text-xs">
          <thead>
            <tr class="bg-slate-50 border-b border-slate-200 text-slate-600 font-bold uppercase text-[10px] select-none">
              <th class="py-3 px-4">
                <button
                  type="button"
                  @click="toggleSubjectSort('code')"
                  class="group inline-flex items-center gap-1 cursor-pointer transition hover:text-indigo-600 uppercase font-bold"
                  :class="subjectSortKey === 'code' ? 'text-indigo-600' : 'text-slate-600'"
                  title="Urutkan berdasarkan Kode Mapel"
                >
                  <span>Kode Mapel</span>
                  <svg v-if="subjectSortKey === 'code' && subjectSortOrder === 'asc'" class="w-3 h-3 text-indigo-600" viewBox="0 0 20 20" fill="currentColor"><path fill-rule="evenodd" d="M5.293 9.707a1 1 0 010-1.414l4-4a1 1 0 011.414 0l4 4a1 1 0 01-1.414 1.414L11 7.414V15a1 1 0 11-2 0V7.414L6.707 9.707a1 1 0 01-1.414 0z" clip-rule="evenodd" /></svg>
                  <svg v-else-if="subjectSortKey === 'code' && subjectSortOrder === 'desc'" class="w-3 h-3 text-indigo-600" viewBox="0 0 20 20" fill="currentColor"><path fill-rule="evenodd" d="M14.707 10.293a1 1 0 010 1.414l-4 4a1 1 0 01-1.414 0l-4-4a1 1 0 111.414-1.414L9 12.586V5a1 1 0 012 0v7.586l2.293-2.293a1 1 0 011.414 0z" clip-rule="evenodd" /></svg>
                  <svg v-else class="w-3 h-3 text-slate-300 group-hover:text-slate-400" viewBox="0 0 20 20" fill="currentColor"><path d="M5 8l5-5 5 5H5zm10 4l-5 5-5-5h10z" /></svg>
                </button>
              </th>
              <th class="py-3 px-4">
                <button
                  type="button"
                  @click="toggleSubjectSort('name')"
                  class="group inline-flex items-center gap-1 cursor-pointer transition hover:text-indigo-600 uppercase font-bold"
                  :class="subjectSortKey === 'name' ? 'text-indigo-600' : 'text-slate-600'"
                  title="Urutkan berdasarkan Nama Mapel"
                >
                  <span>Nama Mata Pelajaran</span>
                  <svg v-if="subjectSortKey === 'name' && subjectSortOrder === 'asc'" class="w-3 h-3 text-indigo-600" viewBox="0 0 20 20" fill="currentColor"><path fill-rule="evenodd" d="M5.293 9.707a1 1 0 010-1.414l4-4a1 1 0 011.414 0l4 4a1 1 0 01-1.414 1.414L11 7.414V15a1 1 0 11-2 0V7.414L6.707 9.707a1 1 0 01-1.414 0z" clip-rule="evenodd" /></svg>
                  <svg v-else-if="subjectSortKey === 'name' && subjectSortOrder === 'desc'" class="w-3 h-3 text-indigo-600" viewBox="0 0 20 20" fill="currentColor"><path fill-rule="evenodd" d="M14.707 10.293a1 1 0 010 1.414l-4 4a1 1 0 01-1.414 0l-4-4a1 1 0 111.414-1.414L9 12.586V5a1 1 0 012 0v7.586l2.293-2.293a1 1 0 011.414 0z" clip-rule="evenodd" /></svg>
                  <svg v-else class="w-3 h-3 text-slate-300 group-hover:text-slate-400" viewBox="0 0 20 20" fill="currentColor"><path d="M5 8l5-5 5 5H5zm10 4l-5 5-5-5h10z" /></svg>
                </button>
              </th>
              <th class="py-3 px-4 text-center">Bank Soal</th>
              <th class="py-3 px-4 text-center">Alokasi Kelas</th>
              <th class="py-3 px-4 text-right">Aksi</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-slate-100">
            <tr v-if="filteredSubjects.length === 0">
              <td colspan="5" class="py-8 text-center text-slate-400">
                <div v-if="subjects.length === 0">Belum ada mata pelajaran terdaftar. Silakan klik Tambah Mapel.</div>
                <div v-else class="space-y-1.5">
                  <div class="font-bold text-slate-700 text-xs">Tidak ada mata pelajaran yang sesuai filter</div>
                  <button
                    @click="subjectSearchQuery = ''; subjectCurrentPage = 1"
                    class="px-3 py-1 bg-slate-100 hover:bg-slate-200 text-slate-700 font-bold text-xs rounded-xl transition cursor-pointer mt-1"
                  >
                    Reset Filter
                  </button>
                </div>
              </td>
            </tr>
            <tr v-for="s in paginatedSubjects" :key="s.id" class="hover:bg-slate-50/60 transition">
              <td class="py-3 px-4 font-mono font-bold text-indigo-700">{{ s.code }}</td>
              <td class="py-3 px-4 font-bold text-slate-900">{{ s.name }}</td>
              <td class="py-3 px-4 text-center">
                <span class="px-2 py-0.5 rounded-full text-[10px] font-bold bg-slate-100 text-slate-700">
                  {{ getSubjectBankCount(s.id) }} Paket
                </span>
              </td>
              <td class="py-3 px-4 text-center">
                <span class="px-2 py-0.5 rounded-full text-[10px] font-bold bg-teal-50 text-teal-700 border border-teal-100">
                  {{ getSubjectClassCount(s.id) }} Kelas
                </span>
              </td>
              <td class="py-3 px-4 text-right">
                <div class="flex items-center justify-end gap-1.5">
                  <button
                    @click="openEditSubject(s)"
                    class="p-1.5 bg-slate-100 hover:bg-slate-200 text-slate-600 rounded-xl transition cursor-pointer"
                    title="Ubah Mapel"
                  >
                    <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z" />
                    </svg>
                  </button>
                  <button
                    @click="deleteSubject(s)"
                    class="p-1.5 bg-rose-50 hover:bg-rose-100 text-rose-600 rounded-xl transition cursor-pointer"
                    title="Hapus Mapel"
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

      <!-- Pagination Controls Footer -->
      <div v-if="filteredSubjects.length > 0" class="border-t border-slate-200/80 bg-slate-50/60 px-4 py-3 flex flex-col sm:flex-row items-center justify-between gap-3 text-xs text-slate-600 select-none">
        <div class="flex items-center gap-3 flex-wrap justify-center sm:justify-start">
          <span>
            Menampilkan
            <span class="font-bold text-slate-800">{{ (subjectCurrentPage - 1) * subjectPerPage + 1 }}</span>
            –
            <span class="font-bold text-slate-800">{{ Math.min(subjectCurrentPage * subjectPerPage, filteredSubjects.length) }}</span>
            dari
            <span class="font-bold text-slate-800">{{ filteredSubjects.length }}</span> mapel
          </span>

          <div class="flex items-center gap-1.5 pl-2.5 border-l border-slate-200">
            <span class="text-slate-400 text-[11px]">Tampilkan:</span>
            <select
              v-model.number="subjectPerPage"
              class="bg-white border border-slate-200 rounded-lg px-2 py-1 text-xs font-semibold text-slate-700 focus:outline-none focus:ring-1 focus:ring-indigo-500 cursor-pointer shadow-2xs"
            >
              <option :value="5">5 / hal</option>
              <option :value="10">10 / hal</option>
              <option :value="25">25 / hal</option>
              <option :value="50">50 / hal</option>
            </select>
          </div>
        </div>

        <!-- Page Navigation -->
        <div v-if="totalSubjectPages > 1" class="flex items-center gap-1">
          <button
            type="button"
            @click="setSubjectPage(subjectCurrentPage - 1)"
            :disabled="subjectCurrentPage === 1"
            class="px-2.5 py-1 rounded-lg border border-slate-200 bg-white hover:bg-slate-100 disabled:opacity-40 disabled:cursor-not-allowed transition font-medium text-xs flex items-center gap-1 cursor-pointer shadow-2xs"
            title="Halaman Sebelumnya"
          >
            <svg class="w-3.5 h-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 19l-7-7 7-7" />
            </svg>
            <span class="hidden sm:inline">Sebelumnya</span>
          </button>

          <div class="flex items-center gap-1">
            <button
              v-for="(p, idx) in displayedSubjectPages"
              :key="idx"
              type="button"
              @click="setSubjectPage(p)"
              :disabled="p === '...'"
              :class="[
                'min-w-[28px] h-7 px-2 rounded-lg text-xs font-bold transition flex items-center justify-center',
                p === '...' ? 'cursor-default text-slate-400' :
                p === subjectCurrentPage ? 'bg-indigo-600 text-white shadow-xs' : 'bg-white border border-slate-200 text-slate-700 hover:bg-slate-100 cursor-pointer shadow-2xs'
              ]"
            >
              {{ p }}
            </button>
          </div>

          <button
            type="button"
            @click="setSubjectPage(subjectCurrentPage + 1)"
            :disabled="subjectCurrentPage === totalSubjectPages"
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
  canManageMaster,
  classSubjects,
  deleteSubject,
  displayedSubjectPages,
  filteredSubjects,
  getSubjectBankCount,
  getSubjectClassCount,
  openCreateSubject,
  openEditSubject,
  paginatedSubjects,
  setSubjectPage,
  showImportSubjectModal,
  subjectCurrentPage,
  subjectPerPage,
  subjectSearchQuery,
  subjectSortKey,
  subjectSortOrder,
  subjects,
  subjectsWithBanksCount,
  toggleSubjectSort,
  totalSubjectPages,
} = useDashboard()
</script>
