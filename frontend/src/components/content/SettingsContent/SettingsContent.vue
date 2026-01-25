<template>
  <div class="settings-content">
    <div class="settings-header">
      <h1>⚙️ Settings</h1>
      <p>{{ getHeaderText() }}</p>
    </div>

    <div class="tabs-content">
      <div v-if="settingsSubsection === 'google-sheets'" class="tab-pane">
        <GoogleSheetsSettings />
      </div>
      <div v-else-if="settingsSubsection === 'script-monitor'" class="tab-pane">
        <ScriptMonitor />
      </div>
    </div>
  </div>
</template>

<script lang="ts">
import { defineComponent, computed } from "vue";
import GoogleSheetsSettings from "../../GoogleSheetsSettings.vue";
import ScriptMonitor from "./ScriptMonitor.vue";
import { useUIStore } from "@/store/ui";

export default defineComponent({
  name: "SettingsContent",
  components: {
    GoogleSheetsSettings,
    ScriptMonitor,
  },
  props: { loading: { type: Boolean, default: false } },
  emits: ["refresh-all"],
  setup() {
    const uiStore = useUIStore();

    const settingsSubsection = computed(() => uiStore.settingsSubsection);

    const getHeaderText = () => {
      const texts: Record<string, string> = {
        "google-sheets": "Kelola integrasi Google Sheets",
        "script-monitor":
          "Monitor script execution, queue, dan auto-function settings",
      };
      return texts[settingsSubsection.value] || "Settings";
    };

    return { settingsSubsection, getHeaderText };
  },
});
</script>

<style scoped>
.settings-content {
  height: 100%;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.settings-header {
  padding: 16px 20px;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: white;
  border-bottom: 1px solid rgba(255, 255, 255, 0.1);
}

.settings-header h1 {
  font-size: 1.4em;
  margin: 0 0 4px 0;
  text-shadow: 0 1px 2px rgba(0, 0, 0, 0.2);
  font-weight: 600;
}

.settings-header p {
  margin: 0;
  opacity: 0.85;
  font-size: 0.9em;
  font-weight: 400;
}

.tabs-content {
  flex: 1;
  overflow-y: auto;
  background: white;
}

.tab-pane {
  padding: 20px;
}

@media (max-width: 768px) {
  .tab-pane {
    padding: 16px;
  }
}
</style>
