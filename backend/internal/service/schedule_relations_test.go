package service

import (
	"errors"
	"reflect"
	"testing"
	"time"

	"cbt-backend/internal/domain"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// addSession membuat satu sesi ujian siswa pada jadwal (menandakan siswa sudah memulai).
func (f *accessFixture) addSession(t *testing.T, scheduleID uuid.UUID) {
	t.Helper()
	now := time.Now()
	sess := domain.ExamSession{
		ID: uuid.New(), ScheduleID: scheduleID, StudentID: f.siswa.ID,
		StartedAt: now, ServerDeadline: now.Add(time.Hour), Status: domain.StatusInProgress,
		CreatedAt: now, UpdatedAt: now,
	}
	if err := f.db.Create(&sess).Error; err != nil {
		t.Fatalf("gagal membuat sesi ujian: %v", err)
	}
}

// bankOf membaca bank_id jadwal langsung dari basis data.
func (f *accessFixture) bankOf(t *testing.T, scheduleID uuid.UUID) *uuid.UUID {
	t.Helper()
	var s domain.ExamSchedule
	if err := f.db.First(&s, "id = ?", scheduleID).Error; err != nil {
		t.Fatalf("gagal membaca jadwal: %v", err)
	}
	return s.BankID
}

func TestBuildScheduleRelations(t *testing.T) {
	tests := []struct {
		name                         string
		mengampu, bankSaya, pengawas bool
		want                         []string
	}{
		{"tanpa relasi menghasilkan irisan kosong non-nil", false, false, false, []string{}},
		{"hanya mengampu", true, false, false, []string{"mengampu"}},
		{"hanya bank_saya", false, true, false, []string{"bank_saya"}},
		{"hanya pengawas", false, false, true, []string{"pengawas"}},
		{"gabungan dua relasi", true, false, true, []string{"mengampu", "pengawas"}},
		{"gabungan semua dengan urutan tetap", true, true, true, []string{"mengampu", "bank_saya", "pengawas"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := buildScheduleRelations(tt.mengampu, tt.bankSaya, tt.pengawas)
			if got == nil {
				t.Fatal("hasil tidak boleh nil")
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("relasi = %v, ingin %v", got, tt.want)
			}
		})
	}
}

func TestScheduleRelationsFor(t *testing.T) {
	f := newAccessFixture(t)

	// Jadwal Biologi kelas Y memakai bank Matematika buatan guru A: hanya relasi bank_saya bagi guru A.
	schedBankOnly := f.makeScheduleFor(t, "BIO Y (bank A)", f.classY, f.subjectBio, &f.bankMathByA)
	// Jadwal Matematika kelas X dengan bank buatan guru A: mengampu dan bank_saya sekaligus.
	schedTeachAndBank := f.makeScheduleFor(t, "MTK X (bank A)", f.classX, f.subjectMath, &f.bankMathByA)
	// Jadwal Matematika kelas Y tanpa relasi apa pun bagi guru A (mapel sama, kelas berbeda).
	schedMathY := f.makeScheduleFor(t, "MTK Y", f.classY, f.subjectMath, nil)
	// Jadwal dengan bank buatan admin, dipakai untuk menguji pemanggil berizin lihat-semua.
	schedAdminBank := f.makeScheduleFor(t, "MTK Y (bank admin)", f.classY, f.subjectMath, &f.bankMathByAdmin)

	// Guru A juga tercatat sebagai pengawas di jadwal X (yang sudah ia ampu) dan jadwal Y-Biologi.
	if err := f.svc.SetScheduleProctors(f.schedX, []uuid.UUID{f.guruA.ID, f.pengawas.ID}); err != nil {
		t.Fatalf("SetScheduleProctors gagal: %v", err)
	}
	if err := f.svc.SetScheduleProctors(schedMathY, []uuid.UUID{f.pengawas.ID}); err != nil {
		t.Fatalf("SetScheduleProctors gagal: %v", err)
	}
	if err := f.svc.SetScheduleProctors(f.schedY, []uuid.UUID{f.guruA.ID}); err != nil {
		t.Fatalf("SetScheduleProctors gagal: %v", err)
	}

	tests := []struct {
		name string
		user domain.User
		want map[uuid.UUID][]string
	}{
		{
			"guru A: tiap relasi terpisah dan gabungan tanpa duplikat",
			f.guruA,
			map[uuid.UUID][]string{
				f.schedX:          {RelationMengampu, RelationPengawas},
				f.schedY:          {RelationPengawas},
				schedBankOnly:     {RelationBankSaya},
				schedTeachAndBank: {RelationMengampu, RelationBankSaya},
			},
		},
		{
			"guru B: mengampu BIO kelas Y; bank buatan guru A bukan relasinya",
			f.guruB,
			map[uuid.UUID][]string{
				f.schedY:      {RelationMengampu},
				schedBankOnly: {RelationMengampu},
			},
		},
		{
			"pengawas: hanya relasi pengawas",
			f.pengawas,
			map[uuid.UUID][]string{
				f.schedX:   {RelationPengawas},
				schedMathY: {RelationPengawas},
			},
		},
		{
			"admin tetap mendapat relasi (bank_saya) walau melihat semua jadwal",
			f.admin,
			map[uuid.UUID][]string{
				schedAdminBank: {RelationBankSaya},
			},
		},
		{"pengguna tanpa relasi menghasilkan peta kosong", f.siswa, map[uuid.UUID][]string{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := f.svc.ScheduleRelationsFor(tt.user)
			if err != nil {
				t.Fatalf("ScheduleRelationsFor gagal: %v", err)
			}
			if got == nil {
				t.Fatal("peta tidak boleh nil")
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("relasi = %v, ingin %v", got, tt.want)
			}
			for id, rel := range got {
				seen := make(map[string]bool)
				for _, r := range rel {
					if seen[r] {
						t.Fatalf("relasi duplikat pada jadwal %s: %v", id, rel)
					}
					seen[r] = true
				}
			}
		})
	}
}

func TestScheduleVisibilityAndRelationsForConsistent(t *testing.T) {
	f := newAccessFixture(t)
	if err := f.svc.SetScheduleProctors(f.schedY, []uuid.UUID{f.pengawas.ID}); err != nil {
		t.Fatalf("SetScheduleProctors gagal: %v", err)
	}

	// Pengguna terbatas: cakupan persis sama dengan kunci peta relasi dan dengan ScheduleVisibilityFor.
	scope, relations, err := f.svc.ScheduleVisibilityAndRelationsFor(f.guruA)
	if err != nil {
		t.Fatalf("ScheduleVisibilityAndRelationsFor gagal: %v", err)
	}
	plain, err := f.svc.ScheduleVisibilityFor(f.guruA)
	if err != nil {
		t.Fatalf("ScheduleVisibilityFor gagal: %v", err)
	}
	if scope.All || plain.All || !reflect.DeepEqual(scope.ScheduleIDs, plain.ScheduleIDs) {
		t.Fatalf("cakupan tidak konsisten: %+v vs %+v", scope, plain)
	}
	if len(scope.ScheduleIDs) != len(relations) {
		t.Fatalf("cakupan %v harus sama dengan kunci relasi %v", scope.ScheduleIDs, relations)
	}
	for id := range relations {
		if !scope.Allows(id) {
			t.Fatalf("jadwal %s berelasi tetapi tidak terlihat", id)
		}
	}

	// Pemegang izin lihat-semua: All=true dan relasi tetap dihitung.
	scope, relations, err = f.svc.ScheduleVisibilityAndRelationsFor(f.pengawas)
	if err != nil {
		t.Fatalf("ScheduleVisibilityAndRelationsFor gagal: %v", err)
	}
	if scope.All {
		t.Fatal("pengawas bawaan tidak boleh berlingkup semua jadwal")
	}
	if !reflect.DeepEqual(relations[f.schedY], []string{RelationPengawas}) {
		t.Fatalf("relasi pengawas salah: %v", relations)
	}
	adminScope, adminRelations, err := f.svc.ScheduleVisibilityAndRelationsFor(f.admin)
	if err != nil {
		t.Fatalf("ScheduleVisibilityAndRelationsFor gagal: %v", err)
	}
	if !adminScope.All || !adminScope.Allows(f.schedX) || !adminScope.Allows(f.schedY) {
		t.Fatalf("admin seharusnya melihat semua jadwal: %+v", adminScope)
	}
	if adminRelations == nil {
		t.Fatal("peta relasi admin tidak boleh nil")
	}
}

func TestScheduleRelationsForFailsClosed(t *testing.T) {
	f := newAccessFixture(t)
	sqlDB, err := f.db.DB()
	if err != nil {
		t.Fatalf("gagal mengambil sql.DB: %v", err)
	}
	_ = sqlDB.Close()

	scope, relations, err := f.svc.ScheduleVisibilityAndRelationsFor(f.guruA)
	if err == nil {
		t.Fatal("galat kueri seharusnya dikembalikan")
	}
	if scope.All || len(scope.ScheduleIDs) != 0 || len(relations) != 0 {
		t.Fatalf("gagal kueri harus menghasilkan cakupan kosong: %+v %v", scope, relations)
	}
	// Pemegang izin lihat-semua pun tidak boleh lolos bila kueri relasi gagal (gagal tertutup).
	if scope, err := f.svc.ScheduleVisibilityFor(f.admin); err == nil || scope.All {
		t.Fatalf("admin: galat kueri seharusnya menolak (All=%v, err=%v)", scope.All, err)
	}
}

func TestScheduleHasSessions(t *testing.T) {
	f := newAccessFixture(t)

	has, err := f.svc.ScheduleHasSessions(f.schedX)
	if err != nil || has {
		t.Fatalf("jadwal tanpa sesi: has=%v err=%v", has, err)
	}

	f.addSession(t, f.schedX)
	has, err = f.svc.ScheduleHasSessions(f.schedX)
	if err != nil || !has {
		t.Fatalf("jadwal dengan sesi: has=%v err=%v", has, err)
	}
	has, err = f.svc.ScheduleHasSessions(f.schedY)
	if err != nil || has {
		t.Fatalf("sesi jadwal lain tidak boleh terhitung: has=%v err=%v", has, err)
	}

	sqlDB, err := f.db.DB()
	if err != nil {
		t.Fatalf("gagal mengambil sql.DB: %v", err)
	}
	_ = sqlDB.Close()
	if has, err := f.svc.ScheduleHasSessions(f.schedX); err == nil || has {
		t.Fatalf("galat kueri harus dikembalikan (has=%v err=%v)", has, err)
	}
}

func TestSameBankLink(t *testing.T) {
	a, b, zero := uuid.New(), uuid.New(), uuid.Nil
	tests := []struct {
		name            string
		current, target *uuid.UUID
		want            bool
	}{
		{"keduanya nil", nil, nil, true},
		{"nil dan uuid.Nil setara", nil, &zero, true},
		{"bank sama", &a, &a, true},
		{"bank sama nilainya walau pointer berbeda", &a, func() *uuid.UUID { v := a; return &v }(), true},
		{"bank berbeda", &a, &b, false},
		{"melepas bank", &a, nil, false},
		{"menautkan bank pada jadwal tanpa bank", nil, &a, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := SameBankLink(tt.current, tt.target); got != tt.want {
				t.Fatalf("SameBankLink = %v, ingin %v", got, tt.want)
			}
		})
	}
}

func TestEnsureBankChangeAllowed(t *testing.T) {
	f := newAccessFixture(t)
	bankA, bankB := f.bankMathByA, f.bankBioByB

	// Tanpa sesi: semua perubahan diizinkan.
	for _, target := range []*uuid.UUID{&bankA, &bankB, nil} {
		if err := f.svc.EnsureBankChangeAllowed(f.schedX, &bankA, target); err != nil {
			t.Fatalf("jadwal tanpa sesi seharusnya boleh diubah: %v", err)
		}
	}

	f.addSession(t, f.schedX)

	// Dengan sesi: bank yang sama (no-op) boleh; mengganti dan melepas ditolak.
	if err := f.svc.EnsureBankChangeAllowed(f.schedX, &bankA, &bankA); err != nil {
		t.Fatalf("bank yang sama seharusnya no-op sukses: %v", err)
	}
	if err := f.svc.EnsureBankChangeAllowed(f.schedX, nil, nil); err != nil {
		t.Fatalf("jadwal tanpa bank tetap tanpa bank seharusnya no-op sukses: %v", err)
	}
	if err := f.svc.EnsureBankChangeAllowed(f.schedX, &bankA, &bankB); !errors.Is(err, ErrScheduleHasSessions) {
		t.Fatalf("mengganti bank seharusnya ditolak, dapat %v", err)
	}
	if err := f.svc.EnsureBankChangeAllowed(f.schedX, &bankA, nil); !errors.Is(err, ErrScheduleHasSessions) {
		t.Fatalf("melepas bank (bank_id null) seharusnya ditolak, dapat %v", err)
	}
	if err := f.svc.EnsureBankChangeAllowed(f.schedX, nil, &bankB); !errors.Is(err, ErrScheduleHasSessions) {
		t.Fatalf("menautkan bank baru seharusnya ditolak, dapat %v", err)
	}
	if err := f.svc.EnsureScheduleWithoutSessions(f.schedX); !errors.Is(err, ErrScheduleHasSessions) {
		t.Fatalf("EnsureScheduleWithoutSessions seharusnya menolak, dapat %v", err)
	}
	if err := f.svc.EnsureScheduleWithoutSessions(f.schedY); err != nil {
		t.Fatalf("jadwal Y tanpa sesi seharusnya lolos: %v", err)
	}
	if ErrScheduleHasSessions.Error() != "Jadwal sudah dimulai oleh siswa, bank soal tidak dapat diubah" {
		t.Fatalf("pesan galat tidak sesuai kontrak: %q", ErrScheduleHasSessions.Error())
	}
}

func TestEnsureBankChangeAllowedFailsClosed(t *testing.T) {
	f := newAccessFixture(t)
	sqlDB, err := f.db.DB()
	if err != nil {
		t.Fatalf("gagal mengambil sql.DB: %v", err)
	}
	_ = sqlDB.Close()

	bankA, bankB := f.bankMathByA, f.bankBioByB
	err = f.svc.EnsureBankChangeAllowed(f.schedX, &bankA, &bankB)
	if err == nil || errors.Is(err, ErrScheduleHasSessions) {
		t.Fatalf("galat kueri seharusnya dikembalikan sebagai galat sistem, dapat %v", err)
	}
	// Tautan yang sama tidak butuh kueri sehingga tetap lolos.
	if err := f.svc.EnsureBankChangeAllowed(f.schedX, &bankA, &bankA); err != nil {
		t.Fatalf("no-op tidak boleh gagal: %v", err)
	}
}

func TestLinkBankGuarded(t *testing.T) {
	f := newAccessFixture(t)
	bankA, bankB := f.bankMathByA, f.bankBioByB

	// Tanpa sesi: bank tertaut, lalu bisa dilepas.
	if err := f.svc.LinkBankGuarded(f.schedX, &bankA, nil); err != nil {
		t.Fatalf("LinkBankGuarded gagal: %v", err)
	}
	if got := f.bankOf(t, f.schedX); got == nil || *got != bankA {
		t.Fatalf("bank jadwal = %v, ingin %s", got, bankA)
	}
	if err := f.svc.LinkBankGuarded(f.schedX, nil, nil); err != nil {
		t.Fatalf("melepas bank gagal: %v", err)
	}
	if got := f.bankOf(t, f.schedX); got != nil {
		t.Fatalf("bank seharusnya null setelah dilepas, dapat %v", got)
	}

	// Mapel diisi hanya bila subjectID diberikan.
	subj := f.subjectBio
	if err := f.svc.LinkBankGuarded(f.schedX, &bankB, &subj); err != nil {
		t.Fatalf("LinkBankGuarded gagal: %v", err)
	}
	var s domain.ExamSchedule
	if err := f.db.First(&s, "id = ?", f.schedX).Error; err != nil {
		t.Fatalf("gagal membaca jadwal: %v", err)
	}
	if s.SubjectID == nil || *s.SubjectID != subj {
		t.Fatalf("mapel jadwal seharusnya terisi, dapat %v", s.SubjectID)
	}

	// Dengan sesi: ditolak dan tidak ada yang berubah.
	f.addSession(t, f.schedX)
	if err := f.svc.LinkBankGuarded(f.schedX, &bankA, nil); !errors.Is(err, ErrScheduleHasSessions) {
		t.Fatalf("seharusnya ditolak, dapat %v", err)
	}
	if err := f.svc.LinkBankGuarded(f.schedX, nil, nil); !errors.Is(err, ErrScheduleHasSessions) {
		t.Fatalf("melepas bank seharusnya ditolak, dapat %v", err)
	}
	if got := f.bankOf(t, f.schedX); got == nil || *got != bankB {
		t.Fatalf("bank tidak boleh berubah saat ada sesi, dapat %v", got)
	}

	// Jadwal yang tidak ada: not found, bukan konflik.
	if err := f.svc.LinkBankGuarded(uuid.New(), &bankA, nil); !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatalf("jadwal tak ada seharusnya ErrRecordNotFound, dapat %v", err)
	}
}

func TestCreateAndLinkBankGuarded(t *testing.T) {
	f := newAccessFixture(t)
	now := time.Now()
	newBank := func(title string) *domain.QuestionBank {
		return &domain.QuestionBank{ID: uuid.New(), Title: title, SubjectID: f.subjectMath, CreatedByID: f.guruA.ID, CreatedAt: now}
	}

	ok := newBank("Naskah baru")
	if err := f.svc.CreateAndLinkBankGuarded(f.schedX, ok); err != nil {
		t.Fatalf("CreateAndLinkBankGuarded gagal: %v", err)
	}
	if got := f.bankOf(t, f.schedX); got == nil || *got != ok.ID {
		t.Fatalf("bank baru seharusnya tertaut, dapat %v", got)
	}

	// Jadwal Y sudah dimulai siswa: transaksi dibatalkan, bank tidak tersimpan.
	f.addSession(t, f.schedY)
	blocked := newBank("Naskah ditolak")
	if err := f.svc.CreateAndLinkBankGuarded(f.schedY, blocked); !errors.Is(err, ErrScheduleHasSessions) {
		t.Fatalf("seharusnya ditolak, dapat %v", err)
	}
	var count int64
	f.db.Model(&domain.QuestionBank{}).Where("id = ?", blocked.ID).Count(&count)
	if count != 0 {
		t.Fatal("bank tidak boleh tersimpan ketika penautan ditolak")
	}
	if got := f.bankOf(t, f.schedY); got != nil {
		t.Fatalf("bank jadwal Y tidak boleh berubah, dapat %v", got)
	}
}

func TestRedactionHelpers(t *testing.T) {
	f := newAccessFixture(t)

	owner := domain.User{ID: uuid.New(), Username: "penyusun", FullName: "Penyusun", Permissions: []string{"*"}}
	bank := &domain.QuestionBank{ID: uuid.New(), CreatedBy: owner}
	schedules := []domain.ExamSchedule{{ID: uuid.New(), Bank: bank}, {ID: uuid.New()}}
	ClearBankOwners(schedules)
	if !reflect.DeepEqual(schedules[0].Bank.CreatedBy, domain.User{}) {
		t.Fatalf("objek penyusun pada jadwal seharusnya kosong: %+v", schedules[0].Bank.CreatedBy)
	}
	if schedules[1].Bank != nil {
		t.Fatal("jadwal tanpa bank tidak boleh berubah")
	}

	banks := []domain.QuestionBank{{ID: uuid.New(), CreatedBy: owner}}
	StripBankOwnerPermissions(banks)
	if banks[0].CreatedBy.Permissions != nil {
		t.Fatalf("izin penyusun seharusnya dikosongkan: %v", banks[0].CreatedBy.Permissions)
	}
	if banks[0].CreatedBy.FullName != "Penyusun" || banks[0].CreatedBy.Username != "penyusun" {
		t.Fatalf("field lain penyusun harus tetap: %+v", banks[0].CreatedBy)
	}

	withPerms := func(perms ...domain.Permission) domain.User {
		u := f.guruA
		u.Permissions = make([]string, 0, len(perms))
		for _, p := range perms {
			u.Permissions = append(u.Permissions, string(p))
		}
		return u
	}
	usernameTests := []struct {
		name string
		user domain.User
		want bool
	}{
		{"admin", f.admin, true},
		{"guru berdaftar * bukan izin penuh", withPerms(domain.PermAll), false},
		{"master:manage", withPerms(domain.PermMasterManage), true},
		{"users:manage", withPerms(domain.PermUsersManage), true},
		{"guru bawaan", f.guruA, false},
		{"schedules:manage saja", withPerms(domain.PermSchedulesManage), false},
	}
	for _, tt := range usernameTests {
		t.Run("username "+tt.name, func(t *testing.T) {
			if got := CanSeeTeacherUsername(tt.user); got != tt.want {
				t.Fatalf("CanSeeTeacherUsername = %v, ingin %v", got, tt.want)
			}
		})
	}

	mk := func() []domain.ClassSubject {
		return []domain.ClassSubject{{ID: uuid.New(), TeacherID: owner.ID, Teacher: owner}}
	}
	items := mk()
	RedactClassSubjectTeachers(items, false)
	if items[0].Teacher.Permissions != nil || items[0].Teacher.Username != "" {
		t.Fatalf("izin dan username seharusnya kosong: %+v", items[0].Teacher)
	}
	if items[0].Teacher.FullName != "Penyusun" || items[0].Teacher.ID != owner.ID {
		t.Fatalf("id dan full_name harus tetap: %+v", items[0].Teacher)
	}
	items = mk()
	RedactClassSubjectTeachers(items, true)
	if items[0].Teacher.Permissions != nil || items[0].Teacher.Username != "penyusun" {
		t.Fatalf("izin kosong, username tetap: %+v", items[0].Teacher)
	}
}
