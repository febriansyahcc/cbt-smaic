package middleware

import (
	"errors"
	"os"
	"strings"
	"time"

	"cbt-backend/internal/domain"
	"cbt-backend/internal/repository"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

var jwtSecret = []byte(getSecret())

func getSecret() string {
	s := os.Getenv("JWT_SECRET")
	if s == "" {
		return "cbt-highschool-super-secret-key-2026"
	}
	return s
}

type JWTClaims struct {
	UserID      uuid.UUID   `json:"user_id"`
	Username    string      `json:"username"`
	Role        domain.Role `json:"role"`
	Permissions []string    `json:"permissions"`
	SessionID   string      `json:"sid"`
	jwt.RegisteredClaims
}

func GenerateToken(user domain.User, sessionID string) (string, error) {
	claims := JWTClaims{
		UserID:      user.ID,
		Username:    user.Username,
		Role:        user.Role,
		Permissions: user.EffectivePermissions(),
		SessionID:   sessionID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

func AuthRequired(db *repository.Database) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var tokenStr string
		authHeader := c.Get("Authorization")
		if authHeader != "" && strings.HasPrefix(authHeader, "Bearer ") {
			tokenStr = strings.TrimPrefix(authHeader, "Bearer ")
		} else if qToken := c.Query("token"); qToken != "" {
			tokenStr = qToken
		}

		if tokenStr == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"success": false,
				"message": "Akses ditolak: token autentikasi tidak ditemukan",
			})
		}
		token, err := jwt.ParseWithClaims(tokenStr, &JWTClaims{}, func(t *jwt.Token) (interface{}, error) {
			return jwtSecret, nil
		})

		if err != nil || !token.Valid {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"success": false,
				"message": "Token tidak valid atau telah kedaluwarsa",
			})
		}

		claims, ok := token.Claims.(*JWTClaims)
		if !ok {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"success": false,
				"message": "Klaim token tidak valid",
			})
		}

		// Single-device active check
		var user domain.User
		if err := db.DB.First(&user, "id = ?", claims.UserID).Error; err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"success": false,
				"message": "Pengguna tidak ditemukan",
			})
		}

		if !user.IsActive {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"success": false,
				"message": "Akun Anda dinonaktifkan oleh administrator",
			})
		}

		// For students, check single device session token
		if user.Role == domain.RoleSiswa && user.SessionToken != "" && user.SessionToken != claims.SessionID {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"success": false,
				"code":    "CONCURRENT_LOGIN",
				"message": "Sesi berakhir. Akun Anda telah digunakan untuk login di perangkat lain.",
			})
		}

		// Set locals
		c.Locals("user", user)
		c.Locals("user_id", user.ID)
		c.Locals("role", user.Role)
		c.Locals("permissions", user.EffectivePermissions())

		return c.Next()
	}
}

func RequirePermission(requiredPerms ...string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		user, ok := c.Locals("user").(domain.User)
		if !ok {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"success": false,
				"message": "Autentikasi diperlukan",
			})
		}

		if user.Role == domain.RoleAdmin {
			return c.Next()
		}

		for _, perm := range requiredPerms {
			if user.HasPermission(perm) {
				return c.Next()
			}
		}

		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"success": false,
			"message": "Akses ditolak: Anda tidak memiliki izin untuk operasi ini",
		})
	}
}

// StaffOnly meloloskan pengguna yang memiliki minimal satu izin staf (selain exam:take).
// Dipakai sebagai pintu masuk grup route staf; izin spesifik dicek per route lewat RequirePermission.
func StaffOnly() fiber.Handler {
	return func(c *fiber.Ctx) error {
		user, ok := c.Locals("user").(domain.User)
		if !ok {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"success": false,
				"message": "Autentikasi diperlukan",
			})
		}
		if user.IsStaff() {
			return c.Next()
		}
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"success": false,
			"message": "Anda tidak memiliki hak akses untuk operasi ini",
		})
	}
}

func RoleGuard(allowedRoles ...domain.Role) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userRole, ok := c.Locals("role").(domain.Role)
		if !ok {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"success": false,
				"message": "Autentikasi diperlukan",
			})
		}

		for _, r := range allowedRoles {
			if r == userRole {
				return c.Next()
			}
		}

		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"success": false,
			"message": "Anda tidak memiliki hak akses untuk operasi ini",
		})
	}
}

func GetCurrentUser(c *fiber.Ctx) (domain.User, error) {
	u, ok := c.Locals("user").(domain.User)
	if !ok {
		return domain.User{}, errors.New("user not found in context")
	}
	return u, nil
}
