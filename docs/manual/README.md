# Buku Panduan CBT SMAS Islamic Centre Demak

Buku panduan dibagi per peran pengguna. Setiap panduan tersedia sebagai HTML (sumber) dan PDF siap cetak A4.

| Panduan | Untuk | PDF |
|---|---|---|
| Panduan Siswa | Peserta ujian | [pdf/panduan-siswa.pdf](pdf/panduan-siswa.pdf) |
| Panduan Pengawas | Pengawas ruang / proktor | [pdf/panduan-pengawas.pdf](pdf/panduan-pengawas.pdf) |
| Panduan Guru | Guru pengampu penyusun soal | [pdf/panduan-guru.pdf](pdf/panduan-guru.pdf) |
| Panduan Admin Kurikulum | Admin kurikulum | [pdf/panduan-admin.pdf](pdf/panduan-admin.pdf) |
| Panduan Teknis | Operator IT / server | [pdf/panduan-teknis.pdf](pdf/panduan-teknis.pdf) |

## Struktur folder

- `panduan-*.html` — isi panduan. Buka langsung di browser untuk pratinjau.
- `assets/manual.css` — gaya bersama (warna aplikasi dan logo sekolah).
- `img/` — screenshot aplikasi per peran, diambil dari aplikasi yang berjalan dengan data contoh.
- `build/` — skrip pembuat PDF.
- `pdf/` — hasil cetak.

## Membuat ulang PDF

Butuh Node.js 18+ dan browser Chromium untuk Playwright.

```bash
cd docs/manual/build
npm install
npx playwright install chromium   # sekali saja, bila Chromium belum ada
node build.mjs                    # semua panduan
node build.mjs panduan-siswa.html # satu panduan
```

Sampul dicetak tanpa kop; halaman isi diberi kop dan nomor halaman otomatis.

## Memperbarui isi

Saat fitur aplikasi berubah, perbarui teks di HTML dan ganti screenshot terkait di `img/`, lalu buat ulang PDF. Nama tombol dan menu ditulis persis seperti di aplikasi agar mudah dicocokkan pembaca.
