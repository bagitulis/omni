import { Space, Select, Button, Tag, theme } from "antd";
import { FilterOutlined, ClearOutlined } from "@ant-design/icons";
import { useInventoryFilterStore } from "@/stores/inventoryFilterStore";

export function InventoryToolbar() {
  const {
    token: { colorBgContainer, colorBorderSecondary },
  } = theme.useToken();

  const {
    platform,
    syncStatus,
    setPlatformFilter,
    setSyncStatusFilter,
    clearFilters,
    getActiveFilterCount,
  } = useInventoryFilterStore();

  const activeCount = getActiveFilterCount();

  return (
    <div
      style={{
        display: "flex",
        alignItems: "center",
        gap: 12,
        padding: "8px 16px",
        background: colorBgContainer,
        borderBottom: `1px solid ${colorBorderSecondary}`,
        minHeight: 48,
      }}
    >
      {/* Filter Icon + Label */}
      <Space size={4}>
        <FilterOutlined style={{ fontSize: 14 }} />
        <span style={{ fontSize: 12, fontWeight: 500 }}>Filters:</span>
      </Space>

      {/* Platform Filter */}
      <Select
        mode="multiple"
        placeholder="Platform"
        value={platform}
        onChange={setPlatformFilter}
        style={{ minWidth: 160, fontSize: 12 }}
        options={[
          { label: "Shopee", value: "shopee" },
          { label: "Lazada", value: "lazada" },
          { label: "TikTok", value: "tiktok" },
        ]}
        maxTagCount="responsive"
        size="small"
      />

      {/* Sync Status Filter */}
      <Select
        mode="multiple"
        placeholder="Sync Status"
        value={syncStatus}
        onChange={setSyncStatusFilter}
        style={{ minWidth: 160, fontSize: 12 }}
        options={[
          { label: "Synced", value: "synced" },
          { label: "Not Synced", value: "not_synced" },
          { label: "Error", value: "error" },
        ]}
        maxTagCount="responsive"
        size="small"
      />

      {/* Filter Chips (Active Filters) */}
      <div style={{ display: "flex", gap: 4, flex: 1, overflowX: "auto" }}>
        {platform.map((p) => (
          <Tag
            key={p}
            closable
            onClose={() => setPlatformFilter(platform.filter((x) => x !== p))}
            style={{ fontSize: 12 }}
          >
            {p}
          </Tag>
        ))}
        {syncStatus.map((s) => (
          <Tag
            key={s}
            closable
            onClose={() =>
              setSyncStatusFilter(syncStatus.filter((x) => x !== s))
            }
            style={{ fontSize: 12 }}
          >
            {s.replace("_", " ")}
          </Tag>
        ))}
      </div>

      {/* Clear All Button */}
      {activeCount > 0 && (
        <Button
          type="text"
          danger
          size="small"
          icon={<ClearOutlined />}
          onClick={clearFilters}
          style={{ fontSize: 12 }}
        >
          Clear ({activeCount})
        </Button>
      )}
    </div>
  );
}
