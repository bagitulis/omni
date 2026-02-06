import { Tag } from "antd";

interface StatusPipelineProps {
  status: string;
}

const STATUS_CONFIG: Record<string, { color: string; label: string }> = {
  PENDING: { color: "orange", label: "Pending" },
  UNPAID: { color: "red", label: "Unpaid" },
  READY_TO_SHIP: { color: "blue", label: "Ready to Ship" },
  PROCESSED: { color: "cyan", label: "Processed" },
  SHIPPED: { color: "geekblue", label: "Shipped" },
  COMPLETED: { color: "green", label: "Completed" },
  CANCELLED: { color: "default", label: "Cancelled" },
  IN_CANCEL: { color: "volcano", label: "Cancelling" },
};

export function StatusPipeline({ status }: StatusPipelineProps) {
  const config = STATUS_CONFIG[status] || { color: "default", label: status };
  return <Tag color={config.color}>{config.label}</Tag>;
}

