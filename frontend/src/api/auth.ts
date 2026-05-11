import axios from "axios";
import api from "./client";
import { useAuthStore } from "@/stores/authStore";
import type { ApiResponse } from "./client";
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

type AuthApiResponse = ApiResponse<LoginResponse> | DevLoginResponse;

function isWrappedAuthResponse(
  response: AuthApiResponse,
): response is ApiResponse<LoginResponse> {
  return "data" in response;
}

function getAuthErrorMessage(response: AuthApiResponse): string | undefined {
  return "error" in response && typeof response.error === "string"
    ? response.error
    : undefined;
}

function normalizeAuthResponse(
  response: AuthApiResponse,
  fallbackMessage: string,
): LoginResponse | DevLoginResponse {
  if (!response.success) {
    throw new Error(getAuthErrorMessage(response) || fallbackMessage);
  }

  const authPayload = isWrappedAuthResponse(response)
    ? response.data
    : response;
  if (!authPayload) {
    throw new Error(fallbackMessage);
  }

  if (!authPayload.user || (!authPayload.access_token && !authPayload.token)) {
    throw new Error(fallbackMessage);
  }

  return authPayload;
}

/**
 * Authentication API Service
 * Auth endpoints now use standard {success, data} wrapper format.
 */
export const login = async (payload: LoginPayload): Promise<LoginResponse> => {
  const response = await api.post<LoginResponse>("/auth/login", payload, {
    withCredentials: true,
  });
  return normalizeAuthResponse(response, "Login failed") as LoginResponse;
};

export const devLogin = async (
  payload: DevLoginPayload,
): Promise<DevLoginResponse> => {
  const response = await api.client.post<DevLoginResponse>(
    "/auth/dev-login",
    payload,
    {
      withCredentials: true,
    },
  );
  return normalizeAuthResponse(
    response.data,
    "Dev login failed",
  ) as DevLoginResponse;
};

export const logout = async (): Promise<void> => {
  await useAuthStore.getState().logout();
};

export const getCurrentUser = async (): Promise<User | null> => {
  try {
    const response = await api.get<User>("/auth/me");
    return response.data || null;
  } catch (err) {
    // Rethrow auth errors so callers can distinguish "no user" from "server error"
    if (axios.isAxiosError(err) && err.response?.status === 401) {
      return null; // Not authenticated — expected case
    }
    console.error("[Auth] Failed to fetch current user:", err);
    throw err; // Propagate unexpected errors
  }
};

export const changePassword = async (
  currentPassword: string,
  newPassword: string,
): Promise<void> => {
  const response = await api.post("/auth/change-password", {
    current_password: currentPassword,
    new_password: newPassword,
  });
  if (!response.success) {
    throw new Error(response.error || "Failed to change password");
  }
};

export interface DevInfoResponse {
  success: boolean;
  dev_mode: boolean;
  tenants: Array<{ id: string; name: string }>;
  username: string;
  error?: string;
}

export const getDevInfo = async (): Promise<DevInfoResponse> => {
  const response = await api.client.get<DevInfoResponse>("/auth/dev-info");
  return response.data;
};
