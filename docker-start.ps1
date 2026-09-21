Write-Host "===================================================" -ForegroundColor Cyan
Write-Host "  Memulai CBT High School System via Docker Compose" -ForegroundColor Green
Write-Host "===================================================" -ForegroundColor Cyan

docker compose up -d --build
if ($LASTEXITCODE -eq 0) {
    Write-Host ""
    Write-Host "[SUKSES] Sistem CBT berhasil dijalankan di Docker!" -ForegroundColor Green
    Write-Host "Akses Aplikasi : http://localhost" -ForegroundColor Cyan
    Write-Host "Dokumentasi    : docs/PANDUAN_OPERASIONAL.md" -ForegroundColor Yellow
} else {
    Write-Host ""
    Write-Host "[ERROR] Pastikan Docker Desktop telah dibuka dan aktif di Windows." -ForegroundColor Red
}