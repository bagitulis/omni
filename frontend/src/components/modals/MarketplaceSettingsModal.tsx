import {
  Modal,
  Button,
  Select,
  Slider,
  Typography,
  Row,
  Col,
  Card,
  message,
} from "antd";
import { useState, useEffect, useMemo } from "react";

const { Text, Title } = Typography;
const { Option } = Select;

interface MarketplaceSettings {
  totalColumn: string;
  autoColumn: string;
  shopeeRatio: number;
  tiktokRatio: number;
}

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
  const [settings, setSettings] = useState<MarketplaceSettings>({
    totalColumn: "",
    autoColumn: "",
    shopeeRatio: 0.6,
    tiktokRatio: 0.3,
  });

  useEffect(() => {
    if (open) {
      // Mock load settings
      // In real app, load from API/Storage
      const loadedSettings = localStorage.getItem("marketplace_settings");
      if (loadedSettings) {
        setSettings(JSON.parse(loadedSettings));
      }
    }
  }, [open]);

  const handleSave = () => {
    // Mock save settings
    localStorage.setItem("marketplace_settings", JSON.stringify(settings));
    message.success("Settings saved");
    onClose();
  };

  const handleReset = () => {
    setSettings({
      totalColumn: "",
      autoColumn: "",
      shopeeRatio: 0.6,
      tiktokRatio: 0.3,
    });
  };

  // Filter columns logic (simplified from Vue)
  const numericColumns = useMemo(() => {
    return schemaColumns.map((c) => c.column_name);
    // In real implementation, filter by type if available
  }, [schemaColumns]);

  const booleanColumns = useMemo(() => {
    return schemaColumns.map((c) => c.column_name);
  }, [schemaColumns]);

  // Preview calculation
  const total = 20;
  const previewShopee = Math.min(
    Math.ceil(settings.shopeeRatio * total),
    total,
  );
  const remainingAfterShopee = Math.max(0, total - previewShopee);
  const previewTiktok = Math.min(
    Math.ceil(settings.tiktokRatio * total),
    remainingAfterShopee,
  );
  const previewLazada = Math.max(0, total - previewShopee - previewTiktok);

  return (
    <Modal
      title="⚙️ Marketplace Allocation Settings"
      open={open}
      onCancel={onClose}
      footer={[
        <Button key="reset" onClick={handleReset}>
          Reset Defaults
        </Button>,
        <Button key="save" type="primary" onClick={handleSave}>
          Save Settings
        </Button>,
      ]}
      width={600}
    >
      <div style={{ display: "flex", flexDirection: "column", gap: 24 }}>
        {/* Column Mapping */}
        <div>
          <Title level={5}>📊 Column Mapping</Title>
          <Text type="secondary">
            Select which columns to use for Total and Auto values
          </Text>
          <Row gutter={16} style={{ marginTop: 12 }}>
            <Col span={12}>
              <Text strong>Total Column</Text>
              <Select
                style={{ width: "100%" }}
                value={settings.totalColumn}
                onChange={(val) =>
                  setSettings({ ...settings, totalColumn: val })
                }
                placeholder="Select Column"
              >
                {numericColumns.map((col) => (
                  <Option key={col} value={col}>
                    {col}
                  </Option>
                ))}
              </Select>
            </Col>
            <Col span={12}>
              <Text strong>Auto Column (Boolean)</Text>
              <Select
                style={{ width: "100%" }}
                value={settings.autoColumn}
                onChange={(val) =>
                  setSettings({ ...settings, autoColumn: val })
                }
                placeholder="Select Column"
              >
                {booleanColumns.map((col) => (
                  <Option key={col} value={col}>
                    {col}
                  </Option>
                ))}
              </Select>
            </Col>
          </Row>
        </div>

        {/* Allocation Ratios */}
        <div>
          <Title level={5}>📈 Allocation Ratios</Title>
          <Text type="secondary">
            Set the percentage ratio for each marketplace
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
              onChange={(val) => setSettings({ ...settings, shopeeRatio: val })}
            />
          </div>

          <div style={{ marginTop: 16 }}>
            <Text>
              Tiktok Ratio: {(settings.tiktokRatio * 100).toFixed(0)}%
            </Text>
            <Slider
              min={0}
              max={1}
              step={0.05}
              value={settings.tiktokRatio}
              onChange={(val) => setSettings({ ...settings, tiktokRatio: val })}
            />
          </div>

          <Card size="small" style={{ marginTop: 16, background: "#f5f5f5" }}>
            <div
              style={{
                display: "flex",
                justifyContent: "space-between",
                fontWeight: 500,
              }}
            >
              <span style={{ color: "#ee4d2d" }}>Shopee: {previewShopee}</span>
              <span style={{ color: "#000000" }}>Tiktok: {previewTiktok}</span>
              <span style={{ color: "#0f146d" }}>Lazada: {previewLazada}</span>
              <span>Total: {total}</span>
            </div>
          </Card>
        </div>

        {/* Formula Reference */}
        <div>
          <Title level={5}>📝 Formula Reference</Title>
          <ul>
            <li>
              <strong>Shopee:</strong> MIN(CEILING(ratio × Total), Total)
            </li>
            <li>
              <strong>Tiktok:</strong> MIN(CEILING(ratio × Total), Total -
              Shopee)
            </li>
            <li>
              <strong>Lazada:</strong> Total - Shopee - Tiktok
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
