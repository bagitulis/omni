<template>
  <div class="delivery-selector">
    <label class="form-label">
      Delivery Options
      <span v-if="required" class="required">*</span>
    </label>

    <div v-if="loading" class="loading-indicator">
      <i class="pi pi-spin pi-spinner"></i> Loading delivery options...
    </div>

    <div v-else-if="options.length === 0" class="empty-state">
      No delivery options available
    </div>

    <div v-else class="options-grid">
      <label
        v-for="option in options"
        :key="option.id"
        class="option-item"
        :class="{ selected: isSelected(option.id) }"
      >
        <input
          type="checkbox"
          :value="option.id"
          :checked="isSelected(option.id)"
          @change="toggleOption(option.id)"
          class="option-checkbox"
        />
        <span class="option-content">
          <span class="option-name">{{ option.name }}</span>
          <span v-if="option.description" class="option-desc">
            {{ option.description }}
          </span>
        </span>
      </label>
    </div>

    <p v-if="selectedCount > 0" class="selection-count">
      {{ selectedCount }} option{{ selectedCount > 1 ? 's' : '' }} selected
    </p>
  </div>
</template>

<script setup lang="ts">
import { computed } from "vue";

interface DeliveryOption {
  id: string;
  name: string;
  description?: string;
}

const props = defineProps<{
  modelValue: string[];
  options: DeliveryOption[];
  loading?: boolean;
  required?: boolean;
}>();

const emit = defineEmits<{
  "update:modelValue": [value: string[]];
}>();

const selectedCount = computed(() => props.modelValue.length);

function isSelected(id: string): boolean {
  return props.modelValue.includes(id);
}

function toggleOption(id: string) {
  const newValue = isSelected(id)
    ? props.modelValue.filter((v) => v !== id)
    : [...props.modelValue, id];
  emit("update:modelValue", newValue);
}
</script>

<style scoped>
.delivery-selector {
  margin-top: 16px;
}

.form-label {
  display: block;
  font-weight: 600;
  margin-bottom: 10px;
  color: #374151;
}

.required {
  color: #ef4444;
}

.options-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(200px, 1fr));
  gap: 10px;
}

.option-item {
  display: flex;
  align-items: flex-start;
  gap: 10px;
  padding: 12px;
  border: 1px solid #e5e7eb;
  border-radius: 8px;
  cursor: pointer;
  transition: all 0.2s;
  background: #fafafa;
}

.option-item:hover {
  border-color: #3b82f6;
  background: #f0f9ff;
}

.option-item.selected {
  border-color: #3b82f6;
  background: #eff6ff;
}

.option-checkbox {
  margin-top: 2px;
  width: 16px;
  height: 16px;
  cursor: pointer;
}

.option-content {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.option-name {
  font-weight: 500;
  color: #1f2937;
  font-size: 14px;
}

.option-desc {
  font-size: 12px;
  color: #6b7280;
}

.loading-indicator,
.empty-state {
  padding: 16px;
  text-align: center;
  color: #6b7280;
  font-size: 14px;
}

.loading-indicator i {
  margin-right: 8px;
}

.selection-count {
  margin-top: 10px;
  font-size: 13px;
  color: #3b82f6;
  font-weight: 500;
}
</style>
