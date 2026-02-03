-- Fix tiktok_order_items with product_id = 0
-- This script updates product_id by matching seller_sku with tiktok_skus table

BEGIN;

-- Update items where product_id is 0 and we can match by seller_sku = sku_id
UPDATE tiktok_order_items toi
SET product_id = ts.product_id
FROM tiktok_skus ts
WHERE toi.product_id = 0
  AND toi.seller_sku = ts.sku_id
  AND toi.tenant_id = ts.tenant_id;

-- Update items where product_id is 0 and we can match by seller_sku = seller_sku
UPDATE tiktok_order_items toi
SET product_id = ts.product_id
FROM tiktok_skus ts
WHERE toi.product_id = 0
  AND toi.seller_sku = ts.seller_sku
  AND toi.tenant_id = ts.tenant_id;

-- Log the number of updated rows (optional, depending on tool support)
-- SELECT count(*) as fixed_items FROM tiktok_order_items WHERE product_id > 0;

COMMIT;
