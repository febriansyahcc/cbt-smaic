# STANDAR DESAIN UI / UX DAN ATURAN SISTEM CBT
> Diadopsi dari penyempurnaan Menu Jadwal Ujian SMAS Islamic Centre Demak.
> Berlaku untuk seluruh halaman, menu, modal, dan komponen di seluruh aplikasi CBT.

---

## 1. Zero Native Browser Dialogs (Dilarang Memakai alert() & confirm())
- **Wajib Reaktif:** Tidak boleh ada pemanggilan native `window.alert()` maupun `window.confirm()`.
- **Standar Konfirmasi (`showConfirmModal`):**
  - Mengembalikan `Promise<boolean>` yang dapat di-`await`.
  - Tipe aksi: `danger` (merah untuk hapus), `warning` (kuning untuk reset/lepas kunci), atau `confirm` (biru/indigo).
- **Standar Peringatan/Validasi (`showAlertModal`):**
  - Untuk error validasi form, peringatan status, atau instruksi langkah berikutnya.
- **Standar Notifikasi Ringan (`showToast`):**
  - Untuk umpan balik sukses non-blocking (contoh: *Data berhasil disimpan*), hilang otomatis setelah 3 detik.

---

## 2. Standar Animasi Pop-up Modal (Clean Separated Transitions)
- **Backdrop (Overlay Gelap):**
  - **Hanya animasi fade opacity murni** (`opacity-0` -> `opacity-100`).
  - **Dilarang keras memberikan efek scale/zoom pada backdrop** agar latar belakang tidak terlihat membesar atau bergeser.
  - Backdrop blur ringan (`backdrop-blur-xs` atau `backdrop-blur-sm`).
- **Kartu Dialog (Modal Card):**
  - Menggunakan animasi scale & translate lembut (`scale-95 translate-y-2` -> `scale-100 translate-y-0`).
- **Estetika Kartu:**
  - Latar belakang putih bersih (*clean white*), sudut membulat modern (`rounded-3xl`), bayangan halus (`shadow-2xl`).
  - **Tanpa garis aksen/warna di bagian atas kartu.**
  - Tombol penutup silang (✕) di pojok kanan atas.
  - Tombol aksi di footer menggunakan efek responsif `active:scale-95`.

---

## 3. Prinsip Tampilan Minimalis & Efisiensi Informasi (Minimalist Information Design)
1. **Penggabungan Data yang Terkait (Logical Data Grouping):**
   - Gabungkan data yang berkaitan erat ke dalam satu kolom/komponen (contoh: Mata Pelajaran disatukan dengan Nama Sesi & Waktu).
2. **Eliminasi Informasi Redundan:**
   - Hilangkan label atau subtitle yang berulang atau sudah jelas dari konteks (contoh: menghilangkan label Tingkat jika nama kelas sudah spesifik).
3. **Status Ringkas & Berorientasi Aksi (Action-Oriented Status):**
   - Hindari badge status yang menumpuk atau membingungkan.
   - Tampilkan status yang langsung membuka tindakan pengguna (contoh: jika sudah ada soal cukup tampilkan *Ganti Soal*; jika belum ada tampilkan *Belum Ada Soal & Tautkan Soal*).
4. **Pembersihan Ikon & Emoji Berlebih:**
   - Hilangkan ikon dekoratif/emoji yang tidak perlu di header modal atau tabel agar antarmuka terlihat profesional, bersih, dan tidak melelahkan mata.
5. **Aksi Footer Esensial:**
   - Footer modal hanya menyajikan tombol tindakan utama dan tombol tutup (contoh: hanya tombol Edit dan Tutup).

---

## 4. Standar Identitas Resmi Sekolah
- **Nama Sekolah:** SMAS ISLAMIC CENTRE DEMAK
- **Yayasan:** Yayasan Islamic Centre Sultan Fatah Demak
- **Logo Resmi:** `/logo-smic.png` (Logo Kemenag/Yayasan hijau keemasan)
