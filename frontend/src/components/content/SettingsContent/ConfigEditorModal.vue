<template>
  <div v-if="config" class="config-editor">
    <div class="editor-overlay" @click="$emit('close-editor')"></div>
    <div class="editor-modal">
      <h3>
        {{ config.id ? "Configure " + config.name : "Add New Auto-Function" }}
      </h3>

      <!-- Function Name Select -->
      <div v-if="!config.id" class="form-group">
        <label>Function Name:</label>
        <select v-model="config.name" class="form-input form-select">
          <option value="">-- Select a function --</option>
          <option
            v-for="func in availableFunctions"
            :key="func.value"
            :value="func.value"
          >
            {{ func.label }}
          </option>
        </select>
      </div>

      <!-- Interval Input -->
      <div class="form-group">
        <label>Interval (minutes):</label>
        <div class="input-group-number">
          <button
            @click="decrementInterval"
            class="btn-spin"
            type="button"
            aria-label="Decrease interval"
          >
            −
          </button>
          <input
            v-model.number="config.interval_minutes"
            type="number"
            min="1"
            class="form-input"
          />
          <button
            @click="incrementInterval"
            class="btn-spin"
            type="button"
            aria-label="Increase interval"
          >
            +
          </button>
        </div>
      </div>

      <!-- Time Window -->
      <div class="time-window-row">
        <div class="form-group flex-1">
          <label>Start Time (HH:mm) - Optional:</label>
          <div class="time-picker">
            <input
              v-model="config.start_time"
              type="time"
              class="form-input time-input"
            />
            <div class="time-controls">
              <button
                @click="$emit('increment-time', 'start_time', 15)"
                class="btn-time-control"
                title="Add 15 min"
                type="button"
                aria-label="Increase start time by 15 minutes"
              >
                ↑
              </button>
              <button
                @click="$emit('decrement-time', 'start_time', 15)"
                class="btn-time-control"
                title="Sub 15 min"
                type="button"
                aria-label="Decrease start time by 15 minutes"
              >
                ↓
              </button>
            </div>
          </div>
        </div>

        <div class="form-group flex-1">
          <label>End Time (HH:mm) - Optional:</label>
          <div class="time-picker">
            <input
              v-model="config.end_time"
              type="time"
              class="form-input time-input"
            />
            <div class="time-controls">
              <button
                @click="$emit('increment-time', 'end_time', 15)"
                class="btn-time-control"
                title="Add 15 min"
                type="button"
                aria-label="Increase end time by 15 minutes"
              >
                ↑
              </button>
              <button
                @click="$emit('decrement-time', 'end_time', 15)"
                class="btn-time-control"
                title="Sub 15 min"
                type="button"
                aria-label="Decrease end time by 15 minutes"
              >
                ↓
              </button>
            </div>
          </div>
        </div>
      </div>

      <!-- Enable Checkbox -->
      <div class="form-group">
        <label>
          <input type="checkbox" v-model="config.enabled" class="checkbox" />
          <span class="ml-2">Enable this function</span>
        </label>
      </div>

      <!-- Actions -->
      <div class="editor-actions">
        <button @click="$emit('save-config')" class="btn btn-primary">
          {{ config.id ? "Save" : "Create" }}
        </button>
        <button @click="$emit('close-editor')" class="btn btn-secondary">
          Cancel
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
/**
 * Config Editor Modal
 * Single Responsibility: Handle configuration form only
 * JSON uses snake_case as per AGENTS.md standard
 */
import { computed } from "vue";
import { type AutoFunctionConfig } from "./configTableUtils";

const props = defineProps<{
  config: AutoFunctionConfig | null;
  availableFunctions: Array<{ value: string; label: string }>;
}>();

defineEmits<{
  "close-editor": [];
  "save-config": [];
  "increment-time": [field: "start_time" | "end_time", minutes: number];
  "decrement-time": [field: "start_time" | "end_time", minutes: number];
}>();

const config = computed(() => props.config);

const incrementInterval = () => {
  if (config.value) {
    config.value.interval_minutes++;
  }
};

const decrementInterval = () => {
  if (config.value) {
    config.value.interval_minutes = Math.max(
      1,
      config.value.interval_minutes - 1,
    );
  }
};
</script>

<style scoped lang="css">
@import "./ScriptMonitor.styles.css";
</style>
