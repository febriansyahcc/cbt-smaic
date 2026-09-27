<template>
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
</template>

<script setup>
import { useDashboard } from './context'

const {
  classSubjectForm,
  classes,
  isEditClassSubject,
  isEditSubject,
  showClassSubjectModal,
  showSubjectModal,
  subjectForm,
  subjects,
  submitClassSubjectForm,
  submitSubjectForm,
  teachers,
} = useDashboard()
</script>
