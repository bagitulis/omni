import { Modal, Descriptions, Tag, Button, Typography } from "antd";
import { Product } from "@/types/product";
import { theme } from "antd";

const { Text } = Typography;

interface ProductDetailModalProps {
  open: boolean;
  onClose: () => void;
  product: Product | null;
}

export function ProductDetailModal({
  open,
  onClose,
  product,
}: ProductDetailModalProps) {
  const { token } = theme.useToken();

  if (!product) return null;

  return (
    <Modal
      title="Product Details"
      open={open}
      onCancel={onClose}
      footer={[
        <Button key="close" onClick={onClose}>
          Close
        </Button>,
      ]}
      width={600}
    >
      <div style={{ display: "flex", flexDirection: "column", gap: 24 }}>
        <div
          style={{
            borderBottom: `1px solid ${token.colorBorderSecondary}`,
            paddingBottom: 16,
          }}
        >
          <Typography.Title level={5} style={{ margin: 0 }}>
            Item ID: {product.item_id}
          </Typography.Title>
          <Text type="secondary">Shop ID: {product.tenant_id}</Text>
        </div>

        <Descriptions column={1} bordered size="small">
          <Descriptions.Item label="ID">{product.id}</Descriptions.Item>
          <Descriptions.Item label="Item ID">
            <Text code>{product.item_id}</Text>
          </Descriptions.Item>
          <Descriptions.Item label="Shop ID">
            {product.tenant_id}
          </Descriptions.Item>
          <Descriptions.Item label="Status">
            <Tag
              color={
                product.status === "active"
                  ? "success"
                  : product.status === "inactive"
                    ? "error"
                    : "default"
              }
            >
              {product.status.toUpperCase()}
            </Tag>
          </Descriptions.Item>
          <Descriptions.Item label="Created At">
            {new Date(product.created_at).toLocaleString()}
          </Descriptions.Item>
          <Descriptions.Item label="Updated At">
            {new Date(product.updated_at).toLocaleString()}
          </Descriptions.Item>
        </Descriptions>
      </div>
    </Modal>
  );
}
