<template>
  <div v-if="queueInfo.queueSize > 0" class="queue-section">
    <!-- Queue Header -->
    <div class="section-header">
      <div class="header-left">
        <span class="section-icon">⏳</span>
        <span class="section-title">Navigation Queue</span>
        <span class="queue-badge">{{ queueInfo.queueSize }}</span>
      </div>
      <div class="header-right">
        <p class="section-subtitle">Pending requests waiting for processing</p>
      </div>
    </div>

    <!-- Queue Content -->
    <div class="queue-container">
      <div class="queue-stats">
        <div class="stat-item">
          <span class="stat-label">Pending Requests</span>
          <span class="stat-value">{{ queueInfo.queueSize }}</span>
        </div>
        <div v-if="queueInfo.isProcessing" class="stat-item">
          <span class="stat-label">Currently Processing</span>
          <span class="stat-value processing">{{ queueInfo.isProcessing }}</span>
        </div>
      </div>

      <div v-if="queueInfo.queued.length > 0" class="queued-items">
        <div 
          v-for="(item, idx) in queueInfo.queued" 
          :key="`queue-${idx}`"
          class="queued-item"
          :style="{ animationDelay: `${idx * 50}ms` }"
        >
          <div class="item-position">{{ idx + 1 }}</div>
          <div class="item-details">
            <div class="item-name">{{ item.routeName }}</div>
            <div class="item-meta">
              <span class="wait-time">⏱️ {{ formatWaitTime(item.waitTime) }} waiting</span>
              <span class="priority-badge" :class="`priority-${item.priority}`">
                {{ item.priority }}
              </span>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
/**
 * Route Queue Section
 * Single Responsibility: Display navigation queue status
 */

interface QueueItem {
  routeName: string;
  waitTime: number;
  priority: 'normal' | 'high';
}

interface QueueInfo {
  queueSize: number;
  isProcessing: string | null;
  queued: QueueItem[];
}

defineProps<{
  queueInfo: QueueInfo;
}>();

const formatWaitTime = (ms: number): string => {
  const seconds = Math.floor(ms / 1000);
  if (seconds < 60) return `${seconds}s`;
  const minutes = Math.floor(seconds / 60);
  return `${minutes}m ${seconds % 60}s`;
};
</script>

<style scoped lang="css">
.queue-section {
  animation: slideInUp 0.4s ease-out;
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

.section-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 16px;
  padding-bottom: 12px;
  border-bottom: 2px solid #f0f1f3;
}

.header-left {
  display: flex;
  align-items: center;
  gap: 12px;
}

.section-icon {
  font-size: 20px;
  animation: float 3s ease-in-out infinite;
}

@keyframes float {
  0%, 100% { transform: translateY(0); }
  50% { transform: translateY(-4px); }
}

.section-title {
  font-size: 16px;
  font-weight: 700;
  color: #1a1a1a;
}

.queue-badge {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-width: 28px;
  height: 28px;
  padding: 0 8px;
  background: #e3f2fd;
  color: #0066cc;
  border-radius: 6px;
  font-size: 12px;
  font-weight: 700;
}

.section-subtitle {
  font-size: 13px;
  color: #6b7280;
  margin: 0;
}

.queue-container {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.queue-stats {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(150px, 1fr));
  gap: 12px;
  padding: 12px;
  background: #f8f9fa;
  border-radius: 8px;
}

.stat-item {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.stat-label {
  font-size: 12px;
  color: #6b7280;
  text-transform: uppercase;
  font-weight: 600;
  letter-spacing: 0.3px;
}

.stat-value {
  font-size: 18px;
  font-weight: 700;
  color: #1a1a1a;
}

.stat-value.processing {
  color: #0066cc;
  font-family: 'Courier New', monospace;
}

.queued-items {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.queued-item {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px;
  background: #ffffff;
  border: 1px solid #e0e0e0;
  border-radius: 8px;
  transition: all 0.2s ease;
  animation: slideInLeft 0.3s ease-out forwards;
  opacity: 0;
}

@keyframes slideInLeft {
  from {
    opacity: 0;
    transform: translateX(-12px);
  }
  to {
    opacity: 1;
    transform: translateX(0);
  }
}

.queued-item:hover {
  background: #f8f9fa;
  border-color: #0066cc;
  box-shadow: 0 2px 8px rgba(0, 102, 204, 0.1);
}

.item-position {
  display: flex;
  align-items: center;
  justify-content: center;
  min-width: 32px;
  height: 32px;
  background: #e3f2fd;
  color: #0066cc;
  border-radius: 50%;
  font-weight: 700;
  font-size: 14px;
  flex-shrink: 0;
}

.item-details {
  flex: 1;
  min-width: 0;
}

.item-name {
  font-size: 14px;
  font-weight: 600;
  color: #1a1a1a;
  margin-bottom: 4px;
}

.item-meta {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 12px;
}

.wait-time {
  color: #6b7280;
}

.priority-badge {
  padding: 2px 8px;
  border-radius: 4px;
  font-size: 11px;
  font-weight: 600;
  text-transform: uppercase;
}

.priority-normal {
  background: #e3f2fd;
  color: #0066cc;
}

.priority-high {
  background: #fff3cd;
  color: #856404;
  animation: pulse 2s ease-in-out infinite;
}

@keyframes pulse {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.7; }
}
</style>
