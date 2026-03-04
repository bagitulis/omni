<template>
  <div class="url-section">
    <div class="url-card">
      <h3>📍 Your Integration URLs</h3>
      <p class="subtitle">Use these URLs in your platform developer settings</p>

      <div class="url-group">
        <label for="oauth-callback-url"
          >OAuth Callback URL (All Platforms)</label
        >
        <div class="url-display">
          <input
            id="oauth-callback-url"
            type="text"
            :value="baseUrl"
            readonly
            class="url-input"
          />
          <button @click="$emit('copy', baseUrl)" class="copy-btn">
            📋 Copy
          </button>
        </div>
        <span class="url-hint">
          Append platform: /api/platform-auth/shopee/callback,
          /api/platform-auth/tiktok/callback, etc.
        </span>
      </div>

      <div class="url-group">
        <label for="webhook-url">Webhook URL (All Platforms)</label>
        <div class="url-display">
          <input
            id="webhook-url"
            type="text"
            :value="webhookBaseUrl"
            readonly
            class="url-input"
          />
          <button @click="$emit('copy', webhookBaseUrl)" class="copy-btn">
            📋 Copy
          </button>
        </div>
        <span class="url-hint">
          {{
            tenantId
              ? `Full URL: ${webhookBaseUrl}/shopee, ${webhookBaseUrl}/tiktok, etc.`
              : "Login to see tenant-specific webhook URL"
          }}
        </span>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from "vue";

const props = defineProps<{
  baseUrl: string;
  tenantId: string;
}>();

defineEmits<{
  (e: "copy", url: string): void;
}>();

// Compute webhook base URL with tenant ID
const webhookBaseUrl = computed(() => {
  if (props.tenantId) {
    return `${props.baseUrl}/api/webhooks/${props.tenantId}`;
  }
  return `${props.baseUrl}/api/webhooks`;
});
</script>

<style scoped>
.url-section {
  margin-bottom: 32px;
}

.url-card {
  background: hsl(var(--b2));
  border-radius: 12px;
  padding: 24px;
  border: 1px solid hsl(var(--bc) / 0.1);
}

.url-card h3 {
  font-size: 1.1rem;
  margin-bottom: 4px;
}

.subtitle {
  color: hsl(var(--bc) / 0.6);
  font-size: 0.85rem;
  margin-bottom: 16px;
}

.url-group {
  margin-bottom: 16px;
}

.url-group label {
  display: block;
  font-weight: 500;
  margin-bottom: 8px;
  font-size: 0.9rem;
}

.url-display {
  display: flex;
  gap: 8px;
}

.url-input {
  flex: 1;
  padding: 10px 12px;
  border: 1px solid hsl(var(--bc) / 0.2);
  border-radius: 8px;
  background: hsl(var(--b1));
  color: hsl(var(--bc));
  font-family: monospace;
  font-size: 0.9rem;
}

.copy-btn {
  padding: 10px 16px;
  background: hsl(var(--p));
  color: hsl(var(--pc));
  border: none;
  border-radius: 8px;
  cursor: pointer;
  font-weight: 500;
  transition: opacity 0.2s;
}

.copy-btn:hover {
  opacity: 0.9;
}

.url-hint {
  display: block;
  margin-top: 6px;
  font-size: 0.8rem;
  color: hsl(var(--bc) / 0.5);
}
</style>
