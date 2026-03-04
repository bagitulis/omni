import fs from "fs";
import path from "path";
import { exec } from "child_process";
import { promisify } from "util";
import { getLogger } from "../utils/logger";

const execAsync = promisify(exec);
const logger = getLogger("TenantManagementService");

interface TenantConfig {
  dbPath: string;
  shopName: string;
  isGlobal?: boolean; // true for system tenant (GlobalConfig only, no jobs db)
}

interface TenantsConfig {
  [tenantId: string]: TenantConfig;
}

/**
 * TenantManagementService - Manages multi-tenant provisioning
 *
 * Responsibilities:
 *   ✅ Setup new tenant databases (create files + migrations)
 *   ✅ Delete tenant (cleanup databases)
 *   ✅ Check tenant existence
 *   ✅ Load/update tenants configuration
 */
export class TenantManagementService {
  private tenantsConfigPath = path.resolve("config/static/tenants.json");
  private databasesDir = path.resolve("config/databases");

  /**
   * Check if tenant database exists
   */
  async databaseExists(tenantId: string): Promise<boolean> {
    try {
      const config = this.loadTenantsConfig();
      if (!config[tenantId]) {
        return false;
      }

      const dbPath = path.resolve(config[tenantId].dbPath);
      return fs.existsSync(dbPath);
    } catch (error) {
      logger.error(`Failed to check database existence: ${error}`);
      return false;
    }
  }

  /**
   * Setup new tenant with Prisma database
   * Creates database file and runs migrations
   */
  async setupTenantDatabase(tenantId: string, shopName: string): Promise<void> {
    try {
      // 1. Validate tenant ID
      if (!tenantId || tenantId.length === 0) {
        throw new Error("Invalid tenant ID");
      }

      // 2. Check if already exists
      const config = this.loadTenantsConfig();
      if (config[tenantId]) {
        throw new Error(`Tenant ${tenantId} already exists`);
      }

      // 3. Determine database path
      const dbFileName = `${tenantId}_bertigamart.db`;
      const dbPath = path.join(this.databasesDir, dbFileName);

      logger.info(`🚀 Setting up database for tenant: ${tenantId}`);

      // 4. Create directory if needed
      if (!fs.existsSync(this.databasesDir)) {
        fs.mkdirSync(this.databasesDir, { recursive: true });
        logger.info(`📁 Created databases directory`);
      }

      // 5. Create empty database file
      if (!fs.existsSync(dbPath)) {
        fs.writeFileSync(dbPath, "");
        logger.info(`📄 Created database file: ${dbPath}`);
      }

      // 6. Run Prisma migrations
      await this.runMigrations(dbPath);

      // 7. Update tenants.json
      this.addTenantToConfig(tenantId, dbPath, shopName);

      logger.info(`✅ Tenant ${tenantId} database setup complete`);
    } catch (error: any) {
      logger.error(`❌ Failed to setup tenant database: ${error.message}`);
      throw error;
    }
  }

  /**
   * Setup job database for tenant
   * Auto-called by getTenantJobDatabase if doesn't exist
   */
  setupJobDatabase(tenantId: string): void {
    const jobDbPath = path.join(this.databasesDir, `${tenantId}_jobs.db`);

    if (!fs.existsSync(jobDbPath)) {
      fs.writeFileSync(jobDbPath, "");
      logger.info(`📄 Created job database: ${jobDbPath}`);
    }
  }

  /**
   * Delete tenant completely (cleanup both databases)
   */
  async deleteTenant(tenantId: string): Promise<void> {
    try {
      const config = this.loadTenantsConfig();

      if (!config[tenantId]) {
        throw new Error(`Tenant ${tenantId} not found`);
      }

      logger.info(`🗑️ Deleting tenant: ${tenantId}`);

      // 1. Delete Prisma database file
      const dbPath = path.resolve(config[tenantId].dbPath);
      if (fs.existsSync(dbPath)) {
        fs.unlinkSync(dbPath);
        logger.info(`🗑️ Deleted Prisma database: ${dbPath}`);
      }

      // 2. Delete job database file
      const jobDbPath = path.join(this.databasesDir, `${tenantId}_jobs.db`);
      if (fs.existsSync(jobDbPath)) {
        fs.unlinkSync(jobDbPath);
        logger.info(`🗑️ Deleted job database: ${jobDbPath}`);
      }

      // 3. Delete from tenants.json
      this.removeTenantFromConfig(tenantId);

      logger.info(`✅ Tenant ${tenantId} deleted successfully`);
    } catch (error: any) {
      logger.error(`❌ Failed to delete tenant: ${error.message}`);
      throw error;
    }
  }

  /**
   * Get all tenant IDs (includes system for backward compatibility)
   * ⚠️ For operations requiring real tenants, use getRealTenantIds() instead!
   */
  getAllTenantIds(): string[] {
    try {
      const config = this.loadTenantsConfig();
      return Object.keys(config);
    } catch (error) {
      logger.error(`Failed to get tenant IDs: ${error}`);
      return []; // Return empty array, no hardcoded fallback
    }
  }

  /**
   * Get only REAL tenant IDs (excludes 'system' and global tenants)
   * Use this for job processing, auto-functions, and operations requiring real tenants
   *
   * Real tenants: yumna_bertigamart, tika_nusseyba
   * NOT included: system (GlobalConfig only), default
   */
  getRealTenantIds(): string[] {
    try {
      const config = this.loadTenantsConfig();
      return Object.keys(config).filter((id) => {
        if (id === "system" || id === "default") return false;
        if (config[id]?.isGlobal) return false;
        return true;
      });
    } catch (error) {
      logger.error(`Failed to get real tenant IDs: ${error}`);
      return [];
    }
  }

  /**
   * Get all tenants configuration
   */
  getAllTenants(): TenantsConfig {
    try {
      return this.loadTenantsConfig();
    } catch (error) {
      logger.error(`Failed to load tenants config: ${error}`);
      return {};
    }
  }

  /**
   * Load tenants configuration from JSON
   * No hardcoded fallback - returns empty config if not found
   */
  private loadTenantsConfig(): TenantsConfig {
    try {
      if (!fs.existsSync(this.tenantsConfigPath)) {
        logger.warn(`Tenants config not found at ${this.tenantsConfigPath}`);
        return {};
      }

      const content = fs.readFileSync(this.tenantsConfigPath, "utf-8");
      return JSON.parse(content);
    } catch (error) {
      logger.error(`Failed to load tenants config: ${error}`);
      return {};
    }
  }

  /**
   * Add tenant to configuration
   */
  private addTenantToConfig(
    tenantId: string,
    dbPath: string,
    shopName: string,
  ): void {
    try {
      const config = this.loadTenantsConfig();
      config[tenantId] = {
        dbPath,
        shopName,
      };

      fs.writeFileSync(this.tenantsConfigPath, JSON.stringify(config, null, 2));

      logger.info(`✅ Updated tenants.json with ${tenantId}`);
    } catch (error) {
      logger.error(`Failed to update tenants config: ${error}`);
      throw error;
    }
  }

  /**
   * Remove tenant from configuration
   */
  private removeTenantFromConfig(tenantId: string): void {
    try {
      const config = this.loadTenantsConfig();
      delete config[tenantId];

      fs.writeFileSync(this.tenantsConfigPath, JSON.stringify(config, null, 2));

      logger.info(`✅ Removed ${tenantId} from tenants.json`);
    } catch (error) {
      logger.error(`Failed to update tenants config: ${error}`);
      throw error;
    }
  }

  /**
   * Run Prisma migrations for specific database
   */
  private async runMigrations(dbPath: string): Promise<void> {
    try {
      logger.info(`🔧 Running migrations for ${dbPath}...`);

      const env = {
        ...process.env,
        DATABASE_URL: `file:${dbPath}`,
      };

      const { stdout } = await execAsync("npx prisma db push --skip-generate", {
        env,
        cwd: process.cwd(),
      });

      logger.info(`✅ Migrations completed for ${dbPath}`);
      if (stdout) logger.info(stdout);
    } catch (error: any) {
      logger.error(`❌ Migration failed: ${error.message}`);
      throw error;
    }
  }
}

export function getTenantManagementService(): TenantManagementService {
  return new TenantManagementService();
}
