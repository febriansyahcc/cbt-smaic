<template>
  <router-view />
  <GlobalModal />
</template>

<script setup>
import { onMounted, onUnmounted } from 'vue'
import GlobalModal from '@/components/common/GlobalModal.vue'

// localStorage dibagi antar tab pada origin yang sama. Bila siswa memindai QR kartu dua kali lalu
// masuk di salah satu tab, tab lain masih memegang identitas sesi yang sudah digantikan dan
// request-nya akan ditolak. Muat ulang tab itu agar mengikuti sesi yang sekarang aktif; jawaban
// yang belum tersinkron tetap aman karena antreannya disimpan di localStorage.
const onSessionTokenChanged = (e) => {
  if (e.key !== 'cbt_token' || e.oldValue === e.newValue) return
  // Token dihapus (logout / sesi dibersihkan) sudah ditangani interceptor di tab ini sendiri.
  if (!e.newValue) return
  window.location.reload()
}

onMounted(() => window.addEventListener('storage', onSessionTokenChanged))
onUnmounted(() => window.removeEventListener('storage', onSessionTokenChanged))
</script>
