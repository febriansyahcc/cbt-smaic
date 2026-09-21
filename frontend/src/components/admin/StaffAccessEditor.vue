<script setup>
// Editor akses akun staf: pilih template (Administrator, Guru, Pengawas) lalu sesuaikan izin.
// v-model: { role: 'ADMIN' | 'GURU', permissions: string[] }. Administrator memakai ['*'] dan terkunci.
// Mode `readonly` menampilkan pilihan saat ini tanpa bisa diubah (untuk pemanggil yang bukan grantor).
import { computed, ref, watch } from 'vue'
import {
  applyImplications,
  canUncheck,
  dependentsOf,
  samePermissionSet,
  templateByKey,
  templateKeyFor,
  TEMPLATE_KEYS,
} from '../../utils/staffAccess'

const props = defineProps({
  modelValue: { type: Object, required: true },
  catalog: { type: Object, required: true },
  readonly: { type: Boolean, default: false },
  // async ({ title, message, type, confirmText, cancelText }) => boolean
  confirmFn: { type: Function, required: true },
})

const emit = defineEmits(['update:modelValue', 'update:valid'])

const groups = computed(() => props.catalog?.groups || [])
const implies = computed(() => props.catalog?.implies || {})
const templates = computed(() => props.catalog?.templates || [])

const staffTemplates = computed(() =>
  [TEMPLATE_KEYS.ADMIN, TEMPLATE_KEYS.GURU, TEMPLATE_KEYS.PENGAWAS]
    .map((key) => templateByKey(templates.value, key))
    .filter(Boolean)
)

const itemLabels = computed(() => {
  const map = {}
  for (const g of groups.value) {
    for (const item of g.items || []) map[item.key] = item.label
  }
  return map
})

// Izin yang benar-benar punya checkbox. Kunci lain (usang atau tak dikenal) tidak dihitung.
const catalogKeys = computed(() => new Set(Object.keys(itemLabels.value)))

// Template dasar: akun kustom dianggap turunan Guru.
const initialKey = templateKeyFor(props.modelValue, templates.value, implies.value)
const baseKey = ref(
  initialKey === TEMPLATE_KEYS.CUSTOM || initialKey === TEMPLATE_KEYS.UNKNOWN ? TEMPLATE_KEYS.GURU : initialKey
)

const baseTemplate = computed(() => templateByKey(templates.value, baseKey.value))
const isAdmin = computed(() => props.modelValue.role === 'ADMIN')
const permissions = computed(() => props.modelValue.permissions || [])

const isCustom = computed(() => {
  if (isAdmin.value) return false
  const basePerms = applyImplications(baseTemplate.value?.permissions, implies.value)
  return !samePermissionSet(applyImplications(permissions.value, implies.value), basePerms)
})

const isValid = computed(
  () => isAdmin.value || permissions.value.some((p) => catalogKeys.value.has(p))
)
watch(isValid, (v) => emit('update:valid', v), { immediate: true })

const setValue = (role, perms) => emit('update:modelValue', { role, permissions: perms })

const applyTemplate = (tpl) => {
  if (tpl.role === 'ADMIN') {
    setValue('ADMIN', ['*'])
  } else {
    setValue(tpl.role, applyImplications(tpl.permissions, implies.value))
  }
}

const selectTemplate = async (tpl) => {
  if (props.readonly) return
  if (tpl.key === baseKey.value) {
    if (isCustom.value) await resetToBase()
    return
  }
  if (tpl.role === 'ADMIN' && !isAdmin.value) {
    const ok = await props.confirmFn({
      title: 'Berikan Akses Administrator',
      message:
        'Template Administrator memberi akses penuh ke seluruh fitur, termasuk mengelola akun dan izin akun lain.' +
        (isCustom.value ? ' Izin kustom yang sudah diatur akan dibuang.' : '') +
        ' Lanjutkan?',
      type: 'danger',
      confirmText: 'Berikan Akses',
      cancelText: 'Batal',
    })
    if (!ok) return
  } else if (isCustom.value) {
    const ok = await props.confirmFn({
      title: 'Ganti Template Akses',
      message: `Izin kustom yang sudah diatur akan dibuang dan diganti izin bawaan ${tpl.label}. Lanjutkan?`,
      type: 'warning',
      confirmText: 'Ganti Template',
      cancelText: 'Batal',
    })
    if (!ok) return
  }
  baseKey.value = tpl.key
  applyTemplate(tpl)
}

const resetToBase = async () => {
  if (props.readonly || !baseTemplate.value) return
  if (isCustom.value) {
    const ok = await props.confirmFn({
      title: 'Kembalikan ke Bawaan',
      message: `Izin kustom akan dibuang dan diganti izin bawaan ${baseTemplate.value.label}. Lanjutkan?`,
      type: 'warning',
      confirmText: 'Kembalikan',
      cancelText: 'Batal',
    })
    if (!ok) return
  }
  applyTemplate(baseTemplate.value)
}

const isChecked = (key) => isAdmin.value || permissions.value.includes(key)

const owners = (key) => dependentsOf(key, permissions.value, implies.value)

const isLocked = (key) =>
  props.readonly || isAdmin.value || (isChecked(key) && !canUncheck(key, permissions.value, implies.value))

const lockedNote = (key) => {
  if (props.readonly || isAdmin.value) return ''
  const first = owners(key)[0]
  return first ? `Diperlukan oleh ${itemLabels.value[first] || first}` : ''
}

const onToggle = async (item, event) => {
  const el = event.target
  if (props.readonly || isAdmin.value) {
    el.checked = isChecked(item.key)
    return
  }
  if (el.checked) {
    if (item.sensitive) {
      // Kembalikan tampilan dulu; centang baru muncul setelah pengguna menyetujui.
      el.checked = false
      const ok = await props.confirmFn({
        title: 'Izin Sensitif',
        message: `Izin "${item.label}" tergolong sensitif. Berikan izin ini kepada akun tersebut?`,
        type: 'warning',
        confirmText: 'Berikan Izin',
        cancelText: 'Batal',
      })
      if (!ok) return
    }
    setValue(props.modelValue.role, applyImplications([...permissions.value, item.key], implies.value))
  } else {
    if (!canUncheck(item.key, permissions.value, implies.value)) {
      el.checked = true
      return
    }
    setValue(props.modelValue.role, permissions.value.filter((p) => p !== item.key))
  }
}
</script>

<template>
  <div class="space-y-3 text-xs">
    <!-- Pemilih template -->
    <div>
      <div class="block font-bold text-slate-700 mb-1.5">Template Akses:</div>
      <div class="grid grid-cols-3 gap-2" role="radiogroup" aria-label="Template akses">
        <button
          v-for="tpl in staffTemplates"
          :key="tpl.key"
          type="button"
          role="radio"
          :aria-checked="baseKey === tpl.key"
          :disabled="readonly"
          @click="selectTemplate(tpl)"
          :class="[
            'min-h-[48px] px-2 py-2.5 rounded-xl border text-center font-bold transition',
            readonly ? 'cursor-not-allowed' : 'active:scale-95 cursor-pointer',
            baseKey === tpl.key
              ? 'border-indigo-500 bg-indigo-50 text-indigo-700 ring-1 ring-indigo-500'
              : readonly
                ? 'border-slate-200 bg-slate-50 text-slate-400 opacity-60'
                : 'border-slate-200 bg-white text-slate-600 hover:bg-slate-50'
          ]"
        >
          {{ tpl.label }}
        </button>
      </div>
      <p v-if="baseTemplate?.description" class="mt-1.5 text-[11px] text-slate-500">{{ baseTemplate.description }}</p>
    </div>

    <!-- Status izin -->
    <div class="flex items-center justify-between gap-2">
      <div class="flex items-center gap-2 min-w-0">
        <span class="font-bold text-slate-700">Izin Akun:</span>
        <span
          v-if="isCustom"
          class="px-2 py-0.5 rounded-full text-[10px] font-bold bg-amber-100 text-amber-800"
        >
          Kustom
        </span>
      </div>
      <button
        v-if="isCustom && !readonly"
        type="button"
        @click="resetToBase"
        class="font-bold text-indigo-600 hover:underline active:scale-95 transition cursor-pointer shrink-0"
      >
        Kembalikan ke bawaan
      </button>
    </div>

    <p v-if="isAdmin" class="px-3 py-2 rounded-xl bg-slate-50 border border-slate-200 font-semibold text-slate-700">
      Akses penuh ke seluruh fitur
    </p>
    <p v-else-if="!isValid && !readonly" class="font-semibold text-rose-600">Pilih minimal satu izin</p>

    <!-- Daftar izin per kelompok -->
    <div class="space-y-2.5" :class="isAdmin || readonly ? 'opacity-70' : ''">
      <div
        v-for="group in groups"
        :key="group.key"
        class="rounded-xl border border-slate-200 overflow-hidden"
      >
        <div class="px-3 py-1.5 bg-slate-50 text-[11px] font-bold text-slate-600 uppercase tracking-wide">
          {{ group.label }}
        </div>
        <label
          v-for="item in group.items"
          :key="item.key"
          :class="[
            'flex items-start gap-2.5 px-3 py-2.5 min-h-[44px] border-t border-slate-100 transition',
            isLocked(item.key) ? 'cursor-not-allowed' : 'cursor-pointer hover:bg-slate-50/70'
          ]"
        >
          <input
            type="checkbox"
            class="mt-0.5 w-4 h-4 shrink-0 accent-indigo-600"
            :class="isLocked(item.key) ? 'cursor-not-allowed' : 'cursor-pointer'"
            :checked="isChecked(item.key)"
            :disabled="isLocked(item.key)"
            @change="onToggle(item, $event)"
          />
          <span class="min-w-0 flex-1">
            <span class="flex items-center gap-1.5 flex-wrap">
              <span class="font-semibold text-slate-800">{{ item.label }}</span>
              <span
                v-if="item.sensitive"
                class="px-1.5 py-px rounded-md text-[10px] font-bold bg-amber-50 text-amber-700 border border-amber-200"
              >
                Sensitif
              </span>
            </span>
            <span v-if="item.description" class="block text-[11px] leading-snug text-slate-500">{{ item.description }}</span>
            <span v-if="lockedNote(item.key)" class="block text-[10px] font-medium text-indigo-600">{{ lockedNote(item.key) }}</span>
          </span>
        </label>
      </div>
    </div>
  </div>
</template>
