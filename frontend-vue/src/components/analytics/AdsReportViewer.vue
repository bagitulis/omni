<template>
  <div class="report-viewer">
    <div class="viewer-header">
      <h2 class="viewer-title">
        <span aria-hidden="true">🤖</span> AI Insights
      </h2>
      <div class="report-selector">
        <label for="report-select" class="sr-only">Select report</label>
        <select
          id="report-select"
          v-model="selectedReport"
          class="report-select"
        >
          <option value="">Select Report</option>
          <option v-for="r in reports" :key="r.filename" :value="r.filename">
            {{ formatReportName(r.filename) }}
          </option>
        </select>
        <button
          type="button"
          class="btn-refresh"
          @click="fetchReports"
          aria-label="Refresh report list"
        >
          <span aria-hidden="true">🔄</span>
        </button>
      </div>
    </div>

    <!-- Loading State -->
    <div v-if="loading" class="loading-state">
      <span aria-hidden="true">⏳</span> Loading reports...
    </div>

    <!-- Empty State -->
    <div v-else-if="reports.length === 0" class="empty-state">
      <span aria-hidden="true">📊</span>
      <p>No AI reports available yet</p>
      <p class="empty-hint">Run the analysis notebook to generate insights</p>
    </div>

    <!-- Report Embed -->
    <div v-else-if="selectedReport" class="report-frame-container">
      <iframe
        :src="reportUrl"
        class="report-frame"
        :title="`${platform} AI Insights Report`"
        sandbox="allow-same-origin allow-scripts"
      ></iframe>
    </div>

    <!-- Select Prompt -->
    <div v-else class="select-prompt">
      <span aria-hidden="true">👆</span>
      <p>Select a report from the dropdown above</p>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch, onMounted } from "vue";
import { useAdsReports, type ReportInfo } from "@/composables/useAdsReports";

type Platform = "shopee" | "tiktok";

const props = defineProps<{
  platform: Platform;
}>();

const {
  loading,
  shopeeReports,
  tiktokReports,
  fetchShopeeReports,
  fetchTiktokReports,
  getReportFileUrl,
} = useAdsReports();

const selectedReport = ref("");

// Computed
const reports = computed<ReportInfo[]>(() => {
  return props.platform === "shopee"
    ? shopeeReports.value
    : tiktokReports.value;
});

const reportUrl = computed(() => {
  if (!selectedReport.value) return "";
  return getReportFileUrl(props.platform, selectedReport.value);
});

// Methods
function fetchReports() {
  if (props.platform === "shopee") {
    fetchShopeeReports();
  } else {
    fetchTiktokReports();
  }
}

function formatReportName(filename: string): string {
  // Convert "executive_report_20250615.html" to "Executive Report - Jun 15, 2025"
  const match = filename.match(/(\d{4})(\d{2})(\d{2})/);
  if (match) {
    const date = new Date(`${match[1]}-${match[2]}-${match[3]}`);
    const dateStr = date.toLocaleDateString("en-US", {
      month: "short",
      day: "numeric",
      year: "numeric",
    });
    const name = filename
      .replace(/_\d{8}\.html$/, "")
      .replace(/_/g, " ")
      .replace(/\b\w/g, (c) => c.toUpperCase());
    return `${name} - ${dateStr}`;
  }
  return filename;
}

// Auto-select latest report
watch(reports, (newReports) => {
  if (newReports.length > 0 && !selectedReport.value) {
    selectedReport.value = newReports[0].filename;
  }
});

onMounted(() => {
  fetchReports();
});
</script>

<style scoped>
.report-viewer {
  display: flex;
  flex-direction: column;
  height: 100%;
  min-height: 600px;
}

.viewer-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 16px 20px;
  background: white;
  border-bottom: 1px solid #e5e7eb;
}

.viewer-title {
  font-size: 16px;
  color: #1f2937;
  margin: 0;
  display: flex;
  align-items: center;
  gap: 8px;
}

.report-selector {
  display: flex;
  gap: 8px;
}

.sr-only {
  position: absolute;
  width: 1px;
  height: 1px;
  padding: 0;
  margin: -1px;
  overflow: hidden;
  clip: rect(0, 0, 0, 0);
  white-space: nowrap;
  border: 0;
}

.report-select {
  padding: 8px 12px;
  border: 1px solid #d1d5db;
  border-radius: 6px;
  font-size: 14px;
  min-width: 250px;
}

.btn-refresh {
  padding: 8px 12px;
  border: 1px solid #d1d5db;
  border-radius: 6px;
  background: white;
  cursor: pointer;
}

.btn-refresh:hover {
  background: #f9fafb;
}

.loading-state,
.empty-state,
.select-prompt {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  color: #6b7280;
  gap: 8px;
  font-size: 24px;
}

.empty-state p,
.select-prompt p {
  font-size: 14px;
  margin: 0;
}

.empty-hint {
  font-size: 12px;
  color: #9ca3af;
}

.report-frame-container {
  flex: 1;
  padding: 0;
}

.report-frame {
  width: 100%;
  height: 100%;
  min-height: 600px;
  border: none;
  background: white;
}
</style>
