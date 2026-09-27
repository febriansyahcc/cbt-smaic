package handler

import (
	"fmt"
	"strings"
	"time"

	"cbt-backend/internal/domain"
	"cbt-backend/internal/middleware"
	"cbt-backend/internal/repository"
	"cbt-backend/internal/service"
	"cbt-backend/pkg/excel"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// ---------------- ADMIN CRUD HANDLERS ----------------

func (h *Handlers) HandleGetAdminStats(c *fiber.Ctx) error {
	var totalStudents, totalTeachers, totalClasses, activeSchedules, totalBanks, totalSubjects, totalClassSubjects int64
	h.repo.DB.Model(&domain.User{}).Where("role = ?", domain.RoleSiswa).Count(&totalStudents)
	h.repo.DB.Model(&domain.User{}).Where("role IN ?", []domain.Role{domain.RoleGuru, domain.RoleAdmin}).Count(&totalTeachers)
	h.repo.DB.Model(&domain.ClassRoom{}).Count(&totalClasses)
	h.repo.DB.Model(&domain.ExamSchedule{}).Where("is_active = ?", true).Count(&activeSchedules)
	h.repo.DB.Model(&domain.QuestionBank{}).Count(&totalBanks)
	h.repo.DB.Model(&domain.Subject{}).Count(&totalSubjects)
	h.repo.DB.Model(&domain.ClassSubject{}).Count(&totalClassSubjects)

	return c.JSON(fiber.Map{
		"success": true,
		"data": fiber.Map{
			"total_students":       totalStudents,
			"total_teachers":       totalTeachers,
			"total_classes":        totalClasses,
			"active_schedules":     activeSchedules,
			"total_banks":          totalBanks,
			"total_subjects":       totalSubjects,
			"total_class_subjects": totalClassSubjects,
		},
	})
}

func (h *Handlers) HandleGetClasses(c *fiber.Ctx) error {
	var classes []domain.ClassRoom
	h.repo.DB.Order("grade ASC, name ASC").Find(&classes)
	return c.JSON(fiber.Map{"success": true, "data": classes})
}

type CreateClassRequest struct {
	Name  string `json:"name"`
	Grade string `json:"grade"`
	Major string `json:"major"`
}

func (h *Handlers) HandleCreateClass(c *fiber.Ctx) error {
	var req CreateClassRequest
	if err := c.BodyParser(&req); err != nil || req.Name == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "message": "Nama kelas wajib diisi"})
	}
	class := domain.ClassRoom{
		ID:        uuid.New(),
		Name:      req.Name,
		Grade:     req.Grade,
		Major:     req.Major,
		CreatedAt: time.Now(),
	}
	if err := h.repo.DB.Create(&class).Error; err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "message": err.Error()})
	}
	return c.JSON(fiber.Map{"success": true, "data": class, "message": "Kelas berhasil ditambahkan"})
}

func (h *Handlers) HandleUpdateClass(c *fiber.Ctx) error {
	classID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "message": "ID kelas tidak valid"})
	}
	var class domain.ClassRoom
	if err := h.repo.DB.First(&class, "id = ?", classID).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"success": false, "message": "Kelas tidak ditemukan"})
	}
	var req CreateClassRequest
	if err := c.BodyParser(&req); err != nil || strings.TrimSpace(req.Name) == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "message": "Nama kelas wajib diisi"})
	}
	class.Name = strings.TrimSpace(req.Name)
	class.Grade = strings.TrimSpace(req.Grade)
	class.Major = strings.TrimSpace(req.Major)
	if err := h.repo.DB.Save(&class).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false, "message": "Gagal memperbarui kelas"})
	}
	return c.JSON(fiber.Map{"success": true, "data": class, "message": "Data kelas berhasil diperbarui"})
}

func (h *Handlers) HandleDeleteClass(c *fiber.Ctx) error {
	classID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "message": "ID kelas tidak valid"})
	}
	var studentCount int64
	h.repo.DB.Model(&domain.StudentProfile{}).Where("class_room_id = ?", classID).Count(&studentCount)
	if studentCount > 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": fmt.Sprintf("Kelas tidak dapat dihapus karena masih ada %d siswa terdaftar. Pindahkan atau hapus siswa terlebih dahulu.", studentCount),
		})
	}
	var scheduleCount int64
	h.repo.DB.Model(&domain.ExamSchedule{}).Where("class_room_id = ?", classID).Count(&scheduleCount)
	if scheduleCount > 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": fmt.Sprintf("Kelas tidak dapat dihapus karena masih ada %d jadwal ujian terhubung. Hapus jadwal terlebih dahulu.", scheduleCount),
		})
	}
	if err := h.repo.DB.Delete(&domain.ClassRoom{}, "id = ?", classID).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false, "message": "Gagal menghapus kelas"})
	}
	return c.JSON(fiber.Map{"success": true, "message": "Kelas berhasil dihapus"})
}

func (h *Handlers) HandleGetStudents(c *fiber.Ctx) error {
	var profiles []domain.StudentProfile
	h.repo.DB.Preload("User").Preload("ClassRoom").Find(&profiles)
	return c.JSON(fiber.Map{"success": true, "data": profiles})
}

type CreateStudentRequest struct {
	Username string    `json:"username"`
	Password string    `json:"password"`
	FullName string    `json:"full_name"`
	NIS      string    `json:"nis"`
	NISN     string    `json:"nisn"`
	ClassID  uuid.UUID `json:"class_id"`
	Gender   string    `json:"gender"`
}

func (h *Handlers) HandleCreateStudent(c *fiber.Ctx) error {
	var req CreateStudentRequest
	if err := c.BodyParser(&req); err != nil || req.Username == "" || req.NIS == "" || req.ClassID == uuid.Nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "message": "Data siswa tidak lengkap"})
	}

	pass := req.Password
	if pass == "" {
		pass = "siswa123"
	}

	userID := uuid.New()
	user := domain.User{
		ID:           userID,
		Username:     req.Username,
		PasswordHash: repository.HashPassword(pass),
		FullName:     req.FullName,
		Role:         domain.RoleSiswa,
		IsActive:     true,
		CreatedAt:    time.Now(),
	}
	if err := h.repo.DB.Create(&user).Error; err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "message": "Username sudah terdaftar"})
	}

	profile := domain.StudentProfile{
		ID:          uuid.New(),
		UserID:      userID,
		NIS:         req.NIS,
		NISN:        req.NISN,
		ClassRoomID: req.ClassID,
		Gender:      req.Gender,
		CreatedAt:   time.Now(),
	}
	h.repo.DB.Create(&profile)

	return c.JSON(fiber.Map{"success": true, "message": "Akun siswa berhasil dibuat"})
}

type UpdateStudentRequest struct {
	Username string    `json:"username"`
	Password string    `json:"password"`
	FullName string    `json:"full_name"`
	NIS      string    `json:"nis"`
	NISN     string    `json:"nisn"`
	ClassID  uuid.UUID `json:"class_id"`
	Gender   string    `json:"gender"`
}

func (h *Handlers) HandleUpdateStudent(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "message": "ID tidak valid"})
	}

	var req UpdateStudentRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "message": "Data tidak valid"})
	}

	var profile domain.StudentProfile
	if err := h.repo.DB.First(&profile, "id = ? OR user_id = ?", id, id).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"success": false, "message": "Data siswa tidak ditemukan"})
	}

	var user domain.User
	if err := h.repo.DB.First(&user, "id = ?", profile.UserID).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"success": false, "message": "User siswa tidak ditemukan"})
	}
	// Akun yang sudah berganti role (guru/staf/admin) tidak boleh diambil alih lewat jalur siswa.
	if user.Role != domain.RoleSiswa && !h.callerCanGrant(c) {
		h.forbidden(c, "Akses ditolak: hanya administrator yang dapat mengubah akun guru dan staf")
		return nil
	}

	if req.FullName != "" {
		user.FullName = req.FullName
	}
	if req.Username != "" && req.Username != user.Username {
		var exists int64
		h.repo.DB.Model(&domain.User{}).Where("username = ? AND id != ?", req.Username, user.ID).Count(&exists)
		if exists > 0 {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "message": "Username sudah digunakan oleh akun lain"})
		}
		user.Username = req.Username
	}
	if req.Password != "" {
		user.PasswordHash = repository.HashPassword(req.Password)
	}
	h.repo.DB.Save(&user)

	if req.NIS != "" {
		profile.NIS = req.NIS
	}
	if req.NISN != "" {
		profile.NISN = req.NISN
	}
	if req.ClassID != uuid.Nil {
		profile.ClassRoomID = req.ClassID
	}
	if req.Gender != "" {
		profile.Gender = req.Gender
	}
	h.repo.DB.Save(&profile)

	h.repo.DB.Preload("User").Preload("ClassRoom").First(&profile, "id = ?", profile.ID)

	return c.JSON(fiber.Map{"success": true, "data": profile, "message": "Data siswa berhasil diperbarui"})
}

func (h *Handlers) HandleDeleteStudent(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "message": "ID tidak valid"})
	}

	var profile domain.StudentProfile
	if err := h.repo.DB.First(&profile, "id = ? OR user_id = ?", id, id).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"success": false, "message": "Data siswa tidak ditemukan"})
	}

	// Profil siswa yang akunnya sudah berganti role (guru/staf/admin) tidak boleh dihapus
	// lewat jalur siswa oleh pemanggil tanpa izin penuh.
	if h.denyPrivilegedTarget(c, profile.UserID) {
		return nil
	}

	userID := profile.UserID
	h.repo.DB.Delete(&profile)
	h.repo.DB.Where("user_id = ?", userID).Delete(&domain.EventParticipant{})
	h.repo.DB.Delete(&domain.User{}, "id = ?", userID)

	return c.JSON(fiber.Map{"success": true, "message": "Data siswa berhasil dihapus"})
}

func (h *Handlers) HandleGetTeachers(c *fiber.Ctx) error {
	roleQuery := strings.ToUpper(strings.TrimSpace(c.Query("role")))
	db := h.repo.DB.Model(&domain.User{})
	if roleQuery == "ADMIN" {
		db = db.Where("role = ?", domain.RoleAdmin)
	} else if roleQuery == "GURU" {
		db = db.Where("role = ?", domain.RoleGuru)
	} else {
		db = db.Where("role IN ?", []domain.Role{domain.RoleGuru, domain.RoleAdmin})
	}
	var teachers []domain.User
	db.Order("role ASC, full_name ASC").Find(&teachers)
	return c.JSON(fiber.Map{"success": true, "data": teachers})
}

type CreateTeacherRequest struct {
	Username    string      `json:"username"`
	Password    string      `json:"password"`
	FullName    string      `json:"full_name"`
	Role        domain.Role `json:"role"`
	Permissions []string    `json:"permissions"`
}

func (h *Handlers) HandleCreateTeacher(c *fiber.Ctx) error {
	var req CreateTeacherRequest
	if err := c.BodyParser(&req); err != nil || req.Username == "" || req.FullName == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "message": "Data guru / staf tidak lengkap"})
	}
	pass := req.Password
	if pass == "" {
		pass = "guru123"
	}
	role := req.Role
	if role == "" {
		role = domain.RoleGuru
	}
	if role != domain.RoleGuru && role != domain.RoleAdmin {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "message": "Role guru / staf tidak valid"})
	}
	canGrant := h.callerCanGrant(c)
	if role == domain.RoleAdmin && !canGrant {
		h.forbidden(c, "Akses ditolak: hanya administrator yang dapat membuat akun administrator")
		return nil
	}
	// Izin khusus hanya dapat diatur oleh pemegang izin penuh (selain itu 403); tanpa daftar izin
	// dipakai template guru. Akun GURU selalu menyimpan izin secara eksplisit.
	perms, err := service.ResolveStaffPermissionsOnCreate(role, canGrant, req.Permissions)
	if err != nil {
		return h.respondPermissionError(c, err)
	}
	user := domain.User{
		ID:           uuid.New(),
		Username:     req.Username,
		PasswordHash: repository.HashPassword(pass),
		FullName:     req.FullName,
		Role:         role,
		Permissions:  perms,
		IsActive:     true,
		CreatedAt:    time.Now(),
	}
	if err := h.repo.DB.Create(&user).Error; err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "message": "Username sudah digunakan"})
	}
	return c.JSON(fiber.Map{"success": true, "message": "Akun guru / staf berhasil ditambahkan"})
}

func (h *Handlers) HandleUpdateTeacher(c *fiber.Ctx) error {
	teacherID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "message": "ID guru/staf tidak valid"})
	}
	var user domain.User
	if err := h.repo.DB.First(&user, "id = ?", teacherID).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"success": false, "message": "Akun guru/staf tidak ditemukan"})
	}
	var req CreateTeacherRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "message": "Data tidak valid"})
	}
	canGrant := h.callerCanGrant(c)
	if user.Role == domain.RoleAdmin && !canGrant {
		h.forbidden(c, "Akses ditolak: hanya administrator yang dapat mengubah akun administrator")
		return nil
	}
	// Pemanggil tanpa izin penuh hanya boleh mengubah nama akun guru/staf. Mengganti username
	// atau kata sandi sama dengan mengambil alih akun yang izinnya bisa lebih besar.
	if !canGrant && user.Role != domain.RoleSiswa &&
		(req.Password != "" || (req.Username != "" && req.Username != user.Username)) {
		h.forbidden(c, "Akses ditolak: hanya administrator yang dapat mengubah username dan kata sandi akun guru dan staf")
		return nil
	}
	if req.FullName != "" {
		user.FullName = req.FullName
	}
	if req.Username != "" && req.Username != user.Username {
		var exists int64
		h.repo.DB.Model(&domain.User{}).Where("username = ? AND id != ?", req.Username, user.ID).Count(&exists)
		if exists > 0 {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "message": "Username sudah digunakan"})
		}
		user.Username = req.Username
	}
	prevRole := user.Role
	if req.Role != "" && req.Role != user.Role {
		if !canGrant {
			h.forbidden(c, "Akses ditolak: hanya administrator yang dapat mengubah role")
			return nil
		}
		if req.Role != domain.RoleGuru && req.Role != domain.RoleAdmin {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "message": "Role guru / staf tidak valid"})
		}
		user.Role = req.Role
	}
	if prevRole == domain.RoleAdmin && user.Role != domain.RoleAdmin {
		if h.denyLastAdmin(c, user.ID, "Tidak dapat menurunkan administrator terakhir") {
			return nil
		}
	}
	// Izin lama tidak terbawa saat role berubah (mis. "*" setelah admin diturunkan); akun GURU
	// selalu menyimpan izin eksplisit, daftar izin kosong ditolak (400), dan non-grantor yang
	// mengirim daftar izin ditolak (403).
	perms, err := service.ResolveStaffPermissionsOnUpdate(prevRole, user.Role, user.Permissions, canGrant, req.Permissions)
	if err != nil {
		return h.respondPermissionError(c, err)
	}
	user.Permissions = perms
	if req.Password != "" {
		user.PasswordHash = repository.HashPassword(req.Password)
	}
	if err := h.repo.DB.Save(&user).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false, "message": "Gagal memperbarui data"})
	}
	return c.JSON(fiber.Map{"success": true, "data": user, "message": "Akun guru/staf berhasil diperbarui"})
}

func (h *Handlers) HandleDeleteTeacher(c *fiber.Ctx) error {
	teacherID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "message": "ID tidak valid"})
	}
	if h.denyPrivilegedTarget(c, teacherID) {
		return nil
	}
	currentAdmin, _ := middleware.GetCurrentUser(c)
	if currentAdmin.ID == teacherID {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "message": "Anda tidak dapat menghapus akun Anda sendiri"})
	}
	if h.denyLastAdmin(c, teacherID, "Tidak dapat menghapus administrator terakhir") {
		return nil
	}
	if err := h.accessService.DeleteUserWithProctorAssignments(teacherID); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false, "message": "Gagal menghapus akun guru/staf"})
	}
	return c.JSON(fiber.Map{"success": true, "message": "Akun guru/staf berhasil dihapus"})
}

// ---------------- SUBJECT (MATA PELAJARAN) HANDLERS ----------------

type CreateSubjectRequest struct {
	Code string `json:"code"`
	Name string `json:"name"`
}

func (h *Handlers) HandleGetSubjects(c *fiber.Ctx) error {
	var subjects []domain.Subject
	h.repo.DB.Order("name ASC").Find(&subjects)
	return c.JSON(fiber.Map{"success": true, "data": subjects})
}

func (h *Handlers) HandleCreateSubject(c *fiber.Ctx) error {
	var req CreateSubjectRequest
	if err := c.BodyParser(&req); err != nil || strings.TrimSpace(req.Code) == "" || strings.TrimSpace(req.Name) == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "message": "Kode dan nama mata pelajaran wajib diisi"})
	}
	subject := domain.Subject{
		ID:        uuid.New(),
		Code:      strings.ToUpper(strings.TrimSpace(req.Code)),
		Name:      strings.TrimSpace(req.Name),
		CreatedAt: time.Now(),
	}
	if err := h.repo.DB.Create(&subject).Error; err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "message": "Kode mata pelajaran sudah digunakan"})
	}
	return c.JSON(fiber.Map{"success": true, "data": subject, "message": "Mata pelajaran berhasil ditambahkan"})
}

func (h *Handlers) HandleUpdateSubject(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "message": "ID mata pelajaran tidak valid"})
	}
	var req CreateSubjectRequest
	if err := c.BodyParser(&req); err != nil || strings.TrimSpace(req.Code) == "" || strings.TrimSpace(req.Name) == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "message": "Kode dan nama mata pelajaran wajib diisi"})
	}
	var subject domain.Subject
	if err := h.repo.DB.First(&subject, "id = ?", id).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"success": false, "message": "Mata pelajaran tidak ditemukan"})
	}
	subject.Code = strings.ToUpper(strings.TrimSpace(req.Code))
	subject.Name = strings.TrimSpace(req.Name)
	if err := h.repo.DB.Save(&subject).Error; err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "message": err.Error()})
	}
	return c.JSON(fiber.Map{"success": true, "data": subject, "message": "Mata pelajaran berhasil diperbarui"})
}

func (h *Handlers) HandleDeleteSubject(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "message": "ID tidak valid"})
	}
	var count int64
	h.repo.DB.Model(&domain.QuestionBank{}).Where("subject_id = ?", id).Count(&count)
	if count > 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": fmt.Sprintf("Mata pelajaran tidak dapat dihapus karena terhubung dengan %d bank soal", count),
		})
	}
	var classSubCount int64
	h.repo.DB.Model(&domain.ClassSubject{}).Where("subject_id = ?", id).Count(&classSubCount)
	if classSubCount > 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": fmt.Sprintf("Mata pelajaran tidak dapat dihapus karena terhubung dengan %d alokasi kelas mapel", classSubCount),
		})
	}
	if err := h.repo.DB.Delete(&domain.Subject{}, "id = ?", id).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false, "message": "Gagal menghapus mata pelajaran"})
	}
	return c.JSON(fiber.Map{"success": true, "message": "Mata pelajaran berhasil dihapus"})
}

// ---------------- CLASS SUBJECT (KELAS MAPEL) HANDLERS ----------------

type CreateClassSubjectRequest struct {
	ClassID      uuid.UUID `json:"class_id"`
	SubjectID    uuid.UUID `json:"subject_id"`
	TeacherID    uuid.UUID `json:"teacher_id"`
	AcademicYear string    `json:"academic_year"`
}

func (h *Handlers) HandleGetClassSubjects(c *fiber.Ctx) error {
	caller, err := middleware.GetCurrentUser(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"success": false, "message": "Autentikasi diperlukan"})
	}

	var classSubjects []domain.ClassSubject
	query := h.repo.DB.Preload("ClassRoom").Preload("Subject").Preload("Teacher")

	if classIDStr := c.Query("class_id"); classIDStr != "" {
		if cid, err := uuid.Parse(classIDStr); err == nil {
			query = query.Where("class_room_id = ?", cid)
		}
	}
	if subjectIDStr := c.Query("subject_id"); subjectIDStr != "" {
		if sid, err := uuid.Parse(subjectIDStr); err == nil {
			query = query.Where("subject_id = ?", sid)
		}
	}

	query.Order("academic_year DESC, created_at DESC").Find(&classSubjects)

	// Endpoint ini terbuka bagi seluruh staf: izin guru tidak pernah dikirim, dan username guru
	// hanya untuk pengelola data master atau akun (tab alokasi di frontend memakainya).
	service.RedactClassSubjectTeachers(classSubjects, service.CanSeeTeacherUsername(caller))
	return c.JSON(fiber.Map{"success": true, "data": classSubjects})
}

func (h *Handlers) HandleCreateClassSubject(c *fiber.Ctx) error {
	var req CreateClassSubjectRequest
	if err := c.BodyParser(&req); err != nil || req.ClassID == uuid.Nil || req.SubjectID == uuid.Nil || req.TeacherID == uuid.Nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "message": "Kelas, Mata Pelajaran, dan Guru Pengampu wajib dipilih"})
	}

	acadYear := strings.TrimSpace(req.AcademicYear)
	if acadYear == "" {
		acadYear = "2026/2027"
	}

	var existing domain.ClassSubject
	if err := h.repo.DB.Where("class_room_id = ? AND subject_id = ? AND academic_year = ?", req.ClassID, req.SubjectID, acadYear).First(&existing).Error; err == nil {
		existing.TeacherID = req.TeacherID
		h.repo.DB.Save(&existing)
		h.repo.DB.Preload("ClassRoom").Preload("Subject").Preload("Teacher").First(&existing, "id = ?", existing.ID)
		return c.JSON(fiber.Map{"success": true, "data": existing, "message": "Alokasi guru pengampu berhasil diperbarui"})
	}

	item := domain.ClassSubject{
		ID:           uuid.New(),
		ClassRoomID:  req.ClassID,
		SubjectID:    req.SubjectID,
		TeacherID:    req.TeacherID,
		AcademicYear: acadYear,
		CreatedAt:    time.Now(),
	}
	if err := h.repo.DB.Create(&item).Error; err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "message": err.Error()})
	}
	h.repo.DB.Preload("ClassRoom").Preload("Subject").Preload("Teacher").First(&item, "id = ?", item.ID)

	return c.JSON(fiber.Map{"success": true, "data": item, "message": "Alokasi kelas mapel berhasil ditambahkan"})
}

func (h *Handlers) HandleUpdateClassSubject(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "message": "ID tidak valid"})
	}

	var item domain.ClassSubject
	if err := h.repo.DB.First(&item, "id = ?", id).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"success": false, "message": "Alokasi kelas mapel tidak ditemukan"})
	}

	var req CreateClassSubjectRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "message": "Format data tidak valid"})
	}

	if req.ClassID != uuid.Nil {
		item.ClassRoomID = req.ClassID
	}
	if req.SubjectID != uuid.Nil {
		item.SubjectID = req.SubjectID
	}
	if req.TeacherID != uuid.Nil {
		item.TeacherID = req.TeacherID
	}
	if strings.TrimSpace(req.AcademicYear) != "" {
		item.AcademicYear = strings.TrimSpace(req.AcademicYear)
	}

	if err := h.repo.DB.Save(&item).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false, "message": "Gagal memperbarui alokasi kelas mapel"})
	}

	h.repo.DB.Preload("ClassRoom").Preload("Subject").Preload("Teacher").First(&item, "id = ?", item.ID)

	return c.JSON(fiber.Map{
		"success": true,
		"data":    item,
		"message": "Alokasi kelas mapel berhasil diperbarui",
	})
}

func (h *Handlers) HandleDeleteClassSubject(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "message": "ID tidak valid"})
	}
	if err := h.repo.DB.Delete(&domain.ClassSubject{}, "id = ?", id).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false, "message": "Gagal menghapus alokasi kelas mapel"})
	}
	return c.JSON(fiber.Map{"success": true, "message": "Alokasi kelas mapel berhasil dihapus"})
}

// ---------------- IMPORT/TEMPLATE HANDLERS (CLASSES, SUBJECTS, TEACHERS) ----------------

func (h *Handlers) HandleGetClassesTemplate(c *fiber.Ctx) error {
	buf, err := excel.GenerateClassesTemplate()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false, "message": "Gagal membuat template"})
	}
	c.Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Set("Content-Disposition", `attachment; filename="Template_Import_Kelas_CBT.xlsx"`)
	return c.Send(buf)
}

func (h *Handlers) HandleImportClassesExcel(c *fiber.Ctx) error {
	fileHeader, err := c.FormFile("file")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "message": "File Excel wajib diunggah"})
	}
	src, err := fileHeader.Open()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false, "message": "Gagal membaca file"})
	}
	defer src.Close()

	parsed, err := excel.ParseClassesFromExcel(src)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "message": err.Error()})
	}

	imported, skipped := 0, 0
	for _, cl := range parsed {
		var existing domain.ClassRoom
		if err := h.repo.DB.First(&existing, "name = ?", cl.Name).Error; err == nil {
			skipped++
			continue
		}
		item := domain.ClassRoom{
			ID:        uuid.New(),
			Name:      cl.Name,
			Grade:     cl.Grade,
			Major:     cl.Major,
			CreatedAt: time.Now(),
		}
		if err := h.repo.DB.Create(&item).Error; err == nil {
			imported++
		} else {
			skipped++
		}
	}
	return c.JSON(fiber.Map{
		"success":        true,
		"imported_count": imported,
		"skipped_count":  skipped,
		"message":        fmt.Sprintf("Berhasil mengimpor %d kelas (%d dilewati/duplikat)", imported, skipped),
	})
}

func (h *Handlers) HandleGetSubjectsTemplate(c *fiber.Ctx) error {
	buf, err := excel.GenerateSubjectsTemplate()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false, "message": "Gagal membuat template"})
	}
	c.Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Set("Content-Disposition", `attachment; filename="Template_Import_Mapel_CBT.xlsx"`)
	return c.Send(buf)
}

func (h *Handlers) HandleImportSubjectsExcel(c *fiber.Ctx) error {
	fileHeader, err := c.FormFile("file")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "message": "File Excel wajib diunggah"})
	}
	src, err := fileHeader.Open()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false, "message": "Gagal membaca file"})
	}
	defer src.Close()

	parsed, err := excel.ParseSubjectsFromExcel(src)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "message": err.Error()})
	}

	imported, skipped := 0, 0
	for _, s := range parsed {
		var existing domain.Subject
		if err := h.repo.DB.First(&existing, "code = ?", s.Code).Error; err == nil {
			skipped++
			continue
		}
		item := domain.Subject{
			ID:        uuid.New(),
			Code:      s.Code,
			Name:      s.Name,
			CreatedAt: time.Now(),
		}
		if err := h.repo.DB.Create(&item).Error; err == nil {
			imported++
		} else {
			skipped++
		}
	}
	return c.JSON(fiber.Map{
		"success":        true,
		"imported_count": imported,
		"skipped_count":  skipped,
		"message":        fmt.Sprintf("Berhasil mengimpor %d mata pelajaran (%d dilewati/duplikat)", imported, skipped),
	})
}

func (h *Handlers) HandleGetTeachersTemplate(c *fiber.Ctx) error {
	buf, err := excel.GenerateTeachersTemplate()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false, "message": "Gagal membuat template"})
	}
	c.Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Set("Content-Disposition", `attachment; filename="Template_Import_Guru_CBT.xlsx"`)
	return c.Send(buf)
}

func (h *Handlers) HandleImportTeachersExcel(c *fiber.Ctx) error {
	fileHeader, err := c.FormFile("file")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "message": "File Excel wajib diunggah"})
	}
	src, err := fileHeader.Open()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false, "message": "Gagal membaca file"})
	}
	defer src.Close()

	parsed, err := excel.ParseTeachersFromExcel(src)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "message": err.Error()})
	}

	imported, skipped := 0, 0
	for _, t := range parsed {
		var existing domain.User
		if err := h.repo.DB.First(&existing, "username = ?", t.Username).Error; err == nil {
			skipped++
			continue
		}
		role := domain.RoleGuru
		if t.Role == "ADMIN" {
			role = domain.RoleAdmin
		}
		perms, _ := service.ResolveStaffPermissionsOnCreate(role, role == domain.RoleAdmin, nil)
		user := domain.User{
			ID:           uuid.New(),
			Username:     t.Username,
			PasswordHash: repository.HashPassword(t.Password),
			FullName:     t.FullName,
			Role:         role,
			Permissions:  perms,
			IsActive:     true,
			CreatedAt:    time.Now(),
		}
		if err := h.repo.DB.Create(&user).Error; err == nil {
			imported++
		} else {
			skipped++
		}
	}
	return c.JSON(fiber.Map{
		"success":        true,
		"imported_count": imported,
		"skipped_count":  skipped,
		"message":        fmt.Sprintf("Berhasil mengimpor %d akun guru/staf (%d dilewati/duplikat)", imported, skipped),
	})
}
