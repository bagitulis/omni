import { useMemo } from "react";
import { Table } from "antd";
import { GroupedOrder, OrderTableProps } from "./OrderTable.types";
import { getOrderTableColumns } from "./OrderTableColumns";

export function OrderTable({
  orders,
  loading,
  pagination,
  selectedRowKeys,
  onSelectionChange,
  onShip,
  onPrint,
  onViewDetail,
}: OrderTableProps) {
  // Group orders by order_no
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
        });
      }

      const order = orderMap.get(key)!;
      order.items.push({
        sku: item.sku,
        product_name: item.product_name,
        variation_name: item.variation_name,
        qty: item.qty,
        price: item.price,
        product_image: item.product_image,
      });
    });

    return Array.from(orderMap.values());
  }, [orders]);

  const columns = getOrderTableColumns({
    onShip,
    onPrint,
    onViewDetail,
  });

  return (
    <Table<GroupedOrder>
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
      }}
      scroll={{ x: 900 }}
      style={{ backgroundColor: "#fff" }}
    />
  );
}
