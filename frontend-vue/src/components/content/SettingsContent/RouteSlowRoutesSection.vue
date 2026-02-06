<template>
  <div v-if="slowRoutes.length > 0" class="slow-routes-section">
    <!-- Section Header -->
    <div class="section-header alert">
      <div class="header-left">
        <span class="section-icon warning">⚠️</span>
        <span class="section-title">Slow Routes Detected</span>
        <span class="route-count-badge">{{ slowRoutes.length }}</span>
      </div>
    </div>

    <!-- Slow Routes List -->
    <div class="slow-routes-list">
      <div 
        v-for="(route, idx) in slowRoutes" 
        :key="`slow-${route.name}`"
        class="slow-route-card"
        :style="{ animationDelay: `${idx * 50}ms` }"
      >
        <div class="route-header">
          <span class="route-icon">🐢</span>
          <span class="route-name">{{ route.name }}</span>
        </div>
        <div class="route-details">
          <div class="detail-item">
            <span class="detail-label">Avg Response Time</span>
            <span class="detail-value">{{ route.avgTime }}</span>
          </div>
          <div class="detail-item">
            <span class="detail-label">Slowness Factor</span>
            <span class="detail-value slowness">{{ route.slowness }}</span>
          </div>
        </div>
      </div>
    </div>

    <!-- Tips -->
    <div class="tips-box">
      <p class="tips-title">💡 Optimization Tips:</p>
      <ul class="tips-list">
        <li>Check database query performance</li>
        <li>Review API response times</li>
        <li>Consider caching frequently accessed data</li>
      </ul>
    </div>
  </div>
</template>

<script setup lang="ts">
/**
 * Slow Routes Section
 * Single Responsibility: Display slow routes alert and details
 */

interface SlowRoute {
  name: string;
  avgTime: string;
  slowness: string;
}

defineProps<{
  slowRoutes: SlowRoute[];
}>();
</script>

<style scoped lang="css">
.slow-routes-section {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.section-header {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px;
  background: linear-gradient(135deg, #fffbf0 0%, #fff8f0 100%);
  border-left: 4px solid #f39c12;
  border-radius: 8px;
}

.section-header.alert {
  border-left-color: #f39c12;
}

.header-left {
  display: flex;
  align-items: center;
  gap: 12px;
  flex: 1;
}

.section-icon {
  font-size: 20px;
  animation: bounce 2s ease-in-out infinite;
}

@keyframes bounce {
  0%, 100% { transform: translateY(0); }
  50% { transform: translateY(-4px); }
}

.section-title {
  font-size: 16px;
  font-weight: 700;
  color: #856404;
}

.route-count-badge {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-width: 28px;
  height: 28px;
  padding: 0 8px;
  background: #f39c12;
  color: white;
  border-radius: 6px;
  font-size: 12px;
  font-weight: 700;
}

.slow-routes-list {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.slow-route-card {
  display: flex;
  flex-direction: column;
  gap: 10px;
  padding: 12px;
  background: #ffffff;
  border: 1px solid #ffe8c8;
  border-radius: 8px;
  animation: slideInUp 0.3s ease-out forwards;
  opacity: 0;
  transition: all 0.2s ease;
}

@keyframes slideInUp {
  from {
    opacity: 0;
    transform: translateY(8px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

.slow-route-card:hover {
  background: #fffbf0;
  border-color: #f39c12;
  box-shadow: 0 2px 8px rgba(243, 156, 18, 0.1);
}

.route-header {
  display: flex;
  align-items: center;
  gap: 10px;
}

.route-icon {
  font-size: 16px;
}

.route-name {
  font-size: 14px;
  font-weight: 600;
  color: #1a1a1a;
}

.route-details {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 10px;
}

.detail-item {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.detail-label {
  font-size: 11px;
  color: #6b7280;
  text-transform: uppercase;
  font-weight: 600;
  letter-spacing: 0.3px;
}

.detail-value {
  font-size: 13px;
  font-weight: 700;
  color: #1a1a1a;
  font-family: 'Courier New', monospace;
}

.detail-value.slowness {
  color: #f39c12;
}

.tips-box {
  padding: 12px;
  background: #f8f9fa;
  border-radius: 8px;
  border-left: 3px solid #f39c12;
}

.tips-title {
  margin: 0 0 8px 0;
  font-size: 13px;
  font-weight: 700;
  color: #1a1a1a;
}

.tips-list {
  margin: 0;
  padding-left: 20px;
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.tips-list li {
  font-size: 12px;
  color: #666;
  line-height: 1.4;
}
</style>
