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
        @after-leave="onAfterLeave"
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
              <h3 class="text-base font-bold text-slate-900">Cetak {{ docTypeLabel }}</h3>
              <p class="text-[11px] text-slate-500 mt-0.5 truncate">
                {{ bankInfo?.title || bank?.title }} &middot; {{ bankInfo?.subject?.name || bank?.subject?.name || '-' }}
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
            <div class="flex items-center gap-1 bg-slate-100 p-1 rounded-xl" role="tablist" aria-label="Jenis dokumen">
              <button
                v-for="opt in docTypeOptions"
                :key="opt.value"
                type="button"
                role="tab"
                :aria-selected="docType === opt.value"
                @click="docType = opt.value"
                :class="[
                  'px-3 py-1.5 rounded-lg text-xs font-bold transition active:scale-95 cursor-pointer',
                  docType === opt.value ? 'bg-white text-emerald-700 shadow-xs' : 'text-slate-500 hover:text-slate-800'
                ]"
              >
                {{ opt.label }}
              </button>
            </div>
            <select
              v-model="paper"
              class="px-3 py-2 rounded-xl border border-slate-300 bg-white text-xs font-semibold text-slate-700 focus:ring-2 focus:ring-indigo-600 focus:border-indigo-600"
              aria-label="Ukuran kertas"
            >
              <option v-for="(p, key) in PAPERS" :key="key" :value="key">{{ p.label }}</option>
            </select>
            <template v-if="docType === 'naskah'">
              <select
                v-model.number="columns"
                class="px-3 py-2 rounded-xl border border-slate-300 bg-white text-xs font-semibold text-slate-700 focus:ring-2 focus:ring-indigo-600 focus:border-indigo-600"
                aria-label="Jumlah kolom"
              >
                <option :value="1">1 kolom</option>
                <option :value="2">2 kolom</option>
              </select>
              <label v-if="hasEssay" class="inline-flex items-center gap-1.5 px-2 py-2 font-semibold text-slate-600 cursor-pointer select-none">
                <input v-model="essayLines" type="checkbox" class="rounded border-slate-300 text-emerald-600 focus:ring-emerald-600" />
                Ruang jawab uraian
              </label>
            </template>
            <span class="text-slate-500 ml-auto">{{ questions.length }} soal &middot; urutan sesuai bank</span>
          </div>

          <!-- Body -->
          <div class="flex-1 min-h-0 overflow-hidden flex flex-col bg-slate-100">
            <div v-if="isLoading" class="py-16 flex flex-col items-center gap-2 text-slate-400 text-xs">
              <svg class="w-5 h-5 animate-spin" fill="none" viewBox="0 0 24 24">
                <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
                <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8v4a4 4 0 00-4 4H4z"></path>
              </svg>
              <span>Menyiapkan naskah…</span>
            </div>

            <div v-else-if="loadError" class="py-16 px-6 text-center text-xs text-rose-600">{{ loadError }}</div>

            <div v-else-if="questions.length === 0" class="py-16 px-6 text-center space-y-1">
              <p class="text-sm font-bold text-slate-700">Bank soal masih kosong</p>
              <p class="text-xs text-slate-500">Tambahkan butir soal terlebih dahulu sebelum mencetak.</p>
            </div>

            <iframe
              v-else
              ref="previewFrame"
              :srcdoc="documentHtml"
              sandbox="allow-same-origin allow-modals"
              :title="`Pratinjau ${docTypeLabel.toLowerCase()}`"
              class="flex-1 w-full min-h-[50vh] border-0 bg-slate-100"
              @load="onFrameLoad"
            ></iframe>
          </div>

          <!-- Footer -->
          <div class="px-5 sm:px-6 py-3 border-t border-slate-100 flex flex-wrap items-center justify-between gap-3 shrink-0">
            <p class="text-[11px] text-slate-500">
              Kertas <strong>{{ PAPERS[paper].label }}</strong>, margin <strong>Default</strong>, skala 100%.
              <template v-if="docType === 'kunci'">Dokumen rahasia; simpan terpisah dari naskah soal.</template>
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
                @click="printDocument"
                :disabled="!frameReady || questions.length === 0"
                class="px-4 py-2 bg-emerald-600 hover:bg-emerald-700 rounded-xl text-xs font-bold text-white transition active:scale-95 cursor-pointer disabled:opacity-50 disabled:cursor-not-allowed"
              >
                Cetak {{ docTypeLabel }}
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
import katexCssUrl from 'katex/dist/katex.min.css?url'
import api from '../../services/api'
import { formatRichContent } from '../../utils/richContent'

// Modal cetak naskah soal dan kunci jawaban dari bank soal yang sudah terkunci.
// Dokumen dirender sebagai HTML terpisah di dalam iframe (pratinjau sekaligus sumber cetak),
// mengikuti pola EventCardPrintModal. Urutan soal mengikuti nomor di bank (tidak diacak);
// soal dikelompokkan per tipe dengan penomoran berurutan yang sama pada naskah dan kunci.
const props = defineProps({
  modelValue: { type: Boolean, default: false },
  bank: { type: Object, default: null },
})
const emit = defineEmits(['update:modelValue'])

const SCHOOL_NAME = 'SMAS ISLAMIC CENTRE DEMAK'
const FOUNDATION_NAME = 'Yayasan Islamic Centre Sultan Fatah Demak'

const PAPERS = {
  a4: { label: 'A4', size: '210mm 297mm', width: 210 },
  f4: { label: 'F4 / Folio', size: '215mm 330mm', width: 215 },
}
const PAGE_MARGIN_MM = 15

const docTypeOptions = [
  { value: 'naskah', label: 'Naskah Soal' },
  { value: 'kunci', label: 'Kunci Jawaban' },
]
const docType = ref('naskah')
const docTypeLabel = computed(() => docTypeOptions.find((o) => o.value === docType.value)?.label || 'Naskah')

const SECTIONS = [
  { type: 'MULTIPLE_CHOICE', title: 'Pilihan Ganda', hint: 'Pilihlah satu jawaban yang paling tepat!' },
  { type: 'SHORT_ANSWER', title: 'Isian Singkat', hint: 'Isilah titik-titik di bawah ini dengan jawaban yang tepat!' },
  { type: 'ESSAY', title: 'Uraian', hint: 'Jawablah pertanyaan di bawah ini dengan jelas dan lengkap!' },
]
const ROMAN = ['A', 'B', 'C']

const rendered = ref(false)
const isLoading = ref(false)
const loadError = ref('')
const bankInfo = ref(null)
const questions = ref([])
const paper = ref('a4')
const columns = ref(1)
const essayLines = ref(true)
const logoDataUrl = ref('')
const previewFrame = ref(null)
const frameReady = ref(false)

let loadToken = 0

const close = () => emit('update:modelValue', false)

// Bila modal dibuka lagi sebelum animasi tutup selesai, after-leave yang tertunda tidak boleh
// menyembunyikan modal yang sudah terbuka kembali.
const onAfterLeave = () => {
  if (!props.modelValue) rendered.value = false
}

const escapeHtml = (value) =>
  String(value ?? '').replace(/[&<>"']/g, (ch) => ({ '&': '&amp;', '<': '&lt;', '>': '&gt;', '"': '&quot;', "'": '&#39;' }[ch]))

// Logo sebagai data URL agar pasti termuat di dokumen iframe sebelum dicetak.
const ensureLogo = async () => {
  if (logoDataUrl.value) return
  try {
    const img = await new Promise((resolve, reject) => {
      const el = new Image()
      el.onload = () => resolve(el)
      el.onerror = reject
      el.src = '/logo-smic.png'
    })
    const canvas = document.createElement('canvas')
    canvas.width = img.naturalWidth
    canvas.height = img.naturalHeight
    canvas.getContext('2d').drawImage(img, 0, 0)
    logoDataUrl.value = canvas.toDataURL('image/png')
  } catch {
    logoDataUrl.value = ''
  }
}

const load = async () => {
  const token = ++loadToken
  isLoading.value = true
  loadError.value = ''
  try {
    const [res] = await Promise.all([api.get(`/admin/question-banks/${props.bank.id}/print`), ensureLogo()])
    if (token !== loadToken) return
    bankInfo.value = res.data.bank || null
    questions.value = res.data.data || []
  } catch (err) {
    if (token !== loadToken) return
    loadError.value = err.response?.data?.message || 'Gagal memuat naskah soal'
  } finally {
    if (token === loadToken) isLoading.value = false
  }
}

const hasEssay = computed(() => questions.value.some((q) => q.type === 'ESSAY'))

// Kelompok soal per tipe dengan nomor cetak berurutan. Urutan di dalam kelompok mengikuti bank.
const groups = computed(() => {
  let no = 0
  const known = new Set(SECTIONS.map((s) => s.type))
  return SECTIONS.map((section) => {
    const items = questions.value
      .filter((q) => (known.has(q.type) ? q.type : 'MULTIPLE_CHOICE') === section.type)
      .map((q) => ({ ...q, printNo: ++no }))
    return { ...section, items }
  }).filter((g) => g.items.length > 0)
})

const classLabel = computed(() => {
  const b = bankInfo.value
  if (!b) return '-'
  if (b.classes?.length) return b.classes.map((c) => c.name).join(', ')
  return b.grade ? `Kelas ${b.grade}` : '-'
})

const formatWeight = (w) => {
  const n = Number(w ?? 0)
  return Number.isInteger(n) ? String(n) : n.toFixed(2).replace(/0+$/, '').replace(/\.$/, '')
}

const letterhead = (title, subtitle = '') => `
  <header class="kop">
    ${logoDataUrl.value ? `<img class="logo" src="${logoDataUrl.value}" alt="">` : ''}
    <div class="kop-text">
      <div class="yayasan">${escapeHtml(FOUNDATION_NAME)}</div>
      <div class="school">${escapeHtml(SCHOOL_NAME)}</div>
      <div class="doc-title">${escapeHtml(title)}</div>
      ${subtitle ? `<div class="doc-sub">${escapeHtml(subtitle)}</div>` : ''}
    </div>
  </header>`

const metaTable = (extraRows = []) => {
  const b = bankInfo.value || {}
  const rows = [
    ['Mata Pelajaran', b.subject?.name || '-'],
    ['Kelas', classLabel.value],
    ['Naskah', b.title || '-'],
    ['Jumlah Soal', `${questions.value.length} butir`],
    ...extraRows,
  ]
  return `<table class="meta">${rows
    .map(([k, v]) => `<tr><td class="k">${escapeHtml(k)}</td><td class="sep">:</td><td>${escapeHtml(v)}</td></tr>`)
    .join('')}</table>`
}

const renderOptions = (q) => {
  const opts = q.options || []
  if (!opts.length) return ''
  // Opsi pendek disusun menyamping agar hemat kertas; opsi panjang atau bergambar tetap satu per baris.
  const plain = (o) => String(o.text || '').replace(/<[^>]+>/g, '')
  const compact = opts.every((o) => !o.image_url && plain(o).length <= 22 && !/\$/.test(o.text || ''))
  return `<ol class="opts${compact ? ' compact' : ''}">${opts
    .map(
      (o) => `<li><span class="opt-key">${escapeHtml(o.key)}.</span><div class="opt-body">${formatRichContent(o.text || '')}${
        o.image_url ? `<img class="opt-img" src="${escapeHtml(o.image_url)}" alt="">` : ''
      }</div></li>`
    )
    .join('')}</ol>`
}

const naskahBody = () => {
  const identity = `
    <table class="identity">
      <tr><td>Nama</td><td>:</td><td class="line"></td><td class="gap"></td><td>No. Ujian</td><td>:</td><td class="line"></td></tr>
      <tr><td>Kelas</td><td>:</td><td class="line"></td><td class="gap"></td><td>Tanggal</td><td>:</td><td class="line"></td></tr>
    </table>`
  const multi = groups.value.length > 1
  const sections = groups.value
    .map((g, gi) => {
      const items = g.items
        .map((q) => {
          let answer = ''
          if (q.type === 'SHORT_ANSWER') answer = '<div class="short-ans">Jawab: <span class="dots"></span></div>'
          if (q.type === 'ESSAY' && essayLines.value) answer = `<div class="essay-lines">${'<div></div>'.repeat(5)}</div>`
          return `<li class="q"><span class="q-no">${q.printNo}.</span><div class="q-body"><div class="stem">${formatRichContent(
            q.content_html || ''
          )}</div>${q.type === 'MULTIPLE_CHOICE' ? renderOptions(q) : ''}${answer}</div></li>`
        })
        .join('')
      const heading = multi ? `<h2 class="section">${ROMAN[gi]}. ${escapeHtml(g.title)}</h2>` : ''
      return `${heading}<p class="hint">${escapeHtml(g.hint)}</p><ol class="questions">${items}</ol>`
    })
    .join('')
  return `${letterhead('NASKAH SOAL')}${metaTable()}${identity}<main class="${columns.value === 2 ? 'cols-2' : ''}">${sections}</main>`
}

const kunciBody = () => {
  const multi = groups.value.length > 1
  const totalWeight = questions.value.reduce((sum, q) => sum + Number(q.score_weight || 0), 0)
  const sections = groups.value
    .map((g, gi) => {
      const heading = multi ? `<h2 class="section">${ROMAN[gi]}. ${escapeHtml(g.title)}</h2>` : ''
      if (g.type === 'MULTIPLE_CHOICE') {
        const cells = g.items
          .map((q) => `<div class="key-cell"><span class="n">${q.printNo}.</span><span class="v">${escapeHtml(q.correct_key || '-')}</span></div>`)
          .join('')
        return `${heading}<div class="key-grid">${cells}</div>`
      }
      const isEssay = g.type === 'ESSAY'
      const rows = g.items
        .map((q) => {
          const answer = isEssay
            ? formatRichContent(q.rubric_guide || '') || '<span class="muted">Belum ada pedoman penskoran</span>'
            : escapeHtml((q.correct_key || '').split('|').map((k) => k.trim()).filter(Boolean).join(' / ') || '-')
          return `<tr><td class="c">${q.printNo}</td><td>${answer}</td><td class="c">${formatWeight(q.score_weight)}</td></tr>`
        })
        .join('')
      return `${heading}<table class="key-table"><thead><tr><th class="c w-no">No</th><th>${
        isEssay ? 'Pedoman Penskoran' : 'Kunci Jawaban'
      }</th><th class="c w-bobot">Bobot</th></tr></thead><tbody>${rows}</tbody></table>`
    })
    .join('')
  return `${letterhead('KUNCI JAWABAN', 'DOKUMEN RAHASIA')}${metaTable([['Total Bobot', formatWeight(totalWeight)]])}<main>${sections}</main>`
}

const documentHtml = computed(() => {
  if (!bankInfo.value) return ''
  const p = PAPERS[paper.value]
  const body = docType.value === 'kunci' ? kunciBody() : naskahBody()
  const title = `${docTypeLabel.value} - ${bankInfo.value.title || ''}`
  return `<!DOCTYPE html>
<html lang="id"><head><meta charset="utf-8"><title>${escapeHtml(title)}</title>
<link rel="stylesheet" href="${katexCssUrl}">
<link rel="stylesheet" href="https://fonts.googleapis.com/css2?family=Amiri:wght@400;700&family=Noto+Sans+KR:wght@400;600&display=swap">
<style>
  @page { size: ${p.size}; margin: ${PAGE_MARGIN_MM}mm; }
  * { box-sizing: border-box; }
  html { background: #e2e8f0; }
  body { margin: 0; padding: 16px 0; font-family: 'Times New Roman', Times, serif; font-size: 11.5pt; color: #111; line-height: 1.45; }
  .sheet { width: ${p.width}mm; min-height: 100mm; margin: 0 auto; padding: ${PAGE_MARGIN_MM}mm; background: #fff; box-shadow: 0 1px 4px rgba(0,0,0,.15); }
  img { max-width: 100%; height: auto; }
  .kop { display: flex; align-items: center; gap: 4mm; border-bottom: 0.8mm double #111; padding-bottom: 2.5mm; margin-bottom: 3mm; }
  .kop .logo { width: 20mm; height: 20mm; object-fit: contain; }
  .kop-text { flex: 1; text-align: center; padding-right: 20mm; }
  .yayasan { font-size: 10pt; }
  .school { font-size: 15pt; font-weight: bold; letter-spacing: .02em; }
  .doc-title { font-size: 12.5pt; font-weight: bold; margin-top: 1mm; letter-spacing: .06em; }
  .doc-sub { display: inline-block; margin-top: 1mm; padding: .3mm 3mm; border: .3mm solid #111; font-size: 9pt; font-weight: bold; letter-spacing: .1em; }
  .meta { border-collapse: collapse; font-size: 10.5pt; margin-bottom: 2mm; }
  .meta td { padding: .3mm 1.5mm .3mm 0; vertical-align: top; }
  .meta .k { width: 32mm; }
  .meta .sep { width: 3mm; }
  .identity { width: 100%; border-collapse: collapse; font-size: 10.5pt; margin: 1mm 0 3mm; border-top: .3mm solid #111; border-bottom: .3mm solid #111; }
  .identity td { padding: 1.6mm 1mm 1mm 0; white-space: nowrap; width: 1%; }
  .identity .line { width: 35%; border-bottom: .2mm dotted #555; }
  .identity .gap { width: 6%; }
  main.cols-2 { column-count: 2; column-gap: 7mm; column-rule: .2mm solid #bbb; }
  .section { font-size: 11.5pt; margin: 4mm 0 1mm; break-after: avoid; }
  .hint { margin: 0 0 2mm; font-style: italic; font-size: 10.5pt; break-after: avoid; }
  ol { list-style: none; margin: 0; padding: 0; }
  .q { display: flex; gap: 2mm; margin-bottom: 3mm; break-inside: avoid; }
  .q-no { min-width: 7mm; font-weight: bold; }
  .q-body { flex: 1; min-width: 0; }
  .stem p { margin: 0 0 1mm; }
  .stem img { display: block; max-height: 70mm; margin: 1mm 0; }
  .opts { margin-top: 1mm; }
  .opts li { display: flex; gap: 1.5mm; margin-bottom: .6mm; }
  .opts.compact { display: flex; flex-wrap: wrap; column-gap: 6mm; }
  .opts.compact li { min-width: 28mm; }
  .opt-key { min-width: 5mm; }
  .opt-body { flex: 1; min-width: 0; }
  .opt-body p { margin: 0; }
  .opt-img { display: block; max-height: 35mm; margin: .5mm 0; }
  .short-ans { margin-top: 1.5mm; display: flex; gap: 2mm; }
  .short-ans .dots { flex: 1; border-bottom: .2mm dotted #555; }
  .essay-lines div { height: 7mm; border-bottom: .2mm solid #999; }
  .font-arabic { font-family: 'Amiri', serif; font-size: 1.25em; line-height: 2; }
  .font-korean { font-family: 'Noto Sans KR', sans-serif; }
  .katex { font-size: 1.05em; }
  .katex-display { margin: 1mm 0; overflow: hidden; }
  .key-grid { display: grid; grid-template-columns: repeat(5, 1fr); border-top: .3mm solid #111; border-left: .3mm solid #111; margin-bottom: 3mm; }
  .key-cell { display: flex; justify-content: space-between; padding: 1mm 2.5mm; border-right: .3mm solid #111; border-bottom: .3mm solid #111; break-inside: avoid; }
  .key-cell .v { font-weight: bold; }
  .key-table { width: 100%; border-collapse: collapse; font-size: 10.5pt; margin-bottom: 3mm; }
  .key-table th, .key-table td { border: .3mm solid #111; padding: 1.2mm 2mm; vertical-align: top; text-align: left; }
  .key-table th { background: #eee; }
  .key-table tr { break-inside: avoid; }
  .key-table .c { text-align: center; }
  .key-table .w-no { width: 12mm; }
  .key-table .w-bobot { width: 16mm; }
  .key-table p { margin: 0 0 1mm; }
  .muted { color: #777; font-style: italic; }
  @media print {
    html, body { background: #fff; padding: 0; }
    .sheet { width: auto; min-height: 0; padding: 0; box-shadow: none; }
    .key-table th { -webkit-print-color-adjust: exact; print-color-adjust: exact; }
  }
</style></head>
<body><div class="sheet">${body}</div></body></html>`
})

// Pratinjau diperkecil agar lembar muat pada lebar modal (termasuk layar ponsel).
const onFrameLoad = () => {
  const frame = previewFrame.value
  const doc = frame?.contentDocument
  if (!doc?.body) return
  const pageWidthPx = PAPERS[paper.value].width * (96 / 25.4) + 32
  const scale = Math.min(1, frame.clientWidth / pageWidthPx)
  doc.body.style.zoom = scale < 1 ? String(scale) : ''
  frameReady.value = true
}

const printDocument = () => {
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
    if (open && props.bank?.id) {
      rendered.value = true
      docType.value = 'naskah'
      bankInfo.value = null
      questions.value = []
      load()
    } else if (!open) {
      loadToken++
    }
  },
  { immediate: true }
)
</script>
