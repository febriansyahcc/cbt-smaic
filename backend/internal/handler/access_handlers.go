package handler

import (
	"errors"
	"log"

	"cbt-backend/internal/domain"
	"cbt-backend/internal/middleware"
	"cbt-backend/internal/service"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ---------------- PBAC: HELPER PENOLAKAN AKSES ----------------
//
// Setiap helper deny* menulis respons 403 dan mengembalikan true bila akses ditolak.
// Pemakaian di handler:
//
//	if h.denyBank(c, bankID) { return nil }

func (h *Handlers) forbidden(c *fiber.Ctx, message string) {
	_ = c.Status(fiber.StatusForbidden).JSON(fiber.Map{"success": false, "message": message})
}

func (h *Handlers) denyBank(c *fiber.Ctx, bankID uuid.UUID) bool {
	user, err := middleware.GetCurrentUser(c)
	if err != nil || !h.accessService.CanAccessBank(user, bankID) {
		h.forbidden(c, "Akses ditolak: bank soal ini bukan milik Anda")
		return true
	}
	return false
}

func (h *Handlers) denyScheduleControl(c *fiber.Ctx, scheduleID uuid.UUID) bool {
	user, err := middleware.GetCurrentUser(c)
	if err != nil || !h.accessService.CanControlSchedule(user, scheduleID) {
		h.forbidden(c, "Akses ditolak: jadwal ini tidak ditugaskan kepada Anda sebagai pengawas")
		return true
	}
	return false
}

func (h *Handlers) denyScheduleView(c *fiber.Ctx, scheduleID uuid.UUID) bool {
	user, err := middleware.GetCurrentUser(c)
	if err != nil || !h.accessService.CanViewSchedule(user, scheduleID) {
		h.forbidden(c, "Akses ditolak: jadwal ini tidak ditugaskan kepada Anda sebagai pengawas")
		return true
	}
	return false
}

// denySessionView memeriksa cakupan lihat berdasarkan jadwal milik sesi.
// Sesi yang tidak ditemukan diteruskan ke service agar pesan galat aslinya tampil.
// Galat lain (mis. basis data) menolak akses (fail-closed) dengan 500 dan pesan generik.
func (h *Handlers) denySessionView(c *fiber.Ctx, sessionID uuid.UUID) bool {
	scheduleID, denied := h.scheduleIDForSessionOrDeny(c, sessionID)
	if denied {
		return true
	}
	if scheduleID == uuid.Nil {
		return false
	}
	return h.denyScheduleView(c, scheduleID)
}

// denySessionControl memeriksa kendali berdasarkan jadwal milik sesi.
// Sesi yang tidak ditemukan diteruskan ke service agar pesan galat aslinya tampil.
// Galat lain (mis. basis data) menolak akses (fail-closed) dengan 500 dan pesan generik.
func (h *Handlers) denySessionControl(c *fiber.Ctx, sessionID uuid.UUID) bool {
	scheduleID, denied := h.scheduleIDForSessionOrDeny(c, sessionID)
	if denied {
		return true
	}
	if scheduleID == uuid.Nil {
		return false
	}
	return h.denyScheduleControl(c, scheduleID)
}

// scheduleIDForSessionOrDeny mencari jadwal milik sesi. Mengembalikan:
//   - (jadwal, false): sesi ditemukan
//   - (uuid.Nil, false): sesi tidak ada, diteruskan agar handler/service memberi respons "tidak ditemukan"
//   - (uuid.Nil, true): galat lain; respons 500 sudah ditulis dan akses ditolak
func (h *Handlers) scheduleIDForSessionOrDeny(c *fiber.Ctx, sessionID uuid.UUID) (uuid.UUID, bool) {
	scheduleID, err := h.accessService.ScheduleIDForSession(sessionID)
	if err == nil {
		return scheduleID, false
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return uuid.Nil, false
	}
	log.Printf("gagal memeriksa cakupan sesi %s: %v", sessionID, err)
	_ = c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false, "message": "Terjadi kesalahan saat memeriksa hak akses"})
	return uuid.Nil, true
}

// respondBankChangeError menerjemahkan galat pemeriksaan/penulisan tautan bank jadwal menjadi respons:
// jadwal sudah punya sesi siswa = 409, jadwal hilang = 404, selain itu 500 dengan pesan generik
// (gagal tertutup: perubahan tidak dilanjutkan). Selalu menulis respons dan mengembalikan galat Fiber-nya.
func (h *Handlers) respondBankChangeError(c *fiber.Ctx, scheduleID uuid.UUID, err error) error {
	switch {
	case errors.Is(err, service.ErrScheduleHasSessions):
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{"success": false, "message": err.Error()})
	case errors.Is(err, gorm.ErrRecordNotFound):
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"success": false, "message": "Jadwal ujian tidak ditemukan"})
	default:
		log.Printf("gagal mengubah bank soal jadwal %s: %v", scheduleID, err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false, "message": "Gagal memeriksa atau menyimpan tautan bank soal"})
	}
}

func (h *Handlers) denyStudentControl(c *fiber.Ctx, studentUserID uuid.UUID) bool {
	user, err := middleware.GetCurrentUser(c)
	if err != nil || !h.accessService.CanControlStudent(user, studentUserID) {
		h.forbidden(c, "Akses ditolak: siswa ini tidak berada pada jadwal yang ditugaskan kepada Anda")
		return true
	}
	return false
}

func (h *Handlers) denyScheduleExport(c *fiber.Ctx, scheduleID uuid.UUID) bool {
	user, err := middleware.GetCurrentUser(c)
	if err != nil || !h.accessService.CanExportSchedule(user, scheduleID) {
		h.forbidden(c, "Akses ditolak: Anda bukan pengawas atau pengampu jadwal ini")
		return true
	}
	return false
}

// ---------------- KATALOG IZIN & TEMPLATE ROLE ----------------

// HandleGetPermissionCatalog mengirim daftar izin per kelompok, template role bawaan, dan
// peta implikasi izin (izin -> izin yang otomatis ikut) agar frontend tidak perlu menulis
// ulang daftar izin maupun aturan ketergantungannya.
func (h *Handlers) HandleGetPermissionCatalog(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"success": true,
		"data": fiber.Map{
			"groups":    domain.PermissionCatalog(),
			"templates": domain.RoleTemplates(),
			"implies":   domain.PermissionImplications(),
		},
	})
}

// ---------------- PENUGASAN PENGAWAS PER JADWAL ----------------

type proctorUserItem struct {
	ID       uuid.UUID   `json:"id"`
	Username string      `json:"username"`
	FullName string      `json:"full_name"`
	Role     domain.Role `json:"role"`
}

func (h *Handlers) HandleGetScheduleProctors(c *fiber.Ctx) error {
	scheduleID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "message": "ID jadwal tidak valid"})
	}

	caller, err := middleware.GetCurrentUser(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"success": false, "message": "Autentikasi diperlukan"})
	}

	// Jadwal di luar cakupan keterlihatan dibalas 404 agar keberadaannya tidak terungkap.
	// Gagal kueri = gagal tertutup (500).
	visible, err := h.accessService.ScheduleVisibilityFor(caller)
	if err != nil {
		log.Printf("gagal memeriksa keterlihatan jadwal %s: %v", scheduleID, err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false, "message": "Gagal memeriksa hak akses jadwal"})
	}
	if !visible.Allows(scheduleID) {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"success": false, "message": "Jadwal ujian tidak ditemukan"})
	}

	users, err := h.accessService.GetScheduleProctors(scheduleID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false, "message": "Gagal memuat daftar pengawas"})
	}

	// Pemegang schedules:manage menerima data lengkap; pemegang schedules:read saja
	// hanya menerima id dan nama lengkap (tanpa username dan role).
	if !caller.HasPermission(string(domain.PermSchedulesManage)) {
		refs := make([]proctorRefItem, 0, len(users))
		for _, u := range users {
			refs = append(refs, proctorRefItem{ID: u.ID, FullName: u.FullName})
		}
		return c.JSON(fiber.Map{"success": true, "data": refs})
	}

	items := make([]proctorUserItem, 0, len(users))
	for _, u := range users {
		items = append(items, proctorUserItem{ID: u.ID, Username: u.Username, FullName: u.FullName, Role: u.Role})
	}
	return c.JSON(fiber.Map{"success": true, "data": items})
}

type SetScheduleProctorsRequest struct {
	UserIDs []uuid.UUID `json:"user_ids"`
}

func (h *Handlers) HandleSetScheduleProctors(c *fiber.Ctx) error {
	scheduleID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "message": "ID jadwal tidak valid"})
	}

	var req SetScheduleProctorsRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "message": "Format data tidak valid"})
	}

	if err := h.accessService.SetScheduleProctors(scheduleID, req.UserIDs); err != nil {
		// Galat validasi diteruskan apa adanya; galat sistem (mis. basis data) tidak dibocorkan.
		if service.IsValidationError(err) {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "message": err.Error()})
		}
		log.Printf("gagal menyimpan penugasan pengawas jadwal %s: %v", scheduleID, err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false, "message": "Gagal menyimpan penugasan pengawas"})
	}

	return c.JSON(fiber.Map{"success": true, "message": "Penugasan pengawas berhasil disimpan"})
}

// HandleGetProctorCandidates mengirim akun yang dapat ditugaskan sebagai pengawas
// (aktif, bukan siswa, berizin proctor:control atau proctor:control_all).
// Hanya id, username, nama lengkap, dan role yang dikirim.
func (h *Handlers) HandleGetProctorCandidates(c *fiber.Ctx) error {
	users, err := h.accessService.ProctorCandidates()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false, "message": "Gagal memuat daftar calon pengawas"})
	}

	items := make([]proctorUserItem, 0, len(users))
	for _, u := range users {
		items = append(items, proctorUserItem{ID: u.ID, Username: u.Username, FullName: u.FullName, Role: u.Role})
	}
	return c.JSON(fiber.Map{"success": true, "data": items})
}

// proctorScheduleItem membungkus jadwal dengan penanda apakah pengguna boleh mengendalikannya.
type proctorScheduleItem struct {
	domain.ExamSchedule
	CanControl bool `json:"can_control"`
}

// proctorRefItem adalah ringkasan pengawas yang tampil pada daftar jadwal admin.
type proctorRefItem struct {
	ID       uuid.UUID `json:"id"`
	FullName string    `json:"full_name"`
}

// adminScheduleItem membungkus jadwal dengan daftar pengawas yang ditugaskan.
type adminScheduleItem struct {
	domain.ExamSchedule
	Proctors []proctorRefItem `json:"proctors"`
	// Relations: subset dari "mengampu", "bank_saya", "pengawas" (tanpa duplikat, [] bila kosong).
	Relations []string `json:"relations"`
}

// ---------------- PBAC: PROTEKSI AKUN STAF ----------------

// callerCanGrant bernilai true bila pemanggil memiliki izin penuh ("*") atau berperan ADMIN.
// Hanya pemanggil seperti ini yang boleh mengubah role, mengatur izin, atau menyentuh akun staf.
func (h *Handlers) callerCanGrant(c *fiber.Ctx) bool {
	caller, err := middleware.GetCurrentUser(c)
	return err == nil && caller.HasPermission(string(domain.PermAll))
}

// respondPermissionError membalas galat dari resolver izin: pengiriman izin oleh non-grantor
// dibalas 403, galat validasi (mis. GURU tanpa izin) dibalas 400.
func (h *Handlers) respondPermissionError(c *fiber.Ctx, err error) error {
	if service.IsPermissionsNotAllowed(err) {
		h.forbidden(c, "Akses ditolak: hanya administrator yang dapat mengatur izin akun")
		return nil
	}
	return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "message": err.Error()})
}

// denyLastAdmin membalas 409 dengan pesan yang diberikan (dan mengembalikan true) bila akun
// dengan ID tersebut, sesuai data tersimpan, adalah ADMIN aktif terakhir. Dipakai sebelum
// menurunkan role atau menghapus akun. Akun ADMIN nonaktif tidak dihitung: mengubah atau
// menghapusnya tidak mengurangi jumlah administrator yang dapat masuk. Akun yang tidak
// ditemukan tidak dicegah (pemanggil menangani 404 sendiri). Gagal kueri = gagal tertutup (500).
func (h *Handlers) denyLastAdmin(c *fiber.Ctx, targetID uuid.UUID, message string) bool {
	var target domain.User
	if err := h.repo.DB.First(&target, "id = ?", targetID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false
		}
		_ = c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false, "message": "Gagal memeriksa akun administrator"})
		return true
	}
	if target.Role != domain.RoleAdmin || !target.IsActive {
		return false
	}
	var others int64
	err := h.repo.DB.Model(&domain.User{}).
		Where("role = ? AND is_active = ? AND id <> ?", domain.RoleAdmin, true, target.ID).
		Count(&others).Error
	if err != nil {
		_ = c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false, "message": "Gagal memeriksa akun administrator"})
		return true
	}
	if others == 0 {
		_ = c.Status(fiber.StatusConflict).JSON(fiber.Map{"success": false, "message": message})
		return true
	}
	return false
}

// denyPrivilegedTarget menolak aksi pada akun non-siswa (guru, staf, admin) bila pemanggil
// bukan pemegang izin penuh. Akun siswa tetap dapat dikelola pemegang users:manage/master:manage.
func (h *Handlers) denyPrivilegedTarget(c *fiber.Ctx, targetID uuid.UUID) bool {
	if h.callerCanGrant(c) {
		return false
	}
	var target domain.User
	if err := h.repo.DB.First(&target, "id = ?", targetID).Error; err != nil {
		return false
	}
	if target.Role != domain.RoleSiswa {
		h.forbidden(c, "Akses ditolak: hanya administrator yang dapat mengelola akun guru dan staf")
		return true
	}
	return false
}
