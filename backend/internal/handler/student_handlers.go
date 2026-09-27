package handler

import (
	"time"

	"cbt-backend/internal/middleware"
	"cbt-backend/internal/service"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

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

	count, deadline, err := h.examService.SyncAnswers(req.SessionID, user.ID, req.Answers)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success":         true,
		"synced":          count,
		"server_time":     time.Now(),
		"server_deadline": deadline,
		"message":         "Jawaban berhasil disinkronkan",
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
