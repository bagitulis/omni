import { useState, useMemo } from "react";
import { Alert, Button, Empty, Popconfirm, Typography } from "antd";
import { batchDeleteBySkus, batchShopeeMpq } from "@/api/wholesale";
import type { BulkPricingItem } from "../utils/bulkPricingItems";
import { logger } from "@/lib/logger";

interface ResetPricingTabProps {
	items: BulkPricingItem[];
}

export function ResetPricingTab({ items }: ResetPricingTabProps) {
	const [processing, setProcessing] = useState(false);
	const [resultMessage, setResultMessage] = useState<string | null>(null);
	const [resultType, setResultType] = useState<
		"success" | "warning" | "error" | null
	>(null);

	const shopeeItems = useMemo(() => {
		const unique = new Map<string, number>();
		for (const item of items) {
			if (item.platform === "shopee" && !unique.has(item.sku)) {
				unique.set(item.sku, item.price);
			}
		}
		return Array.from(unique.entries()).map(([sku, price]) => ({ sku, price }));
	}, [items]);

	const tiktokItems = useMemo(() => {
		const unique = new Map<string, number>();
		for (const item of items) {
			if (item.platform === "tiktok" && !unique.has(item.sku)) {
				unique.set(item.sku, item.price);
			}
		}
		return Array.from(unique.entries()).map(([sku, price]) => ({ sku, price }));
	}, [items]);

	const handleReset = async () => {
		if (shopeeItems.length === 0 && tiktokItems.length === 0) return;

		setProcessing(true);
		setResultMessage(null);
		setResultType(null);

		const results: string[] = [];
		let hasError = false;

		try {
			// Step 1: Delete wholesale tiers (Shopee only)
			if (shopeeItems.length > 0) {
				const deleteResult = await batchDeleteBySkus(
					shopeeItems.map((i) => i.sku),
				);
				results.push(
					`Wholesale: ${deleteResult.data.processed} cleared, ${deleteResult.data.failed} failed`,
				);
				if (deleteResult.data.failed > 0) hasError = true;
			}

			// Step 2: Reset MPQ to 1 (Shopee)
			if (shopeeItems.length > 0) {
				try {
					await batchShopeeMpq(shopeeItems, 1);
					results.push(`Shopee MPQ: reset to 1`);
				} catch (err) { logger.warn("Operation failed:", { err: err });
					results.push(`Shopee MPQ: failed to reset`);
					hasError = true;
				}
			}

			const summary = results.join(" | ");
			setResultMessage(summary);
			setResultType(hasError ? "warning" : "success");
		} catch (error) {
			const msg =
				error instanceof Error ? error.message : "Failed to reset pricing";
			setResultType("error");
			setResultMessage(msg);
		} finally {
			setProcessing(false);
		}
	};

	const totalItems = shopeeItems.length + tiktokItems.length;

	return (
		<div style={{ display: "flex", flexDirection: "column", gap: 16 }}>
			<Alert
				type="warning"
				showIcon
				message="Reset selected items to clean state"
				description={
					<>
						<div>This will:</div>
						<ul style={{ margin: "4px 0", paddingLeft: 20 }}>
							<li>Delete all wholesale tiers (Shopee)</li>
							<li>Reset MPQ to 1 (Shopee)</li>
						</ul>
						<div>
							Items affected: <strong>{shopeeItems.length}</strong> Shopee
							{tiktokItems.length > 0 && (
								<>
									{" "}| <strong>{tiktokItems.length}</strong> TikTok (MPQ only)
								</>
							)}
						</div>
					</>
				}
			/>

			{resultMessage && resultType ? (
				<Alert
					type={resultType}
					showIcon
					message={resultMessage}
					closable
					onClose={() => {
						setResultMessage(null);
						setResultType(null);
					}}
				/>
			) : null}

			{totalItems > 0 ? (
				<Popconfirm
					title="Reset pricing for selected items?"
					description={`This will clear wholesale tiers and reset MPQ to 1 for ${shopeeItems.length} Shopee items. This cannot be undone.`}
					onConfirm={handleReset}
					okText="Yes, Reset"
					okButtonProps={{ danger: true }}
				>
					<Button
						type="primary"
						danger
						loading={processing}
						style={{ alignSelf: "flex-start" }}
					>
						Reset Pricing ({totalItems} items)
					</Button>
				</Popconfirm>
			) : (
				<Empty description="No items available for reset" />
			)}

			{shopeeItems.length > 0 ? (
				<Typography.Text type="secondary">
					Preview: {shopeeItems.slice(0, 5).map((i) => i.sku).join(", ")}
					{shopeeItems.length > 5
						? ` ... +${shopeeItems.length - 5} more`
						: ""}
				</Typography.Text>
			) : null}
		</div>
	);
}
