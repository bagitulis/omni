import {
  CloudDownloadOutlined,
  DeleteOutlined,
  DownOutlined,
  EditOutlined,
  FormOutlined,
  PictureOutlined,
  ShopOutlined,
} from "@ant-design/icons";
import type { MenuProps, TableColumnsType } from "antd";
import { Badge, Button, Dropdown, Space, Typography } from "antd";
import { PlatformStatusCell } from "@/pages/products/components/PlatformStatusCell";
import type { UnifiedProductRow } from "@/types/shared";
import { toImageSrc } from "./unifiedProductUtils";



const idrNumberFormatter = new Intl.NumberFormat("id-ID", {
  maximumFractionDigits: 0,
});

export function formatIdr(value: number): string {
  return `Rp ${idrNumberFormatter.format(value)}`;
}

export function collectVariantNames(skus: UnifiedProductRow["skus"]): string[] {
  const names = skus
    .map((sku) => sku.variant_name.trim())
    .filter((name) => name !== "");

  return Array.from(new Set(names));
}

export function buildVariantSummary(skus: UnifiedProductRow["skus"]): string {
  const variantNames = collectVariantNames(skus);
  if (variantNames.length === 0) {
    return "";
  }

  const visible = variantNames.slice(0, 2).join(", ");
  const hiddenCount = variantNames.length - 2;
  if (hiddenCount <= 0) {
    return visible;
  }

  return `${visible} +${hiddenCount}`;
}

export function getPriceDisplayText(skus: UnifiedProductRow["skus"]): string {
  if (skus.length === 0) {
    return "-";
  }

  const prices = skus.map((sku) => sku.price);
  const minPrice = Math.min(...prices);
  const maxPrice = Math.max(...prices);

  if (minPrice === maxPrice) {
    return formatIdr(minPrice);
  }

  return `${formatIdr(minPrice)} - ${formatIdr(maxPrice)}`;
}

export function getTotalStock(skus: UnifiedProductRow["skus"]): number {
  return skus.reduce((sum, sku) => sum + sku.stock, 0);
}

/** Get the first SKU's platform price for a given platform. Returns 0 if not linked. */
export function getPlatformPrice(
  skus: UnifiedProductRow["skus"],
  platform: string,
): number {
  for (const sku of skus) {
    const pp = sku.platform_prices?.find((p) => p.platform === platform);
    if (pp && pp.platform_price > 0) return pp.platform_price;
  }
  return 0;
}

/** Get the first SKU's inventory reference price. Returns 0 if not in inventory. */
export function getInventoryPrice(skus: UnifiedProductRow["skus"]): number {
  for (const sku of skus) {
    if (sku.inventory_price && sku.inventory_price > 0) return sku.inventory_price;
  }
  return 0;
}

export function getSkuCountLabel(count: number): string {
  return `${count} SKU`;
}

export function getDisplayTitle(title: string, fallbackSku: string): string {
  const trimmed = title.trim();
  if (trimmed) {
    return trimmed;
  }

  return fallbackSku.trim() || "Unnamed Product";
}

export type RowActionKey =
  | "edit"
  | "clone"
  | "sku_mapping"
  | "update_price"
  | "update_stock"
  | "delete";

interface BuildProductColumnsOptions {
  onRowAction: (action: RowActionKey, record: UnifiedProductRow) => void;
}

export function buildProductColumns({
  onRowAction,
}: BuildProductColumnsOptions): Record<
  string,
  TableColumnsType<UnifiedProductRow>[number]
> {
  return {
    image: {
      title: "Image",
      dataIndex: "images",
      key: "image",
      width: 160,
      render: (_, record) => {
        const imageSrc = toImageSrc(record.images);

        return (
          <div
            style={{
              width: 120,
              height: 120,
              borderRadius: 4,
              border: "1px solid #e5e7eb",
              background: "#f8fafc",
              display: "flex",
              alignItems: "center",
              justifyContent: "center",
              overflow: "hidden",
            }}
          >
            {imageSrc ? (
              <img
                src={imageSrc}
                alt={record.title || record.primary_sku}
                style={{
                  width: "100%",
                  height: "100%",
                  objectFit: "contain",
                }}
              />
            ) : (
              <PictureOutlined style={{ color: "#94a3b8", fontSize: 18 }} />
            )}
          </div>
        );
      },
    },
    name: {
      title: "Name",
      key: "name",
      width: 300,
      render: (_, record) => {
        const variantSummary = buildVariantSummary(record.skus);
        const displayTitle = getDisplayTitle(record.title, record.primary_sku);

        return (
          <Space direction="vertical" size={2} style={{ maxWidth: 300 }}>
            <Typography.Text strong>{displayTitle}</Typography.Text>
            <Typography.Text type="secondary">
              SKU: {record.primary_sku}
            </Typography.Text>
            {variantSummary ? (
              <Typography.Text
                type="secondary"
                ellipsis={{ tooltip: `Variant: ${variantSummary}` }}
              >
                Variant: {variantSummary}
              </Typography.Text>
            ) : null}
          </Space>
        );
      },
    },
    price: {
      title: "Price",
      key: "price",
      width: 110,
      render: (_, record) => {
        const invPrice = getInventoryPrice(record.skus);
        if (invPrice > 0) {
          return <Typography.Text>{formatIdr(invPrice)}</Typography.Text>;
        }
        // Fallback to master price
        const masterPrice = record.primary_price;
        return (
          <Typography.Text type={masterPrice > 0 ? undefined : "secondary"}>
            {masterPrice > 0 ? formatIdr(masterPrice) : "—"}
          </Typography.Text>
        );
      },
    },

    stock: {
      title: "Stock",
      key: "stock",
      width: 100,
      render: (_, record) => {
        return (
          <Typography.Text>
            {idrNumberFormatter.format(getTotalStock(record.skus))}
          </Typography.Text>
        );
      },
    },
    platforms: {
      title: "Platforms",
      key: "platforms",
      width: 180,
      render: (_, record) => <PlatformStatusCell product={record} />,
    },
    category: {
      title: "Category",
      key: "category",
      width: 140,
      render: (_, record) => (
        <Typography.Text>
          {record.description ? "General" : "-"}
        </Typography.Text>
      ),
    },
    status: {
      title: "Status",
      key: "status",
      width: 120,
      render: (_, record) => (
        <Badge
          status={
            record.status === "active"
              ? "success"
              : record.status === "draft"
                ? "processing"
                : "default"
          }
          text={record.status}
        />
      ),
    },
    actions: {
      title: "Actions",
      key: "actions",
      width: 120,
      render: (_, record) => {
        const items: MenuProps["items"] = [
          { key: "edit", icon: <EditOutlined />, label: "Edit" },
          { key: "clone", icon: <CloudDownloadOutlined />, label: "Clone" },
          {
            key: "sku_mapping",
            icon: <ShopOutlined />,
            label: "SKU Mapping",
          },
          {
            key: "update_stock",
            icon: <FormOutlined />,
            label: "Update Stock",
          },
          {
            key: "update_price",
            icon: <FormOutlined />,
            label: "Update Price",
          },
          { type: "divider" },
          {
            key: "delete",
            icon: <DeleteOutlined />,
            label: "Delete",
            danger: true,
          },
        ];

        return (
          <Dropdown
            menu={{
              items,
              onClick: ({ key }) => onRowAction(key as RowActionKey, record),
            }}
            trigger={["click"]}
          >
            <Button size="small">
              Actions <DownOutlined />
            </Button>
          </Dropdown>
        );
      },
    },
  };
}
