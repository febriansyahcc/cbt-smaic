<template>
  <!-- TAB 0: BERANDA / RINGKASAN EKSEKUTIF CBT -->
  <div v-if="activeTab === 'dashboard'" class="space-y-6">
    <!-- Kartu statistik global (khusus pengelola master data) -->
    <div v-if="canManageMaster" class="grid grid-cols-2 sm:grid-cols-4 gap-3 sm:gap-4">
      <div class="bg-white rounded-3xl p-4 border border-slate-200 shadow-xs">
        <span class="text-xs font-semibold text-slate-500">Total Siswa</span>
        <div class="text-2xl font-black text-slate-900 mt-1">
          {{ stats.total_students || 0 }} <span class="text-xs font-medium text-slate-400">Siswa</span>
        </div>
      </div>
      <div class="bg-white rounded-3xl p-4 border border-slate-200 shadow-xs">
        <span class="text-xs font-semibold text-slate-500">Guru / Pengawas</span>
        <div class="text-2xl font-black text-indigo-600 mt-1">
          {{ stats.total_teachers || 0 }} <span class="text-xs font-medium text-slate-400">Guru</span>
        </div>
      </div>
      <div class="bg-white rounded-3xl p-4 border border-slate-200 shadow-xs">
        <span class="text-xs font-semibold text-slate-500">Kelas</span>
        <div class="text-2xl font-black text-slate-900 mt-1">
          {{ stats.total_classes || 0 }}
        </div>
      </div>
      <div class="bg-white rounded-3xl p-4 border border-emerald-200 shadow-xs bg-emerald-50/20">
        <span class="text-xs font-semibold text-emerald-700">{{ selectedExamEvent ? 'Jadwal Aktif (event ini)' : 'Jadwal Ujian Aktif' }}</span>
        <div class="text-2xl font-black text-emerald-700 mt-1">
          {{ selectedExamEvent ? activeSchedulesCount : (stats.active_schedules || 0) }} <span class="text-xs font-medium text-emerald-400">Sesi</span>
        </div>
      </div>
    </div>

    <!-- Beranda guru / pengawas (staf tanpa hak kelola): ringkasan tugas milik pengguna -->
    <template v-if="showWelcomeCard">
      <!-- Sapaan, konteks event, dan pintasan -->
      <div class="bg-white rounded-3xl p-5 sm:p-6 border border-slate-200/80 shadow-xs flex flex-col sm:flex-row sm:items-center justify-between gap-4">
        <div class="min-w-0">
          <h3 class="text-base font-black text-slate-900">Selamat datang, {{ authStore.user?.full_name || 'Pengguna' }}</h3>
          <p class="text-xs text-slate-500 mt-1">
            <template v-if="selectedExamEvent">Event: <span class="font-semibold text-slate-700">{{ selectedExamEvent.title }}</span></template>
            <template v-else>Semua event. Pilih event untuk membuka menu pelaksanaan ujian.</template>
          </p>
        </div>
        <div class="flex flex-wrap gap-2.5 shrink-0">
          <button
            @click="switchTab('events')"
            :class="selectedExamEvent
              ? 'bg-white hover:bg-slate-100 text-slate-700 border border-slate-200'
              : 'bg-indigo-600 hover:bg-indigo-700 text-white shadow-xs'"
            class="px-4 py-2.5 active:scale-95 text-xs font-bold rounded-2xl transition"
          >
            {{ selectedExamEvent ? 'Ganti Event' : 'Pilih Event' }}
          </button>
          <button
            v-if="canShowTeacherQuestionSection"
            @click="switchTab('questions')"
            class="px-4 py-2.5 bg-white hover:bg-slate-100 active:scale-95 text-slate-700 text-xs font-bold rounded-2xl border border-slate-200 transition"
          >
            Bank Soal
          </button>
          <button
            v-if="canShowTeacherProctorSection"
            @click="switchTab('proctor')"
            class="px-4 py-2.5 bg-white hover:bg-slate-100 active:scale-95 text-slate-700 text-xs font-bold rounded-2xl border border-slate-200 transition"
          >
            Live Proctor
          </button>
        </div>
      </div>

      <!-- Kartu angka -->
      <div :class="['grid gap-3 sm:gap-4', teacherHomeGridClass]">
        <div class="bg-white rounded-3xl p-4 border border-slate-200 shadow-xs">
          <span class="text-xs font-semibold text-slate-500">Jadwal Saya</span>
          <div class="text-2xl font-black text-slate-900 mt-1">{{ teacherHomeSchedules.length }}</div>
          <div class="text-[11px] text-slate-500 mt-0.5">{{ teacherHomeTodayScheduleCount }} hari ini</div>
        </div>

        <template v-if="canShowTeacherQuestionSection">
          <div class="bg-white rounded-3xl p-4 border border-slate-200 shadow-xs">
            <span class="text-xs font-semibold text-slate-500">Bank Soal Saya</span>
            <div class="text-2xl font-black text-slate-900 mt-1">{{ teacherHomeBanks.length }}</div>
            <div class="text-[11px] text-slate-500 mt-0.5">{{ teacherHomeLockedBankCount }} terkunci · {{ teacherHomeDraftBankCount }} draf</div>
          </div>
          <div :class="['rounded-3xl p-4 border shadow-xs', teacherHomeSchedulesWithoutBank.length > 0 ? 'bg-amber-50/40 border-amber-200' : 'bg-white border-slate-200']">
            <span :class="['text-xs font-semibold', teacherHomeSchedulesWithoutBank.length > 0 ? 'text-amber-700' : 'text-slate-500']">Perlu Dibuat</span>
            <div :class="['text-2xl font-black mt-1', teacherHomeSchedulesWithoutBank.length > 0 ? 'text-amber-700' : 'text-slate-900']">{{ teacherHomeSchedulesWithoutBank.length }}</div>
            <div class="text-[11px] text-slate-500 mt-0.5">jadwal belum punya bank soal</div>
          </div>
          <div :class="['rounded-3xl p-4 border shadow-xs', teacherHomeUnfinishedBanks.length > 0 ? 'bg-amber-50/40 border-amber-200' : 'bg-white border-slate-200']">
            <span :class="['text-xs font-semibold', teacherHomeUnfinishedBanks.length > 0 ? 'text-amber-700' : 'text-slate-500']">Perlu Diselesaikan</span>
            <div :class="['text-2xl font-black mt-1', teacherHomeUnfinishedBanks.length > 0 ? 'text-amber-700' : 'text-slate-900']">{{ teacherHomeUnfinishedBanks.length }}</div>
            <div class="text-[11px] text-slate-500 mt-0.5">bank soal belum selesai</div>
          </div>
        </template>

        <div v-if="canShowTeacherProctorSection" :class="['bg-white rounded-3xl p-4 border border-slate-200 shadow-xs', teacherHomeProctorCardSpanClass]">
          <span class="text-xs font-semibold text-slate-500">Tugas Mengawasi</span>
          <div class="text-2xl font-black text-indigo-600 mt-1">{{ teacherHomeProctorTasks.length }}</div>
          <div class="text-[11px] text-slate-500 mt-0.5">{{ teacherHomeProctorTodayCount }} hari ini</div>
        </div>
      </div>

      <!-- Perlu Tindakan -->
      <div v-if="canShowTeacherQuestionSection" class="bg-white rounded-3xl p-5 border border-slate-200 shadow-xs space-y-3">
        <div class="flex items-center justify-between gap-2">
          <h3 class="text-sm font-bold text-slate-900">Perlu Tindakan</h3>
          <button
            v-if="teacherHomeActionItems.length > teacherHomeVisibleActionItems.length"
            type="button"
            @click="switchTab('questions')"
            class="text-[11px] font-bold text-indigo-600 hover:text-indigo-800 active:scale-95 transition shrink-0"
          >
            Lihat semua ({{ teacherHomeActionItems.length }})
          </button>
        </div>

        <div v-if="teacherHomeVisibleActionItems.length > 0" class="space-y-2">
          <div
            v-for="item in teacherHomeVisibleActionItems"
            :key="item.key"
            class="flex items-center justify-between gap-3 p-3 bg-slate-50 border border-slate-200 rounded-2xl"
          >
            <div class="min-w-0">
              <div class="text-xs font-bold text-slate-900 truncate">{{ item.title }}</div>
              <div class="text-[11px] mt-0.5 leading-snug">
                <span class="font-semibold text-amber-600">{{ item.label }}</span>
                <span v-if="item.detail" class="text-slate-500"> · {{ item.detail }}</span>
              </div>
            </div>
            <button
              type="button"
              @click="runTeacherHomeAction(item)"
              class="px-3 py-1.5 bg-white hover:bg-slate-100 active:scale-95 text-slate-700 border border-slate-300 font-bold text-[11px] rounded-xl transition shrink-0"
            >
              {{ item.actionLabel }}
            </button>
          </div>
        </div>
        <div v-else class="p-4 bg-emerald-50/60 rounded-2xl border border-emerald-100 text-center text-xs font-semibold text-emerald-700">
          Semua tugas Anda sudah beres.
        </div>
      </div>

      <!-- Jadwal Saya & Tugas Mengawasi -->
      <div :class="['grid grid-cols-1 gap-5 items-start', teacherHomeProctorUpcomingList.length > 0 ? 'lg:grid-cols-2' : '']">
        <div class="bg-white rounded-3xl p-5 border border-slate-200 shadow-xs space-y-3">
          <h3 class="text-sm font-bold text-slate-900">Jadwal Saya</h3>
          <div v-if="teacherHomeUpcomingList.length > 0" class="space-y-2">
            <div
              v-for="sch in teacherHomeUpcomingList"
              :key="sch.id"
              class="p-3 bg-slate-50 border border-slate-200 rounded-2xl"
            >
              <div class="flex items-start justify-between gap-3">
                <div class="min-w-0">
                  <div class="text-xs font-bold text-slate-900 leading-snug">{{ sch.title }}</div>
                  <div class="text-[11px] text-slate-500 mt-0.5">
                    {{ sch.class_room?.name || '-' }} · {{ formatScheduleTimeRange(sch.start_time, sch.end_time) }}
                    <template v-if="teacherHomeRoleLabel(sch)"> · {{ teacherHomeRoleLabel(sch) }}</template>
                  </div>
                </div>
                <span
                  v-if="teacherHomeIsInactive(sch)"
                  class="text-[11px] font-semibold text-slate-400 shrink-0"
                >Nonaktif</span>
                <span
                  v-else-if="canShowTeacherQuestionSection"
                  :class="['text-[11px] font-semibold shrink-0', teacherHomeScheduleBankStatus(sch).textClass]"
                >{{ teacherHomeScheduleBankStatus(sch).label }}</span>
              </div>
              <div v-if="Array.isArray(sch.proctors)" class="text-[11px] mt-1 leading-snug">
                <span
                  v-if="sch.proctors.length > 0"
                  class="text-slate-500"
                  :title="sch.proctors.map(p => p.full_name).join(', ')"
                >Pengawas: {{ formatProctorNames(sch.proctors) }}</span>
                <span v-else class="text-amber-600">Belum ada pengawas</span>
              </div>
            </div>
          </div>
          <div v-else class="p-4 bg-slate-50 rounded-2xl border border-slate-100 text-center text-xs text-slate-400">
            Belum ada jadwal yang ditugaskan kepada Anda.
          </div>
        </div>

        <div v-if="teacherHomeProctorUpcomingList.length > 0" class="bg-white rounded-3xl p-5 border border-slate-200 shadow-xs space-y-3">
          <h3 class="text-sm font-bold text-slate-900">Tugas Mengawasi</h3>
          <div class="space-y-2">
            <div
              v-for="sch in teacherHomeProctorUpcomingList"
              :key="sch.id"
              class="flex items-center justify-between gap-3 p-3 bg-slate-50 border border-slate-200 rounded-2xl"
            >
              <div class="min-w-0">
                <div class="text-xs font-bold text-slate-900 leading-snug">{{ sch.title }}</div>
                <div class="text-[11px] text-slate-500 mt-0.5">{{ sch.class_room?.name || '-' }} · {{ formatScheduleTimeRange(sch.start_time, sch.end_time) }}</div>
              </div>
              <button
                type="button"
                @click="switchTab('proctor', sch.id)"
                class="px-3 py-1.5 bg-indigo-50 hover:bg-indigo-100 active:scale-95 text-indigo-700 font-bold text-[11px] rounded-xl transition shrink-0"
              >
                Pantau
              </button>
            </div>
          </div>
        </div>
      </div>
    </template>

    <!-- KESIAPAN EVENT UJIAN (khusus pengelola jadwal) -->
    <div v-if="canManageSchedules && events.length > 0" class="bg-white rounded-3xl p-5 border border-slate-200/80 shadow-xs">
      <div class="flex items-center justify-between mb-3">
        <div class="text-sm font-bold text-slate-900">Kesiapan event ujian</div>
        <button
          type="button"
          @click="showEventReadiness = !showEventReadiness"
          class="text-[11px] font-bold text-indigo-600 hover:text-indigo-800 active:scale-95 transition"
        >
          {{ showEventReadiness ? 'Sembunyikan' : 'Tampilkan' }}
        </button>
      </div>

      <div v-if="showEventReadiness" class="divide-y divide-slate-100">
        <div
          v-for="event in events"
          :key="event.id"
          class="flex items-center justify-between gap-3 py-3 first:pt-0 last:pb-0"
        >
          <div class="flex items-center gap-2 min-w-0">
            <span :class="['inline-flex items-center gap-1 px-2 py-0.5 rounded-full text-[11px] font-semibold shrink-0', eventStatusBadgeClass(event)]">
              {{ eventStatusLabel(event) }}
            </span>
            <span class="text-xs text-slate-800 truncate">{{ event.title }}</span>
            <span class="text-[11px] text-slate-400 shrink-0">· {{ event.total_schedules || 0 }} jadwal</span>
          </div>
          <button
            type="button"
            @click="eventActionClick(event)"
            :class="['shrink-0 text-[11px] font-semibold px-3 py-1.5 rounded-xl transition active:scale-95', eventActionBtnClass(event)]"
          >
            {{ eventActionLabel(event) }}
          </button>
        </div>
      </div>
    </div>

    <!-- Pintasan cepat (per izin) & sesi login siswa aktif -->
    <div v-if="showQuickActions || canManageLoginSessions" class="grid grid-cols-1 lg:grid-cols-2 gap-5 items-start">
      <!-- Pintasan Cepat -->
      <div v-if="showQuickActions" class="bg-white rounded-3xl p-5 border border-slate-200 shadow-xs space-y-4">
        <h3 class="text-sm font-bold text-slate-900">Pintasan Cepat</h3>
        <div class="grid grid-cols-2 gap-2.5">
          <button
            v-if="canManageSchedules"
            @click="openCreateSchedule()"
            class="p-3 bg-slate-50 hover:bg-indigo-50 border border-slate-200 hover:border-indigo-200 rounded-2xl text-left transition active:scale-95 group"
          >
            <div class="w-8 h-8 rounded-xl bg-indigo-600 text-white flex items-center justify-center mb-2 group-hover:scale-105 transition">
              <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                <path stroke-linecap="round" stroke-linejoin="round" d="M12 4v16m8-8H4" />
              </svg>
            </div>
            <div class="text-xs font-bold text-slate-800 group-hover:text-indigo-600">Buat Jadwal Baru</div>
            <div class="text-[10px] text-slate-500 mt-0.5">Alokasikan sesi & token</div>
          </button>

          <button
            v-if="canImportStudents"
            @click="showImportModal = true"
            class="p-3 bg-slate-50 hover:bg-emerald-50 border border-slate-200 hover:border-emerald-200 rounded-2xl text-left transition active:scale-95 group"
          >
            <div class="w-8 h-8 rounded-xl bg-emerald-600 text-white flex items-center justify-center mb-2 group-hover:scale-105 transition">
              <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                <path stroke-linecap="round" stroke-linejoin="round" d="M4 16v2a2 2 0 002 2h12a2 2 0 002-2v-2M16 8l-4-4m0 0L8 8m4-4v12" />
              </svg>
            </div>
            <div class="text-xs font-bold text-slate-800 group-hover:text-emerald-600">Impor Siswa Excel</div>
            <div class="text-[10px] text-slate-500 mt-0.5">Unggah data peserta massal</div>
          </button>

          <button
            v-if="canUploadQuestionBank"
            @click="switchTab('questions')"
            class="p-3 bg-slate-50 hover:bg-amber-50 border border-slate-200 hover:border-amber-200 rounded-2xl text-left transition active:scale-95 group"
          >
            <div class="w-8 h-8 rounded-xl bg-amber-500 text-white flex items-center justify-center mb-2 group-hover:scale-105 transition">
              <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                <path stroke-linecap="round" stroke-linejoin="round" d="M12 6.253v13m0-13C10.832 5.477 9.246 5 7.5 5S4.168 5.477 3 6.253v13C4.168 18.477 5.754 18 7.5 18s3.332.477 4.5 1.253m0-13C13.168 5.477 14.754 5 16.5 5c1.747 0 3.332.477 4.5 1.253v13C19.832 18.477 18.247 18 16.5 18c-1.746 0-3.332.477-4.5 1.253" />
              </svg>
            </div>
            <div class="text-xs font-bold text-slate-800 group-hover:text-amber-600">Unggah Bank Soal</div>
            <div class="text-[10px] text-slate-500 mt-0.5">Format file Excel .xlsx</div>
          </button>

          <button
            v-if="canManageEvents"
            @click="openCreateEvent"
            class="p-3 bg-slate-50 hover:bg-purple-50 border border-slate-200 hover:border-purple-200 rounded-2xl text-left transition active:scale-95 group"
          >
            <div class="w-8 h-8 rounded-xl bg-purple-600 text-white flex items-center justify-center mb-2 group-hover:scale-105 transition">
              <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                <path stroke-linecap="round" stroke-linejoin="round" d="M8 7V3m8 4V3m-9 8h10M5 21h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z" />
              </svg>
            </div>
            <div class="text-xs font-bold text-slate-800 group-hover:text-purple-600">Buat Event Ujian</div>
            <div class="text-[10px] text-slate-500 mt-0.5">Periode semester/tahunan</div>
          </button>
        </div>
      </div>

      <!-- Sesi Login Siswa Aktif (butuh izin kelola akun; daftar akun hanya dimuat untuk pengguna ini) -->
      <div v-if="canManageLoginSessions" class="bg-white rounded-3xl p-5 border border-slate-200 shadow-xs space-y-3">
        <div class="flex items-center justify-between gap-2">
          <h3 class="text-sm font-bold text-slate-900">Sesi Login Siswa Aktif</h3>
          <span class="text-xs font-semibold text-slate-600 shrink-0">{{ activeLoginSessions.length }} sesi aktif</span>
        </div>
        <p class="text-xs text-slate-500">
          Setiap siswa hanya aktif di satu perangkat; login baru otomatis mengeluarkan perangkat lama. Reset sesi untuk mengeluarkan siswa dari perangkat yang sedang dipakai.
        </p>

        <div v-if="activeLoginSessions.length > 0" class="space-y-2 pt-1">
          <div v-for="u in activeLoginSessions.slice(0, 3)" :key="u.id" class="flex items-center justify-between gap-2 p-2.5 bg-slate-50 border border-slate-200 rounded-2xl text-xs">
            <div class="min-w-0">
              <div class="font-bold text-slate-900 truncate">{{ u.full_name }}</div>
              <div class="text-[10px] text-slate-500 font-mono truncate">{{ u.username }} · {{ u.class_name || '-' }}</div>
            </div>
            <button
              @click="resetUserSession(u)"
              class="px-2.5 py-1 bg-white hover:bg-slate-100 active:scale-95 text-slate-700 border border-slate-300 font-bold text-[10px] rounded-xl transition shrink-0"
            >
              Reset Sesi
            </button>
          </div>
          <p v-if="activeLoginSessions.length > 3" class="text-[10px] text-slate-400">
            Menampilkan 3 dari {{ activeLoginSessions.length }} sesi aktif.
          </p>
        </div>
        <div v-else class="p-4 bg-slate-50 rounded-2xl border border-slate-100 text-center text-xs text-slate-400">
          Tidak ada sesi login siswa yang aktif.
        </div>
      </div>
    </div>

    <!-- Kesiapan bank soal (khusus pengguna yang boleh membuka menu Bank Soal) -->
    <div v-if="authStore.canAccessTab('questions') && !showWelcomeCard" class="bg-white rounded-3xl px-5 py-3.5 border border-slate-200 shadow-xs text-xs text-slate-500 font-medium">
      Bank soal terkunci: {{ lockedBanksCount }} dari {{ (readinessData.question_banks || []).length }}
    </div>
  </div>
</template>

<script setup>
import { useDashboard } from './context'

const {
  activeLoginSessions,
  activeSchedulesCount,
  activeTab,
  authStore,
  canImportStudents,
  canManageEvents,
  canManageLoginSessions,
  canManageMaster,
  canManageSchedules,
  canShowTeacherProctorSection,
  canShowTeacherQuestionSection,
  canUploadQuestionBank,
  eventActionBtnClass,
  eventActionClick,
  eventActionLabel,
  eventStatusBadgeClass,
  eventStatusLabel,
  events,
  formatProctorNames,
  formatScheduleTimeRange,
  lockedBanksCount,
  openCreateEvent,
  openCreateSchedule,
  readinessData,
  resetUserSession,
  runTeacherHomeAction,
  selectedExamEvent,
  showEventReadiness,
  showImportModal,
  showQuickActions,
  showWelcomeCard,
  stats,
  switchTab,
  teacherHomeActionItems,
  teacherHomeBanks,
  teacherHomeDraftBankCount,
  teacherHomeGridClass,
  teacherHomeIsInactive,
  teacherHomeLockedBankCount,
  teacherHomeProctorCardSpanClass,
  teacherHomeProctorTasks,
  teacherHomeProctorTodayCount,
  teacherHomeProctorUpcomingList,
  teacherHomeRoleLabel,
  teacherHomeScheduleBankStatus,
  teacherHomeSchedules,
  teacherHomeSchedulesWithoutBank,
  teacherHomeTodayScheduleCount,
  teacherHomeUnfinishedBanks,
  teacherHomeUpcomingList,
  teacherHomeVisibleActionItems,
} = useDashboard()
</script>
