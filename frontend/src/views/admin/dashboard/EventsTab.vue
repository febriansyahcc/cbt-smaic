<template>
  <!-- TAB: EVENT UJIAN (PERIODE ASESMEN) -->
  <div v-if="activeTab === 'events'" class="space-y-4">
    <!-- Accessibility & Action Toolbar (Search, Status Filters, & Buat Event Baru) -->
    <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-3 bg-white p-3 rounded-2xl border border-slate-200/80 shadow-xs">
      <!-- Search & Filter Controls -->
      <div class="flex flex-col sm:flex-row items-stretch sm:items-center gap-2.5 flex-1">
        <!-- Search Bar -->
        <div class="relative flex-1 sm:max-w-xs">
          <svg class="w-4 h-4 absolute left-3 top-1/2 -translate-y-1/2 text-slate-400 pointer-events-none" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
          </svg>
          <input
            v-model="eventSearchQuery"
            type="text"
            placeholder="Cari nama atau kode event..."
            class="w-full pl-9 pr-8 py-2 bg-slate-50 hover:bg-slate-100/80 focus:bg-white rounded-xl border border-slate-200 text-xs font-medium focus:outline-none focus:ring-2 focus:ring-indigo-500 transition"
          />
          <button
            v-if="eventSearchQuery"
            @click="eventSearchQuery = ''"
            class="absolute right-2.5 top-1/2 -translate-y-1/2 text-slate-400 hover:text-slate-600 text-xs font-bold"
          >
            ✕
          </button>
        </div>

        <!-- Status Filter Pills -->
        <div class="flex items-center gap-1 bg-slate-100 p-1 rounded-xl shrink-0 text-[11px] font-semibold">
          <button
            type="button"
            @click="eventStatusFilter = 'all'"
            :class="['px-3 py-1 rounded-lg transition', eventStatusFilter === 'all' ? 'bg-white text-indigo-700 shadow-2xs font-bold' : 'text-slate-600 hover:text-slate-900']"
          >
            Semua ({{ events.length }})
          </button>
          <button
            type="button"
            @click="eventStatusFilter = 'active'"
            :class="['px-3 py-1 rounded-lg transition', eventStatusFilter === 'active' ? 'bg-white text-emerald-700 shadow-2xs font-bold' : 'text-slate-600 hover:text-slate-900']"
          >
            Aktif ({{ activeEventsCount }})
          </button>
          <button
            type="button"
            @click="eventStatusFilter = 'archived'"
            :class="['px-3 py-1 rounded-lg transition', eventStatusFilter === 'archived' ? 'bg-white text-slate-800 shadow-2xs font-bold' : 'text-slate-600 hover:text-slate-900']"
          >
            Arsip ({{ events.length - activeEventsCount }})
          </button>
        </div>
      </div>

      <!-- Create Event Button (Tunggal: 1 icon plus SVG tanpa karakter plus ganda di teks) -->
      <button
        v-if="authStore.hasPermission('events:manage')"
        @click="openCreateEvent"
        class="px-4 py-2 bg-indigo-600 hover:bg-indigo-700 active:scale-95 text-white font-bold text-xs rounded-xl shadow-xs transition flex items-center justify-center gap-1.5 shrink-0 cursor-pointer self-stretch sm:self-auto"
      >
        <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
        </svg>
        <span>Buat Event Baru</span>
      </button>
    </div>

    <!-- Event Cards Grid -->
    <div v-if="filteredEvents.length > 0" class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
      <div
        v-for="ev in filteredEvents"
        :key="ev.id"
        @click="enterEvent(ev)"
        :class="[
          'bg-white rounded-3xl border p-5 shadow-xs transition-all flex flex-col justify-between space-y-4 relative cursor-pointer group hover:border-indigo-400 hover:shadow-md',
          selectedExamEvent?.id === ev.id ? 'border-indigo-600 ring-2 ring-indigo-500/30' : (ev.is_active ? 'border-slate-200' : 'border-slate-200 bg-slate-50/60 opacity-90')
        ]"
      >
        <div class="space-y-3">
          <!-- Top Row: Badges & 3-Dots Action Menu -->
          <div class="flex items-center justify-between gap-2">
            <div class="flex items-center flex-wrap gap-1.5">
              <span class="px-2 py-0.5 bg-slate-100 text-slate-700 font-mono text-[10px] font-bold rounded-md">
                {{ ev.code }}
              </span>
              <span class="px-2 py-0.5 bg-indigo-50 text-indigo-700 text-[10px] font-bold rounded-md">
                T.A. {{ ev.academic_year }} • {{ ev.semester }}
              </span>
              <span v-if="selectedExamEvent?.id === ev.id" class="px-2 py-0.5 bg-indigo-600 text-white text-[10px] font-black rounded-md flex items-center gap-1 uppercase tracking-wider">
                🎯 Terpilih
              </span>
              <span v-else-if="ev.is_active" class="px-2 py-0.5 bg-emerald-50 text-emerald-700 border border-emerald-200 text-[10px] font-bold rounded-md">
                Berlangsung
              </span>
              <span v-else class="px-2 py-0.5 bg-slate-100 text-slate-500 border border-slate-200 text-[10px] font-bold rounded-md">
                Diarsipkan
              </span>
            </div>

            <!-- 3-Dots Dropdown Trigger -->
            <div v-if="authStore.hasPermission('events:manage')" class="relative shrink-0" @click.stop>
              <button
                type="button"
                @click.stop="toggleEventMenu(ev.id)"
                class="w-8 h-8 flex items-center justify-center rounded-xl text-slate-400 hover:text-slate-700 hover:bg-slate-100 transition cursor-pointer"
                title="Opsi Event"
              >
                <svg class="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 5v.01M12 12v.01M12 19v.01M12 6a1 1 0 110-2 1 1 0 010 2zm0 7a1 1 0 110-2 1 1 0 010 2zm0 7a1 1 0 110-2 1 1 0 010 2z" />
                </svg>
              </button>

              <!-- Dropdown Menu -->
              <div
                v-if="activeEventMenuId === ev.id"
                class="absolute right-0 mt-1 w-44 bg-white rounded-2xl shadow-xl border border-slate-100 py-1.5 z-20"
              >
                <button
                  type="button"
                  @click.stop="openEditEvent(ev); activeEventMenuId = null"
                  class="w-full text-left px-3.5 py-2 text-xs font-semibold text-slate-700 hover:bg-indigo-50 hover:text-indigo-600 flex items-center gap-2 transition cursor-pointer"
                >
                  <svg class="w-4 h-4 text-slate-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z" />
                  </svg>
                  <span>Edit Event</span>
                </button>
                <button
                  type="button"
                  @click.stop="toggleEventActive(ev); activeEventMenuId = null"
                  class="w-full text-left px-3.5 py-2 text-xs font-semibold text-slate-700 hover:bg-slate-50 flex items-center gap-2 transition cursor-pointer"
                >
                  <svg class="w-4 h-4 text-slate-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 8h14M5 8a2 2 0 110-4h14a2 2 0 110 4M5 8v10a2 2 0 002 2h10a2 2 0 002-2V8m-9 4h4" />
                  </svg>
                  <span>{{ ev.is_active ? 'Arsipkan Event' : 'Aktifkan Event' }}</span>
                </button>
                <button
                  type="button"
                  @click.stop="openCardPrint(ev); activeEventMenuId = null"
                  class="w-full text-left px-3.5 py-2 text-xs font-semibold text-slate-700 hover:bg-emerald-50 hover:text-emerald-700 flex items-center gap-2 transition cursor-pointer"
                >
                  <svg class="w-4 h-4 text-slate-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17 17h2a2 2 0 002-2v-4a2 2 0 00-2-2H5a2 2 0 00-2 2v4a2 2 0 002 2h2m2 4h6a2 2 0 002-2v-4a2 2 0 00-2-2H9a2 2 0 00-2 2v4a2 2 0 002 2zm8-12V5a2 2 0 00-2-2H9a2 2 0 00-2 2v4h10z" />
                  </svg>
                  <span>Cetak Kartu Peserta</span>
                </button>
                <div class="my-1 border-t border-slate-100"></div>
                <button
                  type="button"
                  @click.stop="deleteEvent(ev); activeEventMenuId = null"
                  class="w-full text-left px-3.5 py-2 text-xs font-semibold text-rose-600 hover:bg-rose-50 flex items-center gap-2 transition cursor-pointer"
                >
                  <svg class="w-4 h-4 text-rose-500" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
                  </svg>
                  <span>Hapus Event</span>
                </button>
              </div>
            </div>
          </div>

          <!-- Title & Description -->
          <h3 class="text-sm font-bold text-slate-900 leading-snug group-hover:text-indigo-600 transition">
            {{ ev.title }}
          </h3>
          <p v-if="ev.description && ev.description.toLowerCase() !== ev.title.toLowerCase()" class="text-xs text-slate-500 line-clamp-2">
            {{ ev.description }}
          </p>

          <!-- Dates & Schedules -->
          <div class="pt-2 border-t border-slate-100 grid grid-cols-2 gap-2 text-xs">
            <div>
              <span class="text-[10px] text-slate-400 block font-semibold uppercase">Periode Tanggal</span>
              <span class="font-medium text-slate-700 text-[11px]">
                {{ formatDate(ev.start_date) }} - {{ formatDate(ev.end_date) }}
              </span>
            </div>
            <div>
              <span class="text-[10px] text-slate-400 block font-semibold uppercase">Total Jadwal</span>
              <span class="font-bold text-indigo-700 font-mono text-[11px]">
                {{ ev.total_schedules || 0 }} Sesi Ujian
              </span>
            </div>
          </div>
        </div>

        <!-- Footer Action indicator -->
        <div class="pt-3 flex items-center justify-end border-t border-slate-100 text-xs">
          <span class="inline-flex items-center gap-1 text-xs font-bold text-indigo-600 group-hover:text-indigo-700 transition">
            <span>Buka Event</span>
            <svg class="w-4 h-4 transform group-hover:translate-x-1 transition duration-150" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 7l5 5m0 0l-5 5m5-5H6" />
            </svg>
          </span>
        </div>
      </div>
    </div>

    <!-- Empty State jika pencarian tidak menemukan hasil -->
    <div v-else-if="events.length > 0" class="p-10 text-center bg-white rounded-3xl border border-slate-200 shadow-xs space-y-2">
      <div class="text-3xl">🔍</div>
      <h3 class="text-sm font-bold text-slate-800">Tidak Ada Event yang Sesuai</h3>
      <p class="text-xs text-slate-500">
        Tidak ditemukan event dengan kata kunci "{{ eventSearchQuery }}".
      </p>
      <button
        @click="eventSearchQuery = ''; eventStatusFilter = 'all'"
        class="px-3.5 py-1.5 bg-slate-100 hover:bg-slate-200 text-slate-700 font-bold text-xs rounded-xl transition cursor-pointer"
      >
        Reset Filter
      </button>
    </div>

    <!-- Empty State jika belum ada event sama sekali -->
    <div v-else class="p-12 text-center bg-white rounded-3xl border border-slate-200 shadow-xs space-y-3">
      <div class="text-4xl">🎯</div>
      <h3 class="text-sm font-bold text-slate-800">Belum Ada Event Ujian Terdaftar</h3>
      <p class="text-xs text-slate-500 max-w-md mx-auto">
        Buat event ujian baru untuk mulai mengatur sesi jadwal pelaksanaan ujian dan naskah bank soal.
      </p>
      <button
        v-if="authStore.hasPermission('events:manage')"
        @click="openCreateEvent"
        class="px-4 py-2 bg-indigo-600 hover:bg-indigo-700 text-white font-bold text-xs rounded-xl shadow-xs transition"
      >
        Buat Event Baru
      </button>
    </div>
  </div>
</template>

<script setup>
import { useDashboard } from './context'

const {
  activeEventMenuId,
  activeEventsCount,
  activeTab,
  authStore,
  deleteEvent,
  enterEvent,
  eventSearchQuery,
  eventStatusFilter,
  events,
  filteredEvents,
  formatDate,
  openCardPrint,
  openCreateEvent,
  openEditEvent,
  selectedExamEvent,
  toggleEventActive,
  toggleEventMenu,
} = useDashboard()
</script>
