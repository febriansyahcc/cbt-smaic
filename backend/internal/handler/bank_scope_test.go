package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"cbt-backend/internal/domain"
	"cbt-backend/internal/repository"

	"github.com/google/uuid"
)

func (e *staffEnv) mkClass(t *testing.T, name, grade string) domain.ClassRoom {
	t.Helper()
	c := domain.ClassRoom{ID: uuid.New(), Name: name, Grade: grade, CreatedAt: time.Now()}
	if err := e.db.Create(&c).Error; err != nil {
		t.Fatalf("gagal membuat kelas %s: %v", name, err)
	}
	return c
}

func (e *staffEnv) sendJSON(t *testing.T, method, path string, body any) (int, map[string]any) {
	t.Helper()
	raw, _ := json.Marshal(body)
	req := httptest.NewRequest(method, path, bytes.NewReader(raw))
	req.Header.Set("Content-Type", "application/json")
	resp, err := e.app.Test(req, -1)
	if err != nil {
		t.Fatalf("request %s %s gagal: %v", method, path, err)
	}
	var out map[string]any
	_ = json.NewDecoder(resp.Body).Decode(&out)
	return resp.StatusCode, out
}

func (e *staffEnv) bankClassIDs(t *testing.T, id uuid.UUID) (domain.QuestionBank, map[uuid.UUID]bool) {
	t.Helper()
	var b domain.QuestionBank
	if err := e.db.Preload("Classes").First(&b, "id = ?", id).Error; err != nil {
		t.Fatalf("gagal membaca bank: %v", err)
	}
	ids := make(map[uuid.UUID]bool)
	for _, c := range b.Classes {
		ids[c.ID] = true
	}
	return b, ids
}

// Bank soal default per angkatan, tetapi bisa dibatasi ke kelas tertentu lintas angkatan.
func TestQuestionBankClassScope(t *testing.T) {
	env := newStaffEnv(t)
	admin := env.mkUser(t, "admin", domain.RoleAdmin, []string{string(domain.PermQuestionsAll)})
	env.caller = &admin

	h := NewHandlers(&repository.Database{DB: env.db})
	env.app.Post("/banks", h.HandleCreateQuestionBank)
	env.app.Put("/banks/:id", h.HandleUpdateQuestionBank)
	env.app.Delete("/banks/:id", h.HandleDeleteQuestionBank)

	subj := domain.Subject{ID: uuid.New(), Code: "INF", Name: "Informatika", CreatedAt: time.Now()}
	env.db.Create(&subj)
	x1 := env.mkClass(t, "X 1", "X")
	xi2 := env.mkClass(t, "XI 2", "XI")
	xii1 := env.mkClass(t, "XII 1", "XII")

	// Default: per angkatan
	code, out := env.sendJSON(t, http.MethodPost, "/banks", map[string]any{"subject_id": subj.ID, "grade": "XI", "title": "INF XI"})
	if code != http.StatusOK {
		t.Fatalf("buat bank per angkatan: status %d %v", code, out)
	}
	gradeID := uuid.MustParse(out["data"].(map[string]any)["id"].(string))
	b, ids := env.bankClassIDs(t, gradeID)
	if b.Grade != "XI" || len(ids) != 0 {
		t.Fatalf("bank per angkatan salah: grade=%q kelas=%v", b.Grade, ids)
	}

	// Kustom: X 1 + XI 2 (grade dikosongkan otomatis)
	code, out = env.sendJSON(t, http.MethodPost, "/banks", map[string]any{"subject_id": subj.ID, "grade": "X", "class_ids": []uuid.UUID{x1.ID, xi2.ID, x1.ID}, "title": "INF X 1, XI 2"})
	if code != http.StatusOK {
		t.Fatalf("buat bank kustom: status %d %v", code, out)
	}
	customID := uuid.MustParse(out["data"].(map[string]any)["id"].(string))
	b, ids = env.bankClassIDs(t, customID)
	if b.Grade != "" || len(ids) != 2 || !ids[x1.ID] || !ids[xi2.ID] {
		t.Fatalf("bank kustom salah: grade=%q kelas=%v", b.Grade, ids)
	}

	// Ganti kelas
	code, out = env.sendJSON(t, http.MethodPut, "/banks/"+customID.String(), map[string]any{"class_ids": []uuid.UUID{xii1.ID}})
	if code != http.StatusOK {
		t.Fatalf("perbarui kelas: status %d %v", code, out)
	}
	if _, ids = env.bankClassIDs(t, customID); len(ids) != 1 || !ids[xii1.ID] {
		t.Fatalf("kelas setelah diperbarui salah: %v", ids)
	}

	// Kembali ke per angkatan: kelas dikosongkan
	code, _ = env.sendJSON(t, http.MethodPut, "/banks/"+customID.String(), map[string]any{"grade": "XII", "class_ids": []uuid.UUID{}})
	if code != http.StatusOK {
		t.Fatalf("kembali per angkatan: status %d", code)
	}
	if b, ids = env.bankClassIDs(t, customID); b.Grade != "XII" || len(ids) != 0 {
		t.Fatalf("kembali per angkatan salah: grade=%q kelas=%v", b.Grade, ids)
	}

	// Kelas tak dikenal ditolak
	code, _ = env.sendJSON(t, http.MethodPost, "/banks", map[string]any{"subject_id": subj.ID, "class_ids": []uuid.UUID{uuid.New()}, "title": "x"})
	if code != http.StatusBadRequest {
		t.Fatalf("kelas tak dikenal seharusnya 400, dapat %d", code)
	}

	// Hapus bank ikut membersihkan tautan kelas
	env.sendJSON(t, http.MethodPut, "/banks/"+customID.String(), map[string]any{"class_ids": []uuid.UUID{x1.ID}})
	if code, _ = env.sendJSON(t, http.MethodDelete, "/banks/"+customID.String(), nil); code != http.StatusOK {
		t.Fatalf("hapus bank: status %d", code)
	}
	var links int64
	env.db.Table("question_bank_classes").Where("question_bank_id = ?", customID).Count(&links)
	if links != 0 {
		t.Fatalf("tautan kelas masih tersisa: %d", links)
	}
}
