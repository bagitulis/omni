<template>
  <div class="view-container">
    <div v-if="!routes.length" class="empty-state">
      <p>✅ No disconnected routes found!</p>
      <p>All frontend API calls have matching backend endpoints.</p>
    </div>

    <div v-else>
      <div class="disconnected-header">
        <h2>❌ Disconnected Routes</h2>
        <p>Frontend is calling these endpoints, but they don't exist in backend:</p>
      </div>

      <div
        v-for="route in routes"
        :key="route.endpoint"
        class="disconnected-route-item"
      >
        <div class="route-main-disconnected">
          <code class="endpoint-path">{{ route.endpoint }}</code>
          <span class="error-badge">❌ NOT IMPLEMENTED</span>
        </div>

        <div class="route-meta">
          <span class="component-count">
            Called by {{ route.components.length }} component(s)
          </span>
        </div>

        <div class="calling-components">
          <strong>📍 Called From:</strong>
          <div class="components-list">
            <span v-for="comp in route.components" :key="comp" class="component-badge">
              {{ comp }}
            </span>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
interface DisconnectedRoute {
  endpoint: string;
  components: string[];
}

interface Props {
  routes: DisconnectedRoute[];
}

defineProps<Props>();
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

.disconnected-header {
  margin-bottom: 16px;
}

.disconnected-header h2 {
  margin: 0 0 8px 0;
  color: #e74c3c;
}

.disconnected-header p {
  margin: 0;
  color: #7f8c8d;
}

.disconnected-route-item {
  background: white;
  border: 2px solid #ffe0e0;
  border-radius: 8px;
  padding: 16px;
  margin-bottom: 12px;
}

.route-main-disconnected {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 12px;
}

.endpoint-path {
  flex: 1;
  padding: 8px 12px;
  background: #f8f9fa;
  border-radius: 4px;
  font-family: monospace;
  font-size: 0.9em;
}

.error-badge {
  background: #e74c3c;
  color: white;
  padding: 6px 12px;
  border-radius: 4px;
  font-weight: bold;
  font-size: 0.9em;
}

.route-meta {
  margin-bottom: 12px;
  padding-bottom: 12px;
  border-bottom: 1px solid #f0f0f0;
}

.component-count {
  color: #7f8c8d;
  font-size: 0.9em;
}

.calling-components {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.calling-components strong {
  color: #2c3e50;
  font-size: 0.95em;
}

.components-list {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.component-badge {
  background: #e8f1f8;
  color: #2c3e50;
  padding: 6px 12px;
  border-radius: 20px;
  font-size: 0.85em;
  font-weight: 500;
}
</style>
