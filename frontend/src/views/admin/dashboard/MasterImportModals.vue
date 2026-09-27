<template>
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
</template>

<script setup>
import { useDashboard } from './context'

const {
  downloadClassTemplate,
  downloadStudentTemplate,
  downloadSubjectTemplate,
  downloadTeacherTemplate,
  handleExcelFileSelect,
  importClassFile,
  importClassResult,
  importSubjectFile,
  importSubjectResult,
  importTeacherFile,
  importTeacherResult,
  isImporting,
  isImportingClass,
  isImportingSubject,
  isImportingTeacher,
  showImportClassModal,
  showImportModal,
  showImportSubjectModal,
  showImportTeacherModal,
  submitImportClasses,
  submitImportExcel,
  submitImportSubjects,
  submitImportTeachers,
} = useDashboard()
</script>
