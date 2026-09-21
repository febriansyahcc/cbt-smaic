package repository

import (
	"encoding/json"
	"log"
	"time"

	"cbt-backend/internal/domain"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

func HashPassword(plain string) string {
	b, _ := bcrypt.GenerateFromPassword([]byte(plain), bcrypt.DefaultCost)
	return string(b)
}

func (d *Database) SeedInitialData() {
	var count int64
	d.DB.Model(&domain.User{}).Count(&count)
	if count > 0 {
		return // Data already exists
	}

	log.Println("Seeding initial database with Master, Teacher, Student, and Question Bank...")

	// 1. Users
	adminID := uuid.New()
	guruID := uuid.New()
	siswa1ID := uuid.New()
	siswa2ID := uuid.New()
	siswa3ID := uuid.New()

	users := []domain.User{
		{
			ID:           adminID,
			Username:     "admin",
			PasswordHash: HashPassword("admin123"),
			FullName:     "Administrator Kurikulum",
			Role:         domain.RoleAdmin,
			IsActive:     true,
			CreatedAt:    time.Now(),
		},
		{
			ID:           guruID,
			Username:     "guru1",
			PasswordHash: HashPassword("guru123"),
			FullName:     "Drs. Bambang Sudarsono, M.Pd",
			Role:         domain.RoleGuru,
			Permissions:  domain.TemplatePermissions("guru"),
			IsActive:     true,
			CreatedAt:    time.Now(),
		},
		{
			ID:           siswa1ID,
			Username:     "siswa1",
			PasswordHash: HashPassword("siswa123"),
			FullName:     "Ahmad Fauzi",
			Role:         domain.RoleSiswa,
			IsActive:     true,
			CreatedAt:    time.Now(),
		},
		{
			ID:           siswa2ID,
			Username:     "siswa2",
			PasswordHash: HashPassword("siswa123"),
			FullName:     "Siti Nurhaliza",
			Role:         domain.RoleSiswa,
			IsActive:     true,
			CreatedAt:    time.Now(),
		},
		{
			ID:           siswa3ID,
			Username:     "siswa3",
			PasswordHash: HashPassword("siswa123"),
			FullName:     "Budi Wicaksono",
			Role:         domain.RoleSiswa,
			IsActive:     true,
			CreatedAt:    time.Now(),
		},
	}
	d.DB.Create(&users)

	// 2. Class
	classID := uuid.New()
	class := domain.ClassRoom{
		ID:        classID,
		Name:      "XII MIPA 1",
		Grade:     "XII",
		Major:     "MIPA",
		CreatedAt: time.Now(),
	}
	d.DB.Create(&class)

	// 3. Student Profiles
	students := []domain.StudentProfile{
		{
			ID:          uuid.New(),
			UserID:      siswa1ID,
			NISN:        "0051234501",
			NIS:         "1001",
			ClassRoomID: classID,
			Gender:      "L",
			CreatedAt:   time.Now(),
		},
		{
			ID:          uuid.New(),
			UserID:      siswa2ID,
			NISN:        "0051234502",
			NIS:         "1002",
			ClassRoomID: classID,
			Gender:      "P",
			CreatedAt:   time.Now(),
		},
		{
			ID:          uuid.New(),
			UserID:      siswa3ID,
			NISN:        "0051234503",
			NIS:         "1003",
			ClassRoomID: classID,
			Gender:      "L",
			CreatedAt:   time.Now(),
		},
	}
	d.DB.Create(&students)

	// 4. Subject
	subjectID := uuid.New()
	subject := domain.Subject{
		ID:        subjectID,
		Code:      "MAT-WJB-XII",
		Name:      "Matematika Wajib",
		CreatedAt: time.Now(),
	}
	d.DB.Create(&subject)

	// 5. Question Bank
	bankID := uuid.New()
	bank := domain.QuestionBank{
		ID:             bankID,
		Title:          "Penilaian Akhir Semester (PAS) Ganjil - Matematika Wajib",
		SubjectID:      subjectID,
		CreatedByID:    guruID,
		TotalQuestions: 10,
		IsLocked:       true,
		CreatedAt:      time.Now(),
	}
	d.DB.Create(&bank)

	// 6. Questions
	type qSeed struct {
		Number  int
		Content string
		Opts    []domain.OptionItem
		Key     string
	}

	rawQuestions := []qSeed{
		{
			Number:  1,
			Content: "Diketahui persamaan kuadrat 2x² - 5x + 3 = 0. Akar-akar persamaan tersebut adalah x₁ dan x₂. Nilai dari x₁ + x₂ adalah...",
			Opts: []domain.OptionItem{
				{Key: "A", Text: "5/2"},
				{Key: "B", Text: "-5/2"},
				{Key: "C", Text: "3/2"},
				{Key: "D", Text: "-3/2"},
				{Key: "E", Text: "5"},
			},
			Key: "A",
		},
		{
			Number:  2,
			Content: "Turunan pertama dari f(x) = 3x⁴ - 4x³ + 2x² - 5x + 7 adalah f'(x) = ...",
			Opts: []domain.OptionItem{
				{Key: "A", Text: "12x³ - 12x² + 4x - 5"},
				{Key: "B", Text: "12x³ - 12x² + 4x + 7"},
				{Key: "C", Text: "7x³ - 7x² + 4x - 5"},
				{Key: "D", Text: "12x³ - 8x² + 2x - 5"},
				{Key: "E", Text: "3x³ - 4x² + 2x - 5"},
			},
			Key: "A",
		},
		{
			Number:  3,
			Content: "Nilai limit dari lim (x→3) (x² - 9) / (x - 3) adalah...",
			Opts: []domain.OptionItem{
				{Key: "A", Text: "0"},
				{Key: "B", Text: "3"},
				{Key: "C", Text: "6"},
				{Key: "D", Text: "9"},
				{Key: "E", Text: "Tak Hingga"},
			},
			Key: "C",
		},
		{
			Number:  4,
			Content: "Sebuah dadu setimbang bersisi enam dilempar sekali. Peluang munculnya mata dadu bilangan prima adalah...",
			Opts: []domain.OptionItem{
				{Key: "A", Text: "1/6"},
				{Key: "B", Text: "1/3"},
				{Key: "C", Text: "1/2"},
				{Key: "D", Text: "2/3"},
				{Key: "E", Text: "5/6"},
			},
			Key: "C",
		},
		{
			Number:  5,
			Content: "Jika log 2 = a dan log 3 = b, maka nilai dari log 18 adalah...",
			Opts: []domain.OptionItem{
				{Key: "A", Text: "a + 2b"},
				{Key: "B", Text: "2a + b"},
				{Key: "C", Text: "a + b²"},
				{Key: "D", Text: "2(a + b)"},
				{Key: "E", Text: "3a + b"},
			},
			Key: "A",
		},
		{
			Number:  6,
			Content: "Suku ke-n dari barisan aritmetika dinyatakan dengan Un = 4n - 1. Jumlah 10 suku pertama (S₁₀) adalah...",
			Opts: []domain.OptionItem{
				{Key: "A", Text: "210"},
				{Key: "B", Text: "220"},
				{Key: "C", Text: "230"},
				{Key: "D", Text: "240"},
				{Key: "E", Text: "250"},
			},
			Key: "A",
		},
		{
			Number:  7,
			Content: "Himpunan penyelesaian pertidaksamaan 2x - 3 < 5x + 6 adalah...",
			Opts: []domain.OptionItem{
				{Key: "A", Text: "x > -3"},
				{Key: "B", Text: "x < -3"},
				{Key: "C", Text: "x > 3"},
				{Key: "D", Text: "x < 3"},
				{Key: "E", Text: "x > -1"},
			},
			Key: "A",
		},
		{
			Number:  8,
			Content: "Diketahui matriks A = [[2, 3], [1, 4]]. Determinan dari matriks A adalah...",
			Opts: []domain.OptionItem{
				{Key: "A", Text: "5"},
				{Key: "B", Text: "8"},
				{Key: "C", Text: "11"},
				{Key: "D", Text: "-5"},
				{Key: "E", Text: "14"},
			},
			Key: "A",
		},
		{
			Number:  9,
			Content: "Nilai dari sin 30° + cos 60° adalah...",
			Opts: []domain.OptionItem{
				{Key: "A", Text: "0"},
				{Key: "B", Text: "1/2"},
				{Key: "C", Text: "1"},
				{Key: "D", Text: "√2"},
				{Key: "E", Text: "√3"},
			},
			Key: "C",
		},
		{
			Number:  10,
			Content: "Banyaknya susunan huruf berbeda yang dapat dibentuk dari kata 'MALAM' adalah...",
			Opts: []domain.OptionItem{
				{Key: "A", Text: "30"},
				{Key: "B", Text: "60"},
				{Key: "C", Text: "120"},
				{Key: "D", Text: "24"},
				{Key: "E", Text: "10"},
			},
			Key: "A",
		},
	}

	for _, rq := range rawQuestions {
		optsBytes, _ := json.Marshal(rq.Opts)
		q := domain.Question{
			ID:             uuid.New(),
			BankID:         bankID,
			QuestionNumber: rq.Number,
			ContentHTML:    rq.Content,
			OptionsJSON:    string(optsBytes),
			CorrectKey:     rq.Key,
			ScoreWeight:    1.0,
			CreatedAt:      time.Now(),
		}
		d.DB.Create(&q)
	}

	// 7. Class Subject (Penugasan Mengajar Guru)
	classSubject := domain.ClassSubject{
		ID:           uuid.New(),
		ClassRoomID:  classID,
		SubjectID:    subjectID,
		TeacherID:    guruID,
		AcademicYear: "2026/2027",
		CreatedAt:    time.Now(),
	}
	d.DB.Create(&classSubject)

	// 8. Exam Schedule
	now := time.Now()
	schedule := domain.ExamSchedule{
		ID:                 uuid.New(),
		Title:              "PAS Ganjil 2026/2027 - Matematika Wajib",
		SubjectID:          &subjectID,
		BankID:             &bankID,
		ClassRoomID:        classID,
		ExamToken:          "CBT2026",
		StartTime:          now.Add(-2 * time.Hour),
		EndTime:            now.Add(24 * time.Hour),
		DurationMinutes:    90,
		MaxViolations:      3,
		RandomizeQuestions: true,
		RandomizeOptions:   true,
		IsActive:           true,
		CreatedAt:          now,
	}
	d.DB.Create(&schedule)

	log.Println("Seeding completed successfully! Token ujian: CBT2026")
}

func (d *Database) EnsureClassSubjectAssignments() {
	var count int64
	d.DB.Model(&domain.ClassSubject{}).Count(&count)
	if count > 0 {
		return
	}

	var teachers []domain.User
	d.DB.Where("role = ?", domain.RoleGuru).Find(&teachers)
	if len(teachers) == 0 {
		return
	}

	var classes []domain.ClassRoom
	d.DB.Find(&classes)
	var subjects []domain.Subject
	d.DB.Find(&subjects)

	if len(classes) == 0 || len(subjects) == 0 {
		return
	}

	// Petakan penugasan default guru pengampu
	for _, c := range classes {
		for _, s := range subjects {
			cs := domain.ClassSubject{
				ID:           uuid.New(),
				ClassRoomID:  c.ID,
				SubjectID:    s.ID,
				TeacherID:    teachers[0].ID,
				AcademicYear: "2026/2027",
				CreatedAt:    time.Now(),
			}
			d.DB.Create(&cs)
		}
	}
	log.Println("Default ClassSubject relations initialized for existing teacher.")
}

