<template>
  <div class="video-uploader">
    <label class="uploader-label">
      Product Video
      <span class="optional">(Optional)</span>
    </label>

    <!-- Video preview -->
    <div v-if="videoUrl" class="video-preview">
      <video :src="videoUrl" controls class="video-player">
        Your browser does not support video playback.
      </video>
      <button
        type="button"
        class="remove-video-btn"
        aria-label="Remove video"
        @click="removeVideo"
      >
        <i class="pi pi-trash" aria-hidden="true"></i>
        Remove Video
      </button>
    </div>

    <!-- Upload area -->
    <div v-else class="upload-area" @click="showUrlInput = true">
      <div class="upload-content">
        <i class="pi pi-video" aria-hidden="true"></i>
        <span class="upload-text">Add Video</span>
        <span class="upload-hint">MP4 format, max {{ maxSizeMb }}MB</span>
      </div>
    </div>

    <!-- URL input modal -->
    <div v-if="showUrlInput" class="url-input-overlay" @click.self="showUrlInput = false">
      <div class="url-input-card" role="dialog" aria-labelledby="video-url-title">
        <h4 id="video-url-title">Add Video by URL</h4>
        <div class="form-group">
          <label for="video-url-input">Video URL</label>
          <input
            id="video-url-input"
            v-model="inputUrl"
            type="url"
            placeholder="https://example.com/video.mp4"
            @keydown.enter="addVideoUrl"
          />
        </div>
        <p class="url-hint">
          Supported formats: MP4. Max duration: {{ maxDuration }} seconds.
        </p>
        <div class="url-actions">
          <button type="button" class="btn btn-secondary" @click="showUrlInput = false">
            Cancel
          </button>
          <button type="button" class="btn btn-primary" @click="addVideoUrl">
            Add Video
          </button>
        </div>
      </div>
    </div>

    <!-- Platform-specific info -->
    <div class="platform-info">
      <div v-if="platform === 'tiktok'" class="info-card">
        <i class="pi pi-info-circle" aria-hidden="true"></i>
        TikTok: 9:16 ratio recommended, max 30MB
      </div>
      <div v-else-if="platform === 'shopee'" class="info-card">
        <i class="pi pi-info-circle" aria-hidden="true"></i>
        Shopee: 1:1 or 9:16 ratio, 10-60 seconds
      </div>
      <div v-else-if="platform === 'lazada'" class="info-card">
        <i class="pi pi-info-circle" aria-hidden="true"></i>
        Lazada: Max 100MB, 60 seconds
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from "vue";

const props = defineProps<{
  videoUrl: string;
  platform?: "tiktok" | "shopee" | "lazada";
}>();

const emit = defineEmits<{
  "update:videoUrl": [url: string];
}>();

const showUrlInput = ref(false);
const inputUrl = ref("");

const maxSizeMb = computed(() => {
  if (props.platform === "lazada") return 100;
  return 30;
});

const maxDuration = computed(() => {
  return 60;
});

function addVideoUrl() {
  if (!inputUrl.value) return;
  emit("update:videoUrl", inputUrl.value);
  inputUrl.value = "";
  showUrlInput.value = false;
}

function removeVideo() {
  emit("update:videoUrl", "");
}
</script>

<style scoped>
.video-uploader {
  margin-bottom: 24px;
}

.uploader-label {
  display: block;
  font-weight: 600;
  font-size: 14px;
  color: #374151;
  margin-bottom: 12px;
}

.optional {
  font-weight: 400;
  color: #9ca3af;
}

.video-preview {
  position: relative;
  max-width: 400px;
}

.video-player {
  width: 100%;
  border-radius: 8px;
  background: #000;
}

.remove-video-btn {
  margin-top: 12px;
  padding: 8px 16px;
  background: #fef2f2;
  color: #ef4444;
  border: 1px solid #fecaca;
  border-radius: 6px;
  cursor: pointer;
  font-size: 13px;
  font-weight: 500;
  display: inline-flex;
  align-items: center;
  gap: 8px;
}

.remove-video-btn:hover {
  background: #fee2e2;
}

.upload-area {
  max-width: 400px;
  padding: 40px;
  border: 2px dashed #d1d5db;
  border-radius: 8px;
  text-align: center;
  cursor: pointer;
  transition: all 0.2s;
}

.upload-area:hover {
  border-color: #3b82f6;
  background: #eff6ff;
}

.upload-content {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 8px;
  color: #6b7280;
}

.upload-content i {
  font-size: 32px;
}

.upload-text {
  font-weight: 600;
  font-size: 14px;
  color: #374151;
}

.upload-hint {
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
  margin-bottom: 12px;
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

.url-hint {
  margin: 0 0 16px;
  font-size: 12px;
  color: #6b7280;
}

.url-actions {
  display: flex;
  gap: 12px;
  justify-content: flex-end;
}

.platform-info {
  margin-top: 12px;
}

.info-card {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 10px 12px;
  background: #f0f9ff;
  border-radius: 6px;
  font-size: 12px;
  color: #0369a1;
}

.btn { padding: 10px 16px; border: none; border-radius: 6px; cursor: pointer; font-weight: 600; font-size: 14px; }
.btn-primary { background: #3b82f6; color: white; }
.btn-primary:hover { background: #2563eb; }
.btn-secondary { background: #e5e7eb; color: #374151; }
.btn-secondary:hover { background: #d1d5db; }
</style>
