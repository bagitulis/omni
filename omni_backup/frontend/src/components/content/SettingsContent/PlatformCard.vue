<template>
  <div class="platform-card" :class="platform">
    <div class="card-header">
      <div class="platform-info">
        <span class="platform-icon">{{ icon }}</span>
        <h3>{{ title }}</h3>
      </div>
      <div class="status-indicator" :class="getStatusClass()">
        {{ getStatusText() }}
      </div>
    </div>

    <div class="card-body">
      <!-- Connection Status -->
      <div class="status-section" v-if="status.connected && status.expiresAt">
        <span class="status-label">Token Expires:</span>
        <span
          class="status-value"
          :class="{
            'text-warning': isExpiringSoon(),
            'text-error': status.isExpired,
          }"
        >
          {{ formatExpiry(status.expiresAt) }}
        </span>
      </div>

      <!-- URLs -->
      <div class="url-item">
        <label>Callback URL</label>
        <div class="url-row">
          <code>{{ callbackUrl }}</code>
          <button
            @click="$emit('copy', callbackUrl)"
            class="mini-btn"
            title="Copy"
          >
            📋
          </button>
        </div>
      </div>

      <div class="url-item">
        <label>Webhook URL</label>
        <div class="url-row">
          <code>{{ webhookUrl }}</code>
          <button
            @click="$emit('copy', webhookUrl)"
            class="mini-btn"
            title="Copy"
          >
            📋
          </button>
        </div>
      </div>
    </div>

    <div class="card-footer">
      <button
        @click="$emit('authorize')"
        class="authorize-btn"
        :class="platform"
      >
        {{ status.connected ? "🔄 Re-authorize" : "🔗 Connect" }}
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
const props = defineProps<{
  platform: string;
  title: string;
  icon: string;
  status: { connected: boolean; isExpired?: boolean; expiresAt: string | null };
  callbackUrl: string;
  webhookUrl: string;
}>();

defineEmits<{
  (e: "authorize"): void;
  (e: "copy", url: string): void;
}>();

function getStatusClass() {
  if (!props.status.connected) return "disconnected";
  if (props.status.isExpired) return "expired";
  if (isExpiringSoon()) return "expiring";
  return "connected";
}

function getStatusText() {
  if (!props.status.connected) return "❌ Disconnected";
  if (props.status.isExpired) return "⚠️ Expired";
  if (isExpiringSoon()) return "⏳ Expiring Soon";
  return "✅ Connected";
}

function isExpiringSoon() {
  if (!props.status.expiresAt) return false;
  const date = new Date(props.status.expiresAt);
  const now = new Date();
  const hoursLeft = (date.getTime() - now.getTime()) / (1000 * 60 * 60);
  return hoursLeft > 0 && hoursLeft < 24;
}

function formatExpiry(dateStr: string | null) {
  if (!dateStr) return "N/A";
  const date = new Date(dateStr);
  const now = new Date();
  const diff = date.getTime() - now.getTime();
  const days = Math.floor(diff / (1000 * 60 * 60 * 24));

  if (days < 0) return "Expired";
  if (days === 0) return "Today";
  if (days === 1) return "Tomorrow";
  return `${days} days`;
}
</script>

<style scoped>
.platform-card {
  background: hsl(var(--b2));
  border-radius: 12px;
  padding: 20px;
  border: 1px solid hsl(var(--bc) / 0.1);
  transition:
    transform 0.2s,
    box-shadow 0.2s;
}

.platform-card:hover {
  transform: translateY(-2px);
  box-shadow: 0 4px 12px hsl(var(--bc) / 0.1);
}

.platform-card.shopee {
  border-top: 3px solid #ee4d2d;
}

.platform-card.tiktok {
  border-top: 3px solid #00f2ea;
}

.platform-card.lazada {
  border-top: 3px solid #0f1369;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
}

.platform-info {
  display: flex;
  align-items: center;
  gap: 8px;
}

.platform-icon {
  font-size: 1.5rem;
}

.platform-info h3 {
  font-size: 1.1rem;
  font-weight: 600;
  margin: 0;
}

.status-indicator {
  font-size: 0.85rem;
  padding: 4px 10px;
  border-radius: 20px;
}

.status-indicator.connected {
  background: hsl(var(--su) / 0.15);
  color: hsl(var(--su));
}

.status-indicator.disconnected {
  background: hsl(var(--er) / 0.15);
  color: hsl(var(--er));
}

.status-indicator.expired {
  background: hsl(var(--er) / 0.15);
  color: hsl(var(--er));
}

.status-indicator.expiring {
  background: hsl(var(--wa) / 0.15);
  color: hsl(var(--wa));
}

.text-warning {
  color: hsl(var(--wa)) !important;
}

.text-error {
  color: hsl(var(--er)) !important;
}

.card-body {
  margin-bottom: 16px;
}

.status-section {
  display: flex;
  justify-content: space-between;
  padding: 8px 12px;
  background: hsl(var(--b3));
  border-radius: 8px;
  margin-bottom: 12px;
  font-size: 0.9rem;
}

.status-label {
  color: hsl(var(--bc) / 0.7);
}

.status-value {
  font-weight: 500;
}

.url-item {
  margin-bottom: 12px;
}

.url-item label {
  display: block;
  font-size: 0.8rem;
  color: hsl(var(--bc) / 0.6);
  margin-bottom: 4px;
}

.url-row {
  display: flex;
  align-items: center;
  gap: 8px;
}

.url-row code {
  flex: 1;
  padding: 8px 10px;
  background: hsl(var(--b1));
  border-radius: 6px;
  font-size: 0.75rem;
  color: hsl(var(--bc) / 0.8);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.mini-btn {
  padding: 6px 10px;
  background: hsl(var(--b3));
  border: none;
  border-radius: 6px;
  cursor: pointer;
  transition: background 0.2s;
}

.mini-btn:hover {
  background: hsl(var(--bc) / 0.1);
}

.card-footer {
  padding-top: 12px;
  border-top: 1px solid hsl(var(--bc) / 0.1);
}

.authorize-btn {
  width: 100%;
  padding: 12px;
  border: none;
  border-radius: 8px;
  font-weight: 600;
  cursor: pointer;
  transition: opacity 0.2s;
}

.authorize-btn:hover {
  opacity: 0.9;
}

.authorize-btn.shopee {
  background: #ee4d2d;
  color: white;
}

.authorize-btn.tiktok {
  background: linear-gradient(135deg, #00f2ea, #ff0050);
  color: white;
}

.authorize-btn.lazada {
  background: #0f1369;
  color: white;
}
</style>
