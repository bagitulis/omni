<template>
  <div class="sidebar-menu" v-if="!collapsed">
    <!-- Product Manager Section -->
    <div class="menu-section" :class="{ active: isActive('product-manager') }">
      <div
        class="menu-item parent"
        @click="$emit('toggle-expand', 'product-manager')"
      >
        <div class="menu-label-flex">
          <span>🛍️</span>
          <span>Product Manager</span>
        </div>
        <span
          class="expand-icon"
          :class="{ rotated: expandedSections['product-manager'] }"
          >▼</span
        >
      </div>
      <MenuProductManagerSection
        v-if="expandedSections['product-manager']"
        :product-platforms="productPlatforms"
        :is-active="isActive('product-manager')"
        :active-platform="activePlatform"
      />
    </div>

    <!-- Order Manager -->
    <router-link
      to="/order-manager"
      class="menu-item"
      :class="{ active: isActive('order-manager') }"
    >
      <span>📋</span>
      <span>Order Manager</span>
    </router-link>

    <!-- Inventory -->
    <router-link
      to="/inventory"
      class="menu-item"
      :class="{ active: isActive('inventory') }"
    >
      <span>📦</span>
      <span>Inventory</span>
    </router-link>

    <!-- Route Mapper -->
    <router-link
      to="/route-mapping"
      class="menu-item"
      :class="{ active: isActive('route-mapping') }"
    >
      <span>🗺️</span>
      <span>Route Mapper</span>
    </router-link>

    <!-- Script Monitor with Sub-menus (Top Level) -->
    <div class="menu-section" :class="{ active: isActive('script-monitor') }">
      <div
        class="menu-item parent"
        @click="$emit('toggle-expand', 'script-monitor')"
      >
        <div class="menu-label-flex">
          <span>🎬</span>
          <span>Script Monitor</span>
        </div>
        <span
          class="expand-icon"
          :class="{ rotated: expandedSections['script-monitor'] }"
          >▼</span
        >
      </div>

      <!-- Submenu untuk Script Monitor -->
      <div class="submenu" v-if="expandedSections['script-monitor']">
        <router-link
          to="/script-monitor/current"
          class="submenu-item"
          :class="{ active: currentPathIs('/script-monitor/current') }"
        >
          <span>⏳</span>
          <span>Current Running</span>
        </router-link>
        <router-link
          to="/script-monitor/queue"
          class="submenu-item"
          :class="{ active: currentPathIs('/script-monitor/queue') }"
        >
          <span>📋</span>
          <span>Queue</span>
        </router-link>
        <router-link
          to="/script-monitor/history"
          class="submenu-item"
          :class="{ active: currentPathIs('/script-monitor/history') }"
        >
          <span>✅</span>
          <span>History</span>
        </router-link>
        <router-link
          to="/script-monitor/auto-functions"
          class="submenu-item"
          :class="{ active: currentPathIs('/script-monitor/auto-functions') }"
        >
          <span>⚙️</span>
          <span>Auto-Functions</span>
        </router-link>
      </div>
    </div>

    <!-- Report Section (Shopee & TikTok Reports) -->
    <div class="menu-section" :class="{ active: isActive('report') }">
      <div class="menu-item parent" @click="$emit('toggle-expand', 'report')">
        <div class="menu-label-flex">
          <span>📑</span>
          <span>Report</span>
        </div>
        <span
          class="expand-icon"
          :class="{ rotated: expandedSections['report'] }"
          >▼</span
        >
      </div>

      <!-- Submenu untuk Report -->
      <div class="submenu" v-if="expandedSections['report']">
        <router-link
          to="/report/shopee"
          class="submenu-item"
          :class="{ active: currentPathIs('/report/shopee') }"
        >
          <span>🟠</span>
          <span>Shopee</span>
        </router-link>
        <router-link
          to="/report/tiktok"
          class="submenu-item"
          :class="{ active: currentPathIs('/report/tiktok') }"
        >
          <span>🎵</span>
          <span>TikTok</span>
        </router-link>
      </div>
    </div>

    <!-- Analytics Section (TikTok Ads only) -->
    <div class="menu-section" :class="{ active: isActive('analytics') }">
      <div
        class="menu-item parent"
        @click="$emit('toggle-expand', 'analytics')"
      >
        <div class="menu-label-flex">
          <span>📊</span>
          <span>Analytics</span>
        </div>
        <span
          class="expand-icon"
          :class="{ rotated: expandedSections['analytics'] }"
          >▼</span
        >
      </div>

      <!-- Submenu untuk Analytics -->
      <div class="submenu" v-if="expandedSections['analytics']">
        <router-link
          to="/analytics/hub"
          class="submenu-item"
          :class="{ active: currentPathIs('/analytics/hub') }"
        >
          <span>📊</span>
          <span>Analytics Hub</span>
        </router-link>
        <router-link
          to="/analytics/simulator"
          class="submenu-item"
          :class="{ active: currentPathIs('/analytics/simulator') }"
        >
          <span>🎯</span>
          <span>Budget Simulator</span>
        </router-link>
        <router-link
          to="/analytics/classification"
          class="submenu-item"
          :class="{ active: currentPathIs('/analytics/classification') }"
        >
          <span>📋</span>
          <span>Product Classification</span>
        </router-link>
        <router-link
          to="/analytics/ml"
          class="submenu-item"
          :class="{ active: currentPathIs('/analytics/ml') }"
        >
          <span>🧠</span>
          <span>ML Dashboard</span>
        </router-link>
        <router-link
          to="/analytics/shopee-ads"
          class="submenu-item"
          :class="{ active: currentPathIs('/analytics/shopee-ads') }"
        >
          <span>🛒</span>
          <span>Shopee Ads</span>
        </router-link>
        <router-link
          to="/analytics/tiktok-ads"
          class="submenu-item"
          :class="{ active: currentPathIs('/analytics/tiktok-ads') }"
        >
          <span>📈</span>
          <span>TikTok Ads</span>
        </router-link>
        <router-link
          to="/analytics/ai-reports"
          class="submenu-item"
          :class="{ active: currentPathIs('/analytics/ai-reports') }"
        >
          <span>🤖</span>
          <span>AI Reports</span>
        </router-link>
      </div>
    </div>

    <!-- Settings Section -->
    <MenuSettingsSection
      :expanded="expandedSections.settings"
      :is-active="isActive('settings')"
      :is-settings-google-sheets-active="isActive('settings-google-sheets')"
      :is-settings-logs-active="isActive('settings-logs')"
      :is-settings-resources-active="isActive('settings-resources')"
      :is-settings-webhook-active="isActive('settings-webhook')"
      @toggle="$emit('toggle-expand', 'settings')"
    />
  </div>
</template>

<script lang="ts">
import { defineComponent, PropType } from "vue";
import { useRoute } from "vue-router";
import MenuProductManagerSection from "./MenuProductManagerSection.vue";
import MenuSettingsSection from "./MenuSettingsSection.vue";
import type { ExpandedSections } from "./composables/useMenuExpansion";

interface Platform {
  value: string;
  label: string;
  icon: string;
}

export default defineComponent({
  name: "SidebarMenu",
  components: {
    MenuProductManagerSection,
    MenuSettingsSection,
  },
  props: {
    collapsed: {
      type: Boolean,
      default: false,
    },
    platforms: {
      type: Array as PropType<Platform[]>,
      required: true,
    },
    productPlatforms: {
      type: Array as PropType<Platform[]>,
      required: true,
    },
    expandedSections: {
      type: Object as PropType<ExpandedSections>,
      required: true,
    },
    activePlatform: {
      type: String,
      default: "shopee",
    },
  },
  emits: ["toggle-expand"],
  setup() {
    const route = useRoute();

    const isActive = (section: string): boolean => {
      const currentPath = route.path;
      switch (section) {
        case "operation":
          return currentPath.startsWith("/operation");
        case "product-manager":
          return currentPath.startsWith("/product-manager");
        case "order-manager":
          return currentPath.startsWith("/order-manager");
        case "inventory":
          return currentPath === "/inventory";
        case "settings":
          return currentPath.startsWith("/settings");
        case "settings-google-sheets":
          return currentPath === "/settings/google-sheets";
        case "settings-logs":
          return currentPath === "/settings/logs";
        case "settings-resources":
          return currentPath === "/settings/resources";
        case "settings-webhook":
          return currentPath === "/settings/webhook";
        case "script-monitor":
          return currentPath.startsWith("/script-monitor");
        case "analytics":
          return currentPath.startsWith("/analytics");
        case "report":
          return currentPath.startsWith("/report");
        case "route-mapping":
          return currentPath === "/route-mapping";
        default:
          return false;
      }
    };

    const currentPathIs = (path: string): boolean => {
      return route.path === path;
    };

    return { isActive, currentPathIs };
  },
});
</script>

<style src="./SidebarMenu.styles.css" scoped></style>
