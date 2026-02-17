import React from "react";
import { Button, Space, Tooltip, Divider } from "antd";
import { CloseOutlined } from "@ant-design/icons";
import { BatchActionItem, BatchActionType } from "@/types/shared";

export interface UnifiedBatchBarProps {
  selectedCount: number;
  actions: BatchActionItem[];
  onAction: (actionKey: BatchActionType) => void;
  onClear: () => void;
  className?: string;
  style?: React.CSSProperties;
}

export const UnifiedBatchBar: React.FC<UnifiedBatchBarProps> = ({
  selectedCount,
  actions,
  onAction,
  onClear,
  className,
  style,
}) => {
  if (selectedCount <= 0) {
    return null;
  }

  return (
    <div
      className={`fixed bottom-6 left-1/2 transform -translate-x-1/2 z-50 bg-white shadow-lg rounded-lg border border-gray-200 px-6 py-3 flex items-center justify-between gap-6 transition-all duration-300 ease-in-out ${className || ""}`}
      style={{
        minWidth: "400px",
        maxWidth: "90vw",
        ...style,
      }}
      role="toolbar"
      aria-label="Batch Actions"
    >
      <div className="flex items-center gap-4">
        <Space>
          <div className="bg-blue-50 text-blue-600 font-semibold px-3 py-1 rounded-full text-sm">
            {selectedCount} Selected
          </div>
          <Button
            type="text"
            size="small"
            icon={<CloseOutlined />}
            onClick={onClear}
            className="text-gray-500 hover:text-gray-700"
            aria-label="Clear selection"
          >
            Clear
          </Button>
        </Space>
      </div>

      <Divider type="vertical" className="h-6" />

      <Space size="small" wrap className="justify-end">
        {actions.map((action) => (
          <Tooltip
            key={action.key}
            title={action.disabled ? "Not available" : action.label}
          >
            <Button
              type={action.danger ? "primary" : "default"}
              danger={action.danger}
              icon={action.icon}
              disabled={action.disabled}
              onClick={() => onAction(action.key)}
              className={
                action.danger ? "" : "hover:border-blue-500 hover:text-blue-500"
              }
            >
              {action.label}
            </Button>
          </Tooltip>
        ))}
      </Space>
    </div>
  );
};
