package service

import (
	"errors"
	"time"

	"cbt-backend/internal/domain"
	"cbt-backend/internal/repository"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// AccessService memusatkan aturan cakupan data (scope) untuk PBAC.
// Izin (permission) menentukan aksi yang boleh dilakukan; service ini menentukan
// atas data mana aksi tersebut boleh dilakukan.
type AccessService struct {
	repo *repository.Database
}

func NewAccessService(repo *repository.Database) *AccessService {
	return &AccessService{repo: repo}
}

// ---------------- BANK SOAL ----------------

// CanAccessBank menentukan apakah pengguna boleh menyentuh sebuah bank soal.
// Aturan: questions:read_all (atau admin) lolos; selain itu bank harus dibuat oleh
// pengguna tersebut. Mengampu mata pelajaran yang sama tidak memberi akses ke bank orang lain.
// Bank yang tidak ditemukan mengembalikan true agar handler memberi respons 404 sendiri.
func (s *AccessService) CanAccessBank(user domain.User, bankID uuid.UUID) bool {
	if user.HasPermission(string(domain.PermQuestionsAll)) {
		return true
	}

	var bank domain.QuestionBank
	if err := s.repo.DB.First(&bank, "id = ?", bankID).Error; err != nil {
		return errors.Is(err, gorm.ErrRecordNotFound)
	}
	return bank.CreatedByID == user.ID
}

// ---------------- PENGAWASAN ----------------

// ControlScope menyatakan jadwal mana yang boleh dikendalikan seorang pengguna.
type ControlScope struct {
	All         bool
	ScheduleIDs map[uuid.UUID]bool
}

// Allows memeriksa apakah jadwal termasuk dalam cakupan kendali.
func (c ControlScope) Allows(scheduleID uuid.UUID) bool {
	return c.All || c.ScheduleIDs[scheduleID]
}

// ControlScopeFor menghitung cakupan kendali pengguna.
//   - proctor:control_all (atau admin): semua jadwal
//   - proctor:control: hanya jadwal di schedule_proctors
//   - selain itu: tidak ada
func (s *AccessService) ControlScopeFor(user domain.User) ControlScope {
	scope := ControlScope{ScheduleIDs: make(map[uuid.UUID]bool)}
	if user.HasPermission(string(domain.PermProctorControlAll)) {
		scope.All = true
		return scope
	}
	if !user.HasPermission(string(domain.PermProctorControl)) {
		return scope
	}

	var rows []domain.ScheduleProctor
	if err := s.repo.DB.Where("user_id = ?", user.ID).Find(&rows).Error; err != nil {
		return scope
	}
	for _, r := range rows {
		scope.ScheduleIDs[r.ScheduleID] = true
	}
	return scope
}

// CanControlSchedule menentukan apakah pengguna boleh mengendalikan sesi pada jadwal tertentu.
func (s *AccessService) CanControlSchedule(user domain.User, scheduleID uuid.UUID) bool {
	return s.ControlScopeFor(user).Allows(scheduleID)
}

// ViewScopeFor menghitung cakupan lihat pengawasan pengguna.
//   - proctor:view, proctor:control_all (atau admin): semua jadwal aktif
//   - proctor:control saja: hanya jadwal di schedule_proctors (sama dengan cakupan kendali)
//   - selain itu: tidak ada
func (s *AccessService) ViewScopeFor(user domain.User) ControlScope {
	if user.HasPermission(string(domain.PermProctorView)) {
		return ControlScope{All: true, ScheduleIDs: make(map[uuid.UUID]bool)}
	}
	return s.ControlScopeFor(user)
}

// CanViewSchedule menentukan apakah pengguna boleh memantau sesi pada jadwal tertentu.
func (s *AccessService) CanViewSchedule(user domain.User, scheduleID uuid.UUID) bool {
	return s.ViewScopeFor(user).Allows(scheduleID)
}

// TokenScopeFor menghitung jadwal yang boleh diketahui token ujiannya oleh pengguna.
//   - schedules:manage, proctor:view, proctor:control_all (atau admin): semua jadwal
//   - selain itu: hanya jadwal dalam cakupan kendali (proctor:control + penugasan)
//
// Hitung sekali per permintaan, lalu pakai Allows(scheduleID) per jadwal.
func (s *AccessService) TokenScopeFor(user domain.User) ControlScope {
	if user.HasPermission(string(domain.PermSchedulesManage)) ||
		user.HasPermission(string(domain.PermProctorView)) ||
		user.HasPermission(string(domain.PermProctorControlAll)) {
		return ControlScope{All: true, ScheduleIDs: make(map[uuid.UUID]bool)}
	}
	return s.ControlScopeFor(user)
}

// ScheduleVisibilityFor menghitung jadwal ujian mana yang boleh dilihat pengguna staf.
// Ini adalah satu-satunya sumber kebenaran keterlihatan jadwal (daftar jadwal admin dan
// matriks kesiapan memakainya).
//
//   - schedules:manage, proctor:view, proctor:control_all, atau questions:read_all
//     (ADMIN dan izin "*" otomatis lolos): semua jadwal
//   - selain itu, jadwal terlihat bila salah satu benar:
//     (a) pasangan (kelas, mapel) jadwal ada di class_subjects dengan guru = pengguna
//     (kecocokan pasangan tepat; mengampu mapel yang sama di kelas lain tidak cukup),
//     (b) bank soal jadwal dibuat oleh pengguna,
//     (c) pengguna tercatat di schedule_proctors untuk jadwal itu.
//
// Cakupan dihitung dengan satu query lalu dipakai per jadwal lewat Allows(scheduleID).
// Gagal kueri berarti gagal tertutup: cakupan kosong beserta galatnya.
func (s *AccessService) ScheduleVisibilityFor(user domain.User) (ControlScope, error) {
	scope, _, err := s.ScheduleVisibilityAndRelationsFor(user)
	return scope, err
}

// Nama relasi pengguna terhadap sebuah jadwal (nilai pada field "relations" daftar jadwal admin).
const (
	RelationMengampu = "mengampu" // pasangan kelas + mapel jadwal ada di class_subjects milik pengguna
	RelationBankSaya = "bank_saya" // bank soal jadwal dibuat pengguna
	RelationPengawas = "pengawas"  // pengguna tercatat di schedule_proctors jadwal itu
)

// hasSeeAllSchedules menyatakan apakah izin efektif pengguna membuka semua jadwal
// (ADMIN dan izin "*" otomatis lolos lewat HasPermission).
func hasSeeAllSchedules(user domain.User) bool {
	for _, perm := range []domain.Permission{
		domain.PermSchedulesManage,
		domain.PermProctorView,
		domain.PermProctorControlAll,
		domain.PermQuestionsAll,
	} {
		if user.HasPermission(string(perm)) {
			return true
		}
	}
	return false
}

// buildScheduleRelations menyusun daftar relasi dengan urutan tetap dan tanpa duplikat.
// Hasilnya selalu berupa irisan non-nil ([] bila tidak ada relasi).
func buildScheduleRelations(mengampu, bankSaya, pengawas bool) []string {
	relations := make([]string, 0, 3)
	if mengampu {
		relations = append(relations, RelationMengampu)
	}
	if bankSaya {
		relations = append(relations, RelationBankSaya)
	}
	if pengawas {
		relations = append(relations, RelationPengawas)
	}
	return relations
}

// scheduleRelationRow adalah satu baris hasil kueri relasi (satu baris per jadwal).
type scheduleRelationRow struct {
	ID       uuid.UUID
	Mengampu int
	BankSaya int
	Pengawas int
}

// ScheduleRelationsFor menghitung relasi pengguna terhadap setiap jadwal dengan satu kueri
// (tanpa N+1). Peta hanya memuat jadwal yang punya minimal satu relasi; nilainya berisi
// subset dari "mengampu", "bank_saya", "pengawas" (urutan tetap, tanpa duplikat).
// Perhitungan tidak bergantung pada izin: berlaku juga bagi pemegang izin lihat-semua.
// Gagal kueri mengembalikan peta kosong beserta galatnya.
func (s *AccessService) ScheduleRelationsFor(user domain.User) (map[uuid.UUID][]string, error) {
	result := make(map[uuid.UUID][]string)

	var rows []scheduleRelationRow
	err := s.repo.DB.Table("exam_schedules").
		Select(`exam_schedules.id AS id,
			CASE WHEN EXISTS (SELECT 1 FROM class_subjects cs
				WHERE cs.class_room_id = exam_schedules.class_room_id
				AND cs.subject_id = exam_schedules.subject_id
				AND cs.teacher_id = ?) THEN 1 ELSE 0 END AS mengampu,
			CASE WHEN exam_schedules.bank_id IN
				(SELECT qb.id FROM question_banks qb WHERE qb.created_by_id = ?) THEN 1 ELSE 0 END AS bank_saya,
			CASE WHEN EXISTS (SELECT 1 FROM schedule_proctors sp
				WHERE sp.schedule_id = exam_schedules.id AND sp.user_id = ?) THEN 1 ELSE 0 END AS pengawas`,
			user.ID, user.ID, user.ID).
		Scan(&rows).Error
	if err != nil {
		return make(map[uuid.UUID][]string), err
	}
	for _, r := range rows {
		relations := buildScheduleRelations(r.Mengampu == 1, r.BankSaya == 1, r.Pengawas == 1)
		if len(relations) > 0 {
			result[r.ID] = relations
		}
	}
	return result, nil
}

// ScheduleVisibilityAndRelationsFor menghitung cakupan keterlihatan sekaligus relasi per jadwal
// dari himpunan id yang sama, dengan satu kueri. Aturan keterlihatan sama dengan
// ScheduleVisibilityFor: pemegang izin lihat-semua mendapat All=true, selain itu jadwal terlihat
// bila pengguna punya minimal satu relasi. Peta relasi dihitung untuk semua pemanggil.
// Gagal kueri berarti gagal tertutup: cakupan kosong beserta galatnya.
func (s *AccessService) ScheduleVisibilityAndRelationsFor(user domain.User) (ControlScope, map[uuid.UUID][]string, error) {
	relations, err := s.ScheduleRelationsFor(user)
	if err != nil {
		return ControlScope{ScheduleIDs: make(map[uuid.UUID]bool)}, make(map[uuid.UUID][]string), err
	}

	scope := ControlScope{ScheduleIDs: make(map[uuid.UUID]bool)}
	if hasSeeAllSchedules(user) {
		scope.All = true
		return scope, relations, nil
	}
	for id := range relations {
		scope.ScheduleIDs[id] = true
	}
	return scope, relations, nil
}

// ---------------- SESI UJIAN & BANK SOAL JADWAL ----------------

// ErrScheduleHasSessions menandai bank soal jadwal tidak boleh diubah karena siswa sudah memulai ujian.
var ErrScheduleHasSessions = errors.New("Jadwal sudah dimulai oleh siswa, bank soal tidak dapat diubah")

// ScheduleHasSessions memeriksa apakah jadwal sudah punya sesi ujian siswa (baris exam_sessions).
// Pemanggil harus memperlakukan galat sebagai gagal tertutup (jangan lanjutkan perubahan).
func (s *AccessService) ScheduleHasSessions(scheduleID uuid.UUID) (bool, error) {
	var count int64
	if err := s.repo.DB.Model(&domain.ExamSession{}).Where("schedule_id = ?", scheduleID).Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

// SameBankLink membandingkan dua tautan bank soal. nil dan uuid.Nil sama-sama berarti "tanpa bank".
func SameBankLink(current, target *uuid.UUID) bool {
	var a, b uuid.UUID
	if current != nil {
		a = *current
	}
	if target != nil {
		b = *target
	}
	return a == b
}

// EnsureBankChangeAllowed memeriksa apakah tautan bank jadwal boleh diubah dari current ke target.
// Tautan yang sama (no-op) selalu diizinkan. Perubahan (termasuk melepas bank) ditolak dengan
// ErrScheduleHasSessions bila jadwal sudah punya sesi ujian siswa. Galat kueri dikembalikan
// apa adanya (gagal tertutup).
func (s *AccessService) EnsureBankChangeAllowed(scheduleID uuid.UUID, current, target *uuid.UUID) error {
	if SameBankLink(current, target) {
		return nil
	}
	return s.EnsureScheduleWithoutSessions(scheduleID)
}

// EnsureScheduleWithoutSessions mengembalikan ErrScheduleHasSessions bila jadwal sudah punya sesi siswa.
func (s *AccessService) EnsureScheduleWithoutSessions(scheduleID uuid.UUID) error {
	has, err := s.ScheduleHasSessions(scheduleID)
	if err != nil {
		return err
	}
	if has {
		return ErrScheduleHasSessions
	}
	return nil
}

// noSessionsGuard adalah kondisi SQL yang hanya benar bila jadwal belum punya sesi ujian siswa.
// Dipakai pada UPDATE agar pemeriksaan dan penulisan satu langkah (menutup celah siswa yang
// memulai ujian tepat di antara pemeriksaan dan penyimpanan).
const noSessionsGuard = "id = ? AND NOT EXISTS (SELECT 1 FROM exam_sessions es WHERE es.schedule_id = ?)"

// LinkBankGuarded menautkan (atau melepas, bila bankID nil) bank soal pada jadwal dan mengisi
// mapel jadwal bila subjectID tidak nil. Penulisan hanya terjadi bila jadwal belum punya sesi
// siswa; selain itu ErrScheduleHasSessions dikembalikan dan tidak ada yang berubah.
func (s *AccessService) LinkBankGuarded(scheduleID uuid.UUID, bankID, subjectID *uuid.UUID) error {
	updates := map[string]interface{}{"bank_id": nil}
	if bankID != nil && *bankID != uuid.Nil {
		updates["bank_id"] = *bankID
	}
	if subjectID != nil {
		updates["subject_id"] = *subjectID
	}
	res := s.repo.DB.Model(&domain.ExamSchedule{}).Where(noSessionsGuard, scheduleID, scheduleID).Updates(updates)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return scheduleUnchangedError(s.repo.DB, scheduleID)
	}
	return nil
}

// CreateAndLinkBankGuarded menyimpan bank soal baru dan menautkannya ke jadwal dalam satu
// transaksi. Bila jadwal sudah punya sesi siswa, transaksi dibatalkan (bank tidak tersimpan)
// dan ErrScheduleHasSessions dikembalikan.
func (s *AccessService) CreateAndLinkBankGuarded(scheduleID uuid.UUID, bank *domain.QuestionBank) error {
	return s.repo.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(bank).Error; err != nil {
			return err
		}
		res := tx.Model(&domain.ExamSchedule{}).Where(noSessionsGuard, scheduleID, scheduleID).
			Update("bank_id", bank.ID)
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			// Pemeriksaan memakai handle transaksi agar tidak meminta koneksi kedua.
			return scheduleUnchangedError(tx, scheduleID)
		}
		return nil
	})
}

// scheduleUnchangedError menjelaskan mengapa UPDATE bersyarat tidak mengubah baris:
// jadwal sudah punya sesi siswa, atau jadwal tidak ada lagi.
func scheduleUnchangedError(db *gorm.DB, scheduleID uuid.UUID) error {
	var count int64
	if err := db.Model(&domain.ExamSession{}).Where("schedule_id = ?", scheduleID).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return ErrScheduleHasSessions
	}
	return gorm.ErrRecordNotFound
}

// ScheduleIDForSession mencari jadwal dari sebuah sesi ujian siswa.
func (s *AccessService) ScheduleIDForSession(sessionID uuid.UUID) (uuid.UUID, error) {
	var session domain.ExamSession
	if err := s.repo.DB.First(&session, "id = ?", sessionID).Error; err != nil {
		return uuid.Nil, err
	}
	return session.ScheduleID, nil
}

// CanControlStudent menentukan apakah pengguna boleh mereset perangkat login seorang siswa.
// Untuk pengawas terbatas: siswa harus berada pada kelas dari minimal satu jadwal aktif
// yang ditugaskan kepada pengguna.
func (s *AccessService) CanControlStudent(user domain.User, studentUserID uuid.UUID) bool {
	scope := s.ControlScopeFor(user)
	if scope.All {
		return true
	}
	if len(scope.ScheduleIDs) == 0 {
		return false
	}

	var profile domain.StudentProfile
	if err := s.repo.DB.Where("user_id = ?", studentUserID).First(&profile).Error; err != nil {
		return false
	}

	ids := make([]uuid.UUID, 0, len(scope.ScheduleIDs))
	for id := range scope.ScheduleIDs {
		ids = append(ids, id)
	}

	var count int64
	s.repo.DB.Model(&domain.ExamSchedule{}).
		Where("id IN ? AND class_room_id = ? AND is_active = ?", ids, profile.ClassRoomID, true).
		Count(&count)
	return count > 0
}

// CanExportSchedule menentukan apakah pengguna boleh mengekspor laporan sebuah jadwal.
// Lolos bila: dapat mengendalikan jadwal, memiliki questions:read_all, atau mengampu
// kombinasi kelas dan mata pelajaran jadwal tersebut.
func (s *AccessService) CanExportSchedule(user domain.User, scheduleID uuid.UUID) bool {
	if s.CanControlSchedule(user, scheduleID) {
		return true
	}
	if user.HasPermission(string(domain.PermQuestionsAll)) {
		return true
	}

	var sched domain.ExamSchedule
	if err := s.repo.DB.First(&sched, "id = ?", scheduleID).Error; err != nil {
		return errors.Is(err, gorm.ErrRecordNotFound)
	}
	if sched.SubjectID == nil {
		return false
	}

	var count int64
	s.repo.DB.Model(&domain.ClassSubject{}).
		Where("teacher_id = ? AND class_room_id = ? AND subject_id = ?", user.ID, sched.ClassRoomID, *sched.SubjectID).
		Count(&count)
	return count > 0
}

// ---------------- PENUGASAN PENGAWAS ----------------

// GetScheduleProctors mengembalikan daftar pengawas yang ditugaskan pada sebuah jadwal.
func (s *AccessService) GetScheduleProctors(scheduleID uuid.UUID) ([]domain.User, error) {
	var rows []domain.ScheduleProctor
	if err := s.repo.DB.Where("schedule_id = ?", scheduleID).Find(&rows).Error; err != nil {
		return nil, err
	}
	users := make([]domain.User, 0, len(rows))
	if len(rows) == 0 {
		return users, nil
	}

	ids := make([]uuid.UUID, 0, len(rows))
	for _, r := range rows {
		ids = append(ids, r.UserID)
	}
	if err := s.repo.DB.Where("id IN ?", ids).Order("full_name ASC").Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}

// ScheduleProctorRef adalah ringkasan pengawas yang ditugaskan pada satu jadwal.
type ScheduleProctorRef struct {
	ScheduleID uuid.UUID
	ID         uuid.UUID
	FullName   string
}

// ProctorsBySchedule mengambil pengawas untuk banyak jadwal sekaligus dengan satu query IN.
// Hasilnya dipetakan per jadwal, terurut nama lengkap. Jadwal tanpa pengawas tidak ada di peta.
func (s *AccessService) ProctorsBySchedule(scheduleIDs []uuid.UUID) (map[uuid.UUID][]ScheduleProctorRef, error) {
	result := make(map[uuid.UUID][]ScheduleProctorRef)
	if len(scheduleIDs) == 0 {
		return result, nil
	}

	var rows []ScheduleProctorRef
	err := s.repo.DB.Table("schedule_proctors").
		Select("schedule_proctors.schedule_id AS schedule_id, users.id AS id, users.full_name AS full_name").
		Joins("JOIN users ON users.id = schedule_proctors.user_id").
		Where("schedule_proctors.schedule_id IN ?", scheduleIDs).
		Order("users.full_name ASC").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	for _, r := range rows {
		result[r.ScheduleID] = append(result[r.ScheduleID], r)
	}
	return result, nil
}

// ValidationError menandai galat validasi masukan (aturan bisnis), berbeda dari galat sistem
// seperti kegagalan basis data. Handler membalas galat ini dengan 400 beserta pesan aslinya.
type ValidationError struct {
	Message string
}

func (e *ValidationError) Error() string { return e.Message }

func newValidationError(message string) error {
	return &ValidationError{Message: message}
}

// IsValidationError memeriksa apakah galat (atau galat yang dibungkusnya) adalah galat validasi.
func IsValidationError(err error) bool {
	var ve *ValidationError
	return errors.As(err, &ve)
}

// canBeProctor menyatakan apakah izin efektif akun mencakup pengendalian pengawasan
// (proctor:control atau proctor:control_all). Tanpa izin ini, penugasan tidak berefek.
// Akun ADMIN dan pemegang izin penuh otomatis lolos. Izin kosong memakai template role.
func canBeProctor(user *domain.User) bool {
	return user.HasPermission(string(domain.PermProctorControl)) ||
		user.HasPermission(string(domain.PermProctorControlAll))
}

// ProctorCandidates mengembalikan akun yang dapat ditugaskan sebagai pengawas:
// aktif, bukan SISWA, dan izin efektifnya mencakup proctor:control atau proctor:control_all,
// terurut nama lengkap. Hanya kolom yang perlu dimuat; izin dipakai untuk penyaringan
// lalu dikosongkan kembali agar tidak ikut terkirim ke klien.
func (s *AccessService) ProctorCandidates() ([]domain.User, error) {
	var users []domain.User
	err := s.repo.DB.Select("id", "username", "full_name", "role", "permissions").
		Where("role <> ? AND is_active = ?", domain.RoleSiswa, true).
		Order("full_name ASC").
		Find(&users).Error
	if err != nil {
		return nil, err
	}

	candidates := make([]domain.User, 0, len(users))
	for i := range users {
		// Akun yang sudah dapat mengendalikan semua jadwal (admin, izin "*", proctor:control_all)
		// tidak perlu ditugaskan per jadwal, jadi tidak ditawarkan sebagai pilihan.
		if users[i].HasPermission(string(domain.PermProctorControlAll)) {
			continue
		}
		if !users[i].HasPermission(string(domain.PermProctorControl)) {
			continue
		}
		users[i].Permissions = nil
		candidates = append(candidates, users[i])
	}
	return candidates, nil
}

// DeleteUserWithProctorAssignments menghapus akun beserta penugasan pengawasnya dalam satu
// transaksi agar tidak tersisa baris schedule_proctors yatim.
func (s *AccessService) DeleteUserWithProctorAssignments(userID uuid.UUID) error {
	return s.repo.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("user_id = ?", userID).Delete(&domain.ScheduleProctor{}).Error; err != nil {
			return err
		}
		return tx.Delete(&domain.User{}, "id = ?", userID).Error
	})
}

// DeleteScheduleWithProctors menghapus jadwal beserta penugasan pengawasnya dalam satu transaksi.
func (s *AccessService) DeleteScheduleWithProctors(scheduleID uuid.UUID) error {
	return s.repo.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("schedule_id = ?", scheduleID).Delete(&domain.ScheduleProctor{}).Error; err != nil {
			return err
		}
		return tx.Delete(&domain.ExamSchedule{}, "id = ?", scheduleID).Error
	})
}

// SetScheduleProctors mengganti seluruh penugasan pengawas sebuah jadwal.
// Hanya akun staf (bukan SISWA) yang aktif dan berizin proctor:control atau
// proctor:control_all yang dapat ditugaskan.
//
// Validasi dan penulisan berjalan dalam satu transaksi. Pelanggaran aturan dikembalikan
// sebagai *ValidationError; galat lain (mis. basis data) dikembalikan apa adanya.
func (s *AccessService) SetScheduleProctors(scheduleID uuid.UUID, userIDs []uuid.UUID) error {
	unique := make([]uuid.UUID, 0, len(userIDs))
	seen := make(map[uuid.UUID]bool)
	for _, id := range userIDs {
		if id == uuid.Nil || seen[id] {
			continue
		}
		seen[id] = true
		unique = append(unique, id)
	}

	return s.repo.DB.Transaction(func(tx *gorm.DB) error {
		var sched domain.ExamSchedule
		if err := tx.Select("id").First(&sched, "id = ?", scheduleID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return newValidationError("jadwal ujian tidak ditemukan")
			}
			return err
		}

		if len(unique) > 0 {
			var users []domain.User
			if err := tx.Select("id", "full_name", "role", "permissions", "is_active").
				Where("id IN ?", unique).Find(&users).Error; err != nil {
				return err
			}
			if len(users) != len(unique) {
				return newValidationError("sebagian pengguna yang dipilih tidak ditemukan")
			}
			for i := range users {
				u := &users[i]
				if u.Role == domain.RoleSiswa {
					return newValidationError("siswa tidak dapat ditugaskan sebagai pengawas")
				}
				if !u.IsActive {
					return newValidationError("akun " + u.FullName + " sedang dinonaktifkan")
				}
				if !canBeProctor(u) {
					return newValidationError("akun " + u.FullName + " tidak memiliki izin pengendali pengawasan")
				}
			}
		}

		if err := tx.Where("schedule_id = ?", scheduleID).Delete(&domain.ScheduleProctor{}).Error; err != nil {
			return err
		}
		now := time.Now()
		for _, id := range unique {
			row := domain.ScheduleProctor{ScheduleID: scheduleID, UserID: id, CreatedAt: now}
			if err := tx.Create(&row).Error; err != nil {
				return err
			}
		}
		return nil
	})
}
