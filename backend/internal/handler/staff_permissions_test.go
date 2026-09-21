package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"

	"cbt-backend/internal/domain"
	"cbt-backend/internal/repository"

	"github.com/glebarez/sqlite"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type staffEnv struct {
	app    *fiber.App
	db     *gorm.DB
	caller *domain.User
}

// newStaffEnv menyiapkan aplikasi Fiber minimal (tanpa JWT) yang menyuntikkan pemanggil
// dari env.caller ke Locals, meniru hasil middleware.AuthRequired.
func newStaffEnv(t *testing.T) *staffEnv {
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
	if err := db.AutoMigrate(
		&domain.User{}, &domain.ClassRoom{}, &domain.StudentProfile{}, &domain.Subject{},
		&domain.ClassSubject{}, &domain.QuestionBank{}, &domain.Question{}, &domain.ExamEvent{},
		&domain.ExamSchedule{}, &domain.ExamSession{}, &domain.StudentAnswer{},
		&domain.ViolationLog{}, &domain.ScheduleProctor{},
	); err != nil {
		t.Fatalf("auto-migrate gagal: %v", err)
	}

	env := &staffEnv{db: db}
	h := NewHandlers(&repository.Database{DB: db})
	app := fiber.New()
	app.Use(func(c *fiber.Ctx) error {
		if env.caller != nil {
			c.Locals("user", *env.caller)
		}
		return c.Next()
	})
	app.Get("/catalog", h.HandleGetPermissionCatalog)
	app.Post("/teachers", h.HandleCreateTeacher)
	app.Put("/teachers/:id", h.HandleUpdateTeacher)
	app.Post("/users", h.HandleCreateUser)
	app.Put("/users/:id", h.HandleUpdateUser)
	app.Delete("/teachers/:id", h.HandleDeleteTeacher)
	app.Delete("/users/:id", h.HandleDeleteUser)
	app.Put("/students/:id", h.HandleUpdateStudent)
	app.Delete("/students/:id", h.HandleDeleteStudent)
	env.app = app
	return env
}

func (e *staffEnv) mkUser(t *testing.T, name string, role domain.Role, perms []string) domain.User {
	t.Helper()
	u := domain.User{ID: uuid.New(), Username: name, PasswordHash: "x", FullName: name, Role: role, Permissions: perms, IsActive: true, CreatedAt: time.Now()}
	if err := e.db.Create(&u).Error; err != nil {
		t.Fatalf("gagal membuat user %s: %v", name, err)
	}
	return u
}

func (e *staffEnv) stored(t *testing.T, id uuid.UUID) domain.User {
	t.Helper()
	var u domain.User
	if err := e.db.First(&u, "id = ?", id).Error; err != nil {
		t.Fatalf("gagal membaca user: %v", err)
	}
	return u
}

func (e *staffEnv) storedByName(t *testing.T, name string) domain.User {
	t.Helper()
	var u domain.User
	if err := e.db.First(&u, "username = ?", name).Error; err != nil {
		t.Fatalf("gagal membaca user %s: %v", name, err)
	}
	return u
}

// do mengirim request JSON. body berupa string JSON mentah agar `permissions: []`
// dan permissions yang dihilangkan dapat dibedakan.
func (e *staffEnv) do(t *testing.T, method, path, body string) (int, map[string]interface{}) {
	t.Helper()
	req := httptest.NewRequest(method, path, bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := e.app.Test(req, -1)
	if err != nil {
		t.Fatalf("request gagal: %v", err)
	}
	defer resp.Body.Close()
	var out map[string]interface{}
	_ = json.NewDecoder(resp.Body).Decode(&out)
	return resp.StatusCode, out
}

func TestCreateTeacherWithoutPermissionsStoresGuruTemplate(t *testing.T) {
	e := newStaffEnv(t)
	admin := e.mkUser(t, "admin", domain.RoleAdmin, nil)
	e.caller = &admin

	status, _ := e.do(t, http.MethodPost, "/teachers", `{"username":"g1","full_name":"Guru Satu","role":"GURU"}`)
	if status != http.StatusOK {
		t.Fatalf("status = %d, ingin 200", status)
	}
	got := e.storedByName(t, "g1")
	if want := domain.TemplatePermissions("guru"); !reflect.DeepEqual(got.Permissions, want) {
		t.Fatalf("izin tersimpan = %v, ingin template guru %v", got.Permissions, want)
	}

	// Role dikosongkan -> default GURU, tetap eksplisit.
	status, _ = e.do(t, http.MethodPost, "/teachers", `{"username":"g2","full_name":"Guru Dua"}`)
	if status != http.StatusOK {
		t.Fatalf("status = %d, ingin 200", status)
	}
	if got := e.storedByName(t, "g2"); !reflect.DeepEqual(got.Permissions, domain.TemplatePermissions("guru")) {
		t.Fatalf("izin g2 = %v", got.Permissions)
	}
}

func TestCreateTeacherWithEmptyPermissionsIsRejected(t *testing.T) {
	e := newStaffEnv(t)
	admin := e.mkUser(t, "admin", domain.RoleAdmin, nil)
	e.caller = &admin

	for _, body := range []string{
		`{"username":"g1","full_name":"Guru","role":"GURU","permissions":[]}`,
		`{"username":"g1","full_name":"Guru","role":"GURU","permissions":["*","exam:take","bogus"]}`,
	} {
		status, resp := e.do(t, http.MethodPost, "/teachers", body)
		if status != http.StatusBadRequest {
			t.Fatalf("status = %d, ingin 400 untuk %s", status, body)
		}
		if resp["message"] != "Pilih minimal satu izin untuk akun guru" {
			t.Fatalf("pesan = %v", resp["message"])
		}
	}
	var n int64
	e.db.Model(&domain.User{}).Where("username = ?", "g1").Count(&n)
	if n != 0 {
		t.Fatalf("akun tidak boleh terbuat saat ditolak, ditemukan %d", n)
	}
}

func TestCreateTeacherWithPermissionsAppliesImplications(t *testing.T) {
	e := newStaffEnv(t)
	admin := e.mkUser(t, "admin", domain.RoleAdmin, nil)
	e.caller = &admin

	status, _ := e.do(t, http.MethodPost, "/teachers", `{"username":"g1","full_name":"Guru","role":"GURU","permissions":["schedules:manage","*"]}`)
	if status != http.StatusOK {
		t.Fatalf("status = %d, ingin 200", status)
	}
	want := []string{"schedules:manage", "schedules:read"}
	if got := e.storedByName(t, "g1"); !reflect.DeepEqual(got.Permissions, want) {
		t.Fatalf("izin = %v, ingin %v", got.Permissions, want)
	}
}

func TestUpdateTeacherAdminToGuruStoresExplicitTemplate(t *testing.T) {
	e := newStaffEnv(t)
	admin := e.mkUser(t, "admin", domain.RoleAdmin, nil)
	target := e.mkUser(t, "target", domain.RoleAdmin, []string{"*"})
	e.caller = &admin

	status, _ := e.do(t, http.MethodPut, "/teachers/"+target.ID.String(), `{"role":"GURU"}`)
	if status != http.StatusOK {
		t.Fatalf("status = %d, ingin 200", status)
	}
	got := e.stored(t, target.ID)
	if got.Role != domain.RoleGuru {
		t.Fatalf("role = %s, ingin GURU", got.Role)
	}
	if want := domain.TemplatePermissions("guru"); !reflect.DeepEqual(got.Permissions, want) {
		t.Fatalf("izin = %v, ingin template guru %v", got.Permissions, want)
	}
	if got.HasPermission("users:manage") {
		t.Fatalf("admin yang diturunkan tidak boleh mempertahankan izin penuh")
	}
}

func TestUpdateTeacherAdminToGuruWithPermissions(t *testing.T) {
	e := newStaffEnv(t)
	admin := e.mkUser(t, "admin", domain.RoleAdmin, nil)
	target := e.mkUser(t, "target", domain.RoleAdmin, nil)
	e.caller = &admin

	status, _ := e.do(t, http.MethodPut, "/teachers/"+target.ID.String(), `{"role":"GURU","permissions":["reports:export"]}`)
	if status != http.StatusOK {
		t.Fatalf("status = %d, ingin 200", status)
	}
	if got := e.stored(t, target.ID); !reflect.DeepEqual(got.Permissions, []string{"reports:export"}) {
		t.Fatalf("izin = %v", got.Permissions)
	}
}

func TestUpdateTeacherGuruToAdminClearsPermissions(t *testing.T) {
	e := newStaffEnv(t)
	admin := e.mkUser(t, "admin", domain.RoleAdmin, nil)
	target := e.mkUser(t, "target", domain.RoleGuru, []string{"reports:export"})
	e.caller = &admin

	status, _ := e.do(t, http.MethodPut, "/teachers/"+target.ID.String(), `{"role":"ADMIN"}`)
	if status != http.StatusOK {
		t.Fatalf("status = %d, ingin 200", status)
	}
	got := e.stored(t, target.ID)
	if got.Role != domain.RoleAdmin || len(got.Permissions) != 0 {
		t.Fatalf("role=%s izin=%v, ingin ADMIN tanpa izin tersimpan", got.Role, got.Permissions)
	}
	if !reflect.DeepEqual(got.EffectivePermissions(), []string{"*"}) {
		t.Fatalf("izin efektif admin = %v", got.EffectivePermissions())
	}
}

func TestUpdateTeacherWithEmptyPermissionsIsRejectedAndKeepsData(t *testing.T) {
	e := newStaffEnv(t)
	admin := e.mkUser(t, "admin", domain.RoleAdmin, nil)
	target := e.mkUser(t, "target", domain.RoleGuru, []string{"reports:export"})
	e.caller = &admin

	status, resp := e.do(t, http.MethodPut, "/teachers/"+target.ID.String(), `{"full_name":"Nama Baru","permissions":[]}`)
	if status != http.StatusBadRequest {
		t.Fatalf("status = %d, ingin 400", status)
	}
	if resp["message"] != "Pilih minimal satu izin untuk akun guru" {
		t.Fatalf("pesan = %v", resp["message"])
	}
	got := e.stored(t, target.ID)
	if got.FullName != "target" || !reflect.DeepEqual(got.Permissions, []string{"reports:export"}) {
		t.Fatalf("data berubah walau ditolak: nama=%q izin=%v", got.FullName, got.Permissions)
	}

	// Admin diturunkan ke GURU dengan daftar kosong juga ditolak.
	adminTarget := e.mkUser(t, "adm2", domain.RoleAdmin, nil)
	status, _ = e.do(t, http.MethodPut, "/teachers/"+adminTarget.ID.String(), `{"role":"GURU","permissions":[]}`)
	if status != http.StatusBadRequest {
		t.Fatalf("status = %d, ingin 400", status)
	}
	if got := e.stored(t, adminTarget.ID); got.Role != domain.RoleAdmin {
		t.Fatalf("role berubah walau ditolak: %s", got.Role)
	}
}

// Pemanggil tanpa izin penuh (bukan grantor) hanya boleh mengganti nama akun guru/staf.
func TestUpdateTeacherByNonGrantorMayOnlyChangeFullName(t *testing.T) {
	e := newStaffEnv(t)
	staf := e.mkUser(t, "staf", domain.RoleGuru, []string{"master:manage"})
	target := e.mkUser(t, "target", domain.RoleGuru, []string{"reports:export"})
	e.caller = &staf
	path := "/teachers/" + target.ID.String()
	oldHash := e.stored(t, target.ID).PasswordHash

	const msg = "Akses ditolak: hanya administrator yang dapat mengubah username dan kata sandi akun guru dan staf"
	for name, body := range map[string]string{
		"kata sandi":                 `{"password":"rahasia-baru"}`,
		"kata sandi dan nama":        `{"full_name":"Diubah","password":"rahasia-baru"}`,
		"username":                   `{"username":"pengambilalih"}`,
		"username dan nama":          `{"full_name":"Diubah","username":"pengambilalih"}`,
		"username sama, kata sandi":  `{"username":"target","password":"x"}`,
		"kata sandi bersama izin":    `{"password":"x","permissions":["reports:export"]}`,
		"username bersama role guru": `{"username":"pengambilalih","role":"GURU"}`,
	} {
		status, resp := e.do(t, http.MethodPut, path, body)
		if status != http.StatusForbidden {
			t.Fatalf("%s: status = %d, ingin 403", name, status)
		}
		if name == "kata sandi" || name == "username" {
			if resp["message"] != msg {
				t.Fatalf("%s: pesan = %v", name, resp["message"])
			}
		}
		got := e.stored(t, target.ID)
		if got.Username != "target" || got.FullName != "target" || got.PasswordHash != oldHash {
			t.Fatalf("%s: data berubah walau ditolak: username=%q nama=%q", name, got.Username, got.FullName)
		}
	}

	// Mengganti nama saja diperbolehkan, juga bila username dikirim tanpa berubah dan kata sandi kosong.
	status, _ := e.do(t, http.MethodPut, path, `{"full_name":"Nama Baru"}`)
	if status != http.StatusOK {
		t.Fatalf("ganti nama: status = %d, ingin 200", status)
	}
	status, _ = e.do(t, http.MethodPut, path, `{"full_name":"Nama Lagi","username":"target","password":""}`)
	if status != http.StatusOK {
		t.Fatalf("ganti nama dengan username sama: status = %d, ingin 200", status)
	}
	got := e.stored(t, target.ID)
	if got.FullName != "Nama Lagi" || got.Username != "target" || got.PasswordHash != oldHash {
		t.Fatalf("hasil: nama=%q username=%q", got.FullName, got.Username)
	}
	if !reflect.DeepEqual(got.Permissions, []string{"reports:export"}) {
		t.Fatalf("izin berubah: %v", got.Permissions)
	}
}

// Grantor (administrator) tetap bebas mengganti username dan kata sandi.
func TestUpdateTeacherByGrantorMayChangeUsernameAndPassword(t *testing.T) {
	e := newStaffEnv(t)
	admin := e.mkUser(t, "admin", domain.RoleAdmin, nil)
	target := e.mkUser(t, "target", domain.RoleGuru, []string{"reports:export"})
	e.caller = &admin

	status, _ := e.do(t, http.MethodPut, "/teachers/"+target.ID.String(), `{"username":"baru","password":"rahasia-baru","full_name":"Nama Baru"}`)
	if status != http.StatusOK {
		t.Fatalf("status = %d, ingin 200", status)
	}
	got := e.stored(t, target.ID)
	if got.Username != "baru" || got.FullName != "Nama Baru" || got.PasswordHash == "x" {
		t.Fatalf("data tidak berubah: username=%q nama=%q hash=%q", got.Username, got.FullName, got.PasswordHash)
	}
}

// Non-grantor yang mengirim permissions ditolak 403 di semua jalur, bukan diabaikan diam-diam.
func TestNonGrantorSendingPermissionsIsForbidden(t *testing.T) {
	e := newStaffEnv(t)
	staf := e.mkUser(t, "staf", domain.RoleGuru, []string{"master:manage", "users:manage"})
	guru := e.mkUser(t, "guru", domain.RoleGuru, []string{"reports:export"})
	siswa := e.mkUser(t, "siswa", domain.RoleSiswa, nil)
	e.caller = &staf

	const msg = "Akses ditolak: hanya administrator yang dapat mengatur izin akun"
	cases := []struct{ name, method, path, body string }{
		{"create teacher", http.MethodPost, "/teachers", `{"username":"g1","full_name":"G","role":"GURU","permissions":["users:manage"]}`},
		{"create teacher kosong", http.MethodPost, "/teachers", `{"username":"g1","full_name":"G","role":"GURU","permissions":[]}`},
		{"update teacher", http.MethodPut, "/teachers/" + guru.ID.String(), `{"permissions":["users:manage"]}`},
		{"update teacher kosong", http.MethodPut, "/teachers/" + guru.ID.String(), `{"permissions":[]}`},
		{"create user siswa", http.MethodPost, "/users", `{"username":"s1","full_name":"S","permissions":["users:manage"]}`},
		{"create user siswa kosong", http.MethodPost, "/users", `{"username":"s1","full_name":"S","permissions":[]}`},
		{"update user siswa", http.MethodPut, "/users/" + siswa.ID.String(), `{"full_name":"S2","permissions":["users:manage"]}`},
		{"update user siswa kosong", http.MethodPut, "/users/" + siswa.ID.String(), `{"permissions":[]}`},
	}
	for _, tc := range cases {
		status, resp := e.do(t, tc.method, tc.path, tc.body)
		if status != http.StatusForbidden {
			t.Fatalf("%s: status = %d, ingin 403", tc.name, status)
		}
		if resp["message"] != msg {
			t.Fatalf("%s: pesan = %v", tc.name, resp["message"])
		}
	}
	for _, name := range []string{"g1", "s1"} {
		var n int64
		e.db.Model(&domain.User{}).Where("username = ?", name).Count(&n)
		if n != 0 {
			t.Fatalf("akun %s tidak boleh terbuat", name)
		}
	}
	if got := e.stored(t, guru.ID); !reflect.DeepEqual(got.Permissions, []string{"reports:export"}) {
		t.Fatalf("izin guru berubah: %v", got.Permissions)
	}
	if got := e.stored(t, siswa.ID); got.FullName != "siswa" || len(got.Permissions) != 0 {
		t.Fatalf("siswa berubah: %+v", got)
	}
}

// Tanpa permissions, non-grantor tetap boleh: create -> template guru, update -> izin lama.
func TestNonGrantorWithoutPermissionsKeepsCurrentBehavior(t *testing.T) {
	e := newStaffEnv(t)
	staf := e.mkUser(t, "staf", domain.RoleGuru, []string{"master:manage"})
	target := e.mkUser(t, "target", domain.RoleGuru, []string{"reports:export"})
	e.caller = &staf

	status, _ := e.do(t, http.MethodPost, "/teachers", `{"username":"g1","full_name":"Guru","role":"GURU"}`)
	if status != http.StatusOK {
		t.Fatalf("create: status = %d, ingin 200", status)
	}
	if got := e.storedByName(t, "g1"); !reflect.DeepEqual(got.Permissions, domain.TemplatePermissions("guru")) {
		t.Fatalf("izin = %v, ingin template guru", got.Permissions)
	}

	status, _ = e.do(t, http.MethodPut, "/teachers/"+target.ID.String(), `{"full_name":"Diubah"}`)
	if status != http.StatusOK {
		t.Fatalf("update: status = %d, ingin 200", status)
	}
	got := e.stored(t, target.ID)
	if got.FullName != "Diubah" || !reflect.DeepEqual(got.Permissions, []string{"reports:export"}) {
		t.Fatalf("nama=%q izin=%v", got.FullName, got.Permissions)
	}
}

func TestUpdateTeacherByNonGrantorCannotChangeRoleOrTouchAdmin(t *testing.T) {
	e := newStaffEnv(t)
	staf := e.mkUser(t, "staf", domain.RoleGuru, []string{"master:manage"})
	target := e.mkUser(t, "target", domain.RoleGuru, []string{"reports:export"})
	adminTarget := e.mkUser(t, "adm", domain.RoleAdmin, nil)
	e.caller = &staf

	status, _ := e.do(t, http.MethodPut, "/teachers/"+target.ID.String(), `{"role":"ADMIN"}`)
	if status != http.StatusForbidden {
		t.Fatalf("ubah role: status = %d, ingin 403", status)
	}
	status, _ = e.do(t, http.MethodPut, "/teachers/"+adminTarget.ID.String(), `{"role":"GURU"}`)
	if status != http.StatusForbidden {
		t.Fatalf("ubah admin: status = %d, ingin 403", status)
	}
	status, _ = e.do(t, http.MethodPut, "/teachers/"+adminTarget.ID.String(), `{"full_name":"Diubah"}`)
	if status != http.StatusForbidden {
		t.Fatalf("ubah nama admin: status = %d, ingin 403", status)
	}
	if got := e.stored(t, adminTarget.ID); got.Role != domain.RoleAdmin || got.FullName != "adm" {
		t.Fatalf("akun admin berubah: %+v", got)
	}
	if got := e.stored(t, target.ID); got.Role != domain.RoleGuru {
		t.Fatalf("role guru berubah: %s", got.Role)
	}
}

// Data lama: GURU berdaftar ["*"] tidak boleh menjadi grantor (callerCanGrant).
func TestGuruWithStoredWildcardIsNotGrantor(t *testing.T) {
	e := newStaffEnv(t)
	fake := e.mkUser(t, "guru-bintang", domain.RoleGuru, []string{"*"})
	target := e.mkUser(t, "target", domain.RoleGuru, []string{"reports:export"})
	adminTarget := e.mkUser(t, "adm", domain.RoleAdmin, nil)
	e.caller = &fake

	tests := []struct{ name, method, path, body string }{
		{"ubah role", http.MethodPut, "/teachers/" + target.ID.String(), `{"role":"ADMIN"}`},
		{"ganti kata sandi", http.MethodPut, "/teachers/" + target.ID.String(), `{"password":"x"}`},
		{"ganti username", http.MethodPut, "/teachers/" + target.ID.String(), `{"username":"pengambilalih"}`},
		{"kirim izin", http.MethodPut, "/teachers/" + target.ID.String(), `{"permissions":["users:manage"]}`},
		{"ubah admin", http.MethodPut, "/teachers/" + adminTarget.ID.String(), `{"full_name":"x"}`},
		{"buat admin", http.MethodPost, "/teachers", `{"username":"a1","full_name":"A","role":"ADMIN"}`},
		{"buat guru dengan izin", http.MethodPost, "/teachers", `{"username":"g1","full_name":"G","role":"GURU","permissions":["users:manage"]}`},
		{"buat user guru", http.MethodPost, "/users", `{"username":"g2","full_name":"G","role":"GURU"}`},
	}
	for _, tt := range tests {
		if status, _ := e.do(t, tt.method, tt.path, tt.body); status != http.StatusForbidden {
			t.Errorf("%s: status = %d, ingin 403", tt.name, status)
		}
	}
	if got := e.stored(t, target.ID); got.Role != domain.RoleGuru || got.Username != "target" || !reflect.DeepEqual(got.Permissions, []string{"reports:export"}) {
		t.Fatalf("akun target berubah: %+v", got)
	}
}

func TestCreateUserGuruFollowsSameRules(t *testing.T) {
	e := newStaffEnv(t)
	admin := e.mkUser(t, "admin", domain.RoleAdmin, nil)
	e.caller = &admin

	status, _ := e.do(t, http.MethodPost, "/users", `{"username":"g1","full_name":"Guru","role":"GURU"}`)
	if status != http.StatusOK {
		t.Fatalf("status = %d, ingin 200", status)
	}
	if got := e.storedByName(t, "g1"); !reflect.DeepEqual(got.Permissions, domain.TemplatePermissions("guru")) {
		t.Fatalf("izin = %v, ingin template guru", got.Permissions)
	}

	status, resp := e.do(t, http.MethodPost, "/users", `{"username":"g2","full_name":"Guru","role":"GURU","permissions":[]}`)
	if status != http.StatusBadRequest || resp["message"] != "Pilih minimal satu izin untuk akun guru" {
		t.Fatalf("status=%d pesan=%v, ingin 400 dan pesan izin minimal", status, resp["message"])
	}

	status, _ = e.do(t, http.MethodPost, "/users", `{"username":"s1","full_name":"Siswa"}`)
	if status != http.StatusOK {
		t.Fatalf("status = %d, ingin 200", status)
	}
	if got := e.storedByName(t, "s1"); got.Role != domain.RoleSiswa || len(got.Permissions) != 0 {
		t.Fatalf("siswa: role=%s izin=%v", got.Role, got.Permissions)
	}
}

func TestUpdateUserFollowsSameRules(t *testing.T) {
	e := newStaffEnv(t)
	admin := e.mkUser(t, "admin", domain.RoleAdmin, nil)
	target := e.mkUser(t, "target", domain.RoleAdmin, nil)
	e.caller = &admin

	status, _ := e.do(t, http.MethodPut, "/users/"+target.ID.String(), `{"role":"GURU"}`)
	if status != http.StatusOK {
		t.Fatalf("status = %d, ingin 200", status)
	}
	if got := e.stored(t, target.ID); !reflect.DeepEqual(got.Permissions, domain.TemplatePermissions("guru")) {
		t.Fatalf("izin = %v, ingin template guru", got.Permissions)
	}

	status, resp := e.do(t, http.MethodPut, "/users/"+target.ID.String(), `{"permissions":[]}`)
	if status != http.StatusBadRequest || resp["message"] != "Pilih minimal satu izin untuk akun guru" {
		t.Fatalf("status=%d pesan=%v, ingin 400 dan pesan izin minimal", status, resp["message"])
	}
	if got := e.stored(t, target.ID); !reflect.DeepEqual(got.Permissions, domain.TemplatePermissions("guru")) {
		t.Fatalf("izin berubah walau ditolak: %v", got.Permissions)
	}

	status, _ = e.do(t, http.MethodPut, "/users/"+target.ID.String(), `{"permissions":["proctor:control_all"]}`)
	if status != http.StatusOK {
		t.Fatalf("status = %d, ingin 200", status)
	}
	want := []string{"proctor:control_all", "proctor:control"}
	if got := e.stored(t, target.ID); !reflect.DeepEqual(got.Permissions, want) {
		t.Fatalf("izin = %v, ingin %v", got.Permissions, want)
	}
}

func TestUpdateUserByNonAdminCannotTouchStaff(t *testing.T) {
	e := newStaffEnv(t)
	staf := e.mkUser(t, "staf", domain.RoleGuru, []string{"users:manage"})
	target := e.mkUser(t, "target", domain.RoleGuru, []string{"reports:export"})
	e.caller = &staf

	status, _ := e.do(t, http.MethodPut, "/users/"+target.ID.String(), `{"permissions":["users:manage"]}`)
	if status != http.StatusForbidden {
		t.Fatalf("status = %d, ingin 403", status)
	}
	if got := e.stored(t, target.ID); !reflect.DeepEqual(got.Permissions, []string{"reports:export"}) {
		t.Fatalf("izin berubah: %v", got.Permissions)
	}
}

func TestPermissionCatalogResponseShape(t *testing.T) {
	e := newStaffEnv(t)
	admin := e.mkUser(t, "admin", domain.RoleAdmin, nil)
	e.caller = &admin

	req := httptest.NewRequest(http.MethodGet, "/catalog", nil)
	resp, err := e.app.Test(req, -1)
	if err != nil {
		t.Fatalf("request gagal: %v", err)
	}
	defer resp.Body.Close()
	var body struct {
		Success bool `json:"success"`
		Data    struct {
			Groups []struct {
				Key   string `json:"key"`
				Items []struct {
					Key       string `json:"key"`
					Sensitive *bool  `json:"sensitive"`
				} `json:"items"`
			} `json:"groups"`
			Templates []struct {
				Key string `json:"key"`
			} `json:"templates"`
			Implies map[string][]string `json:"implies"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("gagal decode: %v", err)
	}
	if !body.Success || len(body.Data.Groups) == 0 {
		t.Fatalf("respons tidak berisi grup: %+v", body)
	}
	sensitive := map[string]bool{}
	for _, g := range body.Data.Groups {
		for _, it := range g.Items {
			if it.Sensitive == nil {
				t.Fatalf("item %q tidak memiliki field sensitive", it.Key)
			}
			sensitive[it.Key] = *it.Sensitive
		}
	}
	for _, k := range []string{"users:manage", "master:manage", "proctor:control_all", "questions:read_all"} {
		if !sensitive[k] {
			t.Errorf("izin %q harus sensitive=true", k)
		}
	}
	if sensitive["reports:export"] {
		t.Errorf("reports:export tidak boleh sensitif")
	}
	var keys []string
	for _, tpl := range body.Data.Templates {
		keys = append(keys, tpl.Key)
	}
	if !reflect.DeepEqual(keys, []string{"admin", "guru", "pengawas", "siswa"}) {
		t.Errorf("template = %v", keys)
	}
	if !reflect.DeepEqual(body.Data.Implies, domain.PermissionImplications()) {
		t.Errorf("implies = %v", body.Data.Implies)
	}
	if !reflect.DeepEqual(body.Data.Implies["schedules:manage"], []string{"schedules:read"}) {
		t.Errorf("implies[schedules:manage] = %v", body.Data.Implies["schedules:manage"])
	}
}
