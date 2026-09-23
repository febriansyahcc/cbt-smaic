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

	peer := mk(f.classY, start.Add(2*time.Hour), "DDDDDD", time.Hour)
	unrelated := mk(f.classX, start.Add(4*time.Hour), "EEEEEE", time.Hour)

	tokenOf := func(id uuid.UUID) string {
		var s domain.ExamSchedule
		f.db.First(&s, "id = ?", id)
		return s.ExamToken
	}

	// Centang jadwal a (sesi 07:30) dan other (sesi 09:30): satu token untuk keduanya,
	// dan jadwal lain di kedua sesi itu (b, peer) ikut seragam.
	token, n, err := RegenerateSessionTokens(f.db, []uuid.UUID{a, other})
	if err != nil || n != 2 {
		t.Fatalf("RegenerateSessionTokens = %d, %v; ingin 2 sesi", n, err)
	}
	for _, id := range []uuid.UUID{a, b, other, peer} {
		if tokenOf(id) != token {
			t.Fatalf("jadwal %s bertoken %q, ingin %q", id, tokenOf(id), token)
		}
	}
	if tokenOf(unrelated) != "EEEEEE" {
		t.Fatal("sesi yang tidak dicentang tidak boleh berubah")
	}
}
