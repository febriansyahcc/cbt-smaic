package service

import (
	"crypto/rand"
	"math/big"
	"time"

	"cbt-backend/internal/domain"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

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

// Token dibagi per sesi waktu: seluruh jadwal pada event yang sama dengan waktu mulai yang
// sama (tanggal + jam) memakai satu token, apa pun kelasnya. Pengawas cukup mengumumkan satu
// token untuk semua ruang pada sesi itu. Jam selesai boleh berbeda (durasi per tingkat).

const examTokenAlphabet = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"

// GenerateExamToken membuat token ujian acak 6 karakter tanpa karakter yang mirip.
func GenerateExamToken() (string, error) {
	max := big.NewInt(int64(len(examTokenAlphabet)))
	b := make([]byte, 6)
	for i := range b {
		n, err := rand.Int(rand.Reader, max)
		if err != nil {
			return "", err
		}
		b[i] = examTokenAlphabet[n.Int64()]
	}
	return string(b), nil
}

// sessionSlot membatasi query ke jadwal satu sesi waktu (event + waktu mulai).
func sessionSlot(db *gorm.DB, eventID *uuid.UUID, start time.Time) *gorm.DB {
	q := db.Model(&domain.ExamSchedule{}).Where("start_time = ?", start)
	if eventID == nil {
		return q.Where("event_id IS NULL")
	}
	return q.Where("event_id = ?", *eventID)
}

// SessionToken mengembalikan token yang sudah dipakai jadwal lain pada sesi yang sama
// (jadwal tertua menjadi acuan bila data lama belum seragam).
func SessionToken(db *gorm.DB, eventID *uuid.UUID, start time.Time, excludeID uuid.UUID) (string, bool) {
	var other domain.ExamSchedule
	err := sessionSlot(db, eventID, start).
		Where("id <> ? AND exam_token <> ''", excludeID).
		Order("created_at ASC").
		First(&other).Error
	if err != nil {
		return "", false
	}
	return other.ExamToken, true
}

// ApplySessionToken menyamakan token seluruh jadwal pada sesi yang sama.
func ApplySessionToken(db *gorm.DB, eventID *uuid.UUID, start time.Time, token string) error {
	return sessionSlot(db, eventID, start).Update("exam_token", token).Error
}

// RegenerateSessionTokens membuat satu token baru untuk setiap sesi waktu yang memuat jadwal
// terpilih, lalu menerapkannya ke semua jadwal di sesi tersebut. Mengembalikan jumlah sesi.
func RegenerateSessionTokens(db *gorm.DB, scheduleIDs []uuid.UUID) (int, error) {
	var picked []domain.ExamSchedule
	if err := db.Where("id IN ?", scheduleIDs).Find(&picked).Error; err != nil {
		return 0, err
	}
	type slotKey struct {
		event uuid.UUID
		start int64
	}
	done := map[slotKey]bool{}
	err := db.Transaction(func(tx *gorm.DB) error {
		for _, s := range picked {
			key := slotKey{start: s.StartTime.UnixNano()}
			if s.EventID != nil {
				key.event = *s.EventID
			}
			if done[key] {
				continue
			}
			done[key] = true
			token, err := GenerateExamToken()
			if err != nil {
				return err
			}
			if err := ApplySessionToken(tx, s.EventID, s.StartTime, token); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return 0, err
	}
	return len(done), nil
}
