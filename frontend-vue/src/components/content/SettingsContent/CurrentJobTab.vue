<template>
  <div v-if="activeTab === 'current'" class="tab-content">
    <div v-if="!currentJob" class="empty-state">
      <p>No job currently running</p>
      <p class="text-muted">Waiting for next job in queue...</p>
    </div>

    <CurrentJobCard
      v-else
      :job="currentJob"
      :isJobStuck="isJobStuck"
      @cancel="$emit('cancel-job', currentJob.id)"
      @force-cancel="$emit('force-cancel', currentJob.id)"
    />
  </div>
</template>

<script setup lang="ts">
import { type Job } from "./currentJobUtils";
import CurrentJobCard from "./CurrentJobCard.vue";

defineProps<{
  activeTab: string;
  currentJob: Job | null;
  isJobStuck: boolean;
}>();

defineEmits<{
  "cancel-job": [jobId: string];
  "force-cancel": [jobId: string];
}>();
</script>

<style scoped lang="css">
@import "./ScriptMonitor.styles.css";

.tab-content {
  animation: slideInUp 0.3s cubic-bezier(0.4, 0, 0.2, 1);
}

@keyframes slideInUp {
  from {
    opacity: 0;
    transform: translateY(12px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 16px;
  padding: 80px 20px;
  background: #f8f9fa;
  border: 2px dashed #e0e0e0;
  border-radius: 12px;
  text-align: center;
}

.empty-state p {
  margin: 0;
  font-size: 16px;
  font-weight: 600;
  color: #1a1a1a;
}

.empty-state p.text-muted {
  font-weight: 400;
  color: #6b7280;
  font-size: 14px;
}
</style>
