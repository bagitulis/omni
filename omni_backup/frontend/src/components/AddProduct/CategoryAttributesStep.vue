<template>
  <div class="category-attributes-step">
    <h3 class="step-title">Category Attributes</h3>
    <p class="step-description">
      Fill in the required attributes for your selected category.
    </p>

    <div v-if="loading" class="loading-state">
      <i class="pi pi-spin pi-spinner"></i>
      <span>Loading attributes...</span>
    </div>

    <div v-else-if="attributes.length === 0" class="empty-state">
      <i class="pi pi-check-circle"></i>
      <span>No additional attributes required for this category.</span>
    </div>

    <div v-else class="attributes-form">
      <div
        v-for="attr in attributes"
        :key="attr.id"
        class="attribute-field"
      >
        <label :for="`attr-${attr.id}`" class="field-label">
          {{ attr.name }}
          <span v-if="attr.required" class="required">*</span>
        </label>

        <!-- Dropdown for attributes with predefined values -->
        <select
          v-if="attr.values && attr.values.length > 0"
          :id="`attr-${attr.id}`"
          :value="modelValue[attr.id] || ''"
          class="field-select"
          :required="attr.required"
          @change="updateAttribute(attr.id, ($event.target as HTMLSelectElement).value)"
        >
          <option value="">Select {{ attr.name }}</option>
          <option
            v-for="val in attr.values"
            :key="val.id"
            :value="val.id"
          >
            {{ val.name }}
          </option>
        </select>

        <!-- Text input for free-form attributes -->
        <input
          v-else
          :id="`attr-${attr.id}`"
          type="text"
          :value="modelValue[attr.id] || ''"
          class="field-input"
          :placeholder="`Enter ${attr.name}`"
          :required="attr.required"
          @input="updateAttribute(attr.id, ($event.target as HTMLInputElement).value)"
        />

        <p v-if="attr.description" class="field-hint">
          {{ attr.description }}
        </p>
      </div>
    </div>

    <div v-if="missingRequired.length > 0" class="validation-warning">
      <i class="pi pi-exclamation-triangle"></i>
      Please fill in: {{ missingRequired.join(', ') }}
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from "vue";

interface AttributeValue {
  id: string;
  name: string;
}

interface CategoryAttribute {
  id: string;
  name: string;
  required: boolean;
  description?: string;
  values?: AttributeValue[];
}

const props = defineProps<{
  attributes: CategoryAttribute[];
  modelValue: Record<string, string>;
  loading?: boolean;
}>();

const emit = defineEmits<{
  "update:modelValue": [value: Record<string, string>];
}>();

const missingRequired = computed(() => {
  return props.attributes
    .filter((attr) => attr.required && !props.modelValue[attr.id])
    .map((attr) => attr.name);
});

function updateAttribute(attrId: string, value: string) {
  emit("update:modelValue", {
    ...props.modelValue,
    [attrId]: value,
  });
}
</script>

<style scoped>
.category-attributes-step {
  padding: 8px 0;
}

.step-title {
  font-size: 18px;
  font-weight: 600;
  color: #1f2937;
  margin-bottom: 4px;
}

.step-description {
  font-size: 14px;
  color: #6b7280;
  margin-bottom: 20px;
}

.loading-state,
.empty-state {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 10px;
  padding: 32px;
  background: #f9fafb;
  border-radius: 8px;
  color: #6b7280;
}

.empty-state i {
  font-size: 20px;
  color: #10b981;
}

.attributes-form {
  display: grid;
  gap: 16px;
}

.attribute-field {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.field-label {
  font-weight: 500;
  font-size: 14px;
  color: #374151;
}

.required {
  color: #ef4444;
}

.field-select,
.field-input {
  width: 100%;
  padding: 10px 12px;
  border: 1px solid #d1d5db;
  border-radius: 6px;
  font-size: 14px;
  transition: border-color 0.2s, box-shadow 0.2s;
  background: white;
}

.field-select:focus,
.field-input:focus {
  outline: none;
  border-color: #3b82f6;
  box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.1);
}

.field-hint {
  font-size: 12px;
  color: #9ca3af;
  margin: 0;
}

.validation-warning {
  margin-top: 16px;
  padding: 12px;
  background: #fef3c7;
  border: 1px solid #f59e0b;
  border-radius: 6px;
  color: #92400e;
  font-size: 13px;
  display: flex;
  align-items: center;
  gap: 8px;
}

.validation-warning i {
  color: #f59e0b;
}
</style>
