# Panduan Operasional & Arsitektur Sistem CBT (Computer-Based Test)
SMA / Sekolah Menengah Level

Dokumen ini berisi panduan lengkap pengoperasian sistem CBT untuk Siswa, Pengawas Ruang / Proktor, Guru Mata Pelajaran, dan Administrator Kurikulum.

---

## 1. Kredensial Akun Default (Demo & Pengujian)

Sistem telah dilengkapi dengan *database seed* otomatis yang siap digunakan:

| Peran | Username | Password | Keterangan / Identitas |
|---|---|---|---|
| **Siswa 1** | `siswa1` | `siswa123` | Ahmad Fauzi (NIS: `1001`, Kelas XII MIPA 1) |
| **Siswa 2** | `siswa2` | `siswa123` | Siti Nurhaliza (NIS: `1002`, Kelas XII MIPA 1) |
| **Siswa 3** | `siswa3` | `siswa123` | Budi Wicaksono (NIS: `1003`, Kelas XII MIPA 1) |
| **Guru / Proktor** | `guru1` | `guru123` | Drs. Bambang Sudarsono, M.Pd (Pengawas Ruang) |
| **Kurikulum / Admin** | `admin` | `admin123` | Administrator Utama Kurikulum |

* **Token Ruang Ujian Aktif:** `CBT2026`
* **Mata Pelajaran:** Matematika Wajib (10 Soal Pilihan Ganda)

---

## 2. Fitur & Alur Kerja Siswa

1. **Login & Single-Device Lock:**
   * Siswa login menggunakan NIS atau Username dan Password.
   * Satu akun hanya diperbolehkan login pada satu perangkat aktif. Jika login di HP lain tanpa reset oleh proktor, sesi sebelumnya akan otomatis terputus.
2. **Mulai Ujian:**
   * Memilih jadwal ujian aktif dan memasukkan token `CBT2026`.
   * Sistem otomatis memuat seluruh paket soal ke `localStorage` (Dual-State Resilience) dan mengaktifkan mode fullscreen.
3. **Pengerjaan Soal (3-Zone Mobile Layout):**
   * **Zone 1 (Topbar):** Timer waktu server monospace tabular, nama mata pelajaran, dan indikator status sinkronisasi.
   * **Zone 2 (Card Soal):** Teks soal dengan tombol pengatur ukuran font (`A-`, `A`, `A+`) dan pilihan jawaban (A–E) berbentuk kartu sentuh ramah jempol (*thumb-friendly*, min 48px).
   * **Zone 3 (Bottombar):** Tombol *Sebelumnya*, *Ragu-ragu* (amber), *Kisi Soal* (bottom-sheet drawer), dan *Berikutnya*.
4. **Resiliensi Jaringan (Offline-to-Online):**
   * Setiap jawaban langsung tersimpan seketika di memori lokal.
   * Sinkronisasi ke server berjalan di latar belakang secara *idempotent* (`UPSERT`).
   * Jika sinyal Wi-Fi putus/drop, jawaban tetap tersimpan dan otomatis disinkronkan saat koneksi kembali stabil.
5. **Anti-Cheat Smartphone:**
   * Deteksi otomatis jika siswa berpindah aplikasi (*tab switch*), membuka notifikasi (*blur*), atau keluar dari layar penuh.
   * Pelanggaran 1 & 2: Muncul dialog peringatan merah.
   * Pelanggaran 3 (Kuota Habis): Ujian terkunci total (*Auto-Lock*). Siswa wajib memanggil pengawas untuk dibuka kembali.
6. **Pengumpulan Ujian (Submit Ganda):**
   * Konfirmasi ganda menampilkan rincian: jumlah soal terjawab, ragu-ragu, dan belum terisi.
   * Sesaat setelah submit, nilai PG langsung dihitung otomatis (*Auto-Grading*).

---

## 3. Fitur & Alur Kerja Pengawas / Proktor

1. **Live Proctoring Dashboard (`/proctor`):**
   * Pemantauan real-time status seluruh siswa dalam ruang ujian (diperbarui otomatis tiap 4 detik).
   * Statistik ringkas: Total Peserta, Sedang Mengerjakan, Sudah Selesai, dan Terkunci.
   * *Progress Bar* persentase pengerjaan soal tiap siswa.
2. **Buka Kunci Pelanggaran (*Unlock*):**
   * Tombol satu klik **"Buka Kunci"** untuk membuka blokir siswa yang terkunci akibat kuota tab switch.
3. **Reset Sesi Perangkat (*Device Reset*):**
   * Jika HP siswa mati kehabisan baterai atau rusak, pengawas dapat menekan **"Reset HP"** agar siswa dapat segera login dari perangkat cadangan di laboratorium.
4. **Ekspor Nilai Excel (.xlsx):**
   * Klik tombol **"Rekap Nilai (.xlsx)"** untuk mengunduh rekap nilai lengkap (NIS, NISN, Nama, Kelas, Jam Mulai, Jam Selesai, Jumlah Pelanggaran, Nilai Akhir) yang dihasilkan dengan engine stream `excelize`.
5. **Cetak Berita Acara Resmi PDF (.pdf):**
   * Klik tombol **"Berita Acara (.pdf)"** untuk mengunduh dokumen formal Berita Acara Ujian ber-Kop resmi dinas pendidikan lengkap dengan catatan insiden dan kolom tanda tangan pengawas.

---

## 4. Fitur Guru & Kurikulum

1. **Unduh Template Soal Excel:**
   * Klik **"Unduh Template Excel Soal"** untuk mendapatkan format resmi 9 kolom: `No`, `Pertanyaan`, `Opsi A-E`, `Kunci (A-E)`, `Bobot`.
2. **Impor Bank Soal Instan:**
   * Pilih bank soal target dan unggah file `.xlsx`. Sistem Go akan memvalidasi dan memproses butir soal dalam hitungan milidetik.
3. **Readiness Matrix:**
   * Memantau kesiapan soal guru mata pelajaran dan status penguncian (*Lock*) sebelum ujian dilaksanakan.

---

## 5. Cara Menjalankan Sistem Secara Lokal

### Cara A: Menjalankan Native (Go + Node.js)

1. **Jalankan Backend (Terminal 1):**
   ```powershell
   cd d:\cbt\backend
   .\cbt-backend.exe
   ```
   *Backend berjalan pada `http://localhost:8080` (menggunakan SQLite otomatis jika PostgreSQL lokal belum diisi DSN).*

2. **Jalankan Frontend (Terminal 2):**
   ```powershell
   cd d:\cbt\frontend
   npm run dev
   ```
   *Buka browser di `http://localhost:5173`.*

### Cara B: Menjalankan dengan Docker Compose

```bash
cd d:\cbt\deploy
docker compose up -d --build
```
*Aplikasi siap diakses di port 80.*
---

## 6. CI/CD & Update Otomatis Server

### Alur rilis

1. Push ke branch `main` menjalankan workflow `.github/workflows/ci-cd.yml`: `go vet`, `go test`, build backend, dan build frontend.
2. Jika semua lulus, image `mochamaddwifebriansyah13/cbt-backend` dan `cbt-frontend` di-push ke Docker Hub dengan tag `latest` dan `sha-<commit>`.
3. Push tag `v1.2.3` menghasilkan tag image `1.2.3` dan `1.2` untuk keperluan rollback.
4. Pull request ke `main` hanya menjalankan tes, tanpa push image.

Secret yang wajib diisi di GitHub (**Settings → Secrets and variables → Actions**):

| Nama | Isi |
|------|-----|
| `DOCKERHUB_USERNAME` | Username Docker Hub |
| `DOCKERHUB_TOKEN` | Personal access token Docker Hub dengan izin *Read & Write* |

### Keamanan image (tanda tangan cosign)

Setiap image yang dibuat workflow ditandatangani dengan [cosign](https://github.com/sigstore/cosign) secara *keyless*: tanda tangan terikat pada identitas workflow `ci-cd.yml` di branch `main` repo `febriansyahcc/cbt-smaic`, tanpa kunci privat yang perlu disimpan. Skrip auto-update hanya memasang image yang lolos verifikasi ini. Kalau token Docker Hub bocor dan ada orang yang mem-push image lain, server menolaknya dan menulis `PERINGATAN: tanda tangan ... tidak valid` di log.

Pasang cosign di server (sekali saja):

```bash
curl -fsSLO https://github.com/sigstore/cosign/releases/download/v3.1.3/cosign-linux-amd64
sudo install -m 755 cosign-linux-amd64 /usr/local/bin/cosign
cosign version
```

Server perlu akses HTTPS keluar ke `registry-1.docker.io`, `auth.docker.io`, `tuf-repo-cdn.sigstore.dev`, dan `rekor.sigstore.dev`. Tanpa akses ke Sigstore, verifikasi gagal dan update tidak dipasang (aman, tetapi server tidak ter-update).

### Update otomatis di server

Skrip `deploy/auto-update.sh` memeriksa Docker Hub, memverifikasi tanda tangan image baru, lalu pull (berdasarkan digest yang sudah diverifikasi) dan restart container hanya jika image berubah. Update **ditunda** selama ada jadwal ujian aktif atau sesi siswa yang belum selesai, supaya ujian tidak terputus.

Pasang di cron server (contoh folder `/opt/cbt/deploy`, cek setiap 10 menit):

```bash
chmod +x /opt/cbt/deploy/auto-update.sh
sudo crontab -e
# tambahkan baris:
*/10 * * * * /opt/cbt/deploy/auto-update.sh >> /var/log/cbt-auto-update.log 2>&1
```

Perintah yang berguna:

```bash
tail -f /var/log/cbt-auto-update.log            # lihat riwayat update
touch /opt/cbt/deploy/.update-paused            # jeda update (mis. minggu ujian)
rm /opt/cbt/deploy/.update-paused               # aktifkan lagi
FORCE=1 /opt/cbt/deploy/auto-update.sh          # update sekarang tanpa cek ujian
```

Rollback ke versi tertentu: ubah `image:` di `deploy/docker-compose.yml` ke tag versi (mis. `cbt-backend:1.2.3`), buat `.update-paused`, lalu `docker compose up -d`.
