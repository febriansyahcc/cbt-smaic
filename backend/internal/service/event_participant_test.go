package service

import (
	"strings"
	"testing"
	"time"

	"cbt-backend/internal/domain"
	"cbt-backend/internal/repository"

	"github.com/google/uuid"
)

func TestGradeDigitAndExamNumber(t *testing.T) {
	cases := map[string]string{"X": "1", "xi": "2", "XII": "3", "10": "1", "12": "3", "": "0"}
	for in, want := range cases {
		if got := GradeDigit(in); got != want {
			t.Fatalf("GradeDigit(%q) = %q, ingin %q", in, got, want)
		}
	}
	if got := FormatExamNumber("psat", "1", 1); got != "PSAT10001" {
		t.Fatalf("FormatExamNumber = %q", got)
	}
}

func TestGenerateCardPassword(t *testing.T) {
	p, err := GenerateCardPassword()
	if err != nil {
		t.Fatal(err)
	}
	if len(p) != cardPasswordLength {
		t.Fatalf("panjang kata sandi = %d", len(p))
	}
	for _, r := range p {
		if !strings.ContainsRune(cardPasswordAlphabet, r) {
			t.Fatalf("karakter di luar alfabet: %q", r)
		}
	}
}

func TestEventParticipantGenerateAndLogin(t *testing.T) {
	f := newAccessFixture(t)
	if err := f.db.AutoMigrate(&domain.EventParticipant{}); err != nil {
		t.Fatal(err)
	}
	repo := &repository.Database{DB: f.db}
	svc := NewEventParticipantService(repo)

	event := domain.ExamEvent{ID: uuid.New(), Title: "PSAT", Code: "PSAT", AcademicYear: "2025/2026", Semester: "GENAP", IsActive: true, CreatedAt: time.Now()}
	if err := f.db.Create(&event).Error; err != nil {
		t.Fatal(err)
	}
	// Siswa fixture berada di classX; hubungkan jadwal classX ke event.
	if err := f.db.Model(&domain.ExamSchedule{}).Where("id = ?", f.schedX).Update("event_id", event.ID).Error; err != nil {
		t.Fatal(err)
	}
	profile := domain.StudentProfile{ID: uuid.New(), UserID: f.siswa.ID, NIS: "1023", NISN: "991023", ClassRoomID: f.classX, CreatedAt: time.Now()}
	if err := f.db.Create(&profile).Error; err != nil {
		t.Fatal(err)
	}

	created, err := svc.Generate(event.ID, false)
	if err != nil || created == 0 {
		t.Fatalf("Generate = %d, %v", created, err)
	}
	again, err := svc.Generate(event.ID, false)
	if err != nil || again != 0 {
		t.Fatalf("Generate kedua harus idempoten, dapat %d, %v", again, err)
	}

	cards, err := svc.List(event.ID, nil)
	if err != nil || len(cards) == 0 {
		t.Fatalf("List = %v, %v", cards, err)
	}
	var card ParticipantCard
	for _, c := range cards {
		if c.UserID == f.siswa.ID {
			card = c
		}
	}
	if !strings.HasPrefix(card.ExamNumber, "PSAT") || card.Password == "" {
		t.Fatalf("kartu siswa tidak valid: %+v", card)
	}

	if u, ok := findParticipantLogin(f.db, strings.ToLower(card.ExamNumber), card.Password); !ok || u.ID != f.siswa.ID {
		t.Fatal("login nomor ujian seharusnya berhasil saat event aktif")
	}
	if _, ok := findParticipantLogin(f.db, card.ExamNumber, "SALAH1"); ok {
		t.Fatal("kata sandi salah tidak boleh diterima")
	}
	f.db.Model(&event).Update("is_active", false)
	if _, ok := findParticipantLogin(f.db, card.ExamNumber, card.Password); ok {
		t.Fatal("login nomor ujian tidak boleh berlaku saat event nonaktif")
	}
}
