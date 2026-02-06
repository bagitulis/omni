import { useState, useEffect } from "react";
import { Avatar, Button, Typography, Checkbox, Dropdown, MenuProps } from "antd";
import { UserOutlined, MessageOutlined, CopyOutlined, CheckOutlined, MoreOutlined } from "@ant-design/icons";
import { GroupedOrder } from "./OrderTypes";
import { OrderProductList } from "./OrderProductList";
import { StatusPipeline } from "../ui/StatusPipeline";

const { Text } = Typography;

interface OrderCardProps {
  order: GroupedOrder;
  selected: boolean;
  onSelect: (checked: boolean) => void;
  onShip: (order: GroupedOrder) => void;
  onPrint: (order: GroupedOrder) => void;
  onViewDetail?: (order: GroupedOrder) => void;
}

export function OrderCard({
  order,
  selected,
  onSelect,
  onShip,
  onPrint,
  onViewDetail,
}: OrderCardProps) {
  const [copied, setCopied] = useState(false);
  const [now, setNow] = useState(Date.now());

  useEffect(() => {
    if (!order.ship_by_date) return;
    const interval = setInterval(() => setNow(Date.now()), 60000);
    return () => clearInterval(interval);
  }, [order.ship_by_date]);

  const getCountdown = () => {
    if (!order.ship_by_date) return "-";
    const deadline = order.ship_by_date * 1000;
    const diff = deadline - now;
    if (diff <= 0) return "Overdue";
    
    const hours = Math.floor(diff / (1000 * 60 * 60));
    const minutes = Math.floor((diff % (1000 * 60 * 60)) / (1000 * 60));
    
    if (hours >= 24) {
      const days = Math.floor(hours / 24);
      return `${days}d ${hours % 24}h`;
    }
    return `${hours}h ${minutes}m`;
  };

  const getCountdownClass = () => {
    if (!order.ship_by_date) return "text-slate-600";
    const deadline = order.ship_by_date * 1000;
    const diff = deadline - now;
    if (diff <= 0) return "text-red-600 font-bold";
    if (diff <= 24 * 60 * 60 * 1000) return "text-orange-500 font-semibold";
    return "text-slate-600";
  };

  const handleCopy = () => {
    navigator.clipboard.writeText(order.order_no);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  };

  const formatAmount = (amt: number, cur: string) => {
    return new Intl.NumberFormat("id-ID", {
      style: "currency",
      currency: cur || "IDR",
      minimumFractionDigits: 0,
    }).format(amt);
  };

  const actionMenu: MenuProps["items"] = [
    { key: "view", label: "View Details", onClick: () => onViewDetail?.(order) },
    { key: "print", label: "Print Label", onClick: () => onPrint(order) },
  ];

  return (
    <div className="bg-white border border-slate-200 rounded mb-3 overflow-hidden shadow-sm hover:shadow-md transition-shadow">
      {/* Buyer Header - Full Width Gray Bar */}
      <div 
        className="px-4 py-2 border-b border-slate-200 flex justify-between items-center"
        style={{ backgroundColor: '#f5f5f5' }}
      >
        <div className="flex items-center gap-3">
          <Checkbox checked={selected} onChange={(e) => onSelect(e.target.checked)} />
          <Avatar size={28} icon={<UserOutlined />} className="bg-sky-600" />
          <Text strong className="text-slate-800 text-sm">{order.buyer_username}</Text>
          <Button type="text" size="small" icon={<MessageOutlined className="text-slate-400" />} />
        </div>
        
        <div className="flex items-center gap-3">
          <Text type="secondary" className="text-xs">Order ID</Text>
          <Text strong className="text-sm text-slate-700">{order.order_no}</Text>
          <Button
            type="text"
            size="small"
            icon={copied ? <CheckOutlined className="text-green-500" /> : <CopyOutlined className="text-slate-400" />}
            onClick={handleCopy}
          />
        </div>
      </div>

      {/* Order Content - 6 Column Grid */}
      <div 
        className="px-4 py-3 gap-3 items-start text-sm"
        style={{ display: 'grid', gridTemplateColumns: '2.5fr 1fr 0.8fr 1fr 1fr 1fr' }}
      >
        {/* Product Column */}
        <div className="min-w-0 pr-3">
          <OrderProductList items={order.items} platform={order.platform} />
        </div>

        {/* Amount Paid Column */}
        <div className="flex flex-col gap-1 px-3 border-l border-slate-200">
          <Text strong className="text-slate-800">{formatAmount(order.total_amount, order.currency)}</Text>
          <Text type="secondary" className="text-xs">{order.payment_method || "Online Payment"}</Text>
        </div>

        {/* Status Column */}
        <div className="flex flex-col gap-1 px-3 border-l border-slate-200">
          <StatusPipeline status={order.status} />
        </div>

        {/* Countdown Column */}
        <div className="flex flex-col gap-1 px-3 border-l border-slate-200">
          <span className={`text-sm ${getCountdownClass()}`}>{getCountdown()}</span>
          {order.ship_by_date && (
            <Text type="secondary" className="text-xs">
              Ship by: {new Date(order.ship_by_date * 1000).toLocaleDateString()}
            </Text>
          )}
        </div>

        {/* Shipping Column */}
        <div className="flex flex-col gap-1 px-3 border-l border-slate-200">
          <Text className="text-slate-700">{order.shipping_carrier || "-"}</Text>
        </div>

        {/* Action Column */}
        <div className="flex flex-col gap-2 px-3 border-l border-slate-200">
          <Button 
            type="primary" 
            size="small"
            disabled={order.status !== "READY_TO_SHIP"}
            onClick={() => onShip(order)}
            block
          >
            Ship
          </Button>
          <div className="flex justify-between items-center">
            <Button type="link" size="small" className="p-0 h-auto text-xs" onClick={() => onViewDetail?.(order)}>
              Details
            </Button>
            <Dropdown menu={{ items: actionMenu }} trigger={["click"]}>
              <Button type="text" size="small" icon={<MoreOutlined />} />
            </Dropdown>
          </div>
        </div>
      </div>
    </div>
  );
}
