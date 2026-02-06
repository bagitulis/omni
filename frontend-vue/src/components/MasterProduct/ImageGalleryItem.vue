<template>
  <div
    class="image-item"
    :class="{ selected: selected }"
    @click="$emit('toggle', image)"
  >
    <div class="image-wrapper">
      <img :src="imageUrl" :alt="image.filename" loading="lazy" />
    </div>
    <div class="image-info">
      <span class="filename" :title="image.filename">
        {{ image.filename }}
      </span>
      <span class="meta">{{ image.width }}x{{ image.height }}</span>
    </div>
    <div class="selection-indicator" v-if="selected">✓</div>
  </div>
</template>

<script setup lang="ts">
import type { GalleryImage } from "@/services/imageService";

defineProps<{
  image: GalleryImage;
  imageUrl: string;
  selected: boolean;
}>();

defineEmits<{
  (e: "toggle", image: GalleryImage): void;
}>();
</script>

<style scoped>
.image-item {
  background: white;
  border: 2px solid transparent;
  border-radius: 0.5rem;
  overflow: hidden;
  cursor: pointer;
  transition: all 0.2s;
  position: relative;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
}

.image-item:hover {
  transform: translateY(-2px);
  box-shadow: 0 4px 6px rgba(0, 0, 0, 0.1);
}

.image-item.selected {
  border-color: #ff6b2c;
  background: #fff5f0;
}

.image-wrapper {
  aspect-ratio: 1;
  background: #f3f4f6;
  overflow: hidden;
}

.image-wrapper img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.image-info {
  padding: 0.5rem;
  font-size: 0.75rem;
}

.filename {
  display: block;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  font-weight: 500;
  color: #374151;
  margin-bottom: 0.25rem;
}

.meta {
  color: #9ca3af;
}

.selection-indicator {
  position: absolute;
  top: 0.5rem;
  right: 0.5rem;
  background: #ff6b2c;
  color: white;
  width: 20px;
  height: 20px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 0.75rem;
  font-weight: bold;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.2);
}
</style>
