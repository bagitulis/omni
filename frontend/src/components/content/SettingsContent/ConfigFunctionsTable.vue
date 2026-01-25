<template>
  <div class="config-table-container">
    <table class="config-table" aria-label="Auto-function configurations">
      <thead>
        <tr>
          <th scope="col">Function Name</th>
          <th scope="col">Interval</th>
          <th scope="col">Time Window</th>
          <th scope="col">Status</th>
          <th scope="col">Last Executed</th>
          <th scope="col">Next Trigger</th>
          <th scope="col">Actions</th>
        </tr>
      </thead>
      <tbody>
        <tr
          v-for="config in configs"
          :key="config.id"
          :class="{ disabled: !config.enabled }"
        >
          <td class="function-name">{{ config.name }}</td>
          <td class="interval">{{ config.intervalMinutes }}m</td>
          <td class="time-window">
            <span v-if="config.startTime"
              >{{ config.startTime }} - {{ config.endTime }}</span
            >
            <span v-else class="text-muted">-</span>
          </td>
          <td class="status">
            <label class="toggle-switch">
              <input
                type="checkbox"
                :checked="config.enabled"
                @change="$emit('toggle-function', config.name, !config.enabled)"
              />
              <span class="toggle-slider"></span>
            </label>
          </td>
          <td class="last-executed">
            <span v-if="config.lastExecuted" class="time-value">
              {{ formatDateTime(config.lastExecuted) }}
            </span>
            <span v-else class="text-muted">Never</span>
          </td>
          <td
            class="next-trigger"
            :class="{ overdue: isScheduleOverdue(config) }"
          >
            <div v-if="config.nextScheduledExecution">
              <div class="time-value" :class="getNextTriggerClass(config)">
                {{ formatDateTime(config.nextScheduledExecution) }}
              </div>
              <div class="countdown-value">
                {{ getTimeUntilTrigger(config) }}
              </div>
            </div>
            <span v-else class="text-muted">-</span>
          </td>
          <td class="actions">
            <button
              @click="$emit('show-editor', config)"
              class="btn btn-secondary btn-xs"
              title="Edit configuration"
            >
              ✎ Edit
            </button>
            <button
              @click="$emit('delete-function', config.name)"
              class="btn btn-danger btn-xs"
              title="Delete this auto-function"
            >
              ✕ Delete
            </button>
          </td>
        </tr>
      </tbody>
    </table>
  </div>
</template>

<script setup lang="ts">
/**
 * Config Functions Table
 * Single Responsibility: Display configuration table only
 */
import {
  formatDateTime,
  getTimeUntilTrigger,
  getNextTriggerClass,
  isScheduleOverdue,
  type AutoFunctionConfig,
} from "./configTableUtils";

defineProps<{
  configs: AutoFunctionConfig[];
}>();

defineEmits<{
  "toggle-function": [name: string, enabled: boolean];
  "show-editor": [config: AutoFunctionConfig];
  "delete-function": [name: string];
}>();
</script>

<style scoped lang="css">
@import "./ScriptMonitor.styles.css";

.config-table-container {
  overflow-x: auto;
}

.config-table {
  width: 100%;
  border-collapse: collapse;
  background: #ffffff;
  border-radius: 10px;
  overflow: hidden;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.05);
}

.config-table thead {
  background: #f8f9fa;
  border-bottom: 2px solid #e0e0e0;
}

.config-table th {
  padding: 12px 16px;
  text-align: left;
  font-size: 13px;
  font-weight: 700;
  color: #1a1a1a;
  text-transform: uppercase;
  letter-spacing: 0.3px;
}

.config-table tbody tr {
  border-bottom: 1px solid #f0f1f3;
  transition: all 0.2s ease;
}

.config-table tbody tr:hover {
  background: #f8f9fa;
}

.config-table tbody tr.disabled {
  opacity: 0.6;
}

.config-table td {
  padding: 12px 16px;
  font-size: 13px;
  color: #1a1a1a;
}

.function-name {
  font-weight: 600;
  color: #0066cc;
  font-family: "Courier New", monospace;
}

.toggle-switch {
  position: relative;
  display: inline-flex;
  cursor: pointer;
}

.toggle-switch input {
  display: none;
}

.toggle-slider {
  display: inline-block;
  width: 40px;
  height: 24px;
  background: #ccc;
  border-radius: 12px;
  transition: all 0.3s ease;
}

.toggle-slider::after {
  content: "";
  position: absolute;
  width: 20px;
  height: 20px;
  background: white;
  border-radius: 50%;
  top: 2px;
  left: 2px;
  transition: all 0.3s ease;
}

.toggle-switch input:checked + .toggle-slider {
  background: #27ae60;
}

.toggle-switch input:checked + .toggle-slider::after {
  left: 18px;
}

.time-value {
  font-weight: 600;
  color: #1a1a1a;
}

.trigger-overdue .time-value {
  color: #f39c12;
  animation: pulse 1s ease-in-out infinite;
}

.trigger-soon .time-value {
  color: #e74c3c;
}

.countdown-value {
  font-size: 11px;
  color: #6b7280;
}

.actions {
  display: flex;
  gap: 6px;
  white-space: nowrap;
}

.btn {
  padding: 6px 12px;
  border: none;
  border-radius: 6px;
  font-size: 12px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s ease;
}

.btn-secondary {
  background: #f8f9fa;
  color: #1a1a1a;
  border: 1px solid #e0e0e0;
}

.btn-secondary:hover {
  background: #f0f1f3;
  border-color: #0066cc;
}

.btn-danger {
  background: #e74c3c;
  color: white;
}

.btn-danger:hover {
  background: #c0392b;
}

.btn-xs {
  padding: 4px 8px;
  font-size: 11px;
}

.text-muted {
  color: #6b7280;
}

@media (max-width: 768px) {
  .config-table {
    font-size: 12px;
  }

  .config-table th,
  .config-table td {
    padding: 10px 12px;
  }

  .actions {
    flex-direction: column;
  }
}

@keyframes pulse {
  0%,
  100% {
    opacity: 1;
  }
  50% {
    opacity: 0.7;
  }
}
</style>
