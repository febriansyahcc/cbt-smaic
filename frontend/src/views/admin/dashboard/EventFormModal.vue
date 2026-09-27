<template>
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
</template>

<script setup>
import { useDashboard } from './context'

const {
  eventForm,
  isEditEvent,
  showEventModal,
  submitEventForm,
} = useDashboard()
</script>
