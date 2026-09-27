package handler

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// ---------------- MEDIA / IMAGE UPLOAD HANDLER ----------------

func (h *Handlers) HandleUploadImage(c *fiber.Ctx) error {
	file, err := c.FormFile("image")
	if err != nil {
		file, err = c.FormFile("file")
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"success": false,
				"message": "File gambar tidak ditemukan dalam request",
			})
		}
	}

	// Max 5 MB
	if file.Size > 5*1024*1024 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Ukuran file gambar melebihi batas maksimal 5 MB",
		})
	}

	ext := strings.ToLower(filepath.Ext(file.Filename))
	allowedExtensions := map[string]bool{
		".jpg":  true,
		".jpeg": true,
		".png":  true,
		".webp": true,
		".gif":  true,
		// .svg sengaja ditolak: SVG bisa memuat script yang berjalan di origin aplikasi.
	}

	if !allowedExtensions[ext] {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Format file tidak didukung. Gunakan format JPG, PNG, WEBP, atau GIF",
		})
	}

	uploadDir := "./uploads/questions"
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Gagal menyiapkan direktori penyimpanan gambar",
		})
	}

	uniqueFilename := fmt.Sprintf("%s%s", uuid.New().String(), ext)
	targetPath := filepath.Join(uploadDir, uniqueFilename)

	if err := c.SaveFile(file, targetPath); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Gagal menyimpan file gambar ke server",
		})
	}

	fileURL := "/uploads/questions/" + uniqueFilename
	return c.JSON(fiber.Map{
		"success":  true,
		"url":      fileURL,
		"filename": uniqueFilename,
		"message":  "Gambar berhasil diunggah",
	})
}
