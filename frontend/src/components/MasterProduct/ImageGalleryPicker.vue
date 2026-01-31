<template>
  <div v-if="visible" class="modal-overlay" @click.self="close">
    <div class="modal-container">
      <!-- Header -->
      <div class="modal-header">
        <h3>Pilih Gambar dari Galeri</h3>
        <button class="btn-close" @click="close">✕</button>
      </div>

      <!-- Toolbar -->
      <div class="toolbar">
        <div class="search-wrapper">
          <span class="search-icon">🔍</span>
          <input
            v-model="searchQuery"
            @input="debouncedSearch"
            placeholder="Cari nama file..."
            class="search-input"
          />
          <span v-if="searchQuery" class="clear-icon" @click="clearSearch"
            >✕</span
          >
        </div>

        <div class="actions">
          <input
            type="file"
            ref="fileInput"
            accept="image/*"
            multiple
            hidden
            @change="handleFileUpload"
          />
          <button
            class="btn-upload"
            @click="triggerUpload"
            :disabled="uploading"
          >
            <span v-if="uploading" class="spinner-sm"></span>
            <span v-else>+ Upload Baru</span>
          </button>
        </div>
      </div>

      <!-- Content -->
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
          <div
            v-for="img in images"
            :key="img.id"
            class="image-item"
            :class="{ selected: isSelected(img) }"
            @click="toggleSelection(img)"
          >
            <div class="image-wrapper">
              <img :src="getImageUrl(img)" :alt="img.filename" loading="lazy" />
            </div>
            <div class="image-info">
              <span class="filename" :title="img.filename">{{
                img.filename
              }}</span>
              <span class="meta">{{ img.width }}x{{ img.height }}</span>
            </div>
            <div class="selection-indicator" v-if="isSelected(img)">✓</div>
          </div>
        </div>
      </div>

      <!-- Footer -->
      <div class="modal-footer">
        <div class="footer-left">
          <div class="selection-info">
            Terpilih: {{ selectedImages.length }} / {{ maxImages }}
          </div>
          <div class="pagination" v-if="meta.pages > 1">
            <button
              class="page-btn"
              :disabled="meta.page === 1"
              @click="changePage(meta.page - 1)"
            >
              ‹
            </button>
            <span class="page-info">{{ meta.page }} / {{ meta.pages }}</span>
            <button
              class="page-btn"
              :disabled="meta.page === meta.pages"
              @click="changePage(meta.page + 1)"
            >
              ›
            </button>
          </div>
        </div>

        <div class="footer-actions">
          <button class="btn-cancel" @click="close">Batal</button>
          <button
            class="btn-confirm"
            @click="confirm"
            :disabled="selectedImages.length === 0"
          >
            Pilih ({{ selectedImages.length }})
          </button>
        </div>
      </div>
    </div>

    <!-- Messages -->
    <transition name="fade">
      <div v-if="errorMessage" class="message error">
        <span class="message-icon">✕</span>
        {{ errorMessage }}
      </div>
    </transition>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch } from "vue";
import imageService, {
  type GalleryImage,
  type GalleryResponse,
} from "@/services/imageService";

// Props
const props = defineProps({
  visible: {
    type: Boolean,
    default: false,
  },
  multiple: {
    type: Boolean,
    default: true,
  },
  maxImages: {
    type: Number,
    default: 8,
  },
  initialSelection: {
    type: Array as () => string[], // Array of filenames or paths
    default: () => [],
  },
});

// Emits
const emit = defineEmits(["update:visible", "select", "close"]);

// State
const images = ref<GalleryImage[]>([]);
const selectedImages = ref<GalleryImage[]>([]);
const loading = ref(false);
const uploading = ref(false);
const searchQuery = ref("");
const errorMessage = ref("");
const fileInput = ref<HTMLInputElement | null>(null);
const searchTimeout = ref<number | null>(null);

const meta = ref({
  total: 0,
  page: 1,
  page_size: 20,
  pages: 1,
});

// Methods
const getBackendUrl = () => {
  if (import.meta.env.DEV) {
    const currentUrl = window.location.origin;
    const isLocalhost =
      window.location.hostname === "localhost" ||
      window.location.hostname === "127.0.0.1";
    return isLocalhost ? `${currentUrl.replace(/:\d+$/, "")}:3000` : currentUrl;
  }
  return import.meta.env.VITE_API_URL || "";
};

const getImageUrl = (img: GalleryImage) => {
  // Construct full URL assuming backend serves uploads statically
  // If local_path already has /uploads, use it, otherwise prepend
  // Note: Backend might return path relative to tenant dir
  // Let's assume standardized path: /uploads/images/{tenant}/{path}
  // But wait, backend handles static files. Let's use relative path if it starts with /
  if (img.local_path.startsWith("http")) return img.local_path;

  const baseUrl = getBackendUrl();
  const path = img.local_path.startsWith("/")
    ? img.local_path
    : `/${img.local_path}`;

  // Quick fix: if path doesn't include /uploads, prepend it (based on common pattern)
  // However, looking at instructions: "Backend returns local_path like gallery/1738345678_filename.jpg"
  // And "Full URL construction: /uploads/images/{tenant_id}/{local_path}"

  // Since we don't have tenant_id easily here (unless passed or stored),
  // relying on backend to return full usable relative path is better,
  // OR the static middleware handles it.
  // Let's assume the API returns a path that is accessible via the static file server.

  // If the backend returns `gallery/file.jpg`, we need the full path.
  // For now, let's trust the path provided or prepend /uploads if needed.
  if (path.startsWith("/uploads")) return `${baseUrl}${path}`;

  // Fallback if tenant_id is needed but not available: try to use the path as is if it looks complete
  return `${baseUrl}/uploads/images/${img.tenant_id}/${img.local_path}`;
};

const fetchImages = async (page = 1) => {
  loading.value = true;
  errorMessage.value = "";
  try {
    const response = await imageService.getGallery(
      page,
      meta.value.page_size,
      searchQuery.value,
    );
    if (response.success) {
      images.value = response.data;
      meta.value = response.meta;
    }
  } catch (error: any) {
    console.error("Failed to fetch gallery:", error);
    errorMessage.value = "Gagal memuat galeri gambar";
  } finally {
    loading.value = false;
  }
};

const triggerUpload = () => {
  fileInput.value?.click();
};

const handleFileUpload = async (event: Event) => {
  const input = event.target as HTMLInputElement;
  if (!input.files || input.files.length === 0) return;

  uploading.value = true;
  errorMessage.value = "";
  let successCount = 0;

  try {
    // Process one by one or Promise.all. One by one for better error handling/feedback?
    // Let's do parallel
    const uploads = Array.from(input.files).map((file) =>
      imageService.uploadImage(file, "gallery"),
    );

    const results = await Promise.allSettled(uploads);

    // Check results
    results.forEach((res) => {
      if (res.status === "fulfilled" && res.value.success) {
        successCount++;
      }
    });

    if (successCount > 0) {
      // Refresh gallery
      await fetchImages(1);
    }

    if (successCount < input.files.length) {
      errorMessage.value = `Gagal mengupload ${input.files.length - successCount} gambar`;
    }
  } catch (error: any) {
    console.error("Upload error:", error);
    errorMessage.value = "Terjadi kesalahan saat upload";
  } finally {
    uploading.value = false;
    // Reset input
    if (fileInput.value) fileInput.value.value = "";
  }
};

const isSelected = (img: GalleryImage) => {
  return selectedImages.value.some((i) => i.id === img.id);
};

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
  searchTimeout.value = window.setTimeout(() => {
    fetchImages(1);
  }, 500);
};

const clearSearch = () => {
  searchQuery.value = "";
  fetchImages(1);
};

const changePage = (page: number) => {
  fetchImages(page);
};

const close = () => {
  emit("update:visible", false);
  emit("close");
  selectedImages.value = [];
};

const confirm = () => {
  emit("select", selectedImages.value);
  close();
};

// Lifecycle & Watchers
watch(
  () => props.visible,
  (newVal) => {
    if (newVal) {
      fetchImages(1);
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

/* Header */
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

/* Toolbar */
.toolbar {
  padding: 1rem;
  border-bottom: 1px solid #e5e7eb;
  display: flex;
  gap: 1rem;
  justify-content: space-between;
  align-items: center;
  background: white;
}

.search-wrapper {
  position: relative;
  flex: 1;
  max-width: 400px;
}

.search-input {
  width: 100%;
  padding: 0.625rem 2.5rem;
  border: 1px solid #e5e7eb;
  border-radius: 0.5rem;
  font-size: 0.9375rem;
  transition: all 0.2s;
}

.search-input:focus {
  outline: none;
  border-color: #ff6b2c;
  box-shadow: 0 0 0 3px rgba(255, 107, 44, 0.1);
}

.search-icon {
  position: absolute;
  left: 0.75rem;
  top: 50%;
  transform: translateY(-50%);
  color: #9ca3af;
}

.clear-icon {
  position: absolute;
  right: 0.75rem;
  top: 50%;
  transform: translateY(-50%);
  color: #9ca3af;
  cursor: pointer;
}

.btn-upload {
  padding: 0.625rem 1.25rem;
  background: #ff6b2c;
  color: white;
  border: none;
  border-radius: 0.5rem;
  font-weight: 600;
  cursor: pointer;
  display: flex;
  align-items: center;
  gap: 0.5rem;
  transition: all 0.2s;
}

.btn-upload:hover:not(:disabled) {
  background: #ff5511;
  transform: translateY(-1px);
}

.btn-upload:disabled {
  opacity: 0.7;
  cursor: not-allowed;
}

/* Content */
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

/* Loading & Empty States */
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

.spinner-sm {
  width: 16px;
  height: 16px;
  border: 2px solid rgba(255, 255, 255, 0.3);
  border-top-color: white;
  border-radius: 50%;
  animation: spin 0.8s linear infinite;
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

/* Footer */
.modal-footer {
  padding: 1rem 1.25rem;
  border-top: 1px solid #e5e7eb;
  display: flex;
  justify-content: space-between;
  align-items: center;
  background: white;
}

.footer-left {
  display: flex;
  gap: 1.5rem;
  align-items: center;
}

.selection-info {
  font-size: 0.875rem;
  font-weight: 600;
  color: #374151;
}

.pagination {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  font-size: 0.875rem;
}

.page-btn {
  padding: 0.25rem 0.5rem;
  border: 1px solid #e5e7eb;
  background: white;
  border-radius: 0.25rem;
  cursor: pointer;
}

.page-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.footer-actions {
  display: flex;
  gap: 0.75rem;
}

.btn-cancel {
  padding: 0.625rem 1.25rem;
  background: white;
  border: 1px solid #d1d5db;
  color: #374151;
  border-radius: 0.5rem;
  font-weight: 500;
  cursor: pointer;
}

.btn-confirm {
  padding: 0.625rem 1.5rem;
  background: #ff6b2c;
  color: white;
  border: none;
  border-radius: 0.5rem;
  font-weight: 600;
  cursor: pointer;
}

.btn-confirm:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

/* Messages */
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
