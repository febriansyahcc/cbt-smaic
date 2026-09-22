package excel

import (
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"cbt-backend/internal/domain"

	"github.com/google/uuid"
	"github.com/xuri/excelize/v2"
)

type ParsedQuestion struct {
	QuestionNumber int
	Type           domain.QuestionType
	ContentHTML    string
	Options        []domain.OptionItem
	CorrectKey     string
	RubricGuide    string
	ScoreWeight    float64
}

// ParseQuestionsFromExcel reads an uploaded Excel file and extracts question items
func ParseQuestionsFromExcel(r io.Reader) ([]ParsedQuestion, error) {
	f, err := excelize.OpenReader(r)
	if err != nil {
		return nil, fmt.Errorf("gagal membaca file excel: %w", err)
	}
	defer f.Close()

	sheetList := f.GetSheetList()
	if len(sheetList) == 0 {
		return nil, fmt.Errorf("file excel tidak memiliki sheet aktif")
	}

	sheetName := sheetList[0]
	rows, err := f.GetRows(sheetName)
	if err != nil {
		return nil, fmt.Errorf("gagal membaca baris sheet: %w", err)
	}

	if len(rows) < 2 {
		return nil, fmt.Errorf("sheet harus memiliki minimal 1 baris header dan 1 baris data")
	}

	// Detect if header includes Type column
	headerRow := rows[0]
	isNewFormat := false
	if len(headerRow) > 1 {
		firstColUpper := strings.ToUpper(strings.TrimSpace(headerRow[1]))
		if strings.Contains(firstColUpper, "TIPE") || strings.Contains(firstColUpper, "TYPE") || strings.Contains(firstColUpper, "JENIS") {
			isNewFormat = true
		}
	}

	var results []ParsedQuestion
	for i := 1; i < len(rows); i++ {
		row := rows[i]
		if len(row) < 2 {
			continue // skip empty rows
		}

		qNum := i
		if len(row) > 0 && row[0] != "" {
			if n, err := strconv.Atoi(strings.TrimSpace(row[0])); err == nil {
				qNum = n
			}
		}

		var qType domain.QuestionType = domain.TypeMultipleChoice
		var soal string
		var optA, optB, optC, optD, optE string
		var kunci string
		var rubrik string
		var bobot float64 = 1.0

		if isNewFormat {
			// New format: [0: No, 1: Tipe, 2: Soal, 3: Opsi A, 4: Opsi B, 5: Opsi C, 6: Opsi D, 7: Opsi E, 8: Kunci, 9: Rubrik, 10: Bobot]
			typeStr := strings.ToUpper(strings.TrimSpace(row[1]))
			if strings.Contains(typeStr, "ISIAN") || strings.Contains(typeStr, "SHORT") || strings.Contains(typeStr, "SINGKAT") {
				qType = domain.TypeShortAnswer
			} else if strings.Contains(typeStr, "ESSAY") || strings.Contains(typeStr, "URAIAN") {
				qType = domain.TypeEssay
			} else {
				qType = domain.TypeMultipleChoice
			}

			if len(row) > 2 {
				soal = strings.TrimSpace(row[2])
			}
			if len(row) > 3 {
				optA = strings.TrimSpace(row[3])
			}
			if len(row) > 4 {
				optB = strings.TrimSpace(row[4])
			}
			if len(row) > 5 {
				optC = strings.TrimSpace(row[5])
			}
			if len(row) > 6 {
				optD = strings.TrimSpace(row[6])
			}
			if len(row) > 7 {
				optE = strings.TrimSpace(row[7])
			}
			if len(row) > 8 {
				kunci = strings.TrimSpace(row[8])
			}
			if len(row) > 9 {
				rubrik = strings.TrimSpace(row[9])
			}
			if len(row) > 10 && strings.TrimSpace(row[10]) != "" {
				if b, err := strconv.ParseFloat(strings.TrimSpace(row[10]), 64); err == nil && b > 0 {
					bobot = b
				}
			}
		} else {
			// Legacy format: [0: No, 1: Soal, 2: Opsi A, 3: Opsi B, 4: Opsi C, 5: Opsi D, 6: Opsi E, 7: Kunci, 8: Bobot]
			soal = strings.TrimSpace(row[1])
			if len(row) > 2 {
				optA = strings.TrimSpace(row[2])
			}
			if len(row) > 3 {
				optB = strings.TrimSpace(row[3])
			}
			if len(row) > 4 {
				optC = strings.TrimSpace(row[4])
			}
			if len(row) > 5 {
				optD = strings.TrimSpace(row[5])
			}
			if len(row) > 6 {
				optE = strings.TrimSpace(row[6])
			}
			if len(row) > 7 {
				kunci = strings.TrimSpace(row[7])
			}
			if len(row) > 8 && strings.TrimSpace(row[8]) != "" {
				if b, err := strconv.ParseFloat(strings.TrimSpace(row[8]), 64); err == nil && b > 0 {
					bobot = b
				}
			}
		}

		if soal == "" {
			continue
		}

		if qType == domain.TypeMultipleChoice && kunci == "" {
			kunci = "A"
		}
		if qType == domain.TypeMultipleChoice {
			kunci = strings.ToUpper(kunci)
		}

		var options []domain.OptionItem
		if qType == domain.TypeMultipleChoice {
			if optA != "" {
				options = append(options, domain.OptionItem{Key: "A", Text: optA})
			}
			if optB != "" {
				options = append(options, domain.OptionItem{Key: "B", Text: optB})
			}
			if optC != "" {
				options = append(options, domain.OptionItem{Key: "C", Text: optC})
			}
			if optD != "" {
				options = append(options, domain.OptionItem{Key: "D", Text: optD})
			}
			if optE != "" {
				options = append(options, domain.OptionItem{Key: "E", Text: optE})
			}
		}

		results = append(results, ParsedQuestion{
			QuestionNumber: qNum,
			Type:           qType,
			ContentHTML:    soal,
			Options:        options,
			CorrectKey:     kunci,
			RubricGuide:    rubrik,
			ScoreWeight:    bobot,
		})
	}

	if len(results) == 0 {
		return nil, fmt.Errorf("tidak ada butir soal valid yang ditemukan dalam file excel")
	}

	return results, nil
}

// GenerateQuestionTemplate creates a blank/sample template file for teachers to download
func GenerateQuestionTemplate() (*excelize.File, error) {
	f := excelize.NewFile()
	sheet := "Bank Soal"
	f.SetSheetName("Sheet1", sheet)

	headers := []string{"No", "Tipe Soal (PG / ISIAN / ESSAY)", "Pertanyaan (Soal)", "Opsi A", "Opsi B", "Opsi C", "Opsi D", "Opsi E", "Kunci / Jawaban Benar", "Pedoman Rubrik (Essay)", "Bobot"}
	for colIdx, h := range headers {
		cell, _ := excelize.CoordinatesToCellName(colIdx+1, 1)
		f.SetCellValue(sheet, cell, h)
	}

	// Sample Rows showcasing PG, Isian Singkat, and Essay
	samples := [][]interface{}{
		{1, "PG", "Sebuah mobil bergerak dengan kelajuan 72 km/jam. Nilai ini setara dengan...", "10 m/s", "15 m/s", "20 m/s", "25 m/s", "30 m/s", "C", "", 1.0},
		{2, "ISIAN", "Ibukota negara Indonesia yang berada di pulau Kalimantan adalah...", "", "", "", "", "", "Nusantara|IKN|Ibu Kota Nusantara", "", 2.0},
		{3, "ESSAY", "Jelaskan proses terjadinya fotosintesis pada tumbuhan hijau beserta reaksi kimianya secara singkat!", "", "", "", "", "", "", "Jawaban harus memuat: klorofil, cahaya matahari, air (H2O), karbondioksida (CO2), menghasilkan glukosa (C6H12O6) dan oksigen (O2).", 5.0},
	}

	for rIdx, sample := range samples {
		for cIdx, val := range sample {
			cell, _ := excelize.CoordinatesToCellName(cIdx+1, rIdx+2)
			f.SetCellValue(sheet, cell, val)
		}
	}

	// Style headers
	style, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Color: "#FFFFFF"},
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"#4F46E5"}, Pattern: 1},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
	})
	f.SetRowStyle(sheet, 1, 1, style)
	f.SetColWidth(sheet, "B", "B", 18)
	f.SetColWidth(sheet, "C", "C", 40)
	f.SetColWidth(sheet, "D", "H", 20)
	f.SetColWidth(sheet, "I", "I", 25)
	f.SetColWidth(sheet, "J", "J", 35)
	f.SetColWidth(sheet, "K", "K", 10)

	return f, nil
}

type GradeReportItem struct {
	No          int
	NIS         string
	NISN        string
	StudentName string
	ClassName   string
	StartedAt   time.Time
	SubmittedAt *time.Time
	Violations  int
	TotalScore  float64
	Status      string
}

// GenerateGradeReport creates a styled Excel report of exam scores
func GenerateGradeReport(examTitle, className string, items []GradeReportItem) (*excelize.File, error) {
	f := excelize.NewFile()
	sheet := "Rekap Nilai"
	f.SetSheetName("Sheet1", sheet)

	// Title Banner
	f.SetCellValue(sheet, "A1", fmt.Sprintf("REKAPITULASI NILAI UJIAN: %s", strings.ToUpper(examTitle)))
	f.SetCellValue(sheet, "A2", fmt.Sprintf("Kelas: %s | Dicetak: %s", className, time.Now().Format("02-01-2006 15:04")))

	headers := []string{"No", "NIS", "NISN", "Nama Siswa", "Kelas", "Mulai Ujian", "Selesai", "Pelanggaran", "Nilai Akhir", "Status"}
	for colIdx, h := range headers {
		cell, _ := excelize.CoordinatesToCellName(colIdx+1, 4)
		f.SetCellValue(sheet, cell, h)
	}

	for idx, item := range items {
		rowIdx := idx + 5
		f.SetCellValue(sheet, fmt.Sprintf("A%d", rowIdx), item.No)
		f.SetCellValue(sheet, fmt.Sprintf("B%d", rowIdx), item.NIS)
		f.SetCellValue(sheet, fmt.Sprintf("C%d", rowIdx), item.NISN)
		f.SetCellValue(sheet, fmt.Sprintf("D%d", rowIdx), item.StudentName)
		f.SetCellValue(sheet, fmt.Sprintf("E%d", rowIdx), item.ClassName)
		f.SetCellValue(sheet, fmt.Sprintf("F%d", rowIdx), item.StartedAt.Format("15:04:05"))
		subStr := "-"
		if item.SubmittedAt != nil {
			subStr = item.SubmittedAt.Format("15:04:05")
		}
		f.SetCellValue(sheet, fmt.Sprintf("G%d", rowIdx), subStr)
		f.SetCellValue(sheet, fmt.Sprintf("H%d", rowIdx), item.Violations)
		f.SetCellValue(sheet, fmt.Sprintf("I%d", rowIdx), item.TotalScore)
		f.SetCellValue(sheet, fmt.Sprintf("J%d", rowIdx), item.Status)
	}

	headerStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Color: "#FFFFFF"},
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"#4F46E5"}, Pattern: 1},
		Alignment: &excelize.Alignment{Horizontal: "center"},
	})
	f.SetRowStyle(sheet, 4, 4, headerStyle)
	f.SetColWidth(sheet, "A", "A", 6)
	f.SetColWidth(sheet, "B", "C", 16)
	f.SetColWidth(sheet, "D", "D", 28)
	f.SetColWidth(sheet, "E", "E", 16)
	f.SetColWidth(sheet, "F", "G", 14)
	f.SetColWidth(sheet, "H", "J", 14)

	return f, nil
}

// ConvertParsedToQuestions converts parsed items to domain.Question slice with JSON serialization
func ConvertParsedToQuestions(bankID uuid.UUID, parsed []ParsedQuestion) ([]domain.Question, error) {
	var questions []domain.Question
	for _, p := range parsed {
		optsBytes, err := json.Marshal(p.Options)
		if err != nil {
			optsBytes = []byte("[]")
		}
		qType := p.Type
		if qType == "" {
			qType = domain.TypeMultipleChoice
		}
		questions = append(questions, domain.Question{
			ID:             uuid.New(),
			BankID:         bankID,
			QuestionNumber: p.QuestionNumber,
			Type:           qType,
			ContentHTML:    p.ContentHTML,
			OptionsJSON:    string(optsBytes),
			CorrectKey:     p.CorrectKey,
			RubricGuide:    p.RubricGuide,
			ScoreWeight:    p.ScoreWeight,
			CreatedAt:      time.Now(),
		})
	}
	return questions, nil
}

type ParsedStudent struct {
	No        int
	NIS       string
	NISN      string
	FullName  string
	ClassName string
	Gender    string
	Password  string
}

// GenerateStudentTemplate creates an Excel template for bulk importing student accounts
func GenerateStudentTemplate() (*excelize.File, error) {
	f := excelize.NewFile()
	sheet := "Data Siswa"
	f.SetSheetName("Sheet1", sheet)

	headers := []string{"No", "NIS", "NISN", "Nama Lengkap Siswa", "Nama Kelas", "Jenis Kelamin (L/P)", "Password (Opsional)"}
	for colIdx, h := range headers {
		cell, _ := excelize.CoordinatesToCellName(colIdx+1, 1)
		f.SetCellValue(sheet, cell, h)
	}

	// Sample Rows
	samples := [][]interface{}{
		{1, "1004", "0051234504", "Rian Hidayat", "XII MIPA 1", "L", "siswa123"},
		{2, "1005", "0051234505", "Dewi Lestari", "XII MIPA 1", "P", "siswa123"},
	}

	for rIdx, sample := range samples {
		for cIdx, val := range sample {
			cell, _ := excelize.CoordinatesToCellName(cIdx+1, rIdx+2)
			f.SetCellValue(sheet, cell, val)
		}
	}

	headerStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Color: "#FFFFFF"},
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"#4F46E5"}, Pattern: 1},
		Alignment: &excelize.Alignment{Horizontal: "center"},
	})
	f.SetRowStyle(sheet, 1, 1, headerStyle)
	f.SetColWidth(sheet, "D", "D", 30)
	f.SetColWidth(sheet, "B", "C", 16)
	f.SetColWidth(sheet, "E", "G", 20)

	return f, nil
}

// ============ CLASSES IMPORT ============

type ParsedClass struct {
	Name  string
	Grade string
	Major string
}

func ParseClassesFromExcel(r io.Reader) ([]ParsedClass, error) {
	f, err := excelize.OpenReader(r)
	if err != nil {
		return nil, fmt.Errorf("gagal membaca file excel: %w", err)
	}
	defer f.Close()
	rows, err := f.GetRows(f.GetSheetList()[0])
	if err != nil || len(rows) < 2 {
		return nil, fmt.Errorf("sheet harus memiliki minimal 1 baris header dan 1 baris data")
	}
	var out []ParsedClass
	for i := 1; i < len(rows); i++ {
		row := rows[i]
		if len(row) < 1 || strings.TrimSpace(row[0]) == "" {
			continue
		}
		name := strings.TrimSpace(row[0])
		grade := ""
		if len(row) > 1 {
			grade = strings.TrimSpace(row[1])
		}
		major := ""
		if len(row) > 2 {
			major = strings.TrimSpace(row[2])
		}
		// auto-detect grade from name if not provided
		if grade == "" {
			nameLow := strings.ToUpper(name)
			if strings.Contains(nameLow, "XII") {
				grade = "XII"
			} else if strings.Contains(nameLow, "XI") {
				grade = "XI"
			} else if strings.Contains(nameLow, "X") {
				grade = "X"
			}
		}
		out = append(out, ParsedClass{Name: name, Grade: grade, Major: major})
	}
	return out, nil
}

func GenerateClassesTemplate() ([]byte, error) {
	f := excelize.NewFile()
	defer f.Close()
	sheet := "Template Kelas"
	f.SetSheetName("Sheet1", sheet)
	headers := []string{"Nama Kelas", "Tingkat (X/XI/XII)", "Jurusan/Program"}
	for i, h := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(sheet, cell, h)
	}
	examples := [][]string{
		{"X MIPA 1", "X", "MIPA"},
		{"X IPS 1", "X", "IPS"},
		{"XI MIPA 1", "XI", "MIPA"},
		{"XI IPS 1", "XI", "IPS"},
		{"XII MIPA 1", "XII", "MIPA"},
		{"XII MIPA 2", "XII", "MIPA"},
		{"XII IPS 1", "XII", "IPS"},
	}
	for r, ex := range examples {
		for c, val := range ex {
			cell, _ := excelize.CoordinatesToCellName(c+1, r+2)
			f.SetCellValue(sheet, cell, val)
		}
	}
	buf, err := f.WriteToBuffer()
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// ============ SUBJECTS IMPORT ============

type ParsedSubject struct {
	Code string
	Name string
}

func ParseSubjectsFromExcel(r io.Reader) ([]ParsedSubject, error) {
	f, err := excelize.OpenReader(r)
	if err != nil {
		return nil, fmt.Errorf("gagal membaca file excel: %w", err)
	}
	defer f.Close()
	rows, err := f.GetRows(f.GetSheetList()[0])
	if err != nil || len(rows) < 2 {
		return nil, fmt.Errorf("sheet harus memiliki minimal 1 baris header dan 1 baris data")
	}
	var out []ParsedSubject
	for i := 1; i < len(rows); i++ {
		row := rows[i]
		if len(row) < 2 || strings.TrimSpace(row[0]) == "" || strings.TrimSpace(row[1]) == "" {
			continue
		}
		out = append(out, ParsedSubject{
			Code: strings.ToUpper(strings.TrimSpace(row[0])),
			Name: strings.TrimSpace(row[1]),
		})
	}
	return out, nil
}

func GenerateSubjectsTemplate() ([]byte, error) {
	f := excelize.NewFile()
	defer f.Close()
	sheet := "Template Mapel"
	f.SetSheetName("Sheet1", sheet)
	headers := []string{"Kode Mapel", "Nama Mata Pelajaran"}
	for i, h := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(sheet, cell, h)
	}
	examples := [][]string{
		{"MTK", "Matematika"},
		{"BIN", "Bahasa Indonesia"},
		{"BING", "Bahasa Inggris"},
		{"FIS", "Fisika"},
		{"KIM", "Kimia"},
		{"BIO", "Biologi"},
		{"SEJ", "Sejarah"},
		{"GEO", "Geografi"},
		{"EKO", "Ekonomi"},
		{"SOS", "Sosiologi"},
		{"PAI", "Pendidikan Agama Islam"},
		{"PKN", "Pendidikan Kewarganegaraan"},
	}
	for r, ex := range examples {
		for c, val := range ex {
			cell, _ := excelize.CoordinatesToCellName(c+1, r+2)
			f.SetCellValue(sheet, cell, val)
		}
	}
	buf, err := f.WriteToBuffer()
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// ============ TEACHERS IMPORT ============

type ParsedTeacher struct {
	FullName string
	Username string
	Password string
	Role     string // GURU or ADMIN
}

func ParseTeachersFromExcel(r io.Reader) ([]ParsedTeacher, error) {
	f, err := excelize.OpenReader(r)
	if err != nil {
		return nil, fmt.Errorf("gagal membaca file excel: %w", err)
	}
	defer f.Close()
	rows, err := f.GetRows(f.GetSheetList()[0])
	if err != nil || len(rows) < 2 {
		return nil, fmt.Errorf("sheet harus memiliki minimal 1 baris header dan 1 baris data")
	}
	var out []ParsedTeacher
	for i := 1; i < len(rows); i++ {
		row := rows[i]
		if len(row) < 2 || strings.TrimSpace(row[0]) == "" || strings.TrimSpace(row[1]) == "" {
			continue
		}
		fullName := strings.TrimSpace(row[0])
		username := strings.TrimSpace(row[1])
		password := "guru123"
		if len(row) > 2 && strings.TrimSpace(row[2]) != "" {
			password = strings.TrimSpace(row[2])
		}
		role := "GURU"
		if len(row) > 3 {
			r4 := strings.ToUpper(strings.TrimSpace(row[3]))
			if r4 == "ADMIN" || r4 == "ADMINISTRATOR" {
				role = "ADMIN"
			}
		}
		out = append(out, ParsedTeacher{FullName: fullName, Username: username, Password: password, Role: role})
	}
	return out, nil
}

func GenerateTeachersTemplate() ([]byte, error) {
	f := excelize.NewFile()
	defer f.Close()
	sheet := "Template Guru"
	f.SetSheetName("Sheet1", sheet)
	headers := []string{"Nama Lengkap & Gelar", "Username Login", "Password Awal", "Role (GURU/ADMIN)"}
	for i, h := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(sheet, cell, h)
	}
	examples := [][]string{
		{"Drs. Ahmad Fauzi, M.Pd", "afauzi", "guru123", "GURU"},
		{"Hj. Siti Rahayu, S.Pd", "srahayu", "guru123", "GURU"},
		{"Bambang Supriyadi, S.Pd.I", "bsupriyadi", "guru123", "GURU"},
	}
	for r, ex := range examples {
		for c, val := range ex {
			cell, _ := excelize.CoordinatesToCellName(c+1, r+2)
			f.SetCellValue(sheet, cell, val)
		}
	}
	buf, err := f.WriteToBuffer()
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// ParseStudentsFromExcel parses uploaded student spreadsheet
func ParseStudentsFromExcel(r io.Reader) ([]ParsedStudent, error) {
	f, err := excelize.OpenReader(r)
	if err != nil {
		return nil, fmt.Errorf("gagal membaca file excel siswa: %w", err)
	}
	defer f.Close()

	sheetList := f.GetSheetList()
	if len(sheetList) == 0 {
		return nil, fmt.Errorf("file excel tidak memiliki sheet")
	}

	rows, err := f.GetRows(sheetList[0])
	if err != nil {
		return nil, fmt.Errorf("gagal membaca baris siswa: %w", err)
	}

	if len(rows) < 2 {
		return nil, fmt.Errorf("file harus memiliki baris header dan minimal 1 baris data siswa")
	}

	var results []ParsedStudent
	for i := 1; i < len(rows); i++ {
		row := rows[i]
		if len(row) < 4 {
			continue
		}

		nis := strings.TrimSpace(row[1])
		nama := strings.TrimSpace(row[3])
		if nis == "" || nama == "" {
			continue
		}

		nisn := ""
		if len(row) > 2 {
			nisn = strings.TrimSpace(row[2])
		}

		className := ""
		if len(row) > 4 {
			className = strings.TrimSpace(row[4])
		}

		gender := "L"
		if len(row) > 5 && strings.ToUpper(strings.TrimSpace(row[5])) == "P" {
			gender = "P"
		}

		pass := "siswa123"
		if len(row) > 6 && strings.TrimSpace(row[6]) != "" {
			pass = strings.TrimSpace(row[6])
		}

		results = append(results, ParsedStudent{
			No:        i,
			NIS:       nis,
			NISN:      nisn,
			FullName:  nama,
			ClassName: className,
			Gender:    gender,
			Password:  pass,
		})
	}

	if len(results) == 0 {
		return nil, fmt.Errorf("tidak ada data siswa valid dalam file")
	}

	return results, nil
}
