package handler

import (
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"cbt-backend/internal/domain"
	"cbt-backend/internal/middleware"
	"cbt-backend/internal/repository"
	"cbt-backend/internal/service"
	"cbt-backend/pkg/excel"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ---------------- USER MANAGEMENT HANDLERS ----------------

type UserDetailResponse struct {
	ID               uuid.UUID   `json:"id"`
	Username         string      `json:"username"`
	FullName         string      `json:"full_name"`
	Role             domain.Role `json:"role"`
	Permissions      []string    `json:"permissions"`
	IsActive         bool        `json:"is_active"`
	HasActiveSession bool        `json:"has_active_session"`
	NIS              string      `json:"nis,omitempty"`
	NISN             string      `json:"nisn,omitempty"`
	ClassName        string      `json:"class_name,omitempty"`
	ClassID          *uuid.UUID  `json:"class_id,omitempty"`
	Gender           string      `json:"gender,omitempty"`
	CreatedAt        time.Time   `json:"created_at"`
}

func (h *Handlers) HandleGetUsers(c *fiber.Ctx) error {
	roleQuery := c.Query("role")
	searchQuery := c.Query("search")

	db := h.repo.DB.Model(&domain.User{})
	if roleQuery != "" {
		db = db.Where("role = ?", strings.ToUpper(roleQuery))
	}
	if searchQuery != "" {
		db = db.Where("username ILIKE ? OR full_name ILIKE ?", "%"+searchQuery+"%", "%"+searchQuery+"%")
	}

	var users []domain.User
	if err := db.Order("created_at DESC").Find(&users).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false, "message": err.Error()})
	}

	// Fetch student profiles map
	var profiles []domain.StudentProfile
	h.repo.DB.Preload("ClassRoom").Find(&profiles)
	profileMap := make(map[uuid.UUID]domain.StudentProfile)
	for _, p := range profiles {
		profileMap[p.UserID] = p
	}

	var results []UserDetailResponse
	for _, u := range users {
		item := UserDetailResponse{
			ID:               u.ID,
			Username:         u.Username,
			FullName:         u.FullName,
			Role:             u.Role,
			Permissions:      u.EffectivePermissions(),
			IsActive:         u.IsActive,
			HasActiveSession: u.SessionToken != "",
			CreatedAt:        u.CreatedAt,
		}
		if p, ok := profileMap[u.ID]; ok {
			item.NIS = p.NIS
			item.NISN = p.NISN
			item.ClassName = p.ClassRoom.Name
			item.ClassID = &p.ClassRoomID
			item.Gender = p.Gender
		}
		results = append(results, item)
	}

	return c.JSON(fiber.Map{"success": true, "data": results})
}

type ManageUserRequest struct {
	Username    string      `json:"username"`
	Password    string      `json:"password"`
	FullName    string      `json:"full_name"`
	Role        domain.Role `json:"role"`
	Permissions []string    `json:"permissions"`
	NIS         string      `json:"nis"`
	NISN        string      `json:"nisn"`
	ClassID     *uuid.UUID  `json:"class_id"`
	Gender      string      `json:"gender"`
}

func (h *Handlers) HandleCreateUser(c *fiber.Ctx) error {
	var req ManageUserRequest
	if err := c.BodyParser(&req); err != nil || req.Username == "" || req.FullName == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "message": "Username dan Nama Lengkap wajib diisi"})
	}

	role := req.Role
	if role == "" {
		role = domain.RoleSiswa
	}
	if role != domain.RoleSiswa && role != domain.RoleGuru && role != domain.RoleAdmin {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "message": "Role tidak valid"})
	}
	canGrant := h.callerCanGrant(c)
	if role != domain.RoleSiswa && !canGrant {
		h.forbidden(c, "Akses ditolak: hanya administrator yang dapat membuat akun guru, staf, atau administrator")
		return nil
	}
	perms, err := service.ResolveStaffPermissionsOnCreate(role, canGrant, req.Permissions)
	if err != nil {
		return h.respondPermissionError(c, err)
	}
	pass := req.Password
	if pass == "" {
		pass = "123456"
	}

	userID := uuid.New()
	user := domain.User{
		ID:           userID,
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

	if role == domain.RoleSiswa && req.ClassID != nil {
		nis := req.NIS
		if nis == "" {
			nis = req.Username
		}
		profile := domain.StudentProfile{
			ID:          uuid.New(),
			UserID:      userID,
			NIS:         nis,
			NISN:        req.NISN,
			ClassRoomID: *req.ClassID,
			Gender:      req.Gender,
			CreatedAt:   time.Now(),
		}
		h.repo.DB.Create(&profile)
	}

	return c.JSON(fiber.Map{"success": true, "message": "Pengguna baru berhasil ditambahkan"})
}

func (h *Handlers) HandleUpdateUser(c *fiber.Ctx) error {
	userID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "message": "ID tidak valid"})
	}

	var user domain.User
	if err := h.repo.DB.First(&user, "id = ?", userID).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"success": false, "message": "Pengguna tidak ditemukan"})
	}

	var req ManageUserRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "message": "Data tidak valid"})
	}

	canGrant := h.callerCanGrant(c)
	if !canGrant && (user.Role != domain.RoleSiswa || (req.Role != "" && req.Role != domain.RoleSiswa)) {
		h.forbidden(c, "Akses ditolak: hanya administrator yang dapat mengubah akun guru, staf, atau role")
		return nil
	}
	if req.Role != "" && req.Role != domain.RoleSiswa && req.Role != domain.RoleGuru && req.Role != domain.RoleAdmin {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "message": "Role tidak valid"})
	}

	if req.FullName != "" {
		user.FullName = req.FullName
	}
	if req.Username != "" && req.Username != user.Username {
		var exists int64
		if err := h.repo.DB.Model(&domain.User{}).Where("username = ? AND id != ?", req.Username, user.ID).Count(&exists).Error; err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false, "message": "Gagal memeriksa username"})
		}
		if exists > 0 {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "message": "Username sudah digunakan"})
		}
		user.Username = req.Username
	}
	prevRole := user.Role
	if req.Role != "" && req.Role != user.Role {
		user.Role = req.Role
	}
	if prevRole == domain.RoleAdmin && user.Role != domain.RoleAdmin {
		if h.denyLastAdmin(c, user.ID, "Tidak dapat menurunkan administrator terakhir") {
			return nil
		}
	}
	// Izin lama tidak terbawa saat role berubah; akun GURU selalu menyimpan izin eksplisit,
	// daftar izin kosong ditolak (400), dan non-grantor yang mengirim daftar izin ditolak (403).
	perms, err := service.ResolveStaffPermissionsOnUpdate(prevRole, user.Role, user.Permissions, canGrant, req.Permissions)
	if err != nil {
		return h.respondPermissionError(c, err)
	}
	user.Permissions = perms
	if req.Password != "" {
		user.PasswordHash = repository.HashPassword(req.Password)
	}

	// Akun dan profil siswa (bila ada) disimpan atomik agar gagal simpan tidak menyisakan data setengah jalan.
	saveErr := h.repo.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(&user).Error; err != nil {
			return err
		}
		if user.Role != domain.RoleSiswa {
			return nil
		}
		var profile domain.StudentProfile
		if err := tx.First(&profile, "user_id = ?", user.ID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil // siswa tanpa profil: tidak ada yang diperbarui
			}
			return err
		}
		if req.NIS != "" {
			profile.NIS = req.NIS
		}
		if req.NISN != "" {
			profile.NISN = req.NISN
		}
		if req.ClassID != nil {
			profile.ClassRoomID = *req.ClassID
		}
		if req.Gender != "" {
			profile.Gender = req.Gender
		}
		return tx.Save(&profile).Error
	})
	if saveErr != nil {
		log.Printf("gagal memperbarui pengguna %s: %v", user.ID, saveErr)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false, "message": "Gagal memperbarui data pengguna"})
	}

	return c.JSON(fiber.Map{"success": true, "message": "Data pengguna berhasil diperbarui"})
}

func (h *Handlers) HandleToggleUserStatus(c *fiber.Ctx) error {
	userID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "message": "ID tidak valid"})
	}
	if h.denyPrivilegedTarget(c, userID) {
		return nil
	}

	currentAdmin, _ := middleware.GetCurrentUser(c)
	if currentAdmin.ID == userID {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "message": "Anda tidak dapat menonaktifkan akun sendiri"})
	}

	var user domain.User
	if err := h.repo.DB.First(&user, "id = ?", userID).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"success": false, "message": "Pengguna tidak ditemukan"})
	}

	user.IsActive = !user.IsActive
	h.repo.DB.Save(&user)

	statusText := "diaktifkan"
	if !user.IsActive {
		statusText = "dinonaktifkan"
	}
	return c.JSON(fiber.Map{"success": true, "is_active": user.IsActive, "message": fmt.Sprintf("Akun %s berhasil %s", user.FullName, statusText)})
}

type ResetPasswordRequest struct {
	NewPassword string `json:"new_password"`
}

func (h *Handlers) HandleResetUserPassword(c *fiber.Ctx) error {
	userID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "message": "ID tidak valid"})
	}
	if h.denyPrivilegedTarget(c, userID) {
		return nil
	}

	var req ResetPasswordRequest
	if err := c.BodyParser(&req); err != nil || req.NewPassword == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "message": "Password baru wajib diisi"})
	}

	var user domain.User
	if err := h.repo.DB.First(&user, "id = ?", userID).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"success": false, "message": "Pengguna tidak ditemukan"})
	}

	user.PasswordHash = repository.HashPassword(req.NewPassword)
	user.SessionToken = "" // force re-login
	h.repo.DB.Save(&user)

	return c.JSON(fiber.Map{"success": true, "message": fmt.Sprintf("Password akun %s berhasil direset", user.Username)})
}

func (h *Handlers) HandleResetUserSession(c *fiber.Ctx) error {
	userID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "message": "ID tidak valid"})
	}
	if h.denyPrivilegedTarget(c, userID) {
		return nil
	}

	var user domain.User
	if err := h.repo.DB.First(&user, "id = ?", userID).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"success": false, "message": "Pengguna tidak ditemukan"})
	}

	user.SessionToken = ""
	h.repo.DB.Model(&user).Update("session_token", "")

	return c.JSON(fiber.Map{"success": true, "message": fmt.Sprintf("Sesi perangkat %s berhasil direset", user.Username)})
}

func (h *Handlers) HandleDeleteUser(c *fiber.Ctx) error {
	userID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "message": "ID tidak valid"})
	}
	if h.denyPrivilegedTarget(c, userID) {
		return nil
	}

	currentAdmin, _ := middleware.GetCurrentUser(c)
	if currentAdmin.ID == userID {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "message": "Tidak dapat menghapus akun Anda sendiri"})
	}
	if h.denyLastAdmin(c, userID, "Tidak dapat menghapus administrator terakhir") {
		return nil
	}

	if err := h.accessService.DeleteUserWithProctorAssignments(userID); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false, "message": "Gagal menghapus akun"})
	}

	return c.JSON(fiber.Map{"success": true, "message": "Akun pengguna berhasil dihapus"})
}

func (h *Handlers) HandleGetStudentTemplate(c *fiber.Ctx) error {
	file, err := excel.GenerateStudentTemplate()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString(err.Error())
	}
	c.Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Set("Content-Disposition", `attachment; filename="Template_Import_Siswa_CBT.xlsx"`)
	return file.Write(c.Response().BodyWriter())
}

func (h *Handlers) HandleImportStudentsExcel(c *fiber.Ctx) error {
	fileHeader, err := c.FormFile("file")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "message": "File Excel wajib diunggah"})
	}

	src, err := fileHeader.Open()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false, "message": "Gagal membaca file"})
	}
	defer src.Close()

	parsed, err := excel.ParseStudentsFromExcel(src)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "message": err.Error()})
	}

	// Fetch existing classes for name lookup
	var classes []domain.ClassRoom
	h.repo.DB.Find(&classes)
	classMap := make(map[string]uuid.UUID)
	for _, cl := range classes {
		classMap[strings.ToLower(strings.TrimSpace(cl.Name))] = cl.ID
	}

	var fallbackClassID uuid.UUID
	if len(classes) > 0 {
		fallbackClassID = classes[0].ID
	} else {
		// create default class
		fallbackClassID = uuid.New()
		h.repo.DB.Create(&domain.ClassRoom{
			ID:        fallbackClassID,
			Name:      "XII MIPA 1",
			Grade:     "XII",
			Major:     "MIPA",
			CreatedAt: time.Now(),
		})
	}

	importedCount := 0
	for _, st := range parsed {
		// Target class
		targetClassID := fallbackClassID
		if cid, found := classMap[strings.ToLower(st.ClassName)]; found {
			targetClassID = cid
		}

		// Check if user with this username already exists
		var existing domain.User
		if err := h.repo.DB.First(&existing, "username = ?", st.NIS).Error; err == nil {
			continue // skip duplicate
		}

		userID := uuid.New()
		user := domain.User{
			ID:           userID,
			Username:     st.NIS,
			PasswordHash: repository.HashPassword(st.Password),
			FullName:     st.FullName,
			Role:         domain.RoleSiswa,
			IsActive:     true,
			CreatedAt:    time.Now(),
		}
		if err := h.repo.DB.Create(&user).Error; err != nil {
			continue
		}

		profile := domain.StudentProfile{
			ID:          uuid.New(),
			UserID:      userID,
			NIS:         st.NIS,
			NISN:        st.NISN,
			ClassRoomID: targetClassID,
			Gender:      st.Gender,
			CreatedAt:   time.Now(),
		}
		h.repo.DB.Create(&profile)
		importedCount++
	}

	return c.JSON(fiber.Map{
		"success":        true,
		"imported_count": importedCount,
		"message":        fmt.Sprintf("Berhasil mengimpor %d akun siswa dari file Excel", importedCount),
	})
}
