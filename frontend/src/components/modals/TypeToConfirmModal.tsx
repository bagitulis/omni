import { useEffect, useState } from "react";
import { Input, Modal, Typography } from "antd";

const { Text } = Typography;

export interface TypeToConfirmModalProps {
  open: boolean;
  title: string;
  description: string;
  confirmText: string;
  onConfirm: () => void;
  onCancel: () => void;
  danger?: boolean;
  loading?: boolean;
}

export function TypeToConfirmModal({
  open,
  title,
  description,
  confirmText,
  onConfirm,
  onCancel,
  danger = false,
  loading = false,
}: TypeToConfirmModalProps) {
  const [inputValue, setInputValue] = useState("");

  const isMatch = inputValue === confirmText;

  useEffect(() => {
    if (!open) {
      setInputValue("");
    }
  }, [open]);

  return (
    <Modal
      title={title}
      open={open}
      onCancel={onCancel}
      destroyOnClose
      width={480}
      okText="Confirm"
      okButtonProps={{
        danger,
        disabled: !isMatch,
        loading,
      }}
      onOk={onConfirm}
    >
      <div style={{ marginTop: 8 }}>
        <Text>{description}</Text>
        <div style={{ marginTop: 16 }}>
          <Text type="secondary" style={{ fontSize: 12 }}>
            Type <Text strong code>{confirmText}</Text> to confirm
          </Text>
          <Input
            style={{ marginTop: 8 }}
            placeholder={`Type "${confirmText}" to confirm`}
            value={inputValue}
            onChange={(e) => setInputValue(e.target.value)}
            autoFocus
          />
        </div>
      </div>
    </Modal>
  );
}
