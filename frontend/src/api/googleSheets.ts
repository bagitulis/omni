import apiClient from "./client";
import type {
  SpreadsheetLinks,
  GoogleSheetsSettings,
  ValidationResult,
  ValidateLinkPayload,
  SaveLinksPayload,
  UpdateSettingsPayload,
} from "@/types/googleSheets";

/**
 * Get saved Google Sheets links
 */
export async function getSavedLinks(): Promise<SpreadsheetLinks> {
  const response = await apiClient.get<SpreadsheetLinks>(
    "/google/settings/saved-links",
  );
  if (!response.success) {
    throw new Error(response.error || "Failed to fetch saved links");
  }
  return response.data as SpreadsheetLinks;
}

/**
 * Get detailed Google Sheets settings (links + metadata)
 */
export async function getDetailedSettings(): Promise<GoogleSheetsSettings> {
  const response = await apiClient.get<GoogleSheetsSettings>(
    "/google/settings/detailed",
  );
  if (!response.success) {
    throw new Error(response.error || "Failed to fetch detailed settings");
  }
  return response.data as GoogleSheetsSettings;
}

/**
 * Validate Google Sheets link
 */
export async function validateLink(
  payload: ValidateLinkPayload,
): Promise<ValidationResult> {
  const response = await apiClient.post<ValidationResult>(
    "/google/settings/validate-link",
    payload,
  );
  if (!response.success) {
    throw new Error(response.error || "Failed to validate link");
  }
  return response.data as ValidationResult;
}

/**
 * Save Google Sheets links
 */
export async function saveLinks(payload: SaveLinksPayload): Promise<void> {
  const response = await apiClient.post("/google/settings/save-links", payload);
  if (!response.success) {
    throw new Error(response.error || "Failed to save links");
  }
}

/**
 * Update detailed Google Sheets settings
 */
export async function updateSettings(
  payload: UpdateSettingsPayload,
): Promise<void> {
  const response = await apiClient.post(
    "/google/settings/update-detailed",
    payload,
  );
  if (!response.success) {
    throw new Error(response.error || "Failed to update settings");
  }
}
