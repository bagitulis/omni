<template>
  <div class="resources-section">
    <div class="resources-header">
      <h2>📊 Resource Usage Monitoring</h2>
      <button
        @click="refreshResources"
        class="btn btn-refresh"
        :disabled="resourcesLoading"
      >
        <span v-if="!resourcesLoading">🔄 Refresh</span>
        <span v-else>Loading...</span>
      </button>
    </div>

    <div v-if="resourcesError" class="error-message">
      ⚠️ {{ resourcesError }}
    </div>

    <div v-if="resourcesData" class="resources-grid">
      <!-- Backend Resources -->
      <div class="resource-card backend-card">
        <h3>🔧 Backend Server</h3>

        <div class="resource-item">
          <label>Memory Usage:</label>
          <div class="progress-bar">
            <div
              class="progress-fill"
              :style="{ width: resourcesData.backend.memory.percent + '%' }"
              :class="getResourceClass(resourcesData.backend.memory.percent)"
            ></div>
          </div>
          <span class="resource-value">
            {{ resourcesData.backend.memory.rss_mb }} MB /
            {{ resourcesData.backend.memory.percent }}%
          </span>
        </div>

        <div class="resource-item">
          <label>CPU Usage:</label>
          <div class="progress-bar">
            <div
              class="progress-fill"
              :style="{ width: resourcesData.backend.cpu.percent + '%' }"
              :class="getResourceClass(resourcesData.backend.cpu.percent)"
            ></div>
          </div>
          <span class="resource-value"
            >{{ resourcesData.backend.cpu.percent }}%</span
          >
        </div>

        <div class="resource-item">
          <label>Threads:</label>
          <span class="resource-value">{{
            resourcesData.backend.threads
          }}</span>
        </div>

        <div class="resource-item">
          <label>Process ID:</label>
          <span class="resource-value">{{ resourcesData.backend.pid }}</span>
        </div>
      </div>

      <!-- System Resources -->
      <div class="resource-card system-card">
        <h3>💻 System Resources</h3>

        <div class="resource-item">
          <label>Total Memory:</label>
          <span class="resource-value"
            >{{ resourcesData.system.memory.total_gb }} GB</span
          >
        </div>

        <div class="resource-item">
          <label>Memory Usage:</label>
          <div class="progress-bar">
            <div
              class="progress-fill"
              :style="{ width: resourcesData.system.memory.percent + '%' }"
              :class="getResourceClass(resourcesData.system.memory.percent)"
            ></div>
          </div>
          <span class="resource-value">
            {{ resourcesData.system.memory.used_gb }} GB /
            {{ resourcesData.system.memory.percent }}%
          </span>
        </div>

        <div class="resource-item">
          <label>Available Memory:</label>
          <span class="resource-value"
            >{{ resourcesData.system.memory.available_gb }} GB</span
          >
        </div>

        <div class="resource-item">
          <label>CPU Usage:</label>
          <div class="progress-bar">
            <div
              class="progress-fill"
              :style="{ width: resourcesData.system.cpu.percent + '%' }"
              :class="getResourceClass(resourcesData.system.cpu.percent)"
            ></div>
          </div>
          <span class="resource-value"
            >{{ resourcesData.system.cpu.percent }}%</span
          >
        </div>

        <div class="resource-item">
          <label>CPU Cores:</label>
          <span class="resource-value">{{
            resourcesData.system.cpu.count
          }}</span>
        </div>
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

<script setup lang="ts">
import { ref, onMounted, onUnmounted } from "vue";
import { formatTime, getResourceClass } from "@/utils/helpers";
import { getAuthHeaders, getApiBaseUrl } from "@/utils/apiHeaders";

interface ResourcesData {
  backend: {
    memory: { rss_mb: number; percent: number };
    cpu: { percent: number };
    threads: number;
    pid: number;
  };
  system: {
    memory: {
      total_gb: number;
      used_gb: number;
      available_gb: number;
      percent: number;
    };
    cpu: { percent: number; count: number };
  };
  timestamp: string;
}

const resourcesLoading = ref(false);
const resourcesError = ref("");
const resourcesData = ref<ResourcesData | null>(null);

let resourcesInterval: ReturnType<typeof setInterval> | null = null;
let pageIsVisible = !document.hidden;

const refreshResources = async () => {
  resourcesLoading.value = true;
  resourcesError.value = "";
  try {
    const response = await fetch(getApiBaseUrl("/settings/resource-usage"), {
      headers: getAuthHeaders(),
    });
    const result = await response.json();
    if (result.success) {
      resourcesData.value = result.data;
    } else {
      resourcesError.value = result.error || "Failed to load resource data";
    }
  } catch {
    resourcesError.value = "Error connecting to server";
  } finally {
    resourcesLoading.value = false;
  }
};

const startResourcesPolling = () => {
  if (resourcesInterval) clearInterval(resourcesInterval);

  resourcesInterval = setInterval(() => {
    if (pageIsVisible) {
      refreshResources();
    }
  }, 5000);
};

const stopResourcesPolling = () => {
  if (resourcesInterval) {
    clearInterval(resourcesInterval);
    resourcesInterval = null;
  }
};

onMounted(async () => {
  await refreshResources();
  startResourcesPolling();

  document.addEventListener("visibilitychange", () => {
    pageIsVisible = !document.hidden;
  });
});

onUnmounted(() => {
  stopResourcesPolling();
});
</script>

<style src="./ResourceUsageSettings.styles.css" scoped></style>
