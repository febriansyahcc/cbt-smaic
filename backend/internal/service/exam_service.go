package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"cbt-backend/internal/domain"
	"cbt-backend/internal/repository"
	"cbt-backend/pkg/prng"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type ExamService struct {
	repo *repository.Database
}

func NewExamService(repo *repository.Database) *ExamService {
	return &ExamService{repo: repo}
}

type ClientQuestion struct {
	ID             uuid.UUID           `json:"id"`
	QuestionNumber int                 `json:"question_number"`
	Type           domain.QuestionType `json:"type"`
	ContentHTML    string              `json:"content_html"`
	Options        []domain.OptionItem `json:"options"`
}

type ExamPayloadResponse struct {
	SessionID         uuid.UUID              `json:"session_id"`
	ScheduleTitle     string                 `json:"schedule_title"`
	SubjectName       string                 `json:"subject_name"`
	DurationMinutes   int                    `json:"duration_minutes"`
	ServerTime        time.Time              `json:"server_time"`
	ServerDeadline    time.Time              `json:"server_deadline"`
	RemainingSeconds  int64                  `json:"remaining_seconds"`
	MaxViolations     int                    `json:"max_violations"`
	CurrentViolations int                    `json:"current_violations"`
	Questions         []ClientQuestion       `json:"questions"`
	SavedAnswers      map[string]SavedAnswer `json:"saved_answers"`
}

type SavedAnswer struct {
	SelectedOption string `json:"selected_option"`
	AnswerText     string `json:"answer_text"`
	IsDoubtful     bool   `json:"is_doubtful"`
}

type SyncAnswerItem struct {
	QuestionID     uuid.UUID `json:"question_id"`
	SelectedOption string    `json:"selected_option"`
	AnswerText     string    `json:"answer_text"`
	IsDoubtful     bool      `json:"is_doubtful"`
}

// GetStudentSchedules finds active exam schedules available for the student.
// exam_token selalu dikosongkan: siswa harus mengetik token yang diumumkan pengawas,
// sehingga token tidak boleh ikut terkirim ke klien. Validasi token tetap memakai
// data jadwal yang dimuat sendiri oleh StartOrResumeExam.
func (s *ExamService) GetStudentSchedules(studentUserID uuid.UUID) ([]domain.ExamSchedule, error) {
	var profile domain.StudentProfile
	if err := s.repo.DB.Where("user_id = ?", studentUserID).First(&profile).Error; err != nil {
		return nil, errors.New("profil siswa tidak ditemukan")
	}

	var schedules []domain.ExamSchedule
	now := time.Now()
	err := s.repo.DB.
		Preload("Event").
		Preload("Subject").
		Preload("Bank").
		Preload("Bank.Subject").
		Preload("ClassRoom").
		Joins("LEFT JOIN exam_events ON exam_events.id = exam_schedules.event_id").
		Joins("LEFT JOIN question_banks ON question_banks.id = exam_schedules.bank_id").
		Where("class_room_id = ? AND exam_schedules.is_active = ? AND start_time <= ? AND end_time >= ? AND (exam_schedules.event_id IS NULL OR exam_events.is_active = ?) AND exam_schedules.bank_id IS NOT NULL AND question_banks.is_locked = ?",
			profile.ClassRoomID, true, now, now, true, true).
		Order("start_time ASC").
		Find(&schedules).Error

	ClearExamTokens(schedules)
	return schedules, err
}

// StartOrResumeExam handles token validation, session init, PRNG shuffling, and strips answer keys
func (s *ExamService) StartOrResumeExam(studentUserID uuid.UUID, scheduleID uuid.UUID, token, clientIP, userAgent string) (*ExamPayloadResponse, error) {
	var schedule domain.ExamSchedule
	if err := s.repo.DB.Preload("Event").Preload("Subject").Preload("Bank").Preload("Bank.Subject").First(&schedule, "id = ?", scheduleID).Error; err != nil {
		return nil, errors.New("jadwal ujian tidak ditemukan")
	}

	if !schedule.IsActive {
		return nil, errors.New("ujian ini belum dibuka atau telah dinonaktifkan")
	}

	if schedule.EventID != nil && schedule.Event != nil && !schedule.Event.IsActive {
		return nil, errors.New("event ujian ini telah ditutup atau diarsipkan")
	}

	if schedule.BankID == nil || *schedule.BankID == uuid.Nil || schedule.Bank == nil {
		return nil, errors.New("naskah soal untuk jadwal ujian ini belum ditautkan oleh kurikulum")
	}

	if !schedule.Bank.IsLocked {
		return nil, errors.New("naskah soal ujian ini masih dalam tahap penyusunan dan belum dikunci/siap diujikan")
	}

	now := time.Now()
	if now.Before(schedule.StartTime) || now.After(schedule.EndTime) {
		return nil, errors.New("jadwal ujian belum dimulai atau sudah berakhir")
	}

	if strings.ToUpper(strings.TrimSpace(token)) != strings.ToUpper(strings.TrimSpace(schedule.ExamToken)) {
		return nil, errors.New("token ujian salah atau tidak valid")
	}

	// Check existing session
	var session domain.ExamSession
	err := s.repo.DB.Where("schedule_id = ? AND student_id = ?", schedule.ID, studentUserID).First(&session).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		// Create new session
		deadline := now.Add(time.Duration(schedule.DurationMinutes) * time.Minute)
		if deadline.After(schedule.EndTime) {
			deadline = schedule.EndTime
		}

		session = domain.ExamSession{
			ID:             uuid.New(),
			ScheduleID:     schedule.ID,
			StudentID:      studentUserID,
			StartedAt:      now,
			ServerDeadline: deadline,
			Status:         domain.StatusInProgress,
			ViolationCount: 0,
			ClientIP:       clientIP,
			UserAgent:      userAgent,
			CreatedAt:      now,
			UpdatedAt:      now,
		}
		if err := s.repo.DB.Create(&session).Error; err != nil {
			return nil, fmt.Errorf("gagal membuat sesi ujian: %w", err)
		}
	} else if err != nil {
		return nil, err
	} else {
		// Existing session check
		if session.Status == domain.StatusSubmitted {
			return nil, errors.New("ujian ini telah Anda selesaikan dan kumpulkan")
		}
		if session.Status == domain.StatusBlocked {
			return nil, errors.New("ujian Anda terkunci karena kuota pelanggaran terlampaui. Hubungi pengawas ruangan")
		}
		// Refresh IP & UserAgent
		session.ClientIP = clientIP
		session.UserAgent = userAgent
		session.UpdatedAt = now
		s.repo.DB.Save(&session)
	}

	// Check deadline
	if now.After(session.ServerDeadline) {
		session.Status = domain.StatusSubmitted
		nowVal := now
		session.SubmittedAt = &nowVal
		s.repo.DB.Save(&session)
		return nil, errors.New("waktu pengerjaan ujian telah habis")
	}

	// Fetch questions
	var questions []domain.Question
	if err := s.repo.DB.Where("bank_id = ?", schedule.BankID).Order("question_number ASC").Find(&questions).Error; err != nil {
		return nil, fmt.Errorf("gagal memuat soal: %w", err)
	}

	// Seeded Randomization
	masterSeed := prng.GenerateSeed(studentUserID.String(), schedule.ID.String())
	if schedule.RandomizeQuestions {
		questions = prng.ShuffleQuestions(questions, masterSeed)
	}

	var clientQuestions []ClientQuestion
	for idx, q := range questions {
		var opts []domain.OptionItem
		if err := json.Unmarshal([]byte(q.OptionsJSON), &opts); err != nil {
			opts = []domain.OptionItem{}
		}

		if schedule.RandomizeOptions {
			subSeed := prng.GenerateSubSeed(masterSeed, q.ID.String())
			opts = prng.ShuffleOptions(opts, subSeed)
		}

		qType := q.Type
		if qType == "" {
			qType = domain.TypeMultipleChoice
		}

		clientQuestions = append(clientQuestions, ClientQuestion{
			ID:             q.ID,
			QuestionNumber: idx + 1,
			Type:           qType,
			ContentHTML:    q.ContentHTML,
			Options:        opts, // Notice: CorrectKey is omitted!
		})
	}

	// Fetch existing student answers
	var answers []domain.StudentAnswer
	s.repo.DB.Where("session_id = ?", session.ID).Find(&answers)
	savedMap := make(map[string]SavedAnswer)
	for _, a := range answers {
		savedMap[a.QuestionID.String()] = SavedAnswer{
			SelectedOption: a.SelectedOption,
			AnswerText:     a.AnswerText,
			IsDoubtful:     a.IsDoubtful,
		}
	}

	remaining := int64(session.ServerDeadline.Sub(now).Seconds())
	if remaining < 0 {
		remaining = 0
	}

	subjectName := "Ujian CBT"
	if schedule.Subject != nil && schedule.Subject.Name != "" {
		subjectName = schedule.Subject.Name
	} else if schedule.Bank != nil && schedule.Bank.Subject.Name != "" {
		subjectName = schedule.Bank.Subject.Name
	}

	return &ExamPayloadResponse{
		SessionID:         session.ID,
		ScheduleTitle:     schedule.Title,
		SubjectName:       subjectName,
		DurationMinutes:   schedule.DurationMinutes,
		ServerTime:        now,
		ServerDeadline:    session.ServerDeadline,
		RemainingSeconds:  remaining,
		MaxViolations:     schedule.MaxViolations,
		CurrentViolations: session.ViolationCount,
		Questions:         clientQuestions,
		SavedAnswers:      savedMap,
	}, nil
}

// SyncAnswers handles idempotent batch upserts from client
func (s *ExamService) SyncAnswers(sessionID uuid.UUID, studentUserID uuid.UUID, items []SyncAnswerItem) (int, error) {
	var session domain.ExamSession
	if err := s.repo.DB.First(&session, "id = ? AND student_id = ?", sessionID, studentUserID).Error; err != nil {
		return 0, errors.New("sesi tidak valid")
	}

	if session.Status == domain.StatusSubmitted {
		return 0, errors.New("ujian telah dikumpulkan")
	}
	if session.Status == domain.StatusBlocked {
		return 0, errors.New("sesi sedang terblokir")
	}

	now := time.Now()
	// Check grace deadline (30 seconds grace for network latency)
	if now.After(session.ServerDeadline.Add(30 * time.Second)) {
		return 0, errors.New("waktu pengerjaan telah habis")
	}

	count := 0
	for _, item := range items {
		ans := domain.StudentAnswer{
			ID:             uuid.New(),
			SessionID:      sessionID,
			QuestionID:     item.QuestionID,
			SelectedOption: item.SelectedOption,
			AnswerText:     item.AnswerText,
			IsDoubtful:     item.IsDoubtful,
			LastUpdatedAt:  now,
		}

		// Idempotent UPSERT
		err := s.repo.DB.Clauses(clause.OnConflict{
			Columns: []clause.Column{
				{Name: "session_id"},
				{Name: "question_id"},
			},
			DoUpdates: clause.AssignmentColumns([]string{"selected_option", "answer_text", "is_doubtful", "last_updated_at"}),
		}).Create(&ans).Error

		if err == nil {
			count++
		}
	}

	return count, nil
}

// RecordViolation logs anti-cheat event and blocks session if quota reached
func (s *ExamService) RecordViolation(sessionID uuid.UUID, studentUserID uuid.UUID, eventType, details string) (int, bool, error) {
	var session domain.ExamSession
	if err := s.repo.DB.Preload("Schedule").First(&session, "id = ? AND student_id = ?", sessionID, studentUserID).Error; err != nil {
		return 0, false, errors.New("sesi tidak ditemukan")
	}

	if session.Status == domain.StatusSubmitted {
		return session.ViolationCount, false, nil
	}

	// Insert log
	vLog := domain.ViolationLog{
		ID:         uuid.New(),
		SessionID:  sessionID,
		EventType:  eventType,
		Details:    details,
		OccurredAt: time.Now(),
	}
	s.repo.DB.Create(&vLog)

	session.ViolationCount++
	isBlocked := false

	if session.ViolationCount >= session.Schedule.MaxViolations {
		session.Status = domain.StatusBlocked
		isBlocked = true
	}
	s.repo.DB.Save(&session)

	return session.ViolationCount, isBlocked, nil
}

// SubmitExam finalizes the exam and auto-calculates total score for multiple choice & short answers
func (s *ExamService) SubmitExam(sessionID uuid.UUID, studentUserID uuid.UUID) (float64, error) {
	var session domain.ExamSession
	if err := s.repo.DB.Preload("Schedule").First(&session, "id = ? AND student_id = ?", sessionID, studentUserID).Error; err != nil {
		return 0, errors.New("sesi tidak ditemukan")
	}

	if session.Status == domain.StatusSubmitted {
		return session.TotalScore, nil
	}

	// Fetch all questions in bank with correct keys
	var questions []domain.Question
	if err := s.repo.DB.Where("bank_id = ?", session.Schedule.BankID).Find(&questions).Error; err != nil {
		return 0, errors.New("gagal memuat kunci jawaban")
	}

	// Fetch student answers
	var answers []domain.StudentAnswer
	s.repo.DB.Where("session_id = ?", session.ID).Find(&answers)
	answerMap := make(map[string]domain.StudentAnswer)
	for _, a := range answers {
		answerMap[a.QuestionID.String()] = a
	}

	var earnedScore float64 = 0
	var totalWeight float64 = 0

	for _, q := range questions {
		totalWeight += q.ScoreWeight
		studentAns, exists := answerMap[q.ID.String()]
		var awarded float64 = 0
		isGraded := true

		qType := q.Type
		if qType == "" {
			qType = domain.TypeMultipleChoice
		}

		switch qType {
		case domain.TypeShortAnswer:
			rawAnswer := ""
			if exists {
				rawAnswer = studentAns.AnswerText
				if rawAnswer == "" {
					rawAnswer = studentAns.SelectedOption
				}
			}
			normalizedStudent := strings.ToLower(strings.TrimSpace(rawAnswer))
			if normalizedStudent != "" && strings.TrimSpace(q.CorrectKey) != "" {
				keys := strings.Split(q.CorrectKey, "|")
				for _, k := range keys {
					if strings.ToLower(strings.TrimSpace(k)) == normalizedStudent {
						awarded = q.ScoreWeight
						earnedScore += q.ScoreWeight
						break
					}
				}
			}
		case domain.TypeEssay:
			// Essay is graded manually by teacher
			isGraded = false
			awarded = 0
		default: // Multiple Choice
			studentChoice := ""
			if exists {
				studentChoice = studentAns.SelectedOption
			}
			if studentChoice != "" && strings.EqualFold(studentChoice, q.CorrectKey) {
				awarded = q.ScoreWeight
				earnedScore += q.ScoreWeight
			}
		}

		if exists {
			scoreVal := awarded
			s.repo.DB.Model(&domain.StudentAnswer{}).
				Where("id = ?", studentAns.ID).
				Updates(map[string]interface{}{
					"score_awarded": &scoreVal,
					"is_graded":     isGraded,
				})
		}
	}

	finalGrade := 0.0
	if totalWeight > 0 {
		finalGrade = (earnedScore / totalWeight) * 100.0
	}

	now := time.Now()
	session.Status = domain.StatusSubmitted
	session.SubmittedAt = &now
	session.TotalScore = finalGrade
	session.MaxScore = 100.0
	s.repo.DB.Save(&session)

	return finalGrade, nil
}
