<template>
    <!-- TAB 1: JADWAL & TOKEN UJIAN -->
  <div v-if="activeTab === 'schedules'" class="space-y-4">
    <!-- Prerequisite Banner if no active event -->
    <div v-if="!activeExamEvent && canManageSchedules" class="p-4 bg-amber-50 border border-amber-200 rounded-3xl flex flex-col sm:flex-row sm:items-center justify-between gap-3 text-amber-900">
      <div class="flex items-center space-x-3">
        <div class="w-9 h-9 rounded-2xl bg-amber-200 text-amber-900 flex items-center justify-center text-base shrink-0 font-bold">
          ⚠️
        </div>
        <div>
          <div class="font-bold text-xs">Belum Ada Event Ujian yang Aktif</div>
          <div class="text-[11px] text-amber-700">Disarankan untuk menetapkan dan mengaktifkan Event Ujian terlebih dahulu agar setiap sesi jadwal terorganisasi dengan baik.</div>
        </div>
      </div>
      <button @click="switchTab('events')" class="px-4 py-2 bg-amber-600 hover:bg-amber-700 active:scale-95 text-white font-bold text-xs rounded-2xl shrink-0 transition">
        Atur Event Ujian
      </button>
    </div>

    <!-- Contextual Cards: Schedules -->
    <div class="grid grid-cols-2 sm:grid-cols-4 gap-3">
      <div class="bg-white rounded-3xl p-4 border border-slate-200 shadow-xs">
        <span class="text-[11px] font-semibold text-slate-500">Total Sesi Jadwal</span>
        <div class="text-2xl font-black text-slate-900 mt-1">{{ totalSchedulesCount }} <span class="text-xs font-medium text-slate-400">Sesi</span></div>
      </div>
      <div class="bg-white rounded-3xl p-4 border border-emerald-200 shadow-xs bg-emerald-50/20">
        <span class="text-[11px] font-semibold text-emerald-700">Jadwal Aktif</span>
        <div class="text-2xl font-black text-emerald-700 mt-1">{{ activeSchedulesCount }} <span class="text-xs font-medium text-emerald-500">Aktif</span></div>
      </div>
      <div class="bg-white rounded-3xl p-4 border border-slate-200 shadow-xs">
        <span class="text-[11px] font-semibold text-slate-500">Rombel Terjadwal</span>
        <div class="text-2xl font-black text-indigo-600 mt-1">{{ scheduledClassesCount }} <span class="text-xs font-medium text-slate-400">Kelas</span></div>
      </div>
      <div class="bg-white rounded-3xl p-4 border border-slate-200 shadow-xs">
        <span class="text-[11px] font-semibold text-slate-500">Event Terpilih</span>
        <div class="text-sm font-bold text-slate-800 mt-2 truncate">{{ selectedEventFilterName }}</div>
      </div>
    </div>

    <!-- Schedule Action & Accessibility Toolbar -->
    <div class="space-y-3">
      <div class="flex flex-col md:flex-row md:items-center justify-between gap-3 bg-white p-3.5 rounded-3xl border border-slate-200 shadow-xs">
        <!-- Search & Grade Filters -->
        <div class="flex flex-wrap items-center gap-2.5 flex-1">
          <!-- Search Box -->
          <div class="relative min-w-[200px] max-w-xs flex-1">
            <span class="absolute inset-y-0 left-0 flex items-center pl-3 pointer-events-none text-slate-400">
              <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
              </svg>
            </span>
            <input
              v-model="scheduleSearchQuery"
              type="text"
              placeholder="Cari mapel, kelas, token..."
              class="w-full pl-9 pr-8 py-2 bg-slate-50 border border-slate-200 rounded-2xl text-xs text-slate-800 placeholder-slate-400 focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:bg-white transition"
            />
            <button
              v-if="scheduleSearchQuery"
              @click="scheduleSearchQuery = ''"
              class="absolute inset-y-0 right-0 pr-2.5 flex items-center text-slate-400 hover:text-slate-600 cursor-pointer"
            >
              ✕
            </button>
          </div>

          <!-- Grade Filter Pills -->
          <div class="flex items-center gap-1.5 flex-wrap">
            <button
              type="button"
              @click="selectedScheduleGrade = 'all'"
              :class="[
                'px-3 py-1.5 rounded-xl text-xs font-bold transition cursor-pointer',
                selectedScheduleGrade === 'all'
                  ? 'bg-indigo-600 text-white shadow-xs'
                  : 'bg-slate-100 text-slate-600 hover:bg-slate-200'
              ]"
            >
              Semua Kelas
            </button>
            <button
              v-for="gr in availableScheduleGrades"
              :key="gr"
              type="button"
              @click="selectedScheduleGrade = gr"
              :class="[
                'px-3 py-1.5 rounded-xl text-xs font-bold transition cursor-pointer',
                selectedScheduleGrade === gr
                  ? 'bg-indigo-600 text-white shadow-xs'
                  : 'bg-slate-100 text-slate-600 hover:bg-slate-200'
              ]"
            >
              Kelas {{ gr }}
            </button>
          </div>

          <!-- Time Filter Pills (Hari Ini / Mendatang / Semua) -->
          <div class="flex items-center gap-1.5 flex-wrap pl-2 border-l border-slate-200">
            <button
              type="button"
              @click="selectedScheduleTimeFilter = 'all'"
              :class="[
                'px-2.5 py-1.5 rounded-xl text-xs font-semibold transition cursor-pointer',
                selectedScheduleTimeFilter === 'all'
                  ? 'bg-slate-800 text-white shadow-xs'
                  : 'bg-slate-100 text-slate-600 hover:bg-slate-200'
              ]"
            >
              Semua Waktu
            </button>
            <button
              type="button"
              @click="selectedScheduleTimeFilter = 'today'"
              :class="[
                'px-2.5 py-1.5 rounded-xl text-xs font-semibold transition cursor-pointer',
                selectedScheduleTimeFilter === 'today'
                  ? 'bg-emerald-600 text-white shadow-xs'
                  : 'bg-emerald-50 text-emerald-700 hover:bg-emerald-100 border border-emerald-200'
              ]"
            >
              Hari Ini
            </button>
            <button
              type="button"
              @click="selectedScheduleTimeFilter = 'upcoming'"
              :class="[
                'px-2.5 py-1.5 rounded-xl text-xs font-semibold transition cursor-pointer',
                selectedScheduleTimeFilter === 'upcoming'
                  ? 'bg-indigo-600 text-white shadow-xs'
                  : 'bg-slate-100 text-slate-600 hover:bg-slate-200'
              ]"
            >
              Mendatang
            </button>
          </div>
        </div>

        <!-- Action Button (hanya pemegang schedules:manage) -->
        <div v-if="canManageSchedules" class="flex items-center justify-end">
          <button
            @click="openCreateSchedule()"
            class="px-4 py-2 bg-indigo-600 hover:bg-indigo-700 active:scale-95 text-white font-bold text-xs rounded-xl shadow-xs transition flex items-center space-x-1.5 cursor-pointer shrink-0"
          >
            <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
            </svg>
            <span>Buat Jadwal Ujian Baru</span>
          </button>
        </div>
      </div>
    </div>

    <!-- Floating Contextual Bulk Action Bar -->
    <div v-if="canManageSchedules && selectedScheduleIds.length > 0" class="p-3 bg-indigo-50 border-2 border-indigo-300 rounded-2xl flex flex-col sm:flex-row sm:items-center justify-between gap-3 text-xs shadow-md">
      <div class="flex items-center gap-2.5">
        <span class="w-6 h-6 rounded-full bg-indigo-600 text-white font-black flex items-center justify-center text-xs">
          {{ selectedScheduleIds.length }}
        </span>
        <span class="font-bold text-slate-900">Sesi Jadwal Terpilih</span>
      </div>

      <div class="flex items-center gap-2 flex-wrap">
        <button
          @click="bulkActivateSchedules()"
          class="px-3 py-1.5 bg-emerald-600 hover:bg-emerald-700 active:scale-95 text-white font-bold rounded-xl transition flex items-center gap-1.5 shadow-xs cursor-pointer"
        >
          <span>Aktifkan Terpilih</span>
        </button>
        <button
          @click="bulkDeactivateSchedules()"
          class="px-3 py-1.5 bg-slate-700 hover:bg-slate-800 active:scale-95 text-white font-bold rounded-xl transition flex items-center gap-1.5 shadow-xs cursor-pointer"
        >
          <span>Nonaktifkan</span>
        </button>
        <button
          @click="bulkRegenerateTokens()"
          class="px-3 py-1.5 bg-white border border-slate-300 hover:bg-slate-100 active:scale-95 text-slate-700 font-bold rounded-xl transition flex items-center gap-1.5 shadow-2xs cursor-pointer"
        >
          <span>Samakan Token</span>
        </button>
        <button
          @click="selectedScheduleIds = []"
          class="px-2.5 py-1.5 text-slate-500 hover:text-slate-800 active:scale-95 font-semibold transition cursor-pointer"
        >
          Batal
        </button>
      </div>
    </div>

    <div class="bg-white rounded-3xl border border-slate-200 shadow-sm overflow-hidden">
      <div class="overflow-x-auto">
        <table class="w-full text-left text-xs">
          <thead>
            <tr class="bg-slate-50 border-b border-slate-200 text-slate-600 font-bold uppercase text-[10px] select-none">
              <!-- Col 0: Checkbox Seleksi -->
              <th v-if="canManageSchedules" class="py-3 pl-4 pr-1 w-8 text-center">
                <input
                  type="checkbox"
                  :checked="isAllPaginatedSelected"
                  @change="toggleSelectAllPaginatedSchedules"
                  class="w-3.5 h-3.5 rounded text-indigo-600 border-slate-300 focus:ring-indigo-500 cursor-pointer"
                  title="Pilih Semua di Halaman Ini"
                />
              </th>

              <!-- Col 1: Sesi & Waktu -->
              <th class="py-3 px-4">
                <div class="flex items-center gap-2">
                  <button
                    type="button"
                    @click="toggleScheduleSort('title')"
                    class="group inline-flex items-center gap-1 cursor-pointer transition hover:text-indigo-600"
                    :class="scheduleSortKey === 'title' ? 'text-indigo-600' : 'text-slate-600'"
                    title="Urutkan berdasarkan Nama Sesi"
                  >
                    <span>Nama Sesi</span>
                    <svg v-if="scheduleSortKey === 'title' && scheduleSortOrder === 'asc'" class="w-3 h-3 text-indigo-600" viewBox="0 0 20 20" fill="currentColor"><path fill-rule="evenodd" d="M5.293 9.707a1 1 0 010-1.414l4-4a1 1 0 011.414 0l4 4a1 1 0 01-1.414 1.414L11 7.414V15a1 1 0 11-2 0V7.414L6.707 9.707a1 1 0 01-1.414 0z" clip-rule="evenodd" /></svg>
                    <svg v-else-if="scheduleSortKey === 'title' && scheduleSortOrder === 'desc'" class="w-3 h-3 text-indigo-600" viewBox="0 0 20 20" fill="currentColor"><path fill-rule="evenodd" d="M14.707 10.293a1 1 0 010 1.414l-4 4a1 1 0 01-1.414 0l-4-4a1 1 0 111.414-1.414L9 12.586V5a1 1 0 012 0v7.586l2.293-2.293a1 1 0 011.414 0z" clip-rule="evenodd" /></svg>
                    <svg v-else class="w-3 h-3 text-slate-300 group-hover:text-slate-400" viewBox="0 0 20 20" fill="currentColor"><path d="M5 8l5-5 5 5H5zm10 4l-5 5-5-5h10z" /></svg>
                  </button>
                  <span class="text-slate-300 font-normal">/</span>
                  <button
                    type="button"
                    @click="toggleScheduleSort('time')"
                    class="group inline-flex items-center gap-1 cursor-pointer transition hover:text-indigo-600"
                    :class="scheduleSortKey === 'time' ? 'text-indigo-600' : 'text-slate-600'"
                    title="Urutkan berdasarkan Waktu Pelaksanaan"
                  >
                    <span>Waktu</span>
                    <svg v-if="scheduleSortKey === 'time' && scheduleSortOrder === 'asc'" class="w-3 h-3 text-indigo-600" viewBox="0 0 20 20" fill="currentColor"><path fill-rule="evenodd" d="M5.293 9.707a1 1 0 010-1.414l4-4a1 1 0 011.414 0l4 4a1 1 0 01-1.414 1.414L11 7.414V15a1 1 0 11-2 0V7.414L6.707 9.707a1 1 0 01-1.414 0z" clip-rule="evenodd" /></svg>
                    <svg v-else-if="scheduleSortKey === 'time' && scheduleSortOrder === 'desc'" class="w-3 h-3 text-indigo-600" viewBox="0 0 20 20" fill="currentColor"><path fill-rule="evenodd" d="M14.707 10.293a1 1 0 010 1.414l-4 4a1 1 0 01-1.414 0l-4-4a1 1 0 111.414-1.414L9 12.586V5a1 1 0 012 0v7.586l2.293-2.293a1 1 0 011.414 0z" clip-rule="evenodd" /></svg>
                    <svg v-else class="w-3 h-3 text-slate-300 group-hover:text-slate-400" viewBox="0 0 20 20" fill="currentColor"><path d="M5 8l5-5 5 5H5zm10 4l-5 5-5-5h10z" /></svg>
                  </button>
                </div>
              </th>

              <!-- Col 2: Kelas -->
              <th class="py-3 px-4">
                <button
                  type="button"
                  @click="toggleScheduleSort('class')"
                  class="group inline-flex items-center gap-1 cursor-pointer transition hover:text-indigo-600 font-bold uppercase"
                  :class="scheduleSortKey === 'class' ? 'text-indigo-600' : 'text-slate-600'"
                  title="Urutkan berdasarkan Kelas"
                >
                  <span>Kelas</span>
                  <svg v-if="scheduleSortKey === 'class' && scheduleSortOrder === 'asc'" class="w-3 h-3 text-indigo-600" viewBox="0 0 20 20" fill="currentColor"><path fill-rule="evenodd" d="M5.293 9.707a1 1 0 010-1.414l4-4a1 1 0 011.414 0l4 4a1 1 0 01-1.414 1.414L11 7.414V15a1 1 0 11-2 0V7.414L6.707 9.707a1 1 0 01-1.414 0z" clip-rule="evenodd" /></svg>
                  <svg v-else-if="scheduleSortKey === 'class' && scheduleSortOrder === 'desc'" class="w-3 h-3 text-indigo-600" viewBox="0 0 20 20" fill="currentColor"><path fill-rule="evenodd" d="M14.707 10.293a1 1 0 010 1.414l-4 4a1 1 0 01-1.414 0l-4-4a1 1 0 111.414-1.414L9 12.586V5a1 1 0 012 0v7.586l2.293-2.293a1 1 0 011.414 0z" clip-rule="evenodd" /></svg>
                  <svg v-else class="w-3 h-3 text-slate-300 group-hover:text-slate-400" viewBox="0 0 20 20" fill="currentColor"><path d="M5 8l5-5 5 5H5zm10 4l-5 5-5-5h10z" /></svg>
                </button>
              </th>

              <!-- Col 3: Catatan / Status Soal -->
              <th class="py-3 px-4">
                <button
                  type="button"
                  @click="toggleScheduleSort('subject')"
                  class="group inline-flex items-center gap-1 cursor-pointer transition hover:text-indigo-600 font-bold uppercase"
                  :class="scheduleSortKey === 'subject' ? 'text-indigo-600' : 'text-slate-600'"
                  title="Urutkan berdasarkan Mata Pelajaran & Soal"
                >
                  <span>Catatan / Status Soal</span>
                  <svg v-if="scheduleSortKey === 'subject' && scheduleSortOrder === 'asc'" class="w-3 h-3 text-indigo-600" viewBox="0 0 20 20" fill="currentColor"><path fill-rule="evenodd" d="M5.293 9.707a1 1 0 010-1.414l4-4a1 1 0 011.414 0l4 4a1 1 0 01-1.414 1.414L11 7.414V15a1 1 0 11-2 0V7.414L6.707 9.707a1 1 0 01-1.414 0z" clip-rule="evenodd" /></svg>
                  <svg v-else-if="scheduleSortKey === 'subject' && scheduleSortOrder === 'desc'" class="w-3 h-3 text-indigo-600" viewBox="0 0 20 20" fill="currentColor"><path fill-rule="evenodd" d="M14.707 10.293a1 1 0 010 1.414l-4 4a1 1 0 01-1.414 0l-4-4a1 1 0 111.414-1.414L9 12.586V5a1 1 0 012 0v7.586l2.293-2.293a1 1 0 011.414 0z" clip-rule="evenodd" /></svg>
                  <svg v-else class="w-3 h-3 text-slate-300 group-hover:text-slate-400" viewBox="0 0 20 20" fill="currentColor"><path d="M5 8l5-5 5 5H5zm10 4l-5 5-5-5h10z" /></svg>
                </button>
              </th>

              <!-- Col 4: Token & Status -->
              <th class="py-3 px-4 text-center">
                <button
                  type="button"
                  @click="toggleScheduleSort('status')"
                  class="group inline-flex items-center gap-1 cursor-pointer transition hover:text-indigo-600 font-bold uppercase mx-auto"
                  :class="scheduleSortKey === 'status' ? 'text-indigo-600' : 'text-slate-600'"
                  title="Urutkan berdasarkan Status & Token"
                >
                  <span>Token & Status</span>
                  <svg v-if="scheduleSortKey === 'status' && scheduleSortOrder === 'asc'" class="w-3 h-3 text-indigo-600" viewBox="0 0 20 20" fill="currentColor"><path fill-rule="evenodd" d="M5.293 9.707a1 1 0 010-1.414l4-4a1 1 0 011.414 0l4 4a1 1 0 01-1.414 1.414L11 7.414V15a1 1 0 11-2 0V7.414L6.707 9.707a1 1 0 01-1.414 0z" clip-rule="evenodd" /></svg>
                  <svg v-else-if="scheduleSortKey === 'status' && scheduleSortOrder === 'desc'" class="w-3 h-3 text-indigo-600" viewBox="0 0 20 20" fill="currentColor"><path fill-rule="evenodd" d="M14.707 10.293a1 1 0 010 1.414l-4 4a1 1 0 01-1.414 0l-4-4a1 1 0 111.414-1.414L9 12.586V5a1 1 0 012 0v7.586l2.293-2.293a1 1 0 011.414 0z" clip-rule="evenodd" /></svg>
                  <svg v-else class="w-3 h-3 text-slate-300 group-hover:text-slate-400" viewBox="0 0 20 20" fill="currentColor"><path d="M5 8l5-5 5 5H5zm10 4l-5 5-5-5h10z" /></svg>
                </button>
              </th>

              <!-- Col 5: Aksi -->
              <th class="py-3 px-4 text-right">Aksi</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-slate-100">
            <tr v-if="filteredSchedules.length === 0">
              <td :colspan="canManageSchedules ? 6 : 5" class="py-10 text-center text-slate-400">
                <div v-if="currentEventSchedules.length === 0" class="space-y-1">
                  <template v-if="canManageSchedules">
                    <div class="font-bold text-slate-700 text-xs">Belum ada jadwal ujian untuk event ini.</div>
                    <div class="text-[11px] text-slate-400">Klik tombol "Buat Jadwal Ujian Baru" di atas untuk menambahkan sesi.</div>
                  </template>
                  <div v-else class="font-bold text-slate-700 text-xs">Belum ada jadwal ujian yang ditugaskan kepada Anda.</div>
                </div>
                <div v-else class="space-y-1.5">
                  <div class="font-bold text-slate-700 text-xs">Tidak ada jadwal ujian yang sesuai dengan filter</div>
                  <div class="text-[11px] text-slate-400">Coba sesuaikan kata kunci pencarian, filter jenjang, atau filter waktu.</div>
                  <button
                    @click="scheduleSearchQuery = ''; selectedScheduleGrade = 'all'; selectedScheduleTimeFilter = 'all'; scheduleSortKey = 'time'; scheduleSortOrder = 'asc'; scheduleCurrentPage = 1"
                    class="px-3 py-1 bg-slate-100 hover:bg-slate-200 text-slate-700 font-bold text-xs rounded-xl transition cursor-pointer mt-1"
                  >
                    Reset Filter
                  </button>
                </div>
              </td>
            </tr>
            <tr v-for="sch in paginatedSchedules" :key="sch.id" class="hover:bg-slate-50/60 transition" :class="selectedScheduleIds.includes(sch.id) ? 'bg-indigo-50/30' : ''">
              <!-- Col 0: Checkbox Seleksi (hanya pemegang schedules:manage) -->
              <td v-if="canManageSchedules" class="py-3.5 pl-4 pr-1 text-center">
                <input
                  type="checkbox"
                  :value="sch.id"
                  v-model="selectedScheduleIds"
                  class="w-3.5 h-3.5 rounded text-indigo-600 border-slate-300 focus:ring-indigo-500 cursor-pointer"
                />
              </td>

              <!-- Col 1: Sesi & Waktu (Merged Mapel) -->
              <td class="py-3.5 px-4">
                <div class="flex items-center gap-2 flex-wrap">
                  <span class="font-bold text-slate-900 leading-snug text-xs">{{ sch.title }}</span>
                  <span
                    v-if="sch.subject?.name && !sch.title.toLowerCase().includes(sch.subject.name.toLowerCase())"
                    class="px-2 py-0.5 bg-indigo-50 text-indigo-700 text-[10px] font-semibold rounded-md border border-indigo-100"
                  >
                    {{ sch.subject.name }}
                  </span>
                  <span v-if="sch.is_makeup" class="px-2 py-0.5 bg-orange-100 text-orange-700 rounded-full text-[10px] font-bold">
                    Susulan
                  </span>
                </div>
                <div class="text-[11px] text-slate-500 mt-1 flex items-center gap-2 flex-wrap">
                  <span class="font-medium text-slate-600 bg-slate-100 px-2 py-0.5 rounded-md text-[11px]">
                    {{ formatScheduleTimeRange(sch.start_time, sch.end_time) }} • {{ sch.duration_minutes }} mnt
                  </span>
                </div>
                <div v-if="Array.isArray(sch.proctors)" class="text-[11px] mt-1 leading-snug">
                  <span
                    v-if="sch.proctors.length > 0"
                    class="text-slate-500"
                    :title="sch.proctors.map(p => p.full_name).join(', ')"
                  >Pengawas: {{ formatProctorNames(sch.proctors) }}</span>
                  <span v-else class="text-amber-600">Belum ada pengawas</span>
                </div>
              </td>

              <!-- Col 2: Kelas (Clean without "Tingkat") -->
              <td class="py-3.5 px-4 font-bold text-slate-800 text-xs whitespace-nowrap">
                {{ sch.class_room?.name }}
              </td>

              <!-- Col 3: Catatan / Status Soal -->
              <td class="py-3.5 px-4">
                <div class="flex items-center gap-2 flex-wrap">
                  <button
                    v-if="sch.bank_id && canManageSchedules"
                    type="button"
                    @click="openLinkBankModal(sch)"
                    class="text-xs font-semibold text-indigo-600 hover:text-indigo-800 hover:underline transition cursor-pointer"
                  >
                    Ganti Soal
                  </button>
                  <button
                    v-else-if="!sch.bank_id && canManageSchedules"
                    type="button"
                    @click="openLinkBankModal(sch)"
                    class="text-xs font-semibold text-indigo-600 hover:text-indigo-800 transition cursor-pointer flex items-center gap-1.5"
                  >
                    <span class="px-2 py-0.5 rounded-md bg-amber-50 border border-amber-200 text-amber-700 text-[11px] font-medium">Belum Ada Soal</span>
                    <span class="underline">Tautkan Soal</span>
                  </button>
                  <!-- Tanpa izin schedules:manage: hanya informasi status soal -->
                  <span
                    v-else-if="!sch.bank_id"
                    class="px-2 py-0.5 rounded-md bg-amber-50 border border-amber-200 text-amber-700 text-[11px] font-medium"
                  >Belum Ada Soal</span>
                  <span v-else class="text-xs text-slate-500">Soal tertaut</span>
                </div>
              </td>

              <!-- Col 4: Token & Status -->
              <td class="py-3.5 px-4 text-center">
                <div class="inline-flex flex-col items-center gap-1.5">
                  <!-- Token with Click-to-Copy (server dapat mengosongkan token bagi non-pengelola) -->
                  <div v-if="sch.exam_token" class="flex items-center gap-1">
                    <span class="px-2.5 py-0.5 bg-indigo-50 border border-indigo-200 text-indigo-700 font-mono font-bold text-xs rounded-lg tracking-wider">
                      {{ sch.exam_token }}
                    </span>
                    <button
                      type="button"
                      @click="copyTokenToClipboard(sch.exam_token)"
                      class="p-1 hover:bg-slate-100 text-slate-400 hover:text-indigo-600 rounded-md transition cursor-pointer"
                      title="Salin Kode Token"
                    >
                      <svg class="w-3.5 h-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 16H6a2 2 0 01-2-2V6a2 2 0 012-2h8a2 2 0 012 2v2m-6 12h8a2 2 0 002-2v-8a2 2 0 00-2-2h-8a2 2 0 00-2 2v8a2 2 0 002 2z" />
                      </svg>
                    </button>
                  </div>

                  <button
                    type="button"
                    :disabled="!canManageSchedules"
                    @click="canManageSchedules && toggleScheduleStatus(sch)"
                    :class="[
                      'px-2.5 py-0.5 rounded-full text-[10px] font-bold transition inline-flex items-center gap-1.5',
                      canManageSchedules ? 'cursor-pointer active:scale-95' : 'cursor-default',
                      sch.is_active
                        ? ['bg-emerald-50 text-emerald-700 border border-emerald-200', canManageSchedules ? 'hover:bg-emerald-100' : '']
                        : ['bg-slate-100 text-slate-500 border border-slate-200', canManageSchedules ? 'hover:bg-slate-200' : '']
                    ]"
                    :title="canManageSchedules ? (sch.is_active ? 'Klik untuk nonaktifkan sesi' : 'Klik untuk aktifkan sesi') : (sch.is_active ? 'Sesi aktif' : 'Sesi nonaktif')"
                  >
                    <span class="w-1.5 h-1.5 rounded-full" :class="sch.is_active ? 'bg-emerald-500' : 'bg-slate-400'"></span>
                    <span>{{ sch.is_active ? 'Aktif' : 'Nonaktif' }}</span>
                  </button>
                </div>
              </td>

              <!-- Col 5: Aksi (Live Monitor + Cetak Berkas + Ikon Detail + Edit + Hapus) -->
              <td class="py-3.5 px-4 text-right">
                <div class="flex items-center justify-end gap-1.5">
                  <!-- Live Monitor Button -->
                  <button
                    v-if="authStore.canAccessTab('proctor')"
                    @click="switchTab('proctor', sch.id)"
                    class="px-2.5 py-1.5 bg-indigo-50 hover:bg-indigo-100 text-indigo-700 font-bold text-xs rounded-xl transition flex items-center gap-1.5 cursor-pointer"
                    title="Buka Live Monitoring Ruang Ujian"
                  >
                    <span class="w-1.5 h-1.5 rounded-full bg-emerald-500 animate-ping"></span>
                    <span>Live Monitor</span>
                  </button>

                  <!-- Cetak Berkas Button -->
                  <button
                    v-if="canReadPeople"
                    @click="openPrintModal(sch, 'attendance')"
                    class="p-1.5 bg-slate-100 hover:bg-emerald-50 hover:text-emerald-700 text-slate-600 rounded-xl transition cursor-pointer"
                    title="Cetak Berkas Resmi (Daftar Hadir / Berita Acara)"
                  >
                    <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17 17h2a2 2 0 002-2v-4a2 2 0 00-2-2H5a2 2 0 00-2 2v4a2 2 0 002 2h2m2 4h6a2 2 0 002-2v-4a2 2 0 00-2-2H9a2 2 0 00-2 2v4a2 2 0 002 2zm8-12V5a2 2 0 00-2-2H9a2 2 0 00-2 2v4h10z" />
                    </svg>
                  </button>

                  <!-- Detail Button -->
                  <button
                    @click="openScheduleDetail(sch)"
                    class="p-1.5 bg-slate-100 hover:bg-indigo-50 hover:text-indigo-600 text-slate-600 rounded-xl transition cursor-pointer"
                    title="Rincian Informasi Detail"
                  >
                    <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
                    </svg>
                  </button>

                  <!-- Atur Pengawas Button -->
                  <button
                    v-if="authStore.hasPermission('schedules:manage')"
                    type="button"
                    @click="openProctorAssign(sch)"
                    class="p-1.5 bg-slate-100 hover:bg-indigo-50 hover:text-indigo-600 text-slate-600 rounded-xl transition active:scale-95 cursor-pointer"
                    title="Atur Pengawas"
                  >
                    <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m5.618-4.016A11.955 11.955 0 0112 2.944a11.955 11.955 0 01-8.618 3.04A12.02 12.02 0 003 9c0 5.591 3.824 10.29 9 11.622 5.176-1.332 9-6.03 9-11.622 0-1.042-.133-2.052-.382-3.016z" />
                    </svg>
                  </button>

                  <!-- Edit Button -->
                  <button
                    v-if="canManageSchedules"
                    @click="openEditSchedule(sch)"
                    class="p-1.5 bg-slate-100 hover:bg-slate-200 text-slate-600 rounded-xl transition active:scale-95 cursor-pointer"
                    title="Edit Sesi Jadwal"
                  >
                    <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z" />
                    </svg>
                  </button>

                  <!-- Delete Button -->
                  <button
                    v-if="canManageSchedules"
                    @click="deleteSchedule(sch)"
                    class="p-1.5 bg-rose-50 hover:bg-rose-100 text-rose-600 rounded-xl transition active:scale-95 cursor-pointer"
                    title="Hapus Sesi Jadwal"
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
      <div v-if="filteredSchedules.length > 0" class="border-t border-slate-200/80 bg-slate-50/60 px-4 py-3 flex flex-col sm:flex-row items-center justify-between gap-3 text-xs text-slate-600 select-none">
        <!-- Left: Info & Per-Page Selector -->
        <div class="flex items-center gap-3 flex-wrap justify-center sm:justify-start">
          <span>
            Menampilkan
            <span class="font-bold text-slate-800">{{ (scheduleCurrentPage - 1) * schedulePerPage + 1 }}</span>
            –
            <span class="font-bold text-slate-800">{{ Math.min(scheduleCurrentPage * schedulePerPage, filteredSchedules.length) }}</span>
            dari
            <span class="font-bold text-slate-800">{{ filteredSchedules.length }}</span> sesi
          </span>

          <div class="flex items-center gap-1.5 pl-2.5 border-l border-slate-200">
            <span class="text-slate-400 text-[11px]">Tampilkan:</span>
            <select
              v-model.number="schedulePerPage"
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
        <div v-if="totalSchedulePages > 1" class="flex items-center gap-1">
          <!-- Previous Button -->
          <button
            type="button"
            @click="setSchedulePage(scheduleCurrentPage - 1)"
            :disabled="scheduleCurrentPage === 1"
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
              v-for="(p, idx) in displayedSchedulePages"
              :key="idx"
              type="button"
              @click="setSchedulePage(p)"
              :disabled="p === '...'"
              :class="[
                'min-w-[28px] h-7 px-2 rounded-lg text-xs font-bold transition flex items-center justify-center',
                p === '...' ? 'cursor-default text-slate-400' :
                p === scheduleCurrentPage ? 'bg-indigo-600 text-white shadow-xs' : 'bg-white border border-slate-200 text-slate-700 hover:bg-slate-100 cursor-pointer shadow-2xs'
              ]"
            >
              {{ p }}
            </button>
          </div>

          <!-- Next Button -->
          <button
            type="button"
            @click="setSchedulePage(scheduleCurrentPage + 1)"
            :disabled="scheduleCurrentPage === totalSchedulePages"
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
  activeSchedulesCount,
  activeTab,
  authStore,
  availableScheduleGrades,
  bulkActivateSchedules,
  bulkDeactivateSchedules,
  bulkRegenerateTokens,
  canManageSchedules,
  canReadPeople,
  copyTokenToClipboard,
  currentEventSchedules,
  deleteSchedule,
  displayedSchedulePages,
  filteredSchedules,
  formatProctorNames,
  formatScheduleTimeRange,
  isAllPaginatedSelected,
  openCreateSchedule,
  openEditSchedule,
  openLinkBankModal,
  openPrintModal,
  openProctorAssign,
  openScheduleDetail,
  paginatedSchedules,
  scheduleCurrentPage,
  schedulePerPage,
  scheduleSearchQuery,
  scheduleSortKey,
  scheduleSortOrder,
  scheduledClassesCount,
  selectedEventFilterName,
  selectedScheduleGrade,
  selectedScheduleIds,
  selectedScheduleTimeFilter,
  setSchedulePage,
  switchTab,
  toggleScheduleSort,
  toggleScheduleStatus,
  toggleSelectAllPaginatedSchedules,
  totalSchedulePages,
  totalSchedulesCount,
} = useDashboard()
</script>
