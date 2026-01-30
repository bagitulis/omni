<template>
  <div class="token-dropdown" ref="dropdownRef">
    <button
      @click="toggleDropdown"
      class="token-trigger"
      :class="{ 'has-issues': hasTokenIssues }"
      :title="hasTokenIssues ? 'Some tokens need attention' : 'Token Status'"
      :aria-label="
        hasTokenIssues
          ? 'Token status - some tokens need attention'
          : 'View token status'
      "
      aria-haspopup="true"
      :aria-expanded="isOpen"
    >
      <span class="token-icon">🔑</span>
      <span v-if="hasTokenIssues" class="issue-indicator">!</span>
    </button>

    <Transition name="dropdown">
      <div v-if="isOpen" class="dropdown-panel">
        <div class="dropdown-header">
          <h4>Token Status</h4>
          <button
            @click="refreshStatus"
            :disabled="loading"
            class="btn-refresh-all"
            title="Refresh All Tokens"
            aria-label="Refresh all platform tokens"
          >
            <span :class="{ 'spin-animation': loading }">🔄</span>
            <span class="refresh-text">Refresh</span>
          </button>
        </div>

        <div v-if="tokenStatus" class="token-list">
          <div
            v-for="(statusData, platform) in tokenStatus"
            :key="platform"
            class="token-item"
            :class="getTokenStatusClass(statusData)"
          >
            <div class="token-header">
              <div class="token-platform">
                <PlatformBadge
                  :platform="String(platform)"
                  size="sm"
                  :status="getTokenStatus(statusData)"
                />
                <span class="platform-name">{{
                  String(platform).toUpperCase()
                }}</span>
              </div>
              <span class="status-badge" :class="getTokenStatus(statusData)">
                {{
                  getTokenStatus(statusData) === "valid"
                    ? "✓ Valid"
                    : getTokenStatus(statusData) === "expiring"
                      ? "⚠ Expiring"
                      : "✗ Invalid"
                }}
              </span>
            </div>
            <div class="token-details">
              <div class="detail-row">
                <span class="detail-label">Access Token:</span>
                <span
                  class="detail-value"
                  :class="statusData.isExpired ? 'expired' : 'valid'"
                >
                  {{ statusData.isExpired ? "❌ Expired" : "✅ Valid" }}
                </span>
              </div>
              <div class="detail-row indent">
                <span class="detail-label">Expires in:</span>
                <span class="detail-value">{{
                  formatTimeRemaining(statusData.expiresAt)
                }}</span>
              </div>
              <div class="detail-row">
                <span class="detail-label">Refresh Token:</span>
                <span
                  class="detail-value"
                  :class="isRefreshExpired(statusData) ? 'expired' : 'valid'"
                >
                  {{ isRefreshExpired(statusData) ? "❌ Expired" : "✅ Valid" }}
                </span>
              </div>
              <div class="detail-row indent">
                <span class="detail-label">Expires in:</span>
                <span class="detail-value">{{
                  formatTimeRemaining(statusData.refreshTokenExpiresAt)
                }}</span>
              </div>
            </div>
          </div>
        </div>

        <div v-else class="loading-state">
          <span>{{ loading ? "Loading..." : "Click refresh to load" }}</span>
        </div>
      </div>
    </Transition>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onBeforeUnmount } from "vue";
import PlatformBadge from "@/components/PlatformBadge.vue";
import apiService from "@/services/api";
import "./TokenStatusDropdown.styles.css";

const dropdownRef = ref<HTMLElement | null>(null);
const isOpen = ref(false);
const loading = ref(false);
const tokenStatus = ref<Record<string, any> | null>(null);

const hasTokenIssues = computed(() => {
  if (!tokenStatus.value) return false;
  return Object.values(tokenStatus.value).some((s: any) => {
    if (s.isExpired === true) return true;
    if (s.expiresAt) {
      const expiresAt = new Date(s.expiresAt).getTime();
      const hoursLeft = (expiresAt - Date.now()) / (1000 * 60 * 60);
      if (hoursLeft < 24 && hoursLeft > 0) return true;
    }
    return false;
  });
});

const toggleDropdown = async () => {
  isOpen.value = !isOpen.value;
  if (isOpen.value && !tokenStatus.value) {
    await loadTokenStatus();
  }
};

const loadTokenStatus = async () => {
  loading.value = true;
  try {
    const response = await apiService.getTokenStatus();
    if (response.status === "success" && response.data) {
      tokenStatus.value = response.data;
    }
  } catch (error) {
    console.error("Failed to load token status:", error);
  } finally {
    loading.value = false;
  }
};

const refreshStatus = async () => {
  loading.value = true;
  try {
    // Trigger refresh all tokens on backend
    await apiService.refreshAllTokens();
    // Then reload the status
    await loadTokenStatus();
  } catch (error) {
    console.error("Failed to refresh tokens:", error);
  } finally {
    loading.value = false;
  }
};

const getTokenStatusClass = (statusData: any) => {
  if (!statusData) return "unknown";
  // Check expired - API returns isExpired boolean
  if (
    statusData.isExpired === true ||
    statusData.status === "expired" ||
    statusData.valid === false
  ) {
    return "expired";
  }
  // Check expiring soon (within 24 hours)
  if (statusData.expiresAt) {
    const expiresAt = new Date(statusData.expiresAt).getTime();
    const now = Date.now();
    const hoursLeft = (expiresAt - now) / (1000 * 60 * 60);
    if (hoursLeft < 24 && hoursLeft > 0) return "expiring";
  }
  if (statusData.status === "expiring") return "expiring";
  return "valid";
};

const getTokenStatus = (
  statusData: any,
): "valid" | "expiring" | "expired" | "unknown" => {
  if (!statusData) return "unknown";
  // Check expired - API returns isExpired boolean
  if (
    statusData.isExpired === true ||
    statusData.status === "expired" ||
    statusData.valid === false
  ) {
    return "expired";
  }
  // Check expiring soon (within 24 hours)
  if (statusData.expiresAt) {
    const expiresAt = new Date(statusData.expiresAt).getTime();
    const now = Date.now();
    const hoursLeft = (expiresAt - now) / (1000 * 60 * 60);
    if (hoursLeft < 24 && hoursLeft > 0) return "expiring";
  }
  if (statusData.status === "expiring") return "expiring";
  return "valid";
};

const formatTimeRemaining = (dateString: string | null | undefined): string => {
  if (!dateString) return "Unknown";

  const expiresAt = new Date(dateString);
  const now = new Date();
  const diffMs = expiresAt.getTime() - now.getTime();

  if (diffMs <= 0) return "Expired";

  const days = Math.floor(diffMs / (1000 * 60 * 60 * 24));
  const hours = Math.floor((diffMs % (1000 * 60 * 60 * 24)) / (1000 * 60 * 60));
  const minutes = Math.floor((diffMs % (1000 * 60 * 60)) / (1000 * 60));

  return `${days}d ${hours}h ${minutes}m`;
};

const isRefreshExpired = (statusData: any): boolean => {
  if (!statusData?.refreshTokenExpiresAt) return false;
  return new Date(statusData.refreshTokenExpiresAt).getTime() < Date.now();
};

// Click outside to close
const handleClickOutside = (e: MouseEvent) => {
  if (dropdownRef.value && !dropdownRef.value.contains(e.target as Node)) {
    isOpen.value = false;
  }
};

onMounted(() => {
  document.addEventListener("click", handleClickOutside);
});

onBeforeUnmount(() => {
  document.removeEventListener("click", handleClickOutside);
});
</script>
