package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"sort"
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

	now := time.Now()

	// Query 1: schedule non-susulan untuk kelas siswa
	var regularSchedules []domain.ExamSchedule
	err := s.repo.DB.
		Preload("Event").
		Preload("Subject").
		Preload("Bank").
		Preload("Bank.Subject").
		Preload("ClassRoom").
		Joins("LEFT JOIN exam_events ON exam_events.id = exam_schedules.event_id").
		Joins("LEFT JOIN question_banks ON question_banks.id = exam_schedules.bank_id").
		Where("class_room_id = ? AND exam_schedules.is_active = ? AND start_time <= ? AND end_time >= ? AND (exam_schedules.event_id IS NULL OR exam_events.is_active = ?) AND exam_schedules.bank_id IS NOT NULL AND question_banks.is_locked = ? AND exam_schedules.is_makeup = ?",
			profile.ClassRoomID, true, now, now, true, true, false).
		Order("start_time ASC").
		Find(&regularSchedules).Error
	if err != nil {
		return nil, err
	}

	// Query 2: schedule susulan di mana siswa ada di whitelist
	var makeupSchedules []domain.ExamSchedule
	err = s.repo.DB.
		Preload("Event").
		Preload("Subject").
		Preload("Bank").
		Preload("Bank.Subject").
		Preload("ClassRoom").
		Joins("LEFT JOIN exam_events ON exam_events.id = exam_schedules.event_id").
		Joins("LEFT JOIN question_banks ON question_banks.id = exam_schedules.bank_id").
		Joins("JOIN exam_makeup_students ON exam_makeup_students.schedule_id = exam_schedules.id AND exam_makeup_students.student_id = ?", studentUserID).
		Where("exam_schedules.is_active = ? AND start_time <= ? AND end_time >= ? AND (exam_schedules.event_id IS NULL OR exam_events.is_active = ?) AND exam_schedules.bank_id IS NOT NULL AND question_banks.is_locked = ? AND exam_schedules.is_makeup = ?",
			true, now, now, true, true, true).
		Order("start_time ASC").
		Find(&makeupSchedules).Error
	if err != nil {
		return nil, err
	}

	// Gabungkan dan sort by start_time ASC
	schedules := append(regularSchedules, makeupSchedules...)
	sort.Slice(schedules, func(i, j int) bool {
		return schedules[i].StartTime.Before(schedules[j].StartTime)
	})

	ClearExamTokens(schedules)
	return schedules, nil
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

	// Jika susulan: validasi whitelist siswa
	if schedule.IsMakeup {
		var count int64
		s.repo.DB.Model(&domain.ExamMakeupStudent{}).
			Where("schedule_id = ? AND student_id = ?", schedule.ID, studentUserID).
			Count(&count)
		if count == 0 {
			return nil, errors.New("Anda tidak terdaftar sebagai peserta ujian susulan ini")
		}
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
		// SEMENTARA DINONAKTIFKAN: auto-unblock sesi terkunci agar siswa bisa langsung lanjut.
		// Aktifkan kembali blok di bawah setelah fitur pelanggaran dihidupkan ulang.
		if session.Status == domain.StatusBlocked {
			s.repo.DB.Model(&domain.ExamSession{}).Where("id = ?", session.ID).Updates(map[string]interface{}{
				"status":     domain.StatusInProgress,
				"updated_at": now,
			})
			session.Status = domain.StatusInProgress
		}
		// Refresh IP & UserAgent tanpa menimpa kolom lain (status, pelanggaran) yang bisa berubah bersamaan.
		s.repo.DB.Model(&domain.ExamSession{}).Where("id = ?", session.ID).Updates(map[string]interface{}{
			"client_ip":  clientIP,
			"user_agent": userAgent,
			"updated_at": now,
		})
	}

	// Waktu habis: kumpulkan dan nilai jawaban yang sudah tersinkron, bukan sekadar menandai selesai.
	if now.After(session.ServerDeadline) {
		if _, err := s.SubmitExam(session.ID, studentUserID, nil); err != nil {
			return nil, err
		}
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
	if err := s.repo.DB.Where("session_id = ?", session.ID).Find(&answers).Error; err != nil {
		return nil, fmt.Errorf("gagal memuat jawaban tersimpan: %w", err)
	}
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
		CurrentViolations: session.ActiveViolations(),
		Questions:         clientQuestions,
		SavedAnswers:      savedMap,
	}, nil
}

// SyncAnswers handles idempotent batch upserts from client. Deadline sesi ikut dikembalikan agar
// klien menyesuaikan timer, mis. setelah pengawas menambah waktu. Daftar jawaban boleh kosong
// (heartbeat) untuk sekadar mengambil deadline terbaru.
func (s *ExamService) SyncAnswers(sessionID uuid.UUID, studentUserID uuid.UUID, items []SyncAnswerItem) (int, time.Time, error) {
	var session domain.ExamSession
	if err := s.repo.DB.Preload("Schedule").First(&session, "id = ? AND student_id = ?", sessionID, studentUserID).Error; err != nil {
		return 0, time.Time{}, errors.New("sesi tidak valid")
	}

	if session.Status == domain.StatusSubmitted {
		return 0, session.ServerDeadline, errors.New("ujian telah dikumpulkan")
	}
	if session.Status == domain.StatusBlocked {
		return 0, session.ServerDeadline, errors.New("sesi sedang terblokir")
	}

	now := time.Now()
	// Check grace deadline (30 seconds grace for network latency)
	if now.After(session.ServerDeadline.Add(30 * time.Second)) {
		return 0, session.ServerDeadline, errors.New("waktu pengerjaan telah habis")
	}

	// Hanya soal dari bank jadwal sesi ini yang diterima; ID lain dilewati agar tabel jawaban
	// tidak terisi baris sampah dari request yang dimanipulasi.
	validIDs, err := s.bankQuestionIDs(session.Schedule.BankID, items)
	if err != nil {
		return 0, session.ServerDeadline, errors.New("gagal memvalidasi soal")
	}

	count := 0
	for _, item := range items {
		if !validIDs[item.QuestionID] {
			continue
		}
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

	return count, session.ServerDeadline, nil
}

// bankQuestionIDs mengembalikan himpunan ID soal pada items yang benar-benar milik bank.
func (s *ExamService) bankQuestionIDs(bankID *uuid.UUID, items []SyncAnswerItem) (map[uuid.UUID]bool, error) {
	valid := make(map[uuid.UUID]bool)
	if bankID == nil || len(items) == 0 {
		return valid, nil
	}
	ids := make([]uuid.UUID, 0, len(items))
	for _, it := range items {
		ids = append(ids, it.QuestionID)
	}
	var found []uuid.UUID
	if err := s.repo.DB.Model(&domain.Question{}).
		Where("bank_id = ? AND id IN ?", *bankID, ids).
		Pluck("id", &found).Error; err != nil {
		return nil, err
	}
	for _, id := range found {
		valid[id] = true
	}
	return valid, nil
}

// RecordViolation logs anti-cheat event and blocks session if quota reached.
// Penambahan hitungan dan penguncian dilakukan dengan UPDATE bersyarat di satu transaksi, sehingga
// laporan yang datang bersamaan tidak saling menghilangkan dan tidak menimpa sesi yang sudah dikumpulkan.
func (s *ExamService) RecordViolation(sessionID uuid.UUID, studentUserID uuid.UUID, eventType, details string) (int, bool, error) {
	var session domain.ExamSession
	isBlocked := false
	err := s.repo.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Preload("Schedule").First(&session, "id = ? AND student_id = ?", sessionID, studentUserID).Error; err != nil {
			return errSessionNotFound
		}
		if session.Status == domain.StatusSubmitted {
			return nil
		}

		if err := tx.Create(&domain.ViolationLog{
			ID:         uuid.New(),
			SessionID:  sessionID,
			EventType:  eventType,
			Details:    details,
			OccurredAt: time.Now(),
		}).Error; err != nil {
			return err
		}

		res := tx.Model(&domain.ExamSession{}).
			Where("id = ? AND status <> ?", sessionID, domain.StatusSubmitted).
			Updates(map[string]interface{}{
				"violation_count": gorm.Expr("violation_count + 1"),
				"updated_at":      time.Now(),
			})
		if res.Error != nil {
			return res.Error
		}

		res = tx.Model(&domain.ExamSession{}).
			Where("id = ? AND status = ? AND violation_count - violation_base >= ?",
				sessionID, domain.StatusInProgress, session.Schedule.MaxViolations).
			Update("status", domain.StatusBlocked)
		if res.Error != nil {
			return res.Error
		}
		isBlocked = res.RowsAffected > 0

		return tx.First(&session, "id = ?", sessionID).Error
	})
	if err != nil {
		return 0, false, err
	}
	return session.ActiveViolations(), isBlocked, nil
}

var errSessionNotFound = errors.New("sesi tidak ditemukan")

var errSessionBlocked = errors.New("ujian Anda sedang terkunci. Hubungi pengawas ruangan")

// SubmitExam dipanggil siswa. Sesi yang terkunci tidak bisa dikumpulkan sendiri: pengawas yang
// memutuskan membuka kunci atau mengumpulkan paksa (ForceSubmit).
// finalAnswers adalah sisa pending queue klien yang belum tersinkron; disimpan atomik sebelum penilaian.
func (s *ExamService) SubmitExam(sessionID uuid.UUID, studentUserID uuid.UUID, finalAnswers []SyncAnswerItem) (float64, error) {
	return s.submitExam(sessionID, studentUserID, false, finalAnswers)
}

// ForceSubmit dipanggil pengawas dan boleh mengumpulkan sesi yang terkunci.
func (s *ExamService) ForceSubmit(sessionID uuid.UUID, studentUserID uuid.UUID) (float64, error) {
	return s.submitExam(sessionID, studentUserID, true, nil)
}

// submitExam finalizes the exam and auto-calculates total score for multiple choice & short answers.
// Sesi "diklaim" lebih dulu dengan UPDATE bersyarat status <> SUBMITTED di dalam transaksi: dari
// beberapa submit bersamaan (siswa, timer habis, pengawas) hanya satu yang menilai, sisanya
// mengembalikan nilai yang sudah tersimpan.
func (s *ExamService) submitExam(sessionID uuid.UUID, studentUserID uuid.UUID, allowBlocked bool, finalAnswers []SyncAnswerItem) (float64, error) {
	var finalGrade float64
	err := s.repo.DB.Transaction(func(tx *gorm.DB) error {
		var session domain.ExamSession
		if err := tx.Preload("Schedule").First(&session, "id = ? AND student_id = ?", sessionID, studentUserID).Error; err != nil {
			return errSessionNotFound
		}
		if session.Status == domain.StatusSubmitted {
			finalGrade = session.TotalScore
			return nil
		}
		if session.Status == domain.StatusBlocked && !allowBlocked {
			return errSessionBlocked
		}

		now := time.Now()
		claimable := []domain.SessionStatus{domain.StatusInProgress, domain.StatusNotStarted}
		if allowBlocked {
			claimable = append(claimable, domain.StatusBlocked)
		}
		claim := tx.Model(&domain.ExamSession{}).
			Where("id = ? AND status IN ?", sessionID, claimable).
			Updates(map[string]interface{}{
				"status":       domain.StatusSubmitted,
				"submitted_at": now,
				"updated_at":   now,
			})
		if claim.Error != nil {
			return claim.Error
		}
		if claim.RowsAffected == 0 {
			// Didahului submit lain (kembalikan nilainya) atau sesi baru saja terkunci.
			if err := tx.First(&session, "id = ?", sessionID).Error; err != nil {
				return err
			}
			if session.Status == domain.StatusBlocked {
				return errSessionBlocked
			}
			finalGrade = session.TotalScore
			return nil
		}

		// Simpan sisa jawaban yang belum tersinkron dari klien sebelum penilaian.
		// Ini adalah pengaman terakhir: jika sync berkala gagal karena jaringan atau
		// grace deadline habis, jawaban dari pending queue tetap tersimpan.
		if len(finalAnswers) > 0 {
			validIDs, _ := s.bankQuestionIDs(session.Schedule.BankID, finalAnswers)
			now2 := time.Now()
			for _, item := range finalAnswers {
				if !validIDs[item.QuestionID] {
					continue
				}
				ans := domain.StudentAnswer{
					ID:             uuid.New(),
					SessionID:      sessionID,
					QuestionID:     item.QuestionID,
					SelectedOption: item.SelectedOption,
					AnswerText:     item.AnswerText,
					IsDoubtful:     item.IsDoubtful,
					LastUpdatedAt:  now2,
				}
				tx.Clauses(clause.OnConflict{
					Columns:   []clause.Column{{Name: "session_id"}, {Name: "question_id"}},
					DoUpdates: clause.AssignmentColumns([]string{"selected_option", "answer_text", "is_doubtful", "last_updated_at"}),
				}).Create(&ans)
			}
		}

		// Fetch all questions in bank with correct keys
		var questions []domain.Question
		if err := tx.Where("bank_id = ?", session.Schedule.BankID).Find(&questions).Error; err != nil {
			return errors.New("gagal memuat kunci jawaban")
		}

		// Fetch student answers
		var answers []domain.StudentAnswer
		if err := tx.Where("session_id = ?", session.ID).Find(&answers).Error; err != nil {
			return errors.New("gagal memuat jawaban siswa")
		}
		answerMap := make(map[string]domain.StudentAnswer)
		for _, a := range answers {
			answerMap[a.QuestionID.String()] = a
		}

		var earnedScore float64 = 0
		var totalWeight float64 = 0

		for _, q := range questions {
			totalWeight += q.ScoreWeight
			studentAns, exists := answerMap[q.ID.String()]
			awarded, isGraded := gradeAnswer(q, studentAns, exists)
			earnedScore += awarded

			if exists {
				scoreVal := awarded
				if err := tx.Model(&domain.StudentAnswer{}).
					Where("id = ?", studentAns.ID).
					Updates(map[string]interface{}{
						"score_awarded": &scoreVal,
						"is_graded":     isGraded,
					}).Error; err != nil {
					return err
				}
			}
		}

		if totalWeight > 0 {
			finalGrade = (earnedScore / totalWeight) * 100.0
		}

		return tx.Model(&domain.ExamSession{}).Where("id = ?", sessionID).Updates(map[string]interface{}{
			"total_score": finalGrade,
			"max_score":   100.0,
		}).Error
	})
	if err != nil {
		return 0, err
	}
	return finalGrade, nil
}

// gradeAnswer menilai satu jawaban. Esai dinilai manual oleh guru (isGraded = false).
func gradeAnswer(q domain.Question, studentAns domain.StudentAnswer, exists bool) (awarded float64, isGraded bool) {
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
			for _, k := range strings.Split(q.CorrectKey, "|") {
				if strings.ToLower(strings.TrimSpace(k)) == normalizedStudent {
					return q.ScoreWeight, true
				}
			}
		}
		return 0, true
	case domain.TypeEssay:
		return 0, false
	default: // Multiple Choice
		if exists && studentAns.SelectedOption != "" && strings.EqualFold(studentAns.SelectedOption, q.CorrectKey) {
			return q.ScoreWeight, true
		}
		return 0, true
	}
}
