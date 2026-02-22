import { Col, Row, Space, Tag, Typography } from "antd";
import { formatIdr } from "@/pages/products/utils/productColumns";
import type { Platform, UnifiedProductRow } from "@/types/shared";
import shopeeIcon from "@/assets/icons/shopee.svg";
import tiktokIcon from "@/assets/icons/tiktok.webp";
import lazadaIcon from "@/assets/icons/lazada.webp";

const PLATFORM_COLORS: Record<Platform, string> = {
  shopee: "#ee4d2d",
  tiktok: "#000000",
  lazada: "#0f1689",
};

const PLATFORM_ICONS: Record<Platform, string> = {
  shopee: shopeeIcon,
  tiktok: tiktokIcon,
  lazada: lazadaIcon,
};

interface ProductVariantExpandedRowProps {
  product: UnifiedProductRow;
}

function PlatformPriceTag({ platform, price }: { platform: Platform; price: number }) {
  return (
    <Tag
      style={{
        fontSize: 10,
        margin: 0,
        lineHeight: "18px",
        display: "inline-flex",
        alignItems: "center",
        gap: 3,
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
    </Tag>
  );
}

export function ProductVariantExpandedRow({
  product,
}: ProductVariantExpandedRowProps) {
  if (product.skus.length <= 1) {
    return null;
  }

  return (
    <div
      style={{
        padding: "16px",
        backgroundColor: "#f8fafc",
        borderRadius: "3px",
        border: "1px solid #e2e8f0",
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
        <Typography.Text strong style={{ color: "#0f172a", fontSize: 13 }}>
          {product.skus.length} Variations
        </Typography.Text>
        <Typography.Text type="secondary" style={{ fontSize: 12 }}>
          Variant details. Use modal actions to update stock and price.
        </Typography.Text>
      </div>

      <div
        style={{
          border: "1px solid #e2e8f0",
          borderRadius: "3px",
          backgroundColor: "#ffffff",
        }}
      >
        {/* Header */}
        <Row
          style={{
            padding: "8px 12px",
            backgroundColor: "#f1f5f9",
            borderBottom: "1px solid #e2e8f0",
            fontWeight: 500,
            fontSize: 12,
            borderTopLeftRadius: "3px",
            borderTopRightRadius: "3px",
          }}
        >
          <Col xs={24} md={6} style={{ color: "#475569" }}>
            Variant Details
          </Col>
          <Col xs={4} md={3} style={{ color: "#475569" }}>
            Stock
          </Col>
          <Col xs={4} md={3} style={{ color: "#475569" }}>
            Rec. Price
          </Col>
          <Col xs={12} md={8} style={{ color: "#475569" }}>
            Platform Prices
          </Col>
          <Col xs={12} md={4} style={{ color: "#475569" }}>
            Linked
          </Col>
        </Row>

        {/* Rows */}
        {product.skus.map((sku, index) => {
          const linkedPlatforms = (sku.platform_links || [])
            .filter((l) => l.sync_status === "synced" || l.sync_status === "outdated")
            .map((l) => l.platform);
          const uniquePlatforms = [...new Set(linkedPlatforms)];

          return (
            <Row
              key={sku.id}
              align="middle"
              style={{
                padding: "12px",
                borderBottom:
                  index < product.skus.length - 1 ? "1px solid #f1f5f9" : "none",
                fontSize: 12,
              }}
            >
              <Col xs={24} md={6}>
                <Space direction="vertical" size={2}>
                  <Typography.Text
                    strong
                    style={{ color: "#0369a1", fontSize: 12 }}
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
                  type={sku.inventory_price && sku.inventory_price > 0 ? undefined : "secondary"}
                >
                  {sku.inventory_price && sku.inventory_price > 0
                    ? formatIdr(sku.inventory_price)
                    : "—"}
                </Typography.Text>
              </Col>

              <Col xs={12} md={8}>
                <Space size={4} wrap>
                  {(["shopee", "tiktok", "lazada"] as Platform[]).map((p) => {
                    const pp = sku.platform_prices?.find((x) => x.platform === p);
                    return (
                      <PlatformPriceTag
                        key={p}
                        platform={p}
                        price={pp?.platform_price ?? 0}
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
                        color={PLATFORM_COLORS[p]}
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
