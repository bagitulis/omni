<template>
  <div class="step-content">
    <div class="form-group">
      <label for="package-weight">Package Weight (grams) *</label>
      <input
        id="package-weight"
        :value="packageWeight"
        type="number"
        placeholder="Enter weight in grams"
        min="1"
        required
        @input="$emit('update:packageWeight', parseInt(($event.target as HTMLInputElement).value) || 0)"
      />
    </div>

    <div class="form-group">
      <label>Package Dimensions (cm)</label>
      <div class="dimensions-row">
        <div class="dimension-field">
          <label for="dimension-length" class="sr-only">Length</label>
          <input
            id="dimension-length"
            :value="dimensions.length"
            type="number"
            placeholder="Length"
            min="0"
            @input="updateDimension('length', parseFloat(($event.target as HTMLInputElement).value) || 0)"
          />
        </div>
        <span class="dimension-separator" aria-hidden="true">×</span>
        <div class="dimension-field">
          <label for="dimension-width" class="sr-only">Width</label>
          <input
            id="dimension-width"
            :value="dimensions.width"
            type="number"
            placeholder="Width"
            min="0"
            @input="updateDimension('width', parseFloat(($event.target as HTMLInputElement).value) || 0)"
          />
        </div>
        <span class="dimension-separator" aria-hidden="true">×</span>
        <div class="dimension-field">
          <label for="dimension-height" class="sr-only">Height</label>
          <input
            id="dimension-height"
            :value="dimensions.height"
            type="number"
            placeholder="Height"
            min="0"
            @input="updateDimension('height', parseFloat(($event.target as HTMLInputElement).value) || 0)"
          />
        </div>
      </div>
    </div>

    <!-- Delivery Options Selector -->
    <DeliverySelector
      :model-value="deliveryOptionIds"
      :options="deliveryOptions"
      :loading="deliveryLoading"
      @update:model-value="$emit('update:deliveryOptionIds', $event)"
    />

    <fieldset class="form-group">
      <legend>Save Mode</legend>
      <div class="save-mode-options">
        <label class="radio-option">
          <input 
            type="radio" 
            :checked="saveMode === 'LISTING'" 
            value="LISTING" 
            name="saveMode"
            @change="$emit('update:saveMode', 'LISTING')" 
          />
          <span>Publish immediately</span>
        </label>
        <label class="radio-option">
          <input 
            type="radio" 
            :checked="saveMode === 'AS_DRAFT'" 
            value="AS_DRAFT" 
            name="saveMode"
            @change="$emit('update:saveMode', 'AS_DRAFT')" 
          />
          <span>Save as draft</span>
        </label>
      </div>
    </fieldset>
  </div>
</template>

<script setup lang="ts">
import DeliverySelector from "./DeliverySelector.vue";

interface Dimensions {
  length: number;
  width: number;
  height: number;
}

interface DeliveryOption {
  id: string;
  name: string;
  description?: string;
}

const props = defineProps<{
  packageWeight: number;
  dimensions: Dimensions;
  saveMode: "AS_DRAFT" | "LISTING";
  deliveryOptionIds: string[];
  deliveryOptions: DeliveryOption[];
  deliveryLoading?: boolean;
}>();

const emit = defineEmits<{
  "update:packageWeight": [value: number];
  "update:dimensions": [value: Dimensions];
  "update:saveMode": [value: "AS_DRAFT" | "LISTING"];
  "update:deliveryOptionIds": [value: string[]];
}>();

function updateDimension(field: keyof Dimensions, value: number) {
  emit("update:dimensions", {
    ...props.dimensions,
    [field]: value,
  });
}
</script>

<style scoped>
@import "./AddProductForm.styles.css";

.dimensions-row {
  display: flex;
  align-items: center;
  gap: 8px;
}

.dimension-field {
  flex: 1;
}

.dimension-separator {
  color: #9ca3af;
  font-weight: 500;
}

.save-mode-options {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.radio-option {
  display: flex;
  align-items: center;
  gap: 8px;
  cursor: pointer;
  font-weight: 400;
}

.radio-option input[type="radio"] {
  width: 18px;
  height: 18px;
  accent-color: #1d4ed8;
}

@media (max-width: 640px) {
  .dimensions-row {
    flex-wrap: wrap;
  }

  .dimension-field {
    min-width: calc(33% - 16px);
  }
}
</style>
