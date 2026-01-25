/**
 * SnapshotService - Manages SheetSnapshot for Delta Sync
 *
 * SINGLE RESPONSIBILITY: Store and retrieve Google Sheet snapshots
 * - Save sheet state after sync from sheets
 * - Save sheet state after successful export
 * - Retrieve snapshot for comparison
 * - Multi-tenant support via tenantContext
 */

import { PrismaClient } from "@prisma/client";
import { getPrisma } from "../prismaClient";
import { tenantContext } from "../../utils/tenantContext";
import crypto from "crypto";

export interface SnapshotRecord {
  sku: string;
  rowIndex: number;
  data: Record<string, any>;
  dataHash: string;
}

export interface SnapshotSaveResult {
  saved: number;
  updated: number;
  errors: string[];
}

export class SnapshotService {
  /**
   * Get tenant-aware Prisma client
   * Called per-method to ensure correct tenant DB is used
   */
  private getPrismaClient(): PrismaClient {
    const tenantId = tenantContext.getTenantId();
    return getPrisma(tenantId);
  }

  /**
   * Generate hash for quick change detection
   */
  private generateHash(data: Record<string, any>): string {
    const sortedKeys = Object.keys(data).sort();
    const normalized = sortedKeys.map((k) => `${k}:${data[k] ?? ""}`).join("|");
    return crypto.createHash("md5").update(normalized).digest("hex");
  }

  /**
   * Save snapshot from Google Sheet data
   * Called after: Sync From Sheet or Export To Sheet
   */
  async saveSnapshot(
    sheetData: any[][],
    headers: string[],
    skuColumnName: string = "SKU"
  ): Promise<SnapshotSaveResult> {
    const tenantId = tenantContext.getTenantId();
    const prisma = this.getPrismaClient();
    const skuIndex = headers.findIndex(
      (h) => h.toLowerCase() === skuColumnName.toLowerCase()
    );

    if (skuIndex === -1) {
      return { saved: 0, updated: 0, errors: ["SKU column not found"] };
    }

    let saved = 0;
    let updated = 0;
    const errors: string[] = [];

    // Process data rows (skip header at index 0)
    for (let rowIndex = 1; rowIndex < sheetData.length; rowIndex++) {
      const row = sheetData[rowIndex];
      const sku = String(row[skuIndex] || "").trim();

      if (!sku) continue;

      // Build data object from row
      const data: Record<string, any> = {};
      for (let colIdx = 0; colIdx < headers.length; colIdx++) {
        data[headers[colIdx]] = row[colIdx] ?? "";
      }

      const dataHash = this.generateHash(data);

      try {
        await prisma.sheetSnapshot.upsert({
          where: { tenantId_sku: { tenantId, sku } },
          update: {
            snapshotData: JSON.stringify(data),
            rowIndex,
            dataHash,
            syncedAt: new Date(),
            updatedAt: new Date(),
          },
          create: {
            tenantId,
            sku,
            snapshotData: JSON.stringify(data),
            rowIndex,
            dataHash,
            syncedAt: new Date(),
          },
        });

        // Check if it was an update or create
        const existing = await prisma.sheetSnapshot.findUnique({
          where: { tenantId_sku: { tenantId, sku } },
        });
        if (existing && existing.createdAt < existing.updatedAt) {
          updated++;
        } else {
          saved++;
        }
      } catch (error: any) {
        errors.push(`SKU ${sku}: ${error.message}`);
      }
    }

    return { saved, updated, errors };
  }

  /**
   * Get all snapshots for current tenant
   */
  async getAllSnapshots(): Promise<Map<string, SnapshotRecord>> {
    const tenantId = tenantContext.getTenantId();
    const prisma = this.getPrismaClient();
    const snapshots = await prisma.sheetSnapshot.findMany({
      where: { tenantId },
    });

    const map = new Map<string, SnapshotRecord>();
    for (const snapshot of snapshots) {
      map.set(snapshot.sku, {
        sku: snapshot.sku,
        rowIndex: snapshot.rowIndex,
        data: JSON.parse(snapshot.snapshotData),
        dataHash: snapshot.dataHash || "",
      });
    }

    return map;
  }

  /**
   * Get snapshot by SKU
   */
  async getSnapshotBySku(sku: string): Promise<SnapshotRecord | null> {
    const tenantId = tenantContext.getTenantId();
    const prisma = this.getPrismaClient();
    const snapshot = await prisma.sheetSnapshot.findUnique({
      where: { tenantId_sku: { tenantId, sku } },
    });

    if (!snapshot) return null;

    return {
      sku: snapshot.sku,
      rowIndex: snapshot.rowIndex,
      data: JSON.parse(snapshot.snapshotData),
      dataHash: snapshot.dataHash || "",
    };
  }

  /**
   * Update single snapshot after cell update
   */
  async updateSnapshotCell(
    sku: string,
    column: string,
    newValue: any
  ): Promise<boolean> {
    const tenantId = tenantContext.getTenantId();
    const prisma = this.getPrismaClient();
    const snapshot = await prisma.sheetSnapshot.findUnique({
      where: { tenantId_sku: { tenantId, sku } },
    });

    if (!snapshot) return false;

    const data = JSON.parse(snapshot.snapshotData);
    data[column] = newValue;
    const dataHash = this.generateHash(data);

    await prisma.sheetSnapshot.update({
      where: { tenantId_sku: { tenantId, sku } },
      data: {
        snapshotData: JSON.stringify(data),
        dataHash,
        syncedAt: new Date(),
        updatedAt: new Date(),
      },
    });

    return true;
  }

  /**
   * Check if snapshot exists and is fresh (within TTL)
   */
  async isSnapshotFresh(ttlMinutes: number = 60): Promise<boolean> {
    const tenantId = tenantContext.getTenantId();
    const prisma = this.getPrismaClient();
    const latest = await prisma.sheetSnapshot.findFirst({
      where: { tenantId },
      orderBy: { syncedAt: "desc" },
    });

    if (!latest) return false;

    const ageMs = Date.now() - latest.syncedAt.getTime();
    const ttlMs = ttlMinutes * 60 * 1000;

    return ageMs < ttlMs;
  }

  /**
   * Clear all snapshots for current tenant
   */
  async clearSnapshots(): Promise<number> {
    const tenantId = tenantContext.getTenantId();
    const prisma = this.getPrismaClient();
    const result = await prisma.sheetSnapshot.deleteMany({
      where: { tenantId },
    });
    return result.count;
  }

  /**
   * Get snapshot count
   */
  async getSnapshotCount(): Promise<number> {
    const tenantId = tenantContext.getTenantId();
    const prisma = this.getPrismaClient();
    return prisma.sheetSnapshot.count({ where: { tenantId } });
  }
}

// Singleton instance
let instance: SnapshotService;

export function getSnapshotService(): SnapshotService {
  if (!instance) {
    instance = new SnapshotService();
  }
  return instance;
}
