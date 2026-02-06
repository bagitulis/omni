import { defineStore } from "pinia";
import { ref, computed, Ref } from "vue";

export interface ExpandedSections {
  "product-manager": boolean;
  "order-manager": boolean;
  "master-product": boolean;
  settings: boolean;
  "script-monitor": boolean;
  analytics: boolean;
  report: boolean;
  [key: string]: boolean;
}

export const useUIStore = defineStore("ui", () => {
  // State
  const leftSidebarCollapsed: Ref<boolean> = ref(false);
  const rightSidebarCollapsed: Ref<boolean> = ref(true);
  const activeTab: Ref<
    | "logs"
    | "settings"
    | "product-management"
    | "order-management"
    | "inventory"
    | "script-monitor"
    | "analytics"
  > = ref("product-management"); // Default to product-management tab
  const settingsSubsection: Ref<
    "google-sheets" | "logs" | "resources" | "script-monitor"
  > = ref("google-sheets");
  const activePlatform: Ref<"shopee" | "lazada" | "tiktok"> = ref("shopee");
  const autoScroll: Ref<boolean> = ref(true);
  const activeProduct: Ref<string | number | null> = ref(null);

  // Sidebar menu expansion state - persisted across route changes
  const expandedSections: Ref<ExpandedSections> = ref({
    "product-manager": false,
    "order-manager": false,
    "master-product": false,
    settings: false,
    "script-monitor": false,
    analytics: false,
    report: false,
  });

  // Getters
  const isLeftSidebarOpen = computed(() => !leftSidebarCollapsed.value);
  const isRightSidebarOpen = computed(() => !rightSidebarCollapsed.value);

  // Actions
  function toggleLeftSidebar(): void {
    leftSidebarCollapsed.value = !leftSidebarCollapsed.value;
  }

  function toggleRightSidebar(): void {
    rightSidebarCollapsed.value = !rightSidebarCollapsed.value;
  }

  function setActiveTab(
    tab:
      | "logs"
      | "settings"
      | "product-management"
      | "order-management"
      | "inventory"
      | "script-monitor"
      | "analytics",
  ): void {
    activeTab.value = tab;
  }

  function setSettingsSubsection(
    subsection: "google-sheets" | "logs" | "resources",
  ): void {
    settingsSubsection.value = subsection;
  }

  function setActivePlatform(platform: "shopee" | "lazada" | "tiktok"): void {
    activePlatform.value = platform;
  }

  function setAutoScroll(value: boolean): void {
    autoScroll.value = value;
  }

  function setActiveProduct(itemId: string | number | null): void {
    activeProduct.value = itemId;
  }

  function clearActiveProduct(): void {
    activeProduct.value = null;
  }

  // Toggle menu section expansion
  function toggleMenuSection(section: string): void {
    if (section in expandedSections.value) {
      expandedSections.value[section] = !expandedSections.value[section];
    }
  }

  // Auto-expand section based on current route and close unrelated sections
  function expandSectionForRoute(routePath: string): void {
    // Determine which section should be open based on route
    const sectionMap: Record<string, string> = {
      "/product-manager": "product-manager",
      "/master-products": "master-product",
      "/inventory": "master-product",
      "/script-monitor": "script-monitor",
      "/settings": "settings",
      "/analytics": "analytics",
      "/report": "report",
    };

    // Find which section matches current route
    let activeSection: string | null = null;
    for (const [prefix, section] of Object.entries(sectionMap)) {
      if (routePath.startsWith(prefix)) {
        activeSection = section;
        break;
      }
    }

    // Close all sections that are not the active one
    for (const section of Object.keys(expandedSections.value)) {
      if (section === activeSection) {
        expandedSections.value[section] = true;
      } else {
        expandedSections.value[section] = false;
      }
    }
  }

  return {
    // State
    leftSidebarCollapsed,
    rightSidebarCollapsed,
    activeTab,
    settingsSubsection,
    activePlatform,
    autoScroll,
    activeProduct,
    expandedSections,

    // Getters
    isLeftSidebarOpen,
    isRightSidebarOpen,

    // Actions
    toggleLeftSidebar,
    toggleRightSidebar,
    setActiveTab,
    setSettingsSubsection,
    setActivePlatform,
    setAutoScroll,
    setActiveProduct,
    clearActiveProduct,
    toggleMenuSection,
    expandSectionForRoute,
  };
});
