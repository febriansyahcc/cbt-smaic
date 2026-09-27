<template>
  <!-- MODAL: CETAK DAFTAR HADIR & BERITA ACARA DARI DATA PENGAWASAN LIVE -->
  <ProctorPrintModal v-model="showProctorPrint" :proctor-data="proctorData" />
  <EventCardPrintModal v-model="showCardPrint" :event="cardPrintEvent" />
  <QuestionPrintModal v-model="showQuestionPrint" :bank="questionPrintBank" />

  <!-- MODAL: CETAK DOKUMEN RESMI UJIAN (DAFTAR HADIR & BERITA ACARA) -->
  <div v-if="showPrintModal && printSchedule" class="fixed inset-0 z-50 flex items-center justify-center p-2 sm:p-4 bg-slate-900/70 backdrop-blur-xs overflow-y-auto">
    <div class="bg-white rounded-3xl max-w-4xl w-full shadow-2xl overflow-hidden flex flex-col max-h-[95vh] my-auto">
      <!-- Modal Toolbar Header (no-print) -->
      <div class="flex items-center justify-between px-6 py-4 border-b border-slate-200 shrink-0 bg-slate-50 no-print">
        <div class="flex items-center space-x-3">
          <div class="w-9 h-9 rounded-2xl bg-emerald-100 text-emerald-700 flex items-center justify-center font-black text-base">
            📄
          </div>
          <div>
            <h3 class="text-sm font-bold text-slate-900">Format Cetak Dokumen Resmi Ujian</h3>
            <p class="text-[11px] text-slate-500">SMAS Islamic Centre Demak • Standar A4 Cetak Siap Pakai</p>
          </div>
        </div>

        <!-- Document Type Selector Pills -->
        <div class="flex items-center gap-1.5 bg-slate-200/70 p-1 rounded-2xl">
          <button
            type="button"
            @click="printDocType = 'attendance'"
            :class="[
              'px-3 py-1.5 rounded-xl text-xs font-bold transition cursor-pointer',
              printDocType === 'attendance' ? 'bg-white text-emerald-700 shadow-xs' : 'text-slate-600 hover:text-slate-900'
            ]"
          >
            📄 Daftar Hadir
          </button>
          <button
            type="button"
            @click="printDocType = 'report'"
            :class="[
              'px-3 py-1.5 rounded-xl text-xs font-bold transition cursor-pointer',
              printDocType === 'report' ? 'bg-white text-indigo-700 shadow-xs' : 'text-slate-600 hover:text-slate-900'
            ]"
          >
            📝 Berita Acara
          </button>
        </div>

        <button @click="showPrintModal = false" class="text-slate-400 hover:text-slate-600 p-1.5 rounded-xl hover:bg-slate-200 transition cursor-pointer">
          ✕
        </button>
      </div>

      <!-- Printable Document Area -->
      <div class="p-6 sm:p-8 overflow-y-auto flex-1 bg-slate-100/50">
        <div id="printable-document" class="bg-white p-8 max-w-[210mm] mx-auto shadow-sm border border-slate-200 text-black font-serif text-[12px] leading-relaxed">

          <!-- Kop Surat Resmi SMAS Islamic Centre Demak -->
          <div class="flex items-center gap-4 pb-3 border-b-4 border-double border-slate-900 mb-5">
            <img src="/logo-smic.png" alt="Logo SMAS Islamic Centre Demak" class="w-20 h-20 object-contain shrink-0" />
            <div class="text-center flex-1 space-y-0.5">
              <div class="font-bold text-sm tracking-wide uppercase text-slate-800">Yayasan Islamic Centre Demak</div>
              <div class="font-black text-lg tracking-wider uppercase text-slate-950">SMAS ISLAMIC CENTRE DEMAK</div>
              <div class="text-[11px] font-semibold tracking-wide text-slate-700">STATUS : TERAKREDITASI "A"</div>
              <div class="text-[10px] text-slate-600">Alamat: Jl. Diponegoro No. 12 Demak, Jawa Tengah 59515 | Telp: (0291) 685261 | Website: smicdemak.sch.id</div>
            </div>
          </div>

          <!-- DOKUMEN 1: DAFTAR HADIR PESERTA -->
          <div v-if="printDocType === 'attendance'" class="space-y-4">
            <div class="text-center space-y-1">
              <h2 class="text-base font-black uppercase tracking-wider underline">DAFTAR HADIR PESERTA UJIAN</h2>
              <p class="text-xs font-bold uppercase tracking-wide text-slate-800">{{ selectedExamEvent?.name || 'PENILAIAN AKHIR TAHUN / SUMATIF' }}</p>
              <p class="text-[11px] text-slate-600">TAHUN PELAJARAN 2025/2026</p>
            </div>

            <!-- Metadata Grid -->
            <table class="w-full text-xs border-collapse my-3">
              <tbody>
                <tr>
                  <td class="py-1 w-32 font-bold">Mata Pelajaran</td>
                  <td class="py-1 w-4">:</td>
                  <td class="py-1 font-semibold">{{ printSchedule.subject?.name || printSchedule.bank?.subject?.name || '-' }}</td>
                  <td class="py-1 w-28 font-bold">Hari / Tanggal</td>
                  <td class="py-1 w-4">:</td>
                  <td class="py-1">{{ formatScheduleDateFull(printSchedule.start_time) }}</td>
                </tr>
                <tr>
                  <td class="py-1 font-bold">Kelas / Rombel</td>
                  <td class="py-1">:</td>
                  <td class="py-1 font-semibold">{{ printSchedule.class_room?.name || '-' }}</td>
                  <td class="py-1 font-bold">Waktu / Sesi</td>
                  <td class="py-1">:</td>
                  <td class="py-1">{{ formatScheduleTimeOnly(printSchedule.start_time) }} - {{ formatScheduleTimeOnly(printSchedule.end_time) }} WIB</td>
                </tr>
                <tr>
                  <td class="py-1 font-bold">Naskah Bank Soal</td>
                  <td class="py-1">:</td>
                  <td class="py-1">{{ printSchedule.bank?.title || 'Bank Soal Terintegrasi CBT' }}</td>
                  <td class="py-1 font-bold">Token Sesi</td>
                  <td class="py-1">:</td>
                  <td class="py-1 font-mono font-bold">{{ printSchedule.exam_token }}</td>
                </tr>
              </tbody>
            </table>

            <!-- Tabel Daftar Hadir Siswa -->
            <table class="w-full border-collapse border border-slate-900 text-xs mt-4">
              <thead>
                <tr class="bg-slate-100 border-b border-slate-900 text-center font-bold">
                  <th class="border border-slate-900 py-2 w-10">No.</th>
                  <th class="border border-slate-900 py-2 w-32">Nomor Peserta / NISN</th>
                  <th class="border border-slate-900 py-2 px-3 text-left">Nama Lengkap Siswa</th>
                  <th class="border border-slate-900 py-2 w-12">L/P</th>
                  <th class="border border-slate-900 py-2 w-44" colspan="2">Tanda Tangan</th>
                  <th class="border border-slate-900 py-2 w-24">Keterangan</th>
                </tr>
              </thead>
              <tbody>
                <tr v-if="printStudentsList.length === 0">
                  <td colspan="7" class="border border-slate-900 py-6 text-center text-slate-500 italic">
                    Belum ada data siswa teralokasi di kelas ini.
                  </td>
                </tr>
                <tr v-for="(stu, idx) in printStudentsList" :key="stu.id" class="border-b border-slate-900/60">
                  <td class="border border-slate-900 py-2.5 text-center font-semibold">{{ idx + 1 }}.</td>
                  <td class="border border-slate-900 py-2.5 px-2 text-center font-mono">{{ stu.nisn || stu.student_id_number || stu.user?.username || '-' }}</td>
                  <td class="border border-slate-900 py-2.5 px-3 font-semibold uppercase">{{ stu.user?.full_name || stu.name || '-' }}</td>
                  <td class="border border-slate-900 py-2.5 text-center">{{ stu.gender || stu.user?.gender || 'L' }}</td>
                  <!-- Tanda Tangan Selang-Seling Kolom Ganjil / Genap -->
                  <td class="border border-slate-900 py-2.5 px-2 w-24 text-left font-mono text-[10px] text-slate-500 align-top">
                    <span v-if="idx % 2 === 0">{{ idx + 1 }}. ......................</span>
                  </td>
                  <td class="border border-slate-900 py-2.5 px-2 w-24 text-left font-mono text-[10px] text-slate-500 align-top">
                    <span v-if="idx % 2 === 1">{{ idx + 1 }}. ......................</span>
                  </td>
                  <td class="border border-slate-900 py-2.5 text-center text-[10px] text-slate-400 font-sans">
                    [ &nbsp; ] Hadir
                  </td>
                </tr>
              </tbody>
            </table>

            <!-- Rekap Kehadiran Singkat -->
            <div class="flex justify-between items-start text-xs pt-2">
              <div class="space-y-1">
                <div>Jumlah Peserta Seharusnya : <strong>{{ printStudentsList.length }}</strong> Siswa</div>
                <div>Jumlah Peserta Hadir &nbsp; &nbsp; &nbsp; &nbsp; : .......... Siswa</div>
                <div>Jumlah Peserta Tidak Hadir &nbsp;: .......... Siswa</div>
              </div>
            </div>

            <!-- Tanda Tangan Footer Pengawas -->
            <div class="grid grid-cols-2 gap-8 pt-8 text-center text-xs">
              <div>
                <p class="font-bold">Mengetahui,</p>
                <p class="font-semibold">Kepala Sekolah / Ketua Panitia</p>
                <div class="h-20"></div>
                <p class="font-black underline uppercase">Drs. H. M. Sodiq, M.Pd.</p>
                <p class="text-[11px] text-slate-600">NIP. 19680512 199403 1 005</p>
              </div>
              <div>
                <p class="font-bold">Demak, {{ formatScheduleDateFull(printSchedule.start_time) }}</p>
                <p class="font-semibold">Pengawas Ruang Ujian / Proktor</p>
                <div class="h-20"></div>
                <p class="font-black underline uppercase">....................................................</p>
                <p class="text-[11px] text-slate-600">NIP. ............................................</p>
              </div>
            </div>
          </div>

          <!-- DOKUMEN 2: BERITA ACARA UJIAN -->
          <div v-else class="space-y-4">
            <div class="text-center space-y-1">
              <h2 class="text-base font-black uppercase tracking-wider underline">BERITA ACARA PELAKSANAAN UJIAN</h2>
              <p class="text-xs font-bold uppercase tracking-wide text-slate-800">{{ selectedExamEvent?.name || 'UJIAN SEKOLAH BERBASIS KOMPUTER (CBT)' }}</p>
              <p class="text-[11px] text-slate-600">TAHUN PELAJARAN 2025/2026</p>
            </div>

            <p class="text-xs leading-relaxed text-justify pt-2">
              Pada hari ini <strong>{{ getDayName(printSchedule.start_time) }}</strong>, tanggal <strong>{{ formatScheduleDateFull(printSchedule.start_time) }}</strong>, bertempat di <strong>SMAS Islamic Centre Demak</strong>, telah diselenggarakan Ujian Berbasis Komputer (CBT) untuk sesi:
            </p>

            <!-- Metadata Berita Acara -->
            <table class="w-full text-xs border-collapse pl-4">
              <tbody>
                <tr>
                  <td class="py-1 w-36 font-bold">Mata Pelajaran</td>
                  <td class="py-1 w-4">:</td>
                  <td class="py-1 font-semibold">{{ printSchedule.subject?.name || printSchedule.bank?.subject?.name || '-' }}</td>
                </tr>
                <tr>
                  <td class="py-1 font-bold">Tingkat / Kelas</td>
                  <td class="py-1">:</td>
                  <td class="py-1 font-semibold">{{ printSchedule.class_room?.name || '-' }}</td>
                </tr>
                <tr>
                  <td class="py-1 font-bold">Waktu Pelaksanaan</td>
                  <td class="py-1">:</td>
                  <td class="py-1">{{ formatScheduleTimeOnly(printSchedule.start_time) }} - {{ formatScheduleTimeOnly(printSchedule.end_time) }} WIB ({{ printSchedule.duration_minutes }} Menit)</td>
                </tr>
                <tr>
                  <td class="py-1 font-bold">Kode Token Sesi</td>
                  <td class="py-1">:</td>
                  <td class="py-1 font-mono font-bold">{{ printSchedule.exam_token }}</td>
                </tr>
              </tbody>
            </table>

            <!-- Tabel Rekapitulasi Peserta -->
            <div class="space-y-2 pt-2">
              <p class="font-bold text-xs">I. REKAPITULASI KEHADIRAN PESERTA :</p>
              <table class="w-full border-collapse border border-slate-900 text-xs">
                <thead>
                  <tr class="bg-slate-100 border-b border-slate-900 text-center font-bold">
                    <th class="border border-slate-900 py-2 px-3">Jumlah Terdaftar</th>
                    <th class="border border-slate-900 py-2 px-3">Jumlah Hadir</th>
                    <th class="border border-slate-900 py-2 px-3">Jumlah Tidak Hadir</th>
                    <th class="border border-slate-900 py-2 px-3">Nomor / Nama Peserta Tidak Hadir</th>
                  </tr>
                </thead>
                <tbody>
                  <tr class="text-center">
                    <td class="border border-slate-900 py-4 font-bold">{{ printStudentsList.length }} Siswa</td>
                    <td class="border border-slate-900 py-4 font-bold text-emerald-800">.......... Siswa</td>
                    <td class="border border-slate-900 py-4 font-bold text-rose-800">.......... Siswa</td>
                    <td class="border border-slate-900 py-4 text-left px-3 text-[11px] text-slate-500">
                      .........................................................................
                    </td>
                  </tr>
                </tbody>
              </table>
            </div>

            <!-- Catatan Kejadian Khusus -->
            <div class="space-y-2 pt-2">
              <p class="font-bold text-xs">II. CATATAN / KEJADIAN KHUSUS SELAMA PELAKSANAAN UJIAN :</p>
              <div class="border border-slate-900 p-3 min-h-[90px] text-xs text-slate-600">
                <p class="italic text-slate-400">Pelaksanaan ujian berlangsung tertib, aman, dan lancar tanpa kendala teknis yang berarti.</p>
              </div>
            </div>

            <p class="text-xs pt-2">
              Demikian Berita Acara ini dibuat dengan sesungguhnya dan penuh tanggung jawab untuk dipergunakan sebagaimana mestinya.
            </p>

            <!-- Signatures -->
            <div class="grid grid-cols-2 gap-8 pt-6 text-center text-xs">
              <div>
                <p class="font-bold">Pengawas Ujian 1,</p>
                <div class="h-20"></div>
                <p class="font-black underline uppercase">....................................................</p>
                <p class="text-[11px] text-slate-600">NIP. ............................................</p>
              </div>
              <div>
                <p class="font-bold">Pengawas Ujian 2 / Proktor,</p>
                <div class="h-20"></div>
                <p class="font-black underline uppercase">....................................................</p>
                <p class="text-[11px] text-slate-600">NIP. ............................................</p>
              </div>
            </div>
          </div>

        </div>
      </div>

      <!-- Footer Modal Actions (no-print) -->
      <div class="px-6 py-4 bg-slate-50 border-t border-slate-200 flex items-center justify-between no-print shrink-0">
        <div class="text-xs text-slate-500">
          💡 Tips: Pastikan pengaturan printer memilih ukuran kertas <strong>A4</strong> dan margin <strong>Default / None</strong>.
        </div>
        <div class="flex items-center gap-2">
          <button
            type="button"
            @click="triggerPrintDocument()"
            class="px-4 py-2 bg-emerald-600 hover:bg-emerald-700 active:scale-95 text-white font-bold text-xs rounded-xl shadow-xs transition flex items-center gap-1.5 cursor-pointer"
          >
            <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17 17h2a2 2 0 002-2v-4a2 2 0 00-2-2H5a2 2 0 00-2 2v4a2 2 0 002 2h2m2 4h6a2 2 0 002-2v-4a2 2 0 00-2-2H9a2 2 0 00-2 2v4a2 2 0 002 2zm8-12V5a2 2 0 00-2-2H9a2 2 0 00-2 2v4h10z" />
            </svg>
            <span>🖨️ Cetak / Print Dokumen</span>
          </button>
          <button
            type="button"
            @click="showPrintModal = false"
            class="px-4 py-2 bg-slate-200 hover:bg-slate-300 text-slate-700 font-bold text-xs rounded-xl transition cursor-pointer"
          >
            Tutup
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import EventCardPrintModal from '../../../components/admin/EventCardPrintModal.vue'
import ProctorPrintModal from '@/components/proctor/ProctorPrintModal.vue'
import QuestionPrintModal from '../../../components/admin/QuestionPrintModal.vue'
import { useDashboard } from './context'

const {
  cardPrintEvent,
  formatScheduleDateFull,
  formatScheduleTimeOnly,
  getDayName,
  printDocType,
  printSchedule,
  printStudentsList,
  proctorData,
  questionPrintBank,
  selectedExamEvent,
  showCardPrint,
  showPrintModal,
  showProctorPrint,
  showQuestionPrint,
  triggerPrintDocument,
} = useDashboard()
</script>
