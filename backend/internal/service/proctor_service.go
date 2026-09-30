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
	"gorm.io/gorm"
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

	// Fetch students: for makeup schedules, only load whitelisted students
	var studentProfiles []domain.StudentProfile
	if schedule.IsMakeup {
		// Susulan: hanya siswa di whitelist
		var makeupStudents []domain.ExamMakeupStudent
		s.repo.DB.Where("schedule_id = ?", scheduleID).Find(&makeupStudents)
		if len(makeupStudents) > 0 {
			ids := make([]uuid.UUID, len(makeupStudents))
			for i, ms := range makeupStudents {
				ids[i] = ms.StudentID
			}
			s.repo.DB.Preload("User").Where("user_id IN ?", ids).Find(&studentProfiles)
		}
	} else {
		s.repo.DB.Preload("User").Where("class_room_id = ?", schedule.ClassRoomID).Find(&studentProfiles)
	}

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

	// Lazy Auto-Close Trigger: jika ada sesi yang telah melewati batas deadline server (+ 30 detik toleransi),
	// otomatis kumpulkan dan nilai agar tampilan pengawas langsung mutakhir menjadi "Selesai"
	now := time.Now()
	cutoff := now.Add(-30 * time.Second)
	examSvc := NewExamService(s.repo)
	needReload := false
	for _, sess := range sessions {
		if (sess.Status == domain.StatusInProgress || sess.Status == domain.StatusBlocked) && sess.ServerDeadline.Before(cutoff) {
			if _, err := examSvc.submitExam(sess.ID, sess.StudentID, true, nil); err == nil {
				needReload = true
			}
		}
	}
	if needReload {
		sessions = nil
		s.repo.DB.Where("schedule_id = ?", schedule.ID).Find(&sessions)
	}

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
	now = time.Now()
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

	if session.Status == domain.StatusSubmitted {
		return errors.New("sesi ini sudah dikumpulkan dan tidak dapat dibuka kembali")
	}
	if session.Status != domain.StatusBlocked {
		return nil // tidak sedang terkunci
	}

	// Syarat status di WHERE mencegah menimpa sesi yang baru saja dikumpulkan.
	res := s.repo.DB.Model(&domain.ExamSession{}).
		Where("id = ? AND status = ?", sessionID, domain.StatusBlocked).
		Updates(reopenSessionFields())
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return errors.New("status sesi telah berubah, muat ulang data")
	}

	_ = s.repo.DB.Create(&domain.ViolationLog{
		ID:         uuid.New(),
		SessionID:  sessionID,
		EventType:  "UNLOCKED",
		Details:    fmt.Sprintf("Kunci dibuka pengawas (total pelanggaran %d)", session.ViolationCount),
		OccurredAt: time.Now(),
	}).Error
	return nil
}

// reopenSessionFields membuka sesi terkunci. Pelanggaran yang sudah terjadi diputihkan lewat
// ViolationBase sehingga siswa kembali mendapat kuota penuh, sementara ViolationCount tetap
// menyimpan total untuk audit pengawas.
// Nilai dasar diambil dari kolom itu sendiri agar pelanggaran yang tercatat bersamaan tidak terlewat.
func reopenSessionFields() map[string]interface{} {
	return map[string]interface{}{
		"status":         domain.StatusInProgress,
		"violation_base": gorm.Expr("violation_count"),
		"updated_at":     time.Now(),
	}
}

// extendSessionDeadline memperpanjang deadline sesi dan membuka kuncinya bila terkunci. Hanya kolom
// terkait yang ditulis, sehingga hitungan pelanggaran atau status submit yang berubah bersamaan aman.
func extendSessionDeadline(db *gorm.DB, session domain.ExamSession, extraMinutes int, now time.Time) error {
	deadline := session.ServerDeadline
	if deadline.Before(now) {
		deadline = now
	}
	deadline = deadline.Add(time.Duration(extraMinutes) * time.Minute)

	if err := db.Model(&domain.ExamSession{}).Where("id = ?", session.ID).Updates(map[string]interface{}{
		"server_deadline": deadline,
		"updated_at":      now,
	}).Error; err != nil {
		return err
	}
	return db.Model(&domain.ExamSession{}).
		Where("id = ? AND status = ?", session.ID, domain.StatusBlocked).
		Updates(reopenSessionFields()).Error
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

// ResetExamSession menghapus sesi ujian dan semua jawaban siswa agar dapat mengulang dari awal.
func (s *ProctorService) ResetExamSession(sessionID uuid.UUID) error {
	var session domain.ExamSession
	if err := s.repo.DB.First(&session, "id = ?", sessionID).Error; err != nil {
		return errors.New("sesi ujian tidak ditemukan")
	}
	if err := s.repo.DB.Where("session_id = ?", sessionID).Delete(&domain.StudentAnswer{}).Error; err != nil {
		return fmt.Errorf("gagal menghapus jawaban siswa: %w", err)
	}
	if err := s.repo.DB.Delete(&session).Error; err != nil {
		return fmt.Errorf("gagal menghapus sesi ujian: %w", err)
	}
	return nil
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
	if err := extendSessionDeadline(s.repo.DB, session, extraMinutes, now); err != nil {
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
		if err := extendSessionDeadline(s.repo.DB, sess, extraMinutes, now); err == nil {
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
	return examService.ForceSubmit(sessionID, session.StudentID)
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

// ExportMergedGrades menghasilkan Excel rekap nilai satu kelas yang menggabungkan
// jadwal reguler (scheduleID) dengan semua jadwal susulan yang merujuk ke jadwal ini.
// Setiap siswa diambil nilainya dari: jadwal susulan (prioritas) jika ada, jadwal reguler jika tidak.
func (s *ProctorService) ExportMergedGrades(scheduleID uuid.UUID) (*excelize.File, string, error) {
	// Load jadwal induk
	var schedule domain.ExamSchedule
	if err := s.repo.DB.Preload("Bank").Preload("Bank.Subject").Preload("Subject").Preload("ClassRoom").
		First(&schedule, "id = ?", scheduleID).Error; err != nil {
		return nil, "", errors.New("jadwal tidak ditemukan")
	}
	if schedule.IsMakeup {
		return nil, "", errors.New("gunakan jadwal induk (reguler) untuk export gabungan")
	}

	// Load semua jadwal susulan yang merujuk ke jadwal ini
	var makeupSchedules []domain.ExamSchedule
	s.repo.DB.Where("parent_schedule_id = ? AND is_makeup = ?", scheduleID, true).Find(&makeupSchedules)

	// Load semua siswa kelas reguler
	var studentProfiles []domain.StudentProfile
	s.repo.DB.Preload("User").Where("class_room_id = ?", schedule.ClassRoomID).Find(&studentProfiles)

	// Load semua sesi dari jadwal induk
	var regularSessions []domain.ExamSession
	s.repo.DB.Where("schedule_id = ?", scheduleID).Find(&regularSessions)
	regularSessionMap := make(map[uuid.UUID]domain.ExamSession)
	for _, sess := range regularSessions {
		regularSessionMap[sess.StudentID] = sess
	}

	// Load semua sesi dari jadwal susulan (semua makeup children)
	makeupSessionMap := make(map[uuid.UUID]domain.ExamSession)
	for _, ms := range makeupSchedules {
		var mkSessions []domain.ExamSession
		s.repo.DB.Where("schedule_id = ?", ms.ID).Find(&mkSessions)
		for _, sess := range mkSessions {
			makeupSessionMap[sess.StudentID] = sess
		}
	}

	// Tentukan nama kelas & mapel
	className := schedule.ClassRoom.Name
	subjectName := "Ujian CBT"
	if schedule.Subject != nil && schedule.Subject.Name != "" {
		subjectName = schedule.Subject.Name
	} else if schedule.Bank != nil && schedule.Bank.Subject.Name != "" {
		subjectName = schedule.Bank.Subject.Name
	}

	var reportItems []excel.GradeReportItem
	for i, sp := range studentProfiles {
		var sess domain.ExamSession
		var sumber string

		if mkSess, ok := makeupSessionMap[sp.UserID]; ok {
			sess = mkSess
			sumber = "Susulan"
		} else if regSess, ok := regularSessionMap[sp.UserID]; ok {
			sess = regSess
			sumber = "Reguler"
		} else {
			// Tidak ikut ujian sama sekali
			reportItems = append(reportItems, excel.GradeReportItem{
				No:          i + 1,
				NIS:         sp.NIS,
				NISN:        sp.NISN,
				StudentName: sp.User.FullName,
				ClassName:   className,
				StartedAt:   time.Time{},
				Status:      "Tidak Hadir",
			})
			continue
		}

		statusLabel := string(sess.Status)
		switch sess.Status {
		case domain.StatusSubmitted:
			statusLabel = "Selesai (" + sumber + ")"
		case domain.StatusInProgress:
			statusLabel = "Mengerjakan"
		case domain.StatusBlocked:
			statusLabel = "Terkunci"
		default:
			statusLabel = "Belum Mulai"
		}

		stStarted := time.Now()
		if !sess.StartedAt.IsZero() {
			stStarted = sess.StartedAt
		}

		reportItems = append(reportItems, excel.GradeReportItem{
			No:          i + 1,
			NIS:         sp.NIS,
			NISN:        sp.NISN,
			StudentName: sp.User.FullName,
			ClassName:   className,
			StartedAt:   stStarted,
			SubmittedAt: sess.SubmittedAt,
			Violations:  sess.ViolationCount,
			TotalScore:  sess.TotalScore,
			Status:      statusLabel,
		})
	}

	title := schedule.Title + " — " + subjectName + " (Rekap Gabungan)"
	file, err := excel.GenerateGradeReport(title, className, reportItems)
	if err != nil {
		return nil, "", err
	}

	filename := fmt.Sprintf("Nilai_Gabungan_%s_%s.xlsx", className, time.Now().Format("20060102"))
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
