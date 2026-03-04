<template>
  <div class="operation-header">
    <div class="platform-info">
      <div class="platform-header">
        <span class="platform-badge" :class="`platform-${platform}`">
          {{ getPlatformIcon(platform) }}
        </span>
        <div class="platform-details">
          <h2>
            {{ platform.charAt(0).toUpperCase() + platform.slice(1) }}
            Operations
          </h2>
          <p class="subtitle">Manage orders and platform authentication</p>
        </div>
      </div>

      <div class="status-indicator" :class="getPlatformStatusClass()">
        <span class="status-dot"></span>
        <span class="status-text">{{ getPlatformStatusText() }}</span>
      </div>
    </div>

    <div class="header-actions">
      <button
        class="action-btn refresh-btn"
        type="button"
        aria-label="Refresh status"
        @click="$emit('refresh-status')"
        :disabled="loading"
      >
        <i class="pi pi-refresh" aria-hidden="true"></i>
        <span>Refresh</span>
      </button>
    </div>
  </div>
</template>

<script lang="ts">
import { defineComponent } from "vue";
import { useOperationStatus } from "./composables/useOperationStatus";

export default defineComponent({
  name: "OperationHeader",
  props: {
    platform: {
      type: String,
      default: "shopee",
    },
    status: {
      type: Object,
      default: null,
    },
    loading: {
      type: Boolean,
      default: false,
    },
  },
  emits: ["refresh-status"],
  setup(props) {
    const { getPlatformIcon, getPlatformStatusClass, getPlatformStatusText } =
      useOperationStatus(props);

    return {
      getPlatformIcon,
      getPlatformStatusClass,
      getPlatformStatusText,
    };
  },
});
</script>

<style scoped>
.operation-header {
  background: white;
  padding: 16px 20px;
  border-bottom: 2px solid #e8eef5;
  display: flex;
  justify-content: space-between;
  align-items: center;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.05);
  flex-wrap: wrap;
  gap: 16px;
}

.platform-info {
  display: flex;
  align-items: center;
  gap: 16px;
  flex: 1;
  min-width: 250px;
}

.platform-header {
  display: flex;
  align-items: center;
  gap: 12px;
}

.platform-badge {
  width: 50px;
  height: 50px;
  border-radius: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 28px;
  font-weight: 600;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
  flex-shrink: 0;
}

.platform-badge.platform-shopee {
  background: linear-gradient(135deg, #ff6b6b, #ff8787);
}

.platform-badge.platform-lazada {
  background: linear-gradient(135deg, #4a90e2, #357abd);
}

.platform-badge.platform-tiktok {
  background: linear-gradient(135deg, #000000, #25f4ee);
}

.platform-details h2 {
  font-size: 1.3rem;
  font-weight: 700;
  color: #2c3e50;
  margin: 0;
}

.platform-details .subtitle {
  font-size: 0.85rem;
  color: #4b5563; /* Improved from #7a8fa6 for WCAG AA 4.5:1 contrast ratio */
  margin: 4px 0 0 0;
  font-weight: 400;
}

.status-indicator {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 12px 20px;
  border-radius: 8px;
  font-weight: 600;
  font-size: 0.95rem;
}

.status-indicator .status-dot {
  width: 10px;
  height: 10px;
  border-radius: 50%;
  animation: pulse 2s infinite;
}

@keyframes pulse {
  0%,
  100% {
    opacity: 1;
  }
  50% {
    opacity: 0.5;
  }
}

.status-indicator.status-success {
  background: #d4edda;
  color: #155724;
  border: 1px solid #c3e6cb;
}

.status-indicator.status-success .status-dot {
  background: #28a745;
}

.status-indicator.status-error {
  background: #f8d7da;
  color: #721c24;
  border: 1px solid #f5c6cb;
}

.status-indicator.status-error .status-dot {
  background: #dc3545;
}

.status-indicator.status-warning {
  background: #fff3cd;
  color: #856404;
  border: 1px solid #ffeaa7;
}

.status-indicator.status-warning .status-dot {
  background: #ffc107;
}

.status-indicator.status-unknown {
  background: #e2e3e5;
  color: #383d41;
  border: 1px solid #d6d8db;
}

.status-indicator.status-unknown .status-dot {
  background: #6c757d;
}

.header-actions {
  display: flex;
  gap: 12px;
}

.action-btn {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 10px 16px;
  border: 2px solid #dfe4f0;
  background: white;
  border-radius: 8px;
  font-size: 0.9rem;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  color: #2c3e50;
  white-space: nowrap;
}

.action-btn:hover:not(:disabled) {
  border-color: #4a90e2;
  background: #f0f5ff;
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(74, 144, 226, 0.2);
}

.action-btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.action-btn i {
  font-size: 1.1rem;
}

@media (max-width: 768px) {
  .operation-header {
    flex-direction: column;
    align-items: flex-start;
    padding: 16px;
  }

  .platform-info {
    width: 100%;
    gap: 16px;
  }

  .platform-details h2 {
    font-size: 1.2rem;
  }

  .header-actions {
    width: 100%;
    gap: 10px;
  }

  .action-btn {
    flex: 1;
    justify-content: center;
    padding: 10px 16px;
  }
}

@media (max-width: 480px) {
  .platform-header {
    gap: 12px;
  }

  .platform-badge {
    width: 48px;
    height: 48px;
    font-size: 24px;
  }

  .platform-details h2 {
    font-size: 1rem;
  }

  .platform-details .subtitle {
    font-size: 0.8rem;
  }

  .action-btn {
    padding: 8px 12px;
    font-size: 0.85rem;
  }

  .action-btn span {
    display: none;
  }
}
</style>
