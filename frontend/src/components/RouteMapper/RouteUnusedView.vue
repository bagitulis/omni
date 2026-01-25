<template>
  <div class="view-container">
    <div v-if="!routes.length" class="empty-state">
      <p>{{ emptyTitle }}</p>
      <p>{{ emptyDescription }}</p>
    </div>

    <div v-else>
      <div class="unused-header">
        <h2>⚠️ {{ title }}</h2>
        <p>{{ description }}</p>
      </div>

      <div v-for="route in routes" :key="route.endpoint" class="unused-route-item">
        <div class="route-main-unused">
          <span class="method-badge" :class="`method-${route.method}`">
            {{ route.method }}
          </span>
          <code class="endpoint-path">{{ route.endpoint }}</code>
          <span v-if="route.is_dynamic" class="dynamic-badge">🔄 Dynamic</span>
        </div>

        <div class="route-info-unused">
          <span class="warning-badge">⚠️ NOT USED</span>
          <span class="status-text">No component is calling this endpoint</span>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { withDefaults } from 'vue';

interface UnusedRoute {
  endpoint: string;
  method: string;
  is_dynamic?: boolean;
}

interface Props {
  routes: UnusedRoute[];
  title?: string;
  description?: string;
  emptyTitle?: string;
  emptyDescription?: string;
}

withDefaults(defineProps<Props>(), {
  title: 'Unused Routes',
  description: 'Routes defined in backend but NOT called by any frontend component:',
  emptyTitle: '✅ All backend routes are being used!',
  emptyDescription: 'Every route in the backend is called by at least one component.',
});
</script>

<style scoped>
.view-container {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.empty-state {
  padding: 40px 20px;
  text-align: center;
  background: #f0fff4;
  border: 1px solid #bbf7d0;
  border-radius: 8px;
  color: #065f46;
}

.empty-state p {
  margin: 8px 0;
}

.unused-header {
  margin-bottom: 16px;
}

.unused-header h2 {
  margin: 0 0 8px 0;
  color: #f39c12;
}

.unused-header p {
  margin: 0;
  color: #7f8c8d;
}

.unused-route-item {
  background: white;
  border: 2px solid #ffe8b6;
  border-radius: 8px;
  padding: 16px;
  margin-bottom: 12px;
}

.route-main-unused {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 12px;
}

.method-badge {
  padding: 6px 10px;
  border-radius: 4px;
  font-size: 0.85em;
  font-weight: bold;
  color: white;
  min-width: 50px;
  text-align: center;
}

.method-GET {
  background: #3498db;
}

.method-POST {
  background: #27ae60;
}

.method-PUT {
  background: #f39c12;
}

.method-DELETE {
  background: #e74c3c;
}

.endpoint-path {
  flex: 1;
  padding: 8px 12px;
  background: #f8f9fa;
  border-radius: 4px;
  font-family: monospace;
  font-size: 0.9em;
}

.dynamic-badge {
  background: #9b59b6;
  color: white;
  padding: 6px 10px;
  border-radius: 4px;
  font-size: 0.85em;
  font-weight: bold;
}

.route-info-unused {
  display: flex;
  align-items: center;
  gap: 12px;
  padding-top: 12px;
  border-top: 1px solid #f0f0f0;
}

.warning-badge {
  background: #f39c12;
  color: white;
  padding: 6px 12px;
  border-radius: 4px;
  font-weight: bold;
  font-size: 0.85em;
}

.status-text {
  color: #7f8c8d;
  font-size: 0.9em;
}
</style>
