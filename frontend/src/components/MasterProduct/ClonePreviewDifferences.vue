<template>
  <div v-if="differences?.length" class="differences-section">
    <h4>Perbedaan</h4>
    <div class="differences-list">
      <div v-for="diff in differences" :key="diff.field" class="diff-item">
        <span class="diff-field">{{ formatFieldName(diff.field) }}</span>
        <span class="diff-source">{{ diff.source_value }}</span>
        <span class="diff-arrow">&rarr;</span>
        <span class="diff-target">{{ diff.target_value }}</span>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import type { Difference } from "./ClonePreview.types";

defineProps<{
  differences?: Difference[];
}>();

const formatFieldName = (field: string) => {
  const names: Record<string, string> = {
    name: "Nama",
    price: "Harga",
    stock: "Stok",
    image_count: "Jumlah Gambar",
  };
  return names[field] || field;
};
</script>

<style scoped>
.differences-section {
  margin-top: 1.5rem;
  padding-top: 1.5rem;
  border-top: 1px solid #e5e7eb;
}

.differences-section h4 {
  margin: 0 0 1rem 0;
  font-size: 0.875rem;
  font-weight: 700;
  color: #1a1a1a;
}

.differences-list {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.diff-item {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  padding: 0.5rem;
  background: #f9fafb;
  border-radius: 0.25rem;
  font-size: 0.875rem;
}

.diff-field {
  font-weight: 600;
  min-width: 100px;
}

.diff-source {
  color: #6b7280;
}

.diff-arrow {
  color: #9ca3af;
}

.diff-target {
  color: #f59e0b;
  font-weight: 600;
}
</style>
