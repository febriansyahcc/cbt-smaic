package service

import (
	"testing"
	"time"

	"cbt-backend/internal/domain"
	"cbt-backend/internal/repository"

	"github.com/glebarez/sqlite"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// fixture berisi data uji: dua guru, satu pengawas, satu admin, satu siswa, dan dua jadwal aktif.
type accessFixture struct {
	svc *AccessService
	db  *gorm.DB

	admin, guruA, guruB, pengawas, siswa domain.User

	subjectMath, subjectBio uuid.UUID
	classX, classY          uuid.UUID
	bankMathByA, bankBioByB uuid.UUID
	bankMathByAdmin         uuid.UUID
	schedX, schedY          uuid.UUID
}

func newAccessFixture(t *testing.T) *accessFixture {
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

	if err := db.AutoMigrate(
		&domain.User{}, &domain.ClassRoom{}, &domain.StudentProfile{}, &domain.Subject{},
		&domain.ClassSubject{}, &domain.QuestionBank{}, &domain.Question{}, &domain.ExamEvent{},
		&domain.ExamSchedule{}, &domain.ExamSession{}, &domain.StudentAnswer{},
		&domain.ViolationLog{}, &domain.ScheduleProctor{},
	); err != nil {
		t.Fatalf("auto-migrate gagal: %v", err)
	}

	f := &accessFixture{
		svc: NewAccessService(&repository.Database{DB: db}),
		db:  db,
	}
	now := time.Now()

	mkUser := func(username string, role domain.Role, perms []string) domain.User {
		u := domain.User{ID: uuid.New(), Username: username, PasswordHash: "x", FullName: username, Role: role, Permissions: perms, IsActive: true, CreatedAt: now}
		if err := db.Create(&u).Error; err != nil {
			t.Fatalf("gagal membuat user %s: %v", username, err)
		}
		return u
	}
	f.admin = mkUser("admin", domain.RoleAdmin, nil)
	f.guruA = mkUser("guru-a", domain.RoleGuru, nil)
	f.guruB = mkUser("guru-b", domain.RoleGuru, nil)
	f.pengawas = mkUser("pengawas", domain.RoleGuru, domain.TemplatePermissions("pengawas"))
	f.siswa = mkUser("siswa", domain.RoleSiswa, nil)

	mkSubject := func(code string) uuid.UUID {
		s := domain.Subject{ID: uuid.New(), Code: code, Name: code, CreatedAt: now}
		if err := db.Create(&s).Error; err != nil {
			t.Fatalf("gagal membuat mapel: %v", err)
		}
		return s.ID
	}
	f.subjectMath = mkSubject("MTK")
	f.subjectBio = mkSubject("BIO")

	mkClass := func(name string) uuid.UUID {
		c := domain.ClassRoom{ID: uuid.New(), Name: name, Grade: "XII", CreatedAt: now}
		if err := db.Create(&c).Error; err != nil {
			t.Fatalf("gagal membuat kelas: %v", err)
		}
		return c.ID
	}
	f.classX = mkClass("XII-X")
	f.classY = mkClass("XII-Y")

	// Guru A mengampu Matematika di kelas X. Guru B mengampu Biologi di kelas Y.
	for _, cs := range []domain.ClassSubject{
		{ID: uuid.New(), ClassRoomID: f.classX, SubjectID: f.subjectMath, TeacherID: f.guruA.ID, AcademicYear: "2025/2026", CreatedAt: now},
		{ID: uuid.New(), ClassRoomID: f.classY, SubjectID: f.subjectBio, TeacherID: f.guruB.ID, AcademicYear: "2025/2026", CreatedAt: now},
	} {
		row := cs
		if err := db.Create(&row).Error; err != nil {
			t.Fatalf("gagal membuat class_subject: %v", err)
		}
	}

	mkBank := func(title string, subject uuid.UUID, by uuid.UUID) uuid.UUID {
		b := domain.QuestionBank{ID: uuid.New(), Title: title, SubjectID: subject, CreatedByID: by, CreatedAt: now}
		if err := db.Create(&b).Error; err != nil {
			t.Fatalf("gagal membuat bank: %v", err)
		}
		return b.ID
	}
	f.bankMathByA = mkBank("MTK A", f.subjectMath, f.guruA.ID)
	f.bankBioByB = mkBank("BIO B", f.subjectBio, f.guruB.ID)
	f.bankMathByAdmin = mkBank("MTK oleh admin", f.subjectMath, f.admin.ID)

	mkSchedule := func(title string, class uuid.UUID, subject uuid.UUID) uuid.UUID {
		sub := subject
		s := domain.ExamSchedule{
			ID: uuid.New(), Title: title, ClassRoomID: class, SubjectID: &sub, ExamToken: "TKN",
			StartTime: now, EndTime: now.Add(time.Hour), DurationMinutes: 60, IsActive: true, CreatedAt: now,
		}
		if err := db.Create(&s).Error; err != nil {
			t.Fatalf("gagal membuat jadwal: %v", err)
		}
		return s.ID
	}
	f.schedX = mkSchedule("Ujian MTK X", f.classX, f.subjectMath)
	f.schedY = mkSchedule("Ujian BIO Y", f.classY, f.subjectBio)

	return f
}

func TestCanAccessBank(t *testing.T) {
	f := newAccessFixture(t)

	tests := []struct {
		name   string
		user   domain.User
		bankID uuid.UUID
		want   bool
	}{
		{"admin melihat semua bank", f.admin, f.bankBioByB, true},
		{"guru mengakses bank buatannya", f.guruA, f.bankMathByA, true},
		{"guru ditolak pada bank mapel yang diampu bila dibuat admin", f.guruA, f.bankMathByAdmin, false},
		{"guru ditolak pada bank mapel yang tidak diampu", f.guruA, f.bankBioByB, false},
		{"guru lain ditolak pada bank buatan guru A", f.guruB, f.bankMathByA, false},
		{"bank tidak ditemukan diteruskan ke handler", f.guruA, uuid.New(), true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := f.svc.CanAccessBank(tt.user, tt.bankID); got != tt.want {
				t.Fatalf("CanAccessBank = %v, ingin %v", got, tt.want)
			}
		})
	}
}

func TestCanAccessBankWithReadAllPermission(t *testing.T) {
	f := newAccessFixture(t)
	kurikulum := f.guruB
	kurikulum.Permissions = []string{string(domain.PermQuestionsAll)}
	if !f.svc.CanAccessBank(kurikulum, f.bankMathByA) {
		t.Fatal("pemegang questions:read_all seharusnya dapat mengakses semua bank")
	}
}

func TestPengawasControlsOnlyAssignedSchedule(t *testing.T) {
	f := newAccessFixture(t)

	// Belum ditugaskan: tidak boleh mengendalikan jadwal mana pun.
	if f.svc.CanControlSchedule(f.pengawas, f.schedX) {
		t.Fatal("pengawas tanpa penugasan tidak boleh mengendalikan jadwal")
	}

	if err := f.svc.SetScheduleProctors(f.schedX, []uuid.UUID{f.pengawas.ID}); err != nil {
		t.Fatalf("SetScheduleProctors gagal: %v", err)
	}
	if !f.svc.CanControlSchedule(f.pengawas, f.schedX) {
		t.Fatal("pengawas seharusnya dapat mengendalikan jadwal yang ditugaskan")
	}
	if f.svc.CanControlSchedule(f.pengawas, f.schedY) {
		t.Fatal("pengawas tidak boleh mengendalikan jadwal yang tidak ditugaskan")
	}
}

func TestControlScopeFor(t *testing.T) {
	f := newAccessFixture(t)
	if err := f.svc.SetScheduleProctors(f.schedX, []uuid.UUID{f.pengawas.ID}); err != nil {
		t.Fatalf("SetScheduleProctors gagal: %v", err)
	}

	if scope := f.svc.ControlScopeFor(f.admin); !scope.All {
		t.Fatal("admin seharusnya berlingkup semua jadwal")
	}

	controlAll := f.guruA
	controlAll.Permissions = []string{string(domain.PermProctorControlAll)}
	if scope := f.svc.ControlScopeFor(controlAll); !scope.All {
		t.Fatal("proctor:control_all seharusnya berlingkup semua jadwal")
	}

	scope := f.svc.ControlScopeFor(f.pengawas)
	if scope.All || !scope.Allows(f.schedX) || scope.Allows(f.schedY) {
		t.Fatalf("lingkup pengawas salah: %+v", scope)
	}

	// Izin proctor:control dicabut walau masih tercatat di penugasan.
	viewOnly := f.pengawas
	viewOnly.Permissions = []string{string(domain.PermProctorView)}
	if scope := f.svc.ControlScopeFor(viewOnly); scope.All || scope.Allows(f.schedX) {
		t.Fatal("tanpa izin proctor:control tidak boleh ada hak kendali")
	}
}

func TestCanControlStudent(t *testing.T) {
	f := newAccessFixture(t)

	profile := domain.StudentProfile{ID: uuid.New(), UserID: f.siswa.ID, NIS: "1001", NISN: "9001", ClassRoomID: f.classX, CreatedAt: time.Now()}
	if err := f.db.Create(&profile).Error; err != nil {
		t.Fatalf("gagal membuat profil siswa: %v", err)
	}

	if f.svc.CanControlStudent(f.pengawas, f.siswa.ID) {
		t.Fatal("pengawas tanpa penugasan tidak boleh mereset siswa")
	}
	if err := f.svc.SetScheduleProctors(f.schedY, []uuid.UUID{f.pengawas.ID}); err != nil {
		t.Fatalf("SetScheduleProctors gagal: %v", err)
	}
	if f.svc.CanControlStudent(f.pengawas, f.siswa.ID) {
		t.Fatal("penugasan pada kelas lain tidak boleh memberi hak atas siswa kelas X")
	}
	if err := f.svc.SetScheduleProctors(f.schedX, []uuid.UUID{f.pengawas.ID}); err != nil {
		t.Fatalf("SetScheduleProctors gagal: %v", err)
	}
	if !f.svc.CanControlStudent(f.pengawas, f.siswa.ID) {
		t.Fatal("pengawas jadwal kelas X seharusnya boleh mereset siswa kelas X")
	}
	if !f.svc.CanControlStudent(f.admin, f.siswa.ID) {
		t.Fatal("admin seharusnya boleh mereset siswa mana pun")
	}
}

func TestCanExportSchedule(t *testing.T) {
	f := newAccessFixture(t)

	if !f.svc.CanExportSchedule(f.guruA, f.schedX) {
		t.Fatal("guru pengampu seharusnya boleh mengekspor jadwal kelas dan mapelnya")
	}
	if f.svc.CanExportSchedule(f.guruA, f.schedY) {
		t.Fatal("guru tidak boleh mengekspor jadwal kelas dan mapel yang tidak diampu")
	}
	if !f.svc.CanExportSchedule(f.admin, f.schedY) {
		t.Fatal("admin seharusnya boleh mengekspor semua jadwal")
	}
	if err := f.svc.SetScheduleProctors(f.schedY, []uuid.UUID{f.pengawas.ID}); err != nil {
		t.Fatalf("SetScheduleProctors gagal: %v", err)
	}
	if !f.svc.CanExportSchedule(f.pengawas, f.schedY) {
		t.Fatal("pengawas yang ditugaskan seharusnya boleh mengekspor jadwalnya")
	}
}

func TestSetScheduleProctorsValidation(t *testing.T) {
	f := newAccessFixture(t)

	if err := f.svc.SetScheduleProctors(uuid.New(), nil); err == nil {
		t.Fatal("jadwal yang tidak ada seharusnya menghasilkan galat")
	}
	if err := f.svc.SetScheduleProctors(f.schedX, []uuid.UUID{f.siswa.ID}); err == nil {
		t.Fatal("siswa tidak boleh ditugaskan sebagai pengawas")
	}
	if err := f.svc.SetScheduleProctors(f.schedX, []uuid.UUID{uuid.New()}); err == nil {
		t.Fatal("pengguna yang tidak ada seharusnya menghasilkan galat")
	}

	inactive := f.guruB
	if err := f.db.Model(&domain.User{}).Where("id = ?", inactive.ID).Update("is_active", false).Error; err != nil {
		t.Fatalf("gagal menonaktifkan akun: %v", err)
	}
	if err := f.svc.SetScheduleProctors(f.schedX, []uuid.UUID{inactive.ID}); err == nil {
		t.Fatal("akun nonaktif tidak boleh ditugaskan")
	}
}

func TestSetScheduleProctorsReplacesAndDeduplicates(t *testing.T) {
	f := newAccessFixture(t)

	if err := f.svc.SetScheduleProctors(f.schedX, []uuid.UUID{f.pengawas.ID, f.pengawas.ID, f.guruA.ID}); err != nil {
		t.Fatalf("SetScheduleProctors gagal: %v", err)
	}
	users, err := f.svc.GetScheduleProctors(f.schedX)
	if err != nil {
		t.Fatalf("GetScheduleProctors gagal: %v", err)
	}
	if len(users) != 2 {
		t.Fatalf("jumlah pengawas = %d, ingin 2 (duplikat harus dibuang)", len(users))
	}

	if err := f.svc.SetScheduleProctors(f.schedX, []uuid.UUID{f.guruA.ID}); err != nil {
		t.Fatalf("SetScheduleProctors gagal: %v", err)
	}
	users, _ = f.svc.GetScheduleProctors(f.schedX)
	if len(users) != 1 || users[0].ID != f.guruA.ID {
		t.Fatalf("penugasan seharusnya diganti seluruhnya, dapat: %+v", users)
	}

	if err := f.svc.SetScheduleProctors(f.schedX, nil); err != nil {
		t.Fatalf("mengosongkan penugasan gagal: %v", err)
	}
	users, _ = f.svc.GetScheduleProctors(f.schedX)
	if len(users) != 0 {
		t.Fatalf("penugasan seharusnya kosong, dapat %d", len(users))
	}
}

func TestViewScopeFor(t *testing.T) {
	f := newAccessFixture(t)
	if err := f.svc.SetScheduleProctors(f.schedX, []uuid.UUID{f.pengawas.ID}); err != nil {
		t.Fatalf("SetScheduleProctors gagal: %v", err)
	}

	if scope := f.svc.ViewScopeFor(f.admin); !scope.All {
		t.Fatal("admin seharusnya melihat semua jadwal")
	}

	viewAll := f.guruA
	viewAll.Permissions = []string{string(domain.PermProctorView)}
	if scope := f.svc.ViewScopeFor(viewAll); !scope.All {
		t.Fatal("proctor:view seharusnya melihat semua jadwal")
	}
	// proctor:view tidak otomatis memberi hak kendali.
	if f.svc.CanControlSchedule(viewAll, f.schedX) {
		t.Fatal("proctor:view saja tidak boleh memberi hak kendali")
	}

	controlAll := f.guruA
	controlAll.Permissions = []string{string(domain.PermProctorControlAll)}
	if scope := f.svc.ViewScopeFor(controlAll); !scope.All {
		t.Fatal("proctor:control_all seharusnya melihat semua jadwal")
	}

	// Pengawas terbatas (hanya proctor:control): hanya jadwal yang ditugaskan.
	scope := f.svc.ViewScopeFor(f.pengawas)
	if scope.All || !scope.Allows(f.schedX) || scope.Allows(f.schedY) {
		t.Fatalf("cakupan lihat pengawas terbatas salah: %+v", scope)
	}
	if !f.svc.CanViewSchedule(f.pengawas, f.schedX) {
		t.Fatal("pengawas seharusnya boleh melihat jadwal yang ditugaskan")
	}
	if f.svc.CanViewSchedule(f.pengawas, f.schedY) {
		t.Fatal("pengawas terbatas tidak boleh melihat jadwal yang tidak ditugaskan")
	}

	// Guru bawaan (template tanpa proctor:view) juga terbatas pada penugasan.
	if f.svc.CanViewSchedule(f.guruA, f.schedX) {
		t.Fatal("guru bawaan tanpa penugasan tidak boleh melihat jadwal")
	}
	if err := f.svc.SetScheduleProctors(f.schedY, []uuid.UUID{f.guruA.ID}); err != nil {
		t.Fatalf("SetScheduleProctors gagal: %v", err)
	}
	if !f.svc.CanViewSchedule(f.guruA, f.schedY) || f.svc.CanViewSchedule(f.guruA, f.schedX) {
		t.Fatal("guru bawaan hanya boleh melihat jadwal yang ditugaskan")
	}

	// Tanpa izin pengawasan sama sekali, penugasan lama tidak memberi hak lihat.
	noPerm := f.pengawas
	noPerm.Permissions = []string{string(domain.PermReportsExport)}
	if scope := f.svc.ViewScopeFor(noPerm); scope.All || scope.Allows(f.schedX) {
		t.Fatal("tanpa izin pengawasan tidak boleh ada hak lihat")
	}
}

func TestProctorCandidatesExcludeStudentsAndInactive(t *testing.T) {
	f := newAccessFixture(t)
	if err := f.db.Model(&domain.User{}).Where("id = ?", f.guruB.ID).Update("is_active", false).Error; err != nil {
		t.Fatalf("gagal menonaktifkan akun: %v", err)
	}

	users, err := f.svc.ProctorCandidates()
	if err != nil {
		t.Fatalf("ProctorCandidates gagal: %v", err)
	}

	want := []string{"guru-a", "pengawas"} // urut full_name ASC; admin tidak ditawarkan
	if len(users) != len(want) {
		t.Fatalf("jumlah kandidat = %d, ingin %d: %+v", len(users), len(want), users)
	}
	for i, u := range users {
		if u.FullName != want[i] {
			t.Fatalf("kandidat ke-%d = %q, ingin %q", i, u.FullName, want[i])
		}
		if u.Role == domain.RoleSiswa || u.ID == f.siswa.ID {
			t.Fatalf("siswa tidak boleh menjadi kandidat: %+v", u)
		}
		if u.ID == f.guruB.ID {
			t.Fatal("akun nonaktif tidak boleh menjadi kandidat")
		}
		if u.PasswordHash != "" || u.SessionToken != "" {
			t.Fatalf("kolom sensitif tidak boleh dimuat: %+v", u)
		}
		if len(u.Permissions) != 0 {
			t.Fatalf("izin tidak boleh ikut dikembalikan: %+v", u)
		}
	}
}

func TestProctorCandidatesRequireControlPermission(t *testing.T) {
	f := newAccessFixture(t)
	now := time.Now()

	mk := func(username string, role domain.Role, perms []string) domain.User {
		u := domain.User{ID: uuid.New(), Username: username, PasswordHash: "x", FullName: username, Role: role, Permissions: perms, IsActive: true, CreatedAt: now}
		if err := f.db.Create(&u).Error; err != nil {
			t.Fatalf("gagal membuat user %s: %v", username, err)
		}
		return u
	}
	viewOnly := mk("hanya-lihat", domain.RoleGuru, []string{string(domain.PermProctorView), string(domain.PermSchedulesRead)})
	noProctor := mk("tanpa-pengawasan", domain.RoleGuru, []string{string(domain.PermQuestionsAssigned)})
	controlAll := mk("kendali-semua", domain.RoleGuru, []string{string(domain.PermProctorControlAll)})
	fullPerm := mk("izin-penuh", domain.RoleGuru, []string{string(domain.PermAll)})

	users, err := f.svc.ProctorCandidates()
	if err != nil {
		t.Fatalf("ProctorCandidates gagal: %v", err)
	}
	got := make(map[uuid.UUID]bool)
	for _, u := range users {
		got[u.ID] = true
	}

	tests := []struct {
		name string
		id   uuid.UUID
		want bool
	}{
		{"admin (sudah mengendalikan semua jadwal)", f.admin.ID, false},
		{"guru dengan izin kosong (template guru)", f.guruA.ID, true},
		{"pengawas (template)", f.pengawas.ID, true},
		{"proctor:control_all (sudah mengendalikan semua jadwal)", controlAll.ID, false},
		// Data lama GURU berdaftar "*" tidak berlaku sebagai izin penuh: kembali ke template guru.
		{"guru berdaftar * (dianggap template guru)", fullPerm.ID, true},
		{"hanya proctor:view", viewOnly.ID, false},
		{"tanpa izin pengawasan", noProctor.ID, false},
		{"siswa", f.siswa.ID, false},
	}
	for _, tt := range tests {
		if got[tt.id] != tt.want {
			t.Errorf("kandidat %q = %v, ingin %v", tt.name, got[tt.id], tt.want)
		}
	}
}

func TestSetScheduleProctorsRejectsAccountWithoutControlPermission(t *testing.T) {
	f := newAccessFixture(t)

	noProctor := domain.User{ID: uuid.New(), Username: "guru-c", PasswordHash: "x", FullName: "Guru C", Role: domain.RoleGuru,
		Permissions: []string{string(domain.PermQuestionsAssigned)}, IsActive: true, CreatedAt: time.Now()}
	if err := f.db.Create(&noProctor).Error; err != nil {
		t.Fatalf("gagal membuat user: %v", err)
	}

	err := f.svc.SetScheduleProctors(f.schedX, []uuid.UUID{f.pengawas.ID, noProctor.ID})
	if err == nil {
		t.Fatal("akun tanpa izin pengendali pengawasan seharusnya ditolak")
	}
	if !IsValidationError(err) {
		t.Fatalf("penolakan seharusnya berupa galat validasi, dapat: %v", err)
	}
	if want := "akun Guru C tidak memiliki izin pengendali pengawasan"; err.Error() != want {
		t.Fatalf("pesan = %q, ingin %q", err.Error(), want)
	}

	// Penolakan tidak boleh menyisakan penugasan sebagian.
	users, _ := f.svc.GetScheduleProctors(f.schedX)
	if len(users) != 0 {
		t.Fatalf("penugasan seharusnya tidak berubah saat ditolak, dapat %d", len(users))
	}
}

func TestSetScheduleProctorsErrorClassification(t *testing.T) {
	f := newAccessFixture(t)

	tests := []struct {
		name string
		run  func() error
	}{
		{"jadwal tidak ada", func() error { return f.svc.SetScheduleProctors(uuid.New(), nil) }},
		{"siswa", func() error { return f.svc.SetScheduleProctors(f.schedX, []uuid.UUID{f.siswa.ID}) }},
		{"pengguna tidak ada", func() error { return f.svc.SetScheduleProctors(f.schedX, []uuid.UUID{uuid.New()}) }},
	}
	for _, tt := range tests {
		if err := tt.run(); err == nil || !IsValidationError(err) {
			t.Errorf("%s: seharusnya galat validasi, dapat %v", tt.name, err)
		}
	}

	// Galat basis data bukan galat validasi.
	if err := f.db.Migrator().DropTable(&domain.ScheduleProctor{}); err != nil {
		t.Fatalf("gagal menghapus tabel: %v", err)
	}
	err := f.svc.SetScheduleProctors(f.schedX, []uuid.UUID{f.pengawas.ID})
	if err == nil {
		t.Fatal("penulisan ke tabel yang hilang seharusnya gagal")
	}
	if IsValidationError(err) {
		t.Fatalf("galat basis data tidak boleh dianggap galat validasi: %v", err)
	}
}

func TestTokenScopeFor(t *testing.T) {
	f := newAccessFixture(t)
	if err := f.svc.SetScheduleProctors(f.schedX, []uuid.UUID{f.pengawas.ID}); err != nil {
		t.Fatalf("SetScheduleProctors gagal: %v", err)
	}

	withPerms := func(perms ...domain.Permission) domain.User {
		u := f.guruA
		u.Permissions = make([]string, 0, len(perms))
		for _, p := range perms {
			u.Permissions = append(u.Permissions, string(p))
		}
		return u
	}

	tests := []struct {
		name         string
		user         domain.User
		wantX, wantY bool
	}{
		{"admin melihat semua token", f.admin, true, true},
		{"schedules:manage melihat semua token", withPerms(domain.PermSchedulesManage), true, true},
		{"proctor:view melihat semua token", withPerms(domain.PermProctorView), true, true},
		{"proctor:control_all melihat semua token", withPerms(domain.PermProctorControlAll), true, true},
		{"pengawas hanya jadwal yang ditugaskan", f.pengawas, true, false},
		{"guru bawaan tanpa penugasan tidak melihat token", f.guruA, false, false},
		{"hanya schedules:read tidak melihat token", withPerms(domain.PermSchedulesRead), false, false},
		{"siswa tidak melihat token", f.siswa, false, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scope := f.svc.TokenScopeFor(tt.user)
			if got := scope.Allows(f.schedX); got != tt.wantX {
				t.Errorf("token jadwal X = %v, ingin %v", got, tt.wantX)
			}
			if got := scope.Allows(f.schedY); got != tt.wantY {
				t.Errorf("token jadwal Y = %v, ingin %v", got, tt.wantY)
			}
		})
	}
}

func TestDeleteUserWithProctorAssignmentsRemovesOrphans(t *testing.T) {
	f := newAccessFixture(t)
	if err := f.svc.SetScheduleProctors(f.schedX, []uuid.UUID{f.pengawas.ID, f.guruA.ID}); err != nil {
		t.Fatalf("SetScheduleProctors gagal: %v", err)
	}
	if err := f.svc.SetScheduleProctors(f.schedY, []uuid.UUID{f.pengawas.ID}); err != nil {
		t.Fatalf("SetScheduleProctors gagal: %v", err)
	}

	if err := f.svc.DeleteUserWithProctorAssignments(f.pengawas.ID); err != nil {
		t.Fatalf("DeleteUserWithProctorAssignments gagal: %v", err)
	}

	var userCount, orphanRows, otherRows int64
	f.db.Model(&domain.User{}).Where("id = ?", f.pengawas.ID).Count(&userCount)
	f.db.Model(&domain.ScheduleProctor{}).Where("user_id = ?", f.pengawas.ID).Count(&orphanRows)
	f.db.Model(&domain.ScheduleProctor{}).Where("user_id = ?", f.guruA.ID).Count(&otherRows)
	if userCount != 0 {
		t.Fatal("akun seharusnya terhapus")
	}
	if orphanRows != 0 {
		t.Fatalf("penugasan akun yang dihapus seharusnya ikut terhapus, tersisa %d", orphanRows)
	}
	if otherRows != 1 {
		t.Fatalf("penugasan pengguna lain tidak boleh terpengaruh, jumlah = %d", otherRows)
	}
}

func TestProctorsBySchedule(t *testing.T) {
	f := newAccessFixture(t)

	got, err := f.svc.ProctorsBySchedule(nil)
	if err != nil || len(got) != 0 {
		t.Fatalf("masukan kosong seharusnya menghasilkan peta kosong, dapat %v (galat: %v)", got, err)
	}

	if err := f.svc.SetScheduleProctors(f.schedX, []uuid.UUID{f.pengawas.ID, f.guruA.ID}); err != nil {
		t.Fatalf("SetScheduleProctors gagal: %v", err)
	}

	got, err = f.svc.ProctorsBySchedule([]uuid.UUID{f.schedX, f.schedY})
	if err != nil {
		t.Fatalf("ProctorsBySchedule gagal: %v", err)
	}
	if len(got[f.schedY]) != 0 {
		t.Fatalf("jadwal tanpa penugasan seharusnya kosong, dapat %+v", got[f.schedY])
	}
	px := got[f.schedX]
	if len(px) != 2 || px[0].ID != f.guruA.ID || px[1].ID != f.pengawas.ID {
		t.Fatalf("pengawas jadwal X salah atau tidak terurut nama: %+v", px)
	}
	if px[0].FullName != "guru-a" || px[0].ScheduleID != f.schedX {
		t.Fatalf("data ringkasan pengawas salah: %+v", px[0])
	}
}

func TestDeleteScheduleWithProctorsRemovesAssignments(t *testing.T) {
	f := newAccessFixture(t)
	if err := f.svc.SetScheduleProctors(f.schedX, []uuid.UUID{f.pengawas.ID, f.guruA.ID}); err != nil {
		t.Fatalf("SetScheduleProctors gagal: %v", err)
	}
	if err := f.svc.SetScheduleProctors(f.schedY, []uuid.UUID{f.pengawas.ID}); err != nil {
		t.Fatalf("SetScheduleProctors gagal: %v", err)
	}

	if err := f.svc.DeleteScheduleWithProctors(f.schedX); err != nil {
		t.Fatalf("DeleteScheduleWithProctors gagal: %v", err)
	}

	var schedCount, rowsX, rowsY int64
	f.db.Model(&domain.ExamSchedule{}).Where("id = ?", f.schedX).Count(&schedCount)
	f.db.Model(&domain.ScheduleProctor{}).Where("schedule_id = ?", f.schedX).Count(&rowsX)
	f.db.Model(&domain.ScheduleProctor{}).Where("schedule_id = ?", f.schedY).Count(&rowsY)
	if schedCount != 0 {
		t.Fatal("jadwal X seharusnya terhapus")
	}
	if rowsX != 0 {
		t.Fatalf("penugasan jadwal X seharusnya ikut terhapus, tersisa %d", rowsX)
	}
	if rowsY != 1 {
		t.Fatalf("penugasan jadwal Y tidak boleh terpengaruh, jumlah = %d", rowsY)
	}
	if f.svc.CanControlSchedule(f.pengawas, f.schedX) {
		t.Fatal("pengawas tidak boleh lagi mengendalikan jadwal yang sudah dihapus")
	}
}

// makeScheduleFor membuat jadwal tambahan untuk uji keterlihatan jadwal.
func (f *accessFixture) makeScheduleFor(t *testing.T, title string, class, subject uuid.UUID, bank *uuid.UUID) uuid.UUID {
	t.Helper()
	now := time.Now()
	sub := subject
	s := domain.ExamSchedule{
		ID: uuid.New(), Title: title, ClassRoomID: class, SubjectID: &sub, BankID: bank, ExamToken: "TKN",
		StartTime: now, EndTime: now.Add(time.Hour), DurationMinutes: 60, IsActive: true, CreatedAt: now,
	}
	if err := f.db.Create(&s).Error; err != nil {
		t.Fatalf("gagal membuat jadwal %s: %v", title, err)
	}
	return s.ID
}

func TestScheduleVisibilityFor(t *testing.T) {
	f := newAccessFixture(t)

	// Mapel yang sama (Matematika) di kelas lain: guru A tidak mengampunya di kelas Y.
	schedMathY := f.makeScheduleFor(t, "Ujian MTK Y", f.classY, f.subjectMath, nil)
	// Jadwal kelas Y mapel Biologi yang memakai bank Matematika buatan guru A.
	schedBankA := f.makeScheduleFor(t, "Ujian BIO Y (bank A)", f.classY, f.subjectBio, &f.bankMathByA)
	// Jadwal tanpa mapel dan bank: tidak boleh memicu kecocokan apa pun.
	noSubject := domain.ExamSchedule{
		ID: uuid.New(), Title: "Tanpa mapel", ClassRoomID: f.classX, ExamToken: "TKN",
		StartTime: time.Now(), EndTime: time.Now().Add(time.Hour), DurationMinutes: 60, IsActive: true, CreatedAt: time.Now(),
	}
	if err := f.db.Create(&noSubject).Error; err != nil {
		t.Fatalf("gagal membuat jadwal tanpa mapel: %v", err)
	}

	// Pengawas ditugaskan pada jadwal Y walau bukan pengampu.
	if err := f.svc.SetScheduleProctors(f.schedY, []uuid.UUID{f.pengawas.ID}); err != nil {
		t.Fatalf("SetScheduleProctors gagal: %v", err)
	}

	withPerms := func(u domain.User, perms ...domain.Permission) domain.User {
		u.Permissions = make([]string, 0, len(perms))
		for _, p := range perms {
			u.Permissions = append(u.Permissions, string(p))
		}
		return u
	}

	all := []uuid.UUID{f.schedX, f.schedY, schedMathY, schedBankA, noSubject.ID}
	tests := []struct {
		name string
		user domain.User
		want []uuid.UUID
	}{
		{"guru pengampu: pasangan kelas-mapel dan bank buatannya", f.guruA, []uuid.UUID{f.schedX, schedBankA}},
		{"guru B: hanya pasangan kelas-mapelnya (BIO kelas Y)", f.guruB, []uuid.UUID{f.schedY, schedBankA}},
		{"pengawas: hanya jadwal yang ditugaskan", f.pengawas, []uuid.UUID{f.schedY}},
		{"admin melihat semua", f.admin, all},
		{"schedules:manage melihat semua", withPerms(f.guruA, domain.PermSchedulesManage), all},
		{"proctor:view melihat semua", withPerms(f.guruA, domain.PermProctorView), all},
		{"proctor:control_all melihat semua", withPerms(f.guruA, domain.PermProctorControlAll), all},
		{"questions:read_all melihat semua", withPerms(f.guruA, domain.PermQuestionsAll), all},
		// GURU berdaftar "*" bukan pemegang izin penuh: hanya jadwal relasinya (template guru).
		{"guru berdaftar * tidak melihat semua", withPerms(f.guruA, domain.PermAll), []uuid.UUID{f.schedX, schedBankA}},
		{"hanya schedules:read tidak melihat semua", withPerms(f.guruB, domain.PermSchedulesRead), []uuid.UUID{f.schedY, schedBankA}},
		{"akun tanpa relasi tidak melihat jadwal", f.siswa, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scope, err := f.svc.ScheduleVisibilityFor(tt.user)
			if err != nil {
				t.Fatalf("ScheduleVisibilityFor gagal: %v", err)
			}
			want := make(map[uuid.UUID]bool)
			for _, id := range tt.want {
				want[id] = true
			}
			for _, id := range all {
				if got := scope.Allows(id); got != want[id] {
					t.Errorf("jadwal %s terlihat = %v, ingin %v", id, got, want[id])
				}
			}
		})
	}
}

func TestScheduleVisibilityForRequiresExactClassSubjectPair(t *testing.T) {
	f := newAccessFixture(t)

	// Guru A mengampu Matematika di kelas X saja; jadwal Matematika kelas Y bukan miliknya.
	schedMathY := f.makeScheduleFor(t, "Ujian MTK Y", f.classY, f.subjectMath, nil)
	// Jadwal Biologi kelas X: kelasnya diampu guru A, tetapi mapelnya bukan.
	schedBioX := f.makeScheduleFor(t, "Ujian BIO X", f.classX, f.subjectBio, nil)

	scope, err := f.svc.ScheduleVisibilityFor(f.guruA)
	if err != nil {
		t.Fatalf("ScheduleVisibilityFor gagal: %v", err)
	}
	if !scope.Allows(f.schedX) {
		t.Fatal("guru A seharusnya melihat jadwal Matematika kelas X")
	}
	if scope.Allows(schedMathY) {
		t.Fatal("guru A tidak boleh melihat jadwal Matematika kelas lain")
	}
	if scope.Allows(schedBioX) {
		t.Fatal("guru A tidak boleh melihat jadwal mapel lain di kelas yang ia ampu")
	}
	if scope.All {
		t.Fatal("guru bawaan tidak boleh berlingkup semua jadwal")
	}
}

func TestScheduleVisibilityForProctorWithoutTeachingRelation(t *testing.T) {
	f := newAccessFixture(t)

	scope, err := f.svc.ScheduleVisibilityFor(f.pengawas)
	if err != nil {
		t.Fatalf("ScheduleVisibilityFor gagal: %v", err)
	}
	if scope.All || scope.Allows(f.schedX) || scope.Allows(f.schedY) {
		t.Fatalf("pengawas tanpa penugasan dan tanpa pengampuan tidak boleh melihat jadwal: %+v", scope)
	}

	if err := f.svc.SetScheduleProctors(f.schedX, []uuid.UUID{f.pengawas.ID}); err != nil {
		t.Fatalf("SetScheduleProctors gagal: %v", err)
	}
	scope, err = f.svc.ScheduleVisibilityFor(f.pengawas)
	if err != nil {
		t.Fatalf("ScheduleVisibilityFor gagal: %v", err)
	}
	if !scope.Allows(f.schedX) || scope.Allows(f.schedY) {
		t.Fatalf("pengawas seharusnya hanya melihat jadwal yang ditugaskan: %+v", scope)
	}
}
