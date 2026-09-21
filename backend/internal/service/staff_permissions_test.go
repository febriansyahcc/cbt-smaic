package service

import (
	"reflect"
	"testing"

	"cbt-backend/internal/domain"
)

func TestResolveStaffPermissionsOnCreate(t *testing.T) {
	guruTpl := domain.TemplatePermissions("guru")
	tests := []struct {
		name      string
		role      domain.Role
		canGrant  bool
		requested []string
		want      []string
		wantErr   bool
	}{
		{"guru tanpa izin menyimpan template guru", domain.RoleGuru, true, nil, guruTpl, false},
		{"guru oleh non-admin menyimpan template guru", domain.RoleGuru, false, nil, guruTpl, false},
		{"guru dengan izin kustom disanitasi dan diberi implikasi", domain.RoleGuru, true, []string{"schedules:manage", "*", "bogus"}, []string{"schedules:manage", "schedules:read"}, false},
		{"guru dengan daftar kosong ditolak", domain.RoleGuru, true, []string{}, nil, true},
		{"guru dengan hanya izin tak sah ditolak", domain.RoleGuru, true, []string{"*", "exam:take", "bogus"}, nil, true},
		{"admin tanpa izin tidak menyimpan izin", domain.RoleAdmin, true, nil, nil, false},
		{"admin dengan izin selalu bintang", domain.RoleAdmin, true, []string{"reports:export"}, []string{"*"}, false},
		{"admin dengan daftar kosong tetap sah", domain.RoleAdmin, true, []string{}, []string{"*"}, false},
		{"siswa tidak menyimpan izin", domain.RoleSiswa, true, []string{"users:manage"}, nil, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ResolveStaffPermissionsOnCreate(tt.role, tt.canGrant, tt.requested)
			if tt.wantErr {
				if err == nil || !IsValidationError(err) {
					t.Fatalf("ingin galat validasi, dapat %v", err)
				}
				if err.Error() != "Pilih minimal satu izin untuk akun guru" {
					t.Fatalf("pesan galat = %q", err.Error())
				}
				return
			}
			if err != nil {
				t.Fatalf("galat tak terduga: %v", err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("izin = %v, ingin %v", got, tt.want)
			}
		})
	}
}

func TestResolveStaffPermissionsOnUpdate(t *testing.T) {
	guruTpl := domain.TemplatePermissions("guru")
	custom := []string{"reports:export"}
	tests := []struct {
		name      string
		prev, now domain.Role
		current   []string
		canGrant  bool
		requested []string
		want      []string
		wantErr   bool
	}{
		{"admin ke guru mengisi template guru", domain.RoleAdmin, domain.RoleGuru, nil, true, nil, guruTpl, false},
		{"admin ke guru tidak membawa bintang lama", domain.RoleAdmin, domain.RoleGuru, []string{"*"}, true, nil, guruTpl, false},
		{"admin ke guru memakai izin kiriman bila ada", domain.RoleAdmin, domain.RoleGuru, nil, true, []string{"reports:export"}, []string{"reports:export"}, false},
		{"admin ke guru dengan daftar kosong ditolak", domain.RoleAdmin, domain.RoleGuru, nil, true, []string{}, nil, true},
		{"guru ke admin mengosongkan izin", domain.RoleGuru, domain.RoleAdmin, custom, true, nil, nil, false},
		{"guru tetap guru mempertahankan izin kustom", domain.RoleGuru, domain.RoleGuru, custom, true, nil, custom, false},
		{"guru tetap guru mengganti izin", domain.RoleGuru, domain.RoleGuru, custom, true, []string{"proctor:control_all"}, []string{"proctor:control_all", "proctor:control"}, false},
		{"guru tetap guru dengan daftar kosong ditolak", domain.RoleGuru, domain.RoleGuru, custom, true, []string{}, nil, true},
		{"guru baris lama berizin kosong dijadikan eksplisit", domain.RoleGuru, domain.RoleGuru, nil, false, nil, guruTpl, false},
		{"non-admin tanpa izin kiriman tidak mengubah izin guru", domain.RoleGuru, domain.RoleGuru, custom, false, nil, custom, false},
		{"guru ke siswa mengosongkan izin", domain.RoleGuru, domain.RoleSiswa, custom, true, []string{"users:manage"}, nil, false},
		{"siswa tetap siswa tidak berubah", domain.RoleSiswa, domain.RoleSiswa, nil, true, []string{"users:manage"}, nil, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ResolveStaffPermissionsOnUpdate(tt.prev, tt.now, tt.current, tt.canGrant, tt.requested)
			if tt.wantErr {
				if err == nil || !IsValidationError(err) {
					t.Fatalf("ingin galat validasi, dapat %v", err)
				}
				return
			}
			if err != nil {
				t.Fatalf("galat tak terduga: %v", err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("izin = %v, ingin %v", got, tt.want)
			}
		})
	}
}

// Pemanggil tanpa izin penuh yang mengirim daftar izin (termasuk daftar kosong) ditolak
// dengan ErrPermissionsNotAllowed untuk semua role, baik create maupun update. Galat ini
// bukan galat validasi karena handler membalasnya 403, bukan 400.
func TestResolveStaffPermissionsRejectsNonGrantorInput(t *testing.T) {
	for _, requested := range [][]string{{}, {"users:manage"}, {"reports:export", "master:manage"}} {
		for _, role := range []domain.Role{domain.RoleGuru, domain.RoleAdmin, domain.RoleSiswa} {
			if _, err := ResolveStaffPermissionsOnCreate(role, false, requested); !IsPermissionsNotAllowed(err) || IsValidationError(err) {
				t.Errorf("create %s requested=%v: err = %v, ingin ErrPermissionsNotAllowed", role, requested, err)
			}
			if _, err := ResolveStaffPermissionsOnUpdate(role, role, nil, false, requested); !IsPermissionsNotAllowed(err) || IsValidationError(err) {
				t.Errorf("update %s requested=%v: err = %v, ingin ErrPermissionsNotAllowed", role, requested, err)
			}
		}
	}
	// Tanpa daftar izin, non-grantor tetap boleh (create -> template guru, update -> izin lama).
	got, err := ResolveStaffPermissionsOnCreate(domain.RoleGuru, false, nil)
	if err != nil || !reflect.DeepEqual(got, domain.TemplatePermissions("guru")) {
		t.Errorf("create guru tanpa izin oleh non-grantor: %v, %v", got, err)
	}
	got, err = ResolveStaffPermissionsOnCreate(domain.RoleSiswa, false, nil)
	if err != nil || got != nil {
		t.Errorf("create siswa tanpa izin oleh non-grantor: %v, %v", got, err)
	}
}

// Hasil resolusi untuk GURU tidak pernah kosong, sehingga tidak ada eskalasi diam-diam
// lewat fallback template pada EffectivePermissions.
func TestResolveStaffPermissionsNeverEmptyForGuru(t *testing.T) {
	for _, requested := range [][]string{nil, {}, {"*"}, {"reports:export"}} {
		for _, canGrant := range []bool{true, false} {
			if got, err := ResolveStaffPermissionsOnCreate(domain.RoleGuru, canGrant, requested); err == nil && len(got) == 0 {
				t.Errorf("create guru menghasilkan izin kosong (requested=%v canGrant=%v)", requested, canGrant)
			}
			if got, err := ResolveStaffPermissionsOnUpdate(domain.RoleAdmin, domain.RoleGuru, nil, canGrant, requested); err == nil && len(got) == 0 {
				t.Errorf("update ke guru menghasilkan izin kosong (requested=%v canGrant=%v)", requested, canGrant)
			}
		}
	}
}
