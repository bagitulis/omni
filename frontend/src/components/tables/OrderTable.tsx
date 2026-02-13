import React, { useMemo } from "react";
import { theme } from "antd";
import { getOrderTableColumns } from "./OrderTableColumns";
import type { GroupedOrder, OrderTableProps } from "./OrderTable.types";
import { VirtualTable } from "@/components/common/VirtualTable";

function OrderTable({
  orders,
  loading,
  pagination,
  selectedRowKeys,
  onSelectionChange,
  onShip,
  onPrint,
  onCancel,
  onViewDetail,
}: OrderTableProps) {
  const { token } = theme.useToken();
  // Group orders by order_no, then accumulate same-SKU items within each order
  const groupedOrders = useMemo(() => {
    const orderMap = new Map<string, GroupedOrder>();

    orders.forEach((item) => {
      const key = item.order_no;

      if (!orderMap.has(key)) {
        orderMap.set(key, {
          key,
          ...item,
          items: [],
          status: item.status || item.order_status,
          countdown: item.countdown,
          tracking_number: item.tracking_number,
          buyer_message: item.buyer_message,
        });
      }

      const order = orderMap.get(key);
      if (!order) {
        return;
      }

      // Accumulate by SKU + variation_name as grouping key (safety net for un-grouped backend data)
      const itemKey = `${item.sku || ""}|${item.variation_name || ""}|${item.product_name || ""}`;
      const existing = order.items.find(
        (i) =>
          `${i.sku || ""}|${i.variation_name || ""}|${i.product_name || ""}` ===
          itemKey,
      );

      if (existing) {
        existing.qty += item.qty || 1;
      } else {
        order.items.push({
          sku: item.sku,
          product_name: item.product_name,
          variation_name: item.variation_name,
          qty: item.qty || 1,
          price: item.price,
          product_image: item.product_image,
        });
      }
    });

    return Array.from(orderMap.values());
  }, [orders]);

  const columns = getOrderTableColumns({
    onShip,
    onPrint,
    onCancel,
    onViewDetail,
  });

  return (
    <VirtualTable<GroupedOrder>
      columns={columns}
      dataSource={groupedOrders}
      loading={loading}
      rowKey="key"
      size="middle"
      pagination={{
        current: pagination.current,
        pageSize: pagination.pageSize,
        total: pagination.total,
        onChange: pagination.onChange,
        showSizeChanger: true,
        showTotal: (total) => `Total ${total} orders`,
        size: "small",
      }}
      rowSelection={{
        selectedRowKeys,
        onChange: onSelectionChange,
        columnWidth: 40,
      }}
      scroll={{ x: 900 }}
      style={{ backgroundColor: token.colorBgContainer }}
      enableVirtual={groupedOrders.length > 50}
      offsetBottom={320}
    />
  );
}

export const OrderTableMemo = React.memo(OrderTable);
export { OrderTableMemo as OrderTable };
