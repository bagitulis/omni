import { useEffect, useMemo, useState } from "react";
import { Button, Col, Modal, Row, Select, Typography } from "antd";
import {
  useInventoryConfig,
  useUpdateInventoryConfig,
} from "@/hooks/useInventory";
import {
  defaultMarketplaceAllocationSettings,
  deriveMarketplaceAllocationSettings,
  encodeColumns,
  parseColumns,
  saveMarketplaceAllocationSettings,
  type MarketplaceAllocationSettings,
} from "@/pages/inventory/utils/marketplaceAllocation";
import { AllocationRatioSection } from "./AllocationRatioSection";
import { message } from "@/components/AntStaticApi";

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

  const handleSave = () => {
    if (!settings.keyColumn) {
      message.error("Key column is required");
      return;
    }

    const existingColumns = parseColumns(inventoryConfig?.selected_columns);
    const mergedColumns = encodeColumns(inventoryConfig?.selected_columns, [
      ...existingColumns,
      ...(settings.totalColumn ? [settings.totalColumn] : []),
      ...(settings.autoColumn ? [settings.autoColumn] : []),
    ]);

    updateConfigMutation.mutate(
      {
        key_column: settings.keyColumn,
        total_column: settings.totalColumn,
        raw_total_column: settings.rawTotalColumn,
        auto_column: settings.autoColumn,
        selected_columns: mergedColumns,
        price_column: settings.priceColumn || undefined,
        price_column_shopee: settings.priceColumnShopee || undefined,
        price_column_tiktok: settings.priceColumnTiktok || undefined,
        price_column_lazada: settings.priceColumnLazada || undefined,
        shopee_ratio: settings.shopeeRatio,
        tiktok_ratio: settings.tiktokRatio,
      },
      {
        onSuccess: () => {
          // Keep localStorage as fallback for backward compat during migration
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
          <Title level={5}>Key Column</Title>
          <Text type="secondary">
            The column used as the unique identifier (row key) in the inventory
            table. This is typically the SKU column.
          </Text>
          <Select
            style={{ width: "100%", marginTop: 8 }}
            value={settings.keyColumn || undefined}
            onChange={(value) =>
              setSettings({ ...settings, keyColumn: value })
            }
            placeholder="Select key column (e.g. SKU)"
            options={selectOptions}
            loading={configLoading}
          />
        </div>

        <div>
          <Title level={5}>Column Mapping</Title>
          <Text type="secondary">
            Configure which columns from your inventory sheet map to stock,
            allocation, and price values.
          </Text>

          <Text strong style={{ display: "block", marginTop: 16, marginBottom: 4 }}>Stock & Allocation</Text>
          <Row gutter={16}>
            <Col span={12}>
              <Text type="secondary" style={{ fontSize: 12 }}>
                Stock Total (e.g. "TOTAL")
              </Text>
              <Select
                style={{ width: "100%", marginTop: 4 }}
                value={settings.rawTotalColumn || undefined}
                onChange={(value) =>
                  setSettings({ ...settings, rawTotalColumn: value })
                }
                placeholder="Select column"
                options={selectOptions}
                loading={configLoading}
                allowClear
              />
            </Col>
            <Col span={12}>
              <Text type="secondary" style={{ fontSize: 12 }}>
                Allocation Source (e.g. "Sellable")
              </Text>
              <Select
                style={{ width: "100%", marginTop: 4 }}
                value={settings.totalColumn || undefined}
                onChange={(value) =>
                  setSettings({ ...settings, totalColumn: value })
                }
                placeholder="Select column"
                options={selectOptions}
                loading={configLoading}
                allowClear
              />
            </Col>
          </Row>
          <Row gutter={16} style={{ marginTop: 8 }}>
            <Col span={12}>
              <Text type="secondary" style={{ fontSize: 12 }}>
                Auto Mode Column (boolean)
              </Text>
              <Select
                style={{ width: "100%", marginTop: 4 }}
                value={settings.autoColumn || undefined}
                onChange={(value) =>
                  setSettings({ ...settings, autoColumn: value })
                }
                placeholder="Select column"
                options={selectOptions}
                loading={configLoading}
                allowClear
              />
            </Col>
          </Row>

          <Text strong style={{ display: "block", marginTop: 16, marginBottom: 4 }}>Price Columns</Text>
          <Text type="secondary" style={{ fontSize: 12, display: "block", marginBottom: 8 }}>
            Used for price recommendations in Sync Price modal. Select which columns contain per-platform prices.
          </Text>
          <Row gutter={16}>
            <Col span={12}>
              <Text type="secondary" style={{ fontSize: 12 }}>
                Base Price
              </Text>
              <Select
                style={{ width: "100%", marginTop: 4 }}
                value={settings.priceColumn || undefined}
                onChange={(value) =>
                  setSettings({ ...settings, priceColumn: value })
                }
                placeholder="e.g. HARGA"
                options={selectOptions}
                loading={configLoading}
                allowClear
              />
            </Col>
            <Col span={12}>
              <Text type="secondary" style={{ fontSize: 12 }}>
                Shopee Price
              </Text>
              <Select
                style={{ width: "100%", marginTop: 4 }}
                value={settings.priceColumnShopee || undefined}
                onChange={(value) =>
                  setSettings({ ...settings, priceColumnShopee: value })
                }
                placeholder="e.g. HARGA_SHOPEE"
                options={selectOptions}
                loading={configLoading}
                allowClear
              />
            </Col>
          </Row>
          <Row gutter={16} style={{ marginTop: 8 }}>
            <Col span={12}>
              <Text type="secondary" style={{ fontSize: 12 }}>
                TikTok Price
              </Text>
              <Select
                style={{ width: "100%", marginTop: 4 }}
                value={settings.priceColumnTiktok || undefined}
                onChange={(value) =>
                  setSettings({ ...settings, priceColumnTiktok: value })
                }
                placeholder="e.g. HARGA_TIKTOK"
                options={selectOptions}
                loading={configLoading}
                allowClear
              />
            </Col>
            <Col span={12}>
              <Text type="secondary" style={{ fontSize: 12 }}>
                Lazada Price
              </Text>
              <Select
                style={{ width: "100%", marginTop: 4 }}
                value={settings.priceColumnLazada || undefined}
                onChange={(value) =>
                  setSettings({ ...settings, priceColumnLazada: value })
                }
                placeholder="e.g. HARGA_LAZADA"
                options={selectOptions}
                loading={configLoading}
                allowClear
              />
            </Col>
          </Row>
        </div>

        <AllocationRatioSection
          settings={settings}
          onChangeShopeeRatio={(value) =>
            setSettings({ ...settings, shopeeRatio: value })
          }
          onChangeTiktokRatio={(value) =>
            setSettings({ ...settings, tiktokRatio: value })
          }
        />
      </div>
    </Modal>
  );
}
