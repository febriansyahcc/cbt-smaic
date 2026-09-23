package domain

import (
	"time"

	"github.com/google/uuid"
)

type Role string

const (
	RoleAdmin Role = "ADMIN"
	RoleGuru  Role = "GURU"
	RoleSiswa Role = "SISWA"
)

type Permission string

const (
	PermAll               Permission = "*"
	PermQuestionsAssigned Permission = "questions:read_assigned"
	PermQuestionsAll      Permission = "questions:read_all"
	PermQuestionsUpload   Permission = "questions:upload"
	PermQuestionsLock     Permission = "questions:lock"
	PermSchedulesRead     Permission = "schedules:read"
	PermSchedulesManage   Permission = "schedules:manage"
	PermProctorView       Permission = "proctor:view"
	PermProctorControl    Permission = "proctor:control"
	PermReportsExport     Permission = "reports:export"
	PermMasterManage      Permission = "master:manage"
	PermExamTake          Permission = "exam:take"
)

type SessionStatus string

const (
	StatusNotStarted SessionStatus = "NOT_STARTED"
	StatusInProgress SessionStatus = "IN_PROGRESS"
	StatusBlocked    SessionStatus = "BLOCKED"
	StatusSubmitted  SessionStatus = "SUBMITTED"
)

type User struct {
	ID           uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	Username     string    `gorm:"size:100;uniqueIndex;not null" json:"username"`
	PasswordHash string    `gorm:"size:255;not null" json:"-"`
	FullName     string    `gorm:"size:255;not null" json:"full_name"`
	Role         Role      `gorm:"size:20;not null;default:'SISWA'" json:"role"`
	Permissions  []string  `gorm:"serializer:json" json:"permissions"`
	SessionToken string    `gorm:"size:255" json:"-"`
	IsActive     bool      `gorm:"default:true" json:"is_active"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

func GetDefaultPermissions(role Role) []string {
	switch role {
	case RoleAdmin:
		return []string{string(PermAll)}
	case RoleGuru:
		// Sumber tunggal: template "guru" di permissions.go
		return TemplatePermissions("guru")
	case RoleSiswa:
		return TemplatePermissions("siswa")
	default:
		return []string{}
	}
}

// EffectivePermissions menghitung izin yang berlaku dari data tersimpan. Data lama yang
// tidak bersih tidak boleh memberi izin berlebih:
//   - ADMIN selalu "*".
//   - SISWA selalu template siswa (exam:take); Permissions tersimpan diabaikan.
//   - Role lain: daftar tersimpan disanitasi (membuang "*", exam:take, dan kunci tak dikenal,
//     menambah izin turunan). Bila tidak ada daftar, atau hasil sanitasi kosong, dipakai
//     izin bawaan role.
func (u *User) EffectivePermissions() []string {
	switch u.Role {
	case RoleAdmin:
		return []string{string(PermAll)}
	case RoleSiswa:
		return TemplatePermissions("siswa")
	}
	if len(u.Permissions) > 0 {
		if sanitized := SanitizePermissions(u.Role, u.Permissions); len(sanitized) > 0 {
			return sanitized
		}
	}
	return GetDefaultPermissions(u.Role)
}

func (u *User) HasPermission(perm string) bool {
	if u.Role == RoleAdmin {
		return true
	}
	eff := u.EffectivePermissions()
	for _, p := range eff {
		if p == string(PermAll) || p == perm {
			return true
		}
	}
	return false
}

type ClassRoom struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	Name      string    `gorm:"size:100;not null" json:"name"`
	Grade     string    `gorm:"size:20;not null" json:"grade"`
	Major     string    `gorm:"size:50" json:"major"`
	CreatedAt time.Time `json:"created_at"`
}

type StudentProfile struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	UserID      uuid.UUID `gorm:"type:uuid;uniqueIndex;not null" json:"user_id"`
	User        User      `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"user,omitempty"`
	NISN        string    `gorm:"size:20;uniqueIndex" json:"nisn"`
	NIS         string    `gorm:"size:20;uniqueIndex;not null" json:"nis"`
	ClassRoomID uuid.UUID `gorm:"type:uuid;index;not null" json:"class_id"`
	ClassRoom   ClassRoom `gorm:"foreignKey:ClassRoomID" json:"class_room,omitempty"`
	Gender      string    `gorm:"size:10" json:"gender"`
	CreatedAt   time.Time `json:"created_at"`
}

type Subject struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	Code      string    `gorm:"size:50;uniqueIndex;not null" json:"code"`
	Name      string    `gorm:"size:100;not null" json:"name"`
	CreatedAt time.Time `json:"created_at"`
}

type ClassSubject struct {
	ID           uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	ClassRoomID  uuid.UUID `gorm:"type:uuid;index;not null" json:"class_id"`
	ClassRoom    ClassRoom `gorm:"foreignKey:ClassRoomID;constraint:OnDelete:CASCADE" json:"class_room,omitempty"`
	SubjectID    uuid.UUID `gorm:"type:uuid;index;not null" json:"subject_id"`
	Subject      Subject   `gorm:"foreignKey:SubjectID;constraint:OnDelete:CASCADE" json:"subject,omitempty"`
	TeacherID    uuid.UUID `gorm:"type:uuid;index;not null" json:"teacher_id"`
	Teacher      User      `gorm:"foreignKey:TeacherID;constraint:OnDelete:CASCADE" json:"teacher,omitempty"`
	AcademicYear string    `gorm:"size:20;not null" json:"academic_year"`
	CreatedAt    time.Time `json:"created_at"`
}

type QuestionType string

const (
	TypeMultipleChoice QuestionType = "MULTIPLE_CHOICE"
	TypeShortAnswer    QuestionType = "SHORT_ANSWER"
	TypeEssay          QuestionType = "ESSAY"
)

type OptionItem struct {
	Key      string `json:"key"`
	Text     string `json:"text"`
	ImageURL string `json:"image_url,omitempty"`
}

type QuestionBank struct {
	ID             uuid.UUID   `gorm:"type:uuid;primaryKey" json:"id"`
	Title          string      `gorm:"size:255;not null" json:"title"`
	SubjectID      uuid.UUID   `gorm:"type:uuid;index;not null" json:"subject_id"`
	Subject        Subject     `gorm:"foreignKey:SubjectID" json:"subject,omitempty"`
	Grade          string      `gorm:"size:20;not null;default:''" json:"grade"`                                   // Tingkat kelas (X/XI/XII); kosong = bank lama tanpa tingkat
	Classes        []ClassRoom `gorm:"many2many:question_bank_classes;constraint:OnDelete:CASCADE" json:"classes"` // Cakupan kelas khusus; kosong = ikut Grade
	CreatedByID    uuid.UUID   `gorm:"type:uuid;not null" json:"created_by_id"`
	CreatedBy      User        `gorm:"foreignKey:CreatedByID" json:"created_by,omitempty"`
	TotalQuestions int         `gorm:"default:0" json:"total_questions"`
	IsLocked       bool        `gorm:"default:false" json:"is_locked"`
	CreatedAt      time.Time   `json:"created_at"`
	Questions      []Question  `gorm:"foreignKey:BankID;constraint:OnDelete:CASCADE" json:"questions,omitempty"`
}

type Question struct {
	ID             uuid.UUID    `gorm:"type:uuid;primaryKey" json:"id"`
	BankID         uuid.UUID    `gorm:"type:uuid;index;not null" json:"bank_id"`
	QuestionNumber int          `gorm:"not null" json:"question_number"`
	Type           QuestionType `gorm:"size:20;not null;default:'MULTIPLE_CHOICE'" json:"type"`
	ContentHTML    string       `gorm:"type:text;not null" json:"content_html"`
	OptionsJSON    string       `gorm:"type:text;not null;default:'[]'" json:"options_json"`
	CorrectKey     string       `gorm:"size:255;not null;default:''" json:"correct_key,omitempty"`
	RubricGuide    string       `gorm:"type:text" json:"rubric_guide,omitempty"`
	ScoreWeight    float64      `gorm:"default:1.0" json:"score_weight"`
	CreatedAt      time.Time    `json:"created_at"`
}

type ExamEvent struct {
	ID           uuid.UUID      `gorm:"type:uuid;primaryKey" json:"id"`
	Title        string         `gorm:"size:255;not null" json:"title"`
	Code         string         `gorm:"size:50;not null;uniqueIndex" json:"code"`
	AcademicYear string         `gorm:"size:20;not null" json:"academic_year"` // e.g. "2025/2026"
	Semester     string         `gorm:"size:20;not null" json:"semester"`      // "GANJIL" / "GENAP"
	StartDate    time.Time      `json:"start_date"`
	EndDate      time.Time      `json:"end_date"`
	IsActive     bool           `gorm:"default:false" json:"is_active"`
	Description  string         `gorm:"type:text" json:"description"`
	CreatedAt    time.Time      `json:"created_at"`
	Schedules    []ExamSchedule `gorm:"foreignKey:EventID" json:"schedules,omitempty"`
}

type ExamSchedule struct {
	ID                 uuid.UUID     `gorm:"type:uuid;primaryKey" json:"id"`
	EventID            *uuid.UUID    `gorm:"type:uuid;index" json:"event_id,omitempty"`
	Event              *ExamEvent    `gorm:"foreignKey:EventID" json:"event,omitempty"`
	Title              string        `gorm:"size:255;not null" json:"title"`
	SubjectID          *uuid.UUID    `gorm:"type:uuid;index" json:"subject_id,omitempty"`
	Subject            *Subject      `gorm:"foreignKey:SubjectID" json:"subject,omitempty"`
	BankID             *uuid.UUID    `gorm:"type:uuid;index" json:"bank_id,omitempty"`
	Bank               *QuestionBank `gorm:"foreignKey:BankID" json:"bank,omitempty"`
	ClassRoomID        uuid.UUID     `gorm:"type:uuid;index;not null" json:"class_id"`
	ClassRoom          ClassRoom     `gorm:"foreignKey:ClassRoomID" json:"class_room,omitempty"`
	ExamToken          string        `gorm:"size:20;not null" json:"exam_token"`
	StartTime          time.Time     `json:"start_time"`
	EndTime            time.Time     `json:"end_time"`
	DurationMinutes    int           `gorm:"not null" json:"duration_minutes"`
	MaxViolations      int           `gorm:"default:3" json:"max_violations"`
	RandomizeQuestions bool          `gorm:"default:true" json:"randomize_questions"`
	RandomizeOptions   bool          `gorm:"default:true" json:"randomize_options"`
	IsActive           bool          `gorm:"default:true" json:"is_active"`
	CreatedAt          time.Time     `json:"created_at"`
}

type ExamSession struct {
	ID                uuid.UUID     `gorm:"type:uuid;primaryKey" json:"id"`
	ScheduleID        uuid.UUID     `gorm:"type:uuid;index;not null" json:"schedule_id"`
	Schedule          ExamSchedule  `gorm:"foreignKey:ScheduleID" json:"schedule,omitempty"`
	StudentID         uuid.UUID     `gorm:"type:uuid;index;not null" json:"student_id"`
	Student           User          `gorm:"foreignKey:StudentID" json:"student,omitempty"`
	StartedAt         time.Time     `json:"started_at"`
	ServerDeadline    time.Time     `json:"server_deadline"`
	SubmittedAt       *time.Time    `json:"submitted_at,omitempty"`
	Status            SessionStatus `gorm:"size:20;default:'IN_PROGRESS'" json:"status"`
	ViolationCount    int           `gorm:"default:0" json:"violation_count"`
	ClientIP          string        `gorm:"size:100" json:"client_ip"`
	UserAgent         string        `gorm:"size:255" json:"user_agent"`
	DeviceFingerprint string        `gorm:"size:255" json:"device_fingerprint"`
	TotalScore        float64       `gorm:"default:0.0" json:"total_score"`
	MaxScore          float64       `gorm:"default:100.0" json:"max_score"`
	CreatedAt         time.Time     `json:"created_at"`
	UpdatedAt         time.Time     `json:"updated_at"`
}

type StudentAnswer struct {
	ID             uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	SessionID      uuid.UUID `gorm:"type:uuid;uniqueIndex:idx_session_question;not null" json:"session_id"`
	QuestionID     uuid.UUID `gorm:"type:uuid;uniqueIndex:idx_session_question;not null" json:"question_id"`
	SelectedOption string    `gorm:"size:10" json:"selected_option"`
	AnswerText     string    `gorm:"type:text" json:"answer_text,omitempty"`
	ScoreAwarded   *float64  `gorm:"default:null" json:"score_awarded,omitempty"`
	IsGraded       bool      `gorm:"default:true" json:"is_graded"`
	TeacherComment string    `gorm:"type:text" json:"teacher_comment,omitempty"`
	IsDoubtful     bool      `gorm:"default:false" json:"is_doubtful"`
	LastUpdatedAt  time.Time `json:"last_updated_at"`
}

type ViolationLog struct {
	ID         uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	SessionID  uuid.UUID `gorm:"type:uuid;index;not null" json:"session_id"`
	EventType  string    `gorm:"size:50;not null" json:"event_type"`
	Details    string    `gorm:"size:255" json:"details"`
	OccurredAt time.Time `json:"occurred_at"`
}

// EventParticipant adalah akun peserta per event yang dicetak pada kartu peserta.
// Siswa dapat login memakai ExamNumber + Password selama event aktif; login NIS tetap berlaku.
// Password disimpan apa adanya karena harus bisa dicetak ulang; berlaku hanya selama event aktif.
type EventParticipant struct {
	ID         uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	EventID    uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_event_participant_user" json:"event_id"`
	UserID     uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_event_participant_user;index" json:"user_id"`
	ExamNumber string    `gorm:"size:50;not null;uniqueIndex" json:"exam_number"`
	Password   string    `gorm:"size:20;not null" json:"password"`
	CreatedAt  time.Time `json:"created_at"`
}
