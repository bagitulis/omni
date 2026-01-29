<template>
  <div class="dashboard">
    <div class="main-layout">
      <!-- Left Sidebar -->
      <LeftSidebar
        :collapsed="uiStore.leftSidebarCollapsed"
        :active-tab="uiStore.activeTab"
        :active-platform="uiStore.activePlatform"
        @toggle="uiStore.toggleLeftSidebar"
        @tab-change="handleTabChange"
        @platform-change="handlePlatformChange"
      />

      <!-- Main Content -->
      <main class="main-content">
        <div class="gallery-page">
          <!-- Header -->
          <div class="gallery-header">
            <div class="header-left">
              <h1><span aria-hidden="true">🤖</span> AI Report Gallery</h1>
              <p class="subtitle">ML-Generated Analytics Reports</p>
            </div>
            <div class="header-right">
              <button
                class="btn-generate"
                :disabled="generating"
                @click="handleGenerateReport"
                type="button"
              >
                <span v-if="generating" aria-hidden="true">⏳</span>
                <span v-else aria-hidden="true">✨</span>
                {{ generating ? "Generating..." : "Generate New Report" }}
              </button>
            </div>
          </div>

          <!-- Platform Filter -->
          <div class="filter-section">
            <div class="filter-group">
              <label>Platform:</label>
              <select v-model="selectedPlatform" class="platform-select">
                <option value="all">All Platforms</option>
                <option value="shopee">Shopee</option>
                <option value="tiktok">TikTok</option>
              </select>
            </div>
          </div>

          <!-- Loading State -->
          <div v-if="loading" class="loading-state">
            <div class="spinner"></div>
            <p>Loading reports...</p>
          </div>

          <!-- Error State -->
          <div v-else-if="error" class="error-state">
            <span aria-hidden="true">❌</span>
            <p>{{ error }}</p>
            <button @click="loadReports" class="btn-retry" type="button">
              Retry
            </button>
          </div>

          <!-- Empty State -->
          <div v-else-if="filteredReports.length === 0" class="empty-state">
            <span aria-hidden="true">📭</span>
            <p>No reports found</p>
            <p class="empty-subtitle">
              Generate your first AI report to get started
            </p>
          </div>

          <!-- Reports Grid -->
          <div v-else class="reports-grid">
            <div
              v-for="report in filteredReports"
              :key="report.id"
              class="report-card"
              @click="viewReport(report)"
            >
              <div class="card-header">
                <span class="platform-badge" :class="report.platform">
                  {{ report.platform }}
                </span>
                <span class="report-type">{{ report.report_type }}</span>
              </div>
              <div class="card-body">
                <h3>{{ report.period_label || report.file_name }}</h3>
                <div class="card-meta">
                  <div class="meta-item">
                    <span aria-hidden="true">📅</span>
                    <span>{{ formatDate(report.created_at) }}</span>
                  </div>
                  <div class="meta-item">
                    <span aria-hidden="true">📊</span>
                    <span>{{ formatFileSize(report.file_size) }}</span>
                  </div>
                </div>
              </div>
              <div class="card-actions">
                <button
                  @click.stop="viewReport(report)"
                  class="btn-view"
                  type="button"
                >
                  View Report
                </button>
                <button
                  @click.stop="downloadReportFile(report)"
                  class="btn-download"
                  type="button"
                >
                  <span aria-hidden="true">⬇️</span>
                </button>
              </div>
            </div>
          </div>

          <!-- Report Viewer Modal -->
          <ReportModal
            :report="selectedReport"
            :loading="loadingReport"
            :html="reportHTML"
            @close="closeReport"
          />
        </div>
      </main>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from "vue";
import { useRouter } from "vue-router";
import { useUIStore } from "@/store/ui";
import { useAppStore } from "@/store/app";
import { useMLReports } from "@/composables/useMLReports";
import LeftSidebar from "@/components/layout/LeftSidebar.vue";
import ReportModal from "./ReportModal.vue";

const router = useRouter();
const uiStore = useUIStore();
const appStore = useAppStore();

const {
  loading,
  generating,
  error,
  reports,
  reportHTML,
  fetchReports,
  generateReport,
  fetchReport,
  downloadReport: downloadReportFile,
  formatFileSize,
  formatDate,
} = useMLReports();

const selectedPlatform = ref<"all" | "shopee" | "tiktok">("all");
const selectedReport = ref<any>(null);
const loadingReport = ref(false);

const filteredReports = computed(() => {
  if (selectedPlatform.value === "all") {
    return reports.value;
  }
  return reports.value.filter(
    (report) => report.platform === selectedPlatform.value,
  );
});

type TabType =
  | "logs"
  | "settings"
  | "operation"
  | "product-management"
  | "order-management"
  | "inventory"
  | "script-monitor"
  | "analytics";
type PlatformType = "shopee" | "lazada" | "tiktok";

function handleTabChange(tab: string) {
  uiStore.setActiveTab(tab as TabType);
  if (tab !== "analytics") {
    router.push(`/${tab}`);
  }
}

function handlePlatformChange(platform: string) {
  uiStore.setActivePlatform(platform as PlatformType);
}

async function handleGenerateReport() {
  const platform =
    selectedPlatform.value === "all" ? "shopee" : selectedPlatform.value;
  try {
    await generateReport({
      platform: platform as "shopee" | "tiktok",
      report_type: "full",
    });
    await loadReports();
  } catch (err) {
    console.error("Failed to generate report:", err);
  }
}

async function viewReport(report: any) {
  selectedReport.value = report;
  loadingReport.value = true;
  try {
    await fetchReport(report.platform, report.file_name);
  } catch (err) {
    console.error("Failed to load report:", err);
  }
  loadingReport.value = false;
}

function closeReport() {
  selectedReport.value = null;
}

async function loadReports() {
  const platform =
    selectedPlatform.value === "all" ? "tiktok" : selectedPlatform.value;
  try {
    await fetchReports(platform as "shopee" | "tiktok");
  } catch (err) {
    console.error("Failed to load reports:", err);
  }
}

onMounted(async () => {
  uiStore.setActiveTab("analytics");

  if (appStore.connectionStatus !== "connected") {
    appStore.initializeApp();
  }

  await loadReports();
});
</script>

<style scoped>
@import "./AIReportGallery.styles.css";
</style>
