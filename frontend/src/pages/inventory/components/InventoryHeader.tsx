import { useState } from "react";
import {
  Input,
  Button,
  Grid,
  Typography,
  Tooltip,
  Dropdown,
} from "antd";
import {
  SearchOutlined,
  ReloadOutlined,
  CloudDownloadOutlined,
  CloudUploadOutlined,
  ShopOutlined,
  SyncOutlined,
  SettingOutlined,
  DownOutlined,
} from "@ant-design/icons";
import { MarketplaceSettingsModal } from "@/components/modals/MarketplaceSettingsModal";
import {
  buildInventoryColumnControlOrder,
  type InventoryColumnMoveDirection,
} from "../utils/inventoryColumnOrder";
import { ColumnSettingsPopover } from "./ColumnSettingsPopover";

interface InventoryHeaderProps {
  searchText: string;
  onSearch: (value: string) => void;
  onRefresh: () => void;
  loading: boolean;
  onSyncFromSheets: () => void;
  syncingFromSheets: boolean;
  onSyncToSheets: () => void;
  syncingToSheets: boolean;
  availableColumns: string[];
  visibleColumns: string[];
  onToggleColumn: (column: string, checked: boolean) => void;
  onMoveColumn: (column: string, direction: InventoryColumnMoveDirection) => void;
  onOpenBulkPricing?: () => void;
}

export function InventoryHeader({
  searchText,
  onSearch,
  onRefresh,
  loading,
  onSyncFromSheets,
  syncingFromSheets,
  onSyncToSheets,
  syncingToSheets,
  availableColumns,
  visibleColumns,
  onToggleColumn,
  onMoveColumn,
  onOpenBulkPricing,
}: InventoryHeaderProps) {
  const screens = Grid.useBreakpoint();
  const isMobile = !screens.md;
  const [settingsModalOpen, setSettingsModalOpen] = useState(false);

  const orderedColumns = buildInventoryColumnControlOrder(availableColumns, visibleColumns);
  const visibleSet = new Set(visibleColumns);
  const selectedColumns = orderedColumns.filter((c) => visibleSet.has(c));
  const hiddenColumns = orderedColumns.filter((c) => !visibleSet.has(c));

  const syncMenuItems = [
    { key: "sync-from", label: "Sync From Sheets", icon: <CloudDownloadOutlined />, onClick: onSyncFromSheets, disabled: syncingFromSheets },
    { key: "sync-to", label: "Sync To Sheets", icon: <CloudUploadOutlined />, onClick: onSyncToSheets, disabled: syncingToSheets },
    { key: "bulk-pricing", label: "Bulk Pricing (MPQ/Wholesale)", icon: <SettingOutlined />, onClick: onOpenBulkPricing },
  ];

  return (
    <div
      style={{
        marginBottom: 16,
        display: "flex",
        flexDirection: isMobile ? "column" : "row",
        justifyContent: "space-between",
        alignItems: isMobile ? "stretch" : "center",
        gap: 12,
      }}
    >
      <Typography.Title
        level={isMobile ? 4 : 2}
        style={{ margin: 0, whiteSpace: isMobile ? "normal" : "nowrap", lineHeight: 1.2 }}
      >
        Inventory
      </Typography.Title>

      <div
        style={{
          display: "flex",
          flexWrap: "wrap",
          alignItems: isMobile ? "stretch" : "center",
          gap: 8,
          justifyContent: isMobile ? "flex-start" : "flex-end",
          width: "100%",
        }}
      >
        <Input
          placeholder="Search"
          prefix={<SearchOutlined />}
          value={searchText}
          onChange={(e) => onSearch(e.target.value)}
          style={{
            width: isMobile ? "100%" : 220,
            minWidth: isMobile ? "100%" : 180,
            flex: isMobile ? "1 1 100%" : undefined,
          }}
          size="middle"
          allowClear
        />

        <Tooltip title="Refresh from database">
          <Button
            icon={<ReloadOutlined />}
            onClick={onRefresh}
            loading={loading}
            size="middle"
            style={isMobile ? { flex: "1 1 calc(33% - 6px)", minWidth: 96 } : undefined}
          />
        </Tooltip>

        <Dropdown menu={{ items: syncMenuItems }}>
          <Button
            icon={<SyncOutlined />}
            loading={syncingFromSheets || syncingToSheets}
            size="middle"
            style={isMobile ? { width: "100%" } : undefined}
          >
            {isMobile ? <DownOutlined /> : <>Sync <DownOutlined /></>}
          </Button>
        </Dropdown>

        <Button
          icon={<ShopOutlined />}
          onClick={() => setSettingsModalOpen(true)}
          size="middle"
          style={isMobile ? { flex: "1 1 calc(50% - 8px)", minWidth: 140 } : undefined}
        >
          {isMobile ? "Marketplace" : "Marketplace Settings"}
        </Button>

        <ColumnSettingsPopover
          selectedColumns={selectedColumns}
          hiddenColumns={hiddenColumns}
          onToggle={onToggleColumn}
          onMove={onMoveColumn}
          isMobile={isMobile}
        />
      </div>

      <MarketplaceSettingsModal
        open={settingsModalOpen}
        onClose={() => setSettingsModalOpen(false)}
        schemaColumns={availableColumns.map((c) => ({ column_name: c }))}
      />
    </div>
  );
}
