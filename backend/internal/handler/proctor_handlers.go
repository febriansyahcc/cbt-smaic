package handler

import (
	"encoding/json"
	"fmt"
	"strings"

	"cbt-backend/internal/domain"
	"cbt-backend/internal/middleware"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

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

func (h *Handlers) HandleResetExamSession(c *fiber.Ctx) error {
	sessionID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "message": "ID sesi tidak valid"})
	}
	if err := h.proctorService.ResetExamSession(sessionID); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "message": err.Error()})
	}
	return c.JSON(fiber.Map{
		"success": true,
		"message": "Sesi ujian berhasil direset. Siswa dapat mengulang dari awal.",
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

func (h *Handlers) HandleExportMergedGradesExcel(c *fiber.Ctx) error {
	scheduleID, err := uuid.Parse(c.Params("schedule_id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).SendString("Invalid schedule ID")
	}
	if h.denyScheduleExport(c, scheduleID) {
		return nil
	}
	file, filename, err := h.proctorService.ExportMergedGrades(scheduleID)
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

func (h *Handlers) HandleGetSessionAnswers(c *fiber.Ctx) error {
	sessionID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "message": "ID sesi tidak valid"})
	}

	if h.denySessionView(c, sessionID) {
		return nil
	}

	var session domain.ExamSession
	if err := h.repo.DB.Preload("Schedule").First(&session, "id = ?", sessionID).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"success": false, "message": "Sesi ujian tidak ditemukan"})
	}

	isSubmitted := session.Status == domain.StatusSubmitted

	var student domain.User
	h.repo.DB.First(&student, "id = ?", session.StudentID)
	var profile domain.StudentProfile
	h.repo.DB.Where("user_id = ?", session.StudentID).First(&profile)

	if session.Schedule.BankID == nil {
		return c.JSON(fiber.Map{"success": true, "data": fiber.Map{"questions": []interface{}{}}})
	}

	var questions []domain.Question
	h.repo.DB.Where("bank_id = ?", session.Schedule.BankID).Order("question_number ASC").Find(&questions)

	var answers []domain.StudentAnswer
	h.repo.DB.Where("session_id = ?", sessionID).Find(&answers)
	answerMap := make(map[uuid.UUID]domain.StudentAnswer)
	for _, a := range answers {
		answerMap[a.QuestionID] = a
	}

	type OptionResult struct {
		Key      string `json:"key"`
		Text     string `json:"text"`
		ImageURL string `json:"image_url,omitempty"`
	}
	type QuestionResult struct {
		QuestionNumber int           `json:"question_number"`
		Type           string        `json:"type"`
		ContentHTML    string        `json:"content_html"`
		ScoreWeight    float64       `json:"score_weight"`
		Options        []OptionResult `json:"options"`
		CorrectKey     string        `json:"correct_key"`
		SelectedOption string        `json:"selected_option"`
		AnswerText     string        `json:"answer_text"`
		IsCorrect      *bool         `json:"is_correct"`
		ScoreAwarded   *float64      `json:"score_awarded"`
		IsGraded       bool          `json:"is_graded"`
	}

	result := make([]QuestionResult, 0, len(questions))
	for _, q := range questions {
		var opts []domain.OptionItem
		_ = json.Unmarshal([]byte(q.OptionsJSON), &opts)
		optResults := make([]OptionResult, 0, len(opts))
		for _, o := range opts {
			optResults = append(optResults, OptionResult{Key: o.Key, Text: o.Text, ImageURL: o.ImageURL})
		}

		ans := answerMap[q.ID]
		var isCorrect *bool
		correctKey := ""

		if isSubmitted {
			correctKey = q.CorrectKey
			if q.Type == domain.TypeMultipleChoice || q.Type == "" {
				correct := ans.SelectedOption != "" && strings.EqualFold(ans.SelectedOption, q.CorrectKey)
				isCorrect = &correct
			}
		}

		result = append(result, QuestionResult{
			QuestionNumber: q.QuestionNumber,
			Type:           string(q.Type),
			ContentHTML:    q.ContentHTML,
			ScoreWeight:    q.ScoreWeight,
			Options:        optResults,
			CorrectKey:     correctKey,
			SelectedOption: ans.SelectedOption,
			AnswerText:     ans.AnswerText,
			IsCorrect:      isCorrect,
			ScoreAwarded:   ans.ScoreAwarded,
			IsGraded:       ans.IsGraded,
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data": fiber.Map{
			"student_name":   student.FullName,
			"student_nis":    profile.NIS,
			"schedule_title": session.Schedule.Title,
			"total_score":    session.TotalScore,
			"status":         string(session.Status),
			"is_submitted":   isSubmitted,
			"questions":      result,
		},
	})
}
