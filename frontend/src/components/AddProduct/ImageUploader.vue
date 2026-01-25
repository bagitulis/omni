<template>
  <div class="image-uploader">
    <label class="uploader-label">
      Product Images *
      <span class="image-count">({{ images.length }}/{{ maxImages }})</span>
    </label>

    <div class="images-grid">
      <!-- Image thumbnails -->
      <div
        v-for="(image, index) in images"
        :key="index"
        class="image-item"
        :class="{ 'is-main': index === 0 }"
        draggable="true"
        @dragstart="handleDragStart(index)"
        @dragover.prevent="handleDragOver(index)"
        @drop="handleDrop(index)"
      >
        <img :src="image" :alt="`Product image ${index + 1}`" />
        <div class="image-overlay">
          <span v-if="index === 0" class="main-badge">Main</span>
          <button
            type="button"
            class="remove-btn"
            :aria-label="`Remove image ${index + 1}`"
            @click="$emit('remove-image', index)"
          >
            <i class="pi pi-times" aria-hidden="true"></i>
          </button>
        </div>
        <div class="drag-hint">Drag to reorder</div>
      </div>

      <!-- Add image button -->
      <div v-if="images.length < maxImages" class="add-image-box">
        <div class="add-content" @click="showUrlInput = true">
          <i class="pi pi-plus" aria-hidden="true"></i>
          <span>Add Image</span>
        </div>
      </div>
    </div>

    <!-- URL input modal -->
    <div v-if="showUrlInput" class="url-input-overlay" @click.self="showUrlInput = false">
      <div class="url-input-card" role="dialog" aria-labelledby="url-input-title">
        <h4 id="url-input-title">Add Image by URL</h4>
        <div class="form-group">
          <label for="image-url">Image URL</label>
          <input
            id="image-url"
            v-model="imageUrl"
            type="url"
            placeholder="https://example.com/image.jpg"
            @keydown.enter="addImageUrl"
          />
        </div>
        <div class="url-actions">
          <button type="button" class="btn btn-secondary" @click="showUrlInput = false">
            Cancel
          </button>
          <button type="button" class="btn btn-primary" @click="addImageUrl">
            Add Image
          </button>
        </div>
      </div>
    </div>

    <p class="uploader-hint">
      <i class="pi pi-info-circle" aria-hidden="true"></i>
      First image will be the main product image. Drag to reorder.
    </p>
  </div>
</template>

<script setup lang="ts">
import { ref } from "vue";

const props = defineProps<{
  images: string[];
  maxImages?: number;
}>();

const emit = defineEmits<{
  "add-image": [url: string];
  "remove-image": [index: number];
  "reorder-images": [images: string[]];
}>();

const maxImages = props.maxImages || 9;
const showUrlInput = ref(false);
const imageUrl = ref("");
const dragIndex = ref<number | null>(null);

function addImageUrl() {
  if (!imageUrl.value) return;
  emit("add-image", imageUrl.value);
  imageUrl.value = "";
  showUrlInput.value = false;
}

function handleDragStart(index: number) {
  dragIndex.value = index;
}

function handleDragOver(_index: number) {
  // Visual feedback handled by CSS
}

function handleDrop(targetIndex: number) {
  if (dragIndex.value === null || dragIndex.value === targetIndex) return;
  
  const newImages = [...props.images];
  const [draggedItem] = newImages.splice(dragIndex.value, 1);
  newImages.splice(targetIndex, 0, draggedItem);
  
  emit("reorder-images", newImages);
  dragIndex.value = null;
}
</script>

<style scoped>
.image-uploader {
  margin-bottom: 24px;
}

.uploader-label {
  display: block;
  font-weight: 600;
  font-size: 14px;
  color: #374151;
  margin-bottom: 12px;
}

.image-count {
  font-weight: 400;
  color: #6b7280;
}

.images-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(100px, 1fr));
  gap: 12px;
  max-width: 600px;
}

.image-item {
  position: relative;
  aspect-ratio: 1;
  border-radius: 8px;
  overflow: hidden;
  border: 2px solid #e5e7eb;
  cursor: grab;
  transition: border-color 0.2s;
}

.image-item:active {
  cursor: grabbing;
}

.image-item.is-main {
  border-color: #3b82f6;
}

.image-item img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.image-overlay {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  padding: 6px;
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  background: linear-gradient(180deg, rgba(0,0,0,0.5) 0%, transparent 100%);
}

.main-badge {
  background: #3b82f6;
  color: white;
  padding: 2px 8px;
  border-radius: 4px;
  font-size: 11px;
  font-weight: 600;
}

.remove-btn {
  background: rgba(239, 68, 68, 0.9);
  color: white;
  border: none;
  width: 24px;
  height: 24px;
  border-radius: 50%;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 12px;
}

.remove-btn:hover {
  background: #ef4444;
}

.drag-hint {
  position: absolute;
  bottom: 0;
  left: 0;
  right: 0;
  background: rgba(0,0,0,0.6);
  color: white;
  font-size: 10px;
  text-align: center;
  padding: 4px;
  opacity: 0;
  transition: opacity 0.2s;
}

.image-item:hover .drag-hint {
  opacity: 1;
}

.add-image-box {
  aspect-ratio: 1;
  border: 2px dashed #d1d5db;
  border-radius: 8px;
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  transition: all 0.2s;
}

.add-image-box:hover {
  border-color: #3b82f6;
  background: #eff6ff;
}

.add-content {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 8px;
  color: #6b7280;
}

.add-content i {
  font-size: 24px;
}

.add-content span {
  font-size: 12px;
}

.url-input-overlay {
  position: fixed;
  inset: 0;
  background: rgba(0,0,0,0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
}

.url-input-card {
  background: white;
  padding: 24px;
  border-radius: 12px;
  width: 90%;
  max-width: 400px;
  box-shadow: 0 20px 25px -5px rgba(0,0,0,0.1);
}

.url-input-card h4 {
  margin: 0 0 16px;
  font-size: 16px;
  color: #1f2937;
}

.form-group {
  margin-bottom: 16px;
}

.form-group label {
  display: block;
  margin-bottom: 6px;
  font-weight: 600;
  font-size: 14px;
  color: #374151;
}

.form-group input {
  width: 100%;
  padding: 10px 12px;
  border: 1px solid #d1d5db;
  border-radius: 6px;
  font-size: 14px;
}

.form-group input:focus {
  outline: none;
  border-color: #3b82f6;
  box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.1);
}

.url-actions {
  display: flex;
  gap: 12px;
  justify-content: flex-end;
}

.uploader-hint {
  margin-top: 12px;
  font-size: 12px;
  color: #6b7280;
  display: flex;
  align-items: center;
  gap: 6px;
}

.btn { padding: 10px 16px; border: none; border-radius: 6px; cursor: pointer; font-weight: 600; font-size: 14px; }
.btn-primary { background: #3b82f6; color: white; }
.btn-primary:hover { background: #2563eb; }
.btn-secondary { background: #e5e7eb; color: #374151; }
.btn-secondary:hover { background: #d1d5db; }
</style>
