package handler

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"cbt-backend/internal/domain"
	"cbt-backend/internal/service"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

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
