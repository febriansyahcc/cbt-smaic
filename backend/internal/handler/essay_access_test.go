package handler

import (
	"net/http"
	"testing"
	"time"

	"cbt-backend/internal/domain"
	"cbt-backend/internal/repository"

	"github.com/google/uuid"
)

// Koreksi essay hanya untuk administrator dan guru pengampu pasangan kelas + mapel jadwal.
// Pengawas dan pembuat bank soal yang tidak mengampu kelas itu ditolak.
func TestEssayGradingScope(t *testing.T) {
	env := newStaffEnv(t)
	h := NewHandlers(&repository.Database{DB: env.db})
	env.app.Get("/essay/:id", h.HandleGetEssayAnswers)
	env.app.Patch("/essay/:id", h.HandleGradeEssayAnswers)

	guruPerms := domain.TemplatePermissions("guru")
	admin := env.mkUser(t, "admin", domain.RoleAdmin, nil)
	guruA := env.mkUser(t, "guruA", domain.RoleGuru, guruPerms) // pengampu X-1, pembuat bank
	guruB := env.mkUser(t, "guruB", domain.RoleGuru, guruPerms) // pengampu X-2
	pengawas := env.mkUser(t, "pengawas", domain.RoleGuru, guruPerms)
	siswa := env.mkUser(t, "siswa", domain.RoleSiswa, nil)

	subj := domain.Subject{ID: uuid.New(), Code: "MTK", Name: "Matematika", CreatedAt: time.Now()}
	env.db.Create(&subj)
	x1 := env.mkClass(t, "X-1", "X")
	x2 := env.mkClass(t, "X-2", "X")
	env.db.Create(&domain.ClassSubject{ID: uuid.New(), ClassRoomID: x1.ID, SubjectID: subj.ID, TeacherID: guruA.ID, AcademicYear: "2026/2027"})
	env.db.Create(&domain.ClassSubject{ID: uuid.New(), ClassRoomID: x2.ID, SubjectID: subj.ID, TeacherID: guruB.ID, AcademicYear: "2026/2027"})

	bank := domain.QuestionBank{ID: uuid.New(), Title: "MTK X", SubjectID: subj.ID, CreatedByID: guruA.ID, IsLocked: true, CreatedAt: time.Now()}
	env.db.Create(&bank)
	q := domain.Question{ID: uuid.New(), BankID: bank.ID, QuestionNumber: 1, Type: domain.TypeEssay, ContentHTML: "Jelaskan", OptionsJSON: "[]", ScoreWeight: 5}
	env.db.Create(&q)

	mkSchedule := func(class domain.ClassRoom) (domain.ExamSchedule, domain.StudentAnswer) {
		sch := domain.ExamSchedule{ID: uuid.New(), Title: "PAS MTK " + class.Name, SubjectID: &subj.ID, BankID: &bank.ID, ClassRoomID: class.ID,
			StartTime: time.Now().Add(-time.Hour), EndTime: time.Now().Add(time.Hour), DurationMinutes: 60, MaxViolations: 3, IsActive: true, CreatedAt: time.Now()}
		if err := env.db.Create(&sch).Error; err != nil {
			t.Fatalf("buat jadwal: %v", err)
		}
		sess := domain.ExamSession{ID: uuid.New(), ScheduleID: sch.ID, StudentID: siswa.ID, Status: domain.StatusSubmitted, ServerDeadline: time.Now()}
		if err := env.db.Create(&sess).Error; err != nil {
			t.Fatalf("buat sesi: %v", err)
		}
		ans := domain.StudentAnswer{ID: uuid.New(), SessionID: sess.ID, QuestionID: q.ID, AnswerText: "jawaban", LastUpdatedAt: time.Now()}
		env.db.Create(&ans)
		return sch, ans
	}
	schX1, ansX1 := mkSchedule(x1)
	schX2, _ := mkSchedule(x2)
	env.db.Create(&domain.ScheduleProctor{ScheduleID: schX1.ID, UserID: pengawas.ID})

	cases := []struct {
		name     string
		user     domain.User
		schedule domain.ExamSchedule
		want     int
	}{
		{"admin boleh", admin, schX1, http.StatusOK},
		{"pengampu X-1 boleh", guruA, schX1, http.StatusOK},
		{"pembuat bank tidak mengampu X-2 ditolak", guruA, schX2, http.StatusForbidden},
		{"pengampu X-2 boleh", guruB, schX2, http.StatusOK},
		{"pengampu X-2 tidak boleh X-1", guruB, schX1, http.StatusForbidden},
		{"pengawas X-1 ditolak", pengawas, schX1, http.StatusForbidden},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			u := tc.user
			env.caller = &u
			if code, _ := env.do(t, http.MethodGet, "/essay/"+tc.schedule.ID.String(), ""); code != tc.want {
				t.Fatalf("GET status %d, ingin %d", code, tc.want)
			}
			if code, _ := env.do(t, http.MethodPatch, "/essay/"+tc.schedule.ID.String(), `[]`); code != tc.want {
				t.Fatalf("PATCH status %d, ingin %d", code, tc.want)
			}
		})
	}

	t.Run("nilai melebihi bobot ditolak", func(t *testing.T) {
		env.caller = &guruA
		body := `[{"answer_id":"` + ansX1.ID.String() + `","score_awarded":6}]`
		if code, _ := env.do(t, http.MethodPatch, "/essay/"+schX1.ID.String(), body); code != http.StatusBadRequest {
			t.Fatalf("status %d, ingin 400", code)
		}
		body = `[{"answer_id":"` + ansX1.ID.String() + `","score_awarded":5}]`
		if code, out := env.do(t, http.MethodPatch, "/essay/"+schX1.ID.String(), body); code != http.StatusOK {
			t.Fatalf("nilai sama dengan bobot harus diterima: %d %v", code, out)
		}
	})
}
