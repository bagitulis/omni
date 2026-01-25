<template>
  <div class="checkbox-item">
    <label class="checkbox-label">
      <input
        type="checkbox"
        :checked="isSelected"
        @change="$emit('toggle')"
        class="item-checkbox"
      />
      <span class="column-text">{{ column.name }}</span>
    </label>
  </div>
</template>

<script setup lang="ts">
defineProps<{
  column: { name: string; [key: string]: any };
  isSelected: boolean;
}>();

defineEmits<{
  toggle: [];
}>();
</script>

<style scoped>
.checkbox-item {
  display: flex;
  align-items: center;
}

.checkbox-label {
  display: flex;
  align-items: center;
  gap: 8px;
  cursor: pointer;
  padding: 8px;
  border-radius: 4px;
  transition: background-color 0.2s;
  user-select: none;
  flex: 1;
}

.checkbox-label:hover {
  background-color: #f5f5f5;
}

.item-checkbox {
  width: 18px;
  height: 18px;
  cursor: pointer;
  accent-color: #3498db;
}

.column-text {
  font-size: 13px;
  color: #333;
  font-weight: 500;
}
</style>
