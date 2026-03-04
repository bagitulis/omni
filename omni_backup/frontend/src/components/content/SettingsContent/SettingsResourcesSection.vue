<template>
  <div class="resources-section">
    <div class="resources-header">
      <h2>📊 Resource Usage</h2>
      <button
        class="btn btn-refresh"
        @click="refreshResources"
        :disabled="resourcesLoading"
      >
        {{ resourcesLoading ? "Loading..." : "Refresh" }}
      </button>
    </div>

    <div v-if="resourcesError" class="error-message">
      ❌ {{ resourcesError }}
    </div>

    <div v-if="resourcesData" class="resources-grid">
      <div class="resource-card backend-card">
        <h3>🖥️ Backend Resources</h3>
        <ResourceMetric
          label="Memory Usage:"
          :value="`${resourcesData.backend.memory.used_mb} MB`"
          :percent="resourcesData.backend.memory.percent"
        />
        <ResourceMetric
          label="CPU Usage:"
          :value="`${resourcesData.backend.cpu.percent}%`"
          :percent="resourcesData.backend.cpu.percent"
        />
        <ResourceMetric
          label="Threads:"
          :value="resourcesData.backend.threads"
          :bare="true"
        />
        <ResourceMetric
          label="Process ID:"
          :value="resourcesData.backend.pid"
          :bare="true"
        />
      </div>

      <div class="resource-card system-card">
        <h3>💻 System Resources</h3>
        <ResourceMetric
          label="Total Memory:"
          :value="`${resourcesData.system.memory.total_gb} GB`"
          :bare="true"
        />
        <ResourceMetric
          label="Memory Usage:"
          :value="`${resourcesData.system.memory.used_gb} GB`"
          :percent="resourcesData.system.memory.percent"
        />
        <ResourceMetric
          label="Available Memory:"
          :value="`${resourcesData.system.memory.available_gb} GB`"
          :bare="true"
        />
        <ResourceMetric
          label="CPU Usage:"
          :value="`${resourcesData.system.cpu.percent}%`"
          :percent="resourcesData.system.cpu.percent"
        />
        <ResourceMetric
          label="CPU Cores:"
          :value="resourcesData.system.cpu.count"
          :bare="true"
        />
      </div>
    </div>

    <div v-if="resourcesLoading" class="loading-message">
      ⏳ Loading resource data...
    </div>

    <div v-if="resourcesData" class="resources-footer">
      <small>⏰ Last Updated: {{ formatTime(resourcesData.timestamp) }}</small>
    </div>
  </div>
</template>

<script lang="ts">
import { defineComponent } from "vue";
import ResourceMetric from "./ResourceMetric.vue";
import { useSettingsResources } from "./composables/useSettingsResources";

export default defineComponent({
  name: "SettingsResourcesSection",
  components: { ResourceMetric },
  setup() {
    const {
      resourcesData,
      resourcesLoading,
      resourcesError,
      refreshResources,
      formatTime,
    } = useSettingsResources();

    return {
      resourcesData,
      resourcesLoading,
      resourcesError,
      refreshResources,
      formatTime,
    };
  },
});
</script>

<style scoped>
.resources-section {
  display: flex;
  flex-direction: column;
  gap: 24px;
}

.resources-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 16px;
}

.resources-header h2 {
  margin: 0;
  color: #333;
}

.btn {
  padding: 8px 16px;
  border: none;
  border-radius: 6px;
  cursor: pointer;
  font-size: 0.9em;
  font-weight: 500;
  transition: all 0.3s ease;
}

.btn-refresh {
  background: #3b82f6;
  color: white;
}

.btn-refresh:hover:not(:disabled) {
  background: #2563eb;
}

.btn-refresh:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.error-message {
  padding: 16px;
  background: #ffebee;
  color: #c62828;
  border-radius: 8px;
  border-left: 4px solid #c62828;
}

.resources-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
  gap: 20px;
}

.resource-card {
  padding: 20px;
  border-radius: 12px;
  background: white;
  border: 2px solid #e0e0e0;
}

.resource-card h3 {
  margin: 0 0 20px 0;
  color: #333;
  font-size: 1.1em;
}

.backend-card {
  border-color: #3b82f6;
  background: linear-gradient(
    135deg,
    rgba(59, 130, 246, 0.05) 0%,
    rgba(29, 78, 216, 0.05) 100%
  );
}

.system-card {
  border-color: #10b981;
  background: linear-gradient(
    135deg,
    rgba(16, 185, 129, 0.05) 0%,
    rgba(5, 150, 105, 0.05) 100%
  );
}

.loading-message {
  text-align: center;
  padding: 40px;
  color: #94a3b8;
  font-size: 1em;
}

.resources-footer {
  text-align: right;
  color: #94a3b8;
  font-size: 0.85em;
  padding-top: 16px;
  border-top: 1px solid #e0eaf7;
}

@media (max-width: 768px) {
  .resources-grid {
    grid-template-columns: 1fr;
  }
}
</style>
