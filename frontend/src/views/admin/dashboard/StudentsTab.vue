<template>
  <!-- TAB: DATA MASTER SISWA -->
    <div v-if="activeTab === 'students'" class="space-y-4">
      <!-- Contextual Cards: Students -->
      <div class="grid grid-cols-2 sm:grid-cols-4 gap-3">
        <div class="bg-white rounded-3xl p-4 border border-slate-200 shadow-xs">
          <span class="text-[11px] font-semibold text-slate-500">Total Siswa</span>
          <div class="text-2xl font-black text-slate-900 mt-1">{{ students.length }} <span class="text-xs font-medium text-slate-400">Siswa</span></div>
        </div>
        <div class="bg-white rounded-3xl p-4 border border-blue-200 shadow-xs bg-blue-50/20">
          <span class="text-[11px] font-semibold text-blue-700">Laki-Laki (L)</span>
          <div class="text-2xl font-black text-blue-700 mt-1">{{ maleStudentsCount }} <span class="text-xs font-medium text-blue-400">Siswa</span></div>
        </div>
        <div class="bg-white rounded-3xl p-4 border border-pink-200 shadow-xs bg-pink-50/20">
          <span class="text-[11px] font-semibold text-pink-700">Perempuan (P)</span>
          <div class="text-2xl font-black text-pink-700 mt-1">{{ femaleStudentsCount }} <span class="text-xs font-medium text-pink-400">Siswi</span></div>
        </div>
        <div :class="['rounded-3xl p-4 border shadow-xs transition', lockedStudentsCount > 0 ? 'bg-amber-50 border-amber-300' : 'bg-white border-slate-200']">
          <span :class="['text-[11px] font-semibold', lockedStudentsCount > 0 ? 'text-amber-800 font-bold' : 'text-slate-500']">🔒 Sesi HP Terkunci</span>
          <div :class="['text-2xl font-black mt-1', lockedStudentsCount > 0 ? 'text-amber-800' : 'text-slate-900']">
            {{ lockedStudentsCount }} <span class="text-xs font-medium opacity-75">Perlu Reset</span>
          </div>
        </div>
      </div>

      <!-- Student Action & Filter Toolbar -->
      <div class="space-y-3">
        <div class="flex flex-col md:flex-row md:items-center justify-between gap-3 bg-white p-3.5 rounded-3xl border border-slate-200 shadow-xs">
          <!-- Search & Filter Controls -->
          <div class="flex flex-wrap items-center gap-2.5 flex-1">
            <!-- Search Box -->
            <div class="relative min-w-[200px] max-w-xs flex-1">
              <span class="absolute inset-y-0 left-0 flex items-center pl-3 pointer-events-none text-slate-400">
                <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
                </svg>
              </span>
              <input
                v-model="studentSearchQuery"
                type="text"
                placeholder="Cari nama, username, NIS, kelas..."
                class="w-full pl-9 pr-8 py-2 bg-slate-50 border border-slate-200 rounded-2xl text-xs text-slate-800 placeholder-slate-400 focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:bg-white transition"
              />
              <button
                v-if="studentSearchQuery"
                @click="studentSearchQuery = ''"
                class="absolute inset-y-0 right-0 pr-2.5 flex items-center text-slate-400 hover:text-slate-600 cursor-pointer"
              >
                ✕
              </button>
            </div>

            <!-- Grade Filter Pills -->
            <div class="flex items-center gap-1.5 flex-wrap">
              <button
                type="button"
                @click="selectedStudentGrade = 'all'"
                :class="[
                  'px-3 py-1.5 rounded-xl text-xs font-bold transition cursor-pointer',
                  selectedStudentGrade === 'all'
                    ? 'bg-indigo-600 text-white shadow-xs'
                    : 'bg-slate-100 text-slate-600 hover:bg-slate-200'
                ]"
              >
                Semua Kelas
              </button>
              <button
                v-for="gr in availableStudentGrades"
                :key="gr"
                type="button"
                @click="selectedStudentGrade = gr"
                :class="[
                  'px-3 py-1.5 rounded-xl text-xs font-bold transition cursor-pointer',
                  selectedStudentGrade === gr
                    ? 'bg-indigo-600 text-white shadow-xs'
                    : 'bg-slate-100 text-slate-600 hover:bg-slate-200'
                ]"
              >
                Kelas {{ gr }}
              </button>
            </div>
          </div>

          <!-- Action Buttons -->
          <div class="flex items-center gap-2 flex-wrap justify-end">
            <button
              @click="openCreateStudentModal"
              class="px-4 py-2 bg-indigo-600 hover:bg-indigo-700 active:scale-95 text-white font-bold text-xs rounded-xl shadow-xs transition flex items-center space-x-1.5 cursor-pointer"
            >
              <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
              </svg>
              <span>Tambah Siswa</span>
            </button>
          </div>
        </div>
      </div>

      <!-- Students Table -->
      <div class="bg-white rounded-3xl border border-slate-200 shadow-sm overflow-hidden">
        <div class="overflow-x-auto">
          <table class="w-full text-left text-xs">
            <thead>
              <tr class="bg-slate-50 border-b border-slate-200 text-slate-600 font-bold uppercase text-[10px] select-none">
                <!-- Col: Nama Siswa -->
                <th class="py-3 px-4">
                  <button
                    type="button"
                    @click="toggleStudentSort('name')"
                    class="group inline-flex items-center gap-1 cursor-pointer transition hover:text-indigo-600 uppercase font-bold"
                    :class="studentSortKey === 'name' ? 'text-indigo-600' : 'text-slate-600'"
                    title="Urutkan berdasarkan Nama Siswa"
                  >
                    <span>Nama Siswa</span>
                    <svg v-if="studentSortKey === 'name' && studentSortOrder === 'asc'" class="w-3 h-3 text-indigo-600" viewBox="0 0 20 20" fill="currentColor"><path fill-rule="evenodd" d="M5.293 9.707a1 1 0 010-1.414l4-4a1 1 0 011.414 0l4 4a1 1 0 01-1.414 1.414L11 7.414V15a1 1 0 11-2 0V7.414L6.707 9.707a1 1 0 01-1.414 0z" clip-rule="evenodd" /></svg>
                    <svg v-else-if="studentSortKey === 'name' && studentSortOrder === 'desc'" class="w-3 h-3 text-indigo-600" viewBox="0 0 20 20" fill="currentColor"><path fill-rule="evenodd" d="M14.707 10.293a1 1 0 010 1.414l-4 4a1 1 0 01-1.414 0l-4-4a1 1 0 111.414-1.414L9 12.586V5a1 1 0 012 0v7.586l2.293-2.293a1 1 0 011.414 0z" clip-rule="evenodd" /></svg>
                    <svg v-else class="w-3 h-3 text-slate-300 group-hover:text-slate-400" viewBox="0 0 20 20" fill="currentColor"><path d="M5 8l5-5 5 5H5zm10 4l-5 5-5-5h10z" /></svg>
                  </button>
                </th>

                <!-- Col: Username -->
                <th class="py-3 px-4">
                  <button
                    type="button"
                    @click="toggleStudentSort('username')"
                    class="group inline-flex items-center gap-1 cursor-pointer transition hover:text-indigo-600 uppercase font-bold"
                    :class="studentSortKey === 'username' ? 'text-indigo-600' : 'text-slate-600'"
                    title="Urutkan berdasarkan Username"
                  >
                    <span>Username</span>
                    <svg v-if="studentSortKey === 'username' && studentSortOrder === 'asc'" class="w-3 h-3 text-indigo-600" viewBox="0 0 20 20" fill="currentColor"><path fill-rule="evenodd" d="M5.293 9.707a1 1 0 010-1.414l4-4a1 1 0 011.414 0l4 4a1 1 0 01-1.414 1.414L11 7.414V15a1 1 0 11-2 0V7.414L6.707 9.707a1 1 0 01-1.414 0z" clip-rule="evenodd" /></svg>
                    <svg v-else-if="studentSortKey === 'username' && studentSortOrder === 'desc'" class="w-3 h-3 text-indigo-600" viewBox="0 0 20 20" fill="currentColor"><path fill-rule="evenodd" d="M14.707 10.293a1 1 0 010 1.414l-4 4a1 1 0 01-1.414 0l-4-4a1 1 0 111.414-1.414L9 12.586V5a1 1 0 012 0v7.586l2.293-2.293a1 1 0 011.414 0z" clip-rule="evenodd" /></svg>
                    <svg v-else class="w-3 h-3 text-slate-300 group-hover:text-slate-400" viewBox="0 0 20 20" fill="currentColor"><path d="M5 8l5-5 5 5H5zm10 4l-5 5-5-5h10z" /></svg>
                  </button>
                </th>

                <!-- Col: NIS / NISN -->
                <th class="py-3 px-4">
                  <button
                    type="button"
                    @click="toggleStudentSort('nis')"
                    class="group inline-flex items-center gap-1 cursor-pointer transition hover:text-indigo-600 uppercase font-bold"
                    :class="studentSortKey === 'nis' ? 'text-indigo-600' : 'text-slate-600'"
                    title="Urutkan berdasarkan NIS"
                  >
                    <span>NIS / NISN</span>
                    <svg v-if="studentSortKey === 'nis' && studentSortOrder === 'asc'" class="w-3 h-3 text-indigo-600" viewBox="0 0 20 20" fill="currentColor"><path fill-rule="evenodd" d="M5.293 9.707a1 1 0 010-1.414l4-4a1 1 0 011.414 0l4 4a1 1 0 01-1.414 1.414L11 7.414V15a1 1 0 11-2 0V7.414L6.707 9.707a1 1 0 01-1.414 0z" clip-rule="evenodd" /></svg>
                    <svg v-else-if="studentSortKey === 'nis' && studentSortOrder === 'desc'" class="w-3 h-3 text-indigo-600" viewBox="0 0 20 20" fill="currentColor"><path fill-rule="evenodd" d="M14.707 10.293a1 1 0 010 1.414l-4 4a1 1 0 01-1.414 0l-4-4a1 1 0 111.414-1.414L9 12.586V5a1 1 0 012 0v7.586l2.293-2.293a1 1 0 011.414 0z" clip-rule="evenodd" /></svg>
                    <svg v-else class="w-3 h-3 text-slate-300 group-hover:text-slate-400" viewBox="0 0 20 20" fill="currentColor"><path d="M5 8l5-5 5 5H5zm10 4l-5 5-5-5h10z" /></svg>
                  </button>
                </th>

                <!-- Col: Kelas -->
                <th class="py-3 px-4">
                  <button
                    type="button"
                    @click="toggleStudentSort('class')"
                    class="group inline-flex items-center gap-1 cursor-pointer transition hover:text-indigo-600 uppercase font-bold"
                    :class="studentSortKey === 'class' ? 'text-indigo-600' : 'text-slate-600'"
                    title="Urutkan berdasarkan Kelas"
                  >
                    <span>Kelas</span>
                    <svg v-if="studentSortKey === 'class' && studentSortOrder === 'asc'" class="w-3 h-3 text-indigo-600" viewBox="0 0 20 20" fill="currentColor"><path fill-rule="evenodd" d="M5.293 9.707a1 1 0 010-1.414l4-4a1 1 0 011.414 0l4 4a1 1 0 01-1.414 1.414L11 7.414V15a1 1 0 11-2 0V7.414L6.707 9.707a1 1 0 01-1.414 0z" clip-rule="evenodd" /></svg>
                    <svg v-else-if="studentSortKey === 'class' && studentSortOrder === 'desc'" class="w-3 h-3 text-indigo-600" viewBox="0 0 20 20" fill="currentColor"><path fill-rule="evenodd" d="M14.707 10.293a1 1 0 010 1.414l-4 4a1 1 0 01-1.414 0l-4-4a1 1 0 111.414-1.414L9 12.586V5a1 1 0 012 0v7.586l2.293-2.293a1 1 0 011.414 0z" clip-rule="evenodd" /></svg>
                    <svg v-else class="w-3 h-3 text-slate-300 group-hover:text-slate-400" viewBox="0 0 20 20" fill="currentColor"><path d="M5 8l5-5 5 5H5zm10 4l-5 5-5-5h10z" /></svg>
                  </button>
                </th>

                <!-- Col: Sesi HP -->
                <th class="py-3 px-4 text-center">
                  <button
                    type="button"
                    @click="toggleStudentSort('session')"
                    class="group inline-flex items-center gap-1 cursor-pointer transition hover:text-indigo-600 uppercase font-bold mx-auto"
                    :class="studentSortKey === 'session' ? 'text-indigo-600' : 'text-slate-600'"
                    title="Urutkan berdasarkan Status Kunci Sesi HP"
                  >
                    <span>Sesi HP</span>
                    <svg v-if="studentSortKey === 'session' && studentSortOrder === 'asc'" class="w-3 h-3 text-indigo-600" viewBox="0 0 20 20" fill="currentColor"><path fill-rule="evenodd" d="M5.293 9.707a1 1 0 010-1.414l4-4a1 1 0 011.414 0l4 4a1 1 0 01-1.414 1.414L11 7.414V15a1 1 0 11-2 0V7.414L6.707 9.707a1 1 0 01-1.414 0z" clip-rule="evenodd" /></svg>
                    <svg v-else-if="studentSortKey === 'session' && studentSortOrder === 'desc'" class="w-3 h-3 text-indigo-600" viewBox="0 0 20 20" fill="currentColor"><path fill-rule="evenodd" d="M14.707 10.293a1 1 0 010 1.414l-4 4a1 1 0 01-1.414 0l-4-4a1 1 0 111.414-1.414L9 12.586V5a1 1 0 012 0v7.586l2.293-2.293a1 1 0 011.414 0z" clip-rule="evenodd" /></svg>
                    <svg v-else class="w-3 h-3 text-slate-300 group-hover:text-slate-400" viewBox="0 0 20 20" fill="currentColor"><path d="M5 8l5-5 5 5H5zm10 4l-5 5-5-5h10z" /></svg>
                  </button>
                </th>

                <!-- Col: Aksi -->
                <th class="py-3 px-4 text-right">Aksi</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-slate-100">
              <tr v-if="filteredStudents.length === 0">
                <td colspan="6" class="py-10 text-center text-slate-400">
                  <div v-if="students.length === 0" class="space-y-1">
                    <div class="font-bold text-slate-700 text-xs">Belum ada data siswa terdaftar.</div>
                    <div class="text-[11px] text-slate-400">Klik tombol "+ Tambah Siswa" atau "Impor Excel" di atas untuk menambahkan peserta.</div>
                  </div>
                  <div v-else class="space-y-1.5">
                    <div class="font-bold text-slate-700 text-xs">Tidak ada data siswa yang sesuai dengan filter</div>
                    <div class="text-[11px] text-slate-400">Coba sesuaikan kata kunci pencarian, filter kelas, atau filter status sesi.</div>
                    <button
                      @click="resetStudentFilters"
                      class="px-3 py-1 bg-slate-100 hover:bg-slate-200 text-slate-700 font-bold text-xs rounded-xl transition cursor-pointer mt-1"
                    >
                      Reset Filter
                    </button>
                  </div>
                </td>
              </tr>
              <tr v-for="st in paginatedStudents" :key="st.id" class="hover:bg-slate-50/60 transition">
                <td class="py-3 px-4 font-bold text-slate-900">
                  <div class="flex items-center gap-2">
                    <span>{{ st.user?.full_name }}</span>
                    <span v-if="st.gender === 'P'" class="px-1.5 py-0.5 bg-pink-50 text-pink-700 text-[10px] font-bold rounded-md border border-pink-200" title="Perempuan">P</span>
                    <span v-else-if="st.gender === 'L'" class="px-1.5 py-0.5 bg-blue-50 text-blue-700 text-[10px] font-bold rounded-md border border-blue-200" title="Laki-Laki">L</span>
                  </div>
                </td>
                <td class="py-3 px-4 font-mono text-slate-600 text-xs">{{ st.user?.username }}</td>
                <td class="py-3 px-4">
                  <span class="font-mono font-semibold text-slate-800">{{ st.nis }}</span>
                  <span class="text-slate-400 font-mono text-[10px] ml-1.5" v-if="st.nisn">({{ st.nisn }})</span>
                </td>
                <td class="py-3 px-4 font-semibold text-indigo-700 whitespace-nowrap">{{ st.class_room?.name }}</td>
                <td class="py-3 px-4 text-center">
                  <span v-if="st.user?.session_token" class="px-2.5 py-0.5 rounded-full text-[10px] font-bold bg-amber-100 text-amber-800 border border-amber-200 inline-flex items-center gap-1">
                    🔒 Terkunci
                  </span>
                  <span v-else class="px-2.5 py-0.5 rounded-full text-[10px] font-bold bg-emerald-50 text-emerald-700 border border-emerald-200 inline-flex items-center gap-1">
                    Bebas
                  </span>
                </td>
                <td class="py-3 px-4 text-right">
                  <div class="flex items-center justify-end gap-1.5">
                    <!-- Reset Sesi HP Button (if locked) -->
                    <button
                      v-if="st.user?.session_token"
                      @click="resetStudentSession(st)"
                      class="px-2 py-1 bg-amber-50 hover:bg-amber-100 text-amber-800 font-bold rounded-xl text-[10px] border border-amber-300 transition flex items-center gap-1 cursor-pointer"
                      title="Lepas kunci sesi login HP agar siswa bisa login di perangkat lain"
                    >
                      <span>🔓 Reset</span>
                    </button>

                    <!-- Detail Button -->
                    <button
                      @click="openStudentDetail(st)"
                      class="p-1.5 bg-slate-100 hover:bg-indigo-50 hover:text-indigo-600 text-slate-600 rounded-xl transition cursor-pointer"
                      title="Rincian Detail Siswa"
                    >
                      <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
                      </svg>
                    </button>

                    <!-- Edit Button -->
                    <button
                      @click="openEditStudent(st)"
                      class="p-1.5 bg-slate-100 hover:bg-slate-200 text-slate-600 rounded-xl transition cursor-pointer"
                      title="Edit Data Siswa"
                    >
                      <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z" />
                      </svg>
                    </button>

                    <!-- Delete Button -->
                    <button
                      @click="deleteStudent(st)"
                      class="p-1.5 bg-rose-50 hover:bg-rose-100 text-rose-600 rounded-xl transition cursor-pointer"
                      title="Hapus Data Siswa"
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
        <div v-if="filteredStudents.length > 0" class="border-t border-slate-200/80 bg-slate-50/60 px-4 py-3 flex flex-col sm:flex-row items-center justify-between gap-3 text-xs text-slate-600 select-none">
          <!-- Left: Info & Per-Page Selector -->
          <div class="flex items-center gap-3 flex-wrap justify-center sm:justify-start">
            <span>
              Menampilkan
              <span class="font-bold text-slate-800">{{ (studentCurrentPage - 1) * studentPerPage + 1 }}</span>
              –
              <span class="font-bold text-slate-800">{{ Math.min(studentCurrentPage * studentPerPage, filteredStudents.length) }}</span>
              dari
              <span class="font-bold text-slate-800">{{ filteredStudents.length }}</span> siswa
            </span>

            <div class="flex items-center gap-1.5 pl-2.5 border-l border-slate-200">
              <span class="text-slate-400 text-[11px]">Tampilkan:</span>
              <select
                v-model.number="studentPerPage"
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
          <div v-if="totalStudentPages > 1" class="flex items-center gap-1">
            <!-- Previous Button -->
            <button
              type="button"
              @click="setStudentPage(studentCurrentPage - 1)"
              :disabled="studentCurrentPage === 1"
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
                v-for="(p, idx) in displayedStudentPages"
                :key="idx"
                type="button"
                @click="setStudentPage(p)"
                :disabled="p === '...'"
                :class="[
                  'min-w-[28px] h-7 px-2 rounded-lg text-xs font-bold transition flex items-center justify-center',
                  p === '...' ? 'cursor-default text-slate-400' :
                  p === studentCurrentPage ? 'bg-indigo-600 text-white shadow-xs' : 'bg-white border border-slate-200 text-slate-700 hover:bg-slate-100 cursor-pointer shadow-2xs'
                ]"
              >
                {{ p }}
              </button>
            </div>

            <!-- Next Button -->
            <button
              type="button"
              @click="setStudentPage(studentCurrentPage + 1)"
              :disabled="studentCurrentPage === totalStudentPages"
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
  availableStudentGrades,
  deleteStudent,
  displayedStudentPages,
  femaleStudentsCount,
  filteredStudents,
  lockedStudentsCount,
  maleStudentsCount,
  openCreateStudentModal,
  openEditStudent,
  openStudentDetail,
  paginatedStudents,
  resetStudentFilters,
  resetStudentSession,
  selectedStudentGrade,
  setStudentPage,
  studentCurrentPage,
  studentPerPage,
  studentSearchQuery,
  studentSortKey,
  studentSortOrder,
  students,
  toggleStudentSort,
  totalStudentPages,
} = useDashboard()
</script>
