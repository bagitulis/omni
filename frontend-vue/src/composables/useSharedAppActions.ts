/**
 * Shared App Actions Composable
 * Extracted to break circular dependency between stores
 * SRP: Only handles shared app actions (logging, operations)
 */
import { ref, Ref } from "vue";
import apiService from "@/services/api";
import { useToast } from "@/composables/useToast";

// Shared logging state (singleton pattern)
const logs: Ref<string[]> = ref([]);

/**
 * Add log message with timestamp
 */
export function addLog(message: string): void {
  const timestamp = new Date().toLocaleTimeString();
  logs.value.push(`[${timestamp}] ${message}`);

  // Keep last 100 logs
  if (logs.value.length > 100) {
    logs.value = logs.value.slice(-100);
  }
}

/**
 * Clear all logs
 */
export function clearLogs(): void {
  logs.value = [];
}

/**
 * Get logs ref
 */
export function getLogs(): Ref<string[]> {
  return logs;
}

/**
 * Execute operation via API
 */
export async function executeOperation(
  operation: string,
  params: Record<string, any> = {}
): Promise<any> {
  const toast = useToast();

  try {
    addLog(`🔄 Executing: ${operation}`);
    const response = await apiService.executeOperation(operation, params);

    if ((response as any).success) {
      addLog(`✅ ${operation} completed`);
    } else {
      addLog(
        `❌ ${operation} failed: ${(response as any).message || "Unknown error"}`
      );
    }

    return response;
  } catch (error: any) {
    addLog(`❌ ${operation} error: ${error.message}`);
    toast.error(`Operation failed: ${error.message}`);
    throw error;
  }
}
