import { Tag } from "antd";
import type { InventoryRecord } from "@/types/inventory";
import {
  resolveMarketplaceAllocationForRecord,
  type MarketplaceAllocationSettings,
} from "../utils/marketplaceAllocation";

export function estimateInventoryColumnWidth(
  columnName: string,
  records: InventoryRecord[],
): number {
  const normalized = columnName.trim().toLowerCase();

  if (normalized === "nama barang") {
    return 360;
  }

  if (normalized === "nama variasi") {
    return 260;
  }

  if (normalized === "harga" || normalized === "price") {
    return 130;
  }

  if (normalized === "total") {
    return 90;
  }

  if (normalized === "stock" || normalized === "stok") {
    return 110;
  }

  const sampleRecords = records.slice(0, 120);
  const longestValueLength = sampleRecords.reduce((max, record) => {
    const value = record.data?.[columnName];
    const currentLength = value == null ? 0 : String(value).trim().length;
    return Math.max(max, currentLength);
  }, columnName.length);

  const estimated = longestValueLength * 8 + 40;
  return Math.min(280, Math.max(120, estimated));
}

export function renderMarketplaceCell(
  record: InventoryRecord,
  platformName: "shopee" | "tiktok" | "lazada",
  marketplaceSettings: MarketplaceAllocationSettings,
) {
  const allocation = resolveMarketplaceAllocationForRecord(
    record.data || {},
    marketplaceSettings,
  );
  const value = allocation[platformName];

  if (value === null || value === undefined || String(value).trim() === "") {
    return <span style={{ color: "var(--color-text-secondary)", fontSize: 11 }}>-</span>;
  }

  return (
    <Tag
      color="blue"
      style={{
        fontSize: 11,
        margin: 0,
        minWidth: 30,
        display: "inline-flex",
        justifyContent: "center",
        textAlign: "center",
      }}
    >
      {String(value)}
    </Tag>
  );
}

export function resolveSyncStatus(syncStatus?: string): {
  color: string;
  label: string;
} {
  const statusMap: Record<string, { color: string; label: string }> = {
    synced: { color: "green", label: "Synced" },
    not_synced: { color: "default", label: "Not Synced" },
    error: { color: "red", label: "Error" },
  };

  return statusMap[syncStatus || "not_synced"] || statusMap.not_synced;
}
