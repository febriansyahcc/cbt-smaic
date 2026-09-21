import { onMounted, onUnmounted, ref } from 'vue'

export function useAntiCheat(examStore, onWarningModal) {
  const isFullscreen = ref(false)

  const handleVisibilityChange = () => {
    if (document.hidden) {
      triggerViolation('TAB_SWITCH', 'Meninggalkan tab atau membuka aplikasi lain')
    }
  }

  const handleWindowBlur = () => {
    triggerViolation('BLUR', 'Jendela browser kehilangan fokus (split-screen/status bar)')
  }

  const handleFullscreenChange = () => {
    isFullscreen.value = !!document.fullscreenElement
    if (!document.fullscreenElement && !examStore.isBlocked) {
      triggerViolation('FULLSCREEN_EXIT', 'Keluar dari mode layar penuh (Fullscreen)')
    }
  }

  const handleContextMenu = (e) => {
    e.preventDefault()
  }

  const handleKeyDown = (e) => {
    // Block inspect element (F12, Ctrl+Shift+I, Ctrl+Shift+J)
    if (
      e.key === 'F12' ||
      (e.ctrlKey && e.shiftKey && (e.key === 'I' || e.key === 'i' || e.key === 'J' || e.key === 'j')) ||
      (e.ctrlKey && (e.key === 'u' || e.key === 'U')) ||
      (e.ctrlKey && (e.key === 'c' || e.key === 'C')) ||
      (e.ctrlKey && (e.key === 'v' || e.key === 'V')) ||
      (e.ctrlKey && (e.key === 'a' || e.key === 'A'))
    ) {
      e.preventDefault()
    }
  }

  let lastTriggerTime = 0
  const triggerViolation = (type, message) => {
    const now = Date.now()
    // Debounce triggers by 1.5 seconds so a blur followed by visibilitychange counts as 1 violation
    if (now - lastTriggerTime < 1500) return
    lastTriggerTime = now

    examStore.reportViolation(type, message)
    if (onWarningModal) {
      onWarningModal(message)
    }
  }

  const enterFullscreen = async () => {
    try {
      const docEl = document.documentElement
      if (docEl.requestFullscreen) {
        await docEl.requestFullscreen()
      } else if (docEl.webkitRequestFullscreen) {
        await docEl.webkitRequestFullscreen()
      }
      isFullscreen.value = true
    } catch (e) {
      console.warn('Fullscreen request rejected by browser/user gesture')
    }
  }

  onMounted(() => {
    document.addEventListener('visibilitychange', handleVisibilityChange)
    window.addEventListener('blur', handleWindowBlur)
    document.addEventListener('fullscreenchange', handleFullscreenChange)
    document.addEventListener('contextmenu', handleContextMenu)
    window.addEventListener('keydown', handleKeyDown)
  })

  onUnmounted(() => {
    document.removeEventListener('visibilitychange', handleVisibilityChange)
    window.removeEventListener('blur', handleWindowBlur)
    document.removeEventListener('fullscreenchange', handleFullscreenChange)
    document.removeEventListener('contextmenu', handleContextMenu)
    window.removeEventListener('keydown', handleKeyDown)
  })

  return {
    isFullscreen,
    enterFullscreen,
  }
}
