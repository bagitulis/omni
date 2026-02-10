import { Modal, InputNumber, Typography, Button } from "antd";
import { useState, useEffect } from "react";

const { Text } = Typography;

interface ProductBatchModalProps {
  open: boolean;
  onClose: () => void;
  type: "price" | "stock";
  count: number;
  onConfirm: (value: number) => Promise<void>;
  loading?: boolean;
}

export function ProductBatchModal({
  open,
  onClose,
  type,
  count,
  onConfirm,
  loading = false,
}: ProductBatchModalProps) {
  const [value, setValue] = useState<number | null>(null);

  useEffect(() => {
    if (open) {
      setValue(null);
    }
  }, [open]);

  const handleConfirm = async () => {
    if (value !== null) {
      await onConfirm(value);
      onClose();
    }
  };

  const title =
    type === "price" ? "💰 Update Batch Price" : "📦 Update Batch Stock";
  const placeholder =
    type === "price" ? "Enter new price..." : "Enter new stock...";
  const step = type === "price" ? 1000 : 1;

  return (
    <Modal
      title={title}
      open={open}
      onCancel={onClose}
      confirmLoading={loading}
      footer={[
        <Button key="cancel" onClick={onClose} disabled={loading}>
          Cancel
        </Button>,
        <Button
          key="submit"
          type="primary"
          onClick={handleConfirm}
          loading={loading}
          disabled={value === null || value < 0}
        >
          Update All
        </Button>,
      ]}
      width={400}
    >
      <div style={{ display: "flex", flexDirection: "column", gap: 16 }}>
        <Text type="secondary">
          You are about to update <strong>{count}</strong> variants at once.
        </Text>
        <div>
          <div style={{ marginBottom: 8 }}>
            <Text strong>
              {type === "price" ? "New Price (Rp)" : "New Stock"}
            </Text>
          </div>
          <InputNumber
            style={{ width: "100%" }}
            value={value}
            onChange={setValue}
            placeholder={placeholder}
            min={0}
            step={step}
            // formatter={(value) =>
            //   type === "price"
            //     ? `${value}`.replace(/\B(?=(\d{3})+(?!\d))/g, ",")
            //     : `${value}`
            // }
            // parser={(value) =>
            //   value?.replace(/\$\s?|(,*)/g, "") as unknown as number
            // }
          />
        </div>
      </div>
    </Modal>
  );
}
