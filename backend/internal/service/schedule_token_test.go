package service

import (
	"testing"
	"time"

	"cbt-backend/internal/domain"
	"cbt-backend/internal/repository"

	"github.com/google/uuid"
)

func TestClearExamTokens(t *testing.T) {
	schedules := []domain.ExamSchedule{
		{ID: uuid.New(), ExamToken: "AAA"},
		{ID: uuid.New(), ExamToken: "BBB"},
	}
	ClearExamTokens(schedules)
	for i, s := range schedules {
		if s.ExamToken != "" {
			t.Errorf("jadwal %d masih membawa token %q", i, s.ExamToken)
		}
	}
	ClearExamTokens(nil) // tidak boleh panik
}

func TestControlScopeRedactTokens(t *testing.T) {
	f := newAccessFixture(t)
	if err := f.svc.SetScheduleProctors(f.schedX, []uuid.UUID{f.pengawas.ID}); err != nil {
		t.Fatalf("SetScheduleProctors gagal: %v", err)
	}

	tests := []struct {
		name         string
		user         domain.User
		wantX, wantY string
	}{
		{"admin menerima semua token", f.admin, "TKN", "TKN"},
		{"pengawas hanya token jadwal yang ditugaskan", f.pengawas, "TKN", ""},
		{"guru tanpa penugasan tidak menerima token", f.guruA, "", ""},
		{"pemegang questions:read_all non-admin tidak otomatis menerima token", func() domain.User {
			u := f.guruB
			u.Permissions = []string{string(domain.PermQuestionsAll)}
			return u
		}(), "", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			schedules := []domain.ExamSchedule{
				{ID: f.schedX, ExamToken: "TKN"},
				{ID: f.schedY, ExamToken: "TKN"},
			}
			f.svc.TokenScopeFor(tt.user).RedactTokens(schedules)
			if schedules[0].ExamToken != tt.wantX {
				t.Errorf("token jadwal X = %q, ingin %q", schedules[0].ExamToken, tt.wantX)
			}
			if schedules[1].ExamToken != tt.wantY {
				t.Errorf("token jadwal Y = %q, ingin %q", schedules[1].ExamToken, tt.wantY)
			}
		})
	}
}

// Respons jadwal siswa tidak boleh membawa token, sedangkan data di basis data tetap utuh
// sehingga validasi token pada StartOrResumeExam tidak terpengaruh.
func TestGetStudentSchedulesClearsExamToken(t *testing.T) {
	f := newAccessFixture(t)
	now := time.Now()

	bank := domain.QuestionBank{ID: uuid.New(), Title: "Bank siap", SubjectID: f.subjectMath, CreatedByID: f.guruA.ID, IsLocked: true, CreatedAt: now}
	if err := f.db.Create(&bank).Error; err != nil {
		t.Fatalf("gagal membuat bank: %v", err)
	}
	if err := f.db.Model(&domain.ExamSchedule{}).Where("id = ?", f.schedX).Updates(map[string]interface{}{
		"bank_id":    bank.ID,
		"start_time": now.Add(-time.Minute),
		"end_time":   now.Add(time.Hour),
	}).Error; err != nil {
		t.Fatalf("gagal memperbarui jadwal: %v", err)
	}
	profile := domain.StudentProfile{ID: uuid.New(), UserID: f.siswa.ID, ClassRoomID: f.classX, NISN: "0001"}
	if err := f.db.Create(&profile).Error; err != nil {
		t.Fatalf("gagal membuat profil siswa: %v", err)
	}

	svc := NewExamService(&repository.Database{DB: f.db})
	got, err := svc.GetStudentSchedules(f.siswa.ID)
	if err != nil {
		t.Fatalf("GetStudentSchedules gagal: %v", err)
	}
	if len(got) != 1 || got[0].ID != f.schedX {
		t.Fatalf("jadwal siswa = %d item, ingin tepat jadwal X", len(got))
	}
	if got[0].ExamToken != "" {
		t.Fatalf("exam_token bocor ke siswa: %q", got[0].ExamToken)
	}

	var stored domain.ExamSchedule
	if err := f.db.First(&stored, "id = ?", f.schedX).Error; err != nil {
		t.Fatalf("gagal membaca jadwal: %v", err)
	}
	if stored.ExamToken != "TKN" {
		t.Fatalf("token di basis data berubah: %q", stored.ExamToken)
	}
}
