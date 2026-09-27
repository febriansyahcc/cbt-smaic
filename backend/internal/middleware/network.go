package middleware

import (
	"encoding/json"
	"os"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/limiter"
)

// defaultTrustedProxies mencakup loopback dan jaringan privat, tempat kontainer nginx
// (jaringan Docker) atau Vite dev server berada. Backend tidak dipublikasikan langsung.
var defaultTrustedProxies = []string{"127.0.0.1", "::1", "10.0.0.0/8", "172.16.0.0/12", "192.168.0.0/16"}

// TrustedProxies membaca TRUSTED_PROXIES (dipisah koma, IP atau CIDR). Header X-Real-IP hanya
// dipercaya bila request datang dari alamat ini, sehingga klien tidak bisa memalsukan IP-nya.
func TrustedProxies() []string {
	raw := strings.TrimSpace(os.Getenv("TRUSTED_PROXIES"))
	if raw == "" {
		return defaultTrustedProxies
	}
	var out []string
	for _, p := range strings.Split(raw, ",") {
		if p = strings.TrimSpace(p); p != "" {
			out = append(out, p)
		}
	}
	return out
}

const (
	loginMaxFailures = 10
	loginWindow      = 10 * time.Minute
)

// LoginRateLimiter membatasi percobaan login gagal per pasangan IP + username. Kunci per pasangan
// (bukan per IP saja) agar satu kelas di balik NAT yang sama tidak ikut terblokir, dan (bukan per
// username saja) agar teman sekelas tidak bisa sengaja mengunci akun siswa lain dari perangkatnya.
// Login yang berhasil tidak dihitung.
func LoginRateLimiter() fiber.Handler {
	return limiter.New(limiter.Config{
		Max:                    loginMaxFailures,
		Expiration:             loginWindow,
		SkipSuccessfulRequests: true,
		KeyGenerator: func(c *fiber.Ctx) string {
			var body struct {
				Username string `json:"username"`
			}
			_ = json.Unmarshal(c.Body(), &body)
			return c.IP() + "|" + strings.ToLower(strings.TrimSpace(body.Username))
		},
		LimitReached: func(c *fiber.Ctx) error {
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"success": false,
				"message": "Terlalu banyak percobaan login yang gagal. Coba lagi dalam 10 menit atau hubungi pengawas.",
			})
		},
	})
}
