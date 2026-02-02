/**
 * useOrderFetch Composable
 * Handles order data fetching and synchronization
 */

import { ref } from "vue";
import { useToast } from "@/composables/useToast";
import { getAuthHeaders, getApiBaseUrl } from "@/utils/apiHeaders";

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

export function useOrderFetch() {
  const toast = useToast();
  const orders = ref<Order[]>([]);
  const loading = ref<boolean>(false);
  const error = ref<string | null>(null);

  const handleError = (err: any): void => {
    error.value = err.message;
    console.error("Error:", err);
    toast.add({
      severity: "error",
      summary: "Error",
      detail: err.message,
      life: 3000,
    });
  };

  const fetchOrdersByCategory = async (category: string): Promise<void> => {
    try {
      loading.value = true;
      error.value = null;
      orders.value = [];

      const response = await fetch(`${API_BASE_URL}/${category}`, {
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
      } else {
        throw new Error(data.error || "Failed to fetch orders");
      }
    } catch (err: any) {
      handleError(err);
    } finally {
      loading.value = false;
    }
  };

  const syncCategory = async (category: string): Promise<void> => {
    try {
      loading.value = true;
      error.value = null;

      const syncResponse = await fetch(`${API_BASE_URL}/sync/${category}`, {
        method: "POST",
        headers: getAuthHeaders(),
        body: JSON.stringify({ days: 7 }),
      });

      const syncData = await syncResponse.json();
      if (!syncResponse.ok) {
        console.warn(
          `Sync warning for ${category}:`,
          syncData.error || "Sync completed with issues",
        );
      }

      await fetchOrdersByCategory(category);
    } catch (err: any) {
      handleError(err);
    } finally {
      loading.value = false;
    }
  };

  const syncAll = async (): Promise<void> => {
    try {
      loading.value = true;
      error.value = null;

      const syncResponse = await fetch(`${API_BASE_URL}/sync-all`, {
        method: "POST",
        headers: getAuthHeaders(),
        body: JSON.stringify({ days: 7 }),
      });

      const syncData = await syncResponse.json();
      if (!syncResponse.ok) {
        console.warn(
          "Sync-all warning:",
          syncData.error || "Sync completed with issues",
        );
      }
    } catch (err: any) {
      handleError(err);
    } finally {
      loading.value = false;
    }
  };

  const fetchLockedOrders = async (): Promise<void> => {
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
      } else {
        throw new Error(data.error || "Failed to fetch locked orders");
      }
    } catch (err: any) {
      handleError(err);
    } finally {
      loading.value = false;
    }
  };

  const fetchTodayOrders = async (): Promise<void> => {
    try {
      loading.value = true;
      error.value = null;

      const response = await fetch(`${API_BASE_URL}/today`, {
        method: "POST",
        headers: getAuthHeaders(),
        body: JSON.stringify({ days: 7 }),
      });

      if (!response.ok) {
        const errorData = await response.json();
        throw new Error(errorData.error || "Failed to fetch today's orders");
      }

      const data = await response.json();

      if (data.success) {
        orders.value = data.items || [];
      } else {
        throw new Error(data.error || "Failed to fetch today's orders");
      }
    } catch (err: any) {
      handleError(err);
    } finally {
      loading.value = false;
    }
  };

  return {
    orders,
    loading,
    error,
    fetchOrdersByCategory,
    syncCategory,
    syncAll,
    fetchLockedOrders,
    fetchTodayOrders,
    handleError,
  };
}
