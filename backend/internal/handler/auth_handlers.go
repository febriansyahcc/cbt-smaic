package handler

import (
	"cbt-backend/internal/domain"
	"cbt-backend/internal/middleware"

	"github.com/gofiber/fiber/v2"
)

// ---------------- AUTH HANDLERS ----------------

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func (h *Handlers) HandleLogin(c *fiber.Ctx) error {
	var req LoginRequest
	if err := c.BodyParser(&req); err != nil || req.Username == "" || req.Password == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Username dan password wajib diisi",
		})
	}

	ip := c.IP()
	userAgent := c.Get("User-Agent")

	res, err := h.authService.Login(req.Username, req.Password, ip, userAgent)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"message": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    res,
	})
}

func (h *Handlers) HandleGetMe(c *fiber.Ctx) error {
	user, err := middleware.GetCurrentUser(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"success": false, "message": "Unauthorized"})
	}

	user.Permissions = user.EffectivePermissions()

	resp := fiber.Map{
		"user":        user,
		"permissions": user.EffectivePermissions(),
	}

	if user.Role == domain.RoleSiswa {
		var profile domain.StudentProfile
		if err := h.repo.DB.Preload("ClassRoom").Where("user_id = ?", user.ID).First(&profile).Error; err == nil {
			resp["student_profile"] = profile
		}
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    resp,
	})
}

func (h *Handlers) HandleLogout(c *fiber.Ctx) error {
	user, _ := middleware.GetCurrentUser(c)
	h.repo.DB.Model(&user).Update("session_token", "")
	return c.JSON(fiber.Map{
		"success": true,
		"message": "Logout berhasil",
	})
}
