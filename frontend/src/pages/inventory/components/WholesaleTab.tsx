import { useEffect, useMemo, useState } from "react";
import { Alert, Button, Empty, Popconfirm, Table, message } from "antd";
import {
  batchWholesaleWithReset,
  calculateTiersLocal,
  getSettings,
  type WholesaleSettings,
} from "@/api/wholesale";
import type { BulkPricingItem } from "../utils/bulkPricingItems";

interface WholesaleTabProps {
  items: BulkPricingItem[];
}

interface PreviewRow {
  key: string;
  sku: string;
  price: number;
  tier_1: number;
  tier_2: number;
  tier_3: number;
}

const DEFAULT_SETTINGS: WholesaleSettings = {
  admin_fee: 1500,
  min_order_1: 2,
  max_order_1: 3,
  max_order_tier_3: 1000,
};

function formatCurrency(value: number): string {
  return new Intl.NumberFormat("id-ID").format(value);
}

function readCount(payload: unknown, field: string): number {
  if (typeof payload !== "object" || payload === null) {
    return 0;
  }

  const raw = (payload as Record<string, unknown>)[field];
  return typeof raw === "number" && Number.isFinite(raw) ? raw : 0;
}

export function WholesaleTab({ items }: WholesaleTabProps) {
  const [settings, setSettings] = useState<WholesaleSettings>(DEFAULT_SETTINGS);
  const [loadingSettings, setLoadingSettings] = useState(false);
  const [processing, setProcessing] = useState(false);
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
          message.warning("Could not load wholesale settings — using defaults");
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

  const tiktokCount = useMemo(
    () => items.filter((i) => i.platform === "tiktok").length,
    [items],
  );

  const previewRows = useMemo<PreviewRow[]>(() => {
    return shopeeItems.slice(0, 5).map((item) => {
      const tiers = calculateTiersLocal(item.price, settings);

      return {
        key: item.sku,
        sku: item.sku,
        price: item.price,
        tier_1: tiers[0]?.unit_price ?? item.price,
        tier_2: tiers[1]?.unit_price ?? item.price,
        tier_3: tiers[2]?.unit_price ?? item.price,
      };
    });
  }, [settings, shopeeItems]);

  const columns = [
    {
      title: "SKU",
      dataIndex: "sku",
      key: "sku",
    },
    {
      title: "Base Price",
      dataIndex: "price",
      key: "price",
      render: (value: number) => formatCurrency(value),
    },
    {
      title: `Tier 1 (${settings.min_order_1}-${settings.max_order_1})`,
      dataIndex: "tier_1",
      key: "tier_1",
      render: (value: number) => formatCurrency(value),
    },
    {
      title: `Tier 2 (${settings.max_order_1 + 1}-${settings.max_order_1 + 2})`,
      dataIndex: "tier_2",
      key: "tier_2",
      render: (value: number) => formatCurrency(value),
    },
    {
      title: `Tier 3 (${settings.max_order_1 + 3}-${settings.max_order_tier_3})`,
      dataIndex: "tier_3",
      key: "tier_3",
      render: (value: number) => formatCurrency(value),
    },
  ];

  const handleWholesaleUpdate = async () => {
    if (shopeeItems.length === 0) {
      return;
    }

    setProcessing(true);
    setResultMessage(null);
    setResultType(null);

    try {
      const result = await batchWholesaleWithReset(shopeeItems);
      const processed = readCount(result.data, "processed");
      const failed = readCount(result.data, "failed");
      const summary = `Wholesale update completed: ${processed} processed, ${failed} failed`;

      setResultMessage(summary);
      if (failed > 0) {
        setResultType("warning");
      } else {
        setResultType("success");
      }
    } catch (error) {
      const rawMessage =
        error instanceof Error ? error.message : "Failed to update wholesale";
      setResultType("error");
      setResultMessage(rawMessage);
    } finally {
      setProcessing(false);
    }
  };

  return (
    <div style={{ display: "flex", flexDirection: "column", gap: 16 }}>
      <Alert
        type="info"
        showIcon
        message={`${shopeeItems.length} Shopee SKU ready for wholesale update`}
        description="This action resets MPQ to 1, then applies wholesale tiers using current settings."
      />

      {tiktokCount > 0 ? (
        <Alert
          type="warning"
          showIcon
          message={`${tiktokCount} TikTok items skipped`}
          description="Wholesale tiers are only available for Shopee. TikTok items will not be affected."
        />
      ) : null}

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
        title="Update wholesale tiers?"
        description={`Apply wholesale tiers for ${shopeeItems.length} Shopee items? This will reset MPQ to 1.`}
        onConfirm={handleWholesaleUpdate}
        okText="Yes, Update"
        disabled={shopeeItems.length === 0}
      >
        <Button
          type="primary"
          loading={processing}
          disabled={shopeeItems.length === 0}
        >
          Update Wholesale
        </Button>
      </Popconfirm>

      <Table
        dataSource={previewRows}
        columns={columns}
        rowKey="key"
        loading={loadingSettings}
        pagination={false}
        title={() => `Preview (${Math.min(5, shopeeItems.length)} of ${shopeeItems.length})`}
        locale={{
          emptyText: (
            <Empty description="No Shopee items with valid price in current selection" />
          ),
        }}
      />
    </div>
  );
}
