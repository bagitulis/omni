import { useState } from "react";
import { Avatar, Button, Typography, Tag, Space, Flex } from "antd";
import {
  UserOutlined,
  MessageOutlined,
  CopyOutlined,
  CheckOutlined,
  ShoppingOutlined,
} from "@ant-design/icons";
import { GroupedOrder } from "./OrderTable.types";

const { Text } = Typography;

export function CopyButton({ text }: { text: string }) {
  const [copied, setCopied] = useState(false);

  const handleCopy = () => {
    navigator.clipboard.writeText(text);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  };

  return (
    <Button
      type="text"
      size="small"
      icon={
        copied ? (
          <CheckOutlined style={{ color: "#52c41a" }} />
        ) : (
          <CopyOutlined style={{ color: "#999" }} />
        )
      }
      onClick={handleCopy}
      style={{ padding: 0, height: "auto" }}
    />
  );
}

// Product Cell with Buyer Header
export function ProductCellWithBuyer({ order }: { order: GroupedOrder }) {
  return (
    <Flex vertical gap={8}>
      {/* Buyer Header */}
      <Flex
        justify="space-between"
        align="center"
        style={{
          backgroundColor: "#f5f5f5",
          padding: "8px 12px",
          borderRadius: 4,
          marginBottom: 4,
        }}
      >
        <Space size="small">
          <Avatar
            size={24}
            icon={<UserOutlined />}
            style={{ backgroundColor: "#0369a1" }}
          />
          <Text strong style={{ fontSize: 13 }}>
            {order.buyer_username}
          </Text>
          <Button
            type="text"
            size="small"
            icon={<MessageOutlined style={{ color: "#999", fontSize: 12 }} />}
            style={{ padding: 0 }}
          />
        </Space>
        <Space size={4}>
          <Text type="secondary" style={{ fontSize: 11 }}>
            ID:
          </Text>
          <Text style={{ fontSize: 12, fontFamily: "monospace" }}>
            {order.order_no}
          </Text>
          <CopyButton text={order.order_no} />
        </Space>
      </Flex>

      {/* Products */}
      <Flex vertical gap={10}>
        {order.items.map((item, idx) => (
          <Flex
            key={`${item.sku || "item"}-${idx}`}
            gap={10}
            align="flex-start"
          >
            <Avatar
              shape="square"
              size={52}
              src={item.product_image}
              icon={<ShoppingOutlined />}
              style={{
                flexShrink: 0,
                border: "1px solid #e0e0e0",
                backgroundColor: "#fafafa",
              }}
            />
            <Flex vertical gap={2} style={{ flex: 1, minWidth: 0 }}>
              <Flex justify="space-between" align="flex-start" gap={8}>
                <Text
                  style={{
                    fontSize: 13,
                    fontWeight: 500,
                    lineHeight: 1.3,
                    display: "-webkit-box",
                    WebkitLineClamp: 2,
                    WebkitBoxOrient: "vertical",
                    overflow: "hidden",
                  }}
                  title={item.product_name}
                >
                  {item.product_name}
                </Text>
                <Tag
                  color="default"
                  style={{ margin: 0, flexShrink: 0, fontSize: 11 }}
                >
                  x{item.qty}
                </Tag>
              </Flex>
              {item.variation_name && (
                <Text type="secondary" style={{ fontSize: 11 }}>
                  Var: {item.variation_name}
                </Text>
              )}
              {item.sku && (
                <Text
                  type="secondary"
                  style={{ fontSize: 10, fontFamily: "monospace" }}
                >
                  SKU: {item.sku}
                </Text>
              )}
            </Flex>
          </Flex>
        ))}
      </Flex>
    </Flex>
  );
}
