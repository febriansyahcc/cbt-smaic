package repository

import (
	"reflect"
	"testing"
	"time"

	"cbt-backend/internal/domain"

	"github.com/glebarez/sqlite"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

func newBackfillDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("gagal membuka sqlite memori: %v", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("gagal mengambil sql.DB: %v", err)
	}
	sqlDB.SetMaxOpenConns(1)
	if err := db.AutoMigrate(&domain.User{}); err != nil {
		t.Fatalf("auto-migrate gagal: %v", err)
	}
	return db
}

func mkBackfillUser(t *testing.T, db *gorm.DB, name string, role domain.Role, perms []string) uuid.UUID {
	t.Helper()
	u := domain.User{ID: uuid.New(), Username: name, PasswordHash: "x", FullName: name, Role: role, Permissions: perms, IsActive: true, CreatedAt: time.Now()}
	if err := db.Create(&u).Error; err != nil {
		t.Fatalf("gagal membuat user %s: %v", name, err)
	}
	return u.ID
}

func storedPermissions(t *testing.T, db *gorm.DB, id uuid.UUID) []string {
	t.Helper()
	var u domain.User
	if err := db.First(&u, "id = ?", id).Error; err != nil {
		t.Fatalf("gagal membaca user: %v", err)
	}
	return u.Permissions
}

func TestBackfillStaffPermissions(t *testing.T) {
	db := newBackfillDB(t)
	custom := []string{"reports:export"}
	tpl := domain.TemplatePermissions("guru")

	guruNil := mkBackfillUser(t, db, "guru-nil", domain.RoleGuru, nil)
	guruEmpty := mkBackfillUser(t, db, "guru-empty", domain.RoleGuru, []string{})
	guruCustom := mkBackfillUser(t, db, "guru-custom", domain.RoleGuru, custom)
	admin := mkBackfillUser(t, db, "admin", domain.RoleAdmin, nil)
	adminPerms := mkBackfillUser(t, db, "admin-perms", domain.RoleAdmin, []string{"*", "bogus"})
	siswa := mkBackfillUser(t, db, "siswa", domain.RoleSiswa, nil)
	// Baris lama dengan kolom NULL murni (tidak melewati serializer).
	guruNull := mkBackfillUser(t, db, "guru-null", domain.RoleGuru, custom)
	if err := db.Exec("UPDATE users SET permissions = NULL WHERE id = ?", guruNull).Error; err != nil {
		t.Fatalf("gagal mengosongkan kolom: %v", err)
	}

	n, err := BackfillStaffPermissions(db)
	if err != nil {
		t.Fatalf("backfill gagal: %v", err)
	}
	if n != 3 {
		t.Fatalf("jumlah akun diperbarui = %d, ingin 3", n)
	}

	for name, id := range map[string]uuid.UUID{"nil": guruNil, "empty": guruEmpty, "null": guruNull} {
		if got := storedPermissions(t, db, id); !reflect.DeepEqual(got, tpl) {
			t.Errorf("guru-%s: izin = %v, ingin template guru %v", name, got, tpl)
		}
	}
	if got := storedPermissions(t, db, guruCustom); !reflect.DeepEqual(got, custom) {
		t.Errorf("izin kustom yang bersih berubah: %v", got)
	}
	if got := storedPermissions(t, db, admin); len(got) != 0 {
		t.Errorf("izin admin tidak boleh diisi: %v", got)
	}
	if got := storedPermissions(t, db, adminPerms); !reflect.DeepEqual(got, []string{"*", "bogus"}) {
		t.Errorf("admin tidak boleh disentuh: %v", got)
	}
	if got := storedPermissions(t, db, siswa); len(got) != 0 {
		t.Errorf("izin siswa tidak boleh diisi: %v", got)
	}

	// Idempotent: jalan kedua tidak mengubah apa pun.
	n, err = BackfillStaffPermissions(db)
	if err != nil || n != 0 {
		t.Fatalf("jalan kedua: n=%d err=%v, ingin 0 tanpa galat", n, err)
	}
}

// Daftar GURU yang tidak bersih disanitasi ulang: "*", exam:take, dan kunci tak dikenal dibuang,
// izin turunan ditambahkan; bila tidak ada yang tersisa diisi template guru.
func TestBackfillStaffPermissionsSanitizesDirtyGuru(t *testing.T) {
	db := newBackfillDB(t)
	tpl := domain.TemplatePermissions("guru")

	wildcard := mkBackfillUser(t, db, "guru-wildcard", domain.RoleGuru, []string{"*"})
	examOnly := mkBackfillUser(t, db, "guru-exam", domain.RoleGuru, []string{"exam:take"})
	junk := mkBackfillUser(t, db, "guru-junk", domain.RoleGuru, []string{"bogus", "*", "exam:take"})
	mixed := mkBackfillUser(t, db, "guru-mixed", domain.RoleGuru, []string{"*", "reports:export", "bogus", "reports:export", "exam:take"})
	needsImplied := mkBackfillUser(t, db, "guru-implied", domain.RoleGuru, []string{"schedules:manage"})
	clean := mkBackfillUser(t, db, "guru-clean", domain.RoleGuru, []string{"proctor:control", "reports:export"})

	n, err := BackfillStaffPermissions(db)
	if err != nil {
		t.Fatalf("backfill gagal: %v", err)
	}
	if n != 5 {
		t.Fatalf("jumlah akun diperbarui = %d, ingin 5", n)
	}

	want := map[string][]string{
		"wildcard": tpl,
		"exam":     tpl,
		"junk":     tpl,
		"mixed":    {"reports:export"},
		"implied":  {"schedules:manage", "schedules:read"},
		"clean":    {"proctor:control", "reports:export"},
	}
	ids := map[string]uuid.UUID{"wildcard": wildcard, "exam": examOnly, "junk": junk, "mixed": mixed, "implied": needsImplied, "clean": clean}
	for name, id := range ids {
		if got := storedPermissions(t, db, id); !reflect.DeepEqual(got, want[name]) {
			t.Errorf("guru-%s: izin = %v, ingin %v", name, got, want[name])
		}
	}

	// Hasil akhir sudah setara dengan izin efektifnya dan tidak pernah memuat "*".
	for name, id := range ids {
		var u domain.User
		if err := db.First(&u, "id = ?", id).Error; err != nil {
			t.Fatalf("gagal membaca user: %v", err)
		}
		if !reflect.DeepEqual(u.Permissions, u.EffectivePermissions()) {
			t.Errorf("guru-%s: tersimpan %v berbeda dari efektif %v", name, u.Permissions, u.EffectivePermissions())
		}
		if u.HasPermission("*") {
			t.Errorf("guru-%s masih memegang izin penuh", name)
		}
	}

	// Idempotent.
	if n, err := BackfillStaffPermissions(db); err != nil || n != 0 {
		t.Fatalf("jalan kedua: n=%d err=%v, ingin 0 tanpa galat", n, err)
	}
}

// Siswa yang punya daftar izin tersimpan dikosongkan; role lain tidak ikut terpengaruh.
func TestBackfillStaffPermissionsClearsSiswaPermissions(t *testing.T) {
	db := newBackfillDB(t)

	dirty := mkBackfillUser(t, db, "siswa-dirty", domain.RoleSiswa, []string{"*", "users:manage"})
	cleanSiswa := mkBackfillUser(t, db, "siswa-clean", domain.RoleSiswa, nil)
	guru := mkBackfillUser(t, db, "guru", domain.RoleGuru, []string{"reports:export"})

	n, err := BackfillStaffPermissions(db)
	if err != nil {
		t.Fatalf("backfill gagal: %v", err)
	}
	if n != 1 {
		t.Fatalf("jumlah akun diperbarui = %d, ingin 1", n)
	}
	if got := storedPermissions(t, db, dirty); len(got) != 0 {
		t.Errorf("izin siswa seharusnya dikosongkan: %v", got)
	}
	if got := storedPermissions(t, db, cleanSiswa); len(got) != 0 {
		t.Errorf("siswa bersih berubah: %v", got)
	}
	if got := storedPermissions(t, db, guru); !reflect.DeepEqual(got, []string{"reports:export"}) {
		t.Errorf("guru berubah: %v", got)
	}

	var u domain.User
	if err := db.First(&u, "id = ?", dirty).Error; err != nil {
		t.Fatalf("gagal membaca user: %v", err)
	}
	if !reflect.DeepEqual(u.EffectivePermissions(), []string{"exam:take"}) {
		t.Errorf("izin efektif siswa = %v, ingin [exam:take]", u.EffectivePermissions())
	}

	if n, err := BackfillStaffPermissions(db); err != nil || n != 0 {
		t.Fatalf("jalan kedua: n=%d err=%v, ingin 0 tanpa galat", n, err)
	}
}
