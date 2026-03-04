<template>
  <div class="route-mapping-viewer">
    <div class="viewer-header">
      <h1>Route & Component Mapping</h1>
      <div class="header-stats">
        <div class="stat-card">
          <span class="stat-label">Total Routes</span>
          <span class="stat-value">{{ data?.total_routes || 0 }}</span>
        </div>
        <div class="stat-card">
          <span class="stat-label">Components (with routes)</span>
          <span class="stat-value">{{
            Object.keys(filteredComponents).length || 0
          }}</span>
        </div>
        <div class="stat-card">
          <span class="stat-label">Categories</span>
          <span class="stat-value">{{ data?.total_categories || 0 }}</span>
        </div>
        <div class="stat-card">
          <span class="stat-label">Connected</span>
          <span class="stat-value stat-success">{{
            categoryStats.connected
          }}</span>
        </div>
        <div class="stat-card">
          <span class="stat-label">Frontend Missing Backend</span>
          <span class="stat-value stat-error">{{
            categoryStats.frontend_only
          }}</span>
        </div>
        <div class="stat-card">
          <span class="stat-label">Backend Only</span>
          <span class="stat-value stat-warn">{{
            categoryStats.backend_only
          }}</span>
        </div>
        <div class="stat-card">
          <span class="stat-label">Unused</span>
          <span class="stat-value stat-warn">{{ categoryStats.unused }}</span>
        </div>
        <div class="stat-card">
          <span class="stat-label">Connection Rate</span>
          <span class="stat-value">{{ data?.connection_rate || "0%" }}</span>
        </div>
      </div>
    </div>

    <div class="viewer-controls">
      <div class="filter-group">
        <label for="view-mode">View Mode:</label>
        <select
          v-model="viewMode"
          @change="updateURLQuery(viewMode)"
          id="view-mode"
          class="select-input"
        >
          <option value="categories">By Category</option>
          <option value="component">By Component</option>
          <option value="disconnected">Frontend Missing Backend</option>
          <option value="backend">Backend-only</option>
          <option value="unused">Unused</option>
        </select>
      </div>

      <div class="filter-group">
        <label for="category-filter">Filter:</label>
        <input
          v-model="searchQuery"
          id="category-filter"
          type="text"
          placeholder="Search routes, components, or endpoints..."
          class="search-input"
        />
      </div>

      <button @click="refreshData" class="refresh-btn" :disabled="loading">
        {{ loading ? "Loading..." : "Refresh" }}
      </button>
    </div>

    <div v-if="loading" class="loading-state">
      <div class="spinner"></div>
      <p>Loading route mapping data...</p>
    </div>

    <div v-else-if="error" class="error-state">
      <p>Error: {{ error }}</p>
      <button @click="refreshData" class="retry-btn">Try Again</button>
    </div>

    <div v-else class="viewer-content">
      <RouteCategoryView
        v-if="viewMode === 'categories'"
        :categories="normalizedCategories"
        :labels="data?.category_labels || {}"
        :searchQuery="searchQuery"
      />

      <RouteComponentView
        v-else-if="viewMode === 'component'"
        :components="filteredComponents"
        :buttonToEndpoints="data?.button_to_endpoints || {}"
      />

      <RouteDisconnectedView
        v-else-if="viewMode === 'disconnected'"
        :routes="filteredDisconnected"
      />

      <RouteUnusedView
        v-else-if="viewMode === 'backend'"
        :routes="filteredBackendOnly"
        title="Backend-only Routes"
        description="Backend routes that are not called from the frontend but look like internal utilities."
        emptyTitle="No backend-only routes detected"
        emptyDescription="All backend routes are connected to the frontend or unused."
      />

      <RouteUnusedView v-else :routes="filteredUnused" />
    </div>

    <div class="viewer-footer">
      <small>Last updated: {{ formatDate(data?.timestamp) }}</small>
    </div>
  </div>
</template>

<script setup lang="ts">
import RouteCategoryView from "./RouteCategoryView.vue";
import RouteComponentView from "./RouteComponentView.vue";
import RouteDisconnectedView from "./RouteDisconnectedView.vue";
import RouteUnusedView from "./RouteUnusedView.vue";
import { useRouteMappingData } from "./useRouteMappingData";

const {
  data,
  loading,
  error,
  viewMode,
  searchQuery,
  normalizedCategories,
  categoryStats,
  filteredComponents,
  filteredDisconnected,
  filteredBackendOnly,
  filteredUnused,
  refreshData,
  formatDate,
  updateURLQuery,
} = useRouteMappingData();
</script>

<style scoped src="./RouteMappingViewer.styles.css"></style>
