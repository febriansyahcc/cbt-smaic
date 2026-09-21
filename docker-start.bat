@echo off
echo ===================================================
echo   Memulai CBT High School System via Docker Compose
echo ===================================================
docker compose up -d --build
if %ERRORLEVEL% equ 0 (
    echo.
    echo [SUKSES] Sistem CBT berhasil dijalankan!
    echo Akses Aplikasi: http://localhost
    echo Dokumentasi: docs\PANDUAN_OPERASIONAL.md
) else (
    echo.
    echo [ERROR] Gagal menjalankan Docker Compose. Pastikan Docker Desktop sedang aktif.
)
pause