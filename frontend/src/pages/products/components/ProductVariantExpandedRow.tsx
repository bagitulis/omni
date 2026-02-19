import { Col, Divider, Row, Space, Typography } from "antd";
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
    <Space direction="vertical" size={10} style={{ width: "100%" }}>
      <Row justify="space-between" align="middle" wrap>
        <Col>
          <Typography.Text strong>{product.skus.length} SKU</Typography.Text>
        </Col>
        <Col>
          <Typography.Text type="secondary" style={{ fontSize: 12 }}>
            Variant details. Use modal actions to update stock and price.
          </Typography.Text>
        </Col>
      </Row>

      <Row gutter={12} style={{ fontSize: 12 }}>
        <Col xs={24} md={10}>
          <Typography.Text type="secondary">Variant</Typography.Text>
        </Col>
        <Col xs={12} md={7}>
          <Typography.Text type="secondary">Stock</Typography.Text>
        </Col>
        <Col xs={12} md={7}>
          <Typography.Text type="secondary">Price</Typography.Text>
        </Col>
      </Row>

      {product.skus.map((sku, index) => (
        <div key={sku.id}>
          {index > 0 ? <Divider style={{ margin: "8px 0" }} /> : null}
          <Row gutter={12} align="middle">
            <Col xs={24} md={10}>
              <Space direction="vertical" size={0}>
                <Typography.Text strong>
                  {sku.variant_name.trim() || "Default Variant"}
                </Typography.Text>
                <Typography.Text type="secondary" style={{ fontSize: 12 }}>
                  {sku.seller_sku}
                </Typography.Text>
              </Space>
            </Col>

            <Col xs={12} md={7}>
              <Typography.Text>
                {sku.stock.toLocaleString("id-ID")}
              </Typography.Text>
            </Col>

            <Col xs={12} md={7}>
              <Typography.Text>{formatIdr(sku.price)}</Typography.Text>
            </Col>
          </Row>
        </div>
      ))}
    </Space>
  );
}
