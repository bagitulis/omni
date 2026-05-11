import api from "./client";
import type { ApiResponse } from "./client";

export interface UserResponse {
  id: string;
  username: string;
  email: string;
  role: string;
  failed_login_attempts: number;
  account_locked_until: string | null;
  created_at: string;
  updated_at: string;
}

export interface UsersListResponse {
  users: UserResponse[];
  total: number;
  page: number;
  limit: number;
}

export interface CreateUserPayload {
  username: string;
  email: string;
  password: string;
  role: string;
}

export interface UpdateUserPayload {
  username?: string;
  email?: string;
  role?: string;
}

export const usersApi = {
  list: async (params?: {
    page?: number;
    limit?: number;
  }): Promise<ApiResponse<UsersListResponse>> => {
    return api.get<UsersListResponse>("/users", { params });
  },
  get: async (id: string): Promise<ApiResponse<UserResponse>> => {
    return api.get<UserResponse>(`/users/${id}`);
  },
  create: async (
    data: CreateUserPayload,
  ): Promise<ApiResponse<UserResponse>> => {
    return api.post<UserResponse>("/users", data);
  },
  update: async (
    id: string,
    data: UpdateUserPayload,
  ): Promise<ApiResponse<UserResponse>> => {
    return api.put<UserResponse>(`/users/${id}`, data);
  },
  delete: async (id: string): Promise<ApiResponse<void>> => {
    return api.delete<void>(`/users/${id}`);
  },
  unlock: async (id: string): Promise<ApiResponse<void>> => {
    return api.post<void>(`/users/${id}/unlock`, {});
  },
};
