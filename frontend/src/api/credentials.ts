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

export interface CredentialRequestContext {
  tenant_id: string;
}

function withTenantContext(context: CredentialRequestContext) {
  return {
    headers: {
      "x-tenant-id": context.tenant_id,
    },
    params: {
      tenant_id: context.tenant_id,
    },
  };
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

export async function getCredentialPlatforms(context: CredentialRequestContext) {
  const response = await apiClient.get<CredentialPlatformListResponse>(
    "/credentials/platforms",
    withTenantContext(context),
  );
  if (!response.success || !response.data) {
    throw new Error(response.error || "Failed to load credential platforms");
  }
  return response.data.platforms;
}

export async function getCredentialAudit(
  platform: string,
  context: CredentialRequestContext,
) {
  const response = await apiClient.get<CredentialAuditResponse>(
    `/credentials/platforms/${platform}/audit`,
    withTenantContext(context),
  );
  if (!response.success || !response.data) {
    throw new Error(response.error || "Failed to load credential history");
  }
  return response.data.events;
}

export async function saveManualToken(
  platform: string,
  payload: ManualTokenPayload,
  context: CredentialRequestContext,
) {
  const response = await apiClient.post(
    `/credentials/platforms/${platform}/connections/manual-token`,
    payload,
    withTenantContext(context),
  );
  if (!response.success) {
    throw new Error(response.error || "Failed to save manual token");
  }
}

export interface CredentialActionResult {
  status?: CredentialStatus;
  code?: string;
  message?: string;
  remote_revoke_status?: string;
  audit_event_id?: string;
}

export async function disconnectCredentialStore(
  platform: string,
  storeIdentifier: string,
  context: CredentialRequestContext,
) {
  const response = await apiClient.post<CredentialActionResult>(
    `/credentials/platforms/${platform}/connections/${encodeURIComponent(storeIdentifier)}/disconnect`,
    { reason: "disconnect_requested", revoke_remote: true },
    withTenantContext(context),
  );
  if (!response.success) {
    throw new Error(response.error || "disconnect_failed");
  }
  return response.data || { code: "disconnect_accepted" };
}
