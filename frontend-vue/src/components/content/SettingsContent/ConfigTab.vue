<template>
  <div v-if="activeTab === 'config'" class="tab-content">
    <!-- Manual Trigger Section -->
    <ManualTriggerSection
      :configs="routeExecutionConfigs"
      @show-add-new="
        showRouteModal = true;
        editingRouteConfig = null;
      "
      @edit="handleEditRoute"
      @delete="handleDeleteRoute"
      @update-mode="handleUpdateMode"
      @update-priority="handleUpdatePriority"
    />

    <!-- Auto Scheduled Section -->
    <div class="auto-scheduled-section">
      <div class="section-header">
        <div class="section-title">
          <Icon name="clock" size="md" class="section-icon" />
          <h3>Auto Scheduled</h3>
          <span class="section-subtitle"
            >Timer-based execution dengan repeat</span
          >
        </div>
        <button @click="$emit('show-add-new')" class="btn btn-add">
          <Icon name="plus" size="xs" /> Add Schedule
        </button>
      </div>

      <!-- Empty State -->
      <div v-if="autoFunctionConfigs.length === 0" class="empty-state-small">
        <p>No auto-functions configured</p>
      </div>

      <!-- Config Table -->
      <ConfigFunctionsTable
        v-else
        :configs="autoFunctionConfigs"
        @toggle-function="
          (name, enabled) => $emit('toggle-function', name, enabled)
        "
        @show-editor="(config) => $emit('show-editor', config)"
        @delete-function="(name) => $emit('delete-function', name)"
      />
    </div>

    <!-- Execution Stats -->
    <div class="execution-stats">
      <div class="stat-item stat-completed">
        <Icon name="check" size="md" class="stat-icon text-green-600" />
        <span class="stat-value">{{ stats.completed }}</span>
        <span class="stat-label">Completed</span>
      </div>
      <div class="stat-item stat-failed">
        <Icon name="close" size="md" class="stat-icon text-red-600" />
        <span class="stat-value">{{ stats.failed }}</span>
        <span class="stat-label">Failed</span>
      </div>
      <div class="stat-item stat-pending">
        <Icon name="spinner" size="md" class="stat-icon text-yellow-600" spin />
        <span class="stat-value">{{ stats.pending }}</span>
        <span class="stat-label">Pending</span>
      </div>
      <div class="stat-item stat-queue">
        <Icon name="document" size="md" class="stat-icon text-blue-600" />
        <span class="stat-value">{{ stats.inQueue }}</span>
        <span class="stat-label">In Queue</span>
      </div>
    </div>

    <!-- Config Editor Modal (Auto Functions) -->
    <ConfigEditorModal
      :config="editingConfig"
      :availableFunctions="availableFunctions"
      @close-editor="() => $emit('close-editor')"
      @save-config="() => $emit('save-config')"
      @increment-time="
        (field, minutes) => $emit('increment-time', field, minutes)
      "
      @decrement-time="
        (field, minutes) => $emit('decrement-time', field, minutes)
      "
    />

    <!-- Route Config Modal (Manual Trigger) -->
    <RouteConfigModal
      :show="showRouteModal"
      :config="editingRouteConfig"
      @close="showRouteModal = false"
      @save="handleSaveRoute"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, computed } from "vue";
import Icon from "@/components/ui/Icon.vue";
import ConfigFunctionsTable from "./ConfigFunctionsTable.vue";
import ConfigEditorModal from "./ConfigEditorModal.vue";
import ManualTriggerSection from "./ManualTriggerSection.vue";
import RouteConfigModal from "./RouteConfigModal.vue";
import { useRouteExecutionConfig } from "@/composables/useRouteExecutionConfig";
import { type AutoFunctionConfig } from "./configTableUtils";
import type { RouteExecutionConfig } from "@/types/routeExecutionConfig";

const props = defineProps<{
  activeTab: string;
  autoFunctionConfigs: AutoFunctionConfig[];
  editingConfig: AutoFunctionConfig | null;
  availableFunctions: Array<{ value: string; label: string }>;
}>();

defineEmits<{
  "show-add-new": [];
  "show-editor": [config: AutoFunctionConfig];
  "close-editor": [];
  "toggle-function": [name: string, enabled: boolean];
  "delete-function": [name: string];
  "save-config": [];
  "increment-time": [field: "start_time" | "end_time", minutes: number];
  "decrement-time": [field: "start_time" | "end_time", minutes: number];
}>();

// Route Execution Config
const routeConfig = useRouteExecutionConfig();
const showRouteModal = ref(false);
const editingRouteConfig = ref<RouteExecutionConfig | null>(null);

const routeExecutionConfigs = computed(() => routeConfig.configs.value);

// Stats (placeholder - can be connected to real data)
const stats = ref({
  completed: 0,
  failed: 0,
  pending: 0,
  inQueue: 0,
});

onMounted(async () => {
  await routeConfig.fetchConfigs();
});

function handleEditRoute(config: RouteExecutionConfig) {
  editingRouteConfig.value = config;
  showRouteModal.value = true;
}

async function handleDeleteRoute(routeKey: string) {
  if (!confirm(`Delete route "${routeKey}"?`)) return;
  await routeConfig.deleteConfig(routeKey);
}

async function handleUpdateMode(routeKey: string, mode: string) {
  await routeConfig.updateConfig(routeKey, {
    execution_mode: mode as "queue" | "direct",
  });
}

async function handleUpdatePriority(routeKey: string, priority: string) {
  await routeConfig.updateConfig(routeKey, {
    priority: priority as "low" | "normal" | "high",
  });
}

async function handleSaveRoute(data: any) {
  if (editingRouteConfig.value) {
    await routeConfig.updateConfig(editingRouteConfig.value.route_key, data);
  } else {
    await routeConfig.createConfig(data);
  }
  showRouteModal.value = false;
  editingRouteConfig.value = null;
}
</script>

<style scoped lang="css">
.tab-content {
  display: flex;
  flex-direction: column;
  gap: 20px;
  animation: slideInUp 0.3s cubic-bezier(0.4, 0, 0.2, 1);
}

@keyframes slideInUp {
  from {
    opacity: 0;
    transform: translateY(12px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

.auto-scheduled-section {
  background: #fff;
  border: 1px solid #e0e0e0;
  border-left: 4px solid #9b59b6;
  border-radius: 10px;
  padding: 20px;
}

.section-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
}

.section-title {
  display: flex;
  align-items: center;
  gap: 10px;
}

.section-icon {
  font-size: 20px;
}

.section-title h3 {
  margin: 0;
  font-size: 16px;
  font-weight: 700;
  color: #1a1a1a;
}

.section-subtitle {
  font-size: 12px;
  color: #6b7280;
}

.btn-add {
  padding: 6px 12px;
  background: #9b59b6;
  color: white;
  border: none;
  border-radius: 6px;
  font-size: 12px;
  font-weight: 600;
  cursor: pointer;
}

.btn-add:hover {
  background: #8e44ad;
}

.empty-state-small {
  text-align: center;
  padding: 30px 20px;
  background: #f8f9fa;
  border-radius: 8px;
  color: #666;
}

.execution-stats {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 16px;
  background: #fff;
  border: 1px solid #e0e0e0;
  border-left: 4px solid #2ecc71;
  border-radius: 10px;
  padding: 20px;
}

.stat-item {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 4px;
  padding: 12px;
  background: #f8f9fa;
  border-radius: 8px;
}

.stat-icon {
  font-size: 20px;
}
.stat-value {
  font-size: 24px;
  font-weight: 700;
  color: #1a1a1a;
}
.stat-label {
  font-size: 11px;
  color: #6b7280;
  text-transform: uppercase;
}

.stat-completed .stat-value {
  color: #27ae60;
}
.stat-failed .stat-value {
  color: #e74c3c;
}
.stat-pending .stat-value {
  color: #f39c12;
}
.stat-queue .stat-value {
  color: #3498db;
}

@media (max-width: 768px) {
  .execution-stats {
    grid-template-columns: repeat(2, 1fr);
  }
}
</style>
