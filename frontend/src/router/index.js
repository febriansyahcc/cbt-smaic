import { createRouter, createWebHistory } from 'vue-router'
import LoginView from '../views/LoginView.vue'
import StudentHomeView from '../views/student/StudentHomeView.vue'
import ExamView from '../views/student/ExamView.vue'
import ExamFinishedView from '../views/student/ExamFinishedView.vue'
import { homePathFor, isStaffUser } from '../utils/access'
import { useAuthStore } from '../stores/auth'

const routes = [
  {
    path: '/',
    redirect: () => {
      const userRaw = localStorage.getItem('cbt_user')
      if (userRaw) {
        try {
          return homePathFor(JSON.parse(userRaw))
        } catch (e) {
          return '/login'
        }
      }
      return '/login'
    }
  },
  {
    path: '/login',
    name: 'login',
    component: LoginView,
  },
  {
    path: '/admin',
    name: 'admin-dashboard',
    // Dimuat terpisah: siswa tidak perlu mengunduh kode dasbor staf. Halaman siswa tetap di
    // bundle utama agar alur ujian tidak bergantung pada unduhan chunk di tengah jalan.
    component: () => import('../views/admin/AdminDashboardView.vue'),
    meta: { requiresAuth: true, staff: true }
  },
  {
    path: '/student',
    name: 'student-home',
    component: StudentHomeView,
    meta: { requiresAuth: true, role: 'SISWA' }
  },
  {
    path: '/exam',
    name: 'exam',
    component: ExamView,
    meta: { requiresAuth: true, role: 'SISWA' }
  },
  {
    path: '/exam-finished',
    name: 'exam-finished',
    component: ExamFinishedView,
    meta: { requiresAuth: true }
  },
  {
    path: '/proctor',
    redirect: '/admin?tab=proctor'
  },
  {
    path: '/admin/questions',
    redirect: '/admin?tab=questions'
  }
]

const router = createRouter({
  history: createWebHistory(),
  routes,
})

let sessionValidated = false

// Token di localStorage bisa milik sesi yang sudah berakhir atau — pada komputer sekolah yang
// dipakai bergantian — milik siswa sebelumnya yang menutup browser tanpa logout. Sekali per
// pemuatan halaman, token itu divalidasi ke server sebelum halaman apa pun dirender, sehingga
// identitas yang dipakai selalu yang diakui server. Kegagalan jaringan sengaja tidak memblokir
// (ujian harus tetap bisa dimasuki saat jaringan tersendat); hanya 401 yang membersihkan sesi,
// dan itu ditangani interceptor di services/api.js.
const validateStoredSession = async () => {
  if (sessionValidated) return
  // Perangkat sedang offline: validasi dilewati tanpa menandai selesai, agar alur ujian tetap
  // bisa dibuka dari data lokal dan pemeriksaan terjadi saat jaringan kembali.
  if (typeof navigator !== 'undefined' && navigator.onLine === false) return
  sessionValidated = true
  try {
    await useAuthStore().fetchMe({ timeout: 8000 })
  } catch (e) {
    // Jaringan tersendat: lanjutkan dengan data lokal, jangan halangi siswa masuk ujian.
  }
}

router.beforeEach(async (to, from, next) => {
  if (!to.meta.requiresAuth) {
    return next()
  }

  if (!localStorage.getItem('cbt_token') || !localStorage.getItem('cbt_user')) {
    return next('/login')
  }

  await validateStoredSession()

  // Dibaca ulang: validasi di atas bisa mengganti atau menghapus data sesi.
  const userRaw = localStorage.getItem('cbt_user')
  if (!localStorage.getItem('cbt_token') || !userRaw) {
    return next('/login')
  }

  let user
  try {
    user = JSON.parse(userRaw)
  } catch (e) {
    return next('/login')
  }

  if (to.meta.staff && !isStaffUser(user)) {
    const home = homePathFor(user)
    return next(home === to.fullPath ? '/login' : home)
  }
  if (to.meta.role) {
    const allowed = Array.isArray(to.meta.role) ? to.meta.role : [to.meta.role]
    if (!allowed.includes(user.role) && user.role !== 'ADMIN') {
      return next('/login')
    }
  }
  next()
})

// Setelah redeploy, tab lama bisa merujuk chunk yang sudah tidak ada. Muat ulang halaman
// tujuan sekali agar index.html dan nama chunk terbaru terambil.
router.onError((err, to) => {
  const isChunkError = /dynamically imported module|Importing a module script failed/i.test(err?.message || '')
  if (!isChunkError) return
  try {
    if (sessionStorage.getItem('cbt_chunk_reload') === to.fullPath) return
    sessionStorage.setItem('cbt_chunk_reload', to.fullPath)
  } catch (e) {
    // sessionStorage tidak tersedia: tetap coba muat ulang sekali
  }
  window.location.assign(to.fullPath)
})

router.afterEach(() => {
  try {
    sessionStorage.removeItem('cbt_chunk_reload')
  } catch (e) {
    // abaikan
  }
})

export default router
