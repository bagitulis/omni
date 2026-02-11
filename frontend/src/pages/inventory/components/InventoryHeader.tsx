import { useState } from "react";
import {
  Input,
  Button,
  Popover,
  Checkbox,
  Space,
  Typography,
  Tooltip,
} from "antd";
import {
  SearchOutlined,
  SettingOutlined,
  ReloadOutlined,
  CloudDownloadOutlined,
  CloudUploadOutlined,
  ShopOutlined,
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

        <Tooltip title="Pull latest data from Google Sheets">
          <Button
            icon={<CloudDownloadOutlined />}
            onClick={onSyncFromSheets}
            loading={syncingFromSheets}
          >
            Sync from Sheets
          </Button>
        </Tooltip>

        <Tooltip title="Push changes to Google Sheets">
          <Button
            icon={<CloudUploadOutlined />}
            onClick={onSyncToSheets}
            loading={syncingToSheets}
          >
            Sync to Sheets
          </Button>
        </Tooltip>

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
                <div style={{ padding: 8, color: "#999" }}>
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
