/**
 * User Management API Service
 * Uses centralized API service with tenant headers
 */

import api from "./api";

export interface UserData {
  username?: string;
  email?: string;
  password?: string;
  shopName?: string;
  role?: string;
}

export interface User extends UserData {
  id: string;
  tenantId: string;
  createdAt: string;
  updatedAt: string;
}

class UserManagementService {
  /**
   * List all users
   */
  async listAllUsers(): Promise<User[]> {
    const data = await api.get("/admin/users");
    return data.data?.users || [];
  }

  /**
   * Create new user
   */
  async createUser(user: UserData): Promise<User> {
    const data = await api.post("/admin/users", user);
    return data.data;
  }

  /**
   * Update user
   */
  async updateUser(tenantId: string, updates: Partial<UserData>): Promise<User> {
    const data = await api.patch(`/admin/users/${tenantId}`, updates);
    return data.data;
  }

  /**
   * Delete user
   */
  async deleteUser(tenantId: string): Promise<void> {
    await api.delete(`/admin/users/${tenantId}`);
  }

  /**
   * Change password
   */
  async changePassword(
    tenantId: string,
    currentPassword: string,
    newPassword: string
  ): Promise<void> {
    await api.post(`/admin/users/${tenantId}/change-password`, {
      currentPassword,
      newPassword,
    });
  }
}

export default new UserManagementService();
