package service

import (
	"errors"

	"cbt-backend/internal/domain"
	"cbt-backend/internal/middleware"
	"cbt-backend/internal/repository"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	repo *repository.Database
}

func NewAuthService(repo *repository.Database) *AuthService {
	return &AuthService{repo: repo}
}

type LoginResponse struct {
	Token   string                 `json:"token"`
	User    domain.User            `json:"user"`
	Profile *domain.StudentProfile `json:"student_profile,omitempty"`
}

func (s *AuthService) Login(username, password, clientIP, userAgent string) (*LoginResponse, error) {
	var user domain.User
	if err := s.repo.DB.Where("username = ?", username).First(&user).Error; err != nil {
		return nil, errors.New("username atau password salah")
	}

	if !user.IsActive {
		return nil, errors.New("akun dinonaktifkan")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, errors.New("username atau password salah")
	}

	// Generate new session ID for single-device lock
	sessionID := uuid.New().String()
	user.SessionToken = sessionID
	s.repo.DB.Model(&user).Update("session_token", sessionID)

	token, err := middleware.GenerateToken(user, sessionID)
	if err != nil {
		return nil, errors.New("gagal menerbitkan token autentikasi")
	}

	user.Permissions = user.EffectivePermissions()

	res := &LoginResponse{
		Token: token,
		User:  user,
	}

	if user.Role == domain.RoleSiswa {
		var profile domain.StudentProfile
		if err := s.repo.DB.Preload("ClassRoom").Where("user_id = ?", user.ID).First(&profile).Error; err == nil {
			res.Profile = &profile
		}
	}

	return res, nil
}
