package handler

import (
	"fmt"
	"log"
	"strings"
	"time"

	"cbt-backend/internal/domain"
	"cbt-backend/internal/middleware"
	"cbt-backend/internal/service"
	"cbt-backend/pkg/excel"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

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
