import {
  Alert,
  Button,
  Checkbox,
  Empty,
  message,
  Radio,
  Table,
  Typography,
} from "antd";
import { useEffect, useMemo, useState } from "react";
import {
  batchShopeeMpq,
  getSettings,
  type WholesaleSettings,
} from "@/api/wholesale";
import type { BulkPricingItem } from "../utils/bulkPricingItems";
import {
  DEFAULT_SETTINGS,
  formatCurrency,
  getAdjustedPrice,
  type PreviewRow,
  readCount,
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
  const [enableTiktok, setEnableTiktok] = useState(false); // TikTok MPQ not yet supported
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
      } catch (error) {
        if (!cancelled) {
          const rawMessage =
            error instanceof Error
              ? error.message
              : "Failed to load wholesale settings";
          message.error(rawMessage);
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
      if (item.platform !== "shopee") {
        continue;
      }
      if (!unique.has(item.sku)) {
        unique.set(item.sku, item.price);
      }
    }
    return Array.from(unique.entries()).map(([sku, price]) => ({ sku, price }));
  }, [items]);

  const tiktokItems = useMemo(() => {
    const unique = new Map<string, number>();
    for (const item of items) {
      if (item.platform !== "tiktok") {
        continue;
      }
      if (!unique.has(item.sku)) {
        unique.set(item.sku, item.price);
      }
    }
    return Array.from(unique.entries()).map(([sku, price]) => ({ sku, price }));
  }, [items]);

  const selectedMinQty = useMemo(() => {
    if (selectedTier === "normal") {
      return 1;
    }

    if (selectedTier === "tier1") {
      return settings.min_order_1;
    }

    if (selectedTier === "tier2") {
      return settings.max_order_1 + 1;
    }

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
    const shopeeActive = enableShopee && shopeeItems.length > 0;
    const tiktokActive = enableTiktok && tiktokItems.length > 0;

    if (!shopeeActive && !tiktokActive) {
      message.warning(
        "No platform selected or no items available for selected platforms",
      );
      return;
    }

    setProcessing(true);
    setResultMessage(null);
    setResultType(null);

    let totalFailed = 0;
    const summaries: string[] = [];

    try {
      if (shopeeActive) {
        const shopeePayload = shopeeItems.map((item) => ({
          sku: item.sku,
          price: getAdjustedPrice(item.price, settings, selectedTier),
        }));

        const shopeeResult = await batchShopeeMpq(
          shopeePayload,
          selectedMinQty,
        );
        const processed = readCount(shopeeResult.data, "processed");
        const failed = readCount(shopeeResult.data, "failed");
        totalFailed += failed;
        summaries.push(`Shopee ${processed} processed, ${failed} failed`);
      }

      // TikTok MPQ: will be enabled when backend TikTok API integration is ready

      const summary = `MPQ update (min_qty=${selectedMinQty}): ${summaries.join(" | ")}`;
      setResultMessage(summary);

      if (totalFailed > 0) {
        setResultType("warning");
        message.warning(summary);
      } else {
        setResultType("success");
        message.success(summary);
      }
    } catch (error) {
      const rawMessage =
        error instanceof Error ? error.message : "Failed to update MPQ";
      setResultType("error");
      setResultMessage(rawMessage);
      message.error(rawMessage);
    } finally {
      setProcessing(false);
    }
  };

  const columns = [
    {
      title: "Platform",
      dataIndex: "platform",
      key: "platform",
    },
    {
      title: "SKU",
      dataIndex: "sku",
      key: "sku",
    },
    {
      title: "Base Price",
      dataIndex: "base_price",
      key: "base_price",
      render: (value: number) => formatCurrency(value),
    },
    {
      title: "Updated Price",
      dataIndex: "updated_price",
      key: "updated_price",
      render: (value: number) => formatCurrency(value),
    },
    {
      title: "MPQ",
      dataIndex: "mpq",
      key: "mpq",
    },
  ];

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
            disabled
          >
            TikTok ({tiktokItems.length} SKU) — Coming soon
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
            {
              label: `Tier 1 (${settings.min_order_1})`,
              value: "tier1",
            },
            {
              label: `Tier 2 (${settings.max_order_1 + 1})`,
              value: "tier2",
            },
            {
              label: `Tier 3 (${settings.max_order_1 + 3})`,
              value: "tier3",
            },
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

      <Button
        type="primary"
        loading={processing}
        onClick={handleMpqUpdate}
        disabled={
          (!enableShopee || shopeeItems.length === 0) &&
          (!enableTiktok || tiktokItems.length === 0)
        }
      >
        Update MPQ
      </Button>

      <Table
        dataSource={previewRows}
        columns={columns}
        rowKey="key"
        loading={loadingSettings}
        pagination={false}
        locale={{
          emptyText: (
            <Empty description="No Shopee/TikTok items with valid price in current selection" />
          ),
        }}
      />
    </div>
  );
}
