# Standar Pengembangan Frontend (Vue 3, Vite & Tailwind CSS)

Pedoman teknis untuk pengembangan antarmuka pengguna (UI/UX) pada sistem CBT SMAS Islamic Centre Demak.

---

## 1. Komposisi Komponen & State Management
- **Vue 3 Composition API:** Selalu gunakan sintaks `<script setup>` yang bersih dan modular.
- **Pinia Stores:** Pisahkan state global (otentikasi, sesi ujian aktif, pengaturan tema/sekolah) ke dalam folder `frontend/src/stores/`.
- **API Services Layer:** Semua pemanggilan endpoint HTTP ditempatkan di `frontend/src/services/` menggunakan Axios client terpusat yang menangani Authorization Bearer token secara otomatis.

---

## 2. Standar Mobile-First & Responsivitas
Antarmuka pengerjaan ujian siswa dirancang dengan **Mobile-First 3-Zone Layout**:
1. **Fixed Topbar (Zona Header):**
   - Menampilkan sisa waktu ujian (Countdown Timer), status koneksi sinkronisasi (Online/Offline), dan tombol drawer navigasi soal.
2. **Scrollable Question Canvas (Zona Konten Utama):**
   - Menampilkan nomor soal, bobot/ragam soal, stimulus gambar/teks, dan pilihan ganda/esay dengan area sentuh jempol yang lega (`min-h-[48px]`, `p-4`).
3. **Fixed Bottombar (Zona Aksi Bawah):**
   - Tombol navigasi 'Sebelumnya', 'Ragu-ragu', dan 'Berikutnya' / 'Selesai'.
   - Dilengkapi **Bottom-sheet Drawer** untuk nomor daftar soal dengan status warna (Belum Dijawab: Abu-abu, Ragu-ragu: Kuning, Terjawab: Hijau/Biru).

---

## 3. Komponen Dialog & Modal Reaktif
- Jangan gunakan `window.alert()` atau `window.confirm()`.
- Wajib menggunakan custom modal helper terpusat atau composable modal (`showConfirmModal`, `showAlertModal`, `showToast`).
- Transisi modal:
  - Overlay: `transition-opacity duration-200 ease-out`
  - Dialog Card: `transition-all duration-200 ease-out transform scale-95 -> scale-100`

---

## 4. Verifikasi & Build
Setiap perubahan kode frontend wajib diverifikasi dengan perintah:
```powershell
cd frontend
npm run build
```
