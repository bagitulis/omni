import {
  CloudDownloadOutlined,
  DeleteOutlined,
  DownOutlined,
  EditOutlined,
  PictureOutlined,
  ShopOutlined,
} from "@ant-design/icons";
import type { MenuProps, TableColumnsType } from "antd";
import { Avatar, Badge, Button, Dropdown, Space, Typography } from "antd";
import { InlineEditCell } from "@/components/shared/InlineEditCell";
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

export type RowActionKey = "edit" | "clone" | "sku_mapping" | "delete";

interface BuildProductColumnsOptions {
  onInlinePriceSave: (skuId: number, price: number) => Promise<void>;
  onInlineStockSave: (
    skuId: number,
    sellerSku: string,
    stock: number,
  ) => Promise<void>;
  onRowAction: (action: RowActionKey, record: UnifiedProductRow) => void;
}

export function buildProductColumns({
  onInlinePriceSave,
  onInlineStockSave,
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
      width: 80,
      render: (_, record) => (
        <Avatar
          shape="square"
          size={48}
          src={toImageSrc(record.images)}
          icon={<PictureOutlined />}
        />
      ),
    },
    name: {
      title: "Name",
      key: "name",
      width: 300,
      render: (_, record) => {
        const variantSummary = buildVariantSummary(record.skus);
        return (
          <Space direction="vertical" size={0}>
            <Typography.Text strong>{record.title}</Typography.Text>
            <Typography.Text type="secondary">
              SKU: {record.primary_sku}
            </Typography.Text>
            {variantSummary ? (
              <Typography.Text type="secondary">
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
      width: 140,
      render: (_, record) => {
        if (record.skus.length === 0) {
          return <Typography.Text type="secondary">-</Typography.Text>;
        }

        if (record.skus.length === 1) {
          const sku = record.skus[0];
          return (
            <div className="inline-edit-cell" data-field="price">
              <InlineEditCell
                value={sku.price}
                mode="price"
                prefix="Rp"
                onSave={(value) => onInlinePriceSave(sku.id, value)}
              />
            </div>
          );
        }

        return (
          <Space direction="vertical" size={0}>
            <Typography.Text>
              {getPriceDisplayText(record.skus)}
            </Typography.Text>
            <Typography.Text type="secondary">
              {record.skus.length} variants
            </Typography.Text>
          </Space>
        );
      },
    },
    stock: {
      title: "Stock",
      key: "stock",
      width: 120,
      render: (_, record) => {
        if (record.skus.length === 0) {
          return <Typography.Text type="secondary">-</Typography.Text>;
        }

        if (record.skus.length === 1) {
          const sku = record.skus[0];
          return (
            <div className="inline-edit-cell" data-field="stock">
              <InlineEditCell
                value={sku.stock}
                mode="stock"
                onSave={(value) =>
                  onInlineStockSave(sku.id, sku.seller_sku, value)
                }
              />
            </div>
          );
        }

        return (
          <Space direction="vertical" size={0}>
            <Typography.Text>
              {idrNumberFormatter.format(getTotalStock(record.skus))}
            </Typography.Text>
            <Typography.Text type="secondary">
              {record.skus.length} variants
            </Typography.Text>
          </Space>
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
