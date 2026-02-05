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
 */
export const login = async (payload: LoginPayload): Promise<LoginResponse> => {
  const response = await api.post<LoginResponse>("/auth/login", payload);
  if (!response.success || !response.data) {
    throw new Error(response.error || "Login failed");
  }
  return response.data;
};

export const devLogin = async (
  payload: DevLoginPayload,
): Promise<DevLoginResponse> => {
  const response = await api.post<DevLoginResponse>("/auth/dev-login", payload);
  if (!response.success || !response.data) {
    throw new Error(response.error || "Dev login failed");
  }
  return response.data;
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
