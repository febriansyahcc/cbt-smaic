package handler

import (
	"errors"
	"fmt"

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
