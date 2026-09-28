package service

import (
	"archive/tar"
	"compress/gzip"
	"errors"
	"io"
	"os"
	"path/filepath"
	"testing"
	"time"

	"cbt-backend/internal/domain"
	"cbt-backend/internal/repository"

	"github.com/glebarez/sqlite"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// closeDB menutup koneksi *sql.DB milik GORM saat tes selesai. Wajib dipanggil
// setelah t.TempDir() supaya cleanup-nya berjalan sebelum direktori sementara
// dihapus, sebab Windows menolak menghapus berkas database yang masih terbuka.
func closeDB(t *testing.T, db *gorm.DB) {
	t.Helper()
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("gagal mengambil *sql.DB: %v", err)
	}
	t.Cleanup(func() {
		if err := sqlDB.Close(); err != nil {
			t.Errorf("gagal menutup koneksi database: %v", err)
		}
	})
}

func newBackupFixture(t *testing.T) (*BackupService, *gorm.DB, string) {
	t.Helper()
	root := t.TempDir()
	db, err := gorm.Open(sqlite.Open(filepath.Join(root, "cbt.db")), &gorm.Config{})
	if err != nil {
		t.Fatalf("gagal membuka sqlite: %v", err)
	}
	// Didaftarkan setelah t.TempDir() agar berjalan lebih dulu (cleanup LIFO):
	// di Windows berkas cbt.db tidak bisa dihapus selama handle-nya masih terbuka.
	closeDB(t, db)
	if err := db.AutoMigrate(&domain.User{}, &domain.ClassRoom{}, &domain.ExamSchedule{}, &domain.ExamSession{}); err != nil {
		t.Fatalf("auto-migrate gagal: %v", err)
	}
	if err := db.Create(&domain.User{ID: uuid.New(), Username: "admin", PasswordHash: "x", FullName: "Admin", Role: domain.RoleAdmin, IsActive: true}).Error; err != nil {
		t.Fatalf("gagal membuat user: %v", err)
	}

	uploads := filepath.Join(root, "uploads", "questions")
	if err := os.MkdirAll(uploads, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(uploads, "soal-1.png"), []byte("gambar"), 0o644); err != nil {
		t.Fatal(err)
	}

	s := &BackupService{
		repo:       &repository.Database{DB: db},
		dir:        filepath.Join(root, "backups"),
		uploadsDir: filepath.Join(root, "uploads"),
		keep:       2,
		autoOn:     true,
		autoHour:   1,
		now:        time.Now,
	}
	return s, db, root
}

func archiveEntries(t *testing.T, path string) map[string][]byte {
	t.Helper()
	f, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	gz, err := gzip.NewReader(f)
	if err != nil {
		t.Fatalf("arsip bukan gzip: %v", err)
	}
	tr := tar.NewReader(gz)
	out := map[string][]byte{}
	for {
		h, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatalf("arsip tar rusak: %v", err)
		}
		b, _ := io.ReadAll(tr)
		out[h.Name] = b
	}
	return out
}

func TestBackupCreateSQLiteArchive(t *testing.T) {
	s, _, _ := newBackupFixture(t)

	item, err := s.Create(BackupKindManual)
	if err != nil {
		t.Fatalf("Create gagal: %v", err)
	}
	if item.Kind != BackupKindManual || item.Size == 0 {
		t.Fatalf("item tidak sesuai: %+v", item)
	}

	path, err := s.Path(item.Name)
	if err != nil {
		t.Fatalf("Path gagal: %v", err)
	}
	entries := archiveEntries(t, path)
	for _, want := range []string{"manifest.json", "database.sqlite", "uploads/questions/soal-1.png"} {
		if _, ok := entries[want]; !ok {
			t.Errorf("arsip tidak memuat %s", want)
		}
	}
	if string(entries["uploads/questions/soal-1.png"]) != "gambar" {
		t.Errorf("isi berkas unggahan berubah")
	}

	// Salinan database harus bisa dibuka dan berisi data.
	restored := filepath.Join(t.TempDir(), "restored.sqlite")
	if err := os.WriteFile(restored, entries["database.sqlite"], 0o600); err != nil {
		t.Fatal(err)
	}
	rdb, err := gorm.Open(sqlite.Open(restored), &gorm.Config{})
	if err != nil {
		t.Fatalf("salinan database tidak bisa dibuka: %v", err)
	}
	closeDB(t, rdb)
	var count int64
	rdb.Model(&domain.User{}).Count(&count)
	if count != 1 {
		t.Errorf("jumlah user di salinan = %d, ingin 1", count)
	}

	// Folder kerja sementara harus sudah dibersihkan.
	leftovers, _ := filepath.Glob(filepath.Join(s.dir, ".work-*"))
	if len(leftovers) != 0 {
		t.Errorf("folder kerja tertinggal: %v", leftovers)
	}
}

func TestBackupPruneKeepsNewestPerKind(t *testing.T) {
	s, _, _ := newBackupFixture(t)
	base := time.Date(2026, 9, 1, 2, 0, 0, 0, time.Local)
	for i := 0; i < 4; i++ {
		ts := base.Add(time.Duration(i) * time.Hour)
		s.now = func() time.Time { return ts }
		if _, err := s.Create(BackupKindAuto); err != nil {
			t.Fatalf("Create auto #%d gagal: %v", i, err)
		}
	}
	s.now = func() time.Time { return base.Add(10 * time.Hour) }
	if _, err := s.Create(BackupKindManual); err != nil {
		t.Fatal(err)
	}

	items, err := s.List()
	if err != nil {
		t.Fatal(err)
	}
	auto, manual := 0, 0
	for _, it := range items {
		if it.Kind == BackupKindAuto {
			auto++
		} else {
			manual++
		}
	}
	if auto != 2 || manual != 1 {
		t.Fatalf("auto=%d manual=%d, ingin auto=2 manual=1 (%v)", auto, manual, items)
	}
	if items[0].Kind != BackupKindManual || items[1].CreatedAt.Hour() != 5 {
		t.Errorf("urutan atau backup yang tersisa salah: %v", items)
	}
}

func TestBackupPathRejectsTraversal(t *testing.T) {
	s, _, _ := newBackupFixture(t)
	for _, name := range []string{"../cbt.db", "..%2Fcbt.db", "cbt-backup-20260101-010101-auto.tar.gz/../../x", "passwd", ""} {
		if _, err := s.Path(name); !errors.Is(err, ErrBackupBadName) {
			t.Errorf("Path(%q) = %v, ingin ErrBackupBadName", name, err)
		}
	}
	if _, err := s.Path("cbt-backup-20260101-010101-auto.tar.gz"); !errors.Is(err, ErrBackupNotFound) {
		t.Errorf("nama valid yang tidak ada harus ErrBackupNotFound, dapat %v", err)
	}
}

func TestBackupRejectsConcurrentRun(t *testing.T) {
	s, _, _ := newBackupFixture(t)
	if !s.begin() {
		t.Fatal("begin pertama harus berhasil")
	}
	if err := s.StartAsync(BackupKindManual); !errors.Is(err, ErrBackupRunning) {
		t.Fatalf("StartAsync saat berjalan = %v, ingin ErrBackupRunning", err)
	}
	s.finish(nil)
}

func TestAutoBackupWaitsForExamAndHour(t *testing.T) {
	s, db, _ := newBackupFixture(t)
	at := time.Date(2026, 9, 2, 0, 30, 0, 0, time.Local)
	s.now = func() time.Time { return at }

	s.TickAuto()
	if items, _ := s.List(); len(items) != 0 {
		t.Fatalf("backup otomatis tidak boleh jalan sebelum BACKUP_HOUR")
	}

	at = time.Date(2026, 9, 2, 1, 5, 0, 0, time.Local)
	session := domain.ExamSession{
		ID: uuid.New(), ScheduleID: uuid.New(), StudentID: uuid.New(), StartedAt: at.Add(-10 * time.Minute),
		ServerDeadline: at.Add(50 * time.Minute), Status: domain.StatusInProgress,
	}
	if err := db.Create(&session).Error; err != nil {
		t.Fatal(err)
	}
	s.TickAuto()
	if items, _ := s.List(); len(items) != 0 {
		t.Fatalf("backup otomatis tidak boleh jalan saat siswa sedang mengerjakan")
	}

	db.Model(&session).Update("status", domain.StatusSubmitted)
	s.TickAuto()
	items, _ := s.List()
	if len(items) != 1 || items[0].Kind != BackupKindAuto {
		t.Fatalf("backup otomatis harus jalan setelah ujian selesai, dapat %v", items)
	}

	s.TickAuto()
	if items, _ := s.List(); len(items) != 1 {
		t.Fatalf("backup otomatis hanya sekali per hari, dapat %d", len(items))
	}
}

func TestSplitPostgresPassword(t *testing.T) {
	cases := []struct{ in, conn, pass string }{
		{"postgres://cbt_user:rahasia@postgres:5432/cbt_db?sslmode=disable", "postgres://cbt_user@postgres:5432/cbt_db?sslmode=disable", "rahasia"},
		{"postgres://cbt_user:p%40ss@db/cbt", "postgres://cbt_user@db/cbt", "p@ss"},
		{"host=db user=cbt password=rahasia dbname=cbt", "host=db user=cbt dbname=cbt", "rahasia"},
	}
	for _, c := range cases {
		conn, pass, err := splitPostgresPassword(c.in)
		if err != nil || conn != c.conn || pass != c.pass {
			t.Errorf("splitPostgresPassword(%q) = %q, %q, %v; ingin %q, %q", c.in, conn, pass, err, c.conn, c.pass)
		}
	}
	if _, _, err := splitPostgresPassword(""); err == nil {
		t.Error("DSN kosong harus galat")
	}
}
