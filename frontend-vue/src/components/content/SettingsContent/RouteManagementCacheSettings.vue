<template>
  <div class="cache-settings-wrapper">
    <div class="settings-header">
      <p class="header-desc">
        Control how frontend caching works for API routes
      </p>
    </div>

    <div class="settings-content">
      <!-- Enable Caching -->
      <div class="setting-card">
        <label class="setting-checkbox">
          <input
            type="checkbox"
            :checked="cacheManager.cacheConfig.value.enabled"
            @change="(e) => onToggleCaching((e.target as any).checked)"
          />
          <span>Enable Frontend Caching</span>
        </label>
        <p class="setting-desc">
          Cache API responses locally to reduce server load
        </p>
      </div>

      <!-- Cache TTL -->
      <div v-if="cacheManager.cacheConfig.value.enabled" class="setting-card">
        <label class="setting-label">Cache Duration (TTL)</label>
        <div class="ttl-input-group">
          <input
            type="range"
            :value="cacheManager.cacheConfig.value.ttlSeconds"
            @change="(e) => onTTLChange(Number((e.target as any).value))"
            min="0"
            max="3600"
            step="30"
            class="ttl-slider"
          />
          <div class="ttl-display">
            <span class="ttl-value">{{ ttlDisplay }}</span>
            <button @click="ttlQuickSet(0)" class="ttl-quick">Fresh</button>
            <button @click="ttlQuickSet(300)" class="ttl-quick">5min</button>
            <button @click="ttlQuickSet(3600)" class="ttl-quick">1h</button>
          </div>
        </div>
        <p class="setting-desc">
          How long cached data stays valid (0 = always fresh from API)
        </p>
      </div>

      <!-- Auto Refresh -->
      <div v-if="cacheManager.cacheConfig.value.enabled" class="setting-card">
        <label class="setting-checkbox">
          <input
            type="checkbox"
            :checked="cacheManager.cacheConfig.value.autoRefreshMs > 0"
            @change="(e) => onAutoRefreshChange((e.target as any).checked)"
          />
          <span>Auto-refresh Cache</span>
        </label>
        <p class="setting-desc">
          Automatically refresh cached data every 30 seconds
        </p>
      </div>

      <!-- Cache Status -->
      <div class="setting-card status-card">
        <div class="status-info">
          <span class="status-label">Cache Status:</span>
          <span
            class="status-value"
            :class="{ 'is-enabled': cacheManager.cacheConfig.value.enabled }"
          >
            {{
              cacheManager.cacheConfig.value.enabled
                ? "✅ Active"
                : "❌ Disabled"
            }}
          </span>
        </div>
        <button @click="onClearCache" class="btn btn-clear">
          🗑️ Clear Cache Now
        </button>
      </div>
    </div>

    <!-- Route Exclusion Manager -->
    <div
      v-if="cacheManager.cacheConfig.value.enabled"
      class="exclusion-section"
    >
      <RouteExcludedRoutesManager
        :cache-manager="cacheManager"
        @update-exclude-routes="onExcludeRoutesChange"
      />
    </div>

    <!-- Cache Information (shown when showInfoCards is true) -->
    <div
      v-if="showInfoCards && cacheManager.cacheConfig.value.enabled"
      class="cache-info-section"
    >
      <h4 class="info-section-title">📊 Cache Information</h4>
      <div class="cache-info-grid">
        <div class="info-card">
          <span class="info-label">Caching Status:</span>
          <span
            class="info-value"
            :class="{ 'is-enabled': cacheManager.cacheConfig.value.enabled }"
          >
            {{
              cacheManager.cacheConfig.value.enabled
                ? "✅ Enabled"
                : "❌ Disabled"
            }}
          </span>
        </div>
        <div v-if="cacheManager.cacheConfig.value.enabled" class="info-card">
          <span class="info-label">Cache TTL:</span>
          <span class="info-value"
            >{{ cacheManager.cacheConfig.value.ttlSeconds }} seconds</span
          >
        </div>
        <div v-if="cacheManager.cacheConfig.value.enabled" class="info-card">
          <span class="info-label">Excluded Routes:</span>
          <span class="info-value">{{ excludedRoutesCount }}</span>
        </div>
        <div
          v-if="
            cacheManager.cacheConfig.value.enabled && excludedRoutesCount > 0
          "
          class="info-card full-width"
        >
          <span class="info-label">Excluded Routes:</span>
          <div class="excluded-routes-list">
            <span
              v-for="(route, index) in cacheManager.cacheConfig.value
                .excludeRoutes"
              :key="index"
              class="route-tag"
            >
              {{ route }}
            </span>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from "vue";
import RouteExcludedRoutesManager from "./RouteExcludedRoutesManager.vue";

interface Props {
  cacheManager: any;
  showInfoCards?: boolean;
}

interface Emits {
  (e: "toggle-caching", enabled: boolean): void;
  (e: "update-ttl", ttl: number): void;
  (e: "update-auto-refresh", enabled: boolean): void;
  (e: "clear-cache"): void;
  (e: "update-exclude-routes", routes: string[]): void;
}

const props = withDefaults(defineProps<Props>(), {
  showInfoCards: false,
});
const emit = defineEmits<Emits>();

const ttlDisplay = computed(() => {
  const ttl = props.cacheManager.cacheConfig.value.ttlSeconds;
  if (ttl === 0) return "Always Fresh (0s)";
  if (ttl < 60) return `${ttl} seconds`;
  if (ttl < 3600) return `${Math.round(ttl / 60)} minutes`;
  return `${Math.round(ttl / 3600)} hour(s)`;
});

const excludedRoutesCount = computed(() => {
  return (
    (props.cacheManager.cacheConfig.value as any).excludeRoutes?.length || 0
  );
});

const onToggleCaching = (enabled: boolean) => {
  emit("toggle-caching", enabled);
};

const onTTLChange = (ttl: number) => {
  emit("update-ttl", ttl);
};

const onAutoRefreshChange = (enabled: boolean) => {
  emit("update-auto-refresh", enabled);
};

const onClearCache = () => {
  emit("clear-cache");
};

const ttlQuickSet = (seconds: number) => {
  onTTLChange(seconds);
};

const onExcludeRoutesChange = (routes: string[]) => {
  emit("update-exclude-routes", routes);
};
</script>

<style src="./RouteManagementCacheSettings.styles.css" scoped></style>
