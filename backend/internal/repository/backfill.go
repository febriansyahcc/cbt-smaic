package repository

import (
	"log"
	"reflect"
	"time"

	"cbt-backend/internal/domain"

	"gorm.io/gorm"
)

// BackfillStaffPermissions merapikan kolom users.permissions agar cocok dengan model
// "izin selalu eksplisit":
//
//   - GURU dengan daftar kosong (NULL, "", "null", "[]") diisi TemplatePermissions("guru").
//   - GURU dengan daftar tidak kosong disanitasi ulang (SanitizePermissions): "*", exam:take,
//     dan kunci tak dikenal dibuang, izin turunan ditambahkan. Hasil ditulis balik hanya bila
//     berbeda; bila hasil sanitasi kosong, diisi template guru.
//   - SISWA dengan daftar terisi dikosongkan (nil). Izin siswa selalu dibaca dari template
//     (EffectivePermissions mengabaikan daftar tersimpan), jadi mengosongkannya hanya
//     membersihkan data dan aman dijalankan ulang.
//   - ADMIN tidak disentuh (selalu "*").
//
// Idempotent: setelah dijalankan, jalan berikutnya tidak mengubah apa pun, sehingga aman pada
// setiap startup. Mengembalikan jumlah akun yang diperbarui. Fallback di EffectivePermissions
// tetap ada sebagai pengaman untuk baris yang belum sempat dirapikan.
func BackfillStaffPermissions(db *gorm.DB) (int, error) {
	// Kosong ditentukan setelah deserialisasi agar tidak bergantung pada representasi
	// NULL/"null"/"[]" yang berbeda antar SQLite dan PostgreSQL.
	var users []domain.User
	if err := db.Select("id", "role", "permissions").
		Where("role IN ?", []domain.Role{domain.RoleGuru, domain.RoleSiswa}).
		Find(&users).Error; err != nil {
		return 0, err
	}

	updated := 0
	for i := range users {
		u := &users[i]
		var want []string
		switch u.Role {
		case domain.RoleGuru:
			want = domain.TemplatePermissions("guru")
			if len(u.Permissions) > 0 {
				if sanitized := domain.SanitizePermissions(u.Role, u.Permissions); len(sanitized) > 0 {
					want = sanitized
				}
			}
			if reflect.DeepEqual(want, u.Permissions) {
				continue
			}
		case domain.RoleSiswa:
			if len(u.Permissions) == 0 {
				continue
			}
			want = nil
		default:
			continue
		}

		u.Permissions = want
		if err := db.Model(u).Select("Permissions").Updates(u).Error; err != nil {
			return updated, err
		}
		updated++
	}
	return updated, nil
}

// BackfillStaffPermissions dijalankan saat startup setelah migrasi dan seed.
func (d *Database) BackfillStaffPermissions() {
	n, err := BackfillStaffPermissions(d.DB)
	if err != nil {
		log.Printf("Gagal merapikan izin akun guru/siswa: %v", err)
		return
	}
	if n > 0 {
		log.Printf("Izin dirapikan untuk %d akun (guru: eksplisit dan tersanitasi, siswa: dikosongkan)", n)
	}
}

// AppMigration mencatat migrasi data satu kali yang sudah dijalankan, agar tidak diulang
// pada startup berikutnya (mis. izin yang sengaja dicabut admin tidak diberikan kembali).
type AppMigration struct {
	Name      string `gorm:"size:100;primaryKey"`
	AppliedAt time.Time
}

const migrationGrantQuestionsPrint = "grant_questions_print_v1"

// GrantQuestionsPrintOnce memberi izin questions:print kepada akun GURU yang sudah dapat
// membuka bank soal saat izin itu diperkenalkan. Dijalankan sekali; sesudahnya izin cetak
// sepenuhnya diatur lewat editor akses. Mengembalikan jumlah akun yang diperbarui.
func GrantQuestionsPrintOnce(db *gorm.DB) (int, error) {
	if err := db.AutoMigrate(&AppMigration{}); err != nil {
		return 0, err
	}
	var done int64
	if err := db.Model(&AppMigration{}).Where("name = ?", migrationGrantQuestionsPrint).Count(&done).Error; err != nil {
		return 0, err
	}
	if done > 0 {
		return 0, nil
	}

	bankPerms := map[string]bool{
		string(domain.PermQuestionsAssigned): true,
		string(domain.PermQuestionsAll):      true,
		string(domain.PermQuestionsUpload):   true,
		string(domain.PermQuestionsManage):   true,
		string(domain.PermQuestionsLock):     true,
	}

	updated := 0
	err := db.Transaction(func(tx *gorm.DB) error {
		var users []domain.User
		if err := tx.Select("id", "role", "permissions").Where("role = ?", domain.RoleGuru).Find(&users).Error; err != nil {
			return err
		}
		for i := range users {
			u := &users[i]
			canRead, hasPrint := false, false
			for _, p := range u.Permissions {
				canRead = canRead || bankPerms[p]
				hasPrint = hasPrint || p == string(domain.PermQuestionsPrint)
			}
			if !canRead || hasPrint {
				continue
			}
			u.Permissions = domain.SanitizePermissions(u.Role, append(u.Permissions, string(domain.PermQuestionsPrint)))
			if err := tx.Model(u).Select("Permissions").Updates(u).Error; err != nil {
				return err
			}
			updated++
		}
		return tx.Create(&AppMigration{Name: migrationGrantQuestionsPrint, AppliedAt: time.Now()}).Error
	})
	return updated, err
}

// GrantQuestionsPrintOnce dijalankan saat startup setelah BackfillStaffPermissions.
func (d *Database) GrantQuestionsPrintOnce() {
	n, err := GrantQuestionsPrintOnce(d.DB)
	if err != nil {
		log.Printf("Gagal memberi izin cetak soal: %v", err)
		return
	}
	if n > 0 {
		log.Printf("Izin cetak soal diberikan kepada %d akun yang dapat membuka bank soal", n)
	}
}
