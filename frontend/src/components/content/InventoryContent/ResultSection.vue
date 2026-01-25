<template>
  <div class="result-section" v-if="visible && result">
    <div :class="['result-banner', result.success ? 'success' : 'error']">
      <span class="result-icon">{{ result.success ? "✅" : "❌" }}</span>
      <div class="result-content">
        <strong>{{
          result.success ? "Update Berhasil" : "Update Gagal"
        }}</strong>
        <p class="result-message">{{ resultMessage }}</p>
      </div>
      <button
        class="btn-close"
        type="button"
        aria-label="Close result"
        @click="emit('close')"
      >
        ✕
      </button>
    </div>

    <!-- Detailed Results -->
    <div v-if="hasMultipleResults" class="result-details">
      <div
        v-for="(platformResult, idx) in result.results"
        :key="idx"
        class="platform-result"
      >
        <h3 class="platform-title">
          <span :class="['platform-badge', platformResult.platform]">
            {{ platformResult.platform?.toUpperCase() }}
          </span>
        </h3>
        <div class="stats">
          <span class="stat success"
            >✓ {{ platformResult.processed || 0 }} berhasil</span
          >
          <span class="stat failed"
            >✗ {{ platformResult.failed || 0 }} gagal</span
          >
        </div>

        <!-- Failed SKUs -->
        <div
          v-if="getFailedSkus(platformResult).length > 0"
          class="failed-list"
        >
          <p class="failed-title">SKU Gagal:</p>
          <ul>
            <li
              v-for="sku in getFailedSkus(platformResult).slice(0, 5)"
              :key="sku"
            >
              {{ sku }}
            </li>
            <li
              v-if="getFailedSkus(platformResult).length > 5"
              class="more-note"
            >
              ... +{{ getFailedSkus(platformResult).length - 5 }} lainnya
            </li>
          </ul>
        </div>
      </div>
    </div>

    <!-- Single Result (for wholesale) -->
    <div v-else-if="result.data" class="result-details">
      <div class="stats">
        <span class="stat success"
          >✓ {{ result.data.processed || 0 }} berhasil</span
        >
        <span class="stat failed">✗ {{ result.data.failed || 0 }} gagal</span>
      </div>
      <div v-if="getDataFailedSkus().length > 0" class="failed-list">
        <p class="failed-title">SKU Gagal:</p>
        <ul>
          <li v-for="sku in getDataFailedSkus().slice(0, 5)" :key="sku">
            {{ sku }}
          </li>
          <li v-if="getDataFailedSkus().length > 5" class="more-note">
            ... +{{ getDataFailedSkus().length - 5 }} lainnya
          </li>
        </ul>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from "vue";

interface PlatformResultItem {
  platform: string;
  processed?: number;
  failed?: number;
  failedSkus?: string[];
}

interface ResultData {
  success: boolean;
  message?: string;
  results?: PlatformResultItem[];
  data?: {
    processed?: number;
    failed?: number;
    failedSkus?: string[];
  };
}

const props = defineProps<{
  result: ResultData | null;
  visible: boolean;
}>();

const emit = defineEmits<{
  close: [];
}>();

const hasMultipleResults = computed(() => {
  return props.result?.results && props.result.results.length > 0;
});

function getFailedSkus(platformResult: PlatformResultItem): string[] {
  return platformResult.failedSkus || [];
}

function getDataFailedSkus(): string[] {
  return props.result?.data?.failedSkus || [];
}

const resultMessage = computed(() => {
  if (!props.result) return "";
  if (props.result.message) return props.result.message;

  if (props.result.results) {
    const total = props.result.results.reduce(
      (acc, r) => acc + (r.processed || 0),
      0
    );
    return `${total} item berhasil diproses`;
  }

  if (props.result.data) {
    return `${props.result.data.processed || 0} item berhasil diproses`;
  }

  return props.result.success ? "Operasi selesai" : "Terjadi kesalahan";
});
</script>

<style scoped>
.result-section {
  margin-top: 16px;
}

.result-banner {
  display: flex;
  align-items: flex-start;
  gap: 12px;
  padding: 12px 16px;
  border-radius: 8px;
}

.result-banner.success {
  background: #e8f5e9;
  border: 1px solid #a5d6a7;
}
.result-banner.error {
  background: #ffebee;
  border: 1px solid #ef9a9a;
}

.result-icon {
  font-size: 24px;
}

.result-content {
  flex: 1;
}
.result-content strong {
  display: block;
  margin-bottom: 4px;
}
.result-message {
  font-size: 13px;
  color: #666;
  margin: 0;
}

.btn-close {
  background: transparent;
  border: none;
  font-size: 16px;
  cursor: pointer;
  color: #6b7280;
}

.btn-close:hover {
  color: #333;
}

.result-details {
  margin-top: 12px;
  padding: 12px;
  background: #fafafa;
  border-radius: 8px;
}

.platform-result {
  margin-bottom: 12px;
}
.platform-result:last-child {
  margin-bottom: 0;
}

.platform-title {
  font-size: 13px;
  margin: 0 0 8px;
}

.platform-badge {
  padding: 2px 8px;
  border-radius: 4px;
  font-size: 11px;
  color: white;
}

.platform-badge.shopee {
  background: #ee4d2d;
}
.platform-badge.tiktok {
  background: #000;
}

.stats {
  display: flex;
  gap: 16px;
}

.stat {
  font-size: 13px;
  font-weight: 500;
}

.stat.success {
  color: #2e7d32;
}
.stat.failed {
  color: #c62828;
}

.failed-list {
  margin-top: 8px;
  padding: 8px;
  background: #fff3e0;
  border-radius: 4px;
}

.failed-title {
  font-size: 12px;
  font-weight: 600;
  color: #e65100;
  margin: 0 0 4px;
}

.failed-list ul {
  margin: 0;
  padding-left: 20px;
  font-size: 12px;
  font-family: monospace;
}

.more-note {
  font-style: italic;
  color: #6b7280; /* Improved from #888 for WCAG AA contrast */
}
</style>
