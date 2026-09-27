<template>
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
</template>

<script setup>
import { useDashboard } from './context'

const {
  currentSectionMeta,
  isMobileSidebarOpen,
  selectedExamEvent,
  switchTab,
} = useDashboard()
</script>
