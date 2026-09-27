<template>
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
</template>

<script setup>
import StaffAccessEditor from '../../../components/admin/StaffAccessEditor.vue'
import { useDashboard } from './context'

const {
  createTeacher,
  editTeacherAccess,
  editTeacherAccessValid,
  editTeacherCanSave,
  editTeacherForm,
  isGrantor,
  newTeacher,
  newTeacherAccess,
  newTeacherAccessValid,
  newTeacherCanSave,
  permissionCatalog,
  showConfirmModal,
  showEditTeacherModal,
  showTeacherModal,
  updateTeacher,
} = useDashboard()
</script>
