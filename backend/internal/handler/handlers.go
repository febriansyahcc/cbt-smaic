package handler

import (
	"cbt-backend/internal/repository"
	"cbt-backend/internal/service"
)

type Handlers struct {
	repo           *repository.Database
	authService    *service.AuthService
	examService    *service.ExamService
	proctorService *service.ProctorService
	accessService  *service.AccessService
	participants   *service.EventParticipantService
	backupService  *service.BackupService
}

// BackupService dipakai main.go untuk menjalankan penjadwal backup otomatis.
func (h *Handlers) BackupService() *service.BackupService {
	return h.backupService
}

// ExamService dipakai main.go untuk menjalankan sweeper sesi ujian kedaluwarsa.
func (h *Handlers) ExamService() *service.ExamService {
	return h.examService
}

func NewHandlers(repo *repository.Database) *Handlers {
	return &Handlers{
		repo:           repo,
		authService:    service.NewAuthService(repo),
		examService:    service.NewExamService(repo),
		proctorService: service.NewProctorService(repo),
		accessService:  service.NewAccessService(repo),
		participants:   service.NewEventParticipantService(repo),
		backupService:  service.NewBackupService(repo),
	}
}
