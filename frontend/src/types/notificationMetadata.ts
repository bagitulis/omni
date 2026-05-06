/** Stored in notification.metadata JSONB field for bulk operations */
export interface BulkOperationMetadata {
  operation_type: "stock_sync" | "price_sync";
  total: number;
  succeeded: number;
  failed: number;
  platforms: Record<string, PlatformSyncStats>;
  failed_items?: FailedItemDetail[];
  request_id?: string;
}

export interface PlatformSyncStats {
  succeeded: number;
  failed: number;
  request_id?: string;
}

export interface FailedItemDetail {
  sku: string;
  platform: string;
  error: string;
  request_id?: string;
}
