<template>
  <div class="space-y-4">
    <!-- Top Bar: Session Info & Token -->
    <div class="bg-white rounded-3xl p-5 border border-slate-200 shadow-xs flex flex-col md:flex-row md:items-center justify-between gap-4">
      <div class="space-y-1.5 min-w-0">
        <div class="flex flex-wrap items-center gap-2">
          <span class="px-2.5 py-0.5 bg-indigo-50 text-indigo-700 font-bold text-xs rounded-full border border-indigo-100">
            {{ proctorData?.class_name || 'Semua Kelas' }}
          </span>
          <span class="text-xs font-semibold text-slate-500">
            Durasi: {{ proctorData?.duration_minutes || 90 }} Menit
          </span>
          <span class="px-2.5 py-0.5 bg-emerald-50 text-emerald-700 border border-emerald-200 text-[10px] font-bold rounded-full flex items-center gap-1.5 shadow-2xs">
            <span class="w-1.5 h-1.5 rounded-full bg-emerald-500 animate-pulse"></span>
            <span>LIVE MONITORING</span>
          </span>
        </div>
        <h2 class="text-base sm:text-lg font-black text-slate-900 leading-tight truncate">
          {{ proctorData?.schedule_title || (schedulesList.length === 0 ? 'Belum Ada Jadwal Ujian Aktif' : 'Pilih Sesi Jadwal Ujian') }}
        </h2>
        <p class="text-xs text-slate-500 flex items-center gap-2 flex-wrap">
          <span>Mata Pelajaran: <strong class="text-slate-700">{{ proctorData?.subject_name || '-' }}</strong></span>
        </p>
      </div>

      <!-- Sesi Selector & Live Token Cluster -->
      <div class="flex items-center gap-3 shrink-0 flex-wrap sm:flex-nowrap">
        <!-- Schedule Dropdown Selector -->
        <div v-if="schedulesList.length > 0" class="flex items-center gap-2 bg-slate-50 border border-slate-200 rounded-2xl px-3 py-1.5 shadow-2xs">
          <span class="text-xs font-bold text-slate-500 shrink-0">Sesi:</span>
          <select
            v-model="selectedScheduleId"
            @change="handleScheduleChange"
            class="bg-transparent text-xs font-bold text-slate-800 focus:outline-none cursor-pointer max-w-[170px] sm:max-w-[210px] truncate"
          >
            <option v-for="sch in orderedSchedules" :key="sch.id" :value="sch.id">
              {{ sch.title }} ({{ sch.class_room?.name || 'Kelas' }}){{ sch.can_control === false ? ' (pantau saja)' : '' }}
            </option>
          </select>
        </div>

        <!-- Token Ruang Ujian -->
        <div class="flex items-center gap-2.5 bg-indigo-50/80 border border-indigo-200 px-3.5 py-1.5 rounded-2xl shadow-2xs">
          <div>
            <span class="text-[9px] font-bold text-indigo-600 uppercase tracking-wider block leading-none">Token</span>
            <span class="font-mono text-base font-black text-indigo-900 tracking-wider leading-tight">
              {{ proctorData?.exam_token || '------' }}
            </span>
          </div>
          <button
            @click="refreshData"
            class="p-1.5 bg-white hover:bg-indigo-100 active:scale-95 rounded-xl border border-indigo-200 text-indigo-700 transition cursor-pointer shadow-2xs flex items-center justify-center"
            title="Segarkan Data Real-Time"
          >
            <svg class="w-4 h-4" :class="{ 'animate-spin': isRefreshing }" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
            </svg>
          </button>
        </div>
      </div>
    </div>

    <!-- Banner: hanya memantau (bukan pengawas yang ditugaskan) -->
    <div
      v-if="selectedScheduleId && !canControl"
      class="flex items-start gap-2.5 px-4 py-3 bg-amber-50 border border-amber-200 rounded-2xl text-xs text-amber-800"
    >
      <svg class="w-4 h-4 shrink-0 mt-px" fill="none" viewBox="0 0 24 24" stroke="currentColor">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z" />
      </svg>
      <span>Anda hanya memantau jadwal ini; kendali hanya untuk pengawas yang ditugaskan.</span>
    </div>

    <!-- Toolbar: Pencarian & Aksi (baris atas), Filter Status (baris bawah) -->
    <div class="bg-white rounded-3xl p-3.5 border border-slate-200 shadow-xs space-y-3">
      <div class="flex flex-col lg:flex-row lg:items-center justify-between gap-3">
        <!-- Search Input -->
        <div class="relative w-full lg:max-w-sm">
          <svg class="w-4 h-4 text-slate-400 absolute left-3 top-1/2 -translate-y-1/2 pointer-events-none" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
          </svg>
          <input
            v-model="searchQuery"
            type="text"
            placeholder="Cari siswa, NIS, NISN..."
            class="w-full h-9 pl-9 pr-9 bg-white border border-slate-200 rounded-xl text-xs font-medium text-slate-800 placeholder:text-slate-400 focus:border-indigo-500 focus:ring-4 focus:ring-indigo-100 focus:outline-none transition"
          />
          <button
            v-if="searchQuery"
            type="button"
            @click="searchQuery = ''"
            class="absolute right-2 top-1/2 -translate-y-1/2 p-1 rounded-lg text-slate-400 hover:text-slate-600 hover:bg-slate-100 cursor-pointer transition"
            title="Hapus pencarian"
          >
            <svg class="w-3.5 h-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
            </svg>
          </button>
        </div>

        <!-- Action Buttons Group on Right -->
        <div class="flex items-center gap-2 flex-wrap lg:justify-end">
          <!-- Tambah Waktu Seluruh Kelas -->
          <button
            v-if="selectedScheduleId"
            type="button"
            @click="openBulkExtendTimeModal"
            :disabled="!canControl"
            class="px-3.5 py-2 bg-amber-500 hover:bg-amber-600 active:scale-95 text-white font-bold text-xs rounded-xl shadow-xs transition flex items-center gap-1.5 cursor-pointer whitespace-nowrap disabled:opacity-50 disabled:cursor-not-allowed disabled:active:scale-100"
            :title="canControl ? 'Tambah waktu ujian untuk seluruh siswa pada jadwal ini' : controlDisabledHint"
          >
            <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
            </svg>
            <span>Tambah Waktu Kelas</span>
          </button>

          <!-- Slot Header Actions (Rekap Nilai Excel & Berita Acara PDF) -->
          <slot name="header-actions" :schedule-id="selectedScheduleId" :proctor-data="proctorData"></slot>
        </div>
      </div>

      <!-- Filter Status: pil seragam dengan filter di tab Jadwal -->
      <div class="flex items-center gap-1.5 flex-wrap">
        <button
          v-for="f in statusFilterOptions"
          :key="f.key"
          type="button"
          @click="statusFilter = f.key"
          :class="['px-3 py-1.5 rounded-xl text-xs font-bold transition cursor-pointer whitespace-nowrap active:scale-95', statusFilterClass(f)]"
        >
          {{ f.label }} ({{ f.count }})
        </button>
      </div>
    </div>

    <!-- Unified Live Monitoring Table Card -->
    <div class="bg-white rounded-3xl border border-slate-200 shadow-sm overflow-hidden">
      <div class="overflow-x-auto">
        <table class="w-full text-left text-xs">
          <thead>
            <tr class="bg-slate-50 border-b border-slate-200 text-slate-600 font-bold uppercase text-[10px] select-none whitespace-nowrap">
              <!-- Col 1: Nama Siswa & Identitas (Sortable) -->
              <th class="py-3 px-4">
                <button
                  type="button"
                  @click="toggleSort('name')"
                  class="group inline-flex items-center gap-1 cursor-pointer transition hover:text-indigo-600 font-bold uppercase whitespace-nowrap"
                  :class="sortBy.startsWith('name') ? 'text-indigo-600' : 'text-slate-600'"
                  title="Urutkan berdasarkan Nama Siswa"
                >
                  <span>Siswa & Identitas</span>
                  <svg v-if="sortBy === 'name_asc'" class="w-3 h-3 text-indigo-600 shrink-0" viewBox="0 0 20 20" fill="currentColor"><path fill-rule="evenodd" d="M5.293 9.707a1 1 0 010-1.414l4-4a1 1 0 011.414 0l4 4a1 1 0 01-1.414 1.414L11 7.414V15a1 1 0 11-2 0V7.414L6.707 9.707a1 1 0 01-1.414 0z" clip-rule="evenodd" /></svg>
                  <svg v-else-if="sortBy === 'name_desc'" class="w-3 h-3 text-indigo-600 shrink-0" viewBox="0 0 20 20" fill="currentColor"><path fill-rule="evenodd" d="M14.707 10.293a1 1 0 010 1.414l-4 4a1 1 0 01-1.414 0l-4-4a1 1 0 111.414-1.414L9 12.586V5a1 1 0 012 0v7.586l2.293-2.293a1 1 0 011.414 0z" clip-rule="evenodd" /></svg>
                  <svg v-else class="w-3 h-3 text-slate-300 group-hover:text-slate-400 shrink-0" viewBox="0 0 20 20" fill="currentColor"><path d="M5 8l5-5 5 5H5zm10 4l-5 5-5-5h10z" /></svg>
                </button>
              </th>

              <!-- Col 2: Status Pengerjaan (Sortable) -->
              <th class="py-3 px-4 text-center">
                <button
                  type="button"
                  @click="toggleSort('status')"
                  class="group inline-flex items-center gap-1 cursor-pointer transition hover:text-indigo-600 font-bold uppercase mx-auto whitespace-nowrap"
                  :class="sortBy.startsWith('status') ? 'text-indigo-600' : 'text-slate-600'"
                  title="Urutkan berdasarkan Status"
                >
                  <span>Status</span>
                  <svg v-if="sortBy === 'status_asc'" class="w-3 h-3 text-indigo-600 shrink-0" viewBox="0 0 20 20" fill="currentColor"><path fill-rule="evenodd" d="M5.293 9.707a1 1 0 010-1.414l4-4a1 1 0 011.414 0l4 4a1 1 0 01-1.414 1.414L11 7.414V15a1 1 0 11-2 0V7.414L6.707 9.707a1 1 0 01-1.414 0z" clip-rule="evenodd" /></svg>
                  <svg v-else-if="sortBy === 'status_desc'" class="w-3 h-3 text-indigo-600 shrink-0" viewBox="0 0 20 20" fill="currentColor"><path fill-rule="evenodd" d="M14.707 10.293a1 1 0 010 1.414l-4 4a1 1 0 01-1.414 0l-4-4a1 1 0 111.414-1.414L9 12.586V5a1 1 0 012 0v7.586l2.293-2.293a1 1 0 011.414 0z" clip-rule="evenodd" /></svg>
                  <svg v-else class="w-3 h-3 text-slate-300 group-hover:text-slate-400 shrink-0" viewBox="0 0 20 20" fill="currentColor"><path d="M5 8l5-5 5 5H5zm10 4l-5 5-5-5h10z" /></svg>
                </button>
              </th>

              <!-- Col 3: Sisa Waktu -->
              <th class="py-3 px-4 text-center whitespace-nowrap">
                <span>Sisa Waktu</span>
              </th>

              <!-- Col 4: Progres Jawaban (Sortable) -->
              <th class="py-3 px-4 min-w-[140px]">
                <button
                  type="button"
                  @click="toggleSort('progress')"
                  class="group inline-flex items-center gap-1 cursor-pointer transition hover:text-indigo-600 font-bold uppercase whitespace-nowrap"
                  :class="sortBy.startsWith('progress') ? 'text-indigo-600' : 'text-slate-600'"
                  title="Urutkan berdasarkan Progres"
                >
                  <span>Progres Jawaban</span>
                  <svg v-if="sortBy === 'progress_desc'" class="w-3 h-3 text-indigo-600 shrink-0" viewBox="0 0 20 20" fill="currentColor"><path fill-rule="evenodd" d="M14.707 10.293a1 1 0 010 1.414l-4 4a1 1 0 01-1.414 0l-4-4a1 1 0 111.414-1.414L9 12.586V5a1 1 0 012 0v7.586l2.293-2.293a1 1 0 011.414 0z" clip-rule="evenodd" /></svg>
                  <svg v-else-if="sortBy === 'progress_asc'" class="w-3 h-3 text-indigo-600 shrink-0" viewBox="0 0 20 20" fill="currentColor"><path fill-rule="evenodd" d="M5.293 9.707a1 1 0 010-1.414l4-4a1 1 0 011.414 0l4 4a1 1 0 01-1.414 1.414L11 7.414V15a1 1 0 11-2 0V7.414L6.707 9.707a1 1 0 01-1.414 0z" clip-rule="evenodd" /></svg>
                  <svg v-else class="w-3 h-3 text-slate-300 group-hover:text-slate-400 shrink-0" viewBox="0 0 20 20" fill="currentColor"><path d="M5 8l5-5 5 5H5zm10 4l-5 5-5-5h10z" /></svg>
                </button>
              </th>

              <!-- Col 5: Pelanggaran (Sortable) -->
              <th class="py-3 px-4 text-center">
                <button
                  type="button"
                  @click="toggleSort('violations')"
                  class="group inline-flex items-center gap-1 cursor-pointer transition hover:text-indigo-600 font-bold uppercase mx-auto whitespace-nowrap"
                  :class="sortBy.startsWith('violations') ? 'text-indigo-600' : 'text-slate-600'"
                  title="Urutkan berdasarkan Pelanggaran"
                >
                  <span>Pelanggaran</span>
                  <svg v-if="sortBy === 'violations_desc'" class="w-3 h-3 text-indigo-600 shrink-0" viewBox="0 0 20 20" fill="currentColor"><path fill-rule="evenodd" d="M14.707 10.293a1 1 0 010 1.414l-4 4a1 1 0 01-1.414 0l-4-4a1 1 0 111.414-1.414L9 12.586V5a1 1 0 012 0v7.586l2.293-2.293a1 1 0 011.414 0z" clip-rule="evenodd" /></svg>
                  <svg v-else class="w-3 h-3 text-slate-300 group-hover:text-slate-400 shrink-0" viewBox="0 0 20 20" fill="currentColor"><path d="M5 8l5-5 5 5H5zm10 4l-5 5-5-5h10z" /></svg>
                </button>
              </th>

              <!-- Col 6: Nilai PG (Sortable) -->
              <th class="py-3 px-4 text-center">
                <button
                  type="button"
                  @click="toggleSort('score')"
                  class="group inline-flex items-center gap-1 cursor-pointer transition hover:text-indigo-600 font-bold uppercase mx-auto whitespace-nowrap"
                  :class="sortBy.startsWith('score') ? 'text-indigo-600' : 'text-slate-600'"
                  title="Urutkan berdasarkan Nilai PG"
                >
                  <span>Nilai PG</span>
                  <svg v-if="sortBy === 'score_desc'" class="w-3 h-3 text-indigo-600 shrink-0" viewBox="0 0 20 20" fill="currentColor"><path fill-rule="evenodd" d="M14.707 10.293a1 1 0 010 1.414l-4 4a1 1 0 01-1.414 0l-4-4a1 1 0 111.414-1.414L9 12.586V5a1 1 0 012 0v7.586l2.293-2.293a1 1 0 011.414 0z" clip-rule="evenodd" /></svg>
                  <svg v-else class="w-3 h-3 text-slate-300 group-hover:text-slate-400 shrink-0" viewBox="0 0 20 20" fill="currentColor"><path d="M5 8l5-5 5 5H5zm10 4l-5 5-5-5h10z" /></svg>
                </button>
              </th>

              <!-- Col 7: Aksi Tindakan -->
              <th class="py-3 px-4 text-right whitespace-nowrap">Aksi</th>
            </tr>
          </thead>

          <tbody class="divide-y divide-slate-100">
            <tr v-if="filteredAndSortedStudents.length === 0">
              <td colspan="7" class="py-10 text-center text-slate-400">
                <div class="space-y-1.5">
                  <div class="font-bold text-slate-700 text-xs">Tidak ada siswa yang sesuai dengan filter</div>
                  <div class="text-[11px] text-slate-400">Coba sesuaikan kata kunci pencarian atau filter status pengerjaan.</div>
                  <button
                    type="button"
                    @click="searchQuery = ''; statusFilter = 'ALL'; sortBy = 'name_asc'; currentPage = 1"
                    class="px-3 py-1 bg-slate-100 hover:bg-slate-200 text-slate-700 font-bold text-xs rounded-xl transition cursor-pointer mt-1"
                  >
                    Reset Filter
                  </button>
                </div>
              </td>
            </tr>

            <tr
              v-for="st in paginatedStudents"
              :key="st.student_id"
              class="hover:bg-slate-50/60 transition"
            >
              <!-- Col 1: Siswa & Identitas -->
              <td class="py-3.5 px-4">
                <div class="flex items-center gap-2 flex-wrap">
                  <span class="font-bold text-slate-900 text-xs">{{ st.full_name }}</span>
                  <span class="px-2 py-0.5 bg-indigo-50 text-indigo-700 text-[10px] font-semibold rounded-md border border-indigo-100">
                    {{ st.class_name || 'Kelas' }}
                  </span>
                </div>
                <div class="text-[11px] text-slate-400 mt-0.5 font-mono flex items-center gap-2">
                  <span>NIS: {{ st.nis || '-' }}</span>
                  <span v-if="st.nisn">• NISN: {{ st.nisn }}</span>
                  <span v-if="st.client_ip" class="text-slate-400 font-normal hidden lg:inline">({{ st.client_ip }})</span>
                </div>
              </td>

              <!-- Col 2: Status Pengerjaan -->
              <td class="py-3.5 px-4 text-center">
                <span :class="['px-2.5 py-1 rounded-full text-[10px] font-bold inline-block', getStatusBadge(st.status)]">
                  {{ getStatusLabel(st.status) }}
                </span>
              </td>

              <!-- Col 3: Sisa Waktu -->
              <td class="py-3.5 px-4 text-center font-mono">
                <span
                  v-if="st.status === 'IN_PROGRESS' || st.status === 'BLOCKED'"
                  :class="[
                    'px-2 py-0.5 rounded-lg text-xs font-bold tracking-wider inline-block',
                    (countdownMap[st.student_id] ?? st.remaining_seconds ?? 0) <= 300
                      ? 'bg-rose-100 text-rose-700 animate-pulse border border-rose-300'
                      : 'bg-indigo-50 text-indigo-800 border border-indigo-200'
                  ]"
                >
                  ⏱️ {{ formatSeconds(countdownMap[st.student_id] ?? st.remaining_seconds ?? 0) }}
                </span>
                <span v-else-if="st.status === 'SUBMITTED'" class="text-[11px] text-emerald-600 font-semibold">
                  Selesai
                </span>
                <span v-else class="text-slate-400 text-xs">-</span>
              </td>

              <!-- Col 4: Progres Jawaban -->
              <td class="py-3.5 px-4 min-w-[140px]">
                <div class="flex items-center justify-between text-[11px] text-slate-600 mb-1">
                  <span>{{ st.answered_count }} / {{ st.total_questions }} Soal</span>
                  <span class="font-bold text-slate-800">{{ st.progress_percent }}%</span>
                </div>
                <div class="w-full bg-slate-100 h-2 rounded-full overflow-hidden">
                  <div
                    class="bg-indigo-600 h-full rounded-full transition-all duration-300"
                    :style="{ width: `${st.progress_percent}%` }"
                  ></div>
                </div>
              </td>

              <!-- Col 5: Pelanggaran Anti-Cheat -->
              <td class="py-3.5 px-4 text-center">
                <button
                  v-if="st.violation_count > 0"
                  type="button"
                  @click="openViolationLogModal(st)"
                  class="font-mono font-bold text-xs hover:underline cursor-pointer inline-flex items-center gap-1 group text-rose-600"
                  title="Klik untuk melihat catatan pelanggaran"
                >
                  <span>{{ st.violation_count }} Pelanggaran</span>
                  <svg class="w-3 h-3 opacity-60 group-hover:opacity-100 transition" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z" />
                  </svg>
                </button>
                <span v-else class="text-slate-400 text-xs">-</span>
              </td>

              <!-- Col 6: Nilai PG -->
              <td class="py-3.5 px-4 text-center font-mono font-bold">
                <span v-if="st.status === 'SUBMITTED'" class="text-xs text-indigo-700 bg-indigo-50 px-2 py-0.5 rounded-md border border-indigo-100">
                  {{ st.total_score.toFixed(1) }}
                </span>
                <span v-else class="text-slate-400 font-normal text-xs">-</span>
              </td>

              <!-- Col 7: Aksi Tindakan (Matching exact screenshot button styles) -->
              <td class="py-3.5 px-4 text-right">
                <div class="flex items-center justify-end gap-1.5 flex-wrap">
                  <!-- Buka Blokir (Jika BLOCKED) -->
                  <button
                    v-if="st.status === 'BLOCKED'"
                    type="button"
                    @click="handleUnlock(st)"
                    :disabled="!canControl"
                    class="px-2.5 py-1 bg-rose-50 text-rose-700 hover:bg-rose-100 border border-rose-200 rounded-lg text-xs font-bold transition shadow-2xs cursor-pointer flex items-center gap-1 disabled:opacity-50 disabled:cursor-not-allowed"
                    :title="canControl ? 'Buka blokir sesi siswa ini' : controlDisabledHint"
                  >
                    <span>🔓 Buka Blokir</span>
                  </button>

                  <!-- Tambah Waktu Individu -->
                  <button
                    v-if="st.status === 'IN_PROGRESS' || st.status === 'BLOCKED'"
                    type="button"
                    @click="openSingleExtendTimeModal(st)"
                    :disabled="!canControl"
                    class="px-2.5 py-1 bg-indigo-50 hover:bg-indigo-100 text-indigo-700 border border-indigo-200 rounded-lg text-xs font-bold transition cursor-pointer flex items-center gap-1 shadow-2xs disabled:opacity-50 disabled:cursor-not-allowed"
                    :title="canControl ? 'Tambah waktu ujian individu' : controlDisabledHint"
                  >
                    <span>⏱️</span>
                    <span>+Waktu</span>
                  </button>

                  <!-- Selesaikan / Force Submit -->
                  <button
                    v-if="st.status === 'IN_PROGRESS' || st.status === 'BLOCKED'"
                    type="button"
                    @click="handleForceSubmit(st)"
                    :disabled="!canControl"
                    class="px-2.5 py-1 bg-emerald-600 hover:bg-emerald-700 text-white rounded-lg text-xs font-bold transition shadow-xs cursor-pointer flex items-center gap-1 disabled:opacity-50 disabled:cursor-not-allowed"
                    :title="canControl ? 'Kumpulkan paksa ujian siswa' : controlDisabledHint"
                  >
                    <span>Selesaikan</span>
                  </button>

                  <!-- Reset Login HP -->
                  <button
                    type="button"
                    @click="handleResetDevice(st)"
                    :disabled="!canControl"
                    class="px-2.5 py-1 bg-slate-100 hover:bg-slate-200 text-slate-700 rounded-lg text-xs font-semibold transition cursor-pointer flex items-center gap-1 disabled:opacity-50 disabled:cursor-not-allowed"
                    :title="canControl ? 'Reset login perangkat HP / Komputer' : controlDisabledHint"
                  >
                    <svg class="w-3.5 h-3.5 text-slate-500" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
                    </svg>
                    <span>Reset HP</span>
                  </button>

                  <!-- Log Pelanggaran Icon Button -->
                  <button
                    v-if="st.session_id"
                    type="button"
                    @click="openViolationLogModal(st)"
                    class="p-1.5 text-slate-400 hover:text-indigo-600 hover:bg-indigo-50 rounded-lg transition cursor-pointer"
                    title="Lihat Rekaman Log Pelanggaran"
                  >
                    <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
                    </svg>
                  </button>
                </div>
              </td>
            </tr>
          </tbody>
        </table>
      </div>

      <!-- Unified Pagination Footer (Matching screenshot: Menampilkan 1 - 1 dari 1 paket | Tampilkan: 10/hal) -->
      <div class="px-5 py-3.5 border-t border-slate-100 bg-slate-50/50 flex flex-col sm:flex-row items-center justify-between gap-3 text-xs text-slate-500">
        <div class="flex items-center gap-4">
          <div>
            Menampilkan <strong>{{ paginationInfo.from }} – {{ paginationInfo.to }}</strong> dari <strong>{{ filteredAndSortedStudents.length }}</strong> siswa
          </div>

          <!-- Per Page Selector Dropdown -->
          <div class="flex items-center space-x-1.5">
            <label class="text-[11px] text-slate-500 font-medium">Tampilkan:</label>
            <select
              v-model.number="perPage"
              class="px-2.5 py-1 rounded-xl border border-slate-200 text-xs font-semibold focus:ring-2 focus:ring-indigo-500 bg-white cursor-pointer shadow-2xs"
            >
              <option :value="10">10 / hal</option>
              <option :value="25">25 / hal</option>
              <option :value="50">50 / hal</option>
              <option :value="100">100 / hal</option>
              <option :value="-1">Semua</option>
            </select>
          </div>
        </div>

        <!-- Navigation Buttons -->
        <div v-if="perPage !== -1 && totalPages > 1" class="flex items-center gap-1">
          <button
            type="button"
            @click="currentPage = Math.max(1, currentPage - 1)"
            :disabled="currentPage === 1"
            class="px-2.5 py-1 rounded-lg border border-slate-200 bg-white hover:bg-slate-100 disabled:opacity-40 disabled:pointer-events-none transition cursor-pointer"
          >
            ‹
          </button>

          <template v-for="p in visiblePages" :key="p">
            <span v-if="p === '...'" class="px-2 text-slate-400">...</span>
            <button
              v-else
              type="button"
              @click="currentPage = p"
              :class="[
                'w-7 h-7 rounded-lg text-xs font-bold transition cursor-pointer',
                currentPage === p ? 'bg-indigo-600 text-white shadow-2xs' : 'border border-slate-200 bg-white hover:bg-slate-100 text-slate-700'
              ]"
            >
              {{ p }}
            </button>
          </template>

          <button
            type="button"
            @click="currentPage = Math.min(totalPages, currentPage + 1)"
            :disabled="currentPage === totalPages"
            class="px-2.5 py-1 rounded-lg border border-slate-200 bg-white hover:bg-slate-100 disabled:opacity-40 disabled:pointer-events-none transition cursor-pointer"
          >
            ›
          </button>
        </div>
      </div>
    </div>

    <!-- MODAL: TAMBAH WAKTU UJIAN (INDIVIDU / MASSAL) -->
    <div
      v-if="showExtendTimeModal"
      class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-slate-900/60 backdrop-blur-xs transition-opacity"
    >
      <div class="bg-white rounded-3xl max-w-md w-full p-6 shadow-2xl border border-slate-100 transform transition-all scale-100 animate-in fade-in zoom-in-95">
        <div class="flex items-center justify-between pb-3 border-b border-slate-100">
          <div class="flex items-center space-x-2.5">
            <div class="w-9 h-9 rounded-2xl bg-indigo-100 text-indigo-700 flex items-center justify-center font-bold text-base">
              ⏱️
            </div>
            <div>
              <h3 class="text-sm font-bold text-slate-900">
                {{ extendTimeMode === 'bulk' ? 'Tambah Waktu Seluruh Kelas' : 'Tambah Waktu Ujian Siswa' }}
              </h3>
              <p class="text-[11px] text-slate-500">
                {{ extendTimeMode === 'bulk' ? `Jadwal: ${proctorData?.schedule_title}` : `${selectedStudentForTime?.full_name} (${selectedStudentForTime?.nis})` }}
              </p>
            </div>
          </div>
          <button @click="showExtendTimeModal = false" class="text-slate-400 hover:text-slate-600 p-1 rounded-lg cursor-pointer">
            ✕
          </button>
        </div>

        <div class="py-4 space-y-4 text-xs">
          <!-- Preset Minutes Selector -->
          <div class="space-y-1.5">
            <label class="font-bold text-slate-700 block">Pilihan Cepat Tambahan Waktu:</label>
            <div class="grid grid-cols-4 gap-2">
              <button
                v-for="preset in [5, 10, 15, 30]"
                :key="preset"
                type="button"
                @click="extraMinutesInput = preset"
                :class="[
                  'py-2 rounded-xl font-bold text-xs transition border cursor-pointer',
                  extraMinutesInput === preset
                    ? 'bg-indigo-600 text-white border-indigo-700 shadow-2xs'
                    : 'bg-slate-50 text-slate-700 border-slate-200 hover:bg-slate-100'
                ]"
              >
                +{{ preset }} Mnt
              </button>
            </div>
          </div>

          <!-- Custom Minutes Input -->
          <div class="space-y-1">
            <label class="font-bold text-slate-700 block">Atau Tentukan Jumlah Menit:</label>
            <div class="relative">
              <input
                v-model.number="extraMinutesInput"
                type="number"
                min="1"
                max="180"
                class="w-full px-3 py-2 bg-slate-50 border border-slate-200 rounded-xl text-xs font-bold focus:bg-white focus:ring-2 focus:ring-indigo-500 focus:outline-none"
              />
              <span class="absolute right-3 top-1/2 -translate-y-1/2 text-slate-400 font-semibold">Menit</span>
            </div>
          </div>

          <!-- Reason / Catatan Pengawas -->
          <div class="space-y-1">
            <label class="font-bold text-slate-700 block">Alasan Penambahan Waktu (Untuk Berita Acara):</label>
            <input
              v-model="extendTimeReason"
              type="text"
              placeholder="Contoh: Kendala jaringan lab / restart perangkat"
              class="w-full px-3 py-2 bg-slate-50 border border-slate-200 rounded-xl text-xs focus:bg-white focus:ring-2 focus:ring-indigo-500 focus:outline-none"
            />
          </div>

          <div v-if="extendTimeMode === 'bulk'" class="p-3 bg-indigo-50 border border-indigo-200 rounded-2xl text-[11px] text-indigo-800">
            ℹ️ <strong>Informasi:</strong> Penambahan waktu akan diterapkan secara serentak kepada seluruh siswa yang sedang aktif atau terblokir pada jadwal ini.
          </div>
        </div>

        <div class="flex items-center justify-end gap-2 pt-3 border-t border-slate-100">
          <button
            type="button"
            @click="showExtendTimeModal = false"
            class="px-4 py-2 bg-slate-100 hover:bg-slate-200 text-slate-700 font-semibold text-xs rounded-xl transition cursor-pointer"
          >
            Batal
          </button>
          <button
            type="button"
            @click="submitExtendTime"
            :disabled="isSubmittingTime || extraMinutesInput <= 0"
            class="px-4 py-2 bg-indigo-600 hover:bg-indigo-700 active:scale-95 disabled:opacity-50 text-white font-bold text-xs rounded-xl shadow-xs transition flex items-center gap-1.5 cursor-pointer"
          >
            <span v-if="isSubmittingTime" class="animate-spin">⌛</span>
            <span>Simpan & Tambah Waktu</span>
          </button>
        </div>
      </div>
    </div>

    <!-- MODAL: LOG PELANGGARAN & AKTIVITAS SISWA -->
    <div
      v-if="showViolationModal"
      class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-slate-900/60 backdrop-blur-xs transition-opacity"
    >
      <div class="bg-white rounded-3xl max-w-xl w-full p-6 shadow-2xl border border-slate-100 flex flex-col max-h-[85vh]">
        <div class="flex items-center justify-between pb-3 border-b border-slate-100 shrink-0">
          <div class="flex items-center space-x-2.5">
            <div class="w-9 h-9 rounded-2xl bg-rose-100 text-rose-700 flex items-center justify-center font-bold text-base">
              📋
            </div>
            <div>
              <h3 class="text-sm font-bold text-slate-900">
                Log Pelanggaran Anti-Cheat & Aktivitas
              </h3>
              <p class="text-[11px] text-slate-500">
                {{ selectedStudentForViolation?.full_name }} (NIS: {{ selectedStudentForViolation?.nis || '-' }})
              </p>
            </div>
          </div>
          <button @click="showViolationModal = false" class="text-slate-400 hover:text-slate-600 p-1 rounded-lg cursor-pointer">
            ✕
          </button>
        </div>

        <!-- Violation Log Timeline Area -->
        <div class="py-4 overflow-y-auto flex-1 space-y-3">
          <div v-if="isLoadingViolations" class="py-12 text-center text-slate-400">
            <div class="animate-spin text-2xl mb-2">⌛</div>
            <p class="text-xs">Memuat rekaman log pelanggaran...</p>
          </div>
          <div v-else-if="violationLogsList.length === 0" class="py-12 text-center text-slate-400">
            <span class="text-3xl block mb-2">🛡️</span>
            <p class="text-xs font-bold text-slate-700">Tidak ada catatan pelanggaran</p>
            <p class="text-[11px] text-slate-400 mt-0.5">Siswa ini mengerjakan ujian dengan tertib dan belum pernah terdeteksi melanggar layar.</p>
          </div>
          <div v-else class="space-y-2">
            <div
              v-for="(log, lIdx) in violationLogsList"
              :key="log.id || lIdx"
              class="p-3 bg-slate-50 rounded-2xl border border-slate-200/80 flex items-start justify-between gap-3 text-xs"
            >
              <div class="space-y-1">
                <div class="flex items-center gap-2">
                  <span :class="['px-2 py-0.5 rounded-md font-bold text-[10px]', getViolationBadge(log.event_type)]">
                    {{ formatViolationType(log.event_type) }}
                  </span>
                  <span class="text-slate-400 text-[11px] font-mono">
                    {{ formatLogTime(log.occurred_at) }}
                  </span>
                </div>
                <p class="text-slate-700 font-medium text-xs">{{ log.details || 'Tidak ada keterangan tambahan.' }}</p>
              </div>
            </div>
          </div>
        </div>

        <div class="pt-3 border-t border-slate-100 flex items-center justify-between shrink-0">
          <span class="text-[11px] text-slate-400">
            Total tercatat: <strong>{{ violationLogsList.length }}</strong> aktivitas
          </span>
          <button
            type="button"
            @click="showViolationModal = false"
            class="px-4 py-2 bg-slate-200 hover:bg-slate-300 text-slate-700 font-bold text-xs rounded-xl transition cursor-pointer"
          >
            Tutup
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, onUnmounted, watch } from 'vue'
import api from '@/services/api'
import { useDialog } from '@/composables/useDialog'
import { useAuthStore } from '@/stores/auth'

const authStore = useAuthStore()

const props = defineProps({
  schedules: {
    type: Array,
    default: () => []
  },
  initialScheduleId: {
    type: String,
    default: null
  }
})

const emit = defineEmits(['schedule-changed', 'data-refreshed'])

const { confirm: showConfirmModal, alert: showAlertModal, toast: showToast } = useDialog()

// State
const schedulesList = ref([])
const selectedScheduleId = ref(null)
const proctorData = ref(null)
const isRefreshing = ref(false)

// Table Toolbar Filter & Sort State
const searchQuery = ref('')
const statusFilter = ref('ALL')
const statusFilterOptions = computed(() => [
  { key: 'ALL', label: 'Semua', count: allStudentsList.value.length },
  { key: 'IN_PROGRESS', label: 'Mengerjakan', count: inProgressCount.value },
  { key: 'BLOCKED', label: 'Terblokir', count: blockedCount.value },
  { key: 'SUBMITTED', label: 'Selesai', count: submittedCount.value },
  { key: 'NOT_STARTED', label: 'Belum Mulai', count: notStartedCount.value },
])
// Kelas pil filter status: aktif berwarna per status, tidak aktif abu-abu (Terblokir bernuansa merah bila ada)
const statusFilterActiveClass = {
  ALL: 'bg-indigo-600 text-white shadow-xs border border-transparent',
  IN_PROGRESS: 'bg-blue-600 text-white shadow-xs border border-transparent',
  BLOCKED: 'bg-rose-600 text-white shadow-xs border border-transparent',
  SUBMITTED: 'bg-emerald-600 text-white shadow-xs border border-transparent',
  NOT_STARTED: 'bg-slate-800 text-white shadow-xs border border-transparent',
}
const statusFilterClass = (f) => {
  if (statusFilter.value === f.key) return statusFilterActiveClass[f.key]
  if (f.key === 'BLOCKED' && f.count > 0) {
    return 'bg-rose-50 text-rose-700 hover:bg-rose-100 border border-rose-200'
  }
  return 'bg-slate-100 text-slate-600 hover:bg-slate-200 border border-transparent'
}
const sortBy = ref('name_asc')
const perPage = ref(10)
const currentPage = ref(1)

// Live Countdown Local Map (student_id -> remaining seconds)
const countdownMap = ref({})
let pollInterval = null
let tickInterval = null

// Modals State
const showExtendTimeModal = ref(false)
const extendTimeMode = ref('single') // 'single' | 'bulk'
const selectedStudentForTime = ref(null)
const extraMinutesInput = ref(10)
const extendTimeReason = ref('')
const isSubmittingTime = ref(false)

const showViolationModal = ref(false)
const selectedStudentForViolation = ref(null)
const violationLogsList = ref([])
const isLoadingViolations = ref(false)

// Counts
const allStudentsList = computed(() => {
  return proctorData.value?.students || []
})

const inProgressCount = computed(() => {
  return allStudentsList.value.filter(s => s.status === 'IN_PROGRESS').length
})

const blockedCount = computed(() => {
  return allStudentsList.value.filter(s => s.status === 'BLOCKED').length
})

const submittedCount = computed(() => {
  return allStudentsList.value.filter(s => s.status === 'SUBMITTED').length
})

const notStartedCount = computed(() => {
  return allStudentsList.value.filter(s => s.status === 'NOT_STARTED').length
})

// Sort helper toggler
const toggleSort = (column) => {
  if (column === 'name') {
    sortBy.value = sortBy.value === 'name_asc' ? 'name_desc' : 'name_asc'
  } else if (column === 'status') {
    sortBy.value = sortBy.value === 'status_asc' ? 'status_desc' : 'status_asc'
  } else if (column === 'progress') {
    sortBy.value = sortBy.value === 'progress_desc' ? 'progress_asc' : 'progress_desc'
  } else if (column === 'violations') {
    sortBy.value = sortBy.value === 'violations_desc' ? 'violations_asc' : 'violations_desc'
  } else if (column === 'score') {
    sortBy.value = sortBy.value === 'score_desc' ? 'score_asc' : 'score_desc'
  }
}

const filteredAndSortedStudents = computed(() => {
  let list = [...allStudentsList.value]

  // Filter Search Query
  if (searchQuery.value.trim()) {
    const q = searchQuery.value.trim().toLowerCase()
    list = list.filter(s => {
      const name = (s.full_name || '').toLowerCase()
      const nis = (s.nis || '').toLowerCase()
      const nisn = (s.nisn || '').toLowerCase()
      const cls = (s.class_name || '').toLowerCase()
      return name.includes(q) || nis.includes(q) || nisn.includes(q) || cls.includes(q)
    })
  }

  // Filter Status
  if (statusFilter.value !== 'ALL') {
    list = list.filter(s => s.status === statusFilter.value)
  }

  // Sort
  list.sort((a, b) => {
    switch (sortBy.value) {
      case 'name_desc':
        return (b.full_name || '').localeCompare(a.full_name || '')
      case 'status_asc':
        return (a.status || '').localeCompare(b.status || '')
      case 'status_desc':
        return (b.status || '').localeCompare(a.status || '')
      case 'violations_desc':
        return (b.violation_count || 0) - (a.violation_count || 0)
      case 'violations_asc':
        return (a.violation_count || 0) - (b.violation_count || 0)
      case 'progress_desc':
        return (b.progress_percent || 0) - (a.progress_percent || 0)
      case 'progress_asc':
        return (a.progress_percent || 0) - (b.progress_percent || 0)
      case 'score_desc':
        return (b.total_score || 0) - (a.total_score || 0)
      case 'score_asc':
        return (a.total_score || 0) - (b.total_score || 0)
      case 'name_asc':
      default:
        return (a.full_name || '').localeCompare(b.full_name || '')
    }
  })

  return list
})

const totalPages = computed(() => {
  if (perPage.value === -1) return 1
  return Math.max(1, Math.ceil(filteredAndSortedStudents.value.length / perPage.value))
})

const paginatedStudents = computed(() => {
  if (perPage.value === -1) return filteredAndSortedStudents.value
  const start = (currentPage.value - 1) * perPage.value
  return filteredAndSortedStudents.value.slice(start, start + perPage.value)
})

const paginationInfo = computed(() => {
  const total = filteredAndSortedStudents.value.length
  if (total === 0) return { from: 0, to: 0 }
  if (perPage.value === -1) return { from: 1, to: total }
  const from = (currentPage.value - 1) * perPage.value + 1
  const to = Math.min(currentPage.value * perPage.value, total)
  return { from, to }
})

const visiblePages = computed(() => {
  const pages = []
  const total = totalPages.value
  const current = currentPage.value

  if (total <= 7) {
    for (let i = 1; i <= total; i++) pages.push(i)
  } else {
    pages.push(1)
    if (current > 3) pages.push('...')
    const start = Math.max(2, current - 1)
    const end = Math.min(total - 1, current + 1)
    for (let i = start; i <= end; i++) pages.push(i)
    if (current < total - 2) pages.push('...')
    pages.push(total)
  }
  return pages
})

// Watch filters to reset page
watch([searchQuery, statusFilter, perPage, sortBy], () => {
  currentPage.value = 1
})

// Hak kendali: can_control === false berarti pengguna hanya boleh memantau (tidak ada = boleh, perilaku lama).
const orderedSchedules = computed(() => {
  const list = [...schedulesList.value]
  // Array.prototype.sort stabil, urutan asli tetap terjaga di dalam tiap kelompok.
  return list.sort((a, b) => Number(b.can_control !== false) - Number(a.can_control !== false))
})

const canControl = computed(() => {
  const sch = schedulesList.value.find(s => s.id === selectedScheduleId.value)
  if (sch) return sch.can_control !== false
  // Jadwal di luar daftar (mis. dibuka lewat Live Monitor): hanya pemegang control_all yang boleh mengendalikan.
  return authStore.hasPermission('proctor:control_all')
})

const controlDisabledHint = 'Hanya pengawas yang ditugaskan pada jadwal ini yang dapat melakukan kendali'

// Helpers
const getStatusBadge = (status) => {
  switch (status) {
    case 'SUBMITTED':
      return 'bg-emerald-100 text-emerald-800'
    case 'IN_PROGRESS':
      return 'bg-blue-100 text-blue-800'
    case 'BLOCKED':
      return 'bg-rose-100 text-rose-800 border border-rose-300'
    default:
      return 'bg-slate-100 text-slate-600'
  }
}

const getStatusLabel = (status) => {
  switch (status) {
    case 'SUBMITTED':
      return 'Selesai'
    case 'IN_PROGRESS':
      return 'Sedang Mengerjakan'
    case 'BLOCKED':
      return 'Terblokir'
    default:
      return 'Belum Mulai'
  }
}

const formatSeconds = (seconds) => {
  if (!seconds || seconds <= 0) return '00:00'
  const m = Math.floor(seconds / 60)
  const s = Math.floor(seconds % 60)
  const pad = (n) => String(n).padStart(2, '0')
  if (m >= 60) {
    const h = Math.floor(m / 60)
    const remM = m % 60
    return `${pad(h)}:${pad(remM)}:${pad(s)}`
  }
  return `${pad(m)}:${pad(s)}`
}

const formatLogTime = (dateStr) => {
  if (!dateStr) return '-'
  const d = new Date(dateStr)
  return d.toLocaleTimeString('id-ID', { hour: '2-digit', minute: '2-digit', second: '2-digit' }) + ' WIB'
}

const getViolationBadge = (type) => {
  switch (type) {
    case 'TIME_EXTENDED':
      return 'bg-indigo-100 text-indigo-800'
    case 'BLUR_WINDOW':
    case 'TAB_SWITCH':
      return 'bg-amber-100 text-amber-800'
    case 'FULLSCREEN_EXIT':
      return 'bg-rose-100 text-rose-800'
    default:
      return 'bg-slate-100 text-slate-800'
  }
}

const formatViolationType = (type) => {
  switch (type) {
    case 'TIME_EXTENDED':
      return '⏱️ Penambahan Waktu'
    case 'BLUR_WINDOW':
      return '🔲 Pindah Jendela / Fokus'
    case 'TAB_SWITCH':
      return '📑 Pindah Tab Browser'
    case 'FULLSCREEN_EXIT':
      return '🖥️ Keluar Fullscreen'
    case 'DEVTOOLS':
      return '⚠️ Inspeksi DevTools'
    default:
      return type || 'Pelanggaran'
  }
}

// Data Fetching
const handleScheduleChange = () => {
  emit('schedule-changed', selectedScheduleId.value)
  refreshData()
}

const refreshData = async () => {
  if (!selectedScheduleId.value) return
  isRefreshing.value = true
  try {
    const res = await api.get(`/proctor/live/${selectedScheduleId.value}`)
    proctorData.value = res.data.data
    emit('data-refreshed', proctorData.value)

    // Sync local countdown counters
    if (proctorData.value?.students) {
      const newMap = {}
      for (const st of proctorData.value.students) {
        newMap[st.student_id] = st.remaining_seconds || 0
      }
      countdownMap.value = newMap
    }
  } catch (err) {
    console.error('Failed to fetch proctor live data', err)
  } finally {
    isRefreshing.value = false
  }
}

// Dipasang setelah deklarasi refreshData (hindari TDZ saat immediate berjalan).
watch(() => props.schedules, (newVal) => {
  if (newVal && newVal.length > 0) {
    schedulesList.value = newVal
    if (!selectedScheduleId.value) {
      const firstControllable = newVal.find(s => s.can_control !== false)
      selectedScheduleId.value = props.initialScheduleId || (firstControllable || newVal[0]).id
      refreshData()
    }
  }
}, { immediate: true })

// Local 1s ticker for smooth countdown
const startTicker = () => {
  tickInterval = setInterval(() => {
    for (const sid in countdownMap.value) {
      if (countdownMap.value[sid] > 0) {
        countdownMap.value[sid]--
      }
    }
  }, 1000)
}

// Actions
const handleUnlock = async (st) => {
  if (!canControl.value || !st.session_id) return
  try {
    await api.post(`/proctor/sessions/${st.session_id}/unlock`)
    showToast(`Berhasil membuka blokir untuk ${st.full_name}`, 'success')
    await refreshData()
  } catch (err) {
    await showAlertModal({
      title: 'Gagal Membuka Kunci',
      message: err.response?.data?.message || 'Terjadi kesalahan saat membuka kunci siswa.',
      type: 'danger'
    })
  }
}

const handleResetDevice = async (st) => {
  if (!canControl.value) return
  const confirmed = await showConfirmModal({
    title: 'Reset Sesi Perangkat',
    message: `Reset sesi login perangkat untuk ${st.full_name}? Siswa akan dapat login kembali di smartphone/komputer baru.`,
    type: 'warning',
    confirmText: 'Reset Sesi',
    cancelText: 'Batal'
  })
  if (!confirmed) return

  try {
    await api.post(`/proctor/students/${st.student_id}/reset-session`)
    showToast(`Sesi perangkat ${st.full_name} berhasil direset.`, 'success')
    await refreshData()
  } catch (err) {
    await showAlertModal({
      title: 'Gagal Reset Sesi',
      message: err.response?.data?.message || 'Terjadi kesalahan saat mereset sesi perangkat.',
      type: 'danger'
    })
  }
}

const handleForceSubmit = async (st) => {
  if (!canControl.value || !st.session_id) return
  const confirmed = await showConfirmModal({
    title: 'Kumpulkan Paksa Ujian Siswa',
    message: `Kumpulkan paksa ujian untuk ${st.full_name}? Seluruh jawaban yang telah tersimpan akan dinilai otomatis dan sesi siswa akan ditutup.`,
    type: 'danger',
    confirmText: 'Kumpulkan Paksa',
    cancelText: 'Batal'
  })
  if (!confirmed) return

  try {
    const res = await api.post(`/proctor/sessions/${st.session_id}/force-submit`)
    showToast(`Ujian ${st.full_name} berhasil dikumpulkan (Skor: ${res.data.total_score?.toFixed(1) ?? '-'}).`, 'success')
    await refreshData()
  } catch (err) {
    await showAlertModal({
      title: 'Gagal Mengumpulkan Ujian',
      message: err.response?.data?.message || 'Terjadi kesalahan saat mengumpulkan paksa ujian.',
      type: 'danger'
    })
  }
}

// Modal Handlers
const openSingleExtendTimeModal = (st) => {
  extendTimeMode.value = 'single'
  selectedStudentForTime.value = st
  extraMinutesInput.value = 10
  extendTimeReason.value = ''
  showExtendTimeModal.value = true
}

const openBulkExtendTimeModal = () => {
  extendTimeMode.value = 'bulk'
  selectedStudentForTime.value = null
  extraMinutesInput.value = 15
  extendTimeReason.value = ''
  showExtendTimeModal.value = true
}

const submitExtendTime = async () => {
  if (!canControl.value || extraMinutesInput.value <= 0) return
  isSubmittingTime.value = true
  try {
    if (extendTimeMode.value === 'bulk') {
      const res = await api.post(`/proctor/schedules/${selectedScheduleId.value}/extend-time-all`, {
        extra_minutes: Number(extraMinutesInput.value),
        reason: extendTimeReason.value
      })
      showToast(res.data.message || `Berhasil menambah +${extraMinutesInput.value} menit untuk seluruh kelas.`, 'success')
    } else {
      const sessId = selectedStudentForTime.value?.session_id
      if (!sessId) return
      const res = await api.post(`/proctor/sessions/${sessId}/extend-time`, {
        extra_minutes: Number(extraMinutesInput.value),
        reason: extendTimeReason.value
      })
      showToast(res.data.message || `Berhasil menambah +${extraMinutesInput.value} menit untuk ${selectedStudentForTime.value.full_name}.`, 'success')
    }
    showExtendTimeModal.value = false
    await refreshData()
  } catch (err) {
    await showAlertModal({
      title: 'Gagal Menambah Waktu',
      message: err.response?.data?.message || 'Terjadi kesalahan saat menambahkan waktu ujian.',
      type: 'danger'
    })
  } finally {
    isSubmittingTime.value = false
  }
}

const openViolationLogModal = async (st) => {
  selectedStudentForViolation.value = st
  violationLogsList.value = []
  showViolationModal.value = true
  if (!st.session_id) return

  isLoadingViolations.value = true
  try {
    const res = await api.get(`/proctor/sessions/${st.session_id}/violations`)
    violationLogsList.value = res.data.data || []
  } catch (err) {
    console.error('Failed to load violation logs', err)
  } finally {
    isLoadingViolations.value = false
  }
}

onMounted(() => {
  startTicker()
  // Poll every 4 seconds
  pollInterval = setInterval(refreshData, 4000)
})

onUnmounted(() => {
  if (pollInterval) clearInterval(pollInterval)
  if (tickInterval) clearInterval(tickInterval)
})

defineExpose({
  refreshData,
  selectedScheduleId,
  proctorData
})
</script>
