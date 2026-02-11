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

const { Text } = Typography;
const GROUPS = ["Product Info", "Stock", "Price", "Platform", "Other"] as const;

interface InventoryColumnsModalProps {
  open: boolean;
  onClose: () => void;
}

function detectGroup(column: string): (typeof GROUPS)[number] {
  const normalized = column.toLowerCase();
  if (
    normalized === "name" ||
    normalized === "product_name" ||
    normalized === "sku" ||
    normalized === "key_value" ||
    normalized.includes("name") ||
    normalized.includes("sku")
  ) {
    return "Product Info";
  }
  if (
    normalized === "stock" ||
    normalized === "stok" ||
    normalized.includes("stock") ||
    normalized.includes("stok")
  ) {
    return "Stock";
  }
  if (
    normalized === "price" ||
    normalized === "harga" ||
    normalized.includes("price") ||
    normalized.includes("harga")
  ) {
    return "Price";
  }
  if (
    normalized.startsWith("shopee_") ||
    normalized.startsWith("tiktok_") ||
    normalized.startsWith("lazada_")
  ) {
    return "Platform";
  }
  return "Other";
}

function buildGroups(columns: string[]) {
  const grouped = GROUPS.reduce<Record<(typeof GROUPS)[number], string[]>>(
    (acc, group) => ({ ...acc, [group]: [] }),
    { "Product Info": [], Stock: [], Price: [], Platform: [], Other: [] },
  );
  for (const column of columns) {
    grouped[detectGroup(column)].push(column);
  }
  return GROUPS.map((group) => ({
    key: group,
    label: `${group} (${grouped[group].length})`,
    children: grouped[group],
  })).filter((group) => group.children.length > 0);
}

export function InventoryColumnsModal({
  open,
  onClose,
}: InventoryColumnsModalProps) {
  const {
    token: {
      colorFillAlter,
      colorBgContainer,
      colorBorderSecondary,
      borderRadius,
    },
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
      queryClient.invalidateQueries({
        queryKey: ["inventory-columns-selected"],
      });
      queryClient.invalidateQueries({ queryKey: ["inventory"] });
      message.success("Columns configuration saved");
      onClose();
    },
    onError: (error: Error) => {
      message.error(error.message || "Failed to save configuration");
    },
  });

  useEffect(() => {
    if (open) {
      setSelectedColumns(selectedColumnsFromApi);
    }
  }, [open, selectedColumnsFromApi]);

  const groupedColumns = useMemo(
    () => buildGroups(availableColumns),
    [availableColumns],
  );
  const loading = isLoadingAvailable || isLoadingSelected;
  const hasError = isAvailableError || isSelectedError;
  const errorMessage =
    (availableError as Error | null)?.message ||
    (selectedError as Error | null)?.message;
  const hasChanges = useMemo(() => {
    if (selectedColumns.length !== selectedColumnsFromApi.length) return true;
    const selectedSet = new Set(selectedColumns);
    return selectedColumnsFromApi.some((column) => !selectedSet.has(column));
  }, [selectedColumns, selectedColumnsFromApi]);

  const toggleColumn = (columnName: string) => {
    setSelectedColumns((prev) =>
      prev.includes(columnName)
        ? prev.filter((column) => column !== columnName)
        : [...prev, columnName],
    );
  };

  return (
    <Modal
      title="Inventory Column Configuration"
      open={open}
      onCancel={onClose}
      width={700}
      footer={[
        <Button key="cancel" onClick={onClose}>
          Cancel
        </Button>,
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
      {loading ? (
        <div style={{ textAlign: "center", padding: 24 }}>
          <Spin tip="Loading columns..." />
        </div>
      ) : hasError ? (
        <Alert
          type="error"
          showIcon
          message="Failed to load column configuration"
          description={errorMessage || "Please try again."}
          action={
            <Button
              size="small"
              onClick={() => {
                refetchAvailable();
                refetchSelected();
              }}
            >
              Retry
            </Button>
          }
        />
      ) : (
        <div style={{ display: "flex", flexDirection: "column", gap: 16 }}>
          <Text type="secondary">
            Select which columns from Google Sheets to display in the inventory
            table. Total: {availableColumns.length}
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
                {selectedColumns.map((column) => (
                  <Tag
                    key={column}
                    closable
                    onClose={(event) => {
                      event.preventDefault();
                      toggleColumn(column);
                    }}
                  >
                    {column}
                  </Tag>
                ))}
              </Space>
            </div>
          )}

          <div>
            <div
              style={{
                display: "flex",
                justifyContent: "space-between",
                marginBottom: 8,
              }}
            >
              <Text strong>Available</Text>
              <Space>
                <Button
                  size="small"
                  onClick={() => setSelectedColumns(availableColumns)}
                >
                  Select All
                </Button>
                <Button size="small" onClick={() => setSelectedColumns([])}>
                  Clear All
                </Button>
              </Space>
            </div>

            {groupedColumns.length === 0 ? (
              <Empty
                image={Empty.PRESENTED_IMAGE_SIMPLE}
                description="No columns available"
              />
            ) : (
              <Collapse
                size="small"
                defaultActiveKey={groupedColumns.map((group) => group.key)}
                items={groupedColumns.map((group) => ({
                  key: group.key,
                  label: group.label,
                  children: (
                    <div
                      style={{
                        display: "grid",
                        gridTemplateColumns:
                          "repeat(auto-fill, minmax(180px, 1fr))",
                        gap: 8,
                        paddingTop: 4,
                      }}
                    >
                      {group.children.map((column) => (
                        <Checkbox
                          key={column}
                          checked={selectedColumns.includes(column)}
                          onChange={() => toggleColumn(column)}
                          style={{
                            background: colorBgContainer,
                            border: `1px solid ${colorBorderSecondary}`,
                            borderRadius,
                            padding: "6px 8px",
                          }}
                        >
                          {column}
                        </Checkbox>
                      ))}
                    </div>
                  ),
                }))}
              />
            )}
          </div>
        </div>
      )}
    </Modal>
  );
}
