import {
  Button,
  Card,
  Col,
  Empty,
  Pagination,
  Row,
  Space,
  Typography,
} from "antd";
import { DeleteOutlined, EditOutlined } from "@ant-design/icons";
import { PlatformBadge } from "@/components/ui/PlatformBadge";
import type { Product } from "@/types/product";

interface ProductGridViewProps {
  products: Product[];
  page: number;
  pageSize: number;
  total: number;
  onPageChange: (page: number, pageSize: number) => void;
  onDelete: (id: string) => void;
}

export function ProductGridView({
  products,
  page,
  pageSize,
  total,
  onPageChange,
  onDelete,
}: ProductGridViewProps) {
  if (!products.length) return <Empty description="No products found" />;

  return (
    <>
      <Row gutter={[16, 16]}>
        {products.map((product) => (
          <Col key={product.item_id} xs={24} sm={12} md={8} lg={6} xl={4}>
            <Card
              hoverable
              cover={
                <img
                  alt={product.item_name}
                  src={product.image_url}
                  style={{ height: 200, objectFit: "cover" }}
                />
              }
              actions={[
                <Button type="text" icon={<EditOutlined />} key="edit" />,
                <Button
                  type="text"
                  danger
                  icon={<DeleteOutlined />}
                  key="delete"
                  onClick={() => onDelete(product.item_id)}
                />,
              ]}
            >
              <Card.Meta
                title={
                  <div
                    style={{
                      display: "flex",
                      justifyContent: "space-between",
                      alignItems: "center",
                    }}
                  >
                    <Typography.Text strong ellipsis>
                      {product.item_name}
                    </Typography.Text>
                  </div>
                }
                description={
                  <Space
                    direction="vertical"
                    size={4}
                    style={{ width: "100%" }}
                  >
                    <Space>
                      <PlatformBadge platform={product.platform} />
                      <Typography.Text type="secondary">
                        {product.item_sku}
                      </Typography.Text>
                    </Space>
                    <div
                      style={{
                        display: "flex",
                        justifyContent: "space-between",
                      }}
                    >
                      <Typography.Text strong type="warning">
                        {product.price === null || product.price === undefined
                          ? "—"
                          : new Intl.NumberFormat("id-ID", {
                              style: "currency",
                              currency: "IDR",
                              maximumFractionDigits: 0,
                            }).format(product.price)}
                      </Typography.Text>
                      <Typography.Text
                        type={
                          product.stock === null || product.stock === undefined
                            ? "secondary"
                            : product.stock === 0
                              ? "danger"
                              : product.stock <= 10
                                ? "warning"
                                : "secondary"
                        }
                      >
                        Stock:{" "}
                        {product.stock === null || product.stock === undefined
                          ? "—"
                          : product.stock}
                      </Typography.Text>
                    </div>
                  </Space>
                }
              />
            </Card>
          </Col>
        ))}
      </Row>
      <div style={{ marginTop: 16, textAlign: "right" }}>
        <Pagination
          current={page}
          pageSize={pageSize}
          total={total}
          onChange={(p, ps) => {
            onPageChange(p, ps);
          }}
          showSizeChanger
        />
      </div>
    </>
  );
}
