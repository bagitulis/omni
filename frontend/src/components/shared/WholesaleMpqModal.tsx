import { Alert, Button, Modal, Tabs, Typography } from "antd";
import { useMemo } from "react";
import { MpqTab } from "@/pages/inventory/components/MpqTab";
import { WholesaleTab } from "@/pages/inventory/components/WholesaleTab";
import { extractBulkPricingItems } from "@/pages/inventory/utils/bulkPricingItems";
import type { InventoryRecord } from "@/types/inventory";
import type { UnifiedProductRow } from "@/types/shared";

interface WholesaleMpqModalProps {
	open: boolean;
	onClose: () => void;
	selectedRecords: InventoryRecord[] | UnifiedProductRow[];
	defaultTab?: "wholesale" | "mpq";
}

/**
 * Adapter: Convert UnifiedProductRow[] to InventoryRecord[] shape
 * for compatibility with extractBulkPricingItems()
 */
function adaptUnifiedProductsToInventoryRecords(
	products: UnifiedProductRow[],
): InventoryRecord[] {
	const records: InventoryRecord[] = [];

	for (const product of products) {
		for (const sku of product.skus) {
			// Map platforms from platform_links to platform_status shape
			const platform_status = sku.platform_links.map((link) => ({
				platform: link.platform,
				platform_product_id: link.platform_product_id ?? "",
				platform_item_id: "",
				platform_sku: link.platform_sku_id ?? "",
				status: link.sync_status,
				stock: sku.stock,
				price: sku.price,
			}));

			records.push({
				id: `${product.id}-${sku.id}`,
				key_value: sku.seller_sku,
				key_column_name: "SKU",
				data: {
					price: sku.price,
					stock: sku.stock,
					variant: sku.variant_name,
				},
				platform_status,
				created_at: product.created_at,
				updated_at: product.updated_at,
			});
		}
	}

	return records;
}

/**
 * Shared WholesaleMpqModal for both Inventory and Products pages
 * - Supports InventoryRecord[] (legacy) and UnifiedProductRow[] (new unified products)
 * - Adapter converts UnifiedProductRow[] to InventoryRecord[] for extractBulkPricingItems
 * - defaultTab prop allows caller to control initial tab (wholesale | mpq)
 */
export function WholesaleMpqModal({
	open,
	onClose,
	selectedRecords,
	defaultTab = "wholesale",
}: WholesaleMpqModalProps) {
	// Convert UnifiedProductRow[] to InventoryRecord[] if needed
	const inventoryRecords: InventoryRecord[] = useMemo(() => {
		if (selectedRecords.length === 0) {
			return [];
		}

		// Type guard: check if first record is UnifiedProductRow
		const first = selectedRecords[0];
		if ("skus" in first && Array.isArray(first.skus)) {
			return adaptUnifiedProductsToInventoryRecords(
				selectedRecords as UnifiedProductRow[],
			);
		}

		// Otherwise, treat as InventoryRecord[]
		return selectedRecords as InventoryRecord[];
	}, [selectedRecords]);

	const { items: selectedItems, skipped_skus } = useMemo(
		() => extractBulkPricingItems(inventoryRecords),
		[inventoryRecords],
	);

	const items = [
		{
			key: "wholesale",
			label: "Wholesale",
			children: <WholesaleTab items={selectedItems} />,
		},
		{
			key: "mpq",
			label: "MPQ",
			children: <MpqTab items={selectedItems} />,
		},
	];

	return (
		<Modal
			title="Bulk Pricing"
			open={open}
			onCancel={onClose}
			destroyOnHidden
			width={900}
			footer={[
				<Button key="close" onClick={onClose}>
					Close
				</Button>,
			]}
			styles={{ body: { height: "600px", overflowY: "auto" } }}
			centered
		>
			{selectedRecords.length === 0 ? (
				<Alert
					type="warning"
					showIcon
					style={{ marginBottom: 12 }}
					message="Select at least one row first"
					description="Bulk Pricing uses SKU and price from the selected rows."
				/>
			) : null}

			{skipped_skus.length > 0 ? (
				<Alert
					type="info"
					showIcon
					style={{ marginBottom: 12 }}
					message={`${skipped_skus.length} SKU skipped`}
					description={`Missing/invalid price for: ${skipped_skus.slice(0, 5).join(", ")}${skipped_skus.length > 5 ? "..." : ""}`}
				/>
			) : null}

			<Typography.Text
				type="secondary"
				style={{ display: "block", marginBottom: 12 }}
			>
				Valid items: {selectedItems.length}
			</Typography.Text>

			<Tabs defaultActiveKey={defaultTab} items={items} />
		</Modal>
	);
}
