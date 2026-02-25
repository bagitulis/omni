import apiClient from "./client";

export interface AuditLog {
  action: string;
  user_id: string;
  target_user_id?: string;
  target_tenant_id?: string;
  details?: Record<string, unknown>;
  timestamp?: string;
  status: "success" | "error";
  error_message?: string;
}

interface AuditLogPayload {
  logs?: AuditLog[];
}

export async function getUserAuditLogs(
  limit: number = 50,
): Promise<AuditLog[]> {
  const response = await apiClient.get<AuditLogPayload>(
    `/audit/logs/user?limit=${limit}`,
  );
  if (!response.success) {
    throw new Error(response.error || "Failed to fetch user audit logs");
  }
  return response.data?.logs ?? [];
}

export async function getTenantAuditLogs(
  limit: number = 100,
): Promise<AuditLog[]> {
  const response = await apiClient.get<AuditLogPayload>(
    `/audit/logs/tenant?limit=${limit}`,
  );
  if (!response.success) {
    throw new Error(response.error || "Failed to fetch tenant audit logs");
  }
  return response.data?.logs ?? [];
}
