/*
  Warnings:

  - Added the required column `tenantId` to the `LazadaOrder` table without a default value. This is not possible if the table is not empty.
  - Added the required column `tenantId` to the `ShopeeOrder` table without a default value. This is not possible if the table is not empty.

*/
-- RedefineTables
PRAGMA defer_foreign_keys=ON;
PRAGMA foreign_keys=OFF;
CREATE TABLE "new_LazadaOrder" (
    "id" INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
    "tenantId" TEXT NOT NULL,
    "orderSn" TEXT NOT NULL,
    "shopId" BIGINT,
    "orderStatus" TEXT NOT NULL,
    "createdAt" DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updatedAt" DATETIME NOT NULL,
    "trackingNumber" TEXT,
    "shippingCarrier" TEXT
);
INSERT INTO "new_LazadaOrder" ("createdAt", "id", "orderSn", "orderStatus", "shippingCarrier", "shopId", "trackingNumber", "updatedAt", "tenantId") SELECT "createdAt", "id", "orderSn", "orderStatus", "shippingCarrier", "shopId", "trackingNumber", "updatedAt", 'yumna' FROM "LazadaOrder";
DROP TABLE "LazadaOrder";
ALTER TABLE "new_LazadaOrder" RENAME TO "LazadaOrder";
CREATE INDEX "LazadaOrder_tenantId_idx" ON "LazadaOrder"("tenantId");
CREATE INDEX "LazadaOrder_orderStatus_idx" ON "LazadaOrder"("orderStatus");
CREATE INDEX "LazadaOrder_createdAt_idx" ON "LazadaOrder"("createdAt");
CREATE UNIQUE INDEX "LazadaOrder_tenantId_orderSn_key" ON "LazadaOrder"("tenantId", "orderSn");
CREATE UNIQUE INDEX "LazadaOrder_orderSn_key" ON "LazadaOrder"("orderSn");
CREATE TABLE "new_ShopeeOrder" (
    "id" INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
    "tenantId" TEXT NOT NULL,
    "orderSn" TEXT NOT NULL,
    "shopId" BIGINT,
    "orderStatus" TEXT NOT NULL,
    "orderTimestamp" INTEGER,
    "trackingNumber" TEXT,
    "shippingCarrier" TEXT,
    "createdAt" DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updatedAt" DATETIME NOT NULL
);
INSERT INTO "new_ShopeeOrder" ("createdAt", "id", "orderSn", "orderStatus", "orderTimestamp", "shippingCarrier", "shopId", "trackingNumber", "updatedAt", "tenantId") SELECT "createdAt", "id", "orderSn", "orderStatus", "orderTimestamp", "shippingCarrier", "shopId", "trackingNumber", "updatedAt", 'yumna' FROM "ShopeeOrder";
DROP TABLE "ShopeeOrder";
ALTER TABLE "new_ShopeeOrder" RENAME TO "ShopeeOrder";
CREATE INDEX "ShopeeOrder_tenantId_idx" ON "ShopeeOrder"("tenantId");
CREATE INDEX "ShopeeOrder_orderStatus_idx" ON "ShopeeOrder"("orderStatus");
CREATE INDEX "ShopeeOrder_createdAt_idx" ON "ShopeeOrder"("createdAt");
CREATE UNIQUE INDEX "ShopeeOrder_tenantId_orderSn_key" ON "ShopeeOrder"("tenantId", "orderSn");
CREATE UNIQUE INDEX "ShopeeOrder_orderSn_key" ON "ShopeeOrder"("orderSn");
CREATE TABLE "new_TiktokOrder" (
    "id" INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
    "tenantId" TEXT NOT NULL,
    "orderSn" TEXT NOT NULL,
    "shopId" BIGINT,
    "orderStatus" TEXT NOT NULL,
    "createdAt" DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updatedAt" DATETIME NOT NULL,
    "trackingNumber" TEXT,
    "shippingCarrier" TEXT
);
INSERT INTO "new_TiktokOrder" ("createdAt", "id", "orderSn", "orderStatus", "shippingCarrier", "shopId", "trackingNumber", "updatedAt", "tenantId") SELECT "createdAt", "id", "orderSn", "orderStatus", "shippingCarrier", "shopId", "trackingNumber", "updatedAt", 'yumna' FROM "TiktokOrder";
DROP TABLE "TiktokOrder";
ALTER TABLE "new_TiktokOrder" RENAME TO "TiktokOrder";
CREATE INDEX "TiktokOrder_tenantId_idx" ON "TiktokOrder"("tenantId");
CREATE INDEX "TiktokOrder_orderStatus_idx" ON "TiktokOrder"("orderStatus");
CREATE INDEX "TiktokOrder_createdAt_idx" ON "TiktokOrder"("createdAt");
CREATE UNIQUE INDEX "TiktokOrder_tenantId_orderSn_key" ON "TiktokOrder"("tenantId", "orderSn");
CREATE UNIQUE INDEX "TiktokOrder_orderSn_key" ON "TiktokOrder"("orderSn");
PRAGMA foreign_keys=ON;
PRAGMA defer_foreign_keys=OFF;
