---
name: cbt-feature-development
description: End-to-end workflow for adding or modifying fullstack features in the CBT application (Go Fiber backend + Vue 3 frontend + GORM database).
---

# CBT Fullstack Feature Development Workflow

Gunakan skill ini saat diminta untuk menambahkan modul atau fitur baru di aplikasi CBT SMAS Islamic Centre Demak.

---

## Tahapan Eksekusi (Workflow Steps)

### 1. Perancangan Model & Database (`backend/internal/domain/`)
1. Definisikan struct entitas baru di `backend/internal/domain/models.go` lengkap dengan tag GORM (`gorm:"..."`) dan JSON (`json:"..."`).
2. Pastikan relasi foreign key dan constraint indeks dideklarasikan dengan benar.
3. Tambahkan struct Request DTO & Response DTO jika payload berbeda dengan model basis data.

### 2. Implementasi Layer Repository (`backend/internal/repository/`)
1. Buat method baru di file repository terkait (atau tambahkan di `backend/internal/repository/db.go`).
2. Gunakan query GORM yang efisien (`Find`, `First`, `Where`, `Preload`).
3. Jika melibatkan operasi tulis multi-tabel, bungkus dalam transaksi GORM.

### 3. Implementasi Layer Service (`backend/internal/service/`)
1. Buat business logic service di `backend/internal/service/`.
2. Lakukan validasi input, verifikasi otorisasi/role, dan perhitungan bisnis.
3. Kembalikan error yang jelas jika terjadi kegagalan validasi.

### 4. Implementasi Layer Handler & Routing (`backend/internal/handler/`)
1. Tambahkan HTTP controller handler di `backend/internal/handler/handlers.go`.
2. Lakukan parsing request body `c.BodyParser(&req)` dan parsing query/param.
3. Kembalikan response JSON standar `{ success: true, message: "...", data: ... }`.
4. Daftarkan endpoint route baru di `backend/cmd/api/main.go` dengan middleware auth/role yang sesuai.

### 5. Implementasi Frontend Service & Views (`frontend/src/`)
1. Daftarkan fungsi pemanggilan endpoint di `frontend/src/services/api.js` (atau service terkait).
2. Buat / perbarui state store di `frontend/src/stores/` jika data dibutuhkan secara global.
3. Bangun komponen antarmuka di `frontend/src/views/` atau `frontend/src/components/` mengikuti aturan UI/UX (Zero native alert, custom modal, clean minimalist layout).

### 6. Verifikasi & Pengujian
Jalankan verifikasi backend dan frontend:
```powershell
# Verifikasi Backend Go
cd backend; go build ./cmd/api; cd ..

# Verifikasi Frontend Vue
cd frontend; npm run build; cd ..
```
