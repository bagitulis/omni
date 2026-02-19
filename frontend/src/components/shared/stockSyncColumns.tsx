import { Checkbox, InputNumber } from "antd";
import type { ColumnType } from "antd/es/table";
import type { Dispatch, SetStateAction } from "react";
import type { Platform, UnifiedProductRow } from "@/types/shared";
import type { StockRecommendationMap } from "./stockSyncRecommendations";

export interface PerPlatformConfig {
  [sku: string]: {
    platforms: Record<
      Platform,
      {
        enabled: boolean;
        stock: number;
      }
    >;
  };
}

export interface PerPlatformRow {
  key: string;
  sku: string;
}

export const PLATFORM_OPTIONS: Array<{ key: Platform; label: string }> = [
  { key: "shopee", label: "Shopee" },
  { key: "tiktok", label: "TikTok" },
  { key: "lazada", label: "Lazada" },
];

export const getDefaultPerPlatformConfig = (
  selectedProducts: UnifiedProductRow[],
  linkedPlatformsBySku: Record<string, Record<Platform, boolean>>,
): PerPlatformConfig => {
  const config: PerPlatformConfig = {};

  for (const product of selectedProducts) {
    for (const sku of product.skus) {
      config[sku.seller_sku] = {
        platforms: {
          shopee: {
            enabled: linkedPlatformsBySku[sku.seller_sku]?.shopee ?? false,
            stock: sku.stock,
          },
          tiktok: {
            enabled: linkedPlatformsBySku[sku.seller_sku]?.tiktok ?? false,
            stock: sku.stock,
          },
          lazada: {
            enabled: linkedPlatformsBySku[sku.seller_sku]?.lazada ?? false,
            stock: sku.stock,
          },
        },
      };
    }
  }

  return config;
};

export const getPerPlatformColumns = (
  linkedPlatformsBySku: Record<string, Record<Platform, boolean>>,
  perPlatformConfig: PerPlatformConfig,
  setPerPlatformConfig: Dispatch<SetStateAction<PerPlatformConfig>>,
  recommendations: StockRecommendationMap,
): ColumnType<PerPlatformRow>[] => {
  return [
    {
      title: "SKU",
      dataIndex: "sku",
      key: "sku",
      width: 180,
    },
    {
      title: "Inventory Hint",
      key: "inventory_hint",
      width: 220,
      render: (_: unknown, record: PerPlatformRow) => {
        const hint = recommendations[record.sku];
        if (!hint) {
          return <span style={{ color: "#999", fontSize: 12 }}>No hint</span>;
        }

        return (
          <span style={{ fontSize: 12 }}>
            S:{hint.shopee} T:{hint.tiktok} L:{hint.lazada} ({hint.source})
          </span>
        );
      },
    },
    ...PLATFORM_OPTIONS.map(
      (platformOption): ColumnType<PerPlatformRow> => ({
        title: platformOption.label,
        key: platformOption.key,
        width: 170,
        render: (_: unknown, record: PerPlatformRow) => {
          const linked =
            linkedPlatformsBySku[record.sku]?.[platformOption.key] ?? false;
          const platformConfig =
            perPlatformConfig[record.sku]?.platforms[platformOption.key];

          return (
            <div style={{ display: "flex", gap: 8, alignItems: "center" }}>
              <Checkbox
                disabled={!linked}
                checked={linked ? (platformConfig?.enabled ?? false) : false}
                onChange={(event) =>
                  setPerPlatformConfig((previous) => {
                    const existing = previous[record.sku];
                    if (!existing) return previous;

                    const current = existing.platforms[platformOption.key];

                    return {
                      ...previous,
                      [record.sku]: {
                        ...existing,
                        platforms: {
                          ...existing.platforms,
                          [platformOption.key]: {
                            enabled: linked ? event.target.checked : false,
                            stock: current.stock,
                          },
                        },
                      },
                    };
                  })
                }
              />
              <InputNumber
                size="small"
                min={0}
                value={platformConfig?.stock ?? 0}
                disabled={!linked || !(platformConfig?.enabled ?? false)}
                onChange={(value) =>
                  setPerPlatformConfig((previous) => {
                    const existing = previous[record.sku];
                    if (!existing) return previous;

                    const current = existing.platforms[platformOption.key];
                    return {
                      ...previous,
                      [record.sku]: {
                        ...existing,
                        platforms: {
                          ...existing.platforms,
                          [platformOption.key]: {
                            ...current,
                            stock: value ?? 0,
                          },
                        },
                      },
                    };
                  })
                }
                style={{ width: 92 }}
              />
            </div>
          );
        },
      }),
    ),
  ];
};
