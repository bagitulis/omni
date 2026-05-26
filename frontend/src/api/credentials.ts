import apiClient from "./client";

export type CredentialStatus =
  | "disconnected"
  | "incomplete"
  | "connected"
  | "expired"
  | "refresh_failed"
  | "action_required";

export interface CredentialStoreConnection {
  store_identifier: string;
  store_name?: string;
  status: CredentialStatus;
  expires_at?: string;
  last_refresh_at?: string;
  refresh_status?: CredentialStatus;
  masked_access_token?: string;
  masked_refresh_token?: string;
  region?: string;
}

export interface CredentialAppConfig {
  platform: string;
  region?: string;
  status: CredentialStatus;
  app_configured: boolean;
  secret_mask: string;
  updated_at?: string;
  updated_by?: string;
  store_identifier?: string | null;
  fields?: string[];
}

export interface CredentialPlatformSummary {
  platform: string;
  region?: string;
  status: CredentialStatus;
  app_configured: boolean;
  secret_mask?: string;
  app_secret_mask?: string;
  stores: CredentialStoreConnection[];
  app_config?: CredentialAppConfig | null;
  audit_summary?: {
    last_event_type?: string;
    last_event_at?: string;
  };
}

export interface CredentialPlatformListResponse {
  platforms: CredentialPlatformSummary[];
}

export interface CredentialAuditEvent {
  event_type: string;
  platform: string;
  store_identifier_mask?: string;
  code?: string;
  reason?: string;
  created_at: string;
}

export interface CredentialAuditResponse {
  events: CredentialAuditEvent[];
}

export interface ManualTokenPayload {
  store_identifier: string;
  region?: string;
  access_token: string;
  refresh_token?: string;
  expires_at?: string;
  shop_cipher?: string;
  reason: string;
}

export async function getCredentialPlatforms() {
  const response = await apiClient.get<CredentialPlatformListResponse>(
    "/credentials/platforms",
  );
  if (!response.success || !response.data) {
    throw new Error(response.error || "Failed to load credential platforms");
  }
  return response.data.platforms;
}

export async function getCredentialAudit(platform: string) {
  const response = await apiClient.get<CredentialAuditResponse>(
    `/credentials/platforms/${platform}/audit`,
  );
  if (!response.success || !response.data) {
    throw new Error(response.error || "Failed to load credential history");
  }
  return response.data.events;
}

export async function saveManualToken(
  platform: string,
  payload: ManualTokenPayload,
) {
  const response = await apiClient.post(
    `/credentials/platforms/${platform}/connections/manual-token`,
    payload,
  );
  if (!response.success) {
    throw new Error(response.error || "Failed to save manual token");
  }
}
