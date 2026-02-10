import {
  Modal,
  Button,
  Spin,
  Checkbox,
  message,
  Typography,
  Space,
} from "antd";
import { useState, useEffect } from "react";

const { Text } = Typography;

interface InventoryColumn {
  name: string;
  type: string;
}

interface InventoryColumnsModalProps {
  open: boolean;
  onClose: () => void;
}

export function InventoryColumnsModal({
  open,
  onClose,
}: InventoryColumnsModalProps) {
  const [loading, setLoading] = useState(false);
  const [saving, setSaving] = useState(false);
  const [availableColumns, setAvailableColumns] = useState<InventoryColumn[]>(
    [],
  );
  const [selectedColumns, setSelectedColumns] = useState<string[]>([]);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const loadColumns = async () => {
      setLoading(true);
      setError(null);
      try {
        // Mock data
        await new Promise((resolve) => setTimeout(resolve, 500));
        setAvailableColumns([
          { name: "Product Name", type: "string" },
          { name: "SKU", type: "string" },
          { name: "Stock", type: "number" },
          { name: "Price", type: "number" },
        ]);
        setSelectedColumns(["Product Name", "SKU", "Stock"]);
      } catch (err) {
        setError("Failed to load columns");
      } finally {
        setLoading(false);
      }
    };

    if (open) {
      loadColumns();
    }
  }, [open]);

  const handleSave = async () => {
    setSaving(true);
    try {
      await new Promise((resolve) => setTimeout(resolve, 500));
      message.success("Columns configuration saved");
      onClose();
    } catch (err) {
      message.error("Failed to save configuration");
    } finally {
      setSaving(false);
    }
  };

  const toggleColumn = (columnName: string) => {
    if (selectedColumns.includes(columnName)) {
      setSelectedColumns(selectedColumns.filter((c) => c !== columnName));
    } else {
      setSelectedColumns([...selectedColumns, columnName]);
    }
  };

  const selectAll = () => {
    setSelectedColumns(availableColumns.map((c) => c.name));
  };

  const clearAll = () => {
    setSelectedColumns([]);
  };

  return (
    <Modal
      title="⚙️ Inventory Column Configuration"
      open={open}
      onCancel={onClose}
      footer={[
        <Button key="cancel" onClick={onClose}>
          Cancel
        </Button>,
        <Button
          key="save"
          type="primary"
          onClick={handleSave}
          loading={saving}
          disabled={loading}
        >
          Save
        </Button>,
      ]}
      width={600}
    >
      {loading ? (
        <div style={{ textAlign: "center", padding: 24 }}>
          <Spin tip="Loading columns..." />
        </div>
      ) : error ? (
        <div style={{ color: "red", textAlign: "center" }}>{error}</div>
      ) : (
        <div style={{ display: "flex", flexDirection: "column", gap: 16 }}>
          <Text type="secondary">
            Select which columns from Google Sheets to display in the inventory
            table. Total: {availableColumns.length}
          </Text>

          {/* Selected Columns */}
          {selectedColumns.length > 0 && (
            <div
              style={{
                background: "#f5f5f5",
                padding: 12,
                borderRadius: 6,
              }}
            >
              <Text strong>Selected ({selectedColumns.length})</Text>
              <div
                style={{
                  display: "flex",
                  flexWrap: "wrap",
                  gap: 8,
                  marginTop: 8,
                }}
              >
                {selectedColumns.map((col) => (
                  <div
                    key={col}
                    style={{
                      background: "white",
                      padding: "4px 8px",
                      borderRadius: 4,
                      border: "1px solid #d9d9d9",
                      fontSize: 12,
                      display: "flex",
                      alignItems: "center",
                      gap: 4,
                    }}
                  >
                    {col}
                    <button
                      type="button"
                      style={{
                        cursor: "pointer",
                        color: "#999",
                        fontWeight: "bold",
                        background: "none",
                        border: "none",
                        padding: 0,
                        marginLeft: 4,
                      }}
                      onClick={() => toggleColumn(col)}
                      aria-label={`Remove ${col}`}
                    >
                      ×
                    </button>
                  </div>
                ))}
              </div>
            </div>
          )}

          {/* Available Columns */}
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
                <Button size="small" onClick={selectAll}>
                  Select All
                </Button>
                <Button size="small" onClick={clearAll}>
                  Clear All
                </Button>
              </Space>
            </div>
            <div
              style={{
                display: "grid",
                gridTemplateColumns: "repeat(auto-fill, minmax(140px, 1fr))",
                gap: 8,
                maxHeight: 300,
                overflowY: "auto",
              }}
            >
              {availableColumns.map((col) => (
                <Checkbox
                  key={col.name}
                  checked={selectedColumns.includes(col.name)}
                  onChange={() => toggleColumn(col.name)}
                >
                  {col.name}
                </Checkbox>
              ))}
            </div>
          </div>
        </div>
      )}
    </Modal>
  );
}
