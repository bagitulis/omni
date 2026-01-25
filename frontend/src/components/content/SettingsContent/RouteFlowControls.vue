<template>
  <div class="route-flow-card">
    <div class="card-header">
      <div>
        <h4>🎛️ Route Flow Controls</h4>
        <p class="subtitle">
          Routes marked as <strong>direct</strong> bypass the navigation queue but are still tracked in Script Monitor.
        </p>
      </div>
      <div class="header-actions">
        <button class="btn ghost" type="button" @click="resetDefaults">
          Reset Default
        </button>
      </div>
    </div>

    <div class="chip-row" v-if="directRoutes.length">
      <span class="chip" v-for="route in directRoutes" :key="route">
        {{ route }}
        <button class="chip-close" type="button" @click="removeRoute(route)" aria-label="Remove route">✕</button>
      </span>
    </div>
    <div v-else class="empty">No direct routes configured.</div>

    <form class="add-form" @submit.prevent="addRoute">
      <input
        v-model="newRoute"
        type="text"
        placeholder="e.g. OrderManagerShopee or /order-manager"
        class="input"
      />
      <button class="btn primary" type="submit">Add</button>
    </form>

    <div class="preset-row">
      <span class="preset-label">Quick presets:</span>
      <button
        v-for="preset in defaultDirectRoutes"
        :key="preset"
        type="button"
        class="btn tag"
        :class="{ active: isActivePreset(preset) }"
        @click="togglePreset(preset)"
      >
        {{ preset }}
      </button>
    </div>

    <div class="flow-log">
      <div class="flow-log__header">
        <h5>Recent Route Flows</h5>
        <span class="hint">Last 10 navigations</span>
      </div>
      <div v-if="recentFlows.length" class="flow-list">
        <div class="flow-item" v-for="flow in recentFlows" :key="flow.id">
          <div class="flow-item__main">
            <span class="flow-name">{{ flow.routeName || flow.path }}</span>
            <span class="flow-tag" :class="flow.flow">{{ flow.flow }}</span>
            <span class="flow-status" :class="flow.status">{{ flow.status }}</span>
          </div>
          <div class="flow-item__meta">
            <span>{{ formatDuration(flow.durationMs) }}</span>
            <span>{{ formatTimestamp(flow.startedAt) }}</span>
          </div>
        </div>
      </div>
      <div v-else class="empty">No route activity recorded yet.</div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from "vue";
import { useRouteFlowController } from "@/composables/useRouteFlowController";

const flowController = useRouteFlowController();
const newRoute = ref("");

const directRoutes = computed(() => flowController.directRoutes.value);
const recentFlows = computed(() => flowController.recentFlows.value);
const defaultDirectRoutes = computed(() => flowController.defaultDirectRoutes);

const addRoute = () => {
  const value = newRoute.value.trim();
  if (!value) return;
  flowController.addDirectRoute(value);
  newRoute.value = "";
};

const removeRoute = (route: string) => {
  flowController.removeDirectRoute(route);
};

const togglePreset = (preset: string) => {
  if (isActivePreset(preset)) {
    removeRoute(preset);
  } else {
    flowController.addDirectRoute(preset);
  }
};

const isActivePreset = (preset: string) => {
  return directRoutes.value.some(
    (route) => route.toLowerCase() === preset.toLowerCase()
  );
};

const resetDefaults = () => {
  flowController.resetDirectRoutes();
};

const formatDuration = (duration?: number) => {
  if (!duration && duration !== 0) return "-";
  if (duration < 1000) return `${Math.round(duration)} ms`;
  return `${(duration / 1000).toFixed(1)} s`;
};

const formatTimestamp = (timestamp: number) => {
  const date = new Date(timestamp);
  return date.toLocaleTimeString();
};
</script>

<style scoped>
.route-flow-card {
  display: flex;
  flex-direction: column;
  gap: 12px;
  padding: 16px;
  background: #ffffff;
  border: 1px solid #e5e7eb;
  border-radius: 10px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: 12px;
}

.subtitle {
  margin: 4px 0 0 0;
  color: #6b7280;
  font-size: 13px;
}

.header-actions {
  display: flex;
  gap: 8px;
}

.btn {
  border: none;
  border-radius: 6px;
  padding: 8px 12px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s ease;
}

.btn.primary {
  background: #2563eb;
  color: white;
}

.btn.ghost {
  background: #f3f4f6;
  color: #1f2937;
}

.btn.tag {
  background: #eef2ff;
  color: #4338ca;
}

.btn.tag.active {
  background: #4338ca;
  color: #ffffff;
}

.input {
  flex: 1;
  min-width: 220px;
  padding: 10px 12px;
  border: 1px solid #d1d5db;
  border-radius: 8px;
  font-size: 14px;
}

.add-form {
  display: flex;
  gap: 10px;
  align-items: center;
}

.chip-row {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.chip {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 6px 10px;
  background: #f3f4f6;
  border-radius: 999px;
  font-size: 13px;
}

.chip-close {
  background: none;
  border: none;
  cursor: pointer;
  font-size: 12px;
}

.empty {
  color: #9ca3af;
  font-size: 13px;
}

.preset-row {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  align-items: center;
}

.preset-label {
  font-weight: 600;
  color: #374151;
}

.flow-log {
  border-top: 1px solid #e5e7eb;
  padding-top: 8px;
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.flow-log__header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.flow-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.flow-item {
  display: flex;
  flex-direction: column;
  gap: 4px;
  padding: 10px 12px;
  background: #f9fafb;
  border: 1px solid #e5e7eb;
  border-radius: 8px;
}

.flow-item__main {
  display: flex;
  gap: 8px;
  align-items: center;
  flex-wrap: wrap;
}

.flow-name {
  font-weight: 700;
  color: #111827;
}

.flow-tag {
  padding: 4px 8px;
  border-radius: 999px;
  font-size: 12px;
  font-weight: 700;
  text-transform: uppercase;
}

.flow-tag.queued {
  background: #dbeafe;
  color: #1d4ed8;
}

.flow-tag.direct {
  background: #dcfce7;
  color: #15803d;
}

.flow-status {
  font-size: 12px;
  font-weight: 700;
  text-transform: uppercase;
}

.flow-status.success {
  color: #15803d;
}

.flow-status.pending {
  color: #f59e0b;
}

.flow-status.error {
  color: #b91c1c;
}

.flow-item__meta {
  display: flex;
  gap: 10px;
  color: #6b7280;
  font-size: 12px;
}

.hint {
  color: #6b7280;
  font-size: 12px;
}
</style>
