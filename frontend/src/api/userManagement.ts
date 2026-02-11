import apiClient from "./client";

export interface UserData {
  username?: string;
  email?: string;
  password?: string;
  shop_name?: string;
  role?: string;
}

export interface User {
  id: string;
  tenant_id: string;
  username?: string;
  email?: string;
  shop_name?: string;
  role?: string;
  created_at: string;
  updated_at: string;
}

interface UserListPayload {
  users?: User[];
}

export async function listAllUsers(): Promise<User[]> {
  const response = await apiClient.get<UserListPayload>("/users");
  if (!response.success) {
    throw new Error(response.error || "Failed to fetch users");
  }
  return response.data?.users ?? [];
}

export async function createUser(user: UserData): Promise<User> {
  const response = await apiClient.post<User>("/users", user);
  if (!response.success || !response.data) {
    throw new Error(response.error || "Failed to create user");
  }
  return response.data;
}

export async function updateUser(
  tenantId: string,
  updates: Partial<UserData>,
): Promise<User> {
  const response = await apiClient.patch<User>(`/users/${tenantId}`, updates);
  if (!response.success || !response.data) {
    throw new Error(response.error || "Failed to update user");
  }
  return response.data;
}

export async function deleteUser(tenantId: string): Promise<void> {
  const response = await apiClient.delete(`/users/${tenantId}`);
  if (!response.success) {
    throw new Error(response.error || "Failed to delete user");
  }
}

export async function changePassword(
  tenantId: string,
  currentPassword: string,
  newPassword: string,
): Promise<void> {
  const response = await apiClient.post(`/users/${tenantId}/change-password`, {
    current_password: currentPassword,
    new_password: newPassword,
  });
  if (!response.success) {
    throw new Error(response.error || "Failed to change password");
  }
}
