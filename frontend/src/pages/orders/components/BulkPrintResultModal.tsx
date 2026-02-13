import { Alert, Button, List, Modal, Space, Tag, Typography } from "antd";
import type { BulkResult } from "../hooks/bulkActionTypes";

const { Text } = Typography;

interface BulkPrintResultModalProps {
  open: boolean;
  result: BulkResult;
  onClose: () => void;
  onRetryFailed: () => void;
  isRetrying: boolean;
}

export function BulkPrintResultModal({
  open,
  result,
  onClose,
  onRetryFailed,
  isRetrying,
}: BulkPrintResultModalProps) {
  const hasFailures = result.failed.length > 0;

  return (
    <Modal
      open={open}
      title="Bulk Print Result Details"
      onCancel={onClose}
      width={760}
      footer={[
        <Button key="close" onClick={onClose}>
          Close
        </Button>,
        <Button
          key="retry"
          type="primary"
          onClick={onRetryFailed}
          loading={isRetrying}
          disabled={!hasFailures}
        >
          Retry Failed ({result.failed.length})
        </Button>,
      ]}
    >
      <Space direction="vertical" size={12} style={{ width: "100%" }}>
        <Alert
          type={hasFailures ? "warning" : "success"}
          showIcon
          message={
            hasFailures
              ? `${result.succeeded.length} labels downloaded, ${result.failed.length} failed`
              : `All ${result.succeeded.length} labels downloaded successfully`
          }
        />

        <div>
          <Text strong>Downloaded Labels</Text>
          <List
            size="small"
            bordered
            locale={{ emptyText: "No downloaded labels" }}
            dataSource={result.succeeded}
            style={{ marginTop: 8 }}
            renderItem={(orderSn) => (
              <List.Item>
                <Space>
                  <Tag color="success">Downloaded</Tag>
                  <Text>{orderSn}</Text>
                </Space>
              </List.Item>
            )}
          />
        </div>

        <div>
          <Text strong>Failed Labels (Raw Platform Error)</Text>
          <List
            size="small"
            bordered
            locale={{ emptyText: "No failed labels" }}
            dataSource={result.failed}
            style={{ marginTop: 8 }}
            renderItem={(item) => (
              <List.Item>
                <Space direction="vertical" size={2} style={{ width: "100%" }}>
                  <Text strong>{item.order_sn}</Text>
                  <Text code style={{ whiteSpace: "normal" }}>
                    {item.error}
                  </Text>
                </Space>
              </List.Item>
            )}
          />
        </div>
      </Space>
    </Modal>
  );
}
