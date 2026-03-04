import { message } from "antd";
import { importFromStaging } from "@/api/products";
import { syncSelectedProducts } from "@/api/productManager";
import { getErrorMessage } from "@/pages/products/utils/unifiedProductUtils";

/**
 * Execute sync marketplace for selected products:
 * 1. Sync from marketplace APIs → staging tables
 * 2. Import staging → master_products for each synced platform
 *
 * Shows per-platform results with raw API errors.
 * No false positives — only success if everything actually succeeded.
 */
export async function executeSyncMarketplace(
	productIds: number[],
	refreshProducts: () => Promise<void>,
	clearSelection: () => void,
): Promise<void> {
	// Step 1: Sync selected products from marketplace APIs → staging tables
	const result = await syncSelectedProducts(productIds);

	// Step 2: Parse which platforms were synced from result.details
	// Backend returns details like ["shopee: 3 synced", "tiktok: error msg"]
	const PLATFORMS = ["shopee", "tiktok", "lazada"] as const;
	const syncedPlatforms: Array<"shopee" | "tiktok" | "lazada"> = [];

	for (const platform of PLATFORMS) {
		const detail = result.details.find((d) =>
			d.toLowerCase().startsWith(platform),
		);
		if (detail && detail.includes("synced")) {
			syncedPlatforms.push(platform);
		}
	}

	// Log sync result per platform
	for (const detail of result.details) {
		if (
			detail.toLowerCase().includes("error") ||
			detail.toLowerCase().includes("fail")
		) {
			message.error(`Sync: ${detail}`);
		} else {
			message.info(`Sync: ${detail}`);
		}
	}

	if (result.failed > 0 && result.synced === 0) {
		message.error(
			`Sync failed for all ${result.failed} products. Check platform credentials.`,
		);
		return;
	}

	// Step 3: Import staging → master for each synced platform
	let importErrors = 0;
	for (const platform of syncedPlatforms) {
		try {
			const importResult = await importFromStaging(platform);
			message.success(
				`${platform}: imported ${importResult.products_created} new, ${importResult.products_matched} matched, ${importResult.skus_created} SKUs`,
			);
			if (importResult.errors.length > 0) {
				for (const err of importResult.errors) {
					message.warning(`${platform} import warning: ${err}`);
				}
			}
		} catch (importError) {
			importErrors++;
			message.error(
				`${platform} import failed: ${getErrorMessage(importError)}`,
			);
		}
	}

	if (importErrors === 0 && result.synced > 0) {
		message.success(
			`Sync complete: ${result.synced} products synced and imported`,
		);
	}

	await refreshProducts();
	clearSelection();
}
