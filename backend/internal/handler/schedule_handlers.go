package handler

import (
	"fmt"
	"log"
	"strings"
	"time"

	"cbt-backend/internal/domain"
	"cbt-backend/internal/middleware"
	"cbt-backend/internal/service"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

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
	IsMakeup           bool       `json:"is_makeup"`
	ParentScheduleID   *uuid.UUID `json:"parent_schedule_id"`
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
	IsMakeup           *bool      `json:"is_makeup"`
}

type LinkScheduleBankRequest struct {
	BankID *uuid.UUID `json:"bank_id"`
}

func (h *Handlers) HandleCreateSchedule(c *fiber.Ctx) error {
	var req CreateScheduleRequest
	if err := c.BodyParser(&req); err != nil || req.Title == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "message": "Judul jadwal wajib diisi"})
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

	// Jika parent_id diisi, jadwal ini adalah susulan: inherit kelas/mapel/event dari induk
	if req.ParentScheduleID != nil && *req.ParentScheduleID != uuid.Nil {
		var parent domain.ExamSchedule
		if err := h.repo.DB.First(&parent, "id = ?", *req.ParentScheduleID).Error; err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "message": "Jadwal induk tidak ditemukan"})
		}
		if parent.IsMakeup {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "message": "Jadwal induk tidak boleh merupakan jadwal susulan"})
		}
		req.IsMakeup = true
		if req.ClassID == uuid.Nil {
			req.ClassID = parent.ClassRoomID
		}
		if targetSubjectID == nil {
			targetSubjectID = parent.SubjectID
		}
		if targetEventID == nil {
			targetEventID = parent.EventID
		}
	}

	// Validasi kelas wajib ada (setelah inherit dari parent)
	if req.ClassID == uuid.Nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "message": "Kelas wajib ditentukan untuk jadwal ujian"})
	}

	if targetSubjectID == nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "message": "Mata pelajaran wajib ditentukan untuk jadwal ujian"})
	}

	st, et := parseScheduleTimes(req.ExamDate, req.StartTime, req.EndTime, duration)
	now := time.Now()

	// Satu sesi waktu (event + jam mulai) memakai satu token untuk semua kelas.
	token := strings.ToUpper(strings.TrimSpace(req.ExamToken))
	if shared, ok := service.SessionToken(h.repo.DB, targetEventID, st, uuid.Nil); ok {
		token = shared
	} else if token == "" {
		generated, err := service.GenerateExamToken()
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false, "message": "Gagal membuat token ujian"})
		}
		token = generated
	}

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
		IsMakeup:           req.IsMakeup,
		ParentScheduleID:   req.ParentScheduleID,
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
	oldEventID, oldStart, oldToken := sched.EventID, sched.StartTime, sched.ExamToken

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
	if req.IsMakeup != nil {
		sched.IsMakeup = *req.IsMakeup
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

	// Token dibagi per sesi waktu: token yang diubah diterapkan ke seluruh jadwal sesi itu,
	// sedangkan jadwal yang pindah ke sesi lain mengikuti token sesi tujuan.
	newToken := strings.ToUpper(strings.TrimSpace(req.ExamToken))
	tokenChanged := newToken != "" && newToken != oldToken
	slotMoved := !sched.StartTime.Equal(oldStart) || !sameEventID(sched.EventID, oldEventID)
	if tokenChanged {
		sched.ExamToken = newToken
	} else if slotMoved {
		if shared, ok := service.SessionToken(h.repo.DB, sched.EventID, sched.StartTime, sched.ID); ok {
			sched.ExamToken = shared
		}
	}

	if err := h.repo.DB.Save(&sched).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false, "message": "Gagal memperbarui jadwal"})
	}
	if tokenChanged {
		if err := service.ApplySessionToken(h.repo.DB, sched.EventID, sched.StartTime, sched.ExamToken); err != nil {
			log.Printf("gagal menyamakan token sesi jadwal %s: %v", sched.ID, err)
		}
	}

	h.repo.DB.Preload("Event").Preload("Subject").Preload("Bank").Preload("Bank.Subject").Preload("ClassRoom").First(&sched, "id = ?", sched.ID)

	return c.JSON(fiber.Map{"success": true, "data": sched, "message": "Jadwal ujian berhasil diperbarui"})
}

func sameEventID(a, b *uuid.UUID) bool {
	if a == nil || b == nil {
		return a == nil && b == nil
	}
	return *a == *b
}

type RegenerateTokensRequest struct {
	ScheduleIDs []uuid.UUID `json:"schedule_ids"`
}

// HandleRegenerateSessionTokens memberi satu token baru yang sama untuk seluruh jadwal terpilih.
// Jadwal lain pada sesi waktu yang sama ikut mendapat token itu agar tetap seragam.
func (h *Handlers) HandleRegenerateSessionTokens(c *fiber.Ctx) error {
	var req RegenerateTokensRequest
	if err := c.BodyParser(&req); err != nil || len(req.ScheduleIDs) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "message": "Pilih minimal satu jadwal"})
	}
	token, sessions, err := service.RegenerateSessionTokens(h.repo.DB, req.ScheduleIDs)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false, "message": "Gagal mengacak token"})
	}
	return c.JSON(fiber.Map{
		"success":    true,
		"exam_token": token,
		"sessions":   sessions,
		"message":    fmt.Sprintf("Token %s diterapkan ke %d sesi waktu", token, sessions),
	})
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

// HandleGetMakeupStudents — GET /admin/schedules/:id/makeup-students
func (h *Handlers) HandleGetMakeupStudents(c *fiber.Ctx) error {
	schedID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "message": "ID jadwal tidak valid"})
	}
	var sched domain.ExamSchedule
	if err := h.repo.DB.First(&sched, "id = ?", schedID).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"success": false, "message": "Jadwal tidak ditemukan"})
	}

	type MakeupStudentRow struct {
		StudentID uuid.UUID `json:"student_id"`
		FullName  string    `json:"full_name"`
		NIS       string    `json:"nis"`
		ClassName string    `json:"class_name"`
	}
	var rows []MakeupStudentRow
	h.repo.DB.Raw(`
		SELECT ems.student_id, u.full_name, sp.nis, cr.name as class_name
		FROM exam_makeup_students ems
		JOIN users u ON u.id = ems.student_id
		JOIN student_profiles sp ON sp.user_id = ems.student_id
		JOIN class_rooms cr ON cr.id = sp.class_room_id
		WHERE ems.schedule_id = ?
		ORDER BY cr.name, u.full_name
	`, schedID).Scan(&rows)

	return c.JSON(fiber.Map{"success": true, "data": rows})
}

// HandleSetMakeupStudents — POST /admin/schedules/:id/makeup-students
// Body: { "student_ids": ["uuid1", "uuid2"] }  — REPLACE seluruh whitelist
func (h *Handlers) HandleSetMakeupStudents(c *fiber.Ctx) error {
	schedID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "message": "ID jadwal tidak valid"})
	}
	var sched domain.ExamSchedule
	if err := h.repo.DB.First(&sched, "id = ?", schedID).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"success": false, "message": "Jadwal tidak ditemukan"})
	}
	if !sched.IsMakeup {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "message": "Jadwal ini bukan jadwal susulan"})
	}

	var body struct {
		StudentIDs []uuid.UUID `json:"student_ids"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "message": "Format data tidak valid"})
	}

	// Replace whitelist dalam satu transaksi
	if err := h.repo.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("schedule_id = ?", schedID).Delete(&domain.ExamMakeupStudent{}).Error; err != nil {
			return err
		}
		for _, sid := range body.StudentIDs {
			row := domain.ExamMakeupStudent{ScheduleID: schedID, StudentID: sid, CreatedAt: time.Now()}
			if err := tx.Create(&row).Error; err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false, "message": "Gagal menyimpan peserta susulan"})
	}

	return c.JSON(fiber.Map{"success": true, "message": fmt.Sprintf("%d peserta susulan berhasil disimpan", len(body.StudentIDs))})
}

// HandleRemoveMakeupStudent — DELETE /admin/schedules/:id/makeup-students/:studentId
func (h *Handlers) HandleRemoveMakeupStudent(c *fiber.Ctx) error {
	schedID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "message": "ID jadwal tidak valid"})
	}
	studentID, err := uuid.Parse(c.Params("studentId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "message": "ID siswa tidak valid"})
	}

	result := h.repo.DB.Where("schedule_id = ? AND student_id = ?", schedID, studentID).Delete(&domain.ExamMakeupStudent{})
	if result.Error != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false, "message": "Gagal menghapus peserta"})
	}

	return c.JSON(fiber.Map{"success": true, "message": "Peserta berhasil dihapus dari daftar susulan"})
}
