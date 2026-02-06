<template>
  <label class="column-checkbox">
    <input
      type="checkbox"
      :checked="isSelected"
      @change="$emit('toggle')"
      class="checkbox-input"
    />
    <span class="column-info">
      <strong>{{ column.spreadsheet_column }}</strong>
      <span>{{ column.name }}</span>
    </span>
  </label>
</template>

<script setup lang="ts">
defineProps<{
  column: any;
  isSelected: boolean;
}>();

defineEmits<{
  toggle: [];
}>();
</script>

<style scoped>
.column-checkbox {
  display: flex;
  align-items: flex-start;
  gap: 8px;
  cursor: pointer;
  padding: 8px;
  border-radius: 4px;
  transition: background 0.2s;
}

.column-checkbox:hover {
  background: #f5f5f5;
}

.column-checkbox .checkbox-input {
  margin-top: 3px;
  cursor: pointer;
}

.column-info {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.column-info strong {
  font-size: 13px;
  color: #333;
}

.column-info span {
  font-size: 12px;
  color: #666;
}
</style>
