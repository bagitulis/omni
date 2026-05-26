import type { CredentialStatus } from "@/api/credentials";

export const CREDENTIAL_STATUS_COLOR: Record<CredentialStatus, string> = {
  disconnected: "default",
  incomplete: "warning",
  connected: "success",
  expired: "error",
  refresh_failed: "error",
  action_required: "warning",
};

export const CREDENTIAL_STATUS_LABEL: Record<CredentialStatus, string> = {
  disconnected: "Disconnected",
  incomplete: "Incomplete",
  connected: "Connected",
  expired: "Expired",
  refresh_failed: "Refresh Failed",
  action_required: "Action Required",
};

export function formatRegion(region?: string): string {
  if (!region) return "-";
  return region === "id" ? "Indonesia" : region.toUpperCase();
}
