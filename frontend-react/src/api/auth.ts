import api from "./client";
import { STORAGE_KEYS } from "@/lib/constants";
import type { LoginResponse, User } from "@/types/auth";

export interface LoginPayload {
  username?: string;
  password?: string;
  recaptchaToken?: string;
}

export interface DevLoginPayload {
  tenant_id: string;
}

export interface DevLoginResponse extends LoginResponse {
  dev_mode: boolean;
  success: boolean;
}

/**
 * Authentication API Service
 * Note: Auth endpoints return flat responses (not wrapped in 'data' field)
 */
export const login = async (payload: LoginPayload): Promise<LoginResponse> => {
  // Auth endpoints return flat response: {success, token, user, tenant_id}
  const response = await api.client.post("/auth/login", payload);
  const result = response.data;
  if (!result.success) {
    throw new Error(result.error || "Login failed");
  }
  return result as LoginResponse;
};

export const devLogin = async (
  payload: DevLoginPayload,
): Promise<DevLoginResponse> => {
  // Auth endpoints return flat response: {success, token, user, tenant_id, dev_mode}
  const response = await api.client.post("/auth/dev-login", payload);
  const result = response.data;
  if (!result.success) {
    throw new Error(result.error || "Dev login failed");
  }
  return result as DevLoginResponse;
};

export const logout = async (): Promise<void> => {
  try {
    await api.post("/auth/logout");
  } catch (error) {
    console.error("Logout error:", error);
  } finally {
    // Clear all auth-related localStorage items
    localStorage.removeItem(STORAGE_KEYS.AUTH_TOKEN);
    localStorage.removeItem(STORAGE_KEYS.AUTH_USER);
    localStorage.removeItem(STORAGE_KEYS.TENANT_ID);
    localStorage.removeItem(STORAGE_KEYS.USER_ROLE);
    localStorage.removeItem(STORAGE_KEYS.USER_NAME);
  }
};

export const getCurrentUser = async (): Promise<User | null> => {
  try {
    const response = await api.get<User>("/auth/me");
    return response.data || null;
  } catch {
    return null;
  }
};
