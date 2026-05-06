import { Checkbox, InputNumber, theme } from "antd";
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

// ─── Platform Price Cell ─────────────────────────────────────────────────────

// eslint-disable-next-line react-refresh/only-export-components
  const PricePlatformCell = ({
  linked,
  isChecked,
  price,
  currentPrice,
  onCheckChange,
  onPriceChange,
  }: {
  linked: boolean;
  isChecked: boolean;
  price: number;
  currentPrice: number;
  onCheckChange: (checked: boolean) => void;
  onPriceChange: (value: number) => void;
}) => {
  const { token } = theme.useToken();

  return (
    <div>
      <div style={{ display: "flex", alignItems: "center", gap: 4 }}>
        <Checkbox
          disabled={!linked}
          checked={isChecked}
          onChange={(e) => onCheckChange(e.target.checked)}
        />
        <InputNumber
          size="small"
          min={0}
          disabled={!linked || !isChecked}
          value={price}
          onChange={(value) => onPriceChange(value ?? 0)}
          style={{ width: 100 }}
          formatter={(v) => v != null ? `${Number(v).toLocaleString("id-ID")}` : ""}
          parser={(v) => Number((v ?? "").replace(/\./g, "")) || 0}
        />
      </div>
      {currentPrice > 0 && (
        <div style={{ fontSize: 10, color: token.colorTextQuaternary, marginLeft: 24, marginTop: 2 }}>
          now: Rp {idrFormatter.format(currentPrice)}
        </div>
      )}
    </div>
  );
};

// ─── Column Builder ──────────────────────────────────────────────────────────

export const getPricePerPlatformColumns = (
  linkedPlatformsBySku: Record<string, Record<Platform, boolean>>,
  perPlatformConfig: PricePerPlatformConfig,
  setPerPlatformConfig: Dispatch<SetStateAction<PricePerPlatformConfig>>,
  _recommendations: PriceRecommendationMap,
  productNameBySku?: Record<string, string>,
  currentPricesBySku?: Record<string, Record<Platform, number>>,
): ColumnType<PricePerPlatformRow>[] => {
  return [
    {
      title: "SKU",
      dataIndex: "sku",
      key: "sku",
      width: 160,
      fixed: "left" as const,
      render: (_: unknown, record: PricePerPlatformRow) => {
        const name = productNameBySku?.[record.sku];
        return (
          <div>
            <div style={{ fontWeight: 500, fontSize: 12 }}>{record.sku}</div>
            {name && (
              <div style={{ fontSize: 11, color: "var(--color-text-secondary)", lineHeight: 1.2, marginTop: 2, overflow: "hidden", textOverflow: "ellipsis", whiteSpace: "nowrap", maxWidth: 140 }}>
                {name}
              </div>
            )}
          </div>
        );
      },
    },
    ...PRICE_PLATFORM_OPTIONS.map((platformOption): ColumnType<PricePerPlatformRow> => ({
      title: <span style={{ fontSize: 12 }}>{platformOption.label}</span>,
      key: `price_${platformOption.key}`,
      width: 175,
      render: (_: unknown, record: PricePerPlatformRow) => {
        const linked = linkedPlatformsBySku[record.sku]?.[platformOption.key] ?? false;
        const config = perPlatformConfig[record.sku];
        const isChecked = linked ? (config?.platforms[platformOption.key] ?? false) : false;
        const price = config?.prices[platformOption.key] ?? 0;
        const currentPrice = currentPricesBySku?.[record.sku]?.[platformOption.key] ?? 0;

        return (
          <PricePlatformCell
            linked={linked}
            isChecked={isChecked}
            price={price}
            currentPrice={currentPrice}
            onCheckChange={(checked) =>
              setPerPlatformConfig((prev) => {
                const existing = prev[record.sku];
                if (!existing) return prev;
                return {
                  ...prev,
                  [record.sku]: {
                    ...existing,
                    platforms: {
                      ...existing.platforms,
                      [platformOption.key]: linked ? checked : false,
                    },
                  },
                };
              })
            }
            onPriceChange={(value) =>
              setPerPlatformConfig((prev) => {
                const existing = prev[record.sku];
                if (!existing) return prev;
                return {
                  ...prev,
                  [record.sku]: {
                    ...existing,
                    prices: {
                      ...existing.prices,
                      [platformOption.key]: value,
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
