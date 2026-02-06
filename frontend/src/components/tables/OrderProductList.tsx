import { Avatar, Typography } from "antd";
import { OrderItem } from "./OrderTypes";

const { Text } = Typography;

interface OrderProductListProps {
  items: OrderItem[];
  platform: string;
}

export function OrderProductList({ items }: OrderProductListProps) {
  return (
    <div className="flex flex-col gap-3">
      {items.map((item, index) => (
        <div key={`${item.sku || 'nosku'}-${index}`} className="flex gap-3 items-start p-1 rounded-sm hover:bg-gray-50 transition-colors">
          <div className="flex-shrink-0 w-[60px] h-[60px]">
             <Avatar 
                shape="square" 
                size={60} 
                src={item.product_image} 
                alt={item.product_name}
                className="border border-slate-200"
              >
                {item.product_name?.charAt(0)}
              </Avatar>
          </div>
          
          <div className="flex flex-col flex-1 min-w-0 gap-0.5">
            <div className="flex justify-between items-start gap-2">
              <Text className="text-sm font-medium leading-tight line-clamp-2 overflow-hidden text-ellipsis" title={item.product_name}>
                {item.product_name}
              </Text>
              <span className="text-xs font-semibold bg-slate-100 px-1.5 py-0.5 rounded-full whitespace-nowrap text-slate-600">
                x{item.qty}
              </span>
            </div>
            
            {item.variation_name && (
              <Text type="secondary" className="text-xs truncate" title={item.variation_name}>
                Var: {item.variation_name}
              </Text>
            )}
            
            {item.sku && (
              <Text type="secondary" className="text-xs font-mono text-slate-400">
                SKU: {item.sku}
              </Text>
            )}
          </div>
        </div>
      ))}
    </div>
  );
}
