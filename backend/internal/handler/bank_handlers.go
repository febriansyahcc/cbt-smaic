package handler

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"cbt-backend/internal/domain"
	"cbt-backend/internal/middleware"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

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

	return c.JSON(fiber.Map{
		"success": true,
		"bank":    bank,
		"data":    h.loadBankQuestionDetails(bankID),
	})
}

// HandleGetBankPrint mengirim naskah lengkap (soal, opsi, kunci, rubrik) untuk dicetak.
// Hanya bank yang sudah terkunci agar naskah cetak sama dengan yang dipakai ujian.
// Urutan mengikuti nomor soal di bank (tidak diacak).
func (h *Handlers) HandleGetBankPrint(c *fiber.Ctx) error {
	bankID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "message": "ID bank soal tidak valid"})
	}
	if h.denyBank(c, bankID) {
		return nil
	}

	var bank domain.QuestionBank
	if err := h.repo.DB.Preload("Subject").Preload("Classes").First(&bank, "id = ?", bankID).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"success": false, "message": "Bank soal tidak ditemukan"})
	}
	if !bank.IsLocked {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{"success": false, "message": "Kunci naskah terlebih dahulu sebelum mencetak"})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"bank":    bank,
		"data":    h.loadBankQuestionDetails(bankID),
	})
}

// BankQuestionDetail adalah butir soal beserta opsi yang sudah diurai dari OptionsJSON.
type BankQuestionDetail struct {
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

func (h *Handlers) loadBankQuestionDetails(bankID uuid.UUID) []BankQuestionDetail {
	var questions []domain.Question
	h.repo.DB.Where("bank_id = ?", bankID).Order("question_number ASC").Find(&questions)

	res := make([]BankQuestionDetail, 0, len(questions))
	for _, q := range questions {
		var opts []domain.OptionItem
		if err := json.Unmarshal([]byte(q.OptionsJSON), &opts); err != nil {
			opts = []domain.OptionItem{}
		}
		qType := q.Type
		if qType == "" {
			qType = domain.TypeMultipleChoice
		}
		res = append(res, BankQuestionDetail{
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

	return res
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
