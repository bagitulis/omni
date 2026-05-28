import { Alert, Descriptions, Drawer, Empty, Flex, Grid, Space, Spin, Table, Typography, theme } from "antd";
import type { ColumnsType } from "antd/es/table";
import { useOrderItems, useSkuOrders } from "@/hooks/useAnalytics";
import { formatCurrency, formatDate } from "@/lib/analyticsHelpers";
import type {
  ReportPlatform,
  ShopeeOrderItem,
  ShopeeShippingOrder,
  ShopeeSkuOrder,
  SkuGroup,
  TiktokOrderItem,
  TiktokShippingOrder,
  TiktokSkuGroup,
  TiktokSkuOrder,
} from "@/types/analytics";

type ReconciliationRow = SkuGroup | TiktokSkuGroup;
type ShippingRow = ShopeeShippingOrder | TiktokShippingOrder;
type SkuOrderRow = ShopeeSkuOrder | TiktokSkuOrder;
type OrderItemRow = ShopeeOrderItem | TiktokOrderItem;

interface PeriodProps {
  platform: ReportPlatform;
  month: number;
  year: number;
}

interface ReconciliationDrilldownDrawerProps extends PeriodProps {
  row: ReconciliationRow | null;
  open: boolean;
  onClose: () => void;
}

interface ShippingFeeDrilldownDrawerProps extends PeriodProps {
  row: ShippingRow | null;
  open: boolean;
  onClose: () => void;
}

function isShopeeSkuOrder(row: SkuOrderRow): row is ShopeeSkuOrder {
  return "escrow_amount" in row;
}

function isShopeeOrderItem(row: OrderItemRow): row is ShopeeOrderItem {
  return "item_name" in row;
}

function isShopeeShippingRow(row: ShippingRow): row is ShopeeShippingOrder {
  return "buyer_paid" in row;
}

function productName(row: ReconciliationRow) {
  return "item_name" in row ? row.item_name : row.product_name;
}

function skuLabel(row: ReconciliationRow) {
  return "model_sku" in row ? row.model_sku : row.seller_sku;
}

function valuesLabel(values: number[]) {
  return values.length === 0 ? "-" : values.map((value) => formatCurrency(value)).join(" / ");
}

function safeNumber(value: number | null | undefined) {
  return typeof value === "number" && Number.isFinite(value) ? value : 0;
}

function drawerWidth(screens: ReturnType<typeof Grid.useBreakpoint>) {
  return screens.md ? 720 : "100vw";
}

const skuOrderColumns: ColumnsType<SkuOrderRow> = [
  { title: "Order SN", key: "order", width: 170, render: (_, row) => (isShopeeSkuOrder(row) ? row.order_sn : row.order_id) },
  { title: "Buyer", dataIndex: "buyer_name", key: "buyer_name", width: 140, render: (value: string) => value || "-" },
  { title: "Order Date", dataIndex: "order_date", key: "order_date", width: 140, render: (value: string) => formatDate(value) },
  { title: "Quantity", dataIndex: "quantity", key: "quantity", width: 90, align: "right" },
  { title: "Unit Price", key: "unit_price", width: 120, align: "right", render: (_, row) => formatCurrency(isShopeeSkuOrder(row) ? row.original_price : row.sale_price) },
  { title: "Settlement", key: "settlement", width: 130, align: "right", render: (_, row) => formatCurrency(isShopeeSkuOrder(row) ? row.escrow_amount : row.total_settlement_amount) },
];

const itemColumns: ColumnsType<OrderItemRow> = [
  { title: "Product", key: "product", render: (_, row) => (isShopeeOrderItem(row) ? row.item_name : row.product_name) },
  { title: "SKU", key: "sku", width: 150, render: (_, row) => (isShopeeOrderItem(row) ? (row.model_sku ?? row.sku) : row.seller_sku) },
  { title: "Quantity", dataIndex: "quantity", key: "quantity", width: 90, align: "right" },
  { title: "Original", dataIndex: "original_price", key: "original_price", width: 110, align: "right", render: (value: number) => formatCurrency(value) },
  { title: "Net/Sale", key: "net", width: 110, align: "right", render: (_, row) => formatCurrency(isShopeeOrderItem(row) ? row.selling_price : row.sale_price) },
  { title: "Fees", key: "fees", width: 110, align: "right", render: (_, row) => formatCurrency(isShopeeOrderItem(row) ? safeNumber(row.ams_commission_fee) + safeNumber(row.seller_order_processing_fee) : safeNumber(row.commission) + safeNumber(row.transaction_fee_item)) },
];

export function ReconciliationDrilldownDrawer({ row, open, onClose, platform, month, year }: ReconciliationDrilldownDrawerProps) {
  const screens = Grid.useBreakpoint();
  const { token } = theme.useToken();
  const sku = row?.sku ?? "";
  const query = useSkuOrders(platform, sku, month, year);
  const orders = query.data?.orders ?? [];

  return (
    <Drawer title={row ? `SKU ${row.sku}` : "SKU Detail"} open={open} onClose={onClose} width={drawerWidth(screens)} styles={{ body: { padding: 16, background: token.colorBgLayout } }}>
      {query.isLoading ? (
        <Flex justify="center" align="center" style={{ minHeight: 240 }}><Spin /></Flex>
      ) : query.isError ? (
        <Alert type="error" message="Unable to load SKU order detail" showIcon />
      ) : row ? (
        <Space direction="vertical" size={16} style={{ width: "100%" }}>
          <section>
            <Typography.Title level={5} style={{ marginTop: 0, marginBottom: 8 }}>SKU Metadata</Typography.Title>
            <Descriptions column={1} size="small" bordered>
              <Descriptions.Item label="SKU">{row.sku}</Descriptions.Item>
              <Descriptions.Item label={platform === "shopee" ? "Model SKU" : "Seller SKU"}>{skuLabel(row) || "-"}</Descriptions.Item>
              <Descriptions.Item label="Product">{productName(row) || "-"}</Descriptions.Item>
              <Descriptions.Item label="Variant">{row.variant_name || "-"}</Descriptions.Item>
              <Descriptions.Item label="Expected Income">{row.expected_income == null ? "-" : formatCurrency(row.expected_income)}</Descriptions.Item>
              <Descriptions.Item label="Inventory Price">{row.inventory_price == null ? "-" : formatCurrency(row.inventory_price)}</Descriptions.Item>
              <Descriptions.Item label="Unit Prices">{valuesLabel(row.unique_unit_prices)}</Descriptions.Item>
              <Descriptions.Item label="Actual Incomes">{valuesLabel(row.unique_actual_incomes)}</Descriptions.Item>
              <Descriptions.Item label="Status">{row.status}</Descriptions.Item>
            </Descriptions>
          </section>
          <section>
            <Typography.Title level={5} style={{ marginTop: 0, marginBottom: 8 }}>SKU Orders</Typography.Title>
            <Table<SkuOrderRow> columns={skuOrderColumns} dataSource={orders} rowKey="id" size="small" pagination={false} scroll={{ x: 760 }} locale={{ emptyText: <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description="No SKU orders" /> }} />
          </section>
        </Space>
      ) : (
        <Empty description="No SKU selected" />
      )}
    </Drawer>
  );
}

export function ShippingFeeDrilldownDrawer({ row, open, onClose, platform, month, year }: ShippingFeeDrilldownDrawerProps) {
  const screens = Grid.useBreakpoint();
  const { token } = theme.useToken();
  const orderSn = row?.order_sn ?? "";
  const query = useOrderItems(platform, orderSn, month, year);
  const items = query.data?.items ?? [];

  return (
    <Drawer title={row ? `Order ${row.order_sn}` : "Order Detail"} open={open} onClose={onClose} width={drawerWidth(screens)} styles={{ body: { padding: 16, background: token.colorBgLayout } }}>
      {query.isLoading ? (
        <Flex justify="center" align="center" style={{ minHeight: 240 }}><Spin /></Flex>
      ) : query.isError ? (
        <Alert type="error" message="Unable to load order item detail" showIcon />
      ) : row ? (
        <Space direction="vertical" size={16} style={{ width: "100%" }}>
          <section>
            <Typography.Title level={5} style={{ marginTop: 0, marginBottom: 8 }}>Order Metadata</Typography.Title>
            <Descriptions column={1} size="small" bordered>
              <Descriptions.Item label="Order SN">{row.order_sn}</Descriptions.Item>
              <Descriptions.Item label="Order Date">{formatDate(row.order_date)}</Descriptions.Item>
              <Descriptions.Item label="Status">{isShopeeShippingRow(row) ? row.status : row.order_status || row.status}</Descriptions.Item>
              <Descriptions.Item label="Buyer">{isShopeeShippingRow(row) ? row.buyer_name || "-" : "-"}</Descriptions.Item>
              <Descriptions.Item label="Payment Method">{isShopeeShippingRow(row) ? row.payment_method || "-" : "-"}</Descriptions.Item>
              <Descriptions.Item label="Currency">{isShopeeShippingRow(row) ? "-" : row.currency || "-"}</Descriptions.Item>
              <Descriptions.Item label={isShopeeShippingRow(row) ? "Buyer Paid" : "Customer Paid"}>{formatCurrency(isShopeeShippingRow(row) ? row.buyer_paid : row.customer_paid)}</Descriptions.Item>
              <Descriptions.Item label={isShopeeShippingRow(row) ? "Shopee Rebate" : "Platform Discount"}>{formatCurrency(isShopeeShippingRow(row) ? row.shopee_rebate : row.platform_discount)}</Descriptions.Item>
              <Descriptions.Item label="Actual Fee">{formatCurrency(row.actual_fee)}</Descriptions.Item>
              <Descriptions.Item label="Difference">{formatCurrency(row.difference)}</Descriptions.Item>
            </Descriptions>
          </section>
          <section>
            <Typography.Title level={5} style={{ marginTop: 0, marginBottom: 8 }}>Order Items</Typography.Title>
            <Table<OrderItemRow> columns={itemColumns} dataSource={items} rowKey="id" size="small" pagination={false} scroll={{ x: 720 }} locale={{ emptyText: <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description="No order items" /> }} />
          </section>
        </Space>
      ) : (
        <Empty description="No order selected" />
      )}
    </Drawer>
  );
}
