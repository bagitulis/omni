import { Checkbox, InputNumber } from "antd";
import type { ColumnType } from "antd/es/table";
import type { Platform, UnifiedProductRow } from "@/types/shared";
import type { Dispatch, SetStateAction } from "react";

export interface PerPlatformConfig {
  [sku: string]: {
    stock: number;
    platforms: Record<Platform, boolean>;
  };
}

export const PLATFORM_OPTIONS: Array<{ key: Platform; label: string }> = [
  { key: "shopee", label: "🟠 Shopee" },
  { key: "tiktok", label: "⬛ TikTok" },
  { key: "lazada", label: "🔵 Lazada" },
];

export interface PerPlatformRow {
  key: string;
  sku: string;
  stock: number;
}

export const getDefaultPerPlatformConfig = (
  selectedProducts: UnifiedProductRow[],
  linkedPlatformsBySku: Record<string, Record<Platform, boolean>>,
): PerPlatformConfig => {
  const config: PerPlatformConfig = {};
  for (const product of selectedProducts) {
    for (const sku of product.skus) {
      const linkedPlatforms = linkedPlatformsBySku[sku.seller_sku] ?? {
        shopee: false,
        tiktok: false,
        lazada: false,
      };

      config[sku.seller_sku] = {
        stock: sku.stock,
        platforms: { ...linkedPlatforms },
      };
    }
  }
  return config;
};

export const getPerPlatformColumns = (
  linkedPlatformsBySku: Record<string, Record<Platform, boolean>>,
  perPlatformConfig: PerPlatformConfig,
  setPerPlatformConfig: Dispatch<SetStateAction<PerPlatformConfig>>,
): ColumnType<PerPlatformRow>[] => {
  return [
    {
      title: "SKU",
      dataIndex: "sku",
      key: "sku",
      width: 150,
    },
    {
      title: "Stock",
      dataIndex: "stock",
      key: "stock",
      width: 100,
      render: (_: unknown, record: PerPlatformRow) => (
        <InputNumber
          size="small"
          min={0}
          value={perPlatformConfig[record.sku]?.stock ?? 0}
          onChange={(value) =>
            setPerPlatformConfig((prev) => {
              const existing = prev[record.sku];
              if (!existing) {
                return prev;
              }

              return {
                ...prev,
                [record.sku]: {
                  ...existing,
                  stock: value ?? 0,
                },
              };
            })
          }
          style={{ width: 80 }}
        />
      ),
    },
    ...PLATFORM_OPTIONS.map((platformOption) => ({
      title: platformOption.label,
      key: platformOption.key,
      width: 100,
      render: (_: unknown, record: PerPlatformRow) => (
        <Checkbox
          disabled={!linkedPlatformsBySku[record.sku]?.[platformOption.key]}
          checked={
            linkedPlatformsBySku[record.sku]?.[platformOption.key]
              ? (perPlatformConfig[record.sku]?.platforms[platformOption.key] ??
                false)
              : false
          }
          onChange={(event) =>
            setPerPlatformConfig((prev) => {
              const existing = prev[record.sku];
              if (!existing) {
                return prev;
              }

              if (!linkedPlatformsBySku[record.sku]?.[platformOption.key]) {
                return {
                  ...prev,
                  [record.sku]: {
                    ...existing,
                    platforms: {
                      ...existing.platforms,
                      [platformOption.key]: false,
                    },
                  },
                };
              }

              return {
                ...prev,
                [record.sku]: {
                  ...existing,
                  platforms: {
                    ...existing.platforms,
                    [platformOption.key]: event.target.checked,
                  },
                },
              };
            })
          }
        />
      ),
    })),
  ];
};
