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

-- CreateIndex
CREATE INDEX "LockedOrder_tenantId_idx" ON "LockedOrder"("tenantId");

-- CreateIndex
CREATE INDEX "LockedOrder_sku_idx" ON "LockedOrder"("sku");

-- CreateIndex
CREATE INDEX "LockedOrder_createdAt_idx" ON "LockedOrder"("createdAt");

-- CreateIndex
CREATE UNIQUE INDEX "LockedOrder_tenantId_sku_productName_variationName_key" ON "LockedOrder"("tenantId", "sku", "productName", "variationName");
