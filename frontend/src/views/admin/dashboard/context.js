import { inject } from 'vue'

// State dan aksi dasbor staf dimiliki AdminDashboardView dan dibagikan ke komponen tab/modal
// lewat provide/inject. Ref yang dibagikan adalah ref yang sama, sehingga perubahan dari mana pun
// langsung terlihat di seluruh dasbor.
export const DASHBOARD_CONTEXT = Symbol('adminDashboard')

export function useDashboard() {
  const ctx = inject(DASHBOARD_CONTEXT, null)
  if (!ctx) throw new Error('useDashboard() hanya bisa dipakai di dalam AdminDashboardView')
  return ctx
}
