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
}

func NewHandlers(repo *repository.Database) *Handlers {
	return &Handlers{
		repo:           repo,
		authService:    service.NewAuthService(repo),
		examService:    service.NewExamService(repo),
		proctorService: service.NewProctorService(repo),
		accessService:  service.NewAccessService(repo),
		participants:   service.NewEventParticipantService(repo),
	}
}
