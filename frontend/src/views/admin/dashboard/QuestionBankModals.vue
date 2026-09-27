<template>
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
            <label class="block font-bold text-slate-700 mb-1">Cakupan Bank Soal:</label>
            <div class="grid grid-cols-2 gap-1 p-1 bg-slate-100 rounded-xl">
              <button
                v-for="opt in bankScopeOptions"
                :key="opt.value"
                type="button"
                @click="setBankScope(opt.value)"
                class="py-2 rounded-lg font-bold transition active:scale-95 cursor-pointer"
                :class="newBankForm.scope === opt.value ? 'bg-white text-indigo-700 shadow-xs' : 'text-slate-500 hover:text-slate-700'"
              >
                {{ opt.label }}
              </button>
            </div>
          </div>

          <!-- Mode per angkatan (default) -->
          <div v-if="newBankForm.scope === 'grade'">
            <label class="block font-bold text-slate-700 mb-1">Pilih Mata Pelajaran &amp; Tingkat:</label>
            <select
              v-model="newBankOptionKey"
              @change="onBankSubjectChange"
              required
              :disabled="availableBankSubjects.length === 0 && !isEditBank"
              class="w-full px-3 py-2.5 rounded-xl border border-slate-300 font-medium focus:ring-2 focus:ring-indigo-600 bg-slate-50 text-xs cursor-pointer disabled:bg-slate-100 disabled:text-slate-400"
            >
              <option value="" disabled>
                {{ (availableBankSubjects.length === 0 && !isEditBank) ? '-- Semua Mata Pelajaran Sudah Memiliki Bank Soal --' : '-- Pilih Mata Pelajaran --' }}
              </option>
              <option v-for="s in availableBankSubjects" :key="s.key" :value="s.key">
                {{ s.name }}{{ s.grade ? ` — Kelas ${s.grade}` : '' }} ({{ s.code }})
              </option>
            </select>
            <p v-if="availableBankSubjects.length > 0" class="text-[11px] text-slate-500 mt-1.5 flex items-center gap-1.5">
              <span class="w-1.5 h-1.5 rounded-full bg-indigo-600 shrink-0"></span>
              <span>Satu naskah untuk seluruh kelas di angkatan tersebut.</span>
            </p>
            <p v-else-if="!isEditBank" class="text-[11px] text-emerald-600 font-semibold mt-1.5 flex items-center gap-1.5">
              <span class="w-1.5 h-1.5 rounded-full bg-emerald-500 shrink-0"></span>
              <span>Seluruh mata pelajaran jadwal event ini sudah memiliki bank soal.</span>
            </p>
          </div>

          <!-- Mode kelas tertentu -->
          <template v-else>
            <div>
              <label class="block font-bold text-slate-700 mb-1">Pilih Mata Pelajaran:</label>
              <select
                v-model="newBankForm.subject_id"
                @change="onBankCustomSubjectChange"
                required
                class="w-full px-3 py-2.5 rounded-xl border border-slate-300 font-medium focus:ring-2 focus:ring-indigo-600 bg-slate-50 text-xs cursor-pointer"
              >
                <option value="" disabled>-- Pilih Mata Pelajaran --</option>
                <option v-for="s in bankCustomSubjects" :key="s.id" :value="s.id">{{ s.name }} ({{ s.code }})</option>
              </select>
            </div>
            <div v-if="newBankForm.subject_id">
              <label class="block font-bold text-slate-700 mb-1">Pilih Kelas:</label>
              <div v-if="bankCustomClassOptions.length > 0" class="flex flex-wrap gap-1.5">
                <button
                  v-for="c in bankCustomClassOptions"
                  :key="c.id"
                  type="button"
                  :disabled="c.covered"
                  @click="toggleBankClass(c.id)"
                  :title="c.covered ? 'Kelas ini sudah memiliki bank soal untuk mapel ini' : ''"
                  class="px-3 py-1.5 rounded-lg border font-bold transition active:scale-95 cursor-pointer disabled:cursor-not-allowed disabled:opacity-40 disabled:line-through"
                  :class="newBankForm.class_ids.includes(c.id) ? 'bg-indigo-600 border-indigo-600 text-white' : 'bg-white border-slate-300 text-slate-700 hover:border-indigo-400'"
                >
                  {{ c.name }}
                </button>
              </div>
              <p v-else class="text-[11px] text-slate-500">Belum ada kelas untuk mapel ini.</p>
              <p class="text-[11px] text-slate-500 mt-1.5 flex items-center gap-1.5">
                <span class="w-1.5 h-1.5 rounded-full bg-indigo-600 shrink-0"></span>
                <span>Boleh lintas angkatan. Kelas yang dicoret sudah punya bank soal mapel ini.</span>
              </p>
            </div>
          </template>

          <div>
            <label class="block font-bold text-slate-700 mb-1">Judul / Nama Paket Bank Soal:</label>
            <input
              v-model="newBankForm.title"
              type="text"
              required
              :disabled="bankFormLocked"
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
              :disabled="isCreatingBank || bankFormLocked"
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
</template>

<script setup>
import { useDashboard } from './context'

const {
  availableBankSubjects,
  bankCustomClassOptions,
  bankCustomSubjects,
  bankFormLocked,
  bankScopeOptions,
  downloadQuestionBankTemplate,
  handleQuestionBankFileChange,
  isCreatingBank,
  isEditBank,
  isUploadingQuestionBank,
  newBankForm,
  newBankOptionKey,
  onBankCustomSubjectChange,
  onBankSubjectChange,
  questionBankUploadStatus,
  readinessData,
  selectedBankUploadId,
  setBankScope,
  showCreateBankModal,
  showUploadBankModal,
  submitCreateBank,
  submitQuestionBankUpload,
  toggleBankClass,
} = useDashboard()
</script>
