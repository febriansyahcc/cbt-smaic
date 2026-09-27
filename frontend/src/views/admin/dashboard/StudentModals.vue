<template>
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
</template>

<script setup>
import { useDashboard } from './context'

const {
  classes,
  createStudent,
  editStudentForm,
  newStudent,
  openEditStudent,
  resetStudentSession,
  selectedStudentDetail,
  showEditStudentModal,
  showStudentDetailModal,
  showStudentModal,
  updateStudent,
} = useDashboard()
</script>
