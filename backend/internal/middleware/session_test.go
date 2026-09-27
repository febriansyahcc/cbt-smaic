package middleware

import (
	"encoding/json"
	"net/http/httptest"
	"testing"

	"cbt-backend/internal/domain"
	"cbt-backend/internal/repository"

	"github.com/glebarez/sqlite"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

func newSessionTestApp(t *testing.T) (*fiber.App, *gorm.DB) {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("gagal membuka sqlite memori: %v", err)
	}
	sqlDB, _ := db.DB()
	sqlDB.SetMaxOpenConns(1)
	if err := db.AutoMigrate(&domain.User{}); err != nil {
		t.Fatalf("migrasi gagal: %v", err)
	}
	app := fiber.New()
	app.Get("/me", AuthRequired(&repository.Database{DB: db}), func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusOK)
	})
	return app, db
}

func createUser(t *testing.T, db *gorm.DB, role domain.Role, sessionToken string) domain.User {
	t.Helper()
	u := domain.User{
		ID:           uuid.New(),
		Username:     "u-" + uuid.NewString()[:8],
		FullName:     "Pengguna Uji",
		Role:         role,
		IsActive:     true,
		SessionToken: sessionToken,
	}
	if err := db.Create(&u).Error; err != nil {
		t.Fatalf("gagal membuat user: %v", err)
	}
	return u
}

func callMe(t *testing.T, app *fiber.App, user domain.User, sid string) (int, string) {
	t.Helper()
	tok, err := GenerateToken(user, sid)
	if err != nil {
		t.Fatalf("GenerateToken: %v", err)
	}
	req := httptest.NewRequest("GET", "/me", nil)
	req.Header.Set("Authorization", "Bearer "+tok)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("request gagal: %v", err)
	}
	var body struct {
		Code string `json:"code"`
	}
	_ = json.NewDecoder(resp.Body).Decode(&body)
	return resp.StatusCode, body.Code
}

func TestStudentSingleDeviceSession(t *testing.T) {
	app, db := newSessionTestApp(t)

	cases := []struct {
		name       string
		role       domain.Role
		stored     string
		claimSID   string
		wantStatus int
		wantCode   string
	}{
		{"sesi cocok diterima", domain.RoleSiswa, "sid-baru", "sid-baru", 200, ""},
		{"login di perangkat lain menendang sesi lama", domain.RoleSiswa, "sid-baru", "sid-lama", 401, "CONCURRENT_LOGIN"},
		{"sesi direset menolak token lama", domain.RoleSiswa, "", "sid-lama", 401, "SESSION_ENDED"},
		{"sesi direset menolak token tanpa sid", domain.RoleSiswa, "", "", 401, "SESSION_ENDED"},
		{"staf tidak terkena kunci satu perangkat", domain.RoleGuru, "", "apa-saja", 200, ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			u := createUser(t, db, tc.role, tc.stored)
			status, code := callMe(t, app, u, tc.claimSID)
			if status != tc.wantStatus || code != tc.wantCode {
				t.Fatalf("status=%d code=%q; ingin status=%d code=%q", status, code, tc.wantStatus, tc.wantCode)
			}
		})
	}
}
