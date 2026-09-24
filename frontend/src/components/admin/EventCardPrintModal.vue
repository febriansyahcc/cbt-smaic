<template>
  <Teleport to="body">
    <div v-if="rendered" class="fixed inset-0 z-[60] flex items-center justify-center p-2 sm:p-4">
      <!-- Backdrop: hanya fade opacity -->
      <Transition
        appear
        enter-active-class="transition-opacity duration-200 ease-out"
        enter-from-class="opacity-0"
        enter-to-class="opacity-100"
        leave-active-class="transition-opacity duration-150 ease-in"
        leave-from-class="opacity-100"
        leave-to-class="opacity-0"
      >
        <div v-if="modelValue" class="fixed inset-0 bg-slate-900/60 backdrop-blur-sm" @click="close"></div>
      </Transition>

      <!-- Kartu modal: scale + translate halus -->
      <Transition
        appear
        enter-active-class="transition duration-200 ease-out transform"
        enter-from-class="opacity-0 scale-95 translate-y-2"
        enter-to-class="opacity-100 scale-100 translate-y-0"
        leave-active-class="transition duration-150 ease-in transform"
        leave-from-class="opacity-100 scale-100 translate-y-0"
        leave-to-class="opacity-0 scale-95 translate-y-2"
        @after-leave="rendered = false"
      >
        <div
          v-if="modelValue"
          class="relative z-10 bg-white rounded-3xl max-w-4xl w-full shadow-2xl border border-slate-100 flex flex-col max-h-[95vh] overflow-hidden"
          role="dialog"
          aria-modal="true"
          @click.stop
        >
          <!-- Header -->
          <div class="flex items-start justify-between gap-3 px-5 sm:px-6 py-4 border-b border-slate-100 shrink-0">
            <div class="min-w-0">
              <h3 class="text-base font-bold text-slate-900">Cetak {{ cardTypeLabel }}</h3>
              <p class="text-[11px] text-slate-500 mt-0.5 truncate">
                {{ event?.title }} &middot; T.A. {{ event?.academic_year }}
              </p>
            </div>
            <button
              type="button"
              @click="close"
              class="p-1.5 rounded-xl text-slate-400 hover:text-slate-600 hover:bg-slate-100 transition active:scale-95 cursor-pointer shrink-0"
              title="Tutup"
            >
              <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
              </svg>
            </button>
          </div>

          <!-- Toolbar -->
          <div class="px-5 sm:px-6 py-3 border-b border-slate-100 flex flex-wrap items-center gap-2 shrink-0 text-xs">
            <div class="flex items-center gap-1 bg-slate-100 p-1 rounded-xl" role="tablist" aria-label="Jenis kartu">
              <button
                v-for="opt in cardTypeOptions"
                :key="opt.value"
                type="button"
                role="tab"
                :aria-selected="cardType === opt.value"
                @click="cardType = opt.value"
                :class="[
                  'px-3 py-1.5 rounded-lg text-xs font-bold transition active:scale-95 cursor-pointer',
                  cardType === opt.value ? 'bg-white text-emerald-700 shadow-xs' : 'text-slate-500 hover:text-slate-800'
                ]"
              >
                {{ opt.label }}
              </button>
            </div>
            <select
              v-model="classFilter"
              class="px-3 py-2 rounded-xl border border-slate-300 bg-white text-xs font-semibold text-slate-700 focus:ring-2 focus:ring-indigo-600 focus:border-indigo-600"
            >
              <option value="">Semua kelas</option>
              <option v-for="c in classOptions" :key="c.id" :value="c.id">{{ c.name }}</option>
            </select>
            <span class="text-slate-500">
              {{ printable.length }} kartu &middot; {{ pageCount }} lembar A4
            </span>
            <div class="flex-1"></div>
            <button
              v-if="participants.length > 0 && missingCount === 0"
              type="button"
              @click="regenerate"
              :disabled="isGenerating"
              class="px-3 py-2 rounded-xl text-xs font-semibold text-slate-600 hover:bg-slate-100 transition active:scale-95 cursor-pointer disabled:opacity-50"
            >
              Buat ulang nomor &amp; password
            </button>
          </div>

          <!-- Body -->
          <div class="flex-1 min-h-0 overflow-hidden flex flex-col bg-slate-100">
            <div v-if="isLoading" class="py-16 flex flex-col items-center gap-2 text-slate-400 text-xs">
              <svg class="w-5 h-5 animate-spin" fill="none" viewBox="0 0 24 24">
                <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
                <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8v4a4 4 0 00-4 4H4z"></path>
              </svg>
              <span>Menyiapkan kartu peserta…</span>
            </div>

            <div v-else-if="loadError" class="py-16 px-6 text-center text-xs text-rose-600">{{ loadError }}</div>

            <div v-else-if="participants.length === 0" class="py-16 px-6 text-center space-y-1">
              <p class="text-sm font-bold text-slate-700">Belum ada peserta</p>
              <p class="text-xs text-slate-500">Peserta diambil dari siswa di kelas yang memiliki jadwal pada event ini. Tambahkan jadwal ujian terlebih dahulu.</p>
            </div>

            <template v-else>
              <div
                v-if="missingCount > 0"
                class="m-4 mb-0 p-3 rounded-2xl bg-amber-50 border border-amber-200 flex flex-wrap items-center gap-3 text-xs"
              >
                <p class="flex-1 min-w-[12rem] text-amber-800">
                  <strong>{{ missingCount }} siswa</strong> belum memiliki nomor ujian dan password. Buat sekarang agar kartunya bisa dicetak.
                </p>
                <button
                  type="button"
                  @click="generate(false)"
                  :disabled="isGenerating"
                  class="px-4 py-2 rounded-xl bg-amber-600 hover:bg-amber-700 text-white font-bold transition active:scale-95 cursor-pointer disabled:opacity-50"
                >
                  {{ isGenerating ? 'Membuat…' : 'Buat Nomor Ujian' }}
                </button>
              </div>

              <div v-if="printable.length === 0" class="py-16 px-6 text-center text-xs text-slate-500">
                Tidak ada kartu siap cetak untuk kelas ini.
              </div>
              <iframe
                v-else
                ref="previewFrame"
                :srcdoc="documentHtml"
                :title="`Pratinjau ${cardTypeLabel.toLowerCase()}`"
                class="flex-1 w-full min-h-[50vh] border-0 bg-slate-100"
                @load="onFrameLoad"
              ></iframe>
            </template>
          </div>

          <!-- Footer -->
          <div class="px-5 sm:px-6 py-3 border-t border-slate-100 flex flex-wrap items-center justify-between gap-3 shrink-0">
            <p class="text-[11px] text-slate-500">
              Kertas <strong>A4</strong>, margin <strong>Default/None</strong>, skala 100%.
              <template v-if="cardType === 'meja'">Kartu meja tidak memuat password; siswa memakai kartu peserta dari wali kelas untuk login.</template>
              <template v-else>Siswa login dengan No. Ujian &amp; password selama event aktif.</template>
            </p>
            <div class="flex items-center gap-2 ml-auto">
              <button
                type="button"
                @click="close"
                class="px-4 py-2 bg-slate-100 hover:bg-slate-200 rounded-xl text-xs font-semibold text-slate-700 transition active:scale-95 cursor-pointer"
              >
                Tutup
              </button>
              <button
                type="button"
                @click="printCards"
                :disabled="!frameReady || printable.length === 0"
                class="px-4 py-2 bg-emerald-600 hover:bg-emerald-700 rounded-xl text-xs font-bold text-white transition active:scale-95 cursor-pointer disabled:opacity-50 disabled:cursor-not-allowed"
              >
                Cetak {{ printable.length }} {{ cardTypeLabel }}
              </button>
            </div>
          </div>
        </div>
      </Transition>
    </div>
  </Teleport>
</template>

<script setup>
import { ref, computed, watch } from 'vue'
import QRCode from 'qrcode'
import api from '../../services/api'
import { useDialog } from '../../composables/useDialog'

// Modal cetak kartu peserta per event. Kartu dirender sebagai dokumen HTML terpisah di dalam
// iframe (pratinjau sekaligus sumber cetak) agar gaya aplikasi tidak ikut tercetak.
// Tata letak: 8 kartu per lembar A4 (2 kolom x 4 baris).
const props = defineProps({
  modelValue: { type: Boolean, default: false },
  event: { type: Object, default: null },
})
const emit = defineEmits(['update:modelValue'])

const { confirm: showConfirmModal, toast: showToast } = useDialog()

const CARDS_PER_PAGE = 8

// Kartu peserta (dengan password, dibagikan wali kelas) dan kartu meja (ditempel di meja,
// tanpa password). Kartu meja dibedakan sekilas lewat judul berpita gelap dan No. Ujian besar,
// sehingga tetap terbedakan walau dicetak hitam-putih.
const cardTypeOptions = [
  { value: 'peserta', label: 'Kartu Peserta' },
  { value: 'meja', label: 'Kartu Meja' },
]
const cardType = ref('peserta')
const cardTypeLabel = computed(() => cardTypeOptions.find((o) => o.value === cardType.value)?.label || 'Kartu')
const SCHOOL_NAME = 'SMA ISLAMIC CENTRE DEMAK'

const rendered = ref(false)
const isLoading = ref(false)
const isGenerating = ref(false)
const loadError = ref('')
const participants = ref([])
const classFilter = ref('')
const qrMap = ref({})
const logoDataUrl = ref('')
const previewFrame = ref(null)
const frameReady = ref(false)

let loadToken = 0

const classOptions = computed(() => {
  const seen = new Map()
  for (const p of participants.value) {
    if (!seen.has(p.class_id)) seen.set(p.class_id, { id: p.class_id, name: p.class_name })
  }
  return [...seen.values()]
})

const missingCount = computed(() => participants.value.filter((p) => !p.exam_number).length)

const printable = computed(() =>
  participants.value.filter((p) => p.exam_number && (!classFilter.value || p.class_id === classFilter.value))
)

const pageCount = computed(() => Math.ceil(printable.value.length / CARDS_PER_PAGE))

const close = () => emit('update:modelValue', false)

const escapeHtml = (value) =>
  String(value ?? '').replace(/[&<>"']/g, (ch) => ({ '&': '&amp;', '<': '&lt;', '>': '&gt;', '"': '&quot;', "'": '&#39;' }[ch]))

const loadImage = (src) =>
  new Promise((resolve, reject) => {
    const img = new Image()
    img.onload = () => resolve(img)
    img.onerror = reject
    img.src = src
  })

let logoImage = null

// Logo sebagai data URL agar pasti termuat di dokumen iframe sebelum dicetak.
const ensureLogo = async () => {
  if (logoImage) return
  try {
    logoImage = await loadImage('/logo-smic.png')
    const canvas = document.createElement('canvas')
    canvas.width = logoImage.naturalWidth
    canvas.height = logoImage.naturalHeight
    canvas.getContext('2d').drawImage(logoImage, 0, 0)
    logoDataUrl.value = canvas.toDataURL('image/png')
  } catch {
    logoImage = null
    logoDataUrl.value = ''
  }
}

// QR berisi tautan login dengan username terisi (tanpa password), logo sekolah di tengah.
const buildQr = async (examNumber) => {
  const url = `${window.location.origin}/login?u=${encodeURIComponent(examNumber)}`
  const canvas = document.createElement('canvas')
  await QRCode.toCanvas(canvas, url, { errorCorrectionLevel: 'H', margin: 0, width: 240, color: { dark: '#000000', light: '#ffffff' } })
  if (logoImage) {
    const ctx = canvas.getContext('2d')
    const size = canvas.width * 0.24
    const pos = (canvas.width - size) / 2
    ctx.fillStyle = '#ffffff'
    ctx.fillRect(pos - 4, pos - 4, size + 8, size + 8)
    ctx.drawImage(logoImage, pos, pos, size, size)
  }
  return canvas.toDataURL('image/png')
}

const buildQrCodes = async (token) => {
  const map = {}
  for (const p of participants.value) {
    if (!p.exam_number) continue
    map[p.exam_number] = await buildQr(p.exam_number)
    if (token !== loadToken) return
  }
  qrMap.value = map
}

const load = async () => {
  if (!props.event?.id) return
  const token = ++loadToken
  isLoading.value = true
  loadError.value = ''
  frameReady.value = false
  try {
    const [res] = await Promise.all([api.get(`/admin/events/${props.event.id}/participants`), ensureLogo()])
    if (token !== loadToken) return
    participants.value = res.data?.data || []
    if (classFilter.value && !classOptions.value.some((c) => c.id === classFilter.value)) classFilter.value = ''
    await buildQrCodes(token)
  } catch (err) {
    if (token !== loadToken) return
    loadError.value = err.response?.data?.message || 'Gagal memuat data peserta.'
  } finally {
    if (token === loadToken) isLoading.value = false
  }
}

const generate = async (reset) => {
  if (!props.event?.id || isGenerating.value) return
  isGenerating.value = true
  try {
    const res = await api.post(`/admin/events/${props.event.id}/participants/generate`, { reset })
    showToast(res.data?.message || 'Nomor ujian berhasil dibuat', 'success')
    await load()
  } catch (err) {
    showToast(err.response?.data?.message || 'Gagal membuat nomor ujian', 'error')
  } finally {
    isGenerating.value = false
  }
}

const regenerate = async () => {
  const ok = await showConfirmModal({
    title: 'Buat Ulang Nomor & Password?',
    message: 'Seluruh nomor ujian dan password peserta event ini akan diganti. Kartu yang sudah dicetak tidak berlaku lagi dan harus dicetak ulang.',
    type: 'danger',
    confirmText: 'Buat Ulang',
  })
  if (ok) await generate(true)
}

const academicYearLabel = computed(() => String(props.event?.academic_year || '').replace('/', '-'))

const LONG_NAME = 26

const cardHtml = (p) => {
  const isDesk = cardType.value === 'meja'
  return `
  <div class="card${String(p.full_name || '').length > LONG_NAME ? ' long' : ''}${isDesk ? ' desk' : ''}">
    ${logoDataUrl.value ? `<img class="wm" src="${logoDataUrl.value}" alt="">` : ''}
    <div class="inner">
      <div class="head">
        <div>${escapeHtml(SCHOOL_NAME)}</div>
        <div class="title">${isDesk ? 'KARTU MEJA' : 'KARTU PESERTA'}</div>
        <div>${escapeHtml(String(props.event?.title || '').toUpperCase())}</div>
        <div>TAHUN AJARAN ${escapeHtml(academicYearLabel.value)}</div>
      </div>
      <div class="rule"></div>
      <table class="rows">
        <tr><td class="k">Nama Peserta</td><td class="v${String(p.full_name || '').length > LONG_NAME ? ' name' : ''}">: ${escapeHtml(String(p.full_name || '').toUpperCase())}</td></tr>
        <tr><td class="k">Kelas</td><td class="v">: ${escapeHtml(p.class_name)}</td></tr>
        <tr><td class="k">No. Ujian</td><td class="v">: ${escapeHtml(p.exam_number)}</td></tr>
        <tr><td class="k">Username</td><td class="v">: ${escapeHtml(p.exam_number)}</td></tr>
        ${isDesk ? '' : `<tr><td class="k">Password</td><td class="v">: ${escapeHtml(p.password)}</td></tr>`}
      </table>
      <div class="foot">
        <div class="qr">${qrMap.value[p.exam_number] ? `<img src="${qrMap.value[p.exam_number]}" alt="">` : ''}</div>
        <div class="side">
          ${isDesk ? `<div class="big-no">${escapeHtml(p.exam_number)}</div>` : ''}
          <div class="sign">TTD PANITIA</div>
        </div>
      </div>
    </div>
  </div>`
}

const documentHtml = computed(() => {
  const pages = []
  for (let i = 0; i < printable.value.length; i += CARDS_PER_PAGE) {
    pages.push(`<section class="page">${printable.value.slice(i, i + CARDS_PER_PAGE).map(cardHtml).join('')}</section>`)
  }
  return `<!DOCTYPE html>
<html lang="id"><head><meta charset="utf-8"><title>${escapeHtml(cardTypeLabel.value)} - ${escapeHtml(props.event?.title || '')}</title>
<style>
  @page { size: A4 portrait; margin: 0; }
  * { box-sizing: border-box; -webkit-print-color-adjust: exact; print-color-adjust: exact; }
  html, body { margin: 0; padding: 0; }
  body { font-family: 'Times New Roman', Times, serif; color: #000; background: #e2e8f0; }
  .page { width: 210mm; height: 297mm; padding: 8mm; display: grid; grid-template-columns: 95mm 95mm;
    grid-template-rows: repeat(4, 68mm); column-gap: 4mm; row-gap: 3mm; background: #fff; overflow: hidden;
    margin: 16px auto; box-shadow: 0 1px 4px rgba(15, 23, 42, .15); page-break-after: always; break-after: page; }
  .page:last-child { page-break-after: auto; break-after: auto; }
  .card { position: relative; border: 1.2mm solid #4d7a2e; padding: 0.8mm; overflow: hidden; background: #fff; }
  .card .inner { position: relative; height: 100%; border: 0.3mm solid #4d7a2e; padding: 2mm 3mm 2mm; display: flex; flex-direction: column; }
  .wm { position: absolute; left: 50%; top: 54%; width: 58mm; transform: translate(-50%, -50%); opacity: .1; pointer-events: none; }
  .head { position: relative; text-align: center; font-weight: bold; font-size: 9.5pt; line-height: 1.12; }
  .rule { position: relative; margin: 1.2mm 0 1mm; border-top: 0.9mm solid #2f5320; height: 1.3mm; border-bottom: 0.4mm solid #2f5320; }
  .rows { position: relative; border-collapse: collapse; font-size: 10pt; line-height: 1.15; width: 100%; table-layout: fixed; }
  .rows td { padding: 0; vertical-align: top; }
  .rows .k { width: 25mm; white-space: nowrap; }
  .rows .v { overflow-wrap: anywhere; }
  .rows .name { font-size: 9pt; }
  .foot { position: relative; flex: 1; display: flex; align-items: flex-end; justify-content: space-between; min-height: 0; padding-top: 1mm; }
  .qr { width: 19mm; height: 19mm; border: 0.7mm solid #2f6b3a; border-radius: 1.2mm; padding: 0.8mm; background: #fff; }
  .qr img { width: 100%; height: 100%; display: block; }
  .card.long .qr { width: 15.5mm; height: 15.5mm; }
  .side { display: flex; flex-direction: column; align-items: flex-end; justify-content: flex-end; gap: 2mm; min-width: 0; }
  .sign { font-weight: bold; font-size: 9.5pt; padding: 0 4mm 2mm 0; }
  .desk .title { display: inline-block; margin: 0.4mm 0; padding: 0.5mm 5mm; background: #1f3a16; color: #fff; letter-spacing: 0.08em; }
  .desk .big-no { font-weight: bold; font-size: 22pt; line-height: 1; letter-spacing: 0.02em; padding-right: 2mm; white-space: nowrap; }
  @media print {
    body { background: #fff; }
    .page { margin: 0; box-shadow: none; }
  }
</style></head>
<body>${pages.join('')}</body></html>`
})

// Pratinjau diperkecil agar lembar A4 muat pada lebar modal (termasuk layar ponsel).
const onFrameLoad = () => {
  const frame = previewFrame.value
  const doc = frame?.contentDocument
  if (!doc?.body) return
  const pageWidthPx = 210 * (96 / 25.4) + 32
  const scale = Math.min(1, frame.clientWidth / pageWidthPx)
  doc.body.style.zoom = scale < 1 ? String(scale) : ''
  frameReady.value = true
}

const printCards = () => {
  const win = previewFrame.value?.contentWindow
  if (!win) return
  // Zoom pratinjau tidak boleh ikut memengaruhi ukuran cetak.
  const body = win.document.body
  const zoom = body.style.zoom
  body.style.zoom = ''
  const restore = () => {
    body.style.zoom = zoom
    win.removeEventListener('afterprint', restore)
  }
  win.addEventListener('afterprint', restore)
  win.focus()
  win.print()
}

watch(documentHtml, () => {
  frameReady.value = false
})

watch(
  () => props.modelValue,
  (open) => {
    if (open) {
      rendered.value = true
      classFilter.value = ''
      participants.value = []
      qrMap.value = {}
      load()
    } else {
      loadToken++
    }
  },
  { immediate: true }
)
</script>
