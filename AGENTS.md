# CBT System - Agentic Operating Guidelines & System Context

Selamat datang di repositori **CBT System (Computer-Based Test Engine)** untuk **SMAS Islamic Centre Demak**.
Dokumen ini adalah pedoman utama bagi AI Agent saat berinteraksi, mengembangkan, mendebug, atau memelihara kode di seluruh repositori ini.

---

## 🏛️ Identitas & Konteks Proyek

- **Nama Institusi:** SMAS ISLAMIC CENTRE DEMAK
- **Yayasan:** Yayasan Islamic Centre Sultan Fatah Demak
- **Aset Logo Resmi:** `/logo-smic.png` (Logo hijau keemasan)
- **Target Pengguna:** ±200 siswa serentak (High Concurrency, Low Latency, BYOD Smartphone & Komputer Sekolah)

---

## 🏗️ Arsitektur & Teknologi

1. **Backend (Go 1.22+ & Fiber v2):**
   - Pola: Clean Layered Architecture (`domain` → `repository` → `service` → `handler`).
   - ORM & Database: GORM dengan PostgreSQL 16 (dan fallback SQLite `cbt.db` untuk mode dev lokal).
   - Authoritative Server Timer: Waktu ujian dihitung dari deadline server, bukan jam lokal klien.
   - Seeded Randomization: Pengacakan urutan soal & opsi deterministik berbasis seed `student_id + schedule_id`.
   - Idempotent Sync: Penyimpanan jawaban via operasi `UPSERT` yang aman terhadap duplikasi request jaringan.
   - Reporting: `excelize` (Stream Writer untuk Excel Rekap Nilai) & `gofpdf` (Dokumen Berita Acara Ujian PDF).

2. **Frontend (Vue 3 + Vite + Tailwind CSS):**
   - Pola: Composition API (`<script setup>`), Pinia state management, Vue Router.
   - Mobile-First 3-Zone Layout: Fixed Topbar, Scrollable Card, Fixed Bottombar & Bottom-sheet Drawer.
   - Anti-Cheat: Multi-layer detection (Page Visibility API, window blur, fullscreen change, text selection / right click prevention).
   - Offline-Tolerant: Local caching (`localStorage`) + background sync queue saat jaringan tersendat.

3. **DevOps & Deployment:**
   - Multi-container Docker Compose (`cbt-backend`, `cbt-frontend`, `cbt-db`).
   - Skrip Otomasi Windows: `docker-start.ps1`, `docker-start.bat`, `start-dev.ps1`.

---

## 🛡️ Aturan Utama untuk AI Agent (Agentic Rules)

1. **Zero Native Browser Dialogs:**
   - **DILARANG** menggunakan `alert()` atau `confirm()` bawaan browser.
   - Wajib menggunakan modal custom reaktif (`showConfirmModal`, `showAlertModal`, `showToast`).

2. **Animasi Pop-up Modal yang Bersih:**
   - Backdrop (overlay gelap) **hanya animasi fade opacity murni** (`opacity-0` → `opacity-100`). **DILARANG efek scale/zoom pada backdrop**.
   - Kartu modal menggunakan scale & translate halus (`scale-95 translate-y-2` → `scale-100 translate-y-0`).

3. **Prinsip Desain Minimalis & Action-Oriented:**
   - Hindari badge berulang atau emoji berlebih.
   - Gabungkan informasi yang terkait secara logis.
   - Sediakan tombol tindakan yang jelas dan responsif (`active:scale-95`).

4. **Integritas Kode & Verifikasi:**
   - Setiap perubahan backend wajib diverifikasi dengan kompilasi Go (`go build ./cmd/api`).
   - Setiap perubahan frontend wajib diverifikasi dengan build Vite (`npm run build` di folder `frontend`).
   - Selalu pertahankan kompatibilitas data dan skema database yang ada.
