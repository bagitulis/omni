import { Steps } from "antd";

interface StatusPipelineProps {
  status: string;
}

const STEPS = [
  { title: "Paid", key: "PENDING" },
  { title: "Verify", key: "READY_TO_SHIP" },
  { title: "Ship", key: "SHIPPED" },
  { title: "Done", key: "COMPLETED" },
];

const STATUS_MAP: Record<string, number> = {
  PENDING: 0,
  READY_TO_SHIP: 1,
  SHIPPED: 2,
  COMPLETED: 3,
  CANCELLED: -1,
};

export function StatusPipeline({ status }: StatusPipelineProps) {
  const currentStep = STATUS_MAP[status] ?? 0;

  if (status === "CANCELLED") {
    return (
      <div className="flex items-center text-red-600 text-xs font-medium">
        <span className="w-2 h-2 rounded-full bg-red-600 mr-2" />
        Order Cancelled
      </div>
    );
  }

  return (
    <Steps
      size="small"
      current={currentStep}
      items={STEPS.map((_step) => ({
        title: null, // Hide title for compact view in table
        description: null,
      }))}
      progressDot
      style={{ maxWidth: 100 }}
    />
  );
}
