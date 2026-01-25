<template>
  <div class="health-card" :class="healthData.status">
    <div class="health-icon-large">
      {{ healthStatusEmoji[healthData.status] }}
    </div>
    <div class="health-main-content">
      <div class="health-header">
        <span class="health-status-label">{{ healthData.status.toUpperCase() }}</span>
        <span class="health-percentage">{{ healthData.healthyPercent }}%</span>
      </div>
      <div class="health-description">
        {{ healthData.healthy }} of {{ healthData.total }} routes healthy
      </div>
      <div class="health-badges-row">
        <div class="health-badge healthy">
          <span class="badge-icon">✅</span>
          <span class="badge-label">{{ healthData.healthy }}<br><small>Healthy</small></span>
        </div>
        <div class="health-badge warning">
          <span class="badge-icon">⚠️</span>
          <span class="badge-label">{{ healthData.warning }}<br><small>Warning</small></span>
        </div>
        <div class="health-badge error">
          <span class="badge-icon">❌</span>
          <span class="badge-label">{{ healthData.error }}<br><small>Error</small></span>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
/**
 * Route Health Status Card
 * Displays overall system health status
 * Single Responsibility: Show health status metrics only
 */

interface HealthData {
  total: number;
  healthy: number;
  warning: number;
  error: number;
  healthyPercent: number;
  status: 'good' | 'fair' | 'poor';
}

defineProps<{
  healthData: HealthData;
}>();

const healthStatusEmoji: Record<string, string> = {
  'good': '😊',
  'fair': '😐',
  'poor': '😟',
};
</script>

<style scoped lang="css">
@import './ScriptMonitor.styles.css';

.health-card {
  display: flex;
  gap: 20px;
  padding: 24px;
  background: linear-gradient(135deg, #f8f9fa 0%, #ffffff 100%);
  border-radius: 12px;
  border-left: 5px solid #0066cc;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.05);
  transition: all 0.3s ease;
}

.health-card:hover {
  box-shadow: 0 4px 16px rgba(0, 0, 0, 0.1);
  transform: translateY(-2px);
}

.health-card.good {
  border-left-color: #27ae60;
  background: linear-gradient(135deg, #f0fdf4 0%, #ffffff 100%);
}

.health-card.fair {
  border-left-color: #f39c12;
  background: linear-gradient(135deg, #fffbf0 0%, #ffffff 100%);
}

.health-card.poor {
  border-left-color: #e74c3c;
  background: linear-gradient(135deg, #fdf8f7 0%, #ffffff 100%);
}

.health-icon-large {
  font-size: 48px;
  display: flex;
  align-items: center;
  justify-content: center;
  min-width: 64px;
  animation: pulse-icon 2s ease-in-out infinite;
}

@keyframes pulse-icon {
  0%, 100% { transform: scale(1); }
  50% { transform: scale(1.1); }
}

.health-main-content {
  flex: 1;
}

.health-header {
  display: flex;
  align-items: baseline;
  gap: 16px;
  margin-bottom: 8px;
}

.health-status-label {
  font-size: 16px;
  font-weight: 700;
  color: #1a1a1a;
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

.health-percentage {
  font-size: 32px;
  font-weight: 800;
  color: #0066cc;
  font-variant-numeric: tabular-nums;
}

.health-description {
  font-size: 14px;
  color: #666;
  margin-bottom: 12px;
  line-height: 1.5;
}

.health-badges-row {
  display: flex;
  gap: 12px;
}

.health-badge {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 6px;
  padding: 10px 14px;
  border-radius: 8px;
  text-align: center;
  transition: all 0.3s ease;
  font-size: 12px;
  font-weight: 600;
}

.health-badge:hover {
  transform: translateY(-2px);
}

.health-badge.healthy {
  background: #d4edda;
  color: #155724;
  border: 1px solid #c3e6cb;
}

.health-badge.warning {
  background: #fff3cd;
  color: #856404;
  border: 1px solid #ffeaa7;
}

.health-badge.error {
  background: #f8d7da;
  color: #721c24;
  border: 1px solid #f5c6cb;
}

.badge-icon {
  font-size: 16px;
}

.badge-label {
  font-size: 11px;
}

.badge-label small {
  display: block;
  font-size: 10px;
  opacity: 0.8;
  margin-top: 2px;
}

@media (max-width: 768px) {
  .health-card {
    flex-direction: column;
  }

  .health-header {
    flex-direction: column;
    align-items: flex-start;
  }

  .health-badges-row {
    flex-direction: column;
  }

  .health-badge {
    width: 100%;
  }
}
</style>
