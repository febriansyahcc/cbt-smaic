package handler

import (
	"bytes"
	"mime/multipart"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
)

func TestUploadImageRejectsSVG(t *testing.T) {
	h := &Handlers{}
	app := fiber.New()
	app.Post("/upload", h.HandleUploadImage)

	var body bytes.Buffer
	w := multipart.NewWriter(&body)
	part, _ := w.CreateFormFile("image", "soal.svg")
	part.Write([]byte(`<svg xmlns="http://www.w3.org/2000/svg"><script>alert(1)</script></svg>`))
	w.Close()

	req := httptest.NewRequest("POST", "/upload", &body)
	req.Header.Set("Content-Type", w.FormDataContentType())
	resp, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != fiber.StatusBadRequest {
		t.Fatalf("upload SVG berstatus %d, ingin 400", resp.StatusCode)
	}
}
