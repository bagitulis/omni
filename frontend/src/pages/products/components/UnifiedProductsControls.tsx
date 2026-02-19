import { AppstoreOutlined, BarsOutlined } from "@ant-design/icons";
import { Button, Divider, Space } from "antd";
import { ColumnManager } from "@/components/shared/ColumnManager";
import { ProductFilters } from "@/components/shared/ProductFilters";
import type { ProductFilterValues } from "@/types/shared";
import type { ColumnConfig } from "@/types/shared";

interface UnifiedProductsControlsProps {
  filters: ProductFilterValues;
  onFilterChange: (nextFilters: ProductFilterValues) => void;
  viewMode: "grid" | "list";
  onViewModeChange: (mode: "grid" | "list") => void;
  isMobile: boolean;
  columns: ColumnConfig[];
  onColumnsChange: (columns: ColumnConfig[]) => void;
  onResetColumns: () => void;
}

export function UnifiedProductsControls({
  filters,
  onFilterChange,
  viewMode,
  onViewModeChange,
  isMobile,
  columns,
  onColumnsChange,
  onResetColumns,
}: UnifiedProductsControlsProps) {
  return (
    <>
      <ProductFilters
        values={filters}
        onChange={onFilterChange}
        showCategory={true}
        showStatus={true}
      />
      <Divider style={{ margin: "8px 0" }} />

      <Space
        style={{
          width: "100%",
          justifyContent: "space-between",
          marginBottom: 12,
        }}
      >
        <Space className="view-toggle">
          <Button
            data-view="list"
            type={viewMode === "list" ? "primary" : "default"}
            icon={<BarsOutlined />}
            onClick={() => onViewModeChange("list")}
          >
            List
          </Button>
          <Button
            data-view="grid"
            type={viewMode === "grid" ? "primary" : "default"}
            icon={<AppstoreOutlined />}
            onClick={() => onViewModeChange("grid")}
          >
            Grid
          </Button>
        </Space>

        {!isMobile ? (
          <ColumnManager
            columns={columns}
            onChange={onColumnsChange}
            onReset={onResetColumns}
          />
        ) : null}
      </Space>
    </>
  );
}
