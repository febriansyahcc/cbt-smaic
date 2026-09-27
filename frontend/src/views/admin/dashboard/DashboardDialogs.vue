<template>
  <!-- GLOBAL REACTIVE CONFIRM / ALERT MODAL -->
  <div v-if="dialogState.isOpen" class="fixed inset-0 z-50 flex items-center justify-center p-4 select-none">
    <!-- Backdrop with pure opacity fade (NO scale) -->
    <transition
      appear
      enter-active-class="transition-opacity duration-200 ease-out"
      enter-from-class="opacity-0"
      enter-to-class="opacity-100"
      leave-active-class="transition-opacity duration-150 ease-in"
      leave-from-class="opacity-100"
      leave-to-class="opacity-0"
    >
      <div
        class="fixed inset-0 bg-slate-900/60 backdrop-blur-xs transition-opacity"
        @click="handleDialogCancel"
      ></div>
    </transition>

    <!-- Modal Card with smooth scale & fade -->
    <transition
      appear
      enter-active-class="transition duration-200 ease-out transform"
      enter-from-class="opacity-0 scale-95 translate-y-2 sm:translate-y-0"
      enter-to-class="opacity-100 scale-100 translate-y-0"
      leave-active-class="transition duration-150 ease-in transform"
      leave-from-class="opacity-100 scale-100 translate-y-0"
      leave-to-class="opacity-0 scale-95 translate-y-2 sm:translate-y-0"
    >
      <div
        class="relative bg-white rounded-3xl max-w-md w-full p-6 shadow-2xl space-y-4 z-10"
        @click.stop
      >
        <div class="flex items-center justify-between border-b border-slate-100 pb-3">
          <h3 class="text-base font-bold text-slate-900">{{ dialogState.title }}</h3>
          <button
            @click="handleDialogCancel"
            class="text-slate-400 hover:text-slate-600 p-1.5 rounded-xl hover:bg-slate-100 transition cursor-pointer"
          >
            ✕
          </button>
        </div>

        <div class="text-xs text-slate-600 leading-relaxed whitespace-pre-line py-1">
          {{ dialogState.message }}
        </div>

        <div class="flex justify-end gap-2 pt-2 border-t border-slate-100">
          <button
            v-if="dialogState.cancelText"
            type="button"
            @click="handleDialogCancel"
            class="px-4 py-2 bg-slate-100 hover:bg-slate-200 active:scale-95 text-slate-700 font-bold text-xs rounded-xl transition cursor-pointer"
          >
            {{ dialogState.cancelText }}
          </button>
          <button
            type="button"
            @click="handleDialogConfirm"
            :class="[
              'px-5 py-2.5 text-white font-bold text-xs rounded-xl shadow-xs transition active:scale-95 cursor-pointer',
              dialogState.type === 'danger' ? 'bg-rose-600 hover:bg-rose-700' :
              dialogState.type === 'warning' ? 'bg-amber-600 hover:bg-amber-700' :
              'bg-indigo-600 hover:bg-indigo-700'
            ]"
          >
            {{ dialogState.confirmText }}
          </button>
        </div>
      </div>
    </transition>
  </div>

  <!-- Global Toast Notification -->
  <transition enter-active-class="transform ease-out duration-300 transition" enter-from-class="translate-y-2 opacity-0 sm:translate-y-0 sm:translate-x-2" enter-to-class="translate-y-0 opacity-100 sm:translate-x-0" leave-active-class="transition ease-in duration-100" leave-from-class="opacity-100" leave-to-class="opacity-0">
    <div v-if="toastMessage" class="fixed bottom-5 right-5 z-50 flex items-center gap-2 px-4 py-3 bg-slate-900 text-white rounded-2xl shadow-xl text-xs font-semibold no-print">
      <span v-if="toastType === 'success'" class="text-emerald-400">✓</span>
      <span v-else class="text-indigo-400">ℹ</span>
      <span>{{ toastMessage }}</span>
    </div>
  </transition>
</template>

<script setup>
import { useDashboard } from './context'

const {
  dialogState,
  handleDialogCancel,
  handleDialogConfirm,
  toastMessage,
  toastType,
} = useDashboard()
</script>
