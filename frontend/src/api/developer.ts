import api from "./client";
import type { ApiResponse } from "./client";

// --- Existing interfaces ---

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

// --- New interfaces for developer panel ---

export interface TenantDetail {
  id: string;
  name: string;
  is_active: boolean;
  user_count: number;
  created_at: string;
}

export interface CrossTenantUser {
  id: string;
  username: string;
  email: string;
  role: string;
  tenant_id: string;
  tenant_name: string;
  status: string;
}

export interface SystemHealth {
  db_status: string;
  memory_usage: number;
  goroutines: number;
  uptime: string;
  components: { name: string; status: string }[];
}

export interface AuditLogEntry {
  id: string;
  user_id: string;
  action: string;
  details: string;
  created_at: string;
  tenant_id: string;
}

export interface BulkOperationResult {
  success_count: number;
  failure_count: number;
  errors: { user_id: string; error: string }[];
}

export interface AuditLogParams {
  page?: number;
  limit?: number;
  action?: string;
  user_id?: string;
  date_from?: string;
  date_to?: string;
}

export interface EnvironmentInfo {
  go_version: string;
  environment: string;
  version: string;
  build_time: string;
  db_driver: string;
  node_env: string;
  api_base_url: string;
  dev_login_enabled: string;
  active_tenants_count: number;
}

// --- API functions ---

export const developerApi = {
  getOverview: async (): Promise<ApiResponse<DeveloperOverviewResponse>> => {
    return api.get<DeveloperOverviewResponse>("/dev/overview");
  },

  resetUserPassword: async (userId: string, newPassword: string): Promise<ApiResponse<void>> => {
    return api.post<void>("/dev/reset-password", { user_id: userId, new_password: newPassword });
  },

  getTenants: async (): Promise<ApiResponse<TenantDetail[]>> => {
    return api.get<TenantDetail[]>("/dev/tenants");
  },

  createTenant: async (name: string): Promise<ApiResponse<TenantDetail>> => {
    return api.post<TenantDetail>("/dev/tenants", { name });
  },

  deactivateTenant: async (tenantId: string): Promise<ApiResponse<void>> => {
    return api.delete<void>(`/dev/tenants/${tenantId}`);
  },

  searchUsers: async (query: string): Promise<ApiResponse<CrossTenantUser[]>> => {
    return api.get<CrossTenantUser[]>("/dev/users/search", { params: { q: query } });
  },

  bulkResetPasswords: async (userIds: string[], newPassword: string): Promise<ApiResponse<BulkOperationResult>> => {
    return api.post<BulkOperationResult>("/dev/users/bulk-reset-password", { user_ids: userIds, new_password: newPassword });
  },

  bulkDisableUsers: async (userIds: string[]): Promise<ApiResponse<BulkOperationResult>> => {
    return api.post<BulkOperationResult>("/dev/users/bulk-disable", { user_ids: userIds });
  },

  getSystemHealth: async (): Promise<ApiResponse<SystemHealth>> => {
    return api.get<SystemHealth>("/monitoring/health/detailed");
  },

  getAuditLogs: async (params?: AuditLogParams): Promise<ApiResponse<AuditLogEntry[]>> => {
    return api.get<AuditLogEntry[]>("/dev/audit-logs", { params });
  },

  getEnvironmentInfo: async (): Promise<ApiResponse<EnvironmentInfo>> => {
    return api.get<EnvironmentInfo>("/dev/environment");
  },
};
