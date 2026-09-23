#!/bin/bash
# Auto-update CBT: cek image baru di Docker Hub, lalu pull dan restart container.
#
# Dijalankan berkala lewat cron, contoh setiap 10 menit:
#   */10 * * * * /opt/cbt/deploy/auto-update.sh >> /var/log/cbt-auto-update.log 2>&1
#
# Pengaman:
#   - Update DITUNDA selama ada ujian berjalan (jadwal aktif atau sesi siswa belum selesai).
#   - Buat file .update-paused di folder ini untuk menghentikan update sementara:
#       touch /opt/cbt/deploy/.update-paused   (hapus file itu untuk mengaktifkan lagi)
#   - Pengecekan versi memakai request HEAD ke registry, jadi tidak menghabiskan kuota pull
#     Docker Hub. Pull hanya dilakukan kalau digest di Docker Hub berbeda dengan image lokal.
#   - Setiap image baru WAJIB lolos verifikasi tanda tangan cosign: image harus dibuat oleh
#     workflow .github/workflows/ci-cd.yml di branch main repo GitHub resmi. Image yang di-push
#     dengan cara lain (misalnya token Docker Hub bocor) ditolak dan tidak dijalankan.
#     Butuh binary cosign di server (lihat docs/PANDUAN_OPERASIONAL.md bagian 6).
#   - Image di-pull berdasarkan digest yang sudah diverifikasi, bukan berdasarkan tag, supaya
#     tidak tertukar kalau tag latest berubah di antara verifikasi dan pull.
#   - FORCE=1 melewati cek ujian (untuk update manual di luar jam ujian). Verifikasi tanda
#     tangan tetap berlaku.

set -euo pipefail

# PATH bawaan cron hanya /usr/bin:/bin, sedangkan cosign biasanya ada di /usr/local/bin
export PATH="/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin:${PATH:-}"

cd "$(dirname "$(readlink -f "$0")")"

REGISTRY_USER="mochamaddwifebriansyah13"
TAG="${TAG:-latest}"
SERVICES=(backend frontend)
DB_CONTAINER="cbt-postgres"
COSIGN="${COSIGN:-cosign}"
SIGNER_IDENTITY="https://github.com/febriansyahcc/cbt-smaic/.github/workflows/ci-cd.yml@refs/heads/main"
SIGNER_ISSUER="https://token.actions.githubusercontent.com"

log() { echo "[$(date '+%Y-%m-%d %H:%M:%S')] $*"; }

# Cegah dua proses update berjalan bersamaan
exec 9>.auto-update.lock
if ! flock -n 9; then
  log "Proses update lain masih berjalan, lewati."
  exit 0
fi

if [ -f .update-paused ]; then
  log "Update dijeda (.update-paused ada), lewati."
  exit 0
fi

set -a
# shellcheck disable=SC1091
[ -f .env ] && source .env
set +a
DB_USER="${DB_USER:-cbt_user}"
DB_NAME="${DB_NAME:-cbt_db}"

# Digest image :TAG di Docker Hub (request HEAD, tidak dihitung sebagai pull)
remote_digest() {
  local repo="$1" token
  token=$(curl -fsS "https://auth.docker.io/token?service=registry.docker.io&scope=repository:${repo}:pull" \
    | sed -n 's/.*"token":"\([^"]*\)".*/\1/p')
  [ -n "$token" ] || return 1
  curl -fsSI \
    -H "Authorization: Bearer ${token}" \
    -H "Accept: application/vnd.oci.image.index.v1+json" \
    -H "Accept: application/vnd.docker.distribution.manifest.list.v2+json" \
    -H "Accept: application/vnd.oci.image.manifest.v1+json" \
    -H "Accept: application/vnd.docker.distribution.manifest.v2+json" \
    "https://registry-1.docker.io/v2/${repo}/manifests/${TAG}" \
    | tr -d '\r' | awk 'tolower($1)=="docker-content-digest:" {print $2}'
}

# Digest image lokal yang sedang tersimpan (kosong kalau belum pernah di-pull)
local_digest() {
  docker image inspect --format '{{range .RepoDigests}}{{println .}}{{end}}' "$1:${TAG}" 2>/dev/null \
    | sed -n "s|^$1@||p" | head -n1
}

if ! command -v "$COSIGN" >/dev/null 2>&1; then
  log "ERROR: cosign tidak ditemukan. Update dibatalkan karena image tidak bisa diverifikasi."
  exit 1
fi

outdated=()
declare -A new_digest=()
for svc in "${SERVICES[@]}"; do
  repo="${REGISTRY_USER}/cbt-${svc}"
  remote=$(remote_digest "$repo" || true)
  if [ -z "$remote" ]; then
    log "Gagal membaca digest ${repo}:${TAG} dari Docker Hub, lewati."
    continue
  fi
  if [ "$remote" != "$(local_digest "$repo")" ]; then
    outdated+=("$svc")
    new_digest[$svc]="$remote"
  fi
done

if [ ${#outdated[@]} -eq 0 ]; then
  exit 0
fi
log "Ada image baru: ${outdated[*]}"

# Jangan restart saat ujian berjalan: siswa bisa kehilangan koneksi di tengah ujian.
if [ "${FORCE:-0}" != "1" ]; then
  active=$(docker exec "$DB_CONTAINER" psql -U "$DB_USER" -d "$DB_NAME" -tAc "
    SELECT
      (SELECT count(*) FROM exam_schedules
         WHERE is_active AND now() BETWEEN start_time AND end_time)
    + (SELECT count(*) FROM exam_sessions
         WHERE status IN ('IN_PROGRESS','BLOCKED') AND server_deadline > now())
  " 2>/dev/null | tr -d '[:space:]') || active=""

  if [ -z "$active" ]; then
    log "Tidak bisa memeriksa status ujian di database, update ditunda demi aman."
    exit 0
  fi
  if [ "$active" != "0" ]; then
    log "Ujian sedang berlangsung ($active jadwal/sesi aktif), update ditunda."
    exit 0
  fi
fi

verify_err=$(mktemp)
trap 'rm -f "$verify_err"' EXIT

verified=()
for svc in "${outdated[@]}"; do
  ref="${REGISTRY_USER}/cbt-${svc}@${new_digest[$svc]}"
  if ! "$COSIGN" verify \
      --certificate-identity "$SIGNER_IDENTITY" \
      --certificate-oidc-issuer "$SIGNER_ISSUER" \
      "$ref" >/dev/null 2>"$verify_err"; then
    log "PERINGATAN: tanda tangan ${ref} tidak valid atau belum ada, image TIDAK dipakai."
    sed 's/^/    /' "$verify_err" | tail -n 5
    continue
  fi
  log "Tanda tangan valid: ${ref}"
  docker pull -q "$ref" >/dev/null
  docker tag "$ref" "${REGISTRY_USER}/cbt-${svc}:${TAG}"
  verified+=("$svc")
done

if [ ${#verified[@]} -eq 0 ]; then
  log "Tidak ada image terverifikasi untuk dipasang. Dicoba lagi pada jadwal berikutnya."
  exit 0
fi
outdated=("${verified[@]}")

log "Menjalankan ulang container..."
docker compose up -d --no-deps "${outdated[@]}"

docker image prune -f >/dev/null
log "Update selesai."
docker compose ps
