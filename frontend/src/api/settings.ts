import api from "./client";

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

/**
 * Settings API Service
 * Handles user profile, password, and general settings updates
 * TODO: Backend endpoints need to be implemented
 */

export const updateProfile = async (
  payload: ProfileUpdatePayload,
): Promise<void> => {
  // TODO: Implement backend endpoint POST/PUT /api/auth/profile
  const response = await api.post("/auth/profile", payload);
  if (!response.data.success) {
    throw new Error(response.data.error || "Failed to update profile");
  }
};

export const changePassword = async (
  payload: PasswordChangePayload,
): Promise<void> => {
  // TODO: Implement backend endpoint POST /api/auth/change-password
  const response = await api.post("/auth/change-password", {
    current_password: payload.current_password,
    new_password: payload.new_password,
  });
  if (!response.data.success) {
    throw new Error(response.data.error || "Failed to change password");
  }
};

export const saveGeneralSettings = async (
  payload: GeneralSettingsPayload,
): Promise<void> => {
  // TODO: Implement backend endpoint POST/PUT /api/settings/general
  const response = await api.post("/settings/general", payload);
  if (!response.data.success) {
    throw new Error(response.data.error || "Failed to save settings");
  }
};
