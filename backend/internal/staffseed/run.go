package staffseed

import (
	"crypto/rand"
	"errors"
	"fmt"
	"io"
	"time"

	"cbt-backend/internal/domain"
	"cbt-backend/internal/repository"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// errDryRun adalah sentinel internal: memaksa transaksi di-rollback pada mode dry-run
// tanpa dianggap galat oleh pemanggil.
var errDryRun = errors.New("dry-run: transaksi dibatalkan")

// Credential adalah akun guru baru beserta kata sandi awalnya (teks biasa, hanya ada di
// memori). String/GoString sengaja tidak menampilkan kata sandi agar tidak bocor lewat
// format %v atau log yang tidak disengaja.
type Credential struct {
	Username string
	FullName string
	Password string
}

func (c Credential) String() string   { return "Credential{" + c.Username + "}" }
func (c Credential) GoString() string { return c.String() }

// TeacherRename mencatat nama lengkap guru yang sudah ada disamakan dengan berkas.
type TeacherRename struct {
	Username string
	OldName  string
	NewName  string
}

// AssignmentChange mencatat penugasan kelas-mapel yang gurunya diganti.
type AssignmentChange struct {
	Class      string
	Subject    string // "Nama (KODE)"
	OldTeacher string // "Nama (username)"
	NewTeacher string
}

// Gap adalah pasangan mapel x tingkat yang tidak punya guru di berkas (informasi saja).
type Gap struct {
	Subject string // "Nama (KODE)"
	Grade   string
}

// TeacherReport memuat daftar "Mapel - kelas" yang akan dilihat satu guru dari berkas.
type TeacherReport struct {
	Username string
	FullName string
	Items    []string
}

// Report adalah hasil Run. Isinya identik pada dry-run dan apply karena jalur kodenya sama.
type Report struct {
	Apply        bool // mode apply diminta
	Committed    bool // transaksi benar-benar ter-commit (hanya mungkin bila Apply)
	AcademicYear string

	ClassesNew      []string // nama kelas
	ClassesExisting []string

	SubjectsNew      []string // kode mapel
	SubjectsExisting []string

	TeachersNew       []string // username
	TeachersUnchanged []string
	TeachersUpdated   []TeacherRename

	AssignmentsCreated   int
	AssignmentsUnchanged int
	AssignmentsChanged   []AssignmentChange

	Gaps     []Gap
	Warnings []string
	Teachers []TeacherReport

	// Credentials hanya berisi akun BARU dan hanya terisi bila Committed. Jangan dicetak ke log.
	Credentials []Credential
}

// Run memasukkan data ke basis data di dalam SATU transaksi. Bila apply == false transaksi
// di-rollback di akhir, sehingga hitungan pada laporan dry-run dihasilkan oleh jalur kode
// yang sama dengan apply. Galat di tengah jalan me-rollback seluruhnya.
//
// now dan rng boleh nil (bawaan: time.Now dan crypto/rand.Reader); keduanya ada agar tes
// deterministik.
func Run(db *gorm.DB, data Data, apply bool, now func() time.Time, rng io.Reader) (Report, error) {
	if err := data.Validate(); err != nil {
		return Report{}, err
	}
	if now == nil {
		now = time.Now
	}
	if rng == nil {
		rng = rand.Reader
	}

	var rep Report
	err := db.Transaction(func(tx *gorm.DB) error {
		r, err := seed(tx, data, now, rng)
		if err != nil {
			return err
		}
		rep = r
		if !apply {
			return errDryRun
		}
		return nil
	})
	switch {
	case err == nil:
		rep.Committed = true
	case errors.Is(err, errDryRun):
		rep.Credentials = nil // dry-run: tidak ada akun yang benar-benar dibuat
	default:
		return Report{}, err
	}
	rep.Apply = apply
	return rep, nil
}

func teacherLabel(name, username string) string {
	return fmt.Sprintf("%s (%s)", name, username)
}

func seed(tx *gorm.DB, data Data, now func() time.Time, rng io.Reader) (Report, error) {
	rep := Report{AcademicYear: data.AcademicYear}

	// 1. Baca data yang sudah ada.
	classNames := make([]string, 0, len(data.Classes))
	for _, c := range data.Classes {
		classNames = append(classNames, c.Name)
	}
	var classRows []domain.ClassRoom
	if err := tx.Where("name IN ?", classNames).Order("created_at ASC, id ASC").Find(&classRows).Error; err != nil {
		return rep, fmt.Errorf("gagal membaca kelas: %w", err)
	}
	classByName := map[string]domain.ClassRoom{} // yang paling awal dibuat menang
	classCount := map[string]int{}
	for _, c := range classRows {
		classCount[c.Name]++
		if _, ok := classByName[c.Name]; !ok {
			classByName[c.Name] = c
		}
	}

	subjectCodes := make([]string, 0, len(data.Subjects))
	for _, s := range data.Subjects {
		subjectCodes = append(subjectCodes, s.Code)
	}
	var subjectRows []domain.Subject
	if err := tx.Where("code IN ?", subjectCodes).Find(&subjectRows).Error; err != nil {
		return rep, fmt.Errorf("gagal membaca mapel: %w", err)
	}
	subjectByCode := map[string]domain.Subject{}
	for _, s := range subjectRows {
		subjectByCode[s.Code] = s
	}

	usernames := make([]string, 0, len(data.Teachers))
	for _, t := range data.Teachers {
		usernames = append(usernames, t.Username)
	}
	var userRows []domain.User
	if err := tx.Where("username IN ?", usernames).Find(&userRows).Error; err != nil {
		return rep, fmt.Errorf("gagal membaca akun: %w", err)
	}
	userByName := map[string]domain.User{}
	for _, u := range userRows {
		userByName[u.Username] = u
	}

	// 2. Periksa konflik dengan data yang ada SEBELUM menulis apa pun.
	var problems []string
	for _, s := range data.Subjects {
		if ex, ok := subjectByCode[s.Code]; ok && ex.Name != s.Name {
			problems = append(problems, fmt.Sprintf("mapel dengan kode %q sudah ada dengan nama %q, berbeda dari berkas (%q); tidak ditimpa", s.Code, ex.Name, s.Name))
		}
	}
	for _, t := range data.Teachers {
		if ex, ok := userByName[t.Username]; ok && ex.Role != domain.RoleGuru {
			problems = append(problems, fmt.Sprintf("username %q sudah dipakai akun dengan peran %s (bukan GURU); tidak diubah", t.Username, ex.Role))
		}
	}
	if len(problems) > 0 {
		return rep, &ValidationError{Problems: problems}
	}

	// 3. Kelas.
	classID := map[string]uuid.UUID{}
	for _, c := range data.Classes {
		if ex, ok := classByName[c.Name]; ok {
			classID[c.Name] = ex.ID
			rep.ClassesExisting = append(rep.ClassesExisting, c.Name)
			if ex.Grade != c.Grade {
				rep.Warnings = append(rep.Warnings, fmt.Sprintf("kelas %q sudah ada dengan tingkat %q (berkas: %q); dibiarkan, penugasan mengikuti tingkat di berkas", c.Name, ex.Grade, c.Grade))
			}
			if classCount[c.Name] > 1 {
				rep.Warnings = append(rep.Warnings, fmt.Sprintf("ada %d kelas bernama %q; dipakai yang paling awal dibuat", classCount[c.Name], c.Name))
			}
			continue
		}
		row := domain.ClassRoom{ID: uuid.New(), Name: c.Name, Grade: c.Grade, CreatedAt: now()}
		if err := tx.Create(&row).Error; err != nil {
			return rep, fmt.Errorf("gagal membuat kelas %q: %w", c.Name, err)
		}
		classID[c.Name] = row.ID
		rep.ClassesNew = append(rep.ClassesNew, c.Name)
	}

	// 4. Mapel.
	subjectID := map[string]uuid.UUID{}
	subjectName := map[string]string{}
	for _, s := range data.Subjects {
		subjectName[s.Code] = s.Name
		if ex, ok := subjectByCode[s.Code]; ok {
			subjectID[s.Code] = ex.ID
			rep.SubjectsExisting = append(rep.SubjectsExisting, s.Code)
			continue
		}
		row := domain.Subject{ID: uuid.New(), Code: s.Code, Name: s.Name, CreatedAt: now()}
		if err := tx.Create(&row).Error; err != nil {
			return rep, fmt.Errorf("gagal membuat mapel %q: %w", s.Code, err)
		}
		subjectID[s.Code] = row.ID
		rep.SubjectsNew = append(rep.SubjectsNew, s.Code)
	}

	// 5. Guru.
	teacherID := map[string]uuid.UUID{}
	for _, t := range data.Teachers {
		if ex, ok := userByName[t.Username]; ok {
			teacherID[t.Username] = ex.ID
			if !ex.IsActive {
				rep.Warnings = append(rep.Warnings, fmt.Sprintf("akun guru %q sudah ada tetapi nonaktif; tidak diaktifkan, guru belum dapat masuk", t.Username))
			}
			if ex.FullName != t.FullName {
				if err := tx.Model(&domain.User{}).Where("id = ?", ex.ID).Update("full_name", t.FullName).Error; err != nil {
					return rep, fmt.Errorf("gagal memperbarui nama guru %q: %w", t.Username, err)
				}
				rep.TeachersUpdated = append(rep.TeachersUpdated, TeacherRename{Username: t.Username, OldName: ex.FullName, NewName: t.FullName})
			} else {
				rep.TeachersUnchanged = append(rep.TeachersUnchanged, t.Username)
			}
			continue
		}

		password, err := GeneratePassword(rng)
		if err != nil {
			return rep, err
		}
		hash := repository.HashPassword(password)
		if hash == "" {
			return rep, fmt.Errorf("gagal meng-hash kata sandi untuk %q", t.Username)
		}
		ts := now()
		row := domain.User{
			ID:           uuid.New(),
			Username:     t.Username,
			PasswordHash: hash,
			FullName:     t.FullName,
			Role:         domain.RoleGuru,
			Permissions:  domain.TemplatePermissions("guru"),
			IsActive:     true,
			CreatedAt:    ts,
			UpdatedAt:    ts,
		}
		if err := tx.Create(&row).Error; err != nil {
			return rep, fmt.Errorf("gagal membuat akun guru %q: %w", t.Username, err)
		}
		teacherID[t.Username] = row.ID
		rep.TeachersNew = append(rep.TeachersNew, t.Username)
		rep.Credentials = append(rep.Credentials, Credential{Username: t.Username, FullName: t.FullName, Password: password})
	}

	// 6. Penugasan kelas-mapel.
	var csRows []domain.ClassSubject
	if err := tx.Where("academic_year = ?", data.AcademicYear).Find(&csRows).Error; err != nil {
		return rep, fmt.Errorf("gagal membaca penugasan kelas-mapel: %w", err)
	}
	type key struct{ class, subject uuid.UUID }
	existing := map[key][]domain.ClassSubject{}
	for _, cs := range csRows {
		k := key{cs.ClassRoomID, cs.SubjectID}
		existing[k] = append(existing[k], cs)
	}

	type pendingChange struct {
		class, subject string
		oldTeacher     uuid.UUID
		newTeacher     string // username
	}
	var pending []pendingChange
	owned := map[[2]string]bool{} // (kode mapel, tingkat) yang punya guru

	for _, t := range data.Teachers {
		tr := TeacherReport{Username: t.Username, FullName: t.FullName}
		if len(t.Assignments) == 0 {
			rep.Warnings = append(rep.Warnings, fmt.Sprintf("guru %q tidak punya penugasan di berkas", t.Username))
		}
		for _, a := range t.Assignments {
			for _, g := range sortedGrades(a.Grades) {
				owned[[2]string{a.Subject, g}] = true
				for _, c := range data.classesOfGrade(g) {
					tr.Items = append(tr.Items, fmt.Sprintf("%s - %s", subjectName[a.Subject], c.Name))

					k := key{classID[c.Name], subjectID[a.Subject]}
					rows := existing[k]
					if len(rows) == 0 {
						row := domain.ClassSubject{
							ID:           uuid.New(),
							ClassRoomID:  k.class,
							SubjectID:    k.subject,
							TeacherID:    teacherID[t.Username],
							AcademicYear: data.AcademicYear,
							CreatedAt:    now(),
						}
						if err := tx.Create(&row).Error; err != nil {
							return rep, fmt.Errorf("gagal membuat penugasan %s - %s untuk %q: %w", a.Subject, c.Name, t.Username, err)
						}
						existing[k] = []domain.ClassSubject{row}
						rep.AssignmentsCreated++
						continue
					}

					changed := false
					for _, r := range rows {
						if r.TeacherID == teacherID[t.Username] {
							continue
						}
						if !changed {
							pending = append(pending, pendingChange{class: c.Name, subject: fmt.Sprintf("%s (%s)", subjectName[a.Subject], a.Subject), oldTeacher: r.TeacherID, newTeacher: t.Username})
						}
						changed = true
						if err := tx.Model(&domain.ClassSubject{}).Where("id = ?", r.ID).Update("teacher_id", teacherID[t.Username]).Error; err != nil {
							return rep, fmt.Errorf("gagal mengganti guru penugasan %s - %s: %w", a.Subject, c.Name, err)
						}
					}
					if changed {
						// Perbarui salinan lokal agar pemeriksaan berikutnya konsisten.
						for i := range rows {
							rows[i].TeacherID = teacherID[t.Username]
						}
					} else {
						rep.AssignmentsUnchanged++
					}
				}
			}
		}
		rep.Teachers = append(rep.Teachers, tr)
	}

	// Ubah id guru lama menjadi "Nama (username)" untuk laporan.
	if len(pending) > 0 {
		ids := make([]uuid.UUID, 0, len(pending))
		for _, p := range pending {
			ids = append(ids, p.oldTeacher)
		}
		var olds []domain.User
		if err := tx.Select("id", "username", "full_name").Where("id IN ?", ids).Find(&olds).Error; err != nil {
			return rep, fmt.Errorf("gagal membaca guru lama: %w", err)
		}
		oldLabel := map[uuid.UUID]string{}
		for _, u := range olds {
			oldLabel[u.ID] = teacherLabel(u.FullName, u.Username)
		}
		fullName := map[string]string{}
		for _, t := range data.Teachers {
			fullName[t.Username] = t.FullName
		}
		for _, p := range pending {
			label, ok := oldLabel[p.oldTeacher]
			if !ok {
				label = "(akun tidak ditemukan: " + p.oldTeacher.String() + ")"
			}
			rep.AssignmentsChanged = append(rep.AssignmentsChanged, AssignmentChange{
				Class:      p.class,
				Subject:    p.subject,
				OldTeacher: label,
				NewTeacher: teacherLabel(fullName[p.newTeacher], p.newTeacher),
			})
		}
	}

	// 7. Celah: mapel x tingkat (yang punya kelas di berkas) tanpa guru di berkas.
	for _, s := range data.Subjects {
		for _, g := range validGrades {
			if len(data.classesOfGrade(g)) == 0 || owned[[2]string{s.Code, g}] {
				continue
			}
			rep.Gaps = append(rep.Gaps, Gap{Subject: fmt.Sprintf("%s (%s)", s.Name, s.Code), Grade: g})
		}
	}

	return rep, nil
}
