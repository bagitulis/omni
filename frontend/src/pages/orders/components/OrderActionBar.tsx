import { OrderFilters } from "@/components/forms/OrderFilters";
import { Dayjs } from "dayjs";

interface OrderActionBarProps {
  onSearch: (value: string) => void;
  onPlatformChange: (value: string) => void;
  onDateChange: (dates: [Dayjs | null, Dayjs | null] | null) => void;
  onRefresh: () => void;
  onExport: () => void;
  loading: boolean;
  autoRefresh: boolean;
  onAutoRefreshChange: (enabled: boolean) => void;
}

export function OrderActionBar(props: OrderActionBarProps) {
  return <OrderFilters {...props} />;
}
