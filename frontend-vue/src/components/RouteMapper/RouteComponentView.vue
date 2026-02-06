<template>
  <div class="view-container">
    <div
      v-for="(component, compName) in components"
      :key="compName"
      class="component-section"
    >
      <div class="component-header">
        <h2>{{ compName }}</h2>
        <span class="component-meta">
          {{ component.category }} •
          {{ component.routes_called?.length || 0 }} routes
        </span>
      </div>

      <div class="component-details">
        <div v-if="component.path" class="detail-item">
          <strong>📁 Path:</strong>
          <code>{{ component.path }}</code>
        </div>

        <div v-if="component.routes_called?.length" class="detail-item">
          <strong>🔗 API Endpoints Called:</strong>
          <div class="endpoints-detailed">
            <div
              v-for="endpoint in component.routes_called"
              :key="endpoint"
              class="endpoint-detailed-item"
            >
              <span class="method-badge" :class="`method-${getMethodForEndpoint(endpoint)}`">
                {{ getMethodForEndpoint(endpoint) }}
              </span>
              <code class="endpoint-path">{{ endpoint }}</code>
              <span class="status-badge connected-badge">✓ Connected</span>
            </div>
          </div>
        </div>

        <div v-if="component.buttons?.length" class="detail-item">
          <strong>🔘 Actions/Buttons (triggers):</strong>
          <div class="buttons-info">
            <div v-for="btn in component.buttons" :key="btn" class="button-item">
              <span class="button-name">{{ btn }}</span>
              <span class="button-description">
                Calls: {{ getRoutesForButton(compName, btn).join(", ") || "Multiple routes" }}
              </span>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
interface Component {
  category?: string;
  path?: string;
  routes_called?: string[];
  buttons?: string[];
}

interface Props {
  components: Record<string, Component>;
  buttonToEndpoints: Record<string, Record<string, string[]>>;
}

const props = defineProps<Props>();

const getMethodForEndpoint = (endpoint: string): string => {
  const methodMatch = endpoint.match(/^\[(\w+)\]/);
  return methodMatch ? methodMatch[1] : "GET";
};

const getRoutesForButton = (component: string, button: string): string[] => {
  return (props.buttonToEndpoints?.[component]?.[button] as string[]) || [];
};
</script>

<style scoped>
.view-container {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.component-section {
  background: white;
  border: 1px solid #ddd;
  border-radius: 8px;
  padding: 16px;
  margin-bottom: 12px;
}

.component-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12px;
  padding-bottom: 12px;
  border-bottom: 2px solid #f0f0f0;
}

.component-header h2 {
  margin: 0;
  color: #2c3e50;
  font-size: 1.1em;
}

.component-meta {
  color: #7f8c8d;
  font-size: 0.9em;
}

.component-details {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.detail-item {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.detail-item strong {
  color: #2c3e50;
  font-size: 0.95em;
}

.endpoints-detailed,
.buttons-info {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.endpoint-detailed-item,
.button-item {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 8px;
  background: #f8f9fa;
  border-radius: 6px;
}

.method-badge {
  padding: 4px 8px;
  border-radius: 4px;
  font-size: 0.85em;
  font-weight: bold;
  color: white;
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
  padding: 4px 8px;
  background: white;
  border-radius: 4px;
  font-family: monospace;
  font-size: 0.9em;
}

.connected-badge {
  background: #27ae60;
  color: white;
  padding: 4px 8px;
  border-radius: 4px;
  font-size: 0.85em;
}

.button-name {
  font-weight: 500;
  color: #2c3e50;
  min-width: 120px;
}

.button-description {
  flex: 1;
  color: #7f8c8d;
  font-size: 0.9em;
}
</style>
