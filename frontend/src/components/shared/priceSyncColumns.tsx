import { Checkbox, InputNumber } from "antd";
import type { ColumnType } from "antd/es/table";
import type { Dispatch, SetStateAction } from "react";
import type { Platform, UnifiedProductRow } from "@/types/shared";
import type { PriceRecommendationMap } from "./priceSyncRecommendations";

export interface PricePerPlatformConfig {
  [sku: string]: {
    price: number;
    platforms: Record<Platform, boolean>;
  };
}

export interface PricePerPlatformRow {
  key: string;
  sku: string;
  price: number;
}

export const PRICE_PLATFORM_OPTIONS: Array<{ key: Platform; label: string }> = [
  { key: "shopee", label: "Shopee" },
  { key: "tiktok", label: "TikTok" },
  { key: "lazada", label: "Lazada" },
];

export const getDefaultPricePerPlatformConfig = (
  selectedProducts: UnifiedProductRow[],
  linkedPlatformsBySku: Record<string, Record<Platform, boolean>>,
): PricePerPlatformConfig => {
  const config: PricePerPlatformConfig = {};

  for (const product of selectedProducts) {
    for (const sku of product.skus) {
      config[sku.seller_sku] = {
        price: sku.price,
        platforms: {
          shopee: linkedPlatformsBySku[sku.seller_sku]?.shopee ?? false,
          tiktok: linkedPlatformsBySku[sku.seller_sku]?.tiktok ?? false,
          lazada: linkedPlatformsBySku[sku.seller_sku]?.lazada ?? false,
        },
      };
    }
  }

  return config;
};

export const getPricePerPlatformColumns = (
  linkedPlatformsBySku: Record<string, Record<Platform, boolean>>,
  perPlatformConfig: PricePerPlatformConfig,
  setPerPlatformConfig: Dispatch<SetStateAction<PricePerPlatformConfig>>,
  recommendations: PriceRecommendationMap,
): ColumnType<PricePerPlatformRow>[] => {
  return [
    {
      title: "SKU",
      dataIndex: "sku",
      key: "sku",
      width: 180,
    },
    {
      title: "Inventory Price Hint",
      key: "inventory_price_hint",
      width: 220,
      render: (_: unknown, record: PricePerPlatformRow) => {
        const hint = recommendations[record.sku];
        if (!hint) {
          return (
            <span style={{ color: "var(--color-text-secondary)", fontSize: 12 }}>
              No hint
            </span>
          );
        }

        return (
          <span style={{ fontSize: 12 }}>
            S:{hint.shopee} T:{hint.tiktok} L:{hint.lazada} ({hint.source})
          </span>
        );
      },
    },
    {
      title: "Price",
      dataIndex: "price",
      key: "price",
      width: 140,
      render: (_: unknown, record: PricePerPlatformRow) => (
        <InputNumber
          size="small"
          min={0}
          value={perPlatformConfig[record.sku]?.price ?? 0}
          onChange={(value) =>
            setPerPlatformConfig((previous) => {
              const existing = previous[record.sku];
              if (!existing) return previous;

              return {
                ...previous,
                [record.sku]: {
                  ...existing,
                  price: value ?? 0,
                },
              };
            })
          }
          style={{ width: 120 }}
        />
      ),
    },
    ...PRICE_PLATFORM_OPTIONS.map((platformOption) => ({
      title: platformOption.label,
      key: platformOption.key,
      width: 120,
      render: (_: unknown, record: PricePerPlatformRow) => {
        const linked =
          linkedPlatformsBySku[record.sku]?.[platformOption.key] ?? false;

        return (
          <Checkbox
            disabled={!linked}
            checked={
              linked
                ? (perPlatformConfig[record.sku]?.platforms[
                    platformOption.key
                  ] ?? false)
                : false
            }
            onChange={(event) =>
              setPerPlatformConfig((previous) => {
                const existing = previous[record.sku];
                if (!existing) return previous;

                return {
                  ...previous,
                  [record.sku]: {
                    ...existing,
                    platforms: {
                      ...existing.platforms,
                      [platformOption.key]: linked
                        ? event.target.checked
                        : false,
                    },
                  },
                };
              })
            }
          />
        );
      },
    })),
  ];
};
