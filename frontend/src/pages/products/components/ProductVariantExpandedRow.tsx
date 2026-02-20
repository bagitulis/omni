import { Col, Row, Space, Typography } from "antd";
import { formatIdr } from "@/pages/products/utils/productColumns";
import type { UnifiedProductRow } from "@/types/shared";

interface ProductVariantExpandedRowProps {
  product: UnifiedProductRow;
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
          <Col xs={24} md={10} style={{ color: "#475569" }}>
            Variant Details
          </Col>
          <Col xs={12} md={7} style={{ color: "#475569" }}>
            Stock
          </Col>
          <Col xs={12} md={7} style={{ color: "#475569" }}>
            Price
          </Col>
        </Row>

        {/* Rows */}
        {product.skus.map((sku, index) => (
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
            <Col xs={24} md={10}>
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

            <Col xs={12} md={7}>
              <Typography.Text style={{ fontSize: 12 }}>
                {sku.stock.toLocaleString("id-ID")}
              </Typography.Text>
            </Col>

            <Col xs={12} md={7}>
              <Typography.Text style={{ fontSize: 12 }}>
                {formatIdr(sku.price)}
              </Typography.Text>
            </Col>
          </Row>
        ))}
      </div>
    </div>
  );
}
