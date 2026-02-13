import { useState } from "react";
import { message, Modal } from "antd";
import { useOrderActions } from "@/hooks/useOrders";
import type { OrderListResponse } from "@/types/order";
import type { CancelOrderParams } from "@/api/orders";
import { bulkPrintLabels } from "@/api/orders";
import { downloadOrderLabel } from "../utils/labelDownload";
import {
  buildBulkPrintOptions,
  shouldPromptTikTokPackingSlip,
} from "./printOptions";

interface UseOrderBulkActionsProps {
  selectedRowKeys: React.Key[];
  setSelectedRowKeys: (keys: React.Key[]) => void;
  data?: OrderListResponse;
  refetch: () => void;
  platform: string;
}

interface ProgressState {
  current: number;
  total: number;
  status: "idle" | "processing" | "done";
}

interface BulkResult {
  succeeded: string[];
  failed: Array<{ order_sn: string; error: string }>;
}

function askIncludeProductsOption(): Promise<boolean> {
  return new Promise((resolve) => {
    Modal.confirm({
      title: "Print option",
      content:
        "Include product list (packing slip) for TikTok labels? Choose 'With List' or 'Label Only'.",
      okText: "With List",
      cancelText: "Label Only",
      onOk: () => resolve(true),
      onCancel: () => resolve(false),
      centered: true,
    });
  });
}

export function useOrderBulkActions({
  selectedRowKeys,
  setSelectedRowKeys,
  data,
  refetch,
  platform,
}: UseOrderBulkActionsProps) {
  const { shipOrders, cancelOrder, isShipping, isCancelling } =
    useOrderActions();

  const [shipProgress, setShipProgress] = useState<ProgressState>({
    current: 0,
    total: 0,
    status: "idle",
  });

  const [printProgress, setPrintProgress] = useState<ProgressState>({
    current: 0,
    total: 0,
    status: "idle",
  });

  const [cancelProgress, setCancelProgress] = useState<ProgressState>({
    current: 0,
    total: 0,
    status: "idle",
  });

  const [shipResult, setShipResult] = useState<BulkResult>({
    succeeded: [],
    failed: [],
  });

  const [printResult, setPrintResult] = useState<BulkResult>({
    succeeded: [],
    failed: [],
  });

  const [cancelResult, setCancelResult] = useState<BulkResult>({
    succeeded: [],
    failed: [],
  });

  const handleBulkShip = async () => {
    if (selectedRowKeys.length === 0) return;

    const orderSns = selectedRowKeys as string[];
    setShipProgress({
      current: 0,
      total: orderSns.length,
      status: "processing",
    });
    setShipResult({ succeeded: [], failed: [] });

    const succeeded: string[] = [];
    const failed: Array<{ order_sn: string; error: string }> = [];

    try {
      const shipPlatform = platform !== "all" ? platform : undefined;

      // Process orders one by one for progress tracking
      for (let i = 0; i < orderSns.length; i++) {
        const orderSn = orderSns[i];
        try {
          await shipOrders([orderSn], shipPlatform);
          succeeded.push(orderSn);
        } catch (error) {
          failed.push({
            order_sn: orderSn,
            error: error instanceof Error ? error.message : "Unknown error",
          });
        }
        setShipProgress({
          current: i + 1,
          total: orderSns.length,
          status: "processing",
        });
      }

      setShipProgress({
        current: orderSns.length,
        total: orderSns.length,
        status: "done",
      });
      setShipResult({ succeeded, failed });

      // Show summary
      if (failed.length === 0) {
        message.success(`All ${succeeded.length} orders shipped successfully`);
      } else {
        message.warning(`${succeeded.length} shipped, ${failed.length} failed`);
      }

      setSelectedRowKeys([]);
      refetch();
    } catch {
      setShipProgress({ current: 0, total: 0, status: "idle" });
      message.error("Failed to process bulk ship");
    }
  };

  const handleBulkPrint = async () => {
    if (selectedRowKeys.length === 0) return;

    const orderSns = selectedRowKeys as string[];
    let includeProducts: boolean | undefined;
    if (shouldPromptTikTokPackingSlip(platform, data, orderSns)) {
      includeProducts = await askIncludeProductsOption();
    }

    setPrintProgress({
      current: 0,
      total: orderSns.length,
      status: "processing",
    });
    setPrintResult({ succeeded: [], failed: [] });

    try {
      const response = await bulkPrintLabels(
        orderSns,
        buildBulkPrintOptions(platform, includeProducts),
      );

      const succeeded = response.labels.map((label) => label.order_sn);
      const failed = response.failed.map((f) => ({
        order_sn: f.order_sn,
        error: f.error,
      }));

      setPrintProgress({
        current: orderSns.length,
        total: orderSns.length,
        status: "done",
      });
      setPrintResult({ succeeded, failed });

      // Download labels (base64 PDF or platform URL)
      if (response.labels.length > 0) {
        response.labels.forEach((label) => {
          downloadOrderLabel(label.file_data, label.order_sn);
        });
      }

      // Show summary
      if (failed.length === 0) {
        message.success(
          `All ${succeeded.length} labels generated successfully`,
        );
      } else {
        message.warning(
          `${succeeded.length} labels generated, ${failed.length} failed`,
        );
      }
    } catch (error) {
      setPrintProgress({ current: 0, total: 0, status: "idle" });
      message.error(
        error instanceof Error ? error.message : "Failed to print labels",
      );
    }
  };

  const handleBulkCancel = async () => {
    if (selectedRowKeys.length === 0) return;

    Modal.confirm({
      title: "Bulk Cancel Orders",
      content: `Are you sure you want to cancel ${selectedRowKeys.length} orders? This cannot be undone.`,
      okText: "Yes, Cancel All",
      okType: "danger",
      cancelText: "No",
      onOk: async () => {
        const orderSns = selectedRowKeys as string[];
        setCancelProgress({
          current: 0,
          total: orderSns.length,
          status: "processing",
        });
        setCancelResult({ succeeded: [], failed: [] });

        const succeeded: string[] = [];
        const failed: Array<{ order_sn: string; error: string }> = [];

        try {
          for (let i = 0; i < orderSns.length; i++) {
            const orderSn = orderSns[i];
            const order = data?.orders.find(
              (o) => (o.order_sn || o.order_no) === orderSn,
            );
            if (!order) {
              failed.push({ order_sn: orderSn, error: "Order not found" });
              continue;
            }

            try {
              const params: CancelOrderParams = {
                order_no: orderSn,
                platform: (order.platform || "shopee").toLowerCase(),
                cancel_reason: "out_of_stock",
                reason_detail: "Bulk cancellation via OMS",
              };

              await cancelOrder(params);
              succeeded.push(orderSn);
            } catch (e) {
              failed.push({
                order_sn: orderSn,
                error: e instanceof Error ? e.message : "Unknown error",
              });
            }

            setCancelProgress({
              current: i + 1,
              total: orderSns.length,
              status: "processing",
            });
          }

          setCancelProgress({
            current: orderSns.length,
            total: orderSns.length,
            status: "done",
          });
          setCancelResult({ succeeded, failed });

          // Show summary
          if (failed.length === 0) {
            message.success(
              `All ${succeeded.length} orders cancelled successfully`,
            );
          } else {
            message.warning(
              `${succeeded.length} cancelled, ${failed.length} failed`,
            );
          }

          setSelectedRowKeys([]);
          refetch();
        } catch {
          setCancelProgress({ current: 0, total: 0, status: "idle" });
          message.error("Failed to process bulk cancellation");
        }
      },
    });
  };

  return {
    handleBulkShip,
    handleBulkPrint,
    handleBulkCancel,
    isShipping,
    isPrinting: printProgress.status === "processing",
    isCancelling,
    shipProgress,
    printProgress,
    cancelProgress,
    shipResult,
    printResult,
    cancelResult,
  };
}
