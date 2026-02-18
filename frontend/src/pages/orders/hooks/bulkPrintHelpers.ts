import { Modal, message } from "antd";
import { type BulkPrintLabelsOptions, bulkPrintLabels } from "@/api/orders";
import { downloadOrderLabel } from "../utils/labelDownload";
import type {
	BulkActionFailure,
	BulkResult,
	ProgressState,
} from "./bulkActionTypes";

export function askIncludeProductsOption(): Promise<boolean> {
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

export function mergeUniqueOrderSns(items: string[]): string[] {
	return Array.from(new Set(items));
}

export async function runBulkPrint(
	orderSns: string[],
	options: BulkPrintLabelsOptions | undefined,
	previousSucceeded: string[],
	setPrintProgress: (state: ProgressState) => void,
	setPrintResult: (result: BulkResult) => void,
): Promise<void> {
	setPrintProgress({
		current: 0,
		total: orderSns.length,
		status: "processing",
	});

	try {
		const response = await bulkPrintLabels(orderSns, options);

		const succeeded = mergeUniqueOrderSns([
			...previousSucceeded,
			...response.labels.map((label) => label.order_sn),
		]);
		const failed: BulkActionFailure[] = response.failed.map((f) => ({
			order_sn: f.order_sn,
			error: f.error,
		}));

		if (response.labels.length > 0) {
			response.labels.forEach((label) => {
				try {
					downloadOrderLabel(label.file_data, label.order_sn);
				} catch (error) {
					failed.push({
						order_sn: label.order_sn,
						error:
							error instanceof Error
								? error.message
								: "Failed to trigger browser download",
					});
				}
			});
		}

		setPrintProgress({
			current: orderSns.length,
			total: orderSns.length,
			status: "done",
		});
		setPrintResult({ succeeded, failed });

		if (failed.length === 0) {
			message.success(`All ${succeeded.length} labels downloaded successfully`);
		} else {
			message.warning(
				`${succeeded.length} labels downloaded, ${failed.length} failed. Open Print Details for raw errors.`,
			);
		}
	} catch (error) {
		setPrintProgress({ current: 0, total: 0, status: "idle" });
		message.error(
			error instanceof Error ? error.message : "Failed to print labels",
		);
	}
}
