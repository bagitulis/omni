import { useState, useEffect } from "react";
import { InputNumber, message, theme } from "antd";
import { useUpdateInventoryRecord } from "@/hooks/useInventory";
import type { InventoryRecord } from "@/types/inventory";

interface Props {
  record: InventoryRecord;
  dataIndex: string;
  value: unknown;
  locked?: boolean;
}

export function StockCell({ record, dataIndex, value, locked = false }: Props) {
  const [editing, setEditing] = useState(false);
  const [localValue, setLocalValue] = useState<number | null>(null);
  const [flash, setFlash] = useState<"success" | "error" | null>(null);
  const { token } = theme.useToken();

  const { mutate: updateInventoryRecord } = useUpdateInventoryRecord();

  // Sync with prop value when not editing
  useEffect(() => {
    if (!editing) {
      setLocalValue(Number(value) || 0);
    }
  }, [value, editing]);

  useEffect(() => {
    if (locked) {
      setEditing(false);
    }
  }, [locked]);

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

    updateInventoryRecord(
      {
        sku: record.key_value,
        data: {
          ...(record.data || {}),
          [dataIndex]: localValue ?? 0,
        },
      },
      {
        onSuccess: () => {
          setFlash("success");
          message.success("Stock updated");
        },
        onError: (error: Error) => {
          setFlash("error");
          setLocalValue(Number(value) || 0); // Revert
          message.error(error.message || "Failed to update stock");
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

  if (editing && !locked) {
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
    <button
      type="button"
      onClick={() => {
        if (!locked) {
          setEditing(true);
        }
      }}
      disabled={locked}
      style={{
        cursor: locked ? "not-allowed" : "pointer",
        padding: "4px 8px",
        borderRadius: token.borderRadius,
        backgroundColor: getBackgroundColor(),
        transition: "background-color 0.5s ease",
        minHeight: 22,
        width: "100%",
        border: "none",
        textAlign: "left",
        font: "inherit",
        color: locked ? token.colorTextTertiary : token.colorText,
      }}
    >
      {localValue}
    </button>
  );
}
