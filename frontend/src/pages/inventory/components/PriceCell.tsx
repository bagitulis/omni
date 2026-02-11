import { useState, useEffect } from "react";
import { InputNumber, message, theme } from "antd";
import { useUpdatePrice } from "@/hooks/useInventory";
import { InventoryRecord } from "@/types/inventory";

interface Props {
  record: InventoryRecord;
  dataIndex: string;
  value: unknown;
}

export function PriceCell({ record, value }: Props) {
  const [editing, setEditing] = useState(false);
  const [localValue, setLocalValue] = useState<number | null>(null);
  const [flash, setFlash] = useState<"success" | "error" | null>(null);
  const { token } = theme.useToken();

  const { mutate: updatePrice } = useUpdatePrice();

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
    if (localValue === Number(value)) {
      setEditing(false);
      return;
    }

    setEditing(false);

    updatePrice(
      { sku: record.key_value, price: localValue || 0 },
      {
        onSuccess: () => {
          setFlash("success");
          message.success("Price updated");
        },
        onError: () => {
          setFlash("error");
          setLocalValue(Number(value) || 0);
          message.error("Failed to update price");
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

  // Format as IDR
  const formatCurrency = (val: number) => {
    return new Intl.NumberFormat("id-ID", {
      style: "currency",
      currency: "IDR",
      minimumFractionDigits: 0,
      maximumFractionDigits: 0,
    }).format(val);
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
        formatter={(value) =>
          `Rp ${value}`.replace(/\B(?=(\d{3})+(?!\d))/g, ",")
        }
        parser={(value) => Number(value?.replace(/Rp\s?|(,*)/g, ""))}
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
      {localValue != null ? formatCurrency(localValue) : "-"}
    </div>
  );
}
