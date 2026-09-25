import katex from 'katex'

// Pemformatan konten soal (rumus KaTeX, teks Arab dan Korea). Dipakai bersama oleh
// RichContentRenderer (tampilan aplikasi) dan dokumen cetak naskah soal.

/**
 * Render KaTeX math expressions safely.
 */
export const renderMath = (text) => {
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
export const formatMultilingualText = (rawText) => {
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

export const formatRichContent = (raw) => {
  if (!raw || !raw.trim()) return ''
  return formatMultilingualText(renderMath(raw))
}
