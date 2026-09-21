package middleware

import (
	"strings"
	"testing"
)

func TestValidateJWTSecret(t *testing.T) {
	cases := []struct {
		name    string
		value   string
		wantErr bool
	}{
		{"kosong ditolak", "", true},
		{"nilai bawaan publik ditolak", publicDefaultJWTSecret, true},
		{"terlalu pendek ditolak", strings.Repeat("a", minJWTSecretLen-1), true},
		{"panjang minimum diterima", strings.Repeat("a", minJWTSecretLen), false},
		{"hex 64 karakter diterima", strings.Repeat("ab", 32), false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv("JWT_SECRET", tc.value)
			err := ValidateJWTSecret()
			if (err != nil) != tc.wantErr {
				t.Fatalf("ValidateJWTSecret() error = %v, wantErr %v", err, tc.wantErr)
			}
		})
	}
}

func TestJWTSecretNeverFallsBackToPublicDefault(t *testing.T) {
	t.Setenv("JWT_SECRET", "")
	if got := string(jwtSecret()); got == publicDefaultJWTSecret || got == "" {
		t.Fatalf("jwtSecret() tanpa env tidak boleh berupa nilai publik atau kosong")
	}
}
