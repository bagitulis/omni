-- CreateTable
CREATE TABLE "User" (
    "id" TEXT NOT NULL PRIMARY KEY,
    "username" TEXT NOT NULL,
    "email" TEXT NOT NULL,
    "password" TEXT NOT NULL,
    "role" TEXT NOT NULL DEFAULT 'owner',
    "failedLoginAttempts" INTEGER NOT NULL DEFAULT 0,
    "accountLockedUntil" DATETIME,
    "lastFailedLogin" DATETIME,
    "createdAt" DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updatedAt" DATETIME NOT NULL
);

-- CreateTable
CREATE TABLE "PlatformConfig" (
    "id" TEXT NOT NULL PRIMARY KEY,
    "platform" TEXT NOT NULL,
    "configKey" TEXT NOT NULL,
    "configValue" TEXT NOT NULL,
    "dataType" TEXT NOT NULL DEFAULT 'string',
    "isEncrypted" BOOLEAN NOT NULL DEFAULT true,
    "metadata" TEXT,
    "createdAt" DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updatedAt" DATETIME NOT NULL
);

-- CreateTable
CREATE TABLE "ShopeeOrder" (
    "id" INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
    "tenantId" TEXT NOT NULL,
    "orderSn" TEXT NOT NULL,
    "shopId" BIGINT,
    "orderStatus" TEXT NOT NULL,
    "orderTimestamp" INTEGER,
    "createdAt" DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updatedAt" DATETIME NOT NULL
);

-- CreateTable
CREATE TABLE "ShopeeOrderItem" (
    "id" INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
    "orderSn" TEXT NOT NULL,
    "itemId" BIGINT NOT NULL,
    "modelId" BIGINT,
    "itemName" TEXT,
    "modelName" TEXT,
    "itemSku" TEXT,
    "modelSku" TEXT,
    "quantity" INTEGER,
    "price" REAL,
    "createdAt" DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updatedAt" DATETIME NOT NULL,
    CONSTRAINT "ShopeeOrderItem_orderSn_fkey" FOREIGN KEY ("orderSn") REFERENCES "ShopeeOrder" ("orderSn") ON DELETE CASCADE ON UPDATE CASCADE
);

-- CreateTable
CREATE TABLE "LazadaOrder" (
    "id" INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
    "tenantId" TEXT NOT NULL,
    "orderSn" TEXT NOT NULL,
    "shopId" BIGINT,
    "orderStatus" TEXT NOT NULL,
    "createdAt" DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updatedAt" DATETIME NOT NULL
);

-- CreateTable
CREATE TABLE "LazadaOrderItem" (
    "id" INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
    "orderSn" TEXT NOT NULL,
    "itemId" BIGINT NOT NULL,
    "skuId" TEXT,
    "sellerSku" TEXT,
    "productName" TEXT,
    "variationName" TEXT,
    "quantity" INTEGER,
    "price" REAL,
    "createdAt" DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updatedAt" DATETIME NOT NULL,
    CONSTRAINT "LazadaOrderItem_orderSn_fkey" FOREIGN KEY ("orderSn") REFERENCES "LazadaOrder" ("orderSn") ON DELETE CASCADE ON UPDATE CASCADE
);

-- CreateTable
CREATE TABLE "TiktokOrder" (
    "id" INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
    "tenantId" TEXT NOT NULL,
    "orderSn" TEXT NOT NULL,
    "shopId" BIGINT,
    "orderStatus" TEXT NOT NULL,
    "createdAt" DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updatedAt" DATETIME NOT NULL
);

-- CreateTable
CREATE TABLE "TiktokOrderItem" (
    "id" INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
    "orderSn" TEXT NOT NULL,
    "lineItemId" TEXT,
    "productId" BIGINT NOT NULL,
    "skuId" TEXT,
    "sellerSku" TEXT,
    "productName" TEXT,
    "variationName" TEXT,
    "quantity" INTEGER,
    "price" REAL,
    "createdAt" DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updatedAt" DATETIME NOT NULL,
    CONSTRAINT "TiktokOrderItem_orderSn_fkey" FOREIGN KEY ("orderSn") REFERENCES "TiktokOrder" ("orderSn") ON DELETE CASCADE ON UPDATE CASCADE
);

-- CreateTable
CREATE TABLE "OrderTodayItem" (
    "id" INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
    "tenantId" TEXT NOT NULL,
    "platform" TEXT NOT NULL,
    "orderSn" TEXT NOT NULL,
    "trackingNo" TEXT,
    "courier" TEXT,
    "sellerSku" TEXT,
    "productName" TEXT,
    "variationName" TEXT,
    "quantity" INTEGER NOT NULL DEFAULT 1,
    "syncedAt" DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "createdAt" DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updatedAt" DATETIME NOT NULL
);

-- CreateTable
CREATE TABLE "GoogleSheetsSettings" (
    "id" TEXT NOT NULL PRIMARY KEY,
    "spreadsheetId" TEXT,
    "selectedSheet" TEXT,
    "manualMode" BOOLEAN NOT NULL DEFAULT false,
    "walletSpreadsheetId" TEXT,
    "shippingSpreadsheetId" TEXT,
    "inventorySpreadsheetId" TEXT,
    "inventorySheetName" TEXT,
    "inventorySelectedColumns" TEXT,
    "inventoryAvailableWorksheets" TEXT,
    "walletAvailableWorksheets" TEXT,
    "shippingAvailableWorksheets" TEXT,
    "orderAvailableWorksheets" TEXT,
    "orderSpreadsheetId" TEXT,
    "createdAt" DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updatedAt" DATETIME NOT NULL,
    "availableSpreadsheets" TEXT,
    "availableWorksheets" TEXT
);

-- CreateTable
CREATE TABLE "InventorySettings" (
    "id" TEXT NOT NULL PRIMARY KEY,
    "tenantId" TEXT NOT NULL,
    "spreadsheetId" TEXT,
    "sheetName" TEXT,
    "selectedColumns" TEXT,
    "headerRow" INTEGER NOT NULL DEFAULT 1,
    "dataStartRow" INTEGER NOT NULL DEFAULT 2,
    "keyColumn" TEXT,
    "autoSync" BOOLEAN NOT NULL DEFAULT false,
    "syncIntervalSeconds" INTEGER NOT NULL DEFAULT 300,
    "lastSyncTimestamp" DATETIME,
    "createdAt" DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updatedAt" DATETIME NOT NULL,
    "allColumns" TEXT,
    "lastHeadersHash" TEXT,
    "lastSyncStatus" TEXT
);

-- CreateTable
CREATE TABLE "InventoryRecord" (
    "id" TEXT NOT NULL PRIMARY KEY,
    "tenantId" TEXT NOT NULL,
    "data" TEXT NOT NULL,
    "keyValue" TEXT NOT NULL,
    "keyColumnName" TEXT NOT NULL,
    "createdAt" DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updatedAt" DATETIME NOT NULL
);

-- CreateTable
CREATE TABLE "InventorySyncHistory" (
    "id" TEXT NOT NULL PRIMARY KEY,
    "tenantId" TEXT NOT NULL,
    "syncDirection" TEXT NOT NULL,
    "status" TEXT NOT NULL,
    "totalRecords" INTEGER NOT NULL DEFAULT 0,
    "syncedRecords" INTEGER NOT NULL DEFAULT 0,
    "failedRecords" INTEGER NOT NULL DEFAULT 0,
    "duration" REAL NOT NULL DEFAULT 0,
    "message" TEXT,
    "createdAt" DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- CreateTable
CREATE TABLE "SheetSnapshot" (
    "id" TEXT NOT NULL PRIMARY KEY,
    "tenantId" TEXT NOT NULL,
    "sku" TEXT NOT NULL,
    "snapshotData" TEXT NOT NULL,
    "rowIndex" INTEGER NOT NULL,
    "dataHash" TEXT,
    "syncedAt" DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "createdAt" DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updatedAt" DATETIME NOT NULL
);

-- CreateTable
CREATE TABLE "LazadaProduct" (
    "id" INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
    "tenantId" TEXT NOT NULL,
    "itemId" TEXT NOT NULL,
    "name" TEXT NOT NULL,
    "description" TEXT,
    "status" TEXT NOT NULL,
    "price" REAL NOT NULL DEFAULT 0,
    "quantity" INTEGER NOT NULL DEFAULT 0,
    "image" TEXT,
    "createdAt" DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updatedAt" DATETIME NOT NULL
);

-- CreateTable
CREATE TABLE "LazadaSku" (
    "id" INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
    "productId" INTEGER NOT NULL,
    "skuId" TEXT NOT NULL,
    "sellerSku" TEXT,
    "price" REAL NOT NULL DEFAULT 0,
    "quantity" INTEGER NOT NULL DEFAULT 0,
    "createdAt" DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updatedAt" DATETIME NOT NULL,
    "variantData" TEXT,
    "variantName" TEXT,
    CONSTRAINT "LazadaSku_productId_fkey" FOREIGN KEY ("productId") REFERENCES "LazadaProduct" ("id") ON DELETE CASCADE ON UPDATE CASCADE
);

-- CreateTable
CREATE TABLE "ShopeeProduct" (
    "id" INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
    "tenantId" TEXT NOT NULL,
    "itemId" BIGINT NOT NULL,
    "name" TEXT NOT NULL,
    "description" TEXT,
    "status" TEXT NOT NULL,
    "price" REAL NOT NULL DEFAULT 0,
    "quantity" INTEGER NOT NULL DEFAULT 0,
    "image" TEXT,
    "createdAt" DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updatedAt" DATETIME NOT NULL
);

-- CreateTable
CREATE TABLE "ShopeeSku" (
    "id" INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
    "productId" INTEGER NOT NULL,
    "itemId" BIGINT NOT NULL,
    "modelId" BIGINT,
    "sellerSku" TEXT,
    "variantName" TEXT,
    "variantData" TEXT,
    "price" REAL NOT NULL DEFAULT 0,
    "quantity" INTEGER NOT NULL DEFAULT 0,
    "createdAt" DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updatedAt" DATETIME NOT NULL,
    CONSTRAINT "ShopeeSku_productId_fkey" FOREIGN KEY ("productId") REFERENCES "ShopeeProduct" ("id") ON DELETE CASCADE ON UPDATE CASCADE
);

-- CreateTable
CREATE TABLE "TiktokProduct" (
    "id" INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
    "tenantId" TEXT NOT NULL,
    "productId" TEXT NOT NULL,
    "name" TEXT NOT NULL,
    "description" TEXT,
    "status" TEXT NOT NULL,
    "price" REAL NOT NULL DEFAULT 0,
    "quantity" INTEGER NOT NULL DEFAULT 0,
    "image" TEXT,
    "createdAt" DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updatedAt" DATETIME NOT NULL
);

-- CreateTable
CREATE TABLE "TiktokSku" (
    "id" INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
    "productId" INTEGER NOT NULL,
    "skuId" TEXT NOT NULL,
    "price" REAL NOT NULL DEFAULT 0,
    "quantity" INTEGER NOT NULL DEFAULT 0,
    "createdAt" DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updatedAt" DATETIME NOT NULL,
    "sellerSku" TEXT,
    "variantData" TEXT,
    "variantName" TEXT,
    CONSTRAINT "TiktokSku_productId_fkey" FOREIGN KEY ("productId") REFERENCES "TiktokProduct" ("id") ON DELETE CASCADE ON UPDATE CASCADE
);

-- CreateTable
CREATE TABLE "FilterPreference" (
    "id" TEXT NOT NULL PRIMARY KEY,
    "tenantId" TEXT NOT NULL,
    "platform" TEXT NOT NULL,
    "tab" TEXT NOT NULL DEFAULT 'master',
    "visibleColumns" TEXT NOT NULL,
    "columnFilters" TEXT NOT NULL,
    "searchQuery" TEXT NOT NULL DEFAULT '',
    "lockedColumns" TEXT NOT NULL DEFAULT '[]',
    "createdAt" DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updatedAt" DATETIME NOT NULL
);

-- CreateTable
CREATE TABLE "InventorySkuPlatformStatus" (
    "id" TEXT NOT NULL PRIMARY KEY,
    "tenantId" TEXT NOT NULL,
    "sku" TEXT NOT NULL,
    "lazada" BOOLEAN NOT NULL DEFAULT false,
    "shopee" BOOLEAN NOT NULL DEFAULT false,
    "tiktok" BOOLEAN NOT NULL DEFAULT false,
    "checkedAt" DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updatedAt" DATETIME NOT NULL
);

-- CreateTable
CREATE TABLE "RouteConfig" (
    "id" TEXT NOT NULL PRIMARY KEY,
    "tenantId" TEXT NOT NULL,
    "routePath" TEXT NOT NULL,
    "routeMethod" TEXT NOT NULL DEFAULT 'GET',
    "routeName" TEXT,
    "description" TEXT,
    "category" TEXT,
    "enabled" BOOLEAN NOT NULL DEFAULT true,
    "cachingEnabled" BOOLEAN NOT NULL DEFAULT true,
    "cacheTTL" INTEGER NOT NULL DEFAULT 300,
    "cacheStrategy" TEXT NOT NULL DEFAULT 'standard',
    "cacheKeyTemplate" TEXT,
    "queueEnabled" BOOLEAN NOT NULL DEFAULT true,
    "queueMaxSize" INTEGER NOT NULL DEFAULT 100,
    "queuePriority" TEXT NOT NULL DEFAULT 'normal',
    "maxConcurrent" INTEGER NOT NULL DEFAULT 5,
    "rateLimitEnabled" BOOLEAN NOT NULL DEFAULT false,
    "rateLimitWindow" INTEGER NOT NULL DEFAULT 60,
    "rateLimitMax" INTEGER NOT NULL DEFAULT 100,
    "minIntervalMs" INTEGER NOT NULL DEFAULT 0,
    "timeout" INTEGER NOT NULL DEFAULT 30000,
    "retryEnabled" BOOLEAN NOT NULL DEFAULT true,
    "maxRetries" INTEGER NOT NULL DEFAULT 3,
    "retryDelayMs" INTEGER NOT NULL DEFAULT 1000,
    "lastExecuted" DATETIME,
    "executionCount" INTEGER NOT NULL DEFAULT 0,
    "failureCount" INTEGER NOT NULL DEFAULT 0,
    "avgExecutionTimeMs" REAL NOT NULL DEFAULT 0,
    "customConfig" TEXT,
    "createdAt" DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updatedAt" DATETIME NOT NULL
);

-- CreateTable
CREATE TABLE "Spreadsheet" (
    "id" TEXT NOT NULL PRIMARY KEY,
    "spreadsheetId" TEXT NOT NULL,
    "spreadsheetUrl" TEXT NOT NULL,
    "spreadsheetName" TEXT NOT NULL,
    "purpose" TEXT NOT NULL DEFAULT 'other',
    "sheets" TEXT,
    "tenantId" TEXT NOT NULL,
    "registeredBy" TEXT,
    "lastUsedAt" DATETIME,
    "createdAt" DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updatedAt" DATETIME NOT NULL
);

-- CreateTable
CREATE TABLE "LockedOrder" (
    "id" TEXT NOT NULL PRIMARY KEY,
    "tenantId" TEXT NOT NULL,
    "sku" TEXT NOT NULL,
    "productName" TEXT NOT NULL,
    "variationName" TEXT,
    "qty" INTEGER NOT NULL,
    "createdAt" DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updatedAt" DATETIME NOT NULL
);

-- CreateTable
CREATE TABLE "WholesaleSettings" (
    "id" TEXT NOT NULL PRIMARY KEY,
    "tenantId" TEXT NOT NULL,
    "platform" TEXT NOT NULL DEFAULT 'shopee',
    "adminFee" INTEGER NOT NULL DEFAULT 1500,
    "maxOrderTier3" INTEGER NOT NULL DEFAULT 1000,
    "minOrder1" INTEGER NOT NULL DEFAULT 2,
    "maxOrder1" INTEGER NOT NULL DEFAULT 3,
    "createdAt" DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updatedAt" DATETIME NOT NULL
);

-- CreateTable
CREATE TABLE "ShopeeEscrowSync" (
    "id" TEXT NOT NULL PRIMARY KEY,
    "tenantId" TEXT NOT NULL,
    "month" INTEGER NOT NULL,
    "year" INTEGER NOT NULL,
    "totalOrders" INTEGER NOT NULL DEFAULT 0,
    "syncedAt" DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- CreateTable
CREATE TABLE "ShopeeEscrowOrder" (
    "id" TEXT NOT NULL PRIMARY KEY,
    "tenantId" TEXT NOT NULL,
    "orderSn" TEXT NOT NULL,
    "month" INTEGER NOT NULL,
    "year" INTEGER NOT NULL,
    "orderDate" DATETIME,
    "buyerUserName" TEXT,
    "escrowAmount" REAL NOT NULL DEFAULT 0,
    "commissionFee" REAL NOT NULL DEFAULT 0,
    "serviceFee" REAL NOT NULL DEFAULT 0,
    "sellerProcessingFee" REAL NOT NULL DEFAULT 0,
    "buyerPaidShippingFee" REAL NOT NULL DEFAULT 0,
    "actualShippingFee" REAL NOT NULL DEFAULT 0,
    "shopeeShippingRebate" REAL NOT NULL DEFAULT 0,
    "estimatedShippingFee" REAL NOT NULL DEFAULT 0,
    "buyerTotalAmount" REAL NOT NULL DEFAULT 0,
    "buyerPaymentMethod" TEXT,
    "rawOrderIncome" TEXT,
    "rawBuyerPaymentInfo" TEXT,
    "syncedAt" DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "createdAt" DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updatedAt" DATETIME NOT NULL
);

-- CreateTable
CREATE TABLE "ShopeeEscrowItem" (
    "id" TEXT NOT NULL PRIMARY KEY,
    "tenantId" TEXT NOT NULL,
    "escrowOrderId" TEXT NOT NULL,
    "orderId" TEXT,
    "orderSn" TEXT,
    "month" INTEGER,
    "year" INTEGER,
    "itemId" BIGINT,
    "modelId" BIGINT,
    "sku" TEXT,
    "modelSku" TEXT,
    "itemName" TEXT,
    "modelName" TEXT,
    "quantity" INTEGER NOT NULL DEFAULT 0,
    "originalPrice" REAL NOT NULL DEFAULT 0,
    "sellingPrice" REAL NOT NULL DEFAULT 0,
    "discountedPrice" REAL NOT NULL DEFAULT 0,
    "sellerDiscount" REAL NOT NULL DEFAULT 0,
    "shopeeDiscount" REAL NOT NULL DEFAULT 0,
    "discountFromCoin" REAL NOT NULL DEFAULT 0,
    "discountFromVoucherSeller" REAL NOT NULL DEFAULT 0,
    "discountFromVoucherShopee" REAL NOT NULL DEFAULT 0,
    "amsCommissionFee" REAL NOT NULL DEFAULT 0,
    "sellerOrderProcessingFee" REAL NOT NULL DEFAULT 0,
    "rawItemData" TEXT,
    "createdAt" DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updatedAt" DATETIME NOT NULL,
    CONSTRAINT "ShopeeEscrowItem_escrowOrderId_fkey" FOREIGN KEY ("escrowOrderId") REFERENCES "ShopeeEscrowOrder" ("id") ON DELETE CASCADE ON UPDATE CASCADE
);

-- CreateTable
CREATE TABLE "AnalyticsSettings" (
    "id" TEXT NOT NULL PRIMARY KEY,
    "tenantId" TEXT NOT NULL,
    "platform" TEXT NOT NULL DEFAULT 'shopee',
    "priceColumn" TEXT NOT NULL DEFAULT 'HARGA',
    "formulaDeduction" REAL NOT NULL DEFAULT 1500,
    "formulaMultiplier" REAL NOT NULL DEFAULT 0.84,
    "createdAt" DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updatedAt" DATETIME NOT NULL
);

-- CreateTable
CREATE TABLE "TiktokEscrowSync" (
    "id" TEXT NOT NULL PRIMARY KEY,
    "tenantId" TEXT NOT NULL,
    "month" INTEGER NOT NULL,
    "year" INTEGER NOT NULL,
    "totalOrders" INTEGER NOT NULL DEFAULT 0,
    "syncedAt" DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "createdAt" DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updatedAt" DATETIME NOT NULL
);

-- CreateTable
CREATE TABLE "TiktokEscrowOrder" (
    "id" TEXT NOT NULL PRIMARY KEY,
    "tenantId" TEXT NOT NULL,
    "orderId" TEXT NOT NULL,
    "month" INTEGER NOT NULL,
    "year" INTEGER NOT NULL,
    "orderStatus" TEXT,
    "orderDate" DATETIME,
    "buyerName" TEXT,
    "transactionId" TEXT,
    "transactionType" TEXT,
    "statementTime" DATETIME,
    "totalSettlementAmount" REAL NOT NULL DEFAULT 0,
    "productRevenue" REAL NOT NULL DEFAULT 0,
    "platformCommission" REAL NOT NULL DEFAULT 0,
    "transactionFee" REAL NOT NULL DEFAULT 0,
    "shippingFeeCustomerPaid" REAL NOT NULL DEFAULT 0,
    "shippingFeeActual" REAL NOT NULL DEFAULT 0,
    "shippingFeePlatformDiscount" REAL NOT NULL DEFAULT 0,
    "sellerShippingDiscount" REAL NOT NULL DEFAULT 0,
    "refundAmount" REAL NOT NULL DEFAULT 0,
    "adjustment" REAL NOT NULL DEFAULT 0,
    "buyerTotalAmount" REAL NOT NULL DEFAULT 0,
    "currency" TEXT NOT NULL DEFAULT 'IDR',
    "rawTransactionData" TEXT,
    "rawOrderData" TEXT,
    "syncedAt" DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "createdAt" DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updatedAt" DATETIME NOT NULL
);

-- CreateTable
CREATE TABLE "TiktokEscrowItem" (
    "id" TEXT NOT NULL PRIMARY KEY,
    "tenantId" TEXT NOT NULL,
    "escrowOrderId" TEXT NOT NULL,
    "orderId" TEXT NOT NULL,
    "productId" TEXT,
    "productName" TEXT,
    "skuId" TEXT,
    "sellerSku" TEXT,
    "quantity" INTEGER NOT NULL DEFAULT 0,
    "salePrice" REAL NOT NULL DEFAULT 0,
    "originalPrice" REAL NOT NULL DEFAULT 0,
    "subtotalAfterSellerDiscount" REAL NOT NULL DEFAULT 0,
    "platformDiscount" REAL NOT NULL DEFAULT 0,
    "sellerDiscount" REAL NOT NULL DEFAULT 0,
    "commission" REAL NOT NULL DEFAULT 0,
    "transactionFeeItem" REAL NOT NULL DEFAULT 0,
    "settlementAmount" REAL NOT NULL DEFAULT 0,
    "rawItemData" TEXT,
    "syncedAt" DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "createdAt" DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updatedAt" DATETIME NOT NULL,
    CONSTRAINT "TiktokEscrowItem_escrowOrderId_fkey" FOREIGN KEY ("escrowOrderId") REFERENCES "TiktokEscrowOrder" ("id") ON DELETE CASCADE ON UPDATE CASCADE
);

-- CreateTable
CREATE TABLE "OAuthState" (
    "id" TEXT NOT NULL PRIMARY KEY,
    "tenantId" TEXT NOT NULL,
    "platform" TEXT NOT NULL,
    "state" TEXT NOT NULL,
    "redirectUrl" TEXT,
    "metadata" TEXT,
    "expiresAt" DATETIME NOT NULL,
    "createdAt" DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- CreateTable
CREATE TABLE "OAuthLog" (
    "id" TEXT NOT NULL PRIMARY KEY,
    "tenantId" TEXT NOT NULL,
    "platform" TEXT NOT NULL,
    "eventType" TEXT NOT NULL,
    "shopId" TEXT,
    "code" TEXT,
    "state" TEXT,
    "status" TEXT NOT NULL DEFAULT 'received',
    "errorMsg" TEXT,
    "metadata" TEXT,
    "processedAt" DATETIME,
    "createdAt" DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- CreateTable
CREATE TABLE "WebhookLog" (
    "id" TEXT NOT NULL PRIMARY KEY,
    "tenantId" TEXT,
    "platform" TEXT NOT NULL,
    "eventType" TEXT NOT NULL,
    "payload" TEXT NOT NULL,
    "headers" TEXT,
    "status" TEXT NOT NULL DEFAULT 'received',
    "errorMsg" TEXT,
    "processedAt" DATETIME,
    "createdAt" DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- CreateTable
CREATE TABLE "WebhookOrderEvent" (
    "id" INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
    "tenantId" TEXT NOT NULL,
    "platform" TEXT NOT NULL,
    "eventType" TEXT NOT NULL,
    "orderSn" TEXT NOT NULL,
    "shopId" TEXT,
    "oldStatus" TEXT,
    "newStatus" TEXT,
    "fulfillmentStatus" TEXT,
    "packageNumber" TEXT,
    "payload" TEXT,
    "webhookLogId" TEXT,
    "processedAt" DATETIME,
    "createdAt" DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- CreateTable
CREATE TABLE "WebhookProductEvent" (
    "id" INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
    "tenantId" TEXT NOT NULL,
    "eventType" TEXT NOT NULL,
    "shopId" TEXT,
    "itemId" TEXT,
    "variationId" TEXT,
    "action" TEXT,
    "payload" TEXT,
    "processedAt" DATETIME,
    "createdAt" DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- CreateTable
CREATE TABLE "WebhookMarketingEvent" (
    "id" INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
    "tenantId" TEXT NOT NULL,
    "eventType" TEXT NOT NULL,
    "shopId" TEXT,
    "itemId" TEXT,
    "promotionId" TEXT,
    "promotionType" TEXT,
    "action" TEXT,
    "payload" TEXT,
    "processedAt" DATETIME,
    "createdAt" DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- CreateTable
CREATE TABLE "WebhookReturnEvent" (
    "id" INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
    "tenantId" TEXT NOT NULL,
    "eventType" TEXT NOT NULL,
    "shopId" TEXT,
    "orderSn" TEXT,
    "returnSn" TEXT,
    "status" TEXT,
    "reason" TEXT,
    "payload" TEXT,
    "processedAt" DATETIME,
    "createdAt" DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- CreateTable
CREATE TABLE "WebhookShopeeEvent" (
    "id" INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
    "tenantId" TEXT NOT NULL,
    "eventType" TEXT NOT NULL,
    "shopId" TEXT,
    "action" TEXT,
    "expiryTime" DATETIME,
    "penaltyPoints" INTEGER,
    "payload" TEXT,
    "processedAt" DATETIME,
    "createdAt" DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- CreateTable
CREATE TABLE "WebhookWebchatEvent" (
    "id" INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
    "tenantId" TEXT NOT NULL,
    "eventType" TEXT NOT NULL,
    "shopId" TEXT,
    "conversationId" TEXT,
    "messageType" TEXT,
    "senderId" TEXT,
    "payload" TEXT,
    "processedAt" DATETIME,
    "createdAt" DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- CreateTable
CREATE TABLE "WebhookFBSEvent" (
    "id" INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
    "tenantId" TEXT NOT NULL,
    "eventType" TEXT NOT NULL,
    "shopId" TEXT,
    "itemId" TEXT,
    "skuId" TEXT,
    "stockChange" INTEGER,
    "invoiceNumber" TEXT,
    "payload" TEXT,
    "processedAt" DATETIME,
    "createdAt" DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- CreateTable
CREATE TABLE "TiktokAdsUploadBatch" (
    "id" TEXT NOT NULL PRIMARY KEY,
    "tenantId" TEXT NOT NULL,
    "fileName" TEXT NOT NULL,
    "periodStart" DATETIME NOT NULL,
    "periodEnd" DATETIME NOT NULL,
    "totalRows" INTEGER NOT NULL DEFAULT 0,
    "insertedRows" INTEGER NOT NULL DEFAULT 0,
    "skippedRows" INTEGER NOT NULL DEFAULT 0,
    "updatedRows" INTEGER NOT NULL DEFAULT 0,
    "status" TEXT NOT NULL DEFAULT 'pending',
    "errorMessage" TEXT,
    "uploadedBy" TEXT,
    "createdAt" DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updatedAt" DATETIME NOT NULL
);

-- CreateTable
CREATE TABLE "TiktokAdsCreativeData" (
    "id" INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
    "tenantId" TEXT NOT NULL,
    "uploadBatchId" TEXT,
    "periodStart" DATETIME NOT NULL,
    "periodEnd" DATETIME NOT NULL,
    "campaignId" TEXT NOT NULL,
    "campaignName" TEXT NOT NULL,
    "productId" TEXT NOT NULL,
    "creativeType" TEXT NOT NULL,
    "videoTitle" TEXT,
    "videoId" TEXT,
    "tiktokAccount" TEXT,
    "postingTime" DATETIME,
    "status" TEXT,
    "authorizationType" TEXT,
    "cost" REAL NOT NULL DEFAULT 0,
    "ordersSku" INTEGER NOT NULL DEFAULT 0,
    "costPerOrder" REAL NOT NULL DEFAULT 0,
    "grossRevenue" REAL NOT NULL DEFAULT 0,
    "roi" REAL NOT NULL DEFAULT 0,
    "impressions" INTEGER NOT NULL DEFAULT 0,
    "clicks" INTEGER NOT NULL DEFAULT 0,
    "ctr" REAL NOT NULL DEFAULT 0,
    "conversionRate" REAL NOT NULL DEFAULT 0,
    "watchRate2s" REAL,
    "watchRate6s" REAL,
    "watchRate25pct" REAL,
    "watchRate50pct" REAL,
    "watchRate75pct" REAL,
    "watchRate100pct" REAL,
    "currency" TEXT NOT NULL DEFAULT 'IDR',
    "createdAt" DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updatedAt" DATETIME NOT NULL,
    CONSTRAINT "TiktokAdsCreativeData_uploadBatchId_fkey" FOREIGN KEY ("uploadBatchId") REFERENCES "TiktokAdsUploadBatch" ("id") ON DELETE SET NULL ON UPDATE CASCADE
);

-- CreateTable
CREATE TABLE "TiktokAdsProductSummary" (
    "id" TEXT NOT NULL PRIMARY KEY,
    "tenantId" TEXT NOT NULL,
    "productId" TEXT NOT NULL,
    "periodType" TEXT NOT NULL,
    "periodDate" DATETIME NOT NULL,
    "totalCost" REAL NOT NULL DEFAULT 0,
    "totalOrders" INTEGER NOT NULL DEFAULT 0,
    "totalRevenue" REAL NOT NULL DEFAULT 0,
    "avgRoi" REAL NOT NULL DEFAULT 0,
    "avgCtr" REAL NOT NULL DEFAULT 0,
    "avgConversionRate" REAL NOT NULL DEFAULT 0,
    "totalImpressions" INTEGER NOT NULL DEFAULT 0,
    "totalClicks" INTEGER NOT NULL DEFAULT 0,
    "bestCreativeType" TEXT,
    "topVideoId" TEXT,
    "topVideoTitle" TEXT,
    "roiTrend" TEXT,
    "costTrend" TEXT,
    "ordersTrend" TEXT,
    "createdAt" DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updatedAt" DATETIME NOT NULL
);

-- CreateTable
CREATE TABLE "TiktokAdsMLPrediction" (
    "id" TEXT NOT NULL PRIMARY KEY,
    "tenantId" TEXT NOT NULL,
    "productId" TEXT NOT NULL,
    "predictionDate" DATETIME NOT NULL,
    "targetMonth" INTEGER NOT NULL,
    "targetYear" INTEGER NOT NULL,
    "predictedRoi" REAL,
    "roiConfidence" REAL,
    "recommendedBudget" REAL,
    "budgetReason" TEXT,
    "performanceScore" TEXT,
    "performanceReason" TEXT,
    "creativeScore" REAL,
    "bestCreativeType" TEXT,
    "insights" TEXT,
    "hasAnomaly" BOOLEAN NOT NULL DEFAULT false,
    "anomalyType" TEXT,
    "anomalyDescription" TEXT,
    "modelVersion" TEXT NOT NULL DEFAULT 'v1.0',
    "modelType" TEXT,
    "trainingDataPoints" INTEGER,
    "createdAt" DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updatedAt" DATETIME NOT NULL
);

-- CreateIndex
CREATE UNIQUE INDEX "User_username_key" ON "User"("username");

-- CreateIndex
CREATE UNIQUE INDEX "User_email_key" ON "User"("email");

-- CreateIndex
CREATE INDEX "User_username_idx" ON "User"("username");

-- CreateIndex
CREATE INDEX "User_email_idx" ON "User"("email");

-- CreateIndex
CREATE INDEX "User_role_idx" ON "User"("role");

-- CreateIndex
CREATE INDEX "PlatformConfig_platform_idx" ON "PlatformConfig"("platform");

-- CreateIndex
CREATE UNIQUE INDEX "PlatformConfig_platform_configKey_key" ON "PlatformConfig"("platform", "configKey");

-- CreateIndex
CREATE UNIQUE INDEX "ShopeeOrder_orderSn_key" ON "ShopeeOrder"("orderSn");

-- CreateIndex
CREATE INDEX "ShopeeOrder_tenantId_idx" ON "ShopeeOrder"("tenantId");

-- CreateIndex
CREATE INDEX "ShopeeOrder_orderStatus_idx" ON "ShopeeOrder"("orderStatus");

-- CreateIndex
CREATE INDEX "ShopeeOrder_createdAt_idx" ON "ShopeeOrder"("createdAt");

-- CreateIndex
CREATE UNIQUE INDEX "ShopeeOrder_tenantId_orderSn_key" ON "ShopeeOrder"("tenantId", "orderSn");

-- CreateIndex
CREATE INDEX "ShopeeOrderItem_orderSn_idx" ON "ShopeeOrderItem"("orderSn");

-- CreateIndex
CREATE UNIQUE INDEX "LazadaOrder_orderSn_key" ON "LazadaOrder"("orderSn");

-- CreateIndex
CREATE INDEX "LazadaOrder_tenantId_idx" ON "LazadaOrder"("tenantId");

-- CreateIndex
CREATE INDEX "LazadaOrder_orderStatus_idx" ON "LazadaOrder"("orderStatus");

-- CreateIndex
CREATE INDEX "LazadaOrder_createdAt_idx" ON "LazadaOrder"("createdAt");

-- CreateIndex
CREATE UNIQUE INDEX "LazadaOrder_tenantId_orderSn_key" ON "LazadaOrder"("tenantId", "orderSn");

-- CreateIndex
CREATE INDEX "LazadaOrderItem_orderSn_idx" ON "LazadaOrderItem"("orderSn");

-- CreateIndex
CREATE UNIQUE INDEX "TiktokOrder_orderSn_key" ON "TiktokOrder"("orderSn");

-- CreateIndex
CREATE INDEX "TiktokOrder_tenantId_idx" ON "TiktokOrder"("tenantId");

-- CreateIndex
CREATE INDEX "TiktokOrder_orderStatus_idx" ON "TiktokOrder"("orderStatus");

-- CreateIndex
CREATE INDEX "TiktokOrder_createdAt_idx" ON "TiktokOrder"("createdAt");

-- CreateIndex
CREATE UNIQUE INDEX "TiktokOrder_tenantId_orderSn_key" ON "TiktokOrder"("tenantId", "orderSn");

-- CreateIndex
CREATE INDEX "TiktokOrderItem_orderSn_idx" ON "TiktokOrderItem"("orderSn");

-- CreateIndex
CREATE INDEX "OrderTodayItem_tenantId_idx" ON "OrderTodayItem"("tenantId");

-- CreateIndex
CREATE INDEX "OrderTodayItem_platform_idx" ON "OrderTodayItem"("platform");

-- CreateIndex
CREATE INDEX "OrderTodayItem_syncedAt_idx" ON "OrderTodayItem"("syncedAt");

-- CreateIndex
CREATE UNIQUE INDEX "OrderTodayItem_tenantId_platform_orderSn_sellerSku_key" ON "OrderTodayItem"("tenantId", "platform", "orderSn", "sellerSku");

-- CreateIndex
CREATE INDEX "InventorySettings_tenantId_idx" ON "InventorySettings"("tenantId");

-- CreateIndex
CREATE UNIQUE INDEX "InventorySettings_tenantId_key" ON "InventorySettings"("tenantId");

-- CreateIndex
CREATE INDEX "InventoryRecord_tenantId_idx" ON "InventoryRecord"("tenantId");

-- CreateIndex
CREATE INDEX "InventoryRecord_keyColumnName_keyValue_idx" ON "InventoryRecord"("keyColumnName", "keyValue");

-- CreateIndex
CREATE INDEX "InventoryRecord_createdAt_idx" ON "InventoryRecord"("createdAt");

-- CreateIndex
CREATE UNIQUE INDEX "InventoryRecord_tenantId_keyColumnName_keyValue_key" ON "InventoryRecord"("tenantId", "keyColumnName", "keyValue");

-- CreateIndex
CREATE INDEX "InventorySyncHistory_tenantId_idx" ON "InventorySyncHistory"("tenantId");

-- CreateIndex
CREATE INDEX "InventorySyncHistory_syncDirection_idx" ON "InventorySyncHistory"("syncDirection");

-- CreateIndex
CREATE INDEX "InventorySyncHistory_createdAt_idx" ON "InventorySyncHistory"("createdAt");

-- CreateIndex
CREATE INDEX "SheetSnapshot_tenantId_idx" ON "SheetSnapshot"("tenantId");

-- CreateIndex
CREATE INDEX "SheetSnapshot_sku_idx" ON "SheetSnapshot"("sku");

-- CreateIndex
CREATE INDEX "SheetSnapshot_syncedAt_idx" ON "SheetSnapshot"("syncedAt");

-- CreateIndex
CREATE UNIQUE INDEX "SheetSnapshot_tenantId_sku_key" ON "SheetSnapshot"("tenantId", "sku");

-- CreateIndex
CREATE INDEX "LazadaProduct_tenantId_idx" ON "LazadaProduct"("tenantId");

-- CreateIndex
CREATE INDEX "LazadaProduct_status_idx" ON "LazadaProduct"("status");

-- CreateIndex
CREATE INDEX "LazadaProduct_createdAt_idx" ON "LazadaProduct"("createdAt");

-- CreateIndex
CREATE UNIQUE INDEX "LazadaProduct_tenantId_itemId_key" ON "LazadaProduct"("tenantId", "itemId");

-- CreateIndex
CREATE UNIQUE INDEX "LazadaSku_skuId_key" ON "LazadaSku"("skuId");

-- CreateIndex
CREATE INDEX "LazadaSku_productId_idx" ON "LazadaSku"("productId");

-- CreateIndex
CREATE INDEX "LazadaSku_skuId_idx" ON "LazadaSku"("skuId");

-- CreateIndex
CREATE INDEX "ShopeeProduct_tenantId_idx" ON "ShopeeProduct"("tenantId");

-- CreateIndex
CREATE INDEX "ShopeeProduct_status_idx" ON "ShopeeProduct"("status");

-- CreateIndex
CREATE INDEX "ShopeeProduct_createdAt_idx" ON "ShopeeProduct"("createdAt");

-- CreateIndex
CREATE UNIQUE INDEX "ShopeeProduct_tenantId_itemId_key" ON "ShopeeProduct"("tenantId", "itemId");

-- CreateIndex
CREATE INDEX "ShopeeSku_productId_idx" ON "ShopeeSku"("productId");

-- CreateIndex
CREATE INDEX "ShopeeSku_modelId_idx" ON "ShopeeSku"("modelId");

-- CreateIndex
CREATE INDEX "ShopeeSku_itemId_idx" ON "ShopeeSku"("itemId");

-- CreateIndex
CREATE INDEX "TiktokProduct_tenantId_idx" ON "TiktokProduct"("tenantId");

-- CreateIndex
CREATE INDEX "TiktokProduct_status_idx" ON "TiktokProduct"("status");

-- CreateIndex
CREATE INDEX "TiktokProduct_createdAt_idx" ON "TiktokProduct"("createdAt");

-- CreateIndex
CREATE UNIQUE INDEX "TiktokProduct_tenantId_productId_key" ON "TiktokProduct"("tenantId", "productId");

-- CreateIndex
CREATE UNIQUE INDEX "TiktokSku_skuId_key" ON "TiktokSku"("skuId");

-- CreateIndex
CREATE INDEX "TiktokSku_productId_idx" ON "TiktokSku"("productId");

-- CreateIndex
CREATE INDEX "TiktokSku_skuId_idx" ON "TiktokSku"("skuId");

-- CreateIndex
CREATE INDEX "FilterPreference_tenantId_idx" ON "FilterPreference"("tenantId");

-- CreateIndex
CREATE UNIQUE INDEX "FilterPreference_tenantId_platform_tab_key" ON "FilterPreference"("tenantId", "platform", "tab");

-- CreateIndex
CREATE INDEX "InventorySkuPlatformStatus_tenantId_idx" ON "InventorySkuPlatformStatus"("tenantId");

-- CreateIndex
CREATE INDEX "InventorySkuPlatformStatus_sku_idx" ON "InventorySkuPlatformStatus"("sku");

-- CreateIndex
CREATE INDEX "InventorySkuPlatformStatus_updatedAt_idx" ON "InventorySkuPlatformStatus"("updatedAt");

-- CreateIndex
CREATE UNIQUE INDEX "InventorySkuPlatformStatus_tenantId_sku_key" ON "InventorySkuPlatformStatus"("tenantId", "sku");

-- CreateIndex
CREATE INDEX "RouteConfig_tenantId_idx" ON "RouteConfig"("tenantId");

-- CreateIndex
CREATE INDEX "RouteConfig_routePath_idx" ON "RouteConfig"("routePath");

-- CreateIndex
CREATE INDEX "RouteConfig_enabled_idx" ON "RouteConfig"("enabled");

-- CreateIndex
CREATE INDEX "RouteConfig_category_idx" ON "RouteConfig"("category");

-- CreateIndex
CREATE INDEX "RouteConfig_lastExecuted_idx" ON "RouteConfig"("lastExecuted");

-- CreateIndex
CREATE UNIQUE INDEX "RouteConfig_tenantId_routePath_routeMethod_key" ON "RouteConfig"("tenantId", "routePath", "routeMethod");

-- CreateIndex
CREATE INDEX "Spreadsheet_tenantId_idx" ON "Spreadsheet"("tenantId");

-- CreateIndex
CREATE INDEX "Spreadsheet_purpose_idx" ON "Spreadsheet"("purpose");

-- CreateIndex
CREATE UNIQUE INDEX "Spreadsheet_spreadsheetId_tenantId_key" ON "Spreadsheet"("spreadsheetId", "tenantId");

-- CreateIndex
CREATE INDEX "LockedOrder_tenantId_idx" ON "LockedOrder"("tenantId");

-- CreateIndex
CREATE INDEX "LockedOrder_sku_idx" ON "LockedOrder"("sku");

-- CreateIndex
CREATE INDEX "LockedOrder_createdAt_idx" ON "LockedOrder"("createdAt");

-- CreateIndex
CREATE UNIQUE INDEX "LockedOrder_tenantId_sku_productName_variationName_key" ON "LockedOrder"("tenantId", "sku", "productName", "variationName");

-- CreateIndex
CREATE UNIQUE INDEX "WholesaleSettings_tenantId_key" ON "WholesaleSettings"("tenantId");

-- CreateIndex
CREATE INDEX "WholesaleSettings_tenantId_idx" ON "WholesaleSettings"("tenantId");

-- CreateIndex
CREATE INDEX "WholesaleSettings_platform_idx" ON "WholesaleSettings"("platform");

-- CreateIndex
CREATE INDEX "ShopeeEscrowSync_tenantId_idx" ON "ShopeeEscrowSync"("tenantId");

-- CreateIndex
CREATE INDEX "ShopeeEscrowSync_month_year_idx" ON "ShopeeEscrowSync"("month", "year");

-- CreateIndex
CREATE UNIQUE INDEX "ShopeeEscrowSync_tenantId_month_year_key" ON "ShopeeEscrowSync"("tenantId", "month", "year");

-- CreateIndex
CREATE INDEX "ShopeeEscrowOrder_tenantId_idx" ON "ShopeeEscrowOrder"("tenantId");

-- CreateIndex
CREATE INDEX "ShopeeEscrowOrder_month_year_idx" ON "ShopeeEscrowOrder"("month", "year");

-- CreateIndex
CREATE INDEX "ShopeeEscrowOrder_orderSn_idx" ON "ShopeeEscrowOrder"("orderSn");

-- CreateIndex
CREATE UNIQUE INDEX "ShopeeEscrowOrder_tenantId_orderSn_key" ON "ShopeeEscrowOrder"("tenantId", "orderSn");

-- CreateIndex
CREATE INDEX "ShopeeEscrowItem_tenantId_idx" ON "ShopeeEscrowItem"("tenantId");

-- CreateIndex
CREATE INDEX "ShopeeEscrowItem_escrowOrderId_idx" ON "ShopeeEscrowItem"("escrowOrderId");

-- CreateIndex
CREATE INDEX "ShopeeEscrowItem_sku_idx" ON "ShopeeEscrowItem"("sku");

-- CreateIndex
CREATE INDEX "ShopeeEscrowItem_modelSku_idx" ON "ShopeeEscrowItem"("modelSku");

-- CreateIndex
CREATE INDEX "AnalyticsSettings_tenantId_idx" ON "AnalyticsSettings"("tenantId");

-- CreateIndex
CREATE UNIQUE INDEX "AnalyticsSettings_tenantId_platform_key" ON "AnalyticsSettings"("tenantId", "platform");

-- CreateIndex
CREATE INDEX "TiktokEscrowSync_tenantId_idx" ON "TiktokEscrowSync"("tenantId");

-- CreateIndex
CREATE INDEX "TiktokEscrowSync_month_year_idx" ON "TiktokEscrowSync"("month", "year");

-- CreateIndex
CREATE UNIQUE INDEX "TiktokEscrowSync_tenantId_month_year_key" ON "TiktokEscrowSync"("tenantId", "month", "year");

-- CreateIndex
CREATE INDEX "TiktokEscrowOrder_tenantId_idx" ON "TiktokEscrowOrder"("tenantId");

-- CreateIndex
CREATE INDEX "TiktokEscrowOrder_month_year_idx" ON "TiktokEscrowOrder"("month", "year");

-- CreateIndex
CREATE INDEX "TiktokEscrowOrder_orderId_idx" ON "TiktokEscrowOrder"("orderId");

-- CreateIndex
CREATE UNIQUE INDEX "TiktokEscrowOrder_tenantId_orderId_key" ON "TiktokEscrowOrder"("tenantId", "orderId");

-- CreateIndex
CREATE INDEX "TiktokEscrowItem_tenantId_idx" ON "TiktokEscrowItem"("tenantId");

-- CreateIndex
CREATE INDEX "TiktokEscrowItem_escrowOrderId_idx" ON "TiktokEscrowItem"("escrowOrderId");

-- CreateIndex
CREATE INDEX "TiktokEscrowItem_orderId_idx" ON "TiktokEscrowItem"("orderId");

-- CreateIndex
CREATE INDEX "TiktokEscrowItem_sellerSku_idx" ON "TiktokEscrowItem"("sellerSku");

-- CreateIndex
CREATE UNIQUE INDEX "OAuthState_state_key" ON "OAuthState"("state");

-- CreateIndex
CREATE INDEX "OAuthState_tenantId_idx" ON "OAuthState"("tenantId");

-- CreateIndex
CREATE INDEX "OAuthState_platform_idx" ON "OAuthState"("platform");

-- CreateIndex
CREATE INDEX "OAuthState_state_idx" ON "OAuthState"("state");

-- CreateIndex
CREATE INDEX "OAuthState_expiresAt_idx" ON "OAuthState"("expiresAt");

-- CreateIndex
CREATE INDEX "OAuthLog_tenantId_idx" ON "OAuthLog"("tenantId");

-- CreateIndex
CREATE INDEX "OAuthLog_platform_idx" ON "OAuthLog"("platform");

-- CreateIndex
CREATE INDEX "OAuthLog_eventType_idx" ON "OAuthLog"("eventType");

-- CreateIndex
CREATE INDEX "OAuthLog_status_idx" ON "OAuthLog"("status");

-- CreateIndex
CREATE INDEX "OAuthLog_createdAt_idx" ON "OAuthLog"("createdAt");

-- CreateIndex
CREATE INDEX "WebhookLog_tenantId_idx" ON "WebhookLog"("tenantId");

-- CreateIndex
CREATE INDEX "WebhookLog_platform_idx" ON "WebhookLog"("platform");

-- CreateIndex
CREATE INDEX "WebhookLog_eventType_idx" ON "WebhookLog"("eventType");

-- CreateIndex
CREATE INDEX "WebhookLog_status_idx" ON "WebhookLog"("status");

-- CreateIndex
CREATE INDEX "WebhookLog_createdAt_idx" ON "WebhookLog"("createdAt");

-- CreateIndex
CREATE INDEX "WebhookOrderEvent_tenantId_idx" ON "WebhookOrderEvent"("tenantId");

-- CreateIndex
CREATE INDEX "WebhookOrderEvent_orderSn_idx" ON "WebhookOrderEvent"("orderSn");

-- CreateIndex
CREATE INDEX "WebhookOrderEvent_platform_idx" ON "WebhookOrderEvent"("platform");

-- CreateIndex
CREATE INDEX "WebhookOrderEvent_eventType_idx" ON "WebhookOrderEvent"("eventType");

-- CreateIndex
CREATE INDEX "WebhookOrderEvent_createdAt_idx" ON "WebhookOrderEvent"("createdAt");

-- CreateIndex
CREATE INDEX "WebhookProductEvent_tenantId_idx" ON "WebhookProductEvent"("tenantId");

-- CreateIndex
CREATE INDEX "WebhookProductEvent_itemId_idx" ON "WebhookProductEvent"("itemId");

-- CreateIndex
CREATE INDEX "WebhookProductEvent_eventType_idx" ON "WebhookProductEvent"("eventType");

-- CreateIndex
CREATE INDEX "WebhookMarketingEvent_tenantId_idx" ON "WebhookMarketingEvent"("tenantId");

-- CreateIndex
CREATE INDEX "WebhookMarketingEvent_promotionId_idx" ON "WebhookMarketingEvent"("promotionId");

-- CreateIndex
CREATE INDEX "WebhookMarketingEvent_eventType_idx" ON "WebhookMarketingEvent"("eventType");

-- CreateIndex
CREATE INDEX "WebhookReturnEvent_tenantId_idx" ON "WebhookReturnEvent"("tenantId");

-- CreateIndex
CREATE INDEX "WebhookReturnEvent_orderSn_idx" ON "WebhookReturnEvent"("orderSn");

-- CreateIndex
CREATE INDEX "WebhookReturnEvent_returnSn_idx" ON "WebhookReturnEvent"("returnSn");

-- CreateIndex
CREATE INDEX "WebhookShopeeEvent_tenantId_idx" ON "WebhookShopeeEvent"("tenantId");

-- CreateIndex
CREATE INDEX "WebhookShopeeEvent_shopId_idx" ON "WebhookShopeeEvent"("shopId");

-- CreateIndex
CREATE INDEX "WebhookShopeeEvent_eventType_idx" ON "WebhookShopeeEvent"("eventType");

-- CreateIndex
CREATE INDEX "WebhookWebchatEvent_tenantId_idx" ON "WebhookWebchatEvent"("tenantId");

-- CreateIndex
CREATE INDEX "WebhookWebchatEvent_conversationId_idx" ON "WebhookWebchatEvent"("conversationId");

-- CreateIndex
CREATE INDEX "WebhookFBSEvent_tenantId_idx" ON "WebhookFBSEvent"("tenantId");

-- CreateIndex
CREATE INDEX "WebhookFBSEvent_itemId_idx" ON "WebhookFBSEvent"("itemId");

-- CreateIndex
CREATE INDEX "WebhookFBSEvent_eventType_idx" ON "WebhookFBSEvent"("eventType");

-- CreateIndex
CREATE INDEX "TiktokAdsUploadBatch_tenantId_idx" ON "TiktokAdsUploadBatch"("tenantId");

-- CreateIndex
CREATE INDEX "TiktokAdsUploadBatch_status_idx" ON "TiktokAdsUploadBatch"("status");

-- CreateIndex
CREATE INDEX "TiktokAdsUploadBatch_periodStart_periodEnd_idx" ON "TiktokAdsUploadBatch"("periodStart", "periodEnd");

-- CreateIndex
CREATE INDEX "TiktokAdsCreativeData_tenantId_idx" ON "TiktokAdsCreativeData"("tenantId");

-- CreateIndex
CREATE INDEX "TiktokAdsCreativeData_campaignId_idx" ON "TiktokAdsCreativeData"("campaignId");

-- CreateIndex
CREATE INDEX "TiktokAdsCreativeData_productId_idx" ON "TiktokAdsCreativeData"("productId");

-- CreateIndex
CREATE INDEX "TiktokAdsCreativeData_creativeType_idx" ON "TiktokAdsCreativeData"("creativeType");

-- CreateIndex
CREATE INDEX "TiktokAdsCreativeData_periodStart_periodEnd_idx" ON "TiktokAdsCreativeData"("periodStart", "periodEnd");

-- CreateIndex
CREATE INDEX "TiktokAdsCreativeData_uploadBatchId_idx" ON "TiktokAdsCreativeData"("uploadBatchId");

-- CreateIndex
CREATE UNIQUE INDEX "TiktokAdsCreativeData_tenantId_campaignId_productId_videoId_periodStart_periodEnd_key" ON "TiktokAdsCreativeData"("tenantId", "campaignId", "productId", "videoId", "periodStart", "periodEnd");

-- CreateIndex
CREATE INDEX "TiktokAdsProductSummary_tenantId_idx" ON "TiktokAdsProductSummary"("tenantId");

-- CreateIndex
CREATE INDEX "TiktokAdsProductSummary_productId_idx" ON "TiktokAdsProductSummary"("productId");

-- CreateIndex
CREATE INDEX "TiktokAdsProductSummary_periodType_periodDate_idx" ON "TiktokAdsProductSummary"("periodType", "periodDate");

-- CreateIndex
CREATE UNIQUE INDEX "TiktokAdsProductSummary_tenantId_productId_periodType_periodDate_key" ON "TiktokAdsProductSummary"("tenantId", "productId", "periodType", "periodDate");

-- CreateIndex
CREATE INDEX "TiktokAdsMLPrediction_tenantId_idx" ON "TiktokAdsMLPrediction"("tenantId");

-- CreateIndex
CREATE INDEX "TiktokAdsMLPrediction_productId_idx" ON "TiktokAdsMLPrediction"("productId");

-- CreateIndex
CREATE INDEX "TiktokAdsMLPrediction_targetMonth_targetYear_idx" ON "TiktokAdsMLPrediction"("targetMonth", "targetYear");

-- CreateIndex
CREATE INDEX "TiktokAdsMLPrediction_performanceScore_idx" ON "TiktokAdsMLPrediction"("performanceScore");

-- CreateIndex
CREATE UNIQUE INDEX "TiktokAdsMLPrediction_tenantId_productId_targetMonth_targetYear_key" ON "TiktokAdsMLPrediction"("tenantId", "productId", "targetMonth", "targetYear");
