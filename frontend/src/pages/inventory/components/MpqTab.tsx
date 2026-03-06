import {
  Alert,
  Button,
  Checkbox,
  Empty,
  Popconfirm,
  Radio,
  Table,
  Typography,
} from "antd";
import { message } from "@/components/AntStaticHolder";
import { useEffect, useMemo, useState } from "react";
import { getSettings, type WholesaleSettings } from "@/api/wholesale";
import type { BulkPricingItem } from "../utils/bulkPricingItems";
import {
  DEFAULT_SETTINGS,
  executeMpqUpdate,
  getAdjustedPrice,
  MPQ_PREVIEW_COLUMNS,
  type PreviewRow,
  type TierKey,
} from "../utils/mpqTabHelpers";

interface MpqTabProps {
  items: BulkPricingItem[];
}

export function MpqTab({ items }: MpqTabProps) {
  const [settings, setSettings] = useState<WholesaleSettings>(DEFAULT_SETTINGS);
  const [loadingSettings, setLoadingSettings] = useState(false);
  const [processing, setProcessing] = useState(false);
  const [selectedTier, setSelectedTier] = useState<TierKey>("tier1");
  const [enableShopee, setEnableShopee] = useState(true);
  const [enableTiktok, setEnableTiktok] = useState(true);
  const [resultMessage, setResultMessage] = useState<string | null>(null);
  const [resultType, setResultType] = useState<
    "success" | "warning" | "error" | null
  >(null);

  useEffect(() => {
    let cancelled = false;

    async function loadSettings() {
      setLoadingSettings(true);
      try {
        const loaded = await getSettings();
        if (!cancelled && loaded) {
          setSettings(loaded);
        }
      } catch {
        // Settings load error — use defaults and inform user
        if (!cancelled) {
          message.warning("Could not load MPQ settings — using defaults");
        }
      } finally {
        if (!cancelled) {
          setLoadingSettings(false);
        }
      }
    }

    void loadSettings();

    return () => {
      cancelled = true;
    };
  }, []);

  const shopeeItems = useMemo(() => {
    const unique = new Map<string, number>();
    for (const item of items) {
      if (item.platform === "shopee" && !unique.has(item.sku)) {
        unique.set(item.sku, item.price);
      }
    }
    return Array.from(unique.entries()).map(([sku, price]) => ({ sku, price }));
  }, [items]);

  const tiktokItems = useMemo(() => {
    const unique = new Map<string, number>();
    for (const item of items) {
      if (item.platform === "tiktok" && !unique.has(item.sku)) {
        unique.set(item.sku, item.price);
      }
    }
    return Array.from(unique.entries()).map(([sku, price]) => ({ sku, price }));
  }, [items]);

  const selectedMinQty = useMemo(() => {
    if (selectedTier === "normal") return 1;
    if (selectedTier === "tier1") return settings.min_order_1;
    if (selectedTier === "tier2") return settings.max_order_1 + 1;
    return settings.max_order_1 + 3;
  }, [selectedTier, settings.max_order_1, settings.min_order_1]);

  const previewRows = useMemo<PreviewRow[]>(() => {
    const shopeeRows = shopeeItems.slice(0, 3).map((item) => ({
      key: `shopee:${item.sku}`,
      platform: "Shopee",
      sku: item.sku,
      base_price: item.price,
      updated_price: getAdjustedPrice(item.price, settings, selectedTier),
      mpq: selectedMinQty,
    }));

    const tiktokRows = tiktokItems.slice(0, 3).map((item) => ({
      key: `tiktok:${item.sku}`,
      platform: "TikTok",
      sku: item.sku,
      base_price: item.price,
      updated_price: getAdjustedPrice(item.price, settings, selectedTier),
      mpq: selectedMinQty,
    }));

    return [...shopeeRows, ...tiktokRows];
  }, [selectedMinQty, selectedTier, settings, shopeeItems, tiktokItems]);

  const handleMpqUpdate = async () => {
    setProcessing(true);
    setResultMessage(null);
    setResultType(null);

    try {
      const result = await executeMpqUpdate(
        shopeeItems,
        tiktokItems,
        enableShopee,
        enableTiktok,
        settings,
        selectedTier,
        selectedMinQty,
      );
      setResultMessage(result.message);
      setResultType(result.type);
    } catch (error) {
      const rawMessage =
        error instanceof Error ? error.message : "Failed to update MPQ";
      setResultType("error");
      setResultMessage(rawMessage);
    } finally {
      setProcessing(false);
    }
  };

  const isDisabled =
    (!enableShopee || shopeeItems.length === 0) &&
    (!enableTiktok || tiktokItems.length === 0);

  return (
    <div style={{ display: "flex", flexDirection: "column", gap: 16 }}>
      <Alert
        type="info"
        showIcon
        message={`Shopee ${shopeeItems.length} SKU, TikTok ${tiktokItems.length} SKU`}
        description="MPQ update sends calculated price + selected minimum quantity to each platform."
      />

      <div>
        <Typography.Text strong>Platforms</Typography.Text>
        <div style={{ marginTop: 8, display: "flex", gap: 16 }}>
          <Checkbox
            checked={enableShopee}
            onChange={(e) => setEnableShopee(e.target.checked)}
            disabled={shopeeItems.length === 0}
          >
            Shopee ({shopeeItems.length} SKU)
          </Checkbox>
          <Checkbox
            checked={enableTiktok}
            onChange={(e) => setEnableTiktok(e.target.checked)}
            disabled={tiktokItems.length === 0}
          >
            TikTok ({tiktokItems.length} SKU)
          </Checkbox>
        </div>
      </div>

      <div>
        <Typography.Text strong>Select tier target</Typography.Text>
        <Radio.Group
          style={{ display: "block", marginTop: 8 }}
          value={selectedTier}
          onChange={(event) => setSelectedTier(event.target.value as TierKey)}
          optionType="button"
          buttonStyle="solid"
          options={[
            { label: "Normal (1)", value: "normal" },
            { label: `Tier 1 (${settings.min_order_1})`, value: "tier1" },
            { label: `Tier 2 (${settings.max_order_1 + 1})`, value: "tier2" },
            { label: `Tier 3 (${settings.max_order_1 + 3})`, value: "tier3" },
          ]}
        />
      </div>

      {resultMessage && resultType ? (
        <Alert
          type={resultType}
          showIcon
          message={resultMessage}
          closable
          onClose={() => {
            setResultMessage(null);
            setResultType(null);
          }}
        />
      ) : null}

      <Popconfirm
        title="Update MPQ for selected items?"
        description={`Set minimum purchase quantity to ${selectedMinQty} for ${enableShopee ? shopeeItems.length : 0} Shopee + ${enableTiktok ? tiktokItems.length : 0} TikTok items.`}
        onConfirm={handleMpqUpdate}
        okText="Yes, Update"
        disabled={isDisabled}
      >
        <Button type="primary" loading={processing} disabled={isDisabled}>
          Update MPQ
        </Button>
      </Popconfirm>

      <Table
        dataSource={previewRows}
        columns={MPQ_PREVIEW_COLUMNS}
        rowKey="key"
        loading={loadingSettings}
        pagination={false}
        title={() => {
          const shopeeCount = enableShopee ? shopeeItems.length : 0;
          const tiktokCount = enableTiktok ? tiktokItems.length : 0;
          const total = shopeeCount + tiktokCount;
          return `Preview (${previewRows.length} of ${total})`;
        }}
        locale={{
          emptyText: (
            <Empty description="No Shopee/TikTok items with valid price in current selection" />
          ),
        }}
      />
    </div>
  );
}
