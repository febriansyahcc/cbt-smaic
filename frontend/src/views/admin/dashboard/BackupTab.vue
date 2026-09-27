<template>
  <!-- TAB: BACKUP DATA (khusus administrator) -->
  <div v-if="activeTab === 'backup'" class="space-y-4">
    <!-- Ringkasan -->
    <div class="grid grid-cols-2 sm:grid-cols-3 gap-3">
      <div class="bg-white rounded-3xl p-4 border border-slate-200 shadow-xs">
        <span class="text-[11px] font-semibold text-slate-500">Backup Terakhir</span>
        <div class="text-base font-black text-slate-900 mt-1">
          {{ latestBackup ? formatBackupDate(latestBackup.created_at) : 'Belum ada' }}
        </div>
      </div>
      <div class="bg-white rounded-3xl p-4 border border-slate-200 shadow-xs">
        <span class="text-[11px] font-semibold text-slate-500">Tersimpan di Server</span>
        <div class="text-2xl font-black text-slate-900 mt-1">
          {{ backupItems.length }} <span class="text-xs font-medium text-slate-400">Berkas</span>
        </div>
      </div>
      <div class="col-span-2 sm:col-span-1 bg-white rounded-3xl p-4 border border-emerald-200 shadow-xs bg-emerald-50/20">
        <span class="text-[11px] font-semibold text-emerald-700">Backup Otomatis</span>
        <div class="text-sm font-bold text-emerald-800 mt-1">
          <template v-if="backupStatus.auto_enabled">
            Setiap hari mulai pukul {{ String(backupStatus.auto_hour ?? 1).padStart(2, '0') }}.00
          </template>
          <template v-else>Nonaktif</template>
        </div>
        <div class="text-[11px] text-emerald-700/80 mt-0.5">
          Menyimpan {{ backupStatus.keep || 14 }} backup otomatis terakhir
        </div>
      </div>
    </div>

    <!-- Aksi & Keterangan -->
    <div class="bg-white p-4 rounded-3xl border border-slate-200 shadow-xs flex flex-col md:flex-row md:items-center justify-between gap-3">
      <div class="text-xs text-slate-600 space-y-1 max-w-2xl">
        <p>
          Backup berisi seluruh database (akun, soal, jadwal, jawaban, nilai) dan gambar soal dalam satu berkas
          <span class="font-semibold">.tar.gz</span>.
        </p>
        <p class="text-amber-700 font-medium">
          Berkas di daftar ini tersimpan di disk server yang sama. Unduh backup secara berkala dan simpan di laptop
          atau flashdisk agar data tetap aman bila server rusak.
        </p>
      </div>
      <button
        type="button"
        @click="createBackup"
        :disabled="isBackupRunning"
        class="shrink-0 inline-flex items-center justify-center gap-2 px-4 py-2.5 rounded-2xl text-xs font-bold text-white bg-indigo-600 hover:bg-indigo-700 active:scale-95 transition disabled:opacity-60 disabled:cursor-not-allowed cursor-pointer"
      >
        <svg v-if="isBackupRunning" class="w-4 h-4 animate-spin" fill="none" viewBox="0 0 24 24">
          <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
          <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8v4a4 4 0 00-4 4H4z"></path>
        </svg>
        <svg v-else class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 7v10c0 2.21 3.582 4 8 4s8-1.79 8-4V7M4 7c0 2.21 3.582 4 8 4s8-1.79 8-4M4 7c0-2.21 3.582-4 8-4s8 1.79 8 4" />
        </svg>
        {{ isBackupRunning ? 'Sedang Membuat Backup...' : 'Buat Backup Sekarang' }}
      </button>
    </div>

    <!-- Galat backup terakhir -->
    <div
      v-if="backupStatus.last_error && !isBackupRunning"
      class="bg-rose-50 border border-rose-200 text-rose-700 rounded-3xl p-4 text-xs"
    >
      <div class="font-bold mb-0.5">Backup terakhir gagal</div>
      <div class="break-words">{{ backupStatus.last_error }}</div>
    </div>

    <!-- Daftar backup -->
    <div class="bg-white rounded-3xl border border-slate-200 shadow-xs overflow-hidden">
      <div class="overflow-x-auto">
        <table class="w-full text-left text-xs">
          <thead>
            <tr class="bg-slate-50 border-b border-slate-200 text-slate-600 font-bold uppercase text-[10px] select-none">
              <th class="py-3 px-4">Waktu Backup</th>
              <th class="hidden sm:table-cell py-3 px-4">Jenis</th>
              <th class="hidden sm:table-cell py-3 px-4">Ukuran</th>
              <th class="py-3 px-4 text-right">Aksi</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-slate-100">
            <tr v-for="item in backupItems" :key="item.name" class="hover:bg-slate-50/60">
              <td class="py-3 px-4">
                <div class="font-semibold text-slate-800 whitespace-nowrap">{{ formatBackupDate(item.created_at) }}</div>
                <div class="hidden md:block text-[10px] text-slate-400 font-mono break-all">{{ item.name }}</div>
                <div class="sm:hidden text-[11px] text-slate-500 mt-0.5">
                  {{ item.kind === 'auto' ? 'Otomatis' : 'Manual' }} · {{ formatBackupSize(item.size) }}
                </div>
              </td>
              <td class="hidden sm:table-cell py-3 px-4">
                <span
                  :class="[
                    'px-2 py-0.5 rounded-full text-[10px] font-bold',
                    item.kind === 'auto' ? 'bg-emerald-50 text-emerald-700' : 'bg-indigo-50 text-indigo-700'
                  ]"
                >
                  {{ item.kind === 'auto' ? 'Otomatis' : 'Manual' }}
                </span>
              </td>
              <td class="hidden sm:table-cell py-3 px-4 text-slate-600 whitespace-nowrap">{{ formatBackupSize(item.size) }}</td>
              <td class="py-3 px-4">
                <div class="flex items-center justify-end gap-2">
                  <button
                    type="button"
                    @click="downloadBackup(item)"
                    :disabled="!!downloadingBackup"
                    class="px-3 py-1.5 rounded-xl text-xs font-bold bg-slate-100 text-slate-700 hover:bg-slate-200 active:scale-95 transition disabled:opacity-60 cursor-pointer whitespace-nowrap"
                  >
                    {{ downloadingBackup === item.name ? 'Mengunduh...' : 'Unduh' }}
                  </button>
                  <button
                    type="button"
                    @click="deleteBackup(item)"
                    class="px-3 py-1.5 rounded-xl text-xs font-bold text-rose-600 hover:bg-rose-50 active:scale-95 transition cursor-pointer"
                  >
                    Hapus
                  </button>
                </div>
              </td>
            </tr>
            <tr v-if="!backupItems.length">
              <td colspan="4" class="py-10 px-4 text-center text-slate-400">
                {{ isLoadingBackups ? 'Memuat daftar backup...' : 'Belum ada backup. Tekan "Buat Backup Sekarang" untuk membuat backup pertama.' }}
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <p class="text-[11px] text-slate-500 px-1">
      Pemulihan data dari berkas backup dilakukan oleh teknisi di server, lihat Panduan Teknis Bab 5.
      Jangan memulihkan data saat ujian berlangsung.
    </p>
  </div>
</template>

<script setup>
import { useDashboard } from './context'

const {
  activeTab,
  backupItems,
  backupStatus,
  createBackup,
  deleteBackup,
  downloadBackup,
  downloadingBackup,
  formatBackupDate,
  formatBackupSize,
  isBackupRunning,
  isLoadingBackups,
  latestBackup,
} = useDashboard()
</script>
