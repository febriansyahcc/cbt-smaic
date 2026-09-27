import { createRouter, createWebHistory } from 'vue-router'
import LoginView from '../views/LoginView.vue'
import StudentHomeView from '../views/student/StudentHomeView.vue'
import ExamView from '../views/student/ExamView.vue'
import ExamFinishedView from '../views/student/ExamFinishedView.vue'
import { homePathFor, isStaffUser } from '../utils/access'

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

router.beforeEach((to, from, next) => {
  const token = localStorage.getItem('cbt_token')
  const userRaw = localStorage.getItem('cbt_user')

  if (to.meta.requiresAuth) {
    if (!token || !userRaw) {
      return next('/login')
    }
    const user = JSON.parse(userRaw)
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
