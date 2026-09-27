package service

import (
	"testing"
	"time"

	"cbt-backend/internal/domain"
	"cbt-backend/internal/repository"

	"github.com/google/uuid"
)

func newSessionFixture(t *testing.T, status domain.SessionStatus, violations int) (*accessFixture, *ExamService, *ProctorService, uuid.UUID) {
	t.Helper()
	f := newAccessFixture(t)
	if err := f.db.Model(&domain.ExamSchedule{}).Where("id = ?", f.schedX).Update("max_violations", 3).Error; err != nil {
		t.Fatal(err)
	}
	now := time.Now()
	sess := domain.ExamSession{
		ID: uuid.New(), ScheduleID: f.schedX, StudentID: f.siswa.ID, StartedAt: now,
		ServerDeadline: now.Add(time.Hour), Status: status, ViolationCount: violations,
		CreatedAt: now, UpdatedAt: now,
	}
	if err := f.db.Create(&sess).Error; err != nil {
		t.Fatal(err)
	}
	repo := &repository.Database{DB: f.db}
	return f, NewExamService(repo), NewProctorService(repo), sess.ID
}

func loadSession(t *testing.T, f *accessFixture, id uuid.UUID) domain.ExamSession {
	t.Helper()
	var s domain.ExamSession
	if err := f.db.First(&s, "id = ?", id).Error; err != nil {
		t.Fatal(err)
	}
	return s
}

func TestUnlockGivesFreshViolationQuota(t *testing.T) {
	f, exam, proctor, id := newSessionFixture(t, domain.StatusBlocked, 3)

	if err := proctor.UnlockStudentSession(id); err != nil {
		t.Fatalf("unlock gagal: %v", err)
	}
	s := loadSession(t, f, id)
	if s.Status != domain.StatusInProgress || s.ViolationCount != 3 || s.ActiveViolations() != 0 {
		t.Fatalf("setelah unlock: status %s, total %d, aktif %d; ingin IN_PROGRESS, 3, 0", s.Status, s.ViolationCount, s.ActiveViolations())
	}

	// Satu pelanggaran baru tidak boleh langsung mengunci lagi.
	active, blocked, err := exam.RecordViolation(id, f.siswa.ID, "TAB_SWITCH", "")
	if err != nil || blocked || active != 1 {
		t.Fatalf("pelanggaran pertama setelah unlock = %d, blocked %v, err %v; ingin 1, false", active, blocked, err)
	}
	exam.RecordViolation(id, f.siswa.ID, "TAB_SWITCH", "")
	if _, blocked, _ := exam.RecordViolation(id, f.siswa.ID, "TAB_SWITCH", ""); !blocked {
		t.Fatal("kuota penuh berikutnya harus mengunci kembali")
	}
	if s := loadSession(t, f, id); s.ViolationCount != 6 {
		t.Fatalf("total pelanggaran untuk audit = %d, ingin 6", s.ViolationCount)
	}

	var logs int64
	f.db.Model(&domain.ViolationLog{}).Where("session_id = ? AND event_type = ?", id, "UNLOCKED").Count(&logs)
	if logs != 1 {
		t.Fatalf("log UNLOCKED = %d, ingin 1", logs)
	}
}

func TestUnlockRejectsSubmittedSession(t *testing.T) {
	f, _, proctor, id := newSessionFixture(t, domain.StatusSubmitted, 1)

	if err := proctor.UnlockStudentSession(id); err == nil {
		t.Fatal("unlock sesi yang sudah dikumpulkan harus ditolak")
	}
	if s := loadSession(t, f, id); s.Status != domain.StatusSubmitted {
		t.Fatalf("status berubah menjadi %s", s.Status)
	}
}

func TestUnlockInProgressIsNoop(t *testing.T) {
	f, _, proctor, id := newSessionFixture(t, domain.StatusInProgress, 1)

	if err := proctor.UnlockStudentSession(id); err != nil {
		t.Fatalf("unlock sesi yang tidak terkunci: %v", err)
	}
	if s := loadSession(t, f, id); s.ViolationBase != 0 {
		t.Fatalf("violation_base = %d, sesi tidak terkunci tidak boleh diputihkan", s.ViolationBase)
	}
}

func TestExtendTimeReopensBlockedSessionWithFreshQuota(t *testing.T) {
	f, _, proctor, id := newSessionFixture(t, domain.StatusBlocked, 3)

	if err := proctor.ExtendTimeSession(id, 10, ""); err != nil {
		t.Fatal(err)
	}
	if s := loadSession(t, f, id); s.Status != domain.StatusInProgress || s.ActiveViolations() != 0 {
		t.Fatalf("setelah tambah waktu: status %s, aktif %d; ingin IN_PROGRESS, 0", s.Status, s.ActiveViolations())
	}
}
