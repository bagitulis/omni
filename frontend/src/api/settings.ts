import apiClient from "./client";

export interface ProfileUpdatePayload {
  name: string;
  email: string;
  phone?: string;
}

export interface PasswordChangePayload {
  current_password: string;
  new_password: string;
  confirm_password: string;
}

export interface GeneralSettingsPayload {
  language: string;
  timezone: string;
  notifications_email: boolean;
  notifications_browser: boolean;
  auto_sync: boolean;
  sync_interval: string;
}

export interface WebhookConfigPayload {
  custom_url: string;
  secret_key?: string;
}

/**
 * Settings API Service
 * Handles user profile, password, and general settings updates
 */

export const updateProfile = async (
  payload: ProfileUpdatePayload,
): Promise<void> => {
  const response = await apiClient.post("/auth/profile", payload);
  if (!response.success) {
    throw new Error(response.error || "Failed to update profile");
  }
};

export const changePassword = async (
  payload: PasswordChangePayload,
): Promise<void> => {
  const response = await apiClient.post("/auth/change-password", {
    current_password: payload.current_password,
    new_password: payload.new_password,
  });
  if (!response.success) {
    throw new Error(response.error || "Failed to change password");
  }
};

export const saveGeneralSettings = async (
  payload: GeneralSettingsPayload,
): Promise<void> => {
  const response = await apiClient.post("/settings/general", payload);
  if (!response.success) {
    throw new Error(response.error || "Failed to save settings");
  }
};

export const saveWebhookConfig = async (
  payload: WebhookConfigPayload,
): Promise<void> => {
  const response = await apiClient.post("/webhooks/config", payload);
  if (!response.success) {
    throw new Error(response.error || "Failed to save webhook configuration");
  }
};
