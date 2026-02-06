/**
 * Audit Service
 * Uses centralized API service with tenant headers
 */

import api from "./api";

export interface AuditLog {
  action: string;
  userId: string;
  targetUserId?: string;
  targetTenantId?: string;
  details?: Record<string, any>;
  timestamp?: string;
  status: "success" | "error";
  errorMessage?: string;
}

class AuditService {
  /**
   * Get audit logs for current user
   */
  async getUserAuditLogs(_userId: string, limit: number = 50): Promise<AuditLog[]> {
    const data = await api.get(`/audit/logs/user?limit=${limit}`);
    return data.data?.logs || [];
  }

  /**
   * Get audit logs for tenant (admin only)
   */
  async getTenantAuditLogs(limit: number = 100): Promise<AuditLog[]> {
    const data = await api.get(`/audit/logs/tenant?limit=${limit}`);
    return data.data?.logs || [];
  }

  /**
   * Cleanup old audit logs (owner only)
   */
  async cleanupOldLogs(daysOld: number = 90): Promise<void> {
    await api.post("/audit/cleanup", { daysOld });
  }
}

export default new AuditService();
