import { watch } from "vue";
import { useRoute, useRouter } from "vue-router";
import { useUIStore } from "../store/ui";
import { useGoogleSheetsStore } from "../store/googleSheets";

export function useDashboardNavigation() {
  const route = useRoute();
  const router = useRouter();
  const uiStore = useUIStore();
  const googleSheetsStore = useGoogleSheetsStore();

  const routeMap: Record<string, string> = {
    logs: "/",
    settings: "/settings",
    inventory: "/inventory",
    operation: "/operation/shopee",
    "product-management": "/product-manager/shopee",
    "order-management": "/order-manager",
  };

  const handleTabChange = async (tab: string) => {
    if (uiStore.activeTab === tab) return;

    if (tab === "inventory") {
      try {
        await googleSheetsStore.loadAuthStatus();
      } catch {
        // Silently handle error
      }
    }

    const targetRoute = routeMap[tab];
    if (targetRoute) {
      await router.push(targetRoute);
    }

    uiStore.setActiveTab(
      tab as
        | "logs"
        | "settings"
        | "operation"
        | "product-management"
        | "order-management"
        | "inventory"
    );
  };

  const handleNavigateToSettings = () => {
    router.push("/settings");
    uiStore.setActiveTab("settings");
  };

  const handlePlatformChange = (platform: string) => {
    uiStore.setActivePlatform(platform as "shopee" | "lazada" | "tiktok");
  };

  const syncRouteToTab = (path: string) => {
    const platformMatch = path.match(
      /\/(operation|product-manager|order-manager)\/(\w+)/
    );
    if (platformMatch) {
      const platform = platformMatch[2];
      if (["shopee", "lazada", "tiktok"].includes(platform)) {
        uiStore.setActivePlatform(platform as "shopee" | "lazada" | "tiktok");
      }
    }

    if (path.startsWith("/inventory")) {
      uiStore.setActiveTab("inventory");
    } else if (path.startsWith("/settings")) {
      uiStore.setActiveTab("settings");
      const pathParts = path.split("/");
      const settingsSubsection = pathParts[2];
      if (["google-sheets", "logs", "resources"].includes(settingsSubsection)) {
        uiStore.setSettingsSubsection(
          settingsSubsection as "google-sheets" | "logs" | "resources"
        );
      } else {
        uiStore.setSettingsSubsection("google-sheets");
      }
    } else if (path.startsWith("/operation")) {
      uiStore.setActiveTab("operation");
    } else if (path.startsWith("/product-manager")) {
      uiStore.setActiveTab("product-management");
    } else if (path.startsWith("/order-manager")) {
      uiStore.setActiveTab("order-management");
    } else if (path === "/") {
      uiStore.setActiveTab("logs");
    }
  };

  const initRouteSync = () => {
    syncRouteToTab(route.path);
    watch(
      () => route.path,
      (newPath) => {
        syncRouteToTab(newPath);
      }
    );
  };

  return {
    handleTabChange,
    handleNavigateToSettings,
    handlePlatformChange,
    initRouteSync,
  };
}
