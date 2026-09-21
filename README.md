# CBT System - Computer-Based Test Engine (High School Level)

Sistem Ujian Berbasis Komputer & Smartphone (BYOD) yang dirancang khusus untuk memenuhi standar keandalan tinggi pada lingkungan sekolah menengah (±200 siswa serentak).

Containerized dan siap di-deploy langsung menggunakan **Docker & Docker Compose**.

---

## 🚀 Fitur Utama

- 🐳 **Full Containerized Architecture:** Multi-kontainer terisolasi (PostgreSQL 16, Go Backend, Nginx Frontend Reverse Proxy).
- ⚡ **High Concurrency & Low Latency:** Backend Go (Fiber v2) dengan alokasi memori minimal (<100MB RAM) dan respon instan.
- 📱 **Mobile-First 3-Zone Layout:** Antarmuka Vue 3 + Tailwind CSS yang dioptimalkan untuk jempol siswa di smartphone (Fixed Topbar, Scrollable Question Card, Fixed Bottombar & Bottom-sheet Drawer).
- 🛡️ **Anti-Cheat Multi-Layer:** Deteksi Page Visibility API, event Blur (notifikasi/split-screen), deteksi fullscreen, penonaktifan seleksi teks dan klik kanan, serta kuota pelanggaran otomatis (*Auto-Lock*).
- 🔄 **Dual-State Offline-Tolerant Sync:** Penyimpanan instan di memori lokal (`localStorage`/`IndexedDB`) dengan antrean sinkronisasi latar belakang dan operasi basis data yang *idempotent* (`UPSERT`).
- 🎲 **Seeded Randomization:** Pengacakan urutan soal dan opsi berbasis seed `student_id + schedule_id` yang konsisten antar-refresh namun unik antar-siswa.
- ⏱️ **Authoritative Server Timer:** Penghitung sisa waktu berbasis deadline server untuk mencegah manipulasi jam lokal perangkat.
- 📊 **Live Proctoring Dashboard:** Pemantauan progres pengerjaan seluruh siswa di kelas secara real-time dengan tombol buka kunci (*unlock*) dan reset sesi perangkat.
- 📑 **Dual Reporting Engine:** Ekspor Rekap Nilai ke Excel (.xlsx) dengan stream writer `excelize` dan pembuatan dokumen resmi Berita Acara Ujian ke PDF (.pdf) dengan `gofpdf`.

---

## 🐳 Panduan Cepat Menjalankan dengan Docker

### 1. Prasyarat
Pastikan **Docker Desktop** (untuk Windows/Mac) atau Docker Engine (Linux) sudah terpasang dan sedang berjalan.

### 2. Jalankan Sistem
Cukup jalankan satu perintah dari root folder proyek:

**Opsi A (PowerShell):**
```powershell
.\docker-start.ps1
```

**Opsi B (Command Prompt / Double-Click):**
```cmd
docker-start.bat
```

**Opsi C (Docker CLI langsung):**
```bash
docker compose up -d --build
```

Setelah proses build selesai, aplikasi langsung aktif dan dapat diakses di:
👉 **`http://localhost`** (atau IP lokal server/komputer host)

### 3. Menghentikan Kontainer
```bash
docker compose down
# atau klik ganda: docker-stop.bat
```

---

## ⚙️ Konfigurasi Environment (`.env`)

Sistem menggunakan file `.env` di root direktori untuk kustomisasi konfigurasi:

```env
APP_PORT=80
BACKEND_PORT=8080
DB_HOST=postgres
DB_PORT=5432
DB_USER=cbt_user
DB_PASSWORD=cbt_secret_password_2026
DB_NAME=cbt_db
TZ=Asia/Jakarta
JWT_SECRET=<isi dengan hasil: openssl rand -hex 32>
```

---

## 🔑 Akun Percobaan Bawaan (Database Seed)

| Peran | Username | Password | Keterangan |
|---|---|---|---|
| **Siswa 1** | `siswa1` | `siswa123` | Ahmad Fauzi (NIS: `1001`, XII MIPA 1) |
| **Siswa 2** | `siswa2` | `siswa123` | Siti Nurhaliza (NIS: `1002`, XII MIPA 1) |
| **Siswa 3** | `siswa3` | `siswa123` | Budi Wicaksono (NIS: `1003`, XII MIPA 1) |
| **Guru / Proktor** | `guru1` | `guru123` | Drs. Bambang Sudarsono, M.Pd |
| **Admin Kurikulum** | `admin` | `admin123` | Administrator Utama |

* **Token Ruang Ujian Aktif:** `CBT2026`

Dokumentasi lengkap operasional: [docs/PANDUAN_OPERASIONAL.md](docs/PANDUAN_OPERASIONAL.md).