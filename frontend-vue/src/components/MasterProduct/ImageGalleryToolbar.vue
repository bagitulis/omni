<template>
  <div class="toolbar">
    <div class="search-wrapper">
      <span class="search-icon">🔍</span>
      <input
        :value="searchQuery"
        @input="onInput"
        placeholder="Cari nama file..."
        class="search-input"
      />
      <span v-if="searchQuery" class="clear-icon" @click="onClear">✕</span>
    </div>

    <div class="actions">
      <input
        type="file"
        ref="fileInput"
        accept="image/*"
        multiple
        hidden
        @change="handleFileChange"
      />
      <button class="btn-upload" @click="triggerUpload" :disabled="uploading">
        <span v-if="uploading" class="spinner-sm"></span>
        <span v-else>+ Upload Baru</span>
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from "vue";

defineProps<{
  searchQuery: string;
  uploading: boolean;
}>();

const emit = defineEmits<{
  (e: "update:searchQuery", value: string): void;
  (e: "search"): void;
  (e: "clear"): void;
  (e: "upload", files: FileList): void;
}>();

const fileInput = ref<HTMLInputElement | null>(null);

const onInput = (event: Event) => {
  const target = event.target as HTMLInputElement;
  emit("update:searchQuery", target.value);
  emit("search");
};

const onClear = () => {
  emit("clear");
};

const triggerUpload = () => {
  fileInput.value?.click();
};

const handleFileChange = (event: Event) => {
  const input = event.target as HTMLInputElement;
  if (input.files && input.files.length > 0) {
    emit("upload", input.files);
  }
  if (input) input.value = ""; // Reset input
};
</script>

<style scoped>
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
</style>
