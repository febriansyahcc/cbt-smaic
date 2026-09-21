package service

import (
	"errors"

	"cbt-backend/internal/domain"
)

// ErrGuruNeedsPermission dikembalikan bila daftar izin yang dikirim untuk akun GURU
// kosong setelah disanitasi. Daftar kosong berarti fallback ke template guru saat
// pembacaan (EffectivePermissions), yaitu eskalasi diam-diam, sehingga harus ditolak.
var ErrGuruNeedsPermission = newValidationError("Pilih minimal satu izin untuk akun guru")

// ErrPermissionsNotAllowed dikembalikan bila pemanggil yang bukan pemegang izin penuh
// (canGrant=false) mengirim daftar izin. Ini bukan galat validasi: handler membalasnya 403.
var ErrPermissionsNotAllowed = errors.New("hanya administrator yang dapat mengatur izin akun")

// IsPermissionsNotAllowed memeriksa apakah galat berasal dari pengiriman izin oleh non-grantor.
func IsPermissionsNotAllowed(err error) bool {
	return errors.Is(err, ErrPermissionsNotAllowed)
}

// ResolveStaffPermissionsOnCreate menentukan izin yang disimpan untuk akun baru.
//
//   - Pemanggil tanpa izin penuh (canGrant=false) yang mengirim daftar izin (requested != nil)
//     ditolak dengan ErrPermissionsNotAllowed, tidak diabaikan diam-diam.
//   - SISWA: nil (izin siswa dibaca dari template).
//   - GURU: SELALU eksplisit. Tanpa daftar izin dipakai TemplatePermissions("guru").
//     Daftar yang dikirim disanitasi dan tidak boleh kosong.
//   - ADMIN: ["*"] bila daftar dikirim, selain itu nil (ADMIN selalu "*").
func ResolveStaffPermissionsOnCreate(role domain.Role, canGrant bool, requested []string) ([]string, error) {
	if requested != nil && !canGrant {
		return nil, ErrPermissionsNotAllowed
	}
	switch role {
	case domain.RoleSiswa:
		return nil, nil
	case domain.RoleGuru:
		if requested != nil {
			perms := domain.SanitizePermissions(role, requested)
			if len(perms) == 0 {
				return nil, ErrGuruNeedsPermission
			}
			return perms, nil
		}
		return domain.TemplatePermissions("guru"), nil
	}
	if requested != nil {
		return domain.SanitizePermissions(role, requested), nil
	}
	return nil, nil
}

// ResolveStaffPermissionsOnUpdate menentukan izin yang disimpan setelah akun diubah.
// current adalah izin tersimpan saat ini, prevRole dan newRole adalah role sebelum dan sesudah.
//
//   - Pemanggil tanpa izin penuh yang mengirim daftar izin ditolak (ErrPermissionsNotAllowed).
//   - Role berubah: izin lama tidak terbawa. Menjadi GURU -> template guru eksplisit,
//     menjadi ADMIN atau SISWA -> nil.
//   - Daftar izin dikirim (bukan untuk SISWA) -> disanitasi; untuk GURU hasil kosong ditolak.
//   - Akun GURU yang izinnya masih kosong (baris lama) ikut dijadikan eksplisit
//     (template guru), sama dengan izin efektifnya sekarang.
func ResolveStaffPermissionsOnUpdate(prevRole, newRole domain.Role, current []string, canGrant bool, requested []string) ([]string, error) {
	if requested != nil && !canGrant {
		return nil, ErrPermissionsNotAllowed
	}
	perms := current
	if newRole != prevRole {
		perms = nil
		if newRole == domain.RoleGuru {
			perms = domain.TemplatePermissions("guru")
		}
	}
	if requested != nil && newRole != domain.RoleSiswa {
		sanitized := domain.SanitizePermissions(newRole, requested)
		if newRole == domain.RoleGuru && len(sanitized) == 0 {
			return nil, ErrGuruNeedsPermission
		}
		perms = sanitized
	}
	if newRole == domain.RoleGuru && len(perms) == 0 {
		perms = domain.TemplatePermissions("guru")
	}
	return perms, nil
}
