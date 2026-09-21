# PROMPT / SPESIFIKASI PENGEMBANGAN FITUR CBT SYSTEM
## 🎯 Fitur: Simulator Siswa, KaTeX Formula Matematika, Tipografi Arab & Korea, serta Media Gambar pada Pertanyaan & Pilihan Jawaban

> **Target Repository**: `CBT System (Go Fiber + Vue 3 + GORM + Tailwind CSS)`  
> **Institusi**: SMAS Islamic Centre Demak  
> **Tujuan**: Implementasi fitur Pratinjau Ujian Siswa (Simulator), rendering konten kaya (KaTeX LaTeX, Font Arab Amiri berharakat, Hangul Korea Noto Sans), serta integrasi upload/paste gambar di soal dan opsi jawaban (A-E).

---

## 🏛️ Konteks & Pedoman Proyek (Agentic Guidelines)
1. **Zero Native Browser Dialogs**: Jangan gunakan `alert()` / `confirm()`. Gunakan modal reaktif (`showConfirmModal`, `showAlertModal`, `showToast`).
2. **Clean Layered Architecture**:
   - Backend Go di `backend/internal/handler/`, `backend/internal/domain/`, `backend/cmd/api/main.go`.
   - Frontend Vue di `frontend/src/views/`, `frontend/src/components/`, `frontend/src/stores/`.
3. **Verifikasi Wajib**:
   - Frontend build: `npm run build` di folder `frontend` (harus exit code 0).
   - Backend build: `go build ./cmd/api` di folder `backend` (harus exit code 0).
   - Docker: `docker compose up -d --build`.

---

## 📋 DAFTAR TUGAS IMPLEMENTASI (TASK BREAKDOWN)

### TUGAS 1: Backend Media Upload & Static Serving
- [ ] **1.1 Directory Storage**: Pastikan direktori `./uploads/questions` otomatis dibuat jika belum ada.
- [ ] **1.2 Handler Upload Gambar**:
  - Lokasi: `backend/internal/handler/handlers.go`
  - Buat handler `HandleUploadImage(c *fiber.Ctx) error`:
    - Mengambil file dari `c.FormFile("image")`.
    - Validasi format MIME: `.jpg`, `.jpeg`, `.png`, `.webp`, `.gif`, `.svg` (Maks. 5 MB).
    - Simpan dengan nama unik: `fmt.Sprintf("%s%s", uuid.New().String(), ext)`.
    - Return JSON response: `{"success": true, "url": "/uploads/questions/" + filename}`.
- [ ] **1.3 Static Route & Endpoint**:
  - Lokasi: `backend/cmd/api/main.go`
  - Daftarkan route:
    - Static file serve: `app.Static("/uploads", "./uploads")`
    - Route upload: `admin.Post("/upload-image", handlers.HandleUploadImage)`
- [ ] **1.4 Dockerfile Volume/Directory**:
  - Pastikan folder `uploads` ada di container atau di-mount.

---

### TUGAS 2: Frontend Rich Content Renderer (KaTeX + Arab + Korea + Gambar)
- [ ] **2.1 Dependensi & Web Fonts**:
  - Pasang library `katex`: `npm install katex` di `frontend/`.
  - Di `frontend/index.html`:
    - Tambahkan stylesheet KaTeX:
      ```html
      <link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/katex@0.16.11/dist/katex.min.css">
      ```
    - Tambahkan Google Fonts untuk **Amiri** (Arab) dan **Noto Sans KR** (Korea):
      ```html
      <link rel="preconnect" href="https://fonts.googleapis.com">
      <link rel="preconnect" href="https://fonts.gstatic.com" crossorigin>
      <link href="https://fonts.googleapis.com/css2?family=Amiri:ital,wght@0,400;0,700;1,400;1,700&family=Noto+Sans+KR:wght@400;600;700&display=swap" rel="stylesheet">
      ```
  - Di `frontend/src/style.css`:
    - Tambahkan utility classes:
      ```css
      .font-arabic {
        font-family: 'Amiri', serif;
        direction: rtl;
        text-align: right;
        line-height: 1.8;
      }
      .font-korean {
        font-family: 'Noto Sans KR', sans-serif;
      }
      ```
- [ ] **2.2 Komponen Universal `RichContentRenderer.vue`**:
  - Lokasi: `frontend/src/components/common/RichContentRenderer.vue`
  - **Fungsi**:
    - Parsing regex untuk formula LaTeX inline `$...$` dan display `$$...$$` menggunakan `katex.renderToString()`.
    - Auto-detect teks beraksara Arab (`/[\u0600-\u06FF]/`) dan secara otomatis membungkus teks tersebut dalam tag dengan class `font-arabic text-lg`.
    - Auto-detect karakter Hangul Korea (`/[\uAC00-\uD7AF\u1100-\u11FF]/`) dengan class `font-korean`.
    - Render HTML konten aman (termasuk tag `<img>`).
    - Dukungan klik gambar untuk membuka lightbox preview perbesaran gambar (*zoomable lightbox*).

---

### TUGAS 3: Dukungan Media Gambar pada Editor Soal & Opsi (A-E)
- [ ] **3.1 Image Upload pada Form Soal di `AdminDashboardView.vue`**:
  - Input gambar pada pertanyaan:
    - Tombol "Unggah Gambar Pertanyaan", drag-and-drop file area, dan listener `paste` (Ctrl+V screenshot).
    - Otomatis upload ke `POST /api/v1/admin/upload-image` dan memasukkan tag `<img>` atau menyimpan `image_url`.
  - Input gambar pada setiap pilihan jawaban (A, B, C, D, E):
    - Tambahkan tombol upload gambar / paste di samping setiap input opsi text.
    - Simpan field `image_url` pada object opsi: `{ key: 'A', text: '...', image_url: '...' }`.
- [ ] **3.2 Render Gambar pada Kartu Soal & Layar Siswa**:
  - Tampilkan gambar opsi di samping atau di bawah teks opsi jika `opt.image_url` terisi.
  - Komponen siswa `ExamQuestionCard.vue` diupdate agar menampilkan gambar pada opsi dan pertanyaan menggunakan `<RichContentRenderer />`.

---

### TUGAS 4: Fitur Simulator Tampilan Siswa (*Student Exam Simulator*)
- [ ] **4.1 Komponen Simulator `StudentExamSimulatorModal.vue`**:
  - Lokasi: `frontend/src/components/admin/StudentExamSimulatorModal.vue`
  - **Fitur Simulator**:
    - Mode Switcher:
      - 📱 **Mobile Device Mode (375px x 667px)**: Menampilkan frame smartphone interaktif lengkap dengan 3-zone layout (Header timer mockup, Question card scrollable, Bottom navigation sheet & drawer).
      - 💻 **Desktop Mode (Layar Penuh)**: Menampilkan antarmuka desktop dengan sidebar daftar nomor soal di sebelah kanan.
    - Interaktivitas Real-Time:
      - Navigasi nomor soal: Sebelumnya, Selanjutnya, Ragu-ragu, dan Lompat Nomor.
      - Klik opsi jawaban A/B/C/D/E (state tersimpan sementara di local state komponen simulator).
      - Indikator status warna nomor: Abu-abu (Belum dijawab), Hijau (Sudah dijawab), Kuning (Ragu-ragu).
- [ ] **4.2 Integrasi Tombol di `AdminDashboardView.vue`**:
  - Tambahkan tombol `"👁️ Pratinjau Siswa"` pada modal Bank Soal (di sebelah tombol "Tambah Soal").
  - Klik tombol membuka modal simulator dengan data `bankQuestions` yang sedang aktif.

---

## 🔍 5. LANGKAH PENGUJIAN & VERIFIKASI (VERIFICATION CHECKLIST)
1. **Formula Matematika**:
   - Masukkan rumus `$\frac{-b \pm \sqrt{b^2 - 4ac}}{2a}$` dan `$$\int_{0}^{\pi} \sin(x) dx$$`.
   - Formula harus tampil sebagai simbol vektor tajam dan presisi.
2. **Teks Arab & Harakat**:
   - Masukkan ayat Al-Qur'an: `بِسْمِ اللَّهِ الرَّحْمَٰنِ الرَّحِيمِ`.
   - Teks harus tampil dari kanan ke kiri (*RTL*), ukuran huruf proporsional, dan harakat terbaca jelas.
3. **Teks Korea**:
   - Masukkan teks: `안녕하세요. SMAS Islamic Centre Demak`.
   - Glyph Hangul tampil rapi dengan font Noto Sans KR.
4. **Media Gambar**:
   - Paste screenshot (Ctrl+V) dan upload file gambar ke soal serta opsi A/B.
   - Gambar tersimpan di server dan tampil sempurna.
5. **Simulator Siswa**:
   - Buka simulator, uji toggle Mobile/Desktop, klik pilihan jawaban, tandai ragu-ragu, dan navigasi nomor soal.
6. **Build Integrity**:
   - `cd frontend && npm run build` (Exit code: 0)
   - `cd backend && go build ./cmd/api` (Exit code: 0)
   - `docker compose up -d --build` (Semua container healthy)
