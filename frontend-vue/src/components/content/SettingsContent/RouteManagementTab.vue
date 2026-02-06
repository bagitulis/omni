<template>
  <div v-if="activeTab === 'route-management'" class="route-management-tab">
    <!-- Header -->
    <div class="management-header">
      <div class="header-info">
        <h3><Icon name="globe" size="sm" /> Route & Cache Management</h3>
        <p class="subtitle">
          Manage caching, queue, and timing for all routes dynamically
        </p>
      </div>
      <div class="header-actions">
        <button
          @click="logic.fetchRouteConfigs(true)"
          class="btn btn-primary"
          :disabled="logic.isLoading.value"
          aria-label="Refresh routes"
        >
          <Icon v-if="!logic.isLoading.value" name="refresh" size="sm" />
          {{ logic.isLoading.value ? "Loading..." : "Refresh" }}
        </button>
      </div>
    </div>

    <!-- Error Banner -->
    <div v-if="logic.error.value" class="error-banner">
      <Icon name="warning" size="sm" class="text-yellow-600" />
      {{ logic.error.value }}
      <button
        @click="logic.error.value = ''"
        class="btn-close"
        type="button"
        aria-label="Close error"
      >
        ✕
      </button>
    </div>

    <RouteFlowControls />

    <!-- Content Sections -->
    <RouteManagementContent
      :cache-manager="logic.cacheManager"
      :route-count="logic.allRoutes.value.length"
      :total-routes="logic.allRoutes.value.length"
      :enabled-count="logic.enabledCount.value"
      :caching-enabled-count="logic.cachingEnabledCount.value"
      :queue-enabled-count="logic.queueEnabledCount.value"
      @toggle-caching="logic.handleToggleCaching"
      @update-ttl="logic.handleCacheTTLChange"
      @update-auto-refresh="logic.handleAutoRefreshChange"
      @clear-cache="logic.clearCache"
      @update-exclude-routes="logic.handleExcludeRoutesChange"
      @add-route="showAddModal = true"
    >
      <!-- Toolbar -->
      <RouteManagementToolbar
        :search-query="searchQuery"
        :selected-category="selectedCategory"
        :categories="logic.categories.value"
        :all-enabled="allEnabled"
        :selected-count="selectedRouteIds.length"
        @update-search="searchQuery = $event"
        @update-category="selectedCategory = $event"
        @toggle-all-enabled="bulkToggleEnabled"
        @bulk-delete="bulkDelete"
        @apply-preset="applyPreset"
      />

      <!-- Routes Table -->
      <RouteManagementTable
        :routes="filteredRoutes"
        :selected-ids="selectedRouteIds"
        @toggle-selection="toggleRouteSelection"
        @select-all="selectAllRoutes"
        @toggle-enabled="logic.toggleRouteEnabled"
        @edit="editingRoute = $event"
        @delete="logic.deleteRoute"
      />
    </RouteManagementContent>

    <!-- Route Editor Modal -->
    <RouteManagementModal
      v-if="editingRoute"
      :route="editingRoute"
      :is-saving="logic.isSaving.value"
      @save="saveRoute"
      @close="editingRoute = null"
    />

    <!-- Add New Route Modal -->
    <RouteManagementModal
      v-if="showAddModal"
      :route="null"
      :is-saving="logic.isSaving.value"
      @save="saveRoute"
      @close="showAddModal = false"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from "vue";
import Icon from "@/components/ui/Icon.vue";
import { useRouteManagementLogic } from "./composables/useRouteManagementLogic";
import RouteManagementContent from "./RouteManagementContent.vue";
import RouteManagementTable from "./RouteManagementTable.vue";
import RouteManagementToolbar from "./RouteManagementToolbar.vue";
import RouteManagementModal from "./RouteManagementModal.vue";
import RouteFlowControls from "./RouteFlowControls.vue";
import type { RouteConfig, PresetType } from "./types/routeManagement";

defineProps<{ activeTab: string }>();

const logic = useRouteManagementLogic();

const searchQuery = ref("");
const selectedCategory = ref("");
const selectedRouteIds = ref<string[]>([]);
const editingRoute = ref<RouteConfig | null>(null);
const showAddModal = ref(false);

const filteredRoutes = computed(() => {
  return logic.allRoutes.value.filter((route) => {
    const matchesSearch =
      route.route_path
        .toLowerCase()
        .includes(searchQuery.value.toLowerCase()) ||
      route.description
        ?.toLowerCase()
        .includes(searchQuery.value.toLowerCase()) ||
      route.route_name?.toLowerCase().includes(searchQuery.value.toLowerCase());
    const matchesCategory =
      !selectedCategory.value || route.category === selectedCategory.value;
    return matchesSearch && matchesCategory;
  });
});

const allEnabled = computed(() => {
  const enabled = logic.allRoutes.value.filter((r) => r.enabled).length;
  return (
    enabled === logic.allRoutes.value.length && logic.allRoutes.value.length > 0
  );
});

const toggleRouteSelection = (routeId: string) => {
  const index = selectedRouteIds.value.indexOf(routeId);
  index > -1
    ? selectedRouteIds.value.splice(index, 1)
    : selectedRouteIds.value.push(routeId);
};

const selectAllRoutes = () => {
  if (selectedRouteIds.value.length === filteredRoutes.value.length) {
    selectedRouteIds.value = [];
  } else {
    selectedRouteIds.value = filteredRoutes.value
      .map((r) => r.id)
      .filter((id): id is string => Boolean(id));
  }
};

const bulkToggleEnabled = async () => {
  const ids =
    selectedRouteIds.value.length > 0
      ? selectedRouteIds.value
      : logic.allRoutes.value.map((r) => r.id!);
  await logic.bulkUpdateEnabled(ids, !allEnabled.value);
  selectedRouteIds.value = [];
};

const bulkDelete = async () => {
  await logic.bulkDelete(selectedRouteIds.value);
  selectedRouteIds.value = [];
};

const applyPreset = async (preset: string) => {
  await logic.applyPreset(selectedRouteIds.value, preset as PresetType);
  selectedRouteIds.value = [];
};

const saveRoute = async (route: RouteConfig) => {
  await logic.saveRoute(route);
  editingRoute.value = null;
  showAddModal.value = false;
};

onMounted(() => {
  logic.fetchRouteConfigs();
  if (logic.cacheManager.cacheConfig.value.autoRefreshMs > 0) {
    logic.cacheManager.startAutoRefresh();
  }
});

onUnmounted(() => {
  logic.cacheManager.stopAutoRefresh();
});
</script>

<style scoped lang="css">
.route-management-tab {
  display: flex;
  flex-direction: column;
  gap: 24px;
  padding: 0;
}

.management-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 24px;
  padding: 24px;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  border-radius: 8px;
  color: white;
}

.header-info h3 {
  margin: 0 0 8px 0;
  font-size: 20px;
}

.subtitle {
  margin: 0;
  opacity: 0.9;
  font-size: 14px;
}

.header-actions {
  display: flex;
  gap: 12px;
}

.btn {
  padding: 8px 16px;
  border: none;
  border-radius: 4px;
  font-size: 14px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s;
  display: inline-flex;
  align-items: center;
  gap: 6px;
}

.btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.btn-primary {
  background: #2196f3;
  color: white;
}

.btn-primary:hover:not(:disabled) {
  background: #1976d2;
}

.error-banner {
  padding: 12px 16px;
  background: #ffebee;
  border-left: 4px solid #f44336;
  border-radius: 4px;
  color: #c62828;
  font-size: 14px;
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.btn-close {
  background: none;
  border: none;
  font-size: 16px;
  cursor: pointer;
  color: #c62828;
}

@media (max-width: 768px) {
  .management-header {
    flex-direction: column;
    text-align: center;
  }

  .header-actions {
    flex-direction: column;
  }
}
</style>
