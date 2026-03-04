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
            <h1>Script Monitor</h1>
            <p>Monitor all scripts running in the system</p>
          </div>

          <!-- Tab Content -->
          <div class="tabs-content">
            <ScriptMonitorComponent />
          </div>
        </div>
      </main>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from "vue";
import { useRouter } from "vue-router";
import { useUIStore } from "../store/ui";
import { useAppStore } from "../store/app";
import { useModalsStore } from "../store/modals";
import { usePageHeader } from "@/composables/usePageHeader";

// Layout Components
import LeftSidebar from "@/components/layout/LeftSidebar.vue";
import ScriptMonitorComponent from "@/components/content/SettingsContent/ScriptMonitor.vue";

const router = useRouter();
const uiStore = useUIStore();
const appStore = useAppStore();
const modalsStore = useModalsStore();

// Set active tab to script-monitor
uiStore.setActiveTab("script-monitor");

const loadStatus = async () => {
  await appStore.loadStatus(true, true);
};

// Use page header integration
const { connectionStatus, loading, status, requestCount, errorCount } =
  usePageHeader({
    refreshCallback: loadStatus,
    pageTitle: "Script Monitor",
  });

const lastUpdated = ref(new Date().toLocaleString());

onMounted(() => {
  // Non-blocking: defer init to after first paint
  const initAndLoad = async () => {
    if (appStore.connectionStatus === "connecting") {
      await appStore.initializeApp();
    }
    await appStore.loadStatus(true, true);
  };
  
  requestAnimationFrame(() => initAndLoad());
});

const handleTabChange = (tab: string): void => {
  uiStore.setActiveTab(tab as any);
  if (tab !== "script-monitor") {
    router.push("/");
  }
};

const handlePlatformChange = (platform: string): void => {
  uiStore.setActivePlatform(platform as any);
};

const showTokenModal = () => {
  modalsStore.showModal("token", uiStore.activePlatform);
};

const executeOperation = (operationName: string) => {
  appStore.executeOperation(operationName);
};
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
  display: flex;
  flex-direction: column;
}

.settings-container {
  display: flex;
  flex-direction: column;
  gap: 20px;
  padding: 20px;
}

.settings-header {
  border-bottom: 2px solid #e0e0e0;
  padding-bottom: 15px;
}

.settings-header h1 {
  margin: 0 0 8px 0;
  font-size: 28px;
  color: #1f2937;
}

.settings-header p {
  margin: 0;
  color: #6b7280;
  font-size: 14px;
}

.tabs-content {
  flex: 1;
  overflow-y: auto;
}
</style>
