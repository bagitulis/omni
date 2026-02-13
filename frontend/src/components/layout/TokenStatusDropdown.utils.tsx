import {
  CheckCircleFilled,
  CloseCircleFilled,
  WarningFilled,
} from "@ant-design/icons";
import type { GlobalToken } from "antd";
import { type TokenStatusData } from "./TokenStatusDropdown.types";

export function getTokenStatusType(
  data: TokenStatusData | undefined,
): "valid" | "expiring" | "expired" | "unknown" | "not_configured" {
  if (!data) return "unknown";
  if (data.status === "not_configured") return "not_configured";
  if (data.isExpired || data.status === "expired" || data.valid === false) {
    return "expired";
  }
  if (data.expiresAt) {
    const expiresAt = new Date(data.expiresAt).getTime();
    const diff = expiresAt - Date.now();
    if (diff <= 0) return "expired";
    const hoursLeft = diff / (1000 * 60 * 60);
    if (hoursLeft < 24 && hoursLeft > 0) return "expiring";
  }
  return "valid";
}

export function formatTimeRemaining(dateString: string | null): string {
  if (!dateString) return "Unknown";
  const expiresAt = new Date(dateString);
  const diffMs = expiresAt.getTime() - Date.now();
  if (diffMs <= 0) return "Expired";

  const days = Math.floor(diffMs / (1000 * 60 * 60 * 24));
  const hours = Math.floor((diffMs % (1000 * 60 * 60 * 24)) / (1000 * 60 * 60));
  const minutes = Math.floor((diffMs % (1000 * 60 * 60)) / (1000 * 60));

  if (days > 0) return `${days}d ${hours}h`;
  if (hours > 0) return `${hours}h ${minutes}m`;
  return `${minutes}m`;
}

export function getStatusIcon(status: string, token: GlobalToken) {
  switch (status) {
    case "valid":
      return <CheckCircleFilled style={{ color: token.colorSuccess }} />;
    case "expiring":
      return <WarningFilled style={{ color: token.colorWarning }} />;
    case "expired":
      return <CloseCircleFilled style={{ color: token.colorError }} />;
    case "not_configured":
      return <WarningFilled style={{ color: token.colorBorder }} />;
    default:
      return <WarningFilled style={{ color: token.colorTextSecondary }} />;
  }
}

export function getStatusColor(status: string): string {
  switch (status) {
    case "valid":
      return "success";
    case "expiring":
      return "warning";
    case "expired":
      return "error";
    case "not_configured":
      return "default";
    default:
      return "default";
  }
}
