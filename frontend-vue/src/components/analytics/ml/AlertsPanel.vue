<template>
  <div class="alerts-panel">
    <div class="panel-header">
      <h3>
        <Icon name="warning" size="sm" /> Active Alerts
        <span v-if="alerts.length > 0" class="alert-count">{{
          alerts.length
        }}</span>
      </h3>
    </div>

    <!-- Empty State -->
    <div v-if="alerts.length === 0" class="empty-alerts">
      <Icon name="check" size="lg" class="success-icon" />
      <p>No active alerts</p>
    </div>

    <!-- Alerts List -->
    <div v-else class="alerts-list">
      <div
        v-for="alert in displayedAlerts"
        :key="alert.id"
        class="alert-item"
        :class="'severity-' + alert.severity.toLowerCase()"
      >
        <div class="alert-icon">
          <Icon :name="alertIcon(alert.alert_type)" size="md" />
        </div>
        <div class="alert-content">
          <div class="alert-header">
            <span class="alert-type">{{
              formatAlertType(alert.alert_type)
            }}</span>
            <span class="alert-severity">{{ alert.severity }}</span>
          </div>
          <p class="alert-message">{{ alert.message }}</p>
          <span class="alert-product">{{
            alert.product_name || alert.product_id
          }}</span>
        </div>
      </div>

      <!-- Show More -->
      <button
        v-if="alerts.length > 3 && !showAll"
        @click="showAll = true"
        class="show-more-btn"
        type="button"
      >
        Show {{ alerts.length - 3 }} more alerts
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from "vue";
import type { MLAlert } from "@/composables/useMLAnalytics";
import Icon from "@/components/ui/Icon.vue";

const props = defineProps<{
  alerts: MLAlert[];
}>();

const showAll = ref(false);

const displayedAlerts = computed(() => {
  if (showAll.value) return props.alerts;
  return props.alerts.slice(0, 3);
});

function alertIcon(type: string): string {
  switch (type) {
    case "FATIGUE_WARNING":
      return "zap";
    case "CHURN_RISK":
      return "trend-down";
    case "BUDGET_REC":
      return "dollar-sign";
    default:
      return "warning";
  }
}

function formatAlertType(type: string): string {
  switch (type) {
    case "FATIGUE_WARNING":
      return "Creative Fatigue";
    case "CHURN_RISK":
      return "Churn Risk";
    case "BUDGET_REC":
      return "Budget Recommendation";
    default:
      return type.replace(/_/g, " ");
  }
}
</script>

<style scoped>
.alerts-panel {
  background: white;
  border-radius: 12px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.08);
  overflow: hidden;
}

.panel-header {
  padding: 16px 20px;
  border-bottom: 1px solid #e5e7eb;
}

.panel-header h3 {
  margin: 0;
  font-size: 16px;
  font-weight: 600;
  color: #1f2937;
  display: flex;
  align-items: center;
  gap: 8px;
}

.alert-count {
  background: #ef4444;
  color: white;
  font-size: 12px;
  padding: 2px 8px;
  border-radius: 10px;
}

.empty-alerts {
  padding: 40px 20px;
  text-align: center;
  color: #10b981;
}

.empty-alerts :deep(.icon) {
  width: 32px;
  height: 32px;
  margin: 0 auto 8px;
}

.success-icon :deep(.icon) {
  color: #10b981;
}

.empty-alerts p {
  margin: 0;
  color: #6b7280;
}

.alerts-list {
  padding: 12px;
}

.alert-item {
  display: flex;
  gap: 12px;
  padding: 12px;
  border-radius: 8px;
  margin-bottom: 8px;
  border-left: 4px solid;
}

.alert-item:last-child {
  margin-bottom: 0;
}

.severity-high {
  background: #fef2f2;
  border-color: #ef4444;
}

.severity-medium {
  background: #fffbeb;
  border-color: #f59e0b;
}

.severity-low {
  background: #f0f9ff;
  border-color: #3b82f6;
}

.alert-icon {
  font-size: 20px;
  flex-shrink: 0;
}

.alert-content {
  flex: 1;
  min-width: 0;
}

.alert-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 4px;
}

.alert-type {
  font-size: 13px;
  font-weight: 600;
  color: #374151;
}

.alert-severity {
  font-size: 11px;
  font-weight: 600;
  padding: 2px 6px;
  border-radius: 4px;
  text-transform: uppercase;
}

.severity-high .alert-severity {
  background: #fee2e2;
  color: #dc2626;
}

.severity-medium .alert-severity {
  background: #fef3c7;
  color: #d97706;
}

.severity-low .alert-severity {
  background: #dbeafe;
  color: #2563eb;
}

.alert-message {
  font-size: 13px;
  color: #4b5563;
  margin: 0 0 4px 0;
  line-height: 1.4;
}

.alert-product {
  font-size: 12px;
  color: #9ca3af;
}

.show-more-btn {
  width: 100%;
  padding: 10px;
  background: transparent;
  border: 1px dashed #d1d5db;
  border-radius: 6px;
  color: #6b7280;
  font-size: 13px;
  cursor: pointer;
  transition: all 0.2s;
}

.show-more-btn:hover {
  border-color: #9ca3af;
  color: #374151;
}
</style>
