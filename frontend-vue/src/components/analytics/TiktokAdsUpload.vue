<template>
  <div class="upload-container">
    <!-- Upload Zone -->
    <div
      class="upload-zone"
      :class="{ 'drag-active': isDragging }"
      @dragover.prevent="isDragging = true"
      @dragleave.prevent="isDragging = false"
      @drop.prevent="handleDrop"
      role="button"
      tabindex="0"
      aria-label="Drop Excel file here or click to upload"
      @keydown.enter="triggerFileInput"
    >
      <input
        ref="fileInputRef"
        type="file"
        accept=".xlsx,.xls"
        class="file-input"
        aria-label="Select Excel file"
        @change="handleFileSelect"
      />
      <div class="upload-icon" aria-hidden="true">📁</div>
      <h2 class="upload-title">Upload TikTok Ads Export</h2>
      <p class="upload-desc">Drag & drop Excel file or click to browse</p>
      <p class="upload-hint">Accepted formats: .xlsx, .xls (max 50MB)</p>
    </div>

    <!-- Upload Mode Selection -->
    <div class="mode-selection">
      <label class="mode-label">Duplicate Handling:</label>
      <div class="mode-options">
        <label class="mode-option">
          <input type="radio" v-model="uploadMode" value="skip" />
          <span class="mode-text">Skip existing</span>
        </label>
        <label class="mode-option">
          <input type="radio" v-model="uploadMode" value="update" />
          <span class="mode-text">Update existing</span>
        </label>
      </div>
    </div>

    <!-- Selected File Preview -->
    <div v-if="selectedFile" class="file-preview">
      <div class="file-info">
        <span class="file-icon" aria-hidden="true">📄</span>
        <div class="file-details">
          <span class="file-name">{{ selectedFile.name }}</span>
          <span class="file-size">{{ formatFileSize(selectedFile.size) }}</span>
        </div>
        <button
          type="button"
          class="btn-remove"
          aria-label="Remove selected file"
          @click="clearFile"
        >
          <span aria-hidden="true">✕</span>
        </button>
      </div>
      <button
        type="button"
        class="btn-upload"
        :disabled="uploading"
        @click="handleUpload"
      >
        <span v-if="uploading"
          ><span aria-hidden="true">⏳</span> Uploading...</span
        >
        <span v-else><span aria-hidden="true">📤</span> Upload File</span>
      </button>
    </div>

    <!-- Upload Result -->
    <div v-if="uploadResult" class="upload-result" :class="resultClass">
      <h4 class="result-title">{{ resultTitle }}</h4>
      <ul class="result-stats">
        <li>Processed: {{ uploadResult.processedRows }}</li>
        <li>Inserted: {{ uploadResult.insertedRows }}</li>
        <li>Skipped: {{ uploadResult.skippedRows }}</li>
        <li v-if="uploadResult.errorRows > 0">
          Errors: {{ uploadResult.errorRows }}
        </li>
      </ul>
    </div>

    <!-- Upload History -->
    <div v-if="uploads.length > 0" class="upload-history">
      <h3 class="history-title">Recent Uploads</h3>
      <table class="history-table" aria-label="Upload history">
        <thead>
          <tr>
            <th scope="col">Filename</th>
            <th scope="col">Date</th>
            <th scope="col">Status</th>
            <th scope="col">Rows</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="upload in uploads" :key="upload.id">
            <td class="filename">{{ upload.fileName }}</td>
            <td class="date">{{ formatDate(upload.createdAt) }}</td>
            <td>
              <span class="status-badge" :class="statusClass(upload.status)">{{
                upload.status
              }}</span>
            </td>
            <td class="rows">
              {{ upload.insertedRows }} / {{ upload.totalRows }}
            </td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from "vue";
import type {
  UploadBatch,
  UploadResult,
} from "@/composables/useTiktokAdsAnalytics";

defineProps<{ uploads: UploadBatch[]; uploading: boolean }>();
const emit = defineEmits<{ upload: [file: File, mode: "skip" | "update"] }>();

const isDragging = ref(false);
const selectedFile = ref<File | null>(null);
const uploadMode = ref<"skip" | "update">("skip");
const uploadResult = ref<UploadResult | null>(null);
const fileInputRef = ref<HTMLInputElement>();

const resultClass = computed(() =>
  !uploadResult.value
    ? ""
    : uploadResult.value.errorRows > 0
      ? "result-warning"
      : "result-success",
);
const resultTitle = computed(() =>
  !uploadResult.value
    ? ""
    : uploadResult.value.errorRows > 0
      ? "Upload completed with errors"
      : "Upload successful",
);

function handleDrop(e: DragEvent) {
  isDragging.value = false;
  const files = e.dataTransfer?.files;
  if (files && files.length > 0) selectFile(files[0]);
}

function handleFileSelect(e: Event) {
  const input = e.target as HTMLInputElement;
  if (input.files && input.files.length > 0) selectFile(input.files[0]);
}

function selectFile(file: File) {
  if (!file.name.endsWith(".xlsx") && !file.name.endsWith(".xls")) {
    alert("Please select an Excel file (.xlsx or .xls)");
    return;
  }
  if (file.size > 50 * 1024 * 1024) {
    alert("File size must be less than 50MB");
    return;
  }
  selectedFile.value = file;
  uploadResult.value = null;
}

function clearFile() {
  selectedFile.value = null;
  uploadResult.value = null;
  if (fileInputRef.value) fileInputRef.value.value = "";
}

function triggerFileInput() {
  fileInputRef.value?.click();
}
function handleUpload() {
  if (selectedFile.value) emit("upload", selectedFile.value, uploadMode.value);
}

function formatFileSize(bytes: number): string {
  if (bytes < 1024) return `${bytes} B`;
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`;
  return `${(bytes / (1024 * 1024)).toFixed(1)} MB`;
}

function formatDate(dateStr: string): string {
  return new Date(dateStr).toLocaleDateString("id-ID", {
    day: "numeric",
    month: "short",
    year: "numeric",
    hour: "2-digit",
    minute: "2-digit",
  });
}

function statusClass(status: string): string {
  switch (status) {
    case "completed":
      return "status-success";
    case "processing":
      return "status-processing";
    case "failed":
      return "status-failed";
    default:
      return "";
  }
}

function setUploadResult(result: UploadResult) {
  uploadResult.value = result;
  clearFile();
}
defineExpose({ setUploadResult });
</script>

<style scoped src="./TiktokAdsUpload.css"></style>
