<template>
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
</template>

<script setup>
import { useDashboard } from './context'

const {
  createClass,
  editClassForm,
  newClass,
  showClassModal,
  showEditClassModal,
  updateClass,
} = useDashboard()
</script>
