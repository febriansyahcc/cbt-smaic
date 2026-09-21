# Standar Pengembangan Backend (Go & Fiber v2)

Pedoman teknis untuk pengembangan, refaktorisasi, dan pemeliharaan backend CBT SMAS Islamic Centre Demak.

---

## 1. Struktur Arsitektur Bersih (Clean Layered Pattern)
Semua logika backend ditempatkan pada layer yang sesuai di folder `backend/internal/`:

- **`domain/` (Models & Schemas):**
  - Deklarasi struct model GORM, tag JSON, tag validasi, dan DTO request/response.
  - Bebas dari ketergantungan library eksternal selain tag data.
- **`repository/` (Akses Basis Data):**
  - Seluruh query GORM dan manipulasi DB terpusat di sini.
  - Gunakan transaksi (`db.Transaction(func(tx *gorm.DB) error)`) untuk operasi multi-tabel.
- **`service/` (Logika Bisnis):**
  - Memproses aturan bisnis: otentikasi JWT, validasi token ujian, penghitungan nilai otomatis, verifikasi anti-cheat, dll.
  - Tidak berinteraksi langsung dengan Fiber Context (`*fiber.Ctx`).
- **`handler/` (HTTP Controller):**
  - Parsing body JSON / query params / URL params dari `*fiber.Ctx`.
  - Panggil service yang bersangkutan dan kembalikan JSON standar:
    ```json
    {
      "success": true,
      "message": "Deskripsi respon",
      "data": {}
    }
    ```
  - Format respon error standar:
    ```json
    {
      "success": false,
      "message": "Pesan kesalahan yang deskriptif"
    }
    ```

---

## 2. Integritas Data & Concurrency
1. **Authoritative Server Timer:**
   - Sisa waktu pengerjaan dihitung dari `start_time + duration_minutes` di sisi server.
   - Jangan pernah mempercayai waktu yang dikirim oleh client/browser.
2. **Idempotent Answer Synchronization:**
   - Penyimpanan jawaban siswa harus menggunakan operasi `UPSERT` (on conflict do update) agar request ganda dari background sync tidak menduplikasi data.
3. **Deterministik Seeded Randomization:**
   - Pengacakan urutan soal dan pilihan jawaban harus menggunakan fungsi PRNG berbasis seed `(StudentID + ScheduleID)` dari package `pkg/prng`.
4. **Export Engines (`pkg/excel` & `pkg/pdf`):**
   - Penulisan berkas Excel menggunakan `excelize` stream writer untuk efisiensi memori (<100MB RAM untuk 200+ siswa).
   - Penulisan PDF Berita Acara menggunakan `gofpdf` dengan tata letak resmi SMAS Islamic Centre Demak.

---

## 3. Verifikasi & Build
Setiap perubahan kode backend wajib diverifikasi dengan perintah:
```powershell
cd backend
go build ./cmd/api
```
