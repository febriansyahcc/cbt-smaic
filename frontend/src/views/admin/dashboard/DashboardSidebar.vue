<template>
  <!-- Mobile Sidebar Backdrop -->
  <div
    v-if="isMobileSidebarOpen"
    @click="isMobileSidebarOpen = false"
    class="fixed inset-0 z-40 bg-slate-900/50 backdrop-blur-xs lg:hidden"
  ></div>

  <!-- Modern Left Sidebar -->
  <aside
    :class="[
      'fixed inset-y-0 left-0 z-50 flex flex-col bg-slate-900 text-white transition-all duration-300 ease-in-out lg:static shrink-0',
      isSidebarCollapsed ? 'w-20' : 'w-64',
      isMobileSidebarOpen ? 'translate-x-0' : '-translate-x-full lg:translate-x-0'
    ]"
  >
    <!-- Brand Logo / Header -->
    <div :class="['h-16 flex items-center border-b border-slate-800 shrink-0 transition-all', isSidebarCollapsed ? 'justify-center px-2' : 'justify-between px-4']">
      <!-- Expanded Brand Info -->
      <div v-show="!isSidebarCollapsed" class="flex items-center space-x-3 overflow-hidden">
        <img src="/logo-smic.png" alt="Logo SMAS Islamic Centre Demak" class="w-10 h-10 object-contain drop-shadow-xs shrink-0" />
        <div class="truncate">
          <div class="font-bold text-xs text-white tracking-wide leading-tight truncate">SMAS ISLAMIC CENTRE</div>
          <div class="text-[10px] text-emerald-400 font-semibold tracking-wider uppercase">{{ portalLabel }}</div>
        </div>
      </div>

      <!-- Desktop Toggle Button -->
      <button
        @click="isSidebarCollapsed = !isSidebarCollapsed"
        class="hidden lg:flex p-2 rounded-xl text-slate-400 hover:text-white hover:bg-slate-800 transition items-center justify-center cursor-pointer"
        :title="isSidebarCollapsed ? 'Perluas Sidebar' : 'Perkecil Sidebar'"
      >
        <div v-if="isSidebarCollapsed" class="flex flex-col items-center">
          <img src="/logo-smic.png" alt="Logo" class="w-8 h-8 object-contain shrink-0" />
          <svg class="w-3.5 h-3.5 text-slate-400 mt-1" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 5l7 7-7 7M5 5l7 7-7 7" />
          </svg>
        </div>
        <svg v-else class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 19l-7-7 7-7m8 14l-7-7 7-7" />
        </svg>
      </button>

      <!-- Mobile Close Button -->
      <button
        @click="isMobileSidebarOpen = false"
        class="lg:hidden p-1.5 rounded-xl text-slate-400 hover:text-white hover:bg-slate-800"
      >
        <svg class="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
        </svg>
      </button>
    </div>

    <!-- Navigation Menu Items -->
    <div class="flex-1 overflow-y-auto px-3 py-4 space-y-5">
      <!-- Menu Utama -->
      <div class="space-y-1">
        <!-- Beranda -->
        <button
          type="button"
          @click="switchTab('dashboard')"
          :class="[
            'w-full flex items-center rounded-2xl text-xs font-semibold transition',
            isSidebarCollapsed ? 'justify-center py-2.5 px-0' : 'gap-3 px-3 py-2.5',
            activeTab === 'dashboard' ? 'bg-indigo-600 text-white shadow-sm' : 'text-slate-300 hover:bg-slate-800 hover:text-white'
          ]"
          :title="isSidebarCollapsed ? 'Beranda' : ''"
        >
          <svg class="w-5 h-5 shrink-0 text-indigo-300" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 12l2-2m0 0l7-7 7 7M5 10v10a1 1 0 001 1h3m10-11l2 2m-2-2v10a1 1 0 01-1 1h-3m-6 0a1 1 0 001-1v-4a1 1 0 011-1h2a1 1 0 011 1v4a1 1 0 001 1m-6 0h6" />
          </svg>
          <span v-show="!isSidebarCollapsed" class="truncate">Beranda</span>
        </button>

        <!-- TAMPIL JIKA BELUM MEMILIH EVENT: Menu Event Ujian (Khusus Admin / Kurikulum) -->
        <button
          v-if="!selectedExamEvent && authStore.canAccessTab('events')"
          type="button"
          @click="switchTab('events')"
          :class="[
            'w-full flex items-center rounded-2xl text-xs font-semibold transition',
            isSidebarCollapsed ? 'justify-center py-2.5 px-0' : 'gap-3 px-3 py-2.5',
            activeTab === 'events' ? 'bg-indigo-600 text-white shadow-sm' : 'text-slate-300 hover:bg-slate-800 hover:text-white'
          ]"
          :title="isSidebarCollapsed ? 'Event Ujian' : ''"
        >
          <svg class="w-5 h-5 shrink-0 text-amber-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 012-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10" />
          </svg>
          <span v-show="!isSidebarCollapsed" class="truncate flex-1 text-left">Event Ujian</span>
          <span v-show="!isSidebarCollapsed" class="px-1.5 py-0.5 rounded-full text-[9px] font-bold bg-slate-700 text-slate-300">
            {{ events.length }}
          </span>
        </button>

        <!-- TAMPIL JIKA SUDAH MEMILIH EVENT: Tombol Kembali ke Semua Event Tepat di Bawah Beranda -->
        <button
          v-else-if="selectedExamEvent && authStore.canAccessTab('events')"
          type="button"
          @click="exitEventScope"
          :class="[
            'w-full flex items-center rounded-2xl text-xs font-semibold transition group cursor-pointer',
            isSidebarCollapsed ? 'justify-center py-2.5 px-0' : 'gap-3 px-3 py-2.5',
            'text-slate-400 hover:bg-slate-800 hover:text-white'
          ]"
          :title="isSidebarCollapsed ? 'Kembali ke Semua Event' : ''"
        >
          <svg class="w-5 h-5 shrink-0 text-slate-400 group-hover:text-indigo-400 group-hover:-translate-x-0.5 transition" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 19l-7-7m0 0l7-7m-7 7h18" />
          </svg>
          <span v-show="!isSidebarCollapsed" class="truncate flex-1 text-left">Semua Event</span>
        </button>
      </div>

      <div v-if="selectedExamEvent" class="w-full border-b border-slate-800/80"></div>

      <!-- Pelaksanaan Ujian: tampil setelah event dipilih, menu sesuai izin pengguna -->
      <div v-if="selectedExamEvent && (authStore.canAccessTab('schedules') || authStore.canAccessTab('questions') || authStore.canAccessTab('proctor'))">
        <div v-show="!isSidebarCollapsed" class="px-3 mb-2 text-[10px] font-bold text-slate-400 uppercase tracking-wider flex items-center justify-between">
          <span>Pelaksanaan Ujian</span>
        </div>
        <div v-show="isSidebarCollapsed" class="w-8 mx-auto my-2 border-b border-slate-800"></div>
        <nav class="space-y-1">
          <!-- 1. Jadwal Ujian -->
          <button
            v-if="authStore.canAccessTab('schedules')"
            type="button"
            @click="switchTab('schedules')"
            :class="[
              'w-full flex items-center rounded-2xl text-xs font-semibold transition',
              isSidebarCollapsed ? 'justify-center py-2.5 px-0' : 'gap-3 px-3 py-2.5',
              activeTab === 'schedules' ? 'bg-indigo-600 text-white shadow-sm' : 'text-slate-300 hover:bg-slate-800 hover:text-white'
            ]"
            :title="isSidebarCollapsed ? 'Jadwal Ujian' : ''"
          >
            <svg class="w-5 h-5 shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 7V3m8 4V3m-9 8h10M5 21h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z" />
            </svg>
            <span v-show="!isSidebarCollapsed" class="truncate">Jadwal Ujian</span>
          </button>

          <!-- 2. Bank Soal & Kesiapan -->
          <button
            v-if="authStore.canAccessTab('questions')"
            type="button"
            @click="switchTab('questions')"
            :class="[
              'w-full flex items-center rounded-2xl text-xs font-semibold transition',
              isSidebarCollapsed ? 'justify-center py-2.5 px-0' : 'gap-3 px-3 py-2.5',
              activeTab === 'questions' ? 'bg-indigo-600 text-white shadow-sm' : 'text-slate-300 hover:bg-slate-800 hover:text-white'
            ]"
            :title="isSidebarCollapsed ? 'Bank Soal & Kesiapan' : ''"
          >
            <svg class="w-5 h-5 shrink-0 text-emerald-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 6.253v13m0-13C10.832 5.477 9.246 5 7.5 5S4.168 5.477 3 6.253v13C4.168 18.477 5.754 18 7.5 18s3.332.477 4.5 1.253m0-13C13.168 5.477 14.754 5 16.5 5c1.747 0 3.332.477 4.5 1.253v13C19.832 18.477 18.247 18 16.5 18c-1.746 0-3.332.477-4.5 1.253" />
            </svg>
            <span v-show="!isSidebarCollapsed" class="truncate">Bank Soal & Kesiapan</span>
          </button>

          <!-- 3. Live Proctoring -->
          <button
            v-if="authStore.canAccessTab('proctor')"
            type="button"
            @click="switchTab('proctor')"
            :class="[
              'w-full flex items-center rounded-2xl text-xs font-semibold transition',
              isSidebarCollapsed ? 'justify-center py-2.5 px-0' : 'gap-3 px-3 py-2.5',
              activeTab === 'proctor' ? 'bg-indigo-600 text-white shadow-sm' : 'text-slate-300 hover:bg-slate-800 hover:text-white'
            ]"
            :title="isSidebarCollapsed ? 'Live Proctoring' : ''"
          >
            <svg class="w-5 h-5 shrink-0 text-amber-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 10l4.553-2.276A1 1 0 0121 8.618v6.764a1 1 0 01-1.447.894L15 14M5 18h8a2 2 0 002-2V8a2 2 0 00-2-2H5a2 2 0 00-2 2v8a2 2 0 002 2z" />
            </svg>
            <span v-show="!isSidebarCollapsed" class="truncate">Live Proctoring</span>
          </button>
        </nav>
      </div>

      <!-- Group 3: Master Data (Hanya untuk Admin / Kurikulum dengan hak akses master:manage) -->
      <div v-if="authStore.hasPermission('master:manage')">
        <div v-show="!isSidebarCollapsed" class="px-3 mb-2 text-[10px] font-bold text-slate-400 uppercase tracking-wider">
          Master Data
        </div>
        <div v-show="isSidebarCollapsed" class="w-8 mx-auto my-2 border-b border-slate-800"></div>
        <nav class="space-y-1">
          <!-- Data Siswa -->
          <button
            @click="switchTab('students')"
            :class="[
              'w-full flex items-center rounded-2xl text-xs font-semibold transition',
              isSidebarCollapsed ? 'justify-center py-2.5 px-0' : 'gap-3 px-3 py-2.5',
              activeTab === 'students' ? 'bg-indigo-600 text-white shadow-sm' : 'text-slate-300 hover:bg-slate-800 hover:text-white'
            ]"
            :title="isSidebarCollapsed ? 'Data Siswa' : ''"
          >
            <svg class="w-5 h-5 shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z" />
            </svg>
            <span v-show="!isSidebarCollapsed" class="truncate">Data Siswa</span>
          </button>

          <!-- Guru dan Staf -->
          <button
            @click="switchTab('teachers')"
            :class="[
              'w-full flex items-center rounded-2xl text-xs font-semibold transition',
              isSidebarCollapsed ? 'justify-center py-2.5 px-0' : 'gap-3 px-3 py-2.5',
              activeTab === 'teachers' ? 'bg-indigo-600 text-white shadow-sm' : 'text-slate-300 hover:bg-slate-800 hover:text-white'
            ]"
            :title="isSidebarCollapsed ? 'Guru dan Staf' : ''"
          >
            <svg class="w-5 h-5 shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 20H5a2 2 0 01-2-2V6a2 2 0 012-2h10a2 2 0 012 2v1m2 13a2 2 0 01-2-2V7m2 13a2 2 0 002-2V9a2 2 0 00-2-2h-2m-4-3H9M7 16h6M7 8h6v4H7V8z" />
            </svg>
            <span v-show="!isSidebarCollapsed" class="truncate">Guru dan Staf</span>
          </button>

          <!-- Data Kelas -->
          <button
            @click="switchTab('classes')"
            :class="[
              'w-full flex items-center rounded-2xl text-xs font-semibold transition',
              isSidebarCollapsed ? 'justify-center py-2.5 px-0' : 'gap-3 px-3 py-2.5',
              activeTab === 'classes' ? 'bg-indigo-600 text-white shadow-sm' : 'text-slate-300 hover:bg-slate-800 hover:text-white'
            ]"
            :title="isSidebarCollapsed ? 'Data Kelas' : ''"
          >
            <svg class="w-5 h-5 shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 21V5a2 2 0 00-2-2H7a2 2 0 00-2 2v16m14 0h2m-2 0h-5m-9 0H3m2 0h5M9 7h1m-1 4h1m4-4h1m-1 4h1m-5 10v-5a1 1 0 011-1h2a1 1 0 011 1v5m-4 0h4" />
            </svg>
            <span v-show="!isSidebarCollapsed" class="truncate">Data Kelas</span>
          </button>

          <!-- Mata Pelajaran -->
          <button
            @click="switchTab('subjects')"
            :class="[
              'w-full flex items-center rounded-2xl text-xs font-semibold transition',
              isSidebarCollapsed ? 'justify-center py-2.5 px-0' : 'gap-3 px-3 py-2.5',
              activeTab === 'subjects' ? 'bg-indigo-600 text-white shadow-sm' : 'text-slate-300 hover:bg-slate-800 hover:text-white'
            ]"
            :title="isSidebarCollapsed ? 'Mata Pelajaran' : ''"
          >
            <svg class="w-5 h-5 shrink-0 text-cyan-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 6.253v13m0-13C10.832 5.477 9.246 5 7.5 5S4.168 5.477 3 6.253v13C4.168 18.477 5.754 18 7.5 18s3.332.477 4.5 1.253m0-13C13.168 5.477 14.754 5 16.5 5c1.747 0 3.332.477 4.5 1.253v13C19.832 18.477 18.247 18 16.5 18c-1.746 0-3.332.477-4.5 1.253" />
            </svg>
            <span v-show="!isSidebarCollapsed" class="truncate">Mata Pelajaran</span>
          </button>

          <!-- Kelas Mapel -->
          <button
            @click="switchTab('class-subjects')"
            :class="[
              'w-full flex items-center rounded-2xl text-xs font-semibold transition',
              isSidebarCollapsed ? 'justify-center py-2.5 px-0' : 'gap-3 px-3 py-2.5',
              activeTab === 'class-subjects' ? 'bg-indigo-600 text-white shadow-sm' : 'text-slate-300 hover:bg-slate-800 hover:text-white'
            ]"
            :title="isSidebarCollapsed ? 'Kelas Mapel' : ''"
          >
            <svg class="w-5 h-5 shrink-0 text-teal-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2m-6 9l2 2 4-4" />
            </svg>
            <span v-show="!isSidebarCollapsed" class="truncate">Kelas Mapel</span>
          </button>
        </nav>
      </div>
    </div>

    <!-- Sidebar Footer / Admin Profile -->
    <div class="p-3 border-t border-slate-800 shrink-0">
      <!-- Expanded Footer -->
      <div v-if="!isSidebarCollapsed" class="flex items-center justify-between">
        <div class="flex items-center space-x-2.5 overflow-hidden">
          <div class="w-8 h-8 rounded-full bg-indigo-500 text-white flex items-center justify-center font-bold text-xs shrink-0">
            A
          </div>
          <div class="truncate">
            <div class="text-xs font-bold text-white truncate">{{ authStore.user?.full_name || 'Administrator' }}</div>
            <div class="text-[10px] text-slate-400 truncate">{{ portalRoleLabel }}</div>
          </div>
        </div>
        <button
          @click="handleLogout"
          class="p-1.5 rounded-xl text-slate-400 hover:text-rose-400 hover:bg-slate-800 transition"
          title="Keluar / Logout"
        >
          <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17 16l4-4m0 0l-4-4m4 4H7m6 4v1a3 3 0 01-3 3H6a3 3 0 01-3-3V7a3 3 0 013-3h4a3 3 0 013 3v1" />
          </svg>
        </button>
      </div>

      <!-- Collapsed Footer (Cleanly Centered) -->
      <div v-else class="flex flex-col items-center space-y-2">
        <div class="w-8 h-8 rounded-full bg-indigo-500 text-white flex items-center justify-center font-bold text-xs" :title="authStore.user?.full_name || 'Administrator'">
          A
        </div>
        <button
          @click="handleLogout"
          class="p-2 rounded-xl text-slate-400 hover:text-rose-400 hover:bg-slate-800 transition"
          title="Keluar / Logout"
        >
          <svg class="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17 16l4-4m0 0l-4-4m4 4H7m6 4v1a3 3 0 01-3 3H6a3 3 0 01-3-3V7a3 3 0 013-3h4a3 3 0 013 3v1" />
          </svg>
        </button>
      </div>
    </div>
  </aside>
</template>

<script setup>
import { useDashboard } from './context'

const {
  activeTab,
  authStore,
  events,
  exitEventScope,
  handleLogout,
  isMobileSidebarOpen,
  isSidebarCollapsed,
  portalLabel,
  portalRoleLabel,
  selectedExamEvent,
  switchTab,
} = useDashboard()
</script>
