<template>
  <div class="sidebar-section">
    <h3 class="section-title">
      <span>🔑 Token Status</span>
    </h3>
    <div v-if="status" class="token-status-container">
      <div
        v-for="(statusData, platform) in status"
        :key="platform"
        class="token-item"
        :class="getTokenStatusClass(statusData)"
      >
        <div class="token-platform">
          <PlatformBadge
            :platform="platform"
            size="sm"
            :status="getTokenStatus(statusData)"
          />
          <div class="token-info">
            <div class="token-platform-name">{{ platform.toUpperCase() }}</div>
            <div class="token-status-text">
              <div v-html="formatStatusText(statusData)"></div>
            </div>
          </div>
        </div>
      </div>
    </div>
    <div v-else class="text-loading">
      ⏳ Loading token status...
    </div>
    <button
      @click="$emit('load-status')"
      :disabled="loading"
      class="btn-refresh"
    >
      🔄 Refresh Status
    </button>
  </div>
</template>

<script lang="ts">
import { defineComponent } from "vue";
import PlatformBadge from "@/components/PlatformBadge.vue";
import { useRightSidebarStatus } from "@/composables/useRightSidebarStatus";

export default defineComponent({
  name: "TokenStatusSection",
  components: {
    PlatformBadge,
  },
  props: {
    status: {
      type: Object,
      default: null,
    },
    loading: {
      type: Boolean,
      default: false,
    },
  },
  emits: ["load-status"],
  setup() {
    const { getTokenStatusClass, getTokenStatus, formatStatusText } =
      useRightSidebarStatus();

    return {
      getTokenStatusClass,
      getTokenStatus,
      formatStatusText,
    };
  },
});
</script>

<style scoped>
.sidebar-section {
  flex-shrink: 0;
  display: flex;
  flex-direction: column;
  gap: 8px;
  padding: 10px;
  border-radius: 6px;
  background: linear-gradient(
    135deg,
    hsl(var(--b2)) 0%,
    hsl(var(--b2) / 0.8) 100%
  );
  border: 1px solid hsl(var(--b3) / 0.5);
  box-shadow:
    0 2px 6px rgba(0, 0, 0, 0.05),
    inset 0 1px 2px rgba(255, 255, 255, 0.1);
  transition: all 0.2s ease;
  min-height: 0;
  max-height: 45vh;
  overflow-y: auto;
}

.sidebar-section:hover {
  border-color: hsl(var(--p) / 0.2);
  box-shadow:
    0 4px 12px rgba(59, 130, 246, 0.1),
    inset 0 1px 2px rgba(255, 255, 255, 0.15);
}

.section-title {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 0.85rem;
  font-weight: 700;
  color: hsl(var(--bc));
  margin: 0 0 2px 0;
  white-space: nowrap;
  letter-spacing: 0.5px;
  padding-bottom: 6px;
  border-bottom: 2px solid hsl(var(--p) / 0.2);
}

.token-status-container {
  display: flex;
  flex-direction: column;
  gap: 6px;
  max-height: 200px;
  overflow-y: auto;
  padding-right: 4px;
}

.token-item {
  padding: 10px 12px;
  background: hsl(var(--b1));
  border-radius: 6px;
  border-left: 4px solid hsl(var(--b3));
  transition: all 0.2s ease;
  font-size: 0.8rem;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
  flex-shrink: 0;
  display: flex;
  align-items: center;
  gap: 10px;
}

.token-item:hover {
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(59, 130, 246, 0.2);
  border-left-color: hsl(var(--p) / 0.6);
  background: hsl(var(--b1) / 0.9);
}

.token-valid {
  border-left-color: #10b981;
  background: rgba(16, 185, 129, 0.08);
}

.token-expired {
  border-left-color: #ef4444;
  background: rgba(239, 68, 68, 0.08);
}

.token-warning {
  border-left-color: #f59e0b;
  background: rgba(245, 158, 11, 0.08);
}

.token-platform {
  display: flex;
  align-items: center;
  gap: 10px;
  width: 100%;
  min-width: 0;
}

.token-info {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.token-platform-name {
  font-size: 0.75rem;
  font-weight: 700;
  color: hsl(var(--bc));
  text-transform: uppercase;
  letter-spacing: 0.4px;
}

.token-status-text {
  font-size: 0.7rem;
  line-height: 1.4;
  color: hsl(var(--bc) / 0.8);
}

.text-loading {
  text-align: center;
  padding: 0.5rem;
  color: #6b7280;
  font-size: 0.875rem;
}

.btn-refresh {
  width: 100%;
  margin-top: 0.5rem;
  padding: 0.5rem 0.75rem;
  background: #3b82f6;
  color: white;
  border-radius: 0.5rem;
  transition: background-color 0.2s ease;
  font-size: 0.875rem;
  font-weight: 500;
  border: none;
  cursor: pointer;
}

.btn-refresh:hover {
  background: #2563eb;
}

.btn-refresh:disabled {
  background: #9ca3af;
  cursor: not-allowed;
}

.sidebar-section::-webkit-scrollbar {
  width: 5px;
}

.sidebar-section::-webkit-scrollbar-track {
  background: hsl(var(--b2));
}

.sidebar-section::-webkit-scrollbar-thumb {
  background: hsl(var(--b3));
  border-radius: 2px;
}

.sidebar-section::-webkit-scrollbar-thumb:hover {
  background: hsl(var(--p) / 0.5);
}
</style>
