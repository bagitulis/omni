import { useEffect, useMemo, useState } from "react";
import {
  Button,
  Card,
  Col,
  Modal,
  Row,
  Select,
  Slider,
  Typography,
  message,
  theme,
} from "antd";
import {
  useInventoryConfig,
  useUpdateInventoryConfig,
} from "@/hooks/useInventory";
import {
  calculateMarketplaceAllocation,
  defaultMarketplaceAllocationSettings,
  deriveMarketplaceAllocationSettings,
  encodeColumns,
  parseColumns,
  saveMarketplaceAllocationSettings,
  type MarketplaceAllocationSettings,
} from "@/pages/inventory/utils/marketplaceAllocation";

const { Text, Title } = Typography;

interface SchemaColumn {
  column_name: string;
  column_type?: string;
}

interface MarketplaceSettingsModalProps {
  open: boolean;
  onClose: () => void;
  schemaColumns: SchemaColumn[];
}

export function MarketplaceSettingsModal({
  open,
  onClose,
  schemaColumns,
}: MarketplaceSettingsModalProps) {
  const {
    token: {
      colorInfo,
      colorText,
      colorTextSecondary,
      colorBgLayout,
      colorSuccess,
    },
  } = theme.useToken();
  const { data: inventoryConfig, isLoading: configLoading } =
    useInventoryConfig();
  const updateConfigMutation = useUpdateInventoryConfig();
  const [settings, setSettings] = useState<MarketplaceAllocationSettings>(
    defaultMarketplaceAllocationSettings,
  );

  useEffect(() => {
    if (!open || configLoading) {
      return;
    }

    setSettings(deriveMarketplaceAllocationSettings(inventoryConfig));
  }, [open, configLoading, inventoryConfig]);

  const selectableColumns = useMemo(
    () =>
      schemaColumns
        .map((column) => column.column_name)
        .filter(
          (column): column is string =>
            typeof column === "string" && column.trim().length > 0,
        ),
    [schemaColumns],
  );

  const total = 20;
  const preview = useMemo(
    () => calculateMarketplaceAllocation(total, false, settings),
    [settings],
  );

  const handleSave = () => {
    if (!settings.totalColumn) {
      message.error("Total column is required");
      return;
    }

    const existingColumns = parseColumns(inventoryConfig?.selected_columns);
    const mergedColumns = encodeColumns(inventoryConfig?.selected_columns, [
      ...existingColumns,
      settings.totalColumn,
      settings.autoColumn,
    ]);

    updateConfigMutation.mutate(
      {
        key_column: settings.totalColumn,
        selected_columns: mergedColumns,
      },
      {
        onSuccess: () => {
          saveMarketplaceAllocationSettings(settings);
          message.success("Settings saved");
          onClose();
        },
        onError: (error) => {
          message.error(
            error instanceof Error ? error.message : "Failed to save settings",
          );
        },
      },
    );
  };

  const handleReset = () => {
    setSettings(defaultMarketplaceAllocationSettings);
  };
  const selectOptions = selectableColumns.map((column) => ({
    label: column,
    value: column,
  }));

  return (
    <Modal
      title="⚙️ Marketplace Allocation Settings"
      open={open}
      onCancel={onClose}
      footer={[
        <Button key="reset" onClick={handleReset}>
          Reset Defaults
        </Button>,
        <Button
          key="save"
          type="primary"
          onClick={handleSave}
          loading={updateConfigMutation.isPending}
        >
          Save Settings
        </Button>,
      ]}
      width={600}
    >
      <div style={{ display: "flex", flexDirection: "column", gap: 24 }}>
        <div>
          <Title level={5}>Column Mapping</Title>
          <Text type="secondary">
            Select which columns to use for Total and Auto values.
          </Text>
          <Row gutter={16} style={{ marginTop: 12 }}>
            <Col span={12}>
              <Text strong>Total Column</Text>
              <Select
                style={{ width: "100%", marginTop: 6 }}
                value={settings.totalColumn || undefined}
                onChange={(value) =>
                  setSettings({ ...settings, totalColumn: value })
                }
                placeholder="Select column"
                options={selectOptions}
                loading={configLoading}
              />
            </Col>
            <Col span={12}>
              <Text strong>Auto Column (Boolean)</Text>
              <Select
                style={{ width: "100%", marginTop: 6 }}
                value={settings.autoColumn || undefined}
                onChange={(value) =>
                  setSettings({ ...settings, autoColumn: value })
                }
                placeholder="Select column"
                options={selectOptions}
                loading={configLoading}
              />
            </Col>
          </Row>
        </div>

        <div>
          <Title level={5}>Allocation Ratios</Title>
          <Text type="secondary">
            Ratios are used for preview and can be fine-tuned later when backend
            fields are available.
          </Text>

          <div style={{ marginTop: 16 }}>
            <Text>
              Shopee Ratio: {(settings.shopeeRatio * 100).toFixed(0)}%
            </Text>
            <Slider
              min={0}
              max={1}
              step={0.05}
              value={settings.shopeeRatio}
              onChange={(value) =>
                setSettings({ ...settings, shopeeRatio: value })
              }
            />
          </div>

          <div style={{ marginTop: 16 }}>
            <Text>
              TikTok Ratio: {(settings.tiktokRatio * 100).toFixed(0)}%
            </Text>
            <Slider
              min={0}
              max={1}
              step={0.05}
              value={settings.tiktokRatio}
              onChange={(value) =>
                setSettings({ ...settings, tiktokRatio: value })
              }
            />
          </div>

          <Card
            size="small"
            style={{
              marginTop: 16,
              background: colorBgLayout,
              borderRadius: 3,
            }}
          >
            <div
              style={{
                display: "flex",
                justifyContent: "space-between",
                fontWeight: 500,
              }}
            >
              <span style={{ color: colorInfo }}>Shopee: {preview.shopee}</span>
              <span style={{ color: colorText }}>TikTok: {preview.tiktok}</span>
              <span style={{ color: colorSuccess }}>
                Lazada: {preview.lazada}
              </span>
              <span style={{ color: colorTextSecondary }}>Total: {total}</span>
            </div>
          </Card>
        </div>

        <div>
          <Title level={5}>Formula Reference</Title>
          <ul>
            <li>
              <strong>Shopee:</strong> MIN(CEILING(ratio * Total), Total)
            </li>
            <li>
              <strong>TikTok:</strong> MIN(CEILING(ratio * Total), Total -
              Shopee)
            </li>
            <li>
              <strong>Lazada:</strong> Total - Shopee - TikTok
            </li>
            <li>
              <strong>Auto = TRUE:</strong> All platforms = Total
            </li>
          </ul>
        </div>
      </div>
    </Modal>
  );
}
