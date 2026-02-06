<template>
  <div class="product-thumbnail">
    <img
      v-if="imageUrl"
      :src="imageUrl"
      :alt="altText"
      class="thumbnail-image"
      loading="lazy"
    />
    <div v-else class="thumbnail-placeholder">
      <span>📷</span>
    </div>
  </div>
</template>

<script setup lang="ts">
defineProps<{
  imageUrl: string | null;
  altText: string;
}>();
</script>

<style scoped>
.product-thumbnail {
  width: 80px;
  height: 80px;
  border-radius: 0.375rem;
  overflow: hidden;
  flex-shrink: 0;
  background: var(--surface-hover, #f3f4f6);
}

.thumbnail-image {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.thumbnail-placeholder {
  width: 100%;
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 1.5rem;
  color: var(--text-secondary, #6b7280);
}

@media (max-width: 640px) {
  .product-thumbnail {
    width: 60px;
    height: 60px;
  }
}
</style>
