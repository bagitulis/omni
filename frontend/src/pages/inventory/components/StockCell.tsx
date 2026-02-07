import { useState, useEffect, useRef, KeyboardEvent } from "react";
import { InputNumber } from "antd";

export interface StockCellProps {
  value: number;
  itemSku: string;
  isEditing: boolean;
  onStartEdit: () => void;
  onSave: (sku: string, val: number) => void;
  onCancel: () => void;
}

export function StockCell({
  value,
  itemSku,
  isEditing,
  onStartEdit,
  onSave,
  onCancel,
}: StockCellProps) {
  const [editValue, setEditValue] = useState(value);
  const wrapperRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    if (isEditing) {
      setEditValue(value);
      setTimeout(() => {
        const input = wrapperRef.current?.querySelector("input");
        input?.focus();
        input?.select();
      }, 0);
    }
  }, [isEditing, value]);

  const handleKeyDown = (e: KeyboardEvent) => {
    if (e.key === "Escape") {
      e.preventDefault();
      onCancel();
    } else if (e.key === "Enter") {
      e.preventDefault();
      onSave(itemSku, editValue);
    }
  };

  if (!isEditing) {
    return (
      <div
        onDoubleClick={onStartEdit}
        style={{
          cursor: "pointer",
          padding: "4px 8px",
          borderRadius: 3,
          minHeight: 24,
          display: "flex",
          alignItems: "center",
        }}
        title="Double-click to edit"
      >
        {value}
      </div>
    );
  }

  return (
    <div ref={wrapperRef}>
      <InputNumber
        value={editValue}
        min={0}
        onChange={(val) => setEditValue(val ?? 0)}
        onKeyDown={handleKeyDown}
        onBlur={() => onSave(itemSku, editValue)}
        style={{
          width: "100%",
          borderColor: "#0369a1",
          borderRadius: 3,
        }}
      />
    </div>
  );
}
