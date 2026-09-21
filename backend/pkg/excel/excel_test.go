package excel

import (
	"bytes"
	"testing"
	"time"
)

func TestExcelTemplateAndReport(t *testing.T) {
	// 1. Template Generation
	f, err := GenerateQuestionTemplate()
	if err != nil {
		t.Fatalf("Failed to generate question template: %v", err)
	}
	var buf bytes.Buffer
	if err := f.Write(&buf); err != nil {
		t.Fatalf("Failed to write template buffer: %v", err)
	}
	if buf.Len() == 0 {
		t.Fatalf("Generated template buffer is empty")
	}

	// 2. Parse from template
	parsed, err := ParseQuestionsFromExcel(&buf)
	if err != nil {
		t.Fatalf("Failed to parse back from generated template: %v", err)
	}
	if len(parsed) != 3 {
		t.Fatalf("Expected 3 sample questions, got %d", len(parsed))
	}

	// 3. Grade Report Generation
	subTime := time.Now()
	reportItems := []GradeReportItem{
		{
			No:          1,
			NIS:         "1001",
			NISN:        "0051234501",
			StudentName: "Ahmad Fauzi",
			ClassName:   "XII MIPA 1",
			StartedAt:   time.Now().Add(-1 * time.Hour),
			SubmittedAt: &subTime,
			Violations:  0,
			TotalScore:  85.0,
			Status:      "Selesai",
		},
	}
	rf, err := GenerateGradeReport("Ujian Akhir Semester", "XII MIPA 1", reportItems)
	if err != nil {
		t.Fatalf("Failed to generate grade report: %v", err)
	}
	var rBuf bytes.Buffer
	if err := rf.Write(&rBuf); err != nil {
		t.Fatalf("Failed to write report buffer: %v", err)
	}
	if rBuf.Len() == 0 {
		t.Fatalf("Report buffer is empty")
	}
}
