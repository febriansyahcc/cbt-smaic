package middleware

import (
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
)

func newLoginApp() *fiber.App {
	app := fiber.New()
	app.Post("/login", LoginRateLimiter(), func(c *fiber.Ctx) error {
		if strings.Contains(string(c.Body()), `"password":"benar"`) {
			return c.SendStatus(fiber.StatusOK)
		}
		return c.SendStatus(fiber.StatusUnauthorized)
	})
	return app
}

func postLogin(t *testing.T, app *fiber.App, username, password string) int {
	t.Helper()
	req := httptest.NewRequest("POST", "/login", strings.NewReader(`{"username":"`+username+`","password":"`+password+`"}`))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	return resp.StatusCode
}

func TestLoginRateLimiterBlocksRepeatedFailures(t *testing.T) {
	app := newLoginApp()
	for i := 0; i < loginMaxFailures; i++ {
		if code := postLogin(t, app, "siswa1", "salah"); code != fiber.StatusUnauthorized {
			t.Fatalf("percobaan ke-%d berstatus %d, ingin 401", i+1, code)
		}
	}
	if code := postLogin(t, app, "siswa1", "salah"); code != fiber.StatusTooManyRequests {
		t.Fatalf("percobaan setelah batas berstatus %d, ingin 429", code)
	}
	// Akun lain dari perangkat yang sama tidak ikut terblokir.
	if code := postLogin(t, app, "Siswa2", "benar"); code != fiber.StatusOK {
		t.Fatalf("login akun lain berstatus %d, ingin 200", code)
	}
}

func TestLoginRateLimiterIgnoresSuccessfulLogins(t *testing.T) {
	app := newLoginApp()
	for i := 0; i < loginMaxFailures+5; i++ {
		if code := postLogin(t, app, "siswa1", "benar"); code != fiber.StatusOK {
			t.Fatalf("login berhasil ke-%d berstatus %d", i+1, code)
		}
	}
}

func TestTrustedProxiesFromEnv(t *testing.T) {
	t.Setenv("TRUSTED_PROXIES", "")
	if got := TrustedProxies(); len(got) != len(defaultTrustedProxies) {
		t.Fatalf("bawaan = %v", got)
	}
	t.Setenv("TRUSTED_PROXIES", " 10.1.0.0/16 , 127.0.0.1,")
	if got := TrustedProxies(); len(got) != 2 || got[0] != "10.1.0.0/16" || got[1] != "127.0.0.1" {
		t.Fatalf("TRUSTED_PROXIES = %v", got)
	}
}
