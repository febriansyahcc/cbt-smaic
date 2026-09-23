package service

import (
	"testing"
	"time"

	"cbt-backend/internal/domain"

	"github.com/google/uuid"
)

func TestSessionTokenSharedPerSlot(t *testing.T) {
	f := newAccessFixture(t)
	eventID := uuid.New()
	if err := f.db.Create(&domain.ExamEvent{ID: eventID, Title: "PSAT", Code: "PSATX", AcademicYear: "2025/2026", Semester: "GENAP", CreatedAt: time.Now()}).Error; err != nil {
		t.Fatal(err)
	}
	start := time.Date(2026, 10, 5, 7, 30, 0, 0, time.Local)
	mk := func(class uuid.UUID, st time.Time, token string, age time.Duration) uuid.UUID {
		s := domain.ExamSchedule{
			ID: uuid.New(), EventID: &eventID, Title: "Sesi", ClassRoomID: class, ExamToken: token,
			StartTime: st, EndTime: st.Add(time.Hour), DurationMinutes: 60, IsActive: true, CreatedAt: time.Now().Add(-age),
		}
		if err := f.db.Create(&s).Error; err != nil {
			t.Fatal(err)
		}
		return s.ID
	}
	a := mk(f.classX, start, "AAAAAA", 2*time.Hour)
	b := mk(f.classY, start, "BBBBBB", time.Hour)
	other := mk(f.classX, start.Add(2*time.Hour), "CCCCCC", time.Hour)

	if tok, ok := SessionToken(f.db, &eventID, start, uuid.Nil); !ok || tok != "AAAAAA" {
		t.Fatalf("token sesi = %q, %v; ingin token jadwal tertua", tok, ok)
	}
	if _, ok := SessionToken(f.db, &eventID, start.Add(time.Hour), uuid.Nil); ok {
		t.Fatal("sesi kosong tidak boleh punya token")
	}

	n, err := RegenerateSessionTokens(f.db, []uuid.UUID{a, b})
	if err != nil || n != 1 {
		t.Fatalf("RegenerateSessionTokens = %d, %v; ingin 1 sesi", n, err)
	}
	tokenOf := func(id uuid.UUID) string {
		var s domain.ExamSchedule
		f.db.First(&s, "id = ?", id)
		return s.ExamToken
	}
	if tokenOf(a) != tokenOf(b) || tokenOf(a) == "AAAAAA" {
		t.Fatalf("token sesi tidak seragam: %q vs %q", tokenOf(a), tokenOf(b))
	}
	if tokenOf(other) != "CCCCCC" {
		t.Fatal("sesi lain tidak boleh ikut berubah")
	}
}
