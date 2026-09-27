package handler

import (
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"cbt-backend/internal/domain"
	"cbt-backend/internal/repository"

	"github.com/google/uuid"
)

// Daftar siswa harus membawa status sesi login tanpa pernah membocorkan token sesinya.
func TestGetStudentsReportsActiveSession(t *testing.T) {
	env := newStaffEnv(t)
	h := NewHandlers(&repository.Database{DB: env.db})
	env.app.Get("/students-list", h.HandleGetStudents)

	class := domain.ClassRoom{ID: uuid.New(), Name: "XII MIPA 1", Grade: "XII"}
	if err := env.db.Create(&class).Error; err != nil {
		t.Fatalf("gagal membuat kelas: %v", err)
	}
	mk := func(name, nis, sessionToken string) {
		u := domain.User{ID: uuid.New(), Username: name, PasswordHash: "x", FullName: name, Role: domain.RoleSiswa, IsActive: true, SessionToken: sessionToken, CreatedAt: time.Now()}
		if err := env.db.Create(&u).Error; err != nil {
			t.Fatalf("gagal membuat user: %v", err)
		}
		p := domain.StudentProfile{ID: uuid.New(), UserID: u.ID, NIS: nis, NISN: "00" + nis, ClassRoomID: class.ID}
		if err := env.db.Create(&p).Error; err != nil {
			t.Fatalf("gagal membuat profil: %v", err)
		}
	}
	mk("login", "1001", "sid-rahasia")
	mk("belum", "1002", "")

	resp, err := env.app.Test(httptest.NewRequest("GET", "/students-list", nil), -1)
	if err != nil {
		t.Fatalf("request gagal: %v", err)
	}
	defer resp.Body.Close()
	var raw strings.Builder
	var body struct {
		Data []struct {
			NIS              string `json:"nis"`
			HasActiveSession bool   `json:"has_active_session"`
			User             struct {
				Username string `json:"username"`
			} `json:"user"`
		} `json:"data"`
	}
	dec := json.NewDecoder(resp.Body)
	var generic json.RawMessage
	if err := dec.Decode(&generic); err != nil {
		t.Fatalf("respons bukan JSON: %v", err)
	}
	raw.Write(generic)
	if strings.Contains(raw.String(), "sid-rahasia") {
		t.Fatal("token sesi tidak boleh ikut terkirim")
	}
	if err := json.Unmarshal(generic, &body); err != nil {
		t.Fatalf("gagal membaca respons: %v", err)
	}
	got := map[string]bool{}
	for _, s := range body.Data {
		got[s.User.Username] = s.HasActiveSession
		if s.NIS == "" {
			t.Fatalf("field profil harus tetap ada: %+v", s)
		}
	}
	if !got["login"] || got["belum"] || len(got) != 2 {
		t.Fatalf("has_active_session = %v; ingin login=true, belum=false", got)
	}
}
