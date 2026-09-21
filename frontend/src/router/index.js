import { createRouter, createWebHistory } from 'vue-router'
import LoginView from '../views/LoginView.vue'
import StudentHomeView from '../views/student/StudentHomeView.vue'
import ExamView from '../views/student/ExamView.vue'
import ExamFinishedView from '../views/student/ExamFinishedView.vue'
import AdminDashboardView from '../views/admin/AdminDashboardView.vue'
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
    component: AdminDashboardView,
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

export default router
