<template>
  <!-- TAB: KELAS MAPEL -->
  <div v-if="activeTab === 'class-subjects'" class="space-y-4">
    <!-- Contextual Cards: Class Subjects -->
    <div class="grid grid-cols-2 sm:grid-cols-4 gap-3">
      <div class="bg-white rounded-3xl p-4 border border-slate-200 shadow-xs">
        <span class="text-[11px] font-semibold text-slate-500">Total Alokasi</span>
        <div class="text-2xl font-black text-slate-900 mt-1">{{ classSubjects.length }} <span class="text-xs font-medium text-slate-400">Mapel Kelas</span></div>
      </div>
      <div class="bg-white rounded-3xl p-4 border border-teal-200 shadow-xs bg-teal-50/20">
        <span class="text-[11px] font-semibold text-teal-700">Kelas Terlayani</span>
        <div class="text-2xl font-black text-teal-700 mt-1">{{ allocatedClassesCount }} <span class="text-xs font-medium text-teal-400">Rombel</span></div>
      </div>
      <div class="bg-white rounded-3xl p-4 border border-indigo-200 shadow-xs bg-indigo-50/20">
        <span class="text-[11px] font-semibold text-indigo-700">Guru Pengampu</span>
        <div class="text-2xl font-black text-indigo-700 mt-1">{{ allocatedTeachersCount }} <span class="text-xs font-medium text-indigo-400">Pendidik</span></div>
      </div>
      <div class="bg-white rounded-3xl p-4 border border-slate-200 shadow-xs">
        <span class="text-[11px] font-semibold text-slate-500">Tahun Ajaran</span>
        <div class="text-2xl font-black text-slate-900 mt-1">{{ activeExamEvent?.academic_year || '2026/2027' }}</div>
      </div>
    </div>

    <!-- Action & Filter Toolbar -->
    <div class="space-y-3">
      <div class="flex flex-col md:flex-row md:items-center justify-between gap-3 bg-white p-3.5 rounded-3xl border border-slate-200 shadow-xs">
        <!-- Search & Filter Controls -->
        <div class="flex flex-wrap items-center gap-2.5 flex-1">
          <!-- Search Box -->
          <div class="relative min-w-[220px] max-w-xs flex-1">
            <span class="absolute inset-y-0 left-0 flex items-center pl-3 pointer-events-none text-slate-400">
              <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
              </svg>
            </span>
            <input
              v-model="classSubjectSearchQuery"
              type="text"
              placeholder="Cari kelas, mapel, guru pengampu..."
              class="w-full pl-9 pr-8 py-2 bg-slate-50 border border-slate-200 rounded-2xl text-xs text-slate-800 placeholder-slate-400 focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:bg-white transition"
            />
            <button
              v-if="classSubjectSearchQuery"
              @click="classSubjectSearchQuery = ''"
              class="absolute inset-y-0 right-0 pr-2.5 flex items-center text-slate-400 hover:text-slate-600 cursor-pointer"
            >
              ✕
            </button>
          </div>

          <!-- Grade Filter Pills -->
          <div class="flex items-center gap-1.5 flex-wrap">
            <button
              type="button"
              @click="selectedClassSubjectGrade = 'all'; selectedClassSubjectClassId = ''; classSubjectCurrentPage = 1"
              :class="[
                'px-3 py-1.5 rounded-xl text-xs font-bold transition cursor-pointer',
                selectedClassSubjectGrade === 'all' && !selectedClassSubjectClassId
                  ? 'bg-indigo-600 text-white shadow-xs'
                  : 'bg-slate-100 text-slate-600 hover:bg-slate-200'
              ]"
            >
              Semua Tingkat
            </button>
            <button
              v-for="gr in ['X', 'XI', 'XII']"
              :key="gr"
              type="button"
              @click="selectedClassSubjectGrade = gr; selectedClassSubjectClassId = ''; classSubjectCurrentPage = 1"
              :class="[
                'px-3 py-1.5 rounded-xl text-xs font-bold transition cursor-pointer',
                selectedClassSubjectGrade === gr
                  ? 'bg-indigo-600 text-white shadow-xs'
                  : 'bg-slate-100 text-slate-600 hover:bg-slate-200'
              ]"
            >
              Kelas {{ gr }}
            </button>
          </div>

          <!-- Specific Class Dropdown Selector -->
          <div class="flex items-center">
            <select
              v-model="selectedClassSubjectClassId"
              @change="classSubjectCurrentPage = 1"
              class="bg-slate-50 border border-slate-200 rounded-xl px-2.5 py-1.5 text-xs font-medium text-slate-700 focus:outline-none focus:ring-2 focus:ring-indigo-500 cursor-pointer"
            >
              <option value="">Semua Rombel</option>
              <option v-for="c in classes" :key="c.id" :value="c.id">{{ c.name }}</option>
            </select>
          </div>
        </div>

        <!-- Action Button -->
        <div class="flex items-center justify-end">
          <button
            @click="openCreateClassSubject"
            class="px-3.5 py-2 bg-indigo-600 hover:bg-indigo-700 active:scale-95 text-white font-bold text-xs rounded-xl shadow-xs transition flex items-center space-x-1 cursor-pointer shrink-0"
          >
            <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
            </svg>
            <span>+ Alokasikan Mapel</span>
          </button>
        </div>
      </div>
    </div>

    <!-- Class Subjects Table -->
    <div class="bg-white rounded-3xl border border-slate-200 shadow-sm overflow-hidden">
      <div class="overflow-x-auto">
        <table class="w-full text-left text-xs">
          <thead>
            <tr class="bg-slate-50 border-b border-slate-200 text-slate-600 font-bold uppercase text-[10px] select-none">
              <!-- Col 1: Kelas / Rombel -->
              <th class="py-3 px-4">
                <button
                  type="button"
                  @click="toggleClassSubjectSort('class')"
                  class="group inline-flex items-center gap-1 cursor-pointer transition hover:text-indigo-600 uppercase font-bold"
                  :class="classSubjectSortKey === 'class' ? 'text-indigo-600' : 'text-slate-600'"
                >
                  <span>Kelas / Rombel</span>
                  <svg v-if="classSubjectSortKey === 'class' && classSubjectSortOrder === 'asc'" class="w-3 h-3 text-indigo-600" viewBox="0 0 20 20" fill="currentColor"><path fill-rule="evenodd" d="M5.293 9.707a1 1 0 010-1.414l4-4a1 1 0 011.414 0l4 4a1 1 0 01-1.414 1.414L11 7.414V15a1 1 0 11-2 0V7.414L6.707 9.707a1 1 0 01-1.414 0z" clip-rule="evenodd" /></svg>
                  <svg v-else-if="classSubjectSortKey === 'class' && classSubjectSortOrder === 'desc'" class="w-3 h-3 text-indigo-600" viewBox="0 0 20 20" fill="currentColor"><path fill-rule="evenodd" d="M14.707 10.293a1 1 0 010 1.414l-4 4a1 1 0 01-1.414 0l-4-4a1 1 0 111.414-1.414L9 12.586V5a1 1 0 012 0v7.586l2.293-2.293a1 1 0 011.414 0z" clip-rule="evenodd" /></svg>
                  <svg v-else class="w-3 h-3 text-slate-300 group-hover:text-slate-400" viewBox="0 0 20 20" fill="currentColor"><path d="M5 8l5-5 5 5H5zm10 4l-5 5-5-5h10z" /></svg>
                </button>
              </th>

              <!-- Col 2: Mata Pelajaran -->
              <th class="py-3 px-4">
                <button
                  type="button"
                  @click="toggleClassSubjectSort('subject')"
                  class="group inline-flex items-center gap-1 cursor-pointer transition hover:text-indigo-600 uppercase font-bold"
                  :class="classSubjectSortKey === 'subject' ? 'text-indigo-600' : 'text-slate-600'"
                >
                  <span>Mata Pelajaran</span>
                  <svg v-if="classSubjectSortKey === 'subject' && classSubjectSortOrder === 'asc'" class="w-3 h-3 text-indigo-600" viewBox="0 0 20 20" fill="currentColor"><path fill-rule="evenodd" d="M5.293 9.707a1 1 0 010-1.414l4-4a1 1 0 011.414 0l4 4a1 1 0 01-1.414 1.414L11 7.414V15a1 1 0 11-2 0V7.414L6.707 9.707a1 1 0 01-1.414 0z" clip-rule="evenodd" /></svg>
                  <svg v-else-if="classSubjectSortKey === 'subject' && classSubjectSortOrder === 'desc'" class="w-3 h-3 text-indigo-600" viewBox="0 0 20 20" fill="currentColor"><path fill-rule="evenodd" d="M14.707 10.293a1 1 0 010 1.414l-4 4a1 1 0 01-1.414 0l-4-4a1 1 0 111.414-1.414L9 12.586V5a1 1 0 012 0v7.586l2.293-2.293a1 1 0 011.414 0z" clip-rule="evenodd" /></svg>
                  <svg v-else class="w-3 h-3 text-slate-300 group-hover:text-slate-400" viewBox="0 0 20 20" fill="currentColor"><path d="M5 8l5-5 5 5H5zm10 4l-5 5-5-5h10z" /></svg>
                </button>
              </th>

              <!-- Col 3: Guru Pengampu -->
              <th class="py-3 px-4">
                <button
                  type="button"
                  @click="toggleClassSubjectSort('teacher')"
                  class="group inline-flex items-center gap-1 cursor-pointer transition hover:text-indigo-600 uppercase font-bold"
                  :class="classSubjectSortKey === 'teacher' ? 'text-indigo-600' : 'text-slate-600'"
                >
                  <span>Guru Pengampu</span>
                  <svg v-if="classSubjectSortKey === 'teacher' && classSubjectSortOrder === 'asc'" class="w-3 h-3 text-indigo-600" viewBox="0 0 20 20" fill="currentColor"><path fill-rule="evenodd" d="M5.293 9.707a1 1 0 010-1.414l4-4a1 1 0 011.414 0l4 4a1 1 0 01-1.414 1.414L11 7.414V15a1 1 0 11-2 0V7.414L6.707 9.707a1 1 0 01-1.414 0z" clip-rule="evenodd" /></svg>
                  <svg v-else-if="classSubjectSortKey === 'teacher' && classSubjectSortOrder === 'desc'" class="w-3 h-3 text-indigo-600" viewBox="0 0 20 20" fill="currentColor"><path fill-rule="evenodd" d="M14.707 10.293a1 1 0 010 1.414l-4 4a1 1 0 01-1.414 0l-4-4a1 1 0 111.414-1.414L9 12.586V5a1 1 0 012 0v7.586l2.293-2.293a1 1 0 011.414 0z" clip-rule="evenodd" /></svg>
                  <svg v-else class="w-3 h-3 text-slate-300 group-hover:text-slate-400" viewBox="0 0 20 20" fill="currentColor"><path d="M5 8l5-5 5 5H5zm10 4l-5 5-5-5h10z" /></svg>
                </button>
              </th>

              <!-- Col 4: Tahun Ajaran -->
              <th class="py-3 px-4 text-center">
                <button
                  type="button"
                  @click="toggleClassSubjectSort('year')"
                  class="group inline-flex items-center gap-1 cursor-pointer transition hover:text-indigo-600 uppercase font-bold mx-auto"
                  :class="classSubjectSortKey === 'year' ? 'text-indigo-600' : 'text-slate-600'"
                >
                  <span>Tahun Ajaran</span>
                  <svg v-if="classSubjectSortKey === 'year' && classSubjectSortOrder === 'asc'" class="w-3 h-3 text-indigo-600" viewBox="0 0 20 20" fill="currentColor"><path fill-rule="evenodd" d="M5.293 9.707a1 1 0 010-1.414l4-4a1 1 0 011.414 0l4 4a1 1 0 01-1.414 1.414L11 7.414V15a1 1 0 11-2 0V7.414L6.707 9.707a1 1 0 01-1.414 0z" clip-rule="evenodd" /></svg>
                  <svg v-else-if="classSubjectSortKey === 'year' && classSubjectSortOrder === 'desc'" class="w-3 h-3 text-indigo-600" viewBox="0 0 20 20" fill="currentColor"><path fill-rule="evenodd" d="M14.707 10.293a1 1 0 010 1.414l-4 4a1 1 0 01-1.414 0l-4-4a1 1 0 111.414-1.414L9 12.586V5a1 1 0 012 0v7.586l2.293-2.293a1 1 0 011.414 0z" clip-rule="evenodd" /></svg>
                  <svg v-else class="w-3 h-3 text-slate-300 group-hover:text-slate-400" viewBox="0 0 20 20" fill="currentColor"><path d="M5 8l5-5 5 5H5zm10 4l-5 5-5-5h10z" /></svg>
                </button>
              </th>

              <!-- Col 5: Aksi -->
              <th class="py-3 px-4 text-right">Aksi</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-slate-100">
            <tr v-if="filteredClassSubjects.length === 0">
              <td colspan="5" class="py-8 text-center text-slate-400">
                <div v-if="classSubjects.length === 0">Belum ada alokasi mata pelajaran untuk kelas.</div>
                <div v-else class="space-y-1.5">
                  <div class="font-bold text-slate-700 text-xs">Tidak ada data alokasi yang sesuai filter</div>
                  <button
                    @click="classSubjectSearchQuery = ''; selectedClassSubjectGrade = 'all'; selectedClassSubjectClassId = ''; classSubjectCurrentPage = 1"
                    class="px-3 py-1 bg-slate-100 hover:bg-slate-200 text-slate-700 font-bold text-xs rounded-xl transition cursor-pointer mt-1"
                  >
                    Reset Filter
                  </button>
                </div>
              </td>
            </tr>
            <tr v-for="cs in paginatedClassSubjects" :key="cs.id" class="hover:bg-slate-50/60 transition">
              <td class="py-3 px-4 font-bold text-indigo-700">{{ cs.class_room?.name }}</td>
              <td class="py-3 px-4">
                <span class="font-bold text-slate-900">{{ cs.subject?.name }}</span>
                <span class="text-slate-400 font-mono text-[10px] ml-1.5">({{ cs.subject?.code }})</span>
              </td>
              <td class="py-3 px-4">
                <span class="font-semibold text-slate-800">{{ cs.teacher?.full_name }}</span>
                <span class="text-slate-400 font-mono text-[10px] block">@{{ cs.teacher?.username }}</span>
              </td>
              <td class="py-3 px-4 text-center">
                <span class="px-2 py-0.5 rounded-full text-[10px] font-semibold bg-slate-100 text-slate-700">
                  {{ cs.academic_year }}
                </span>
              </td>
              <td class="py-3 px-4 text-right">
                <div class="flex items-center justify-end gap-1.5">
                  <button
                    @click="openEditClassSubject(cs)"
                    class="p-1.5 bg-slate-100 hover:bg-slate-200 text-slate-600 rounded-xl transition cursor-pointer"
                    title="Edit Alokasi"
                  >
                    <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z" />
                    </svg>
                  </button>
                  <button
                    @click="deleteClassSubject(cs)"
                    class="p-1.5 bg-rose-50 hover:bg-rose-100 text-rose-600 rounded-xl transition cursor-pointer"
                    title="Hapus Alokasi"
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
      <div v-if="filteredClassSubjects.length > 0" class="border-t border-slate-200/80 bg-slate-50/60 px-4 py-3 flex flex-col sm:flex-row items-center justify-between gap-3 text-xs text-slate-600 select-none">
        <div class="flex items-center gap-3 flex-wrap justify-center sm:justify-start">
          <span>
            Menampilkan
            <span class="font-bold text-slate-800">{{ (classSubjectCurrentPage - 1) * classSubjectPerPage + 1 }}</span>
            –
            <span class="font-bold text-slate-800">{{ Math.min(classSubjectCurrentPage * classSubjectPerPage, filteredClassSubjects.length) }}</span>
            dari
            <span class="font-bold text-slate-800">{{ filteredClassSubjects.length }}</span> alokasi
          </span>

          <div class="flex items-center gap-1.5 pl-2.5 border-l border-slate-200">
            <span class="text-slate-400 text-[11px]">Tampilkan:</span>
            <select
              v-model.number="classSubjectPerPage"
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
        <div v-if="totalClassSubjectPages > 1" class="flex items-center gap-1">
          <button
            type="button"
            @click="setClassSubjectPage(classSubjectCurrentPage - 1)"
            :disabled="classSubjectCurrentPage === 1"
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
              v-for="(p, idx) in displayedClassSubjectPages"
              :key="idx"
              type="button"
              @click="setClassSubjectPage(p)"
              :disabled="p === '...'"
              :class="[
                'min-w-[28px] h-7 px-2 rounded-lg text-xs font-bold transition flex items-center justify-center',
                p === '...' ? 'cursor-default text-slate-400' :
                p === classSubjectCurrentPage ? 'bg-indigo-600 text-white shadow-xs' : 'bg-white border border-slate-200 text-slate-700 hover:bg-slate-100 cursor-pointer shadow-2xs'
              ]"
            >
              {{ p }}
            </button>
          </div>

          <button
            type="button"
            @click="setClassSubjectPage(classSubjectCurrentPage + 1)"
            :disabled="classSubjectCurrentPage === totalClassSubjectPages"
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
  activeExamEvent,
  activeTab,
  allocatedClassesCount,
  allocatedTeachersCount,
  classSubjectCurrentPage,
  classSubjectPerPage,
  classSubjectSearchQuery,
  classSubjectSortKey,
  classSubjectSortOrder,
  classSubjects,
  classes,
  deleteClassSubject,
  displayedClassSubjectPages,
  filteredClassSubjects,
  openCreateClassSubject,
  openEditClassSubject,
  paginatedClassSubjects,
  selectedClassSubjectClassId,
  selectedClassSubjectGrade,
  setClassSubjectPage,
  toggleClassSubjectSort,
  totalClassSubjectPages,
} = useDashboard()
</script>
