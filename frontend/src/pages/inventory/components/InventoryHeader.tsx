import { useState } from "react";
import {
  Input,
  Button,
  Popover,
  Checkbox,
  Grid,
  Typography,
  Tooltip,
  theme,
  Dropdown,
  Space,
} from "antd";
import {
  SearchOutlined,
  SettingOutlined,
  ReloadOutlined,
  CloudDownloadOutlined,
  CloudUploadOutlined,
  ShopOutlined,
  SyncOutlined,
  DownOutlined,
  ArrowUpOutlined,
  ArrowDownOutlined,
} from "@ant-design/icons";
import { MarketplaceSettingsModal } from "@/components/modals/MarketplaceSettingsModal";
import {
  buildInventoryColumnControlOrder,
  type InventoryColumnMoveDirection,
} from "../utils/inventoryColumnOrder";

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
  onMoveColumn: (
    column: string,
    direction: InventoryColumnMoveDirection,
  ) => void;
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
  const { token } = theme.useToken();
  const [settingsModalOpen, setSettingsModalOpen] = useState(false);
  const schemaColumns = availableColumns.map((column) => ({
    column_name: column,
  }));

  const handleColumnToggle = (col: string, checked: boolean) => {
    onToggleColumn(col, checked);
  };

  const orderedColumns = buildInventoryColumnControlOrder(
    availableColumns,
    visibleColumns,
  );
  const visibleSet = new Set(visibleColumns);
  const selectedColumns = orderedColumns.filter((column) =>
    visibleSet.has(column),
  );
  const hiddenColumns = orderedColumns.filter(
    (column) => !visibleSet.has(column),
  );

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
        style={{
          margin: 0,
          whiteSpace: isMobile ? "normal" : "nowrap",
          lineHeight: 1.2,
        }}
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
            style={
              isMobile
                ? { flex: "1 1 calc(33% - 6px)", minWidth: 96 }
                : undefined
            }
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
              {
                key: "bulk-pricing",
                label: "Bulk Pricing (MPQ/Wholesale)",
                icon: <SettingOutlined />,
                onClick: onOpenBulkPricing,
              },
            ],
          }}
        >
          <Button
            icon={<SyncOutlined />}
            loading={syncingFromSheets || syncingToSheets}
            size="middle"
            style={isMobile ? { width: "100%" } : undefined}
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
          style={
            isMobile
              ? { flex: "1 1 calc(50% - 8px)", minWidth: 140 }
              : undefined
          }
        >
          {isMobile ? "Marketplace" : "Marketplace Settings"}
        </Button>

        <Popover
          trigger="click"
          placement="bottomRight"
          title="Columns & Order"
          content={
            <div
              style={{
                display: "flex",
                flexDirection: "column",
                maxHeight: 300,
                overflowY: "auto",
                minWidth: 280,
                gap: 10,
              }}
            >
              {orderedColumns.length > 0 ? (
                <>
                  <Typography.Text type="secondary" style={{ fontSize: 12 }}>
                    Selected columns (drag order equivalent)
                  </Typography.Text>
                  {selectedColumns.map((col, index) => {
                    const isFirst = index === 0;
                    const isLast = index === selectedColumns.length - 1;

                    return (
                      <div
                        key={col}
                        style={{
                          display: "flex",
                          justifyContent: "space-between",
                          alignItems: "center",
                          gap: 8,
                        }}
                      >
                        <Checkbox
                          checked
                          onChange={(event) =>
                            handleColumnToggle(col, event.target.checked)
                          }
                        >
                          {col}
                        </Checkbox>
                        <Space size={4}>
                          <Tooltip title="Move up">
                            <Button
                              size="small"
                              icon={<ArrowUpOutlined />}
                              onClick={() => onMoveColumn(col, "up")}
                              disabled={isFirst}
                            />
                          </Tooltip>
                          <Tooltip title="Move down">
                            <Button
                              size="small"
                              icon={<ArrowDownOutlined />}
                              onClick={() => onMoveColumn(col, "down")}
                              disabled={isLast}
                            />
                          </Tooltip>
                        </Space>
                      </div>
                    );
                  })}

                  {hiddenColumns.length > 0 ? (
                    <>
                      <Typography.Text
                        type="secondary"
                        style={{ fontSize: 12 }}
                      >
                        Hidden columns
                      </Typography.Text>
                      {hiddenColumns.map((col) => (
                        <Checkbox
                          key={col}
                          checked={false}
                          onChange={(event) =>
                            handleColumnToggle(col, event.target.checked)
                          }
                        >
                          {col}
                        </Checkbox>
                      ))}
                    </>
                  ) : null}
                </>
              ) : (
                <div style={{ padding: 8, color: token.colorTextSecondary }}>
                  No columns available
                </div>
              )}
            </div>
          }
        >
          <Button
            icon={<SettingOutlined />}
            size="middle"
            style={
              isMobile
                ? { flex: "1 1 calc(50% - 8px)", minWidth: 100 }
                : undefined
            }
          >
            Cols
          </Button>
        </Popover>
      </div>

      <MarketplaceSettingsModal
        open={settingsModalOpen}
        onClose={() => setSettingsModalOpen(false)}
        schemaColumns={schemaColumns}
      />
    </div>
  );
}
