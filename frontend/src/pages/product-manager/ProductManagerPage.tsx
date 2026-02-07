import { useEffect, useMemo, useRef, useState } from "react";
import { useNavigate, useParams } from "react-router-dom";
import {
  Button,
  Card,
  Flex,
  Switch,
  Table,
  Tabs,
  Tag,
  Typography,
  message,
} from "antd";
import type { ColumnsType } from "antd/es/table";
import { SyncOutlined } from "@ant-design/icons";
import type {
  DbProductRow,
  ProductManagerPlatform,
} from "@/types/product_manager";
import { useDbProducts } from "@/hooks/useProductManager";
import { syncPlatformProducts } from "@/api/productManager";
import { useMutation } from "@tanstack/react-query";

const { Title, Text } = Typography;

const PLATFORMS: Array<{ key: ProductManagerPlatform; label: string }> = [
  { key: "shopee", label: "Shopee" },
  { key: "lazada", label: "Lazada" },
  { key: "tiktok", label: "TikTok" },
];

function coercePlatform(input: string | undefined): ProductManagerPlatform {
  if (input === "lazada" || input === "tiktok" || input === "shopee") {
    return input;
  }
  return "shopee";
}

function getRowKey(
  platform: ProductManagerPlatform,
  row: DbProductRow,
): string {
  const itemId = typeof row.item_id === "string" ? row.item_id : undefined;
  const modelId = typeof row.model_id === "string" ? row.model_id : undefined;
  const skuId = typeof row.sku_id === "string" ? row.sku_id : undefined;
  const productId =
    typeof row.product_id === "string" ? row.product_id : undefined;

  if (platform === "shopee") {
    return modelId
      ? `${itemId ?? ""}-${modelId}`
      : (itemId ?? JSON.stringify(row));
  }

  if (platform === "lazada") {
    return skuId ? `${itemId ?? ""}-${skuId}` : (itemId ?? JSON.stringify(row));
  }

  return skuId
    ? `${productId ?? ""}-${skuId}`
    : (productId ?? JSON.stringify(row));
}

function formatImageSrc(value: unknown): string | null {
  if (typeof value === "string" && value.trim() !== "") {
    return value;
  }

  if (Array.isArray(value) && value.length > 0) {
    const first = value[0];
    if (typeof first === "string" && first.trim() !== "") {
      return first;
    }
  }

  return null;
}

function getProductImage(row: DbProductRow): string | null {
  const fromLocal = formatImageSrc(row.local_images);
  if (fromLocal) return fromLocal;

  const fromImage = formatImageSrc(row.image);
  if (fromImage) return fromImage;

  return null;
}

function buildColumns(
  platform: ProductManagerPlatform,
): ColumnsType<DbProductRow> {
  const base: ColumnsType<DbProductRow> = [
    {
      title: "Image",
      key: "image",
      width: 64,
      render: (_, row) => {
        const image = getProductImage(row);
        if (!image) return <span style={{ color: "#94a3b8" }}>—</span>;
        return (
          <img
            src={image}
            alt="product"
            width={40}
            height={40}
            style={{
              objectFit: "cover",
              borderRadius: 4,
              border: "1px solid #e2e8f0",
            }}
          />
        );
      },
    },
    {
      title: "Product",
      key: "product",
      render: (_, row) => {
        const name = typeof row.item_name === "string" ? row.item_name : "";
        const sku =
          typeof row.sku === "string"
            ? row.sku
            : typeof row.sku_name === "string"
              ? row.sku_name
              : typeof row.seller_sku === "string"
                ? row.seller_sku
                : typeof row.shop_sku === "string"
                  ? row.shop_sku
                  : "";

        return (
          <div>
            <div style={{ fontWeight: 600, color: "#0f172a" }}>
              {name || "(Unnamed)"}
            </div>
            <div style={{ fontSize: 12, color: "#64748b" }}>{sku}</div>
          </div>
        );
      },
    },
    {
      title: "Price",
      key: "price",
      width: 120,
      align: "right",
      render: (_, row) => {
        const price =
          typeof row.price === "number"
            ? row.price
            : typeof row.special_price === "number"
              ? row.special_price
              : null;
        if (price == null) return "—";
        return `Rp ${Math.round(price).toLocaleString("id-ID")}`;
      },
    },
    {
      title: "Stock",
      key: "stock",
      width: 100,
      align: "right",
      render: (_, row) => {
        const stock =
          typeof row.stock === "number"
            ? row.stock
            : typeof row.quantity === "number"
              ? row.quantity
              : typeof row.available === "number"
                ? row.available
                : null;
        return stock == null ? "—" : stock.toLocaleString("id-ID");
      },
    },
    {
      title: "Status",
      key: "status",
      width: 120,
      render: (_, row) => {
        const status = typeof row.status === "string" ? row.status : "";
        if (!status) return <Tag>unknown</Tag>;
        const normalized = status.toLowerCase();
        const color =
          normalized === "active" || normalized === "normal"
            ? "green"
            : normalized.includes("delete") || normalized.includes("inactive")
              ? "red"
              : "default";
        return <Tag color={color}>{status}</Tag>;
      },
    },
    {
      title: "Updated",
      dataIndex: "updated_at",
      key: "updated_at",
      width: 180,
      render: (value) => {
        if (typeof value !== "string") return "—";
        return value.replace("T", " ").replace("Z", "");
      },
    },
  ];

  if (platform === "shopee") {
    return [
      { title: "Item ID", dataIndex: "item_id", key: "item_id", width: 120 },
      { title: "Model ID", dataIndex: "model_id", key: "model_id", width: 120 },
      ...base,
    ];
  }

  if (platform === "lazada") {
    return [
      { title: "Item ID", dataIndex: "item_id", key: "item_id", width: 120 },
      { title: "SKU ID", dataIndex: "sku_id", key: "sku_id", width: 140 },
      ...base,
    ];
  }

  return [
    {
      title: "Product ID",
      dataIndex: "product_id",
      key: "product_id",
      width: 140,
    },
    { title: "SKU ID", dataIndex: "sku_id", key: "sku_id", width: 140 },
    ...base,
  ];
}

export function ProductManagerPage() {
  const navigate = useNavigate();
  const { platform: platformParam } = useParams<{ platform?: string }>();

  const [activeTab, setActiveTab] = useState<ProductManagerPlatform>(
    coercePlatform(platformParam),
  );
  const [autoRefresh, setAutoRefresh] = useState(true);
  const [autoSync, setAutoSync] = useState(true);
  const [page, setPage] = useState(1);
  const [pageSize, setPageSize] = useState(50);
  const lastAutoSyncedPlatformRef = useRef<ProductManagerPlatform | null>(null);

  useEffect(() => {
    setActiveTab(coercePlatform(platformParam));
  }, [platformParam]);

  const offset = (page - 1) * pageSize;

  const { data, isLoading, refetch } = useDbProducts(
    activeTab,
    { offset, limit: pageSize },
    { autoRefresh, refetchInterval: 15_000 },
  );

  const columns = useMemo(() => buildColumns(activeTab), [activeTab]);

  const syncMutation = useMutation({
    mutationFn: () => syncPlatformProducts(activeTab),
    onSuccess: (result) => {
      const synced = result.data?.synced;
      message.success(
        typeof synced === "number"
          ? `Synced ${synced} products from ${activeTab}`
          : `Products synced from ${activeTab}`,
      );
      void refetch();
    },
    onError: (error: Error) => {
      message.error(error.message || "Failed to sync products");
    },
  });

  useEffect(() => {
    if (!autoSync) return;
    if (syncMutation.isPending) return;
    if (lastAutoSyncedPlatformRef.current === activeTab) return;

    lastAutoSyncedPlatformRef.current = activeTab;
    void syncMutation.mutateAsync();
  }, [activeTab, autoSync, syncMutation]);

  const handleTabChange = (key: string) => {
    const next = coercePlatform(key);
    setActiveTab(next);
    setPage(1);
    navigate(`/product-manager/${next}`);
  };

  const tabItems = PLATFORMS.map((p) => ({
    key: p.key,
    label: p.label,
  }));

  return (
    <div style={{ padding: 24 }}>
      <Flex vertical gap={16}>
        <Flex justify="space-between" align="center">
          <div>
            <Title level={3} style={{ margin: 0 }}>
              Product Manager
            </Title>
            <Text type="secondary" style={{ fontSize: 12 }}>
              Platform products synced into database
            </Text>
          </div>
          <Flex align="center" gap={12}>
            <Flex align="center" gap={6}>
              <Text style={{ fontSize: 12 }}>Auto refresh</Text>
              <Switch
                checked={autoRefresh}
                onChange={setAutoRefresh}
                size="small"
              />
            </Flex>
            <Flex align="center" gap={6}>
              <Text style={{ fontSize: 12 }}>Auto sync</Text>
              <Switch checked={autoSync} onChange={setAutoSync} size="small" />
            </Flex>
            <Button
              data-testid="product-manager-sync"
              type="primary"
              icon={<SyncOutlined />}
              loading={syncMutation.isPending}
              onClick={() => syncMutation.mutate()}
            >
              Sync Now
            </Button>
          </Flex>
        </Flex>

        <Card style={{ borderRadius: 4 }}>
          <Tabs
            activeKey={activeTab}
            onChange={handleTabChange}
            items={tabItems}
          />
          <Table<DbProductRow>
            data-testid="product-manager-table"
            style={{ marginTop: 12 }}
            loading={isLoading}
            dataSource={data?.products ?? []}
            columns={columns}
            rowKey={(row) => getRowKey(activeTab, row)}
            scroll={{ x: 1200 }}
            pagination={{
              current: page,
              pageSize,
              total: data?.total ?? 0,
              onChange: (p, ps) => {
                setPage(p);
                setPageSize(ps);
              },
              showSizeChanger: true,
            }}
          />
        </Card>
      </Flex>
    </div>
  );
}
