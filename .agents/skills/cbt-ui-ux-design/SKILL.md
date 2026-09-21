---
name: cbt-ui-ux-design
description: Guidelines and procedures for creating, auditing, and refactoring UI/UX components in the CBT app according to SMAS Islamic Centre Demak design standards.
---

# CBT UI/UX Design & Component Engineering

Gunakan skill ini saat merancang, mempercantik, atau merefaktorisasi halaman, modal dialog, formulir, atau komponen antarmuka pengguna di CBT.

---

## 🎨 Standar Utama Desain SMAS Islamic Centre Demak

1. **Zero Native Dialogs:**
   - Dilarang keras menggunakan `alert()` dan `confirm()`.
   - Gunakan modal helper reaktif:
     - Konfirmasi: `const confirmed = await showConfirmModal({ title, message, type: 'danger'|'warning'|'confirm' })`
     - Peringatan: `await showAlertModal({ title, message, type: 'info'|'warning'|'error' })`
     - Notifikasi sukses: `showToast({ message, type: 'success' })`

2. **Animasi Modal Terpisah & Bersih:**
   - **Backdrop (Overlay):** Hanya transisi opacity `opacity-0` -> `opacity-100`. **Jangan gunakan scale atau zoom pada backdrop.**
   - **Kartu Modal:** Transisi skala dan translate halus `scale-95 translate-y-2` -> `scale-100 translate-y-0`.
   - Sudut kartu: `rounded-3xl` dengan bayangan `shadow-2xl`.
   - Background kartu: Putih bersih (`bg-white`), tanpa garis pita warna di bagian atas.

3. **Prinsip Efisiensi Informasi (Minimalis):**
   - Gabungkan data yang terkait secara alami (misal: Mata Pelajaran + Tanggal/Waktu Sesi).
   - Jangan menyertakan emoji dekoratif atau ikon visual berlebih di header tabel/modal.
   - Status harus langsung memberikan aksi (misal: tombol 'Ganti Soal' vs 'Tautkan Soal').

4. **Mobile-First Touch Ergonomics:**
   - Target sentuh minimal: `48px` (`min-h-[48px]`, `py-3 px-4`).
   - Efek sentuh responsif: `active:scale-95 transition-transform duration-150`.
   - Tata letak ruang ujian: Fixed Topbar, Scrollable Question Card, Fixed Bottombar + Bottom Sheet Drawer.

---

## 🛠️ Contoh Template Modal Standar (Vue 3)

```vue
<template>
  <Transition
    enter-active-class="transition-opacity duration-200 ease-out"
    enter-from-class="opacity-0"
    enter-to-class="opacity-100"
    leave-active-class="transition-opacity duration-150 ease-in"
    leave-from-class="opacity-100"
    leave-to-class="opacity-0"
  >
    <div v-if="isOpen" class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-slate-900/60 backdrop-blur-xs">
      <Transition
        enter-active-class="transition-all duration-200 ease-out"
        enter-from-class="opacity-0 scale-95 translate-y-2"
        enter-to-class="opacity-100 scale-100 translate-y-0"
        leave-active-class="transition-all duration-150 ease-in"
        leave-from-class="opacity-100 scale-100 translate-y-0"
        leave-to-class="opacity-0 scale-95 translate-y-2"
      >
        <div class="relative w-full max-w-lg bg-white rounded-3xl shadow-2xl overflow-hidden p-6">
          <!-- Header -->
          <div class="flex items-center justify-between pb-4 border-b border-slate-100">
            <h3 class="text-lg font-bold text-slate-900">{{ title }}</h3>
            <button @click="close" class="p-2 text-slate-400 hover:text-slate-600 rounded-full hover:bg-slate-100 transition">
              <svg class="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
              </svg>
            </button>
          </div>

          <!-- Body -->
          <div class="py-4 text-slate-600">
            <slot />
          </div>

          <!-- Footer -->
          <div class="flex items-center justify-end gap-3 pt-4 border-t border-slate-100">
            <button @click="close" class="px-5 py-2.5 text-sm font-medium text-slate-600 hover:bg-slate-100 rounded-xl transition active:scale-95">
              Batal
            </button>
            <button @click="confirm" class="px-5 py-2.5 text-sm font-medium text-white bg-emerald-600 hover:bg-emerald-700 rounded-xl transition shadow-lg shadow-emerald-600/20 active:scale-95">
              Simpan
            </button>
          </div>
        </div>
      </Transition>
    </div>
  </Transition>
</template>
```
