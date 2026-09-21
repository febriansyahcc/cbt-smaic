# PROMPT / SPESIFIKASI PENGEMBANGAN FITUR CBT SYSTEM
## 🎯 Fitur: Live Proctoring & Ruang Kontrol Pengawas (Search, Sort, Pagination, Reset Login, Tambah Waktu, & Log Anti-Cheat)

> **Target Repository**: `CBT System (Go Fiber + Vue 3 + GORM + Tailwind CSS)`  
> **Institusi**: SMAS Islamic Centre Demak  
> **Tujuan**: Mengembangkan antarmuka monitoring ruang ujian (*Live Proctoring*) yang tangguh, dilengkapi toolbar pencarian, filter status, pengurutan cerdas, paginasi konsisten, serta aksi remote pengendali sesi siswa (Reset Login, Buka Blokir, Tambah Waktu, Force Submit, & Log Pelanggaran).

---

## 🏛️ Pedoman Utama & Konteks Proyek (Agentic Rules)
1. **Zero Native Browser Dialogs**: Dilarang menggunakan `alert()` atau `confirm()`. Wajib menggunakan modal custom reaktif (`showConfirmModal`, `showAlertModal`, `showToast`).
2. **Animasi & Responsivitas**:
   - Backdrop modal hanya animasi *fade opacity* (`opacity-0` → `opacity-100`).
   - Kartu modal menggunakan scale halus (`scale-95 translate-y-2` → `scale-100 translate-y-0`).
   - Tombol interaktif dengan efek tekan (`active:scale-95`).
3. **Pola Konsistensi Komponen**:
   - Pola search, filter, sort, dan pagination mengacu pada tab Siswa & Jadwal Ujian di `AdminDashboardView.vue`.
4. **Verifikasi Wajib**:
   - Frontend build: `npm run build` di folder `frontend` (harus exit code 0).
   - Backend build: `go build ./cmd/api` di folder `backend` (harus exit code 0).
   - Docker: `docker compose up -d --build`.

---

## 📋 DAFTAR TUGAS IMPLEMENTASI (TASK BREAKDOWN)

### TUGAS 1: Backend Service & Endpoints Pengawasan (`ProctorService` & `Handlers`)

- [ ] **1.1 Tambah Waktu Ujian Siswa (Individu)**:
  - Lokasi: `backend/internal/service/proctor_service.go` & `backend/internal/handler/handlers.go`
  - Endpoint: `POST /api/v1/proctor/sessions/:id/extend-time`
  - Payload:
    ```json
    {
      "extra_minutes": 10,
      "reason": "Kendala teknis HP restart"
    }
    ```
  - **Logika**: Perbarui `EndTime = EndTime.Add(time.Duration(extraMinutes) * time.Minute)` pada tabel `exam_sessions`. Catat log audit.

- [ ] **1.2 Tambah Waktu Ujian Seluruh Kelas (Massal / 1 Jadwal)**:
  - Endpoint: `POST /api/v1/proctor/schedules/:id/extend-time-all`
  - Payload:
    ```json
    {
      "extra_minutes": 15,
      "reason": "Mati lampu lokal di ruang lab"
    }
    ```
  - **Logika**: Perbarui seluruh sesi aktif (`status = IN_PROGRESS`) pada `schedule_id` terkait.

- [ ] **1.3 Kumpulkan Paksa Ujian Siswa (*Force Submit*)**:
  - Endpoint: `POST /api/v1/proctor/sessions/:id/force-submit`
  - **Logika**:
    - Ubah status sesi menjadi `SUBMITTED`.
    - Panggil fungsi kalkulasi skor otomatis untuk menghitung nilai akhir PG & Isian Singkat.
    - Tutup sesi pengerjaan.

- [ ] **1.4 Riwayat Log Pelanggaran (*Cheat Violation Logs*)**:
  - Endpoint: `GET /api/v1/proctor/sessions/:id/violations`
  - **Logika**: Mengambil daftar `CheatLog` / insiden kecurangan (pindah tab, keluar fullscreen, blur window, dll.) lengkap dengan *timestamp* kejadian.

- [ ] **1.5 Buka Blokir & Reset Login Perangkat**:
  - Verifikasi endpoint yang sudah ada:
    - `POST /api/v1/proctor/sessions/:id/unlock` (Mengembalikan status `BLOCKED` -> `IN_PROGRESS`).
    - `POST /api/v1/proctor/students/:user_id/reset-session` (Mengosongkan `session_token` agar siswa bisa login di HP baru).

- [ ] **1.6 Registrasi Routing di `main.go`**:
  - Pastikan semua endpoint terdaftar di grup route `proctor` atau `admin`.

---

### TUGAS 2: Frontend Dashboard Pengawas (UI/UX, Search, Sort & Pagination)

- [ ] **2.1 Toolbar Kontrol & Filter Real-Time**:
  - Lokasi: Tab `proctor` di `frontend/src/views/admin/AdminDashboardView.vue` (atau `ProctorDashboardView.vue`).
  - **Komponen**:
    - 🔍 **Search Input**: Pencarian real-time berdasarkan Nama Siswa atau NIS/NISN.
    - 🏷️ **Status Filter**: Dropdown filter (`Semua Status`, `Sedang Mengerjakan`, `Terblokir`, `Selesai`, `Belum Mulai`).
    - ↕️ **Sort Selector**: Pengurutan berdasarkan Nama Siswa, Jumlah Pelanggaran (Tertinggi - Terendah), Progres Jawaban (%), atau Nilai PG.
    - 📄 **Per-Page Selector**: Dropdown `10`, `25`, `50`, `100`, `Semua`.
    - ⏱️ **Bulk Action Button**: Tombol *"⏱️ Tambah Waktu Seluruh Kelas"*.

- [ ] **2.2 Peningkatan Tabel Monitoring Siswa**:
  - **Kolom Tabel**:
    1. `No`
    2. `Siswa (Nama Lengkap & NIS)`
    3. `Status Pengerjaan` (Badge: Hijau/Mengerjakan, Merah/Terblokir, Emerald/Selesai, Abu-abu/Belum Mulai)
    4. `Sisa Waktu` (Format MM:SS countdown)
    5. `Progres Jawaban` (Progress bar + Angka `X / Y Soal (Z%)`)
    6. `Pelanggaran Anti-Cheat` (Badge merah/kuning dengan jumlah pelanggaran)
    7. `Nilai Sementara PG`
    8. `Aksi Pengawas` (Action buttons)
  - **Tombol Aksi Dinamis per Baris Siswa**:
    - 🔓 **Buka Blokir** (Hanya tampil jika `status === 'BLOCKED'`)
    - 🔄 **Reset Login** (Hanya tampil jika siswa punya session aktif / kendala login)
    - ⏱️ **+Waktu** (Membuka modal tambah waktu individu)
    - 📥 **Force Submit** (Membuka konfirmasi submit paksa)
    - 📋 **Log Pelanggaran** (Membuka modal rekaman waktu cheat)

- [ ] **2.3 Paginasi Konsisten (*Pagination Footer*)**:
  - Tampilkan ringkasan data: `Menampilkan X - Y dari Z siswa`.
  - Tombol navigasi halaman (*Previous*, Nomor Halaman `1, 2, 3...`, *Next*).

- [ ] **2.4 Modal Tambah Waktu (`ExtendTimeModal`)**:
  - Pilihan cepat preset: `+5 Menit`, `+10 Menit`, `+15 Menit`, `+30 Menit`, atau input custom.
  - Form input alasan penambahan waktu (untuk catatan berita acara).

- [ ] **2.5 Modal Riwayat Pelanggaran (`ViolationLogModal`)**:
  - Menampilkan linimasa rekaman waktu dan tipe pelanggaran siswa secara terperinci.

---

## 🔍 3. LANGKAH PENGUJIAN & VERIFIKASI (VERIFICATION CHECKLIST)
1. **Uji Reset Login**:
   - Login sebagai siswa di satu browser. Coba login di browser lain (harus dicegah karena sesi aktif).
   - Dari dashboard pengawas, klik **Reset Login**. Siswa harus bisa login di browser baru.
2. **Uji Buka Blokir**:
   - Simulasikan siswa melanggar fullscreen/pindah tab hingga terblokir (`BLOCKED`).
   - Dari dashboard pengawas, klik **Buka Blokir**. Siswa harus dapat melanjutkan pengerjaan.
3. **Uji Tambah Waktu**:
   - Tambahkan waktu +10 menit pada siswa tertentu. Periksa apakah sisa waktu di layar siswa bertambah 10 menit.
   - Tambahkan waktu +15 menit ke seluruh kelas. Periksa seluruh siswa aktif mendapatkan kompensasi waktu.
4. **Uji Force Submit**:
   - Klik **Force Submit** pada siswa yang sedang mengerjakan. Sesi siswa harus langsung tertutup dan skor PG dihitung.
5. **Uji Filter, Sort & Paginasi**:
   - Ketik nama siswa di search bar (tabel harus menyaring seketika).
   - Pilih filter status `Terblokir` (hanya siswa terblokir yang tampil).
   - Urutkan berdasarkan `Pelanggaran Terbanyak` (siswa paling banyak melanggar di posisi atas).
   - Ubah per-page ke `10` dan navigasikan halaman.
6. **Build & Deploy Integrity**:
   - `cd frontend && npm run build` (Exit code: 0)
   - `cd backend && go build ./cmd/api` (Exit code: 0)
   - `docker compose up -d --build` (Semua container status Healthy)
