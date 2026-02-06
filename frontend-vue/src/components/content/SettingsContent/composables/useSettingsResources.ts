import { ref, Ref } from "vue";
import { formatTime, getResourceClass } from "@/utils/helpers";
import { getAuthHeaders, getApiBaseUrl } from "@/utils/apiHeaders";

interface BackendResources {
  memory: {
    used_mb: number;
    percent: number;
  };
  cpu: {
    percent: number;
  };
  threads: number;
  pid: number;
}

interface SystemResources {
  memory: {
    total_gb: number;
    used_gb: number;
    available_gb: number;
    percent: number;
  };
  cpu: {
    percent: number;
    count: number;
  };
}

interface ResourcesData {
  backend: BackendResources;
  system: SystemResources;
  timestamp: string;
}

export function useSettingsResources() {
  const resourcesData: Ref<ResourcesData | null> = ref(null);
  const resourcesLoading: Ref<boolean> = ref(false);
  const resourcesError: Ref<string> = ref("");

  const refreshResources = async (): Promise<void> => {
    resourcesLoading.value = true;
    resourcesError.value = "";
    try {
      const response = await fetch(
        getApiBaseUrl("/settings/resource-usage"),
        { headers: getAuthHeaders() }
      );
      const result = await response.json();
      if (result.success) {
        resourcesData.value = result.data;
      } else {
        resourcesError.value = result.error || "Failed to load resource data";
      }
    } catch (error) {
      console.error("Error fetching resources:", error);
      resourcesError.value = "Error connecting to server";
    } finally {
      resourcesLoading.value = false;
    }
  };

  return {
    resourcesData,
    resourcesLoading,
    resourcesError,
    refreshResources,
    formatTime,
    getResourceClass,
  };
}
