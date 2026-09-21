<template>
  <div class="h-full flex flex-col bg-surface-50 exam-protection overflow-hidden select-none">
    <!-- Zone 1: Fixed Topbar -->
    <ExamTopbar />

    <!-- Zone 2: Scrollable Main Question Area -->
    <main class="flex-1 overflow-y-auto pb-6">
      <ExamQuestionCard />
    </main>

    <!-- Zone 3: Fixed Bottom Navigation Bar -->
    <ExamBottombar />

    <!-- Bottom Sheet Drawer for Question Numbers (1-N) -->
    <ExamQuestionDrawer />

    <!-- Modals (Anti-cheat Warning, Total Lockout, Final Submit Confirmation) -->
    <ExamModals 
      :warning-message="warningMsg" 
      @dismiss-warning="warningMsg = ''" 
    />
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useExamStore } from '../../stores/exam'
import { useAntiCheat } from '../../composables/useAntiCheat'

import ExamTopbar from '../../components/exam/ExamTopbar.vue'
import ExamQuestionCard from '../../components/exam/ExamQuestionCard.vue'
import ExamBottombar from '../../components/exam/ExamBottombar.vue'
import ExamQuestionDrawer from '../../components/exam/ExamQuestionDrawer.vue'
import ExamModals from '../../components/exam/ExamModals.vue'

const router = useRouter()
const examStore = useExamStore()
const warningMsg = ref('')

// Initialize Anti-Cheat Engine
useAntiCheat(examStore, (msg) => {
  warningMsg.value = msg
})

onMounted(() => {
  // If exam store has no session or questions (e.g. direct page refresh), try restoring
  if (!examStore.sessionId || examStore.questions.length === 0) {
    const activeScheduleId = localStorage.getItem('cbt_active_schedule_id')
    const activeToken = localStorage.getItem('cbt_active_token')
    if (activeScheduleId && activeToken) {
      examStore.startExam(activeScheduleId, activeToken).catch(() => {
        router.push('/student')
      })
    } else {
      router.push('/student')
    }
  }
})
</script>
