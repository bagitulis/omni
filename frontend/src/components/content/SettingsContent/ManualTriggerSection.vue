<template>
  <div class="manual-trigger-section">
    <!-- Section Header -->
    <div class="section-header">
      <div class="section-title">
        <Icon name="settings" size="md" />
        <h3>Manual Trigger Mode</h3>
        <span class="section-subtitle">Ketika user klik tombol di UI</span>
      </div>
      <button @click="$emit('show-add-new')" class="btn btn-add">
        <Icon name="plus" size="xs" /> Add Route
      </button>
    </div>

    <!-- Empty State -->
    <div v-if="configs.length === 0" class="empty-state">
      <p>No manual trigger routes configured</p>
      <button @click="$emit('show-add-new')" class="btn btn-primary">
        <Icon name="plus" size="xs" /> Add Route
      </button>
    </div>

    <!-- Routes Table -->
    <div v-else class="routes-table-container">
      <table class="routes-table">
        <thead>
          <tr>
            <th class="col-icon"></th>
            <th class="col-route">Route</th>
            <th class="col-mode">Mode</th>
            <th class="col-priority">Priority</th>
            <th class="col-actions">Actions</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="config in configs" :key="config.route_key">
            <td class="col-icon">
              <Icon v-if="config.icon" :name="config.icon" size="sm" /><span
                v-else
                ><Icon name="document" size="sm"
              /></span>
            </td>
            <td class="col-route">
              <div class="route-info">
                <span class="route-name">{{ config.route_name }}</span>
                <span class="route-desc">{{ config.description }}</span>
              </div>
            </td>
            <td class="col-mode">
              <select
                :value="config.execution_mode"
                @change="handleModeChange(config.route_key, $event)"
                class="mode-select"
                :class="'mode-' + config.execution_mode"
              >
                <option value="queue">Queue</option>
                <option value="direct">Direct</option>
              </select>
            </td>
            <td class="col-priority">
              <select
                v-if="config.execution_mode === 'queue'"
                :value="config.priority"
                @change="handlePriorityChange(config.route_key, $event)"
                class="priority-select"
                :class="'priority-' + config.priority"
              >
                <option value="high">High</option>
                <option value="normal">Normal</option>
                <option value="low">Low</option>
              </select>
              <span v-else class="priority-na">-</span>
            </td>
            <td class="col-actions">
              <button
                @click="$emit('edit', config)"
                class="btn-icon"
                title="Edit"
              >
                <Icon name="edit" size="sm" />
              </button>
              <button
                @click="$emit('delete', config.route_key)"
                class="btn-icon btn-danger"
                title="Delete"
              >
                <Icon name="close" size="sm" />
              </button>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <!-- Legend -->
    <div class="legend">
      <span class="legend-item">
        <span class="mode-badge mode-queue">Queue</span>
        Masuk antrian, termonitor
      </span>
      <span class="legend-item">
        <span class="mode-badge mode-direct">Direct</span>
        Langsung execute
      </span>
    </div>
  </div>
</template>

<script setup lang="ts">
import Icon from "@/components/ui/Icon.vue";
import type { RouteExecutionConfig } from "@/types/routeExecutionConfig";

defineProps<{
  configs: RouteExecutionConfig[];
}>();

const emit = defineEmits<{
  "show-add-new": [];
  edit: [config: RouteExecutionConfig];
  delete: [routeKey: string];
  "update-mode": [routeKey: string, mode: string];
  "update-priority": [routeKey: string, priority: string];
}>();

function handleModeChange(routeKey: string, event: Event) {
  const select = event.target as HTMLSelectElement;
  emit("update-mode", routeKey, select.value);
}

function handlePriorityChange(routeKey: string, event: Event) {
  const select = event.target as HTMLSelectElement;
  emit("update-priority", routeKey, select.value);
}
</script>

<style src="./ManualTriggerSection.styles.css" scoped></style>
