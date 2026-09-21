package pdf

import (
	"testing"
)

func TestGenerateBeritaAcara(t *testing.T) {
	data := BeritaAcaraData{
		SchoolName:    "SMA NEGERI TEST",
		ExamTitle:     "Ujian Simulasi CBT",
		SubjectName:   "Matematika",
		ClassName:     "XII MIPA 1",
		RoomName:      "Ruang 01",
		DateStr:       "Senin, 01 Januari 2026",
		TimeStr:       "08:00 - 09:30 WIB",
		TotalStudents: 36,
		PresentCount:  35,
		AbsentCount:   1,
		ViolationNote: "Nihil",
		ProctorName:   "Proktor Test",
		Supervisor1:   "Pengawas 1",
		Supervisor2:   "Pengawas 2",
	}

	pdfBytes, err := GenerateBeritaAcara(data)
	if err != nil {
		t.Fatalf("Failed to generate Berita Acara PDF: %v", err)
	}
	if len(pdfBytes) < 100 {
		t.Fatalf("PDF output is too small (%d bytes)", len(pdfBytes))
	}
}
