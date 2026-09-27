package service

import (
	"sync"
	"testing"
	"time"

	"cbt-backend/internal/domain"

	"github.com/google/uuid"
)

// linkBankWithQuestion menautkan bank terkunci ke jadwal X berisi satu soal PG (kunci A)
// dan menyimpan jawaban siswa A pada sesi, sehingga nilai akhir yang benar adalah 100.
func linkBankWithQuestion(t *testing.T, f *accessFixture, sessionID uuid.UUID) {
	t.Helper()
	if err := f.db.Model(&domain.QuestionBank{}).Where("id = ?", f.bankMathByA).Update("is_locked", true).Error; err != nil {
		t.Fatal(err)
	}
	if err := f.db.Model(&domain.ExamSchedule{}).Where("id = ?", f.schedX).Update("bank_id", f.bankMathByA).Error; err != nil {
		t.Fatal(err)
	}
	q := domain.Question{ID: uuid.New(), BankID: f.bankMathByA, QuestionNumber: 1, Type: domain.TypeMultipleChoice,
		ContentHTML: "1+1?", OptionsJSON: `[{"key":"A","text":"2"},{"key":"B","text":"3"}]`, CorrectKey: "A", ScoreWeight: 1, CreatedAt: time.Now()}
	if err := f.db.Create(&q).Error; err != nil {
		t.Fatal(err)
	}
	ans := domain.StudentAnswer{ID: uuid.New(), SessionID: sessionID, QuestionID: q.ID, SelectedOption: "A", LastUpdatedAt: time.Now()}
	if err := f.db.Create(&ans).Error; err != nil {
		t.Fatal(err)
	}
}

func TestConcurrentViolationsAreAllCounted(t *testing.T) {
	f, exam, _, id := newSessionFixture(t, domain.StatusInProgress, 0)
	f.db.Model(&domain.ExamSchedule{}).Where("id = ?", f.schedX).Update("max_violations", 100)

	const n = 20
	var wg sync.WaitGroup
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if _, _, err := exam.RecordViolation(id, f.siswa.ID, "TAB_SWITCH", ""); err != nil {
				t.Error(err)
			}
		}()
	}
	wg.Wait()

	if s := loadSession(t, f, id); s.ViolationCount != n {
		t.Fatalf("violation_count = %d setelah %d laporan bersamaan", s.ViolationCount, n)
	}
}

func TestViolationDoesNotReopenSubmittedSession(t *testing.T) {
	f, exam, _, id := newSessionFixture(t, domain.StatusSubmitted, 0)

	if _, blocked, err := exam.RecordViolation(id, f.siswa.ID, "TAB_SWITCH", ""); err != nil || blocked {
		t.Fatalf("blocked %v, err %v", blocked, err)
	}
	if s := loadSession(t, f, id); s.Status != domain.StatusSubmitted || s.ViolationCount != 0 {
		t.Fatalf("status %s, pelanggaran %d; sesi yang sudah dikumpulkan tidak boleh berubah", s.Status, s.ViolationCount)
	}
}

func TestConcurrentSubmitsGradeOnce(t *testing.T) {
	f, exam, _, id := newSessionFixture(t, domain.StatusInProgress, 0)
	linkBankWithQuestion(t, f, id)

	const n = 8
	scores := make([]float64, n)
	var wg sync.WaitGroup
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			score, err := exam.SubmitExam(id, f.siswa.ID)
			if err != nil {
				t.Error(err)
			}
			scores[i] = score
		}(i)
	}
	wg.Wait()

	for i, sc := range scores {
		if sc != 100 {
			t.Fatalf("submit ke-%d mengembalikan nilai %v, ingin 100", i, sc)
		}
	}
	s := loadSession(t, f, id)
	if s.Status != domain.StatusSubmitted || s.TotalScore != 100 || s.SubmittedAt == nil {
		t.Fatalf("sesi: status %s, nilai %v, submitted_at %v", s.Status, s.TotalScore, s.SubmittedAt)
	}
}

func TestResumeAfterDeadlineGradesAnswers(t *testing.T) {
	f, exam, _, id := newSessionFixture(t, domain.StatusInProgress, 0)
	linkBankWithQuestion(t, f, id)
	f.db.Model(&domain.ExamSchedule{}).Where("id = ?", f.schedX).Update("start_time", time.Now().Add(-time.Hour))
	f.db.Model(&domain.ExamSession{}).Where("id = ?", id).Update("server_deadline", time.Now().Add(-time.Minute))

	if _, err := exam.StartOrResumeExam(f.siswa.ID, f.schedX, "TKN", "127.0.0.1", "test"); err == nil {
		t.Fatal("melanjutkan ujian setelah deadline harus ditolak")
	}
	s := loadSession(t, f, id)
	if s.Status != domain.StatusSubmitted || s.TotalScore != 100 {
		t.Fatalf("sesi: status %s, nilai %v; jawaban tersinkron harus dinilai (ingin 100)", s.Status, s.TotalScore)
	}
}

func TestExtendTimeKeepsConcurrentViolationCount(t *testing.T) {
	f, exam, proctor, id := newSessionFixture(t, domain.StatusInProgress, 0)
	exam.RecordViolation(id, f.siswa.ID, "TAB_SWITCH", "")

	if err := proctor.ExtendTimeSession(id, 5, ""); err != nil {
		t.Fatal(err)
	}
	if s := loadSession(t, f, id); s.ViolationCount != 1 {
		t.Fatalf("violation_count = %d setelah tambah waktu, ingin 1", s.ViolationCount)
	}
}

func TestSyncHeartbeatReturnsExtendedDeadline(t *testing.T) {
	f, exam, proctor, id := newSessionFixture(t, domain.StatusInProgress, 0)
	before := loadSession(t, f, id).ServerDeadline

	if err := proctor.ExtendTimeSession(id, 10, ""); err != nil {
		t.Fatal(err)
	}
	n, deadline, err := exam.SyncAnswers(id, f.siswa.ID, nil)
	if err != nil || n != 0 {
		t.Fatalf("heartbeat tanpa jawaban = %d, %v", n, err)
	}
	if got := deadline.Sub(before); got < 9*time.Minute || got > 11*time.Minute {
		t.Fatalf("deadline dari sync bergeser %v, ingin sekitar +10 menit", got)
	}
}

func TestSyncIgnoresQuestionsOutsideScheduleBank(t *testing.T) {
	f, exam, _, id := newSessionFixture(t, domain.StatusInProgress, 0)
	linkBankWithQuestion(t, f, id)
	var q domain.Question
	f.db.First(&q, "bank_id = ?", f.bankMathByA)

	foreign := domain.Question{ID: uuid.New(), BankID: f.bankBioByB, QuestionNumber: 1, ContentHTML: "x", CorrectKey: "A", CreatedAt: time.Now()}
	f.db.Create(&foreign)

	n, _, err := exam.SyncAnswers(id, f.siswa.ID, []SyncAnswerItem{
		{QuestionID: q.ID, SelectedOption: "B"},
		{QuestionID: foreign.ID, SelectedOption: "A"},
		{QuestionID: uuid.New(), SelectedOption: "A"},
	})
	if err != nil || n != 1 {
		t.Fatalf("sync = %d, %v; ingin hanya 1 jawaban dari bank jadwal", n, err)
	}
	var rows int64
	f.db.Model(&domain.StudentAnswer{}).Where("session_id = ?", id).Count(&rows)
	if rows != 1 {
		t.Fatalf("baris jawaban = %d, ingin 1", rows)
	}
}

func TestBlockedStudentCannotSubmitButProctorCan(t *testing.T) {
	f, exam, proctor, id := newSessionFixture(t, domain.StatusBlocked, 3)
	linkBankWithQuestion(t, f, id)

	if _, err := exam.SubmitExam(id, f.siswa.ID); err == nil {
		t.Fatal("siswa yang terkunci tidak boleh mengumpulkan sendiri")
	}
	if s := loadSession(t, f, id); s.Status != domain.StatusBlocked {
		t.Fatalf("status berubah menjadi %s", s.Status)
	}

	score, err := proctor.ForceSubmitSession(id)
	if err != nil || score != 100 {
		t.Fatalf("force submit pengawas = %v, %v; ingin 100", score, err)
	}
	if s := loadSession(t, f, id); s.Status != domain.StatusSubmitted {
		t.Fatalf("status setelah force submit = %s", s.Status)
	}
}
