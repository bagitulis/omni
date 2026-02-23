import apiClient from "./client";
import { MonitorData, AutoFunctionConfig } from "@/types/scriptMonitor";

interface AutoFunctionsPayload {
  configs: AutoFunctionConfig[];
  total: number;
}

/**
 * Get monitor data (current job, queue, history)
 * Backend returns camelCase keys nested under `data`, so we map to snake_case MonitorData.
 */
export async function getMonitorData(): Promise<MonitorData> {
  const response = await apiClient.get<{
    currentJob: MonitorData["current_job"];
    pendingQueue: MonitorData["pending_queue"];
    recentHistory: MonitorData["recent_history"];
    totalPending: number;
    totalCompleted: number;
  }>("/jobs/monitor");
  if (!response.success) {
    throw new Error(response.error || "Failed to fetch monitor data");
  }
  const d = response.data!;
  return {
    current_job: d.currentJob ?? null,
    pending_queue: d.pendingQueue ?? [],
    recent_history: d.recentHistory ?? [],
    total_pending: d.totalPending ?? 0,
    total_completed: d.totalCompleted ?? 0,
  };
}

/**
 * Get auto-function configurations
 */
export async function getAutoFunctions(): Promise<AutoFunctionConfig[]> {
  const response = await apiClient.get<
    AutoFunctionConfig[] | AutoFunctionsPayload
  >("/jobs/auto-functions");
  if (!response.success) {
    throw new Error(response.error || "Failed to fetch auto-functions");
  }
  const data = response.data;
  if (Array.isArray(data)) {
    return data;
  }
  return data?.configs || [];
}

export interface AvailableAutoFunction {
  name: string;
  description: string;
}

/**
 * Get available auto-functions for dropdown
 */
export async function getAvailableAutoFunctions(): Promise<
  AvailableAutoFunction[]
> {
  const response = await apiClient.get<AvailableAutoFunction[]>(
    "/jobs/auto-functions/available",
  );
  if (!response.success) {
    throw new Error(
      response.error || "Failed to fetch available auto-functions",
    );
  }
  return response.data ?? [];
}

/**
 * Cancel a running job
 */
export async function cancelJob(jobId: string): Promise<void> {
  const response = await apiClient.post(`/jobs/cancel/${jobId}`);
  if (!response.success) {
    throw new Error(response.error || "Failed to cancel job");
  }
}

/**
 * Force cancel a stuck job
 */
export async function forceCancelJob(jobId: string): Promise<void> {
  const response = await apiClient.post(`/jobs/force-cancel/${jobId}`);
  if (!response.success) {
    throw new Error(response.error || "Failed to force cancel job");
  }
}

/**
 * Clear job history
 */
export async function clearHistory(): Promise<void> {
  const response = await apiClient.delete("/jobs/history");
  if (!response.success) {
    throw new Error(response.error || "Failed to clear history");
  }
}

/**
 * Update auto-function configuration
 */
export async function updateAutoFunction(
  name: string,
  config: Partial<AutoFunctionConfig>,
): Promise<void> {
  const response = await apiClient.put(`/jobs/auto-functions/${name}`, config);
  if (!response.success) {
    throw new Error(response.error || "Failed to update auto-function");
  }
}

/**
 * Create new auto-function
 */
export async function createAutoFunction(
  config: AutoFunctionConfig,
): Promise<void> {
  const response = await apiClient.post("/jobs/auto-functions", config);
  if (!response.success) {
    throw new Error(response.error || "Failed to create auto-function");
  }
}

/**
 * Delete auto-function
 */
export async function deleteAutoFunction(name: string): Promise<void> {
  const response = await apiClient.delete(`/jobs/auto-functions/${name}`);
  if (!response.success) {
    throw new Error(response.error || "Failed to delete auto-function");
  }
}

/**
 * Enable auto-function
 */
export async function enableAutoFunction(name: string): Promise<void> {
  const response = await apiClient.post(`/jobs/auto-functions/${name}/enable`);
  if (!response.success) {
    throw new Error(response.error || "Failed to enable auto-function");
  }
}

/**
 * Disable auto-function
 */
export async function disableAutoFunction(name: string): Promise<void> {
  const response = await apiClient.post(`/jobs/auto-functions/${name}/disable`);
  if (!response.success) {
    throw new Error(response.error || "Failed to disable auto-function");
  }
}

/**
 * Run auto-function manually (immediate execution)
 */
export async function runAutoFunction(name: string): Promise<void> {
  const response = await apiClient.post(`/jobs/auto-functions/${name}/run`);
  if (!response.success) {
    throw new Error(response.error || "Failed to run auto-function");
  }
}

/**
 * Cancel scheduled execution
 */
export async function cancelScheduled(id: number): Promise<void> {
  const response = await apiClient.post(
    `/jobs/auto-functions/${id}/cancel-scheduled`,
  );
  if (!response.success) {
    throw new Error(response.error || "Failed to cancel scheduled execution");
  }
}
