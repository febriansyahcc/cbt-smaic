package staffseed

import (
	"encoding/csv"
	"fmt"
	"io"
	"strings"
)

// errWriter mengumpulkan galat tulis pertama agar kode pencetak tidak penuh pemeriksaan galat.
type errWriter struct {
	w   io.Writer
	err error
}

func (e *errWriter) printf(format string, args ...any) {
	if e.err != nil {
		return
	}
	_, e.err = fmt.Fprintf(e.w, format, args...)
}

// WriteText mencetak laporan yang dapat dibaca manusia (bahasa Indonesia). Kata sandi
// TIDAK PERNAH dicetak di sini; kredensial hanya keluar lewat WriteCredentialsCSV.
func (r Report) WriteText(w io.Writer) error {
	ew := &errWriter{w: w}

	mode := "DRY-RUN (tidak ada yang ditulis ke basis data)"
	if r.Apply {
		mode = "APPLY (perubahan ditulis ke basis data)"
	}
	ew.printf("Mode           : %s\n", mode)
	ew.printf("Tahun ajaran   : %s\n\n", r.AcademicYear)

	ew.printf("Ringkasan\n")
	ew.printf("  Kelas                  : %d baru, %d sudah ada\n", len(r.ClassesNew), len(r.ClassesExisting))
	ew.printf("  Mata pelajaran         : %d baru, %d sudah ada\n", len(r.SubjectsNew), len(r.SubjectsExisting))
	ew.printf("  Guru                   : %d baru, %d sudah ada (tidak berubah), %d diubah (nama)\n",
		len(r.TeachersNew), len(r.TeachersUnchanged), len(r.TeachersUpdated))
	ew.printf("  Penugasan kelas-mapel  : %d dibuat, %d tidak berubah, %d diubah (guru diganti)\n\n",
		r.AssignmentsCreated, r.AssignmentsUnchanged, len(r.AssignmentsChanged))

	if len(r.ClassesNew) > 0 {
		ew.printf("Kelas baru: %s\n", strings.Join(r.ClassesNew, ", "))
	}
	if len(r.SubjectsNew) > 0 {
		ew.printf("Mapel baru (kode): %s\n", strings.Join(r.SubjectsNew, ", "))
	}
	if len(r.TeachersNew) > 0 {
		ew.printf("Guru baru (username): %s\n", strings.Join(r.TeachersNew, ", "))
	}
	if len(r.ClassesNew)+len(r.SubjectsNew)+len(r.TeachersNew) > 0 {
		ew.printf("\n")
	}

	if len(r.TeachersUpdated) > 0 {
		ew.printf("Nama guru yang diubah\n")
		for _, t := range r.TeachersUpdated {
			ew.printf("  %s: %q -> %q\n", t.Username, t.OldName, t.NewName)
		}
		ew.printf("\n")
	}

	if len(r.AssignmentsChanged) > 0 {
		ew.printf("Penugasan yang gurunya diganti\n")
		for _, c := range r.AssignmentsChanged {
			ew.printf("  %s - %s: %s -> %s\n", c.Subject, c.Class, c.OldTeacher, c.NewTeacher)
		}
		ew.printf("\n")
	}

	if len(r.Warnings) > 0 {
		ew.printf("Peringatan\n")
		for _, s := range r.Warnings {
			ew.printf("  - %s\n", s)
		}
		ew.printf("\n")
	}

	if len(r.Gaps) > 0 {
		ew.printf("Celah: pasangan mapel x tingkat tanpa guru di berkas (informasi saja)\n")
		for _, g := range r.Gaps {
			ew.printf("  - %s, tingkat %s\n", g.Subject, g.Grade)
		}
		ew.printf("\n")
	}

	ew.printf("Per guru (mapel - kelas yang akan dilihat guru dari berkas ini)\n")
	for _, t := range r.Teachers {
		ew.printf("  %s\n", teacherLabel(t.FullName, t.Username))
		if len(t.Items) == 0 {
			ew.printf("    (tidak ada penugasan)\n")
		}
		for _, it := range t.Items {
			ew.printf("    - %s\n", it)
		}
	}
	ew.printf("\n")

	switch {
	case r.Committed:
		ew.printf("Selesai: perubahan sudah di-commit. Akun baru: %d (kredensial ada di stdout).\n", len(r.Credentials))
	case r.Apply:
		ew.printf("Tidak ada perubahan yang tersimpan.\n")
	default:
		ew.printf("DRY-RUN selesai: tidak ada yang ditulis. Jalankan ulang dengan --apply untuk menulis.\n")
	}
	return ew.err
}

// WriteCredentialsCSV menulis "username,nama_lengkap,password" untuk akun yang diberikan.
// Pemanggil bertanggung jawab hanya memanggilnya bila Report.Committed.
func WriteCredentialsCSV(w io.Writer, creds []Credential) error {
	cw := csv.NewWriter(w)
	if err := cw.Write([]string{"username", "nama_lengkap", "password"}); err != nil {
		return err
	}
	for _, c := range creds {
		if err := cw.Write([]string{c.Username, c.FullName, c.Password}); err != nil {
			return err
		}
	}
	cw.Flush()
	return cw.Error()
}
