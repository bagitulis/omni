/**
 * Spreadsheet Registry Service
 * Manages registered spreadsheets with metadata and user access
 *
 * Single Responsibility: Spreadsheet registration and metadata management
 */

import { Logger } from "winston";
import { getLogger } from "../utils/logger";
import * as fs from "fs";
import * as path from "path";

export interface SheetMetadata {
  name: string;
  sheetId: number;
  index: number;
  columnCount?: number;
  rowCount?: number;
}

export interface RegisteredSpreadsheet {
  id: string;
  url: string;
  spreadsheetId: string;
  spreadsheetName: string;
  sheets: SheetMetadata[];
  purpose: "inventory" | "wallet" | "shipping" | "order" | "custom";
  createdBy: string;
  createdAt: Date;
  isLocked: boolean;
  lastUsedAt: Date | null;
  syncSettings: {
    autoSync: boolean;
    syncInterval?: number;
    lastSyncAt?: Date;
  };
  version: number;
  lastModifiedBy: string;
  lastModifiedAt: Date;
  isBeingEdited: boolean;
  editingLockedUntil: Date | null;
}

export class SpreadsheetRegistryService {
  private registryPath: string;
  private logger: Logger;
  private registry: Map<string, RegisteredSpreadsheet> = new Map();

  constructor(registryPath?: string) {
    this.logger = getLogger("SpreadsheetRegistry");
    this.registryPath =
      registryPath ||
      path.join(process.cwd(), "config/static/google/registry.json");
    this.loadRegistry();
  }

  /**
   * Load registry from file
   */
  private loadRegistry(): void {
    try {
      if (fs.existsSync(this.registryPath)) {
        const content = fs.readFileSync(this.registryPath, "utf8");
        const data = JSON.parse(content);

        for (const [id, item] of Object.entries(data)) {
          const spreadsheet = item as any;
          this.registry.set(id, {
            ...spreadsheet,
            createdAt: new Date(spreadsheet.createdAt),
            lastUsedAt: spreadsheet.lastUsedAt
              ? new Date(spreadsheet.lastUsedAt)
              : null,
            lastModifiedAt: new Date(spreadsheet.lastModifiedAt),
            editingLockedUntil: spreadsheet.editingLockedUntil
              ? new Date(spreadsheet.editingLockedUntil)
              : null,
            syncSettings: {
              ...spreadsheet.syncSettings,
              lastSyncAt: spreadsheet.syncSettings.lastSyncAt
                ? new Date(spreadsheet.syncSettings.lastSyncAt)
                : undefined,
            },
          });
        }

        this.logger.info(
          `✅ Loaded ${this.registry.size} registered spreadsheets`
        );
      }
    } catch (error) {
      this.logger.error(
        "Error loading registry",
        error instanceof Error ? error.message : String(error)
      );
    }
  }

  /**
   * Save registry to file
   */
  private saveRegistry(): void {
    try {
      const dir = path.dirname(this.registryPath);
      if (!fs.existsSync(dir)) {
        fs.mkdirSync(dir, { recursive: true });
      }

      const data: Record<string, any> = {};
      for (const [id, spreadsheet] of this.registry) {
        data[id] = spreadsheet;
      }

      fs.writeFileSync(
        this.registryPath,
        JSON.stringify(data, null, 2),
        "utf8"
      );
    } catch (error) {
      this.logger.error(
        "Error saving registry",
        error instanceof Error ? error.message : String(error)
      );
    }
  }

  /**
   * Register new spreadsheet
   */
  registerSpreadsheet(
    spreadsheet: Omit<
      RegisteredSpreadsheet,
      "id" | "version" | "lastModifiedAt"
    >
  ): RegisteredSpreadsheet {
    const id = `sheet_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`;

    const registered: RegisteredSpreadsheet = {
      ...spreadsheet,
      id,
      version: 1,
      lastModifiedAt: new Date(),
    };

    this.registry.set(id, registered);
    this.saveRegistry();

    this.logger.info(
      `✅ Registered spreadsheet: ${registered.spreadsheetName}`
    );
    return registered;
  }

  /**
   * Get registered spreadsheet by ID
   */
  getSpreadsheet(id: string): RegisteredSpreadsheet | null {
    return this.registry.get(id) || null;
  }

  /**
   * Get all registered spreadsheets for a user
   */
  getSpreadsheetsByUser(userId: string): RegisteredSpreadsheet[] {
    return Array.from(this.registry.values()).filter(
      (s) => s.createdBy === userId || s.lastModifiedBy === userId
    );
  }

  /**
   * Get spreadsheet by purpose
   */
  getSpreadsheetByPurpose(
    purpose: string,
    userId?: string
  ): RegisteredSpreadsheet | null {
    for (const spreadsheet of this.registry.values()) {
      if (spreadsheet.purpose === purpose) {
        if (!userId || spreadsheet.createdBy === userId) {
          return spreadsheet;
        }
      }
    }
    return null;
  }

  /**
   * Update spreadsheet
   */
  updateSpreadsheet(
    id: string,
    updates: Partial<RegisteredSpreadsheet>,
    userId: string
  ): boolean {
    const spreadsheet = this.registry.get(id);
    if (!spreadsheet) return false;

    // Check version conflict
    if (updates.version && updates.version !== spreadsheet.version) {
      this.logger.warn(
        `Version conflict: expected ${spreadsheet.version}, got ${updates.version}`
      );
      return false;
    }

    // Update with new version
    const updated: RegisteredSpreadsheet = {
      ...spreadsheet,
      ...updates,
      version: spreadsheet.version + 1,
      lastModifiedAt: new Date(),
      lastModifiedBy: userId,
    };

    this.registry.set(id, updated);
    this.saveRegistry();

    this.logger.info(`✅ Updated spreadsheet: ${id}`);
    return true;
  }

  /**
   * Lock spreadsheet for editing
   */
  lockSpreadsheet(id: string, userId: string): boolean {
    const spreadsheet = this.registry.get(id);
    if (!spreadsheet) return false;

    const lockUntil = new Date(Date.now() + 5 * 60 * 1000); // 5 minutes

    return this.updateSpreadsheet(
      id,
      {
        isBeingEdited: true,
        editingLockedUntil: lockUntil,
      },
      userId
    );
  }

  /**
   * Unlock spreadsheet
   */
  unlockSpreadsheet(id: string, userId: string): boolean {
    const spreadsheet = this.registry.get(id);
    if (!spreadsheet) return false;

    return this.updateSpreadsheet(
      id,
      {
        isBeingEdited: false,
        editingLockedUntil: null,
      },
      userId
    );
  }

  /**
   * Check if spreadsheet is locked
   */
  isLocked(id: string): boolean {
    const spreadsheet = this.registry.get(id);
    if (!spreadsheet) return false;

    if (!spreadsheet.isBeingEdited) return false;

    // Check if lock expired
    if (
      spreadsheet.editingLockedUntil &&
      spreadsheet.editingLockedUntil < new Date()
    ) {
      this.unlockSpreadsheet(id, "system");
      return false;
    }

    return true;
  }

  /**
   * Update last used timestamp
   */
  recordUsage(id: string): void {
    const spreadsheet = this.registry.get(id);
    if (spreadsheet) {
      spreadsheet.lastUsedAt = new Date();
      this.saveRegistry();
    }
  }

  /**
   * Delete spreadsheet registration
   */
  deleteSpreadsheet(id: string): boolean {
    const deleted = this.registry.delete(id);
    if (deleted) {
      this.saveRegistry();
      this.logger.info(`✅ Deleted spreadsheet registration: ${id}`);
    }
    return deleted;
  }

  /**
   * Get all registrations
   */
  getAllSpreadsheets(): RegisteredSpreadsheet[] {
    return Array.from(this.registry.values());
  }
}

// Singleton instance
let registryInstance: SpreadsheetRegistryService | null = null;

export function getSpreadsheetRegistry(): SpreadsheetRegistryService {
  if (!registryInstance) {
    registryInstance = new SpreadsheetRegistryService();
  }
  return registryInstance;
}
