import type { JobHistory } from "@/types/scriptMonitor";

const JOB_TYPE_LABELS: Record<string, string> = {
  inventory_sync: "Inventory Sync",
  order_sync: "Order Sync",
  shipping_sync: "Shipping Sync",
  product_sync: "Product Sync",
};

const PRIORITY_LABELS: Record<string, string> = {
  low: "Low",
  normal: "Normal",
  medium: "Medium",
  high: "High",
  urgent: "Urgent",
};

export function getJobTypeLabel(type: string): string {
  if (!type) {
    return "Unknown";
  }

  return (
    JOB_TYPE_LABELS[type] ||
    type
      .split("_")
      .map((part) => part.charAt(0).toUpperCase() + part.slice(1))
      .join(" ")
  );
}

export function getJobPriorityLabel(priority: string): string {
  if (!priority) {
    return "Normal";
  }

  return PRIORITY_LABELS[priority] || priority.toUpperCase();
}

export function getStatusColor(status: string): string {
  switch (status) {
    case "completed":
      return "success";
    case "failed":
      return "error";
    case "cancelled":
      return "warning";
    case "processing":
      return "processing";
    default:
      return "default";
  }
}

export function getUniqueJobTypes(history: JobHistory[]): string[] {
  return Array.from(
    new Set(
      history
        .map((item) => item.job_type)
        .filter(
          (value): value is string =>
            typeof value === "string" && value.length > 0,
        ),
    ),
  ).sort((left, right) => left.localeCompare(right));
}
