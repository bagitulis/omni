<template>
  <div class="script-monitor">
    <!-- Monitor Status Header -->
    <div class="monitor-header">
      <h2>🎬 Script Execution Monitor</h2>
      <div class="monitor-controls">
        <button
          @click="refreshMonitor"
          :disabled="isLoading"
          class="btn btn-refresh"
        >
          {{ isLoading ? "Loading..." : "🔄 Refresh" }}
        </button>
        <button
          @click="clearBrowserCache"
          :disabled="isClearingCache"
          class="btn btn-cache-clear"
          title="Clear browser cache and reload page"
        >
          {{ isClearingCache ? "Clearing..." : "🧹 Clear Cache" }}
        </button>
        <span v-if="lastUpdate" class="last-update">
          Last updated: {{ formatTime(lastUpdate) }}
        </span>
      </div>
    </div>

    <!-- Tab Navigation -->
    <div class="tab-navigation">
      <button
        v-for="tab in tabs"
        :key="tab.id"
        :class="['tab-btn', { active: activeTab === tab.id }]"
        @click="activeTab = tab.id"
      >
        <span class="tab-icon">{{ tab.icon }}</span>
        <span class="tab-label">{{ tab.label }}</span>
        <span v-if="tab.badge" class="tab-badge">{{ tab.badge }}</span>
      </button>
    </div>

    <div v-if="error" class="error-banner">⚠️ {{ error }}</div>

    <!-- Tab Components -->
    <CurrentJobTab
      :activeTab="activeTab"
      :currentJob="currentJob"
      :isJobStuck="isJobStuck"
      @cancel-job="cancelJob"
      @force-cancel="forceCancel"
    />

    <QueueTab
      :activeTab="activeTab"
      :pendingQueue="pendingQueue"
      :autoFunctionConfigs="autoFunctionConfigs"
      @force-cancel-job="forceCancel"
      @cancel-scheduled="cancelScheduledExecution"
    />

    <HistoryTab
      :activeTab="activeTab"
      :recentHistory="recentHistory"
      :isClearing="isClearing"
      @clear-history="clearHistory"
    />

    <ConfigTab
      :activeTab="activeTab"
      :autoFunctionConfigs="autoFunctionConfigs"
      :editingConfig="editingConfig"
      :availableFunctions="availableFunctions"
      @show-add-new="showAddNewFunction"
      @show-editor="showConfigEditor"
      @close-editor="editingConfig = null"
      @toggle-function="toggleAutoFunction"
      @delete-function="deleteAutoFunction"
      @save-config="saveConfigEditor"
      @increment-time="incrementTime"
      @decrement-time="decrementTime"
    />

    <!-- Statistics Footer -->
    <div class="monitor-stats">
      <div class="stat-item">
        <span class="stat-label">Total Completed:</span>
        <span class="stat-value">{{ stats.completed || 0 }}</span>
      </div>
      <div class="stat-item">
        <span class="stat-label">Failed:</span>
        <span class="stat-value text-danger">{{ stats.failed || 0 }}</span>
      </div>
      <div class="stat-item">
        <span class="stat-label">Pending:</span>
        <span class="stat-value">{{ stats.pending || 0 }}</span>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, watch, onMounted, computed } from "vue";
import { useRoute, useRouter } from "vue-router";
import { useScriptMonitorLogic } from "@/composables/useScriptMonitorLogic";
import CurrentJobTab from "./CurrentJobTab.vue";
import QueueTab from "./QueueTab.vue";
import HistoryTab from "./HistoryTab.vue";
import ConfigTab from "./ConfigTab.vue";

/**
 * Script Monitor Component
 * Single Responsibility: Manage job/script monitoring tabs
 */

const route = useRoute();
const router = useRouter();
const activeTab = ref<"current" | "queue" | "history" | "config">("current");
const isClearingCache = ref(false);

const {
  isLoading,
  isClearing,
  error,
  lastUpdate,
  currentJob,
  pendingQueue,
  recentHistory,
  autoFunctionConfigs,
  editingConfig,
  isJobStuck,
  stats,
  refreshMonitor,
  cancelJob,
  forceCancel,
  clearHistory,
  toggleAutoFunction,
  showConfigEditor,
  saveConfigEditor,
  showAddNewFunction,
  incrementTime,
  decrementTime,
  deleteAutoFunction,
  formatTime,
  cancelScheduledExecution,
} = useScriptMonitorLogic();

const availableFunctions = [
  { value: "locked_today", label: "🔒 Locked Today" },
  { value: "auto_update_token", label: "🔑 Auto Update Token" },
  { value: "sync_from_sheets", label: "📥 Sync From Sheets" },
];

// Tab navigation data with dynamic badges
const tabs = computed(() => [
  {
    id: "current" as const,
    icon: "🎬",
    label: "Current Job",
    badge: currentJob.value ? 1 : 0,
  },
  {
    id: "queue" as const,
    icon: "📋",
    label: "Queue",
    badge: pendingQueue.value.length,
  },
  { id: "history" as const, icon: "📜", label: "History", badge: 0 },
  {
    id: "config" as const,
    icon: "⚙️",
    label: "Auto-Functions",
    badge: autoFunctionConfigs.value.length,
  },
]);

onMounted(() => {
  const tabFromRoute = route.meta.tab as string;
  if (tabFromRoute) {
    activeTab.value = tabFromRoute as any;
  }
});

watch(
  () => route.meta.tab,
  (newTab) => {
    if (newTab) {
      activeTab.value = newTab as any;
    }
  }
);

watch(activeTab, (newTab) => {
  const tabRouteMap: Record<string, string> = {
    current: "/script-monitor/current",
    queue: "/script-monitor/queue",
    history: "/script-monitor/history",
    config: "/script-monitor/auto-functions",
  };

  if (tabRouteMap[newTab]) {
    router.push(tabRouteMap[newTab]);
  }
});

async function clearBrowserCache() {
  if (
    !confirm(
      "🧹 Clear all cache? This will:\n• Clear localStorage\n• Clear sessionStorage\n• Unregister service workers\n• Hard reload the page"
    )
  ) {
    return;
  }

  isClearingCache.value = true;
  try {
    localStorage.clear();
    sessionStorage.clear();

    if ("serviceWorker" in navigator) {
      try {
        const registrations = await navigator.serviceWorker.getRegistrations();
        for (let registration of registrations) {
          await registration.unregister();
        }
      } catch (e) {
        // Could not unregister service workers
      }
    }

    if ("caches" in window) {
      try {
        const cacheNames = await caches.keys();
        for (let cacheName of cacheNames) {
          await caches.delete(cacheName);
        }
      } catch (e) {
        // Could not clear caches
      }
    }

    setTimeout(() => {
      window.location.href = window.location.href;
    }, 500);
  } catch (error) {
    console.error("❌ Error clearing cache:", error);
    isClearingCache.value = false;
  }
}
</script>

<style src="./ScriptMonitor.styles.css" scoped></style>
