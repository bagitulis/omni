import api from "./client";
import type { ApiResponse } from "./client";

export interface TenantOverview {
  id: string;
  shop_name: string;
  user_count: number;
  owners: number;
  admins: number;
  users: number;
  error?: string;
}

export interface DeveloperOverviewResponse {
  tenants: TenantOverview[];
  total: number;
}

export const developerApi = {
  getOverview: async (): Promise<ApiResponse<DeveloperOverviewResponse>> => {
    return api.get<DeveloperOverviewResponse>("/dev/overview");
  },
};
