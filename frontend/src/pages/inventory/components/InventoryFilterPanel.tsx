import { useState } from "react";
import {
  Layout,
  Checkbox,
  Badge,
  Button,
  theme,
  Collapse,
  Typography,
  Space,
} from "antd";
import { FilterOutlined, ClearOutlined } from "@ant-design/icons";
import { useInventoryFilterStore } from "@/stores/inventoryFilterStore";

const { Sider } = Layout;
const { Title } = Typography;
const { Panel } = Collapse;

export function InventoryFilterPanel() {
  const {
    token: { colorBgContainer, colorBorderSecondary, colorSuccess },
  } = theme.useToken();

  const [collapsed, setCollapsed] = useState(false);

  const {
    platform,
    stockStatus,
    syncStatus,
    setPlatformFilter,
    setStockStatusFilter,
    setSyncStatusFilter,
    clearFilters,
    getActiveFilterCount,
  } = useInventoryFilterStore();

  const activeCount = getActiveFilterCount();

  const platformOptions = [
    { label: "Shopee", value: "shopee" },
    { label: "Lazada", value: "lazada" },
    { label: "TikTok", value: "tiktok" },
  ];

  const stockOptions = [
    { label: "In Stock", value: "in_stock" },
    { label: "Low Stock", value: "low_stock" },
    { label: "Out of Stock", value: "out_of_stock" },
  ];

  const syncOptions = [
    { label: "Synced", value: "synced" },
    { label: "Not Synced", value: "not_synced" },
    { label: "Error", value: "error" },
  ];

  return (
    <Sider
      theme="light"
      collapsible
      collapsed={collapsed}
      onCollapse={setCollapsed}
      width={240}
      collapsedWidth={50}
      style={{
        background: colorBgContainer,
        borderRight: `1px solid ${colorBorderSecondary}`,
        height: "100%",
        zIndex: 10,
      }}
    >
      <div
        style={{
          display: "flex",
          flexDirection: "column",
          height: "100%",
          padding: collapsed ? "16px 8px" : "16px",
        }}
      >
        {/* Header */}
        <div
          style={{
            display: "flex",
            alignItems: "center",
            justifyContent: collapsed ? "center" : "space-between",
            marginBottom: 16,
          }}
        >
          {collapsed ? (
            <Badge count={activeCount} size="small" offset={[5, 0]}>
              <FilterOutlined style={{ fontSize: 18 }} />
            </Badge>
          ) : (
            <>
              <Space>
                <FilterOutlined />
                <Title level={5} style={{ margin: 0 }}>
                  Filters
                </Title>
              </Space>
              {activeCount > 0 && (
                <Badge
                  count={activeCount}
                  style={{ backgroundColor: colorSuccess }}
                />
              )}
            </>
          )}
        </div>

        {/* Filter Content */}
        {!collapsed && (
          <div style={{ flex: 1, overflowY: "auto" }}>
            <Collapse
              defaultActiveKey={["platform", "stock", "sync"]}
              ghost
              expandIconPosition="end"
              size="small"
            >
              <Panel header="Platform" key="platform">
                <Checkbox.Group
                  options={platformOptions}
                  value={platform}
                  onChange={(checked) => setPlatformFilter(checked as string[])}
                  style={{ display: "flex", flexDirection: "column", gap: 8 }}
                />
              </Panel>
              <Panel header="Stock Status" key="stock">
                <Checkbox.Group
                  options={stockOptions}
                  value={stockStatus}
                  onChange={(checked) =>
                    setStockStatusFilter(checked as string[])
                  }
                  style={{ display: "flex", flexDirection: "column", gap: 8 }}
                />
              </Panel>
              <Panel header="Sync Status" key="sync">
                <Checkbox.Group
                  options={syncOptions}
                  value={syncStatus}
                  onChange={(checked) =>
                    setSyncStatusFilter(checked as string[])
                  }
                  style={{ display: "flex", flexDirection: "column", gap: 8 }}
                />
              </Panel>
            </Collapse>
          </div>
        )}

        {/* Footer Actions */}
        {!collapsed && (
          <div
            style={{
              paddingTop: 16,
              borderTop: `1px solid ${colorBorderSecondary}`,
            }}
          >
            <Button
              type="text"
              danger
              icon={<ClearOutlined />}
              onClick={clearFilters}
              disabled={activeCount === 0}
              block
            >
              Clear All
            </Button>
          </div>
        )}
      </div>
    </Sider>
  );
}
