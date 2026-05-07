import { Typography, theme } from "antd";
import { PlatformIndicator } from "@/components/shared/PlatformIndicator";
import { usePlatformStatus } from "@/hooks/usePlatformStatus";
import type { Platform, UnifiedProductRow } from "@/types/shared";
import { formatIdr } from "@/pages/products/utils/productColumns";

const PLATFORMS: Platform[] = ["shopee", "tiktok", "lazada"];

interface PlatformStatusCellProps {
  product: UnifiedProductRow;
}

/**
 * Renders per-platform breakdown: icon (linked/not) | stock | price
 * Each platform on its own row, vertically stacked.
 */
export function PlatformStatusCell({ product }: PlatformStatusCellProps) {
  const statusByPlatform = usePlatformStatus(product);
  const { token } = theme.useToken();

  return (
    <div style={{ display: "flex", flexDirection: "column", gap: 4 }}>
      {PLATFORMS.map((platform) => {
        const data = statusByPlatform[platform];
        const platformPrice = getPlatformPriceForProduct(product, platform);
        const platformStock = getPlatformStockForProduct(product, platform);

        return (
          <div
            key={`${product.id}-${platform}`}
            style={{
              display: "flex",
              alignItems: "center",
              gap: 8,
              minHeight: 28,
            }}
          >
            <PlatformIndicator data={data} size="small" />
            {data.linked ? (
              <>
                <Typography.Text
                  style={{ fontSize: 12, minWidth: 24, textAlign: "right" }}
                >
                  {platformStock > 0 ? platformStock : "0"}
                </Typography.Text>
                <Typography.Text
                  style={{
                    fontSize: 12,
                    color: platformPrice > 0 ? token.colorText : token.colorTextTertiary,
                  }}
                >
                  {platformPrice > 0 ? formatIdr(platformPrice) : "—"}
                </Typography.Text>
              </>
            ) : (
              <Typography.Text
                type="secondary"
                style={{ fontSize: 11, fontStyle: "italic" }}
              >
                —
              </Typography.Text>
            )}
          </div>
        );
      })}
    </div>
  );
}

/** Get platform-specific price from product's SKU platform_prices array */
function getPlatformPriceForProduct(
  product: UnifiedProductRow,
  platform: Platform,
): number {
  for (const sku of product.skus) {
    const pp = sku.platform_prices?.find((p) => p.platform === platform);
    if (pp && pp.platform_price > 0) return pp.platform_price;
  }
  return 0;
}

/** Get platform-specific stock from product's SKU platform_prices array */
function getPlatformStockForProduct(
  product: UnifiedProductRow,
  platform: Platform,
): number {
  for (const sku of product.skus) {
    const pp = sku.platform_prices?.find((p) => p.platform === platform);
    if (pp && pp.platform_stock > 0) return pp.platform_stock;
  }
  return 0;
}
