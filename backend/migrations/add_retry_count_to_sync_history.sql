-- Add retry_count column to marketplace_sync_history table
-- This column tracks how many times a failed sync operation has been retried
-- Used by auto_function_retry_syncs to prevent infinite retry loops

ALTER TABLE marketplace_sync_history 
ADD COLUMN IF NOT EXISTS retry_count INTEGER DEFAULT 0;

-- Add index for efficient querying of retryable failed syncs
CREATE INDEX IF NOT EXISTS idx_marketplace_sync_history_retry 
ON marketplace_sync_history(tenant_id, status, retry_count, created_at) 
WHERE status = 'failed' AND retry_count < 3;
