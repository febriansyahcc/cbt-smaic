package handler

import (
	"net/http"
	"reflect"
	"testing"
	"time"

	"cbt-backend/internal/domain"

	"github.com/google/uuid"
)

func (e *staffEnv) mkStudent(t *testing.T, name, nis, nisn string) domain.User {
	t.Helper()
	u := e.mkUser(t, name, domain.RoleSiswa, nil)
	var class domain.ClassRoom
	if err := e.db.First(&class).Error; err != nil {
		class = domain.ClassRoom{ID: uuid.New(), Name: "XII-A", Grade: "XII", CreatedAt: time.Now()}
		if err := e.db.Create(&class).Error; err != nil {
			t.Fatalf("gagal membuat kelas: %v", err)
		}
	}
	p := domain.StudentProfile{ID: uuid.New(), UserID: u.ID, NIS: nis, NISN: nisn, ClassRoomID: class.ID, CreatedAt: time.Now()}
	if err := e.db.Create(&p).Error; err != nil {
		t.Fatalf("gagal membuat profil siswa: %v", err)
	}
	return u
}

func (e *staffEnv) exists(t *testing.T, id uuid.UUID) bool {
	t.Helper()
	var n int64
	e.db.Model(&domain.User{}).Where("id = ?", id).Count(&n)
	return n > 0
}

// deactivate menonaktifkan akun lewat UPDATE langsung (Create dengan false memakai default true).
func (e *staffEnv) deactivate(t *testing.T, id uuid.UUID) {
	t.Helper()
	if err := e.db.Model(&domain.User{}).Where("id = ?", id).Update("is_active", false).Error; err != nil {
		t.Fatalf("gagal menonaktifkan akun: %v", err)
	}
}

// Jalur /users/:id harus memakai pra-cek username unik yang sama dengan /teachers/:id.
func TestUpdateUserRejectsDuplicateUsername(t *testing.T) {
	e := newStaffEnv(t)
	admin := e.mkUser(t, "admin", domain.RoleAdmin, nil)
	e.mkUser(t, "sudah-ada", domain.RoleGuru, nil)
	target := e.mkUser(t, "target", domain.RoleGuru, []string{"reports:export"})
	e.caller = &admin

	status, resp := e.do(t, http.MethodPut, "/users/"+target.ID.String(), `{"username":"sudah-ada","full_name":"Nama Baru"}`)
	if status != http.StatusBadRequest || resp["message"] != "Username sudah digunakan" {
		t.Fatalf("status=%d pesan=%v, ingin 400 dan Username sudah digunakan", status, resp["message"])
	}
	if got := e.stored(t, target.ID); got.Username != "target" || got.FullName != "target" {
		t.Fatalf("data berubah walau ditolak: %+v", got)
	}

	// Username sendiri (tidak berubah) dan username baru yang bebas tetap sah.
	if status, _ := e.do(t, http.MethodPut, "/users/"+target.ID.String(), `{"username":"target","full_name":"Nama Baru"}`); status != http.StatusOK {
		t.Fatalf("username sama: status = %d, ingin 200", status)
	}
	if status, _ := e.do(t, http.MethodPut, "/users/"+target.ID.String(), `{"username":"bebas"}`); status != http.StatusOK {
		t.Fatalf("username baru: status = %d, ingin 200", status)
	}
	if got := e.stored(t, target.ID); got.Username != "bebas" {
		t.Fatalf("username = %q", got.Username)
	}
}

// Galat simpan tidak lagi diabaikan: akun dan profil disimpan atomik dan gagal dibalas 500.
func TestUpdateUserReportsSaveErrorAndRollsBack(t *testing.T) {
	e := newStaffEnv(t)
	admin := e.mkUser(t, "admin", domain.RoleAdmin, nil)
	e.mkStudent(t, "siswa-a", "1001", "0001")
	b := e.mkStudent(t, "siswa-b", "1002", "0002")
	e.caller = &admin

	// NIS 1001 sudah dipakai siswa A: penyimpanan profil gagal (indeks unik).
	status, resp := e.do(t, http.MethodPut, "/users/"+b.ID.String(), `{"full_name":"Nama Baru","nis":"1001"}`)
	if status != http.StatusInternalServerError {
		t.Fatalf("status = %d, ingin 500", status)
	}
	if resp["success"] != false || resp["message"] == nil || resp["message"] == "" {
		t.Fatalf("respons galat tidak jelas: %v", resp)
	}
	if got := e.stored(t, b.ID); got.FullName != "siswa-b" {
		t.Fatalf("perubahan akun harus dibatalkan (rollback), nama = %q", got.FullName)
	}

	// Perubahan yang sah tetap tersimpan.
	if status, _ := e.do(t, http.MethodPut, "/users/"+b.ID.String(), `{"full_name":"Nama Baru","nis":"1003"}`); status != http.StatusOK {
		t.Fatalf("status = %d, ingin 200", status)
	}
	if got := e.stored(t, b.ID); got.FullName != "Nama Baru" {
		t.Fatalf("nama = %q", got.FullName)
	}
}

// Akun yang sudah berganti role tidak boleh diambil alih lewat jalur siswa oleh non-grantor.
func TestStudentRoutesProtectPromotedAccounts(t *testing.T) {
	e := newStaffEnv(t)
	staf := e.mkUser(t, "staf", domain.RoleGuru, []string{"master:manage"})
	promoted := e.mkStudent(t, "eks-siswa", "2001", "0201")
	if err := e.db.Model(&domain.User{}).Where("id = ?", promoted.ID).Update("role", domain.RoleGuru).Error; err != nil {
		t.Fatalf("gagal mengubah role: %v", err)
	}
	e.caller = &staf

	status, _ := e.do(t, http.MethodPut, "/students/"+promoted.ID.String(), `{"password":"diambil-alih","username":"pengambilalih"}`)
	if status != http.StatusForbidden {
		t.Fatalf("update: status = %d, ingin 403", status)
	}
	status, _ = e.do(t, http.MethodDelete, "/students/"+promoted.ID.String(), "")
	if status != http.StatusForbidden {
		t.Fatalf("delete: status = %d, ingin 403", status)
	}
	if got := e.stored(t, promoted.ID); got.Username != "eks-siswa" || got.PasswordHash != "x" {
		t.Fatalf("akun berubah: %+v", got)
	}

	// Siswa sungguhan tetap dapat dikelola pemegang master:manage.
	real := e.mkStudent(t, "siswa", "2002", "0202")
	if status, _ := e.do(t, http.MethodPut, "/students/"+real.ID.String(), `{"full_name":"Nama Baru"}`); status != http.StatusOK {
		t.Fatalf("update siswa: status = %d, ingin 200", status)
	}
}

func TestLastActiveAdminCannotBeDemoted(t *testing.T) {
	for _, route := range []string{"/teachers/", "/users/"} {
		t.Run(route, func(t *testing.T) {
			e := newStaffEnv(t)
			only := e.mkUser(t, "admin-satu", domain.RoleAdmin, nil)
			e.caller = &only // administrator menurunkan dirinya sendiri

			status, resp := e.do(t, http.MethodPut, route+only.ID.String(), `{"role":"GURU"}`)
			if status != http.StatusConflict || resp["message"] != "Tidak dapat menurunkan administrator terakhir" {
				t.Fatalf("status=%d pesan=%v, ingin 409", status, resp["message"])
			}
			if got := e.stored(t, only.ID); got.Role != domain.RoleAdmin {
				t.Fatalf("role berubah walau ditolak: %s", got.Role)
			}

			// Admin lain yang nonaktif tidak dihitung.
			other := e.mkUser(t, "admin-nonaktif", domain.RoleAdmin, nil)
			e.deactivate(t, other.ID)
			status, _ = e.do(t, http.MethodPut, route+only.ID.String(), `{"role":"GURU"}`)
			if status != http.StatusConflict {
				t.Fatalf("dengan admin nonaktif: status = %d, ingin 409", status)
			}

			// Ada administrator aktif lain: penurunan diizinkan.
			e.mkUser(t, "admin-dua", domain.RoleAdmin, nil)
			status, _ = e.do(t, http.MethodPut, route+only.ID.String(), `{"role":"GURU"}`)
			if status != http.StatusOK {
				t.Fatalf("dengan admin aktif lain: status = %d, ingin 200", status)
			}
			got := e.stored(t, only.ID)
			if got.Role != domain.RoleGuru || !reflect.DeepEqual(got.Permissions, domain.TemplatePermissions("guru")) {
				t.Fatalf("role=%s izin=%v", got.Role, got.Permissions)
			}
		})
	}
}

// Menurunkan admin nonaktif tidak mengurangi jumlah administrator yang dapat masuk.
func TestDemotingInactiveAdminIsAllowed(t *testing.T) {
	for _, route := range []string{"/teachers/", "/users/"} {
		t.Run(route, func(t *testing.T) {
			e := newStaffEnv(t)
			caller := domain.User{ID: uuid.New(), Username: "peminta", Role: domain.RoleAdmin, IsActive: true}
			e.caller = &caller
			inactive := e.mkUser(t, "admin-nonaktif", domain.RoleAdmin, nil)
			e.deactivate(t, inactive.ID)

			if status, _ := e.do(t, http.MethodPut, route+inactive.ID.String(), `{"role":"GURU"}`); status != http.StatusOK {
				t.Fatalf("status = %d, ingin 200", status)
			}
		})
	}
}

func TestLastActiveAdminCannotBeDeleted(t *testing.T) {
	for _, route := range []string{"/teachers/", "/users/"} {
		t.Run(route, func(t *testing.T) {
			e := newStaffEnv(t)
			// Pemanggil bukan akun tersimpan, jadi pemeriksaan hapus diri sendiri tidak menghalangi.
			caller := domain.User{ID: uuid.New(), Username: "peminta", Role: domain.RoleAdmin, IsActive: true}
			e.caller = &caller
			only := e.mkUser(t, "admin-satu", domain.RoleAdmin, nil)

			status, resp := e.do(t, http.MethodDelete, route+only.ID.String(), "")
			if status != http.StatusConflict || resp["message"] != "Tidak dapat menghapus administrator terakhir" {
				t.Fatalf("status=%d pesan=%v, ingin 409", status, resp["message"])
			}
			if !e.exists(t, only.ID) {
				t.Fatal("administrator terakhir terhapus walau ditolak")
			}

			// Ada administrator aktif lain: penghapusan diizinkan.
			e.mkUser(t, "admin-dua", domain.RoleAdmin, nil)
			if status, _ := e.do(t, http.MethodDelete, route+only.ID.String(), ""); status != http.StatusOK {
				t.Fatalf("dengan admin aktif lain: status = %d, ingin 200", status)
			}
			if e.exists(t, only.ID) {
				t.Fatal("akun seharusnya terhapus")
			}
		})
	}
}

// Menghapus guru dan admin nonaktif tidak terhalang penjagaan administrator terakhir.
func TestDeleteNonLastAdminTargetsAreAllowed(t *testing.T) {
	e := newStaffEnv(t)
	caller := domain.User{ID: uuid.New(), Username: "peminta", Role: domain.RoleAdmin, IsActive: true}
	e.caller = &caller
	guru := e.mkUser(t, "guru", domain.RoleGuru, nil)
	inactiveAdmin := e.mkUser(t, "admin-nonaktif", domain.RoleAdmin, nil)
	e.deactivate(t, inactiveAdmin.ID)

	for _, id := range []uuid.UUID{guru.ID, inactiveAdmin.ID} {
		if status, _ := e.do(t, http.MethodDelete, "/teachers/"+id.String(), ""); status != http.StatusOK {
			t.Fatalf("hapus %s: status = %d, ingin 200", id, status)
		}
		if e.exists(t, id) {
			t.Fatalf("akun %s seharusnya terhapus", id)
		}
	}
}
