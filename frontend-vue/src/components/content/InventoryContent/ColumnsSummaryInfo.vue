<template>
  <div class="summary-info">
    <div class="info-item">
      <span class="label">Total Columns:</span>
      <span class="value">{{ totalColumns }}</span>
    </div>
    <div class="info-item">
      <span class="label">Aktif:</span>
      <span class="value">{{ visibleCount }}</span>
    </div>
    <div class="info-item">
      <span class="label">Terkunci:</span>
      <span class="value lock-color">{{ lockedCount }} 🔒</span>
    </div>
  </div>
</template>

<script setup lang="ts">
defineProps<{
  totalColumns: number;
  visibleCount: number;
  lockedCount: number;
}>();
</script>

<style scoped>
.summary-info {
  display: flex;
  flex-wrap: wrap;
  gap: 20px;
  background-color: #f0f8ff;
  padding: 12px;
  border-radius: 4px;
  border-left: 3px solid #3498db;
  width: 100%;
  box-sizing: border-box;
}

.info-item {
  display: flex;
  gap: 6px;
  font-size: 12px;
}

.info-item .label {
  color: #666;
  font-weight: 500;
}

.info-item .value {
  color: #333;
  font-weight: 600;
}

.lock-color {
  color: #e74c3c !important;
}
</style>
