package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"cbt-backend/internal/domain"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

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
	if h.denyEssayGrading(c, scheduleID) {
		return nil
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

// errScoreAboveWeight menandai nilai essay yang melebihi bobot soalnya.
type errScoreAboveWeight struct{ weight float64 }

func (e errScoreAboveWeight) Error() string {
	return fmt.Sprintf("Nilai tidak boleh melebihi bobot soal (%g)", e.weight)
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
	if h.denyEssayGrading(c, scheduleID) {
		return nil
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
			var weight float64
			tx.Model(&domain.Question{}).Where("id = ?", ans.QuestionID).Select("score_weight").Scan(&weight)
			if item.ScoreAwarded > weight {
				return errScoreAboveWeight{weight: weight}
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

	var overWeight errScoreAboveWeight
	if errors.As(txErr, &overWeight) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": overWeight.Error(),
		})
	}
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

// ---------------- RIWAYAT JAWABAN HANDLERS ----------------

func (h *Handlers) HandleGetScheduleSessions(c *fiber.Ctx) error {
	scheduleID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false, "message": "ID jadwal tidak valid",
		})
	}
	if h.denyEssayGrading(c, scheduleID) {
		return nil
	}

	var sessions []domain.ExamSession
	h.repo.DB.Preload("Student").
		Where("schedule_id = ?", scheduleID).
		Order("started_at ASC").
		Find(&sessions)

	studentIDs := make([]uuid.UUID, 0, len(sessions))
	for _, s := range sessions {
		studentIDs = append(studentIDs, s.StudentID)
	}

	profileMap := make(map[uuid.UUID]domain.StudentProfile)
	if len(studentIDs) > 0 {
		var profiles []domain.StudentProfile
		h.repo.DB.Where("user_id IN ?", studentIDs).Find(&profiles)
		for _, p := range profiles {
			profileMap[p.UserID] = p
		}
	}

	type SessionItem struct {
		SessionID   uuid.UUID `json:"session_id"`
		StudentName string    `json:"student_name"`
		StudentNIS  string    `json:"student_nis"`
		Status      string    `json:"status"`
		TotalScore  float64   `json:"total_score"`
		StartedAt   time.Time `json:"started_at"`
	}

	result := make([]SessionItem, 0, len(sessions))
	for _, s := range sessions {
		result = append(result, SessionItem{
			SessionID:   s.ID,
			StudentName: s.Student.FullName,
			StudentNIS:  profileMap[s.StudentID].NIS,
			Status:      string(s.Status),
			TotalScore:  s.TotalScore,
			StartedAt:   s.StartedAt,
		})
	}

	return c.JSON(fiber.Map{"success": true, "data": result})
}

func (h *Handlers) HandleGetScheduleSessionAnswers(c *fiber.Ctx) error {
	scheduleID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false, "message": "ID jadwal tidak valid",
		})
	}
	sessionID, err := uuid.Parse(c.Params("session_id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false, "message": "ID sesi tidak valid",
		})
	}
	if h.denyEssayGrading(c, scheduleID) {
		return nil
	}

	var session domain.ExamSession
	if err := h.repo.DB.Preload("Schedule").
		First(&session, "id = ? AND schedule_id = ?", sessionID, scheduleID).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false, "message": "Sesi tidak ditemukan",
		})
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
	h.repo.DB.Where("bank_id = ?", session.Schedule.BankID).
		Order("question_number ASC").Find(&questions)

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
		QuestionNumber int            `json:"question_number"`
		Type           string         `json:"type"`
		ContentHTML    string         `json:"content_html"`
		ScoreWeight    float64        `json:"score_weight"`
		Options        []OptionResult `json:"options"`
		CorrectKey     string         `json:"correct_key"`
		SelectedOption string         `json:"selected_option"`
		AnswerText     string         `json:"answer_text"`
		IsCorrect      *bool          `json:"is_correct"`
		ScoreAwarded   *float64       `json:"score_awarded"`
		IsGraded       bool           `json:"is_graded"`
	}

	result := make([]QuestionResult, 0, len(questions))
	for _, q := range questions {
		var opts []domain.OptionItem
		if err := json.Unmarshal([]byte(q.OptionsJSON), &opts); err != nil {
			opts = []domain.OptionItem{}
		}
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
