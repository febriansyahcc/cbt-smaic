package repository

import (
	"errors"
	"fmt"
	"regexp"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Open membuka koneksi PostgreSQL dari DSN tanpa efek samping apa pun: tidak ada
// AutoMigrate, seed, backfill, dan tidak ada fallback ke SQLite. Dipakai alat CLI
// (mis. cmd/seedstaff) yang tidak boleh mengubah skema atau data sebelum diminta.
//
// Logger GORM dimatikan karena logger bawaan menulis ke stdout (mengotori keluaran CSV)
// dan pada galat dapat menampilkan nilai kolom. Galat tetap dikembalikan lewat error.
func Open(dsn string) (*gorm.DB, error) {
	if dsn == "" {
		return nil, errors.New("DB_DSN kosong")
	}
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		// Pesan galat driver dapat memuat DSN; buang rahasianya sebelum dikembalikan.
		return nil, fmt.Errorf("gagal terhubung ke PostgreSQL: %s", RedactSecrets(err.Error()))
	}
	return db, nil
}

var (
	// postgres://user:password@host/db  ->  postgres://user:REDACTED@host/db
	reDSNURLPassword = regexp.MustCompile(`(://[^:/@\s]*:)\S*@`)
	// password=rahasia atau password='ra hasia' (juga sslpassword=...)
	reDSNKVPassword = regexp.MustCompile(`(?i)(password\s*=\s*)('(?:[^'\\]|\\.)*'|\S+)`)
	// ?password=rahasia atau &password=rahasia pada DSN berbentuk URL sudah tercakup reDSNKVPassword.
)

// RedactSecrets menyamarkan kata sandi pada DSN (bentuk URL maupun key=value) atau pada
// teks lain yang memuat DSN, sehingga aman dicetak ke log.
func RedactSecrets(s string) string {
	s = reDSNURLPassword.ReplaceAllString(s, "${1}REDACTED@")
	s = reDSNKVPassword.ReplaceAllString(s, "${1}REDACTED")
	return s
}
