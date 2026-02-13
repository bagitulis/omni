import {
  Button,
  Card,
  Divider,
  Flex,
  Progress,
  Space,
  Typography,
  theme,
} from "antd";
import {
  FileSearchOutlined,
  PrinterOutlined,
  SendOutlined,
} from "@ant-design/icons";
import { useState } from "react";
import { BulkPrintResultModal } from "./BulkPrintResultModal";
import type { BulkResult, ProgressState } from "../hooks/bulkActionTypes";

const { Text } = Typography;

interface OrdersBulkActionsBarProps {
  selectedCount: number;
  onBulkShip: () => void;
  onBulkPrint: () => void;
  onBulkCancel: () => void;
  onRetryFailedPrint: () => void;
  onClearSelection: () => void;
  isShipping: boolean;
  isPrinting: boolean;
  isCancelling: boolean;
  shipProgress: ProgressState;
  printProgress: ProgressState;
  cancelProgress: ProgressState;
  shipResult: BulkResult;
  printResult: BulkResult;
  cancelResult: BulkResult;
}

export function OrdersBulkActionsBar({
  selectedCount,
  onBulkShip,
  onBulkPrint,
  onBulkCancel,
  onRetryFailedPrint,
  onClearSelection,
  isShipping,
  isPrinting,
  isCancelling,
  shipProgress,
  printProgress,
  cancelProgress,
  shipResult,
  printResult,
  cancelResult,
}: OrdersBulkActionsBarProps) {
  const [isPrintDetailsOpen, setIsPrintDetailsOpen] = useState(false);
  const { token } = theme.useToken();
  if (selectedCount <= 0) return null;

  const getProgressText = (progress: ProgressState, action: string) => {
    if (progress.status === "processing") {
      return `${action} ${progress.current}/${progress.total}...`;
    }
    return null;
  };

  const getResultText = (
    result: BulkResult,
    progress: ProgressState,
    action: string,
  ) => {
    if (
      progress.status === "done" &&
      (result.succeeded.length > 0 || result.failed.length > 0)
    ) {
      if (result.failed.length === 0) {
        return `✓ ${result.succeeded.length} ${action}`;
      }
      return `${result.succeeded.length} ${action}, ${result.failed.length} failed`;
    }
    return null;
  };

  return (
    <Card
      size="small"
      style={{
        backgroundColor: token.colorFillQuaternary,
        border: `1px solid ${token.colorBorderSecondary}`,
        borderRadius: 4,
      }}
    >
      <Flex vertical gap={8}>
        <Flex justify="space-between" align="center">
          <Space split={<Divider type="vertical" />}>
            <Text strong style={{ color: token.colorPrimary }}>
              {selectedCount} orders selected
            </Text>
            <Button
              type="primary"
              icon={<SendOutlined />}
              onClick={onBulkShip}
              loading={isShipping}
            >
              Bulk Ship
            </Button>
            <Button
              icon={<PrinterOutlined />}
              onClick={onBulkPrint}
              loading={isPrinting}
            >
              Bulk Print Labels
            </Button>
            <Button danger onClick={onBulkCancel} loading={isCancelling}>
              Bulk Cancel
            </Button>
          </Space>
          <Button type="text" onClick={onClearSelection}>
            Clear Selection
          </Button>
        </Flex>

        {/* Ship Progress */}
        {shipProgress.status === "processing" && (
          <Flex align="center" gap={8}>
            <Text type="secondary" style={{ fontSize: 12 }}>
              {getProgressText(shipProgress, "Shipping")}
            </Text>
            <Progress
              percent={Math.round(
                (shipProgress.current / shipProgress.total) * 100,
              )}
              size="small"
              status="active"
              style={{ flex: 1, margin: 0 }}
            />
          </Flex>
        )}
        {shipProgress.status === "done" &&
          getResultText(shipResult, shipProgress, "shipped") && (
            <Text type="success" style={{ fontSize: 12 }}>
              {getResultText(shipResult, shipProgress, "shipped")}
            </Text>
          )}

        {/* Print Progress */}
        {printProgress.status === "processing" && (
          <Flex align="center" gap={8}>
            <Text type="secondary" style={{ fontSize: 12 }}>
              {getProgressText(printProgress, "Printing")}
            </Text>
            <Progress
              percent={Math.round(
                (printProgress.current / printProgress.total) * 100,
              )}
              size="small"
              status="active"
              style={{ flex: 1, margin: 0 }}
            />
          </Flex>
        )}
        {printProgress.status === "done" &&
          getResultText(printResult, printProgress, "printed") && (
            <Flex align="center" justify="space-between" gap={8}>
              <Text
                type={printResult.failed.length > 0 ? "warning" : "success"}
                style={{ fontSize: 12 }}
              >
                {getResultText(printResult, printProgress, "printed")}
              </Text>
              <Button
                size="small"
                icon={<FileSearchOutlined />}
                onClick={() => setIsPrintDetailsOpen(true)}
              >
                View Print Details
              </Button>
            </Flex>
          )}

        {/* Cancel Progress */}
        {cancelProgress.status === "processing" && (
          <Flex align="center" gap={8}>
            <Text type="secondary" style={{ fontSize: 12 }}>
              {getProgressText(cancelProgress, "Cancelling")}
            </Text>
            <Progress
              percent={Math.round(
                (cancelProgress.current / cancelProgress.total) * 100,
              )}
              size="small"
              status="active"
              style={{ flex: 1, margin: 0 }}
            />
          </Flex>
        )}
        {cancelProgress.status === "done" &&
          getResultText(cancelResult, cancelProgress, "cancelled") && (
            <Text type="success" style={{ fontSize: 12 }}>
              {getResultText(cancelResult, cancelProgress, "cancelled")}
            </Text>
          )}
      </Flex>

      <BulkPrintResultModal
        open={isPrintDetailsOpen}
        result={printResult}
        onClose={() => setIsPrintDetailsOpen(false)}
        onRetryFailed={onRetryFailedPrint}
        isRetrying={isPrinting}
      />
    </Card>
  );
}
