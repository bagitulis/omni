import { Checkbox, InputNumber, Tooltip, theme } from "antd";
import type { ColumnType } from "antd/es/table";
import type { Dispatch, SetStateAction } from "react";
import type { Platform, UnifiedProductRow } from "@/types/shared";
import type { PriceRecommendationMap } from "./priceSyncRecommendations";

export interface PricePerPlatformConfig {
  [sku: string]: {
    prices: Record<Platform, number>;
    platforms: Record<Platform, boolean>;
  };
}

export interface PricePerPlatformRow {
  key: string;
  sku: string;
}

export const PRICE_PLATFORM_OPTIONS: Array<{ key: Platform; label: string }> = [
  { key: "shopee", label: "Shopee" },
  { key: "tiktok", label: "TikTok" },
  { key: "lazada", label: "Lazada" },
];

const idrFormatter = new Intl.NumberFormat("id-ID", { maximumFractionDigits: 0 });

export const getDefaultPricePerPlatformConfig = (
  selectedProducts: UnifiedProductRow[],
  linkedPlatformsBySku: Record<string, Record<Platform, boolean>>,
): PricePerPlatformConfig => {
  const config: PricePerPlatformConfig = {};

  for (const product of selectedProducts) {
    for (const sku of product.skus) {
      const basePrice = sku.price;
      config[sku.seller_sku] = {
        prices: {
          shopee: basePrice,
          tiktok: basePrice,
          lazada: basePrice,
        },
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

function PriceHintCell({ hint }: { hint: PriceRecommendationMap[string] | undefined }) {
  const { token } = theme.useToken();

  if (!hint) {
    return (
      <span style={{ color: token.colorTextQuaternary, fontSize: 11 }}>
        —
      </span>
    );
  }

  const parts: string[] = [];
  if (hint.shopee > 0) parts.push(`S:${idrFormatter.format(hint.shopee)}`);
  if (hint.tiktok > 0) parts.push(`T:${idrFormatter.format(hint.tiktok)}`);
  if (hint.lazada > 0) parts.push(`L:${idrFormatter.format(hint.lazada)}`);

  return (
    <Tooltip title={`Base: Rp ${idrFormatter.format(hint.base_price)} | Source: ${hint.source}`}>
      <span style={{ fontSize: 11, color: token.colorTextSecondary }}>
        {parts.length > 0 ? parts.join(" ") : `Rp ${idrFormatter.format(hint.base_price)}`}
        {hint.source === "fallback" && (
          <span style={{ color: token.colorWarning, marginLeft: 4 }}>(fb)</span>
        )}
      </span>
    </Tooltip>
  );
}

export const getPricePerPlatformColumns = (
  linkedPlatformsBySku: Record<string, Record<Platform, boolean>>,
  perPlatformConfig: PricePerPlatformConfig,
  setPerPlatformConfig: Dispatch<SetStateAction<PricePerPlatformConfig>>,
  recommendations: PriceRecommendationMap,
  productNameBySku?: Record<string, string>,
): ColumnType<PricePerPlatformRow>[] => {
  return [
    {
      title: "SKU",
      dataIndex: "sku",
      key: "sku",
      width: 150,
      fixed: "left" as const,
      render: (_: unknown, record: PricePerPlatformRow) => {
        const name = productNameBySku?.[record.sku];
        return (
          <div>
            <div style={{ fontWeight: 500, fontSize: 12 }}>{record.sku}</div>
            {name && (
              <div style={{ fontSize: 11, color: "var(--color-text-secondary)", lineHeight: 1.2, marginTop: 2, overflow: "hidden", textOverflow: "ellipsis", whiteSpace: "nowrap", maxWidth: 130 }}>
                {name}
              </div>
            )}
          </div>
        );
      },
    },
    {
      title: "Hint",
      key: "inventory_price_hint",
      width: 150,
      render: (_: unknown, record: PricePerPlatformRow) => (
        <PriceHintCell hint={recommendations[record.sku]} />
      ),
    },
    ...PRICE_PLATFORM_OPTIONS.map((platformOption) => ({
      title: (
        <span style={{ fontSize: 12 }}>
          {platformOption.label}
        </span>
      ),
      key: `price_${platformOption.key}`,
      width: 150,
      render: (_: unknown, record: PricePerPlatformRow) => {
        const linked =
          linkedPlatformsBySku[record.sku]?.[platformOption.key] ?? false;
        const config = perPlatformConfig[record.sku];
        const isChecked = linked
          ? (config?.platforms[platformOption.key] ?? false)
          : false;

        return (
          <div style={{ display: "flex", alignItems: "center", gap: 4 }}>
            <Checkbox
              disabled={!linked}
              checked={isChecked}
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
            <InputNumber
              size="small"
              min={0}
              disabled={!linked || !isChecked}
              value={config?.prices[platformOption.key] ?? 0}
              onChange={(value) =>
                setPerPlatformConfig((previous) => {
                  const existing = previous[record.sku];
                  if (!existing) return previous;

                  return {
                    ...previous,
                    [record.sku]: {
                      ...existing,
                      prices: {
                        ...existing.prices,
                        [platformOption.key]: value ?? 0,
                      },
                    },
                  };
                })
              }
              style={{ width: 95 }}
              formatter={(value) => value != null ? `${Number(value).toLocaleString("id-ID")}` : ""}
              parser={(value) => Number((value ?? "").replace(/\./g, "")) || 0}
            />
          </div>
        );
      },
    })),
  ];
};
