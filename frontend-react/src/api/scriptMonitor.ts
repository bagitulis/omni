import apiClient from "./client";
import { MonitorData, AutoFunctionConfig } from "@/types/scriptMonitor";

/**
 * Get monitor data (current job, queue, history)
 */
export async function getMonitorData(): Promise<MonitorData> {
  const response = await apiClient.get<MonitorData>("/jobs/monitor");
  if (!response.success) {
    throw new Error(response.error || "Failed to fetch monitor data");
  }
  return response.data as MonitorData;
}

/**
 * Get auto-function configurations
 */
export async function getAutoFunctions(): Promise<AutoFunctionConfig[]> {
  const response = await apiClient.get<AutoFunctionConfig[]>(
    "/jobs/auto-functions",
  );
  if (!response.success) {
    throw new Error(response.error || "Failed to fetch auto-functions");
  }
  return response.data || [];
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
