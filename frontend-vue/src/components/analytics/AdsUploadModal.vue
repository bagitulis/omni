<template>
  <Teleport to="body">
    <div
      v-if="isOpen"
      class="modal-overlay"
      @click.self="$emit('close')"
      role="dialog"
      aria-modal="true"
      aria-labelledby="ads-upload-title"
    >
      <div class="modal-container">
        <!-- Header -->
        <div class="modal-header">
          <h2 id="ads-upload-title" class="modal-title">
            <Icon name="upload" size="md" /> Upload Ads Report
          </h2>
          <button
            class="modal-close-btn"
            type="button"
            aria-label="Close modal"
            @click="$emit('close')"
          >
            <Icon name="x" size="md" />
          </button>
        </div>

        <!-- Body -->
        <div class="modal-body">
          <p class="description">
            Upload the CSV/Excel file exported from
            <strong>{{ platformName }}</strong
            >.
          </p>

          <div
            class="drop-zone"
            :class="{ 'has-file': selectedFile }"
            @dragover.prevent
            @drop.prevent="handleDrop"
          >
            <div class="drop-content">
              <div class="icon-wrapper">
                <Icon name="upload-cloud" size="xl" class="upload-icon" />
              </div>
              <div class="upload-controls">
                <label for="file-upload" class="file-label">
                  <span>Click to upload</span>
                  <input
                    id="file-upload"
                    name="file-upload"
                    type="file"
                    class="sr-only"
                    accept=".csv, .xlsx, .xls"
                    @change="handleFileSelect"
                  />
                </label>
                <span class="text-secondary">or drag and drop</span>
              </div>
              <p class="file-hint">CSV or XLSX up to 10MB</p>
            </div>
          </div>

          <!-- File Status -->
          <div v-if="selectedFile" class="file-status success">
            <Icon name="check-circle" size="sm" />
            <span class="file-name">{{ selectedFile.name }}</span>
            <button
              type="button"
              class="remove-file"
              @click="selectedFile = null"
              aria-label="Remove file"
            >
              <Icon name="x" size="sm" />
            </button>
          </div>

          <!-- Error Message -->
          <div v-if="error" class="file-status error" role="alert">
            <Icon name="alert-circle" size="sm" />
            <span>{{ error }}</span>
          </div>
        </div>

        <!-- Footer -->
        <div class="modal-footer">
          <button
            class="modal-btn modal-btn-secondary"
            type="button"
            @click="$emit('close')"
          >
            Cancel
          </button>
          <button
            class="modal-btn modal-btn-primary"
            type="button"
            :disabled="!selectedFile || uploading"
            @click="upload"
          >
            <Icon
              v-if="uploading"
              name="loader"
              class="animate-spin"
              size="sm"
            />
            <span v-else><Icon name="upload" size="sm" /> Upload</span>
          </button>
        </div>
      </div>
    </div>
  </Teleport>
</template>

<script setup lang="ts">
import { ref, watch, onMounted, onUnmounted } from "vue";
import { useAnalyticsStore } from "../../stores/analytics";
import Icon from "@/components/ui/Icon.vue";

const props = defineProps<{
  isOpen: boolean;
  platformName: string;
}>();

const emit = defineEmits<{
  (e: "close"): void;
  (e: "uploaded"): void;
}>();

const store = useAnalyticsStore();
const selectedFile = ref<File | null>(null);
const uploading = ref(false);
const error = ref<string | null>(null);

// File Handling
function handleFileSelect(event: Event) {
  const target = event.target as HTMLInputElement;
  if (target.files && target.files.length > 0) {
    validateAndSetFile(target.files[0]);
  }
}

function handleDrop(event: DragEvent) {
  if (event.dataTransfer?.files && event.dataTransfer.files.length > 0) {
    validateAndSetFile(event.dataTransfer.files[0]);
  }
}

function validateAndSetFile(file: File) {
  if (file.size > 10 * 1024 * 1024) {
    error.value = "File size exceeds 10MB limit.";
    selectedFile.value = null;
    return;
  }
  selectedFile.value = file;
  error.value = null;
}

async function upload() {
  if (!selectedFile.value) return;

  uploading.value = true;
  error.value = null;

  try {
    await store.uploadReport(selectedFile.value);
    emit("uploaded");
    emit("close");
    selectedFile.value = null;
  } catch (e: unknown) {
    if (e && typeof e === "object" && "response" in e) {
      const err = e as { response: { data?: { error?: string } } };
      error.value =
        err.response.data?.error ||
        "Upload failed. Please check the file format.";
    } else {
      error.value = "Upload failed. Please check the file format.";
    }
  } finally {
    uploading.value = false;
  }
}

// Accessibility & Lifecycle
const handleKeydown = (e: KeyboardEvent) => {
  if (e.key === "Escape" && props.isOpen) {
    emit("close");
  }
};

watch(
  () => props.isOpen,
  (isOpen) => {
    if (isOpen) {
      document.body.style.overflow = "hidden";
      if (!uploading.value) error.value = null;
    } else {
      document.body.style.overflow = "";
    }
  },
);

onMounted(() => {
  document.addEventListener("keydown", handleKeydown);
});

onUnmounted(() => {
  document.removeEventListener("keydown", handleKeydown);
  document.body.style.overflow = "";
});
</script>

<style src="./AdsUploadModal.styles.css" scoped></style>
