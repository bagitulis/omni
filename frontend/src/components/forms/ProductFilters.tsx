import { Input, Select, Segmented, Space } from "antd";
import {
  SearchOutlined,
  AppstoreOutlined,
  BarsOutlined,
} from "@ant-design/icons";

interface ProductFiltersProps {
  filters: {
    search: string;
    status: string;
    platform: string;
    category: string;
  };
  onFilterChange: (
    key: string,
    value: string | number | boolean | undefined,
  ) => void;
  viewMode: "grid" | "list";
  onViewModeChange: (mode: "grid" | "list") => void;
}

export function ProductFilters({
  filters,
  onFilterChange,
  viewMode,
  onViewModeChange,
}: ProductFiltersProps) {
  return (
    <div
      style={{
        display: "flex",
        justifyContent: "space-between",
        marginBottom: 16,
        flexWrap: "wrap",
        gap: 16,
      }}
    >
      <Space wrap>
        <Input
          prefix={<SearchOutlined />}
          placeholder="Search products..."
          value={filters.search}
          onChange={(e) => onFilterChange("search", e.target.value)}
          style={{ width: 240 }}
          allowClear
        />
        <Select
          value={filters.category}
          onChange={(val) => onFilterChange("category", val)}
          options={[
            { label: "All Categories", value: "all" },
            { label: "Electronics", value: "electronics" },
            { label: "Clothing", value: "clothing" },
            { label: "Home & Living", value: "home" },
          ]}
          style={{ width: 150 }}
          placeholder="Category"
        />
        <Select
          value={filters.platform}
          onChange={(val) => onFilterChange("platform", val)}
          options={[
            { label: "All Platforms", value: "all" },
            { label: "Shopee", value: "shopee" },
            { label: "Lazada", value: "lazada" },
            { label: "TikTok", value: "tiktok" },
          ]}
          style={{ width: 150 }}
          placeholder="Platform"
        />
        <Select
          value={filters.status}
          onChange={(val) => onFilterChange("status", val)}
          options={[
            { label: "All Status", value: "all" },
            { label: "Active", value: "active" },
            { label: "Inactive", value: "inactive" },
            { label: "Draft", value: "draft" },
          ]}
          style={{ width: 150 }}
          placeholder="Status"
        />
      </Space>

      <Segmented
        value={viewMode}
        onChange={(val) => onViewModeChange(val as "grid" | "list")}
        options={[
          { value: "list", icon: <BarsOutlined /> },
          { value: "grid", icon: <AppstoreOutlined /> },
        ]}
      />
    </div>
  );
}
