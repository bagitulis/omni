export function formatAmount(amt: number, cur: string): string {
  return new Intl.NumberFormat("id-ID", {
    style: "currency",
    currency: cur || "IDR",
    minimumFractionDigits: 0,
  }).format(amt);
}

export function getCountdown(shipByDate?: number): string {
  if (!shipByDate) return "-";
  const now = Date.now();
  const deadline = shipByDate * 1000;
  const diff = deadline - now;
  if (diff <= 0) return "Overdue";

  const hours = Math.floor(diff / (1000 * 60 * 60));
  const minutes = Math.floor((diff % (1000 * 60 * 60)) / (1000 * 60));

  if (hours >= 24) {
    const days = Math.floor(hours / 24);
    return `${days}d ${hours % 24}h`;
  }
  return `${hours}h ${minutes}m`;
}

export function getCountdownColor(
  token: GlobalToken,
  shipByDate?: number,
): string {
  if (!shipByDate) return token.colorTextSecondary;
  const now = Date.now();
  const deadline = shipByDate * 1000;
  const diff = deadline - now;
  if (diff <= 0) return token.colorError;
  if (diff <= 24 * 60 * 60 * 1000) return token.colorWarning;
  return token.colorTextSecondary;
}

/**
 * Check if an order can be shipped based on platform-specific statuses.
 * Matches backend validateOrderStatus logic:
 * - Shopee: READY_TO_SHIP
 * - TikTok: AWAITING_SHIPMENT
 * - Lazada: pending | packed
 */
export function canShipOrder(
  status: string | undefined,
  platform: string | undefined,
): boolean {
  if (!status) return false;
  const s = status.toUpperCase();
  const p = (platform || "").toLowerCase();

  switch (p) {
    case "tiktok":
      return s === "AWAITING_SHIPMENT";
    case "lazada":
      return s === "PENDING" || s === "PACKED";
    case "shopee":
    default:
      return s === "READY_TO_SHIP";
  }
}
import type { GlobalToken } from "antd";
