package staffseed

import (
	"fmt"
	"io"
)

const passwordLength = 10

// Alfabet tanpa karakter yang mudah tertukar: 0 O 1 l I.
const passwordAlphabet = "ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz23456789"

// GeneratePassword membuat kata sandi acak sepanjang 10 karakter dari rng (crypto/rand pada
// pemakaian nyata). Byte acak dibuang bila jatuh di luar kelipatan panjang alfabet
// (rejection sampling) supaya setiap karakter berpeluang sama.
func GeneratePassword(rng io.Reader) (string, error) {
	n := len(passwordAlphabet)
	limit := 256 - (256 % n) // byte >= limit dibuang
	out := make([]byte, 0, passwordLength)
	buf := make([]byte, 32)
	for len(out) < passwordLength {
		if _, err := io.ReadFull(rng, buf); err != nil {
			return "", fmt.Errorf("gagal membaca sumber acak: %w", err)
		}
		for _, b := range buf {
			if int(b) >= limit {
				continue
			}
			out = append(out, passwordAlphabet[int(b)%n])
			if len(out) == passwordLength {
				break
			}
		}
	}
	return string(out), nil
}
