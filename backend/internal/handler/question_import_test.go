package handler

import (
	"bytes"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"cbt-backend/internal/domain"
	"cbt-backend/internal/repository"
	"cbt-backend/pkg/excel"

	"github.com/google/uuid"
)

// Impor Excel tunduk pada kunci naskah seperti tambah soal manual, dan jumlah soal bank
// dihitung dari seluruh butir soal, bukan hanya hasil impor terakhir.
func TestImportQuestionsRespectsLockAndCount(t *testing.T) {
	env := newStaffEnv(t)
	admin := env.mkUser(t, "admin", domain.RoleAdmin, []string{string(domain.PermAll)})
	env.caller = &admin
	h := NewHandlers(&repository.Database{DB: env.db})
	env.app.Post("/import", h.HandleImportQuestionsExcel)

	tpl, err := excel.GenerateQuestionTemplate()
	if err != nil {
		t.Fatalf("template: %v", err)
	}
	var xlsx bytes.Buffer
	if err := tpl.Write(&xlsx); err != nil {
		t.Fatalf("tulis template: %v", err)
	}
	parsed, err := excel.ParseQuestionsFromExcel(bytes.NewReader(xlsx.Bytes()))
	if err != nil || len(parsed) == 0 {
		t.Fatalf("template harus berisi contoh soal: %v", err)
	}

	subj := domain.Subject{ID: uuid.New(), Code: "MTK", Name: "Matematika", CreatedAt: time.Now()}
	env.db.Create(&subj)
	mkBank := func(locked bool, existing int) domain.QuestionBank {
		b := domain.QuestionBank{ID: uuid.New(), Title: "Bank", SubjectID: subj.ID, CreatedByID: admin.ID, IsLocked: locked, TotalQuestions: existing, CreatedAt: time.Now()}
		env.db.Create(&b)
		env.db.Model(&b).Update("is_locked", locked)
		for i := 1; i <= existing; i++ {
			env.db.Create(&domain.Question{ID: uuid.New(), BankID: b.ID, QuestionNumber: i, Type: domain.TypeMultipleChoice, ContentHTML: "lama", OptionsJSON: "[]"})
		}
		return b
	}
	upload := func(bankID uuid.UUID) int {
		var body bytes.Buffer
		w := multipart.NewWriter(&body)
		_ = w.WriteField("bank_id", bankID.String())
		fw, _ := w.CreateFormFile("file", "soal.xlsx")
		_, _ = fw.Write(xlsx.Bytes())
		_ = w.Close()
		req := httptest.NewRequest(http.MethodPost, "/import", &body)
		req.Header.Set("Content-Type", w.FormDataContentType())
		resp, err := env.app.Test(req, -1)
		if err != nil {
			t.Fatalf("request gagal: %v", err)
		}
		resp.Body.Close()
		return resp.StatusCode
	}
	countOf := func(bankID uuid.UUID) (int64, int) {
		var n int64
		env.db.Model(&domain.Question{}).Where("bank_id = ?", bankID).Count(&n)
		var b domain.QuestionBank
		env.db.First(&b, "id = ?", bankID)
		return n, b.TotalQuestions
	}

	locked := mkBank(true, 2)
	if code := upload(locked.ID); code != http.StatusBadRequest {
		t.Fatalf("impor ke bank terkunci: status %d, ingin 400", code)
	}
	if n, _ := countOf(locked.ID); n != 2 {
		t.Fatalf("bank terkunci tidak boleh bertambah soal, sekarang %d", n)
	}

	open := mkBank(false, 2)
	if code := upload(open.ID); code != http.StatusOK {
		t.Fatalf("impor ke bank terbuka: status %d", code)
	}
	n, total := countOf(open.ID)
	want := int64(2 + len(parsed))
	if n != want || int64(total) != want {
		t.Fatalf("jumlah soal = %d, total_questions = %d; ingin keduanya %d", n, total, want)
	}
}
