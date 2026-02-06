<template>
  <div class="checkbox-cell">
    <input
      type="checkbox"
      :checked="isSelected"
      @change="$emit('toggle')"
      class="row-checkbox"
      :aria-label="`Select row ${rowIndex}`"
    />
  </div>
</template>

<script setup lang="ts">
defineProps<{
  rowIndex: number;
  isSelected: boolean;
}>();

defineEmits<{
  toggle: [];
}>();
</script>

<style scoped>
.checkbox-cell {
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 8px;
  min-width: 40px;
}

.row-checkbox {
  width: 18px;
  height: 18px;
  cursor: pointer;
  accent-color: #3b82f6;
}

.row-checkbox:hover {
  transform: scale(1.1);
}

.row-checkbox:focus {
  outline: 2px solid #3b82f6;
  outline-offset: 2px;
}
</style>
