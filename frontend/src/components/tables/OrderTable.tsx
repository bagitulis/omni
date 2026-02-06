import { useMemo } from "react";
import { Pagination, Checkbox, Spin, Empty } from "antd";
import { Order } from "@/types/order";
import { OrderCard } from "./OrderCard";
import { GroupedOrder } from "./OrderTypes";

interface OrderTableProps {
  orders: Order[];
  loading: boolean;
  pagination: {
    current: number;
    pageSize: number;
    total: number;
    onChange: (page: number, pageSize: number) => void;
  };
  selectedRowKeys: React.Key[];
  onSelectionChange: (selectedRowKeys: React.Key[]) => void;
  onShip: (order: any) => void; // Using any to accommodate GroupedOrder vs Order mismatch
  onPrint: (order: any) => void;
  onViewDetail?: (order: any) => void;
}

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
      // Use order_no as grouping key
      const key = item.order_no;
      
      if (!orderMap.has(key)) {
        orderMap.set(key, {
          ...item,
          // Explicit overrides to ensure correct mapping if needed
          items: [],
          status: item.status || item.order_status,
        });
      }
      
      const order = orderMap.get(key)!;
      // Add item to the list
      order.items.push({
        sku: item.sku,
        product_name: item.product_name,
        variation_name: item.variation_name,
        qty: item.qty,
        price: item.price,
        product_image: item.product_image,
      });
      
      // Update totals if needed (though usually total_amount is per order)
    });
    
    return Array.from(orderMap.values());
  }, [orders]);

  const allSelected = groupedOrders.length > 0 && groupedOrders.every(o => selectedRowKeys.includes(o.order_sn || o.order_no));
  const indeterminate = groupedOrders.some(o => selectedRowKeys.includes(o.order_sn || o.order_no)) && !allSelected;

  const handleSelectAll = (checked: boolean) => {
    if (checked) {
      const allKeys = groupedOrders.map(o => o.order_sn || o.order_no);
      // Merge with existing keys if we want to keep selection across pages (optional, but standard behavior usually replaces on page select)
      // Here assuming we select all on current page
      onSelectionChange(allKeys);
    } else {
      onSelectionChange([]);
    }
  };

  const handleSelectRow = (key: string, checked: boolean) => {
    if (checked) {
      onSelectionChange([...selectedRowKeys, key]);
    } else {
      onSelectionChange(selectedRowKeys.filter(k => k !== key));
    }
  };

  if (loading) {
    return (
      <div className="flex justify-center items-center h-64">
        <Spin size="large" />
      </div>
    );
  }

  if (orders.length === 0) {
    return <Empty description="No orders found" />;
  }

  return (
    <div className="flex flex-col gap-4">
      {/* Table Header */}
      <div 
        className="bg-gray-100 p-4 border-b-2 border-slate-200 rounded-t-md text-xs font-bold text-slate-500 uppercase tracking-wide items-center sticky top-0 z-10"
        style={{ display: 'grid', gridTemplateColumns: '2.5fr 1fr 0.8fr 1fr 1fr 1fr' }}
      >
        <div className="flex items-center gap-3">
          <Checkbox 
            checked={allSelected} 
            indeterminate={indeterminate}
            onChange={(e) => handleSelectAll(e.target.checked)}
          />
          <span>Product</span>
        </div>
        <div className="pl-2 border-l border-slate-300">Amount Paid</div>
        <div className="pl-2 border-l border-slate-300">Status</div>
        <div className="pl-2 border-l border-slate-300">Countdown</div>
        <div className="pl-2 border-l border-slate-300">Shipping</div>
        <div className="pl-2 border-l border-slate-300">Action</div>
      </div>

      {/* Order List */}
      <div className="flex flex-col">
        {groupedOrders.map((order) => (
          <OrderCard
            key={order.order_sn || order.order_no}
            order={order}
            selected={selectedRowKeys.includes(order.order_sn || order.order_no)}
            onSelect={(checked) => handleSelectRow(order.order_sn || order.order_no, checked)}
            onShip={onShip}
            onPrint={onPrint}
            onViewDetail={onViewDetail}
          />
        ))}
      </div>

      {/* Pagination */}
      <div className="flex justify-end pt-4">
        <Pagination
          current={pagination.current}
          pageSize={pagination.pageSize}
          total={pagination.total}
          onChange={pagination.onChange}
          showSizeChanger
          showTotal={(total) => `Total ${total} orders`}
          size="small"
        />
      </div>
    </div>
  );
}
