<template>
  <div class="h-screen flex bg-slate-50 overflow-hidden font-sans">
    <DashboardSidebar />

    <!-- Right Area: Top Navbar + Main Content -->
    <div class="flex-1 flex flex-col min-w-0 overflow-hidden">
      <DashboardTopbar />

      <!-- Main Scrollable Area: satu tab tampil sesuai activeTab (v-if di tiap komponen) -->
      <main class="flex-1 overflow-y-auto p-4 sm:p-6 space-y-6">
        <HomeTab />
        <SchedulesTab />
        <EventsTab />
        <SubjectsTab />
        <ClassSubjectsTab />
        <StudentsTab />
        <TeachersTab />
        <ClassesTab />
        <ProctorTab />
        <QuestionBanksTab />
      </main>
    </div>

    <!-- Modal: urutan dipertahankan agar tumpukan (z-order) antarmodal tetap sama -->
    <ScheduleModals />
    <EventFormModal />
    <StudentModals />
    <TeacherModals />
    <ClassModals />
    <SubjectModals />
    <MasterImportModals />
    <ExamDocumentModals />
    <QuestionBankModals />
    <QuestionManagerModal />
    <DashboardDialogs />

    <!-- STUDENT EXAM SIMULATOR MODAL -->
    <StudentExamSimulatorModal
      :is-open="showSimulatorModal"
      :bank="viewingBank"
      :questions="bankQuestions"
      @close="showSimulatorModal = false"
    />
  </div>
</template>

<script setup>
import { onMounted, onUnmounted, provide } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '../../stores/auth'
import StudentExamSimulatorModal from '../../components/admin/StudentExamSimulatorModal.vue'
import DashboardSidebar from './dashboard/DashboardSidebar.vue'
import DashboardTopbar from './dashboard/DashboardTopbar.vue'
import HomeTab from './dashboard/HomeTab.vue'
import SchedulesTab from './dashboard/SchedulesTab.vue'
import EventsTab from './dashboard/EventsTab.vue'
import SubjectsTab from './dashboard/SubjectsTab.vue'
import ClassSubjectsTab from './dashboard/ClassSubjectsTab.vue'
import StudentsTab from './dashboard/StudentsTab.vue'
import TeachersTab from './dashboard/TeachersTab.vue'
import ClassesTab from './dashboard/ClassesTab.vue'
import ProctorTab from './dashboard/ProctorTab.vue'
import QuestionBanksTab from './dashboard/QuestionBanksTab.vue'
import ScheduleModals from './dashboard/ScheduleModals.vue'
import EventFormModal from './dashboard/EventFormModal.vue'
import StudentModals from './dashboard/StudentModals.vue'
import TeacherModals from './dashboard/TeacherModals.vue'
import ClassModals from './dashboard/ClassModals.vue'
import SubjectModals from './dashboard/SubjectModals.vue'
import MasterImportModals from './dashboard/MasterImportModals.vue'
import ExamDocumentModals from './dashboard/ExamDocumentModals.vue'
import QuestionBankModals from './dashboard/QuestionBankModals.vue'
import QuestionManagerModal from './dashboard/QuestionManagerModal.vue'
import DashboardDialogs from './dashboard/DashboardDialogs.vue'
import { DASHBOARD_CONTEXT } from './dashboard/context'
import { landingTab } from '../../utils/access'
import { useDashboardUi } from './dashboard/composables/useDashboardUi'
import { useDashboardData } from './dashboard/composables/useDashboardData'
import { useEvents } from './dashboard/composables/useEvents'
import { useDashboardMetrics } from './dashboard/composables/useDashboardMetrics'
import { useStudents } from './dashboard/composables/useStudents'
import { useStaff } from './dashboard/composables/useStaff'
import { useClasses } from './dashboard/composables/useClasses'
import { useSubjects } from './dashboard/composables/useSubjects'
import { useSchedules } from './dashboard/composables/useSchedules'
import { useQuestionBanks } from './dashboard/composables/useQuestionBanks'
import { useQuestionEditor } from './dashboard/composables/useQuestionEditor'
import { useProctoring } from './dashboard/composables/useProctoring'
import { useHome } from './dashboard/composables/useHome'
import { useDashboardNavigation } from './dashboard/composables/useDashboardNavigation'

const router = useRouter()
const authStore = useAuthStore()

// State dan aksi dasbor dibagi per domain di ./dashboard/composables. Setiap composable menerima
// ctx berisi hasil composable sebelumnya, lalu hasilnya digabung ke ctx. Urutan ini penting:
// composable hanya boleh memakai (saat setup) nama dari composable yang dipanggil sebelumnya.
const ctx = { router, authStore }
Object.assign(ctx, useDashboardUi(ctx))
Object.assign(ctx, useDashboardData(ctx))
Object.assign(ctx, useEvents(ctx))
Object.assign(ctx, useDashboardMetrics(ctx))
Object.assign(ctx, useStudents(ctx))
Object.assign(ctx, useStaff(ctx))
Object.assign(ctx, useClasses(ctx))
Object.assign(ctx, useSubjects(ctx))
Object.assign(ctx, useSchedules(ctx))
Object.assign(ctx, useQuestionBanks(ctx))
Object.assign(ctx, useQuestionEditor(ctx))
Object.assign(ctx, useProctoring(ctx))
Object.assign(ctx, useHome(ctx))
Object.assign(ctx, useDashboardNavigation(ctx))

const {
  bankQuestions,
  closeEventMenuOnClickOutside,
  loadAllData,
  loadLoginSessions,
  loadTeacherHomeProctorTasks,
  showSimulatorModal,
  switchTab,
  viewingBank,
} = ctx

onUnmounted(() => {
  window.removeEventListener('click', closeEventMenuOnClickOutside)
})

onMounted(() => {
  window.addEventListener('click', closeEventMenuOnClickOutside)
  const queryTab = router.currentRoute.value.query.tab
  const initialTab = queryTab || landingTab(authStore.user)
  if (initialTab !== 'dashboard') {
    switchTab(initialTab)
  } else {
    loadLoginSessions()
    loadTeacherHomeProctorTasks()
  }
  loadAllData()
})

// Bagikan state dan aksi ke komponen tab/modal di ./dashboard (lihat dashboard/context.js).
provide(DASHBOARD_CONTEXT, ctx)
</script>

