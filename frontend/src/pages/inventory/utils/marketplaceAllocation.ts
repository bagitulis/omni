import type { InventoryConfig } from "@/types/inventory";

export interface MarketplaceAllocationSettings {
  totalColumn: string;
  autoColumn: string;
  shopeeRatio: number;
  tiktokRatio: number;
}

export interface MarketplaceAllocationPreview {
  shopee: number;
  tiktok: number;
  lazada: number;
  total: number;
}

const STORAGE_KEY = "inventory_marketplace_allocation_settings_v1";

export const defaultMarketplaceAllocationSettings: MarketplaceAllocationSettings =
  {
    totalColumn: "",
    autoColumn: "",
    shopeeRatio: 0.6,
    tiktokRatio: 0.3,
  };

export function parseColumns(value: unknown): string[] {
  if (Array.isArray(value)) {
    return value.filter((item): item is string => typeof item === "string");
  }

  if (typeof value !== "string") {
    return [];
  }

  const trimmed = value.trim();
  if (!trimmed) return [];

  if (trimmed.startsWith("[")) {
    try {
      const parsed = JSON.parse(trimmed);
      if (Array.isArray(parsed)) {
        return parsed.filter(
          (item): item is string => typeof item === "string",
        );
      }
    } catch {
      return [];
    }
  }

  return trimmed
    .split(",")
    .map((item) => item.trim())
    .filter(Boolean);
}

export function encodeColumns(original: unknown, columns: string[]): string {
  const unique = Array.from(new Set(columns.filter(Boolean)));
  if (typeof original === "string" && original.trim().startsWith("[")) {
    return JSON.stringify(unique);
  }

  return unique.join(",");
}

export function inferAutoColumnFromSelection(
  selectedColumns: string[],
): string {
  return (
    selectedColumns.find((column) => column.toLowerCase().includes("auto")) ||
    ""
  );
}

export function loadMarketplaceAllocationSettings(): MarketplaceAllocationSettings {
  if (typeof window === "undefined") {
    return defaultMarketplaceAllocationSettings;
  }

  const raw = window.localStorage.getItem(STORAGE_KEY);
  if (!raw) {
    return defaultMarketplaceAllocationSettings;
  }

  try {
    const parsed = JSON.parse(raw) as Partial<MarketplaceAllocationSettings>;
    return {
      totalColumn: parsed.totalColumn || "",
      autoColumn: parsed.autoColumn || "",
      shopeeRatio:
        typeof parsed.shopeeRatio === "number"
          ? parsed.shopeeRatio
          : defaultMarketplaceAllocationSettings.shopeeRatio,
      tiktokRatio:
        typeof parsed.tiktokRatio === "number"
          ? parsed.tiktokRatio
          : defaultMarketplaceAllocationSettings.tiktokRatio,
    };
  } catch {
    return defaultMarketplaceAllocationSettings;
  }
}

export function saveMarketplaceAllocationSettings(
  settings: MarketplaceAllocationSettings,
): void {
  if (typeof window === "undefined") {
    return;
  }

  window.localStorage.setItem(STORAGE_KEY, JSON.stringify(settings));
}

export function deriveMarketplaceAllocationSettings(
  config: InventoryConfig | null | undefined,
): MarketplaceAllocationSettings {
  const saved = loadMarketplaceAllocationSettings();
  const selectedColumns = parseColumns(config?.selected_columns);
  const inferredAuto = inferAutoColumnFromSelection(selectedColumns);

  return {
    totalColumn: config?.key_column || saved.totalColumn,
    autoColumn: saved.autoColumn || inferredAuto,
    shopeeRatio: saved.shopeeRatio,
    tiktokRatio: saved.tiktokRatio,
  };
}

export function getRecordValueByColumn(
  rowData: Record<string, unknown>,
  columnName: string,
): unknown {
  if (!columnName) {
    return undefined;
  }

  if (columnName in rowData) {
    return rowData[columnName];
  }

  const target = columnName.trim().toLowerCase();
  const matchedEntry = Object.entries(rowData).find(
    ([key]) => key.trim().toLowerCase() === target,
  );

  return matchedEntry?.[1];
}

function toBooleanLike(value: unknown): boolean {
  if (typeof value === "boolean") {
    return value;
  }

  if (typeof value === "number") {
    return value !== 0;
  }

  if (typeof value === "string") {
    const normalized = value.trim().toLowerCase();
    return ["1", "true", "yes", "y", "on"].includes(normalized);
  }

  return false;
}

function toSafeNumber(value: unknown): number {
  const parsed = Number(value);
  if (!Number.isFinite(parsed)) {
    return 0;
  }

  return Math.max(0, Math.floor(parsed));
}

export function calculateMarketplaceAllocation(
  total: number,
  autoMode: boolean,
  settings: MarketplaceAllocationSettings,
): MarketplaceAllocationPreview {
  if (total <= 0) {
    return { shopee: 0, tiktok: 0, lazada: 0, total: 0 };
  }

  if (autoMode) {
    return { shopee: total, tiktok: total, lazada: total, total };
  }

  const shopee = Math.min(Math.ceil(settings.shopeeRatio * total), total);
  const remainingAfterShopee = Math.max(0, total - shopee);
  const tiktok = Math.min(
    Math.ceil(settings.tiktokRatio * total),
    remainingAfterShopee,
  );
  const lazada = Math.max(0, total - shopee - tiktok);

  return { shopee, tiktok, lazada, total };
}

export function resolveMarketplaceAllocationForRecord(
  rowData: Record<string, unknown>,
  settings: MarketplaceAllocationSettings,
): MarketplaceAllocationPreview {
  const directShopee = getRecordValueByColumn(rowData, "shopee");
  const directTiktok = getRecordValueByColumn(rowData, "tiktok");
  const directLazada = getRecordValueByColumn(rowData, "lazada");

  const hasDirectValues =
    directShopee !== undefined ||
    directTiktok !== undefined ||
    directLazada !== undefined;
  if (hasDirectValues) {
    const shopee = toSafeNumber(directShopee);
    const tiktok = toSafeNumber(directTiktok);
    const lazada = toSafeNumber(directLazada);
    return {
      shopee,
      tiktok,
      lazada,
      total: shopee + tiktok + lazada,
    };
  }

  const totalValue = toSafeNumber(
    getRecordValueByColumn(rowData, settings.totalColumn),
  );
  const autoMode = toBooleanLike(
    getRecordValueByColumn(rowData, settings.autoColumn),
  );

  return calculateMarketplaceAllocation(totalValue, autoMode, settings);
}
