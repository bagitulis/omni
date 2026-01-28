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
          <div v-if="selectedReport" class="modal-overlay" @click="closeReport">
            <div class="modal-content" @click.stop>
              <div class="modal-header">
                <h2>
                  {{ selectedReport.period_label || selectedReport.file_name }}
                </h2>
                <button @click="closeReport" class="btn-close" type="button">
                  ✕
                </button>
              </div>
              <div class="modal-body">
                <div v-if="loadingReport" class="loading-report">
                  <div class="spinner"></div>
                  <p>Loading report...</p>
                </div>
                <div
                  v-else-if="reportHTML"
                  class="report-content"
                  v-html="reportHTML"
                ></div>
                <div v-else class="error-report">
                  <span aria-hidden="true">❌</span>
                  <p>Failed to load report</p>
                </div>
              </div>
            </div>
          </div>
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

// Local state
const selectedPlatform = ref<"all" | "shopee" | "tiktok">("all");
const selectedReport = ref<any>(null);
const loadingReport = ref(false);

// Computed
const filteredReports = computed(() => {
  if (selectedPlatform.value === "all") {
    return reports.value;
  }
  return reports.value.filter(
    (report) => report.platform === selectedPlatform.value,
  );
});

// Navigation handlers
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

// Report handlers
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

// Lifecycle
onMounted(async () => {
  uiStore.setActiveTab("analytics");

  // Initialize app connection if not already connected
  if (appStore.connectionStatus !== "connected") {
    appStore.initializeApp();
  }

  await loadReports();
});
</script>

<style scoped>
.dashboard {
  display: flex;
  flex-direction: column;
  height: 100vh;
  overflow: hidden;
}

.main-layout {
  display: flex;
  flex: 1;
  overflow: hidden;
}

.main-content {
  flex: 1;
  overflow-y: auto;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
}

.gallery-page {
  padding: 2rem;
  max-width: 1400px;
  margin: 0 auto;
}

/* Header */
.gallery-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 2rem;
  padding: 1.5rem;
  background: white;
  border-radius: 12px;
  box-shadow: 0 4px 6px rgba(0, 0, 0, 0.1);
}

.header-left h1 {
  margin: 0;
  font-size: 2rem;
  color: #2d3748;
}

.subtitle {
  margin: 0.5rem 0 0;
  color: #718096;
  font-size: 1rem;
}

.btn-generate {
  padding: 0.75rem 1.5rem;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: white;
  border: none;
  border-radius: 8px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.3s ease;
}

.btn-generate:hover:not(:disabled) {
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(102, 126, 234, 0.4);
}

.btn-generate:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

/* Filter Section */
.filter-section {
  margin-bottom: 2rem;
  padding: 1rem;
  background: white;
  border-radius: 12px;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
}

.filter-group {
  display: flex;
  align-items: center;
  gap: 1rem;
}

.filter-group label {
  font-weight: 600;
  color: #4a5568;
}

.platform-select {
  padding: 0.5rem 1rem;
  border: 2px solid #e2e8f0;
  border-radius: 8px;
  font-size: 1rem;
  cursor: pointer;
  transition: border-color 0.3s ease;
}

.platform-select:focus {
  outline: none;
  border-color: #667eea;
}

/* States */
.loading-state,
.error-state,
.empty-state {
  text-align: center;
  padding: 4rem 2rem;
  background: white;
  border-radius: 12px;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
}

.spinner {
  width: 40px;
  height: 40px;
  margin: 0 auto 1rem;
  border: 4px solid #e2e8f0;
  border-top-color: #667eea;
  border-radius: 50%;
  animation: spin 1s linear infinite;
}

@keyframes spin {
  to {
    transform: rotate(360deg);
  }
}

.empty-subtitle {
  color: #a0aec0;
  margin-top: 0.5rem;
}

.btn-retry {
  margin-top: 1rem;
  padding: 0.5rem 1.5rem;
  background: #667eea;
  color: white;
  border: none;
  border-radius: 8px;
  cursor: pointer;
  transition: background 0.3s ease;
}

.btn-retry:hover {
  background: #5568d3;
}

/* Reports Grid */
.reports-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(320px, 1fr));
  gap: 1.5rem;
}

.report-card {
  background: white;
  border-radius: 12px;
  overflow: hidden;
  box-shadow: 0 4px 6px rgba(0, 0, 0, 0.1);
  transition: all 0.3s ease;
  cursor: pointer;
}

.report-card:hover {
  transform: translateY(-4px);
  box-shadow: 0 8px 16px rgba(0, 0, 0, 0.15);
}

.card-header {
  display: flex;
  justify-content: space-between;
  padding: 1rem;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: white;
}

.platform-badge {
  padding: 0.25rem 0.75rem;
  background: rgba(255, 255, 255, 0.2);
  border-radius: 12px;
  font-size: 0.875rem;
  font-weight: 600;
  text-transform: capitalize;
}

.report-type {
  font-size: 0.875rem;
  opacity: 0.9;
}

.card-body {
  padding: 1.5rem;
}

.card-body h3 {
  margin: 0 0 1rem;
  font-size: 1.25rem;
  color: #2d3748;
}

.card-meta {
  display: flex;
  gap: 1rem;
}

.meta-item {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  font-size: 0.875rem;
  color: #718096;
}

.card-actions {
  display: flex;
  gap: 0.5rem;
  padding: 1rem;
  border-top: 1px solid #e2e8f0;
}

.btn-view {
  flex: 1;
  padding: 0.5rem 1rem;
  background: #667eea;
  color: white;
  border: none;
  border-radius: 8px;
  font-weight: 500;
  cursor: pointer;
  transition: background 0.3s ease;
}

.btn-view:hover {
  background: #5568d3;
}

.btn-download {
  padding: 0.5rem 1rem;
  background: #48bb78;
  color: white;
  border: none;
  border-radius: 8px;
  cursor: pointer;
  transition: background 0.3s ease;
}

.btn-download:hover {
  background: #38a169;
}

/* Modal */
.modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.7);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
  padding: 2rem;
}

.modal-content {
  background: white;
  border-radius: 12px;
  max-width: 1200px;
  width: 100%;
  max-height: 90vh;
  display: flex;
  flex-direction: column;
  box-shadow: 0 20px 60px rgba(0, 0, 0, 0.3);
}

.modal-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 1.5rem;
  border-bottom: 1px solid #e2e8f0;
}

.modal-header h2 {
  margin: 0;
  font-size: 1.5rem;
  color: #2d3748;
}

.btn-close {
  padding: 0.5rem 1rem;
  background: #f56565;
  color: white;
  border: none;
  border-radius: 8px;
  font-size: 1.25rem;
  font-weight: bold;
  cursor: pointer;
  transition: background 0.3s ease;
}

.btn-close:hover {
  background: #e53e3e;
}

.modal-body {
  flex: 1;
  overflow-y: auto;
  padding: 2rem;
}

.loading-report,
.error-report {
  text-align: center;
  padding: 4rem 2rem;
}

.report-content {
  line-height: 1.6;
}

.report-content :deep(img) {
  max-width: 100%;
  height: auto;
  border-radius: 8px;
  margin: 1rem 0;
}

.report-content :deep(table) {
  width: 100%;
  border-collapse: collapse;
  margin: 1rem 0;
}

.report-content :deep(th),
.report-content :deep(td) {
  padding: 0.75rem;
  border: 1px solid #e2e8f0;
  text-align: left;
}

.report-content :deep(th) {
  background: #f7fafc;
  font-weight: 600;
}
</style>
