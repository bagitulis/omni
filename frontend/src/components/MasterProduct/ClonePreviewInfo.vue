<template>
  <div class="clone-preview-info">
    <!-- Conflict Warning -->
    <div v-if="hasConflict" class="conflict-warning">
      <span class="warning-icon">!</span>
      <div class="warning-text">
        <strong>Produk sudah ada di {{ targetPlatformLabel }}</strong>
        <p>SKU ini sudah terdaftar. Pilih untuk update atau buat baru.</p>
      </div>
    </div>

    <!-- Adjustment Info -->
    <div v-if="adjustments?.title_will_truncate" class="adjustment-info">
      <span class="info-icon">i</span>
      <div class="info-text">
        <strong>Judul akan dipotong</strong>
        <p>
          {{ targetPlatformLabel }} memiliki batas
          {{ adjustments.title_limit }} karakter
        </p>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import type { AdjustmentInfo } from "./ClonePreview.types";

defineProps<{
  hasConflict: boolean;
  targetPlatformLabel: string;
  adjustments?: AdjustmentInfo;
}>();
</script>

<style scoped>
.conflict-warning,
.adjustment-info {
  display: flex;
  gap: 1rem;
  padding: 1rem;
  border-radius: 0.5rem;
  margin-bottom: 1.5rem;
}

.conflict-warning {
  background: #fef3c7;
  border: 1px solid #f59e0b;
}

.adjustment-info {
  background: #e0f2fe;
  border: 1px solid #0ea5e9;
}

.warning-icon,
.info-icon {
  width: 32px;
  height: 32px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-weight: 700;
  flex-shrink: 0;
}

.warning-icon {
  background: #f59e0b;
  color: white;
}

.info-icon {
  background: #0ea5e9;
  color: white;
}

.warning-text strong,
.info-text strong {
  display: block;
  margin-bottom: 0.25rem;
}

.warning-text p,
.info-text p {
  margin: 0;
  font-size: 0.875rem;
  opacity: 0.8;
}
</style>
