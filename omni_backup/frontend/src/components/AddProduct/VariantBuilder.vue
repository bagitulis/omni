<template>
  <div class="variant-builder">
    <div class="variant-header">
      <h4 class="variant-title">Product Variants</h4>
      <label class="toggle-label">
        <input
          type="checkbox"
          :checked="hasVariants"
          @change="$emit('update:hasVariants', ($event.target as HTMLInputElement).checked)"
        />
        <span>Enable variants (color, size, etc.)</span>
      </label>
    </div>

    <div v-if="hasVariants" class="variant-types">
      <div
        v-for="(variantType, typeIndex) in variantTypes"
        :key="typeIndex"
        class="variant-type-card"
      >
        <div class="variant-type-header">
          <label :for="`variant-type-${typeIndex}`">Variant Type {{ typeIndex + 1 }}</label>
          <button
            v-if="variantTypes.length > 1"
            type="button"
            class="btn btn-sm btn-danger"
            aria-label="Remove variant type"
            @click="removeVariantType(typeIndex)"
          >
            <i class="pi pi-trash" aria-hidden="true"></i>
          </button>
        </div>

        <select
          :id="`variant-type-${typeIndex}`"
          :value="variantType.name"
          class="variant-type-select"
          @change="updateVariantTypeName(typeIndex, ($event.target as HTMLSelectElement).value)"
        >
          <option value="">Select type...</option>
          <option value="Color">Color</option>
          <option value="Size">Size</option>
          <option value="Material">Material</option>
          <option value="Style">Style</option>
          <option value="Pattern">Pattern</option>
        </select>

        <div class="variant-values">
          <div
            v-for="(value, valueIndex) in variantType.values"
            :key="valueIndex"
            class="variant-value-chip"
          >
            <span>{{ value }}</span>
            <button
              type="button"
              class="chip-remove"
              :aria-label="`Remove ${value}`"
              @click="removeVariantValue(typeIndex, valueIndex)"
            >
              <i class="pi pi-times" aria-hidden="true"></i>
            </button>
          </div>
          <input
            type="text"
            class="variant-value-input"
            placeholder="Add value..."
            @keydown.enter.prevent="addVariantValue(typeIndex, $event)"
          />
        </div>
      </div>

      <button
        v-if="variantTypes.length < 3"
        type="button"
        class="btn btn-secondary btn-add-type"
        @click="addVariantType"
      >
        <i class="pi pi-plus" aria-hidden="true"></i> Add Variant Type
      </button>
    </div>

    <div v-if="hasVariants && combinations.length > 0" class="variant-preview">
      <p class="preview-info">
        <i class="pi pi-info-circle" aria-hidden="true"></i>
        {{ combinations.length }} variant combinations will be created
      </p>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from "vue";

interface VariantType {
  name: string;
  values: string[];
}

const props = defineProps<{
  hasVariants: boolean;
  variantTypes: VariantType[];
}>();

const emit = defineEmits<{
  "update:hasVariants": [value: boolean];
  "update:variantTypes": [value: VariantType[]];
}>();

const combinations = computed(() => {
  if (!props.hasVariants || props.variantTypes.length === 0) return [];
  
  const validTypes = props.variantTypes.filter(t => t.name && t.values.length > 0);
  if (validTypes.length === 0) return [];

  return validTypes.reduce<string[][]>(
    (acc, type) => {
      if (acc.length === 0) return type.values.map(v => [v]);
      return acc.flatMap(combo => type.values.map(v => [...combo, v]));
    },
    []
  );
});

function addVariantType() {
  if (props.variantTypes.length >= 3) return;
  emit("update:variantTypes", [...props.variantTypes, { name: "", values: [] }]);
}

function removeVariantType(index: number) {
  const updated = props.variantTypes.filter((_, i) => i !== index);
  emit("update:variantTypes", updated);
}

function updateVariantTypeName(index: number, name: string) {
  const updated = props.variantTypes.map((t, i) => 
    i === index ? { ...t, name } : t
  );
  emit("update:variantTypes", updated);
}

function addVariantValue(typeIndex: number, event: KeyboardEvent) {
  const input = event.target as HTMLInputElement;
  const value = input.value.trim();
  if (!value) return;

  const type = props.variantTypes[typeIndex];
  if (type.values.includes(value)) return;

  const updated = props.variantTypes.map((t, i) =>
    i === typeIndex ? { ...t, values: [...t.values, value] } : t
  );
  emit("update:variantTypes", updated);
  input.value = "";
}

function removeVariantValue(typeIndex: number, valueIndex: number) {
  const updated = props.variantTypes.map((t, i) =>
    i === typeIndex
      ? { ...t, values: t.values.filter((_, vi) => vi !== valueIndex) }
      : t
  );
  emit("update:variantTypes", updated);
}
</script>

<style scoped>
.variant-builder {
  margin-top: 16px;
}

.variant-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 16px;
}

.variant-title {
  margin: 0;
  font-size: 16px;
  font-weight: 600;
  color: #1f2937;
}

.toggle-label {
  display: flex;
  align-items: center;
  gap: 8px;
  cursor: pointer;
  font-size: 14px;
}

.toggle-label input {
  width: 18px;
  height: 18px;
  accent-color: #3b82f6;
}

.variant-types {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.variant-type-card {
  padding: 16px;
  border: 1px solid #e5e7eb;
  border-radius: 8px;
  background: #fafafa;
}

.variant-type-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 12px;
}

.variant-type-header label {
  font-weight: 600;
  font-size: 14px;
  color: #374151;
}

.variant-type-select {
  width: 100%;
  padding: 10px 12px;
  border: 1px solid #d1d5db;
  border-radius: 6px;
  font-size: 14px;
  margin-bottom: 12px;
}

.variant-type-select:focus {
  outline: none;
  border-color: #3b82f6;
  box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.1);
}

.variant-values {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  align-items: center;
}

.variant-value-chip {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 6px 10px;
  background: #3b82f6;
  color: white;
  border-radius: 16px;
  font-size: 13px;
}

.chip-remove {
  background: none;
  border: none;
  color: white;
  cursor: pointer;
  padding: 0;
  display: flex;
  opacity: 0.8;
}

.chip-remove:hover {
  opacity: 1;
}

.variant-value-input {
  flex: 1;
  min-width: 120px;
  padding: 6px 10px;
  border: 1px dashed #d1d5db;
  border-radius: 6px;
  font-size: 13px;
}

.variant-value-input:focus {
  outline: none;
  border-color: #3b82f6;
  border-style: solid;
}

.btn-add-type {
  align-self: flex-start;
}

.variant-preview {
  margin-top: 16px;
  padding: 12px;
  background: #eff6ff;
  border-radius: 6px;
}

.preview-info {
  margin: 0;
  font-size: 13px;
  color: #1d4ed8;
  display: flex;
  align-items: center;
  gap: 8px;
}

.btn { padding: 10px 16px; border: none; border-radius: 6px; cursor: pointer; font-weight: 600; font-size: 14px; display: inline-flex; align-items: center; gap: 8px; }
.btn-sm { padding: 6px 12px; font-size: 12px; }
.btn-secondary { background: #e5e7eb; color: #374151; }
.btn-secondary:hover { background: #d1d5db; }
.btn-danger { background: #ef4444; color: white; }
.btn-danger:hover { background: #dc2626; }
</style>
