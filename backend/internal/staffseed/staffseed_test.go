package staffseed

import (
	"bytes"
	"encoding/csv"
	"errors"
	"math/rand"
	"strings"
	"testing"
	"time"

	"cbt-backend/internal/domain"

	"github.com/glebarez/sqlite"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

const year = "2026/2027"

var fixedNow = func() time.Time { return time.Date(2026, 7, 1, 8, 0, 0, 0, time.UTC) }

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
	sqlDB.SetMaxOpenConns(1) // basis data memori harus memakai satu koneksi
	if err := db.AutoMigrate(&domain.User{}, &domain.ClassRoom{}, &domain.Subject{}, &domain.ClassSubject{}); err != nil {
		t.Fatalf("auto-migrate gagal: %v", err)
	}
	return db
}

// smallData: 3 kelas (X 1, X 2, XI 1), 3 mapel, 2 guru. Kombinasi guru-mapel-tingkat: 3
// (budi MAT X, budi MAT XI, sari BIN X) -> 2 + 1 + 2 = 5 penugasan kelas-mapel.
func smallData() Data {
	return Data{
		AcademicYear: year,
		Classes: []ClassData{
			{Name: "X 1", Grade: "X"}, {Name: "X 2", Grade: "X"}, {Name: "XI 1", Grade: "XI"},
		},
		Subjects: []SubjectData{
			{Code: "MAT", Name: "Matematika"}, {Code: "BIN", Name: "Bahasa Indonesia"}, {Code: "FIS", Name: "Fisika"},
		},
		Teachers: []TeacherData{
			{Username: "budi", FullName: "Budi, S.Pd.", Assignments: []AssignmentData{{Subject: "MAT", Grades: []string{"X", "XI"}}}},
			{Username: "sari", FullName: "Sari, M.Pd.", Assignments: []AssignmentData{{Subject: "BIN", Grades: []string{"X"}}}},
		},
	}
}

// demoIDs menyimpan id data demo yang harus tetap utuh.
type demoIDs struct {
	classX1, classMIPA, subjBIN, subjMAT, admin, guru1, classSubject uuid.UUID
}

// seedDemo meniru data demo produksi: kelas X 1 dan XII MIPA 1, mapel BIN-WJB-XII dan
// MAT-WJB-XII, akun admin dan guru1, serta satu class_subject.
func seedDemo(t *testing.T, db *gorm.DB) demoIDs {
	t.Helper()
	now := time.Now()
	d := demoIDs{classX1: uuid.New(), classMIPA: uuid.New(), subjBIN: uuid.New(), subjMAT: uuid.New(), admin: uuid.New(), guru1: uuid.New(), classSubject: uuid.New()}
	must := func(v any) {
		t.Helper()
		if err := db.Create(v).Error; err != nil {
			t.Fatalf("gagal menyemai data demo: %v", err)
		}
	}
	must(&domain.ClassRoom{ID: d.classX1, Name: "X 1", Grade: "X", Major: "UMUM", CreatedAt: now})
	must(&domain.ClassRoom{ID: d.classMIPA, Name: "XII MIPA 1", Grade: "XII", Major: "MIPA", CreatedAt: now})
	must(&domain.Subject{ID: d.subjBIN, Code: "BIN-WJB-XII", Name: "Bahasa Indonesia Wajib XII", CreatedAt: now})
	must(&domain.Subject{ID: d.subjMAT, Code: "MAT-WJB-XII", Name: "Matematika Wajib XII", CreatedAt: now})
	must(&domain.User{ID: d.admin, Username: "admin", PasswordHash: "hash-admin", FullName: "Administrator", Role: domain.RoleAdmin, Permissions: []string{"*"}, IsActive: true, CreatedAt: now, UpdatedAt: now})
	must(&domain.User{ID: d.guru1, Username: "guru1", PasswordHash: "hash-guru1", FullName: "Guru Demo", Role: domain.RoleGuru, Permissions: domain.TemplatePermissions("guru"), IsActive: true, CreatedAt: now, UpdatedAt: now})
	must(&domain.ClassSubject{ID: d.classSubject, ClassRoomID: d.classMIPA, SubjectID: d.subjMAT, TeacherID: d.guru1, AcademicYear: year, CreatedAt: now})
	return d
}

type counts struct{ classes, subjects, users, classSubjects int64 }

func countRows(t *testing.T, db *gorm.DB) counts {
	t.Helper()
	var c counts
	for model, dst := range map[any]*int64{
		&domain.ClassRoom{}: &c.classes, &domain.Subject{}: &c.subjects,
		&domain.User{}: &c.users, &domain.ClassSubject{}: &c.classSubjects,
	} {
		if err := db.Model(model).Count(dst).Error; err != nil {
			t.Fatalf("gagal menghitung baris: %v", err)
		}
	}
	return c
}

func mustRun(t *testing.T, db *gorm.DB, d Data, apply bool) Report {
	t.Helper()
	rep, err := Run(db, d, apply, fixedNow, nil)
	if err != nil {
		t.Fatalf("Run(apply=%v) galat: %v", apply, err)
	}
	return rep
}

func userByName(t *testing.T, db *gorm.DB, username string) domain.User {
	t.Helper()
	var u domain.User
	if err := db.First(&u, "username = ?", username).Error; err != nil {
		t.Fatalf("akun %q tidak ditemukan: %v", username, err)
	}
	return u
}

func TestDryRunTidakMenulisApaPun(t *testing.T) {
	db := newDB(t)
	seedDemo(t, db)
	before := countRows(t, db)

	rep := mustRun(t, db, smallData(), false)
	if after := countRows(t, db); after != before {
		t.Fatalf("dry-run mengubah basis data: sebelum %+v, sesudah %+v", before, after)
	}
	if rep.Apply || rep.Committed {
		t.Fatalf("dry-run tidak boleh Apply/Committed: %+v", rep)
	}
	if len(rep.Credentials) != 0 {
		t.Fatal("dry-run tidak boleh menghasilkan kredensial")
	}
	// Hitungan dry-run harus sama dengan hasil apply.
	if len(rep.ClassesNew) != 2 || len(rep.ClassesExisting) != 1 || len(rep.SubjectsNew) != 3 ||
		len(rep.TeachersNew) != 2 || rep.AssignmentsCreated != 5 {
		t.Fatalf("hitungan dry-run tidak sesuai: %+v", rep)
	}
	applied := mustRun(t, db, smallData(), true)
	if len(applied.ClassesNew) != len(rep.ClassesNew) || len(applied.TeachersNew) != len(rep.TeachersNew) ||
		applied.AssignmentsCreated != rep.AssignmentsCreated {
		t.Fatalf("hitungan apply berbeda dari dry-run: dry %+v vs apply %+v", rep, applied)
	}
}

func TestApplyDuaKaliIdempotent(t *testing.T) {
	db := newDB(t)
	first := mustRun(t, db, smallData(), true)
	if !first.Committed || len(first.Credentials) != 2 {
		t.Fatalf("apply pertama: %+v", first)
	}
	afterFirst := countRows(t, db)
	hashes := map[string]string{"budi": userByName(t, db, "budi").PasswordHash, "sari": userByName(t, db, "sari").PasswordHash}

	second := mustRun(t, db, smallData(), true)
	if !second.Committed {
		t.Fatal("apply kedua seharusnya tetap ter-commit")
	}
	if len(second.ClassesNew) != 0 || len(second.SubjectsNew) != 0 || len(second.TeachersNew) != 0 ||
		len(second.TeachersUpdated) != 0 || second.AssignmentsCreated != 0 || len(second.AssignmentsChanged) != 0 {
		t.Fatalf("apply kedua harus nol perubahan: %+v", second)
	}
	if second.AssignmentsUnchanged != 5 || len(second.TeachersUnchanged) != 2 {
		t.Fatalf("apply kedua: penugasan/guru tidak berubah seharusnya 5/2, dapat %d/%d", second.AssignmentsUnchanged, len(second.TeachersUnchanged))
	}
	if len(second.Credentials) != 0 {
		t.Fatal("apply kedua tidak boleh menghasilkan kredensial baru")
	}
	if countRows(t, db) != afterFirst {
		t.Fatal("jumlah baris berubah pada apply kedua")
	}
	for u, h := range hashes {
		if userByName(t, db, u).PasswordHash != h {
			t.Fatalf("hash kata sandi %s berubah pada apply kedua", u)
		}
	}
}

func TestGuruLamaTidakDiubahKecualiNama(t *testing.T) {
	db := newDB(t)
	old := domain.User{ID: uuid.New(), Username: "budi", PasswordHash: "HASH-LAMA", FullName: "Budi Lama", Role: domain.RoleGuru,
		Permissions: []string{"schedules:read"}, IsActive: true, CreatedAt: time.Now(), UpdatedAt: time.Now()}
	if err := db.Create(&old).Error; err != nil {
		t.Fatal(err)
	}

	rep := mustRun(t, db, smallData(), true)
	got := userByName(t, db, "budi")
	if got.ID != old.ID {
		t.Fatal("akun budi dibuat ulang, seharusnya dipakai yang lama")
	}
	if got.PasswordHash != "HASH-LAMA" {
		t.Fatal("kata sandi guru yang sudah ada berubah")
	}
	if got.Role != domain.RoleGuru || strings.Join(got.Permissions, ",") != "schedules:read" {
		t.Fatalf("peran/izin guru lama berubah: %s %v", got.Role, got.Permissions)
	}
	if got.FullName != "Budi, S.Pd." {
		t.Fatalf("nama lengkap tidak disamakan: %q", got.FullName)
	}
	if len(rep.TeachersUpdated) != 1 || rep.TeachersUpdated[0].OldName != "Budi Lama" || rep.TeachersUpdated[0].NewName != "Budi, S.Pd." {
		t.Fatalf("perubahan nama tidak tercatat: %+v", rep.TeachersUpdated)
	}
	if len(rep.TeachersNew) != 1 || rep.TeachersNew[0] != "sari" || len(rep.Credentials) != 1 || rep.Credentials[0].Username != "sari" {
		t.Fatalf("hanya sari yang baru: %+v", rep)
	}
	// Penugasan memakai id akun lama.
	var n int64
	db.Model(&domain.ClassSubject{}).Where("teacher_id = ?", old.ID).Count(&n)
	if n != 3 {
		t.Fatalf("penugasan budi seharusnya 3 baris, dapat %d", n)
	}
}

func TestGuruBaruIzinDanKataSandi(t *testing.T) {
	db := newDB(t)
	rep := mustRun(t, db, smallData(), true)
	want := domain.TemplatePermissions("guru")
	if len(want) == 0 {
		t.Fatal("template guru kosong")
	}
	for _, c := range rep.Credentials {
		u := userByName(t, db, c.Username)
		if u.Role != domain.RoleGuru || !u.IsActive {
			t.Fatalf("%s: role/aktif salah: %s %v", c.Username, u.Role, u.IsActive)
		}
		if len(u.Permissions) == 0 || strings.Join(u.Permissions, ",") != strings.Join(want, ",") {
			t.Fatalf("%s: izin %v bukan template guru %v", c.Username, u.Permissions, want)
		}
		if bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(c.Password)) != nil {
			t.Fatalf("%s: hash tidak cocok dengan kata sandi yang dilaporkan", c.Username)
		}
		if strings.Contains(u.PasswordHash, c.Password) {
			t.Fatalf("%s: kata sandi tersimpan sebagai teks biasa", c.Username)
		}
		if len(c.Password) != 10 {
			t.Fatalf("%s: panjang kata sandi %d, seharusnya 10", c.Username, len(c.Password))
		}
		if !u.CreatedAt.Equal(fixedNow()) {
			t.Fatalf("%s: CreatedAt tidak memakai now yang diberikan: %v", c.Username, u.CreatedAt)
		}
	}
}

func TestUsernameNonGuruDitolak(t *testing.T) {
	for _, role := range []domain.Role{domain.RoleAdmin, domain.RoleSiswa} {
		t.Run(string(role), func(t *testing.T) {
			db := newDB(t)
			other := domain.User{ID: uuid.New(), Username: "sari", PasswordHash: "H", FullName: "Bukan Guru", Role: role, IsActive: true, CreatedAt: time.Now(), UpdatedAt: time.Now()}
			if err := db.Create(&other).Error; err != nil {
				t.Fatal(err)
			}
			before := countRows(t, db)

			_, err := Run(db, smallData(), true, fixedNow, nil)
			var ve *ValidationError
			if !errors.As(err, &ve) || !strings.Contains(err.Error(), `"sari"`) || !strings.Contains(err.Error(), string(role)) {
				t.Fatalf("seharusnya galat validasi untuk sari (%s), dapat: %v", role, err)
			}
			if countRows(t, db) != before {
				t.Fatal("ada data tertulis padahal validasi gagal")
			}
			if got := userByName(t, db, "sari"); got.Role != role || got.FullName != "Bukan Guru" || got.PasswordHash != "H" {
				t.Fatal("akun non-GURU diubah")
			}
		})
	}
}

func TestKonflikNamaMapelDitolak(t *testing.T) {
	db := newDB(t)
	if err := db.Create(&domain.Subject{ID: uuid.New(), Code: "MAT", Name: "Matematika Peminatan", CreatedAt: time.Now()}).Error; err != nil {
		t.Fatal(err)
	}
	before := countRows(t, db)
	_, err := Run(db, smallData(), true, fixedNow, nil)
	var ve *ValidationError
	if !errors.As(err, &ve) || !strings.Contains(err.Error(), `"MAT"`) {
		t.Fatalf("seharusnya konflik mapel MAT, dapat: %v", err)
	}
	if countRows(t, db) != before {
		t.Fatal("ada data tertulis padahal konflik")
	}
	var s domain.Subject
	db.First(&s, "code = ?", "MAT")
	if s.Name != "Matematika Peminatan" {
		t.Fatal("nama mapel yang ada ditimpa")
	}
}

func TestMapelSamaNamaDilewati(t *testing.T) {
	db := newDB(t)
	id := uuid.New()
	if err := db.Create(&domain.Subject{ID: id, Code: "MAT", Name: "Matematika", CreatedAt: time.Now()}).Error; err != nil {
		t.Fatal(err)
	}
	rep := mustRun(t, db, smallData(), true)
	if len(rep.SubjectsExisting) != 1 || rep.SubjectsExisting[0] != "MAT" || len(rep.SubjectsNew) != 2 {
		t.Fatalf("mapel: %+v", rep)
	}
	var n int64
	db.Model(&domain.ClassSubject{}).Where("subject_id = ?", id).Count(&n)
	if n != 3 {
		t.Fatalf("penugasan MAT seharusnya memakai mapel lama (3 baris), dapat %d", n)
	}
}

func TestValidasiData(t *testing.T) {
	cases := []struct {
		name   string
		mutate func(d *Data)
		want   string
	}{
		{"mapel tak dikenal", func(d *Data) { d.Teachers[0].Assignments[0].Subject = "XYZ" }, `mapel "XYZ" tidak ada di daftar mapel`},
		{"tingkat tak valid", func(d *Data) { d.Teachers[0].Assignments[0].Grades = []string{"XIII"} }, `tingkat "XIII" tidak valid`},
		{"tingkat huruf kecil", func(d *Data) { d.Teachers[0].Assignments[0].Grades = []string{"x"} }, `tingkat "x" tidak valid`},
		{"grades kosong", func(d *Data) { d.Teachers[0].Assignments[0].Grades = nil }, "daftar grades kosong"},
		{"dua guru satu pasangan", func(d *Data) {
			d.Teachers[1].Assignments = append(d.Teachers[1].Assignments, AssignmentData{Subject: "MAT", Grades: []string{"X"}})
		}, `pasangan mapel MAT tingkat X dimiliki dua guru: "budi" dan "sari"`},
		{"pasangan ganda satu guru", func(d *Data) {
			d.Teachers[0].Assignments = append(d.Teachers[0].Assignments, AssignmentData{Subject: "MAT", Grades: []string{"X"}})
		}, "lebih dari sekali"},
		{"username kosong", func(d *Data) { d.Teachers[0].Username = "" }, "username tidak boleh kosong"},
		{"username huruf besar", func(d *Data) { d.Teachers[0].Username = "Budi" }, `username "Budi" tidak valid`},
		{"username berspasi", func(d *Data) { d.Teachers[0].Username = "bu di" }, `username "bu di" tidak valid`},
		{"username kembar", func(d *Data) { d.Teachers[1].Username = "budi" }, `username "budi" muncul lebih dari sekali`},
		{"kode mapel kembar", func(d *Data) { d.Subjects[1].Code = "MAT" }, `kode mapel "MAT" muncul lebih dari sekali`},
		{"kode mapel kosong", func(d *Data) { d.Subjects[0].Code = "" }, "code) tidak boleh kosong"},
		{"tingkat tanpa kelas", func(d *Data) { d.Classes = d.Classes[:2] }, "tingkat XI dipakai tetapi tidak punya satu pun kelas"},
		{"tingkat kelas tak valid", func(d *Data) { d.Classes[0].Grade = "IX" }, `tingkat "IX" tidak valid`},
		{"nama kelas kembar", func(d *Data) { d.Classes[1].Name = "X 1" }, `nama kelas "X 1" muncul lebih dari sekali`},
		{"tahun ajaran kosong", func(d *Data) { d.AcademicYear = "" }, "academic_year tidak boleh kosong"},
		{"nama guru kosong", func(d *Data) { d.Teachers[0].FullName = "  " }, "full_name) tidak boleh kosong"},
		{"tanpa guru", func(d *Data) { d.Teachers = nil }, "tidak memuat satu pun guru"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			d := smallData()
			tc.mutate(&d)

			db := newDB(t)
			before := countRows(t, db)
			_, err := Run(db, d, true, fixedNow, nil)
			var ve *ValidationError
			if !errors.As(err, &ve) {
				t.Fatalf("seharusnya *ValidationError, dapat: %v", err)
			}
			if !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("pesan galat tidak memuat %q:\n%v", tc.want, err)
			}
			if countRows(t, db) != before {
				t.Fatal("ada data tertulis padahal validasi gagal")
			}
		})
	}
}

func TestLoadData(t *testing.T) {
	good := `{"academic_year":"2026/2027","classes":[{"name":"X 1","grade":"X"}],"subjects":[{"code":"MAT","name":"Matematika"}],
		"teachers":[{"username":"budi","full_name":"Budi","assignments":[{"subject":"MAT","grades":["X"]}]}]}`
	if _, err := LoadData(strings.NewReader(good)); err != nil {
		t.Fatalf("data valid ditolak: %v", err)
	}
	bad := map[string]string{
		"bukan json":     `{`,
		"kolom asing":    strings.Replace(good, `"academic_year"`, `"tahun":"x","academic_year"`, 1),
		"isi tambahan":   good + `{}`,
		"validasi gagal": strings.Replace(good, `"X"]`, `"XX"]`, 1),
	}
	for name, in := range bad {
		if _, err := LoadData(strings.NewReader(in)); err == nil {
			t.Errorf("%s: seharusnya galat", name)
		}
	}
}

func TestGuruLamaPadaClassSubjectDilaporkanDiubah(t *testing.T) {
	db := newDB(t)
	demo := seedDemo(t, db) // X 1 sudah ada; guru1 adalah guru demo
	matID := uuid.New()
	if err := db.Create(&domain.Subject{ID: matID, Code: "MAT", Name: "Matematika", CreatedAt: time.Now()}).Error; err != nil {
		t.Fatal(err)
	}
	// MAT di X 1 sudah dipegang guru1 (harus diganti ke budi); BIN di X 1 sudah dipegang sari (harus tidak berubah).
	binID := uuid.New()
	sari := domain.User{ID: uuid.New(), Username: "sari", PasswordHash: "H", FullName: "Sari, M.Pd.", Role: domain.RoleGuru, IsActive: true, CreatedAt: time.Now(), UpdatedAt: time.Now()}
	if err := db.Create(&sari).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Create(&domain.Subject{ID: binID, Code: "BIN", Name: "Bahasa Indonesia", CreatedAt: time.Now()}).Error; err != nil {
		t.Fatal(err)
	}
	oldCS := domain.ClassSubject{ID: uuid.New(), ClassRoomID: demo.classX1, SubjectID: matID, TeacherID: demo.guru1, AcademicYear: year, CreatedAt: time.Now()}
	sameCS := domain.ClassSubject{ID: uuid.New(), ClassRoomID: demo.classX1, SubjectID: binID, TeacherID: sari.ID, AcademicYear: year, CreatedAt: time.Now()}
	otherYear := domain.ClassSubject{ID: uuid.New(), ClassRoomID: demo.classX1, SubjectID: matID, TeacherID: demo.guru1, AcademicYear: "2025/2026", CreatedAt: time.Now()}
	for _, cs := range []*domain.ClassSubject{&oldCS, &sameCS, &otherYear} {
		if err := db.Create(cs).Error; err != nil {
			t.Fatal(err)
		}
	}

	rep := mustRun(t, db, smallData(), true)

	if len(rep.AssignmentsChanged) != 1 {
		t.Fatalf("seharusnya 1 penugasan diubah, dapat %+v", rep.AssignmentsChanged)
	}
	ch := rep.AssignmentsChanged[0]
	if ch.Class != "X 1" || ch.Subject != "Matematika (MAT)" ||
		ch.OldTeacher != "Guru Demo (guru1)" || ch.NewTeacher != "Budi, S.Pd. (budi)" {
		t.Fatalf("rincian perubahan salah: %+v", ch)
	}
	if rep.AssignmentsUnchanged != 1 || rep.AssignmentsCreated != 3 {
		t.Fatalf("tidak berubah/dibuat seharusnya 1/3, dapat %d/%d", rep.AssignmentsUnchanged, rep.AssignmentsCreated)
	}

	teacherOf := func(id uuid.UUID) uuid.UUID {
		var cs domain.ClassSubject
		if err := db.First(&cs, "id = ?", id).Error; err != nil {
			t.Fatalf("penugasan %s hilang: %v", id, err)
		}
		return cs.TeacherID
	}
	if teacherOf(oldCS.ID) != userByName(t, db, "budi").ID {
		t.Fatal("guru pada penugasan lama tidak diganti")
	}
	if teacherOf(sameCS.ID) != sari.ID {
		t.Fatal("penugasan dengan guru sama berubah")
	}
	if teacherOf(otherYear.ID) != demo.guru1 {
		t.Fatal("penugasan tahun ajaran lain ikut berubah")
	}
	var dup int64
	db.Model(&domain.ClassSubject{}).Where("class_room_id = ? AND subject_id = ? AND academic_year = ?", demo.classX1, matID, year).Count(&dup)
	if dup != 1 {
		t.Fatalf("penugasan X 1 - MAT duplikat: %d baris", dup)
	}

	// Teks laporan memuat guru lama dan baru.
	var buf bytes.Buffer
	if err := rep.WriteText(&buf); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), "Guru Demo (guru1) -> Budi, S.Pd. (budi)") {
		t.Fatalf("laporan tidak menampilkan guru lama dan baru:\n%s", buf.String())
	}
}

func TestDataDemoTetapUtuh(t *testing.T) {
	db := newDB(t)
	demo := seedDemo(t, db)
	snap := func() (classes []domain.ClassRoom, subjects []domain.Subject, users []domain.User, cs domain.ClassSubject) {
		db.Order("name").Find(&classes, "id IN ?", []uuid.UUID{demo.classX1, demo.classMIPA})
		db.Order("code").Find(&subjects, "id IN ?", []uuid.UUID{demo.subjBIN, demo.subjMAT})
		db.Order("username").Find(&users, "id IN ?", []uuid.UUID{demo.admin, demo.guru1})
		db.First(&cs, "id = ?", demo.classSubject)
		return
	}
	c1, s1, u1, cs1 := snap()

	// Pakai data yang juga memuat kelas X 1 dan hanya menambah.
	rep := mustRun(t, db, smallData(), true)
	if len(rep.ClassesExisting) != 1 || rep.ClassesExisting[0] != "X 1" {
		t.Fatalf("X 1 seharusnya tercatat sudah ada: %+v", rep.ClassesExisting)
	}
	mustRun(t, db, smallData(), true) // dan ulang sekali lagi

	c2, s2, u2, cs2 := snap()
	if len(c1) != 2 || len(s1) != 2 || len(u1) != 2 {
		t.Fatalf("fixture demo tidak lengkap: %d %d %d", len(c1), len(s1), len(u1))
	}
	if !equalJSON(t, c1, c2) || !equalJSON(t, s1, s2) || !equalJSON(t, u1, u2) || !equalJSON(t, cs1, cs2) {
		t.Fatal("data demo berubah setelah apply")
	}
	// X 1 tetap bergrade/major aslinya.
	var x1 domain.ClassRoom
	db.First(&x1, "id = ?", demo.classX1)
	if x1.Grade != "X" || x1.Major != "UMUM" {
		t.Fatalf("kelas X 1 berubah: %+v", x1)
	}
	// Hash password akun demo tetap.
	if userByName(t, db, "admin").PasswordHash != "hash-admin" || userByName(t, db, "guru1").PasswordHash != "hash-guru1" {
		t.Fatal("hash akun demo berubah")
	}
	var nX1 int64
	db.Model(&domain.ClassRoom{}).Where("name = ?", "X 1").Count(&nX1)
	if nX1 != 1 {
		t.Fatalf("kelas X 1 terduplikasi: %d", nX1)
	}
}

func TestKelasSudahAdaBerbedaTingkatHanyaPeringatan(t *testing.T) {
	db := newDB(t)
	if err := db.Create(&domain.ClassRoom{ID: uuid.New(), Name: "XI 1", Grade: "XII", Major: "IPS", CreatedAt: time.Now()}).Error; err != nil {
		t.Fatal(err)
	}
	rep := mustRun(t, db, smallData(), true)
	if len(rep.Warnings) == 0 || !strings.Contains(strings.Join(rep.Warnings, "\n"), `"XI 1"`) {
		t.Fatalf("seharusnya ada peringatan untuk kelas XI 1: %v", rep.Warnings)
	}
	var c domain.ClassRoom
	db.First(&c, "name = ?", "XI 1")
	if c.Grade != "XII" || c.Major != "IPS" {
		t.Fatalf("kelas yang ada berubah: %+v", c)
	}
}

func TestGalatDiTengahRollbackPenuh(t *testing.T) {
	db := newDB(t)
	seedDemo(t, db)
	before := countRows(t, db)
	// Tabel class_subjects hilang: kelas, mapel, dan guru sudah sempat ditulis sebelum galat muncul.
	if err := db.Migrator().DropTable(&domain.ClassSubject{}); err != nil {
		t.Fatal(err)
	}
	before.classSubjects = 0

	_, err := Run(db, smallData(), true, fixedNow, nil)
	if err == nil {
		t.Fatal("seharusnya galat")
	}
	var ve *ValidationError
	if errors.As(err, &ve) {
		t.Fatalf("seharusnya galat basis data, bukan validasi: %v", err)
	}
	if got := countRows2(t, db); got.classes != before.classes || got.subjects != before.subjects || got.users != before.users {
		t.Fatalf("rollback tidak penuh: sebelum %+v, sesudah %+v", before, got)
	}
}

// countRows2 seperti countRows tetapi mengabaikan tabel class_subjects yang sengaja dihapus.
func countRows2(t *testing.T, db *gorm.DB) counts {
	t.Helper()
	var c counts
	db.Model(&domain.ClassRoom{}).Count(&c.classes)
	db.Model(&domain.Subject{}).Count(&c.subjects)
	db.Model(&domain.User{}).Count(&c.users)
	return c
}

func TestCelahDanLaporanPerGuru(t *testing.T) {
	db := newDB(t)
	rep := mustRun(t, db, smallData(), false)

	gaps := map[string]bool{}
	for _, g := range rep.Gaps {
		gaps[g.Subject+"|"+g.Grade] = true
	}
	if gaps["Matematika (MAT)|XII"] {
		t.Fatal("tingkat tanpa kelas (XII) tidak boleh dihitung sebagai celah")
	}
	for _, k := range []string{"Bahasa Indonesia (BIN)|XI", "Fisika (FIS)|X", "Fisika (FIS)|XI"} {
		if !gaps[k] {
			t.Errorf("celah %q tidak ada: %+v", k, rep.Gaps)
		}
	}
	if gaps["Matematika (MAT)|X"] || gaps["Bahasa Indonesia (BIN)|X"] {
		t.Error("pasangan yang punya guru dilaporkan sebagai celah")
	}

	if len(rep.Teachers) != 2 {
		t.Fatalf("laporan per guru: %+v", rep.Teachers)
	}
	budi := rep.Teachers[0]
	wantItems := []string{"Matematika - X 1", "Matematika - X 2", "Matematika - XI 1"}
	if budi.Username != "budi" || strings.Join(budi.Items, "|") != strings.Join(wantItems, "|") {
		t.Fatalf("laporan budi salah: %+v", budi)
	}
}

func TestKredensialCSVHanyaAkunBaruDanTanpaKataSandiDiLaporan(t *testing.T) {
	db := newDB(t)
	old := domain.User{ID: uuid.New(), Username: "budi", PasswordHash: "HASH-LAMA", FullName: "Budi, S.Pd.", Role: domain.RoleGuru, IsActive: true, CreatedAt: time.Now(), UpdatedAt: time.Now()}
	if err := db.Create(&old).Error; err != nil {
		t.Fatal(err)
	}
	rep := mustRun(t, db, smallData(), true)

	var out, errOut bytes.Buffer
	if err := WriteCredentialsCSV(&out, rep.Credentials); err != nil {
		t.Fatal(err)
	}
	if err := rep.WriteText(&errOut); err != nil {
		t.Fatal(err)
	}
	rows, err := csv.NewReader(&out).ReadAll()
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 2 || strings.Join(rows[0], ",") != "username,nama_lengkap,password" {
		t.Fatalf("CSV salah: %v", rows)
	}
	if rows[1][0] != "sari" || rows[1][1] != "Sari, M.Pd." || len(rows[1][2]) != 10 {
		t.Fatalf("baris CSV salah: %v", rows[1])
	}
	for _, r := range rows {
		if r[0] == "budi" {
			t.Fatal("akun lama muncul di CSV")
		}
	}
	if strings.Contains(errOut.String(), rows[1][2]) {
		t.Fatal("kata sandi muncul di laporan stderr")
	}
	// Format %v/%+v/%#v juga tidak boleh membocorkan kata sandi.
	for _, s := range []string{fmtAll(rep.Credentials[0]), fmtAll(rep)} {
		if strings.Contains(s, rows[1][2]) {
			t.Fatal("kata sandi bocor lewat format fmt")
		}
	}
	if !strings.Contains(errOut.String(), "APPLY") {
		t.Fatalf("laporan tidak menyebut mode APPLY:\n%s", errOut.String())
	}
}

func TestLaporanDryRunMenyebutMode(t *testing.T) {
	db := newDB(t)
	rep := mustRun(t, db, smallData(), false)
	var buf bytes.Buffer
	if err := rep.WriteText(&buf); err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"DRY-RUN", "2 baru, 0 sudah ada", "5 dibuat", "Per guru", "Matematika - X 1"} {
		if !strings.Contains(buf.String(), want) {
			t.Errorf("laporan tidak memuat %q:\n%s", want, buf.String())
		}
	}
}

func TestKredensialDeterministikDenganRng(t *testing.T) {
	run := func() []Credential {
		db := newDB(t)
		rep, err := Run(db, smallData(), true, fixedNow, rand.New(rand.NewSource(42)))
		if err != nil {
			t.Fatal(err)
		}
		return rep.Credentials
	}
	a, b := run(), run()
	if len(a) != 2 || a[0].Password != b[0].Password || a[1].Password != b[1].Password {
		t.Fatalf("rng yang sama seharusnya menghasilkan kata sandi yang sama: %v %v", a, b)
	}
	if a[0].Password == a[1].Password {
		t.Fatal("dua akun mendapat kata sandi yang sama")
	}
}

func TestGeneratePassword(t *testing.T) {
	seen := map[string]bool{}
	for i := 0; i < 500; i++ {
		p, err := GeneratePassword(rand.New(rand.NewSource(int64(i))))
		if err != nil {
			t.Fatal(err)
		}
		if len(p) != 10 {
			t.Fatalf("panjang %d", len(p))
		}
		if strings.ContainsAny(p, "0O1lI") {
			t.Fatalf("mengandung karakter ambigu: %q", p)
		}
		for _, r := range p {
			if !strings.ContainsRune(passwordAlphabet, r) {
				t.Fatalf("karakter di luar alfabet: %q", p)
			}
		}
		seen[p] = true
	}
	if len(seen) < 495 {
		t.Fatalf("terlalu banyak kata sandi kembar: %d unik dari 500", len(seen))
	}
	if _, err := GeneratePassword(strings.NewReader("pendek")); err == nil {
		t.Fatal("sumber acak yang habis seharusnya galat")
	}
	// Bawaan crypto/rand.
	if p, err := GeneratePassword(cryptoRand()); err != nil || len(p) != 10 {
		t.Fatalf("crypto/rand: %q %v", p, err)
	}
}
