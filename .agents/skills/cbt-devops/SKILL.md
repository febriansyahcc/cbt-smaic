---
name: cbt-devops
description: Operational procedures for Docker container management, multi-service compose orchestration, database seeding, environment configuration, and local server startup.
---

# CBT DevOps & Container Orchestration

Gunakan skill ini untuk menjalankan, membangun ulang, mendebug, atau menghentikan layanan Docker dan lingkungan pengembangan lokal CBT.

---

## 🐳 1. Menjalankan Sistem dengan Docker Compose

### Struktur Layanan:
- **`postgres`:** Database PostgreSQL 16 (Port: `5432`).
- **`backend`:** Layanan Go Fiber API (Port: `8080`).
- **`frontend`:** Web Server Nginx + Frontend Vue 3 (Port: `80`).

### Perintah Utama:
1. **Memulai Seluruh Layanan (Build & Run Background):**
   ```powershell
   .\docker-start.ps1
   # atau
   docker compose up -d --build
   ```
2. **Melihat Status Kontainer:**
   ```powershell
   docker compose ps
   ```
3. **Melihat Log Layanan:**
   ```powershell
   docker compose logs -f backend
   # atau seluruh layanan
   docker compose logs -f
   ```
4. **Menghentikan Layanan:**
   ```powershell
   .\docker-stop.ps1
   # atau
   docker compose down
   ```

---

## 💻 2. Menjalankan Mode Pengembangan Lokal (Non-Docker Dev)

Jika ingin mengembangkan langsung di mesin host tanpa Docker:
1. Jalankan skrip pembantu:
   ```powershell
   .\start-dev.ps1
   ```
2. Atau jalankan secara terpisah di dua terminal:
   - **Terminal 1 (Backend Go):**
     ```powershell
     cd backend
     go run ./cmd/api
     ```
   - **Terminal 2 (Frontend Vite):**
     ```powershell
     cd frontend
     npm run dev
     ```

---

## 🔑 3. Verifikasi Akun & Seed Data Bawaan

| Akun | Username | Password | Role |
|---|---|---|---|
| Admin Kurikulum | `admin` | `admin123` | Administrator |
| Guru / Proktor | `guru1` | `guru123` | Guru / Pengawas |
| Siswa 1 | `siswa1` | `siswa123` | Siswa (XII MIPA 1) |
| Siswa 2 | `siswa2` | `siswa123` | Siswa (XII MIPA 1) |
| Siswa 3 | `siswa3` | `siswa123` | Siswa (XII MIPA 1) |

- **Token Ruang Ujian Bawaan:** `CBT2026`
