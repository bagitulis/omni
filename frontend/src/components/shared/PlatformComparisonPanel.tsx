import { Space, Tag, Typography, theme } from "antd";
import type { Platform, UnifiedProductRow } from "@/types/shared";
import shopeeIcon from "@/assets/icons/shopee.svg";
import tiktokIcon from "@/assets/icons/tiktok.webp";
import lazadaIcon from "@/assets/icons/lazada.webp";

const PLATFORM_META: Record<Platform, { icon: string; color: string; label: string }> = {
  shopee: { icon: shopeeIcon, color: "#ee4d2d", label: "Shopee" },
  tiktok: { icon: tiktokIcon, color: "#69727d", label: "TikTok" },
  lazada: { icon: lazadaIcon, color: "#0f1689", label: "Lazada" },
};

const idrFormatter = new Intl.NumberFormat("id-ID", { maximumFractionDigits: 0 });

interface PlatformComparisonPanelProps {
  products: UnifiedProductRow[];
  mode: "price" | "stock";
}

/** Shows current marketplace price+stock for each SKU before editing. */
export function PlatformComparisonPanel({ products, mode }: PlatformComparisonPanelProps) {
  const { token } = theme.useToken();
  if (products.length === 0) return null;

  // Aggregate all SKU platform data
  const skuEntries = products.flatMap((p) =>
    p.skus.map((sku) => ({
      sellerSku: sku.seller_sku,
      productTitle: p.title,
      variantName: sku.variant_name,
      inventoryPrice: sku.inventory_price ?? 0,
      inventoryStock: sku.inventory_stock ?? 0,
      masterPrice: sku.price,
      masterStock: sku.stock,
      platformPrices: sku.platform_prices ?? [],
    })),
  );

  // Show first 5 SKUs max to keep panel compact
  const visibleSkus = skuEntries.slice(0, 5);
  const hiddenCount = skuEntries.length - visibleSkus.length;

  return (
    <div
      style={{
        padding: "10px 12px",
        backgroundColor: token.colorBgLayout,
        borderRadius: 3,
        border: `1px solid ${token.colorBorderSecondary}`,
        marginBottom: 16,
        fontSize: 12,
      }}
    >
      <Typography.Text strong style={{ fontSize: 12, color: token.colorTextSecondary, display: "block", marginBottom: 8 }}>
        Current Marketplace {mode === "price" ? "Prices" : "Stock"}
      </Typography.Text>

      {visibleSkus.map((sku) => {
        const invValue = mode === "price" ? sku.inventoryPrice : sku.inventoryStock;
        const masterValue = mode === "price" ? sku.masterPrice : sku.masterStock;
        const displayRef = invValue > 0 ? invValue : masterValue;

        return (
          <div
            key={sku.sellerSku}
            style={{
              display: "flex",
              alignItems: "center",
              gap: 8,
              marginBottom: 6,
              flexWrap: "wrap",
            }}
          >
            <Typography.Text
              type="secondary"
              style={{ fontSize: 11, minWidth: 100, maxWidth: 180 }}
              ellipsis={{ tooltip: `${sku.productTitle} — ${sku.variantName || sku.sellerSku}` }}
            >
              {sku.variantName
                ? `${sku.productTitle.length > 20 ? sku.productTitle.slice(0, 20) + "…" : sku.productTitle} — ${sku.variantName}`
                : sku.productTitle}
            </Typography.Text>

            <Tag style={{ margin: 0, fontSize: 10, lineHeight: "18px" }}>
              {mode === "price"
                ? (displayRef > 0 ? `Rp ${idrFormatter.format(displayRef)}` : "—")
                : `Stock: ${displayRef}`}
            </Tag>

            <Space size={4}>
              {(["shopee", "tiktok", "lazada"] as Platform[]).map((p) => {
                const pp = sku.platformPrices.find((x) => x.platform === p);
                const val = mode === "price" ? (pp?.platform_price ?? 0) : (pp?.platform_stock ?? 0);
                const meta = PLATFORM_META[p];

                return (
                  <Tag
                    key={p}
                    style={{
                      fontSize: 10,
                      margin: 0,
                      lineHeight: "18px",
                      display: "inline-flex",
                      alignItems: "center",
                      gap: 3,
                      opacity: val > 0 ? 1 : 0.4,
                    }}
                  >
                    <img src={meta.icon} alt={meta.label} width={11} height={11} style={{ objectFit: "contain" }} />
                    {mode === "price"
                      ? (val > 0 ? `Rp ${idrFormatter.format(val)}` : "—")
                      : (val > 0 ? val.toLocaleString("id-ID") : "—")}
                  </Tag>
                );
              })}
            </Space>
          </div>
        );
      })}

      {hiddenCount > 0 && (
        <Typography.Text type="secondary" style={{ fontSize: 10 }}>
          +{hiddenCount} more SKUs
        </Typography.Text>
      )}
    </div>
  );
}
