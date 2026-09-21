package service

import "cbt-backend/internal/domain"

// ClearExamTokens mengosongkan exam_token pada seluruh jadwal (di tempat).
// Dipakai untuk respons yang tidak boleh membawa token sama sekali, misalnya daftar
// jadwal siswa: token hanya boleh diumumkan pengawas, bukan dibaca dari respons API.
// Field tetap ada di JSON (bernilai string kosong) agar bentuk respons tidak berubah.
func ClearExamTokens(schedules []domain.ExamSchedule) {
	for i := range schedules {
		schedules[i].ExamToken = ""
	}
}

// RedactToken mengosongkan exam_token pada jadwal yang berada di luar cakupan token
// (lihat AccessService.TokenScopeFor). Jadwal dalam cakupan tidak diubah.
func (c ControlScope) RedactToken(schedule *domain.ExamSchedule) {
	if !c.Allows(schedule.ID) {
		schedule.ExamToken = ""
	}
}

// RedactTokens menerapkan RedactToken pada setiap jadwal dalam daftar (di tempat).
func (c ControlScope) RedactTokens(schedules []domain.ExamSchedule) {
	for i := range schedules {
		c.RedactToken(&schedules[i])
	}
}
