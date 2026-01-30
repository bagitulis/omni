<template>
  <div class="webhook-settings">
    <div class="section-header">
      <h2>🔗 Platform Integration Settings</h2>
      <p class="description">
        Configure OAuth callback URLs and webhook endpoints for marketplace
        platforms
      </p>
    </div>

    <!-- URL Display Section -->
    <IntegrationUrls
      :baseUrl="baseUrl"
      :tenantId="tenantId"
      @copy="copyToClipboard"
    />

    <!-- Platform Cards -->
    <div class="platforms-grid">
      <PlatformCard
        platform="shopee"
        title="Shopee"
        icon="🛒"
        :status="platformStatus.shopee"
        :callbackUrl="`${baseUrl}/api/platform-auth/shopee/callback`"
        :webhookUrl="
          tenantId
            ? `${baseUrl}/api/webhooks/${tenantId}/shopee`
            : `${baseUrl}/api/webhooks/shopee`
        "
        @authorize="startAuthorization('shopee')"
        @copy="copyToClipboard"
      />

      <PlatformCard
        platform="tiktok"
        title="TikTok Shop"
        icon="🎵"
        :status="platformStatus.tiktok"
        :callbackUrl="`${baseUrl}/api/platform-auth/tiktok/callback`"
        :webhookUrl="
          tenantId
            ? `${baseUrl}/api/webhooks/${tenantId}/tiktok`
            : `${baseUrl}/api/webhooks/tiktok`
        "
        @authorize="startAuthorization('tiktok')"
        @copy="copyToClipboard"
      />

      <PlatformCard
        platform="lazada"
        title="Lazada"
        icon="🏪"
        :status="platformStatus.lazada"
        :callbackUrl="`${baseUrl}/api/platform-auth/lazada/callback`"
        :webhookUrl="
          tenantId
            ? `${baseUrl}/api/webhooks/${tenantId}/lazada`
            : `${baseUrl}/api/webhooks/lazada`
        "
        @authorize="startAuthorization('lazada')"
        @copy="copyToClipboard"
      />
    </div>

    <!-- OAuth Logs Section -->
    <OAuthLogs
      :logs="oauthLogs"
      :loading="oauthLogsLoading"
      @refresh="fetchOAuthLogs"
    />

    <!-- Webhook Logs Section -->
    <WebhookLogs
      :logs="webhookLogs"
      :loading="logsLoading"
      @refresh="fetchWebhookLogs"
    />

    <!-- Toast Notification -->
    <div v-if="toastMessage" class="toast" :class="toastType">
      {{ toastMessage }}
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, computed } from "vue";
import PlatformCard from "./PlatformCard.vue";
import IntegrationUrls from "./IntegrationUrls.vue";
import WebhookLogs from "./WebhookLogs.vue";
import OAuthLogs from "./OAuthLogs.vue";
import { useApi } from "@/composables/useApi";

const api = useApi();

// Base URL
const baseUrl = computed(() => {
  if (typeof window !== "undefined") {
    return window.location.origin;
  }
  return "https://yndigital.my.id";
});

// Get tenant ID from localStorage
const tenantId = computed(() => {
  if (typeof window !== "undefined") {
    return localStorage.getItem("tenantId") || "";
  }
  return "";
});

// Platform connection status
const platformStatus = ref({
  shopee: { connected: false, expiresAt: null as string | null },
  tiktok: { connected: false, expiresAt: null as string | null },
  lazada: { connected: false, expiresAt: null as string | null },
});

// Webhook logs - snake_case to match backend API response
interface LogEntry {
  id: string;
  platform: string;
  event_type: string | null;
  created_at: string;
  status: string;
}
const webhookLogs = ref<LogEntry[]>([]);
const logsLoading = ref(false);

// OAuth logs - snake_case to match backend API response
interface OAuthLog {
  id: string;
  platform: string;
  event_type: string;
  status: string;
  created_at: string;
  processed_at?: string;
  metadata?: Record<string, any>;
}
const oauthLogs = ref<OAuthLog[]>([]);
const oauthLogsLoading = ref(false);

// Toast
const toastMessage = ref("");
const toastType = ref("success");

// Fetch platform status
async function fetchPlatformStatus() {
  try {
    const response = await api.get("/platform-auth/status");
    if (response.success) {
      platformStatus.value = response.data;
    }
  } catch (error) {
    console.error("Failed to fetch platform status:", error);
  }
}

// Fetch webhook logs
async function fetchWebhookLogs() {
  logsLoading.value = true;
  try {
    // Check if tenant is available
    if (!tenantId.value) {
      showToast("Tenant ID not found. Please login again.", "error");
      return;
    }

    const response = await api.get("/webhooks/logs?limit=20");
    if (response.success) {
      webhookLogs.value = response.data || [];
    } else if (
      response.code === "TENANT_REQUIRED" ||
      response.code === "INVALID_TENANT"
    ) {
      showToast(
        response.error || "Invalid tenant. Please login again.",
        "error",
      );
    }
  } catch (error: any) {
    console.error("Failed to fetch webhook logs:", error);
    // Check for specific error codes
    if (error.response?.data?.code === "TENANT_REQUIRED") {
      showToast("Please login to view webhook logs", "error");
    }
  } finally {
    logsLoading.value = false;
  }
}

// Fetch OAuth logs
async function fetchOAuthLogs() {
  oauthLogsLoading.value = true;
  try {
    const response = await api.get("/platform-auth/logs?limit=20");
    if (response.success) {
      oauthLogs.value = response.data || [];
    }
  } catch (error) {
    console.error("Failed to fetch OAuth logs:", error);
  } finally {
    oauthLogsLoading.value = false;
  }
}

// Start OAuth authorization
async function startAuthorization(platform: string) {
  try {
    const response = await api.get(`/platform-auth/${platform}/authorize`);
    if (response.success && response.data?.authUrl) {
      window.open(response.data.authUrl, "_blank", "width=600,height=700");
      showToast(`Opening ${platform} authorization...`, "success");
    } else {
      showToast(`Failed to get authorization URL`, "error");
    }
  } catch (error) {
    console.error("Authorization error:", error);
    showToast(`Authorization failed`, "error");
  }
}

// Copy to clipboard
async function copyToClipboard(text: string) {
  try {
    await navigator.clipboard.writeText(text);
    showToast("Copied to clipboard!", "success");
  } catch {
    showToast("Failed to copy", "error");
  }
}

// Show toast notification
function showToast(message: string, type: "success" | "error") {
  toastMessage.value = message;
  toastType.value = type;
  setTimeout(() => {
    toastMessage.value = "";
  }, 3000);
}

onMounted(() => {
  fetchPlatformStatus();
  fetchWebhookLogs();
  fetchOAuthLogs();
});
</script>

<style scoped>
.webhook-settings {
  padding: 24px;
  max-width: 1200px;
}

.section-header {
  margin-bottom: 24px;
}

.section-header h2 {
  font-size: 1.5rem;
  font-weight: 600;
  color: hsl(var(--bc));
  margin-bottom: 8px;
}

.description {
  color: hsl(var(--bc) / 0.7);
  font-size: 0.95rem;
}

.platforms-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(320px, 1fr));
  gap: 20px;
  margin-bottom: 32px;
}

.toast {
  position: fixed;
  bottom: 24px;
  right: 24px;
  padding: 12px 24px;
  border-radius: 8px;
  font-weight: 500;
  z-index: 1000;
  animation: slideIn 0.3s ease;
}

.toast.success {
  background: hsl(var(--su));
  color: hsl(var(--suc));
}

.toast.error {
  background: hsl(var(--er));
  color: hsl(var(--erc));
}

@keyframes slideIn {
  from {
    transform: translateX(100%);
    opacity: 0;
  }
  to {
    transform: translateX(0);
    opacity: 1;
  }
}
</style>
