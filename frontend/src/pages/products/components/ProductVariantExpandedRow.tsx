import { Col, Divider, Row, Space, Typography } from "antd";
import { InlineEditCell } from "@/components/shared/InlineEditCell";
import { formatIdr } from "@/pages/products/utils/productColumns";
import type { UnifiedProductRow } from "@/types/shared";

interface ProductVariantExpandedRowProps {
  product: UnifiedProductRow;
  onInlinePriceSave: (skuId: number, price: number) => Promise<void>;
  onInlineStockSave: (
    skuId: number,
    sellerSku: string,
    stock: number,
  ) => Promise<void>;
}

export function ProductVariantExpandedRow({
  product,
  onInlinePriceSave,
  onInlineStockSave,
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
            Variant details and inline stock/price updates
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
              <div className="inline-edit-cell" data-field="stock">
                <InlineEditCell
                  value={sku.stock}
                  mode="stock"
                  onSave={(value) =>
                    onInlineStockSave(sku.id, sku.seller_sku, value)
                  }
                />
              </div>
            </Col>

            <Col xs={12} md={7}>
              <div className="inline-edit-cell" data-field="price">
                <InlineEditCell
                  value={sku.price}
                  mode="price"
                  prefix="Rp"
                  onSave={(value) => onInlinePriceSave(sku.id, value)}
                />
              </div>
              <Typography.Text type="secondary" style={{ fontSize: 11 }}>
                {formatIdr(sku.price)}
              </Typography.Text>
            </Col>
          </Row>
        </div>
      ))}
    </Space>
  );
}
