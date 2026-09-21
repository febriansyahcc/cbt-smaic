// Pemetaan menu portal staf ke izin (PBAC). Satu sumber kebenaran untuk router dan sidebar.
// Daftar kosong berarti terbuka bagi semua staf. ADMIN dan izin "*" selalu lolos.
export const TAB_ACCESS = {
  dashboard: [],
  events: [],
  schedules: ['schedules:read', 'schedules:manage'],
  questions: ['questions:read_assigned', 'questions:read_all', 'questions:upload', 'questions:manage'],
  proctor: ['proctor:view', 'proctor:control', 'proctor:control_all'],
  students: ['master:manage'],
  teachers: ['master:manage'],
  classes: ['master:manage'],
  subjects: ['master:manage'],
  'class-subjects': ['master:manage'],
}

// Urutan menu yang dipilih sebagai halaman awal untuk non-admin.
const LANDING_ORDER = ['questions', 'proctor', 'schedules', 'dashboard']

export function hasAnyPermission(user, perms) {
  if (!user) return false
  if (user.role === 'ADMIN') return true
  const owned = user.permissions || []
  if (owned.includes('*')) return true
  return perms.some((p) => owned.includes(p))
}

export function isStaffUser(user) {
  if (!user || user.role === 'SISWA') return false
  if (user.role === 'ADMIN') return true
  return (user.permissions || []).some((p) => p !== 'exam:take')
}

export function canAccessTab(user, tab) {
  const required = TAB_ACCESS[tab]
  if (!required) return false
  if (!isStaffUser(user)) return false
  return required.length === 0 || hasAnyPermission(user, required)
}

export function landingTab(user) {
  if (user?.role === 'ADMIN') return 'dashboard'
  return LANDING_ORDER.find((t) => canAccessTab(user, t)) || 'dashboard'
}

export function homePathFor(user) {
  if (!user) return '/login'
  if (user.role === 'SISWA') return '/student'
  if (!isStaffUser(user)) return '/login'
  const tab = landingTab(user)
  return tab === 'dashboard' ? '/admin' : `/admin?tab=${tab}`
}
