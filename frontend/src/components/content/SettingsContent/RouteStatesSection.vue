<template>
  <div class="route-states-section">
    <!-- Section Header -->
    <div class="section-header">
      <div class="header-left">
        <span class="section-icon">📍</span>
        <span class="section-title">Route States</span>
        <span class="routes-count-badge">{{ routeStates.length }}</span>
      </div>
      <div class="header-right">
        <p class="section-subtitle">Current state of all routes</p>
      </div>
    </div>

    <!-- Route States Grid -->
    <div class="states-grid">
      <div 
        v-for="route in routeStates" 
        :key="`route-state-${route.routeName}`"
        class="state-card"
        :class="`state-${route.state}`"
      >
        <!-- Card Icon -->
        <div class="state-icon-box">
          <span class="state-icon">{{ route.icon }}</span>
        </div>

        <!-- Card Content -->
        <div class="state-content">
          <div class="state-header">
            <span class="state-name">{{ route.routeName }}</span>
          </div>

          <div class="state-badge-row">
            <span class="state-badge" :class="`badge-${route.state}`">
              {{ route.state.toUpperCase() }}
            </span>
          </div>

          <div class="state-timestamp">
            <span class="timestamp-icon">⏰</span>
            <span class="timestamp-text">{{ route.timestamp }}</span>
          </div>

          <div v-if="route.error" class="state-error">
            <span class="error-icon">❗</span>
            <span class="error-message">{{ route.error }}</span>
          </div>
        </div>
      </div>

      <!-- Empty State -->
      <div v-if="routeStates.length === 0" class="empty-state-message">
        <span class="empty-icon">📭</span>
        <p>No route state data available</p>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
/**
 * Route States Section
 * Single Responsibility: Display route state information
 */

interface RouteState {
  routeName: string;
  state: string;
  icon: string;
  error: string | null;
  timestamp: string;
}

defineProps<{
  routeStates: RouteState[];
}>();
</script>

<style scoped lang="css">
.route-states-section {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.section-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
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
  animation: rotate 3s linear infinite;
}

@keyframes rotate {
  from { transform: rotate(0deg); }
  to { transform: rotate(360deg); }
}

.section-title {
  font-size: 16px;
  font-weight: 700;
  color: #1a1a1a;
}

.routes-count-badge {
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

.states-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(240px, 1fr));
  gap: 12px;
  animation: fadeIn 0.3s ease;
}

@keyframes fadeIn {
  from {
    opacity: 0;
  }
  to {
    opacity: 1;
  }
}

.state-card {
  display: flex;
  flex-direction: column;
  gap: 12px;
  padding: 16px;
  background: #ffffff;
  border: 2px solid #e0e0e0;
  border-radius: 10px;
  transition: all 0.2s ease;
}

.state-card:hover {
  border-color: #0066cc;
  box-shadow: 0 4px 12px rgba(0, 102, 204, 0.15);
  transform: translateY(-2px);
}

.state-card.state-idle {
  background: linear-gradient(135deg, #f8f9fa 0%, #ffffff 100%);
  border-color: #27ae60;
}

.state-card.state-idle:hover {
  border-color: #27ae60;
  box-shadow: 0 4px 12px rgba(39, 174, 96, 0.15);
}

.state-card.state-loading {
  background: linear-gradient(135deg, #e3f2fd 0%, #ffffff 100%);
  border-color: #0066cc;
}

.state-card.state-loading:hover {
  border-color: #0066cc;
  box-shadow: 0 4px 12px rgba(0, 102, 204, 0.15);
}

.state-card.state-error {
  background: linear-gradient(135deg, #fdf8f7 0%, #ffffff 100%);
  border-color: #e74c3c;
}

.state-card.state-error:hover {
  border-color: #e74c3c;
  box-shadow: 0 4px 12px rgba(231, 76, 60, 0.15);
}

.state-icon-box {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 44px;
  height: 44px;
  background: #f8f9fa;
  border-radius: 8px;
  font-size: 24px;
}

.state-card.state-loading .state-icon-box {
  animation: pulse 1.5s ease-in-out infinite;
}

@keyframes pulse {
  0%, 100% { background-color: #f8f9fa; }
  50% { background-color: #e3f2fd; }
}

.state-content {
  display: flex;
  flex-direction: column;
  gap: 8px;
  flex: 1;
}

.state-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
}

.state-name {
  font-size: 14px;
  font-weight: 700;
  color: #1a1a1a;
  word-break: break-word;
}

.state-badge-row {
  display: flex;
  gap: 6px;
}

.state-badge {
  display: inline-block;
  padding: 4px 10px;
  border-radius: 6px;
  font-size: 11px;
  font-weight: 700;
  text-transform: uppercase;
  letter-spacing: 0.3px;
}

.badge-idle {
  background: #c6efce;
  color: #270;
}

.badge-loading {
  background: #bde4ff;
  color: #004085;
  animation: fadeInOut 1.5s ease-in-out infinite;
}

@keyframes fadeInOut {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.7; }
}

.badge-error {
  background: #f8d7da;
  color: #721c24;
  animation: shake 0.5s ease-in-out;
}

@keyframes shake {
  0%, 100% { transform: translateX(0); }
  25% { transform: translateX(-2px); }
  75% { transform: translateX(2px); }
}

.state-timestamp {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 12px;
  color: #6b7280;
}

.timestamp-icon {
  font-size: 12px;
}

.timestamp-text {
  font-family: 'Courier New', monospace;
}

.state-error {
  display: flex;
  gap: 8px;
  padding: 8px;
  background: #fdf8f7;
  border-radius: 6px;
  border-left: 3px solid #e74c3c;
}

.error-icon {
  font-size: 14px;
  flex-shrink: 0;
}

.error-message {
  font-size: 12px;
  color: #c33;
  line-height: 1.4;
  word-break: break-word;
}

.empty-state-message {
  grid-column: 1 / -1;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 12px;
  padding: 40px 20px;
  background: #f8f9fa;
  border-radius: 10px;
  text-align: center;
}

.empty-icon {
  font-size: 48px;
}

.empty-state-message p {
  margin: 0;
  font-size: 14px;
  color: #6b7280;
}
</style>
