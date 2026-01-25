-- CreateTable
CREATE TABLE "Spreadsheet" (
    "id" TEXT NOT NULL PRIMARY KEY,
    "spreadsheetId" TEXT NOT NULL,
    "spreadsheetUrl" TEXT NOT NULL,
    "spreadsheetName" TEXT NOT NULL,
    "purpose" TEXT NOT NULL DEFAULT 'other',
    "sheets" TEXT,
    "tenantId" TEXT NOT NULL DEFAULT 'yumna',
    "registeredBy" TEXT,
    "lastUsedAt" DATETIME,
    "createdAt" DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updatedAt" DATETIME NOT NULL
);

-- CreateIndex
CREATE INDEX "Spreadsheet_tenantId_idx" ON "Spreadsheet"("tenantId");

-- CreateIndex
CREATE INDEX "Spreadsheet_purpose_idx" ON "Spreadsheet"("purpose");

-- CreateIndex
CREATE UNIQUE INDEX "Spreadsheet_spreadsheetId_tenantId_key" ON "Spreadsheet"("spreadsheetId", "tenantId");
