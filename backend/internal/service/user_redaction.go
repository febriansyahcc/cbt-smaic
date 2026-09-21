package service

import "cbt-backend/internal/domain"

// Pembantu penyuntingan objek User yang ikut termuat lewat relasi (Preload) agar data akun
// tidak bocor ke klien. Seluruhnya bekerja di tempat pada salinan yang akan dikirim.

// ClearBankOwners mengosongkan objek User penyusun (Bank.CreatedBy) pada bank soal milik jadwal.
// Nama penyusun untuk tampilan harus dihitung sebelum pemanggilan ini.
func ClearBankOwners(schedules []domain.ExamSchedule) {
	for i := range schedules {
		if schedules[i].Bank != nil {
			schedules[i].Bank.CreatedBy = domain.User{}
		}
	}
}

// StripBankOwnerPermissions mengosongkan izin pada objek User penyusun tiap bank soal.
// Field lain (mis. full_name untuk nama penyusun) dipertahankan.
func StripBankOwnerPermissions(banks []domain.QuestionBank) {
	for i := range banks {
		banks[i].CreatedBy.Permissions = nil
	}
}

// CanSeeTeacherUsername menyatakan apakah pemanggil boleh melihat username guru pengampu
// (pengelola data master atau akun). Izin penuh dan ADMIN otomatis lolos lewat HasPermission.
func CanSeeTeacherUsername(user domain.User) bool {
	return user.HasPermission(string(domain.PermMasterManage)) ||
		user.HasPermission(string(domain.PermUsersManage))
}

// RedactClassSubjectTeachers menyunting objek guru pada daftar alokasi kelas-mapel:
// izin selalu dikosongkan, username hanya dipertahankan bila includeUsername true.
func RedactClassSubjectTeachers(items []domain.ClassSubject, includeUsername bool) {
	for i := range items {
		items[i].Teacher.Permissions = nil
		if !includeUsername {
			items[i].Teacher.Username = ""
		}
	}
}
