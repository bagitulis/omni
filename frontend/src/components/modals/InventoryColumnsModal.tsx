import {
  Alert,
  Button,
  Checkbox,
  Collapse,
  Empty,
  Modal,
  Space,
  Spin,
  Tag,
  Typography,
  message,
  theme,
} from "antd";
import { useEffect, useMemo, useState } from "react";
import { useMutation, useQueryClient } from "@tanstack/react-query";
import { saveSelectedColumns } from "@/api/inventory";
import { useAvailableColumns, useSelectedColumns } from "@/hooks/useInventory";
import { buildGroups } from "./inventoryColumnGroups";

const { Text } = Typography;

interface InventoryColumnsModalProps {
  open: boolean;
  onClose: () => void;
}

export function InventoryColumnsModal({
  open,
  onClose,
}: InventoryColumnsModalProps) {
  const {
    token: { colorFillAlter, colorBgContainer, colorBorderSecondary, borderRadius },
  } = theme.useToken();
  const queryClient = useQueryClient();
  const [selectedColumns, setSelectedColumns] = useState<string[]>([]);

  const {
    data: availableColumns = [],
    isLoading: isLoadingAvailable,
    isError: isAvailableError,
    error: availableError,
    refetch: refetchAvailable,
  } = useAvailableColumns();
  const {
    data: selectedColumnsFromApi = [],
    isLoading: isLoadingSelected,
    isError: isSelectedError,
    error: selectedError,
    refetch: refetchSelected,
  } = useSelectedColumns();

  const saveColumnsMutation = useMutation({
    mutationFn: saveSelectedColumns,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["inventory-columns-selected"] });
      queryClient.invalidateQueries({ queryKey: ["inventory"] });
      message.success("Columns configuration saved");
      onClose();
    },
    onError: (error: Error) => {
      message.error(error.message || "Failed to save configuration");
    },
  });

  useEffect(() => {
    if (open) setSelectedColumns(selectedColumnsFromApi);
  }, [open, selectedColumnsFromApi]);

  const groupedColumns = useMemo(() => buildGroups(availableColumns), [availableColumns]);
  const loading = isLoadingAvailable || isLoadingSelected;
  const hasError = isAvailableError || isSelectedError;
  const errorMessage =
    (availableError as Error | null)?.message ||
    (selectedError as Error | null)?.message;

  const hasChanges = useMemo(() => {
    if (selectedColumns.length !== selectedColumnsFromApi.length) return true;
    const set = new Set(selectedColumns);
    return selectedColumnsFromApi.some((c) => !set.has(c));
  }, [selectedColumns, selectedColumnsFromApi]);

  const toggleColumn = (name: string) => {
    setSelectedColumns((prev) =>
      prev.includes(name) ? prev.filter((c) => c !== name) : [...prev, name],
    );
  };

  const renderContent = () => {
    if (loading) {
      return (
        <div style={{ textAlign: "center", padding: 24 }}>
          <Spin tip="Loading columns..." />
        </div>
      );
    }

    if (hasError) {
      return (
        <Alert
          type="error"
          showIcon
          message="Failed to load column configuration"
          description={errorMessage || "Please try again."}
          action={
            <Button size="small" onClick={() => { refetchAvailable(); refetchSelected(); }}>
              Retry
            </Button>
          }
        />
      );
    }

    return (
      <div style={{ display: "flex", flexDirection: "column", gap: 16 }}>
        <Text type="secondary">
          Select which columns from Google Sheets to display in the inventory table.
          Total: {availableColumns.length}
        </Text>

        {selectedColumns.length > 0 && (
          <div
            style={{
              background: colorFillAlter,
              border: `1px solid ${colorBorderSecondary}`,
              borderRadius,
              padding: 12,
            }}
          >
            <Text strong>Selected ({selectedColumns.length})</Text>
            <Space size={[8, 8]} wrap style={{ marginTop: 8 }}>
              {selectedColumns.map((col) => (
                <Tag key={col} closable onClose={(e) => { e.preventDefault(); toggleColumn(col); }}>
                  {col}
                </Tag>
              ))}
            </Space>
          </div>
        )}

        <div>
          <div style={{ display: "flex", justifyContent: "space-between", marginBottom: 8 }}>
            <Text strong>Available</Text>
            <Space>
              <Button size="small" onClick={() => setSelectedColumns(availableColumns)}>
                Select All
              </Button>
              <Button size="small" onClick={() => setSelectedColumns([])}>
                Clear All
              </Button>
            </Space>
          </div>

          {groupedColumns.length === 0 ? (
            <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description="No columns available" />
          ) : (
            <Collapse
              size="small"
              defaultActiveKey={groupedColumns.map((g) => g.key)}
              items={groupedColumns.map((group) => ({
                key: group.key,
                label: group.label,
                children: (
                  <div
                    style={{
                      display: "grid",
                      gridTemplateColumns: "repeat(auto-fill, minmax(180px, 1fr))",
                      gap: 8,
                      paddingTop: 4,
                    }}
                  >
                    {group.children.map((col) => (
                      <Checkbox
                        key={col}
                        checked={selectedColumns.includes(col)}
                        onChange={() => toggleColumn(col)}
                        style={{
                          background: colorBgContainer,
                          border: `1px solid ${colorBorderSecondary}`,
                          borderRadius,
                          padding: "6px 8px",
                        }}
                      >
                        {col}
                      </Checkbox>
                    ))}
                  </div>
                ),
              }))}
            />
          )}
        </div>
      </div>
    );
  };

  return (
    <Modal
      title="Inventory Column Configuration"
      open={open}
      onCancel={onClose}
      width={700}
      footer={[
        <Button key="cancel" onClick={onClose}>Cancel</Button>,
        <Button
          key="save"
          type="primary"
          onClick={() => saveColumnsMutation.mutate(selectedColumns)}
          loading={saveColumnsMutation.isPending}
          disabled={loading || !hasChanges}
        >
          Save
        </Button>,
      ]}
    >
      {renderContent()}
    </Modal>
  );
}
