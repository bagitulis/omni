<template>
  <div class="resource-item">
    <label>{{ label }}</label>
    <div v-if="!bare" class="progress-bar">
      <div
        class="progress-fill"
        :style="{ width: percent + '%' }"
        :class="getResourceClass(percent)"
      ></div>
    </div>
    <span class="resource-value">{{ value }}</span>
  </div>
</template>

<script lang="ts">
import { defineComponent } from "vue";

export default defineComponent({
  name: "ResourceMetric",
  props: {
    label: { type: String, required: true },
    value: { type: [String, Number], required: true },
    percent: { type: Number, default: 0 },
    bare: { type: Boolean, default: false },
  },
  setup() {
    const getResourceClass = (percent: number) => {
      if (percent < 70) return "healthy";
      if (percent < 85) return "warning";
      return "critical";
    };

    return { getResourceClass };
  },
});
</script>

<style scoped>
.resource-item {
  margin-bottom: 18px;
  padding-bottom: 18px;
  border-bottom: 1px solid #e0eaf7;
}

.resource-item:last-child {
  margin-bottom: 0;
  padding-bottom: 0;
  border-bottom: none;
}

label {
  display: block;
  margin-bottom: 8px;
  font-weight: 600;
  color: #475569;
  font-size: 0.9em;
}

.progress-bar {
  height: 20px;
  background: #eef4fb;
  border-radius: 10px;
  overflow: hidden;
  margin-bottom: 6px;
  border: 1px solid #dce9f5;
}

.progress-fill {
  height: 100%;
  transition:
    width 0.5s ease,
    background 0.3s ease;
  display: flex;
  align-items: center;
  justify-content: center;
  color: white;
  font-size: 0.75em;
  font-weight: bold;
}

.progress-fill.healthy {
  background: linear-gradient(90deg, #10b981, #059669);
}

.progress-fill.warning {
  background: linear-gradient(90deg, #f59e0b, #d97706);
}

.progress-fill.critical {
  background: linear-gradient(90deg, #ef4444, #dc2626);
}

.resource-value {
  display: block;
  color: #64748b;
  font-size: 0.9em;
  font-weight: 500;
  font-family: monospace;
}
</style>
