package domain

import (
	"reflect"
	"sync"
	"testing"
)

func TestSanitizePermissions(t *testing.T) {
	tests := []struct {
		name string
		role Role
		in   []string
		want []string
	}{
		{"admin selalu bintang", RoleAdmin, []string{"questions:read_all"}, []string{"*"}},
		{"siswa selalu exam:take", RoleSiswa, []string{"*", "users:manage"}, []string{"exam:take"}},
		{"guru membuang bintang dan izin tak dikenal", RoleGuru, []string{"*", "bogus:perm", "questions:read_all"}, []string{"questions:read_all"}},
		// questions:upload membawa questions:read_assigned, ditambahkan di akhir.
		{"guru membuang duplikat dan menjaga urutan", RoleGuru, []string{"proctor:view", "questions:upload", "proctor:view"}, []string{"proctor:view", "questions:upload", "questions:read_assigned"}},
		{"guru tanpa izin turunan tidak ditambah apa pun", RoleGuru, []string{"proctor:view", "reports:export"}, []string{"proctor:view", "reports:export"}},
		{"guru tidak boleh mengambil exam:take", RoleGuru, []string{"exam:take"}, []string{}},
		{"masukan nil menghasilkan daftar kosong", RoleGuru, nil, []string{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SanitizePermissions(tt.role, tt.in)
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("SanitizePermissions(%s, %v) = %v, ingin %v", tt.role, tt.in, got, tt.want)
			}
		})
	}
}

func TestHasPermission(t *testing.T) {
	admin := User{Role: RoleAdmin}
	guruDefault := User{Role: RoleGuru}
	guruCustom := User{Role: RoleGuru, Permissions: []string{"proctor:view"}}
	guruWildcard := User{Role: RoleGuru, Permissions: []string{"*"}}
	siswa := User{Role: RoleSiswa}

	tests := []struct {
		name string
		user User
		perm string
		want bool
	}{
		{"admin lolos semua", admin, "users:manage", true},
		{"guru bawaan tidak boleh lihat semua jadwal proktor", guruDefault, "proctor:view", false},
		{"guru bawaan boleh kendalikan jadwal yang ditugaskan", guruDefault, "proctor:control", true},
		{"guru bawaan tidak boleh kelola user", guruDefault, "users:manage", false},
		{"guru bawaan tidak boleh kendalikan semua jadwal", guruDefault, "proctor:control_all", false},
		{"daftar kustom menggantikan bawaan", guruCustom, "questions:upload", false},
		{"daftar kustom tetap memberi izinnya", guruCustom, "proctor:view", true},
		// Data lama GURU berdaftar "*" tidak boleh menjadi pemegang izin penuh:
		// daftar itu kosong setelah sanitasi sehingga kembali ke template guru.
		{"bintang pada daftar guru tidak melolosi semua", guruWildcard, "master:manage", false},
		{"bintang pada daftar guru bukan izin bintang", guruWildcard, "*", false},
		{"bintang pada daftar guru kembali ke template guru", guruWildcard, "questions:upload", true},
		{"siswa mengabaikan izin tersimpan", User{Role: RoleSiswa, Permissions: []string{"*", "users:manage"}}, "users:manage", false},
		{"siswa mengabaikan bintang tersimpan", User{Role: RoleSiswa, Permissions: []string{"*"}}, "*", false},
		{"siswa hanya exam:take", siswa, "exam:take", true},
		{"siswa tidak boleh proktor", siswa, "proctor:view", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := tt.user
			if got := u.HasPermission(tt.perm); got != tt.want {
				t.Fatalf("HasPermission(%q) = %v, ingin %v", tt.perm, got, tt.want)
			}
		})
	}
}

func TestIsStaff(t *testing.T) {
	tests := []struct {
		name string
		user User
		want bool
	}{
		{"admin adalah staf", User{Role: RoleAdmin}, true},
		{"guru bawaan adalah staf", User{Role: RoleGuru}, true},
		{"siswa bukan staf", User{Role: RoleSiswa}, false},
		{"akun dengan izin kustom selain exam:take adalah staf", User{Role: RoleGuru, Permissions: []string{"reports:export"}}, true},
		// exam:take tidak pernah berlaku untuk GURU; daftar kosong setelah sanitasi kembali ke template guru.
		{"guru hanya exam:take kembali ke template guru (staf)", User{Role: RoleGuru, Permissions: []string{"exam:take"}}, true},
		{"siswa dengan izin staf tersimpan tetap bukan staf", User{Role: RoleSiswa, Permissions: []string{"users:manage"}}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := tt.user
			if got := u.IsStaff(); got != tt.want {
				t.Fatalf("IsStaff() = %v, ingin %v", got, tt.want)
			}
		})
	}
}

func TestTemplatesOnlyUseKnownPermissions(t *testing.T) {
	for _, tpl := range RoleTemplates() {
		for _, p := range tpl.Permissions {
			if p == string(PermAll) || p == string(PermExamTake) {
				continue
			}
			if !IsKnownPermission(p) {
				t.Errorf("template %q memakai izin yang tidak ada di katalog: %q", tpl.Key, p)
			}
		}
	}
}

func TestCatalogHasNoDuplicateKeys(t *testing.T) {
	seen := make(map[string]bool)
	for _, g := range PermissionCatalog() {
		for _, it := range g.Items {
			if seen[it.Key] {
				t.Errorf("kunci izin ganda di katalog: %q", it.Key)
			}
			seen[it.Key] = true
		}
	}
}

func TestDefaultPermissionsFollowTemplates(t *testing.T) {
	if got, want := GetDefaultPermissions(RoleGuru), TemplatePermissions("guru"); !reflect.DeepEqual(got, want) {
		t.Fatalf("default guru %v berbeda dari template %v", got, want)
	}
	if got := GetDefaultPermissions(RoleAdmin); !reflect.DeepEqual(got, []string{"*"}) {
		t.Fatalf("default admin = %v, ingin [*]", got)
	}
}

func TestPengawasTemplateCannotManageQuestionsOrUsers(t *testing.T) {
	u := User{Role: RoleGuru, Permissions: TemplatePermissions("pengawas")}
	for _, p := range []string{"questions:upload", "questions:manage", "users:manage", "master:manage", "proctor:view", "proctor:control_all"} {
		if u.HasPermission(p) {
			t.Errorf("template pengawas seharusnya tidak memiliki %q", p)
		}
	}
	for _, p := range []string{"proctor:control", "reports:export"} {
		if !u.HasPermission(p) {
			t.Errorf("template pengawas seharusnya memiliki %q", p)
		}
	}
}

// Guru dan pengawas hanya melihat serta mengendalikan jadwal yang ditugaskan:
// keduanya tidak boleh membawa proctor:view maupun proctor:control_all dari template.
func TestGuruDanPengawasTemplateTanpaLihatSemuaJadwal(t *testing.T) {
	for _, key := range []string{"guru", "pengawas"} {
		perms := TemplatePermissions(key)
		hasControl := false
		for _, p := range perms {
			switch p {
			case string(PermProctorView), string(PermProctorControlAll):
				t.Errorf("template %q tidak boleh memuat %q", key, p)
			case string(PermProctorControl):
				hasControl = true
			}
		}
		if !hasControl {
			t.Errorf("template %q harus tetap memuat proctor:control", key)
		}
	}
}

func TestSanitizePermissionsImplicationClosure(t *testing.T) {
	tests := []struct {
		name string
		in   []string
		want []string
	}{
		{
			"jadwal kelola membawa jadwal lihat",
			[]string{"schedules:manage"},
			[]string{"schedules:manage", "schedules:read"},
		},
		{
			"kendali semua membawa kendali",
			[]string{"proctor:control_all"},
			[]string{"proctor:control_all", "proctor:control"},
		},
		{
			"izin turunan yang sudah ada tidak digandakan dan urutan masukan dipertahankan",
			[]string{"schedules:read", "schedules:manage"},
			[]string{"schedules:read", "schedules:manage"},
		},
		{
			"beberapa izin membawa izin yang sama hanya sekali",
			[]string{"questions:lock", "questions:manage", "questions:upload"},
			[]string{"questions:lock", "questions:manage", "questions:upload", "questions:read_assigned"},
		},
		{
			"izin turunan mengikuti urutan katalog, bukan urutan masukan",
			[]string{"proctor:control_all", "schedules:manage", "questions:upload"},
			[]string{"proctor:control_all", "schedules:manage", "questions:upload", "questions:read_assigned", "schedules:read", "proctor:control"},
		},
		{
			"bintang dan exam:take tetap dibuang walau ada izin lain",
			[]string{"*", "exam:take", "questions:manage", "bogus:perm"},
			[]string{"questions:manage", "questions:read_assigned"},
		},
		{
			"hanya izin terbuang menghasilkan daftar kosong",
			[]string{"*", "exam:take", "bogus:perm"},
			[]string{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SanitizePermissions(RoleGuru, tt.in)
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("SanitizePermissions(%v) = %v, ingin %v", tt.in, got, tt.want)
			}
			// Deterministik: pemanggilan berulang menghasilkan urutan yang sama.
			for i := 0; i < 20; i++ {
				if again := SanitizePermissions(RoleGuru, tt.in); !reflect.DeepEqual(again, got) {
					t.Fatalf("hasil tidak deterministik: %v vs %v", again, got)
				}
			}
		})
	}
}

func TestSanitizePermissionsIsIdempotent(t *testing.T) {
	in := []string{"proctor:control_all", "schedules:manage", "questions:lock"}
	once := SanitizePermissions(RoleGuru, in)
	twice := SanitizePermissions(RoleGuru, once)
	if !reflect.DeepEqual(once, twice) {
		t.Fatalf("sanitasi kedua mengubah hasil: %v -> %v", once, twice)
	}
}

func TestSanitizePermissionsDoesNotMutateInput(t *testing.T) {
	in := []string{"questions:upload"}
	SanitizePermissions(RoleGuru, in)
	if !reflect.DeepEqual(in, []string{"questions:upload"}) {
		t.Fatalf("masukan berubah: %v", in)
	}
}

func TestPermissionImplicationsOnlyReferenceKnownPermissions(t *testing.T) {
	for src, deps := range PermissionImplications() {
		if !IsKnownPermission(src) {
			t.Errorf("kunci implikasi tidak ada di katalog: %q", src)
		}
		for _, dep := range deps {
			if !IsKnownPermission(dep) {
				t.Errorf("implikasi %q -> %q merujuk izin yang tidak ada di katalog", src, dep)
			}
			if dep == src {
				t.Errorf("izin %q tidak boleh menyiratkan dirinya sendiri", src)
			}
		}
	}
}

func TestPermissionImplicationsRequiredEntries(t *testing.T) {
	want := map[string][]string{
		"schedules:manage":    {"schedules:read"},
		"proctor:control_all": {"proctor:control"},
		"questions:manage":    {"questions:read_assigned"},
		"questions:upload":    {"questions:read_assigned"},
		"questions:lock":      {"questions:read_assigned"},
	}
	got := PermissionImplications()
	for k, v := range want {
		if !reflect.DeepEqual(got[k], v) {
			t.Errorf("implikasi %q = %v, ingin %v", k, got[k], v)
		}
	}
}

func TestPermissionImplicationsReturnsCopy(t *testing.T) {
	m := PermissionImplications()
	m["schedules:manage"][0] = "diubah"
	delete(m, "questions:lock")
	fresh := PermissionImplications()
	if fresh["schedules:manage"][0] != "schedules:read" || len(fresh["questions:lock"]) == 0 {
		t.Fatalf("peta implikasi bocor antar pemanggilan: %v", fresh)
	}
}

// Template bawaan sudah tertutup terhadap implikasi: memakainya tidak menambah izin baru.
func TestTemplatesAreClosedUnderImplications(t *testing.T) {
	for _, key := range []string{"guru", "pengawas"} {
		perms := TemplatePermissions(key)
		if got := SanitizePermissions(RoleGuru, perms); !reflect.DeepEqual(got, perms) {
			t.Errorf("template %q berubah setelah sanitasi: %v -> %v", key, perms, got)
		}
	}
}

func TestSensitivePermissionFlags(t *testing.T) {
	want := map[string]bool{
		"users:manage":        true,
		"master:manage":       true,
		"proctor:control_all": true,
		"questions:read_all":  true,
	}
	found := 0
	for _, g := range PermissionCatalog() {
		for _, it := range g.Items {
			if want[it.Key] {
				found++
				if !it.Sensitive {
					t.Errorf("izin %q harus bertanda sensitif", it.Key)
				}
			} else if it.Sensitive {
				t.Errorf("izin %q tidak boleh bertanda sensitif", it.Key)
			}
		}
	}
	if found != len(want) {
		t.Errorf("hanya %d dari %d izin sensitif ditemukan di katalog", found, len(want))
	}
}

func TestEffectivePermissions(t *testing.T) {
	guruTpl := TemplatePermissions("guru")
	tests := []struct {
		name string
		user User
		want []string
	}{
		{"admin selalu bintang walau daftar tersimpan lain", User{Role: RoleAdmin, Permissions: []string{"reports:export"}}, []string{"*"}},
		{"siswa mengabaikan daftar tersimpan", User{Role: RoleSiswa, Permissions: []string{"*", "users:manage"}}, []string{"exam:take"}},
		{"siswa tanpa daftar", User{Role: RoleSiswa}, []string{"exam:take"}},
		{"guru tanpa daftar memakai template", User{Role: RoleGuru}, guruTpl},
		{"guru daftar kosong memakai template", User{Role: RoleGuru, Permissions: []string{}}, guruTpl},
		{"guru daftar tersimpan bersih tidak berubah", User{Role: RoleGuru, Permissions: []string{"reports:export"}}, []string{"reports:export"}},
		{"guru membuang bintang, exam:take, dan kunci tak dikenal", User{Role: RoleGuru, Permissions: []string{"*", "exam:take", "bogus", "reports:export"}}, []string{"reports:export"}},
		{"guru hanya berisi izin tak sah kembali ke template", User{Role: RoleGuru, Permissions: []string{"*", "exam:take", "bogus"}}, guruTpl},
		{"guru mendapat izin turunan", User{Role: RoleGuru, Permissions: []string{"schedules:manage"}}, []string{"schedules:manage", "schedules:read"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := tt.user
			if got := u.EffectivePermissions(); !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("EffectivePermissions() = %v, ingin %v", got, tt.want)
			}
		})
	}
}

// Akun non-ADMIN tidak boleh pernah memiliki izin "*" walau tersimpan di data lama.
func TestEffectivePermissionsNeverGrantWildcardOutsideAdmin(t *testing.T) {
	for _, role := range []Role{RoleGuru, RoleSiswa, Role("")} {
		for _, stored := range [][]string{nil, {"*"}, {"*", "reports:export"}, {"exam:take", "*"}} {
			u := User{Role: role, Permissions: stored}
			for _, p := range u.EffectivePermissions() {
				if p == string(PermAll) {
					t.Errorf("role %q dengan daftar %v mendapat izin bintang", role, stored)
				}
			}
			if u.HasPermission(string(PermAll)) {
				t.Errorf("role %q dengan daftar %v lolos HasPermission(*)", role, stored)
			}
		}
	}
}

func TestEffectivePermissionsDoesNotMutateStored(t *testing.T) {
	stored := []string{"*", "questions:upload"}
	u := User{Role: RoleGuru, Permissions: stored}
	_ = u.EffectivePermissions()
	if !reflect.DeepEqual(u.Permissions, []string{"*", "questions:upload"}) || !reflect.DeepEqual(stored, []string{"*", "questions:upload"}) {
		t.Fatalf("data tersimpan berubah: %v", u.Permissions)
	}
}

// IsKnownPermission dan SanitizePermissions dipanggil pada tiap request oleh banyak
// goroutine: hasilnya harus konsisten dan tidak ada data race pada peta memoize.
func TestSanitizePermissionsConcurrent(t *testing.T) {
	in := []string{"schedules:manage", "*", "questions:upload", "bogus", "proctor:control_all"}
	want := SanitizePermissions(RoleGuru, in)
	var wg sync.WaitGroup
	errs := make(chan string, 64)
	for i := 0; i < 64; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 200; j++ {
				if got := SanitizePermissions(RoleGuru, in); !reflect.DeepEqual(got, want) {
					errs <- "hasil berbeda antar goroutine"
					return
				}
				u := User{Role: RoleGuru, Permissions: in}
				_ = u.EffectivePermissions()
				_ = IsKnownPermission("reports:export")
				_ = PermissionImplications()
			}
		}()
	}
	wg.Wait()
	close(errs)
	for e := range errs {
		t.Fatal(e)
	}
}
