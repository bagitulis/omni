<template>
  <div class="platform-item">
    <!-- Sync Button -->
    <button
      v-if="linked && hasUpdate && !syncState"
      class="sync-btn"
      @click.stop="$emit('sync')"
      title="Sinkronisasi perubahan"
    >
      🔄
    </button>

    <!-- Loading Spinner -->
    <div v-else-if="syncState === 'syncing'" class="sync-spinner"></div>

    <!-- Success Indicator -->
    <span v-else-if="syncState === 'success'" class="sync-success">✅</span>

    <!-- Error Indicator -->
    <span
      v-else-if="syncState === 'error'"
      class="sync-error"
      title="Gagal sinkronisasi"
      @click="$emit('sync')"
      >⚠️</span
    >

    <!-- Platform Icon -->
    <span
      class="platform-icon"
      :class="[
        platform,
        {
          linked: linked,
          'has-update': hasUpdate,
          'is-loading': loading,
        },
      ]"
      :title="title"
      @click.stop="$emit('click')"
    >
      {{ loading ? "⏳" : emoji }}
    </span>
  </div>
</template>

<script setup lang="ts">
defineProps<{
  platform: string;
  linked: boolean;
  hasUpdate: boolean;
  loading: boolean;
  syncState?: string;
  title: string;
  emoji: string;
}>();

defineEmits<{
  sync: [];
  click: [];
}>();
</script>

<style scoped>
.platform-item {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 0.25rem;
  position: relative;
}

.sync-btn {
  background: none;
  border: none;
  cursor: pointer;
  font-size: 0.75rem;
  padding: 0;
  line-height: 1;
  margin-bottom: -2px;
  animation: bounce 0.5s cubic-bezier(0.175, 0.885, 0.32, 1.275);
}

@keyframes bounce {
  0% {
    transform: scale(0);
  }
  50% {
    transform: scale(1.2);
  }
  100% {
    transform: scale(1);
  }
}

.sync-spinner {
  width: 12px;
  height: 12px;
  border: 2px solid #e5e7eb;
  border-top-color: #3b82f6;
  border-radius: 50%;
  animation: spin 1s linear infinite;
  margin-bottom: 2px;
}

.sync-success {
  font-size: 0.75rem;
  margin-bottom: -2px;
  animation: popIn 0.3s ease-out;
}

.sync-error {
  font-size: 0.75rem;
  margin-bottom: -2px;
  cursor: pointer;
  animation: shake 0.5s ease-in-out;
}

@keyframes popIn {
  from {
    transform: scale(0);
  }
  to {
    transform: scale(1);
  }
}

@keyframes shake {
  0%,
  100% {
    transform: translateX(0);
  }
  25% {
    transform: translateX(-2px);
  }
  75% {
    transform: translateX(2px);
  }
}

.platform-icon {
  font-size: 1.125rem;
  opacity: 0.2;
  cursor: pointer;
  transition: all 0.2s cubic-bezier(0.4, 0, 0.2, 1);
  filter: grayscale(100%);
  position: relative;
  display: inline-flex;
  align-items: center;
  justify-content: center;
}

.platform-icon:hover {
  opacity: 0.8;
  filter: grayscale(0%);
  transform: scale(1.2);
}

.platform-icon.linked {
  opacity: 1;
  filter: grayscale(0%);
  transform: scale(1.1);
}

.platform-icon.is-loading {
  opacity: 1;
  filter: grayscale(0%);
  animation: pulse 1.5s infinite;
  cursor: wait;
}

.platform-icon.has-update {
  position: relative;
}

.platform-icon.has-update::after {
  content: "!";
  position: absolute;
  top: -4px;
  right: -4px;
  background: #f59e0b;
  color: white;
  font-size: 0.5rem;
  width: 10px;
  height: 10px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  border: 1px solid white;
  font-weight: bold;
}

.platform-icon.linked::after {
  content: "✓";
  position: absolute;
  bottom: -4px;
  right: -4px;
  font-size: 0.5rem;
  background: #10b981;
  color: white;
  width: 12px;
  height: 12px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  border: 1.5px solid white;
  font-weight: bold;
}

.platform-icon:hover {
  transform: scale(1.2);
}

@keyframes spin {
  to {
    transform: rotate(360deg);
  }
}
</style>
