package handler

import (
	"errors"

	"cbt-backend/internal/service"

	"github.com/gofiber/fiber/v2"
)

// ---------------- BACKUP DATA (khusus administrator) ----------------

func (h *Handlers) HandleGetBackups(c *fiber.Ctx) error {
	status, err := h.backupService.Status()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false, "message": "Gagal membaca daftar backup"})
	}
	return c.JSON(fiber.Map{"success": true, "data": status})
}

func (h *Handlers) HandleCreateBackup(c *fiber.Ctx) error {
	if err := h.backupService.StartAsync(service.BackupKindManual); err != nil {
		if errors.Is(err, service.ErrBackupRunning) {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{"success": false, "message": "Backup lain masih berjalan, tunggu hingga selesai"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false, "message": err.Error()})
	}
	return c.Status(fiber.StatusAccepted).JSON(fiber.Map{"success": true, "message": "Backup sedang dibuat"})
}

func (h *Handlers) HandleDownloadBackup(c *fiber.Ctx) error {
	name := c.Params("name")
	path, err := h.backupService.Path(name)
	if err != nil {
		return backupLookupError(c, err)
	}
	c.Set("Cache-Control", "no-store")
	return c.Download(path, name)
}

func (h *Handlers) HandleDeleteBackup(c *fiber.Ctx) error {
	if err := h.backupService.Delete(c.Params("name")); err != nil {
		return backupLookupError(c, err)
	}
	return c.JSON(fiber.Map{"success": true, "message": "Backup berhasil dihapus"})
}

func backupLookupError(c *fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, service.ErrBackupBadName):
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "message": "Nama berkas backup tidak valid"})
	case errors.Is(err, service.ErrBackupNotFound):
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"success": false, "message": "Berkas backup tidak ditemukan"})
	default:
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false, "message": "Gagal memproses berkas backup"})
	}
}
