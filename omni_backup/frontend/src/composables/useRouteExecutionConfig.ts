/**
 * useRouteExecutionConfig Composable
 * Manages manual trigger mode configuration (Queue vs Direct)
 * Single Responsibility: API calls and state for route execution configs
 */

import { ref } from "vue";
import axios from "axios";
import { getAuthHeaders, getApiBaseUrl } from "@/utils/apiHeaders";
import type {
  RouteExecutionConfig,
  RouteExecutionConfigInput,
  ExecutionMode,
} from "@/types/routeExecutionConfig";

const API_BASE = () => getApiBaseUrl("/route-execution-config");

export function useRouteExecutionConfig() {
  const configs = ref<RouteExecutionConfig[]>([]);
  const isLoading = ref(false);
  const error = ref<string | null>(null);

  /**
   * Fetch all route execution configs
   */
  async function fetchConfigs(): Promise<void> {
    isLoading.value = true;
    error.value = null;

    try {
      const response = await axios.get(API_BASE(), {
        headers: getAuthHeaders(),
      });

      if (response.data.success) {
        configs.value = response.data.data;
      }
    } catch (err: any) {
      error.value = err.response?.data?.error || err.message;
      console.error("[RouteExecutionConfig] Fetch failed:", error.value);
    } finally {
      isLoading.value = false;
    }
  }

  /**
   * Get execution mode for a specific route
   */
  async function getExecutionMode(routeKey: string): Promise<ExecutionMode> {
    try {
      const response = await axios.get(`${API_BASE()}/${routeKey}/mode`, {
        headers: getAuthHeaders(),
      });

      if (response.data.success) {
        return response.data.data.executionMode;
      }
      return "direct"; // Default fallback
    } catch {
      return "direct"; // Default fallback
    }
  }

  /**
   * Check if route should use queue
   */
  async function shouldUseQueue(routeKey: string): Promise<boolean> {
    const mode = await getExecutionMode(routeKey);
    return mode === "queue";
  }

  /**
   * Create new route config
   */
  async function createConfig(
    input: RouteExecutionConfigInput
  ): Promise<RouteExecutionConfig | null> {
    isLoading.value = true;
    error.value = null;

    try {
      const response = await axios.post(API_BASE(), input, {
        headers: getAuthHeaders(),
      });

      if (response.data.success) {
        await fetchConfigs(); // Refresh list
        return response.data.data;
      }
      return null;
    } catch (err: any) {
      error.value = err.response?.data?.error || err.message;
      console.error("[RouteExecutionConfig] Create failed:", error.value);
      return null;
    } finally {
      isLoading.value = false;
    }
  }

  /**
   * Update route config
   */
  async function updateConfig(
    routeKey: string,
    updates: Partial<RouteExecutionConfig>
  ): Promise<RouteExecutionConfig | null> {
    isLoading.value = true;
    error.value = null;

    try {
      const response = await axios.put(`${API_BASE()}/${routeKey}`, updates, {
        headers: getAuthHeaders(),
      });

      if (response.data.success) {
        await fetchConfigs(); // Refresh list
        return response.data.data;
      }
      return null;
    } catch (err: any) {
      error.value = err.response?.data?.error || err.message;
      console.error("[RouteExecutionConfig] Update failed:", error.value);
      return null;
    } finally {
      isLoading.value = false;
    }
  }

  /**
   * Toggle execution mode (queue <-> direct)
   */
  async function toggleMode(
    routeKey: string
  ): Promise<RouteExecutionConfig | null> {
    try {
      const response = await axios.post(
        `${API_BASE()}/${routeKey}/toggle`,
        {},
        { headers: getAuthHeaders() }
      );

      if (response.data.success) {
        await fetchConfigs(); // Refresh list
        return response.data.data;
      }
      return null;
    } catch (err: any) {
      error.value = err.response?.data?.error || err.message;
      console.error("[RouteExecutionConfig] Toggle failed:", error.value);
      return null;
    }
  }

  /**
   * Delete route config
   */
  async function deleteConfig(routeKey: string): Promise<boolean> {
    try {
      const response = await axios.delete(`${API_BASE()}/${routeKey}`, {
        headers: getAuthHeaders(),
      });

      if (response.data.success) {
        await fetchConfigs(); // Refresh list
        return true;
      }
      return false;
    } catch (err: any) {
      error.value = err.response?.data?.error || err.message;
      console.error("[RouteExecutionConfig] Delete failed:", error.value);
      return false;
    }
  }

  return {
    configs,
    isLoading,
    error,
    fetchConfigs,
    getExecutionMode,
    shouldUseQueue,
    createConfig,
    updateConfig,
    toggleMode,
    deleteConfig,
  };
}
