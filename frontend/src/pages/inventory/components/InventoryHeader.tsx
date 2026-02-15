import { useState } from "react";
import {
  Input,
  Button,
  Popover,
  Checkbox,
  Space,
  Typography,
  Tooltip,
  theme,
  Dropdown,
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
} from "@ant-design/icons";
import { useSelectedColumns, useAvailableColumns } from "@/hooks/useInventory";
import { useMutation, useQueryClient } from "@tanstack/react-query";
import { saveSelectedColumns } from "@/api/inventory";
import { MarketplaceSettingsModal } from "@/components/modals/MarketplaceSettingsModal";

interface InventoryHeaderProps {
  searchText: string;
  onSearch: (value: string) => void;
  onRefresh: () => void;
  loading: boolean;
  onSyncFromSheets: () => void;
  syncingFromSheets: boolean;
  onSyncToSheets: () => void;
  syncingToSheets: boolean;
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
}: InventoryHeaderProps) {
  const { token } = theme.useToken();
  const [settingsModalOpen, setSettingsModalOpen] = useState(false);
  const queryClient = useQueryClient();
  const { data: selectedCols = [] } = useSelectedColumns();
  const { data: availableCols = [] } = useAvailableColumns();
  const schemaColumns = availableCols.map((column) => ({
    column_name: column,
  }));

  const saveColumnsMutation = useMutation({
    mutationFn: saveSelectedColumns,
    onSuccess: () => {
      queryClient.invalidateQueries({
        queryKey: ["inventory-columns-selected"],
      });
    },
  });

  const handleColumnToggle = (col: string, checked: boolean) => {
    const newCols = checked
      ? [...selectedCols, col]
      : selectedCols.filter((c) => c !== col);
    saveColumnsMutation.mutate(newCols);
  };

  return (
    <div
      style={{
        marginBottom: 16,
        display: "flex",
        justifyContent: "space-between",
        alignItems: "center",
      }}
    >
      <Typography.Title level={2} style={{ margin: 0 }}>
        Inventory
      </Typography.Title>
      <Space>
        <Input
          placeholder="Search"
          prefix={<SearchOutlined />}
          value={searchText}
          onChange={(e) => onSearch(e.target.value)}
          style={{ width: 200 }}
          allowClear
        />
        <Tooltip title="Refresh from database">
          <Button
            icon={<ReloadOutlined />}
            onClick={onRefresh}
            loading={loading}
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
          >
            Sync <DownOutlined />
          </Button>
        </Dropdown>

        <Button
          icon={<ShopOutlined />}
          onClick={() => setSettingsModalOpen(true)}
        >
          Marketplace Settings
        </Button>

        <Popover
          trigger="click"
          placement="bottomRight"
          title="Columns"
          content={
            <div
              style={{
                display: "flex",
                flexDirection: "column",
                maxHeight: 300,
                overflowY: "auto",
              }}
            >
              {availableCols.length > 0 ? (
                availableCols.map((col) => (
                  <Checkbox
                    key={col}
                    checked={selectedCols.includes(col)}
                    onChange={(e) => handleColumnToggle(col, e.target.checked)}
                  >
                    {col}
                  </Checkbox>
                ))
              ) : (
                <div style={{ padding: 8, color: token.colorTextSecondary }}>
                  No columns available
                </div>
              )}
            </div>
          }
        >
          <Button icon={<SettingOutlined />}>Cols</Button>
        </Popover>
      </Space>

      <MarketplaceSettingsModal
        open={settingsModalOpen}
        onClose={() => setSettingsModalOpen(false)}
        schemaColumns={schemaColumns}
      />
    </div>
  );
}
