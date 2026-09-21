<template>
  <div>
    <!-- REACTIVE GLOBAL MODAL (ALERT / CONFIRM) -->
    <Teleport to="body">
      <div
        v-if="dialogState.isOpen"
        class="fixed inset-0 z-[9999] flex items-center justify-center p-4 select-none"
        @keydown.esc="handleCancel"
        tabindex="-1"
      >
        <!-- Backdrop click with pure opacity fade (NO scale) -->
        <Transition
          appear
          enter-active-class="transition-opacity duration-200 ease-out"
          enter-from-class="opacity-0"
          enter-to-class="opacity-100"
          leave-active-class="transition-opacity duration-150 ease-in"
          leave-from-class="opacity-100"
          leave-to-class="opacity-0"
        >
          <div class="fixed inset-0 bg-slate-900/60 backdrop-blur-xs transition-opacity" @click="handleCancel"></div>
        </Transition>

        <!-- Dialog Card with smooth scale & fade -->
        <Transition
          appear
          enter-active-class="transition duration-200 ease-out transform"
          enter-from-class="opacity-0 scale-95 translate-y-2 sm:translate-y-0"
          enter-to-class="opacity-100 scale-100 translate-y-0"
          leave-active-class="transition duration-150 ease-in transform"
          leave-from-class="opacity-100 scale-100 translate-y-0"
          leave-to-class="opacity-0 scale-95 translate-y-2 sm:translate-y-0"
        >
          <div
            class="relative w-full max-w-md bg-white rounded-3xl shadow-2xl border border-slate-100 p-6 z-10 space-y-4"
            @click.stop
          >
            <div class="flex items-start justify-between border-b border-slate-100 pb-3">
              <div class="flex items-center gap-3">
                <!-- Icon Container -->
                <div
                  class="w-10 h-10 rounded-xl flex items-center justify-center shrink-0"
                  :class="{
                    'bg-rose-50 text-rose-600': dialogState.type === 'danger',
                    'bg-amber-50 text-amber-600': dialogState.type === 'warning',
                    'bg-emerald-50 text-emerald-600': dialogState.type === 'success',
                    'bg-indigo-50 text-indigo-600': dialogState.type === 'confirm' || dialogState.type === 'alert' || dialogState.type === 'info'
                  }"
                >
                  <!-- Danger Icon -->
                  <svg v-if="dialogState.type === 'danger'" class="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
                  </svg>

                  <!-- Warning Icon -->
                  <svg v-else-if="dialogState.type === 'warning'" class="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
                  </svg>

                  <!-- Success Icon -->
                  <svg v-else-if="dialogState.type === 'success'" class="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" />
                  </svg>

                  <!-- Info Icon -->
                  <svg v-else class="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
                  </svg>
                </div>

                <h3 class="text-base font-bold text-slate-900 leading-tight">
                  {{ dialogState.title }}
                </h3>
              </div>

              <!-- Close button -->
              <button
                type="button"
                @click="handleCancel"
                class="text-slate-400 hover:text-slate-600 p-1.5 rounded-xl hover:bg-slate-100 transition cursor-pointer"
              >
                ✕
              </button>
            </div>

            <!-- Content Area -->
            <div class="text-xs text-slate-600 leading-relaxed whitespace-pre-line py-1">
              {{ dialogState.message }}
            </div>

            <!-- Modal Action Buttons -->
            <div class="flex items-center justify-end gap-2 pt-3 border-t border-slate-100">
              <button
                v-if="dialogState.cancelText"
                type="button"
                @click="handleCancel"
                class="px-4 py-2 rounded-xl border border-slate-200 text-slate-700 font-bold text-xs hover:bg-slate-50 transition active:scale-95 cursor-pointer"
              >
                {{ dialogState.cancelText }}
              </button>
              
              <button
                type="button"
                @click="handleConfirm"
                class="px-5 py-2.5 rounded-xl text-white font-bold text-xs shadow-xs transition active:scale-95 cursor-pointer flex items-center gap-2"
                :class="{
                  'bg-rose-600 hover:bg-rose-700 shadow-rose-200': dialogState.type === 'danger',
                  'bg-amber-600 hover:bg-amber-700 shadow-amber-200': dialogState.type === 'warning',
                  'bg-emerald-600 hover:bg-emerald-700 shadow-emerald-200': dialogState.type === 'success',
                  'bg-indigo-600 hover:bg-indigo-700 shadow-indigo-200': dialogState.type === 'confirm' || dialogState.type === 'alert' || dialogState.type === 'info'
                }"
              >
                {{ dialogState.confirmText }}
              </button>
            </div>
          </div>
        </Transition>
      </div>
    </Teleport>

    <!-- REACTIVE TOASTS -->
    <Teleport to="body">
      <div class="fixed bottom-5 right-5 z-[9999] flex flex-col gap-2.5 pointer-events-none max-w-md w-full px-4">
        <TransitionGroup
          enter-active-class="transition duration-200 ease-out"
          enter-from-class="opacity-0 translate-y-3 scale-95"
          enter-to-class="opacity-100 translate-y-0 scale-100"
          leave-active-class="transition duration-150 ease-in"
          leave-from-class="opacity-100 translate-y-0 scale-100"
          leave-to-class="opacity-0 translate-y-2 scale-95"
        >
          <div
            v-for="t in toasts"
            :key="t.id"
            class="pointer-events-auto p-4 rounded-xl shadow-lg border backdrop-blur-sm flex items-center justify-between gap-3 text-sm font-medium"
            :class="{
              'bg-emerald-900/90 text-white border-emerald-700': t.type === 'success',
              'bg-rose-900/90 text-white border-rose-700': t.type === 'danger',
              'bg-amber-900/90 text-white border-amber-700': t.type === 'warning',
              'bg-slate-900/90 text-white border-slate-700': t.type === 'info'
            }"
          >
            <div class="flex items-center gap-2.5">
              <span v-if="t.type === 'success'" class="text-emerald-400 font-bold">✓</span>
              <span v-else-if="t.type === 'danger'" class="text-rose-400 font-bold">✕</span>
              <span v-else-if="t.type === 'warning'" class="text-amber-400 font-bold">⚠</span>
              <span v-else class="text-blue-400 font-bold">ℹ</span>
              <span>{{ t.message }}</span>
            </div>
            <button @click="removeToast(t.id)" class="text-white/70 hover:text-white text-xs px-1">✕</button>
          </div>
        </TransitionGroup>
      </div>
    </Teleport>
  </div>
</template>

<script setup>
import { useDialog } from '@/composables/useDialog'

const { dialogState, toasts, handleConfirm, handleCancel, removeToast } = useDialog()
</script>
