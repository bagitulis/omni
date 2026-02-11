import { Button, Flex, Switch, Typography, Dropdown } from "antd";
import {
  SyncOutlined,
  DollarOutlined,
  SearchOutlined,
  EyeOutlined,
} from "@ant-design/icons";
import type { MenuProps } from "antd";

const { Title, Text } = Typography;

interface ProductManagerHeaderProps {
  autoRefresh: boolean;
  setAutoRefresh: (val: boolean) => void;
  autoSync: boolean;
  setAutoSync: (val: boolean) => void;
  syncLoading: boolean;
  onSync: () => void;
  onUpdateBatch?: () => void;
  hasSelection?: boolean;
  onCheckSku: () => void;
  skuCheckLoading: boolean;
  selectedCount: number;
  columnVisibility: Record<string, boolean>;
  onColumnVisibilityChange: (columnVisibility: Record<string, boolean>) => void;
}

export function ProductManagerHeader({
  autoRefresh,
  setAutoRefresh,
  autoSync,
  setAutoSync,
  syncLoading,
  onSync,
  onUpdateBatch,
  hasSelection,
  onCheckSku,
  skuCheckLoading,
  selectedCount,
  columnVisibility,
  onColumnVisibilityChange,
}: ProductManagerHeaderProps) {
  const columns = [
    { key: "image", label: "Image" },
    { key: "product", label: "Product" },
    { key: "price", label: "Price" },
    { key: "stock", label: "Stock" },
    { key: "status", label: "Status" },
    { key: "updated_at", label: "Updated" },
  ];

  const menuItems: MenuProps["items"] = columns.map((col) => ({
    key: col.key,
    label: (
      <Flex align="center" gap={8}>
        <input
          type="checkbox"
          checked={columnVisibility[col.key] !== false}
          onChange={(e) => {
            onColumnVisibilityChange({
              ...columnVisibility,
              [col.key]: e.target.checked,
            });
          }}
          style={{ cursor: "pointer" }}
        />
        <span>{col.label}</span>
      </Flex>
    ),
  }));
  return (
    <Flex justify="space-between" align="center">
      <div>
        <Title level={3} style={{ margin: 0 }}>
          Product Manager
        </Title>
        <Text type="secondary" style={{ fontSize: 12 }}>
          Platform products synced into database
        </Text>
      </div>
      <Flex align="center" gap={12}>
        <Flex align="center" gap={6}>
          <Text style={{ fontSize: 12 }}>Auto refresh</Text>
          <Switch
            checked={autoRefresh}
            onChange={setAutoRefresh}
            size="small"
          />
        </Flex>
        <Flex align="center" gap={6}>
          <Text style={{ fontSize: 12 }}>Auto sync</Text>
          <Switch checked={autoSync} onChange={setAutoSync} size="small" />
        </Flex>
        {onUpdateBatch && (
          <Button
            icon={<DollarOutlined />}
            onClick={onUpdateBatch}
            disabled={!hasSelection}
          >
            Update Prices
          </Button>
        )}
        <Button
          data-testid="product-manager-check-sku"
          icon={<SearchOutlined />}
          loading={skuCheckLoading}
          onClick={onCheckSku}
          disabled={selectedCount === 0}
        >
          Check SKU Status
          {selectedCount > 0 && ` (${selectedCount})`}
        </Button>
        <Dropdown menu={{ items: menuItems }} trigger={["click"]}>
          <Button icon={<EyeOutlined />}>Columns</Button>
        </Dropdown>
        <Button
          data-testid="product-manager-sync"
          type="primary"
          icon={<SyncOutlined />}
          loading={syncLoading}
          onClick={onSync}
        >
          Sync Now
        </Button>
      </Flex>
    </Flex>
  );
}
