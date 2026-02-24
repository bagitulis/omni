import { Modal, Button, Typography, message, List, theme } from "antd";
import { useBatchDeleteInventoryWholesale } from "@/hooks/useWholesale";

interface WholesaleBatchDeleteModalProps {
  open: boolean;
  onClose: () => void;
  selectedSkus: string[];
}

export function WholesaleBatchDeleteModal({
  open,
  onClose,
  selectedSkus,
}: WholesaleBatchDeleteModalProps) {
  const { token } = theme.useToken();
  const { mutate: batchDelete, isPending } = useBatchDeleteInventoryWholesale();

  const handleDelete = () => {
    batchDelete(selectedSkus, {
      onSuccess: () => {
        message.success(
          `Removed wholesale tiers for ${selectedSkus.length} items`,
        );
        onClose();
      },
      onError: (error) => {
        message.error(`Failed to delete: ${error.message}`);
      },
    });
  };

  return (
    <Modal
      title="Delete Wholesale Tiers"
      open={open}
      onCancel={onClose}
      destroyOnHidden
      width={500}
      footer={[
        <Button key="cancel" onClick={onClose}>
          Cancel
        </Button>,
        <Button
          key="delete"
          type="primary"
          danger
          onClick={handleDelete}
          loading={isPending}
        >
          Confirm Delete
        </Button>,
      ]}
    >
      <div style={{ marginBottom: 16 }}>
        <Typography.Text type="danger">
          Are you sure you want to remove wholesale pricing tiers for the
          following <strong>{selectedSkus.length}</strong> items? This action
          cannot be undone.
        </Typography.Text>
      </div>

      <div
        style={{
          maxHeight: 200,
          overflowY: "auto",
          marginTop: 8,
          border: `1px solid ${token.colorBorder}`,
          borderRadius: 4,
          padding: 8,
        }}
      >
        <List
          size="small"
          dataSource={selectedSkus}
          renderItem={(sku) => <List.Item>{sku}</List.Item>}
        />
      </div>
    </Modal>
  );
}
