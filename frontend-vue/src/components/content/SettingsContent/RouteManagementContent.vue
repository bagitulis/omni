<template>
  <!-- Cache Configuration & Information Section -->
  <div class="cache-config-wrapper">
    <div class="cache-section-header" @click="showCacheSettings = !showCacheSettings">
      <div class="header-left">
        <span class="toggle-icon">{{ showCacheSettings ? '▼' : '▶' }}</span>
        <h4>⚙️ Cache Configuration</h4>
      </div>
    </div>

    <transition name="collapse">
      <div v-if="showCacheSettings" class="cache-content">
        <RouteManagementCacheSettings
          :cache-manager="cacheManager"
          :show-info-cards="true"
          @toggle-caching="onToggleCaching"
          @update-ttl="onUpdateTtl"
          @update-auto-refresh="onUpdateAutoRefresh"
          @clear-cache="onClearCache"
          @update-exclude-routes="onUpdateExcludeRoutes"
        />
      </div>
    </transition>
  </div>

  <!-- Routes Table with Collapsible Header -->
  <div class="routes-table-wrapper">
    <div class="table-section-header" @click="showRoutesTable = !showRoutesTable">
      <div class="header-left">
        <span class="toggle-icon">{{ showRoutesTable ? '▼' : '▶' }}</span>
        <h4>📋 Routes Table</h4>
        <span class="route-count">({{ routeCount }} routes)</span>
      </div>
      <button @click.stop="onAddRoute" class="btn btn-add-route">
        ➕ Add Route
      </button>
    </div>

    <transition name="collapse">
      <div v-if="showRoutesTable" class="table-content">
        <slot />
      </div>
    </transition>
  </div>

  <!-- Collapsible Statistics -->
  <div class="stats-wrapper">
    <div class="stats-section-header" @click="showStats = !showStats">
      <div class="header-left">
        <span class="toggle-icon">{{ showStats ? '▼' : '▶' }}</span>
        <h4>📈 Statistics</h4>
      </div>
    </div>

    <transition name="collapse">
      <div v-if="showStats" class="management-stats">
        <div class="stat-item">
          <span class="stat-label">Total Routes:</span>
          <span class="stat-value">{{ totalRoutes }}</span>
        </div>
        <div class="stat-item">
          <span class="stat-label">Enabled:</span>
          <span class="stat-value">{{ enabledCount }}</span>
        </div>
        <div class="stat-item">
          <span class="stat-label">Caching Active:</span>
          <span class="stat-value">{{ cachingEnabledCount }}</span>
        </div>
        <div class="stat-item">
          <span class="stat-label">Queue Active:</span>
          <span class="stat-value">{{ queueEnabledCount }}</span>
        </div>
      </div>
    </transition>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue';
import RouteManagementCacheSettings from './RouteManagementCacheSettings.vue';

interface Props {
  cacheManager: any;
  routeCount: number;
  totalRoutes: number;
  enabledCount: number;
  cachingEnabledCount: number;
  queueEnabledCount: number;
}

const emit = defineEmits<{
  'toggle-caching': [enabled: boolean];
  'update-ttl': [ttl: number];
  'update-auto-refresh': [enabled: boolean];
  'clear-cache': [];
  'update-exclude-routes': [routes: string[]];
  'add-route': [];
}>();

defineProps<Props>();

const showCacheSettings = ref(false);
const showRoutesTable = ref(true);
const showStats = ref(true);

const onToggleCaching = (enabled: boolean) => emit('toggle-caching', enabled);
const onUpdateTtl = (ttl: number) => emit('update-ttl', ttl);
const onUpdateAutoRefresh = (enabled: boolean) => emit('update-auto-refresh', enabled);
const onClearCache = () => emit('clear-cache');
const onUpdateExcludeRoutes = (routes: string[]) => emit('update-exclude-routes', routes);
const onAddRoute = () => emit('add-route');
</script>

<style scoped lang="css">
.cache-config-wrapper,
.routes-table-wrapper,
.stats-wrapper {
  background: white;
  border-radius: 8px;
  border: 1px solid #e0e0e0;
  overflow: hidden;
  margin-bottom: 16px;
}

.cache-section-header,
.table-section-header,
.stats-section-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 16px 20px;
  background: #f8f9fa;
  border-bottom: 1px solid #e0e0e0;
  cursor: pointer;
  transition: background 0.2s;
  user-select: none;
}

.cache-section-header:hover,
.table-section-header:hover,
.stats-section-header:hover {
  background: #f0f2f5;
}

.header-left {
  display: flex;
  align-items: center;
  gap: 12px;
}

.toggle-icon {
  font-size: 12px;
  color: #666;
  display: inline-block;
  min-width: 12px;
}

.cache-section-header h4,
.table-section-header h4,
.stats-section-header h4 {
  margin: 0;
  font-size: 15px;
  font-weight: 700;
  color: #1a1a1a;
}

.route-count {
  font-size: 13px;
  color: #6b7280;
  font-weight: 500;
}

.btn-add-route {
  padding: 8px 16px;
  background: #4caf50;
  color: white;
  border: none;
  border-radius: 4px;
  font-size: 13px;
  font-weight: 600;
  cursor: pointer;
  white-space: nowrap;
  transition: all 0.2s;
}

.btn-add-route:hover {
  background: #388e3c;
  box-shadow: 0 2px 6px rgba(0, 0, 0, 0.1);
}

.cache-content,
.table-content {
  padding: 16px;
  border-top: 1px solid #e0e0e0;
}

.management-stats {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 16px;
  padding: 16px;
  background: white;
}

.stat-item {
  padding: 12px;
  background: #f9f9f9;
  border-radius: 6px;
  text-align: center;
}

.stat-label {
  display: block;
  font-size: 12px;
  color: #666;
  margin-bottom: 6px;
  font-weight: 600;
}

.stat-value {
  display: block;
  font-size: 18px;
  font-weight: 700;
  color: #333;
}

.collapse-enter-active,
.collapse-leave-active {
  transition: all 0.3s ease;
}

.collapse-enter-from,
.collapse-leave-to {
  opacity: 0;
  max-height: 0;
}

.collapse-enter-to,
.collapse-leave-from {
  opacity: 1;
  max-height: 2000px;
}

@media (max-width: 1024px) {
  .management-stats {
    grid-template-columns: repeat(2, 1fr);
  }
}

@media (max-width: 600px) {
  .management-stats {
    grid-template-columns: 1fr;
  }
}
</style>
