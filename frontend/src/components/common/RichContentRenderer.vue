<template>
  <div class="rich-content-wrapper inline-block w-full">
    <!-- Rendered Rich Content HTML -->
    <div
      ref="contentContainer"
      :class="['rich-content-rendered', customClass]"
      @click="handleContainerClick"
      v-html="formattedContent"
    ></div>

    <!-- Zoomable Image Lightbox Modal -->
    <teleport to="body">
      <transition
        enter-active-class="transition-opacity duration-200 ease-out"
        enter-from-class="opacity-0"
        enter-to-class="opacity-100"
        leave-active-class="transition-opacity duration-150 ease-in"
        leave-from-class="opacity-100"
        leave-to-class="opacity-0"
      >
        <div
          v-if="lightboxImage"
          class="fixed inset-0 z-50 flex items-center justify-center p-3 sm:p-6 bg-slate-950/80 backdrop-blur-sm select-none"
          @click="closeLightbox"
          @keydown.esc="closeLightbox"
          tabindex="0"
        >
          <!-- Lightbox Content Card -->
          <transition
            appear
            enter-active-class="transition duration-200 ease-out transform"
            enter-from-class="opacity-0 scale-95"
            enter-to-class="opacity-100 scale-100"
            leave-active-class="transition duration-150 ease-in transform"
            leave-from-class="opacity-100 scale-100"
            leave-to-class="opacity-0 scale-95"
          >
            <div
              class="relative max-w-5xl max-h-[90vh] bg-slate-900 rounded-3xl p-2 sm:p-4 shadow-2xl border border-slate-700 flex flex-col items-center overflow-hidden"
              @click.stop
            >
              <!-- Top Close Button -->
              <div class="w-full flex items-center justify-between pb-2 px-2 text-slate-300">
                <span class="text-xs font-semibold truncate max-w-md text-slate-400">
                  {{ lightboxAlt || 'Pratinjau Gambar' }}
                </span>
                <button
                  type="button"
                  @click="closeLightbox"
                  class="p-2 rounded-xl text-slate-400 hover:text-white hover:bg-slate-800 transition cursor-pointer"
                  title="Tutup (Esc)"
                >
                  <svg class="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
                  </svg>
                </button>
              </div>

              <!-- High Resolution Zoomed Image -->
              <div class="overflow-auto max-h-[80vh] flex items-center justify-center rounded-2xl bg-black/40 p-2">
                <img
                  :src="lightboxImage"
                  :alt="lightboxAlt"
                  class="max-w-full max-h-[75vh] object-contain rounded-xl shadow-lg transition-transform duration-200"
                />
              </div>

              <!-- Bottom Hint -->
              <div class="pt-2 text-[11px] text-slate-400 text-center">
                Klik di luar gambar atau tombol tutup untuk kembali
              </div>
            </div>
          </transition>
        </div>
      </transition>
    </teleport>
  </div>
</template>

<script setup>
import { ref, computed } from 'vue'
import katex from 'katex'

const props = defineProps({
  content: {
    type: String,
    default: ''
  },
  customClass: {
    type: String,
    default: ''
  }
})

const lightboxImage = ref(null)
const lightboxAlt = ref('')
const contentContainer = ref(null)

const closeLightbox = () => {
  lightboxImage.value = null
  lightboxAlt.value = ''
}

const handleContainerClick = (e) => {
  if (e.target && e.target.tagName === 'IMG') {
    e.preventDefault()
    e.stopPropagation()
    lightboxImage.value = e.target.src
    lightboxAlt.value = e.target.alt || 'Gambar Soal'
  }
}

/**
 * Render KaTeX math expressions safely.
 */
const renderMath = (text) => {
  if (!text) return ''

  // 1. Process Display Math: $$...$$
  let processed = text.replace(/\$\$([\s\S]*?)\$\$/g, (match, math) => {
    try {
      return katex.renderToString(math.trim(), {
        displayMode: true,
        throwOnError: false
      })
    } catch (err) {
      console.warn('KaTeX display render error:', err)
      return match
    }
  })

  // 2. Process Inline Math: $...$
  // Match single dollar signs not preceded by another dollar sign and not followed by a digit (to avoid price confusion, though \$ works too)
  processed = processed.replace(/(?<!\$)\$([^\$\n\r]+?)\$(?!\$)/g, (match, math) => {
    try {
      return katex.renderToString(math.trim(), {
        displayMode: false,
        throwOnError: false
      })
    } catch (err) {
      console.warn('KaTeX inline render error:', err)
      return match
    }
  })

  return processed
}

/**
 * Highlight and format Arabic and Korean text.
 */
const formatMultilingualText = (rawText) => {
  if (!rawText) return ''

  // Temporary tokens to protect existing HTML tags & KaTeX output
  const tokens = []
  let textWithPlaceholders = rawText.replace(/<[^>]+>/g, (match) => {
    const placeholder = `___HTML_TOKEN_${tokens.length}___`
    tokens.push(match)
    return placeholder
  })

  // Format Korean Hangul text blocks
  // Regex: matches Korean Hangul syllable blocks [\uAC00-\uD7AF\u1100-\u11FF\u3130-\u318F]
  textWithPlaceholders = textWithPlaceholders.replace(
    /([\uAC00-\uD7AF\u1100-\u11FF\u3130-\u318F]+(?:[\s.,!?;:'"()-]+[\uAC00-\uD7AF\u1100-\u11FF\u3130-\u318F]+)*)/gu,
    (match) => {
      return `<span class="font-korean font-medium text-slate-900">${match}</span>`
    }
  )

  // Format Arabic script blocks
  // Regex: matches Arabic character sequences [\u0600-\u06FF\u0750-\u077F\u08A0-\u08FF\uFB50-\uFDFF\uFE70-\uFEFF]
  textWithPlaceholders = textWithPlaceholders.replace(
    /([\u0600-\u06FF\u0750-\u077F\u08A0-\u08FF\uFB50-\uFDFF\uFE70-\uFEFF]+(?:[\s.,!?;:()«»'’"–-]+[\u0600-\u06FF\u0750-\u077F\u08A0-\u08FF\uFB50-\uFDFF\uFE70-\uFEFF]+)*)/gu,
    (match) => {
      return `<span class="font-arabic font-normal inline-block my-0.5 text-slate-900 text-lg leading-loose" dir="rtl">${match}</span>`
    }
  )

  // Restore protected HTML tags
  tokens.forEach((tag, idx) => {
    textWithPlaceholders = textWithPlaceholders.replace(`___HTML_TOKEN_${idx}___`, tag)
  })

  return textWithPlaceholders
}

const formattedContent = computed(() => {
  const raw = props.content || ''
  if (!raw.trim()) return ''

  // Step 1: Render KaTeX math formulas
  const mathRendered = renderMath(raw)

  // Step 2: Format Arabic & Korean text runs
  const multilingualRendered = formatMultilingualText(mathRendered)

  return multilingualRendered
})
</script>
