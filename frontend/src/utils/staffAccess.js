// Fungsi murni untuk akses akun staf: role sebagai template izin + izin kustom per akun.
// Data templates dan implies berasal dari GET /admin/permission-catalog.

export const TEMPLATE_KEYS = {
  ADMIN: 'admin',
  GURU: 'guru',
  PENGAWAS: 'pengawas',
  CUSTOM: 'custom',
  // Katalog template belum termuat atau gagal: peran non-Administrator belum bisa dipastikan.
  UNKNOWN: 'unknown',
}

export const CUSTOM_LABEL = 'Kustom (dasar Guru)'
export const UNKNOWN_LABEL = 'Memuat...'

const FALLBACK_LABELS = {
  admin: 'Administrator',
  guru: 'Guru',
  pengawas: 'Pengawas',
}

const unique = (list) => [...new Set(Array.isArray(list) ? list : [])]

// Semua izin yang ikut diberikan (langsung maupun berantai) oleh satu izin, tanpa izin itu sendiri.
function impliedBy(perm, implies) {
  const result = new Set()
  const queue = [perm]
  while (queue.length) {
    const current = queue.pop()
    for (const dep of (implies && implies[current]) || []) {
      if (dep !== perm && !result.has(dep)) {
        result.add(dep)
        queue.push(dep)
      }
    }
  }
  return result
}

// Tambahkan seluruh izin turunan ke daftar izin. Urutan asli dipertahankan.
export function applyImplications(perms, implies) {
  const result = new Set(unique(perms))
  const queue = [...result]
  while (queue.length) {
    const current = queue.pop()
    for (const dep of (implies && implies[current]) || []) {
      if (!result.has(dep)) {
        result.add(dep)
        queue.push(dep)
      }
    }
  }
  return [...result]
}

// Izin lain yang sedang tercentang dan masih memerlukan `perm`.
export function dependentsOf(perm, perms, implies) {
  return unique(perms).filter((p) => p !== perm && impliedBy(p, implies).has(perm))
}

// Izin hanya boleh dilepas bila tidak ada izin tercentang lain yang memerlukannya.
export function canUncheck(perm, perms, implies) {
  return dependentsOf(perm, perms, implies).length === 0
}

// Dua daftar dianggap sama bila himpunannya sama, tanpa memedulikan urutan atau duplikat.
export function samePermissionSet(a, b) {
  const setA = new Set(unique(a))
  const setB = new Set(unique(b))
  if (setA.size !== setB.size) return false
  for (const p of setA) {
    if (!setB.has(p)) return false
  }
  return true
}

export function templateByKey(templates, key) {
  return (Array.isArray(templates) ? templates : []).find((t) => t.key === key) || null
}

// Izin efektif akun. Administrator = penuh. Izin kosong pada akun GURU lama = bawaan template Guru.
export function effectivePermissions(user, templates) {
  if (!user) return []
  if (user.role === 'ADMIN') return ['*']
  const own = unique(user.permissions)
  if (own.length > 0) return own
  return unique(templateByKey(templates, TEMPLATE_KEYS.GURU)?.permissions)
}

// Kunci template turunan: 'admin' | 'guru' | 'pengawas' | 'custom', atau 'unknown' bila katalog belum ada.
// Bila `implies` diberikan, kedua sisi dibandingkan setelah izin turunan ditambahkan.
export function templateKeyFor(user, templates, implies) {
  if (user?.role === 'ADMIN') return TEMPLATE_KEYS.ADMIN
  if (!Array.isArray(templates) || templates.length === 0) return TEMPLATE_KEYS.UNKNOWN
  const normalize = (perms) => (implies ? applyImplications(perms, implies) : unique(perms))
  const effective = normalize(effectivePermissions(user, templates))
  for (const key of [TEMPLATE_KEYS.GURU, TEMPLATE_KEYS.PENGAWAS]) {
    const tpl = templateByKey(templates, key)
    if (tpl && samePermissionSet(effective, normalize(tpl.permissions))) return key
  }
  return TEMPLATE_KEYS.CUSTOM
}

// Label turunan: Administrator, Guru, Pengawas, atau "Kustom (dasar Guru)".
export function templateLabelFor(user, templates, implies) {
  const key = templateKeyFor(user, templates, implies)
  if (key === TEMPLATE_KEYS.CUSTOM) return CUSTOM_LABEL
  if (key === TEMPLATE_KEYS.UNKNOWN) return UNKNOWN_LABEL
  return templateByKey(templates, key)?.label || FALLBACK_LABELS[key]
}
