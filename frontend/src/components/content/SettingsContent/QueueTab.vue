<template>
  <div v-if="activeTab === 'queue'" class="tab-content">
    <!-- Pending Queue Jobs -->
    <div v-if="pendingQueue.length > 0" class="queue-section">
      <h3 class="queue-section-title">📋 Pending Jobs (Executing)</h3>
      <div class="queue-list">
        <QueueJobItem
          v-for="(job, index) in pendingQueue"
          :key="job.id"
          :job="job"
          :index="index"
          @cancel="$emit('force-cancel-job', job.id)"
        />
      </div>
    </div>

    <!-- Scheduled Auto-Functions -->
    <div v-if="scheduledAutoFunctions.length > 0" class="queue-section">
      <h3 class="queue-section-title">
        ⏰ Scheduled Auto-Functions (Upcoming)
      </h3>
      <div class="scheduled-list">
        <ScheduledFunctionItem
          v-for="(config, index) in scheduledAutoFunctions"
          :key="`scheduled-${config.id}`"
          :config="config"
          :index="index"
          @cancel="$emit('cancel-scheduled', config.id, config.name)"
        />
      </div>
    </div>

    <!-- Empty State -->
    <div
      v-if="pendingQueue.length === 0 && scheduledAutoFunctions.length === 0"
      class="empty-state"
    >
      <p>Queue is empty</p>
      <p class="text-muted">No pending jobs or scheduled auto-functions</p>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from "vue";
import { type Job, type AutoFunctionConfig } from "./queueUtils";
import QueueJobItem from "./QueueJobItem.vue";
import ScheduledFunctionItem from "./ScheduledFunctionItem.vue";

const props = defineProps<{
  activeTab: string;
  pendingQueue: Job[];
  autoFunctionConfigs: AutoFunctionConfig[];
}>();

defineEmits<{
  "force-cancel-job": [jobId: string];
  "cancel-scheduled": [configId: number, configName: string];
}>();

const scheduledAutoFunctions = computed(() => {
  return props.autoFunctionConfigs.filter(
    (config: AutoFunctionConfig) => config.next_scheduled_execution,
  );
});
</script>

<style scoped lang="css">
@import "./ScriptMonitor.styles.css";

.tab-content {
  display: flex;
  flex-direction: column;
  gap: 20px;
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

.queue-section {
  display: flex;
  flex-direction: column;
  gap: 16px;
  padding: 20px;
  background: #ffffff;
  border: 1px solid #e0e0e0;
  border-radius: 10px;
}

.queue-section-title {
  margin: 0 0 12px 0;
  font-size: 15px;
  font-weight: 700;
  color: #1a1a1a;
  display: flex;
  align-items: center;
  gap: 8px;
}

.queue-list,
.scheduled-list {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 12px;
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

@media (max-width: 768px) {
  .tab-content {
    gap: 16px;
  }

  .queue-section {
    padding: 16px;
  }
}
</style>
