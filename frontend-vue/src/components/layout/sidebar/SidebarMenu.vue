<template>
  <div class="sidebar-menu" v-if="!collapsed">
    <!-- Product Manager Section -->
    <div class="menu-section" :class="{ active: isActive('product-manager') }">
      <div
        class="menu-item parent"
        @click="$emit('toggle-expand', 'product-manager')"
        role="button"
        :aria-expanded="expandedSections['product-manager']"
        aria-controls="submenu-product-manager"
      >
        <div class="menu-label-flex">
          <Icon name="shopping-bag" size="sm" />
          <span>Product Manager</span>
        </div>
        <Icon
          name="chevron-down"
          size="sm"
          :class="{ 'rotate-180': expandedSections['product-manager'] }"
          class="transition-transform"
        />
      </div>
      <MenuProductManagerSection
        v-if="expandedSections['product-manager']"
        id="submenu-product-manager"
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
      <Icon name="document" size="sm" />
      <span>Order Manager</span>
    </router-link>

    <!-- Master Produk Section -->
    <div class="menu-section" :class="{ active: isActive('master-product') }">
      <div
        class="menu-item parent"
        @click="$emit('toggle-expand', 'master-product')"
        role="button"
        :aria-expanded="expandedSections['master-product']"
        aria-controls="submenu-master-product"
      >
        <div class="menu-label-flex">
          <Icon name="shopping-cart" size="sm" />
          <span>Master Produk</span>
        </div>
        <Icon
          name="chevron-down"
          size="sm"
          :class="{ 'rotate-180': expandedSections['master-product'] }"
          class="transition-transform"
        />
      </div>

      <!-- Submenu untuk Master Produk -->
      <div
        id="submenu-master-product"
        class="submenu"
        v-if="expandedSections['master-product']"
      >
        <router-link
          to="/master-products"
          class="submenu-item"
          :class="{ active: currentPathIs('/master-products') }"
        >
          <Icon name="document" size="sm" />
          <span>Daftar Produk</span>
        </router-link>
        <router-link
          to="/master-products/add"
          class="submenu-item"
          :class="{ active: currentPathIs('/master-products/add') }"
        >
          <Icon name="plus" size="sm" />
          <span>Tambah Produk</span>
        </router-link>
        <router-link
          to="/master-products/import"
          class="submenu-item"
          :class="{ active: currentPathIs('/master-products/import') }"
        >
          <Icon name="download" size="sm" />
          <span>Import Produk</span>
        </router-link>
        <router-link
          to="/inventory"
          class="submenu-item"
          :class="{ active: currentPathIs('/inventory') }"
        >
          <Icon name="chart-bar" size="sm" />
          <span>Inventory</span>
        </router-link>
      </div>
    </div>

    <!-- Route Mapper -->
    <router-link
      to="/route-mapping"
      class="menu-item"
      :class="{ active: isActive('route-mapping') }"
    >
      <Icon name="globe" size="sm" />
      <span>Route Mapper</span>
    </router-link>

    <!-- Script Monitor with Sub-menus (Top Level) -->
    <div class="menu-section" :class="{ active: isActive('script-monitor') }">
      <div
        class="menu-item parent"
        @click="$emit('toggle-expand', 'script-monitor')"
        role="button"
        :aria-expanded="expandedSections['script-monitor']"
        aria-controls="submenu-script-monitor"
      >
        <div class="menu-label-flex">
          <Icon name="document" size="sm" />
          <span>Script Monitor</span>
        </div>
        <Icon
          name="chevron-down"
          size="sm"
          :class="{ 'rotate-180': expandedSections['script-monitor'] }"
          class="transition-transform"
        />
      </div>

      <!-- Submenu untuk Script Monitor -->
      <div
        id="submenu-script-monitor"
        class="submenu"
        v-if="expandedSections['script-monitor']"
      >
        <router-link
          to="/script-monitor?tab=current"
          class="submenu-item"
          :class="{
            active:
              currentPathIs('/script-monitor') && route.query.tab === 'current',
          }"
        >
          <Icon name="spinner" size="sm" spin />
          <span>Current Running</span>
        </router-link>
        <router-link
          to="/script-monitor?tab=queue"
          class="submenu-item"
          :class="{
            active:
              currentPathIs('/script-monitor') && route.query.tab === 'queue',
          }"
        >
          <Icon name="document" size="sm" />
          <span>Queue</span>
        </router-link>
        <router-link
          to="/script-monitor?tab=history"
          class="submenu-item"
          :class="{
            active:
              currentPathIs('/script-monitor') && route.query.tab === 'history',
          }"
        >
          <Icon name="check" size="sm" />
          <span>History</span>
        </router-link>
        <router-link
          to="/script-monitor?tab=config"
          class="submenu-item"
          :class="{
            active:
              currentPathIs('/script-monitor') && route.query.tab === 'config',
          }"
        >
          <Icon name="settings" size="sm" />
          <span>Auto-Functions</span>
        </router-link>
      </div>
    </div>

    <!-- Report Section (Shopee & TikTok Reports) -->
    <div class="menu-section" :class="{ active: isActive('report') }">
      <div
        class="menu-item parent"
        @click="$emit('toggle-expand', 'report')"
        role="button"
        :aria-expanded="expandedSections['report']"
        aria-controls="submenu-report"
      >
        <div class="menu-label-flex">
          <Icon name="clipboard" size="sm" />
          <span>Report</span>
        </div>
        <Icon
          name="chevron-down"
          size="sm"
          :class="{ 'rotate-180': expandedSections['report'] }"
          class="transition-transform"
        />
      </div>

      <!-- Submenu untuk Report -->
      <div
        id="submenu-report"
        class="submenu"
        v-if="expandedSections['report']"
      >
        <router-link
          to="/report/shopee"
          class="submenu-item"
          :class="{ active: currentPathIs('/report/shopee') }"
        >
          <Icon name="store" size="sm" />
          <span>Shopee</span>
        </router-link>
        <router-link
          to="/report/tiktok"
          class="submenu-item"
          :class="{ active: currentPathIs('/report/tiktok') }"
        >
          <Icon name="chart" size="sm" />
          <span>TikTok</span>
        </router-link>
      </div>
    </div>

    <!-- Analytics Section (TikTok Ads only) -->
    <div class="menu-section" :class="{ active: isActive('analytics') }">
      <div
        class="menu-item parent"
        @click="$emit('toggle-expand', 'analytics')"
        role="button"
        :aria-expanded="expandedSections['analytics']"
        aria-controls="submenu-analytics"
      >
        <div class="menu-label-flex">
          <Icon name="chart-bar" size="sm" />
          <span>Analytics</span>
        </div>
        <Icon
          name="chevron-down"
          size="sm"
          :class="{ 'rotate-180': expandedSections['analytics'] }"
          class="transition-transform"
        />
      </div>

      <!-- Submenu untuk Analytics -->
      <div
        id="submenu-analytics"
        class="submenu"
        v-if="expandedSections['analytics']"
      >
        <router-link
          to="/analytics/hub"
          class="submenu-item"
          :class="{ active: currentPathIs('/analytics/hub') }"
        >
          <Icon name="chart-bar" size="sm" />
          <span>Analytics Hub</span>
        </router-link>
        <router-link
          to="/analytics/simulator"
          class="submenu-item"
          :class="{ active: currentPathIs('/analytics/simulator') }"
        >
          <Icon name="adjustments" size="sm" />
          <span>Budget Simulator</span>
        </router-link>
        <router-link
          to="/analytics/shopee-ads"
          class="submenu-item"
          :class="{ active: currentPathIs('/analytics/shopee-ads') }"
        >
          <Icon name="store" size="sm" />
          <span>Shopee Ads</span>
        </router-link>
        <router-link
          to="/analytics/tiktok-ads"
          class="submenu-item"
          :class="{ active: currentPathIs('/analytics/tiktok-ads') }"
        >
          <Icon name="chart" size="sm" />
          <span>TikTok Ads</span>
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
import Icon from "@/components/ui/Icon.vue";
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
    Icon,
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
        case "master-product":
          return (
            currentPath.startsWith("/master-products") ||
            currentPath === "/inventory"
          );
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
