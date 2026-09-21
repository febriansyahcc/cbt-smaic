#!/bin/bash
# Setup awal CBT di server baru
# Jalankan: bash setup.sh

set -e

REGISTRY_USER="mochamaddwifebriansyah13"

echo "=== CBT System — Setup Server ==="

# Pastikan .env sudah ada
if [ ! -f .env ]; then
  cp .env.example .env
  echo ""
  echo "File .env dibuat dari .env.example."
  echo "WAJIB isi JWT_SECRET dan DB_PASSWORD sebelum lanjut:"
  echo "  nano .env"
  echo ""
  echo "Generate JWT_SECRET:"
  echo "  openssl rand -hex 32"
  echo ""
  exit 1
fi

# Cek variabel wajib
source .env
if [ -z "$JWT_SECRET" ]; then
  echo "ERROR: JWT_SECRET kosong di .env. Jalankan: openssl rand -hex 32"
  exit 1
fi
if [ -z "$DB_PASSWORD" ]; then
  echo "ERROR: DB_PASSWORD kosong di .env."
  exit 1
fi

echo ">> Pull image terbaru..."
docker pull ${REGISTRY_USER}/cbt-backend:latest
docker pull ${REGISTRY_USER}/cbt-frontend:latest

echo ">> Menjalankan sistem..."
docker compose up -d

echo ""
echo "=== Sistem berjalan ==="
docker compose ps
echo ""
echo "Akses: http://$(hostname -I | awk '{print $1}')"
