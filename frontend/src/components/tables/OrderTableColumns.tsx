import { Button, Typography, Dropdown, Flex, theme } from "antd";
import type { ColumnsType } from "antd/es/table";
import { MoreOutlined } from "@ant-design/icons";
import { StatusPipeline } from "../ui/StatusPipeline";
import { GroupedOrder } from "./OrderTable.types";
import { ProductCellWithBuyer } from "./OrderTableComponents";
import {
  formatAmount,
  getCountdown,
  getCountdownColor,
  canShipOrder,
} from "./OrderTable.utils";

const { Text } = Typography;

interface GetOrderTableColumnsProps {
  onShip: (order: GroupedOrder) => void;
  onPrint: (order: GroupedOrder) => void;
  onCancel?: (order: GroupedOrder) => void;
  onViewDetail?: (order: GroupedOrder) => void;
}

export function getOrderTableColumns({
  onShip,
  onPrint,
  onCancel,
  onViewDetail,
}: GetOrderTableColumnsProps): ColumnsType<GroupedOrder> {
  const token = theme.getDesignToken();
  return [
    {
      title: "Product",
      key: "product",
      width: 320,
      render: (_, record) => <ProductCellWithBuyer order={record} />,
    },
    {
      title: "Amount Paid",
      key: "amount",
      width: 120,
      render: (_, record) => (
        <Flex vertical gap={2}>
          <Text strong style={{ color: token.colorPrimary, fontSize: 14 }}>
            {formatAmount(record.total_amount, record.currency)}
          </Text>
          <Text type="secondary" style={{ fontSize: 11 }}>
            {record.payment_method || "Online Payment"}
          </Text>
        </Flex>
      ),
    },
    {
      title: "Status",
      key: "status",
      width: 130,
      render: (_, record) => <StatusPipeline status={record.status} />,
    },
    {
      title: "Countdown",
      key: "countdown",
      width: 110,
      render: (_, record) => {
        // Prefer backend pre-computed countdown; fallback to client-side calculation
        const countdownText =
          record.countdown || getCountdown(record.ship_by_date);
        return (
          <Flex vertical gap={2}>
            <Text
              style={{
                color: getCountdownColor(token, record.ship_by_date),
                fontWeight: 500,
                fontSize: 13,
              }}
            >
              {countdownText}
            </Text>
            {record.ship_by_date && (
              <Text type="secondary" style={{ fontSize: 10 }}>
                Ship by:{" "}
                {new Date(record.ship_by_date * 1000).toLocaleDateString()}
              </Text>
            )}
          </Flex>
        );
      },
    },
    {
      title: "Shipping",
      key: "shipping",
      width: 120,
      render: (_, record) => (
        <Text style={{ fontWeight: 500, fontSize: 12 }}>
          {record.shipping_carrier || "-"}
        </Text>
      ),
    },
    {
      title: "Action",
      key: "action",
      width: 90,
      fixed: "right",
      render: (_, record) => (
        <Flex vertical gap={6}>
          <Button
            type="primary"
            size="small"
            block
            disabled={!canShipOrder(record.status, record.platform)}
            onClick={() => onShip(record)}
            style={{
              backgroundColor: canShipOrder(record.status, record.platform)
                ? token.colorPrimary
                : undefined,
              fontSize: 12,
            }}
          >
            Ship
          </Button>
          <Flex justify="space-between" align="center">
            <Button
              type="link"
              size="small"
              style={{ padding: 0, fontSize: 11, height: "auto" }}
              onClick={() => onViewDetail?.(record)}
            >
              Details
            </Button>
            <Dropdown
              menu={{
                items: [
                  {
                    key: "view",
                    label: "View Details",
                    onClick: () => onViewDetail?.(record),
                  },
                  {
                    key: "print",
                    label: "Print Label",
                    onClick: () => onPrint(record),
                  },
                  {
                    key: "cancel",
                    label: "Cancel Order",
                    danger: true,
                    onClick: () => onCancel?.(record),
                  },
                ],
              }}
              trigger={["click"]}
            >
              <Button
                type="text"
                size="small"
                icon={<MoreOutlined />}
                style={{ padding: 0 }}
                aria-label="More actions"
              />
            </Dropdown>
          </Flex>
        </Flex>
      ),
    },
  ];
}
