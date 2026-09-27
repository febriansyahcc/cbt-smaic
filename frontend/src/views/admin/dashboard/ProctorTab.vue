<template>
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
</template>

<script setup>
import LiveProctorControl from '@/components/proctor/LiveProctorControl.vue'
import { useDashboard } from './context'

const {
  activeProctorScheduleId,
  activeTab,
  authStore,
  exportBeritaAcaraPDF,
  exportNilaiExcel,
  isExportingNilai,
  isExportingPDF,
  proctorData,
  proctorSchedules,
  showProctorPrint,
} = useDashboard()
</script>
