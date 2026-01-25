-- CreateTable
CREATE TABLE "User" (
    "id" TEXT NOT NULL PRIMARY KEY,
    "username" TEXT NOT NULL,
    "email" TEXT NOT NULL,
    "password" TEXT NOT NULL,
    "role" TEXT NOT NULL DEFAULT 'owner',
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
    "orderSn" TEXT NOT NULL,
    "shopId" BIGINT,
    "orderStatus" TEXT NOT NULL,
    "orderTimestamp" INTEGER,
    "trackingNumber" TEXT,
    "shippingCarrier" TEXT,
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
    "orderSn" TEXT NOT NULL,
    "shopId" BIGINT,
    "orderStatus" TEXT NOT NULL,
    "createdAt" DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updatedAt" DATETIME NOT NULL,
    "trackingNumber" TEXT,
    "shippingCarrier" TEXT
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
    "orderSn" TEXT NOT NULL,
    "shopId" BIGINT,
    "orderStatus" TEXT NOT NULL,
    "createdAt" DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updatedAt" DATETIME NOT NULL,
    "trackingNumber" TEXT,
    "shippingCarrier" TEXT
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
    "orderSpreadsheetId" TEXT,
    "createdAt" DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updatedAt" DATETIME NOT NULL,
    "availableSpreadsheets" TEXT,
    "availableWorksheets" TEXT
);

-- CreateTable
CREATE TABLE "InventorySettings" (
    "id" TEXT NOT NULL PRIMARY KEY,
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
    "data" TEXT NOT NULL,
    "keyValue" TEXT NOT NULL,
    "keyColumnName" TEXT NOT NULL,
    "createdAt" DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updatedAt" DATETIME NOT NULL
);

-- CreateTable
CREATE TABLE "InventorySyncHistory" (
    "id" TEXT NOT NULL PRIMARY KEY,
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
CREATE TABLE "LazadaProduct" (
    "id" INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
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
    "modelId" BIGINT NOT NULL,
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

-- CreateIndex
CREATE UNIQUE INDEX "User_username_key" ON "User"("username");

-- CreateIndex
CREATE UNIQUE INDEX "User_email_key" ON "User"("email");

-- CreateIndex
CREATE INDEX "User_username_idx" ON "User"("username");

-- CreateIndex
CREATE INDEX "User_email_idx" ON "User"("email");

-- CreateIndex
CREATE INDEX "PlatformConfig_platform_idx" ON "PlatformConfig"("platform");

-- CreateIndex
CREATE UNIQUE INDEX "PlatformConfig_platform_configKey_key" ON "PlatformConfig"("platform", "configKey");

-- CreateIndex
CREATE UNIQUE INDEX "ShopeeOrder_orderSn_key" ON "ShopeeOrder"("orderSn");

-- CreateIndex
CREATE INDEX "ShopeeOrder_orderStatus_idx" ON "ShopeeOrder"("orderStatus");

-- CreateIndex
CREATE INDEX "ShopeeOrder_createdAt_idx" ON "ShopeeOrder"("createdAt");

-- CreateIndex
CREATE INDEX "ShopeeOrderItem_orderSn_idx" ON "ShopeeOrderItem"("orderSn");

-- CreateIndex
CREATE UNIQUE INDEX "LazadaOrder_orderSn_key" ON "LazadaOrder"("orderSn");

-- CreateIndex
CREATE INDEX "LazadaOrder_orderStatus_idx" ON "LazadaOrder"("orderStatus");

-- CreateIndex
CREATE INDEX "LazadaOrder_createdAt_idx" ON "LazadaOrder"("createdAt");

-- CreateIndex
CREATE INDEX "LazadaOrderItem_orderSn_idx" ON "LazadaOrderItem"("orderSn");

-- CreateIndex
CREATE UNIQUE INDEX "TiktokOrder_orderSn_key" ON "TiktokOrder"("orderSn");

-- CreateIndex
CREATE INDEX "TiktokOrder_orderStatus_idx" ON "TiktokOrder"("orderStatus");

-- CreateIndex
CREATE INDEX "TiktokOrder_createdAt_idx" ON "TiktokOrder"("createdAt");

-- CreateIndex
CREATE INDEX "TiktokOrderItem_orderSn_idx" ON "TiktokOrderItem"("orderSn");

-- CreateIndex
CREATE INDEX "InventoryRecord_keyColumnName_keyValue_idx" ON "InventoryRecord"("keyColumnName", "keyValue");

-- CreateIndex
CREATE INDEX "InventoryRecord_createdAt_idx" ON "InventoryRecord"("createdAt");

-- CreateIndex
CREATE UNIQUE INDEX "InventoryRecord_keyColumnName_keyValue_key" ON "InventoryRecord"("keyColumnName", "keyValue");

-- CreateIndex
CREATE INDEX "InventorySyncHistory_syncDirection_idx" ON "InventorySyncHistory"("syncDirection");

-- CreateIndex
CREATE INDEX "InventorySyncHistory_createdAt_idx" ON "InventorySyncHistory"("createdAt");

-- CreateIndex
CREATE UNIQUE INDEX "LazadaProduct_itemId_key" ON "LazadaProduct"("itemId");

-- CreateIndex
CREATE INDEX "LazadaProduct_itemId_idx" ON "LazadaProduct"("itemId");

-- CreateIndex
CREATE INDEX "LazadaProduct_status_idx" ON "LazadaProduct"("status");

-- CreateIndex
CREATE INDEX "LazadaProduct_createdAt_idx" ON "LazadaProduct"("createdAt");

-- CreateIndex
CREATE UNIQUE INDEX "LazadaSku_skuId_key" ON "LazadaSku"("skuId");

-- CreateIndex
CREATE INDEX "LazadaSku_productId_idx" ON "LazadaSku"("productId");

-- CreateIndex
CREATE INDEX "LazadaSku_skuId_idx" ON "LazadaSku"("skuId");

-- CreateIndex
CREATE UNIQUE INDEX "ShopeeProduct_itemId_key" ON "ShopeeProduct"("itemId");

-- CreateIndex
CREATE INDEX "ShopeeProduct_itemId_idx" ON "ShopeeProduct"("itemId");

-- CreateIndex
CREATE INDEX "ShopeeProduct_status_idx" ON "ShopeeProduct"("status");

-- CreateIndex
CREATE INDEX "ShopeeProduct_createdAt_idx" ON "ShopeeProduct"("createdAt");

-- CreateIndex
CREATE UNIQUE INDEX "ShopeeSku_modelId_key" ON "ShopeeSku"("modelId");

-- CreateIndex
CREATE INDEX "ShopeeSku_productId_idx" ON "ShopeeSku"("productId");

-- CreateIndex
CREATE INDEX "ShopeeSku_modelId_idx" ON "ShopeeSku"("modelId");

-- CreateIndex
CREATE INDEX "ShopeeSku_itemId_idx" ON "ShopeeSku"("itemId");

-- CreateIndex
CREATE UNIQUE INDEX "TiktokProduct_productId_key" ON "TiktokProduct"("productId");

-- CreateIndex
CREATE INDEX "TiktokProduct_productId_idx" ON "TiktokProduct"("productId");

-- CreateIndex
CREATE INDEX "TiktokProduct_status_idx" ON "TiktokProduct"("status");

-- CreateIndex
CREATE INDEX "TiktokProduct_createdAt_idx" ON "TiktokProduct"("createdAt");

-- CreateIndex
CREATE UNIQUE INDEX "TiktokSku_skuId_key" ON "TiktokSku"("skuId");

-- CreateIndex
CREATE INDEX "TiktokSku_productId_idx" ON "TiktokSku"("productId");

-- CreateIndex
CREATE INDEX "TiktokSku_skuId_idx" ON "TiktokSku"("skuId");

-- CreateIndex
CREATE UNIQUE INDEX "FilterPreference_platform_tab_key" ON "FilterPreference"("platform", "tab");

-- CreateIndex
CREATE UNIQUE INDEX "InventorySkuPlatformStatus_sku_key" ON "InventorySkuPlatformStatus"("sku");

-- CreateIndex
CREATE INDEX "InventorySkuPlatformStatus_sku_idx" ON "InventorySkuPlatformStatus"("sku");

-- CreateIndex
CREATE INDEX "InventorySkuPlatformStatus_updatedAt_idx" ON "InventorySkuPlatformStatus"("updatedAt");

-- CreateIndex
CREATE UNIQUE INDEX "RouteConfig_routePath_key" ON "RouteConfig"("routePath");

-- CreateIndex
CREATE INDEX "RouteConfig_routePath_idx" ON "RouteConfig"("routePath");

-- CreateIndex
CREATE INDEX "RouteConfig_enabled_idx" ON "RouteConfig"("enabled");

-- CreateIndex
CREATE INDEX "RouteConfig_category_idx" ON "RouteConfig"("category");

-- CreateIndex
CREATE INDEX "RouteConfig_lastExecuted_idx" ON "RouteConfig"("lastExecuted");
