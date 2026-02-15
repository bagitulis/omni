import { useState, useEffect } from "react";
import { InputNumber, message, theme } from "antd";
import { useUpdateStock } from "@/hooks/useInventory";
import { InventoryRecord } from "@/types/inventory";

interface Props {
  record: InventoryRecord;
  dataIndex: string;
  value: unknown;
}

export function StockCell({ record, value }: Props) {
  const [editing, setEditing] = useState(false);
  const [localValue, setLocalValue] = useState<number | null>(null);
  const [flash, setFlash] = useState<"success" | "error" | null>(null);
  const { token } = theme.useToken();

  const { mutate: updateStock } = useUpdateStock();

  // Sync with prop value when not editing
  useEffect(() => {
    if (!editing) {
      setLocalValue(Number(value) || 0);
    }
  }, [value, editing]);

  // Flash effect
  useEffect(() => {
    if (flash) {
      const timer = setTimeout(() => {
        setFlash(null);
      }, 1000);
      return () => clearTimeout(timer);
    }
  }, [flash]);

  const handleSave = () => {
    // If value hasn't changed, just exit edit mode
    if (localValue === Number(value)) {
      setEditing(false);
      return;
    }

    setEditing(false);

    updateStock(
      { sku: record.key_value },
      {
        onSuccess: () => {
          setFlash("success");
          message.success("Stock updated");
        },
        onError: () => {
          setFlash("error");
          setLocalValue(Number(value) || 0); // Revert
          message.error("Failed to update stock");
        },
      },
    );
  };

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === "Enter") {
      handleSave();
    } else if (e.key === "Escape") {
      setEditing(false);
      setLocalValue(Number(value) || 0);
    }
  };

  const getBackgroundColor = () => {
    if (flash === "success") return token.colorSuccessBg;
    if (flash === "error") return token.colorErrorBg;
    return "transparent";
  };

  if (editing) {
    return (
      <InputNumber
        value={localValue}
        onChange={(val) => setLocalValue(val)}
        onBlur={handleSave}
        onKeyDown={handleKeyDown}
        autoFocus
        style={{ width: "100%" }}
        size="small"
        min={0}
      />
    );
  }

  return (
    <div
      onClick={() => setEditing(true)}
      role="button"
      tabIndex={0}
      onKeyDown={(e) => {
        if (e.key === "Enter" || e.key === " ") {
          setEditing(true);
        }
      }}
      style={{
        cursor: "pointer",
        padding: "4px 8px",
        borderRadius: token.borderRadius,
        backgroundColor: getBackgroundColor(),
        transition: "background-color 0.5s ease",
        minHeight: 22,
      }}
    >
      {localValue}
    </div>
  );
}
