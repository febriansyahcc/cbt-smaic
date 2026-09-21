// Package staffseed memasukkan guru, mata pelajaran, kelas, dan penugasan kelas-mapel dari
// berkas data (JSON) ke basis data. Logika inti ada di sini agar dapat diuji tanpa CLI;
// cmd/seedstaff hanya membaca flag, membuka koneksi, dan memanggil Run.
//
// Aturan dasarnya: hanya menambah. Tidak ada baris yang dihapus, dan data yang sudah ada
// tidak ditimpa kecuali dua hal yang memang dimaksudkan: nama lengkap guru yang sudah ada
// disamakan dengan berkas, dan guru pengampu pada penugasan kelas-mapel yang sudah ada
// diganti dengan guru di berkas (keduanya dicatat di laporan).
package staffseed

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"regexp"
	"sort"
	"strings"
)

// Tingkat yang dikenali, berurutan.
var validGrades = []string{"X", "XI", "XII"}

var reUsername = regexp.MustCompile(`^[a-z0-9._-]+$`)

// Batas panjang kolom (lihat domain/models.go) agar galat muncul sebagai galat validasi
// yang jelas, bukan galat basis data di tengah jalan.
const (
	maxUsername  = 100
	maxFullName  = 255
	maxClassName = 100
	maxGrade     = 20
	maxSubjCode  = 50
	maxSubjName  = 100
	maxYear      = 20
)

type ClassData struct {
	Name  string `json:"name"`
	Grade string `json:"grade"`
}

type SubjectData struct {
	Code string `json:"code"`
	Name string `json:"name"`
}

type AssignmentData struct {
	Subject string   `json:"subject"`
	Grades  []string `json:"grades"`
}

type TeacherData struct {
	Username    string           `json:"username"`
	FullName    string           `json:"full_name"`
	Assignments []AssignmentData `json:"assignments"`
}

// Data adalah isi berkas seed.
type Data struct {
	AcademicYear string        `json:"academic_year"`
	Classes      []ClassData   `json:"classes"`
	Subjects     []SubjectData `json:"subjects"`
	Teachers     []TeacherData `json:"teachers"`
}

// ValidationError mengumpulkan semua masalah yang ditemukan sekaligus agar pengguna dapat
// memperbaikinya dalam satu putaran.
type ValidationError struct {
	Problems []string
}

func (e *ValidationError) Error() string {
	var b strings.Builder
	fmt.Fprintf(&b, "validasi gagal (%d masalah):", len(e.Problems))
	for _, p := range e.Problems {
		b.WriteString("\n  - ")
		b.WriteString(p)
	}
	return b.String()
}

// LoadData membaca dan memvalidasi data seed dari JSON. Kolom yang tidak dikenal ditolak
// supaya salah ketik pada nama kunci tidak lolos diam-diam.
func LoadData(r io.Reader) (Data, error) {
	dec := json.NewDecoder(r)
	dec.DisallowUnknownFields()
	var d Data
	if err := dec.Decode(&d); err != nil {
		return Data{}, fmt.Errorf("data seed bukan JSON yang valid: %w", err)
	}
	if err := dec.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return Data{}, errors.New("data seed bukan JSON yang valid: ada isi tambahan setelah objek utama")
	}
	if err := d.Validate(); err != nil {
		return Data{}, err
	}
	return d, nil
}

func isValidGrade(g string) bool {
	for _, v := range validGrades {
		if v == g {
			return true
		}
	}
	return false
}

func gradeIndex(g string) int {
	for i, v := range validGrades {
		if v == g {
			return i
		}
	}
	return len(validGrades)
}

// checkText memeriksa string wajib: tidak kosong, tanpa spasi di tepi, dan tidak melebihi kolom.
func checkText(problems *[]string, label, val string, max int) {
	switch {
	case strings.TrimSpace(val) == "":
		*problems = append(*problems, label+" tidak boleh kosong")
	case val != strings.TrimSpace(val):
		*problems = append(*problems, fmt.Sprintf("%s %q tidak boleh diawali atau diakhiri spasi", label, val))
	case len([]rune(val)) > max:
		*problems = append(*problems, fmt.Sprintf("%s %q melebihi %d karakter", label, val, max))
	}
}

// Validate memeriksa data tanpa menyentuh basis data. Hasilnya *ValidationError bila ada masalah.
func (d Data) Validate() error {
	var p []string

	checkText(&p, "academic_year", d.AcademicYear, maxYear)

	// Kelas.
	classNames := map[string]bool{}
	gradesWithClass := map[string]bool{}
	for i, c := range d.Classes {
		label := fmt.Sprintf("kelas #%d", i+1)
		checkText(&p, label+" (name)", c.Name, maxClassName)
		if !isValidGrade(c.Grade) {
			p = append(p, fmt.Sprintf("%s %q: tingkat %q tidak valid (harus X, XI, atau XII)", label, c.Name, c.Grade))
		} else {
			gradesWithClass[c.Grade] = true
		}
		if c.Name != "" {
			if classNames[c.Name] {
				p = append(p, fmt.Sprintf("nama kelas %q muncul lebih dari sekali", c.Name))
			}
			classNames[c.Name] = true
		}
	}

	// Mapel.
	subjectCodes := map[string]bool{}
	for i, s := range d.Subjects {
		label := fmt.Sprintf("mapel #%d", i+1)
		checkText(&p, label+" (code)", s.Code, maxSubjCode)
		checkText(&p, label+" (name)", s.Name, maxSubjName)
		if s.Code != "" {
			if subjectCodes[s.Code] {
				p = append(p, fmt.Sprintf("kode mapel %q muncul lebih dari sekali", s.Code))
			}
			subjectCodes[s.Code] = true
		}
	}

	// Guru dan penugasan.
	if len(d.Teachers) == 0 {
		p = append(p, "data tidak memuat satu pun guru")
	}
	usernames := map[string]bool{}
	type pair struct{ subject, grade string }
	owner := map[pair]string{} // (mapel, tingkat) -> username pemilik pertama
	for i, t := range d.Teachers {
		label := fmt.Sprintf("guru #%d", i+1)
		switch {
		case t.Username == "":
			p = append(p, label+": username tidak boleh kosong")
		case !reUsername.MatchString(t.Username):
			p = append(p, fmt.Sprintf("%s: username %q tidak valid (hanya huruf kecil a-z, angka, titik, garis bawah, dan strip)", label, t.Username))
		case len(t.Username) > maxUsername:
			p = append(p, fmt.Sprintf("%s: username %q melebihi %d karakter", label, t.Username, maxUsername))
		default:
			label = fmt.Sprintf("guru %q", t.Username)
			if usernames[t.Username] {
				p = append(p, fmt.Sprintf("username %q muncul lebih dari sekali", t.Username))
			}
			usernames[t.Username] = true
		}
		checkText(&p, label+" (full_name)", t.FullName, maxFullName)

		for j, a := range t.Assignments {
			alabel := fmt.Sprintf("%s penugasan #%d", label, j+1)
			if !subjectCodes[a.Subject] {
				p = append(p, fmt.Sprintf("%s: mapel %q tidak ada di daftar mapel", alabel, a.Subject))
			}
			if len(a.Grades) == 0 {
				p = append(p, fmt.Sprintf("%s (mapel %q): daftar grades kosong", alabel, a.Subject))
			}
			for _, g := range a.Grades {
				if !isValidGrade(g) {
					p = append(p, fmt.Sprintf("%s (mapel %q): tingkat %q tidak valid (harus X, XI, atau XII)", alabel, a.Subject, g))
					continue
				}
				if !gradesWithClass[g] {
					p = append(p, fmt.Sprintf("%s (mapel %q): tingkat %s dipakai tetapi tidak punya satu pun kelas di daftar kelas", alabel, a.Subject, g))
				}
				k := pair{a.Subject, g}
				if prev, dup := owner[k]; dup {
					if prev == t.Username {
						p = append(p, fmt.Sprintf("guru %q memuat pasangan mapel %s tingkat %s lebih dari sekali", t.Username, a.Subject, g))
					} else {
						p = append(p, fmt.Sprintf("pasangan mapel %s tingkat %s dimiliki dua guru: %q dan %q (hanya boleh satu)", a.Subject, g, prev, t.Username))
					}
					continue
				}
				owner[k] = t.Username
			}
		}
	}

	if len(p) == 0 {
		return nil
	}
	return &ValidationError{Problems: dedupe(p)}
}

// dedupe membuang pesan kembar (mis. tingkat tanpa kelas yang dipakai berkali-kali)
// dengan mempertahankan urutan kemunculan.
func dedupe(in []string) []string {
	seen := map[string]bool{}
	out := make([]string, 0, len(in))
	for _, s := range in {
		if !seen[s] {
			seen[s] = true
			out = append(out, s)
		}
	}
	return out
}

// classesOfGrade mengembalikan kelas pada berkas untuk satu tingkat, urut seperti di berkas.
func (d Data) classesOfGrade(grade string) []ClassData {
	var out []ClassData
	for _, c := range d.Classes {
		if c.Grade == grade {
			out = append(out, c)
		}
	}
	return out
}

// sortedGrades mengurutkan tingkat menurut X, XI, XII.
func sortedGrades(in []string) []string {
	out := append([]string(nil), in...)
	sort.SliceStable(out, func(i, j int) bool { return gradeIndex(out[i]) < gradeIndex(out[j]) })
	return out
}
