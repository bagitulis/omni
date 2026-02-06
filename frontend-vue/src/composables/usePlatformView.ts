import { computed } from "vue";
import { useRouter, useRoute } from "vue-router";
import { useModalsStore } from "@/store/modals";

/**
 * Shared logic for platform-based views (operation, order-manager, product-manager)
 * Handles tab navigation, platform switching, and modal management
 */
export function usePlatformView() {
  const modalsStore = useModalsStore();
  const router = useRouter();
  const route = useRoute();

  // Extract platform from route path (e.g., /operation/shopee -> "shopee")
  const activePlatform = computed(() => {
    const pathSegments = route.path.split("/");
    return pathSegments[2] || "shopee"; // Default to shopee
  });

  // Refresh all data
  const refreshAll = () => {
    // Can be overridden by caller
  };

  // Handle tab changes (operation, order-manager, product-manager, inventory, settings)
  const handleTabChange = (tab: string) => {
    if (tab !== route.params.tab) {
      router.push(`/${tab}`);
    }
  };

  // Handle platform changes (shopee, lazada, tiktok)
  const handlePlatformChange = (platform: string) => {
    const section = route.params.section || route.path.split("/")[1];
    router.push(`/${section}/${platform}`);
  };

  // Modal handlers
  const showTokenModal = (platform: string) => {
    modalsStore.showModal("token", platform);
  };

  const showExportOrdersModal = (platform: string) => {
    modalsStore.showModal("exportOrders", platform);
  };

  const showPriceModal = (platform: string) => {
    modalsStore.showModal("price", platform);
  };

  const showWalletModal = () => {
    modalsStore.showModal("wallet");
  };

  const showShippingModal = () => {
    modalsStore.showModal("shipping");
  };

  return {
    activePlatform,
    refreshAll,
    handleTabChange,
    handlePlatformChange,
    showTokenModal,
    showExportOrdersModal,
    showPriceModal,
    showWalletModal,
    showShippingModal,
  };
}
