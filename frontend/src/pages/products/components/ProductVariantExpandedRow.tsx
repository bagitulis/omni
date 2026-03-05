import { Col, Row, Space, Tag, Typography, theme } from "antd";
import { formatIdr } from "@/pages/products/utils/productColumns";
import type { Platform, UnifiedProductRow } from "@/types/shared";
import { useTheme } from "@/contexts/ThemeContext.hooks";
import shopeeIcon from "@/assets/icons/shopee.svg";
import tiktokIcon from "@/assets/icons/tiktok.webp";
import lazadaIcon from "@/assets/icons/lazada.webp";

const PLATFORM_COLORS: Record<Platform, { light: string; dark: string }> = {
  shopee: { light: "#ee4d2d", dark: "#ee4d2d" },
  tiktok: { light: "#000000", dark: "#e2e8f0" },
  lazada: { light: "#0f1689", dark: "#818cf8" },
};

const PLATFORM_ICONS: Record<Platform, string> = {
  shopee: shopeeIcon,
  tiktok: tiktokIcon,
  lazada: lazadaIcon,
};

interface ProductVariantExpandedRowProps {
  product: UnifiedProductRow;
}

/** Shows platform price + stock in a compact tag. */
function PlatformDetailTag({
  platform,
  price,
  stock,
}: {
  platform: Platform;
  price: number;
  stock: number;
}) {
  const { token } = theme.useToken();
  const hasData = price > 0 || stock > 0;
  return (
    <Tag
      style={{
        fontSize: 10,
        margin: 0,
        lineHeight: "18px",
        display: "inline-flex",
        alignItems: "center",
        gap: 3,
        opacity: hasData ? 1 : 0.5,
      }}
    >
      <img
        src={PLATFORM_ICONS[platform]}
        alt={platform}
        width={12}
        height={12}
        style={{ objectFit: "contain" }}
      />
      {price > 0 ? formatIdr(price) : "—"}
      <span style={{ color: token.colorTextTertiary, margin: "0 1px" }}>·</span>
      <span style={{ color: token.colorTextSecondary }}>{stock > 0 ? stock.toLocaleString("id-ID") : "—"}</span>
    </Tag>
  );
}

export function ProductVariantExpandedRow({
  product,
}: ProductVariantExpandedRowProps) {
  const { token } = theme.useToken();
  const { isDark } = useTheme();

  if (product.skus.length <= 1) {
    return null;
  }

  return (
    <div
      style={{
        padding: "16px",
        backgroundColor: token.colorBgLayout,
        borderRadius: "3px",
        border: `1px solid ${token.colorBorderSecondary}`,
      }}
    >
      <div
        style={{
          display: "flex",
          justifyContent: "space-between",
          alignItems: "center",
          marginBottom: "12px",
        }}
      >
        <Typography.Text strong style={{ fontSize: 13 }}>
          {product.skus.length} Variations
        </Typography.Text>
        <Typography.Text type="secondary" style={{ fontSize: 11 }}>
          Platform columns: price · stock
        </Typography.Text>
      </div>

      <div
        style={{
          border: `1px solid ${token.colorBorderSecondary}`,
          borderRadius: "3px",
          backgroundColor: token.colorBgContainer,
        }}
      >
        {/* Header */}
        <Row
          style={{
            padding: "8px 12px",
            backgroundColor: token.colorBgSpotlight,
            borderBottom: `1px solid ${token.colorBorderSecondary}`,
            fontWeight: 500,
            fontSize: 12,
            borderTopLeftRadius: "3px",
            borderTopRightRadius: "3px",
            color: token.colorTextSecondary,
          }}
        >
          <Col xs={24} md={6}>
            Variant Details
          </Col>
          <Col xs={4} md={3}>
            Stock
          </Col>
          <Col xs={4} md={3}>
            Price
          </Col>
          <Col xs={12} md={8}>
            Platform (Price · Stock)
          </Col>
          <Col xs={12} md={4}>
            Linked
          </Col>
        </Row>

        {/* Rows */}
        {product.skus.map((sku, index) => {
          const linkedPlatforms = (sku.platform_links || [])
            .filter((l) => l.sync_status === "synced" || l.sync_status === "outdated")
            .map((l) => l.platform);
          const uniquePlatforms = [...new Set(linkedPlatforms)];

          // Determine display price: inventory first, then master
          const displayPrice =
            sku.inventory_price && sku.inventory_price > 0
              ? sku.inventory_price
              : sku.price;

          return (
            <Row
              key={sku.id}
              align="middle"
              style={{
                padding: "12px",
                borderBottom:
                  index < product.skus.length - 1 ? `1px solid ${token.colorBorderSecondary}` : "none",
                fontSize: 12,
              }}
            >
              <Col xs={24} md={6}>
                <Space direction="vertical" size={2}>
                  <Typography.Text
                    strong
                    style={{ color: token.colorPrimary, fontSize: 12 }}
                  >
                    {sku.variant_name.trim() || "Default Variant"}
                  </Typography.Text>
                  <Typography.Text type="secondary" style={{ fontSize: 10 }}>
                    SKU: {sku.seller_sku}
                  </Typography.Text>
                </Space>
              </Col>

              <Col xs={4} md={3}>
                <Typography.Text style={{ fontSize: 12 }}>
                  {sku.stock.toLocaleString("id-ID")}
                </Typography.Text>
              </Col>

              <Col xs={4} md={3}>
                <Typography.Text
                  style={{ fontSize: 12 }}
                  type={displayPrice > 0 ? undefined : "secondary"}
                >
                  {displayPrice > 0 ? formatIdr(displayPrice) : "—"}
                </Typography.Text>
              </Col>

              <Col xs={12} md={8}>
                <Space size={4} wrap>
                  {(["shopee", "tiktok", "lazada"] as Platform[]).map((p) => {
                    const pp = sku.platform_prices?.find((x) => x.platform === p);
                    return (
                      <PlatformDetailTag
                        key={p}
                        platform={p}
                        price={pp?.platform_price ?? 0}
                        stock={pp?.platform_stock ?? 0}
                      />
                    );
                  })}
                </Space>
              </Col>

              <Col xs={12} md={4}>
                {uniquePlatforms.length > 0 ? (
                  <Space size={4} wrap>
                    {uniquePlatforms.map((p) => (
                      <Tag
                        key={p}
                        color={isDark ? PLATFORM_COLORS[p].dark : PLATFORM_COLORS[p].light}
                        style={{ fontSize: 10, margin: 0, lineHeight: "18px" }}
                      >
                        {p.charAt(0).toUpperCase() + p.slice(1)}
                      </Tag>
                    ))}
                  </Space>
                ) : (
                  <Typography.Text type="secondary" style={{ fontSize: 10 }}>
                    Not linked
                  </Typography.Text>
                )}
              </Col>
            </Row>
          );
        })}
      </div>
    </div>
  );
}
