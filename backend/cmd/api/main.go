package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"cbt-backend/internal/domain"
	"cbt-backend/internal/handler"
	"cbt-backend/internal/middleware"
	"cbt-backend/internal/repository"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/joho/godotenv"
)

func main() {
	// Load .env if present
	_ = godotenv.Load()

	if err := middleware.ValidateJWTSecret(); err != nil {
		log.Fatalf("Fatal: %v. Isi JWT_SECRET di file .env dengan nilai acak, mis. hasil perintah: openssl rand -hex 32", err)
	}

	// Initialize Database
	db, err := repository.InitDB()
	if err != nil {
		log.Fatalf("Fatal: Database initialization failed: %v", err)
	}

	app := fiber.New(fiber.Config{
		AppName:      "CBT High School Engine v1.0",
		BodyLimit:    20 * 1024 * 1024, // 20 MB for excel/media uploads
		ServerHeader: "CBT-Go-Engine",
	})

	// Global Middlewares
	app.Use(recover.New())
	app.Use(logger.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowHeaders: "Origin, Content-Type, Accept, Authorization",
		AllowMethods: "GET, POST, PUT, DELETE, OPTIONS",
	}))
	// Enable Gzip/Brotli compression for high-concurrency low-bandwidth performance
	app.Use(compress.New(compress.Config{
		Level: compress.LevelBestSpeed,
	}))

	// Ensure uploads directory exists
	_ = os.MkdirAll("./uploads/questions", 0755)

	handlers := handler.NewHandlers(db)

	// Static Media Serving
	app.Static("/uploads", "./uploads")

	// Health Check — performs a real DB ping so external monitors (e.g. UptimeRobot)
	// get an accurate signal; returns 503 when the database is unreachable.
	app.Get("/health", func(c *fiber.Ctx) error {
		ts := time.Now().Unix()
		sqlDB, err := db.DB.DB()
		if err != nil || sqlDB.Ping() != nil {
			msg := "database unreachable"
			if err != nil {
				msg = err.Error()
			}
			return c.Status(http.StatusServiceUnavailable).JSON(fiber.Map{
				"status":    "error",
				"message":   msg,
				"version":   "1.0.0",
				"timestamp": ts,
			})
		}
		return c.JSON(fiber.Map{
			"status":    "ok",
			"engine":    "Go Fiber",
			"message":   "CBT High School Backend is running",
			"version":   "1.0.0",
			"timestamp": ts,
		})
	})

	api := app.Group("/api/v1")

	// Public Auth Routes
	auth := api.Group("/auth")
	auth.Post("/login", handlers.HandleLogin)
	auth.Get("/me", middleware.AuthRequired(db), handlers.HandleGetMe)
	auth.Post("/logout", middleware.AuthRequired(db), handlers.HandleLogout)

	// ---------------- PBAC: definisi izin per kelompok route ----------------
	// Izin menentukan AKSI; cakupan DATA (bank soal milik/ampuan, jadwal yang ditugaskan)
	// dicek lagi di handler lewat AccessService.
	perm := middleware.RequirePermission
	pQuestionsAssigned := string(domain.PermQuestionsAssigned)
	pQuestionsAll := string(domain.PermQuestionsAll)
	pQuestionsUpload := string(domain.PermQuestionsUpload)
	pQuestionsManage := string(domain.PermQuestionsManage)
	pQuestionsLock := string(domain.PermQuestionsLock)
	pSchedulesRead := string(domain.PermSchedulesRead)
	pSchedulesManage := string(domain.PermSchedulesManage)
	pEventsManage := string(domain.PermEventsManage)
	pMaster := string(domain.PermMasterManage)
	pUsers := string(domain.PermUsersManage)
	pProctorView := string(domain.PermProctorView)
	pProctorControl := string(domain.PermProctorControl)
	pProctorControlAll := string(domain.PermProctorControlAll)
	pReports := string(domain.PermReportsExport)

	// Student Protected Routes
	student := api.Group("/student", middleware.AuthRequired(db), perm(string(domain.PermExamTake)))
	student.Get("/schedules", handlers.HandleGetStudentSchedules)
	student.Post("/exams/start", handlers.HandleStartExam)
	student.Post("/exams/sync", handlers.HandleSyncAnswers)
	student.Post("/exams/violation", handlers.HandleRecordViolation)
	student.Post("/exams/submit", handlers.HandleSubmitExam)

	// Proctor Routes: proctor:view/control_all melihat semua jadwal aktif; proctor:control saja
	// hanya melihat dan mengendalikan jadwal yang ditugaskan (dicek lewat AccessService di handler).
	proctor := api.Group("/proctor", middleware.AuthRequired(db), middleware.StaffOnly())
	proctorView := perm(pProctorView, pProctorControl, pProctorControlAll)
	proctorControl := perm(pProctorControl, pProctorControlAll)
	proctor.Get("/schedules", proctorView, handlers.HandleGetProctorSchedules)
	proctor.Get("/live/:schedule_id", proctorView, handlers.HandleGetLiveProctorData)
	proctor.Get("/sessions/:id/violations", proctorView, handlers.HandleGetSessionViolations)
	proctor.Post("/unlock", proctorControl, handlers.HandleUnlockStudentSession)
	proctor.Post("/sessions/:id/unlock", proctorControl, handlers.HandleUnlockStudentSession)
	proctor.Post("/reset-device", proctorControl, handlers.HandleResetStudentSession)
	proctor.Post("/students/:user_id/reset-session", proctorControl, handlers.HandleResetStudentSession)
	proctor.Post("/sessions/:id/extend-time", proctorControl, handlers.HandleExtendTimeSession)
	proctor.Post("/schedules/:id/extend-time-all", proctorControl, handlers.HandleExtendTimeAllSchedule)
	proctor.Post("/sessions/:id/force-submit", proctorControl, handlers.HandleForceSubmitSession)
	proctor.Get("/reports/excel/:schedule_id", perm(pReports), handlers.HandleExportGradesExcel)
	proctor.Get("/reports/pdf/:schedule_id", perm(pReports), handlers.HandleExportBeritaAcaraPDF)

	// Admin / Staff Routes
	// Grup ini hanya pintu masuk untuk staf. Setiap route di bawahnya meminta izin spesifik.
	// Route baca (GET) data referensi non-pribadi (kelas, mapel, jadwal, event) terbuka bagi seluruh staf
	// karena dimuat bersamaan oleh dashboard. Daftar siswa dan akun staf dibatasi lewat peopleRead;
	// data bank soal dan jadwal pada readiness-matrix sudah disaring menurut cakupan pengguna.
	admin := api.Group("/admin", middleware.AuthRequired(db), middleware.StaffOnly())
	master := perm(pMaster)
	usersMgr := perm(pUsers, pMaster)
	eventsMgr := perm(pEventsManage)
	schedulesMgr := perm(pSchedulesManage)
	bankWrite := perm(pQuestionsUpload, pQuestionsManage)
	bankManage := perm(pQuestionsManage)
	bankRead := perm(pQuestionsAssigned, pQuestionsAll, pQuestionsUpload, pQuestionsManage)
	// Daftar siswa dan akun staf memuat data pribadi: hanya untuk pengelola master, akun, atau jadwal.
	peopleRead := perm(pMaster, pUsers, pSchedulesManage)

	admin.Get("/permission-catalog", perm(pMaster, pUsers), handlers.HandleGetPermissionCatalog)
	admin.Get("/template/question-bank.xlsx", perm(pQuestionsUpload), handlers.HandleGetQuestionBankTemplate)
	admin.Post("/questions/import", perm(pQuestionsUpload), handlers.HandleImportQuestionsExcel)
	admin.Get("/readiness-matrix", handlers.HandleGetReadinessMatrix)
	admin.Get("/dashboard-stats", handlers.HandleGetAdminStats)
	admin.Get("/classes", handlers.HandleGetClasses)
	admin.Post("/classes", master, handlers.HandleCreateClass)
	admin.Get("/classes/template", master, handlers.HandleGetClassesTemplate)
	admin.Post("/classes/import", master, handlers.HandleImportClassesExcel)
	admin.Put("/classes/:id", master, handlers.HandleUpdateClass)
	admin.Delete("/classes/:id", master, handlers.HandleDeleteClass)
	admin.Get("/students", peopleRead, handlers.HandleGetStudents)
	admin.Post("/students", master, handlers.HandleCreateStudent)
	admin.Put("/students/:id", master, handlers.HandleUpdateStudent)
	admin.Delete("/students/:id", master, handlers.HandleDeleteStudent)
	admin.Get("/teachers", peopleRead, handlers.HandleGetTeachers)
	admin.Post("/teachers", master, handlers.HandleCreateTeacher)
	admin.Get("/teachers/template", master, handlers.HandleGetTeachersTemplate)
	admin.Post("/teachers/import", master, handlers.HandleImportTeachersExcel)
	admin.Put("/teachers/:id", master, handlers.HandleUpdateTeacher)
	admin.Delete("/teachers/:id", master, handlers.HandleDeleteTeacher)
	admin.Get("/subjects", handlers.HandleGetSubjects)
	admin.Post("/subjects", master, handlers.HandleCreateSubject)
	admin.Get("/subjects/template", master, handlers.HandleGetSubjectsTemplate)
	admin.Post("/subjects/import", master, handlers.HandleImportSubjectsExcel)
	admin.Put("/subjects/:id", master, handlers.HandleUpdateSubject)
	admin.Delete("/subjects/:id", master, handlers.HandleDeleteSubject)
	admin.Get("/class-subjects", handlers.HandleGetClassSubjects)
	admin.Post("/class-subjects", master, handlers.HandleCreateClassSubject)
	admin.Put("/class-subjects/:id", master, handlers.HandleUpdateClassSubject)
	admin.Delete("/class-subjects/:id", master, handlers.HandleDeleteClassSubject)
	admin.Get("/schedules", handlers.HandleGetAdminSchedules)
	admin.Post("/schedules", schedulesMgr, handlers.HandleCreateSchedule)
	admin.Put("/schedules/:id", schedulesMgr, handlers.HandleUpdateSchedule)
	// Menautkan bank soal mengubah jadwal: hanya pengelola jadwal (schedules:manage; ADMIN dan "*" otomatis).
	admin.Post("/schedules/:id/link-bank", schedulesMgr, handlers.HandleLinkScheduleBank)
	admin.Post("/schedules/:id/quick-bank", schedulesMgr, handlers.HandleQuickCreateAndLinkBank)
	admin.Delete("/schedules/:id", schedulesMgr, handlers.HandleDeleteSchedule)
	admin.Post("/schedules/:id/toggle", schedulesMgr, handlers.HandleToggleSchedule)
	essayGrade := perm(pQuestionsUpload, pQuestionsManage, pSchedulesManage)
	admin.Get("/schedules/:id/essay-answers", essayGrade, handlers.HandleGetEssayAnswers)
	admin.Patch("/schedules/:id/essay-answers", essayGrade, handlers.HandleGradeEssayAnswers)
	admin.Get("/schedules/:id/proctors", perm(pSchedulesRead, pSchedulesManage), handlers.HandleGetScheduleProctors)
	admin.Get("/proctor-candidates", schedulesMgr, handlers.HandleGetProctorCandidates)
	admin.Put("/schedules/:id/proctors", schedulesMgr, handlers.HandleSetScheduleProctors)
	admin.Post("/question-banks", bankWrite, handlers.HandleCreateQuestionBank)
	admin.Put("/question-banks/:id", bankManage, handlers.HandleUpdateQuestionBank)
	admin.Delete("/question-banks/:id", bankManage, handlers.HandleDeleteQuestionBank)
	admin.Post("/question-banks/:id/toggle-lock", perm(pQuestionsLock), handlers.HandleToggleBankLock)
	admin.Get("/question-banks/:id/questions", bankRead, handlers.HandleGetBankQuestions)
	admin.Put("/question-banks/:id/reorder", bankManage, handlers.HandleReorderBankQuestions)
	admin.Post("/question-banks/:id/questions", perm(pQuestionsUpload), handlers.HandleCreateQuestion)
	admin.Put("/questions/:id", bankManage, handlers.HandleUpdateQuestion)
	admin.Delete("/questions/:id", bankManage, handlers.HandleDeleteQuestion)
	admin.Post("/upload-image", bankWrite, handlers.HandleUploadImage)
	admin.Post("/questions/upload-image", bankWrite, handlers.HandleUploadImage)

	// Exam Event Routes
	admin.Get("/events", handlers.HandleGetEvents)
	admin.Post("/events", eventsMgr, handlers.HandleCreateEvent)
	admin.Put("/events/:id", eventsMgr, handlers.HandleUpdateEvent)
	admin.Post("/events/:id/toggle-active", eventsMgr, handlers.HandleToggleEventActive)
	admin.Delete("/events/:id", eventsMgr, handlers.HandleDeleteEvent)
	admin.Get("/events/:id/participants", eventsMgr, handlers.HandleGetEventParticipants)
	admin.Post("/events/:id/participants/generate", eventsMgr, handlers.HandleGenerateEventParticipants)

	// User Management Routes
	admin.Get("/users", usersMgr, handlers.HandleGetUsers)
	admin.Post("/users", usersMgr, handlers.HandleCreateUser)
	admin.Put("/users/:id", usersMgr, handlers.HandleUpdateUser)
	admin.Post("/users/:id/toggle-status", usersMgr, handlers.HandleToggleUserStatus)
	admin.Post("/users/:id/reset-password", usersMgr, handlers.HandleResetUserPassword)
	admin.Post("/users/:id/reset-session", usersMgr, handlers.HandleResetUserSession)
	admin.Delete("/users/:id", usersMgr, handlers.HandleDeleteUser)
	admin.Get("/users/template", usersMgr, handlers.HandleGetStudentTemplate)
	admin.Post("/users/import-excel", usersMgr, handlers.HandleImportStudentsExcel)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("CBT Backend Server listening on port %s...", port)
	if err := app.Listen(":" + port); err != nil {
		log.Fatalf("Server stopped: %v", err)
	}
}
