import { Checkbox, InputNumber, theme } from "antd";
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
            stock: 0,
          },
          tiktok: {
            enabled: linkedPlatformsBySku[sku.seller_sku]?.tiktok ?? false,
            stock: 0,
          },
          lazada: {
            enabled: linkedPlatformsBySku[sku.seller_sku]?.lazada ?? false,
            stock: 0,
          },
        },
      };
    }
  }

  return config;
};

// ─── Stock Platform Cell ─────────────────────────────────────────────────────

// eslint-disable-next-line react-refresh/only-export-components
const StockPlatformCell = ({
  linked,
  enabled,
  stock,
  currentStock,
  onEnabledChange,
  onStockChange,
}: {
  linked: boolean;
  enabled: boolean;
  stock: number;
  currentStock: number;
  onEnabledChange: (checked: boolean) => void;
  onStockChange: (value: number) => void;
}) => {
  const { token } = theme.useToken();

  return (
    <div>
      <div style={{ display: "flex", gap: 8, alignItems: "center" }}>
        <Checkbox
          disabled={!linked}
          checked={linked ? enabled : false}
          onChange={(e) => onEnabledChange(e.target.checked)}
        />
        <InputNumber
          size="small"
          min={0}
          value={stock}
          disabled={!linked || !enabled}
          onChange={(value) => onStockChange(value ?? 0)}
          style={{ width: 92 }}
        />
      </div>
      {currentStock > 0 && (
        <div style={{ fontSize: 10, color: token.colorTextQuaternary, marginLeft: 32, marginTop: 2 }}>
          now: {currentStock}
        </div>
      )}
    </div>
  );
};

// ─── Column Builder ──────────────────────────────────────────────────────────

export const getPerPlatformColumns = (
  linkedPlatformsBySku: Record<string, Record<Platform, boolean>>,
  perPlatformConfig: PerPlatformConfig,
  setPerPlatformConfig: Dispatch<SetStateAction<PerPlatformConfig>>,
  _recommendations: StockRecommendationMap,
  productNameBySku?: Record<string, string>,
  currentStockBySku?: Record<string, Record<Platform, number>>,
): ColumnType<PerPlatformRow>[] => {
  return [
    {
      title: "SKU",
      dataIndex: "sku",
      key: "sku",
      width: 180,
      render: (_: unknown, record: PerPlatformRow) => {
        const name = productNameBySku?.[record.sku];
        return (
          <div>
            <div style={{ fontWeight: 500, fontSize: 12 }}>{record.sku}</div>
            {name && (
              <div style={{ fontSize: 11, color: "var(--color-text-secondary)", lineHeight: 1.3, marginTop: 2, maxWidth: 160, display: "-webkit-box", WebkitLineClamp: 2, WebkitBoxOrient: "vertical", overflow: "hidden" }}>
                {name}
              </div>
            )}
          </div>
        );
      },
    },
    ...PLATFORM_OPTIONS.map(
      (platformOption): ColumnType<PerPlatformRow> => {
        // Compute check-all state for this platform column
        const allSkus = Object.keys(perPlatformConfig);
        const linkedSkus = allSkus.filter(
          (sku) => linkedPlatformsBySku[sku]?.[platformOption.key],
        );
        const checkedSkus = linkedSkus.filter(
          (sku) =>
            perPlatformConfig[sku]?.platforms[platformOption.key]?.enabled,
        );
        const allChecked =
          linkedSkus.length > 0 && checkedSkus.length === linkedSkus.length;
        const someChecked =
          checkedSkus.length > 0 && checkedSkus.length < linkedSkus.length;

        return {
          title: (
            <div style={{ display: "flex", gap: 8, alignItems: "center" }}>
              <Checkbox
                checked={allChecked}
                indeterminate={someChecked}
                disabled={linkedSkus.length === 0}
                onChange={(event) => {
                  const nextEnabled = event.target.checked;
                  setPerPlatformConfig((previous) => {
                    const next = { ...previous };
                    for (const sku of linkedSkus) {
                      const existing = next[sku];
                      if (!existing) continue;
                      next[sku] = {
                        ...existing,
                        platforms: {
                          ...existing.platforms,
                          [platformOption.key]: {
                            ...existing.platforms[platformOption.key],
                            enabled: nextEnabled,
                          },
                        },
                      };
                    }
                    return next;
                  });
                }}
              />
              <span>{platformOption.label}</span>
            </div>
          ),
          key: platformOption.key,
          width: 185,
          render: (_: unknown, record: PerPlatformRow) => {
            const linked = linkedPlatformsBySku[record.sku]?.[platformOption.key] ?? false;
            const platformConfig = perPlatformConfig[record.sku]?.platforms[platformOption.key];
            const currentStock = currentStockBySku?.[record.sku]?.[platformOption.key] ?? 0;

            return (
              <StockPlatformCell
                linked={linked}
                enabled={platformConfig?.enabled ?? false}
                stock={platformConfig?.stock ?? 0}
                currentStock={currentStock}
                onEnabledChange={(checked) =>
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
                            enabled: linked ? checked : false,
                            stock: current.stock,
                          },
                        },
                      },
                    };
                  })
                }
                onStockChange={(value) =>
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
                            stock: value,
                          },
                        },
                      },
                    };
                  })
                }
              />
            );
          },
        };
      },
    ),
  ];
};
