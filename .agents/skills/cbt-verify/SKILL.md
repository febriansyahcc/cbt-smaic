---
name: cbt-verify
description: Comprehensive verification runbook to check Go backend compilation, Vue frontend build, unit tests, and system health in the CBT codebase.
---

# CBT Verification & Quality Assurance Runbook

Gunakan skill ini untuk memverifikasi kesehatan kode sebelum menyelesaikan tugas atau commit perubahan.

---

## 📋 Checklist Verifikasi

### 1. Verifikasi Kompilasi Backend Go
Jalankan kompilasi API backend dan pastikan tidak ada error kompilasi:
```powershell
cd backend
go build -v ./cmd/api
```

### 2. Eksekusi Unit Tests Backend
Jalankan seluruh unit test pada sub-paket utility (Excel, PDF, PRNG):
```powershell
cd backend
go test -v ./pkg/...
```

### 3. Verifikasi Build Frontend Vue 3
Jalankan proses bundling Vite frontend dan pastikan tidak ada sintaks/impor yang rusak:
```powershell
cd frontend
npm run build
```

### 4. Audit Kepatuhan UI / UX (Manual / Code Review)
Pastikan hal-hal berikut dipatuhi pada setiap komponen baru:
- [ ] Tidak ada pemanggilan `window.alert()` atau `window.confirm()`.
- [ ] Menggunakan modal reaktif (`showConfirmModal`, `showAlertModal`, `showToast`).
- [ ] Animasi modal bersih (backdrop fade only tanpa zoom, kartu modal skala/translate).
- [ ] Identitas sekolah menggunakan SMAS Islamic Centre Demak dan logo resmi `/logo-smic.png`.
- [ ] Tombol aksi interaktif memiliki class `active:scale-95`.
