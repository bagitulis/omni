import { PrismaClient } from "@prisma/client";
import { tenantContext } from "../../utils/tenantContext";

/**
 * Handles tenant-aware inventory operations
 * - All queries automatically filter by tenantId
 * - Enforces multi-tenant data isolation for inventory
 */
export class InventoryTenantService {
  constructor(private prisma: PrismaClient) {}

  /**
   * Get inventory settings for current tenant
   */
  async getSettings() {
    const tenantId = tenantContext.getTenantId();
    return this.prisma.inventorySettings.findUnique({
      where: { tenantId },
    });
  }

  /**
   * Upsert settings for current tenant
   */
  async upsertSettings(data: any) {
    const tenantId = tenantContext.getTenantId();
    return this.prisma.inventorySettings.upsert({
      where: { tenantId },
      update: {
        ...data,
        updatedAt: new Date(),
      },
      create: {
        tenantId,
        ...data,
      },
    });
  }

  /**
   * Delete settings for current tenant
   */
  async deleteSettings(): Promise<void> {
    const tenantId = tenantContext.getTenantId();
    await this.prisma.inventorySettings.deleteMany({
      where: { tenantId },
    });
  }

  /**
   * Update settings fields for current tenant
   */
  async updateSettings(data: any) {
    const tenantId = tenantContext.getTenantId();
    return this.prisma.inventorySettings.updateMany({
      where: { tenantId },
      data,
    });
  }

  /**
   * Count inventory records for current tenant
   */
  async countRecords(where?: any): Promise<number> {
    const tenantId = tenantContext.getTenantId();
    return this.prisma.inventoryRecord.count({
      where: { ...where, tenantId },
    });
  }

  /**
   * Find inventory records with tenant filter
   */
  async findRecords(options?: any) {
    const tenantId = tenantContext.getTenantId();
    return this.prisma.inventoryRecord.findMany({
      ...options,
      where: { ...options?.where, tenantId },
    });
  }

  /**
   * Find first inventory record matching criteria (tenant-scoped)
   */
  async findFirstRecord(where: any) {
    const tenantId = tenantContext.getTenantId();
    return this.prisma.inventoryRecord.findFirst({
      where: { ...where, tenantId },
    });
  }

  /**
   * Find unique inventory record by ID (tenant-scoped)
   */
  async findUniqueRecord(id: string) {
    const tenantId = tenantContext.getTenantId();
    // Need to verify record belongs to tenant
    const record = await this.prisma.inventoryRecord.findUnique({
      where: { id },
    });
    if (record && record.tenantId !== tenantId) {
      return null; // Record belongs to different tenant
    }
    return record;
  }

  /**
   * Create inventory record for current tenant
   */
  async createRecord(data: any) {
    const tenantId = tenantContext.getTenantId();
    return this.prisma.inventoryRecord.create({
      data: {
        ...data,
        tenantId,
      },
    });
  }

  /**
   * Update inventory record (tenant-scoped)
   */
  async updateRecord(id: string, data: any) {
    // Verify record belongs to tenant before updating
    const record = await this.findUniqueRecord(id);
    if (!record) {
      throw new Error("Record not found or access denied");
    }
    return this.prisma.inventoryRecord.update({
      where: { id },
      data,
    });
  }

  /**
   * Delete inventory record (tenant-scoped)
   */
  async deleteRecord(id: string) {
    const record = await this.findUniqueRecord(id);
    if (!record) {
      throw new Error("Record not found or access denied");
    }
    return this.prisma.inventoryRecord.delete({
      where: { id },
    });
  }
}
