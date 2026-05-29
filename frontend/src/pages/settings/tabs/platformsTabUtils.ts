import type { CredentialPlatformSummary } from "@/api/credentials";
import type { PlatformConnectionSummary } from "../components/PlatformCard";

export const PLATFORM_NAMES: Record<string, string> = {
  shopee: "Shopee",
  tiktok: "TikTok Shop",
  lazada: "Lazada",
};

const PRIVILEGED_ROLES = new Set(["developer", "admin", "owner"]);

export function sanitizeStatusText(value: unknown, fallback: string) {
  if (typeof value !== "string") return fallback;
  const sanitized = value.replace(/[^a-zA-Z0-9_ -]/g, "").slice(0, 80).trim();
  return sanitized || fallback;
}

export function getPrivilegedActionReason(role?: string) {
  return PRIVILEGED_ROLES.has(role || "")
    ? null
    : "Developer or system admin role required for app credentials and manual tokens.";
}

export function getStoreActionReason(tenantId: string | null, role?: string) {
  if (!tenantId) return "Tenant context required before managing store connections.";
  if (!role) return "Authenticated role required before managing store connections.";
  return null;
}

export function formatCredentialExpiry(expiresAt?: string) {
  if (!expiresAt) return null;
  const date = new Date(expiresAt);
  const now = new Date();
  const days = Math.ceil((date.getTime() - now.getTime()) / (1000 * 60 * 60 * 24));
  if (days < 0) return "Expired";
  if (days === 0) return "Expires today";
  if (days <= 7) return `Expires in ${days} days`;
  return `Expires ${date.toLocaleDateString()}`;
}

export function toConnectionSummary(platform: CredentialPlatformSummary): PlatformConnectionSummary {
  const store = platform.stores[0];
  return {
    platform: platform.platform,
    connected: platform.status === "connected",
    store_identifier: store?.store_identifier,
    store_name: store?.store_name,
    expires_at: store?.expires_at,
    expires_soon: platform.status === "expired",
    status: platform.status,
    region: platform.region,
    last_refresh_at: store?.last_refresh_at,
    refresh_status: store?.refresh_status,
  };
}
