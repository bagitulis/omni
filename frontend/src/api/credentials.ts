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
  // Phase 8 additions (Shopee-focused; other platforms leave these unset).
  partner_id?: number;
  test_partner_id?: number;
  test_configured?: boolean;
  partner_key_expires_at?: string;
  test_partner_key_expires_at?: string;
  app_status?: "online" | "offline";
  active_partner_env?: "live" | "test";
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

export interface OAuthInitiatePayload {
  intent: string;
  store_identifier?: string;
  redirect_path: string;
}

export interface OAuthInitiateResult {
  auth_url: string;
  attempt_id: string;
  expires_at: string;
}

export async function initiateOAuth(
  platform: string,
  payload: OAuthInitiatePayload,
  context: CredentialRequestContext,
) {
  const response = await apiClient.post<OAuthInitiateResult>(
    `/credentials/platforms/${platform}/connections/oauth/initiate`,
    payload,
    withTenantContext(context),
  );
  if (!response.success || !response.data) {
    throw new Error(response.error || "Failed to initiate OAuth");
  }
  return response.data;
}

// ── App Credential Management ────────────────────────────────────────

export interface ShopeeAppCredentialPayload {
  partner_id: number;
  partner_key: string;
  region: string;
  reason?: string;
  // Phase 8 — optional sandbox pair + expiry + operational toggles.
  test_partner_id?: number;
  test_partner_key?: string;
  partner_key_expires_at?: string;       // RFC3339
  test_partner_key_expires_at?: string;  // RFC3339
  app_status?: "online" | "offline";
  active_partner_env?: "live" | "test";
}

export interface LazadaAppCredentialPayload {
  app_key: string;
  app_secret: string;
  region: string;
  reason?: string;
}

export interface TiktokAppCredentialPayload {
  app_key: string;
  app_secret: string;
  reason?: string;
}

export type AppCredentialUpsertPayload =
  | ShopeeAppCredentialPayload
  | LazadaAppCredentialPayload
  | TiktokAppCredentialPayload;

export interface AppCredentialMutationResult {
  platform: string;
  status: string;
  configured: boolean;
  secret_mask?: string;
  audit_event_id?: string;
}

export async function upsertAppCredential(
  platform: string,
  payload: AppCredentialUpsertPayload,
  context: CredentialRequestContext,
) {
  const response = await apiClient.put<AppCredentialMutationResult>(
    `/credentials/platforms/${platform}/app`,
    payload,
    withTenantContext(context),
  );
  if (!response.success || !response.data) {
    throw new Error(response.error || "Failed to save app credential");
  }
  return response.data;
}

export async function rotateAppCredential(
  platform: string,
  payload: AppCredentialUpsertPayload,
  context: CredentialRequestContext,
) {
  const response = await apiClient.post<AppCredentialMutationResult>(
    `/credentials/platforms/${platform}/app/rotate`,
    payload,
    withTenantContext(context),
  );
  if (!response.success || !response.data) {
    throw new Error(response.error || "Failed to rotate app credential");
  }
  return response.data;
}
