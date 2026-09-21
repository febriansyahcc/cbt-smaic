package repository

import (
	"fmt"
	"log"
	"os"
	"time"

	"cbt-backend/internal/domain"

	"github.com/glebarez/sqlite"
	"github.com/google/uuid"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type Database struct {
	DB *gorm.DB
}

func InitDB() (*Database, error) {
	dsn := os.Getenv("DB_DSN")
	var db *gorm.DB
	var err error

	gormConfig := &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn),
	}

	if dsn != "" {
		log.Printf("Connecting to PostgreSQL with DSN: %s", dsn)
		db, err = gorm.Open(postgres.Open(dsn), gormConfig)
		if err != nil {
			log.Printf("Failed to connect to PostgreSQL: %v. Falling back to SQLite for seamless operation.", err)
		}
	}

	if db == nil {
		sqliteFile := os.Getenv("SQLITE_FILE")
		if sqliteFile == "" {
			sqliteFile = "cbt.db"
		}
		log.Printf("Using Pure-Go SQLite database: %s", sqliteFile)
		db, err = gorm.Open(sqlite.Open(sqliteFile), gormConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to open database: %w", err)
		}
	}

	// Auto-migrate tables
	err = db.AutoMigrate(
		&domain.User{},
		&domain.ClassRoom{},
		&domain.StudentProfile{},
		&domain.Subject{},
		&domain.ClassSubject{},
		&domain.QuestionBank{},
		&domain.Question{},
		&domain.ExamEvent{},
		&domain.ExamSchedule{},
		&domain.ExamSession{},
		&domain.StudentAnswer{},
		&domain.ViolationLog{},
		&domain.ScheduleProctor{},
	)
	if err != nil {
		return nil, fmt.Errorf("auto-migration failed: %w", err)
	}

	repo := &Database{DB: db}
	repo.SeedInitialData()
	repo.BackfillStaffPermissions()
	repo.EnsureClassSubjectAssignments()
	repo.MigrateOrSeedDefaultEvent()

	return repo, nil
}

func (d *Database) MigrateOrSeedDefaultEvent() {
	var eventCount int64
	d.DB.Model(&domain.ExamEvent{}).Count(&eventCount)

	var defaultEvent domain.ExamEvent
	if eventCount == 0 {
		now := time.Now()
		defaultEvent = domain.ExamEvent{
			ID:           uuid.New(),
			Title:        "Penilaian Akhir Semester (PAS) Ganjil 2026/2027",
			Code:         "PAS-GANJIL-2026",
			AcademicYear: "2026/2027",
			Semester:     "GANJIL",
			StartDate:    now.Add(-7 * 24 * time.Hour),
			EndDate:      now.Add(30 * 24 * time.Hour),
			IsActive:     true,
			Description:  "Event Penilaian Akhir Semester Ganjil Tahun Ajaran 2026/2027",
			CreatedAt:    now,
		}
		if err := d.DB.Create(&defaultEvent).Error; err != nil {
			log.Printf("Failed to create default exam event: %v", err)
			return
		}
		log.Printf("Default Exam Event created: %s (%s)", defaultEvent.Title, defaultEvent.Code)
	} else {
		if err := d.DB.Where("is_active = ?", true).First(&defaultEvent).Error; err != nil {
			d.DB.First(&defaultEvent)
		}
	}

	if defaultEvent.ID != uuid.Nil {
		res := d.DB.Model(&domain.ExamSchedule{}).Where("event_id IS NULL").Update("event_id", defaultEvent.ID)
		if res.RowsAffected > 0 {
			log.Printf("Migrated %d orphan exam schedule(s) to event: %s", res.RowsAffected, defaultEvent.Title)
		}
	}
}