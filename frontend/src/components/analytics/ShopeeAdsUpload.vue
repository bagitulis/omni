<template>
  <div class="upload-container">
    <h2 class="upload-title">Upload Shopee Ads CSV</h2>

    <!-- Upload Area -->
    <div
      class="upload-area"
      :class="{ 'drag-over': isDragging }"
      @dragover.prevent="isDragging = true"
      @dragleave="isDragging = false"
      @drop.prevent="handleDrop"
    >
      <input
        id="csv-file-input"
        type="file"
        accept=".csv"
        class="file-input"
        @change="handleFileSelect"
      />
      <label for="csv-file-input" class="upload-label">
        <span class="upload-icon" aria-hidden="true">📁</span>
        <span class="upload-text">
          Drag & drop CSV file here or
          <span class="upload-link">browse</span>
        </span>
        <span class="upload-hint">
          Export from Shopee Seller Center → Ads → Download Report
        </span>
      </label>
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
          @click="clearFile"
          aria-label="Remove selected file"
        >
          <span aria-hidden="true">✕</span>
        </button>
      </div>

      <!-- Period Label Input -->
      <div class="period-input">
        <label for="period-label">Period Label (e.g., "Jun 2025")</label>
        <input
          id="period-label"
          v-model="periodLabel"
          type="text"
          placeholder="Jun 2025"
          class="input-field"
        />
      </div>

      <button
        type="button"
        class="btn-upload"
        :disabled="uploading || !periodLabel.trim()"
        @click="handleUpload"
      >
        <span v-if="uploading">
          <span aria-hidden="true">⏳</span> Uploading...
        </span>
        <span v-else> <span aria-hidden="true">📤</span> Upload Data </span>
      </button>
    </div>

    <!-- Upload Result -->
    <div v-if="uploadResult" class="upload-result" :class="resultClass">
      <p>{{ uploadResult.message }}</p>
      <ul v-if="uploadResult.details">
        <li>Processed: {{ uploadResult.details.processed }}</li>
        <li>Skipped: {{ uploadResult.details.skipped }}</li>
      </ul>
    </div>

    <!-- Upload History -->
    <div v-if="history.length > 0" class="upload-history">
      <h4>Recent Uploads</h4>
      <table class="history-table" aria-label="Upload history">
        <thead>
          <tr>
            <th scope="col">File</th>
            <th scope="col">Period</th>
            <th scope="col">Records</th>
            <th scope="col">Date</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="batch in history" :key="batch.id">
            <td>{{ truncateFilename(batch.filename) }}</td>
            <td>{{ batch.periodLabel }}</td>
            <td>{{ batch.recordCount }}</td>
            <td>{{ formatDate(batch.uploadedAt) }}</td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from "vue";
import type { UploadBatch } from "@/composables/useShopeeAdsAnalytics";

// Props (used in template)
defineProps<{
  uploading: boolean;
  history: UploadBatch[];
}>();

const emit = defineEmits<{
  upload: [file: File, periodLabel: string];
  "fetch-history": [];
}>();

// State
const selectedFile = ref<File | null>(null);
const periodLabel = ref("");
const isDragging = ref(false);
const uploadResult = ref<{
  success: boolean;
  message: string;
  details?: { processed: number; skipped: number };
} | null>(null);

// Computed
const resultClass = computed(() => ({
  "result-success": uploadResult.value?.success,
  "result-error": !uploadResult.value?.success,
}));

// Methods
function handleFileSelect(event: Event) {
  const input = event.target as HTMLInputElement;
  if (input.files?.[0]) {
    selectedFile.value = input.files[0];
    uploadResult.value = null;
  }
}

function handleDrop(event: DragEvent) {
  isDragging.value = false;
  const files = event.dataTransfer?.files;
  if (files?.[0]?.name.endsWith(".csv")) {
    selectedFile.value = files[0];
    uploadResult.value = null;
  }
}

function clearFile() {
  selectedFile.value = null;
  periodLabel.value = "";
  uploadResult.value = null;
}

async function handleUpload() {
  if (!selectedFile.value || !periodLabel.value.trim()) return;

  emit("upload", selectedFile.value, periodLabel.value.trim());
}

function setUploadResult(result: typeof uploadResult.value) {
  uploadResult.value = result;
  if (result?.success) {
    clearFile();
    emit("fetch-history");
  }
}

function formatFileSize(bytes: number): string {
  if (bytes < 1024) return `${bytes} B`;
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`;
  return `${(bytes / (1024 * 1024)).toFixed(1)} MB`;
}

function formatDate(date: string): string {
  return new Date(date).toLocaleDateString("id-ID", {
    day: "2-digit",
    month: "short",
    year: "numeric",
  });
}

function truncateFilename(name: string): string {
  return name.length > 25 ? name.substring(0, 22) + "..." : name;
}

// Expose for parent to call
defineExpose({ setUploadResult });

onMounted(() => {
  emit("fetch-history");
});
</script>

<style scoped>
.upload-container {
  padding: 20px;
  max-width: 600px;
}

.upload-title {
  margin-bottom: 20px;
  color: #1f2937;
  font-size: 18px;
}

.upload-area {
  border: 2px dashed #d1d5db;
  border-radius: 8px;
  padding: 40px;
  text-align: center;
  transition: all 0.2s;
  background: #fafafa;
}

.upload-area:hover,
.upload-area.drag-over {
  border-color: #f53d2d;
  background: #fef3f2;
}

.file-input {
  display: none;
}

.upload-label {
  cursor: pointer;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 8px;
}

.upload-icon {
  font-size: 48px;
}

.upload-text {
  color: #6b7280;
}

.upload-link {
  color: #dc2626;
  text-decoration: underline;
  font-weight: 600;
}

.upload-hint {
  font-size: 12px;
  color: #6b7280;
}

.file-preview {
  margin-top: 20px;
  padding: 16px;
  background: white;
  border: 1px solid #e5e7eb;
  border-radius: 8px;
}

.file-info {
  display: flex;
  align-items: center;
  gap: 12px;
}

.file-icon {
  font-size: 24px;
}

.file-details {
  flex: 1;
  display: flex;
  flex-direction: column;
}

.file-name {
  font-weight: 500;
  color: #1f2937;
}

.file-size {
  font-size: 12px;
  color: #9ca3af;
}

.btn-remove {
  padding: 4px 8px;
  border: none;
  background: #fee2e2;
  color: #dc2626;
  border-radius: 4px;
  cursor: pointer;
}

.period-input {
  margin-top: 16px;
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.period-input label {
  font-size: 13px;
  color: #6b7280;
}

.input-field {
  padding: 8px 12px;
  border: 1px solid #d1d5db;
  border-radius: 6px;
  font-size: 14px;
}

.btn-upload {
  margin-top: 16px;
  width: 100%;
  padding: 12px;
  background: #f53d2d;
  color: white;
  border: none;
  border-radius: 6px;
  cursor: pointer;
  font-size: 14px;
  font-weight: 500;
}

.btn-upload:hover:not(:disabled) {
  background: #d93025;
}

.btn-upload:disabled {
  background: #9ca3af;
  cursor: not-allowed;
}

.upload-result {
  margin-top: 16px;
  padding: 12px;
  border-radius: 6px;
}

.result-success {
  background: #dcfce7;
  color: #166534;
}

.result-error {
  background: #fee2e2;
  color: #991b1b;
}

.upload-history {
  margin-top: 32px;
}

.upload-history h4 {
  margin-bottom: 12px;
  color: #374151;
}

.history-table {
  width: 100%;
  border-collapse: collapse;
  font-size: 13px;
}

.history-table th,
.history-table td {
  padding: 8px 12px;
  text-align: left;
  border-bottom: 1px solid #e5e7eb;
}

.history-table th {
  background: #f9fafb;
  font-weight: 600;
  color: #374151;
}

.history-table td {
  color: #6b7280;
}
</style>
