package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
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

type Handlers struct {
	repo           *repository.Database
	authService    *service.AuthService
	examService    *service.ExamService
	proctorService *service.ProctorService
	accessService  *service.AccessService
	participants   *service.EventParticipantService
}

func NewHandlers(repo *repository.Database) *Handlers {
	return &Handlers{
		repo:           repo,
		authService:    service.NewAuthService(repo),
		examService:    service.NewExamService(repo),
		proctorService: service.NewProctorService(repo),
		accessService:  service.NewAccessService(repo),
		participants:   service.NewEventParticipantService(repo),
	}
}

// ---------------- AUTH HANDLERS ----------------

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func (h *Handlers) HandleLogin(c *fiber.Ctx) error {
	var req LoginRequest
	if err := c.BodyParser(&req); err != nil || req.Username == "" || req.Password == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Username dan password wajib diisi",
		})
	}

	ip := c.IP()
	userAgent := c.Get("User-Agent")

	res, err := h.authService.Login(req.Username, req.Password, ip, userAgent)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"message": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    res,
	})
}

func (h *Handlers) HandleGetMe(c *fiber.Ctx) error {
	user, err := middleware.GetCurrentUser(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"success": false, "message": "Unauthorized"})
	}

	user.Permissions = user.EffectivePermissions()

	resp := fiber.Map{
		"user":        user,
		"permissions": user.EffectivePermissions(),
	}

	if user.Role == domain.RoleSiswa {
		var profile domain.StudentProfile
		if err := h.repo.DB.Preload("ClassRoom").Where("user_id = ?", user.ID).First(&profile).Error; err == nil {
			resp["student_profile"] = profile
		}
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    resp,
	})
}

func (h *Handlers) HandleLogout(c *fiber.Ctx) error {
	user, _ := middleware.GetCurrentUser(c)
	h.repo.DB.Model(&user).Update("session_token", "")
	return c.JSON(fiber.Map{
		"success": true,
		"message": "Logout berhasil",
	})
}

// ---------------- STUDENT HANDLERS ----------------

func (h *Handlers) HandleGetStudentSchedules(c *fiber.Ctx) error {
	user, err := middleware.GetCurrentUser(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"success": false, "message": "Unauthorized"})
	}

	schedules, err := h.examService.GetStudentSchedules(user.ID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false, "message": err.Error()})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    schedules,
	})
}

type StartExamRequest struct {
	ScheduleID uuid.UUID `json:"schedule_id"`
	Token      string    `json:"token"`
}

func (h *Handlers) HandleStartExam(c *fiber.Ctx) error {
	user, err := middleware.GetCurrentUser(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"success": false, "message": "Unauthorized"})
	}

	var req StartExamRequest
	if err := c.BodyParser(&req); err != nil || req.ScheduleID == uuid.Nil || req.Token == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Jadwal dan Token ujian wajib diisi",
		})
	}

	ip := c.IP()
	userAgent := c.Get("User-Agent")

	payload, err := h.examService.StartOrResumeExam(user.ID, req.ScheduleID, req.Token, ip, userAgent)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    payload,
	})
}

type SyncAnswersRequest struct {
	SessionID uuid.UUID                `json:"session_id"`
	Answers   []service.SyncAnswerItem `json:"answers"`
}

func (h *Handlers) HandleSyncAnswers(c *fiber.Ctx) error {
	user, err := middleware.GetCurrentUser(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"success": false, "message": "Unauthorized"})
	}

	var req SyncAnswersRequest
	if err := c.BodyParser(&req); err != nil || req.SessionID == uuid.Nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Session ID dan data jawaban wajib disertakan",
		})
	}

	count, err := h.examService.SyncAnswers(req.SessionID, user.ID, req.Answers)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"synced":  count,
		"message": "Jawaban berhasil disinkronkan",
	})
}

type ViolationRequest struct {
	SessionID uuid.UUID `json:"session_id"`
	EventType string    `json:"event_type"`
	Details   string    `json:"details"`
}

func (h *Handlers) HandleRecordViolation(c *fiber.Ctx) error {
	user, err := middleware.GetCurrentUser(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"success": false, "message": "Unauthorized"})
	}

	var req ViolationRequest
	if err := c.BodyParser(&req); err != nil || req.SessionID == uuid.Nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "message": "Data tidak valid"})
	}

	count, isBlocked, err := h.examService.RecordViolation(req.SessionID, user.ID, req.EventType, req.Details)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "message": err.Error()})
	}

	return c.JSON(fiber.Map{
		"success":         true,
		"violation_count": count,
		"is_blocked":      isBlocked,
	})
}

type SubmitExamRequest struct {
	SessionID uuid.UUID `json:"session_id"`
}

func (h *Handlers) HandleSubmitExam(c *fiber.Ctx) error {
	user, err := middleware.GetCurrentUser(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"success": false, "message": "Unauthorized"})
	}

	var req SubmitExamRequest
	if err := c.BodyParser(&req); err != nil || req.SessionID == uuid.Nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "message": "Session ID wajib disertakan"})
	}

	score, err := h.examService.SubmitExam(req.SessionID, user.ID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "message": err.Error()})
	}

	return c.JSON(fiber.Map{
		"success":     true,
		"total_score": score,
		"message":     "Ujian telah berhasil dikumpulkan",
	})
}

// ---------------- PROCTOR & REPORT HANDLERS ----------------

func (h *Handlers) HandleGetProctorSchedules(c *fiber.Ctx) error {
	user, err := middleware.GetCurrentUser(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"success": false, "message": "Autentikasi diperlukan"})
	}

	// Cakupan lihat: pemegang proctor:view/control_all melihat semua jadwal aktif,
	// pemegang proctor:control saja hanya melihat jadwal yang ditugaskan kepadanya.
	// Hak kendali ditandai per jadwal (can_control).
	viewScope := h.accessService.ViewScopeFor(user)
	controlScope := viewScope
	if viewScope.All {
		// Pemegang proctor:view saja dapat melihat semua tetapi belum tentu mengendalikan semua.
		controlScope = h.accessService.ControlScopeFor(user)
	}

	items := make([]proctorScheduleItem, 0)
	if viewScope.All || len(viewScope.ScheduleIDs) > 0 {
		query := h.repo.DB.Preload("Bank").Preload("Bank.Subject").Preload("ClassRoom").
			Where("is_active = ?", true)
		if !viewScope.All {
			ids := make([]uuid.UUID, 0, len(viewScope.ScheduleIDs))
			for id := range viewScope.ScheduleIDs {
				ids = append(ids, id)
			}
			query = query.Where("id IN ?", ids)
		}

		var schedules []domain.ExamSchedule
		query.Order("start_time DESC").Find(&schedules)
		for _, sch := range schedules {
			items = append(items, proctorScheduleItem{ExamSchedule: sch, CanControl: controlScope.Allows(sch.ID)})
		}
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    items,
	})
}

func (h *Handlers) HandleGetLiveProctorData(c *fiber.Ctx) error {
	scheduleID, err := uuid.Parse(c.Params("schedule_id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "message": "ID jadwal tidak valid"})
	}

	if h.denyScheduleView(c, scheduleID) {
		return nil
	}

	data, err := h.proctorService.GetLiveProctorData(scheduleID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "message": err.Error()})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    data,
	})
}

type ExtendTimeRequest struct {
	SessionID    uuid.UUID `json:"session_id"`
	ExtraMinutes int       `json:"extra_minutes"`
	Reason       string    `json:"reason"`
}

func (h *Handlers) HandleExtendTimeSession(c *fiber.Ctx) error {
	var req ExtendTimeRequest
	_ = c.BodyParser(&req)

	sessionIDStr := c.Params("id")
	if sessionIDStr != "" {
		if id, err := uuid.Parse(sessionIDStr); err == nil {
			req.SessionID = id
		}
	}

	if req.SessionID == uuid.Nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "message": "ID sesi tidak valid"})
	}
	if req.ExtraMinutes <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "message": "Jumlah menit tambahan harus lebih dari 0"})
	}
	if h.denySessionControl(c, req.SessionID) {
		return nil
	}

	if err := h.proctorService.ExtendTimeSession(req.SessionID, req.ExtraMinutes, req.Reason); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "message": err.Error()})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": fmt.Sprintf("Waktu ujian berhasil ditambah %d menit", req.ExtraMinutes),
	})
}

type ExtendTimeAllRequest struct {
	ScheduleID   uuid.UUID `json:"schedule_id"`
	ExtraMinutes int       `json:"extra_minutes"`
	Reason       string    `json:"reason"`
}

func (h *Handlers) HandleExtendTimeAllSchedule(c *fiber.Ctx) error {
	var req ExtendTimeAllRequest
	_ = c.BodyParser(&req)

	scheduleIDStr := c.Params("id")
	if scheduleIDStr != "" {
		if id, err := uuid.Parse(scheduleIDStr); err == nil {
			req.ScheduleID = id
		}
	}

	if req.ScheduleID == uuid.Nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "message": "ID jadwal tidak valid"})
	}
	if req.ExtraMinutes <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "message": "Jumlah menit tambahan harus lebih dari 0"})
	}

	if h.denyScheduleControl(c, req.ScheduleID) {
		return nil
	}

	count, err := h.proctorService.ExtendTimeAllSchedule(req.ScheduleID, req.ExtraMinutes, req.Reason)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "message": err.Error()})
	}

	return c.JSON(fiber.Map{
		"success":        true,
		"affected_count": count,
		"message":        fmt.Sprintf("Waktu ujian berhasil ditambah %d menit untuk %d siswa aktif", req.ExtraMinutes, count),
	})
}

func (h *Handlers) HandleForceSubmitSession(c *fiber.Ctx) error {
	sessionID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		// Fallback body
		type Body struct {
			SessionID uuid.UUID `json:"session_id"`
		}
		var b Body
		if err := c.BodyParser(&b); err == nil && b.SessionID != uuid.Nil {
			sessionID = b.SessionID
		} else {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "message": "ID sesi tidak valid"})
		}
	}

	if h.denySessionControl(c, sessionID) {
		return nil
	}

	score, err := h.proctorService.ForceSubmitSession(sessionID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "message": err.Error()})
	}

	return c.JSON(fiber.Map{
		"success":     true,
		"total_score": score,
		"message":     "Ujian siswa berhasil dikumpulkan paksa dan nilai telah dihitung",
	})
}

func (h *Handlers) HandleGetSessionViolations(c *fiber.Ctx) error {
	sessionID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "message": "ID sesi tidak valid"})
	}

	if h.denySessionView(c, sessionID) {
		return nil
	}

	logs, err := h.proctorService.GetViolationLogs(sessionID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "message": err.Error()})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    logs,
	})
}

type UnlockRequest struct {
	SessionID uuid.UUID `json:"session_id"`
}

type ResetDeviceRequest struct {
	StudentID uuid.UUID `json:"student_id"`
}

func (h *Handlers) HandleUnlockStudentSession(c *fiber.Ctx) error {
	sessionID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		var req UnlockRequest
		if err := c.BodyParser(&req); err == nil && req.SessionID != uuid.Nil {
			sessionID = req.SessionID
		} else {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "message": "ID sesi tidak valid"})
		}
	}

	if h.denySessionControl(c, sessionID) {
		return nil
	}

	if err := h.proctorService.UnlockStudentSession(sessionID); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "message": err.Error()})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Sesi siswa berhasil dibuka kuncinya",
	})
}

func (h *Handlers) HandleResetStudentSession(c *fiber.Ctx) error {
	userID, err := uuid.Parse(c.Params("user_id"))
	if err != nil {
		var req ResetDeviceRequest
		if err := c.BodyParser(&req); err == nil && req.StudentID != uuid.Nil {
			userID = req.StudentID
		} else {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "message": "ID siswa tidak valid"})
		}
	}

	if h.denyStudentControl(c, userID) {
		return nil
	}

	if err := h.proctorService.ResetStudentDeviceSession(userID); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "message": err.Error()})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Sesi perangkat siswa berhasil direset. Siswa dapat login kembali.",
	})
}

func (h *Handlers) HandleExportGradesExcel(c *fiber.Ctx) error {
	scheduleID, err := uuid.Parse(c.Params("schedule_id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).SendString("Invalid schedule ID")
	}

	if h.denyScheduleExport(c, scheduleID) {
		return nil
	}

	file, filename, err := h.proctorService.ExportClassGrades(scheduleID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).SendString(err.Error())
	}

	c.Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))

	return file.Write(c.Response().BodyWriter())
}

func (h *Handlers) HandleExportBeritaAcaraPDF(c *fiber.Ctx) error {
	scheduleID, err := uuid.Parse(c.Params("schedule_id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).SendString("Invalid schedule ID")
	}

	if h.denyScheduleExport(c, scheduleID) {
		return nil
	}

	pdfBytes, filename, err := h.proctorService.ExportBeritaAcara(
		scheduleID,
		c.Query("spv1"),
		c.Query("spv2"),
		c.Query("proctor"),
		c.Query("room"),
	)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).SendString(err.Error())
	}

	c.Set("Content-Type", "application/pdf")
	c.Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))

	return c.Send(pdfBytes)
}

// ---------------- ADMIN & QUESTION BANK HANDLERS ----------------

func (h *Handlers) HandleGetQuestionBankTemplate(c *fiber.Ctx) error {
	file, err := excel.GenerateQuestionTemplate()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString(err.Error())
	}

	c.Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Set("Content-Disposition", `attachment; filename="Template_Bank_Soal_CBT.xlsx"`)

	return file.Write(c.Response().BodyWriter())
}

func (h *Handlers) HandleImportQuestionsExcel(c *fiber.Ctx) error {
	bankID, err := uuid.Parse(c.FormValue("bank_id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "message": "ID Bank Soal wajib disertakan"})
	}
	if h.denyBank(c, bankID) {
		return nil
	}

	fileHeader, err := c.FormFile("file")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "message": "File Excel tidak ditemukan dalam request"})
	}

	src, err := fileHeader.Open()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false, "message": "Gagal membaca file"})
	}
	defer src.Close()

	parsed, err := excel.ParseQuestionsFromExcel(src)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "message": err.Error()})
	}

	questions, err := excel.ConvertParsedToQuestions(bankID, parsed)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false, "message": err.Error()})
	}

	// Save to DB
	for _, q := range questions {
		h.repo.DB.Create(&q)
	}

	// Update bank count
	h.repo.DB.Model(&domain.QuestionBank{}).Where("id = ?", bankID).Update("total_questions", len(questions))

	return c.JSON(fiber.Map{
		"success":        true,
		"imported_count": len(questions),
		"message":        fmt.Sprintf("Berhasil mengimpor %d butir soal ke dalam bank soal.", len(questions)),
	})
}

func (h *Handlers) HandleGetReadinessMatrix(c *fiber.Ctx) error {
	user, err := middleware.GetCurrentUser(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"success": false, "message": "Autentikasi diperlukan"})
	}
	var banks []domain.QuestionBank
	var schedules []domain.ExamSchedule
	assignedSchedules := make([]fiber.Map, 0)

	hasReadAll := user.HasPermission(string(domain.PermQuestionsAll))
	eventIDQuery := strings.TrimSpace(c.Query("event_id"))

	// Ambil daftar event ujian terdaftar untuk selector filter di UI
	var events []domain.ExamEvent
	h.repo.DB.Order("created_at DESC").Find(&events)

	if !hasReadAll {
		// 1. Bank soal: hanya yang dibuat pengguna ini (mengampu mapel tidak membuka bank orang lain)
		h.repo.DB.Preload("Subject").Preload("CreatedBy").Preload("Classes").
			Where("created_by_id = ?", user.ID).
			Order("created_at DESC").
			Find(&banks)

		// 2. Jadwal ujian: aturan keterlihatan yang sama dengan daftar jadwal admin
		// (AccessService.ScheduleVisibilityFor): semua jadwal untuk pemegang izin lihat-semua,
		// selain itu pasangan kelas-mapel yang diampu, bank buatan sendiri, atau penugasan pengawas.
		visible, visErr := h.accessService.ScheduleVisibilityFor(user)
		if visErr != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false, "message": "Gagal memuat jadwal ujian"})
		}
		schedQuery := h.repo.DB.Preload("Event").Preload("Subject").Preload("Bank").Preload("Bank.Subject").Preload("Bank.CreatedBy").Preload("ClassRoom")
		if eventIDQuery != "" {
			schedQuery = schedQuery.Where("event_id = ?", eventIDQuery)
		}
		var allSchedules []domain.ExamSchedule
		schedQuery.Order("start_time ASC, created_at DESC").Find(&allSchedules)

		for _, sch := range allSchedules {
			if visible.Allows(sch.ID) {
				schedules = append(schedules, sch)
			}
		}
	} else {
		// Pengguna memiliki hak akses questions:read_all atau RoleAdmin: Ambil seluruh data secara holistik
		bankQuery := h.repo.DB.Preload("Subject").Preload("CreatedBy").Preload("Classes").Order("created_at DESC")
		schedQuery := h.repo.DB.Preload("Event").Preload("Subject").Preload("Bank").Preload("Bank.Subject").Preload("Bank.CreatedBy").Preload("ClassRoom").
			Order("start_time ASC, created_at DESC")

		if eventIDQuery != "" {
			schedQuery = schedQuery.Where("event_id = ?", eventIDQuery)
		}

		bankQuery.Find(&banks)
		schedQuery.Find(&schedules)
	}

	// Token ujian hanya dikirim untuk jadwal dalam cakupan pemanggil (aturan sama dengan
	// daftar jadwal admin); cakupan dihitung sekali dan berlaku untuk "schedules"
	// maupun "assigned_schedules".
	h.accessService.TokenScopeFor(user).RedactTokens(schedules)

	// Petakan nama guru pengampu per (Kelas, Mapel)
	var allClassSubjects []domain.ClassSubject
	h.repo.DB.Preload("Teacher").Find(&allClassSubjects)
	teacherMap := make(map[string]string)
	for _, cs := range allClassSubjects {
		k := fmt.Sprintf("%s_%s", cs.ClassRoomID.String(), cs.SubjectID.String())
		if cs.Teacher.FullName != "" {
			teacherMap[k] = cs.Teacher.FullName
		}
	}

	for _, sch := range schedules {
		status := "NO_BANK"
		var bankTitle string
		var bankTotalQ int
		var bankLocked bool
		var bankID *uuid.UUID

		if sch.Bank != nil {
			bID := sch.Bank.ID
			bankID = &bID
			bankTitle = sch.Bank.Title
			bankTotalQ = sch.Bank.TotalQuestions
			bankLocked = sch.Bank.IsLocked
			if bankLocked && bankTotalQ > 0 {
				status = "READY"
			} else {
				status = "DRAFT"
			}
		}

		teacherName := "-"
		if sch.SubjectID != nil {
			k := fmt.Sprintf("%s_%s", sch.ClassRoomID.String(), sch.SubjectID.String())
			if tName, ok := teacherMap[k]; ok {
				teacherName = tName
			}
		}
		if teacherName == "-" && sch.Bank != nil && sch.Bank.CreatedBy.FullName != "" {
			teacherName = sch.Bank.CreatedBy.FullName
		}

		subjectName := "-"
		if sch.Subject != nil && sch.Subject.Name != "" {
			subjectName = sch.Subject.Name
		} else if sch.Bank != nil && sch.Bank.Subject.Name != "" {
			subjectName = sch.Bank.Subject.Name
		}

		eventTitle := "Reguler / Harian"
		if sch.Event != nil && sch.Event.Title != "" {
			eventTitle = sch.Event.Title
		}

		assignedSchedules = append(assignedSchedules, fiber.Map{
			"id":               sch.ID,
			"title":            sch.Title,
			"event_title":      eventTitle,
			"subject_id":       sch.SubjectID,
			"subject_name":     subjectName,
			"class_id":         sch.ClassRoomID,
			"class_name":       sch.ClassRoom.Name,
			"teacher_name":     teacherName,
			"exam_token":       sch.ExamToken,
			"start_time":       sch.StartTime,
			"end_time":         sch.EndTime,
			"duration_minutes": sch.DurationMinutes,
			"is_active":        sch.IsActive,
			"bank_id":          bankID,
			"bank_title":       bankTitle,
			"total_questions":  bankTotalQ,
			"is_locked":        bankLocked,
			"status":           status, // NO_BANK, DRAFT, READY
		})
	}

	// Objek User penyusun tidak boleh ikut terkirim: pada jadwal dikosongkan seluruhnya
	// (nama guru sudah ada di "assigned_schedules[].teacher_name"), pada bank soal hanya
	// izinnya yang dikosongkan karena frontend memakai created_by.full_name sebagai nama penyusun.
	service.ClearBankOwners(schedules)
	service.StripBankOwnerPermissions(banks)

	return c.JSON(fiber.Map{
		"success": true,
		"data": fiber.Map{
			"question_banks":     banks,
			"schedules":          schedules,
			"assigned_schedules": assignedSchedules,
			"events":             events,
			"can_read_all":       hasReadAll,
		},
	})
}

type QuickCreateBankRequest struct {
	Title string `json:"title"`
}

func (h *Handlers) HandleQuickCreateAndLinkBank(c *fiber.Ctx) error {
	user, err := middleware.GetCurrentUser(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"success": false, "message": "Autentikasi diperlukan"})
	}

	schedID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "message": "ID jadwal tidak valid"})
	}

	var sched domain.ExamSchedule
	if err := h.repo.DB.Preload("Subject").Preload("ClassRoom").First(&sched, "id = ?", schedID).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"success": false, "message": "Jadwal ujian tidak ditemukan"})
	}

	// Jadwal di luar cakupan keterlihatan dibalas 404 agar keberadaannya tidak terungkap.
	visible, visErr := h.accessService.ScheduleVisibilityFor(user)
	if visErr != nil {
		log.Printf("gagal memeriksa keterlihatan jadwal %s: %v", schedID, visErr)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false, "message": "Gagal memeriksa hak akses jadwal"})
	}
	if !visible.Allows(schedID) {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"success": false, "message": "Jadwal ujian tidak ditemukan"})
	}

	if sched.SubjectID == nil || *sched.SubjectID == uuid.Nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "message": "Jadwal ini belum memiliki data mata pelajaran"})
	}

	// Quick-bank selalu mengubah tautan bank: tolak bila siswa sudah memulai ujian.
	// Gagal kueri = gagal tertutup (500) dan tidak ada bank yang dibuat.
	if err := h.accessService.EnsureScheduleWithoutSessions(schedID); err != nil {
		return h.respondBankChangeError(c, schedID, err)
	}

	var req QuickCreateBankRequest
	_ = c.BodyParser(&req)

	title := strings.TrimSpace(req.Title)
	if title == "" {
		subjName := "Mata Pelajaran"
		if sched.Subject != nil && sched.Subject.Name != "" {
			subjName = sched.Subject.Name
		}
		title = fmt.Sprintf("Naskah Soal - %s (%s)", subjName, sched.ClassRoom.Name)
	}

	bank := domain.QuestionBank{
		ID:             uuid.New(),
		SubjectID:      *sched.SubjectID,
		Title:          title,
		CreatedByID:    user.ID,
		TotalQuestions: 0,
		IsLocked:       false,
		CreatedAt:      time.Now(),
	}

	// Bank dibuat dan ditautkan dalam satu transaksi; penulisan tautan ikut memeriksa ulang
	// bahwa jadwal belum punya sesi siswa, sehingga bank tidak tersimpan bila siswa keburu memulai.
	if err := h.accessService.CreateAndLinkBankGuarded(schedID, &bank); err != nil {
		return h.respondBankChangeError(c, schedID, err)
	}
	sched.BankID = &bank.ID

	h.repo.DB.Preload("Subject").Preload("CreatedBy").First(&bank, "id = ?", bank.ID)

	// Jadwal sudah tersimpan; hanya salinan respons yang disunting sesuai cakupan token.
	h.accessService.TokenScopeFor(user).RedactToken(&sched)

	return c.JSON(fiber.Map{
		"success": true,
		"data": fiber.Map{
			"bank":     bank,
			"schedule": sched,
		},
		"message": fmt.Sprintf("Bank soal '%s' berhasil dibuat dan ditautkan ke jadwal ujian.", bank.Title),
	})
}

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

// ---------------- EXAM EVENT HANDLERS ----------------

type EventDetailResponse struct {
	domain.ExamEvent
	TotalSchedules int `json:"total_schedules"`
}

func (h *Handlers) HandleGetEvents(c *fiber.Ctx) error {
	var events []domain.ExamEvent
	if err := h.repo.DB.Order("start_date DESC").Find(&events).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false, "message": err.Error()})
	}

	type CountResult struct {
		EventID uuid.UUID `gorm:"column:event_id"`
		Count   int       `gorm:"column:count"`
	}
	var counts []CountResult
	h.repo.DB.Model(&domain.ExamSchedule{}).
		Select("event_id, count(*) as count").
		Group("event_id").
		Scan(&counts)

	countMap := make(map[uuid.UUID]int)
	for _, cnt := range counts {
		countMap[cnt.EventID] = cnt.Count
	}

	var results []EventDetailResponse
	for _, ev := range events {
		results = append(results, EventDetailResponse{
			ExamEvent:      ev,
			TotalSchedules: countMap[ev.ID],
		})
	}

	return c.JSON(fiber.Map{"success": true, "data": results})
}

type ManageEventRequest struct {
	Title        string `json:"title"`
	Code         string `json:"code"`
	AcademicYear string `json:"academic_year"`
	Semester     string `json:"semester"`
	StartDate    string `json:"start_date"`
	EndDate      string `json:"end_date"`
	IsActive     bool   `json:"is_active"`
	Description  string `json:"description"`
}

func (h *Handlers) HandleCreateEvent(c *fiber.Ctx) error {
	var req ManageEventRequest
	if err := c.BodyParser(&req); err != nil || req.Title == "" || req.Code == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "message": "Judul dan Kode Event wajib diisi"})
	}

	ay := req.AcademicYear
	if ay == "" {
		ay = "2026/2027"
	}
	sm := strings.ToUpper(req.Semester)
	if sm == "" {
		sm = "GANJIL"
	}

	startDate := time.Now()
	if t, err := time.Parse("2006-01-02", req.StartDate); err == nil {
		startDate = t
	}
	endDate := startDate.Add(14 * 24 * time.Hour)
	if t, err := time.Parse("2006-01-02", req.EndDate); err == nil {
		endDate = t.Add(23*time.Hour + 59*time.Minute + 59*time.Second)
	}

	event := domain.ExamEvent{
		ID:           uuid.New(),
		Title:        req.Title,
		Code:         strings.ToUpper(strings.TrimSpace(req.Code)),
		AcademicYear: ay,
		Semester:     sm,
		StartDate:    startDate,
		EndDate:      endDate,
		IsActive:     req.IsActive,
		Description:  req.Description,
		CreatedAt:    time.Now(),
	}

	if err := h.repo.DB.Create(&event).Error; err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "message": "Kode event sudah digunakan"})
	}

	return c.JSON(fiber.Map{"success": true, "data": event, "message": "Event ujian berhasil dibuat"})
}

func (h *Handlers) HandleUpdateEvent(c *fiber.Ctx) error {
	eventID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "message": "ID tidak valid"})
	}

	var event domain.ExamEvent
	if err := h.repo.DB.First(&event, "id = ?", eventID).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"success": false, "message": "Event tidak ditemukan"})
	}

	var req ManageEventRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "message": "Data tidak valid"})
	}

	if req.Title != "" {
		event.Title = req.Title
	}
	if req.Code != "" {
		event.Code = strings.ToUpper(strings.TrimSpace(req.Code))
	}
	if req.AcademicYear != "" {
		event.AcademicYear = req.AcademicYear
	}
	if req.Semester != "" {
		event.Semester = strings.ToUpper(req.Semester)
	}
	if req.Description != "" {
		event.Description = req.Description
	}
	if req.StartDate != "" {
		if t, err := time.Parse("2006-01-02", req.StartDate); err == nil {
			event.StartDate = t
		}
	}
	if req.EndDate != "" {
		if t, err := time.Parse("2006-01-02", req.EndDate); err == nil {
			event.EndDate = t.Add(23*time.Hour + 59*time.Minute + 59*time.Second)
		}
	}

	if err := h.repo.DB.Save(&event).Error; err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "message": err.Error()})
	}

	return c.JSON(fiber.Map{"success": true, "data": event, "message": "Data event berhasil diperbarui"})
}

func (h *Handlers) HandleToggleEventActive(c *fiber.Ctx) error {
	eventID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "message": "ID tidak valid"})
	}

	var event domain.ExamEvent
	if err := h.repo.DB.First(&event, "id = ?", eventID).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"success": false, "message": "Event tidak ditemukan"})
	}

	event.IsActive = !event.IsActive
	h.repo.DB.Save(&event)

	statusText := "diaktifkan (Sedang Berlangsung)"
	if !event.IsActive {
		statusText = "dinonaktifkan (Arsip)"
	}

	return c.JSON(fiber.Map{
		"success":   true,
		"is_active": event.IsActive,
		"message":   fmt.Sprintf("Event %s berhasil %s", event.Title, statusText),
	})
}

func (h *Handlers) HandleDeleteEvent(c *fiber.Ctx) error {
	eventID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "message": "ID tidak valid"})
	}

	var schedCount int64
	h.repo.DB.Model(&domain.ExamSchedule{}).Where("event_id = ?", eventID).Count(&schedCount)
	if schedCount > 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": fmt.Sprintf("Event tidak dapat dihapus karena masih memiliki %d sesi jadwal ujian", schedCount),
		})
	}

	h.repo.DB.Where("event_id = ?", eventID).Delete(&domain.EventParticipant{})
	if err := h.repo.DB.Delete(&domain.ExamEvent{}, "id = ?", eventID).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false, "message": "Gagal menghapus event"})
	}

	return c.JSON(fiber.Map{"success": true, "message": "Event berhasil dihapus"})
}

// ---------------- KARTU PESERTA EVENT ----------------

func (h *Handlers) HandleGetEventParticipants(c *fiber.Ctx) error {
	eventID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "message": "ID tidak valid"})
	}
	var classID *uuid.UUID
	if raw := c.Query("class_id"); raw != "" {
		if id, err := uuid.Parse(raw); err == nil {
			classID = &id
		}
	}
	cards, err := h.participants.List(eventID, classID)
	if errors.Is(err, service.ErrEventNotFound) {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"success": false, "message": "Event tidak ditemukan"})
	}
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false, "message": "Gagal memuat data peserta"})
	}
	return c.JSON(fiber.Map{"success": true, "data": cards})
}

type GenerateParticipantsRequest struct {
	Reset bool `json:"reset"`
}

func (h *Handlers) HandleGenerateEventParticipants(c *fiber.Ctx) error {
	eventID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "message": "ID tidak valid"})
	}
	var req GenerateParticipantsRequest
	_ = c.BodyParser(&req)
	created, err := h.participants.Generate(eventID, req.Reset)
	if errors.Is(err, service.ErrEventNotFound) {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"success": false, "message": "Event tidak ditemukan"})
	}
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false, "message": "Gagal membuat akun peserta"})
	}
	msg := fmt.Sprintf("%d akun peserta baru dibuat", created)
	if created == 0 {
		msg = "Semua peserta sudah memiliki nomor ujian"
	}
	return c.JSON(fiber.Map{"success": true, "created": created, "message": msg})
}

func (h *Handlers) HandleGetAdminSchedules(c *fiber.Ctx) error {
	caller, err := middleware.GetCurrentUser(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"success": false, "message": "Autentikasi diperlukan"})
	}

	eventIDStr := c.Query("event_id")
	query := h.repo.DB.Preload("Event").Preload("Subject").Preload("Bank").Preload("Bank.Subject").Preload("ClassRoom")

	if eventIDStr != "" {
		if eventID, err := uuid.Parse(eventIDStr); err == nil {
			query = query.Where("event_id = ?", eventID)
		}
	}

	var allSchedules []domain.ExamSchedule
	query.Order("start_time ASC, created_at DESC").Find(&allSchedules)

	// Guru hanya melihat jadwal yang diberikan kepadanya; cakupan dan relasi per jadwal
	// dihitung sekali per permintaan dari himpunan id yang sama (relasi berlaku untuk semua pemanggil).
	visible, relationsBySchedule, err := h.accessService.ScheduleVisibilityAndRelationsFor(caller)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false, "message": "Gagal memuat jadwal ujian"})
	}
	schedules := make([]domain.ExamSchedule, 0, len(allSchedules))
	for _, sch := range allSchedules {
		if visible.Allows(sch.ID) {
			schedules = append(schedules, sch)
		}
	}

	// Pengawas semua jadwal diambil dengan satu query IN, bukan per jadwal.
	ids := make([]uuid.UUID, 0, len(schedules))
	for _, sch := range schedules {
		ids = append(ids, sch.ID)
	}
	proctorsBySchedule, err := h.accessService.ProctorsBySchedule(ids)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false, "message": "Gagal memuat pengawas jadwal"})
	}

	// Token ujian hanya dikirim untuk jadwal dalam cakupan pemanggil; cakupan dihitung sekali.
	tokenScope := h.accessService.TokenScopeFor(caller)

	items := make([]adminScheduleItem, 0, len(schedules))
	for _, sch := range schedules {
		if !tokenScope.Allows(sch.ID) {
			sch.ExamToken = ""
		}
		proctors := make([]proctorRefItem, 0, len(proctorsBySchedule[sch.ID]))
		for _, p := range proctorsBySchedule[sch.ID] {
			proctors = append(proctors, proctorRefItem{ID: p.ID, FullName: p.FullName})
		}
		// Relasi tidak pernah null di JSON: jadwal tanpa relasi menghasilkan [].
		relations := relationsBySchedule[sch.ID]
		if relations == nil {
			relations = []string{}
		}
		items = append(items, adminScheduleItem{ExamSchedule: sch, Proctors: proctors, Relations: relations})
	}
	return c.JSON(fiber.Map{"success": true, "data": items})
}

func parseScheduleTimes(examDate, startTimeStr, endTimeStr string, defaultDuration int) (time.Time, time.Time) {
	now := time.Now()
	if strings.Contains(startTimeStr, "T") {
		if t, err := time.Parse(time.RFC3339, startTimeStr); err == nil {
			et := t.Add(time.Duration(defaultDuration) * time.Minute)
			if strings.Contains(endTimeStr, "T") {
				if t2, err2 := time.Parse(time.RFC3339, endTimeStr); err2 == nil {
					et = t2
				}
			}
			return t, et
		}
	}

	if examDate == "" {
		examDate = now.Format("2006-01-02")
	}
	if startTimeStr == "" {
		startTimeStr = "08:00"
	}
	if endTimeStr == "" {
		endTimeStr = "10:00"
	}

	startFull := fmt.Sprintf("%s %s", strings.TrimSpace(examDate), strings.TrimSpace(startTimeStr))
	endFull := fmt.Sprintf("%s %s", strings.TrimSpace(examDate), strings.TrimSpace(endTimeStr))

	st, err := time.ParseInLocation("2006-01-02 15:04", startFull, time.Local)
	if err != nil {
		st = now
	}
	et, err := time.ParseInLocation("2006-01-02 15:04", endFull, time.Local)
	if err != nil {
		et = st.Add(time.Duration(defaultDuration) * time.Minute)
	}
	if et.Before(st) {
		et = st.Add(time.Duration(defaultDuration) * time.Minute)
	}
	return st, et
}

type CreateScheduleRequest struct {
	EventID            *uuid.UUID `json:"event_id"`
	Title              string     `json:"title"`
	SubjectID          *uuid.UUID `json:"subject_id"`
	BankID             *uuid.UUID `json:"bank_id"`
	ClassID            uuid.UUID  `json:"class_id"`
	ExamDate           string     `json:"exam_date"`
	StartTime          string     `json:"start_time"`
	EndTime            string     `json:"end_time"`
	ExamToken          string     `json:"exam_token"`
	DurationMinutes    int        `json:"duration_minutes"`
	MaxViolations      int        `json:"max_violations"`
	RandomizeQuestions *bool      `json:"randomize_questions"`
	RandomizeOptions   *bool      `json:"randomize_options"`
}

type UpdateScheduleRequest struct {
	EventID            *uuid.UUID `json:"event_id"`
	Title              string     `json:"title"`
	SubjectID          *uuid.UUID `json:"subject_id"`
	BankID             *uuid.UUID `json:"bank_id"`
	ClassID            uuid.UUID  `json:"class_id"`
	ExamDate           string     `json:"exam_date"`
	StartTime          string     `json:"start_time"`
	EndTime            string     `json:"end_time"`
	ExamToken          string     `json:"exam_token"`
	DurationMinutes    int        `json:"duration_minutes"`
	MaxViolations      int        `json:"max_violations"`
	RandomizeQuestions *bool      `json:"randomize_questions"`
	RandomizeOptions   *bool      `json:"randomize_options"`
	IsActive           *bool      `json:"is_active"`
}

type LinkScheduleBankRequest struct {
	BankID *uuid.UUID `json:"bank_id"`
}

func (h *Handlers) HandleCreateSchedule(c *fiber.Ctx) error {
	var req CreateScheduleRequest
	if err := c.BodyParser(&req); err != nil || req.Title == "" || req.ClassID == uuid.Nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "message": "Judul dan kelas wajib diisi"})
	}

	var targetSubjectID *uuid.UUID
	if req.SubjectID != nil && *req.SubjectID != uuid.Nil {
		targetSubjectID = req.SubjectID
	}

	var targetBankID *uuid.UUID
	if req.BankID != nil && *req.BankID != uuid.Nil {
		targetBankID = req.BankID
		if targetSubjectID == nil {
			var bank domain.QuestionBank
			if err := h.repo.DB.First(&bank, "id = ?", *targetBankID).Error; err == nil {
				targetSubjectID = &bank.SubjectID
			}
		}
	}

	if targetSubjectID == nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "message": "Mata pelajaran wajib ditentukan untuk jadwal ujian"})
	}

	token := req.ExamToken
	if token == "" {
		token = "CBT2026"
	}
	duration := req.DurationMinutes
	if duration <= 0 {
		duration = 90
	}
	maxV := req.MaxViolations
	if maxV <= 0 {
		maxV = 3
	}

	randQ := true
	if req.RandomizeQuestions != nil {
		randQ = *req.RandomizeQuestions
	}
	randO := true
	if req.RandomizeOptions != nil {
		randO = *req.RandomizeOptions
	}

	targetEventID := req.EventID
	if targetEventID == nil || *targetEventID == uuid.Nil {
		var activeEvent domain.ExamEvent
		if err := h.repo.DB.Where("is_active = ?", true).First(&activeEvent).Error; err == nil {
			targetEventID = &activeEvent.ID
		}
	}

	st, et := parseScheduleTimes(req.ExamDate, req.StartTime, req.EndTime, duration)
	now := time.Now()

	sched := domain.ExamSchedule{
		ID:                 uuid.New(),
		EventID:            targetEventID,
		Title:              req.Title,
		SubjectID:          targetSubjectID,
		BankID:             targetBankID,
		ClassRoomID:        req.ClassID,
		ExamToken:          token,
		StartTime:          st,
		EndTime:            et,
		DurationMinutes:    duration,
		MaxViolations:      maxV,
		RandomizeQuestions: randQ,
		RandomizeOptions:   randO,
		IsActive:           true,
		CreatedAt:          now,
	}

	if err := h.repo.DB.Create(&sched).Error; err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "message": err.Error()})
	}

	h.repo.DB.Preload("Event").Preload("Subject").Preload("Bank").Preload("Bank.Subject").Preload("ClassRoom").First(&sched, "id = ?", sched.ID)

	return c.JSON(fiber.Map{"success": true, "data": sched, "message": "Jadwal ujian berhasil diterbitkan"})
}

func (h *Handlers) HandleUpdateSchedule(c *fiber.Ctx) error {
	schedID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "message": "ID jadwal tidak valid"})
	}

	var sched domain.ExamSchedule
	if err := h.repo.DB.First(&sched, "id = ?", schedID).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"success": false, "message": "Jadwal tidak ditemukan"})
	}

	var req UpdateScheduleRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "message": "Format data tidak valid"})
	}

	if req.Title != "" {
		sched.Title = req.Title
	}
	if req.ClassID != uuid.Nil {
		sched.ClassRoomID = req.ClassID
	}
	if req.SubjectID != nil && *req.SubjectID != uuid.Nil {
		sched.SubjectID = req.SubjectID
	}
	if req.EventID != nil && *req.EventID != uuid.Nil {
		sched.EventID = req.EventID
	}
	if req.BankID != nil {
		if *req.BankID == uuid.Nil {
			sched.BankID = nil
		} else {
			sched.BankID = req.BankID
			if sched.SubjectID == nil {
				var b domain.QuestionBank
				if err := h.repo.DB.First(&b, "id = ?", *req.BankID).Error; err == nil {
					sched.SubjectID = &b.SubjectID
				}
			}
		}
	}
	if req.ExamToken != "" {
		sched.ExamToken = req.ExamToken
	}
	if req.DurationMinutes > 0 {
		sched.DurationMinutes = req.DurationMinutes
	}
	if req.MaxViolations > 0 {
		sched.MaxViolations = req.MaxViolations
	}
	if req.RandomizeQuestions != nil {
		sched.RandomizeQuestions = *req.RandomizeQuestions
	}
	if req.RandomizeOptions != nil {
		sched.RandomizeOptions = *req.RandomizeOptions
	}
	if req.IsActive != nil {
		sched.IsActive = *req.IsActive
	}

	if req.ExamDate != "" || req.StartTime != "" || req.EndTime != "" {
		dateStr := req.ExamDate
		if dateStr == "" {
			dateStr = sched.StartTime.Format("2006-01-02")
		}
		startTimeStr := req.StartTime
		if startTimeStr == "" {
			startTimeStr = sched.StartTime.Format("15:04")
		}
		endTimeStr := req.EndTime
		if endTimeStr == "" {
			endTimeStr = sched.EndTime.Format("15:04")
		}
		st, et := parseScheduleTimes(dateStr, startTimeStr, endTimeStr, sched.DurationMinutes)
		sched.StartTime = st
		sched.EndTime = et
	}

	if err := h.repo.DB.Save(&sched).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false, "message": "Gagal memperbarui jadwal"})
	}

	h.repo.DB.Preload("Event").Preload("Subject").Preload("Bank").Preload("Bank.Subject").Preload("ClassRoom").First(&sched, "id = ?", sched.ID)

	return c.JSON(fiber.Map{"success": true, "data": sched, "message": "Jadwal ujian berhasil diperbarui"})
}

func (h *Handlers) HandleLinkScheduleBank(c *fiber.Ctx) error {
	schedID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "message": "ID jadwal tidak valid"})
	}
	caller, callerErr := middleware.GetCurrentUser(c)
	if callerErr != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"success": false, "message": "Autentikasi diperlukan"})
	}

	var sched domain.ExamSchedule
	if err := h.repo.DB.First(&sched, "id = ?", schedID).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"success": false, "message": "Jadwal tidak ditemukan"})
	}

	// Jadwal di luar cakupan keterlihatan dibalas 404 agar keberadaannya tidak terungkap.
	visible, visErr := h.accessService.ScheduleVisibilityFor(caller)
	if visErr != nil {
		log.Printf("gagal memeriksa keterlihatan jadwal %s: %v", schedID, visErr)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false, "message": "Gagal memeriksa hak akses jadwal"})
	}
	if !visible.Allows(schedID) {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"success": false, "message": "Jadwal tidak ditemukan"})
	}

	var req LinkScheduleBankRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "message": "Format data tidak valid"})
	}

	var newBankID, newSubjectID *uuid.UUID
	if req.BankID != nil && *req.BankID != uuid.Nil {
		var bank domain.QuestionBank
		if err := h.repo.DB.First(&bank, "id = ?", *req.BankID).Error; err != nil {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"success": false, "message": "Bank soal tidak ditemukan"})
		}
		if h.denyBank(c, *req.BankID) {
			return nil
		}
		newBankID = req.BankID
		if sched.SubjectID == nil {
			newSubjectID = &bank.SubjectID
		}
	}

	// Bank hanya boleh diganti (atau dilepas) selama belum ada sesi ujian siswa pada jadwal ini.
	// Menautkan bank yang sama dengan yang sekarang tidak mengubah apa pun, jadi selalu diizinkan.
	// Gagal kueri = gagal tertutup (500).
	if err := h.accessService.EnsureBankChangeAllowed(schedID, sched.BankID, newBankID); err != nil {
		return h.respondBankChangeError(c, schedID, err)
	}

	if !service.SameBankLink(sched.BankID, newBankID) {
		// Penulisan ikut memeriksa ulang bahwa belum ada sesi siswa (menutup celah balapan).
		if err := h.accessService.LinkBankGuarded(schedID, newBankID, newSubjectID); err != nil {
			return h.respondBankChangeError(c, schedID, err)
		}
	}

	h.repo.DB.Preload("Event").Preload("Subject").Preload("Bank").Preload("Bank.Subject").Preload("ClassRoom").First(&sched, "id = ?", sched.ID)

	// Token disunting sesuai cakupan pemanggil pada salinan respons saja (jadwal sudah tersimpan).
	h.accessService.TokenScopeFor(caller).RedactToken(&sched)

	return c.JSON(fiber.Map{"success": true, "data": sched, "message": "Naskah bank soal berhasil ditautkan ke jadwal"})
}

func (h *Handlers) HandleDeleteSchedule(c *fiber.Ctx) error {
	schedID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "message": "ID jadwal tidak valid"})
	}

	var sessionCount int64
	h.repo.DB.Model(&domain.ExamSession{}).Where("schedule_id = ?", schedID).Count(&sessionCount)
	if sessionCount > 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": fmt.Sprintf("Jadwal tidak dapat dihapus karena sudah ada %d sesi ujian siswa yang tercatat", sessionCount),
		})
	}

	// Penugasan pengawas ikut dihapus dalam transaksi yang sama.
	if err := h.accessService.DeleteScheduleWithProctors(schedID); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false, "message": "Gagal menghapus jadwal"})
	}

	return c.JSON(fiber.Map{"success": true, "message": "Jadwal ujian berhasil dihapus"})
}

func (h *Handlers) HandleToggleSchedule(c *fiber.Ctx) error {
	schedID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "message": "ID tidak valid"})
	}
	var sched domain.ExamSchedule
	if err := h.repo.DB.Preload("Bank").First(&sched, "id = ?", schedID).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"success": false, "message": "Jadwal tidak ditemukan"})
	}

	// Saat mengaktifkan: wajib ada bank soal yang sudah dikunci
	if !sched.IsActive {
		if sched.BankID == nil || *sched.BankID == uuid.Nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"success": false,
				"message": "Jadwal tidak dapat diaktifkan: bank soal belum ditautkan",
			})
		}
		if sched.Bank == nil || !sched.Bank.IsLocked {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"success": false,
				"message": "Jadwal tidak dapat diaktifkan: bank soal belum dikunci (naskah masih dalam penyusunan)",
			})
		}
	}

	sched.IsActive = !sched.IsActive
	h.repo.DB.Save(&sched)
	return c.JSON(fiber.Map{"success": true, "is_active": sched.IsActive, "message": "Status jadwal berhasil diperbarui"})
}

func (h *Handlers) HandleToggleBankLock(c *fiber.Ctx) error {
	bankID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "message": "ID tidak valid"})
	}
	if h.denyBank(c, bankID) {
		return nil
	}
	var bank domain.QuestionBank
	if err := h.repo.DB.First(&bank, "id = ?", bankID).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"success": false, "message": "Bank soal tidak ditemukan"})
	}
	bank.IsLocked = !bank.IsLocked
	h.repo.DB.Save(&bank)
	return c.JSON(fiber.Map{"success": true, "is_locked": bank.IsLocked, "message": "Status penguncian bank soal diperbarui"})
}

type CreateQuestionBankRequest struct {
	SubjectID uuid.UUID    `json:"subject_id"`
	Grade     *string      `json:"grade"`
	ClassIDs  *[]uuid.UUID `json:"class_ids"`
	Title     string       `json:"title"`
}

func (h *Handlers) HandleCreateQuestionBank(c *fiber.Ctx) error {
	user, err := middleware.GetCurrentUser(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"success": false, "message": "Autentikasi diperlukan"})
	}
	var req CreateQuestionBankRequest
	if err := c.BodyParser(&req); err != nil || req.SubjectID == uuid.Nil || strings.TrimSpace(req.Title) == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "message": "Mata pelajaran dan judul bank soal wajib diisi"})
	}

	grade := ""
	if req.Grade != nil {
		grade = strings.TrimSpace(*req.Grade)
	}

	bank := domain.QuestionBank{
		ID:             uuid.New(),
		SubjectID:      req.SubjectID,
		Grade:          grade,
		Title:          strings.TrimSpace(req.Title),
		CreatedByID:    user.ID,
		TotalQuestions: 0,
		IsLocked:       false,
		CreatedAt:      time.Now(),
	}

	if req.ClassIDs != nil {
		classes, err := h.loadBankClasses(*req.ClassIDs)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "message": err.Error()})
		}
		bank.Classes = classes
		if len(classes) > 0 {
			// Cakupan kelas khusus menggantikan cakupan per angkatan.
			bank.Grade = ""
		}
	}

	if err := h.repo.DB.Create(&bank).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false, "message": "Gagal membuat bank soal: " + err.Error()})
	}

	h.repo.DB.Preload("Subject").Preload("CreatedBy").Preload("Classes").First(&bank, "id = ?", bank.ID)

	return c.JSON(fiber.Map{"success": true, "data": bank, "message": "Bank soal baru berhasil dibuat"})
}

type UpdateQuestionBankRequest struct {
	SubjectID uuid.UUID    `json:"subject_id"`
	Grade     *string      `json:"grade"`
	ClassIDs  *[]uuid.UUID `json:"class_ids"`
	Title     string       `json:"title"`
}

func (h *Handlers) HandleUpdateQuestionBank(c *fiber.Ctx) error {
	bankID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "message": "ID bank soal tidak valid"})
	}
	if h.denyBank(c, bankID) {
		return nil
	}

	var bank domain.QuestionBank
	if err := h.repo.DB.First(&bank, "id = ?", bankID).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"success": false, "message": "Bank soal tidak ditemukan"})
	}

	var req UpdateQuestionBankRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "message": "Format data tidak valid"})
	}

	if req.SubjectID != uuid.Nil {
		bank.SubjectID = req.SubjectID
	}
	if req.Grade != nil {
		bank.Grade = strings.TrimSpace(*req.Grade)
	}
	if strings.TrimSpace(req.Title) != "" {
		bank.Title = strings.TrimSpace(req.Title)
	}

	var newClasses []domain.ClassRoom
	if req.ClassIDs != nil {
		newClasses, err = h.loadBankClasses(*req.ClassIDs)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "message": err.Error()})
		}
		if len(newClasses) > 0 {
			bank.Grade = ""
		}
	}

	err = h.repo.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Omit("Classes").Save(&bank).Error; err != nil {
			return err
		}
		if req.ClassIDs != nil {
			return tx.Model(&bank).Association("Classes").Replace(newClasses)
		}
		return nil
	})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false, "message": "Gagal memperbarui bank soal: " + err.Error()})
	}

	h.repo.DB.Preload("Subject").Preload("CreatedBy").Preload("Classes").First(&bank, "id = ?", bank.ID)

	return c.JSON(fiber.Map{"success": true, "data": bank, "message": "Bank soal berhasil diperbarui"})
}

// loadBankClasses memvalidasi daftar kelas cakupan bank soal; ID duplikat diabaikan.
func (h *Handlers) loadBankClasses(ids []uuid.UUID) ([]domain.ClassRoom, error) {
	unique := make([]uuid.UUID, 0, len(ids))
	seen := make(map[uuid.UUID]bool, len(ids))
	for _, id := range ids {
		if id != uuid.Nil && !seen[id] {
			seen[id] = true
			unique = append(unique, id)
		}
	}
	classes := make([]domain.ClassRoom, 0, len(unique))
	if len(unique) == 0 {
		return classes, nil
	}
	if err := h.repo.DB.Where("id IN ?", unique).Find(&classes).Error; err != nil {
		return nil, fmt.Errorf("Gagal memuat data kelas")
	}
	if len(classes) != len(unique) {
		return nil, fmt.Errorf("Sebagian kelas yang dipilih tidak ditemukan")
	}
	return classes, nil
}

func (h *Handlers) HandleDeleteQuestionBank(c *fiber.Ctx) error {
	bankID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "message": "ID bank soal tidak valid"})
	}
	if h.denyBank(c, bankID) {
		return nil
	}

	var schedCount int64
	h.repo.DB.Model(&domain.ExamSchedule{}).Where("bank_id = ?", bankID).Count(&schedCount)
	if schedCount > 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": fmt.Sprintf("Bank soal tidak dapat dihapus karena sedang ditautkan pada %d sesi jadwal ujian", schedCount),
		})
	}

	// Delete questions & cakupan kelas
	h.repo.DB.Delete(&domain.Question{}, "bank_id = ?", bankID)
	h.repo.DB.Model(&domain.QuestionBank{ID: bankID}).Association("Classes").Clear()
	// Delete bank
	if err := h.repo.DB.Delete(&domain.QuestionBank{}, "id = ?", bankID).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false, "message": "Gagal menghapus bank soal"})
	}

	return c.JSON(fiber.Map{"success": true, "message": "Bank soal berhasil dihapus"})
}

// ---------------- QUESTION MANAGEMENT HANDLERS ----------------

type ManageQuestionRequest struct {
	Type        domain.QuestionType `json:"type"`
	ContentHTML string              `json:"content_html"`
	Options     []domain.OptionItem `json:"options"`
	CorrectKey  string              `json:"correct_key"`
	RubricGuide string              `json:"rubric_guide"`
	ScoreWeight float64             `json:"score_weight"`
}

func (h *Handlers) HandleGetBankQuestions(c *fiber.Ctx) error {
	bankID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "message": "ID bank soal tidak valid"})
	}
	if h.denyBank(c, bankID) {
		return nil
	}

	var bank domain.QuestionBank
	if err := h.repo.DB.Preload("Subject").First(&bank, "id = ?", bankID).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"success": false, "message": "Bank soal tidak ditemukan"})
	}

	var questions []domain.Question
	h.repo.DB.Where("bank_id = ?", bankID).Order("question_number ASC").Find(&questions)

	type QuestionDetail struct {
		ID             uuid.UUID           `json:"id"`
		BankID         uuid.UUID           `json:"bank_id"`
		QuestionNumber int                 `json:"question_number"`
		Type           domain.QuestionType `json:"type"`
		ContentHTML    string              `json:"content_html"`
		OptionsJSON    string              `json:"options_json"`
		Options        []domain.OptionItem `json:"options"`
		CorrectKey     string              `json:"correct_key"`
		RubricGuide    string              `json:"rubric_guide"`
		ScoreWeight    float64             `json:"score_weight"`
		CreatedAt      time.Time           `json:"created_at"`
	}

	res := make([]QuestionDetail, 0, len(questions))
	for _, q := range questions {
		var opts []domain.OptionItem
		if err := json.Unmarshal([]byte(q.OptionsJSON), &opts); err != nil {
			opts = []domain.OptionItem{}
		}
		qType := q.Type
		if qType == "" {
			qType = domain.TypeMultipleChoice
		}
		res = append(res, QuestionDetail{
			ID:             q.ID,
			BankID:         q.BankID,
			QuestionNumber: q.QuestionNumber,
			Type:           qType,
			ContentHTML:    q.ContentHTML,
			OptionsJSON:    q.OptionsJSON,
			Options:        opts,
			CorrectKey:     q.CorrectKey,
			RubricGuide:    q.RubricGuide,
			ScoreWeight:    q.ScoreWeight,
			CreatedAt:      q.CreatedAt,
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"bank":    bank,
		"data":    res,
	})
}

type ReorderBankQuestionsRequest struct {
	QuestionIDs []uuid.UUID `json:"question_ids"`
}

func (h *Handlers) HandleReorderBankQuestions(c *fiber.Ctx) error {
	bankID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "message": "ID bank soal tidak valid"})
	}
	if h.denyBank(c, bankID) {
		return nil
	}

	var bank domain.QuestionBank
	if err := h.repo.DB.First(&bank, "id = ?", bankID).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"success": false, "message": "Bank soal tidak ditemukan"})
	}

	var req ReorderBankQuestionsRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "message": "Format request tidak valid"})
	}

	if len(req.QuestionIDs) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "message": "Daftar ID soal tidak boleh kosong"})
	}

	tx := h.repo.DB.Begin()
	for idx, qID := range req.QuestionIDs {
		if err := tx.Model(&domain.Question{}).Where("id = ? AND bank_id = ?", qID, bankID).Update("question_number", idx+1).Error; err != nil {
			tx.Rollback()
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false, "message": "Gagal memperbarui urutan soal"})
		}
	}
	tx.Commit()

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Urutan soal berhasil diperbarui",
	})
}

func (h *Handlers) HandleCreateQuestion(c *fiber.Ctx) error {
	bankID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "message": "ID bank soal tidak valid"})
	}
	if h.denyBank(c, bankID) {
		return nil
	}

	var bank domain.QuestionBank
	if err := h.repo.DB.First(&bank, "id = ?", bankID).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"success": false, "message": "Bank soal tidak ditemukan"})
	}

	if bank.IsLocked {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Bank soal terkunci. Buka kunci terlebih dahulu untuk menambah soal.",
		})
	}

	var req ManageQuestionRequest
	if err := c.BodyParser(&req); err != nil || strings.TrimSpace(req.ContentHTML) == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "message": "Konten soal wajib diisi"})
	}

	var count int64
	h.repo.DB.Model(&domain.Question{}).Where("bank_id = ?", bankID).Count(&count)

	scoreWeight := req.ScoreWeight
	if scoreWeight <= 0 {
		scoreWeight = 1.0
	}

	qType := req.Type
	if qType == "" {
		qType = domain.TypeMultipleChoice
	}

	correctKey := strings.TrimSpace(req.CorrectKey)
	if qType == domain.TypeMultipleChoice {
		correctKey = strings.ToUpper(correctKey)
		if correctKey == "" {
			correctKey = "A"
		}
	}

	optBytes, err := json.Marshal(req.Options)
	if err != nil {
		optBytes = []byte("[]")
	}

	q := domain.Question{
		ID:             uuid.New(),
		BankID:         bankID,
		QuestionNumber: int(count) + 1,
		Type:           qType,
		ContentHTML:    strings.TrimSpace(req.ContentHTML),
		OptionsJSON:    string(optBytes),
		CorrectKey:     correctKey,
		RubricGuide:    strings.TrimSpace(req.RubricGuide),
		ScoreWeight:    scoreWeight,
		CreatedAt:      time.Now(),
	}

	if err := h.repo.DB.Create(&q).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false, "message": "Gagal menambahkan butir soal"})
	}

	// Update total_questions on bank
	h.repo.DB.Model(&bank).Update("total_questions", int(count)+1)

	return c.JSON(fiber.Map{
		"success": true,
		"data":    q,
		"message": "Butir soal berhasil ditambahkan",
	})
}

func (h *Handlers) HandleUpdateQuestion(c *fiber.Ctx) error {
	qID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "message": "ID butir soal tidak valid"})
	}

	var q domain.Question
	if err := h.repo.DB.First(&q, "id = ?", qID).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"success": false, "message": "Butir soal tidak ditemukan"})
	}
	if h.denyBank(c, q.BankID) {
		return nil
	}

	var bank domain.QuestionBank
	if err := h.repo.DB.First(&bank, "id = ?", q.BankID).Error; err == nil && bank.IsLocked {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Bank soal sedang terkunci. Buka kunci terlebih dahulu untuk mengedit soal.",
		})
	}

	var req ManageQuestionRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "message": "Format data tidak valid"})
	}

	if req.Type != "" {
		q.Type = req.Type
	}
	if strings.TrimSpace(req.ContentHTML) != "" {
		q.ContentHTML = strings.TrimSpace(req.ContentHTML)
	}
	if req.CorrectKey != "" || req.Type == domain.TypeEssay {
		if q.Type == domain.TypeMultipleChoice {
			q.CorrectKey = strings.ToUpper(strings.TrimSpace(req.CorrectKey))
		} else {
			q.CorrectKey = strings.TrimSpace(req.CorrectKey)
		}
	}
	if req.RubricGuide != "" || req.Type != domain.TypeEssay {
		q.RubricGuide = strings.TrimSpace(req.RubricGuide)
	}
	if req.ScoreWeight > 0 {
		q.ScoreWeight = req.ScoreWeight
	}
	if req.Options != nil {
		if optBytes, err := json.Marshal(req.Options); err == nil {
			q.OptionsJSON = string(optBytes)
		}
	}

	if err := h.repo.DB.Save(&q).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false, "message": "Gagal memperbarui soal"})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    q,
		"message": "Butir soal berhasil diperbarui",
	})
}

func (h *Handlers) HandleDeleteQuestion(c *fiber.Ctx) error {
	qID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "message": "ID butir soal tidak valid"})
	}

	var q domain.Question
	if err := h.repo.DB.First(&q, "id = ?", qID).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"success": false, "message": "Butir soal tidak ditemukan"})
	}
	if h.denyBank(c, q.BankID) {
		return nil
	}

	var bank domain.QuestionBank
	if err := h.repo.DB.First(&bank, "id = ?", q.BankID).Error; err == nil && bank.IsLocked {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Bank soal sedang terkunci. Buka kunci terlebih dahulu untuk menghapus soal.",
		})
	}

	bankID := q.BankID
	if err := h.repo.DB.Delete(&q).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false, "message": "Gagal menghapus butir soal"})
	}

	// Renumber remaining questions
	var remaining []domain.Question
	h.repo.DB.Where("bank_id = ?", bankID).Order("question_number ASC, created_at ASC").Find(&remaining)
	for idx, item := range remaining {
		newNum := idx + 1
		if item.QuestionNumber != newNum {
			h.repo.DB.Model(&item).Update("question_number", newNum)
		}
	}

	// Update bank total_questions
	h.repo.DB.Model(&domain.QuestionBank{}).Where("id = ?", bankID).Update("total_questions", len(remaining))

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Butir soal berhasil dihapus",
	})
}

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

// ---------------- MEDIA / IMAGE UPLOAD HANDLER ----------------

func (h *Handlers) HandleUploadImage(c *fiber.Ctx) error {
	file, err := c.FormFile("image")
	if err != nil {
		file, err = c.FormFile("file")
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"success": false,
				"message": "File gambar tidak ditemukan dalam request",
			})
		}
	}

	// Max 5 MB
	if file.Size > 5*1024*1024 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Ukuran file gambar melebihi batas maksimal 5 MB",
		})
	}

	ext := strings.ToLower(filepath.Ext(file.Filename))
	allowedExtensions := map[string]bool{
		".jpg":  true,
		".jpeg": true,
		".png":  true,
		".webp": true,
		".gif":  true,
		".svg":  true,
	}

	if !allowedExtensions[ext] {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Format file tidak didukung. Gunakan format JPG, PNG, WEBP, GIF, atau SVG",
		})
	}

	uploadDir := "./uploads/questions"
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Gagal menyiapkan direktori penyimpanan gambar",
		})
	}

	uniqueFilename := fmt.Sprintf("%s%s", uuid.New().String(), ext)
	targetPath := filepath.Join(uploadDir, uniqueFilename)

	if err := c.SaveFile(file, targetPath); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Gagal menyimpan file gambar ke server",
		})
	}

	fileURL := "/uploads/questions/" + uniqueFilename
	return c.JSON(fiber.Map{
		"success":  true,
		"url":      fileURL,
		"filename": uniqueFilename,
		"message":  "Gambar berhasil diunggah",
	})
}

// ---------------- ESSAY GRADING HANDLERS ----------------

type EssayAnswerItem struct {
	AnswerID       uuid.UUID `json:"answer_id"`
	SessionID      uuid.UUID `json:"session_id"`
	StudentName    string    `json:"student_name"`
	StudentNIS     string    `json:"student_nis"`
	AnswerText     string    `json:"answer_text"`
	ScoreAwarded   *float64  `json:"score_awarded"`
	TeacherComment string    `json:"teacher_comment"`
	IsGraded       bool      `json:"is_graded"`
}

type EssayQuestionResult struct {
	QuestionID     uuid.UUID         `json:"question_id"`
	QuestionNumber int               `json:"question_number"`
	QuestionType   string            `json:"question_type"`
	ContentHTML    string            `json:"content_html"`
	ScoreWeight    float64           `json:"score_weight"`
	GradedCount    int               `json:"graded_count"`
	TotalCount     int               `json:"total_count"`
	Answers        []EssayAnswerItem `json:"answers"`
}

func (h *Handlers) HandleGetEssayAnswers(c *fiber.Ctx) error {
	scheduleID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "ID jadwal tidak valid",
		})
	}

	var schedule domain.ExamSchedule
	if err := h.repo.DB.First(&schedule, "id = ?", scheduleID).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"message": "Jadwal tidak ditemukan",
		})
	}

	if schedule.BankID == nil {
		return c.JSON(fiber.Map{"success": true, "data": []EssayQuestionResult{}})
	}

	var questions []domain.Question
	if err := h.repo.DB.Where("bank_id = ? AND type IN ?", schedule.BankID, []string{"ESSAY", "SHORT_ANSWER"}).
		Order("question_number ASC").Find(&questions).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Gagal mengambil soal",
		})
	}

	if len(questions) == 0 {
		return c.JSON(fiber.Map{"success": true, "data": []EssayQuestionResult{}})
	}

	var sessions []domain.ExamSession
	if err := h.repo.DB.Where("schedule_id = ?", scheduleID).Find(&sessions).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Gagal mengambil sesi ujian",
		})
	}

	sessionIDs := make([]uuid.UUID, 0, len(sessions))
	sessionMap := make(map[uuid.UUID]domain.ExamSession)
	for _, s := range sessions {
		sessionIDs = append(sessionIDs, s.ID)
		sessionMap[s.ID] = s
	}

	questionIDs := make([]uuid.UUID, 0, len(questions))
	for _, q := range questions {
		questionIDs = append(questionIDs, q.ID)
	}

	var answers []domain.StudentAnswer
	if len(sessionIDs) > 0 {
		if err := h.repo.DB.Where("session_id IN ? AND question_id IN ?", sessionIDs, questionIDs).
			Find(&answers).Error; err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"success": false,
				"message": "Gagal mengambil jawaban",
			})
		}
	}

	studentIDs := make([]uuid.UUID, 0, len(sessions))
	for _, s := range sessions {
		studentIDs = append(studentIDs, s.StudentID)
	}

	profileMap := make(map[uuid.UUID]domain.StudentProfile)
	userMap := make(map[uuid.UUID]domain.User)
	if len(studentIDs) > 0 {
		var profiles []domain.StudentProfile
		h.repo.DB.Preload("User").Where("user_id IN ?", studentIDs).Find(&profiles)
		for _, p := range profiles {
			profileMap[p.UserID] = p
			userMap[p.UserID] = p.User
		}
	}

	answersByQuestion := make(map[uuid.UUID][]EssayAnswerItem)
	for _, ans := range answers {
		sess, ok := sessionMap[ans.SessionID]
		if !ok {
			continue
		}
		studentName := ""
		studentNIS := ""
		if profile, pok := profileMap[sess.StudentID]; pok {
			studentNIS = profile.NIS
		}
		if user, uok := userMap[sess.StudentID]; uok {
			studentName = user.FullName
		}
		answersByQuestion[ans.QuestionID] = append(answersByQuestion[ans.QuestionID], EssayAnswerItem{
			AnswerID:       ans.ID,
			SessionID:      ans.SessionID,
			StudentName:    studentName,
			StudentNIS:     studentNIS,
			AnswerText:     ans.AnswerText,
			ScoreAwarded:   ans.ScoreAwarded,
			TeacherComment: ans.TeacherComment,
			IsGraded:       ans.IsGraded,
		})
	}

	result := make([]EssayQuestionResult, 0, len(questions))
	for _, q := range questions {
		items := answersByQuestion[q.ID]
		if items == nil {
			items = []EssayAnswerItem{}
		}
		gradedCount := 0
		for _, item := range items {
			if item.IsGraded {
				gradedCount++
			}
		}
		result = append(result, EssayQuestionResult{
			QuestionID:     q.ID,
			QuestionNumber: q.QuestionNumber,
			QuestionType:   string(q.Type),
			ContentHTML:    q.ContentHTML,
			ScoreWeight:    q.ScoreWeight,
			GradedCount:    gradedCount,
			TotalCount:     len(items),
			Answers:        items,
		})
	}

	return c.JSON(fiber.Map{"success": true, "data": result})
}

type GradeEssayItem struct {
	AnswerID       uuid.UUID `json:"answer_id"`
	ScoreAwarded   float64   `json:"score_awarded"`
	TeacherComment string    `json:"teacher_comment"`
}

func (h *Handlers) HandleGradeEssayAnswers(c *fiber.Ctx) error {
	scheduleID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "ID jadwal tidak valid",
		})
	}

	var items []GradeEssayItem
	if err := c.BodyParser(&items); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Format body tidak valid",
		})
	}

	if len(items) == 0 {
		return c.JSON(fiber.Map{"success": true, "message": "0 jawaban berhasil dinilai"})
	}

	for _, item := range items {
		if item.ScoreAwarded < 0 {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"success": false,
				"message": "Nilai tidak boleh negatif",
			})
		}
	}

	affectedSessionIDs := make(map[uuid.UUID]struct{})

	txErr := h.repo.DB.Transaction(func(tx *gorm.DB) error {
		for _, item := range items {
			var ans domain.StudentAnswer
			if err := tx.Raw(`
				SELECT sa.* FROM student_answers sa
				JOIN exam_sessions es ON es.id = sa.session_id
				WHERE sa.id = ? AND es.schedule_id = ?
			`, item.AnswerID, scheduleID).Scan(&ans).Error; err != nil {
				return err
			}
			if ans.ID == uuid.Nil {
				continue
			}

			scoreVal := item.ScoreAwarded
			if err := tx.Model(&domain.StudentAnswer{}).Where("id = ?", item.AnswerID).
				Updates(map[string]interface{}{
					"score_awarded":   &scoreVal,
					"teacher_comment": item.TeacherComment,
					"is_graded":       true,
				}).Error; err != nil {
				return err
			}
			affectedSessionIDs[ans.SessionID] = struct{}{}
		}

		for sessionID := range affectedSessionIDs {
			var sess domain.ExamSession
			if err := tx.First(&sess, "id = ?", sessionID).Error; err != nil {
				return err
			}
			var sched domain.ExamSchedule
			if err := tx.First(&sched, "id = ?", sess.ScheduleID).Error; err != nil {
				return err
			}
			if sched.BankID == nil {
				continue
			}

			var totalWeight float64
			tx.Model(&domain.Question{}).Where("bank_id = ?", sched.BankID).
				Select("COALESCE(SUM(score_weight), 0)").Scan(&totalWeight)

			var earnedScore float64
			tx.Model(&domain.StudentAnswer{}).Where("session_id = ?", sessionID).
				Select("COALESCE(SUM(score_awarded), 0)").Scan(&earnedScore)

			finalGrade := 0.0
			if totalWeight > 0 {
				finalGrade = (earnedScore / totalWeight) * 100.0
			}
			if err := tx.Model(&domain.ExamSession{}).Where("id = ?", sessionID).
				Update("total_score", finalGrade).Error; err != nil {
				return err
			}
		}
		return nil
	})

	if txErr != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Gagal menyimpan penilaian: " + txErr.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": fmt.Sprintf("%d jawaban berhasil dinilai", len(items)),
	})
}
