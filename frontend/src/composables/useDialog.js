import { ref } from 'vue'

const dialogState = ref({
  isOpen: false,
  isConfirm: false,
  title: '',
  message: '',
  type: 'info', // 'danger' | 'warning' | 'info' | 'success'
  confirmText: 'Lanjutkan',
  cancelText: 'Batal',
  resolve: null
})

const toasts = ref([])
let toastIdCounter = 0

export function useDialog() {
  const confirm = (options = {}) => {
    return new Promise((resolve) => {
      dialogState.value = {
        isOpen: true,
        isConfirm: true,
        title: options.title || 'Konfirmasi Tindakan',
        message: options.message || 'Apakah Anda yakin ingin melanjutkan?',
        type: options.type || 'warning',
        confirmText: options.confirmText || (options.type === 'danger' ? 'Hapus' : 'Ya, Lanjutkan'),
        cancelText: options.cancelText || 'Batal',
        resolve
      }
    })
  }

  const alert = (options = {}) => {
    return new Promise((resolve) => {
      const isString = typeof options === 'string'
      dialogState.value = {
        isOpen: true,
        isConfirm: false,
        title: isString ? 'Informasi' : (options.title || 'Informasi'),
        message: isString ? options : (options.message || ''),
        type: isString ? 'info' : (options.type || 'info'),
        confirmText: isString ? 'Mengerti' : (options.confirmText || 'Mengerti'),
        cancelText: 'Tutup',
        resolve
      }
    })
  }

  const toast = (message, type = 'success', duration = 3000) => {
    const id = ++toastIdCounter
    toasts.value.push({ id, message, type })
    setTimeout(() => {
      toasts.value = toasts.value.filter(t => t.id !== id)
    }, duration)
  }

  const handleConfirm = () => {
    if (dialogState.value.resolve) {
      dialogState.value.resolve(true)
    }
    dialogState.value.isOpen = false
  }

  const handleCancel = () => {
    if (dialogState.value.resolve) {
      dialogState.value.resolve(false)
    }
    dialogState.value.isOpen = false
  }

  const removeToast = (id) => {
    toasts.value = toasts.value.filter(t => t.id !== id)
  }

  return {
    dialogState,
    toasts,
    confirm,
    alert,
    toast,
    handleConfirm,
    handleCancel,
    removeToast
  }
}
