import { useState } from "react";
import { Button, Dropdown, Grid, Input, Tooltip, Typography } from "antd";
import {
  CloudDownloadOutlined,
  CloudUploadOutlined,
  DownOutlined,
  ReloadOutlined,
  SearchOutlined,
  ShopOutlined,
  SyncOutlined,
} from "@ant-design/icons";
import { MarketplaceSettingsModal } from "@/components/modals/MarketplaceSettingsModal";
import { ColumnManager } from "@/components/shared/ColumnManager";
import type { ColumnConfig } from "@/types/shared";
import { SyncStatusIndicator } from "@/components/shared/SyncStatusIndicator";
interface SimplifiedInventoryHeaderProps {
  searchText: string;
  onSearch: (value: string) => void;
  onRefresh: () => void;
  loading: boolean;
  onSyncFromSheets: () => void;
  syncingFromSheets: boolean;
  onSyncToSheets: () => void;
  syncingToSheets: boolean;
  columnConfigs: ColumnConfig[];
  schemaColumns: Array<{ column_name: string; column_type?: string }>;
  onColumnChange: (columns: ColumnConfig[]) => void;
  onColumnReset: () => void;
}

export function SimplifiedInventoryHeader({
  searchText,
  onSearch,
  onRefresh,
  loading,
  onSyncFromSheets,
  syncingFromSheets,
  onSyncToSheets,
  syncingToSheets,
  columnConfigs,
  schemaColumns,
  onColumnChange,
  onColumnReset,
}: SimplifiedInventoryHeaderProps) {
  const screens = Grid.useBreakpoint();
  const isMobile = !screens.md;
  const [settingsModalOpen, setSettingsModalOpen] = useState(false);

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
      <div style={{ display: "flex", alignItems: "center", gap: 12 }}>
        <Typography.Title
          level={isMobile ? 4 : 2}
          style={{
            margin: 0,
            whiteSpace: isMobile ? "normal" : "nowrap",
            lineHeight: 1.2,
          }}
        >
          Inventory
        </Typography.Title>
        <SyncStatusIndicator />
      </div>

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
          />
        </Tooltip>

        <Dropdown
          menu={{
            items: [
              {
                key: "sync-from",
                label: "Sync From Sheets",
                icon: <CloudDownloadOutlined />,
                onClick: onSyncFromSheets,
                disabled: syncingFromSheets,
              },
              {
                key: "sync-to",
                label: "Sync To Sheets",
                icon: <CloudUploadOutlined />,
                onClick: onSyncToSheets,
                disabled: syncingToSheets,
              },
            ],
          }}
        >
          <Button
            icon={<SyncOutlined />}
            loading={syncingFromSheets || syncingToSheets}
            size="middle"
          >
            {isMobile ? (
              <DownOutlined />
            ) : (
              <>
                Sync <DownOutlined />
              </>
            )}
          </Button>
        </Dropdown>

        <Button
          icon={<ShopOutlined />}
          onClick={() => setSettingsModalOpen(true)}
          size="middle"
          style={isMobile ? { width: "100%" } : undefined}
        >
          {isMobile ? "Marketplace" : "Marketplace Allocation"}
        </Button>

        <ColumnManager
          columns={columnConfigs}
          onChange={onColumnChange}
          onReset={onColumnReset}
        />
      </div>

      <MarketplaceSettingsModal
        open={settingsModalOpen}
        onClose={() => setSettingsModalOpen(false)}
        schemaColumns={schemaColumns}
      />
    </div>
  );
}
