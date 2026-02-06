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
        <div class="settings-container">
          <div class="settings-header">
            <h1>Settings</h1>
            <p>Manage application settings and platform integrations</p>
          </div>

          <!-- Tab Content -->
          <div class="tabs-content">
            <!-- Google Sheets Tab -->
            <div v-if="activeTab === 'google-sheets'" class="tab-pane">
              <GoogleSheetsSettings />
            </div>

            <!-- Spreadsheet Registry Tab -->
            <div v-if="activeTab === 'registry'" class="tab-pane">
              <SpreadsheetRegistry />
            </div>

            <!-- General Settings Tab -->
            <div v-if="activeTab === 'general'" class="tab-pane">
              <GeneralSettings />
            </div>

            <!-- Webhook Settings Tab -->
            <div v-if="activeTab === 'webhook'" class="tab-pane">
              <WebhookSettings />
            </div>
          </div>
        </div>
      </main>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch } from "vue";
import { useRoute, useRouter } from "vue-router";
import { useUIStore } from "../store/ui";
import { useAppStore } from "../store/app";
import { useModalsStore } from "../store/modals";
import { usePageHeader } from "@/composables/usePageHeader";

// Layout Components
import LeftSidebar from "@/components/layout/LeftSidebar.vue";

// Settings Components
import GoogleSheetsSettings from "@/components/GoogleSheetsSettings.vue";
import SpreadsheetRegistry from "@/components/SpreadsheetRegistry.vue";
import GeneralSettings from "@/components/content/SettingsContent/GeneralSettings.vue";
import WebhookSettings from "@/components/content/SettingsContent/WebhookSettings.vue";

const route = useRoute();
const router = useRouter();
const uiStore = useUIStore();
const appStore = useAppStore();
const modalsStore = useModalsStore();

// State from stores
// Use page header integration
const { connectionStatus, loading, status, requestCount, errorCount } =
  usePageHeader({
    refreshCallback: async () => await loadStatus(),
    pageTitle: "Settings",
  });

const lastUpdated = ref("");
const activeTab = ref("google-sheets");

// Dashboard methods
const handleTabChange = (tab: any) => {
  uiStore.activeTab = tab;
  router.push({ path: "/" });
};

const handlePlatformChange = (platform: any) => {
  uiStore.activePlatform = platform;
};

const showTokenModal = () => {
  modalsStore.showModal("token", uiStore.activePlatform);
};

const executeOperation = () => {
  // Can be implemented if needed
};

const loadStatus = async () => {
  await appStore.loadStatus(true);
};

// Watch route changes to update activeTab
watch(
  () => route.meta.subsection,
  (newSubsection) => {
    if (newSubsection) {
      activeTab.value = newSubsection as string;
    } else {
      activeTab.value = "google-sheets";
    }
  },
  { immediate: true },
);

onMounted(() => {
  uiStore.activeTab = "settings";

  // Non-blocking: defer init to after first paint
  if (appStore.connectionStatus === "connecting") {
    requestAnimationFrame(() => appStore.initializeApp());
  }

  const subsection = route.meta.subsection as string;
  if (subsection) {
    activeTab.value = subsection;
  }
});
</script>

<style scoped>
.dashboard {
  display: flex;
  flex-direction: column;
  min-height: 100vh;
  width: 100%;
}

.main-layout {
  display: flex;
  flex: 1;
  overflow: hidden;
  position: relative;
}

.main-content {
  flex: 1;
  overflow-y: auto;
  overflow-x: hidden;
  position: relative;
  z-index: 1;
}

.settings-container {
  padding: 24px;
  background: white;
  color: #333;
}

.settings-header {
  margin-bottom: 32px;
  color: #333;
}

.settings-header h1 {
  font-size: 2.5em;
  margin: 0 0 8px 0;
  text-shadow: none;
}

.settings-header p {
  margin: 0;
  opacity: 0.8;
  font-size: 1.1em;
}

/* Tabs Content */
.tabs-content {
  background: rgba(255, 255, 255, 0.98);
  border-radius: 12px;
  box-shadow: 0 2px 12px rgba(59, 130, 246, 0.1);
  overflow: hidden;
}

.tab-pane {
  padding: 32px;
  animation: fadeIn 0.3s ease;
}

@keyframes fadeIn {
  from {
    opacity: 0;
    transform: translateY(10px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

/* Responsive */
@media (max-width: 1024px) {
  .settings-container {
    padding: 16px;
  }

  .settings-header h1 {
    font-size: 1.8em;
  }

  .tab-pane {
    padding: 20px;
  }
}

@media (max-width: 768px) {
  .settings-header h1 {
    font-size: 1.5em;
  }

  .tab-pane {
    padding: 16px;
  }
}
</style>
