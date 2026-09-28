import axios from 'axios'

const TOKEN_KEY = 'cbt_token'

const api = axios.create({
  baseURL: '/api/v1',
  headers: {
    'Content-Type': 'application/json',
  },
})

api.interceptors.request.use((config) => {
  const token = localStorage.getItem(TOKEN_KEY)
  if (token) {
    config.headers.Authorization = `Bearer ${token}`
    // Token yang dipakai dicatat pada request, supaya respons 401 yang datang terlambat bisa
    // dibandingkan dengan token yang sedang aktif (lihat interceptor respons).
    config.cbtToken = token
  }
  return config
})

const setLoginNotice = (message) => {
  try {
    sessionStorage.setItem('cbt_login_notice', message)
  } catch (e) {
    // Mode privat / penyimpanan diblokir: pesan dilewati, alur login tetap jalan.
  }
}

const endSession = () => {
  localStorage.removeItem('cbt_token')
  localStorage.removeItem('cbt_user')
  localStorage.removeItem('cbt_profile')
  if (window.location.pathname !== '/login') {
    window.location.href = '/login'
  }
}

api.interceptors.response.use(
  (response) => response,
  (error) => {
    if (!error.response || error.response.status !== 401) {
      return Promise.reject(error)
    }

    // Kegagalan pada endpoint login bukan sesi yang berakhir; pesannya ditangani halaman login.
    if ((error.config?.url || '').includes('/auth/login')) {
      return Promise.reject(error)
    }

    // Di aula ujian, request bisa menggantung puluhan detik (jaringan padat, layar HP mati,
    // antrean jawaban offline, tab lain yang masih hidup). Respons 401 dari token lama bisa tiba
    // beberapa detik SETELAH siswa berhasil login ulang. Menghapus sesi berdasarkan respons itu
    // membuang token baru yang sah dan melempar siswa kembali ke halaman login berulang kali,
    // yang di kursi siswa tidak bisa dibedakan dari "password saya ditolak". Karena itu sesi
    // hanya diakhiri bila 401 benar-benar milik token yang sekarang dipakai.
    const usedToken = error.config?.cbtToken || null
    const currentToken = localStorage.getItem(TOKEN_KEY)
    if (usedToken && currentToken && usedToken !== currentToken) {
      return Promise.reject(error)
    }

    const code = error.response.data?.code
    if (code === 'CONCURRENT_LOGIN') {
      setLoginNotice('Akun Anda dipakai masuk di perangkat atau tab lain, jadi sesi di sini diakhiri. Password Anda tidak berubah — masuk lagi dengan No. Ujian dan password kartu peserta.')
    } else if (code === 'SESSION_ENDED') {
      setLoginNotice('Sesi perangkat Anda direset pengawas. Password Anda tidak berubah — masuk lagi dengan No. Ujian dan password kartu peserta.')
    }
    endSession()
    return Promise.reject(error)
  }
)

export default api
