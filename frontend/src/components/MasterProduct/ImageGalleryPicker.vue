<template>
  <div v-if="visible" class="modal-overlay" @click.self="close">
    <div class="modal-container">
      <div class="modal-header">
        <h3>Pilih Gambar dari Galeri</h3>
        <button class="btn-close" @click="close">✕</button>
      </div>

      <ImageGalleryToolbar
        :search-query="searchQuery"
        :uploading="uploading"
        @update:search-query="searchQuery = $event"
        @search="debouncedSearch"
        @clear="clearSearch"
        @upload="handleUploadFiles"
      />

      <div class="gallery-content">
        <div v-if="loading" class="loading-state">
          <div class="spinner"></div>
          <p>Memuat gambar...</p>
        </div>

        <div v-else-if="images.length === 0" class="empty-state">
          <span class="empty-icon">🖼️</span>
          <p v-if="searchQuery">Tidak ada gambar yang cocok</p>
          <p v-else>Belum ada gambar di galeri</p>
        </div>

        <div v-else class="image-grid">
          <ImageGalleryItem
            v-for="img in images"
            :key="img.id"
            :image="img"
            :image-url="getImageUrl(img)"
            :selected="isSelected(img)"
            @toggle="toggleSelection"
          />
        </div>
      </div>

      <ImageGalleryFooter
        :selected-count="selectedImages.length"
        :max-images="maxImages"
        :meta="meta"
        @page-change="changePage"
        @cancel="close"
        @confirm="confirm"
      />
    </div>

    <transition name="fade">
      <div v-if="errorMessage" class="message error">
        <span class="message-icon">✕</span>
        {{ errorMessage }}
      </div>
    </transition>
  </div>
</template>

<script setup lang="ts">
import { ref, watch } from "vue";
import type { GalleryImage } from "@/services/imageService";
import { useImageGallery } from "@/composables/useImageGallery";
import ImageGalleryItem from "./ImageGalleryItem.vue";
import ImageGalleryToolbar from "./ImageGalleryToolbar.vue";
import ImageGalleryFooter from "./ImageGalleryFooter.vue";

const props = defineProps({
  visible: { type: Boolean, default: false },
  multiple: { type: Boolean, default: true },
  maxImages: { type: Number, default: 8 },
  initialSelection: { type: Array as () => string[], default: () => [] },
});

const emit = defineEmits(["update:visible", "select", "close"]);

const {
  images,
  loading,
  uploading,
  errorMessage,
  meta,
  getImageUrl,
  fetchImages,
  uploadFiles,
} = useImageGallery();

const selectedImages = ref<GalleryImage[]>([]);
const searchQuery = ref("");
const searchTimeout = ref<number | null>(null);

const handleUploadFiles = async (files: FileList) => {
  const { successCount } = await uploadFiles(files);
  if (successCount > 0) await fetchImages(1, searchQuery.value);
  if (successCount < files.length) {
    errorMessage.value = `Gagal mengupload ${files.length - successCount} gambar`;
  }
};

const isSelected = (img: GalleryImage) =>
  selectedImages.value.some((i) => i.id === img.id);

const toggleSelection = (img: GalleryImage) => {
  const index = selectedImages.value.findIndex((i) => i.id === img.id);
  if (index >= 0) {
    selectedImages.value.splice(index, 1);
  } else {
    if (!props.multiple) {
      selectedImages.value = [img];
    } else if (selectedImages.value.length < props.maxImages) {
      selectedImages.value.push(img);
    } else {
      errorMessage.value = `Maksimal ${props.maxImages} gambar`;
      setTimeout(() => (errorMessage.value = ""), 2000);
    }
  }
};

const debouncedSearch = () => {
  if (searchTimeout.value) clearTimeout(searchTimeout.value);
  searchTimeout.value = window.setTimeout(
    () => fetchImages(1, searchQuery.value),
    500,
  );
};

const clearSearch = () => {
  searchQuery.value = "";
  fetchImages(1, "");
};

const changePage = (page: number) => fetchImages(page, searchQuery.value);

const close = () => {
  emit("update:visible", false);
  emit("close");
  selectedImages.value = [];
};

const confirm = () => {
  emit("select", selectedImages.value);
  close();
};

watch(
  () => props.visible,
  (newVal) => {
    if (newVal) {
      fetchImages(1, searchQuery.value);
      selectedImages.value = [];
    }
  },
);
</script>

<style scoped>
.modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
  backdrop-filter: blur(4px);
}

.modal-container {
  background: white;
  border-radius: 0.75rem;
  width: 90%;
  max-width: 900px;
  height: 85vh;
  display: flex;
  flex-direction: column;
  box-shadow:
    0 20px 25px -5px rgba(0, 0, 0, 0.1),
    0 10px 10px -5px rgba(0, 0, 0, 0.04);
  overflow: hidden;
}

.modal-header {
  padding: 1.25rem;
  border-bottom: 1px solid #e5e7eb;
  display: flex;
  justify-content: space-between;
  align-items: center;
  background: linear-gradient(135deg, #f8f9fa 0%, #ffffff 100%);
}

.modal-header h3 {
  margin: 0;
  font-size: 1.125rem;
  font-weight: 700;
  color: #1a1a1a;
}

.btn-close {
  background: transparent;
  border: none;
  font-size: 1.5rem;
  color: #6b7280;
  cursor: pointer;
  padding: 0.25rem;
  line-height: 1;
  border-radius: 0.375rem;
}

.btn-close:hover {
  background: #f3f4f6;
  color: #1a1a1a;
}

.gallery-content {
  flex: 1;
  overflow-y: auto;
  padding: 1.5rem;
  background: #f9fafb;
}

.image-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(140px, 1fr));
  gap: 1rem;
}

.loading-state,
.empty-state {
  height: 100%;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  color: #6b7280;
  text-align: center;
}

.spinner {
  width: 32px;
  height: 32px;
  border: 3px solid #e5e7eb;
  border-top-color: #ff6b2c;
  border-radius: 50%;
  animation: spin 0.8s linear infinite;
  margin-bottom: 1rem;
}

@keyframes spin {
  to {
    transform: rotate(360deg);
  }
}

.empty-icon {
  font-size: 3rem;
  margin-bottom: 1rem;
  opacity: 0.5;
}

.message {
  position: fixed;
  bottom: 2rem;
  right: 2rem;
  padding: 1rem 1.5rem;
  border-radius: 0.5rem;
  font-weight: 600;
  box-shadow: 0 8px 24px rgba(0, 0, 0, 0.15);
  z-index: 1100;
  display: flex;
  align-items: center;
  gap: 0.75rem;
}

.message.error {
  background: linear-gradient(135deg, #ef4444 0%, #dc2626 100%);
  color: white;
}

.fade-enter-active,
.fade-leave-active {
  transition: all 0.3s;
}

.fade-enter-from,
.fade-leave-to {
  opacity: 0;
  transform: translateY(1rem);
}
</style>
