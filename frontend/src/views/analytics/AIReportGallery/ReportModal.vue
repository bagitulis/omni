<template>
  <div v-if="report" class="modal-overlay" @click="$emit('close')">
    <div class="modal-content" @click.stop>
      <div class="modal-header">
        <h2>{{ report.period_label || report.file_name }}</h2>
        <button @click="$emit('close')" class="btn-close" type="button">
          ✕
        </button>
      </div>
      <div class="modal-body">
        <div v-if="loading" class="loading-report">
          <div class="spinner"></div>
          <p>Loading report...</p>
        </div>
        <div v-else-if="html" class="report-content" v-html="html"></div>
        <div v-else class="error-report">
          <span aria-hidden="true">❌</span>
          <p>Failed to load report</p>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
export interface ReportItem {
  id: number;
  platform: string;
  report_type: string;
  period_label?: string;
  file_name: string;
  created_at: string;
  file_size: number;
}

defineProps<{
  report: ReportItem | null;
  loading: boolean;
  html: string;
}>();

defineEmits<{
  close: [];
}>();
</script>

<style scoped>
@import "./ReportModal.styles.css";

.spinner {
  width: 40px;
  height: 40px;
  margin: 0 auto 1rem;
  border: 4px solid #e2e8f0;
  border-top-color: #667eea;
  border-radius: 50%;
  animation: spin 1s linear infinite;
}

@keyframes spin {
  to {
    transform: rotate(360deg);
  }
}
</style>
