import { Modal, Button, Alert, Typography, List, message } from "antd";
import { useState } from "react";

const { Text } = Typography;

interface WholesaleBatchDeleteModalProps {
  open: boolean;
  onClose: () => void;
  skus: string[];
  onConfirm?: (skus: string[]) => Promise<void>;
}

export function WholesaleBatchDeleteModal({
  open,
  onClose,
  skus,
  onConfirm,
}: WholesaleBatchDeleteModalProps) {
  const [processing, setProcessing] = useState(false);

  const handleDelete = async () => {
    setProcessing(true);
    try {
      if (onConfirm) {
        await onConfirm(skus);
      } else {
        // Mock delete
        await new Promise((resolve) => setTimeout(resolve, 1000));
        message.success("Wholesale pricing removed successfully");
      }
      onClose();
    } catch (error) {
      message.error("Failed to delete wholesale");
    } finally {
      setProcessing(false);
    }
  };

  return (
    <Modal
      title="🗑️ Delete Wholesale"
      open={open}
      onCancel={onClose}
      footer={[
        <Button key="cancel" onClick={onClose} disabled={processing}>
          Cancel
        </Button>,
        <Button
          key="delete"
          type="primary"
          danger
          onClick={handleDelete}
          loading={processing}
          disabled={skus.length === 0}
        >
          Delete Wholesale
        </Button>,
      ]}
      width={500}
    >
      <div style={{ display: "flex", flexDirection: "column", gap: 16 }}>
        <Alert
          message={
            <div>
              <strong>{skus.length} SKU</strong> selected. Wholesale will be
              removed per product (item_id), not per SKU.
            </div>
          }
          type="warning"
          showIcon
        />

        <div>
          <Text strong>Selected SKUs:</Text>
          <div
            style={{
              maxHeight: 200,
              overflowY: "auto",
              marginTop: 8,
              border: "1px solid #f0f0f0",
              borderRadius: 4,
            }}
          >
            <List
              size="small"
              dataSource={skus.slice(0, 10)}
              renderItem={(item) => <List.Item>{item}</List.Item>}
              footer={
                skus.length > 10 ? (
                  <div style={{ padding: "8px 16px", color: "#999" }}>
                    ... and {skus.length - 10} others
                  </div>
                ) : null
              }
            />
          </div>
        </div>
      </div>
    </Modal>
  );
}
