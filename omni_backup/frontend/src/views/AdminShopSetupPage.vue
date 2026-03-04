<template>
  <div class="shop-setup-page">
    <h1 class="visually-hidden">Shop Setup</h1>

    <!-- Global Credentials Section -->
    <div class="section credentials-section">
      <div class="section-header">
        <h2>🔧 App Credentials</h2>
        <p class="subtitle">
          Credentials dari Shopee/TikTok Open Platform (shared untuk semua toko)
        </p>
      </div>

      <!-- Shopee Credentials -->
      <div class="platform-card">
        <div class="platform-header">
          <span class="platform-icon" aria-hidden="true">🛒</span>
          <h3>Shopee</h3>
          <span
            v-if="credentials.shopee.hasPartnerId"
            class="status-badge success"
          >
            ✓ Configured
          </span>
          <span v-else class="status-badge warning">⚠ Not Configured</span>
        </div>

        <div class="credentials-form" v-if="isDeveloper">
          <div class="form-group">
            <label>Partner ID</label>
            <input
              v-model="shopeeForm.partnerId"
              type="text"
              placeholder="Masukkan Partner ID"
              :disabled="saving"
            />
          </div>
          <div class="form-group">
            <label>Partner Key</label>
            <input
              v-model="shopeeForm.partnerKey"
              type="password"
              placeholder="Masukkan Partner Key (encrypted)"
              :disabled="saving"
            />
          </div>
          <div class="form-group">
            <label>Push Partner Key (Optional)</label>
            <input
              v-model="shopeeForm.pushPartnerKey"
              type="password"
              placeholder="Untuk webhook signature verification"
              :disabled="saving"
            />
          </div>
          <button
            @click="saveShopeeCredentials"
            class="save-btn"
            :disabled="saving"
          >
            {{ saving ? "Menyimpan..." : "💾 Simpan Credentials" }}
          </button>
        </div>
        <div v-else class="readonly-notice">
          <p>🔒 Hanya Developer yang dapat mengubah app credentials</p>
          <p v-if="credentials.shopee.partnerId">
            Partner ID: {{ credentials.shopee.partnerId }}
          </p>
        </div>
      </div>
    </div>

    <!-- Shop Connection Section -->
    <div class="section connection-section">
      <div class="section-header">
        <h2>🏪 Koneksi Toko</h2>
        <p class="subtitle">Status koneksi toko Anda ke platform marketplace</p>
      </div>

      <!-- Shopee Connection -->
      <div class="connection-card">
        <div class="connection-header">
          <span class="platform-icon">🛒</span>
          <div class="connection-info">
            <h3>Shopee</h3>
            <p v-if="shopStatus.shopee.connected" class="shop-id">
              Shop ID: {{ shopStatus.shopee.shopId }}
            </p>
          </div>
          <span
            :class="[
              'status-badge',
              shopStatus.shopee.connected ? 'success' : 'error',
            ]"
          >
            {{
              shopStatus.shopee.connected ? "✓ Terhubung" : "✗ Belum Terhubung"
            }}
          </span>
        </div>

        <div class="connection-details" v-if="shopStatus.shopee.connected">
          <p>
            Token:
            <span :class="shopStatus.shopee.tokenValid ? 'valid' : 'expired'">
              {{ shopStatus.shopee.tokenValid ? "Valid" : "Expired" }}
            </span>
          </p>
          <p v-if="shopStatus.shopee.expiresAt">
            Expires: {{ formatDate(shopStatus.shopee.expiresAt) }}
          </p>
        </div>

        <div class="connection-actions">
          <button
            v-if="!shopStatus.shopee.connected"
            @click="authorizeShopee"
            class="authorize-btn"
            :disabled="!credentials.shopee.hasPartnerId"
          >
            🔗 Hubungkan Toko Shopee
          </button>
          <template v-else>
            <button @click="authorizeShopee" class="reconnect-btn">
              🔄 Reconnect
            </button>
            <button @click="disconnectShopee" class="disconnect-btn">
              ❌ Disconnect
            </button>
          </template>
        </div>

        <p v-if="!credentials.shopee.hasPartnerId" class="warning-text">
          ⚠ Credentials belum dikonfigurasi. Hubungi Owner untuk setup.
        </p>
      </div>

      <!-- Webhook URL Info -->
      <div class="webhook-info">
        <h4>📍 Webhook URL</h4>
        <div class="url-display">
          <code>{{ webhookUrl }}</code>
          <button
            @click="copyWebhookUrl"
            class="copy-btn"
            type="button"
            aria-label="Copy webhook URL"
          >
            <span aria-hidden="true">📋</span>
          </button>
        </div>
        <p class="hint">
          Daftarkan URL ini di Shopee Open Platform → App Settings → Push
        </p>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from "vue";
import { useApi } from "@/composables/useApi";

const api = useApi();

const userRole = localStorage.getItem("userRole")?.toLowerCase() || "user";
const tenantId = localStorage.getItem("tenantId") || "";
// Developer can edit credentials (they own the Shopee Open Platform account)
// Owner can only connect/authorize their shop
const isDeveloper = computed(() => userRole === "developer");

const saving = ref(false);
const loading = ref(true);

// Credentials status (masked)
const credentials = ref({
  shopee: {
    hasPartnerId: false,
    hasPartnerKey: false,
    hasPushPartnerKey: false,
    partnerId: "",
  },
  tiktok: {
    hasAppKey: false,
    hasAppSecret: false,
    appKey: "",
  },
});

// Shop connection status
const shopStatus = ref({
  shopee: {
    connected: false,
    shopId: null as string | null,
    tokenValid: false,
    expiresAt: null as string | null,
  },
});

// Form data
const shopeeForm = ref({
  partnerId: "",
  partnerKey: "",
  pushPartnerKey: "",
});

const baseUrl = computed(() => {
  return window.location.origin;
});

const webhookUrl = computed(() => {
  return `${baseUrl.value}/api/webhooks/${tenantId}/shopee`;
});

async function loadStatus() {
  loading.value = true;
  try {
    const [credRes, shopRes] = await Promise.all([
      api.get("/api/admin/shop-setup/credentials"),
      api.get("/api/admin/shop-setup/shop-status"),
    ]);

    if (credRes.success) {
      credentials.value = credRes.data;
      shopeeForm.value.partnerId = credRes.data.shopee.partnerId || "";
    }
    if (shopRes.success) {
      shopStatus.value = shopRes.data;
    }
  } catch (error) {
    console.error("Failed to load status:", error);
  } finally {
    loading.value = false;
  }
}

async function saveShopeeCredentials() {
  saving.value = true;
  try {
    await api.post("/api/admin/shop-setup/credentials", {
      platform: "shopee",
      partnerId: shopeeForm.value.partnerId,
      partnerKey: shopeeForm.value.partnerKey || undefined,
      pushPartnerKey: shopeeForm.value.pushPartnerKey || undefined,
    });
    await loadStatus();
    alert("Credentials saved!");
  } catch (error) {
    console.error("Failed to save:", error);
    alert("Failed to save credentials");
  } finally {
    saving.value = false;
  }
}

function authorizeShopee() {
  // Redirect to OAuth flow
  window.location.href = `/api/platform-auth/shopee/authorize`;
}

async function disconnectShopee() {
  if (!confirm("Disconnect Shopee?")) return;
  try {
    await api.delete("/api/admin/shop-setup/disconnect/shopee");
    await loadStatus();
  } catch (error) {
    console.error("Failed to disconnect:", error);
  }
}

function copyWebhookUrl() {
  navigator.clipboard.writeText(webhookUrl.value);
  alert("URL copied!");
}

function formatDate(dateStr: string) {
  return new Date(dateStr).toLocaleString("id-ID");
}

onMounted(() => {
  loadStatus();
});
</script>

<style scoped>
@import "./styles/AdminShopSetup.styles.css";
</style>
