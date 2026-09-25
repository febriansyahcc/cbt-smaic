package domain

import (
	"sort"
	"sync"
	"time"

	"github.com/google/uuid"
)

// Izin tambahan untuk migrasi ke PBAC (Permission-Based Access Control).
// Izin lama tetap dideklarasikan di models.go agar data pengguna yang ada tetap valid.
const (
	PermQuestionsManage   Permission = "questions:manage"
	PermProctorControlAll Permission = "proctor:control_all"
	PermEventsManage      Permission = "events:manage"
	PermUsersManage       Permission = "users:manage"
	PermQuestionsPrint    Permission = "questions:print"
)

// ScheduleProctor menugaskan seorang pengawas ke satu jadwal ujian.
// Pemegang proctor:control hanya dapat MELIHAT dan MENGENDALIKAN jadwal yang tercatat
// di tabel ini. Melihat semua jadwal aktif butuh proctor:view, mengendalikan semua
// jadwal butuh proctor:control_all.
type ScheduleProctor struct {
	ScheduleID uuid.UUID `gorm:"type:uuid;primaryKey" json:"schedule_id"`
	UserID     uuid.UUID `gorm:"type:uuid;primaryKey;index" json:"user_id"`
	CreatedAt  time.Time `json:"created_at"`
}

// PermissionItem adalah satu izin pada katalog. Sensitive menandai izin berdampak besar
// (mis. mengelola akun atau melihat semua bank soal) agar antarmuka dapat memberi peringatan.
type PermissionItem struct {
	Key         string `json:"key"`
	Label       string `json:"label"`
	Description string `json:"description"`
	Sensitive   bool   `json:"sensitive"`
}

type PermissionGroup struct {
	Key   string           `json:"key"`
	Label string           `json:"label"`
	Items []PermissionItem `json:"items"`
}

// RoleTemplate adalah paket izin bawaan yang mengisi checkbox di antarmuka.
// Role pada template menentukan nilai User.Role saat template dipilih.
type RoleTemplate struct {
	Key         string   `json:"key"`
	Label       string   `json:"label"`
	Description string   `json:"description"`
	Role        Role     `json:"role"`
	Permissions []string `json:"permissions"`
}

// PermissionCatalog mengembalikan seluruh izin yang dapat diatur, dikelompokkan untuk antarmuka.
func PermissionCatalog() []PermissionGroup {
	return []PermissionGroup{
		{
			Key:   "questions",
			Label: "Bank Soal",
			Items: []PermissionItem{
				{Key: string(PermQuestionsAssigned), Label: "Lihat bank soal yang ditugaskan", Description: "Bank soal buatan sendiri dan mata pelajaran yang diampu"},
				{Key: string(PermQuestionsAll), Label: "Lihat semua bank soal", Description: "Seluruh bank soal dan kesiapan jadwal", Sensitive: true},
				{Key: string(PermQuestionsUpload), Label: "Tambah dan unggah soal", Description: "Membuat bank soal, menambah butir soal, impor Excel, unggah gambar"},
				{Key: string(PermQuestionsManage), Label: "Ubah dan hapus bank soal", Description: "Mengedit atau menghapus bank soal dan butir soal"},
				{Key: string(PermQuestionsLock), Label: "Kunci dan buka kunci naskah", Description: "Mengunci bank soal agar tidak dapat diubah"},
				{Key: string(PermQuestionsPrint), Label: "Cetak naskah dan kunci jawaban", Description: "Mencetak naskah soal dan kunci jawaban dari bank soal yang sudah terkunci"},
			},
		},
		{
			Key:   "schedules",
			Label: "Jadwal dan Event",
			Items: []PermissionItem{
				{Key: string(PermSchedulesRead), Label: "Lihat jadwal ujian", Description: "Melihat daftar jadwal ujian"},
				{Key: string(PermSchedulesManage), Label: "Kelola jadwal ujian", Description: "Membuat, mengubah, menghapus, mengaktifkan jadwal, dan menugaskan pengawas"},
				{Key: string(PermEventsManage), Label: "Kelola event ujian", Description: "Membuat, mengubah, dan menghapus event ujian"},
			},
		},
		{
			Key:   "proctor",
			Label: "Pengawasan",
			Items: []PermissionItem{
				{Key: string(PermProctorView), Label: "Lihat pengawasan semua jadwal aktif", Description: "Memantau siswa pada seluruh jadwal yang aktif"},
				{Key: string(PermProctorControl), Label: "Kendalikan jadwal yang ditugaskan", Description: "Melihat, membuka blokir, reset login, tambah waktu, kumpulkan paksa hanya pada jadwal yang ditugaskan"},
				{Key: string(PermProctorControlAll), Label: "Kendalikan semua jadwal", Description: "Aksi pengawasan pada seluruh jadwal tanpa perlu penugasan", Sensitive: true},
			},
		},
		{
			Key:   "reports",
			Label: "Laporan",
			Items: []PermissionItem{
				{Key: string(PermReportsExport), Label: "Ekspor laporan", Description: "Rekap nilai Excel dan Berita Acara PDF"},
			},
		},
		{
			Key:   "admin",
			Label: "Administrasi",
			Items: []PermissionItem{
				{Key: string(PermMasterManage), Label: "Kelola data master", Description: "Siswa, guru dan staf, kelas, mata pelajaran, dan penugasan kelas mapel", Sensitive: true},
				{Key: string(PermUsersManage), Label: "Kelola akun pengguna", Description: "Tambah, ubah, nonaktifkan, dan reset kata sandi akun", Sensitive: true},
			},
		},
	}
}

// RoleTemplates mengembalikan paket izin bawaan. Urutan izin mengikuti urutan katalog.
func RoleTemplates() []RoleTemplate {
	return []RoleTemplate{
		{
			Key:         "admin",
			Label:       "Administrator",
			Description: "Akses penuh ke seluruh fitur dan data",
			Role:        RoleAdmin,
			Permissions: []string{string(PermAll)},
		},
		{
			Key:         "guru",
			Label:       "Guru Pengampu",
			Description: "Mengelola bank soal mata pelajaran yang diampu, mengawasi jadwal yang ditugaskan, dan mengekspor laporan",
			Role:        RoleGuru,
			Permissions: []string{
				string(PermQuestionsAssigned),
				string(PermQuestionsUpload),
				string(PermQuestionsManage),
				string(PermQuestionsLock),
				string(PermQuestionsPrint),
				string(PermSchedulesRead),
				string(PermProctorControl),
				string(PermReportsExport),
			},
		},
		{
			Key:         "pengawas",
			Label:       "Pengawas Ujian",
			Description: "Melihat dan mengendalikan hanya jadwal yang ditugaskan",
			Role:        RoleGuru,
			Permissions: []string{
				string(PermSchedulesRead),
				string(PermProctorControl),
				string(PermReportsExport),
			},
		},
		{
			Key:         "siswa",
			Label:       "Siswa",
			Description: "Mengerjakan ujian",
			Role:        RoleSiswa,
			Permissions: []string{string(PermExamTake)},
		},
	}
}

// TemplatePermissions mengembalikan salinan izin bawaan untuk template dengan key tertentu.
func TemplatePermissions(key string) []string {
	for _, t := range RoleTemplates() {
		if t.Key == key {
			out := make([]string, len(t.Permissions))
			copy(out, t.Permissions)
			return out
		}
	}
	return []string{}
}

// IsKnownPermission memeriksa apakah key ada di katalog (tanpa "*" dan exam:take).
func IsKnownPermission(key string) bool {
	_, ok := permissionOrder()[key]
	return ok
}

// permissionOrder mengembalikan peta kunci izin -> urutan di katalog. Dibangun sekali
// (sync.Once) karena dipanggil pada setiap request lewat EffectivePermissions; peta hanya
// dibaca sesudahnya sehingga aman dipakai bersamaan oleh banyak goroutine.
var (
	permissionOrderOnce sync.Once
	permissionOrderMap  map[string]int
)

func permissionOrder() map[string]int {
	permissionOrderOnce.Do(func() {
		m := make(map[string]int)
		for _, g := range PermissionCatalog() {
			for _, it := range g.Items {
				m[it.Key] = len(m)
			}
		}
		permissionOrderMap = m
	})
	return permissionOrderMap
}

// permissionImplications adalah tabel implikasi yang dipakai internal (hanya dibaca).
var permissionImplications = map[string][]string{
	string(PermSchedulesManage):   {string(PermSchedulesRead)},
	string(PermProctorControlAll): {string(PermProctorControl)},
	string(PermQuestionsManage):   {string(PermQuestionsAssigned)},
	string(PermQuestionsUpload):   {string(PermQuestionsAssigned)},
	string(PermQuestionsLock):     {string(PermQuestionsAssigned)},
	string(PermQuestionsPrint):    {string(PermQuestionsAssigned)},
}

// PermissionImplications mengembalikan peta ketergantungan izin: kunci adalah izin yang
// dipilih, nilainya izin yang otomatis ikut diberikan. Ini satu-satunya sumber kebenaran
// implikasi; SanitizePermissions menerapkannya (termasuk secara transitif) dan
// endpoint katalog mengirimkannya ke antarmuka. Peta yang dikembalikan adalah salinan.
func PermissionImplications() map[string][]string {
	out := make(map[string][]string, len(permissionImplications))
	for k, v := range permissionImplications {
		out[k] = append([]string(nil), v...)
	}
	return out
}

// SanitizePermissions membersihkan daftar izin masukan pengguna:
//   - ADMIN selalu ["*"]
//   - SISWA selalu ["exam:take"]
//   - selain itu hanya izin yang ada di katalog, tanpa duplikat, urutan masukan dipertahankan.
//     Izin turunan (PermissionImplications) yang belum ada ditambahkan di akhir, berurutan
//     mengikuti urutan katalog agar hasilnya deterministik.
//
// Izin "*" dan "exam:take" tidak pernah diteruskan dari masukan untuk role selain ADMIN/SISWA.
func SanitizePermissions(role Role, in []string) []string {
	switch role {
	case RoleAdmin:
		return []string{string(PermAll)}
	case RoleSiswa:
		return []string{string(PermExamTake)}
	}
	seen := make(map[string]bool)
	out := make([]string, 0, len(in))
	for _, p := range in {
		if !IsKnownPermission(p) || seen[p] {
			continue
		}
		seen[p] = true
		out = append(out, p)
	}

	// Closure implikasi: telusuri izin yang dipilih beserta turunannya.
	order := permissionOrder()
	var extra []string
	for i := 0; i < len(out)+len(extra); i++ {
		var p string
		if i < len(out) {
			p = out[i]
		} else {
			p = extra[i-len(out)]
		}
		for _, dep := range permissionImplications[p] {
			if _, known := order[dep]; known && !seen[dep] {
				seen[dep] = true
				extra = append(extra, dep)
			}
		}
	}
	if len(extra) > 0 {
		sort.SliceStable(extra, func(a, b int) bool { return order[extra[a]] < order[extra[b]] })
		out = append(out, extra...)
	}
	return out
}

// IsStaff bernilai true bila pengguna memiliki minimal satu izin selain exam:take.
func (u *User) IsStaff() bool {
	if u.Role == RoleAdmin {
		return true
	}
	for _, p := range u.EffectivePermissions() {
		if p != string(PermExamTake) {
			return true
		}
	}
	return false
}
