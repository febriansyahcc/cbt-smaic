Write-Host "=========================================" -ForegroundColor Cyan
Write-Host "  CBT High School Examination System" -ForegroundColor Green
Write-Host "=========================================" -ForegroundColor Cyan

$backendProcess = Start-Process -FilePath "D:\cbt\backend\cbt-backend.exe" -WorkingDirectory "D:\cbt\backend" -PassThru
Write-Host "[OK] Backend Go Fiber berjalan di http://localhost:8080" -ForegroundColor Green

Write-Host "[INFO] Menjalankan Frontend Vite Dev Server..." -ForegroundColor Yellow
Set-Location -Path "D:\cbt\frontend"
npm run dev