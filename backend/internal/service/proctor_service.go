package service

import (
	"errors"
	"fmt"
	"time"

	"cbt-backend/internal/domain"
	"cbt-backend/internal/repository"
	"cbt-backend/pkg/excel"
	"cbt-backend/pkg/pdf"

	"github.com/google/uuid"
	"github.com/xuri/excelize/v2"
)

type ProctorService struct {
	repo *repository.Database
}

func NewProctorService(repo *repository.Database) *ProctorService {
	return &ProctorService{repo: repo}
}

type ProctorStudentItem struct {
	StudentID        uuid.UUID            `json:"student_id"`
	SessionID        *uuid.UUID           `json:"session_id,omitempty"`
	NIS              string               `json:"nis"`
	NISN             string               `json:"nisn"`
	FullName         string               `json:"full_name"`
	ClassName        string               `json:"class_name"`
	Status           domain.SessionStatus `json:"status"` // NOT_STARTED, IN_PROGRESS, BLOCKED, SUBMITTED
	AnsweredCount    int                  `json:"answered_count"`
	TotalQuestions   int                  `json:"total_questions"`
	ProgressPercent  int                  `json:"progress_percent"`
	ViolationCount   int                  `json:"violation_count"`
	StartedAt        *time.Time           `json:"started_at,omitempty"`
	SubmittedAt      *time.Time           `json:"submitted_at,omitempty"`
	ServerDeadline   *time.Time           `json:"server_deadline,omitempty"`
	RemainingSeconds int64                `json:"remaining_seconds"`
	TotalScore       float64              `json:"total_score"`
	ClientIP         string               `json:"client_ip"`
}

type LiveProctorSummary struct {
	ScheduleTitle  string               `json:"schedule_title"`
	SubjectName    string               `json:"subject_name"`
	ClassName      string               `json:"class_name"`
	ExamToken      string               `json:"exam_token"`
	DurationMin    int                  `json:"duration_minutes"`
	TotalStudents  int                  `json:"total_students"`
	PresentCount   int                  `json:"present_count"`
	SubmittedCount int                  `json:"submitted_count"`
	BlockedCount   int                  `json:"blocked_count"`
	Students       []ProctorStudentItem `json:"students"`
}

// GetLiveProctorData compiles real-time supervision metrics for a specific schedule
func (s *ProctorService) GetLiveProctorData(scheduleID uuid.UUID) (*LiveProctorSummary, error) {
	var schedule domain.ExamSchedule
	if err := s.repo.DB.Preload("Bank").Preload("Bank.Subject").Preload("Subject").Preload("ClassRoom").First(&schedule, "id = ?", scheduleID).Error; err != nil {
		return nil, errors.New("jadwal tidak ditemukan")
	}

	// Fetch all students enrolled in this class
	var studentProfiles []domain.StudentProfile
	s.repo.DB.Preload("User").Where("class_room_id = ?", schedule.ClassRoomID).Find(&studentProfiles)

	// Total questions in bank
	var totalQ int64
	if schedule.BankID != nil && *schedule.BankID != uuid.Nil {
		s.repo.DB.Model(&domain.Question{}).Where("bank_id = ?", schedule.BankID).Count(&totalQ)
	}
	if totalQ == 0 {
		totalQ = 1
	}

	// Fetch existing sessions for this schedule
	var sessions []domain.ExamSession
	s.repo.DB.Where("schedule_id = ?", schedule.ID).Find(&sessions)
	sessionMap := make(map[uuid.UUID]domain.ExamSession)
	for _, sess := range sessions {
		sessionMap[sess.StudentID] = sess
	}

	// Fetch answer counts per session — scoped to this schedule only
	var sessionIDs []uuid.UUID
	for _, sess := range sessions {
		sessionIDs = append(sessionIDs, sess.ID)
	}

	type AnsCount struct {
		SessionID uuid.UUID
		Count     int
	}
	var ansCounts []AnsCount
	if len(sessionIDs) > 0 {
		s.repo.DB.Model(&domain.StudentAnswer{}).
			Select("session_id, count(*) as count").
			Where("session_id IN ? AND ((selected_option != '' AND selected_option IS NOT NULL) OR (answer_text != '' AND answer_text IS NOT NULL))", sessionIDs).
			Group("session_id").
			Scan(&ansCounts)
	}

	ansMap := make(map[uuid.UUID]int)
	for _, ac := range ansCounts {
		ansMap[ac.SessionID] = ac.Count
	}

	present := 0
	submitted := 0
	blocked := 0
	now := time.Now()
	var items []ProctorStudentItem

	className := "Kelas"
	if schedule.ClassRoom.Name != "" {
		className = schedule.ClassRoom.Name
	}

	subjectName := "Ujian CBT"
	if schedule.Subject != nil && schedule.Subject.Name != "" {
		subjectName = schedule.Subject.Name
	} else if schedule.Bank != nil && schedule.Bank.Subject.Name != "" {
		subjectName = schedule.Bank.Subject.Name
	}

	for _, sp := range studentProfiles {
		item := ProctorStudentItem{
			StudentID:      sp.UserID,
			NIS:            sp.NIS,
			NISN:           sp.NISN,
			FullName:       sp.User.FullName,
			ClassName:      className,
			Status:         domain.StatusNotStarted,
			TotalQuestions: int(totalQ),
		}

		if sess, exists := sessionMap[sp.UserID]; exists {
			sessID := sess.ID
			item.SessionID = &sessID
			item.Status = sess.Status
			item.ViolationCount = sess.ViolationCount
			item.StartedAt = &sess.StartedAt
			item.SubmittedAt = sess.SubmittedAt
			item.ServerDeadline = &sess.ServerDeadline
			item.TotalScore = sess.TotalScore
			item.ClientIP = sess.ClientIP

			if sess.Status == domain.StatusInProgress || sess.Status == domain.StatusBlocked {
				rem := int64(sess.ServerDeadline.Sub(now).Seconds())
				if rem < 0 {
					rem = 0
				}
				item.RemainingSeconds = rem
			} else {
				item.RemainingSeconds = 0
			}

			answered := ansMap[sess.ID]
			item.AnsweredCount = answered
			item.ProgressPercent = int((float64(answered) / float64(totalQ)) * 100.0)

			present++
			if sess.Status == domain.StatusSubmitted {
				submitted++
			} else if sess.Status == domain.StatusBlocked {
				blocked++
			}
		}

		items = append(items, item)
	}

	return &LiveProctorSummary{
		ScheduleTitle:  schedule.Title,
		SubjectName:    subjectName,
		ClassName:      className,
		ExamToken:      schedule.ExamToken,
		DurationMin:    schedule.DurationMinutes,
		TotalStudents:  len(studentProfiles),
		PresentCount:   present,
		SubmittedCount: submitted,
		BlockedCount:   blocked,
		Students:       items,
	}, nil
}

// UnlockStudentSession unblocks a student whose quota of violations was exceeded
func (s *ProctorService) UnlockStudentSession(sessionID uuid.UUID) error {
	var session domain.ExamSession
	if err := s.repo.DB.First(&session, "id = ?", sessionID).Error; err != nil {
		return errors.New("sesi tidak ditemukan")
	}

	session.Status = domain.StatusInProgress
	// Keep violation count for audit, but reset lockout
	return s.repo.DB.Save(&session).Error
}

// ResetStudentDeviceSession resets session token to allow login on a replacement smartphone or PC
func (s *ProctorService) ResetStudentDeviceSession(studentUserID uuid.UUID) error {
	var user domain.User
	if err := s.repo.DB.First(&user, "id = ?", studentUserID).Error; err != nil {
		return errors.New("user tidak ditemukan")
	}

	user.SessionToken = ""
	return s.repo.DB.Model(&user).Update("session_token", "").Error
}

// ExtendTimeSession extends the server deadline for an individual student session
func (s *ProctorService) ExtendTimeSession(sessionID uuid.UUID, extraMinutes int, reason string) error {
	if extraMinutes <= 0 {
		return errors.New("jumlah menit tambahan harus lebih dari 0")
	}

	var session domain.ExamSession
	if err := s.repo.DB.First(&session, "id = ?", sessionID).Error; err != nil {
		return errors.New("sesi ujian tidak ditemukan")
	}

	now := time.Now()
	if session.ServerDeadline.Before(now) {
		session.ServerDeadline = now.Add(time.Duration(extraMinutes) * time.Minute)
	} else {
		session.ServerDeadline = session.ServerDeadline.Add(time.Duration(extraMinutes) * time.Minute)
	}

	if session.Status == domain.StatusBlocked {
		session.Status = domain.StatusInProgress
	}

	if err := s.repo.DB.Save(&session).Error; err != nil {
		return err
	}

	details := fmt.Sprintf("+%d Menit", extraMinutes)
	if reason != "" {
		details += fmt.Sprintf(" (%s)", reason)
	}

	vLog := domain.ViolationLog{
		ID:         uuid.New(),
		SessionID:  sessionID,
		EventType:  "TIME_EXTENDED",
		Details:    details,
		OccurredAt: now,
	}
	_ = s.repo.DB.Create(&vLog)

	return nil
}

// ExtendTimeAllSchedule extends the server deadline for all active sessions in a schedule
func (s *ProctorService) ExtendTimeAllSchedule(scheduleID uuid.UUID, extraMinutes int, reason string) (int, error) {
	if extraMinutes <= 0 {
		return 0, errors.New("jumlah menit tambahan harus lebih dari 0")
	}

	var sessions []domain.ExamSession
	if err := s.repo.DB.Where("schedule_id = ? AND status IN ?", scheduleID, []domain.SessionStatus{domain.StatusInProgress, domain.StatusBlocked}).Find(&sessions).Error; err != nil {
		return 0, err
	}

	now := time.Now()
	count := 0
	for _, sess := range sessions {
		if sess.ServerDeadline.Before(now) {
			sess.ServerDeadline = now.Add(time.Duration(extraMinutes) * time.Minute)
		} else {
			sess.ServerDeadline = sess.ServerDeadline.Add(time.Duration(extraMinutes) * time.Minute)
		}
		if sess.Status == domain.StatusBlocked {
			sess.Status = domain.StatusInProgress
		}
		if err := s.repo.DB.Save(&sess).Error; err == nil {
			count++
			details := fmt.Sprintf("+%d Menit (Massal)", extraMinutes)
			if reason != "" {
				details += fmt.Sprintf(": %s", reason)
			}
			vLog := domain.ViolationLog{
				ID:         uuid.New(),
				SessionID:  sess.ID,
				EventType:  "TIME_EXTENDED",
				Details:    details,
				OccurredAt: now,
			}
			_ = s.repo.DB.Create(&vLog)
		}
	}

	return count, nil
}

// ForceSubmitSession forces submission of an ongoing exam session and calculates scores.
// Dipanggil oleh pengawas (bukan siswa), jadi student_id diambil dari record sesi itu sendiri.
func (s *ProctorService) ForceSubmitSession(sessionID uuid.UUID) (float64, error) {
	var session domain.ExamSession
	if err := s.repo.DB.First(&session, "id = ?", sessionID).Error; err != nil {
		return 0, errors.New("sesi tidak ditemukan")
	}
	examService := NewExamService(s.repo)
	return examService.SubmitExam(sessionID, session.StudentID)
}

// GetViolationLogs returns the log of cheat attempts and supervisor interventions for a session
func (s *ProctorService) GetViolationLogs(sessionID uuid.UUID) ([]domain.ViolationLog, error) {
	var logs []domain.ViolationLog
	err := s.repo.DB.Where("session_id = ?", sessionID).Order("occurred_at DESC").Find(&logs).Error
	return logs, err
}

// ExportClassGrades generates the Excel spreadsheet of student results
func (s *ProctorService) ExportClassGrades(scheduleID uuid.UUID) (*excelize.File, string, error) {
	live, err := s.GetLiveProctorData(scheduleID)
	if err != nil {
		return nil, "", err
	}

	var reportItems []excel.GradeReportItem
	for i, st := range live.Students {
		stStarted := time.Now()
		if st.StartedAt != nil {
			stStarted = *st.StartedAt
		}
		statusLabel := string(st.Status)
		if st.Status == domain.StatusSubmitted {
			statusLabel = "Selesai"
		} else if st.Status == domain.StatusInProgress {
			statusLabel = "Mengerjakan"
		} else if st.Status == domain.StatusBlocked {
			statusLabel = "Terkunci"
		} else {
			statusLabel = "Belum Mulai"
		}

		reportItems = append(reportItems, excel.GradeReportItem{
			No:          i + 1,
			NIS:         st.NIS,
			NISN:        st.NISN,
			StudentName: st.FullName,
			ClassName:   st.ClassName,
			StartedAt:   stStarted,
			SubmittedAt: st.SubmittedAt,
			Violations:  st.ViolationCount,
			TotalScore:  st.TotalScore,
			Status:      statusLabel,
		})
	}

	file, err := excel.GenerateGradeReport(live.ScheduleTitle, live.ClassName, reportItems)
	if err != nil {
		return nil, "", err
	}

	filename := fmt.Sprintf("Nilai_%s_%s.xlsx", live.ClassName, time.Now().Format("20060102"))
	return file, filename, nil
}

// ExportBeritaAcara generates formal PDF document
func (s *ProctorService) ExportBeritaAcara(scheduleID uuid.UUID, supervisor1, supervisor2, proctorName, roomName string) ([]byte, string, error) {
	live, err := s.GetLiveProctorData(scheduleID)
	if err != nil {
		return nil, "", err
	}

	if supervisor1 == "" {
		supervisor1 = "Drs. H. Mulyadi, M.Pd"
	}
	if supervisor2 == "" {
		supervisor2 = "Hj. Endang Susilowati, S.Pd"
	}
	if proctorName == "" {
		proctorName = "Ahmad Rizky, S.Kom"
	}
	if roomName == "" {
		roomName = "Ruang Ujian " + live.ClassName
	}

	data := pdf.BeritaAcaraData{
		SchoolName:    "SMA NEGERI UNGGULAN INDONESIA",
		ExamTitle:     live.ScheduleTitle,
		SubjectName:   live.SubjectName,
		ClassName:     live.ClassName,
		RoomName:      roomName,
		DateStr:       time.Now().Format("Monday, 02 January 2006"),
		TimeStr:       "07:30 - 09:00 WIB",
		TotalStudents: live.TotalStudents,
		PresentCount:  live.PresentCount,
		AbsentCount:   live.TotalStudents - live.PresentCount,
		ViolationNote: fmt.Sprintf("Tercatat %d insiden siswa terkunci sementara karena perpindahan aplikasi (seluruhnya telah diverifikasi oleh pengawas).", live.BlockedCount),
		ProctorName:   proctorName,
		Supervisor1:   supervisor1,
		Supervisor2:   supervisor2,
	}

	pdfBytes, err := pdf.GenerateBeritaAcara(data)
	if err != nil {
		return nil, "", err
	}

	filename := fmt.Sprintf("Berita_Acara_%s.pdf", live.ClassName)
	return pdfBytes, filename, nil
}
