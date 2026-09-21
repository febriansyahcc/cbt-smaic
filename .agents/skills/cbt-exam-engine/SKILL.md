---
name: cbt-exam-engine
description: Procedures and architecture guide for managing exam sessions, anti-cheat detection, deterministic seeded randomization, timer countdowns, and idempotent answer sync.
---

# CBT Exam Engine & Anti-Cheat Subsystem

Gunakan skill ini saat mengelola alur pengerjaan ujian siswa, algoritma pengacakan soal, sinkronisasi jawaban luring-daring, atau sistem proteksi kecurangan.

---

## 🔒 1. Mekanisme Anti-Cheat & Auto-Lock

### Event Listeners di Sisi Klien (`frontend/src/views/ExamView.vue`):
1. **Perpindahan Tab / Aplikasi:**
   ```javascript
   document.addEventListener('visibilitychange', () => {
     if (document.hidden) {
       handleCheatViolation('SWITCH_TAB', 'Siswa meninggalkan tab/aplikasi ujian')
     }
   })
   ```
2. **Kehilangan Fokus Jendela (Split Screen / Pop-up Notification):**
   ```javascript
   window.addEventListener('blur', () => {
     handleCheatViolation('WINDOW_BLUR', 'Jendela ujian kehilangan fokus')
   })
   ```
3. **Pencegahan Aksi Terlarang:**
   - Nonaktifkan `contextmenu` (klik kanan).
   - Nonaktifkan `selectstart` (blok tulisan).
   - Cegah kombinasi tombol keyboard: `F12`, `Ctrl+C`, `Ctrl+V`, `Ctrl+Shift+I`, `Ctrl+U`.

### Batas Kuota & Kunci Otomatis:
- Setiap pelanggaran dilaporkan ke backend melalui `POST /api/exam/violation`.
- Backend menghitung total pelanggaran. Jika `violation_count >= MAX_VIOLATIONS` (default: 3 kali), status sesi diubah menjadi `LOCKED`.
- Tampilan siswa seketika terkunci dan menampilkan pesan bahwa siswa wajib melapor ke pengawas ujian.

---

## 🎲 2. Seeded Deterministik Randomization

Pengacakan urutan soal dan pilihan ganda harus **identik saat siswa me-refresh browser**, tetapi **berbeda antar-siswa**.

1. **Seed Formula:**
   ```go
   seed := int64(studentID*100000 + scheduleID)
   rng := prng.NewCustomRNG(seed)
   ```
2. Pengacakan dieksekusi di backend saat payload sesi ujian pertama kali dikompilasi atau di-retrieve oleh siswa.
3. Kunci jawaban dan bobot penilaian tetap terpetakan ke ID asli soal di basis data.

---

## 🔄 3. Dual-State Offline-Tolerant Sync

Untuk mengantisipasi koneksi Wi-Fi sekolah yang tiba-tiba terputus:
1. **Simpan Instan di Memori Lokal (`localStorage` / `IndexedDB`):**
   - Saat siswa memilih opsi atau mengetik esai, jawaban langsung disimpan di `localStorage` dalam hitungan milidetik.
2. **Antrean Sinkronisasi Latar Belakang (Sync Queue):**
   - Jawaban dimasukkan ke queue pengiriman HTTP `POST /api/exam/answers/batch` secara asynchronous.
3. **Idempotent DB Operation di Backend:**
   - Backend menggunakan klausa `ON CONFLICT (student_id, schedule_id, question_id) DO UPDATE` pada PostgreSQL/SQLite untuk mencegah data duplikat atau bentrok versi.
