package pdf

import (
	"bytes"
	"fmt"
	"time"

	"github.com/jung-kurt/gofpdf"
)

type BeritaAcaraData struct {
	SchoolName    string
	ExamTitle     string
	SubjectName   string
	ClassName     string
	RoomName      string
	DateStr       string
	TimeStr       string
	TotalStudents int
	PresentCount  int
	AbsentCount   int
	ViolationNote string
	ProctorName   string
	Supervisor1   string
	Supervisor2   string
}

// GenerateBeritaAcara creates an official PDF document for exam supervision records
func GenerateBeritaAcara(data BeritaAcaraData) ([]byte, error) {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.SetMargins(20, 15, 20)
	pdf.AddPage()

	// Kop Surat
	pdf.SetFont("Arial", "B", 14)
	pdf.CellFormat(0, 7, "PEMERINTAH PROVINSI DAERAH KHUSUS IBUKOTA", "", 1, "C", false, 0, "")
	pdf.CellFormat(0, 6, "DINAS PENDIDIKAN DAN KEBUDAYAAN", "", 1, "C", false, 0, "")
	pdf.SetFont("Arial", "B", 16)
	pdf.CellFormat(0, 8, data.SchoolName, "", 1, "C", false, 0, "")
	pdf.SetFont("Arial", "", 10)
	pdf.CellFormat(0, 5, "Jl. Pendidikan Prestasi No. 10 | Telp: (021) 555-0199 | Web: cbt.sekolah.sch.id", "", 1, "C", false, 0, "")

	// Garis batas kop surat
	pdf.SetLineWidth(0.8)
	pdf.Line(20, 42, 190, 42)
	pdf.SetLineWidth(0.3)
	pdf.Line(20, 43, 190, 43)
	pdf.Ln(8)

	// Judul Dokumen
	pdf.SetFont("Arial", "B", 13)
	pdf.CellFormat(0, 7, "BERITA ACARA PELAKSANAAN UJIAN BERBASIS KOMPUTER (CBT)", "", 1, "C", false, 0, "")
	pdf.SetFont("Arial", "", 10)
	pdf.CellFormat(0, 5, fmt.Sprintf("TAHUN AJARAN %d/%d", time.Now().Year(), time.Now().Year()+1), "", 1, "C", false, 0, "")
	pdf.Ln(6)

	// Paragraf Pembuka
	intro := fmt.Sprintf("Pada hari ini %s, bertempat di %s, telah diselenggarakan Ujian Berbasis Komputer dengan rincian operasional sebagai berikut:",
		data.DateStr, data.SchoolName)
	pdf.MultiCell(0, 5, intro, "", "J", false)
	pdf.Ln(4)

	// Identitas Pelaksanaan
	pdf.SetFont("Arial", "", 10)
	rows := [][]string{
		{"Mata Pelajaran", ": " + data.SubjectName},
		{"Jenis Ujian", ": " + data.ExamTitle},
		{"Kelas / Rombel", ": " + data.ClassName},
		{"Ruang / Sesi", ": " + data.RoomName + " (" + data.TimeStr + ")"},
		{"Jumlah Terdaftar", fmt.Sprintf(": %d Peserta", data.TotalStudents)},
		{"Jumlah Hadir", fmt.Sprintf(": %d Peserta", data.PresentCount)},
		{"Jumlah Tidak Hadir", fmt.Sprintf(": %d Peserta", data.AbsentCount)},
	}

	for _, r := range rows {
		pdf.SetFont("Arial", "B", 10)
		pdf.CellFormat(45, 6, r[0], "", 0, "L", false, 0, "")
		pdf.SetFont("Arial", "", 10)
		pdf.CellFormat(0, 6, r[1], "", 1, "L", false, 0, "")
	}
	pdf.Ln(4)

	// Catatan Kejadian / Pelanggaran
	pdf.SetFont("Arial", "B", 10)
	pdf.CellFormat(0, 6, "Catatan Khusus / Insiden Selama Ujian:", "", 1, "L", false, 0, "")
	pdf.SetFont("Arial", "", 10)
	note := data.ViolationNote
	if note == "" {
		note = "Ujian berlangsung tertib, lancar, dan tidak ada insiden pelanggaran fatal."
	}
	pdf.SetFillColor(245, 245, 245)
	pdf.Rect(20, pdf.GetY(), 170, 20, "F")
	pdf.MultiCell(170, 5, note, "", "L", false)
	pdf.Ln(12)

	// Kolom Tanda Tangan
	pdf.SetFont("Arial", "", 10)
	ySign := pdf.GetY()
	pdf.SetXY(20, ySign)
	pdf.CellFormat(75, 5, "Pengawas Ruang 1,", "", 1, "C", false, 0, "")
	pdf.Ln(18)
	pdf.SetFont("Arial", "U", 10)
	pdf.CellFormat(75, 5, data.Supervisor1, "", 1, "C", false, 0, "")
	pdf.SetFont("Arial", "", 9)
	pdf.CellFormat(75, 4, "NIP. 19850312 201001 1 012", "", 1, "C", false, 0, "")

	pdf.SetXY(115, ySign)
	pdf.SetFont("Arial", "", 10)
	pdf.CellFormat(75, 5, "Proktor Utama CBT,", "", 1, "C", false, 0, "")
	pdf.SetXY(115, ySign+23)
	pdf.SetFont("Arial", "U", 10)
	pdf.CellFormat(75, 5, data.ProctorName, "", 1, "C", false, 0, "")
	pdf.SetXY(115, ySign+28)
	pdf.SetFont("Arial", "", 9)
	pdf.CellFormat(75, 4, "NIP. 19910724 201502 2 004", "", 1, "C", false, 0, "")

	var buf bytes.Buffer
	err := pdf.Output(&buf)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
