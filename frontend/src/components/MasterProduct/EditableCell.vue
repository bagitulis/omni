<template>
  <td
    class="editable-cell text-left"
    :class="{
      editing: isEditing,
      saving: isSaving,
    }"
    @click="$emit('start-edit')"
  >
    <template v-if="isEditing">
      <input
        type="number"
        ref="inputRef"
        :value="editingValue"
        @input="onInput"
        class="inline-edit-input"
        @keydown.enter="$emit('save-edit')"
        @keydown.escape="$emit('cancel-edit')"
        @blur="$emit('save-edit')"
        :min="min"
        :step="step"
      />
    </template>
    <template v-else>
      <span class="cell-value">{{ formattedValue }}</span>
      <span class="edit-hint" v-if="!isSaving">✏️</span>
      <span class="saving-indicator" v-if="isSaving">💾</span>
    </template>
  </td>
</template>

<script setup lang="ts">
import { computed, nextTick, ref, watch } from "vue";

const props = defineProps<{
  value: number;
  editingValue: number;
  isEditing: boolean;
  isSaving: boolean;
  min?: number;
  step?: number;
  formatFn?: (val: number) => string;
}>();

const emit = defineEmits<{
  "start-edit": [];
  "save-edit": [];
  "cancel-edit": [];
  "update:editingValue": [value: number];
}>();

const formattedValue = computed(() => {
  return props.formatFn ? props.formatFn(props.value) : props.value;
});

const onInput = (event: Event) => {
  emit("update:editingValue", (event.target as HTMLInputElement).valueAsNumber);
};

const inputRef = ref<HTMLInputElement | null>(null);

watch(
  () => props.isEditing,
  (newVal) => {
    if (newVal) {
      nextTick(() => {
        inputRef.value?.focus();
      });
    }
  },
  { flush: "post" },
);
</script>

<style scoped>
.text-left {
  text-align: left;
}

.editable-cell {
  cursor: pointer;
  position: relative;
  transition: all 0.2s ease;
  min-width: 100px;
}

.editable-cell:hover:not(.editing):not(.saving) {
  background: linear-gradient(135deg, #f0fdf4 0%, #dcfce7 100%) !important;
  border-radius: 0.25rem;
}

.editable-cell .cell-value {
  display: inline-block;
}

.editable-cell .edit-hint {
  opacity: 0;
  margin-left: 0.5rem;
  font-size: 0.75rem;
  transition: opacity 0.2s ease;
}

.editable-cell:hover .edit-hint {
  opacity: 0.7;
}

.editable-cell.editing {
  padding: 0.5rem !important;
  background: #fffbeb !important;
}

.editable-cell.saving {
  opacity: 0.7;
  pointer-events: none;
}

.saving-indicator {
  margin-left: 0.5rem;
  animation: pulse 1s ease-in-out infinite;
}

.inline-edit-input {
  width: 100%;
  max-width: 120px;
  padding: 0.375rem 0.5rem;
  font-size: 0.875rem;
  font-weight: 600;
  font-variant-numeric: tabular-nums;
  border: 2px solid #3b82f6;
  border-radius: 0.375rem;
  background: white;
  color: inherit;
  outline: none;
  box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.1);
  transition: all 0.2s ease;
}

.inline-edit-input:focus {
  border-color: #2563eb;
  box-shadow: 0 0 0 4px rgba(59, 130, 246, 0.15);
}

.inline-edit-input::-webkit-outer-spin-button,
.inline-edit-input::-webkit-inner-spin-button {
  -webkit-appearance: none;
  margin: 0;
}

.inline-edit-input[type="number"] {
  -moz-appearance: textfield;
}

@keyframes pulse {
  0%,
  100% {
    opacity: 1;
  }
  50% {
    opacity: 0.5;
  }
}
</style>
