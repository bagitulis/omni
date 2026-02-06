<template>
  <div class="form-container">
    <div class="form-header">
      <h3 class="text-lg font-bold text-gray-800">Extract from Wallet</h3>
      <p class="text-sm text-gray-600 mt-1">
        Select the period to extract shipping data from wallet transactions
      </p>
    </div>

    <div class="form-info">
      <div class="info-icon">ℹ️</div>
      <span class="info-text">
        This will download all shipping fee records for the selected month and
        year.
      </span>
    </div>

    <div class="form-inputs">
      <div class="form-group">
        <label for="shipping-month-select" class="form-label"
          >Select Month</label
        >
        <select
          id="shipping-month-select"
          v-model="params.month"
          class="form-select"
          @change="$emit('update:params', params)"
        >
          <option value="">Choose a month...</option>
          <option
            v-for="month in months"
            :key="month.value"
            :value="month.value"
          >
            {{ month.name }}
          </option>
        </select>
      </div>

      <div class="form-group">
        <label for="shipping-year-select" class="form-label">Select Year</label>
        <select
          id="shipping-year-select"
          v-model="params.year"
          class="form-select"
          @change="$emit('update:params', params)"
        >
          <option value="">Choose a year...</option>
          <option v-for="year in years" :key="year.value" :value="year.value">
            {{ year.name }}
          </option>
        </select>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch } from "vue";

interface Props {
  params?: Record<string, any>;
}

const props = withDefaults(defineProps<Props>(), {
  params: () => ({}),
});

defineEmits<{
  "update:params": [params: Record<string, any>];
}>();

const params = ref({ ...props.params });

const years = computed(() =>
  Array.from({ length: 11 }, (_, i) => ({
    name: (2020 + i).toString(),
    value: 2020 + i,
  }))
);

const months = [
  { name: "January", value: 1 },
  { name: "February", value: 2 },
  { name: "March", value: 3 },
  { name: "April", value: 4 },
  { name: "May", value: 5 },
  { name: "June", value: 6 },
  { name: "July", value: 7 },
  { name: "August", value: 8 },
  { name: "September", value: 9 },
  { name: "October", value: 10 },
  { name: "November", value: 11 },
  { name: "December", value: 12 },
];

watch(
  () => props.params,
  (newParams) => {
    params.value = { ...newParams };
  },
  { deep: true }
);
</script>

<style scoped>
.form-container {
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
}

.form-header {
  padding: 1rem;
  background: linear-gradient(135deg, #f0f4f8 0%, #f8fafc 100%);
  border-radius: 0.75rem;
  border-left: 4px solid #3b82f6;
}

.form-header h3 {
  margin: 0;
  color: #1f2937;
  font-size: 1.125rem;
}

.form-header p {
  margin: 0;
  color: #6b7280;
  font-size: 0.875rem;
}

.form-info {
  display: flex;
  align-items: flex-start;
  gap: 0.75rem;
  padding: 1rem;
  background: #eff6ff;
  border: 1px solid #bfdbfe;
  border-radius: 0.5rem;
}

.info-icon {
  font-size: 1.25rem;
  flex-shrink: 0;
}

.info-text {
  color: #1e40af;
  font-size: 0.875rem;
  line-height: 1.5;
}

.form-inputs {
  display: flex;
  flex-direction: column;
  gap: 1.25rem;
}

.form-group {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.form-label {
  color: #1f2937;
  font-weight: 600;
  font-size: 0.875rem;
}

.form-select {
  padding: 0.75rem;
  border: 2px solid #e5e7eb;
  border-radius: 0.5rem;
  background: white;
  color: #1f2937;
  font-size: 0.875rem;
  cursor: pointer;
  transition: all 0.2s ease;
}

.form-select:hover {
  border-color: #d1d5db;
}

.form-select:focus {
  outline: none;
  border-color: #3b82f6;
  background: #f0f9ff;
}
</style>
