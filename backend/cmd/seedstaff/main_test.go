package main

import (
	"bytes"
	"encoding/csv"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"cbt-backend/internal/domain"
	"cbt-backend/internal/staffseed"

	"github.com/glebarez/sqlite"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

func newDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("gagal membuka sqlite memori: %v", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("gagal mengambil sql.DB: %v", err)
	}
	sqlDB.SetMaxOpenConns(1)
	if err := db.AutoMigrate(&domain.User{}, &domain.ClassRoom{}, &domain.Subject{}, &domain.ClassSubject{}); err != nil {
		t.Fatalf("auto-migrate gagal: %v", err)
	}
	return db
}

func seedDemo(t *testing.T, db *gorm.DB) {
	t.Helper()
	now := time.Now()
	for _, v := range []any{
		&domain.ClassRoom{ID: uuid.New(), Name: "X 1", Grade: "X", Major: "UMUM", CreatedAt: now},
		&domain.ClassRoom{ID: uuid.New(), Name: "XII MIPA 1", Grade: "XII", Major: "MIPA", CreatedAt: now},
		&domain.Subject{ID: uuid.New(), Code: "BIN-WJB-XII", Name: "Bahasa Indonesia Wajib XII", CreatedAt: now},
		&domain.User{ID: uuid.New(), Username: "admin", PasswordHash: "hash-admin", FullName: "Administrator", Role: domain.RoleAdmin, Permissions: []string{"*"}, IsActive: true, CreatedAt: now, UpdatedAt: now},
		&domain.User{ID: uuid.New(), Username: "guru1", PasswordHash: "hash-guru1", FullName: "Guru Demo", Role: domain.RoleGuru, Permissions: domain.TemplatePermissions("guru"), IsActive: true, CreatedAt: now, UpdatedAt: now},
	} {
		if err := db.Create(v).Error; err != nil {
			t.Fatalf("gagal menyemai data demo: %v", err)
		}
	}
}

func rowCount(t *testing.T, db *gorm.DB, model any) int64 {
	t.Helper()
	var n int64
	if err := db.Model(model).Count(&n).Error; err != nil {
		t.Fatal(err)
	}
	return n
}

func mustLoadEmbedded(t *testing.T) staffseed.Data {
	t.Helper()
	d, err := staffseed.LoadData(bytes.NewReader(defaultData))
	if err != nil {
		t.Fatalf("data embed tidak valid: %v", err)
	}
	return d
}

// (a) Data embed valid dan menghasilkan 19 guru, 24 mapel, kelas baru 5 (X 1 sudah ada) atau 6,
// dan 124 penugasan (62 kombinasi guru-mapel-tingkat x 2 kelas).
func TestDataEmbedValidDanJumlahnya(t *testing.T) {
	d := mustLoadEmbedded(t)

	combos := 0
	for _, tc := range d.Teachers {
		for _, a := range tc.Assignments {
			combos += len(a.Grades)
		}
	}
	if len(d.Teachers) != 19 || len(d.Subjects) != 24 || len(d.Classes) != 6 || combos != 62 {
		t.Fatalf("isi data embed: guru=%d mapel=%d kelas=%d kombinasi=%d (harap 19/24/6/62)", len(d.Teachers), len(d.Subjects), len(d.Classes), combos)
	}

	t.Run("basis data kosong", func(t *testing.T) {
		rep, err := staffseed.Run(newDB(t), d, false, nil, nil)
		if err != nil {
			t.Fatal(err)
		}
		if len(rep.ClassesNew) != 6 || len(rep.ClassesExisting) != 0 {
			t.Fatalf("kelas: %d baru, %d ada (harap 6/0)", len(rep.ClassesNew), len(rep.ClassesExisting))
		}
	})

	t.Run("dengan data demo", func(t *testing.T) {
		db := newDB(t)
		seedDemo(t, db)
		rep, err := staffseed.Run(db, d, true, nil, nil)
		if err != nil {
			t.Fatal(err)
		}
		if len(rep.TeachersNew) != 19 || len(rep.TeachersUnchanged) != 0 || len(rep.TeachersUpdated) != 0 {
			t.Fatalf("guru: %d baru, %d ada, %d diubah (harap 19/0/0)", len(rep.TeachersNew), len(rep.TeachersUnchanged), len(rep.TeachersUpdated))
		}
		if len(rep.SubjectsNew) != 24 || len(rep.SubjectsExisting) != 0 {
			t.Fatalf("mapel: %d baru, %d ada (harap 24/0)", len(rep.SubjectsNew), len(rep.SubjectsExisting))
		}
		if len(rep.ClassesNew) != 5 || len(rep.ClassesExisting) != 1 || rep.ClassesExisting[0] != "X 1" {
			t.Fatalf("kelas: %v baru, %v ada (harap 5 baru, X 1 sudah ada)", rep.ClassesNew, rep.ClassesExisting)
		}
		if rep.AssignmentsCreated != 124 || rep.AssignmentsUnchanged != 0 || len(rep.AssignmentsChanged) != 0 {
			t.Fatalf("penugasan: %d dibuat, %d tetap, %d diubah (harap 124/0/0)", rep.AssignmentsCreated, rep.AssignmentsUnchanged, len(rep.AssignmentsChanged))
		}
		if len(rep.Credentials) != 19 || len(rep.Teachers) != 19 {
			t.Fatalf("kredensial=%d laporan guru=%d (harap 19/19)", len(rep.Credentials), len(rep.Teachers))
		}
		if got := rowCount(t, db, &domain.ClassSubject{}); got != 124 {
			t.Fatalf("baris class_subjects = %d, harap 124", got)
		}
		if got := rowCount(t, db, &domain.User{}); got != 2+19 {
			t.Fatalf("baris users = %d, harap 21", got)
		}
	})
}

// runCLI menjalankan run() dengan DB sqlite memori dan DSN palsu berkata sandi.
func runCLI(db *gorm.DB, args ...string) (code int, stdout, stderr string) {
	var out, errOut bytes.Buffer
	getenv := func(k string) string {
		if k == "DB_DSN" {
			return "postgres://cbt:S3cretDsnPw@cbt-db:5432/cbt?sslmode=disable"
		}
		return ""
	}
	open := func(string) (*gorm.DB, error) { return db, nil }
	code = run(args, &out, &errOut, getenv, open)
	return code, out.String(), errOut.String()
}

func TestCLIDryRunStdoutKosongDanTidakMenulis(t *testing.T) {
	db := newDB(t)
	seedDemo(t, db)
	users := rowCount(t, db, &domain.User{})

	code, stdout, stderr := runCLI(db)
	if code != 0 {
		t.Fatalf("exit %d, stderr:\n%s", code, stderr)
	}
	if stdout != "" {
		t.Fatalf("stdout dry-run harus kosong, dapat: %q", stdout)
	}
	if !strings.Contains(stderr, "DRY-RUN") || !strings.Contains(stderr, "19 baru") {
		t.Fatalf("laporan dry-run tidak sesuai:\n%s", stderr)
	}
	if strings.Contains(stderr, "S3cretDsnPw") || !strings.Contains(stderr, "REDACTED") {
		t.Fatalf("DSN tidak disamarkan pada stderr:\n%s", stderr)
	}
	if rowCount(t, db, &domain.User{}) != users || rowCount(t, db, &domain.ClassSubject{}) != 0 {
		t.Fatal("dry-run menulis ke basis data")
	}
}

// (j) stdout CSV hanya berisi akun baru; tidak ada kata sandi di stderr.
func TestCLIApplyCSVHanyaAkunBaruDanStderrTanpaKataSandi(t *testing.T) {
	db := newDB(t)
	seedDemo(t, db)
	// Satu guru dari berkas sudah ada: tidak boleh muncul di CSV.
	if err := db.Create(&domain.User{ID: uuid.New(), Username: "azizah", PasswordHash: "HASH-LAMA", FullName: "Azizah, S.Pd., Gr.", Role: domain.RoleGuru,
		Permissions: domain.TemplatePermissions("guru"), IsActive: true, CreatedAt: time.Now(), UpdatedAt: time.Now()}).Error; err != nil {
		t.Fatal(err)
	}

	code, stdout, stderr := runCLI(db, "--apply")
	if code != 0 {
		t.Fatalf("exit %d, stderr:\n%s", code, stderr)
	}
	rows, err := csv.NewReader(strings.NewReader(stdout)).ReadAll()
	if err != nil {
		t.Fatalf("stdout bukan CSV valid: %v\n%s", err, stdout)
	}
	if strings.Join(rows[0], ",") != "username,nama_lengkap,password" {
		t.Fatalf("header CSV: %v", rows[0])
	}
	if len(rows) != 1+18 {
		t.Fatalf("baris CSV = %d, harap 18 akun baru + header", len(rows))
	}
	for _, r := range rows[1:] {
		if r[0] == "azizah" || r[0] == "admin" || r[0] == "guru1" {
			t.Fatalf("akun lama muncul di CSV: %v", r)
		}
		if len(r[2]) != 10 {
			t.Fatalf("panjang kata sandi %q: %d", r[0], len(r[2]))
		}
		if strings.Contains(stderr, r[2]) {
			t.Fatalf("kata sandi %s muncul di stderr", r[0])
		}
	}
	if strings.Contains(stderr, "S3cretDsnPw") {
		t.Fatal("kata sandi DSN muncul di stderr")
	}
	if !strings.Contains(stderr, "APPLY") {
		t.Fatalf("stderr tidak menyebut APPLY:\n%s", stderr)
	}
	if got := rowCount(t, db, &domain.ClassSubject{}); got != 124 {
		t.Fatalf("class_subjects = %d, harap 124", got)
	}

	// Apply kedua: nol perubahan dan stdout kosong (tidak ada akun/kata sandi baru).
	code, stdout, stderr = runCLI(db, "--apply")
	if code != 0 || stdout != "" {
		t.Fatalf("apply kedua: exit %d, stdout %q, stderr:\n%s", code, stdout, stderr)
	}
	if !strings.Contains(stderr, "0 dibuat, 124 tidak berubah, 0 diubah") {
		t.Fatalf("apply kedua seharusnya melaporkan 124 tidak berubah:\n%s", stderr)
	}
}

func TestCLIGalatMenghasilkanExitNonNolDanStdoutKosong(t *testing.T) {
	// Berkas data dengan pasangan (mapel, tingkat) ganda antar guru.
	dir := t.TempDir()
	bad := filepath.Join(dir, "bad.json")
	content := `{"academic_year":"2026/2027","classes":[{"name":"X 1","grade":"X"}],"subjects":[{"code":"MAT","name":"Matematika"}],
	"teachers":[{"username":"a","full_name":"A","assignments":[{"subject":"MAT","grades":["X"]}]},
	            {"username":"b","full_name":"B","assignments":[{"subject":"MAT","grades":["X"]}]}]}`
	if err := os.WriteFile(bad, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	db := newDB(t)
	code, stdout, stderr := runCLI(db, "--file", bad, "--apply")
	if code == 0 || stdout != "" || !strings.Contains(stderr, "dimiliki dua guru") {
		t.Fatalf("exit %d, stdout %q, stderr:\n%s", code, stdout, stderr)
	}
	if rowCount(t, db, &domain.User{}) != 0 {
		t.Fatal("ada akun tertulis padahal data tidak valid")
	}

	// Username milik non-GURU: galat validasi dari basis data, tanpa CSV.
	db = newDB(t)
	seedDemo(t, db)
	good := filepath.Join(dir, "admin.json")
	content = `{"academic_year":"2026/2027","classes":[{"name":"X 1","grade":"X"}],"subjects":[{"code":"MAT","name":"Matematika"}],
	"teachers":[{"username":"admin","full_name":"Bukan Guru","assignments":[{"subject":"MAT","grades":["X"]}]}]}`
	if err := os.WriteFile(good, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	code, stdout, stderr = runCLI(db, "--file", good, "--apply")
	if code == 0 || stdout != "" || !strings.Contains(stderr, "bukan GURU") {
		t.Fatalf("exit %d, stdout %q, stderr:\n%s", code, stdout, stderr)
	}
}

func TestCLIFileMenimpaDataBawaan(t *testing.T) {
	path := filepath.Join(t.TempDir(), "kecil.json")
	content := `{"academic_year":"2026/2027","classes":[{"name":"X 1","grade":"X"}],"subjects":[{"code":"MAT","name":"Matematika"}],
	"teachers":[{"username":"budi","full_name":"Budi","assignments":[{"subject":"MAT","grades":["X"]}]}]}`
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	db := newDB(t)
	code, stdout, stderr := runCLI(db, "--file", path, "--apply")
	if code != 0 {
		t.Fatalf("exit %d, stderr:\n%s", code, stderr)
	}
	rows, _ := csv.NewReader(strings.NewReader(stdout)).ReadAll()
	if len(rows) != 2 || rows[1][0] != "budi" {
		t.Fatalf("CSV: %v", rows)
	}
	if rowCount(t, db, &domain.User{}) != 1 {
		t.Fatal("data bawaan ikut terpakai padahal --file diberikan")
	}
}

func TestCLIArgumenDanEnvBermasalah(t *testing.T) {
	db := newDB(t)

	if code, _, _ := runCLI(db, "--tidak-ada"); code == 0 {
		t.Error("flag tak dikenal seharusnya exit non-nol")
	}
	if code, _, _ := runCLI(db, "sisa"); code == 0 {
		t.Error("argumen posisi seharusnya exit non-nol")
	}
	if code, _, _ := runCLI(db, "--file", filepath.Join(t.TempDir(), "tidak-ada.json")); code == 0 {
		t.Error("berkas tidak ada seharusnya exit non-nol")
	}

	var out, errOut bytes.Buffer
	code := run(nil, &out, &errOut, func(string) string { return "" }, func(string) (*gorm.DB, error) { return db, nil })
	if code == 0 || !strings.Contains(errOut.String(), "DB_DSN") {
		t.Errorf("DB_DSN kosong seharusnya exit non-nol: %d %s", code, errOut.String())
	}

	// Galat koneksi tidak boleh membocorkan kata sandi DSN.
	errOut.Reset()
	code = run(nil, &out, &errOut,
		func(string) string { return "host=db user=cbt password=S3cretDsnPw dbname=cbt" },
		func(dsn string) (*gorm.DB, error) { return nil, errors.New("dial gagal untuk " + dsn) })
	if code == 0 || strings.Contains(errOut.String(), "S3cretDsnPw") {
		t.Errorf("galat koneksi: exit %d, stderr:\n%s", code, errOut.String())
	}
}
