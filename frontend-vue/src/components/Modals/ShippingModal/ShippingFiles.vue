<template>
  <div class="files-container">
    <div class="files-header">
      <h3 class="text-lg font-bold text-gray-800">Process Files</h3>
      <p class="text-sm text-gray-600 mt-1">
        Select a file to process shipping fee differences
      </p>
    </div>

    <div class="files-info">
      <div class="info-icon">📁</div>
      <span class="info-text">Available files in your system</span>
    </div>

    <div v-if="files.length > 0" class="files-list">
      <div v-for="file in files" :key="file.filename" class="file-item">
        <div class="file-icon">📄</div>
        <div class="file-info-content">
          <p class="file-name">{{ file.filename }}</p>
          <p class="file-description">{{ file.description }}</p>
          <div class="file-meta">
            <span class="meta-item">{{ formatFileSize(file.size) }}</span>
            <span class="meta-dot">•</span>
            <span class="meta-item">{{ file.modified }}</span>
          </div>
        </div>
        <button
          class="file-action-btn"
          @click="$emit('process', file.filename)"
        >
          Process
        </button>
      </div>
    </div>

    <div v-else class="empty-state">
      <p class="empty-icon">📭</p>
      <p class="empty-title">No Files Found</p>
      <p class="empty-description">
        There are no shipping files available in the system
      </p>
    </div>
  </div>
</template>

<script setup lang="ts">
interface Props {
  files: any[];
}

defineProps<Props>();

defineEmits<{
  process: [filename: string];
}>();

const formatFileSize = (bytes: number): string => {
  if (bytes === 0) return "0 Bytes";
  const k = 1024;
  const sizes = ["Bytes", "KB", "MB", "GB"];
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + " " + sizes[i];
};
</script>

<style scoped>
.files-container {
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
}

.files-header {
  padding: 1rem;
  background: linear-gradient(135deg, #f0f4f8 0%, #f8fafc 100%);
  border-radius: 0.75rem;
  border-left: 4px solid #3b82f6;
}

.files-header h3 {
  margin: 0;
  color: #1f2937;
  font-size: 1.125rem;
}

.files-header p {
  margin: 0;
  color: #6b7280;
  font-size: 0.875rem;
}

.files-info {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  padding: 1rem;
  background: #f5f3ff;
  border: 1px solid #e9d5ff;
  border-radius: 0.5rem;
}

.files-list {
  display: flex;
  flex-direction: column;
  gap: 1rem;
  max-height: 24rem;
  overflow-y: auto;
  padding-right: 0.5rem;
}

.file-item {
  display: flex;
  align-items: center;
  gap: 1rem;
  padding: 1rem;
  background: white;
  border: 2px solid #e5e7eb;
  border-radius: 0.5rem;
  transition: all 0.3s ease;
}

.file-item:hover {
  border-color: #3b82f6;
  background: #f0f9ff;
  box-shadow: 0 2px 8px rgba(59, 130, 246, 0.1);
}

.file-icon {
  font-size: 1.5rem;
  flex-shrink: 0;
}

.file-info-content {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
}

.file-name {
  margin: 0;
  color: #1f2937;
  font-weight: 600;
  font-size: 0.875rem;
}

.file-description {
  margin: 0;
  color: #6b7280;
  font-size: 0.75rem;
}

.file-meta {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  color: #9ca3af;
  font-size: 0.7rem;
}

.meta-dot {
  color: #d1d5db;
}

.file-action-btn {
  padding: 0.5rem 0.875rem;
  background: #3b82f6;
  color: white;
  border: none;
  border-radius: 0.375rem;
  font-size: 0.75rem;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s ease;
  white-space: nowrap;
  flex-shrink: 0;
}

.file-action-btn:hover {
  background: #2563eb;
  box-shadow: 0 2px 6px rgba(59, 130, 246, 0.3);
}

.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 3rem 1.5rem;
  background: #f9fafb;
  border: 2px dashed #e5e7eb;
  border-radius: 0.75rem;
  text-align: center;
}

.empty-icon {
  font-size: 2.5rem;
  margin-bottom: 0.75rem;
}

.empty-title {
  margin: 0;
  color: #1f2937;
  font-weight: 600;
  font-size: 1rem;
}

.empty-description {
  margin: 0.5rem 0 0 0;
  color: #6b7280;
  font-size: 0.875rem;
}
</style>
