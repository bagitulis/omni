import {
  Space,
  Select,
  Button,
  Tag,
  theme,
  Popover,
  Input,
  Typography,
  Grid,
} from "antd";
import { FilterOutlined, ClearOutlined } from "@ant-design/icons";
import { useInventoryFilterStore } from "@/stores/inventoryFilterStore";

interface InventoryToolbarProps {
  filterableColumns: string[];
}

export function InventoryToolbar({ filterableColumns }: InventoryToolbarProps) {
  const screens = Grid.useBreakpoint();
  const isMobile = !screens.md;
  const {
    token: { colorBgContainer, colorBorderSecondary },
  } = theme.useToken();

  const {
    search,
    platform,
    syncStatus,
    columnFilters,
    setSearch,
    setPlatformFilter,
    setSyncStatusFilter,
    setColumnFilter,
    clearColumnFilters,
    clearFilters,
    getActiveFilterCount,
  } = useInventoryFilterStore();

  const activeCount = getActiveFilterCount();
  const columnFilterEntries = Object.entries(columnFilters);

  const columnFilterContent = (
    <div
      style={{
        maxHeight: 320,
        overflowY: "auto",
        display: "flex",
        flexDirection: "column",
        gap: 8,
        minWidth: 260,
      }}
    >
      {filterableColumns.length === 0 ? (
        <Typography.Text type="secondary" style={{ fontSize: 12 }}>
          No sheet columns available yet
        </Typography.Text>
      ) : (
        filterableColumns.map((column) => (
          <div
            key={column}
            style={{ display: "flex", flexDirection: "column", gap: 4 }}
          >
            <Typography.Text style={{ fontSize: 12 }}>{column}</Typography.Text>
            <Input
              size="small"
              value={columnFilters[column] ?? ""}
              placeholder={`Filter ${column}`}
              allowClear
              onChange={(event) =>
                setColumnFilter(column, event.currentTarget.value)
              }
            />
          </div>
        ))
      )}

      {columnFilterEntries.length > 0 && (
        <Button size="small" onClick={clearColumnFilters}>
          Clear Column Filters
        </Button>
      )}
    </div>
  );

  return (
    <div
      style={{
        display: "flex",
        flexDirection: isMobile ? "column" : "row",
        alignItems: isMobile ? "stretch" : "center",
        flexWrap: "wrap",
        gap: 12,
        padding: isMobile ? "8px 12px" : "8px 16px",
        background: colorBgContainer,
        borderBottom: `1px solid ${colorBorderSecondary}`,
        minHeight: 48,
      }}
    >
      {/* Filter Icon + Label */}
      <Space size={4} style={isMobile ? { width: "100%" } : undefined}>
        <FilterOutlined style={{ fontSize: 14 }} />
        <span style={{ fontSize: 12, fontWeight: 500 }}>Filters:</span>
      </Space>

      {/* Platform Filter */}
      <Select
        mode="multiple"
        placeholder="Platform"
        value={platform}
        onChange={setPlatformFilter}
        style={{
          minWidth: isMobile ? "100%" : 160,
          width: isMobile ? "100%" : undefined,
          fontSize: 12,
          flex: isMobile ? "1 1 100%" : undefined,
        }}
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
        style={{
          minWidth: isMobile ? "100%" : 160,
          width: isMobile ? "100%" : undefined,
          fontSize: 12,
          flex: isMobile ? "1 1 100%" : undefined,
        }}
        options={[
          { label: "Synced", value: "synced" },
          { label: "Not Synced", value: "not_synced" },
          { label: "Error", value: "error" },
        ]}
        maxTagCount="responsive"
        size="small"
      />

      {/* Filter Chips (Active Filters) */}
      <div
        style={{
          display: "flex",
          gap: 4,
          flex: isMobile ? "1 1 100%" : 1,
          flexWrap: "wrap",
          minWidth: isMobile ? "100%" : 0,
        }}
      >
        {search ? (
          <Tag closable onClose={() => setSearch("")} style={{ fontSize: 12 }}>
            search: {search}
          </Tag>
        ) : null}
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
        {columnFilterEntries.map(([column, value]) => (
          <Tag
            key={column}
            closable
            onClose={() => setColumnFilter(column, "")}
            style={{ fontSize: 12 }}
          >
            {column}: {value}
          </Tag>
        ))}
      </div>

      <Popover
        trigger="click"
        placement="bottomRight"
        title={`Column Filters (${columnFilterEntries.length})`}
        content={columnFilterContent}
      >
        <Button size="small" style={isMobile ? { width: "100%" } : undefined}>
          Column Filters
        </Button>
      </Popover>

      {/* Clear All Button */}
      {activeCount > 0 && (
        <Button
          type="text"
          danger
          size="small"
          icon={<ClearOutlined />}
          onClick={clearFilters}
          style={{
            fontSize: 12,
            width: isMobile ? "100%" : undefined,
          }}
        >
          Clear ({activeCount})
        </Button>
      )}
    </div>
  );
}
