<template>
  <div class="h-screen flex bg-slate-50 overflow-hidden font-sans">
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

    <!-- Right Area: Top Navbar + Main Content -->
    <div class="flex-1 flex flex-col min-w-0 overflow-hidden">
      <!-- Top Navbar -->
      <header class="h-16 bg-white border-b border-slate-200 px-4 sm:px-6 flex items-center justify-between shrink-0 shadow-2xs z-10">
        <div class="flex items-center space-x-3">
          <!-- Mobile Toggle Button -->
          <button
            @click="isMobileSidebarOpen = true"
            class="lg:hidden p-2 rounded-xl text-slate-600 hover:bg-slate-100 transition"
            title="Buka Menu"
          >
            <svg class="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 6h16M4 12h16M4 18h16" />
            </svg>
          </button>
          <div>
            <h1 class="text-base sm:text-lg font-bold text-slate-900 tracking-tight">
              {{ currentSectionMeta.title }}
            </h1>
          </div>
        </div>

        <!-- Right Quick Actions -->
        <div class="flex items-center space-x-2.5">
          <div
            v-if="selectedExamEvent"
            @click="switchTab('events')"
            class="hidden md:flex items-center space-x-2 px-3 py-1.5 bg-indigo-50 hover:bg-indigo-100 border border-indigo-200 rounded-2xl text-indigo-900 text-xs font-semibold cursor-pointer transition shadow-2xs"
            title="Klik untuk mengelola atau berpindah event ujian"
          >
            <span class="w-2 h-2 rounded-full" :class="selectedExamEvent.is_active ? 'bg-emerald-500 animate-pulse' : 'bg-slate-400'"></span>
            <span>Event: <strong>{{ selectedExamEvent.title }}</strong></span>
            <span class="font-mono text-[10px] bg-indigo-200/60 text-indigo-800 px-1.5 py-0.5 rounded-md">{{ selectedExamEvent.code }}</span>
          </div>

          <div
            v-else
            @click="switchTab('events')"
            class="hidden md:flex items-center space-x-2 px-3 py-1.5 bg-slate-100 hover:bg-slate-200 border border-slate-200 rounded-2xl text-slate-600 text-xs font-semibold cursor-pointer transition"
            title="Klik untuk memilih event ujian"
          >
            <span class="w-2 h-2 rounded-full bg-slate-400"></span>
            <span>Pilih Event Ujian</span>
          </div>
        </div>
      </header>

      <!-- Main Scrollable Area -->
      <main class="flex-1 overflow-y-auto p-4 sm:p-6 space-y-6">
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
                Setiap siswa hanya dapat login di satu perangkat. Reset sesi bila siswa berpindah perangkat.
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
              <span>Acak Token Baru</span>
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

      <!-- TAB: EVENT UJIAN (PERIODE ASESMEN) -->
      <div v-if="activeTab === 'events'" class="space-y-4">
        <!-- Accessibility & Action Toolbar (Search, Status Filters, & Buat Event Baru) -->
        <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-3 bg-white p-3 rounded-2xl border border-slate-200/80 shadow-xs">
          <!-- Search & Filter Controls -->
          <div class="flex flex-col sm:flex-row items-stretch sm:items-center gap-2.5 flex-1">
            <!-- Search Bar -->
            <div class="relative flex-1 sm:max-w-xs">
              <svg class="w-4 h-4 absolute left-3 top-1/2 -translate-y-1/2 text-slate-400 pointer-events-none" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
              </svg>
              <input
                v-model="eventSearchQuery"
                type="text"
                placeholder="Cari nama atau kode event..."
                class="w-full pl-9 pr-8 py-2 bg-slate-50 hover:bg-slate-100/80 focus:bg-white rounded-xl border border-slate-200 text-xs font-medium focus:outline-none focus:ring-2 focus:ring-indigo-500 transition"
              />
              <button
                v-if="eventSearchQuery"
                @click="eventSearchQuery = ''"
                class="absolute right-2.5 top-1/2 -translate-y-1/2 text-slate-400 hover:text-slate-600 text-xs font-bold"
              >
                ✕
              </button>
            </div>

            <!-- Status Filter Pills -->
            <div class="flex items-center gap-1 bg-slate-100 p-1 rounded-xl shrink-0 text-[11px] font-semibold">
              <button
                type="button"
                @click="eventStatusFilter = 'all'"
                :class="['px-3 py-1 rounded-lg transition', eventStatusFilter === 'all' ? 'bg-white text-indigo-700 shadow-2xs font-bold' : 'text-slate-600 hover:text-slate-900']"
              >
                Semua ({{ events.length }})
              </button>
              <button
                type="button"
                @click="eventStatusFilter = 'active'"
                :class="['px-3 py-1 rounded-lg transition', eventStatusFilter === 'active' ? 'bg-white text-emerald-700 shadow-2xs font-bold' : 'text-slate-600 hover:text-slate-900']"
              >
                Aktif ({{ activeEventsCount }})
              </button>
              <button
                type="button"
                @click="eventStatusFilter = 'archived'"
                :class="['px-3 py-1 rounded-lg transition', eventStatusFilter === 'archived' ? 'bg-white text-slate-800 shadow-2xs font-bold' : 'text-slate-600 hover:text-slate-900']"
              >
                Arsip ({{ events.length - activeEventsCount }})
              </button>
            </div>
          </div>

          <!-- Create Event Button (Tunggal: 1 icon plus SVG tanpa karakter plus ganda di teks) -->
          <button
            v-if="authStore.hasPermission('events:manage')"
            @click="openCreateEvent"
            class="px-4 py-2 bg-indigo-600 hover:bg-indigo-700 active:scale-95 text-white font-bold text-xs rounded-xl shadow-xs transition flex items-center justify-center gap-1.5 shrink-0 cursor-pointer self-stretch sm:self-auto"
          >
            <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
            </svg>
            <span>Buat Event Baru</span>
          </button>
        </div>

        <!-- Event Cards Grid -->
        <div v-if="filteredEvents.length > 0" class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
          <div
            v-for="ev in filteredEvents"
            :key="ev.id"
            @click="enterEvent(ev)"
            :class="[
              'bg-white rounded-3xl border p-5 shadow-xs transition-all flex flex-col justify-between space-y-4 relative cursor-pointer group hover:border-indigo-400 hover:shadow-md',
              selectedExamEvent?.id === ev.id ? 'border-indigo-600 ring-2 ring-indigo-500/30' : (ev.is_active ? 'border-slate-200' : 'border-slate-200 bg-slate-50/60 opacity-90')
            ]"
          >
            <div class="space-y-3">
              <!-- Top Row: Badges & 3-Dots Action Menu -->
              <div class="flex items-center justify-between gap-2">
                <div class="flex items-center flex-wrap gap-1.5">
                  <span class="px-2 py-0.5 bg-slate-100 text-slate-700 font-mono text-[10px] font-bold rounded-md">
                    {{ ev.code }}
                  </span>
                  <span class="px-2 py-0.5 bg-indigo-50 text-indigo-700 text-[10px] font-bold rounded-md">
                    T.A. {{ ev.academic_year }} • {{ ev.semester }}
                  </span>
                  <span v-if="selectedExamEvent?.id === ev.id" class="px-2 py-0.5 bg-indigo-600 text-white text-[10px] font-black rounded-md flex items-center gap-1 uppercase tracking-wider">
                    🎯 Terpilih
                  </span>
                  <span v-else-if="ev.is_active" class="px-2 py-0.5 bg-emerald-50 text-emerald-700 border border-emerald-200 text-[10px] font-bold rounded-md">
                    Berlangsung
                  </span>
                  <span v-else class="px-2 py-0.5 bg-slate-100 text-slate-500 border border-slate-200 text-[10px] font-bold rounded-md">
                    Diarsipkan
                  </span>
                </div>

                <!-- 3-Dots Dropdown Trigger -->
                <div v-if="authStore.hasPermission('events:manage')" class="relative shrink-0" @click.stop>
                  <button
                    type="button"
                    @click.stop="toggleEventMenu(ev.id)"
                    class="w-8 h-8 flex items-center justify-center rounded-xl text-slate-400 hover:text-slate-700 hover:bg-slate-100 transition cursor-pointer"
                    title="Opsi Event"
                  >
                    <svg class="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 5v.01M12 12v.01M12 19v.01M12 6a1 1 0 110-2 1 1 0 010 2zm0 7a1 1 0 110-2 1 1 0 010 2zm0 7a1 1 0 110-2 1 1 0 010 2z" />
                    </svg>
                  </button>

                  <!-- Dropdown Menu -->
                  <div
                    v-if="activeEventMenuId === ev.id"
                    class="absolute right-0 mt-1 w-44 bg-white rounded-2xl shadow-xl border border-slate-100 py-1.5 z-20"
                  >
                    <button
                      type="button"
                      @click.stop="openEditEvent(ev); activeEventMenuId = null"
                      class="w-full text-left px-3.5 py-2 text-xs font-semibold text-slate-700 hover:bg-indigo-50 hover:text-indigo-600 flex items-center gap-2 transition cursor-pointer"
                    >
                      <svg class="w-4 h-4 text-slate-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z" />
                      </svg>
                      <span>Edit Event</span>
                    </button>
                    <button
                      type="button"
                      @click.stop="toggleEventActive(ev); activeEventMenuId = null"
                      class="w-full text-left px-3.5 py-2 text-xs font-semibold text-slate-700 hover:bg-slate-50 flex items-center gap-2 transition cursor-pointer"
                    >
                      <svg class="w-4 h-4 text-slate-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 8h14M5 8a2 2 0 110-4h14a2 2 0 110 4M5 8v10a2 2 0 002 2h10a2 2 0 002-2V8m-9 4h4" />
                      </svg>
                      <span>{{ ev.is_active ? 'Arsipkan Event' : 'Aktifkan Event' }}</span>
                    </button>
                    <div class="my-1 border-t border-slate-100"></div>
                    <button
                      type="button"
                      @click.stop="deleteEvent(ev); activeEventMenuId = null"
                      class="w-full text-left px-3.5 py-2 text-xs font-semibold text-rose-600 hover:bg-rose-50 flex items-center gap-2 transition cursor-pointer"
                    >
                      <svg class="w-4 h-4 text-rose-500" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
                      </svg>
                      <span>Hapus Event</span>
                    </button>
                  </div>
                </div>
              </div>

              <!-- Title & Description -->
              <h3 class="text-sm font-bold text-slate-900 leading-snug group-hover:text-indigo-600 transition">
                {{ ev.title }}
              </h3>
              <p v-if="ev.description && ev.description.toLowerCase() !== ev.title.toLowerCase()" class="text-xs text-slate-500 line-clamp-2">
                {{ ev.description }}
              </p>

              <!-- Dates & Schedules -->
              <div class="pt-2 border-t border-slate-100 grid grid-cols-2 gap-2 text-xs">
                <div>
                  <span class="text-[10px] text-slate-400 block font-semibold uppercase">Periode Tanggal</span>
                  <span class="font-medium text-slate-700 text-[11px]">
                    {{ formatDate(ev.start_date) }} - {{ formatDate(ev.end_date) }}
                  </span>
                </div>
                <div>
                  <span class="text-[10px] text-slate-400 block font-semibold uppercase">Total Jadwal</span>
                  <span class="font-bold text-indigo-700 font-mono text-[11px]">
                    {{ ev.total_schedules || 0 }} Sesi Ujian
                  </span>
                </div>
              </div>
            </div>

            <!-- Footer Action indicator -->
            <div class="pt-3 flex items-center justify-end border-t border-slate-100 text-xs">
              <span class="inline-flex items-center gap-1 text-xs font-bold text-indigo-600 group-hover:text-indigo-700 transition">
                <span>Buka Event</span>
                <svg class="w-4 h-4 transform group-hover:translate-x-1 transition duration-150" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 7l5 5m0 0l-5 5m5-5H6" />
                </svg>
              </span>
            </div>
          </div>
        </div>

        <!-- Empty State jika pencarian tidak menemukan hasil -->
        <div v-else-if="events.length > 0" class="p-10 text-center bg-white rounded-3xl border border-slate-200 shadow-xs space-y-2">
          <div class="text-3xl">🔍</div>
          <h3 class="text-sm font-bold text-slate-800">Tidak Ada Event yang Sesuai</h3>
          <p class="text-xs text-slate-500">
            Tidak ditemukan event dengan kata kunci "{{ eventSearchQuery }}".
          </p>
          <button
            @click="eventSearchQuery = ''; eventStatusFilter = 'all'"
            class="px-3.5 py-1.5 bg-slate-100 hover:bg-slate-200 text-slate-700 font-bold text-xs rounded-xl transition cursor-pointer"
          >
            Reset Filter
          </button>
        </div>

        <!-- Empty State jika belum ada event sama sekali -->
        <div v-else class="p-12 text-center bg-white rounded-3xl border border-slate-200 shadow-xs space-y-3">
          <div class="text-4xl">🎯</div>
          <h3 class="text-sm font-bold text-slate-800">Belum Ada Event Ujian Terdaftar</h3>
          <p class="text-xs text-slate-500 max-w-md mx-auto">
            Buat event ujian baru untuk mulai mengatur sesi jadwal pelaksanaan ujian dan naskah bank soal.
          </p>
          <button
            v-if="authStore.hasPermission('events:manage')"
            @click="openCreateEvent"
            class="px-4 py-2 bg-indigo-600 hover:bg-indigo-700 text-white font-bold text-xs rounded-xl shadow-xs transition"
          >
            Buat Event Baru
          </button>
        </div>
      </div>

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

        <!-- TAB: DATA GURU DAN STAF -->
        <div v-if="activeTab === 'teachers'" class="space-y-4">
          <!-- Contextual Cards: Teachers -->
          <div class="grid grid-cols-2 sm:grid-cols-4 gap-3">
            <div class="bg-white rounded-3xl p-4 border border-slate-200 shadow-xs">
              <span class="text-[11px] font-semibold text-slate-500">Total Guru dan Staf</span>
              <div class="text-2xl font-black text-slate-900 mt-1">{{ teachers.length }} <span class="text-xs font-medium text-slate-400">Akun</span></div>
            </div>
            <div class="bg-white rounded-3xl p-4 border border-indigo-200 shadow-xs bg-indigo-50/20">
              <span class="text-[11px] font-semibold text-indigo-700">Guru dan Pengawas</span>
              <div class="text-2xl font-black text-indigo-700 mt-1">{{ guruCount }} <span class="text-xs font-medium text-indigo-400">Akun</span></div>
            </div>
            <div class="bg-white rounded-3xl p-4 border border-purple-200 shadow-xs bg-purple-50/20">
              <span class="text-[11px] font-semibold text-purple-700">Administrator</span>
              <div class="text-2xl font-black text-purple-700 mt-1">{{ adminStaffCount }} <span class="text-xs font-medium text-purple-400">Akun</span></div>
            </div>
            <div class="bg-white rounded-3xl p-4 border border-emerald-200 shadow-xs bg-emerald-50/20">
              <span class="text-[11px] font-semibold text-emerald-700">Penyusun Bank Soal</span>
              <div class="text-2xl font-black text-emerald-700 mt-1">{{ questionMakersCount }} <span class="text-xs font-medium text-emerald-400">Guru</span></div>
            </div>
          </div>

          <!-- Teacher Action & Filter Toolbar -->
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
                    v-model="teacherSearchQuery"
                    type="text"
                    placeholder="Cari nama, username guru..."
                    class="w-full pl-9 pr-8 py-2 bg-slate-50 border border-slate-200 rounded-2xl text-xs text-slate-800 placeholder-slate-400 focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:bg-white transition"
                  />
                  <button
                    v-if="teacherSearchQuery"
                    @click="teacherSearchQuery = ''"
                    class="absolute inset-y-0 right-0 pr-2.5 flex items-center text-slate-400 hover:text-slate-600 cursor-pointer"
                  >
                    ✕
                  </button>
                </div>

                <!-- Role Filter Pills -->
                <div class="flex items-center gap-1.5 flex-wrap">
                  <button
                    v-for="chip in teacherRoleChips"
                    :key="chip.key"
                    type="button"
                    @click="selectedTeacherRole = chip.key"
                    :class="[
                      'px-3 py-1.5 rounded-xl text-xs font-bold transition cursor-pointer active:scale-95',
                      selectedTeacherRole === chip.key
                        ? (chip.key === 'admin' ? 'bg-purple-600 text-white shadow-xs' : 'bg-indigo-600 text-white shadow-xs')
                        : 'bg-slate-100 text-slate-600 hover:bg-slate-200'
                    ]"
                  >
                    {{ chip.label }}
                  </button>
                </div>
              </div>

              <!-- Action Buttons -->
              <div class="flex items-center justify-end gap-2">
                <button
                  v-if="canManageMaster"
                  @click="showImportTeacherModal = true"
                  class="px-3.5 py-2 bg-emerald-600 hover:bg-emerald-700 active:scale-95 text-white font-bold text-xs rounded-xl shadow-xs transition flex items-center space-x-1 cursor-pointer shrink-0"
                >
                  <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-8l-4-4m0 0L8 8m4-4v12" />
                  </svg>
                  <span>Impor Excel</span>
                </button>
                <button
                  @click="openCreateTeacherModal"
                  class="px-3.5 py-2 bg-indigo-600 hover:bg-indigo-700 active:scale-95 text-white font-bold text-xs rounded-xl shadow-xs transition flex items-center space-x-1 cursor-pointer shrink-0"
                >
                  <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
                  </svg>
                  <span>Tambah Guru / Staf</span>
                </button>
              </div>
            </div>
          </div>

          <!-- Teachers Table -->
          <div class="bg-white rounded-3xl border border-slate-200 shadow-sm overflow-hidden">
            <div class="overflow-x-auto">
              <table class="w-full text-left text-xs">
                <thead>
                  <tr class="bg-slate-50 border-b border-slate-200 text-slate-600 font-bold uppercase text-[10px] select-none">
                    <th class="py-3 px-4">
                      <button
                        type="button"
                        @click="toggleTeacherSort('name')"
                        class="group inline-flex items-center gap-1 cursor-pointer transition hover:text-indigo-600 uppercase font-bold"
                        :class="teacherSortKey === 'name' ? 'text-indigo-600' : 'text-slate-600'"
                      >
                        <span>Nama Lengkap & Gelar</span>
                        <svg v-if="teacherSortKey === 'name' && teacherSortOrder === 'asc'" class="w-3 h-3 text-indigo-600" viewBox="0 0 20 20" fill="currentColor"><path fill-rule="evenodd" d="M5.293 9.707a1 1 0 010-1.414l4-4a1 1 0 011.414 0l4 4a1 1 0 01-1.414 1.414L11 7.414V15a1 1 0 11-2 0V7.414L6.707 9.707a1 1 0 01-1.414 0z" clip-rule="evenodd" /></svg>
                        <svg v-else-if="teacherSortKey === 'name' && teacherSortOrder === 'desc'" class="w-3 h-3 text-indigo-600" viewBox="0 0 20 20" fill="currentColor"><path fill-rule="evenodd" d="M14.707 10.293a1 1 0 010 1.414l-4 4a1 1 0 01-1.414 0l-4-4a1 1 0 111.414-1.414L9 12.586V5a1 1 0 012 0v7.586l2.293-2.293a1 1 0 011.414 0z" clip-rule="evenodd" /></svg>
                        <svg v-else class="w-3 h-3 text-slate-300 group-hover:text-slate-400" viewBox="0 0 20 20" fill="currentColor"><path d="M5 8l5-5 5 5H5zm10 4l-5 5-5-5h10z" /></svg>
                      </button>
                    </th>
                    <th class="py-3 px-4">
                      <button
                        type="button"
                        @click="toggleTeacherSort('username')"
                        class="group inline-flex items-center gap-1 cursor-pointer transition hover:text-indigo-600 uppercase font-bold"
                        :class="teacherSortKey === 'username' ? 'text-indigo-600' : 'text-slate-600'"
                      >
                        <span>Username Login</span>
                        <svg v-if="teacherSortKey === 'username' && teacherSortOrder === 'asc'" class="w-3 h-3 text-indigo-600" viewBox="0 0 20 20" fill="currentColor"><path fill-rule="evenodd" d="M5.293 9.707a1 1 0 010-1.414l4-4a1 1 0 011.414 0l4 4a1 1 0 01-1.414 1.414L11 7.414V15a1 1 0 11-2 0V7.414L6.707 9.707a1 1 0 01-1.414 0z" clip-rule="evenodd" /></svg>
                        <svg v-else-if="teacherSortKey === 'username' && teacherSortOrder === 'desc'" class="w-3 h-3 text-indigo-600" viewBox="0 0 20 20" fill="currentColor"><path fill-rule="evenodd" d="M14.707 10.293a1 1 0 010 1.414l-4 4a1 1 0 01-1.414 0l-4-4a1 1 0 111.414-1.414L9 12.586V5a1 1 0 012 0v7.586l2.293-2.293a1 1 0 011.414 0z" clip-rule="evenodd" /></svg>
                        <svg v-else class="w-3 h-3 text-slate-300 group-hover:text-slate-400" viewBox="0 0 20 20" fill="currentColor"><path d="M5 8l5-5 5 5H5zm10 4l-5 5-5-5h10z" /></svg>
                      </button>
                    </th>
                    <th class="py-3 px-4">
                      <button
                        type="button"
                        @click="toggleTeacherSort('role')"
                        class="group inline-flex items-center gap-1 cursor-pointer transition hover:text-indigo-600 uppercase font-bold"
                        :class="teacherSortKey === 'role' ? 'text-indigo-600' : 'text-slate-600'"
                      >
                        <span>Peran</span>
                        <svg v-if="teacherSortKey === 'role' && teacherSortOrder === 'asc'" class="w-3 h-3 text-indigo-600" viewBox="0 0 20 20" fill="currentColor"><path fill-rule="evenodd" d="M5.293 9.707a1 1 0 010-1.414l4-4a1 1 0 011.414 0l4 4a1 1 0 01-1.414 1.414L11 7.414V15a1 1 0 11-2 0V7.414L6.707 9.707a1 1 0 01-1.414 0z" clip-rule="evenodd" /></svg>
                        <svg v-else-if="teacherSortKey === 'role' && teacherSortOrder === 'desc'" class="w-3 h-3 text-indigo-600" viewBox="0 0 20 20" fill="currentColor"><path fill-rule="evenodd" d="M14.707 10.293a1 1 0 010 1.414l-4 4a1 1 0 01-1.414 0l-4-4a1 1 0 111.414-1.414L9 12.586V5a1 1 0 012 0v7.586l2.293-2.293a1 1 0 011.414 0z" clip-rule="evenodd" /></svg>
                        <svg v-else class="w-3 h-3 text-slate-300 group-hover:text-slate-400" viewBox="0 0 20 20" fill="currentColor"><path d="M5 8l5-5 5 5H5zm10 4l-5 5-5-5h10z" /></svg>
                      </button>
                    </th>
                    <th class="py-3 px-4 text-center">Status</th>
                    <th class="py-3 px-4 text-right">Aksi</th>
                  </tr>
                </thead>
                <tbody class="divide-y divide-slate-100">
                  <tr v-if="filteredTeachers.length === 0">
                    <td colspan="5" class="py-8 text-center text-slate-400">
                      <div v-if="teachers.length === 0">Belum ada data guru/staf terdaftar.</div>
                      <div v-else class="space-y-1.5">
                        <div class="font-bold text-slate-700 text-xs">Tidak ada data guru yang sesuai filter</div>
                        <button
                          @click="teacherSearchQuery = ''; selectedTeacherRole = 'all'; teacherCurrentPage = 1"
                          class="px-3 py-1 bg-slate-100 hover:bg-slate-200 text-slate-700 font-bold text-xs rounded-xl transition cursor-pointer mt-1"
                        >
                          Reset Filter
                        </button>
                      </div>
                    </td>
                  </tr>
                  <tr v-for="g in paginatedTeachers" :key="g.id" class="hover:bg-slate-50/60 transition">
                    <td class="py-3 px-4 font-bold text-slate-900">{{ g.full_name }}</td>
                    <td class="py-3 px-4 font-mono text-slate-600 text-xs">{{ g.username }}</td>
                    <td class="py-3 px-4">
                      <span :class="['px-2 py-0.5 rounded-full text-[10px] font-bold', teacherBadgeClass(g)]">
                        {{ teacherRoleLabel(g) }}
                      </span>
                    </td>
                    <td class="py-3 px-4 text-center">
                      <span class="px-2 py-0.5 rounded-full text-[10px] font-bold bg-emerald-100 text-emerald-800">
                        Aktif
                      </span>
                    </td>
                    <td class="py-3 px-4 text-right">
                      <div v-if="!isGrantor && g.role === 'ADMIN'" class="text-[11px] text-slate-400 text-right">Dikelola administrator</div>
                      <div v-else class="flex items-center justify-end gap-1.5">
                        <button
                          @click="openEditTeacher(g)"
                          class="p-1.5 bg-slate-100 hover:bg-slate-200 text-slate-600 rounded-xl transition cursor-pointer"
                          title="Edit Data Guru / Staf"
                        >
                          <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z" />
                          </svg>
                        </button>
                        <button
                          @click="deleteTeacher(g)"
                          class="p-1.5 bg-rose-50 hover:bg-rose-100 text-rose-600 rounded-xl transition cursor-pointer"
                          title="Hapus Akun Guru / Staf"
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
            <div v-if="filteredTeachers.length > 0" class="border-t border-slate-200/80 bg-slate-50/60 px-4 py-3 flex flex-col sm:flex-row items-center justify-between gap-3 text-xs text-slate-600 select-none">
              <div class="flex items-center gap-3 flex-wrap justify-center sm:justify-start">
                <span>
                  Menampilkan
                  <span class="font-bold text-slate-800">{{ (teacherCurrentPage - 1) * teacherPerPage + 1 }}</span>
                  –
                  <span class="font-bold text-slate-800">{{ Math.min(teacherCurrentPage * teacherPerPage, filteredTeachers.length) }}</span>
                  dari
                  <span class="font-bold text-slate-800">{{ filteredTeachers.length }}</span> akun
                </span>

                <div class="flex items-center gap-1.5 pl-2.5 border-l border-slate-200">
                  <span class="text-slate-400 text-[11px]">Tampilkan:</span>
                  <select
                    v-model.number="teacherPerPage"
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
              <div v-if="totalTeacherPages > 1" class="flex items-center gap-1">
                <button
                  type="button"
                  @click="setTeacherPage(teacherCurrentPage - 1)"
                  :disabled="teacherCurrentPage === 1"
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
                    v-for="(p, idx) in displayedTeacherPages"
                    :key="idx"
                    type="button"
                    @click="setTeacherPage(p)"
                    :disabled="p === '...'"
                    :class="[
                      'min-w-[28px] h-7 px-2 rounded-lg text-xs font-bold transition flex items-center justify-center',
                      p === '...' ? 'cursor-default text-slate-400' :
                      p === teacherCurrentPage ? 'bg-indigo-600 text-white shadow-xs' : 'bg-white border border-slate-200 text-slate-700 hover:bg-slate-100 cursor-pointer shadow-2xs'
                    ]"
                  >
                    {{ p }}
                  </button>
                </div>

                <button
                  type="button"
                  @click="setTeacherPage(teacherCurrentPage + 1)"
                  :disabled="teacherCurrentPage === totalTeacherPages"
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

        <!-- TAB: DATA MASTER KELAS / ROMBEL -->
        <div v-if="activeTab === 'classes'" class="space-y-4">
          <!-- Contextual Cards: Classes -->
          <div class="grid grid-cols-2 sm:grid-cols-4 gap-3">
            <div class="bg-white rounded-3xl p-4 border border-slate-200 shadow-xs">
              <span class="text-[11px] font-semibold text-slate-500">Total Rombel</span>
              <div class="text-2xl font-black text-slate-900 mt-1">{{ classes.length }} <span class="text-xs font-medium text-slate-400">Kelas</span></div>
            </div>
            <div class="bg-white rounded-3xl p-4 border border-blue-200 shadow-xs bg-blue-50/20">
              <span class="text-[11px] font-semibold text-blue-700">Tingkat XII</span>
              <div class="text-2xl font-black text-blue-700 mt-1">{{ gradeCounts['XII'] || 0 }} <span class="text-xs font-medium text-blue-400">Rombel</span></div>
            </div>
            <div class="bg-white rounded-3xl p-4 border border-purple-200 shadow-xs bg-purple-50/20">
              <span class="text-[11px] font-semibold text-purple-700">Tingkat XI</span>
              <div class="text-2xl font-black text-purple-700 mt-1">{{ gradeCounts['XI'] || 0 }} <span class="text-xs font-medium text-purple-400">Rombel</span></div>
            </div>
            <div class="bg-white rounded-3xl p-4 border border-emerald-200 shadow-xs bg-emerald-50/20">
              <span class="text-[11px] font-semibold text-emerald-700">Tingkat X</span>
              <div class="text-2xl font-black text-emerald-700 mt-1">{{ gradeCounts['X'] || 0 }} <span class="text-xs font-medium text-emerald-400">Rombel</span></div>
            </div>
          </div>

          <!-- Class Action & Filter Toolbar -->
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
                    v-model="classSearchQuery"
                    type="text"
                    placeholder="Cari nama kelas, tingkat, jurusan..."
                    class="w-full pl-9 pr-8 py-2 bg-slate-50 border border-slate-200 rounded-2xl text-xs text-slate-800 placeholder-slate-400 focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:bg-white transition"
                  />
                  <button
                    v-if="classSearchQuery"
                    @click="classSearchQuery = ''"
                    class="absolute inset-y-0 right-0 pr-2.5 flex items-center text-slate-400 hover:text-slate-600 cursor-pointer"
                  >
                    ✕
                  </button>
                </div>

                <!-- Grade Filter Pills -->
                <div class="flex items-center gap-1.5 flex-wrap">
                  <button
                    type="button"
                    @click="selectedClassGrade = 'all'"
                    :class="[
                      'px-3 py-1.5 rounded-xl text-xs font-bold transition cursor-pointer',
                      selectedClassGrade === 'all'
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
                    @click="selectedClassGrade = gr"
                    :class="[
                      'px-3 py-1.5 rounded-xl text-xs font-bold transition cursor-pointer',
                      selectedClassGrade === gr
                        ? 'bg-indigo-600 text-white shadow-xs'
                        : 'bg-slate-100 text-slate-600 hover:bg-slate-200'
                    ]"
                  >
                    Kelas {{ gr }}
                  </button>
                </div>
              </div>

              <!-- Action Buttons -->
              <div class="flex items-center justify-end gap-2">
                <button
                  v-if="canManageMaster"
                  @click="showImportClassModal = true"
                  class="px-3.5 py-2 bg-emerald-600 hover:bg-emerald-700 active:scale-95 text-white font-bold text-xs rounded-xl shadow-xs transition flex items-center space-x-1 cursor-pointer shrink-0"
                >
                  <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-8l-4-4m0 0L8 8m4-4v12" />
                  </svg>
                  <span>Impor Excel</span>
                </button>
                <button
                  @click="openCreateClassModal"
                  class="px-3.5 py-2 bg-indigo-600 hover:bg-indigo-700 active:scale-95 text-white font-bold text-xs rounded-xl shadow-xs transition flex items-center space-x-1 cursor-pointer shrink-0"
                >
                  <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
                  </svg>
                  <span>Tambah Kelas</span>
                </button>
              </div>
            </div>
          </div>

          <!-- Classes Table -->
          <div class="bg-white rounded-3xl border border-slate-200 shadow-sm overflow-hidden">
            <div class="overflow-x-auto">
              <table class="w-full text-left text-xs">
                <thead>
                  <tr class="bg-slate-50 border-b border-slate-200 text-slate-600 font-bold uppercase text-[10px] select-none">
                    <th class="py-3 px-4">
                      <button
                        type="button"
                        @click="toggleClassSort('name')"
                        class="group inline-flex items-center gap-1 cursor-pointer transition hover:text-indigo-600 uppercase font-bold"
                        :class="classSortKey === 'name' ? 'text-indigo-600' : 'text-slate-600'"
                      >
                        <span>Nama Kelas</span>
                        <svg v-if="classSortKey === 'name' && classSortOrder === 'asc'" class="w-3 h-3 text-indigo-600" viewBox="0 0 20 20" fill="currentColor"><path fill-rule="evenodd" d="M5.293 9.707a1 1 0 010-1.414l4-4a1 1 0 011.414 0l4 4a1 1 0 01-1.414 1.414L11 7.414V15a1 1 0 11-2 0V7.414L6.707 9.707a1 1 0 01-1.414 0z" clip-rule="evenodd" /></svg>
                        <svg v-else-if="classSortKey === 'name' && classSortOrder === 'desc'" class="w-3 h-3 text-indigo-600" viewBox="0 0 20 20" fill="currentColor"><path fill-rule="evenodd" d="M14.707 10.293a1 1 0 010 1.414l-4 4a1 1 0 01-1.414 0l-4-4a1 1 0 111.414-1.414L9 12.586V5a1 1 0 012 0v7.586l2.293-2.293a1 1 0 011.414 0z" clip-rule="evenodd" /></svg>
                        <svg v-else class="w-3 h-3 text-slate-300 group-hover:text-slate-400" viewBox="0 0 20 20" fill="currentColor"><path d="M5 8l5-5 5 5H5zm10 4l-5 5-5-5h10z" /></svg>
                      </button>
                    </th>
                    <th class="py-3 px-4">
                      <button
                        type="button"
                        @click="toggleClassSort('grade')"
                        class="group inline-flex items-center gap-1 cursor-pointer transition hover:text-indigo-600 uppercase font-bold"
                        :class="classSortKey === 'grade' ? 'text-indigo-600' : 'text-slate-600'"
                      >
                        <span>Tingkat (Grade)</span>
                        <svg v-if="classSortKey === 'grade' && classSortOrder === 'asc'" class="w-3 h-3 text-indigo-600" viewBox="0 0 20 20" fill="currentColor"><path fill-rule="evenodd" d="M5.293 9.707a1 1 0 010-1.414l4-4a1 1 0 011.414 0l4 4a1 1 0 01-1.414 1.414L11 7.414V15a1 1 0 11-2 0V7.414L6.707 9.707a1 1 0 01-1.414 0z" clip-rule="evenodd" /></svg>
                        <svg v-else-if="classSortKey === 'grade' && classSortOrder === 'desc'" class="w-3 h-3 text-indigo-600" viewBox="0 0 20 20" fill="currentColor"><path fill-rule="evenodd" d="M14.707 10.293a1 1 0 010 1.414l-4 4a1 1 0 01-1.414 0l-4-4a1 1 0 111.414-1.414L9 12.586V5a1 1 0 012 0v7.586l2.293-2.293a1 1 0 011.414 0z" clip-rule="evenodd" /></svg>
                        <svg v-else class="w-3 h-3 text-slate-300 group-hover:text-slate-400" viewBox="0 0 20 20" fill="currentColor"><path d="M5 8l5-5 5 5H5zm10 4l-5 5-5-5h10z" /></svg>
                      </button>
                    </th>
                    <th class="py-3 px-4">
                      <button
                        type="button"
                        @click="toggleClassSort('major')"
                        class="group inline-flex items-center gap-1 cursor-pointer transition hover:text-indigo-600 uppercase font-bold"
                        :class="classSortKey === 'major' ? 'text-indigo-600' : 'text-slate-600'"
                      >
                        <span>Jurusan</span>
                        <svg v-if="classSortKey === 'major' && classSortOrder === 'asc'" class="w-3 h-3 text-indigo-600" viewBox="0 0 20 20" fill="currentColor"><path fill-rule="evenodd" d="M5.293 9.707a1 1 0 010-1.414l4-4a1 1 0 011.414 0l4 4a1 1 0 01-1.414 1.414L11 7.414V15a1 1 0 11-2 0V7.414L6.707 9.707a1 1 0 01-1.414 0z" clip-rule="evenodd" /></svg>
                        <svg v-else-if="classSortKey === 'major' && classSortOrder === 'desc'" class="w-3 h-3 text-indigo-600" viewBox="0 0 20 20" fill="currentColor"><path fill-rule="evenodd" d="M14.707 10.293a1 1 0 010 1.414l-4 4a1 1 0 01-1.414 0l-4-4a1 1 0 111.414-1.414L9 12.586V5a1 1 0 012 0v7.586l2.293-2.293a1 1 0 011.414 0z" clip-rule="evenodd" /></svg>
                        <svg v-else class="w-3 h-3 text-slate-300 group-hover:text-slate-400" viewBox="0 0 20 20" fill="currentColor"><path d="M5 8l5-5 5 5H5zm10 4l-5 5-5-5h10z" /></svg>
                      </button>
                    </th>
                    <th class="py-3 px-4 text-right">Aksi</th>
                  </tr>
                </thead>
                <tbody class="divide-y divide-slate-100">
                  <tr v-if="filteredClasses.length === 0">
                    <td colspan="4" class="py-8 text-center text-slate-400">
                      <div v-if="classes.length === 0">Belum ada kelas terdaftar.</div>
                      <div v-else class="space-y-1.5">
                        <div class="font-bold text-slate-700 text-xs">Tidak ada data kelas yang sesuai filter</div>
                        <button
                          @click="classSearchQuery = ''; selectedClassGrade = 'all'; classCurrentPage = 1"
                          class="px-3 py-1 bg-slate-100 hover:bg-slate-200 text-slate-700 font-bold text-xs rounded-xl transition cursor-pointer mt-1"
                        >
                          Reset Filter
                        </button>
                      </div>
                    </td>
                  </tr>
                  <tr v-for="c in paginatedClasses" :key="c.id" class="hover:bg-slate-50/60 transition">
                    <td class="py-3 px-4 font-bold text-slate-900">{{ c.name }}</td>
                    <td class="py-3 px-4 font-mono font-bold text-indigo-700">{{ c.grade }}</td>
                    <td class="py-3 px-4 font-semibold text-slate-700">{{ c.major || '-' }}</td>
                    <td class="py-3 px-4 text-right">
                      <div class="flex items-center justify-end gap-1.5">
                        <button
                          @click="openEditClass(c)"
                          class="p-1.5 bg-slate-100 hover:bg-slate-200 text-slate-600 rounded-xl transition cursor-pointer"
                          title="Edit Kelas"
                        >
                          <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z" />
                          </svg>
                        </button>
                        <button
                          @click="deleteClass(c)"
                          class="p-1.5 bg-rose-50 hover:bg-rose-100 text-rose-600 rounded-xl transition cursor-pointer"
                          title="Hapus Kelas"
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
            <div v-if="filteredClasses.length > 0" class="border-t border-slate-200/80 bg-slate-50/60 px-4 py-3 flex flex-col sm:flex-row items-center justify-between gap-3 text-xs text-slate-600 select-none">
              <div class="flex items-center gap-3 flex-wrap justify-center sm:justify-start">
                <span>
                  Menampilkan
                  <span class="font-bold text-slate-800">{{ (classCurrentPage - 1) * classPerPage + 1 }}</span>
                  –
                  <span class="font-bold text-slate-800">{{ Math.min(classCurrentPage * classPerPage, filteredClasses.length) }}</span>
                  dari
                  <span class="font-bold text-slate-800">{{ filteredClasses.length }}</span> kelas
                </span>

                <div class="flex items-center gap-1.5 pl-2.5 border-l border-slate-200">
                  <span class="text-slate-400 text-[11px]">Tampilkan:</span>
                  <select
                    v-model.number="classPerPage"
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
              <div v-if="totalClassPages > 1" class="flex items-center gap-1">
                <button
                  type="button"
                  @click="setClassPage(classCurrentPage - 1)"
                  :disabled="classCurrentPage === 1"
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
                    v-for="(p, idx) in displayedClassPages"
                    :key="idx"
                    type="button"
                    @click="setClassPage(p)"
                    :disabled="p === '...'"
                    :class="[
                      'min-w-[28px] h-7 px-2 rounded-lg text-xs font-bold transition flex items-center justify-center',
                      p === '...' ? 'cursor-default text-slate-400' :
                      p === classCurrentPage ? 'bg-indigo-600 text-white shadow-xs' : 'bg-white border border-slate-200 text-slate-700 hover:bg-slate-100 cursor-pointer shadow-2xs'
                    ]"
                  >
                    {{ p }}
                  </button>
                </div>

                <button
                  type="button"
                  @click="setClassPage(classCurrentPage + 1)"
                  :disabled="classCurrentPage === totalClassPages"
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

        <!-- (Tab readiness was merged into tab questions) -->

        <!-- TAB: LIVE PROCTORING & PENGAWASAN RUANG UJIAN -->
        <div v-if="activeTab === 'proctor'" class="space-y-4">
          <LiveProctorControl
            :schedules="proctorSchedules"
            :initial-schedule-id="activeProctorScheduleId"
            @schedule-changed="(id) => { activeProctorScheduleId = id }"
            @data-refreshed="(data) => { proctorData = data }"
          >
            <template #header-actions="{ scheduleId }">
              <!-- Export Rekap Nilai Excel -->
              <button
                v-if="scheduleId && authStore.hasPermission('reports:export')"
                @click="exportNilaiExcel"
                :disabled="isExportingNilai"
                class="px-3.5 py-2 bg-emerald-600 hover:bg-emerald-700 active:scale-95 text-white text-xs font-bold rounded-xl shadow-xs transition flex items-center gap-1.5 cursor-pointer"
              >
                <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 10v6m0 0l-3-3m3 3l3-3m2 8H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
                </svg>
                <span>{{ isExportingNilai ? 'Mengunduh...' : 'Rekap Nilai (.xlsx)' }}</span>
              </button>

              <!-- Export Berita Acara PDF -->
              <button
                v-if="scheduleId && authStore.hasPermission('reports:export')"
                @click="exportBeritaAcaraPDF"
                :disabled="isExportingPDF"
                class="px-3.5 py-2 bg-indigo-600 hover:bg-indigo-700 active:scale-95 text-white text-xs font-bold rounded-xl shadow-xs transition flex items-center gap-1.5 cursor-pointer"
              >
                <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M7 21h10a2 2 0 002-2V9.414a1 1 0 00-.293-.707l-5.414-5.414A1 1 0 0012.586 3H7a2 2 0 00-2 2v14a2 2 0 002 2z" />
                </svg>
                <span>{{ isExportingPDF ? 'Menyiapkan...' : 'Berita Acara (.pdf)' }}</span>
              </button>

              <!-- Cetak Daftar Hadir & Berita Acara -->
              <button
                v-if="scheduleId"
                @click="showProctorPrint = true"
                class="px-3.5 py-2 bg-slate-800 hover:bg-slate-900 active:scale-95 text-white text-xs font-bold rounded-xl shadow-xs transition flex items-center gap-1.5 cursor-pointer"
              >
                <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17 17h2a2 2 0 002-2v-4a2 2 0 00-2-2H5a2 2 0 00-2 2v4a2 2 0 002 2h2m2 4h6a2 2 0 002-2v-4a2 2 0 00-2-2H9a2 2 0 00-2 2v4a2 2 0 002 2zm8-12V5a2 2 0 00-2-2H9a2 2 0 00-2 2v4h10z" />
                </svg>
                <span>Cetak Dokumen</span>
              </button>
            </template>
          </LiveProctorControl>
        </div>

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
                          {{ b.subject?.name || '-' }}
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
      </main>
    </div>

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
                <label class="block font-bold text-slate-700 mb-1">Token Ruang:</label>
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
                <div class="text-slate-700 font-medium leading-relaxed" v-html="q.content_html"></div>
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

    <!-- MODAL: BUAT / EDIT EVENT UJIAN -->
    <div v-if="showEventModal" class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-slate-900/60 backdrop-blur-xs">
      <div class="bg-white rounded-3xl max-w-lg w-full p-6 shadow-2xl space-y-4">
        <h3 class="text-base font-bold text-slate-900">
          {{ isEditEvent ? 'Edit Data Event Ujian' : 'Buat Event / Periode Ujian Baru' }}
        </h3>
        <form @submit.prevent="submitEventForm" class="space-y-3 text-xs">
          <div>
            <label class="block font-bold text-slate-700 mb-1">Judul Event Ujian:</label>
            <input v-model="eventForm.title" type="text" required placeholder="Contoh: Asesmen Sumatif Akhir Tahun (ASAT) 2025/2026" class="w-full px-3 py-2 rounded-xl border border-slate-300 focus:outline-none focus:ring-2 focus:ring-indigo-500" />
          </div>

          <div class="grid grid-cols-3 gap-3">
            <div>
              <label class="block font-bold text-slate-700 mb-1">Kode Unik Event:</label>
              <input v-model="eventForm.code" type="text" required placeholder="ASAT-2026" class="w-full uppercase font-mono px-3 py-2 rounded-xl border border-slate-300 focus:outline-none focus:ring-2 focus:ring-indigo-500" />
            </div>
            <div>
              <label class="block font-bold text-slate-700 mb-1">Tahun Ajaran:</label>
              <input v-model="eventForm.academic_year" type="text" required placeholder="2026/2027" class="w-full font-mono px-3 py-2 rounded-xl border border-slate-300 focus:outline-none focus:ring-2 focus:ring-indigo-500" />
            </div>
            <div>
              <label class="block font-bold text-slate-700 mb-1">Semester:</label>
              <select v-model="eventForm.semester" required class="w-full px-3 py-2 rounded-xl border border-slate-300 focus:outline-none focus:ring-2 focus:ring-indigo-500">
                <option value="GANJIL">Ganjil</option>
                <option value="GENAP">Genap</option>
              </select>
            </div>
          </div>

          <div class="grid grid-cols-2 gap-3">
            <div>
              <label class="block font-bold text-slate-700 mb-1">Tanggal Mulai:</label>
              <input v-model="eventForm.start_date" type="date" required class="w-full px-3 py-2 rounded-xl border border-slate-300" />
            </div>
            <div>
              <label class="block font-bold text-slate-700 mb-1">Tanggal Selesai:</label>
              <input v-model="eventForm.end_date" type="date" required class="w-full px-3 py-2 rounded-xl border border-slate-300" />
            </div>
          </div>

          <div>
            <label class="block font-bold text-slate-700 mb-1">Keterangan / Catatan:</label>
            <textarea v-model="eventForm.description" rows="2" placeholder="Catatan kegiatan kurikulum atau petunjuk pekan ujian..." class="w-full px-3 py-2 rounded-xl border border-slate-300"></textarea>
          </div>

          <div class="flex items-center space-x-2 pt-1">
            <input id="event-active-check" v-model="eventForm.is_active" type="checkbox" class="w-4 h-4 text-indigo-600 rounded-md border-slate-300" />
            <label for="event-active-check" class="font-bold text-slate-700 cursor-pointer">
              Setel sebagai Event Aktif (Pekan Ujian Sedang Berlangsung)
            </label>
          </div>

          <div class="pt-3 flex justify-end space-x-2">
            <button type="button" @click="showEventModal = false" class="px-4 py-2 bg-slate-100 rounded-xl text-slate-700 font-semibold">Batal</button>
            <button type="submit" class="px-4 py-2 bg-indigo-600 hover:bg-indigo-700 text-white rounded-xl font-bold shadow-xs">
              {{ isEditEvent ? 'Simpan Perubahan' : 'Buat Event' }}
            </button>
          </div>
        </form>
      </div>
    </div>

    <!-- MODAL: TAMBAH SISWA -->
    <div v-if="showStudentModal" class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-slate-900/60 backdrop-blur-xs">
      <div class="bg-white rounded-3xl max-w-md w-full p-6 shadow-2xl space-y-4">
        <div>
          <h3 class="text-base font-bold text-slate-900">Tambah Siswa Baru</h3>
          <p class="text-xs text-slate-500">Tambahkan akun peserta ujian langsung ke database.</p>
        </div>
        <form @submit.prevent="createStudent" class="space-y-3 text-xs">
          <div>
            <label class="block font-bold text-slate-700 mb-1">Nama Lengkap Siswa:</label>
            <input v-model="newStudent.full_name" type="text" required placeholder="Contoh: Muhammad Rizky" class="w-full px-3 py-2 rounded-xl border border-slate-300 focus:ring-2 focus:ring-indigo-500 focus:outline-none" />
          </div>
          <div class="grid grid-cols-2 gap-3">
            <div>
              <label class="block font-bold text-slate-700 mb-1">Username Login:</label>
              <input v-model="newStudent.username" type="text" required placeholder="siswa4" class="w-full px-3 py-2 rounded-xl border border-slate-300 focus:ring-2 focus:ring-indigo-500 focus:outline-none font-mono" />
            </div>
            <div>
              <label class="block font-bold text-slate-700 mb-1">Password:</label>
              <input v-model="newStudent.password" type="text" placeholder="Default: siswa123" class="w-full px-3 py-2 rounded-xl border border-slate-300 focus:ring-2 focus:ring-indigo-500 focus:outline-none" />
            </div>
          </div>
          <div class="grid grid-cols-2 gap-3">
            <div>
              <label class="block font-bold text-slate-700 mb-1">NIS (Nomor Induk):</label>
              <input v-model="newStudent.nis" type="text" required placeholder="1004" class="w-full px-3 py-2 rounded-xl border border-slate-300 focus:ring-2 focus:ring-indigo-500 focus:outline-none font-mono" />
            </div>
            <div>
              <label class="block font-bold text-slate-700 mb-1">NISN:</label>
              <input v-model="newStudent.nisn" type="text" placeholder="0051234504" class="w-full px-3 py-2 rounded-xl border border-slate-300 focus:ring-2 focus:ring-indigo-500 focus:outline-none font-mono" />
            </div>
          </div>
          <div class="grid grid-cols-2 gap-3">
            <div>
              <label class="block font-bold text-slate-700 mb-1">Rombel Kelas:</label>
              <select v-model="newStudent.class_id" required class="w-full px-3 py-2 rounded-xl border border-slate-300 focus:ring-2 focus:ring-indigo-500 focus:outline-none">
                <option v-for="c in classes" :key="c.id" :value="c.id">{{ c.name }}</option>
              </select>
            </div>
            <div>
              <label class="block font-bold text-slate-700 mb-1">Jenis Kelamin:</label>
              <select v-model="newStudent.gender" class="w-full px-3 py-2 rounded-xl border border-slate-300 focus:ring-2 focus:ring-indigo-500 focus:outline-none">
                <option value="L">Laki-Laki (L)</option>
                <option value="P">Perempuan (P)</option>
              </select>
            </div>
          </div>
          <div class="pt-3 flex justify-end space-x-2">
            <button type="button" @click="showStudentModal = false" class="px-4 py-2 bg-slate-100 hover:bg-slate-200 rounded-xl text-slate-700 font-semibold cursor-pointer">Batal</button>
            <button type="submit" class="px-4 py-2 bg-indigo-600 hover:bg-indigo-700 text-white rounded-xl font-bold shadow-xs cursor-pointer">+ Simpan Siswa</button>
          </div>
        </form>
      </div>
    </div>

    <!-- MODAL: EDIT DATA SISWA -->
    <div v-if="showEditStudentModal" class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-slate-900/60 backdrop-blur-xs">
      <div class="bg-white rounded-3xl max-w-md w-full p-6 shadow-2xl space-y-4">
        <div>
          <h3 class="text-base font-bold text-slate-900">Edit Data Siswa</h3>
          <p class="text-xs text-slate-500">Perbarui identitas, rombel kelas, atau ganti password akun siswa.</p>
        </div>
        <form @submit.prevent="updateStudent" class="space-y-3 text-xs">
          <div>
            <label class="block font-bold text-slate-700 mb-1">Nama Lengkap Siswa:</label>
            <input v-model="editStudentForm.full_name" type="text" required class="w-full px-3 py-2 rounded-xl border border-slate-300 focus:ring-2 focus:ring-indigo-500 focus:outline-none" />
          </div>
          <div class="grid grid-cols-2 gap-3">
            <div>
              <label class="block font-bold text-slate-700 mb-1">Username Login:</label>
              <input v-model="editStudentForm.username" type="text" required class="w-full px-3 py-2 rounded-xl border border-slate-300 focus:ring-2 focus:ring-indigo-500 focus:outline-none font-mono" />
            </div>
            <div>
              <label class="block font-bold text-slate-700 mb-1">Password Baru (Opsional):</label>
              <input v-model="editStudentForm.password" type="text" placeholder="Kosongkan jika tak diubah" class="w-full px-3 py-2 rounded-xl border border-slate-300 focus:ring-2 focus:ring-indigo-500 focus:outline-none" />
            </div>
          </div>
          <div class="grid grid-cols-2 gap-3">
            <div>
              <label class="block font-bold text-slate-700 mb-1">NIS:</label>
              <input v-model="editStudentForm.nis" type="text" required class="w-full px-3 py-2 rounded-xl border border-slate-300 focus:ring-2 focus:ring-indigo-500 focus:outline-none font-mono" />
            </div>
            <div>
              <label class="block font-bold text-slate-700 mb-1">NISN:</label>
              <input v-model="editStudentForm.nisn" type="text" placeholder="0051234504" class="w-full px-3 py-2 rounded-xl border border-slate-300 focus:ring-2 focus:ring-indigo-500 focus:outline-none font-mono" />
            </div>
          </div>
          <div class="grid grid-cols-2 gap-3">
            <div>
              <label class="block font-bold text-slate-700 mb-1">Rombel Kelas:</label>
              <select v-model="editStudentForm.class_id" required class="w-full px-3 py-2 rounded-xl border border-slate-300 focus:ring-2 focus:ring-indigo-500 focus:outline-none">
                <option v-for="c in classes" :key="c.id" :value="c.id">{{ c.name }}</option>
              </select>
            </div>
            <div>
              <label class="block font-bold text-slate-700 mb-1">Jenis Kelamin:</label>
              <select v-model="editStudentForm.gender" class="w-full px-3 py-2 rounded-xl border border-slate-300 focus:ring-2 focus:ring-indigo-500 focus:outline-none">
                <option value="L">Laki-Laki (L)</option>
                <option value="P">Perempuan (P)</option>
              </select>
            </div>
          </div>
          <div class="pt-3 flex justify-end space-x-2">
            <button type="button" @click="showEditStudentModal = false" class="px-4 py-2 bg-slate-100 hover:bg-slate-200 rounded-xl text-slate-700 font-semibold cursor-pointer">Batal</button>
            <button type="submit" class="px-4 py-2 bg-indigo-600 hover:bg-indigo-700 text-white rounded-xl font-bold shadow-xs cursor-pointer">Simpan Perubahan</button>
          </div>
        </form>
      </div>
    </div>

    <!-- MODAL: DETAIL SISWA -->
    <div v-if="showStudentDetailModal && selectedStudentDetail" class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-slate-900/60 backdrop-blur-xs">
      <div class="bg-white rounded-3xl max-w-lg w-full p-6 shadow-2xl space-y-4">
        <div class="flex items-center justify-between border-b border-slate-100 pb-3">
          <div class="flex items-center space-x-3">
            <div class="w-10 h-10 rounded-2xl bg-indigo-50 text-indigo-700 flex items-center justify-center font-bold text-base">
              👤
            </div>
            <div>
              <h3 class="text-base font-bold text-slate-900">{{ selectedStudentDetail.user?.full_name }}</h3>
              <p class="text-xs text-slate-500">Rincian Informasi Master Siswa</p>
            </div>
          </div>
          <button @click="showStudentDetailModal = false" class="p-1 text-slate-400 hover:text-slate-600 rounded-xl cursor-pointer">✕</button>
        </div>

        <div class="grid grid-cols-2 gap-3 text-xs">
          <div class="bg-slate-50 p-3 rounded-2xl border border-slate-200/80">
            <span class="text-[10px] font-semibold text-slate-400 uppercase tracking-wider block">Username Akun</span>
            <span class="font-mono font-bold text-slate-800 text-sm mt-0.5 block">{{ selectedStudentDetail.user?.username }}</span>
          </div>
          <div class="bg-slate-50 p-3 rounded-2xl border border-slate-200/80">
            <span class="text-[10px] font-semibold text-slate-400 uppercase tracking-wider block">Rombel Kelas</span>
            <span class="font-bold text-indigo-700 text-sm mt-0.5 block">{{ selectedStudentDetail.class_room?.name }}</span>
          </div>
          <div class="bg-slate-50 p-3 rounded-2xl border border-slate-200/80">
            <span class="text-[10px] font-semibold text-slate-400 uppercase tracking-wider block">Nomor Induk (NIS)</span>
            <span class="font-mono font-bold text-slate-800 mt-0.5 block">{{ selectedStudentDetail.nis || '-' }}</span>
          </div>
          <div class="bg-slate-50 p-3 rounded-2xl border border-slate-200/80">
            <span class="text-[10px] font-semibold text-slate-400 uppercase tracking-wider block">NISN</span>
            <span class="font-mono font-bold text-slate-800 mt-0.5 block">{{ selectedStudentDetail.nisn || '-' }}</span>
          </div>
          <div class="bg-slate-50 p-3 rounded-2xl border border-slate-200/80">
            <span class="text-[10px] font-semibold text-slate-400 uppercase tracking-wider block">Jenis Kelamin</span>
            <span class="font-semibold text-slate-800 mt-0.5 block">
              {{ selectedStudentDetail.gender === 'P' ? 'Perempuan (P)' : 'Laki-Laki (L)' }}
            </span>
          </div>
          <div class="bg-slate-50 p-3 rounded-2xl border border-slate-200/80">
            <span class="text-[10px] font-semibold text-slate-400 uppercase tracking-wider block">Status Sesi HP</span>
            <span class="mt-0.5 block">
              <span v-if="selectedStudentDetail.user?.session_token" class="px-2 py-0.5 rounded-full text-[10px] font-bold bg-amber-100 text-amber-800 inline-block">
                🔒 Terkunci di Perangkat
              </span>
              <span v-else class="px-2 py-0.5 rounded-full text-[10px] font-bold bg-emerald-100 text-emerald-800 inline-block">
                Bebas (Siap Login)
              </span>
            </span>
          </div>
        </div>

        <div class="pt-2 flex items-center justify-between border-t border-slate-100">
          <div class="flex items-center gap-2">
            <button
              v-if="selectedStudentDetail.user?.session_token"
              @click="resetStudentSession(selectedStudentDetail)"
              class="px-3 py-1.5 bg-amber-50 hover:bg-amber-100 text-amber-800 border border-amber-300 font-bold text-xs rounded-xl transition cursor-pointer"
            >
              🔓 Reset Sesi HP
            </button>
            <button
              @click="showStudentDetailModal = false; openEditStudent(selectedStudentDetail)"
              class="px-3 py-1.5 bg-indigo-50 hover:bg-indigo-100 text-indigo-700 font-bold text-xs rounded-xl transition cursor-pointer"
            >
              ✏️ Edit Data
            </button>
          </div>
          <button
            type="button"
            @click="showStudentDetailModal = false"
            class="px-4 py-2 bg-slate-100 hover:bg-slate-200 rounded-xl text-slate-700 font-semibold text-xs cursor-pointer"
          >
            Tutup
          </button>
        </div>
      </div>
    </div>

    <!-- MODAL: TAMBAH GURU / STAF -->
    <div v-if="showTeacherModal" class="fixed inset-0 z-50 flex items-center justify-center p-4">
      <!-- Backdrop: fade opacity murni (tanpa scale) -->
      <transition
        appear
        enter-active-class="transition-opacity duration-200 ease-out"
        enter-from-class="opacity-0"
        enter-to-class="opacity-100"
        leave-active-class="transition-opacity duration-150 ease-in"
        leave-from-class="opacity-100"
        leave-to-class="opacity-0"
      >
        <div class="fixed inset-0 bg-slate-900/60 backdrop-blur-xs"></div>
      </transition>

      <!-- Kartu modal -->
      <transition
        appear
        enter-active-class="transition duration-200 ease-out transform"
        enter-from-class="opacity-0 scale-95 translate-y-2"
        enter-to-class="opacity-100 scale-100 translate-y-0"
        leave-active-class="transition duration-150 ease-in transform"
        leave-from-class="opacity-100 scale-100 translate-y-0"
        leave-to-class="opacity-0 scale-95 translate-y-2"
      >
        <div class="relative z-10 bg-white rounded-3xl max-w-lg w-full p-6 shadow-2xl flex flex-col gap-4 max-h-[90vh]">
          <div class="shrink-0">
            <h3 class="text-base font-bold text-slate-900">Tambah Akun Guru / Staf</h3>
            <p class="text-xs text-slate-500">Pendaftaran akun pendidik atau pengelola sistem.</p>
          </div>
          <form @submit.prevent="createTeacher" class="flex flex-col gap-3 min-h-0 text-xs">
            <div class="space-y-3 overflow-y-auto min-h-0 px-0.5">
            <div>
              <label class="block font-bold text-slate-700 mb-1">Nama Lengkap & Gelar:</label>
              <input v-model="newTeacher.full_name" type="text" required placeholder="Dra. Sri Wahyuni, M.Pd" class="w-full px-3 py-2 rounded-xl border border-slate-300 focus:ring-2 focus:ring-indigo-500 focus:outline-none" />
            </div>
            <div class="grid grid-cols-2 gap-3">
              <div>
                <label class="block font-bold text-slate-700 mb-1">Username Login:</label>
                <input v-model="newTeacher.username" type="text" required placeholder="guru2" class="w-full px-3 py-2 rounded-xl border border-slate-300 focus:ring-2 focus:ring-indigo-500 focus:outline-none font-mono" />
              </div>
              <div>
                <label class="block font-bold text-slate-700 mb-1">Password:</label>
                <input v-model="newTeacher.password" type="text" placeholder="Default: guru123" class="w-full px-3 py-2 rounded-xl border border-slate-300 focus:ring-2 focus:ring-indigo-500 focus:outline-none" />
              </div>
            </div>
            <p v-if="!isGrantor" class="text-[11px] text-slate-500">Hanya administrator yang dapat mengatur izin akun. Akun baru memakai izin template Guru.</p>
            <StaffAccessEditor
              v-if="permissionCatalog"
              v-model="newTeacherAccess"
              v-model:valid="newTeacherAccessValid"
              :catalog="permissionCatalog"
              :readonly="!isGrantor"
              :confirm-fn="showConfirmModal"
            />
            </div>
            <div class="pt-3 flex justify-end space-x-2 shrink-0">
              <button type="button" @click="showTeacherModal = false" class="px-4 py-2 bg-slate-100 hover:bg-slate-200 rounded-xl text-slate-700 font-semibold cursor-pointer active:scale-95 transition">Batal</button>
              <button type="submit" :disabled="!newTeacherCanSave" class="px-4 py-2 bg-indigo-600 hover:bg-indigo-700 text-white rounded-xl font-bold shadow-xs cursor-pointer active:scale-95 transition disabled:opacity-50 disabled:cursor-not-allowed disabled:active:scale-100">+ Simpan</button>
            </div>
          </form>
        </div>
      </transition>
    </div>

    <!-- MODAL: EDIT GURU / STAF -->
    <div v-if="showEditTeacherModal" class="fixed inset-0 z-50 flex items-center justify-center p-4">
      <!-- Backdrop: fade opacity murni (tanpa scale) -->
      <transition
        appear
        enter-active-class="transition-opacity duration-200 ease-out"
        enter-from-class="opacity-0"
        enter-to-class="opacity-100"
        leave-active-class="transition-opacity duration-150 ease-in"
        leave-from-class="opacity-100"
        leave-to-class="opacity-0"
      >
        <div class="fixed inset-0 bg-slate-900/60 backdrop-blur-xs"></div>
      </transition>

      <!-- Kartu modal -->
      <transition
        appear
        enter-active-class="transition duration-200 ease-out transform"
        enter-from-class="opacity-0 scale-95 translate-y-2"
        enter-to-class="opacity-100 scale-100 translate-y-0"
        leave-active-class="transition duration-150 ease-in transform"
        leave-from-class="opacity-100 scale-100 translate-y-0"
        leave-to-class="opacity-0 scale-95 translate-y-2"
      >
        <div class="relative z-10 bg-white rounded-3xl max-w-lg w-full p-6 shadow-2xl flex flex-col gap-4 max-h-[90vh]">
          <div class="shrink-0">
            <h3 class="text-base font-bold text-slate-900">Edit Data Guru / Staf</h3>
            <p class="text-xs text-slate-500">Perbarui nama, username, akses, atau ganti password.</p>
          </div>
          <form @submit.prevent="updateTeacher" class="flex flex-col gap-3 min-h-0 text-xs">
            <div class="space-y-3 overflow-y-auto min-h-0 px-0.5">
            <div>
              <label class="block font-bold text-slate-700 mb-1">Nama Lengkap & Gelar:</label>
              <input v-model="editTeacherForm.full_name" type="text" required class="w-full px-3 py-2 rounded-xl border border-slate-300 focus:ring-2 focus:ring-indigo-500 focus:outline-none" />
            </div>
            <div class="grid grid-cols-2 gap-3">
              <div>
                <label class="block font-bold text-slate-700 mb-1">Username Login:</label>
                <input v-model="editTeacherForm.username" type="text" required :disabled="!isGrantor" class="w-full px-3 py-2 rounded-xl border border-slate-300 focus:ring-2 focus:ring-indigo-500 focus:outline-none font-mono disabled:bg-slate-100 disabled:text-slate-500 disabled:cursor-not-allowed" />
              </div>
              <div>
                <label class="block font-bold text-slate-700 mb-1">Password Baru (Opsional):</label>
                <input v-model="editTeacherForm.password" type="text" placeholder="Kosongkan jika tak diubah" :disabled="!isGrantor" class="w-full px-3 py-2 rounded-xl border border-slate-300 focus:ring-2 focus:ring-indigo-500 focus:outline-none disabled:bg-slate-100 disabled:text-slate-500 disabled:cursor-not-allowed" />
              </div>
            </div>
            <p v-if="!isGrantor" class="text-[11px] text-slate-500">Hanya administrator yang dapat mengubah username, password, dan izin akun. Anda hanya dapat mengubah nama.</p>
            <StaffAccessEditor
              v-if="permissionCatalog"
              v-model="editTeacherAccess"
              v-model:valid="editTeacherAccessValid"
              :catalog="permissionCatalog"
              :readonly="!isGrantor"
              :confirm-fn="showConfirmModal"
            />
            </div>
            <div class="pt-3 flex justify-end space-x-2 shrink-0">
              <button type="button" @click="showEditTeacherModal = false" class="px-4 py-2 bg-slate-100 hover:bg-slate-200 rounded-xl text-slate-700 font-semibold cursor-pointer active:scale-95 transition">Batal</button>
              <button type="submit" :disabled="!editTeacherCanSave" class="px-4 py-2 bg-indigo-600 hover:bg-indigo-700 text-white rounded-xl font-bold shadow-xs cursor-pointer active:scale-95 transition disabled:opacity-50 disabled:cursor-not-allowed disabled:active:scale-100">Simpan Perubahan</button>
            </div>
          </form>
        </div>
      </transition>
    </div>

    <!-- MODAL: TAMBAH KELAS -->
    <div v-if="showClassModal" class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-slate-900/60 backdrop-blur-xs">
      <div class="bg-white rounded-3xl max-w-md w-full p-6 shadow-2xl space-y-4">
        <div>
          <h3 class="text-base font-bold text-slate-900">Tambah Kelas Baru</h3>
          <p class="text-xs text-slate-500">Pendaftaran rombongan belajar baru.</p>
        </div>
        <form @submit.prevent="createClass" class="space-y-3 text-xs">
          <div>
            <label class="block font-bold text-slate-700 mb-1">Nama Kelas:</label>
            <input v-model="newClass.name" type="text" required placeholder="Contoh: XII MIPA 2" class="w-full px-3 py-2 rounded-xl border border-slate-300 focus:ring-2 focus:ring-indigo-500 focus:outline-none" />
          </div>
          <div class="grid grid-cols-2 gap-3">
            <div>
              <label class="block font-bold text-slate-700 mb-1">Tingkat (Grade):</label>
              <select v-model="newClass.grade" class="w-full px-3 py-2 rounded-xl border border-slate-300 focus:ring-2 focus:ring-indigo-500 focus:outline-none">
                <option value="X">Kelas X</option>
                <option value="XI">Kelas XI</option>
                <option value="XII">Kelas XII</option>
              </select>
            </div>
            <div>
              <label class="block font-bold text-slate-700 mb-1">Jurusan:</label>
              <input v-model="newClass.major" type="text" placeholder="MIPA / IPS / Lainnya" class="w-full px-3 py-2 rounded-xl border border-slate-300 focus:ring-2 focus:ring-indigo-500 focus:outline-none" />
            </div>
          </div>
          <div class="pt-3 flex justify-end space-x-2">
            <button type="button" @click="showClassModal = false" class="px-4 py-2 bg-slate-100 hover:bg-slate-200 rounded-xl text-slate-700 font-semibold cursor-pointer">Batal</button>
            <button type="submit" class="px-4 py-2 bg-indigo-600 hover:bg-indigo-700 text-white rounded-xl font-bold shadow-xs cursor-pointer">+ Simpan Kelas</button>
          </div>
        </form>
      </div>
    </div>

    <!-- MODAL: EDIT KELAS -->
    <div v-if="showEditClassModal" class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-slate-900/60 backdrop-blur-xs">
      <div class="bg-white rounded-3xl max-w-md w-full p-6 shadow-2xl space-y-4">
        <div>
          <h3 class="text-base font-bold text-slate-900">Edit Data Kelas</h3>
          <p class="text-xs text-slate-500">Perbarui nama, tingkatan, atau jurusan rombel.</p>
        </div>
        <form @submit.prevent="updateClass" class="space-y-3 text-xs">
          <div>
            <label class="block font-bold text-slate-700 mb-1">Nama Kelas:</label>
            <input v-model="editClassForm.name" type="text" required class="w-full px-3 py-2 rounded-xl border border-slate-300 focus:ring-2 focus:ring-indigo-500 focus:outline-none" />
          </div>
          <div class="grid grid-cols-2 gap-3">
            <div>
              <label class="block font-bold text-slate-700 mb-1">Tingkat (Grade):</label>
              <select v-model="editClassForm.grade" class="w-full px-3 py-2 rounded-xl border border-slate-300 focus:ring-2 focus:ring-indigo-500 focus:outline-none">
                <option value="X">Kelas X</option>
                <option value="XI">Kelas XI</option>
                <option value="XII">Kelas XII</option>
              </select>
            </div>
            <div>
              <label class="block font-bold text-slate-700 mb-1">Jurusan:</label>
              <input v-model="editClassForm.major" type="text" class="w-full px-3 py-2 rounded-xl border border-slate-300 focus:ring-2 focus:ring-indigo-500 focus:outline-none" />
            </div>
          </div>
          <div class="pt-3 flex justify-end space-x-2">
            <button type="button" @click="showEditClassModal = false" class="px-4 py-2 bg-slate-100 hover:bg-slate-200 rounded-xl text-slate-700 font-semibold cursor-pointer">Batal</button>
            <button type="submit" class="px-4 py-2 bg-indigo-600 hover:bg-indigo-700 text-white rounded-xl font-bold shadow-xs cursor-pointer">Simpan Perubahan</button>
          </div>
        </form>
      </div>
    </div>

    <!-- MODAL: TAMBAH / EDIT MATA PELAJARAN -->
    <div v-if="showSubjectModal" class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-slate-900/60 backdrop-blur-xs">
      <div class="bg-white rounded-3xl max-w-md w-full p-6 shadow-2xl space-y-4">
        <h3 class="text-base font-bold text-slate-900">
          {{ isEditSubject ? 'Edit Mata Pelajaran' : 'Tambah Mata Pelajaran Baru' }}
        </h3>
        <form @submit.prevent="submitSubjectForm" class="space-y-3 text-xs">
          <div>
            <label class="block font-bold text-slate-700 mb-1">Kode Mapel:</label>
            <input v-model="subjectForm.code" type="text" required placeholder="Contoh: MAT-WJB-XII" class="w-full px-3 py-2 rounded-xl border border-slate-300 font-mono uppercase" />
            <p class="text-[10px] text-slate-400 mt-0.5">Gunakan kode unik tanpa spasi.</p>
          </div>
          <div>
            <label class="block font-bold text-slate-700 mb-1">Nama Mata Pelajaran:</label>
            <input v-model="subjectForm.name" type="text" required placeholder="Contoh: Matematika Wajib" class="w-full px-3 py-2 rounded-xl border border-slate-300" />
          </div>
          <div class="pt-3 flex justify-end space-x-2">
            <button type="button" @click="showSubjectModal = false" class="px-4 py-2 bg-slate-100 rounded-xl text-slate-700 font-semibold">Batal</button>
            <button type="submit" class="px-4 py-2 bg-indigo-600 hover:bg-indigo-700 text-white rounded-xl font-bold shadow-xs">
              {{ isEditSubject ? 'Simpan Perubahan' : 'Tambah Mapel' }}
            </button>
          </div>
        </form>
      </div>
    </div>

    <!-- MODAL: ALOKASIKAN / EDIT KELAS MAPEL -->
    <div v-if="showClassSubjectModal" class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-slate-900/60 backdrop-blur-xs">
      <div class="bg-white rounded-3xl max-w-md w-full p-6 shadow-2xl space-y-4">
        <h3 class="text-base font-bold text-slate-900">
          {{ isEditClassSubject ? 'Edit Alokasi Kelas Mapel' : 'Alokasikan Mapel ke Kelas' }}
        </h3>
        <form @submit.prevent="submitClassSubjectForm" class="space-y-3 text-xs">
          <div>
            <label class="block font-bold text-slate-700 mb-1">Pilih Rombel / Kelas:</label>
            <select v-model="classSubjectForm.class_id" required class="w-full px-3 py-2 rounded-xl border border-slate-300">
              <option v-for="c in classes" :key="c.id" :value="c.id">{{ c.name }} ({{ c.grade }} {{ c.major }})</option>
            </select>
          </div>
          <div>
            <label class="block font-bold text-slate-700 mb-1">Pilih Mata Pelajaran:</label>
            <select v-model="classSubjectForm.subject_id" required class="w-full px-3 py-2 rounded-xl border border-slate-300">
              <option v-for="s in subjects" :key="s.id" :value="s.id">{{ s.name }} ({{ s.code }})</option>
            </select>
          </div>
          <div>
            <label class="block font-bold text-slate-700 mb-1">Guru Pengampu:</label>
            <select v-model="classSubjectForm.teacher_id" required class="w-full px-3 py-2 rounded-xl border border-slate-300">
              <option v-for="t in teachers" :key="t.id" :value="t.id">{{ t.full_name }} ({{ t.username }})</option>
            </select>
          </div>
          <div>
            <label class="block font-bold text-slate-700 mb-1">Tahun Ajaran:</label>
            <input v-model="classSubjectForm.academic_year" type="text" placeholder="2026/2027" class="w-full px-3 py-2 rounded-xl border border-slate-300 font-mono" />
          </div>
          <div class="pt-3 flex justify-end space-x-2">
            <button type="button" @click="showClassSubjectModal = false" class="px-4 py-2 bg-slate-100 rounded-xl text-slate-700 font-semibold cursor-pointer">Batal</button>
            <button type="submit" class="px-4 py-2 bg-indigo-600 hover:bg-indigo-700 text-white rounded-xl font-bold shadow-xs cursor-pointer">
              {{ isEditClassSubject ? 'Simpan Perubahan' : 'Simpan Alokasi' }}
            </button>
          </div>
        </form>
      </div>
    </div>

    <!-- MODAL: IMPOR SISWA EXCEL -->
    <div v-if="showImportModal" class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-slate-900/60 backdrop-blur-xs">
      <div class="bg-white rounded-3xl max-w-md w-full p-6 shadow-2xl space-y-4">
        <div class="flex items-center justify-between">
          <h3 class="text-base font-bold text-slate-900">Impor Data Siswa dari Excel</h3>
          <button @click="showImportModal = false" class="text-slate-400 hover:text-slate-600 font-bold text-lg">&times;</button>
        </div>
        <p class="text-xs text-slate-600">
          Gunakan template Excel resmi CBT untuk mengunggah ratusan akun siswa secara instan. Nama kelas yang cocok akan otomatis dialokasikan.
        </p>

        <div class="p-3 bg-indigo-50/50 rounded-2xl border border-indigo-100 flex items-center justify-between">
          <div class="flex items-center space-x-2">
            <svg class="w-5 h-5 text-indigo-600" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 10v6m0 0l-3-3m3 3l3-3m2 8H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
            </svg>
            <span class="text-xs font-semibold text-slate-700">Template_Import_Siswa_CBT.xlsx</span>
          </div>
          <button @click="downloadStudentTemplate" class="text-xs font-bold text-indigo-600 hover:underline">
            Unduh Format
          </button>
        </div>

        <form @submit.prevent="submitImportExcel" class="space-y-4 text-xs">
          <div>
            <label class="block font-bold text-slate-700 mb-1">Pilih Berkas Excel (.xlsx):</label>
            <input
              type="file"
              accept=".xlsx"
              @change="handleExcelFileSelect"
              required
              class="w-full text-xs text-slate-500 file:mr-3 file:py-2 file:px-4 file:rounded-xl file:border-0 file:text-xs file:font-semibold file:bg-indigo-50 file:text-indigo-700 hover:file:bg-indigo-100 cursor-pointer border border-slate-200 rounded-xl p-2"
            />
          </div>

          <div class="pt-2 flex justify-end space-x-2">
            <button type="button" @click="showImportModal = false" class="px-4 py-2 bg-slate-100 rounded-xl text-slate-700 font-semibold" :disabled="isImporting">Batal</button>
            <button type="submit" class="px-4 py-2 bg-emerald-600 hover:bg-emerald-700 text-white rounded-xl font-bold shadow-xs flex items-center space-x-1" :disabled="isImporting">
              <span v-if="isImporting">Sedang Mengunggah...</span>
              <span v-else>Unggah & Impor</span>
            </button>
          </div>
        </form>
      </div>
    </div>

    <!-- Import Kelas Modal -->
    <div v-if="showImportClassModal" class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-slate-900/60 backdrop-blur-xs">
      <div class="bg-white rounded-3xl shadow-2xl w-full max-w-md p-6 space-y-4">
        <div class="flex items-center justify-between">
          <h3 class="text-base font-bold text-slate-900">Impor Data Kelas dari Excel</h3>
          <button @click="showImportClassModal = false" class="text-slate-400 hover:text-slate-600 font-bold text-lg">&times;</button>
        </div>
        <p class="text-xs text-slate-500">
          Gunakan template Excel resmi CBT untuk mengimpor data rombel/kelas secara massal. Kelas yang sudah ada (nama sama) akan dilewati otomatis.
        </p>
        <div class="flex items-center space-x-2 bg-slate-50 border border-slate-200 rounded-2xl p-3">
          <svg class="w-5 h-5 text-emerald-600 shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
          </svg>
          <span class="text-xs font-semibold text-slate-700">Template_Import_Kelas_CBT.xlsx</span>
        </div>
        <button @click="downloadClassTemplate" class="text-xs font-bold text-indigo-600 hover:underline flex items-center space-x-1">
          <svg class="w-3.5 h-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-4l-4 4m0 0l-4-4m4 4V4" /></svg>
          <span>Unduh Template Excel</span>
        </button>
        <form @submit.prevent="submitImportClasses" class="space-y-4 text-xs">
          <div>
            <label class="block font-bold text-slate-700 mb-1">Pilih Berkas Excel (.xlsx):</label>
            <input
              type="file"
              accept=".xlsx,.xls"
              @change="(e) => importClassFile = e.target.files[0]"
              class="block w-full text-xs text-slate-500 file:mr-3 file:py-1.5 file:px-3 file:rounded-lg file:border-0 file:text-xs file:font-semibold file:bg-indigo-50 file:text-indigo-700 hover:file:bg-indigo-100"
              required
            />
          </div>
          <div v-if="importClassResult" :class="importClassResult.success ? 'bg-emerald-50 text-emerald-700 border border-emerald-200' : 'bg-red-50 text-red-700 border border-red-200'" class="p-3 rounded-xl text-xs font-semibold">
            {{ importClassResult.message }}
          </div>
          <div class="flex justify-end space-x-2">
            <button type="button" @click="showImportClassModal = false; importClassResult = null" class="px-4 py-2 bg-slate-100 rounded-xl text-slate-700 font-semibold" :disabled="isImportingClass">Batal</button>
            <button type="submit" class="px-4 py-2 bg-emerald-600 hover:bg-emerald-700 text-white rounded-xl font-bold shadow-xs flex items-center space-x-1" :disabled="isImportingClass">
              <span v-if="isImportingClass">Sedang Mengunggah...</span>
              <span v-else>Impor Kelas</span>
            </button>
          </div>
        </form>
      </div>
    </div>

    <!-- Import Mata Pelajaran Modal -->
    <div v-if="showImportSubjectModal" class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-slate-900/60 backdrop-blur-xs">
      <div class="bg-white rounded-3xl shadow-2xl w-full max-w-md p-6 space-y-4">
        <div class="flex items-center justify-between">
          <h3 class="text-base font-bold text-slate-900">Impor Mata Pelajaran dari Excel</h3>
          <button @click="showImportSubjectModal = false" class="text-slate-400 hover:text-slate-600 font-bold text-lg">&times;</button>
        </div>
        <p class="text-xs text-slate-500">
          Gunakan template Excel resmi CBT untuk mengimpor daftar mata pelajaran. Kode mapel yang sudah ada akan dilewati otomatis.
        </p>
        <div class="flex items-center space-x-2 bg-slate-50 border border-slate-200 rounded-2xl p-3">
          <svg class="w-5 h-5 text-cyan-600 shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
          </svg>
          <span class="text-xs font-semibold text-slate-700">Template_Import_Mapel_CBT.xlsx</span>
        </div>
        <button @click="downloadSubjectTemplate" class="text-xs font-bold text-indigo-600 hover:underline flex items-center space-x-1">
          <svg class="w-3.5 h-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-4l-4 4m0 0l-4-4m4 4V4" /></svg>
          <span>Unduh Template Excel</span>
        </button>
        <form @submit.prevent="submitImportSubjects" class="space-y-4 text-xs">
          <div>
            <label class="block font-bold text-slate-700 mb-1">Pilih Berkas Excel (.xlsx):</label>
            <input
              type="file"
              accept=".xlsx,.xls"
              @change="(e) => importSubjectFile = e.target.files[0]"
              class="block w-full text-xs text-slate-500 file:mr-3 file:py-1.5 file:px-3 file:rounded-lg file:border-0 file:text-xs file:font-semibold file:bg-indigo-50 file:text-indigo-700 hover:file:bg-indigo-100"
              required
            />
          </div>
          <div v-if="importSubjectResult" :class="importSubjectResult.success ? 'bg-emerald-50 text-emerald-700 border border-emerald-200' : 'bg-red-50 text-red-700 border border-red-200'" class="p-3 rounded-xl text-xs font-semibold">
            {{ importSubjectResult.message }}
          </div>
          <div class="flex justify-end space-x-2">
            <button type="button" @click="showImportSubjectModal = false; importSubjectResult = null" class="px-4 py-2 bg-slate-100 rounded-xl text-slate-700 font-semibold" :disabled="isImportingSubject">Batal</button>
            <button type="submit" class="px-4 py-2 bg-cyan-600 hover:bg-cyan-700 text-white rounded-xl font-bold shadow-xs flex items-center space-x-1" :disabled="isImportingSubject">
              <span v-if="isImportingSubject">Sedang Mengunggah...</span>
              <span v-else>Impor Mapel</span>
            </button>
          </div>
        </form>
      </div>
    </div>

    <!-- Import Guru Modal -->
    <div v-if="showImportTeacherModal" class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-slate-900/60 backdrop-blur-xs">
      <div class="bg-white rounded-3xl shadow-2xl w-full max-w-md p-6 space-y-4">
        <div class="flex items-center justify-between">
          <h3 class="text-base font-bold text-slate-900">Impor Data Guru & Staf dari Excel</h3>
          <button @click="showImportTeacherModal = false" class="text-slate-400 hover:text-slate-600 font-bold text-lg">&times;</button>
        </div>
        <p class="text-xs text-slate-500">
          Gunakan template Excel resmi CBT untuk mengimpor akun guru dan staf secara massal. Username yang sudah terdaftar akan dilewati otomatis.
        </p>
        <div class="p-3 bg-amber-50 border border-amber-200 rounded-2xl text-xs text-amber-800 font-medium">
          Password default: <span class="font-bold">guru123</span> (bisa diubah di kolom Password Awal di template)
        </div>
        <div class="flex items-center space-x-2 bg-slate-50 border border-slate-200 rounded-2xl p-3">
          <svg class="w-5 h-5 text-indigo-600 shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
          </svg>
          <span class="text-xs font-semibold text-slate-700">Template_Import_Guru_CBT.xlsx</span>
        </div>
        <button @click="downloadTeacherTemplate" class="text-xs font-bold text-indigo-600 hover:underline flex items-center space-x-1">
          <svg class="w-3.5 h-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-4l-4 4m0 0l-4-4m4 4V4" /></svg>
          <span>Unduh Template Excel</span>
        </button>
        <form @submit.prevent="submitImportTeachers" class="space-y-4 text-xs">
          <div>
            <label class="block font-bold text-slate-700 mb-1">Pilih Berkas Excel (.xlsx):</label>
            <input
              type="file"
              accept=".xlsx,.xls"
              @change="(e) => importTeacherFile = e.target.files[0]"
              class="block w-full text-xs text-slate-500 file:mr-3 file:py-1.5 file:px-3 file:rounded-lg file:border-0 file:text-xs file:font-semibold file:bg-indigo-50 file:text-indigo-700 hover:file:bg-indigo-100"
              required
            />
          </div>
          <div v-if="importTeacherResult" :class="importTeacherResult.success ? 'bg-emerald-50 text-emerald-700 border border-emerald-200' : 'bg-red-50 text-red-700 border border-red-200'" class="p-3 rounded-xl text-xs font-semibold">
            {{ importTeacherResult.message }}
          </div>
          <div class="flex justify-end space-x-2">
            <button type="button" @click="showImportTeacherModal = false; importTeacherResult = null" class="px-4 py-2 bg-slate-100 rounded-xl text-slate-700 font-semibold" :disabled="isImportingTeacher">Batal</button>
            <button type="submit" class="px-4 py-2 bg-indigo-600 hover:bg-indigo-700 text-white rounded-xl font-bold shadow-xs flex items-center space-x-1" :disabled="isImportingTeacher">
              <span v-if="isImportingTeacher">Sedang Mengunggah...</span>
              <span v-else>Impor Guru & Staf</span>
            </button>
          </div>
        </form>
      </div>
    </div>

    <!-- MODAL: CETAK DAFTAR HADIR & BERITA ACARA DARI DATA PENGAWASAN LIVE -->
    <ProctorPrintModal v-model="showProctorPrint" :proctor-data="proctorData" />

    <!-- MODAL: CETAK DOKUMEN RESMI UJIAN (DAFTAR HADIR & BERITA ACARA) -->
    <div v-if="showPrintModal && printSchedule" class="fixed inset-0 z-50 flex items-center justify-center p-2 sm:p-4 bg-slate-900/70 backdrop-blur-xs overflow-y-auto">
      <div class="bg-white rounded-3xl max-w-4xl w-full shadow-2xl overflow-hidden flex flex-col max-h-[95vh] my-auto">
        <!-- Modal Toolbar Header (no-print) -->
        <div class="flex items-center justify-between px-6 py-4 border-b border-slate-200 shrink-0 bg-slate-50 no-print">
          <div class="flex items-center space-x-3">
            <div class="w-9 h-9 rounded-2xl bg-emerald-100 text-emerald-700 flex items-center justify-center font-black text-base">
              📄
            </div>
            <div>
              <h3 class="text-sm font-bold text-slate-900">Format Cetak Dokumen Resmi Ujian</h3>
              <p class="text-[11px] text-slate-500">SMAS Islamic Centre Demak • Standar A4 Cetak Siap Pakai</p>
            </div>
          </div>

          <!-- Document Type Selector Pills -->
          <div class="flex items-center gap-1.5 bg-slate-200/70 p-1 rounded-2xl">
            <button
              type="button"
              @click="printDocType = 'attendance'"
              :class="[
                'px-3 py-1.5 rounded-xl text-xs font-bold transition cursor-pointer',
                printDocType === 'attendance' ? 'bg-white text-emerald-700 shadow-xs' : 'text-slate-600 hover:text-slate-900'
              ]"
            >
              📄 Daftar Hadir
            </button>
            <button
              type="button"
              @click="printDocType = 'report'"
              :class="[
                'px-3 py-1.5 rounded-xl text-xs font-bold transition cursor-pointer',
                printDocType === 'report' ? 'bg-white text-indigo-700 shadow-xs' : 'text-slate-600 hover:text-slate-900'
              ]"
            >
              📝 Berita Acara
            </button>
          </div>

          <button @click="showPrintModal = false" class="text-slate-400 hover:text-slate-600 p-1.5 rounded-xl hover:bg-slate-200 transition cursor-pointer">
            ✕
          </button>
        </div>

        <!-- Printable Document Area -->
        <div class="p-6 sm:p-8 overflow-y-auto flex-1 bg-slate-100/50">
          <div id="printable-document" class="bg-white p-8 max-w-[210mm] mx-auto shadow-sm border border-slate-200 text-black font-serif text-[12px] leading-relaxed">
            
            <!-- Kop Surat Resmi SMAS Islamic Centre Demak -->
            <div class="flex items-center gap-4 pb-3 border-b-4 border-double border-slate-900 mb-5">
              <img src="/logo-smic.png" alt="Logo SMAS Islamic Centre Demak" class="w-20 h-20 object-contain shrink-0" />
              <div class="text-center flex-1 space-y-0.5">
                <div class="font-bold text-sm tracking-wide uppercase text-slate-800">Yayasan Islamic Centre Demak</div>
                <div class="font-black text-lg tracking-wider uppercase text-slate-950">SMAS ISLAMIC CENTRE DEMAK</div>
                <div class="text-[11px] font-semibold tracking-wide text-slate-700">STATUS : TERAKREDITASI "A"</div>
                <div class="text-[10px] text-slate-600">Alamat: Jl. Diponegoro No. 12 Demak, Jawa Tengah 59515 | Telp: (0291) 685261 | Website: smicdemak.sch.id</div>
              </div>
            </div>

            <!-- DOKUMEN 1: DAFTAR HADIR PESERTA -->
            <div v-if="printDocType === 'attendance'" class="space-y-4">
              <div class="text-center space-y-1">
                <h2 class="text-base font-black uppercase tracking-wider underline">DAFTAR HADIR PESERTA UJIAN</h2>
                <p class="text-xs font-bold uppercase tracking-wide text-slate-800">{{ selectedExamEvent?.name || 'PENILAIAN AKHIR TAHUN / SUMATIF' }}</p>
                <p class="text-[11px] text-slate-600">TAHUN PELAJARAN 2025/2026</p>
              </div>

              <!-- Metadata Grid -->
              <table class="w-full text-xs border-collapse my-3">
                <tbody>
                  <tr>
                    <td class="py-1 w-32 font-bold">Mata Pelajaran</td>
                    <td class="py-1 w-4">:</td>
                    <td class="py-1 font-semibold">{{ printSchedule.subject?.name || printSchedule.bank?.subject?.name || '-' }}</td>
                    <td class="py-1 w-28 font-bold">Hari / Tanggal</td>
                    <td class="py-1 w-4">:</td>
                    <td class="py-1">{{ formatScheduleDateFull(printSchedule.start_time) }}</td>
                  </tr>
                  <tr>
                    <td class="py-1 font-bold">Kelas / Rombel</td>
                    <td class="py-1">:</td>
                    <td class="py-1 font-semibold">{{ printSchedule.class_room?.name || '-' }}</td>
                    <td class="py-1 font-bold">Waktu / Sesi</td>
                    <td class="py-1">:</td>
                    <td class="py-1">{{ formatScheduleTimeOnly(printSchedule.start_time) }} - {{ formatScheduleTimeOnly(printSchedule.end_time) }} WIB</td>
                  </tr>
                  <tr>
                    <td class="py-1 font-bold">Naskah Bank Soal</td>
                    <td class="py-1">:</td>
                    <td class="py-1">{{ printSchedule.bank?.title || 'Bank Soal Terintegrasi CBT' }}</td>
                    <td class="py-1 font-bold">Token Sesi</td>
                    <td class="py-1">:</td>
                    <td class="py-1 font-mono font-bold">{{ printSchedule.exam_token }}</td>
                  </tr>
                </tbody>
              </table>

              <!-- Tabel Daftar Hadir Siswa -->
              <table class="w-full border-collapse border border-slate-900 text-xs mt-4">
                <thead>
                  <tr class="bg-slate-100 border-b border-slate-900 text-center font-bold">
                    <th class="border border-slate-900 py-2 w-10">No.</th>
                    <th class="border border-slate-900 py-2 w-32">Nomor Peserta / NISN</th>
                    <th class="border border-slate-900 py-2 px-3 text-left">Nama Lengkap Siswa</th>
                    <th class="border border-slate-900 py-2 w-12">L/P</th>
                    <th class="border border-slate-900 py-2 w-44" colspan="2">Tanda Tangan</th>
                    <th class="border border-slate-900 py-2 w-24">Keterangan</th>
                  </tr>
                </thead>
                <tbody>
                  <tr v-if="printStudentsList.length === 0">
                    <td colspan="7" class="border border-slate-900 py-6 text-center text-slate-500 italic">
                      Belum ada data siswa teralokasi di kelas ini.
                    </td>
                  </tr>
                  <tr v-for="(stu, idx) in printStudentsList" :key="stu.id" class="border-b border-slate-900/60">
                    <td class="border border-slate-900 py-2.5 text-center font-semibold">{{ idx + 1 }}.</td>
                    <td class="border border-slate-900 py-2.5 px-2 text-center font-mono">{{ stu.nisn || stu.student_id_number || stu.user?.username || '-' }}</td>
                    <td class="border border-slate-900 py-2.5 px-3 font-semibold uppercase">{{ stu.user?.full_name || stu.name || '-' }}</td>
                    <td class="border border-slate-900 py-2.5 text-center">{{ stu.gender || stu.user?.gender || 'L' }}</td>
                    <!-- Tanda Tangan Selang-Seling Kolom Ganjil / Genap -->
                    <td class="border border-slate-900 py-2.5 px-2 w-24 text-left font-mono text-[10px] text-slate-500 align-top">
                      <span v-if="idx % 2 === 0">{{ idx + 1 }}. ......................</span>
                    </td>
                    <td class="border border-slate-900 py-2.5 px-2 w-24 text-left font-mono text-[10px] text-slate-500 align-top">
                      <span v-if="idx % 2 === 1">{{ idx + 1 }}. ......................</span>
                    </td>
                    <td class="border border-slate-900 py-2.5 text-center text-[10px] text-slate-400 font-sans">
                      [ &nbsp; ] Hadir
                    </td>
                  </tr>
                </tbody>
              </table>

              <!-- Rekap Kehadiran Singkat -->
              <div class="flex justify-between items-start text-xs pt-2">
                <div class="space-y-1">
                  <div>Jumlah Peserta Seharusnya : <strong>{{ printStudentsList.length }}</strong> Siswa</div>
                  <div>Jumlah Peserta Hadir &nbsp; &nbsp; &nbsp; &nbsp; : .......... Siswa</div>
                  <div>Jumlah Peserta Tidak Hadir &nbsp;: .......... Siswa</div>
                </div>
              </div>

              <!-- Tanda Tangan Footer Pengawas -->
              <div class="grid grid-cols-2 gap-8 pt-8 text-center text-xs">
                <div>
                  <p class="font-bold">Mengetahui,</p>
                  <p class="font-semibold">Kepala Sekolah / Ketua Panitia</p>
                  <div class="h-20"></div>
                  <p class="font-black underline uppercase">Drs. H. M. Sodiq, M.Pd.</p>
                  <p class="text-[11px] text-slate-600">NIP. 19680512 199403 1 005</p>
                </div>
                <div>
                  <p class="font-bold">Demak, {{ formatScheduleDateFull(printSchedule.start_time) }}</p>
                  <p class="font-semibold">Pengawas Ruang Ujian / Proktor</p>
                  <div class="h-20"></div>
                  <p class="font-black underline uppercase">....................................................</p>
                  <p class="text-[11px] text-slate-600">NIP. ............................................</p>
                </div>
              </div>
            </div>

            <!-- DOKUMEN 2: BERITA ACARA UJIAN -->
            <div v-else class="space-y-4">
              <div class="text-center space-y-1">
                <h2 class="text-base font-black uppercase tracking-wider underline">BERITA ACARA PELAKSANAAN UJIAN</h2>
                <p class="text-xs font-bold uppercase tracking-wide text-slate-800">{{ selectedExamEvent?.name || 'UJIAN SEKOLAH BERBASIS KOMPUTER (CBT)' }}</p>
                <p class="text-[11px] text-slate-600">TAHUN PELAJARAN 2025/2026</p>
              </div>

              <p class="text-xs leading-relaxed text-justify pt-2">
                Pada hari ini <strong>{{ getDayName(printSchedule.start_time) }}</strong>, tanggal <strong>{{ formatScheduleDateFull(printSchedule.start_time) }}</strong>, bertempat di <strong>SMAS Islamic Centre Demak</strong>, telah diselenggarakan Ujian Berbasis Komputer (CBT) untuk sesi:
              </p>

              <!-- Metadata Berita Acara -->
              <table class="w-full text-xs border-collapse pl-4">
                <tbody>
                  <tr>
                    <td class="py-1 w-36 font-bold">Mata Pelajaran</td>
                    <td class="py-1 w-4">:</td>
                    <td class="py-1 font-semibold">{{ printSchedule.subject?.name || printSchedule.bank?.subject?.name || '-' }}</td>
                  </tr>
                  <tr>
                    <td class="py-1 font-bold">Tingkat / Kelas</td>
                    <td class="py-1">:</td>
                    <td class="py-1 font-semibold">{{ printSchedule.class_room?.name || '-' }}</td>
                  </tr>
                  <tr>
                    <td class="py-1 font-bold">Waktu Pelaksanaan</td>
                    <td class="py-1">:</td>
                    <td class="py-1">{{ formatScheduleTimeOnly(printSchedule.start_time) }} - {{ formatScheduleTimeOnly(printSchedule.end_time) }} WIB ({{ printSchedule.duration_minutes }} Menit)</td>
                  </tr>
                  <tr>
                    <td class="py-1 font-bold">Kode Token Sesi</td>
                    <td class="py-1">:</td>
                    <td class="py-1 font-mono font-bold">{{ printSchedule.exam_token }}</td>
                  </tr>
                </tbody>
              </table>

              <!-- Tabel Rekapitulasi Peserta -->
              <div class="space-y-2 pt-2">
                <p class="font-bold text-xs">I. REKAPITULASI KEHADIRAN PESERTA :</p>
                <table class="w-full border-collapse border border-slate-900 text-xs">
                  <thead>
                    <tr class="bg-slate-100 border-b border-slate-900 text-center font-bold">
                      <th class="border border-slate-900 py-2 px-3">Jumlah Terdaftar</th>
                      <th class="border border-slate-900 py-2 px-3">Jumlah Hadir</th>
                      <th class="border border-slate-900 py-2 px-3">Jumlah Tidak Hadir</th>
                      <th class="border border-slate-900 py-2 px-3">Nomor / Nama Peserta Tidak Hadir</th>
                    </tr>
                  </thead>
                  <tbody>
                    <tr class="text-center">
                      <td class="border border-slate-900 py-4 font-bold">{{ printStudentsList.length }} Siswa</td>
                      <td class="border border-slate-900 py-4 font-bold text-emerald-800">.......... Siswa</td>
                      <td class="border border-slate-900 py-4 font-bold text-rose-800">.......... Siswa</td>
                      <td class="border border-slate-900 py-4 text-left px-3 text-[11px] text-slate-500">
                        .........................................................................
                      </td>
                    </tr>
                  </tbody>
                </table>
              </div>

              <!-- Catatan Kejadian Khusus -->
              <div class="space-y-2 pt-2">
                <p class="font-bold text-xs">II. CATATAN / KEJADIAN KHUSUS SELAMA PELAKSANAAN UJIAN :</p>
                <div class="border border-slate-900 p-3 min-h-[90px] text-xs text-slate-600">
                  <p class="italic text-slate-400">Pelaksanaan ujian berlangsung tertib, aman, dan lancar tanpa kendala teknis yang berarti.</p>
                </div>
              </div>

              <p class="text-xs pt-2">
                Demikian Berita Acara ini dibuat dengan sesungguhnya dan penuh tanggung jawab untuk dipergunakan sebagaimana mestinya.
              </p>

              <!-- Signatures -->
              <div class="grid grid-cols-2 gap-8 pt-6 text-center text-xs">
                <div>
                  <p class="font-bold">Pengawas Ujian 1,</p>
                  <div class="h-20"></div>
                  <p class="font-black underline uppercase">....................................................</p>
                  <p class="text-[11px] text-slate-600">NIP. ............................................</p>
                </div>
                <div>
                  <p class="font-bold">Pengawas Ujian 2 / Proktor,</p>
                  <div class="h-20"></div>
                  <p class="font-black underline uppercase">....................................................</p>
                  <p class="text-[11px] text-slate-600">NIP. ............................................</p>
                </div>
              </div>
            </div>

          </div>
        </div>

        <!-- Footer Modal Actions (no-print) -->
        <div class="px-6 py-4 bg-slate-50 border-t border-slate-200 flex items-center justify-between no-print shrink-0">
          <div class="text-xs text-slate-500">
            💡 Tips: Pastikan pengaturan printer memilih ukuran kertas <strong>A4</strong> dan margin <strong>Default / None</strong>.
          </div>
          <div class="flex items-center gap-2">
            <button
              type="button"
              @click="triggerPrintDocument()"
              class="px-4 py-2 bg-emerald-600 hover:bg-emerald-700 active:scale-95 text-white font-bold text-xs rounded-xl shadow-xs transition flex items-center gap-1.5 cursor-pointer"
            >
              <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17 17h2a2 2 0 002-2v-4a2 2 0 00-2-2H5a2 2 0 00-2 2v4a2 2 0 002 2h2m2 4h6a2 2 0 002-2v-4a2 2 0 00-2-2H9a2 2 0 00-2 2v4a2 2 0 002 2zm8-12V5a2 2 0 00-2-2H9a2 2 0 00-2 2v4h10z" />
              </svg>
              <span>🖨️ Cetak / Print Dokumen</span>
            </button>
            <button
              type="button"
              @click="showPrintModal = false"
              class="px-4 py-2 bg-slate-200 hover:bg-slate-300 text-slate-700 font-bold text-xs rounded-xl transition cursor-pointer"
            >
              Tutup
            </button>
          </div>
        </div>
      </div>
    </div>

    <!-- MODAL: UNGGAH / IMPOR BANK SOAL EXCEL -->
    <div v-if="showUploadBankModal" class="fixed inset-0 z-50 flex items-center justify-center p-4 select-none">
      <!-- Backdrop with pure opacity fade (NO scale) -->
      <transition
        appear
        enter-active-class="transition-opacity duration-200 ease-out"
        enter-from-class="opacity-0"
        enter-to-class="opacity-100"
        leave-active-class="transition-opacity duration-150 ease-in"
        leave-from-class="opacity-100"
        leave-to-class="opacity-0"
      >
        <div
          class="fixed inset-0 bg-slate-900/60 backdrop-blur-xs transition-opacity"
          @click="showUploadBankModal = false"
        ></div>
      </transition>

      <!-- Modal Card -->
      <transition
        appear
        enter-active-class="transition duration-200 ease-out transform"
        enter-from-class="opacity-0 scale-95 translate-y-2 sm:translate-y-0"
        enter-to-class="opacity-100 scale-100 translate-y-0"
        leave-active-class="transition duration-150 ease-in transform"
        leave-from-class="opacity-100 scale-100 translate-y-0"
        leave-to-class="opacity-0 scale-95 translate-y-2 sm:translate-y-0"
      >
        <div
          class="relative bg-white rounded-3xl max-w-lg w-full p-6 shadow-2xl space-y-4 z-10"
          @click.stop
        >
          <!-- Header -->
          <div class="flex items-center justify-between border-b border-slate-100 pb-3">
            <h3 class="text-base font-bold text-slate-900">Unggah Butir Soal Excel</h3>
            <button
              @click="showUploadBankModal = false"
              class="text-slate-400 hover:text-slate-600 p-1.5 rounded-xl hover:bg-slate-100 transition cursor-pointer"
            >
              ✕
            </button>
          </div>

          <p class="text-xs text-slate-600">
            Unggah paket butir soal pilihan ganda beserta opsi A–E dan kunci jawaban secara instan menggunakan format template resmi.
          </p>

          <!-- Template Download Box -->
          <div class="p-3 bg-indigo-50/50 rounded-2xl border border-indigo-100 flex items-center justify-between">
            <div class="flex items-center space-x-2">
              <svg class="w-5 h-5 text-indigo-600" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 10v6m0 0l-3-3m3 3l3-3m2 8H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
              </svg>
              <span class="text-xs font-semibold text-slate-700">Template_Bank_Soal_CBT.xlsx</span>
            </div>
            <button @click="downloadQuestionBankTemplate" class="text-xs font-bold text-indigo-600 hover:underline cursor-pointer">
              Unduh Format
            </button>
          </div>

          <form @submit.prevent="submitQuestionBankUpload" class="space-y-4 text-xs">
            <div>
              <label class="block font-bold text-slate-700 mb-1">Pilih Bank Soal Tujuan:</label>
              <select
                v-model="selectedBankUploadId"
                required
                class="w-full px-3 py-2.5 rounded-xl border border-slate-300 font-medium focus:ring-2 focus:ring-indigo-600 bg-slate-50 text-xs"
              >
                <option value="" disabled>-- Pilih Bank Soal --</option>
                <option v-for="b in readinessData.question_banks || []" :key="b.id" :value="b.id">
                  {{ b.title }} ({{ b.subject?.name }}) - {{ b.total_questions }} Soal ({{ b.is_locked ? 'Terkunci' : 'Draft' }})
                </option>
              </select>
            </div>

            <div>
              <label class="block font-bold text-slate-700 mb-1">Pilih Berkas Excel (.xlsx):</label>
              <input
                type="file"
                accept=".xlsx, .xls"
                @change="handleQuestionBankFileChange"
                required
                class="w-full text-xs text-slate-500 file:mr-3 file:py-2 file:px-4 file:rounded-xl file:border-0 file:text-xs file:font-semibold file:bg-indigo-50 file:text-indigo-700 hover:file:bg-indigo-100 cursor-pointer border border-slate-200 rounded-xl p-2"
              />
            </div>

            <div v-if="questionBankUploadStatus" :class="['p-3 rounded-xl text-xs font-semibold', questionBankUploadStatus.success ? 'bg-emerald-50 text-emerald-800 border border-emerald-200' : 'bg-rose-50 text-rose-800 border border-rose-200']">
              {{ questionBankUploadStatus.message }}
            </div>

            <div class="pt-2 flex justify-end gap-2 border-t border-slate-100">
              <button
                type="button"
                @click="showUploadBankModal = false"
                class="px-4 py-2 bg-slate-100 hover:bg-slate-200 active:scale-95 text-slate-700 font-bold text-xs rounded-xl transition cursor-pointer"
                :disabled="isUploadingQuestionBank"
              >
                Batal
              </button>
              <button
                type="submit"
                :disabled="isUploadingQuestionBank"
                class="px-5 py-2.5 bg-emerald-600 hover:bg-emerald-700 active:scale-95 disabled:opacity-50 text-white font-bold text-xs rounded-xl shadow-xs transition flex items-center space-x-1 cursor-pointer"
              >
                <svg v-if="isUploadingQuestionBank" class="animate-spin w-3.5 h-3.5 mr-1" fill="none" viewBox="0 0 24 24">
                  <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
                  <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8v8H4z"></path>
                </svg>
                <span>{{ isUploadingQuestionBank ? 'Mengimpor...' : 'Impor Soal' }}</span>
              </button>
            </div>
          </form>
        </div>
      </transition>
    </div>

    <!-- MODAL: BUAT BANK SOAL BARU -->
    <div v-if="showCreateBankModal" class="fixed inset-0 z-50 flex items-center justify-center p-4 select-none">
      <!-- Backdrop with pure opacity fade (NO scale) -->
      <transition
        appear
        enter-active-class="transition-opacity duration-200 ease-out"
        enter-from-class="opacity-0"
        enter-to-class="opacity-100"
        leave-active-class="transition-opacity duration-150 ease-in"
        leave-from-class="opacity-100"
        leave-to-class="opacity-0"
      >
        <div
          class="fixed inset-0 bg-slate-900/60 backdrop-blur-xs transition-opacity"
          @click="showCreateBankModal = false"
        ></div>
      </transition>

      <!-- Modal Card -->
      <transition
        appear
        enter-active-class="transition duration-200 ease-out transform"
        enter-from-class="opacity-0 scale-95 translate-y-2 sm:translate-y-0"
        enter-to-class="opacity-100 scale-100 translate-y-0"
        leave-active-class="transition duration-150 ease-in transform"
        leave-from-class="opacity-100 scale-100 translate-y-0"
        leave-to-class="opacity-0 scale-95 translate-y-2 sm:translate-y-0"
      >
        <div
          class="relative bg-white rounded-3xl max-w-md w-full p-6 shadow-2xl space-y-4 z-10"
          @click.stop
        >
          <!-- Header -->
          <div class="flex items-center justify-between border-b border-slate-100 pb-3">
            <h3 class="text-base font-bold text-slate-900">{{ isEditBank ? 'Edit Paket Bank Soal' : 'Buat Paket Bank Soal Baru' }}</h3>
            <button
              @click="showCreateBankModal = false"
              class="text-slate-400 hover:text-slate-600 p-1.5 rounded-xl hover:bg-slate-100 transition cursor-pointer"
            >
              ✕
            </button>
          </div>

          <form @submit.prevent="submitCreateBank" class="space-y-4 text-xs">
            <div>
              <label class="block font-bold text-slate-700 mb-1">Pilih Mata Pelajaran:</label>
              <select
                v-model="newBankForm.subject_id"
                @change="onBankSubjectChange"
                required
                :disabled="availableBankSubjects.length === 0 && !isEditBank"
                class="w-full px-3 py-2.5 rounded-xl border border-slate-300 font-medium focus:ring-2 focus:ring-indigo-600 bg-slate-50 text-xs cursor-pointer disabled:bg-slate-100 disabled:text-slate-400"
              >
                <option value="" disabled>
                  {{ (availableBankSubjects.length === 0 && !isEditBank) ? '-- Semua Mata Pelajaran Sudah Memiliki Bank Soal --' : '-- Pilih Mata Pelajaran --' }}
                </option>
                <option v-for="s in availableBankSubjects" :key="s.id" :value="s.id">
                  {{ s.name }} ({{ s.code }})
                </option>
              </select>
              <p v-if="availableBankSubjects.length > 0" class="text-[11px] text-slate-500 mt-1.5 flex items-center gap-1.5">
                <span class="w-1.5 h-1.5 rounded-full bg-indigo-600 shrink-0"></span>
                <span>Hanya menampilkan mata pelajaran jadwal yang belum dibuatkan bank soal.</span>
              </p>
              <p v-else-if="!isEditBank" class="text-[11px] text-emerald-600 font-semibold mt-1.5 flex items-center gap-1.5">
                <span class="w-1.5 h-1.5 rounded-full bg-emerald-500 shrink-0"></span>
                <span>Seluruh mata pelajaran jadwal event ini sudah memiliki bank soal.</span>
              </p>
            </div>

            <div>
              <label class="block font-bold text-slate-700 mb-1">Judul / Nama Paket Bank Soal:</label>
              <input
                v-model="newBankForm.title"
                type="text"
                required
                :disabled="availableBankSubjects.length === 0 && !isEditBank"
                placeholder="Contoh: ASAT Matematika Wajib Kelas XII MIPA"
                class="w-full px-3 py-2.5 rounded-xl border border-slate-300 font-medium focus:ring-2 focus:ring-indigo-600 text-xs disabled:bg-slate-100 disabled:text-slate-400"
              />
            </div>

            <div class="pt-2 flex justify-end gap-2 border-t border-slate-100">
              <button
                type="button"
                @click="showCreateBankModal = false"
                class="px-4 py-2 bg-slate-100 hover:bg-slate-200 active:scale-95 text-slate-700 font-bold text-xs rounded-xl transition cursor-pointer"
                :disabled="isCreatingBank"
              >
                Batal
              </button>
              <button
                type="submit"
                :disabled="isCreatingBank || (availableBankSubjects.length === 0 && !isEditBank)"
                class="px-5 py-2.5 bg-indigo-600 hover:bg-indigo-700 active:scale-95 disabled:opacity-50 text-white font-bold text-xs rounded-xl shadow-xs transition flex items-center space-x-1 cursor-pointer"
              >
                <svg v-if="isCreatingBank" class="animate-spin w-3.5 h-3.5 mr-1" fill="none" viewBox="0 0 24 24">
                  <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
                  <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8v8H4z"></path>
                </svg>
                <span>{{ isCreatingBank ? 'Menyimpan...' : (isEditBank ? 'Perbarui Bank Soal' : 'Simpan Bank Soal') }}</span>
              </button>
            </div>
          </form>
        </div>
      </transition>
    </div>

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

    <!-- GLOBAL REACTIVE CONFIRM / ALERT MODAL -->
    <div v-if="dialogState.isOpen" class="fixed inset-0 z-50 flex items-center justify-center p-4 select-none">
      <!-- Backdrop with pure opacity fade (NO scale) -->
      <transition
        appear
        enter-active-class="transition-opacity duration-200 ease-out"
        enter-from-class="opacity-0"
        enter-to-class="opacity-100"
        leave-active-class="transition-opacity duration-150 ease-in"
        leave-from-class="opacity-100"
        leave-to-class="opacity-0"
      >
        <div
          class="fixed inset-0 bg-slate-900/60 backdrop-blur-xs transition-opacity"
          @click="handleDialogCancel"
        ></div>
      </transition>

      <!-- Modal Card with smooth scale & fade -->
      <transition
        appear
        enter-active-class="transition duration-200 ease-out transform"
        enter-from-class="opacity-0 scale-95 translate-y-2 sm:translate-y-0"
        enter-to-class="opacity-100 scale-100 translate-y-0"
        leave-active-class="transition duration-150 ease-in transform"
        leave-from-class="opacity-100 scale-100 translate-y-0"
        leave-to-class="opacity-0 scale-95 translate-y-2 sm:translate-y-0"
      >
        <div
          class="relative bg-white rounded-3xl max-w-md w-full p-6 shadow-2xl space-y-4 z-10"
          @click.stop
        >
          <div class="flex items-center justify-between border-b border-slate-100 pb-3">
            <h3 class="text-base font-bold text-slate-900">{{ dialogState.title }}</h3>
            <button
              @click="handleDialogCancel"
              class="text-slate-400 hover:text-slate-600 p-1.5 rounded-xl hover:bg-slate-100 transition cursor-pointer"
            >
              ✕
            </button>
          </div>

          <div class="text-xs text-slate-600 leading-relaxed whitespace-pre-line py-1">
            {{ dialogState.message }}
          </div>

          <div class="flex justify-end gap-2 pt-2 border-t border-slate-100">
            <button
              v-if="dialogState.cancelText"
              type="button"
              @click="handleDialogCancel"
              class="px-4 py-2 bg-slate-100 hover:bg-slate-200 active:scale-95 text-slate-700 font-bold text-xs rounded-xl transition cursor-pointer"
            >
              {{ dialogState.cancelText }}
            </button>
            <button
              type="button"
              @click="handleDialogConfirm"
              :class="[
                'px-5 py-2.5 text-white font-bold text-xs rounded-xl shadow-xs transition active:scale-95 cursor-pointer',
                dialogState.type === 'danger' ? 'bg-rose-600 hover:bg-rose-700' :
                dialogState.type === 'warning' ? 'bg-amber-600 hover:bg-amber-700' :
                'bg-indigo-600 hover:bg-indigo-700'
              ]"
            >
              {{ dialogState.confirmText }}
            </button>
          </div>
        </div>
      </transition>
    </div>

    <!-- Global Toast Notification -->
    <transition enter-active-class="transform ease-out duration-300 transition" enter-from-class="translate-y-2 opacity-0 sm:translate-y-0 sm:translate-x-2" enter-to-class="translate-y-0 opacity-100 sm:translate-x-0" leave-active-class="transition ease-in duration-100" leave-from-class="opacity-100" leave-to-class="opacity-0">
      <div v-if="toastMessage" class="fixed bottom-5 right-5 z-50 flex items-center gap-2 px-4 py-3 bg-slate-900 text-white rounded-2xl shadow-xl text-xs font-semibold no-print">
        <span v-if="toastType === 'success'" class="text-emerald-400">✓</span>
        <span v-else class="text-indigo-400">ℹ</span>
        <span>{{ toastMessage }}</span>
      </div>
    </transition>

    <!-- STUDENT EXAM SIMULATOR MODAL -->
    <StudentExamSimulatorModal
      :is-open="showSimulatorModal"
      :bank="viewingBank"
      :questions="bankQuestions"
      @close="showSimulatorModal = false"
    />
  </div>
</template>

<script setup>
import { ref, computed, watch, onMounted, onUnmounted } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '../../stores/auth'
import api from '../../services/api'
import RichContentRenderer from '../../components/common/RichContentRenderer.vue'
import StudentExamSimulatorModal from '../../components/admin/StudentExamSimulatorModal.vue'
import LiveProctorControl from '@/components/proctor/LiveProctorControl.vue'
import ProctorPrintModal from '@/components/proctor/ProctorPrintModal.vue'
import ScheduleProctorModal from '@/components/admin/ScheduleProctorModal.vue'
import StaffAccessEditor from '../../components/admin/StaffAccessEditor.vue'
import { landingTab } from '../../utils/access'
import {
  applyImplications,
  effectivePermissions,
  templateByKey,
  templateKeyFor,
  templateLabelFor,
} from '../../utils/staffAccess'

const router = useRouter()
const authStore = useAuthStore()

// Label portal mengikuti izin, bukan role
const portalLabel = computed(() =>
  authStore.hasPermission('master:manage') ? 'Portal Kurikulum & Admin' : 'Portal Guru / Pengawas'
)
const portalRoleLabel = computed(() =>
  authStore.hasPermission('master:manage') ? 'Kurikulum & CBT' : 'Guru & Pengawas'
)

// Sidebar & Layout State
const activeTab = ref('dashboard')
const isSidebarCollapsed = ref(false)
const isMobileSidebarOpen = ref(false)

// Dynamic Header Title & Tag based on current menu
const currentSectionMeta = computed(() => {
  switch (activeTab.value) {
    case 'dashboard':
      return {
        title: 'Beranda',
        tag: 'Ringkasan',
        description: 'Pusat kendali dan monitoring menyeluruh aktivitas ujian sekolah'
      }
    case 'events':
      return {
        title: 'Event Ujian',
        tag: 'Periode Ujian',
        description: 'Kelola kalender pekan ujian (PSAT, ASAT, PTS, PAT, Gladi Bersih, US)'
      }
    case 'schedules':
      return {
        title: 'Jadwal Ujian',
        tag: 'Pelaksanaan',
        description: 'Pengaturan sesi ruang, alokasi rombel, token sesi, dan durasi ujian'
      }
    case 'questions':
      return {
        title: 'Bank Soal & Kesiapan',
        tag: 'Naskah & Kesiapan',
        description: 'Pengelolaan paket naskah butir soal, impor file Excel, dan finalisasi status siap ujian'
      }
    case 'proctor':
      return {
        title: 'Live Proctoring',
        tag: 'Pengawasan',
        description: 'Pemantauan real-time integritas pengerjaan siswa, pelanggaran layar, dan berita acara ujian'
      }
    case 'students':
      return {
        title: 'Data Siswa',
        tag: 'Master Data',
        description: 'Daftar data siswa, nomor induk (NIS/NISN), alokasi kelas, dan reset sesi perangkat'
      }
    case 'teachers':
      return {
        title: 'Guru dan Staf',
        tag: 'Master Data',
        description: 'Daftar akun pendidik pengampu mata pelajaran, penyusun naskah soal, dan staf pengelola CBT'
      }
    case 'classes':
      return {
        title: 'Data Kelas',
        tag: 'Master Data',
        description: 'Struktur ruang kelas, tingkatan (grade), dan konsentrasi keahlian/jurusan'
      }
    case 'subjects':
      return {
        title: 'Mata Pelajaran',
        tag: 'Master Data',
        description: 'Daftar mata pelajaran kurikulum dan kode unik bidang studi asesmen'
      }
    case 'class-subjects':
      return {
        title: 'Kelas Mapel',
        tag: 'Master Data',
        description: 'Alokasi mata pelajaran per rombel dan penetapan guru pengampu'
      }
    default:
      return {
        title: 'Portal CBT',
        tag: 'Kurikulum',
        description: 'Pusat kontrol dan tata kelola ujian sekolah'
      }
  }
})

// Event Management & Scoping States
const events = ref([])
const selectedEventId = ref(sessionStorage.getItem('cbt_selected_event_id') || '')
const selectedEventFilter = ref(sessionStorage.getItem('cbt_selected_event_id') || '')
const isEventDropdownOpen = ref(false)

const selectedExamEvent = computed(() => {
  const id = selectedEventId.value || selectedEventFilter.value
  if (!id) return null
  return events.value.find(e => e.id === id) || null
})

const activeExamEvent = computed(() => {
  return events.value.find(e => e.is_active) || null
})

const currentScopedEventId = computed(() => {
  return selectedExamEvent.value?.id || selectedEventId.value || selectedEventFilter.value || null
})

const currentEventSchedules = computed(() => {
  if (!currentScopedEventId.value) return schedules.value
  return schedules.value.filter(s => s.event_id === currentScopedEventId.value)
})

// Contextual Metric Computeds
const eventsCount = computed(() => events.value.length)
const activeEventsCount = computed(() => events.value.filter(e => e.is_active).length)
const totalSchedulesCount = computed(() => currentEventSchedules.value.length)
const activeSchedulesCount = computed(() => currentEventSchedules.value.filter(s => s.is_active).length)
const scheduledClassesCount = computed(() => new Set(currentEventSchedules.value.map(s => s.class_room_id || s.class_id)).size)
const selectedEventFilterName = computed(() => {
  const ev = selectedExamEvent.value || (currentScopedEventId.value ? events.value.find(e => e.id === currentScopedEventId.value) : null)
  return ev ? ev.code : 'Semua Event'
})

const lockedStudentsCount = computed(() => students.value.filter(s => s.user?.session_token).length)
const maleStudentsCount = computed(() => students.value.filter(s => s.gender === 'L' || s.user?.gender === 'L').length)
const femaleStudentsCount = computed(() => students.value.filter(s => s.gender === 'P' || s.user?.gender === 'P').length)
// Guru dan Pengawas = semua akun non-Administrator (termasuk Kustom), sehingga jumlahnya konsisten dengan Total
const guruCount = computed(() => teachers.value.filter(t => t.role !== 'ADMIN').length)
const adminStaffCount = computed(() => teachers.value.filter(t => t.role === 'ADMIN').length)

const questionMakersCount = computed(() => {
  const banks = readinessData.value.question_banks || []
  return new Set(banks.map(b => b.created_by_id)).size
})

const subjectsWithBanksCount = computed(() => {
  const banks = readinessData.value.question_banks || []
  return new Set(banks.map(b => b.subject_id)).size
})

const allocatedClassesCount = computed(() => new Set(classSubjects.value.map(cs => cs.class_id)).size)
const allocatedTeachersCount = computed(() => new Set(classSubjects.value.map(cs => cs.teacher_id)).size)

const getSubjectBankCount = (subId) => {
  const banks = readinessData.value.question_banks || []
  return banks.filter(b => b.subject_id === subId).length
}

const getSubjectClassCount = (subId) => {
  return classSubjects.value.filter(cs => cs.subject_id === subId).length
}

const gradeCounts = computed(() => {
  const counts = { 'X': 0, 'XI': 0, 'XII': 0 }
  classes.value.forEach(c => {
    if (counts[c.grade] !== undefined) {
      counts[c.grade]++
    }
  })
  return counts
})

const lockedBanksCount = computed(() => (readinessData.value.question_banks || []).filter(b => b.is_locked).length)
const draftBanksCount = computed(() => (readinessData.value.question_banks || []).filter(b => !b.is_locked).length)
const totalQuestionsCount = computed(() => (readinessData.value.question_banks || []).reduce((acc, b) => acc + (b.total_questions || 0), 0))

// Cloudflare-style Event Context Switcher Actions (Two-Tier Model)

const enterEvent = (ev) => {
  selectedEventId.value = ev.id
  selectedEventFilter.value = ev.id
  sessionStorage.setItem('cbt_selected_event_id', ev.id)
  isEventDropdownOpen.value = false
  activeTab.value = authStore.canAccessTab('schedules') ? 'schedules' : landingTab(authStore.user)
  loadSchedules()
}

const selectEvent = (ev) => {
  selectedEventId.value = ev.id
  selectedEventFilter.value = ev.id
  sessionStorage.setItem('cbt_selected_event_id', ev.id)
  isEventDropdownOpen.value = false
  loadSchedules()
}

const exitEventScope = () => {
  selectedEventId.value = ''
  selectedEventFilter.value = ''
  sessionStorage.removeItem('cbt_selected_event_id')
  isEventDropdownOpen.value = false
  activeTab.value = 'events'
}

const stats = ref({})
const schedules = ref([])

// Global Reactive Dialog State & Methods (Replaces native browser alert & confirm)
const dialogState = ref({
  isOpen: false,
  type: 'confirm', // 'confirm' | 'danger' | 'alert' | 'warning'
  title: '',
  message: '',
  confirmText: 'OK',
  cancelText: '',
  resolve: null
})

const showConfirmModal = ({ title = 'Konfirmasi', message = '', type = 'confirm', confirmText = 'Lanjutkan', cancelText = 'Batal' }) => {
  return new Promise((resolve) => {
    dialogState.value = {
      isOpen: true,
      type,
      title,
      message,
      confirmText,
      cancelText,
      resolve
    }
  })
}

const showAlertModal = ({ title = 'Pemberitahuan', message = '', type = 'alert', confirmText = 'Mengerti' }) => {
  return new Promise((resolve) => {
    dialogState.value = {
      isOpen: true,
      type,
      title,
      message,
      confirmText,
      cancelText: '',
      resolve
    }
  })
}

const handleDialogConfirm = () => {
  if (dialogState.value.resolve) dialogState.value.resolve(true)
  dialogState.value.isOpen = false
}

const handleDialogCancel = () => {
  if (dialogState.value.resolve) dialogState.value.resolve(false)
  dialogState.value.isOpen = false
}

// Toast Notification State
const toastMessage = ref('')
const toastType = ref('success')
let toastTimeout = null
const showToast = (msg, type = 'success') => {
  toastMessage.value = msg
  toastType.value = type
  if (toastTimeout) clearTimeout(toastTimeout)
  toastTimeout = setTimeout(() => {
    toastMessage.value = ''
  }, 3000)
}

const copyTokenToClipboard = (token) => {
  if (!token) return
  if (navigator.clipboard && navigator.clipboard.writeText) {
    navigator.clipboard.writeText(token).then(() => {
      showToast(`Token ${token} berhasil disalin ke clipboard!`, 'success')
    }).catch(() => {
      showToast(`Token: ${token}`, 'info')
    })
  } else {
    showToast(`Token: ${token}`, 'info')
  }
}

// Live Status Helper
const getScheduleLiveStatus = (sch) => {
  if (!sch || !sch.start_time || !sch.end_time) return 'upcoming'
  const now = Date.now()
  const start = new Date(sch.start_time).getTime()
  const end = new Date(sch.end_time).getTime()
  if (now >= start && now <= end) return 'live'
  if (now < start) return 'upcoming'
  return 'finished'
}

// Schedule Filtering, Sorting & Detail States
const scheduleSearchQuery = ref('')
const selectedScheduleGrade = ref('all')
const selectedScheduleTimeFilter = ref('all') // 'all' | 'today' | 'upcoming'
const scheduleSortKey = ref('time')
const scheduleSortOrder = ref('asc')
const showScheduleDetailModal = ref(false)
const selectedScheduleDetail = ref(null)

// Essay Correction State (for Schedule Detail Modal)
const scheduleDetailTab = ref('info') // 'info' | 'essay'
const essayQuestions = ref([])
const essayLoading = ref(false)
const essayDraft = ref({}) // keyed by answer_id: { score_awarded: number|null, teacher_comment: string }
const essaySaving = ref({}) // keyed by question_id: boolean

// Bulk Selection States & Logic
const selectedScheduleIds = ref([])

const toggleScheduleSort = (key) => {
  if (scheduleSortKey.value === key) {
    scheduleSortOrder.value = scheduleSortOrder.value === 'asc' ? 'desc' : 'asc'
  } else {
    scheduleSortKey.value = key
    scheduleSortOrder.value = 'asc'
  }
}

const openScheduleDetail = (sch) => {
  selectedScheduleDetail.value = sch
  showScheduleDetailModal.value = true
  scheduleDetailTab.value = 'info'
  essayQuestions.value = []
  essayDraft.value = {}
  essaySaving.value = {}
}

const fetchEssayAnswers = async (scheduleId) => {
  if (essayLoading.value) return
  essayLoading.value = true
  try {
    const res = await api.get(`/api/v1/admin/schedules/${scheduleId}/essay-answers`)
    essayQuestions.value = res.data.data || []
    // Initialize draft from existing data
    const draft = {}
    for (const q of essayQuestions.value) {
      for (const a of q.answers || []) {
        draft[a.answer_id] = {
          score_awarded: a.score_awarded !== null && a.score_awarded !== undefined ? a.score_awarded : null,
          teacher_comment: a.teacher_comment || ''
        }
      }
    }
    essayDraft.value = draft
  } catch (e) {
    showToast('Gagal memuat data jawaban essay.', 'error')
  } finally {
    essayLoading.value = false
  }
}

const switchToEssayTab = () => {
  scheduleDetailTab.value = 'essay'
  if (essayQuestions.value.length === 0 && !essayLoading.value && selectedScheduleDetail.value) {
    fetchEssayAnswers(selectedScheduleDetail.value.id)
  }
}

const essayTotalPending = computed(() => {
  return essayQuestions.value.reduce((sum, q) => sum + (q.total_count - q.graded_count), 0)
})

const saveEssayQuestion = async (question) => {
  const qid = question.question_id
  // Collect answers for this question
  const payload = (question.answers || []).map(a => {
    const d = essayDraft.value[a.answer_id] || {}
    return {
      answer_id: a.answer_id,
      score_awarded: d.score_awarded !== null && d.score_awarded !== undefined ? Number(d.score_awarded) : null,
      teacher_comment: d.teacher_comment || ''
    }
  })

  // Validate scores
  for (const p of payload) {
    if (p.score_awarded !== null) {
      if (p.score_awarded < 0) {
        showToast(`Nilai tidak boleh negatif (Soal No. ${question.question_number}).`, 'error')
        return
      }
      if (p.score_awarded > question.score_weight) {
        showToast(`Nilai melebihi bobot soal ${question.score_weight} (Soal No. ${question.question_number}).`, 'error')
        return
      }
    }
  }

  essaySaving.value = { ...essaySaving.value, [qid]: true }
  try {
    await api.patch(`/api/v1/admin/schedules/${selectedScheduleDetail.value.id}/essay-answers`, payload)
    showToast(`Penilaian soal No. ${question.question_number} berhasil disimpan.`, 'success')
    // Refresh essay data
    await fetchEssayAnswers(selectedScheduleDetail.value.id)
  } catch (e) {
    showToast('Gagal menyimpan penilaian.', 'error')
  } finally {
    essaySaving.value = { ...essaySaving.value, [qid]: false }
  }
}

const availableScheduleGrades = computed(() => {
  const grades = classes.value.map(c => c.grade).filter(Boolean)
  const unique = [...new Set(grades)]
  const order = ['X', 'XI', 'XII', '7', '8', '9', 'VII', 'VIII', 'IX', '10', '11', '12']
  return unique.sort((a, b) => {
    const idxA = order.indexOf(a)
    const idxB = order.indexOf(b)
    if (idxA !== -1 && idxB !== -1) return idxA - idxB
    return a.localeCompare(b)
  })
})

const filteredSchedules = computed(() => {
  let list = currentEventSchedules.value

  // Filter by Grade
  if (selectedScheduleGrade.value !== 'all') {
    list = list.filter(sch => {
      const g = sch.class_room?.grade || (classes.value.find(c => c.id === (sch.class_room_id || sch.class_id))?.grade)
      return g === selectedScheduleGrade.value
    })
  }

  // Filter by Time (Hari Ini / Mendatang / Semua)
  if (selectedScheduleTimeFilter.value === 'today') {
    const todayStr = new Date().toDateString()
    list = list.filter(sch => {
      if (!sch.start_time) return false
      return new Date(sch.start_time).toDateString() === todayStr
    })
  } else if (selectedScheduleTimeFilter.value === 'upcoming') {
    const now = Date.now()
    list = list.filter(sch => {
      if (!sch.start_time) return false
      return new Date(sch.start_time).getTime() > now
    })
  }

  // Filter by Search Query
  if (scheduleSearchQuery.value.trim()) {
    const q = scheduleSearchQuery.value.toLowerCase().trim()
    list = list.filter(sch => {
      const titleMatch = sch.title && sch.title.toLowerCase().includes(q)
      const subMatch = (sch.subject?.name && sch.subject.name.toLowerCase().includes(q)) ||
                       (sch.bank?.subject?.name && sch.bank.subject.name.toLowerCase().includes(q))
      const classMatch = sch.class_room?.name && sch.class_room.name.toLowerCase().includes(q)
      const tokenMatch = sch.exam_token && sch.exam_token.toLowerCase().includes(q)
      return titleMatch || subMatch || classMatch || tokenMatch
    })
  }

  // Sort
  if (scheduleSortKey.value) {
    list = [...list].sort((a, b) => {
      let result = 0
      if (scheduleSortKey.value === 'title') {
        const valA = (a.title || '').trim().toLowerCase()
        const valB = (b.title || '').trim().toLowerCase()
        result = valA.localeCompare(valB, undefined, { numeric: true, sensitivity: 'base' })
      } else if (scheduleSortKey.value === 'time') {
        const timeA = a.start_time ? new Date(a.start_time).getTime() : 0
        const timeB = b.start_time ? new Date(b.start_time).getTime() : 0
        result = timeA - timeB
      } else if (scheduleSortKey.value === 'class') {
        const gradeA = a.class_room?.grade || (classes.value.find(c => c.id === (a.class_room_id || a.class_id))?.grade) || ''
        const gradeB = b.class_room?.grade || (classes.value.find(c => c.id === (b.class_room_id || b.class_id))?.grade) || ''
        const nameA = (a.class_room?.name || '').trim().toLowerCase()
        const nameB = (b.class_room?.name || '').trim().toLowerCase()
        if (gradeA !== gradeB) {
          const order = ['X', 'XI', 'XII', '7', '8', '9', 'VII', 'VIII', 'IX', '10', '11', '12']
          const idxA = order.indexOf(gradeA)
          const idxB = order.indexOf(gradeB)
          if (idxA !== -1 && idxB !== -1) {
            result = idxA - idxB
          } else {
            result = gradeA.localeCompare(gradeB)
          }
        }
        if (result === 0) {
          result = nameA.localeCompare(nameB, undefined, { numeric: true, sensitivity: 'base' })
        }
      } else if (scheduleSortKey.value === 'subject') {
        const nameA = (a.subject?.name || a.bank?.subject?.name || '').trim().toLowerCase()
        const nameB = (b.subject?.name || b.bank?.subject?.name || '').trim().toLowerCase()
        result = nameA.localeCompare(nameB, undefined, { numeric: true, sensitivity: 'base' })
      } else if (scheduleSortKey.value === 'status') {
        const actA = a.is_active ? 1 : 0
        const actB = b.is_active ? 1 : 0
        if (actA !== actB) {
          result = actA - actB
        } else {
          const tokA = (a.exam_token || '').trim().toLowerCase()
          const tokB = (b.exam_token || '').trim().toLowerCase()
          result = tokA.localeCompare(tokB)
        }
      }
      return scheduleSortOrder.value === 'asc' ? result : -result
    })
  }

  return list
})

// Schedule Pagination States & Logic
const scheduleCurrentPage = ref(1)
const schedulePerPage = ref(10)

const totalSchedulePages = computed(() => {
  return Math.max(1, Math.ceil(filteredSchedules.value.length / schedulePerPage.value))
})

const paginatedSchedules = computed(() => {
  const start = (scheduleCurrentPage.value - 1) * schedulePerPage.value
  return filteredSchedules.value.slice(start, start + schedulePerPage.value)
})

const displayedSchedulePages = computed(() => {
  const total = totalSchedulePages.value
  const current = scheduleCurrentPage.value
  if (total <= 7) {
    return Array.from({ length: total }, (_, i) => i + 1)
  }
  if (current <= 4) {
    return [1, 2, 3, 4, 5, '...', total]
  } else if (current >= total - 3) {
    return [1, '...', total - 4, total - 3, total - 2, total - 1, total]
  } else {
    return [1, '...', current - 1, current, current + 1, '...', total]
  }
})

const setSchedulePage = (p) => {
  if (p === '...' || p < 1 || p > totalSchedulePages.value) return
  scheduleCurrentPage.value = p
}

const isAllPaginatedSelected = computed(() => {
  if (!paginatedSchedules.value.length) return false
  return paginatedSchedules.value.every(s => selectedScheduleIds.value.includes(s.id))
})

const toggleSelectAllPaginatedSchedules = () => {
  if (!canManageSchedules.value) return
  const pageIds = paginatedSchedules.value.map(s => s.id)
  if (isAllPaginatedSelected.value) {
    selectedScheduleIds.value = selectedScheduleIds.value.filter(id => !pageIds.includes(id))
  } else {
    selectedScheduleIds.value = [...new Set([...selectedScheduleIds.value, ...pageIds])]
  }
}

// Bulk Actions
const bulkActivateSchedules = async () => {
  if (!canManageSchedules.value || !selectedScheduleIds.value.length) return
  const toActivate = schedules.value.filter(s => selectedScheduleIds.value.includes(s.id))
  const invalid = toActivate.filter(s => !s.bank_id || (s.bank && !s.bank.is_locked))
  if (invalid.length > 0) {
    await showAlertModal({
      title: 'Validasi Bank Soal',
      message: `Ada ${invalid.length} sesi jadwal terpilih yang belum memiliki Bank Soal yang siap (terkunci).\n\nSilakan pastikan semua jadwal yang dipilih telah ditautkan dengan bank soal yang berstatus Siap/Terkunci.`
    })
    return
  }
  const confirmed = await showConfirmModal({
    title: 'Aktivasi Massal',
    message: `Aktifkan ${selectedScheduleIds.value.length} sesi jadwal ujian terpilih?`,
    confirmText: 'Aktifkan',
    cancelText: 'Batal'
  })
  if (!confirmed) return

  try {
    for (const sch of toActivate) {
      if (!sch.is_active) {
        const res = await api.post(`/admin/schedules/${sch.id}/toggle`)
        sch.is_active = res.data.is_active
      }
    }
    showToast(`${selectedScheduleIds.value.length} sesi jadwal berhasil diaktifkan!`, 'success')
  } catch (e) {
    showToast('Terjadi kesalahan saat mengaktifkan jadwal secara massal', 'error')
  }
}

const bulkDeactivateSchedules = async () => {
  if (!canManageSchedules.value || !selectedScheduleIds.value.length) return
  const confirmed = await showConfirmModal({
    title: 'Nonaktifkan Massal',
    message: `Nonaktifkan ${selectedScheduleIds.value.length} sesi jadwal ujian terpilih?`,
    confirmText: 'Nonaktifkan',
    cancelText: 'Batal'
  })
  if (!confirmed) return

  try {
    const toDeactivate = schedules.value.filter(s => selectedScheduleIds.value.includes(s.id))
    for (const sch of toDeactivate) {
      if (sch.is_active) {
        const res = await api.post(`/admin/schedules/${sch.id}/toggle`)
        sch.is_active = res.data.is_active
      }
    }
    showToast(`${selectedScheduleIds.value.length} sesi jadwal dinonaktifkan!`, 'success')
  } catch (e) {
    showToast('Terjadi kesalahan saat menonaktifkan jadwal secara massal', 'error')
  }
}

const bulkRegenerateTokens = async () => {
  if (!canManageSchedules.value || !selectedScheduleIds.value.length) return
  const confirmed = await showConfirmModal({
    title: 'Acak Token Massal',
    message: `Acak ulang token untuk ${selectedScheduleIds.value.length} sesi jadwal terpilih?`,
    confirmText: 'Acak Token',
    cancelText: 'Batal'
  })
  if (!confirmed) return

  try {
    const chars = 'ABCDEFGHJKLMNPQRSTUVWXYZ23456789'
    const toRegen = schedules.value.filter(s => selectedScheduleIds.value.includes(s.id))
    for (const sch of toRegen) {
      let newToken = ''
      for (let i = 0; i < 6; i++) {
        newToken += chars.charAt(Math.floor(Math.random() * chars.length))
      }
      const res = await api.put(`/admin/schedules/${sch.id}`, {
        title: sch.title,
        bank_id: sch.bank_id,
        class_room_id: sch.class_room_id || sch.class_id,
        start_time: sch.start_time,
        end_time: sch.end_time,
        duration_minutes: sch.duration_minutes,
        max_violations: sch.max_violations,
        randomize_questions: sch.randomize_questions,
        randomize_options: sch.randomize_options,
        exam_token: newToken
      })
      sch.exam_token = res.data.exam_token || newToken
    }
    showToast(`Berhasil mengacak token untuk ${selectedScheduleIds.value.length} sesi!`, 'success')
  } catch (e) {
    showToast('Gagal mengacak token massal', 'error')
  }
}

// Print Document States & Methods
const showPrintModal = ref(false)
const printDocType = ref('attendance') // 'attendance' | 'report'
const printSchedule = ref(null)

const openPrintModal = (sch, docType = 'attendance') => {
  printSchedule.value = sch
  printDocType.value = docType
  showPrintModal.value = true
}

const printStudentsList = computed(() => {
  if (!printSchedule.value) return []
  const targetClassId = printSchedule.value.class_room_id || printSchedule.value.class_id
  return students.value
    .filter(s => (s.class_room_id === targetClassId || s.class_id === targetClassId))
    .sort((a, b) => (a.user?.full_name || a.name || '').localeCompare(b.user?.full_name || b.name || ''))
})

const triggerPrintDocument = () => {
  window.print()
}

const formatScheduleDateFull = (dt) => {
  if (!dt) return '-'
  return new Date(dt).toLocaleDateString('id-ID', {
    weekday: 'long',
    day: 'numeric',
    month: 'long',
    year: 'numeric'
  })
}

const formatScheduleTimeOnly = (dt) => {
  if (!dt) return '-'
  return new Date(dt).toLocaleTimeString('id-ID', {
    hour: '2-digit',
    minute: '2-digit'
  })
}

const getDayName = (dt) => {
  if (!dt) return '..................'
  return new Date(dt).toLocaleDateString('id-ID', { weekday: 'long' })
}

watch([scheduleSearchQuery, selectedScheduleGrade, selectedScheduleTimeFilter, schedulePerPage, scheduleSortKey, scheduleSortOrder], () => {
  scheduleCurrentPage.value = 1
})

watch(totalSchedulePages, (newTotal) => {
  if (scheduleCurrentPage.value > newTotal) {
    scheduleCurrentPage.value = Math.max(1, newTotal)
  }
})

const students = ref([])
const teachers = ref([])
const classes = ref([])
const readinessData = ref({})

// Subject & ClassSubject States
const subjects = ref([])
const showSubjectModal = ref(false)
const isEditSubject = ref(false)
const subjectForm = ref({ id: null, code: '', name: '' })

const classSubjects = ref([])
const showClassSubjectModal = ref(false)
const classSubjectForm = ref({
  id: null,
  class_id: '',
  subject_id: '',
  teacher_id: '',
  academic_year: '2026/2027',
})
const filterClassSubjectClass = ref('')

// Event Management States
const eventSearchQuery = ref('')
const eventStatusFilter = ref('all')
const activeEventMenuId = ref(null)

const toggleEventMenu = (id) => {
  activeEventMenuId.value = activeEventMenuId.value === id ? null : id
}
const showEventModal = ref(false)
const isEditEvent = ref(false)
const eventForm = ref({
  id: null,
  title: '',
  code: '',
  academic_year: '2026/2027',
  semester: 'GANJIL',
  start_date: '',
  end_date: '',
  is_active: true,
  description: '',
})

const filteredEvents = computed(() => {
  let list = events.value
  if (eventStatusFilter.value === 'active') {
    list = list.filter(e => e.is_active)
  } else if (eventStatusFilter.value === 'archived') {
    list = list.filter(e => !e.is_active)
  }
  if (eventSearchQuery.value.trim()) {
    const q = eventSearchQuery.value.toLowerCase().trim()
    list = list.filter(e =>
      (e.title && e.title.toLowerCase().includes(q)) ||
      (e.code && e.code.toLowerCase().includes(q)) ||
      (e.academic_year && e.academic_year.toLowerCase().includes(q))
    )
  }
  return list
})

// Modals
const showScheduleModal = ref(false)
const showStudentModal = ref(false)
const showTeacherModal = ref(false)
const showClassModal = ref(false)
const showImportModal = ref(false)
const importExcelFile = ref(null)
const isImporting = ref(false)

// Import states — Kelas
const showImportClassModal = ref(false)
const importClassFile = ref(null)
const isImportingClass = ref(false)
const importClassResult = ref(null)

// Import states — Mata Pelajaran
const showImportSubjectModal = ref(false)
const importSubjectFile = ref(null)
const isImportingSubject = ref(false)
const importSubjectResult = ref(null)

// Import states — Guru
const showImportTeacherModal = ref(false)
const importTeacherFile = ref(null)
const isImportingTeacher = ref(false)
const importTeacherResult = ref(null)

// Schedule Form & Link Bank States
const isEditSchedule = ref(false)
const scheduleForm = ref({
  id: null,
  event_id: '',
  title: '',
  subject_id: '',
  bank_id: '',
  class_id: '',
  exam_date: '',
  start_time: '07:30',
  end_time: '09:00',
  exam_token: 'CBT2026',
  duration_minutes: 90,
  max_violations: 3,
  randomize_questions: true,
  randomize_options: true,
  is_active: true,
  proctors: [],
})

// Penugasan pengawas: modal dari tabel jadwal dan pemilih di dalam form jadwal
const showProctorAssignModal = ref(false)
const scheduleForProctors = ref(null)
const showFormProctorPicker = ref(false)
// Daftar id pengawas saat form dibuka, dipakai untuk mendeteksi perubahan saat menyimpan
const scheduleFormInitialProctorIds = ref([])

const openProctorAssign = (sch) => {
  scheduleForProctors.value = sch
  showProctorAssignModal.value = true
}

const formatProctorNames = (list) => {
  const names = (list || []).map(p => p.full_name).filter(Boolean)
  if (names.length <= 2) return names.join(', ')
  return `${names.slice(0, 2).join(', ')} +${names.length - 2}`
}

const applyFormProctors = (users) => {
  scheduleForm.value.proctors = users || []
}

const removeFormProctor = (id) => {
  scheduleForm.value.proctors = scheduleForm.value.proctors.filter(p => p.id !== id)
}

const sameIdSet = (a, b) => {
  if (a.length !== b.length) return false
  const setB = new Set(b)
  return a.every(id => setB.has(id))
}

const showLinkBankModal = ref(false)
const selectedScheduleForLink = ref(null)
const selectedBankForLink = ref('')

const availableBanksForScheduleForm = computed(() => {
  const allBanks = readinessData.value.question_banks || []
  if (!scheduleForm.value.subject_id) return allBanks
  const matching = allBanks.filter(b => b.subject_id === scheduleForm.value.subject_id)
  return matching.length > 0 ? matching : allBanks
})

const availableBanksForSelectedSchedule = computed(() => {
  const allBanks = readinessData.value.question_banks || []
  if (!selectedScheduleForLink.value) return allBanks
  const targetSubId = selectedScheduleForLink.value.subject_id || selectedScheduleForLink.value.bank?.subject_id
  if (!targetSubId) return allBanks
  const matching = allBanks.filter(b => b.subject_id === targetSubId)
  return matching.length > 0 ? matching : allBanks
})

const formatScheduleTimeRange = (startTime, endTime) => {
  if (!startTime) return '-'
  const st = new Date(startTime)
  const dateStr = st.toLocaleDateString('id-ID', { weekday: 'short', day: 'numeric', month: 'short', year: 'numeric' })
  const pad = (n) => String(n).padStart(2, '0')
  const timeStartStr = `${pad(st.getHours())}:${pad(st.getMinutes())}`
  if (!endTime) return `${dateStr} • ${timeStartStr} WIB`
  const et = new Date(endTime)
  const timeEndStr = `${pad(et.getHours())}:${pad(et.getMinutes())}`
  return `${dateStr} • ${timeStartStr} - ${timeEndStr} WIB`
}

const selectedScheduleClassSubjectId = ref('')

const alreadyScheduledKeys = computed(() => {
  const currentEventId = selectedExamEvent.value?.id || events.value.find(e => e.is_active)?.id || null
  const set = new Set()
  schedules.value.forEach(s => {
    if (!currentEventId || s.event_id === currentEventId) {
      const classId = s.class_room_id || s.class_id
      const subId = s.subject_id || s.bank?.subject_id
      if (classId && subId) {
        if (isEditSchedule.value && scheduleForm.value.id === s.id) {
          return
        }
        set.add(`${classId}_${subId}`)
      }
    }
  })
  return set
})

const availableClassSubjects = computed(() => {
  return classSubjects.value.filter(cs => {
    const key = `${cs.class_id}_${cs.subject_id}`
    if (isEditSchedule.value && selectedScheduleClassSubjectId.value === cs.id) return true
    return !alreadyScheduledKeys.value.has(key)
  })
})

const availableClassSubjectGrades = computed(() => {
  const grades = availableClassSubjects.value.map(cs => cs.class_room?.grade || 'Lainnya')
  const unique = [...new Set(grades)]
  const order = ['X', 'XI', 'XII', '7', '8', '9', 'VII', 'VIII', 'IX', '10', '11', '12']
  return unique.sort((a, b) => {
    const idxA = order.indexOf(a)
    const idxB = order.indexOf(b)
    if (idxA !== -1 && idxB !== -1) return idxA - idxB
    if (idxA !== -1) return -1
    if (idxB !== -1) return 1
    return String(a).localeCompare(String(b))
  })
})

const groupedAvailableClassSubjects = computed(() => {
  const grouped = {}
  availableClassSubjects.value.forEach(cs => {
    const grade = cs.class_room?.grade || 'Lainnya'
    if (!grouped[grade]) grouped[grade] = []
    grouped[grade].push(cs)
  })
  return grouped
})

const onClassSubjectChange = () => {
  const cs = classSubjects.value.find(c => c.id === selectedScheduleClassSubjectId.value)
  if (!cs) return
  scheduleForm.value.class_id = cs.class_id
  scheduleForm.value.subject_id = cs.subject_id
  if (!scheduleForm.value.title || scheduleForm.value.title.startsWith('Ujian ') || scheduleForm.value.title.startsWith('Penilaian ')) {
    scheduleForm.value.title = `Ujian ${cs.subject?.name || ''} - ${cs.class_room?.name || ''}`
  }
}

const openCreateSchedule = (defaultSubjectId = null) => {
  isEditSchedule.value = false
  const now = new Date()
  const pad = (n) => String(n).padStart(2, '0')
  const todayStr = `${now.getFullYear()}-${pad(now.getMonth() + 1)}-${pad(now.getDate())}`
  const activeEv = selectedExamEvent.value || events.value.find(e => e.is_active) || events.value[0]

  // Pick first available class-subject
  const firstCS = availableClassSubjects.value[0]
  selectedScheduleClassSubjectId.value = firstCS ? firstCS.id : ''
  const subId = firstCS ? firstCS.subject_id : (defaultSubjectId || subjects.value[0]?.id || '')
  const clsId = firstCS ? firstCS.class_id : (classes.value[0]?.id || '')
  const defaultTitle = firstCS ? `Ujian ${firstCS.subject?.name || ''} - ${firstCS.class_room?.name || ''}` : ''

  scheduleForm.value = {
    id: null,
    event_id: activeEv ? activeEv.id : '',
    title: defaultTitle,
    subject_id: subId,
    bank_id: '',
    class_id: clsId,
    exam_date: todayStr,
    start_time: '07:30',
    end_time: '09:00',
    exam_token: 'CBT' + Math.floor(1000 + Math.random() * 9000),
    duration_minutes: 90,
    max_violations: 3,
    randomize_questions: true,
    randomize_options: true,
    is_active: true,
    proctors: [],
  }
  scheduleFormInitialProctorIds.value = []
  showScheduleModal.value = true
}

const openEditSchedule = (sch) => {
  isEditSchedule.value = true
  const st = sch.start_time ? new Date(sch.start_time) : new Date()
  const et = sch.end_time ? new Date(sch.end_time) : new Date(st.getTime() + (sch.duration_minutes || 90) * 60000)

  const pad = (n) => String(n).padStart(2, '0')
  const dateStr = `${st.getFullYear()}-${pad(st.getMonth() + 1)}-${pad(st.getDate())}`
  const startTimeStr = `${pad(st.getHours())}:${pad(st.getMinutes())}`
  const endTimeStr = `${pad(et.getHours())}:${pad(et.getMinutes())}`

  const targetClassId = sch.class_room_id || sch.class_id
  const targetSubId = sch.subject_id || sch.bank?.subject_id
  const matchedCS = classSubjects.value.find(cs => cs.class_id === targetClassId && cs.subject_id === targetSubId)
  selectedScheduleClassSubjectId.value = matchedCS ? matchedCS.id : ''

  scheduleForm.value = {
    id: sch.id,
    event_id: sch.event_id || '',
    title: sch.title || '',
    subject_id: targetSubId || (subjects.value[0]?.id || ''),
    bank_id: sch.bank_id || '',
    class_id: targetClassId || (classes.value[0]?.id || ''),
    exam_date: dateStr,
    start_time: startTimeStr,
    end_time: endTimeStr,
    exam_token: sch.exam_token || 'CBT2026',
    duration_minutes: sch.duration_minutes || 90,
    max_violations: sch.max_violations ?? 3,
    randomize_questions: sch.randomize_questions ?? true,
    randomize_options: sch.randomize_options ?? true,
    is_active: sch.is_active ?? true,
    proctors: (sch.proctors || []).map(p => ({ id: p.id, full_name: p.full_name })),
  }
  scheduleFormInitialProctorIds.value = (sch.proctors || []).map(p => p.id)
  showScheduleModal.value = true
}

const submitScheduleForm = async () => {
  try {
    if (!scheduleForm.value.title || !scheduleForm.value.class_id || !scheduleForm.value.subject_id) {
      await showAlertModal({
        title: 'Form Belum Lengkap',
        message: 'Judul sesi, kelas, dan mata pelajaran wajib diisi.',
        type: 'warning'
      })
      return
    }

    const activeEv = selectedExamEvent.value || events.value.find(e => e.is_active) || events.value[0]
    const eventId = scheduleForm.value.event_id || (activeEv ? activeEv.id : null)

    const payload = {
      event_id: eventId,
      title: scheduleForm.value.title,
      subject_id: scheduleForm.value.subject_id,
      bank_id: scheduleForm.value.bank_id || null,
      class_id: scheduleForm.value.class_id,
      exam_date: scheduleForm.value.exam_date,
      start_time: scheduleForm.value.start_time,
      end_time: scheduleForm.value.end_time,
      exam_token: scheduleForm.value.exam_token.toUpperCase(),
      duration_minutes: scheduleForm.value.duration_minutes,
      max_violations: scheduleForm.value.max_violations,
      randomize_questions: scheduleForm.value.randomize_questions,
      randomize_options: scheduleForm.value.randomize_options,
      is_active: scheduleForm.value.is_active,
    }

    let savedScheduleId = scheduleForm.value.id
    if (isEditSchedule.value) {
      await api.put(`/admin/schedules/${scheduleForm.value.id}`, payload)
      showToast('Jadwal ujian berhasil diperbarui!', 'success')
    } else {
      const createRes = await api.post('/admin/schedules', payload)
      savedScheduleId = createRes.data?.data?.id || null
      showToast('Jadwal ujian berhasil diterbitkan!', 'success')
    }

    // Penugasan pengawas dikirim terpisah setelah jadwal tersimpan; kegagalannya tidak membatalkan jadwal.
    let proctorWarning = ''
    if (authStore.hasPermission('schedules:manage')) {
      const proctorIds = scheduleForm.value.proctors.map(p => p.id)
      const needSync = isEditSchedule.value
        ? !sameIdSet(proctorIds, scheduleFormInitialProctorIds.value)
        : proctorIds.length > 0
      if (needSync) {
        if (!savedScheduleId) {
          proctorWarning = 'ID jadwal tidak ditemukan pada respons server.'
        } else {
          try {
            await api.put(`/admin/schedules/${savedScheduleId}/proctors`, { user_ids: proctorIds })
          } catch (pe) {
            proctorWarning = pe.response?.data?.message || 'Terjadi kesalahan saat menyimpan pengawas.'
          }
        }
      }
    }

    showScheduleModal.value = false
    await loadSchedules()
    await loadAllData()

    if (proctorWarning) {
      await showAlertModal({
        title: 'Pengawas Belum Tersimpan',
        message: `Jadwal sudah tersimpan, tetapi penugasan pengawas gagal: ${proctorWarning} Atur ulang pengawas lewat tombol "Atur Pengawas" pada baris jadwal.`,
        type: 'warning'
      })
    }
  } catch (e) {
    await showAlertModal({
      title: 'Gagal Menyimpan Jadwal',
      message: e.response?.data?.message || 'Terjadi kesalahan saat menyimpan data jadwal ujian.',
      type: 'danger'
    })
  }
}

const deleteSchedule = async (sch) => {
  const confirmed = await showConfirmModal({
    title: 'Hapus Sesi Ujian',
    message: `Apakah Anda yakin ingin menghapus sesi ujian "${sch.title}" untuk kelas ${sch.class_room?.name || ''}?`,
    type: 'danger',
    confirmText: 'Hapus Sesi',
    cancelText: 'Batal'
  })
  if (!confirmed) return

  try {
    await api.delete(`/admin/schedules/${sch.id}`)
    await loadSchedules()
    await loadAllData()
    showToast('Jadwal ujian berhasil dihapus', 'success')
  } catch (e) {
    await showAlertModal({
      title: 'Gagal Menghapus Jadwal',
      message: e.response?.data?.message || 'Terjadi kesalahan saat menghapus data jadwal ujian.',
      type: 'danger'
    })
  }
}

const openLinkBankModal = (sch) => {
  selectedScheduleForLink.value = sch
  selectedBankForLink.value = sch.bank_id || ''
  showLinkBankModal.value = true
}

const submitLinkBank = async () => {
  if (!selectedScheduleForLink.value) return
  try {
    await api.post(`/admin/schedules/${selectedScheduleForLink.value.id}/link-bank`, {
      bank_id: selectedBankForLink.value || null
    })
    showLinkBankModal.value = false
    await loadSchedules()
    await loadAllData()
    showToast('Naskah bank soal berhasil ditautkan ke jadwal!', 'success')
  } catch (e) {
    await showAlertModal({
      title: 'Gagal Menautkan Soal',
      message: e.response?.data?.message || 'Terjadi kesalahan saat menautkan bank soal.',
      type: 'danger'
    })
  }
}

// ---------------- MASTER DATA STATES & COMPUTEDS ----------------

// Form states for user/class
const newStudent = ref({ full_name: '', username: '', password: '', nis: '', nisn: '', class_id: '', gender: 'L' })
const newTeacher = ref({ full_name: '', username: '', password: '' })
const newTeacherAccess = ref({ role: 'GURU', permissions: [] })
const newTeacherAccessValid = ref(true)

// Hanya grantor (ADMIN atau pemegang izin "*") yang boleh mengatur izin, username, dan password akun staf.
// Server tetap menjadi penjaga sebenarnya (403); ini hanya untuk kerapian UI.
const isGrantor = computed(() => authStore.role === 'ADMIN' || authStore.permissions.includes('*'))
const newTeacherCanSave = computed(() => !isGrantor.value || newTeacherAccessValid.value)
const newClass = ref({ name: '', grade: 'XII', major: 'MIPA' })

// 1. Data Siswa Filtering, Sorting & Pagination
const studentSearchQuery = ref('')
const selectedStudentGrade = ref('all')
const selectedStudentSessionFilter = ref('all')
const studentSortKey = ref('name')
const studentSortOrder = ref('asc')
const studentCurrentPage = ref(1)
const studentPerPage = ref(10)
const showStudentDetailModal = ref(false)
const selectedStudentDetail = ref(null)
const showEditStudentModal = ref(false)
const editStudentForm = ref({ id: null, user_id: null, full_name: '', username: '', password: '', nis: '', nisn: '', class_id: '', gender: 'L' })

const availableStudentGrades = computed(() => {
  const grades = classes.value.map(c => c.grade).filter(Boolean)
  const unique = [...new Set(grades)]
  const order = ['X', 'XI', 'XII', '7', '8', '9', 'VII', 'VIII', 'IX', '10', '11', '12']
  return unique.sort((a, b) => {
    const idxA = order.indexOf(a)
    const idxB = order.indexOf(b)
    if (idxA !== -1 && idxB !== -1) return idxA - idxB
    return a.localeCompare(b)
  })
})

const toggleStudentSort = (key) => {
  if (studentSortKey.value === key) {
    studentSortOrder.value = studentSortOrder.value === 'asc' ? 'desc' : 'asc'
  } else {
    studentSortKey.value = key
    studentSortOrder.value = 'asc'
  }
}

const resetStudentFilters = () => {
  studentSearchQuery.value = ''
  selectedStudentGrade.value = 'all'
  selectedStudentSessionFilter.value = 'all'
  studentSortKey.value = 'name'
  studentSortOrder.value = 'asc'
  studentCurrentPage.value = 1
}

const filteredStudents = computed(() => {
  let list = students.value || []

  // Filter by Grade
  if (selectedStudentGrade.value !== 'all') {
    list = list.filter(st => {
      const g = st.class_room?.grade || (classes.value.find(c => c.id === (st.class_room_id || st.class_id))?.grade)
      return g === selectedStudentGrade.value
    })
  }

  // Filter by Session
  if (selectedStudentSessionFilter.value === 'locked') {
    list = list.filter(st => st.user?.session_token)
  }

  // Filter by Search Query
  if (studentSearchQuery.value.trim()) {
    const q = studentSearchQuery.value.toLowerCase().trim()
    list = list.filter(st => {
      const nameMatch = st.user?.full_name && st.user.full_name.toLowerCase().includes(q)
      const userMatch = st.user?.username && st.user.username.toLowerCase().includes(q)
      const nisMatch = st.nis && st.nis.toLowerCase().includes(q)
      const nisnMatch = st.nisn && st.nisn.toLowerCase().includes(q)
      const classMatch = st.class_room?.name && st.class_room.name.toLowerCase().includes(q)
      return nameMatch || userMatch || nisMatch || nisnMatch || classMatch
    })
  }

  // Sort
  if (studentSortKey.value) {
    list = [...list].sort((a, b) => {
      let result = 0
      if (studentSortKey.value === 'name') {
        const valA = (a.user?.full_name || '').trim().toLowerCase()
        const valB = (b.user?.full_name || '').trim().toLowerCase()
        result = valA.localeCompare(valB, undefined, { numeric: true, sensitivity: 'base' })
      } else if (studentSortKey.value === 'username') {
        const valA = (a.user?.username || '').trim().toLowerCase()
        const valB = (b.user?.username || '').trim().toLowerCase()
        result = valA.localeCompare(valB, undefined, { numeric: true, sensitivity: 'base' })
      } else if (studentSortKey.value === 'nis') {
        const valA = (a.nis || '').trim().toLowerCase()
        const valB = (b.nis || '').trim().toLowerCase()
        result = valA.localeCompare(valB, undefined, { numeric: true, sensitivity: 'base' })
      } else if (studentSortKey.value === 'class') {
        const valA = (a.class_room?.name || '').trim().toLowerCase()
        const valB = (b.class_room?.name || '').trim().toLowerCase()
        result = valA.localeCompare(valB, undefined, { numeric: true, sensitivity: 'base' })
      } else if (studentSortKey.value === 'session') {
        const lockA = a.user?.session_token ? 1 : 0
        const lockB = b.user?.session_token ? 1 : 0
        result = lockA - lockB
      }
      return studentSortOrder.value === 'asc' ? result : -result
    })
  }

  return list
})

const totalStudentPages = computed(() => Math.ceil(filteredStudents.value.length / studentPerPage.value) || 1)
const paginatedStudents = computed(() => {
  const start = (studentCurrentPage.value - 1) * studentPerPage.value
  return filteredStudents.value.slice(start, start + studentPerPage.value)
})

const displayedStudentPages = computed(() => {
  const total = totalStudentPages.value
  const current = studentCurrentPage.value
  if (total <= 7) {
    return Array.from({ length: total }, (_, i) => i + 1)
  }
  const pages = []
  pages.push(1)
  if (current > 3) pages.push('...')
  const start = Math.max(2, current - 1)
  const end = Math.min(total - 1, current + 1)
  for (let i = start; i <= end; i++) {
    pages.push(i)
  }
  if (current < total - 2) pages.push('...')
  pages.push(total)
  return pages
})

const setStudentPage = (p) => {
  if (p === '...' || p < 1 || p > totalStudentPages.value) return
  studentCurrentPage.value = p
}

watch([studentSearchQuery, selectedStudentGrade, selectedStudentSessionFilter, studentPerPage, studentSortKey, studentSortOrder], () => {
  studentCurrentPage.value = 1
})

// 2. Guru dan Staf Filtering, Sorting & Pagination
const teacherSearchQuery = ref('')
const selectedTeacherRole = ref('all')
const teacherSortKey = ref('name')
const teacherSortOrder = ref('asc')
const teacherCurrentPage = ref(1)
const teacherPerPage = ref(10)
const showEditTeacherModal = ref(false)
const editTeacherForm = ref({ id: null, full_name: '', username: '', password: '' })
const editTeacherAccess = ref({ role: 'GURU', permissions: [] })
const editTeacherAccessValid = ref(true)
const editTeacherCanSave = computed(() => !isGrantor.value || editTeacherAccessValid.value)

// Katalog izin (GET /admin/permission-catalog): dimuat sekali, dipakai editor akses dan label peran turunan.
const permissionCatalog = ref(null)
let permissionCatalogRequest = null
const loadPermissionCatalog = () => {
  if (permissionCatalog.value) return Promise.resolve(permissionCatalog.value)
  if (!permissionCatalogRequest) {
    permissionCatalogRequest = api.get('/admin/permission-catalog')
      .then((res) => {
        const d = res.data?.data || {}
        permissionCatalog.value = {
          groups: d.groups || [],
          templates: d.templates || [],
          implies: d.implies || {},
        }
        return permissionCatalog.value
      })
      .finally(() => { permissionCatalogRequest = null })
  }
  return permissionCatalogRequest
}

const ensurePermissionCatalog = async () => {
  try {
    return await loadPermissionCatalog()
  } catch (e) {
    await showAlertModal({
      title: 'Gagal Memuat Daftar Izin',
      message: e.response?.data?.message || 'Daftar izin akses tidak dapat dimuat. Coba lagi beberapa saat.',
      type: 'danger'
    })
    return null
  }
}

// Peran turunan akun staf: 'admin' | 'guru' | 'pengawas' | 'custom'
const teacherKey = (t) => templateKeyFor(t, permissionCatalog.value?.templates, permissionCatalog.value?.implies)
const teacherRoleLabel = (t) => templateLabelFor(t, permissionCatalog.value?.templates, permissionCatalog.value?.implies)
const TEACHER_BADGE_CLASSES = {
  admin: 'bg-purple-100 text-purple-800',
  guru: 'bg-indigo-100 text-indigo-800',
  pengawas: 'bg-emerald-100 text-emerald-800',
  custom: 'bg-amber-100 text-amber-800',
  unknown: 'bg-slate-100 text-slate-500',
}
const teacherBadgeClass = (t) => TEACHER_BADGE_CLASSES[teacherKey(t)]

const TEACHER_ROLE_CHIPS = [
  { key: 'admin', label: 'Administrator' },
  { key: 'guru', label: 'Guru' },
  { key: 'pengawas', label: 'Pengawas' },
  { key: 'custom', label: 'Kustom' },
]
const teacherRoleChips = computed(() => {
  const list = teachers.value || []
  const counts = { admin: 0, guru: 0, pengawas: 0, custom: 0 }
  for (const t of list) {
    const key = teacherKey(t)
    if (key in counts) counts[key]++ // 'unknown' (katalog belum termuat) tidak dihitung ke kategori mana pun
  }
  const chips = [{ key: 'all', label: `Semua (${list.length})` }]
  for (const c of TEACHER_ROLE_CHIPS) {
    if (counts[c.key] > 0 || selectedTeacherRole.value === c.key) {
      chips.push({ key: c.key, label: `${c.label} (${counts[c.key]})` })
    }
  }
  return chips
})

// Nilai awal editor: template Guru (izin terisi dari template).
const defaultStaffAccess = (catalog) => {
  const guru = templateByKey(catalog.templates, 'guru')
  return { role: 'GURU', permissions: applyImplications(guru?.permissions, catalog.implies) }
}

// Nilai awal editor dari akun: izin kosong pada akun GURU = izin template Guru.
// Kunci yang tidak ada di katalog (usang atau tak dikenal, mis. "*") disaring lebih dulu.
const staffAccessFromAccount = (account, catalog) => {
  if (account.role === 'ADMIN') return { role: 'ADMIN', permissions: ['*'] }
  const known = new Set(catalog.groups.flatMap((g) => (g.items || []).map((i) => i.key)))
  const effective = effectivePermissions(account, catalog.templates).filter((p) => known.has(p))
  return { role: 'GURU', permissions: applyImplications(effective, catalog.implies) }
}

// Payload sesuai kontrak. Grantor: ADMIN tanpa permissions, GURU dengan permissions eksplisit.
// Bukan grantor: tidak boleh mengirim permissions, role dipaksa GURU.
const buildStaffCreatePayload = (form, access) => {
  const payload = { full_name: form.full_name, username: form.username, role: isGrantor.value ? access.role : 'GURU' }
  if (form.password) payload.password = form.password
  if (isGrantor.value && access.role === 'GURU') payload.permissions = [...access.permissions]
  return payload
}

// Bukan grantor: hanya full_name yang boleh diubah (tanpa role, permissions, username, password).
const buildStaffUpdatePayload = (form, access) => {
  if (!isGrantor.value) return { full_name: form.full_name }
  return buildStaffCreatePayload(form, access)
}

const toggleTeacherSort = (key) => {
  if (teacherSortKey.value === key) {
    teacherSortOrder.value = teacherSortOrder.value === 'asc' ? 'desc' : 'asc'
  } else {
    teacherSortKey.value = key
    teacherSortOrder.value = 'asc'
  }
}

const filteredTeachers = computed(() => {
  let list = teachers.value || []
  if (selectedTeacherRole.value !== 'all') {
    list = list.filter(t => teacherKey(t) === selectedTeacherRole.value)
  }
  if (teacherSearchQuery.value.trim()) {
    const q = teacherSearchQuery.value.toLowerCase().trim()
    list = list.filter(t => {
      const nameMatch = t.full_name && t.full_name.toLowerCase().includes(q)
      const userMatch = t.username && t.username.toLowerCase().includes(q)
      return nameMatch || userMatch
    })
  }
  if (teacherSortKey.value) {
    list = [...list].sort((a, b) => {
      let result = 0
      if (teacherSortKey.value === 'name') {
        const valA = (a.full_name || '').trim().toLowerCase()
        const valB = (b.full_name || '').trim().toLowerCase()
        result = valA.localeCompare(valB, undefined, { numeric: true, sensitivity: 'base' })
      } else if (teacherSortKey.value === 'username') {
        const valA = (a.username || '').trim().toLowerCase()
        const valB = (b.username || '').trim().toLowerCase()
        result = valA.localeCompare(valB, undefined, { numeric: true, sensitivity: 'base' })
      } else if (teacherSortKey.value === 'role') {
        const valA = teacherRoleLabel(a).toLowerCase()
        const valB = teacherRoleLabel(b).toLowerCase()
        result = valA.localeCompare(valB, undefined, { numeric: true, sensitivity: 'base' })
      }
      return teacherSortOrder.value === 'asc' ? result : -result
    })
  }
  return list
})

const totalTeacherPages = computed(() => Math.ceil(filteredTeachers.value.length / teacherPerPage.value) || 1)
const paginatedTeachers = computed(() => {
  const start = (teacherCurrentPage.value - 1) * teacherPerPage.value
  return filteredTeachers.value.slice(start, start + teacherPerPage.value)
})

const displayedTeacherPages = computed(() => {
  const total = totalTeacherPages.value
  const current = teacherCurrentPage.value
  if (total <= 7) return Array.from({ length: total }, (_, i) => i + 1)
  const pages = [1]
  if (current > 3) pages.push('...')
  const start = Math.max(2, current - 1)
  const end = Math.min(total - 1, current + 1)
  for (let i = start; i <= end; i++) pages.push(i)
  if (current < total - 2) pages.push('...')
  pages.push(total)
  return pages
})

const setTeacherPage = (p) => {
  if (p === '...' || p < 1 || p > totalTeacherPages.value) return
  teacherCurrentPage.value = p
}

watch([teacherSearchQuery, selectedTeacherRole, teacherPerPage, teacherSortKey, teacherSortOrder], () => {
  teacherCurrentPage.value = 1
})

// 3. Data Kelas Filtering, Sorting & Pagination
const classSearchQuery = ref('')
const selectedClassGrade = ref('all')
const classSortKey = ref('name')
const classSortOrder = ref('asc')
const classCurrentPage = ref(1)
const classPerPage = ref(10)
const showEditClassModal = ref(false)
const editClassForm = ref({ id: null, name: '', grade: 'XII', major: 'MIPA' })

const toggleClassSort = (key) => {
  if (classSortKey.value === key) {
    classSortOrder.value = classSortOrder.value === 'asc' ? 'desc' : 'asc'
  } else {
    classSortKey.value = key
    classSortOrder.value = 'asc'
  }
}

const filteredClasses = computed(() => {
  let list = classes.value || []
  if (selectedClassGrade.value !== 'all') {
    list = list.filter(c => c.grade === selectedClassGrade.value)
  }
  if (classSearchQuery.value.trim()) {
    const q = classSearchQuery.value.toLowerCase().trim()
    list = list.filter(c => {
      const nameMatch = c.name && c.name.toLowerCase().includes(q)
      const gradeMatch = c.grade && c.grade.toLowerCase().includes(q)
      const majorMatch = c.major && c.major.toLowerCase().includes(q)
      return nameMatch || gradeMatch || majorMatch
    })
  }
  if (classSortKey.value) {
    list = [...list].sort((a, b) => {
      let result = 0
      if (classSortKey.value === 'name') {
        const valA = (a.name || '').trim().toLowerCase()
        const valB = (b.name || '').trim().toLowerCase()
        result = valA.localeCompare(valB, undefined, { numeric: true, sensitivity: 'base' })
      } else if (classSortKey.value === 'grade') {
        const order = ['X', 'XI', 'XII', '7', '8', '9', 'VII', 'VIII', 'IX', '10', '11', '12']
        const idxA = order.indexOf(a.grade)
        const idxB = order.indexOf(b.grade)
        if (idxA !== -1 && idxB !== -1) result = idxA - idxB
        else result = (a.grade || '').localeCompare(b.grade || '')
      } else if (classSortKey.value === 'major') {
        const valA = (a.major || '').trim().toLowerCase()
        const valB = (b.major || '').trim().toLowerCase()
        result = valA.localeCompare(valB, undefined, { numeric: true, sensitivity: 'base' })
      }
      return classSortOrder.value === 'asc' ? result : -result
    })
  }
  return list
})

const totalClassPages = computed(() => Math.ceil(filteredClasses.value.length / classPerPage.value) || 1)
const paginatedClasses = computed(() => {
  const start = (classCurrentPage.value - 1) * classPerPage.value
  return filteredClasses.value.slice(start, start + classPerPage.value)
})

const displayedClassPages = computed(() => {
  const total = totalClassPages.value
  const current = classCurrentPage.value
  if (total <= 7) return Array.from({ length: total }, (_, i) => i + 1)
  const pages = [1]
  if (current > 3) pages.push('...')
  const start = Math.max(2, current - 1)
  const end = Math.min(total - 1, current + 1)
  for (let i = start; i <= end; i++) pages.push(i)
  if (current < total - 2) pages.push('...')
  pages.push(total)
  return pages
})

const setClassPage = (p) => {
  if (p === '...' || p < 1 || p > totalClassPages.value) return
  classCurrentPage.value = p
}

watch([classSearchQuery, selectedClassGrade, classPerPage, classSortKey, classSortOrder], () => {
  classCurrentPage.value = 1
})

// 4. Mata Pelajaran Filtering, Sorting & Pagination
const subjectSearchQuery = ref('')
const subjectSortKey = ref('name')
const subjectSortOrder = ref('asc')
const subjectCurrentPage = ref(1)
const subjectPerPage = ref(10)

const toggleSubjectSort = (key) => {
  if (subjectSortKey.value === key) {
    subjectSortOrder.value = subjectSortOrder.value === 'asc' ? 'desc' : 'asc'
  } else {
    subjectSortKey.value = key
    subjectSortOrder.value = 'asc'
  }
}

const filteredSubjects = computed(() => {
  let list = subjects.value || []
  if (subjectSearchQuery.value.trim()) {
    const q = subjectSearchQuery.value.toLowerCase().trim()
    list = list.filter(s => {
      const codeMatch = s.code && s.code.toLowerCase().includes(q)
      const nameMatch = s.name && s.name.toLowerCase().includes(q)
      return codeMatch || nameMatch
    })
  }
  if (subjectSortKey.value) {
    list = [...list].sort((a, b) => {
      let result = 0
      if (subjectSortKey.value === 'code') {
        const valA = (a.code || '').trim().toLowerCase()
        const valB = (b.code || '').trim().toLowerCase()
        result = valA.localeCompare(valB, undefined, { numeric: true, sensitivity: 'base' })
      } else if (subjectSortKey.value === 'name') {
        const valA = (a.name || '').trim().toLowerCase()
        const valB = (b.name || '').trim().toLowerCase()
        result = valA.localeCompare(valB, undefined, { numeric: true, sensitivity: 'base' })
      }
      return subjectSortOrder.value === 'asc' ? result : -result
    })
  }
  return list
})

const totalSubjectPages = computed(() => Math.ceil(filteredSubjects.value.length / subjectPerPage.value) || 1)
const paginatedSubjects = computed(() => {
  const start = (subjectCurrentPage.value - 1) * subjectPerPage.value
  return filteredSubjects.value.slice(start, start + subjectPerPage.value)
})

const displayedSubjectPages = computed(() => {
  const total = totalSubjectPages.value
  const current = subjectCurrentPage.value
  if (total <= 7) return Array.from({ length: total }, (_, i) => i + 1)
  const pages = [1]
  if (current > 3) pages.push('...')
  const start = Math.max(2, current - 1)
  const end = Math.min(total - 1, current + 1)
  for (let i = start; i <= end; i++) pages.push(i)
  if (current < total - 2) pages.push('...')
  pages.push(total)
  return pages
})

const setSubjectPage = (p) => {
  if (p === '...' || p < 1 || p > totalSubjectPages.value) return
  subjectCurrentPage.value = p
}

watch([subjectSearchQuery, subjectPerPage, subjectSortKey, subjectSortOrder], () => {
  subjectCurrentPage.value = 1
})

const formatDate = (d) => {
  if (!d) return '-'
  const date = new Date(d)
  return date.toLocaleDateString('id-ID', { day: 'numeric', month: 'short', year: 'numeric' })
}

const EVENT_SCOPED_TABS = ['schedules', 'questions', 'proctor']

const switchTab = (tab, targetScheduleId = null) => {
  // Menu yang tidak diizinkan dialihkan ke menu pertama yang boleh diakses
  if (!authStore.canAccessTab(tab)) {
    tab = landingTab(authStore.user)
  }
  // Menu pelaksanaan ujian baru dapat dibuka setelah event dipilih
  if (EVENT_SCOPED_TABS.includes(tab) && !currentScopedEventId.value) {
    tab = 'events'
  }
  activeTab.value = tab
  isMobileSidebarOpen.value = false

  if (tab === 'events') {
    loadEvents()
  } else if (tab === 'schedules') {
    loadSchedules()
  } else if (tab === 'students') {
    api.get('/admin/students').then(res => students.value = res.data.data || [])
  } else if (tab === 'teachers') {
    api.get('/admin/teachers').then(res => teachers.value = res.data.data || [])
    // Katalog izin dipakai untuk label peran turunan dan editor akses
    loadPermissionCatalog().catch(() => {})
  } else if (tab === 'classes') {
    api.get('/admin/classes').then(res => classes.value = res.data.data || [])
  } else if (tab === 'subjects') {
    loadSubjects()
  } else if (tab === 'class-subjects') {
    loadClassSubjects()
  } else if (tab === 'proctor') {
    if (targetScheduleId) {
      activeProctorScheduleId.value = targetScheduleId
    }
    // Polling data live dilakukan oleh LiveProctorControl (mengirim data lewat @data-refreshed).
    loadProctorSchedules()
  } else if (tab === 'questions') {
    loadAllData()
    if (readinessData.value.question_banks?.length > 0 && !selectedBankUploadId.value) {
      selectedBankUploadId.value = readinessData.value.question_banks[0].id
    }
  }
}


// Izin kelola jadwal (hanya kerapian UI; server tetap penjaga sebenarnya)
const canManageSchedules = computed(() => authStore.hasPermission('schedules:manage'))

// ---- Beranda: tampilan menurut izin (hanya kerapian UI; server tetap penjaga sebenarnya) ----
const canManageMaster = computed(() => authStore.hasPermission('master:manage'))
const canManageEvents = computed(() => authStore.hasPermission('events:manage'))
const canUploadQuestionBank = computed(() => authStore.hasPermission('questions:upload'))
// Daftar akun siswa (GET /admin/users) dan reset sesi (POST /admin/users/:id/reset-session) butuh izin ini
const canManageLoginSessions = computed(() => authStore.hasAnyPermission(['users:manage', 'master:manage']))
const canImportStudents = canManageLoginSessions
// Daftar siswa dan akun staf (GET /admin/students, /admin/teachers) hanya diberikan server kepada
// pengelola master, akun, atau jadwal; tanpa itu students kosong dan daftar hadir cetak tidak berguna.
const canReadPeople = computed(() => authStore.hasAnyPermission(['master:manage', 'users:manage', 'schedules:manage']))

// Staf tanpa izin kelola apa pun (guru/pengawas biasa) mendapat kartu sambutan sederhana.
// Dalam mode ini kartu Pintasan Cepat disembunyikan karena satu-satunya tile yang mungkin
// (Unggah Bank Soal) sudah tersedia lewat tombol "Bank Soal" pada kartu sambutan.
const showWelcomeCard = computed(() =>
  !(canManageMaster.value || canManageSchedules.value || canManageLoginSessions.value || canManageEvents.value)
)
const showQuickActions = computed(() =>
  !showWelcomeCard.value &&
  (canManageSchedules.value || canImportStudents.value || canUploadQuestionBank.value || canManageEvents.value)
)

const showEventReadiness = ref(true)

const eventReadinessStatus = (event) => {
  if (!event.total_schedules || event.total_schedules === 0) return 'none'
  const isLoaded =
    selectedExamEvent.value?.id === event.id ||
    (!selectedExamEvent.value && activeExamEvent.value?.id === event.id)
  if (isLoaded) {
    const allLinkedAndLocked =
      schedules.value.length > 0 &&
      schedules.value.every((s) => s.bank_id && s.bank && s.bank.is_locked)
    if (allLinkedAndLocked) return 'ready'
    return 'needs_action'
  }
  return 'running'
}

const eventStatusLabel = (event) => {
  const s = eventReadinessStatus(event)
  if (s === 'ready') return 'Siap'
  if (s === 'needs_action') return 'Perlu tindakan'
  if (s === 'running') return 'Berjalan'
  return 'Belum mulai'
}

const eventStatusBadgeClass = (event) => {
  const s = eventReadinessStatus(event)
  if (s === 'ready') return 'bg-emerald-100 text-emerald-700'
  if (s === 'needs_action') return 'bg-amber-100 text-amber-700'
  if (s === 'running') return 'bg-indigo-100 text-indigo-700'
  return 'bg-slate-100 text-slate-500'
}

const eventActionLabel = (event) => {
  const s = eventReadinessStatus(event)
  if (s === 'ready') return 'Live proctor →'
  if (s === 'needs_action') return 'Kunci soal →'
  if (s === 'running') return 'Kelola →'
  return 'Atur event →'
}

const eventActionBtnClass = (event) => {
  const s = eventReadinessStatus(event)
  if (s === 'ready') return 'bg-emerald-50 text-emerald-700 border border-emerald-200 hover:bg-emerald-100'
  if (s === 'needs_action') return 'bg-indigo-600 text-white hover:bg-indigo-700'
  return 'bg-slate-100 text-slate-700 hover:bg-slate-200 border border-slate-200'
}

const eventActionClick = (event) => {
  const s = eventReadinessStatus(event)
  enterEvent(event)
  if (s === 'ready') switchTab('proctor')
  else if (s === 'needs_action') switchTab('questions')
  else if (s === 'running') switchTab('schedules')
  else switchTab('events')
}

// Sesi login siswa aktif. Server menandai akun dengan has_active_session (token sesi terisi);
// token itu sendiri tidak pernah dikirim ke klien (json:"-"), jadi datanya diambil dari daftar akun.
const loginSessionAccounts = ref([])
const activeLoginSessions = computed(() => loginSessionAccounts.value.filter(u => u.has_active_session))
const loadLoginSessions = async () => {
  if (!canManageLoginSessions.value) {
    loginSessionAccounts.value = []
    return
  }
  try {
    const res = await api.get('/admin/users', { params: { role: 'SISWA' } })
    loginSessionAccounts.value = res.data.data || []
  } catch (e) {
    console.error('Failed to load student login sessions', e)
  }
}
// Segarkan setiap kali Beranda dibuka (pemuatan awal dilakukan di onMounted)
watch(
  () => activeTab.value === 'dashboard' && canManageLoginSessions.value,
  (visible) => { if (visible) loadLoginSessions() }
)

// Live Proctoring States
const proctorSchedules = ref([])
const activeProctorScheduleId = ref(null)
const proctorData = ref(null)
const isExportingNilai = ref(false)
const showProctorPrint = ref(false)
const isExportingPDF = ref(false)

const inProgressCount = computed(() => {
  if (!proctorData.value?.students) return 0
  return proctorData.value.students.filter(s => s.status === 'IN_PROGRESS').length
})

const getProctorStatusBadge = (status) => {
  switch (status) {
    case 'SUBMITTED':
      return 'bg-emerald-100 text-emerald-800'
    case 'IN_PROGRESS':
      return 'bg-blue-100 text-blue-800 animate-pulse'
    case 'BLOCKED':
      return 'bg-rose-100 text-rose-800 border border-rose-300'
    default:
      return 'bg-slate-100 text-slate-600'
  }
}

const getProctorStatusLabel = (status) => {
  switch (status) {
    case 'SUBMITTED':
      return 'Selesai'
    case 'IN_PROGRESS':
      return 'Mengerjakan'
    case 'BLOCKED':
      return 'Terkunci'
    default:
      return 'Belum Mulai'
  }
}

const loadProctorSchedules = async () => {
  try {
    const res = await api.get('/proctor/schedules')
    proctorSchedules.value = res.data.data || []
    if (proctorSchedules.value.length > 0 && !activeProctorScheduleId.value) {
      // Utamakan jadwal yang dapat dikendalikan pengguna (can_control tidak ada = dianggap true)
      const firstControllable = proctorSchedules.value.find(s => s.can_control !== false)
      activeProctorScheduleId.value = (firstControllable || proctorSchedules.value[0]).id
    }
  } catch (err) {
    console.error('Failed to load proctor schedules', err)
  }
}

// ---- Beranda guru / pengawas: ringkasan tugas milik pengguna yang sedang login ----
// Sumber data: jadwal (GET /admin/schedules, sudah dibatasi server), bank soal buatan sendiri
// (GET /admin/readiness-matrix), dan jadwal pengawasan (GET /proctor/schedules).
const canShowTeacherQuestionSection = computed(() => authStore.canAccessTab('questions'))
const canShowTeacherProctorSection = computed(() => authStore.canAccessTab('proctor'))
const canLockQuestionBanks = computed(() => authStore.hasPermission('questions:lock'))

const isTodayDate = (value) => !!value && new Date(value).toDateString() === new Date().toDateString()
const scheduleStartTimestamp = (sch) => {
  const t = sch?.start_time ? new Date(sch.start_time).getTime() : 0
  return Number.isNaN(t) ? 0 : t
}
const scheduleEndTimestamp = (sch) => {
  const t = new Date(sch?.end_time || sch?.start_time || 0).getTime()
  return Number.isNaN(t) ? 0 : t
}
const isScheduleNotFinished = (sch) => scheduleEndTimestamp(sch) >= Date.now()
const sortByScheduleStart = (list) => [...list].sort((a, b) => scheduleStartTimestamp(a) - scheduleStartTimestamp(b))

// Relasi pengguna dengan jadwal (GET /admin/schedules -> relations: "mengampu" | "bank_saya" | "pengawas").
// Backend lama belum mengirim properti relations sama sekali; dalam kasus itu semua jadwal yang diterima
// dianggap milik pengguna (perilaku lama) agar Beranda tidak menampilkan angka nol yang keliru.
const teacherHomeLegacyRelations = computed(() =>
  !schedules.value.some(s => s && Object.prototype.hasOwnProperty.call(s, 'relations'))
)
const teacherHomeRelationsOf = (sch) => (Array.isArray(sch?.relations) ? sch.relations : [])
const teacherHomeIsMine = (sch) => teacherHomeLegacyRelations.value || teacherHomeRelationsOf(sch).length > 0
const teacherHomeIsTeaching = (sch) => teacherHomeLegacyRelations.value || teacherHomeRelationsOf(sch).includes('mengampu')
const teacherHomeIsInactive = (sch) => sch?.is_active === false
// Peran sebagai teks netral; kosong pada backend lama (peran tidak diketahui)
const teacherHomeRoleLabel = (sch) => {
  const rel = teacherHomeRelationsOf(sch)
  const roles = []
  if (rel.includes('mengampu')) roles.push('Pengampu')
  if (rel.includes('pengawas')) roles.push('Pengawas')
  if (roles.length > 0) return roles.join(' · ')
  return rel.includes('bank_saya') ? 'Bank saya' : ''
}
// Jadwal yang menjadi tugas membuat bank soal: diampu pengguna, tanpa bank, aktif
const teacherHomeNeedsBank = (sch) =>
  teacherHomeIsTeaching(sch) && (!sch.bank_id || !sch.bank) && !teacherHomeIsInactive(sch)

// Jadwal milik pengguna (sudah difilter event bila ada event terpilih)
const teacherHomeSchedules = computed(() => currentEventSchedules.value.filter(teacherHomeIsMine))
const teacherHomeTodayScheduleCount = computed(() =>
  teacherHomeSchedules.value.filter(s => isTodayDate(s.start_time)).length
)
const teacherHomeUpcomingSchedules = computed(() =>
  sortByScheduleStart(teacherHomeSchedules.value.filter(isScheduleNotFinished))
)
const teacherHomeUpcomingList = computed(() => teacherHomeUpcomingSchedules.value.slice(0, 5))
const teacherHomeSchedulesWithoutBank = computed(() =>
  teacherHomeUpcomingSchedules.value.filter(teacherHomeNeedsBank)
)

// Bank soal buatan pengguna (server sudah membatasi; filter created_by_id sebagai pengaman bila pengguna berizin baca semua)
const teacherHomeBanks = computed(() => {
  const banks = readinessData.value.question_banks || []
  const myId = authStore.user?.id
  return myId ? banks.filter(b => !b.created_by_id || b.created_by_id === myId) : banks
})
const teacherHomeLockedBankCount = computed(() => teacherHomeBanks.value.filter(b => b.is_locked).length)
const teacherHomeDraftBankCount = computed(() => teacherHomeBanks.value.length - teacherHomeLockedBankCount.value)
const teacherHomeUnfinishedBanks = computed(() =>
  teacherHomeBanks.value.filter(b => !(b.total_questions > 0) || !b.is_locked)
)

// Daftar "Perlu Tindakan": jadwal tanpa bank soal + bank soal kosong/belum dikunci,
// diurutkan menurut jadwal terdekat yang memakainya (tanpa jadwal terkait di urutan akhir).
const TEACHER_HOME_ACTION_LIMIT = 6
const teacherHomeActionItems = computed(() => {
  if (!canShowTeacherQuestionSection.value) return []
  const upcoming = teacherHomeUpcomingSchedules.value
  const items = []

  teacherHomeSchedulesWithoutBank.value.forEach(sch => {
    items.push({
      key: `schedule-${sch.id}`,
      kind: 'no-bank',
      title: sch.title,
      label: 'Belum ada bank soal',
      detail: `${sch.class_room?.name || '-'} · ${formatScheduleTimeRange(sch.start_time, sch.end_time)}`,
      actionLabel: 'Buat Bank Soal',
      sortTime: scheduleStartTimestamp(sch)
    })
  })

  teacherHomeUnfinishedBanks.value.forEach(bank => {
    const linkedStarts = upcoming.filter(s => s.bank_id === bank.id).map(scheduleStartTimestamp)
    const isEmpty = !(bank.total_questions > 0)
    items.push({
      key: `bank-${bank.id}`,
      kind: isEmpty ? 'empty-bank' : 'unlocked-bank',
      bank,
      title: bank.title,
      label: isEmpty ? 'Belum ada butir soal' : 'Belum dikunci',
      detail: [bank.subject?.name, isEmpty ? '' : `${bank.total_questions} soal`].filter(Boolean).join(' · '),
      actionLabel: isEmpty ? 'Isi Soal' : (canLockQuestionBanks.value ? 'Kunci' : 'Lihat'),
      sortTime: linkedStarts.length > 0 ? Math.min(...linkedStarts) : Number.MAX_SAFE_INTEGER
    })
  })

  return items.sort((a, b) => a.sortTime - b.sortTime)
})
const teacherHomeVisibleActionItems = computed(() => teacherHomeActionItems.value.slice(0, TEACHER_HOME_ACTION_LIMIT))

const runTeacherHomeAction = (item) => {
  if (item.kind === 'unlocked-bank' && canLockQuestionBanks.value) {
    toggleBankLockWithConfirm(item.bank)
    return
  }
  switchTab('questions')
}

const teacherHomeScheduleBankStatus = (sch) => {
  if (!sch.bank_id || !sch.bank) {
    // Amber hanya bila memang tugas pengguna; jadwal pengawas saja / nonaktif tampil netral
    return { label: 'Belum ada bank soal', textClass: teacherHomeNeedsBank(sch) ? 'text-amber-600' : 'text-slate-500' }
  }
  if (sch.bank.is_locked) return { label: 'Siap', textClass: 'text-emerald-600' }
  return { label: 'Draf', textClass: 'text-slate-500' }
}

// Tugas mengawasi: hanya jadwal yang boleh dikendalikan (can_control), dibatasi event terpilih
const teacherHomeProctorTasks = computed(() => {
  if (!canShowTeacherProctorSection.value) return []
  const evId = currentScopedEventId.value
  return proctorSchedules.value.filter(s => s.can_control === true && (!evId || s.event_id === evId))
})
const teacherHomeProctorTodayCount = computed(() =>
  teacherHomeProctorTasks.value.filter(s => isTodayDate(s.start_time)).length
)
const teacherHomeProctorUpcomingList = computed(() =>
  sortByScheduleStart(teacherHomeProctorTasks.value.filter(isScheduleNotFinished)).slice(0, 5)
)

// Tata letak kartu angka: 1, 2, 4, atau 5 kartu tergantung izin
const teacherHomeStatCount = computed(() =>
  1 + (canShowTeacherQuestionSection.value ? 3 : 0) + (canShowTeacherProctorSection.value ? 1 : 0)
)
const teacherHomeGridClass = computed(() => {
  if (teacherHomeStatCount.value === 5) return 'grid-cols-2 xl:grid-cols-5'
  if (teacherHomeStatCount.value === 4) return 'grid-cols-2 sm:grid-cols-4'
  return 'grid-cols-2'
})
const teacherHomeProctorCardSpanClass = computed(() =>
  teacherHomeStatCount.value === 5 ? 'col-span-2 xl:col-span-1' : ''
)

// Jadwal pengawasan hanya dimuat untuk guru/pengawas yang membuka Beranda
const loadTeacherHomeProctorTasks = () => {
  if (showWelcomeCard.value && canShowTeacherProctorSection.value) loadProctorSchedules()
}
watch(
  () => activeTab.value === 'dashboard' && showWelcomeCard.value && canShowTeacherProctorSection.value,
  (visible) => { if (visible) loadProctorSchedules() }
)

// Galat dari request berjenis blob datang sebagai Blob; baca isinya dengan aman.
const extractExportErrorMessage = async (err, fallback) => {
  try {
    const data = err?.response?.data
    if (data instanceof Blob) {
      const parsed = JSON.parse(await data.text())
      return parsed?.message || parsed?.error || fallback
    }
    return data?.message || fallback
  } catch {
    return fallback
  }
}

const exportNilaiExcel = async () => {
  if (!activeProctorScheduleId.value || isExportingNilai.value) return
  isExportingNilai.value = true
  try {
    const res = await api.get(`/proctor/reports/excel/${activeProctorScheduleId.value}`, {
      responseType: 'blob'
    })
    const url = window.URL.createObjectURL(new Blob([res.data]))
    const link = document.createElement('a')
    link.href = url
    const className = proctorData.value?.class_name || 'Kelas'
    link.setAttribute('download', `Rekap_Nilai_${className}.xlsx`)
    document.body.appendChild(link)
    link.click()
    link.remove()
    window.URL.revokeObjectURL(url)
  } catch (err) {
    await showAlertModal({
      title: 'Gagal Mengunduh Laporan',
      message: await extractExportErrorMessage(err, 'Gagal mengunduh laporan'),
      type: 'danger'
    })
  } finally {
    isExportingNilai.value = false
  }
}

const exportBeritaAcaraPDF = async () => {
  if (!activeProctorScheduleId.value || isExportingPDF.value) return
  isExportingPDF.value = true
  try {
    const res = await api.get(`/proctor/reports/pdf/${activeProctorScheduleId.value}`, {
      responseType: 'blob'
    })
    const url = window.URL.createObjectURL(new Blob([res.data], { type: 'application/pdf' }))
    const link = document.createElement('a')
    link.href = url
    const className = proctorData.value?.class_name || 'Kelas'
    link.setAttribute('download', `Berita_Acara_${className}.pdf`)
    document.body.appendChild(link)
    link.click()
    link.remove()
    window.URL.revokeObjectURL(url)
  } catch (err) {
    await showAlertModal({
      title: 'Gagal Mengunduh Laporan',
      message: await extractExportErrorMessage(err, 'Gagal mengunduh laporan'),
      type: 'danger'
    })
  } finally {
    isExportingPDF.value = false
  }
}

// Bank Soal States & Logic (Unified Bank Soal & Kesiapan)
const questionBankFile = ref(null)
const selectedBankUploadId = ref('')
const isUploadingQuestionBank = ref(false)
const questionBankUploadStatus = ref(null)
const showUploadBankModal = ref(false)
const showCreateBankModal = ref(false)
const isCreatingBank = ref(false)
const isEditBank = ref(false)
const editingBankId = ref(null)
const questionBankSearchQuery = ref('')
const questionBankStatusFilter = ref('all') // 'all', 'locked', 'draft'
const newBankForm = ref({
  subject_id: '',
  title: ''
})

const availableBankSubjects = computed(() => {
  const scheds = currentEventSchedules.value || []
  const existingBanks = readinessData.value.question_banks || []
  const existingBankSubjectIds = new Set(
    existingBanks
      .filter(b => isEditBank.value ? b.id !== editingBankId.value : true)
      .map(b => b.subject_id || b.subject?.id)
      .filter(Boolean)
  )

  let list = []
  if (scheds.length > 0) {
    const map = new Map()
    scheds.forEach(s => {
      let foundSub = null
      const subId = s.subject_id || s.subject?.id || s.bank?.subject_id || s.bank?.subject?.id
      if (subId) {
        foundSub = s.subject || (s.bank?.subject) || subjects.value.find(sub => sub.id === subId)
      }
      if (!foundSub && s.title) {
        foundSub = subjects.value.find(sub => s.title.toLowerCase().includes(sub.name.toLowerCase()))
      }
      if (foundSub && !map.has(foundSub.id)) {
        map.set(foundSub.id, {
          id: foundSub.id,
          name: foundSub.name,
          code: foundSub.code
        })
      }
    })
    // In edit mode, ensure the currently edited bank's subject is in the list
    if (isEditBank.value && editingBankId.value) {
      const currentBank = existingBanks.find(b => b.id === editingBankId.value)
      if (currentBank?.subject && !map.has(currentBank.subject.id)) {
        map.set(currentBank.subject.id, {
          id: currentBank.subject.id,
          name: currentBank.subject.name,
          code: currentBank.subject.code
        })
      }
    }
    list = Array.from(map.values())
  } else {
    list = subjects.value || []
  }

  return list
    .filter(sub => !existingBankSubjectIds.has(sub.id))
    .sort((a, b) => a.name.localeCompare(b.name))
})

const onBankSubjectChange = () => {
  const selSubj = availableBankSubjects.value.find(s => s.id === newBankForm.value.subject_id) || subjects.value.find(s => s.id === newBankForm.value.subject_id)
  if (selSubj) {
    const evTitle = selectedExamEvent.value?.title ? `${selectedExamEvent.value.title} - ` : ''
    newBankForm.value.title = `${evTitle}${selSubj.name}`
  }
}

const questionBankSortKey = ref('title') // 'title', 'creator', 'questions', 'status'
const questionBankSortOrder = ref('asc') // 'asc' | 'desc'
const questionBankCurrentPage = ref(1)
const questionBankPerPage = ref(10)

const toggleQuestionBankSort = (key) => {
  if (questionBankSortKey.value === key) {
    questionBankSortOrder.value = questionBankSortOrder.value === 'asc' ? 'desc' : 'asc'
  } else {
    questionBankSortKey.value = key
    questionBankSortOrder.value = 'asc'
  }
}

const filteredQuestionBanks = computed(() => {
  let list = readinessData.value.question_banks || []
  if (questionBankStatusFilter.value === 'locked') {
    list = list.filter(b => b.is_locked)
  } else if (questionBankStatusFilter.value === 'draft') {
    list = list.filter(b => !b.is_locked)
  }

  if (questionBankSearchQuery.value.trim()) {
    const q = questionBankSearchQuery.value.toLowerCase().trim()
    list = list.filter(b => {
      const title = (b.title || '').toLowerCase()
      const subjectName = (b.subject?.name || '').toLowerCase()
      const subjectCode = (b.subject?.code || '').toLowerCase()
      const creatorName = (b.created_by?.full_name || '').toLowerCase()
      return title.includes(q) || subjectName.includes(q) || subjectCode.includes(q) || creatorName.includes(q)
    })
  }

  // Sort
  if (questionBankSortKey.value) {
    list = [...list].sort((a, b) => {
      let result = 0
      if (questionBankSortKey.value === 'title') {
        const titleA = (a.title || '').trim().toLowerCase()
        const titleB = (b.title || '').trim().toLowerCase()
        result = titleA.localeCompare(titleB, undefined, { numeric: true, sensitivity: 'base' })
        if (result === 0) {
          const subA = (a.subject?.name || '').trim().toLowerCase()
          const subB = (b.subject?.name || '').trim().toLowerCase()
          result = subA.localeCompare(subB, undefined, { numeric: true, sensitivity: 'base' })
        }
      } else if (questionBankSortKey.value === 'creator') {
        const nameA = (a.created_by?.full_name || '').trim().toLowerCase()
        const nameB = (b.created_by?.full_name || '').trim().toLowerCase()
        result = nameA.localeCompare(nameB, undefined, { numeric: true, sensitivity: 'base' })
      } else if (questionBankSortKey.value === 'questions') {
        const countA = a.total_questions || 0
        const countB = b.total_questions || 0
        result = countA - countB
      } else if (questionBankSortKey.value === 'status') {
        const lockA = a.is_locked ? 1 : 0
        const lockB = b.is_locked ? 1 : 0
        result = lockA - lockB
      }
      return questionBankSortOrder.value === 'asc' ? result : -result
    })
  }

  return list
})

const totalQuestionBankPages = computed(() => {
  return Math.max(1, Math.ceil(filteredQuestionBanks.value.length / questionBankPerPage.value))
})

const paginatedQuestionBanks = computed(() => {
  const start = (questionBankCurrentPage.value - 1) * questionBankPerPage.value
  return filteredQuestionBanks.value.slice(start, start + questionBankPerPage.value)
})

const displayedQuestionBankPages = computed(() => {
  const total = totalQuestionBankPages.value
  const current = questionBankCurrentPage.value
  if (total <= 7) {
    return Array.from({ length: total }, (_, i) => i + 1)
  }
  if (current <= 4) {
    return [1, 2, 3, 4, 5, '...', total]
  } else if (current >= total - 3) {
    return [1, '...', total - 4, total - 3, total - 2, total - 1, total]
  } else {
    return [1, '...', current - 1, current, current + 1, '...', total]
  }
})

const setQuestionBankPage = (p) => {
  if (p === '...' || p < 1 || p > totalQuestionBankPages.value) return
  questionBankCurrentPage.value = p
}

const downloadQuestionBankTemplate = async () => {
  try {
    const res = await api.get('/admin/template/question-bank.xlsx', {
      responseType: 'blob'
    })
    const url = window.URL.createObjectURL(new Blob([res.data], {
      type: 'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet'
    }))
    const link = document.createElement('a')
    link.href = url
    link.setAttribute('download', 'Template_Bank_Soal_CBT.xlsx')
    document.body.appendChild(link)
    link.click()
    link.remove()
    window.URL.revokeObjectURL(url)
    showToast('Format template Excel berhasil diunduh!', 'success')
  } catch (err) {
    await showAlertModal({
      title: 'Gagal Mengunduh Template',
      message: await extractExportErrorMessage(err, 'Gagal mengunduh template'),
      type: 'danger'
    })
  }
}

const handleQuestionBankFileChange = (e) => {
  questionBankFile.value = e.target.files[0]
}

const openUploadBankModal = (bank = null) => {
  if (bank) {
    selectedBankUploadId.value = bank.id
  } else if (!selectedBankUploadId.value && (readinessData.value.question_banks || []).length > 0) {
    selectedBankUploadId.value = readinessData.value.question_banks[0].id
  }
  questionBankFile.value = null
  questionBankUploadStatus.value = null
  showUploadBankModal.value = true
}

const openCreateBankModal = () => {
  isEditBank.value = false
  editingBankId.value = null
  const defaultSubj = availableBankSubjects.value[0]
  const defaultSubjId = defaultSubj?.id || subjects.value[0]?.id || ''
  const evTitle = selectedExamEvent.value?.title ? `${selectedExamEvent.value.title} - ` : ''
  newBankForm.value = {
    subject_id: defaultSubjId,
    title: defaultSubj ? `${evTitle}${defaultSubj.name}` : ''
  }
  showCreateBankModal.value = true
}

const openEditBankModal = (bank) => {
  isEditBank.value = true
  editingBankId.value = bank.id
  newBankForm.value = {
    subject_id: bank.subject_id || bank.subject?.id || '',
    title: bank.title || ''
  }
  showCreateBankModal.value = true
}

const submitCreateBank = async () => {
  if (!newBankForm.value.subject_id || !newBankForm.value.title.trim()) {
    await showAlertModal({
      title: 'Form Belum Lengkap',
      message: 'Mata pelajaran dan judul bank soal wajib diisi.',
      type: 'warning'
    })
    return
  }

  isCreatingBank.value = true
  try {
    if (isEditBank.value && editingBankId.value) {
      await api.put(`/admin/question-banks/${editingBankId.value}`, {
        subject_id: newBankForm.value.subject_id,
        title: newBankForm.value.title.trim()
      })
      showToast('Bank soal berhasil diperbarui!', 'success')
    } else {
      const res = await api.post('/admin/question-banks', {
        subject_id: newBankForm.value.subject_id,
        title: newBankForm.value.title.trim()
      })
      showToast('Bank soal baru berhasil dibuat!', 'success')
      if (res.data.data?.id) {
        selectedBankUploadId.value = res.data.data.id
      }
    }
    showCreateBankModal.value = false
    await loadAllData()
  } catch (err) {
    await showAlertModal({
      title: isEditBank.value ? 'Gagal Memperbarui Bank Soal' : 'Gagal Membuat Bank Soal',
      message: err.response?.data?.message || 'Terjadi kesalahan saat menyimpan bank soal.',
      type: 'danger'
    })
  } finally {
    isCreatingBank.value = false
  }
}

const deleteQuestionBank = async (bank) => {
  const confirmed = await showConfirmModal({
    title: 'Hapus Bank Soal',
    message: `Apakah Anda yakin ingin menghapus bank soal "${bank.title}" (${bank.subject?.name || ''}) beserta seluruh butir soalnya?`,
    type: 'danger',
    confirmText: 'Hapus Bank Soal',
    cancelText: 'Batal'
  })
  if (!confirmed) return

  try {
    await api.delete(`/admin/question-banks/${bank.id}`)
    showToast('Bank soal berhasil dihapus!', 'success')
    await loadAllData()
  } catch (err) {
    await showAlertModal({
      title: 'Gagal Menghapus Bank Soal',
      message: err.response?.data?.message || 'Terjadi kesalahan saat menghapus bank soal.',
      type: 'danger'
    })
  }
}

const toggleBankLockWithConfirm = async (bank) => {
  const willLock = !bank.is_locked
  const actionTitle = willLock ? 'Kunci Naskah Bank Soal' : 'Buka Kunci Naskah Bank Soal'
  const actionMessage = willLock
    ? `Kunci naskah "${bank.title}"? Setelah dikunci, naskah dinyatakan Siap Ujian dan butir soal siap disajikan kepada peserta ujian.`
    : `Buka kunci naskah "${bank.title}"? Status naskah akan kembali menjadi Draft Terbuka.`

  const confirmed = await showConfirmModal({
    title: actionTitle,
    message: actionMessage,
    type: willLock ? 'confirm' : 'warning',
    confirmText: willLock ? 'Kunci Naskah' : 'Buka Kunci',
    cancelText: 'Batal'
  })
  if (!confirmed) return

  try {
    await api.post(`/admin/question-banks/${bank.id}/toggle-lock`)
    showToast(willLock ? 'Naskah soal berhasil dikunci (Terkunci)!' : 'Kunci naskah soal berhasil dibuka (Draft)!', 'success')
    await loadAllData()
  } catch (err) {
    await showAlertModal({
      title: 'Gagal Mengubah Status Kunci',
      message: err.response?.data?.message || 'Terjadi kesalahan saat mengubah status penguncian bank soal.',
      type: 'danger'
    })
  }
}

const submitQuestionBankUpload = async () => {
  if (!selectedBankUploadId.value) {
    await showAlertModal({
      title: 'Pilih Bank Soal',
      message: 'Pilih bank soal tujuan terlebih dahulu.',
      type: 'warning'
    })
    return
  }
  if (!questionBankFile.value) {
    await showAlertModal({
      title: 'Pilih File Soal',
      message: 'Pilih file Excel (.xlsx) terlebih dahulu.',
      type: 'warning'
    })
    return
  }
  isUploadingQuestionBank.value = true
  questionBankUploadStatus.value = null
  const formData = new FormData()
  formData.append('bank_id', selectedBankUploadId.value)
  formData.append('file', questionBankFile.value)
  try {
    const res = await api.post('/admin/questions/import', formData, {
      headers: { 'Content-Type': 'multipart/form-data' }
    })
    questionBankUploadStatus.value = {
      success: true,
      message: res.data.message || 'Soal berhasil diimpor ke bank soal!'
    }
    showToast('Soal berhasil diimpor ke bank soal!', 'success')
    questionBankFile.value = null
    showUploadBankModal.value = false
    await loadAllData()
  } catch (err) {
    questionBankUploadStatus.value = {
      success: false,
      message: err.response?.data?.message || 'Gagal mengimpor file Excel bank soal.'
    }
    await showAlertModal({
      title: 'Gagal Impor Soal',
      message: err.response?.data?.message || 'Gagal mengimpor file Excel bank soal.',
      type: 'danger'
    })
  } finally {
    isUploadingQuestionBank.value = false
  }
}

// ---------------- QUESTION VIEWER & EDITOR ----------------
const showBankQuestionsModal = ref(false)
const showSimulatorModal = ref(false)
const viewingBank = ref(null)
const bankQuestions = ref([])
const isLoadingBankQuestions = ref(false)
const editingQuestionId = ref(null) // null = none, 'new' = add new question, uuid = edit question
const isSavingQuestion = ref(false)
const isUploadingQuestionImage = ref(false)
const uploadingOptionKey = ref(null)
const draggedQuestionIdx = ref(null)
const dragOverQuestionIdx = ref(null)
const isReorderingQuestions = ref(false)
const reorderStatusMessage = ref('')

const previewBankDirectly = async (bank) => {
  viewingBank.value = bank
  await fetchBankQuestions(bank.id)
  showSimulatorModal.value = true
}

const handleQuestionImageUpload = async (event) => {
  const file = event.target?.files?.[0]
  if (!file) return
  await uploadAndInsertQuestionImage(file)
  if (event.target) event.target.value = ''
}

const uploadAndInsertQuestionImage = async (file) => {
  if (!file.type.startsWith('image/')) {
    showToast('File harus berupa gambar (JPG, PNG, WEBP, GIF, SVG)', 'warning')
    return
  }
  if (file.size > 5 * 1024 * 1024) {
    showToast('Ukuran gambar maksimal 5 MB', 'warning')
    return
  }

  isUploadingQuestionImage.value = true
  const formData = new FormData()
  formData.append('image', file)

  try {
    const res = await api.post('/admin/upload-image', formData, {
      headers: { 'Content-Type': 'multipart/form-data' }
    })
    const imgUrl = res.data.url
    const imgTag = `\n<img src="${imgUrl}" alt="Gambar Soal" />\n`
    questionForm.value.content_html = (questionForm.value.content_html || '') + imgTag
    showToast('Gambar pertanyaan berhasil diunggah!', 'success')
  } catch (err) {
    showToast('Gagal mengunggah gambar: ' + (err.response?.data?.message || err.message), 'danger')
  } finally {
    isUploadingQuestionImage.value = false
  }
}

const handleQuestionPaste = async (event) => {
  const items = (event.clipboardData || event.originalEvent?.clipboardData)?.items
  if (!items) return

  for (const item of items) {
    if (item.type.indexOf('image') !== -1) {
      event.preventDefault()
      const file = item.getAsFile()
      if (file) {
        await uploadAndInsertQuestionImage(file)
      }
      break
    }
  }
}

const handleQuestionDrop = async (event) => {
  const files = event.dataTransfer?.files
  if (files && files.length > 0) {
    const file = files[0]
    if (file.type.startsWith('image/')) {
      await uploadAndInsertQuestionImage(file)
    }
  }
}

const handleOptionImageUpload = async (event, optKey) => {
  const file = event.target?.files?.[0]
  if (!file) return
  if (!file.type.startsWith('image/')) {
    showToast('File harus berupa gambar (JPG, PNG, WEBP, GIF, SVG)', 'warning')
    return
  }
  if (file.size > 5 * 1024 * 1024) {
    showToast('Ukuran gambar maksimal 5 MB', 'warning')
    return
  }

  uploadingOptionKey.value = optKey
  const formData = new FormData()
  formData.append('image', file)

  try {
    const res = await api.post('/admin/upload-image', formData, {
      headers: { 'Content-Type': 'multipart/form-data' }
    })
    const opt = questionForm.value.options.find(o => o.key === optKey)
    if (opt) {
      opt.image_url = res.data.url
    }
    showToast(`Gambar pilihan ${optKey} berhasil diunggah!`, 'success')
  } catch (err) {
    showToast('Gagal mengunggah gambar opsi: ' + (err.response?.data?.message || err.message), 'danger')
  } finally {
    uploadingOptionKey.value = null
    if (event.target) event.target.value = ''
  }
}

const removeOptionImage = (optKey) => {
  const opt = questionForm.value.options.find(o => o.key === optKey)
  if (opt) {
    opt.image_url = ''
  }
}

const insertMathSnippet = (snippet) => {
  questionForm.value.content_html = (questionForm.value.content_html || '') + ' ' + snippet + ' '
}

const onQuestionDragStart = (event, index) => {
  draggedQuestionIdx.value = index
  if (event.dataTransfer) {
    event.dataTransfer.effectAllowed = 'move'
    event.dataTransfer.setData('text/plain', index.toString())
  }
}

const onQuestionDragOver = (event, index) => {
  event.preventDefault()
  if (event.dataTransfer) {
    event.dataTransfer.dropEffect = 'move'
  }
  if (dragOverQuestionIdx.value !== index) {
    dragOverQuestionIdx.value = index
  }
}

const onQuestionDragLeave = (event, index) => {
  if (dragOverQuestionIdx.value === index) {
    dragOverQuestionIdx.value = null
  }
}

const onQuestionDrop = async (event, targetIndex) => {
  event.preventDefault()
  const fromIndex = draggedQuestionIdx.value
  dragOverQuestionIdx.value = null
  draggedQuestionIdx.value = null

  if (fromIndex === null || fromIndex === targetIndex || fromIndex === undefined || targetIndex === undefined) {
    return
  }

  // Local reorder
  const movedItem = bankQuestions.value.splice(fromIndex, 1)[0]
  bankQuestions.value.splice(targetIndex, 0, movedItem)

  // Update question_number locally
  bankQuestions.value.forEach((q, idx) => {
    q.question_number = idx + 1
  })

  // Sync to backend
  const bankId = viewingBank.value?.id
  if (!bankId) return

  isReorderingQuestions.value = true
  reorderStatusMessage.value = 'Menyimpan urutan...'
  try {
    const questionIds = bankQuestions.value.map(q => q.id)
    await api.put(`/admin/question-banks/${bankId}/reorder`, {
      question_ids: questionIds
    })
    reorderStatusMessage.value = '✓ Urutan tersimpan'
    showToast('Urutan nomor soal berhasil diperbarui!', 'success')
    setTimeout(() => {
      if (reorderStatusMessage.value === '✓ Urutan tersimpan') {
        reorderStatusMessage.value = ''
      }
    }, 2500)
  } catch (err) {
    reorderStatusMessage.value = 'Gagal menyimpan urutan'
    showToast('Gagal menyimpan urutan soal: ' + (err.response?.data?.message || err.message), 'danger')
  } finally {
    isReorderingQuestions.value = false
  }
}

const onQuestionDragEnd = () => {
  draggedQuestionIdx.value = null
  dragOverQuestionIdx.value = null
}

const questionForm = ref({
  question_number: 1,
  type: 'MULTIPLE_CHOICE',
  content_html: '',
  options: [
    { key: 'A', text: '', image_url: '' },
    { key: 'B', text: '', image_url: '' },
    { key: 'C', text: '', image_url: '' },
    { key: 'D', text: '', image_url: '' },
    { key: 'E', text: '', image_url: '' }
  ],
  correct_key: 'A',
  rubric_guide: '',
  score_weight: 1.0
})

const getQuestionTypeLabel = (type) => {
  switch (type) {
    case 'SHORT_ANSWER':
      return 'Jawaban Singkat'
    case 'ESSAY':
      return 'Essay / Uraian'
    default:
      return 'Pilihan Ganda'
  }
}

const getQuestionTypeBadgeClass = (type) => {
  switch (type) {
    case 'SHORT_ANSWER':
      return 'bg-emerald-50 border-emerald-200 text-emerald-700'
    case 'ESSAY':
      return 'bg-purple-50 border-purple-200 text-purple-700'
    default:
      return 'bg-blue-50 border-blue-200 text-blue-700'
  }
}

const openBankQuestionsModal = async (bank) => {
  viewingBank.value = bank
  editingQuestionId.value = null
  showBankQuestionsModal.value = true
  await fetchBankQuestions(bank.id)
}

const fetchBankQuestions = async (bankId = null) => {
  const id = bankId || viewingBank.value?.id
  if (!id) return
  isLoadingBankQuestions.value = true
  try {
    const res = await api.get(`/admin/question-banks/${id}/questions`)
    bankQuestions.value = res.data.data || []
    if (res.data.bank) {
      viewingBank.value = res.data.bank
    }
  } catch (err) {
    showToast('Gagal memuat daftar butir soal: ' + (err.response?.data?.message || err.message), 'danger')
  } finally {
    isLoadingBankQuestions.value = false
  }
}

const startAddQuestion = () => {
  editingQuestionId.value = 'new'
  questionForm.value = {
    question_number: (bankQuestions.value.length || 0) + 1,
    type: 'MULTIPLE_CHOICE',
    content_html: '',
    options: [
      { key: 'A', text: '', image_url: '' },
      { key: 'B', text: '', image_url: '' },
      { key: 'C', text: '', image_url: '' },
      { key: 'D', text: '', image_url: '' },
      { key: 'E', text: '', image_url: '' }
    ],
    correct_key: 'A',
    rubric_guide: '',
    score_weight: 1.0
  }
}

const startEditQuestion = (q) => {
  editingQuestionId.value = q.id
  let parsedOpts = q.options || []
  if (typeof parsedOpts === 'string') {
    try {
      parsedOpts = JSON.parse(parsedOpts)
    } catch {
      parsedOpts = []
    }
  }
  const opts = ['A', 'B', 'C', 'D', 'E'].map(k => {
    const existing = parsedOpts.find(o => o.key === k)
    return {
      key: k,
      text: existing ? existing.text : '',
      image_url: existing ? (existing.image_url || '') : ''
    }
  })
  const qType = q.type || 'MULTIPLE_CHOICE'
  questionForm.value = {
    question_number: q.question_number,
    type: qType,
    content_html: q.content_html,
    options: opts,
    correct_key: q.correct_key || (qType === 'SHORT_ANSWER' ? '' : 'A'),
    rubric_guide: q.rubric_guide || '',
    score_weight: q.score_weight || 1.0
  }
}

const cancelEditQuestion = () => {
  editingQuestionId.value = null
}

const saveQuestion = async () => {
  if (!questionForm.value.content_html.trim()) {
    await showAlertModal({
      title: 'Form Belum Lengkap',
      message: 'Konten / naskah soal wajib diisi.',
      type: 'warning'
    })
    return
  }

  const qType = questionForm.value.type || 'MULTIPLE_CHOICE'

  if (qType === 'MULTIPLE_CHOICE') {
    const filledOptions = questionForm.value.options.filter(o => (o.text && o.text.trim()) || (o.image_url && o.image_url.trim()))
    if (filledOptions.length < 2) {
      await showAlertModal({
        title: 'Pilihan Jawaban Kurang',
        message: 'Harap isi minimal 2 pilihan jawaban (misal A dan B teks atau gambar).',
        type: 'warning'
      })
      return
    }
  } else if (qType === 'SHORT_ANSWER') {
    if (!questionForm.value.correct_key.trim()) {
      await showAlertModal({
        title: 'Kunci Jawaban Kosong',
        message: 'Kunci jawaban singkat wajib diisi.',
        type: 'warning'
      })
      return
    }
  }

  isSavingQuestion.value = true
  try {
    const payload = {
      type: qType,
      content_html: questionForm.value.content_html.trim(),
      options: qType === 'MULTIPLE_CHOICE'
        ? questionForm.value.options.filter(o => (o.text && o.text.trim()) || (o.image_url && o.image_url.trim()))
        : [],
      correct_key: questionForm.value.correct_key.trim(),
      rubric_guide: questionForm.value.rubric_guide.trim(),
      score_weight: Number(questionForm.value.score_weight) || 1.0
    }

    if (editingQuestionId.value === 'new') {
      await api.post(`/admin/question-banks/${viewingBank.value.id}/questions`, payload)
      showToast('Butir soal baru berhasil ditambahkan!', 'success')
    } else {
      await api.put(`/admin/questions/${editingQuestionId.value}`, payload)
      showToast('Butir soal berhasil diperbarui!', 'success')
    }

    editingQuestionId.value = null
    await fetchBankQuestions()
    await loadAllData()
  } catch (err) {
    await showAlertModal({
      title: 'Gagal Menyimpan Soal',
      message: err.response?.data?.message || 'Terjadi kesalahan saat menyimpan soal.',
      type: 'danger'
    })
  } finally {
    isSavingQuestion.value = false
  }
}

const deleteQuestionItem = async (q) => {
  const confirmed = await showConfirmModal({
    title: 'Hapus Butir Soal',
    message: `Apakah Anda yakin ingin menghapus Soal #${q.question_number}? Nomor urut soal lainnya akan disesuaikan otomatis.`,
    type: 'danger',
    confirmText: 'Hapus Soal',
    cancelText: 'Batal'
  })
  if (!confirmed) return

  try {
    await api.delete(`/admin/questions/${q.id}`)
    showToast(`Soal #${q.question_number} berhasil dihapus!`, 'success')
    await fetchBankQuestions()
    await loadAllData()
  } catch (err) {
    await showAlertModal({
      title: 'Gagal Menghapus Soal',
      message: err.response?.data?.message || 'Terjadi kesalahan saat menghapus soal.',
      type: 'danger'
    })
  }
}



const loadAllData = async () => {
  try {
    const evId = currentScopedEventId.value
    const schedParams = evId ? { event_id: evId } : {}
    // Daftar siswa dan akun staf hanya diberikan server kepada pengelola master, akun, atau jadwal
    const emptyList = { data: { data: [] } }
    const [statsRes, schedRes, stRes, tcRes, clRes, rdnRes, evRes, subRes, csRes] = await Promise.all([
      api.get('/admin/dashboard-stats'),
      api.get('/admin/schedules', { params: schedParams }),
      canReadPeople.value ? api.get('/admin/students') : emptyList,
      canReadPeople.value ? api.get('/admin/teachers') : emptyList,
      api.get('/admin/classes'),
      api.get('/admin/readiness-matrix'),
      api.get('/admin/events'),
      api.get('/admin/subjects'),
      api.get('/admin/class-subjects'),
    ])
    stats.value = statsRes.data.data || {}
    schedules.value = schedRes.data.data || []
    students.value = stRes.data.data || []
    teachers.value = tcRes.data.data || []
    classes.value = clRes.data.data || []
    readinessData.value = rdnRes.data.data || {}
    events.value = evRes.data.data || []
    subjects.value = subRes.data.data || []
    classSubjects.value = csRes.data.data || []

    if (selectedEventId.value) {
      const exists = events.value.some(e => e.id === selectedEventId.value)
      if (!exists) {
        selectedEventId.value = ''
        selectedEventFilter.value = ''
        sessionStorage.removeItem('cbt_selected_event_id')
      }
    }

    if (classes.value.length > 0 && !newStudent.value.class_id) {
      newStudent.value.class_id = classes.value[0].id
    }
  } catch (e) {
    console.error('Failed to load admin dashboard data', e)
  }
}

const resetStudentSession = async (st) => {
  const confirmed = await showConfirmModal({
    title: 'Lepas Kunci Sesi Login',
    message: `Lepas kunci sesi login untuk ${st.user?.full_name}? Siswa dapat segera login kembali di perangkat baru.`,
    type: 'warning',
    confirmText: 'Lepas Kunci',
    cancelText: 'Batal'
  })
  if (!confirmed) return

  try {
    await api.post(`/admin/users/${st.user_id}/reset-session`)
    showToast(`✓ Kunci sesi login HP ${st.user?.full_name} berhasil dilepas!`, 'success')
    const res = await api.get('/admin/students')
    students.value = res.data.data || []
  } catch (e) {
    await showAlertModal({
      title: 'Gagal Melepas Kunci Sesi',
      message: e.response?.data?.message || 'Terjadi kesalahan saat melepas kunci sesi.',
      type: 'danger'
    })
  }
}

const loadSubjects = async () => {
  try {
    const res = await api.get('/admin/subjects')
    subjects.value = res.data.data || []
  } catch (e) {
    console.error('Failed to load subjects', e)
  }
}

const openCreateSubject = () => {
  isEditSubject.value = false
  subjectForm.value = { id: null, code: '', name: '' }
  showSubjectModal.value = true
}

const openEditSubject = (s) => {
  isEditSubject.value = true
  subjectForm.value = { id: s.id, code: s.code, name: s.name }
  showSubjectModal.value = true
}

const submitSubjectForm = async () => {
  if (!subjectForm.value.code || !subjectForm.value.name) {
    await showAlertModal({
      title: 'Form Belum Lengkap',
      message: 'Kode dan Nama Mata Pelajaran wajib diisi.',
      type: 'warning'
    })
    return
  }
  try {
    if (isEditSubject.value) {
      await api.put(`/admin/subjects/${subjectForm.value.id}`, subjectForm.value)
      showToast('Mata pelajaran berhasil diperbarui!', 'success')
    } else {
      await api.post('/admin/subjects', subjectForm.value)
      showToast('Mata pelajaran berhasil ditambahkan!', 'success')
    }
    showSubjectModal.value = false
    await loadSubjects()
    await loadAllData()
  } catch (e) {
    await showAlertModal({
      title: 'Gagal Menyimpan Mapel',
      message: e.response?.data?.message || 'Terjadi kesalahan saat menyimpan mata pelajaran.',
      type: 'danger'
    })
  }
}

const deleteSubject = async (s) => {
  const confirmed = await showConfirmModal({
    title: 'Hapus Mata Pelajaran',
    message: `Yakin ingin menghapus mata pelajaran ${s.name} (${s.code})?`,
    type: 'danger',
    confirmText: 'Hapus',
    cancelText: 'Batal'
  })
  if (!confirmed) return

  try {
    await api.delete(`/admin/subjects/${s.id}`)
    await loadSubjects()
    await loadAllData()
    showToast('Mata pelajaran berhasil dihapus', 'success')
  } catch (e) {
    await showAlertModal({
      title: 'Gagal Menghapus Mapel',
      message: e.response?.data?.message || 'Terjadi kesalahan saat menghapus mata pelajaran.',
      type: 'danger'
    })
  }
}

// Class Subjects Search, Filter, Sort, Pagination
const classSubjectSearchQuery = ref('')
const selectedClassSubjectGrade = ref('all')
const selectedClassSubjectClassId = ref('')
const classSubjectSortKey = ref('class')
const classSubjectSortOrder = ref('asc')
const classSubjectCurrentPage = ref(1)
const classSubjectPerPage = ref(10)

const toggleClassSubjectSort = (key) => {
  if (classSubjectSortKey.value === key) {
    classSubjectSortOrder.value = classSubjectSortOrder.value === 'asc' ? 'desc' : 'asc'
  } else {
    classSubjectSortKey.value = key
    classSubjectSortOrder.value = 'asc'
  }
}

const filteredClassSubjects = computed(() => {
  let list = classSubjects.value || []

  // Filter by Grade
  if (selectedClassSubjectGrade.value !== 'all') {
    list = list.filter(cs => {
      const cls = cs.class_room || classes.value.find(c => c.id === cs.class_id)
      return cls && cls.grade === selectedClassSubjectGrade.value
    })
  }

  // Filter by Specific Class ID
  if (selectedClassSubjectClassId.value) {
    list = list.filter(cs => cs.class_id === selectedClassSubjectClassId.value || cs.class_room?.id === selectedClassSubjectClassId.value)
  }

  // Filter by Search Query
  if (classSubjectSearchQuery.value.trim()) {
    const q = classSubjectSearchQuery.value.toLowerCase().trim()
    list = list.filter(cs => {
      const clsName = (cs.class_room?.name || '').toLowerCase()
      const clsGrade = (cs.class_room?.grade || '').toLowerCase()
      const subName = (cs.subject?.name || '').toLowerCase()
      const subCode = (cs.subject?.code || '').toLowerCase()
      const teacherName = (cs.teacher?.full_name || '').toLowerCase()
      const teacherUser = (cs.teacher?.username || '').toLowerCase()
      const academicYear = (cs.academic_year || '').toLowerCase()
      return clsName.includes(q) || clsGrade.includes(q) || subName.includes(q) || subCode.includes(q) || teacherName.includes(q) || teacherUser.includes(q) || academicYear.includes(q)
    })
  }

  // Sorting
  if (classSubjectSortKey.value) {
    list = [...list].sort((a, b) => {
      let result = 0
      if (classSubjectSortKey.value === 'class') {
        const nameA = (a.class_room?.name || '').trim().toLowerCase()
        const nameB = (b.class_room?.name || '').trim().toLowerCase()
        result = nameA.localeCompare(nameB, undefined, { numeric: true, sensitivity: 'base' })
      } else if (classSubjectSortKey.value === 'subject') {
        const nameA = (a.subject?.name || '').trim().toLowerCase()
        const nameB = (b.subject?.name || '').trim().toLowerCase()
        result = nameA.localeCompare(nameB, undefined, { numeric: true, sensitivity: 'base' })
      } else if (classSubjectSortKey.value === 'teacher') {
        const nameA = (a.teacher?.full_name || '').trim().toLowerCase()
        const nameB = (b.teacher?.full_name || '').trim().toLowerCase()
        result = nameA.localeCompare(nameB, undefined, { numeric: true, sensitivity: 'base' })
      } else if (classSubjectSortKey.value === 'year') {
        const valA = (a.academic_year || '').trim().toLowerCase()
        const valB = (b.academic_year || '').trim().toLowerCase()
        result = valA.localeCompare(valB, undefined, { numeric: true, sensitivity: 'base' })
      }
      return classSubjectSortOrder.value === 'asc' ? result : -result
    })
  }

  return list
})

const totalClassSubjectPages = computed(() => {
  return Math.max(1, Math.ceil(filteredClassSubjects.value.length / classSubjectPerPage.value))
})

const paginatedClassSubjects = computed(() => {
  const start = (classSubjectCurrentPage.value - 1) * classSubjectPerPage.value
  return filteredClassSubjects.value.slice(start, start + classSubjectPerPage.value)
})

const displayedClassSubjectPages = computed(() => {
  const total = totalClassSubjectPages.value
  const current = classSubjectCurrentPage.value
  if (total <= 7) {
    return Array.from({ length: total }, (_, i) => i + 1)
  }
  if (current <= 4) {
    return [1, 2, 3, 4, 5, '...', total]
  } else if (current >= total - 3) {
    return [1, '...', total - 4, total - 3, total - 2, total - 1, total]
  } else {
    return [1, '...', current - 1, current, current + 1, '...', total]
  }
})

const setClassSubjectPage = (p) => {
  if (p === '...' || p < 1 || p > totalClassSubjectPages.value) return
  classSubjectCurrentPage.value = p
}

const loadClassSubjects = async () => {
  try {
    const res = await api.get('/admin/class-subjects')
    classSubjects.value = res.data.data || []
  } catch (e) {
    console.error('Failed to load class subjects', e)
  }
}

const isEditClassSubject = ref(false)

const openCreateClassSubject = () => {
  isEditClassSubject.value = false
  classSubjectForm.value = {
    id: null,
    class_id: classes.value[0]?.id || '',
    subject_id: subjects.value[0]?.id || '',
    teacher_id: teachers.value[0]?.id || '',
    academic_year: activeExamEvent.value?.academic_year || '2026/2027',
  }
  showClassSubjectModal.value = true
}

const openEditClassSubject = (cs) => {
  isEditClassSubject.value = true
  classSubjectForm.value = {
    id: cs.id,
    class_id: cs.class_id || cs.class_room?.id || '',
    subject_id: cs.subject_id || cs.subject?.id || '',
    teacher_id: cs.teacher_id || cs.teacher?.id || '',
    academic_year: cs.academic_year || activeExamEvent.value?.academic_year || '2026/2027',
  }
  showClassSubjectModal.value = true
}

const submitClassSubjectForm = async () => {
  if (!classSubjectForm.value.class_id || !classSubjectForm.value.subject_id || !classSubjectForm.value.teacher_id) {
    await showAlertModal({
      title: 'Form Belum Lengkap',
      message: 'Kelas, Mata Pelajaran, dan Guru Pengampu wajib dipilih.',
      type: 'warning'
    })
    return
  }
  try {
    if (isEditClassSubject.value && classSubjectForm.value.id) {
      await api.put(`/admin/class-subjects/${classSubjectForm.value.id}`, classSubjectForm.value)
      showToast('Alokasi kelas mapel berhasil diperbarui!', 'success')
    } else {
      await api.post('/admin/class-subjects', classSubjectForm.value)
      showToast('Alokasi kelas mapel berhasil disimpan!', 'success')
    }
    showClassSubjectModal.value = false
    await loadClassSubjects()
    await loadAllData()
  } catch (e) {
    await showAlertModal({
      title: isEditClassSubject.value ? 'Gagal Memperbarui Alokasi' : 'Gagal Menyimpan Alokasi',
      message: e.response?.data?.message || 'Terjadi kesalahan saat menyimpan alokasi kelas mapel.',
      type: 'danger'
    })
  }
}

const deleteClassSubject = async (cs) => {
  const confirmed = await showConfirmModal({
    title: 'Hapus Alokasi Mapel',
    message: `Hapus alokasi ${cs.subject?.name} untuk kelas ${cs.class_room?.name}?`,
    type: 'danger',
    confirmText: 'Hapus Alokasi',
    cancelText: 'Batal'
  })
  if (!confirmed) return

  try {
    await api.delete(`/admin/class-subjects/${cs.id}`)
    await loadClassSubjects()
    await loadAllData()
    showToast('Alokasi kelas mapel berhasil dihapus', 'success')
  } catch (e) {
    await showAlertModal({
      title: 'Gagal Menghapus Alokasi',
      message: e.response?.data?.message || 'Terjadi kesalahan saat menghapus alokasi.',
      type: 'danger'
    })
  }
}

const loadEvents = async () => {
  try {
    const res = await api.get('/admin/events')
    events.value = res.data.data || []
    if (selectedEventId.value) {
      const exists = events.value.some(e => e.id === selectedEventId.value)
      if (!exists) {
        selectedEventId.value = ''
        selectedEventFilter.value = ''
        sessionStorage.removeItem('cbt_selected_event_id')
      }
    }
  } catch (e) {
    console.error('Failed to load events', e)
  }
}

const loadSchedules = async () => {
  try {
    const params = {}
    const evId = currentScopedEventId.value
    if (evId) {
      params.event_id = evId
    }
    const res = await api.get('/admin/schedules', { params })
    schedules.value = res.data.data || []
  } catch (e) {
    console.error('Failed to load schedules', e)
  }
}

const filterByEventAndOpenSchedules = (ev) => {
  enterEvent(ev)
}

const openCreateEvent = () => {
  isEditEvent.value = false
  const now = new Date()
  const nextMonth = new Date(now.getTime() + 14 * 24 * 60 * 60 * 1000)
  eventForm.value = {
    id: null,
    title: '',
    code: '',
    academic_year: '2026/2027',
    semester: 'GANJIL',
    start_date: now.toISOString().split('T')[0],
    end_date: nextMonth.toISOString().split('T')[0],
    is_active: false,
    description: '',
  }
  showEventModal.value = true
}

const openEditEvent = (ev) => {
  isEditEvent.value = true
  eventForm.value = {
    id: ev.id,
    title: ev.title,
    code: ev.code,
    academic_year: ev.academic_year,
    semester: ev.semester,
    start_date: ev.start_date ? ev.start_date.split('T')[0] : '',
    end_date: ev.end_date ? ev.end_date.split('T')[0] : '',
    is_active: ev.is_active,
    description: ev.description || '',
  }
  showEventModal.value = true
}

const submitEventForm = async () => {
  try {
    if (isEditEvent.value) {
      await api.put(`/admin/events/${eventForm.value.id}`, eventForm.value)
      showToast('✓ Event berhasil diperbarui!', 'success')
    } else {
      await api.post('/admin/events', eventForm.value)
      showToast('✓ Event ujian baru berhasil dibuat!', 'success')
    }
    showEventModal.value = false
    await loadEvents()
    await loadSchedules()
  } catch (e) {
    await showAlertModal({
      title: 'Gagal Menyimpan Event',
      message: e.response?.data?.message || 'Terjadi kesalahan saat menyimpan data event.',
      type: 'danger'
    })
  }
}

const toggleEventActive = async (ev) => {
  try {
    const res = await api.post(`/admin/events/${ev.id}/toggle-active`)
    ev.is_active = res.data.is_active
    await loadEvents()
    await loadSchedules()
    showToast(ev.is_active ? `Event ${ev.title} diaktifkan!` : `Event ${ev.title} dinonaktifkan.`, 'info')
  } catch (e) {
    await showAlertModal({
      title: 'Gagal Mengubah Status',
      message: e.response?.data?.message || 'Gagal mengubah status event.',
      type: 'danger'
    })
  }
}

const deleteEvent = async (ev) => {
  const confirmed = await showConfirmModal({
    title: 'Hapus Event Ujian',
    message: `Hapus event "${ev.title}"?\n\nPerhatian: Event dengan sesi jadwal ujian aktif tidak dapat dihapus.`,
    type: 'danger',
    confirmText: 'Hapus Event',
    cancelText: 'Batal'
  })
  if (!confirmed) return

  try {
    await api.delete(`/admin/events/${ev.id}`)
    showToast('Event berhasil dihapus', 'success')
    await loadEvents()
  } catch (e) {
    await showAlertModal({
      title: 'Gagal Menghapus Event',
      message: e.response?.data?.message || 'Terjadi kesalahan saat menghapus event.',
      type: 'danger'
    })
  }
}

const resetUserSession = async (u) => {
  const confirmed = await showConfirmModal({
    title: 'Lepas Kunci Sesi Login',
    message: `Lepas kunci sesi login untuk pengguna ${u.username}? Siswa akan dapat login kembali di perangkat baru.`,
    type: 'warning',
    confirmText: 'Lepas Kunci',
    cancelText: 'Batal'
  })
  if (!confirmed) return

  try {
    await api.post(`/admin/users/${u.id}/reset-session`)
    u.has_active_session = false
    showToast(`Sesi login untuk ${u.username} berhasil direset`, 'success')
  } catch (e) {
    await showAlertModal({
      title: 'Gagal Mereset Sesi',
      message: e.response?.data?.message || 'Terjadi kesalahan saat mereset sesi.',
      type: 'danger'
    })
  }
}

const downloadStudentTemplate = async () => {
  try {
    const res = await api.get('/admin/users/template', {
      responseType: 'blob'
    })
    const url = window.URL.createObjectURL(new Blob([res.data], {
      type: 'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet'
    }))
    const link = document.createElement('a')
    link.href = url
    link.setAttribute('download', 'Template_Import_Siswa.xlsx')
    document.body.appendChild(link)
    link.click()
    link.remove()
    window.URL.revokeObjectURL(url)
    showToast('Format template siswa berhasil diunduh!', 'success')
  } catch (err) {
    await showAlertModal({
      title: 'Gagal Mengunduh Template',
      message: await extractExportErrorMessage(err, 'Gagal mengunduh template'),
      type: 'danger'
    })
  }
}

const handleExcelFileSelect = (e) => {
  importExcelFile.value = e.target.files[0]
}

const submitImportExcel = async () => {
  if (!importExcelFile.value) {
    await showAlertModal({
      title: 'Pilih File Excel',
      message: 'Pilih file Excel terlebih dahulu!',
      type: 'warning'
    })
    return
  }
  const formData = new FormData()
  formData.append('file', importExcelFile.value)
  isImporting.value = true
  try {
    const res = await api.post('/admin/users/import-excel', formData, {
      headers: { 'Content-Type': 'multipart/form-data' },
    })
    showToast(res.data.message || 'Import berhasil!', 'success')
    showImportModal.value = false
    importExcelFile.value = null
    await loadAllData()
  } catch (e) {
    await showAlertModal({
      title: 'Gagal Impor Excel',
      message: e.response?.data?.message || 'Terjadi kesalahan saat mengimpor file Excel.',
      type: 'danger'
    })
  } finally {
    isImporting.value = false
  }
}

const downloadClassTemplate = async () => {
  try {
    const res = await api.get('/admin/classes/template', { responseType: 'blob' })
    const url = window.URL.createObjectURL(new Blob([res.data], {
      type: 'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet'
    }))
    const a = document.createElement('a')
    a.href = url
    a.download = 'Template_Import_Kelas_CBT.xlsx'
    document.body.appendChild(a)
    a.click()
    a.remove()
    window.URL.revokeObjectURL(url)
    showToast('Template kelas berhasil diunduh!', 'success')
  } catch (err) {
    await showAlertModal({ title: 'Gagal Mengunduh Template', message: 'Gagal mengunduh template kelas.', type: 'danger' })
  }
}

const submitImportClasses = async () => {
  if (!importClassFile.value) return
  isImportingClass.value = true
  importClassResult.value = null
  try {
    const fd = new FormData()
    fd.append('file', importClassFile.value)
    const res = await api.post('/admin/classes/import', fd, {
      headers: { 'Content-Type': 'multipart/form-data' },
    })
    importClassResult.value = { success: true, message: res.data.message || 'Impor kelas berhasil!' }
    await loadAllData()
  } catch (e) {
    importClassResult.value = { success: false, message: e.response?.data?.message || 'Gagal mengunggah file' }
  } finally {
    isImportingClass.value = false
  }
}

const downloadSubjectTemplate = async () => {
  try {
    const res = await api.get('/admin/subjects/template', { responseType: 'blob' })
    const url = window.URL.createObjectURL(new Blob([res.data], {
      type: 'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet'
    }))
    const a = document.createElement('a')
    a.href = url
    a.download = 'Template_Import_Mapel_CBT.xlsx'
    document.body.appendChild(a)
    a.click()
    a.remove()
    window.URL.revokeObjectURL(url)
    showToast('Template mata pelajaran berhasil diunduh!', 'success')
  } catch (err) {
    await showAlertModal({ title: 'Gagal Mengunduh Template', message: 'Gagal mengunduh template mata pelajaran.', type: 'danger' })
  }
}

const submitImportSubjects = async () => {
  if (!importSubjectFile.value) return
  isImportingSubject.value = true
  importSubjectResult.value = null
  try {
    const fd = new FormData()
    fd.append('file', importSubjectFile.value)
    const res = await api.post('/admin/subjects/import', fd, {
      headers: { 'Content-Type': 'multipart/form-data' },
    })
    importSubjectResult.value = { success: true, message: res.data.message || 'Impor mata pelajaran berhasil!' }
    await loadAllData()
  } catch (e) {
    importSubjectResult.value = { success: false, message: e.response?.data?.message || 'Gagal mengunggah file' }
  } finally {
    isImportingSubject.value = false
  }
}

const downloadTeacherTemplate = async () => {
  try {
    const res = await api.get('/admin/teachers/template', { responseType: 'blob' })
    const url = window.URL.createObjectURL(new Blob([res.data], {
      type: 'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet'
    }))
    const a = document.createElement('a')
    a.href = url
    a.download = 'Template_Import_Guru_CBT.xlsx'
    document.body.appendChild(a)
    a.click()
    a.remove()
    window.URL.revokeObjectURL(url)
    showToast('Template guru berhasil diunduh!', 'success')
  } catch (err) {
    await showAlertModal({ title: 'Gagal Mengunduh Template', message: 'Gagal mengunduh template guru.', type: 'danger' })
  }
}

const submitImportTeachers = async () => {
  if (!importTeacherFile.value) return
  isImportingTeacher.value = true
  importTeacherResult.value = null
  try {
    const fd = new FormData()
    fd.append('file', importTeacherFile.value)
    const res = await api.post('/admin/teachers/import', fd, {
      headers: { 'Content-Type': 'multipart/form-data' },
    })
    importTeacherResult.value = { success: true, message: res.data.message || 'Impor guru berhasil!' }
    await loadAllData()
  } catch (e) {
    importTeacherResult.value = { success: false, message: e.response?.data?.message || 'Gagal mengunggah file' }
  } finally {
    isImportingTeacher.value = false
  }
}

const toggleScheduleStatus = async (sch) => {
  if (!sch.is_active) {
    if (!sch.bank_id) {
      await showAlertModal({
        title: 'Naskah Soal Belum Ditautkan',
        message: 'Jadwal belum dapat diaktifkan karena belum ada Naskah Bank Soal yang ditautkan.\n\nSilakan klik "Tautkan Soal" terlebih dahulu.',
        type: 'warning'
      })
      return
    }
    if (sch.bank && !sch.bank.is_locked) {
      await showAlertModal({
        title: 'Bank Soal Belum Siap',
        message: 'Bank Soal masih berstatus Draft (Belum Dikunci).\n\nSilakan kunci/finalisasi bank soal pada menu Kesiapan Soal sebelum mengaktifkan sesi ujian.',
        type: 'warning'
      })
      return
    }
  }
  try {
    const res = await api.post(`/admin/schedules/${sch.id}/toggle`)
    sch.is_active = res.data.is_active
    showToast(sch.is_active ? `Sesi ${sch.title} berhasil diaktifkan!` : `Sesi ${sch.title} dinonaktifkan.`, 'success')
  } catch (e) {
    await showAlertModal({
      title: 'Gagal Mengubah Status',
      message: 'Terjadi kesalahan saat mengubah status aktif jadwal sesi ujian.',
      type: 'danger'
    })
  }
}

const toggleBankLock = async (b) => {
  try {
    const res = await api.post(`/admin/question-banks/${b.id}/toggle-lock`)
    b.is_locked = res.data.is_locked
    showToast(b.is_locked ? 'Bank Soal dikunci (Terkunci)' : 'Bank Soal dibuka (Draft)', 'info')
  } catch (e) {
    await showAlertModal({
      title: 'Gagal Mengubah Status Bank Soal',
      message: 'Terjadi kesalahan saat mengubah status penguncian bank soal.',
      type: 'danger'
    })
  }
}

// --- Student Master Data Methods ---
const openCreateStudentModal = () => {
  newStudent.value = {
    full_name: '',
    username: '',
    password: '',
    nis: '',
    nisn: '',
    class_id: classes.value[0]?.id || '',
    gender: 'L'
  }
  showStudentModal.value = true
}

const openEditStudent = (st) => {
  editStudentForm.value = {
    id: st.id,
    user_id: st.user_id || st.user?.id,
    full_name: st.user?.full_name || '',
    username: st.user?.username || '',
    password: '',
    nis: st.nis || '',
    nisn: st.nisn || '',
    class_id: st.class_room_id || st.class_id || classes.value[0]?.id || '',
    gender: st.gender || 'L'
  }
  showEditStudentModal.value = true
}

const openStudentDetail = (st) => {
  selectedStudentDetail.value = st
  showStudentDetailModal.value = true
}

const createStudent = async () => {
  try {
    await api.post('/admin/students', newStudent.value)
    showStudentModal.value = false
    newStudent.value = { full_name: '', username: '', password: '', nis: '', nisn: '', class_id: classes.value[0]?.id || '', gender: 'L' }
    await loadAllData()
    showToast('Siswa berhasil ditambahkan!', 'success')
  } catch (e) {
    await showAlertModal({
      title: 'Gagal Menambah Siswa',
      message: e.response?.data?.message || 'Terjadi kesalahan saat menambah data siswa.',
      type: 'danger'
    })
  }
}

const updateStudent = async () => {
  try {
    const targetId = editStudentForm.value.id || editStudentForm.value.user_id
    await api.put(`/admin/students/${targetId}`, editStudentForm.value)
    showEditStudentModal.value = false
    await loadAllData()
    showToast('Data siswa berhasil diperbarui!', 'success')
  } catch (e) {
    await showAlertModal({
      title: 'Gagal Memperbarui Siswa',
      message: e.response?.data?.message || 'Terjadi kesalahan saat memperbarui data siswa.',
      type: 'danger'
    })
  }
}

const deleteStudent = async (st) => {
  const confirmed = await showConfirmModal({
    title: 'Hapus Akun Siswa',
    message: `Yakin ingin menghapus akun siswa ${st.user?.full_name || ''} (${st.user?.username || ''})?\n\nTindakan ini akan menghapus profil siswa beserta riwayat nilai terkait.`,
    type: 'danger',
    confirmText: 'Hapus Siswa',
    cancelText: 'Batal'
  })
  if (!confirmed) return

  try {
    const targetId = st.id || st.user_id
    await api.delete(`/admin/students/${targetId}`)
    await loadAllData()
    showToast('Siswa berhasil dihapus', 'success')
  } catch (e) {
    await showAlertModal({
      title: 'Gagal Menghapus Siswa',
      message: e.response?.data?.message || 'Terjadi kesalahan saat menghapus data siswa.',
      type: 'danger'
    })
  }
}

// --- Teacher Master Data Methods ---
const openCreateTeacherModal = async () => {
  const catalog = await ensurePermissionCatalog()
  if (!catalog) return
  newTeacher.value = { full_name: '', username: '', password: '' }
  newTeacherAccess.value = defaultStaffAccess(catalog)
  newTeacherAccessValid.value = true
  showTeacherModal.value = true
}

const openEditTeacher = async (tc) => {
  const catalog = await ensurePermissionCatalog()
  if (!catalog) return
  editTeacherForm.value = {
    id: tc.id,
    full_name: tc.full_name || '',
    username: tc.username || '',
    password: ''
  }
  editTeacherAccess.value = staffAccessFromAccount(tc, catalog)
  editTeacherAccessValid.value = true
  showEditTeacherModal.value = true
}

const createTeacher = async () => {
  if (!newTeacherCanSave.value) return
  try {
    await api.post('/admin/teachers', buildStaffCreatePayload(newTeacher.value, newTeacherAccess.value))
    showTeacherModal.value = false
    newTeacher.value = { full_name: '', username: '', password: '' }
    await loadAllData()
    showToast('Akun guru / staf berhasil ditambahkan!', 'success')
  } catch (e) {
    await showAlertModal({
      title: 'Gagal Menambah Guru / Staf',
      message: e.response?.data?.message || 'Terjadi kesalahan saat menambah guru / staf.',
      type: 'danger'
    })
  }
}

const updateTeacher = async () => {
  if (!editTeacherCanSave.value) return
  try {
    await api.put(`/admin/teachers/${editTeacherForm.value.id}`, buildStaffUpdatePayload(editTeacherForm.value, editTeacherAccess.value))
    showEditTeacherModal.value = false
    await loadAllData()
    showToast('Data guru / staf berhasil diperbarui!', 'success')
  } catch (e) {
    await showAlertModal({
      title: 'Gagal Memperbarui Guru / Staf',
      message: e.response?.data?.message || 'Terjadi kesalahan saat memperbarui data guru / staf.',
      type: 'danger'
    })
  }
}

const deleteTeacher = async (tc) => {
  const confirmed = await showConfirmModal({
    title: 'Hapus Guru / Staf',
    message: `Yakin ingin menghapus akun ${tc.full_name} (${tc.username})?`,
    type: 'danger',
    confirmText: 'Hapus Akun',
    cancelText: 'Batal'
  })
  if (!confirmed) return

  try {
    await api.delete(`/admin/teachers/${tc.id}`)
    await loadAllData()
    showToast('Akun guru / staf berhasil dihapus', 'success')
  } catch (e) {
    await showAlertModal({
      title: 'Gagal Menghapus Guru / Staf',
      message: e.response?.data?.message || 'Terjadi kesalahan saat menghapus akun guru / staf.',
      type: 'danger'
    })
  }
}

// --- Class Master Data Methods ---
const openCreateClassModal = () => {
  newClass.value = { name: '', grade: 'XII', major: 'MIPA' }
  showClassModal.value = true
}

const openEditClass = (c) => {
  editClassForm.value = {
    id: c.id,
    name: c.name || '',
    grade: c.grade || 'XII',
    major: c.major || ''
  }
  showEditClassModal.value = true
}

const createClass = async () => {
  try {
    await api.post('/admin/classes', newClass.value)
    showClassModal.value = false
    newClass.value = { name: '', grade: 'XII', major: 'MIPA' }
    await loadAllData()
    showToast('Kelas berhasil ditambahkan!', 'success')
  } catch (e) {
    await showAlertModal({
      title: 'Gagal Menambah Kelas',
      message: e.response?.data?.message || 'Terjadi kesalahan saat menambah data kelas.',
      type: 'danger'
    })
  }
}

const updateClass = async () => {
  try {
    await api.put(`/admin/classes/${editClassForm.value.id}`, editClassForm.value)
    showEditClassModal.value = false
    await loadAllData()
    showToast('Data kelas berhasil diperbarui!', 'success')
  } catch (e) {
    await showAlertModal({
      title: 'Gagal Memperbarui Kelas',
      message: e.response?.data?.message || 'Terjadi kesalahan saat memperbarui data kelas.',
      type: 'danger'
    })
  }
}

const deleteClass = async (c) => {
  const confirmed = await showConfirmModal({
    title: 'Hapus Kelas',
    message: `Yakin ingin menghapus rombel kelas "${c.name}"?\n\nPerhatian: Siswa dan alokasi mapel yang terhubung ke kelas ini dapat terdampak.`,
    type: 'danger',
    confirmText: 'Hapus Kelas',
    cancelText: 'Batal'
  })
  if (!confirmed) return

  try {
    await api.delete(`/admin/classes/${c.id}`)
    await loadAllData()
    showToast('Kelas berhasil dihapus', 'success')
  } catch (e) {
    await showAlertModal({
      title: 'Gagal Menghapus Kelas',
      message: e.response?.data?.message || 'Terjadi kesalahan saat menghapus data kelas.',
      type: 'danger'
    })
  }
}

const handleLogout = async () => {
  try {
    await authStore.logout()
  } catch (e) {
    console.error('Logout error', e)
  }
  localStorage.clear()
  sessionStorage.clear()
  window.location.href = '/login'
}


const closeEventMenuOnClickOutside = () => {
  if (activeEventMenuId.value !== null) {
    activeEventMenuId.value = null
  }
}

onUnmounted(() => {
  window.removeEventListener('click', closeEventMenuOnClickOutside)
})

onMounted(() => {
  window.addEventListener('click', closeEventMenuOnClickOutside)
  const queryTab = router.currentRoute.value.query.tab
  const initialTab = queryTab || landingTab(authStore.user)
  if (initialTab !== 'dashboard') {
    switchTab(initialTab)
  } else {
    loadLoginSessions()
    loadTeacherHomeProctorTasks()
  }
  loadAllData()
})
</script>

<style scoped>
.card-list-move {
  transition: transform 0.25s cubic-bezier(0.2, 0, 0, 1);
}
</style>
