---
name: cbt-reporting
description: Guide and procedures for generating official exam reports, including Excel score recaps with excelize stream writing and PDF Berita Acara documents with gofpdf.
---

# CBT Reporting Engine (Excel & PDF Generation)

Gunakan skill ini saat membuat, menguji, atau memodifikasi fitur pelaporan rekap nilai ujian dan pencetakan dokumen resmi Berita Acara Ujian.

---

## 📊 1. Rekap Nilai Excel (.xlsx) dengan `excelize`

Lokasi: `backend/pkg/excel/`

### Karakteristik & Standar:
1. **Low Memory Overhead:** Gunakan stream writer (`f.NewStreamWriter(sheetName)`) saat menghasilkan rekap nilai untuk ratusan siswa serentak agar konsumsi RAM tetap di bawah 100MB.
2. **Kop & Header Resmi:**
   - Judul Dokumen: `REKAPITULASI HASIL UJIAN BERBASIS KOMPUTER (CBT)`
   - Identitas Sekolah: `SMAS ISLAMIC CENTRE DEMAK`
   - Mata Pelajaran, Kelas, Sesi Ujian, dan Tanggal Pelaksanaan.
3. **Kolom Data Standar:**
   - No, NIS, NISN, Nama Siswa, Kelas, Jumlah Benar, Jumlah Salah, Nilai Akhir, Status Kelulusan (Tuntas / Belum Tuntas).

### Pengujian Paket Excel:
```powershell
cd backend
go test -v ./pkg/excel/...
```

---

## 📑 2. Berita Acara Ujian PDF (.pdf) dengan `gofpdf`

Lokasi: `backend/pkg/pdf/`

### Karakteristik & Standar:
1. **Tata Letak & Tipografi:**
   - Format kertas A4 standar.
   - Header Dokumen memuat Kop Surat Resmi SMAS Islamic Centre Demak.
2. **Komponen Wajib Berita Acara:**
   - Hari, tanggal, dan rentang waktu pelaksanaan ujian.
   - Ruang ujian dan nama pengawas / proktor yang bertugas.
   - Statistik kehadiran: Jumlah siswa terdaftar, hadir, dan tidak hadir (beserta alasan/keterangan).
   - Kolom tanda tangan resmi Pengawas / Proktor dan Kepala Sekolah / Ketua Panitia Ujian.

### Pengujian Paket PDF:
```powershell
cd backend
go test -v ./pkg/pdf/...
```
