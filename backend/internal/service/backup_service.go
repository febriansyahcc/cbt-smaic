package service

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"cbt-backend/internal/domain"
	"cbt-backend/internal/repository"
)

// Backup berupa satu arsip .tar.gz berisi:
//   - database.dump   (PostgreSQL, format custom pg_dump) atau database.sqlite (mode dev SQLite)
//   - uploads/...     (gambar soal yang diunggah guru)
//   - manifest.json   (waktu, jenis, dan driver database)
//
// Arsip ditulis ke berkas sementara lalu di-rename, sehingga daftar backup tidak pernah
// memuat arsip setengah jadi. Backup otomatis berjalan sekali sehari setelah BACKUP_HOUR
// dan ditunda selama ada ujian berlangsung.

const (
	BackupKindAuto   = "auto"
	BackupKindManual = "manual"

	backupTimeLayout  = "20060102-150405"
	backupDumpName    = "database.dump"
	backupSQLiteName  = "database.sqlite"
	backupDumpTimeout = 15 * time.Minute
	// Jeda sebelum backup otomatis dicoba lagi setelah gagal, agar tidak mengulang tiap menit.
	backupRetryDelay = 30 * time.Minute
)

var backupNamePattern = regexp.MustCompile(`^cbt-backup-(\d{8}-\d{6})-(auto|manual)\.tar\.gz$`)

var (
	ErrBackupRunning  = errors.New("backup lain sedang berjalan")
	ErrBackupNotFound = errors.New("berkas backup tidak ditemukan")
	ErrBackupBadName  = errors.New("nama berkas backup tidak valid")
)

type BackupItem struct {
	Name      string    `json:"name"`
	Kind      string    `json:"kind"`
	Size      int64     `json:"size"`
	CreatedAt time.Time `json:"created_at"`
}

type BackupStatus struct {
	Running        bool         `json:"running"`
	LastError      string       `json:"last_error,omitempty"`
	LastFinishedAt *time.Time   `json:"last_finished_at,omitempty"`
	AutoEnabled    bool         `json:"auto_enabled"`
	AutoHour       int          `json:"auto_hour"`
	Keep           int          `json:"keep"`
	Items          []BackupItem `json:"items"`
}

type BackupService struct {
	repo       *repository.Database
	dir        string
	uploadsDir string
	keep       int
	autoOn     bool
	autoHour   int
	now        func() time.Time

	mu             sync.Mutex
	running        bool
	lastError      string
	lastFinishedAt *time.Time
	autoRetryAfter time.Time
}

// NewBackupService membaca konfigurasi dari environment:
//
//	BACKUP_DIR   folder penyimpanan (default ./backups)
//	BACKUP_KEEP  jumlah backup yang disimpan per jenis (default 14)
//	BACKUP_AUTO  "false" untuk mematikan backup otomatis harian
//	BACKUP_HOUR  jam mulai backup otomatis, 0-23 waktu server (default 1)
func NewBackupService(repo *repository.Database) *BackupService {
	s := &BackupService{
		repo:       repo,
		dir:        envOr("BACKUP_DIR", "./backups"),
		uploadsDir: "./uploads",
		keep:       envInt("BACKUP_KEEP", 14, 1, 365),
		autoOn:     !strings.EqualFold(os.Getenv("BACKUP_AUTO"), "false"),
		autoHour:   envInt("BACKUP_HOUR", 1, 0, 23),
		now:        time.Now,
	}
	return s
}

func envOr(key, def string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return def
}

func envInt(key string, def, min, max int) int {
	v, err := strconv.Atoi(strings.TrimSpace(os.Getenv(key)))
	if err != nil || v < min || v > max {
		return def
	}
	return v
}

// Status mengembalikan daftar backup (terbaru dulu) beserta status proses backup.
func (s *BackupService) Status() (BackupStatus, error) {
	items, err := s.List()
	s.mu.Lock()
	defer s.mu.Unlock()
	return BackupStatus{
		Running:        s.running,
		LastError:      s.lastError,
		LastFinishedAt: s.lastFinishedAt,
		AutoEnabled:    s.autoOn,
		AutoHour:       s.autoHour,
		Keep:           s.keep,
		Items:          items,
	}, err
}

// List mengembalikan backup yang ada di folder, terbaru dulu.
func (s *BackupService) List() ([]BackupItem, error) {
	entries, err := os.ReadDir(s.dir)
	if errors.Is(err, fs.ErrNotExist) {
		return []BackupItem{}, nil
	}
	if err != nil {
		return nil, err
	}
	items := []BackupItem{}
	for _, e := range entries {
		m := backupNamePattern.FindStringSubmatch(e.Name())
		if m == nil || !e.Type().IsRegular() {
			continue
		}
		info, err := e.Info()
		if err != nil {
			continue
		}
		created, err := time.ParseInLocation(backupTimeLayout, m[1], time.Local)
		if err != nil {
			continue
		}
		items = append(items, BackupItem{Name: e.Name(), Kind: m[2], Size: info.Size(), CreatedAt: created})
	}
	sort.Slice(items, func(i, j int) bool { return items[i].Name > items[j].Name })
	return items, nil
}

// Path mengembalikan path berkas backup setelah memvalidasi nama (mencegah path traversal).
func (s *BackupService) Path(name string) (string, error) {
	if !backupNamePattern.MatchString(name) {
		return "", ErrBackupBadName
	}
	p := filepath.Join(s.dir, name)
	info, err := os.Stat(p)
	if err != nil || !info.Mode().IsRegular() {
		return "", ErrBackupNotFound
	}
	return p, nil
}

func (s *BackupService) Delete(name string) error {
	p, err := s.Path(name)
	if err != nil {
		return err
	}
	return os.Remove(p)
}

// StartAsync memulai backup di latar belakang. Mengembalikan ErrBackupRunning bila
// backup lain belum selesai.
func (s *BackupService) StartAsync(kind string) error {
	if !s.begin() {
		return ErrBackupRunning
	}
	go func() {
		_, err := s.run(kind)
		s.finish(err)
	}()
	return nil
}

// Create menjalankan backup secara sinkron (dipakai penjadwal dan tes).
func (s *BackupService) Create(kind string) (BackupItem, error) {
	if !s.begin() {
		return BackupItem{}, ErrBackupRunning
	}
	item, err := s.run(kind)
	s.finish(err)
	return item, err
}

func (s *BackupService) begin() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.running {
		return false
	}
	s.running = true
	return true
}

func (s *BackupService) finish(err error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.running = false
	now := s.now()
	s.lastFinishedAt = &now
	if err != nil {
		s.lastError = err.Error()
		log.Printf("Backup gagal: %v", err)
	} else {
		s.lastError = ""
	}
}

func (s *BackupService) run(kind string) (BackupItem, error) {
	if kind != BackupKindAuto && kind != BackupKindManual {
		kind = BackupKindManual
	}
	if err := os.MkdirAll(s.dir, 0o750); err != nil {
		return BackupItem{}, fmt.Errorf("gagal menyiapkan folder backup: %w", err)
	}

	created := s.now()
	name := fmt.Sprintf("cbt-backup-%s-%s.tar.gz", created.Format(backupTimeLayout), kind)
	finalPath := filepath.Join(s.dir, name)
	if _, err := os.Stat(finalPath); err == nil {
		return BackupItem{}, fmt.Errorf("backup %s sudah ada, coba lagi sebentar", name)
	}

	work, err := os.MkdirTemp(s.dir, ".work-")
	if err != nil {
		return BackupItem{}, fmt.Errorf("gagal membuat folder kerja: %w", err)
	}
	defer os.RemoveAll(work)

	dbFile, driver, err := s.dumpDatabase(work)
	if err != nil {
		return BackupItem{}, err
	}

	manifest, _ := json.MarshalIndent(map[string]any{
		"app":        "cbt-smaic",
		"created_at": created.Format(time.RFC3339),
		"kind":       kind,
		"db_driver":  driver,
		"db_file":    filepath.Base(dbFile),
	}, "", "  ")

	partial := filepath.Join(work, name+".partial")
	if err := s.writeArchive(partial, dbFile, manifest); err != nil {
		return BackupItem{}, err
	}
	if err := os.Chmod(partial, 0o640); err != nil {
		return BackupItem{}, err
	}
	if err := os.Rename(partial, finalPath); err != nil {
		return BackupItem{}, fmt.Errorf("gagal menyimpan backup: %w", err)
	}

	info, err := os.Stat(finalPath)
	if err != nil {
		return BackupItem{}, err
	}
	log.Printf("Backup %s selesai (%d byte)", name, info.Size())
	s.prune(kind)
	return BackupItem{Name: name, Kind: kind, Size: info.Size(), CreatedAt: created}, nil
}

// dumpDatabase menulis salinan database ke folder kerja dan memverifikasinya.
func (s *BackupService) dumpDatabase(work string) (string, string, error) {
	driver := s.repo.DB.Dialector.Name()
	switch driver {
	case "postgres":
		out := filepath.Join(work, backupDumpName)
		if err := pgDump(os.Getenv("DB_DSN"), out); err != nil {
			return "", driver, err
		}
		return out, driver, nil
	case "sqlite":
		out := filepath.Join(work, backupSQLiteName)
		// VACUUM INTO menghasilkan salinan database yang konsisten tanpa menghentikan aplikasi.
		if err := s.repo.DB.Exec("VACUUM INTO ?", out).Error; err != nil {
			return "", driver, fmt.Errorf("gagal menyalin database SQLite: %w", err)
		}
		return out, driver, nil
	default:
		return "", driver, fmt.Errorf("driver database %q belum didukung untuk backup", driver)
	}
}

// pgDump menjalankan pg_dump format custom, lalu memastikan hasilnya terbaca oleh pg_restore.
// Password dikirim lewat PGPASSWORD agar tidak terlihat di daftar proses.
func pgDump(dsn, out string) error {
	conn, password, err := splitPostgresPassword(dsn)
	if err != nil {
		return err
	}
	env := append(os.Environ(), "PGPASSWORD="+password)

	ctx, cancel := context.WithTimeout(context.Background(), backupDumpTimeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, pgTool("pg_dump"), "--format=custom", "--no-owner", "--no-privileges", "--file="+out, "--dbname="+conn)
	cmd.Env = env
	if msg, err := cmd.CombinedOutput(); err != nil {
		if errors.Is(err, exec.ErrNotFound) {
			return errors.New("pg_dump tidak ditemukan di server backend")
		}
		return fmt.Errorf("pg_dump gagal: %v: %s", err, strings.TrimSpace(string(msg)))
	}

	var stderr strings.Builder
	verify := exec.CommandContext(ctx, pgTool("pg_restore"), "--list", out)
	verify.Stdout = io.Discard
	verify.Stderr = &stderr
	if err := verify.Run(); err != nil {
		return fmt.Errorf("hasil pg_dump tidak dapat dibaca pg_restore: %v: %s", err, strings.TrimSpace(stderr.String()))
	}
	return nil
}

// pgTool mencari program klien PostgreSQL. Di image Alpine, program ini biasanya ditautkan
// ke /usr/bin, dengan lokasi asli di /usr/libexec/postgresql16 sebagai cadangan.
func pgTool(name string) string {
	if p, err := exec.LookPath(name); err == nil {
		return p
	}
	if p := filepath.Join("/usr/libexec/postgresql16", name); fileExists(p) {
		return p
	}
	return name
}

func fileExists(p string) bool {
	info, err := os.Stat(p)
	return err == nil && info.Mode().IsRegular()
}

// splitPostgresPassword memisahkan password dari DSN. Mendukung format URL
// (postgres://user:pass@host/db) dan format key=value (host=... password=...).
func splitPostgresPassword(dsn string) (string, string, error) {
	dsn = strings.TrimSpace(dsn)
	if dsn == "" {
		return "", "", errors.New("DB_DSN kosong")
	}
	if strings.HasPrefix(dsn, "postgres://") || strings.HasPrefix(dsn, "postgresql://") {
		u, err := url.Parse(dsn)
		if err != nil {
			return "", "", errors.New("DB_DSN tidak valid")
		}
		password := ""
		if u.User != nil {
			password, _ = u.User.Password()
			u.User = url.User(u.User.Username())
		}
		return u.String(), password, nil
	}
	var kept []string
	password := ""
	for _, part := range strings.Fields(dsn) {
		if strings.HasPrefix(part, "password=") {
			password = strings.Trim(strings.TrimPrefix(part, "password="), "'")
			continue
		}
		kept = append(kept, part)
	}
	return strings.Join(kept, " "), password, nil
}

// writeArchive membuat arsip tar.gz berisi manifest, database, dan folder uploads.
func (s *BackupService) writeArchive(path, dbFile string, manifest []byte) (err error) {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_EXCL, 0o600)
	if err != nil {
		return fmt.Errorf("gagal membuat arsip: %w", err)
	}
	defer func() {
		if cerr := f.Close(); err == nil && cerr != nil {
			err = cerr
		}
	}()
	gz := gzip.NewWriter(f)
	tw := tar.NewWriter(gz)

	now := s.now()
	if err := tw.WriteHeader(&tar.Header{Name: "manifest.json", Mode: 0o644, Size: int64(len(manifest)), ModTime: now}); err != nil {
		return err
	}
	if _, err := tw.Write(manifest); err != nil {
		return err
	}
	if err := addFileToTar(tw, dbFile, filepath.Base(dbFile)); err != nil {
		return fmt.Errorf("gagal menambahkan database ke arsip: %w", err)
	}

	walkErr := filepath.WalkDir(s.uploadsDir, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			if errors.Is(err, fs.ErrNotExist) && p == s.uploadsDir {
				return fs.SkipDir
			}
			return err
		}
		if !d.Type().IsRegular() {
			return nil
		}
		rel, err := filepath.Rel(s.uploadsDir, p)
		if err != nil {
			return err
		}
		return addFileToTar(tw, p, filepath.ToSlash(filepath.Join("uploads", rel)))
	})
	if walkErr != nil {
		return fmt.Errorf("gagal menambahkan berkas unggahan ke arsip: %w", walkErr)
	}

	if err := tw.Close(); err != nil {
		return err
	}
	return gz.Close()
}

func addFileToTar(tw *tar.Writer, src, name string) error {
	f, err := os.Open(src)
	if err != nil {
		return err
	}
	defer f.Close()
	info, err := f.Stat()
	if err != nil {
		return err
	}
	if err := tw.WriteHeader(&tar.Header{Name: name, Mode: 0o644, Size: info.Size(), ModTime: info.ModTime()}); err != nil {
		return err
	}
	_, err = io.CopyN(tw, f, info.Size())
	return err
}

// prune menghapus backup lama dengan jenis yang sama, menyisakan s.keep terbaru.
func (s *BackupService) prune(kind string) {
	items, err := s.List()
	if err != nil {
		return
	}
	kept := 0
	for _, it := range items {
		if it.Kind != kind {
			continue
		}
		kept++
		if kept > s.keep {
			if err := os.Remove(filepath.Join(s.dir, it.Name)); err == nil {
				log.Printf("Backup lama dihapus: %s", it.Name)
			}
		}
	}
}

// examInProgress memeriksa apakah ada siswa yang sedang mengerjakan ujian. Backup otomatis
// ditunda selama kondisi ini benar agar tidak menambah beban server. Yang dicek adalah sesi
// siswa (selalu punya batas waktu), bukan rentang jadwal, supaya jadwal yang dibiarkan aktif
// berhari-hari tidak menunda backup tanpa batas.
func (s *BackupService) examInProgress() (bool, error) {
	var sessions int64
	err := s.repo.DB.Model(&domain.ExamSession{}).
		Where("status IN ? AND server_deadline > ?", []domain.SessionStatus{domain.StatusInProgress, domain.StatusBlocked}, s.now()).
		Count(&sessions).Error
	return sessions > 0, err
}

// autoBackupDue: sudah lewat BACKUP_HOUR hari ini dan belum ada backup otomatis hari ini.
func (s *BackupService) autoBackupDue() bool {
	now := s.now()
	if now.Hour() < s.autoHour {
		return false
	}
	items, err := s.List()
	if err != nil {
		return false
	}
	today := now.Format("20060102")
	for _, it := range items {
		if it.Kind == BackupKindAuto && it.CreatedAt.Format("20060102") == today {
			return false
		}
	}
	return true
}

// TickAuto dijalankan penjadwal setiap menit.
func (s *BackupService) TickAuto() {
	if !s.autoOn || !s.autoBackupDue() {
		return
	}
	s.mu.Lock()
	wait := s.now().Before(s.autoRetryAfter)
	s.mu.Unlock()
	if wait {
		return
	}
	busy, err := s.examInProgress()
	if err != nil || busy {
		return
	}
	if _, err := s.Create(BackupKindAuto); err != nil && !errors.Is(err, ErrBackupRunning) {
		s.mu.Lock()
		s.autoRetryAfter = s.now().Add(backupRetryDelay)
		s.mu.Unlock()
		log.Printf("Backup otomatis gagal, dicoba lagi %s lagi: %v", backupRetryDelay, err)
	}
}

// StartScheduler menjalankan pengecekan backup otomatis setiap menit sampai ctx selesai.
func (s *BackupService) StartScheduler(ctx context.Context) {
	if !s.autoOn {
		log.Printf("Backup otomatis dimatikan (BACKUP_AUTO=false)")
		return
	}
	log.Printf("Backup otomatis aktif: setiap hari mulai pukul %02d:00, menyimpan %d backup terakhir di %s", s.autoHour, s.keep, s.dir)
	go func() {
		t := time.NewTicker(time.Minute)
		defer t.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-t.C:
				s.TickAuto()
			}
		}
	}()
}
