import { Modal } from "antd";
import { message } from "@/components/AntStaticHolder";
import { useState } from "react";
import type { BulkPrintLabelsOptions, CancelOrderParams } from "@/api/orders";
import { useOrderActions } from "@/hooks/useOrders";
import type { OrderListResponse } from "@/types/order";
import type { BulkResult, ProgressState } from "./bulkActionTypes";
import {
	askIncludeProductsOption,
	mergeUniqueOrderSns,
	runBulkPrint,
} from "./bulkPrintHelpers";
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

	const [lastPrintOptions, setLastPrintOptions] =
		useState<BulkPrintLabelsOptions>();

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
		} catch (err) { console.warn("Operation failed:", err);
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

		setPrintResult({ succeeded: [], failed: [] });
		const printOptions = buildBulkPrintOptions(platform, includeProducts);
		setLastPrintOptions(printOptions);
		await runBulkPrint(
			orderSns,
			printOptions,
			[],
			setPrintProgress,
			setPrintResult,
		);
	};

	const handleRetryFailedPrint = async () => {
		if (printResult.failed.length === 0) {
			message.info("No failed labels to retry");
			return;
		}

		const retryOrderSns = mergeUniqueOrderSns(
			printResult.failed.map((item) => item.order_sn),
		);

		await runBulkPrint(
			retryOrderSns,
			lastPrintOptions,
			printResult.succeeded,
			setPrintProgress,
			setPrintResult,
		);
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
				} catch (err) { console.warn("Operation failed:", err);
					setCancelProgress({ current: 0, total: 0, status: "idle" });
					message.error("Failed to process bulk cancellation");
				}
			},
		});
	};

	return {
		handleBulkShip,
		handleBulkPrint,
		handleRetryFailedPrint,
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
