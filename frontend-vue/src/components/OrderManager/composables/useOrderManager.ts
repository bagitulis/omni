/**
 * Order Manager Composable
 * Handles order fetching, syncing, and filtering logic
 */

import { ref, computed, watch } from "vue";
import { useRoute, useRouter } from "vue-router";
import { useToast } from "@/composables/useToast";
import { getAuthHeaders, getApiBaseUrl } from "@/utils/apiHeaders";

// Get base URL for orders API
const API_BASE_URL = getApiBaseUrl("/orders");

export interface Order {
  order_no: string;
  platform: string;
  sku: string;
  product_name: string;
  variation_name?: string;
  qty: number;
  tracking_no?: string;
  courier?: string;
  [key: string]: any;
}

interface TabCount {
  [key: string]: number;
}

export interface OrderTab {
  label: string;
  value: string;
}

export const ORDER_TABS: OrderTab[] = [
  { label: "Unpaid", value: "unpaid" },
  { label: "To Ship", value: "unprocess" },
  { label: "Processed", value: "processed" },
  { label: "Locked Today", value: "locked" },
  { label: "Today's Orders", value: "today" },
];

export function useOrderManager() {
  const router = useRouter();
  const route = useRoute();
  const toast = useToast();

  // State
  const activeTab = ref<string>("unpaid");
  const orders = ref<Order[]>([]);
  const loading = ref<boolean>(false);
  const error = ref<string | null>(null);
  const searchQuery = ref<string>("");
  const selectedPlatform = ref<string>("");
  const tabCounts = ref<TabCount>({
    unpaid: 0,
    unprocess: 0,
    processed: 0,
    locked: 0,
    today: 0,
  });

  // Computed
  const uniquePlatforms = computed<string[]>(() => {
    const platforms = new Set(orders.value.map((o) => o.platform));
    return Array.from(platforms);
  });

  const uniqueOrders = computed<string[]>(() => {
    const orderNos = new Set(orders.value.map((o) => o.order_no));
    return Array.from(orderNos);
  });

  const filteredOrders = computed<Order[] | any[]>(() => {
    if (activeTab.value === "locked") return orders.value;

    let filtered: Order[] = orders.value;

    if (selectedPlatform.value) {
      filtered = filtered.filter((o) => o.platform === selectedPlatform.value);
    }

    if (searchQuery.value) {
      const query = searchQuery.value.toLowerCase();
      filtered = filtered.filter(
        (o) =>
          o.order_no.toLowerCase().includes(query) ||
          o.sku.toLowerCase().includes(query) ||
          o.product_name.toLowerCase().includes(query) ||
          (o.tracking_no && o.tracking_no.toLowerCase().includes(query)),
      );
    }

    return filtered;
  });

  // URL Management
  function updateURLQuery(tabValue: string): void {
    const query = { ...route.query, type: tabValue };
    router.push({ path: route.path, query }).catch(() => {});
  }

  // API Calls
  async function syncAndFetchCategoryData(category: string): Promise<void> {
    try {
      loading.value = true;
      error.value = null;

      const syncResponse = await fetch(`${API_BASE_URL}/sync/${category}`, {
        method: "POST",
        headers: getAuthHeaders(),
        body: JSON.stringify({ days: 7 }),
      });

      // Process sync response and log any errors
      const syncData = await syncResponse.json();
      if (!syncResponse.ok) {
        console.warn(
          `Sync warning for ${category}:`,
          syncData.error || "Sync completed with issues",
        );
      }

      await fetchOrdersForCurrentTab();
    } catch (err: any) {
      handleError(err);
    } finally {
      loading.value = false;
    }
  }

  async function syncAndRefreshData(): Promise<void> {
    try {
      loading.value = true;
      error.value = null;

      const syncResponse = await fetch(`${API_BASE_URL}/sync-all`, {
        method: "POST",
        headers: getAuthHeaders(),
        body: JSON.stringify({ days: 7 }),
      });

      // Process sync response and log any errors
      const syncData = await syncResponse.json();
      if (!syncResponse.ok) {
        console.warn(
          "Sync-all warning:",
          syncData.error || "Sync completed with issues",
        );
      }

      await fetchOrdersForCurrentTab();
    } catch (err: any) {
      handleError(err);
    } finally {
      loading.value = false;
    }
  }

  async function fetchOrdersForCurrentTab(): Promise<void> {
    if (activeTab.value === "locked") {
      await fetchLockedOrdersData();
      return;
    }

    if (activeTab.value === "today") {
      await fetchOrdersTodayData();
      return;
    }

    try {
      loading.value = true;
      error.value = null;
      orders.value = [];

      const response = await fetch(`${API_BASE_URL}/${activeTab.value}`, {
        method: "GET",
        headers: getAuthHeaders(),
      });

      if (!response.ok) {
        const errorData = await response.json();
        throw new Error(errorData.error || "Failed to fetch orders");
      }

      const data = await response.json();

      if (data.success) {
        orders.value = data.items || data.data || [];
        tabCounts.value[activeTab.value] = data.count || 0;
      } else {
        throw new Error(data.error || "Failed to fetch orders");
      }
    } catch (err: any) {
      handleError(err);
    } finally {
      loading.value = false;
    }
  }

  async function fetchLockedOrdersData(): Promise<void> {
    try {
      loading.value = true;
      error.value = null;

      const response = await fetch(`${API_BASE_URL}/locked-today`, {
        method: "POST",
        headers: getAuthHeaders(),
        body: JSON.stringify({ days: 7 }),
      });

      if (!response.ok) {
        const errorData = await response.json();
        throw new Error(errorData.error || "Failed to fetch locked orders");
      }

      const data = await response.json();

      if (data.success) {
        orders.value = data.items || [];
        tabCounts.value["locked"] = data.count || 0;
      } else {
        throw new Error(data.error || "Failed to fetch locked orders");
      }
    } catch (err: any) {
      handleError(err);
    } finally {
      loading.value = false;
    }
  }

  async function fetchOrdersTodayData(): Promise<void> {
    try {
      loading.value = true;
      error.value = null;

      // POST to fetch fresh data from APIs
      const response = await fetch(`${API_BASE_URL}/today`, {
        method: "POST",
        headers: getAuthHeaders(),
        body: JSON.stringify({ days: 7 }),
      });

      if (!response.ok) {
        const errorData = await response.json();
        throw new Error(errorData.error || "Failed to fetch order today");
      }

      const data = await response.json();

      if (data.success) {
        orders.value = data.items || [];
        tabCounts.value["today"] = data.count || 0;
      } else {
        throw new Error(data.error || "Failed to fetch order today");
      }
    } catch (err: any) {
      handleError(err);
    } finally {
      loading.value = false;
    }
  }

  // Tab Management
  async function changeTab(tabValue: string): Promise<void> {
    activeTab.value = tabValue;
    updateURLQuery(tabValue);
    searchQuery.value = "";
    selectedPlatform.value = "";
    orders.value = [];
    error.value = null;

    if (tabValue === "locked") {
      await fetchLockedOrdersData();
    } else if (tabValue === "today") {
      await fetchOrdersTodayData();
    } else {
      await syncAndFetchCategoryData(tabValue);
    }
  }

  async function refreshData(): Promise<void> {
    if (activeTab.value === "locked") {
      await fetchLockedOrdersData();
    } else if (activeTab.value === "today") {
      await fetchOrdersTodayData();
    } else {
      await syncAndRefreshData();
    }
  }

  // Error Handling
  function handleError(err: any): void {
    error.value = err.message;
    console.error("Error:", err);
    toast.add({
      severity: "error",
      summary: "Error",
      detail: err.message,
      life: 3000,
    });
  }

  // Initialize from URL
  async function initializeFromUrl(): Promise<void> {
    const typeFromQuery = route.query.type as string;
    const validTabs = ["unpaid", "unprocess", "processed", "locked", "today"];

    if (typeFromQuery && validTabs.includes(typeFromQuery)) {
      activeTab.value = typeFromQuery;
    }

    if (activeTab.value === "locked") {
      await fetchLockedOrdersData();
    } else if (activeTab.value === "today") {
      await fetchOrdersTodayData();
    } else {
      await syncAndFetchCategoryData(activeTab.value);
    }
  }

  // Watch for URL query changes (browser back/forward, direct URL access)
  watch(
    () => route.query.type,
    async (newType) => {
      const validTabs = ["unpaid", "unprocess", "processed", "locked", "today"];
      const typeValue = newType as string;

      if (
        typeValue &&
        validTabs.includes(typeValue) &&
        typeValue !== activeTab.value
      ) {
        activeTab.value = typeValue;
        orders.value = [];
        error.value = null;

        if (typeValue === "locked") {
          await fetchLockedOrdersData();
        } else if (typeValue === "today") {
          await fetchOrdersTodayData();
        } else {
          await syncAndFetchCategoryData(typeValue);
        }
      }
    },
  );

  return {
    // State
    activeTab,
    orders,
    loading,
    error,
    searchQuery,
    selectedPlatform,
    tabCounts,
    // Computed
    uniquePlatforms,
    uniqueOrders,
    filteredOrders,
    // Methods
    changeTab,
    refreshData,
    initializeFromUrl,
    handleError,
  };
}
