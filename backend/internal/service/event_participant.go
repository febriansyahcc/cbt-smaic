package service

import (
	"crypto/rand"
	"crypto/subtle"
	"errors"
	"fmt"
	"math/big"
	"sort"
	"strconv"
	"strings"
	"time"

	"cbt-backend/internal/domain"
	"cbt-backend/internal/repository"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Huruf dan angka tanpa karakter yang mirip (0/O, 1/I/L) agar mudah dibaca dari kartu cetak.
const cardPasswordAlphabet = "ABCDEFGHJKMNPQRSTUVWXYZ23456789"
const cardPasswordLength = 6

var ErrEventNotFound = errors.New("event tidak ditemukan")

type EventParticipantService struct {
	repo *repository.Database
}

func NewEventParticipantService(repo *repository.Database) *EventParticipantService {
	return &EventParticipantService{repo: repo}
}

// ParticipantCard adalah satu baris data kartu peserta.
type ParticipantCard struct {
	UserID     uuid.UUID `json:"user_id"`
	FullName   string    `json:"full_name"`
	NIS        string    `json:"nis"`
	ClassID    uuid.UUID `json:"class_id"`
	ClassName  string    `json:"class_name"`
	Grade      string    `json:"grade"`
	ExamNumber string    `json:"exam_number"`
	Password   string    `json:"password"`
}

// GradeDigit memetakan tingkat kelas ke satu digit pada nomor ujian: X→1, XI→2, XII→3.
func GradeDigit(grade string) string {
	switch strings.ToUpper(strings.TrimSpace(grade)) {
	case "X", "10":
		return "1"
	case "XI", "11":
		return "2"
	case "XII", "12":
		return "3"
	}
	return "0"
}

// FormatExamNumber menyusun nomor ujian: KODE EVENT + digit tingkat + 4 digit urut, mis. PSAT10001.
func FormatExamNumber(eventCode, gradeDigit string, seq int) string {
	return fmt.Sprintf("%s%s%04d", strings.ToUpper(strings.TrimSpace(eventCode)), gradeDigit, seq)
}

// GenerateCardPassword membuat kata sandi acak kriptografis untuk kartu peserta.
func GenerateCardPassword() (string, error) {
	max := big.NewInt(int64(len(cardPasswordAlphabet)))
	b := make([]byte, cardPasswordLength)
	for i := range b {
		n, err := rand.Int(rand.Reader, max)
		if err != nil {
			return "", err
		}
		b[i] = cardPasswordAlphabet[n.Int64()]
	}
	return string(b), nil
}

type eventStudentRow struct {
	UserID    uuid.UUID
	FullName  string
	NIS       string
	ClassID   uuid.UUID
	ClassName string
	Grade     string
}

// eventStudents mengambil siswa aktif dari kelas yang memiliki jadwal pada event.
func (s *EventParticipantService) eventStudents(db *gorm.DB, eventID uuid.UUID) ([]eventStudentRow, error) {
	var rows []eventStudentRow
	err := db.Table("student_profiles sp").
		Select("sp.user_id AS user_id, u.full_name AS full_name, sp.nis AS nis, sp.class_room_id AS class_id, c.name AS class_name, c.grade AS grade").
		Joins("JOIN users u ON u.id = sp.user_id").
		Joins("JOIN class_rooms c ON c.id = sp.class_room_id").
		Where("u.is_active = ? AND u.role = ?", true, domain.RoleSiswa).
		Where("sp.class_room_id IN (?)", db.Model(&domain.ExamSchedule{}).Select("class_room_id").Where("event_id = ?", eventID)).
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	sort.SliceStable(rows, func(i, j int) bool {
		gi, gj := GradeDigit(rows[i].Grade), GradeDigit(rows[j].Grade)
		if gi != gj {
			return gi < gj
		}
		if rows[i].ClassName != rows[j].ClassName {
			return rows[i].ClassName < rows[j].ClassName
		}
		return strings.ToUpper(rows[i].FullName) < strings.ToUpper(rows[j].FullName)
	})
	return rows, nil
}

// Generate membuat akun peserta untuk siswa event yang belum memilikinya. Nomor dan kata sandi
// yang sudah ada dipertahankan agar kartu yang telah dicetak tetap berlaku; reset=true menghapus
// seluruh akun peserta event lalu membuat ulang semuanya. Mengembalikan jumlah akun baru.
func (s *EventParticipantService) Generate(eventID uuid.UUID, reset bool) (int, error) {
	var event domain.ExamEvent
	if err := s.repo.DB.First(&event, "id = ?", eventID).Error; err != nil {
		return 0, ErrEventNotFound
	}

	created := 0
	err := s.repo.DB.Transaction(func(tx *gorm.DB) error {
		if reset {
			if err := tx.Where("event_id = ?", eventID).Delete(&domain.EventParticipant{}).Error; err != nil {
				return err
			}
		}

		students, err := s.eventStudents(tx, eventID)
		if err != nil {
			return err
		}

		var existing []domain.EventParticipant
		if err := tx.Where("event_id = ?", eventID).Find(&existing).Error; err != nil {
			return err
		}
		has := make(map[uuid.UUID]bool, len(existing))
		for _, p := range existing {
			has[p.UserID] = true
		}

		// Nomor yang sudah terpakai (di event mana pun) dengan awalan kode event ini.
		var taken []string
		if err := tx.Model(&domain.EventParticipant{}).
			Where("exam_number LIKE ?", strings.ToUpper(event.Code)+"%").
			Pluck("exam_number", &taken).Error; err != nil {
			return err
		}
		takenSet := make(map[string]bool, len(taken))
		for _, n := range taken {
			takenSet[n] = true
		}

		// Lanjutkan urutan dari nomor tertinggi per tingkat pada event ini.
		nextSeq := map[string]int{}
		prefix := strings.ToUpper(event.Code)
		for _, p := range existing {
			if !strings.HasPrefix(p.ExamNumber, prefix) || len(p.ExamNumber) < len(prefix)+2 {
				continue
			}
			rest := p.ExamNumber[len(prefix):]
			if n, err := strconv.Atoi(rest[1:]); err == nil && n > nextSeq[rest[:1]] {
				nextSeq[rest[:1]] = n
			}
		}

		now := time.Now()
		for _, st := range students {
			if has[st.UserID] {
				continue
			}
			digit := GradeDigit(st.Grade)
			var number string
			for {
				nextSeq[digit]++
				number = FormatExamNumber(event.Code, digit, nextSeq[digit])
				if !takenSet[number] {
					break
				}
			}
			takenSet[number] = true

			pass, err := GenerateCardPassword()
			if err != nil {
				return err
			}
			p := domain.EventParticipant{
				ID:         uuid.New(),
				EventID:    eventID,
				UserID:     st.UserID,
				ExamNumber: number,
				Password:   pass,
				CreatedAt:  now,
			}
			if err := tx.Create(&p).Error; err != nil {
				return err
			}
			created++
		}
		return nil
	})
	if err != nil {
		return 0, err
	}
	return created, nil
}

// List mengembalikan data kartu peserta event, opsional difilter per kelas. Siswa yang belum
// memiliki akun peserta tetap dimuat dengan ExamNumber kosong, sehingga antarmuka dapat
// menunjukkan bahwa akun perlu dibuat.
func (s *EventParticipantService) List(eventID uuid.UUID, classID *uuid.UUID) ([]ParticipantCard, error) {
	var event domain.ExamEvent
	if err := s.repo.DB.First(&event, "id = ?", eventID).Error; err != nil {
		return nil, ErrEventNotFound
	}
	students, err := s.eventStudents(s.repo.DB, eventID)
	if err != nil {
		return nil, err
	}
	var parts []domain.EventParticipant
	if err := s.repo.DB.Where("event_id = ?", eventID).Find(&parts).Error; err != nil {
		return nil, err
	}
	byUser := make(map[uuid.UUID]domain.EventParticipant, len(parts))
	for _, p := range parts {
		byUser[p.UserID] = p
	}

	cards := make([]ParticipantCard, 0, len(students))
	for _, st := range students {
		if classID != nil && st.ClassID != *classID {
			continue
		}
		p := byUser[st.UserID]
		cards = append(cards, ParticipantCard{
			UserID:     st.UserID,
			FullName:   st.FullName,
			NIS:        st.NIS,
			ClassID:    st.ClassID,
			ClassName:  st.ClassName,
			Grade:      st.Grade,
			ExamNumber: p.ExamNumber,
			Password:   p.Password,
		})
	}
	sort.SliceStable(cards, func(i, j int) bool {
		// Peserta bernomor diurutkan sesuai nomor ujian; yang belum bernomor di akhir.
		if (cards[i].ExamNumber == "") != (cards[j].ExamNumber == "") {
			return cards[j].ExamNumber == ""
		}
		return cards[i].ExamNumber < cards[j].ExamNumber
	})
	return cards, nil
}

// findParticipantLogin mencari akun peserta pada event AKTIF yang cocok dengan nomor ujian dan
// kata sandi. Mengembalikan user siswa pemilik akun tersebut.
func findParticipantLogin(db *gorm.DB, examNumber, password string) (*domain.User, bool) {
	var p domain.EventParticipant
	err := db.Joins("JOIN exam_events e ON e.id = event_participants.event_id AND e.is_active = ?", true).
		Where("event_participants.exam_number = ?", strings.ToUpper(strings.TrimSpace(examNumber))).
		First(&p).Error
	if err != nil {
		return nil, false
	}
	if subtle.ConstantTimeCompare([]byte(strings.ToUpper(password)), []byte(p.Password)) != 1 {
		return nil, false
	}
	var user domain.User
	if err := db.First(&user, "id = ? AND role = ?", p.UserID, domain.RoleSiswa).Error; err != nil {
		return nil, false
	}
	return &user, true
}
