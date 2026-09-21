# Standar Keamanan, Anti-Cheat & Live Proctoring

Pedoman implementasi protokol integritas dan pengawasan ujian di CBT SMAS Islamic Centre Demak.

---

## 1. Lapisan Deteksi Pelanggaran (Client-Side)
Setiap sesi pengerjaan ujian siswa wajib mengaktifkan proteksi:
1. **Page Visibility API (`document.visibilitychange`):**
   - Mendeteksi ketika siswa berpindah tab browser atau meminimalkan aplikasi.
2. **Window Blur Event (`window.onblur`):**
   - Mendeteksi notifikasi pop-up, split-screen pada ponsel, atau aplikasi latar depan lain yang mengambil fokus.
3. **Fullscreen Lock (`fullscreenchange`):**
   - Mengharuskan mode layar penuh (pada desktop) dan mencatat jika siswa keluar dari layar penuh secara sengaja.
4. **Input & Clipboard Restrictions:**
   - Mencegah `contextmenu` (klik kanan), `selectstart` (blok teks), `copy`, `paste`, dan tombol pintas inspect element (`F12`, `Ctrl+Shift+I`, `Ctrl+U`).

---

## 2. Batas Pelanggaran & Auto-Lock
- **Toleransi Pelanggaran:** Sistem memiliki batas kuota pelanggaran (misalnya maksimal 3 kali pergantian layar/tab).
- **Auto-Lock Trigger:** Jika kuota terlampaui, status sesi siswa berubah menjadi `LOCKED`. Layar ujian terkunci seketika dan siswa tidak dapat melanjutkan tanpa verifikasi proktor/guru.
- **Audit Logging:** Setiap pelanggaran dikirimkan ke endpoint backend (`/api/exam/log-violation`) dengan timestamp server.

---

## 3. Live Proctoring Dashboard (Guru / Admin)
- Proktor dapat memantau secara real-time:
  - Status koneksi siswa (Aktif / Putus / Terkunci).
  - Jumlah pelanggaran yang tercatat.
  - Progres jumlah soal yang sudah terjawab.
  - Sisa waktu pengerjaan.
- Tindakan Proktor:
  - **Buka Kunci (`Unlock Session`):** Membuka kembali status siswa yang terkena auto-lock setelah mendapat teguran proktor.
  - **Reset Sesi Perangkat (`Reset Session`):** Mengizinkan siswa login dari perangkat lain jika terjadi kendala teknis perangkat (HP mati/rusak) tanpa menghapus jawaban yang telah tersimpan.
