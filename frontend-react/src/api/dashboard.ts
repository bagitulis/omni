import apiClient from "./client";
import { DashboardData } from "../types/dashboard";

export async function getDashboardData(): Promise<DashboardData> {
  const response = await apiClient.get<DashboardData>("/dashboard");
  if (!response.success) {
    throw new Error(response.error || "Failed to fetch dashboard data");
  }
  return response.data!;
}
