import Database from "better-sqlite3";
import path from "path";
import fs from "fs";
import { getLogger } from "./logger";
import { tenantContext } from "./tenantContext";

const logger = getLogger("JobDatabase");

/**
 * PER-TENANT JOB DATABASES
 * ========================
 *
 * ARCHITECTURE: Separate database file per tenant
 * - Eliminates lock contention between tenants
 * - Allows parallel job processing
 * - Better performance and scalability
 *
 * Structure:
 *   backend/config/databases/
 *   ├── {tenantId}_jobs.db (tenant's jobs, auto-functions, history)
 *   └── ...
 *
 * Benefits:
 *   ✅ Zero cross-tenant contention
 *   ✅ Parallel writes by different tenants
 *   ✅ Better scalability
 *   ✅ Aligns with "NO SHARED SYSTEM" principle
 *
 * Tenant ID determines database file (use FULL tenant ID from tenants.json!):
 *   getTenantJobDatabase('yumna_bertigamart') → yumna_bertigamart_jobs.db
 *   getTenantJobDatabase('tika_nusseyba')     → tika_nusseyba_jobs.db
 */

// Ensure databases directory exists
const dbDir = path.join(process.cwd(), "config/databases");
if (!fs.existsSync(dbDir)) {
  fs.mkdirSync(dbDir, { recursive: true });
}

// Cache for tenant databases (connection pooling)
const tenantDatabases: Map<string, Database.Database> = new Map();

/**
 * Execute SQLite operation with retry on locking errors
 * Handles SQLITE_BUSY, SQLITE_LOCKED, and I/O errors gracefully
 */
export function withRetry<T>(
  operation: () => T,
  options: {
    maxRetries?: number;
    baseDelay?: number;
    operationName?: string;
  } = {},
): T {
  const {
    maxRetries = 3,
    baseDelay = 100,
    operationName = "SQLite operation",
  } = options;
  let lastError: Error | null = null;

  for (let attempt = 1; attempt <= maxRetries; attempt++) {
    try {
      return operation();
    } catch (error: any) {
      lastError = error;
      const errorMsg = error?.message || String(error);

      // Check if it's a retryable error (locking, I/O, protocol, file access)
      const isRetryable =
        errorMsg.includes("locking protocol") ||
        errorMsg.includes("SQLITE_BUSY") ||
        errorMsg.includes("SQLITE_LOCKED") ||
        errorMsg.includes("SQLITE_PROTOCOL") ||
        errorMsg.includes("disk I/O error") ||
        errorMsg.includes("database is locked") ||
        errorMsg.includes("unable to open database file");

      if (!isRetryable || attempt === maxRetries) {
        throw error;
      }

      // Exponential backoff with jitter (longer delays for file access issues)
      const delay = baseDelay * Math.pow(2, attempt - 1) + Math.random() * 100;
      logger.warn(
        `⚠️ [${operationName}] Retry ${attempt}/${maxRetries} after ${delay.toFixed(
          0,
        )}ms: ${errorMsg}`,
      );

      // Synchronous sleep (better-sqlite3 is synchronous anyway)
      const start = Date.now();
      while (Date.now() - start < delay) {
        // Busy wait - acceptable for short delays in synchronous context
      }
    }
  }

  throw lastError;
}

/**
 * Force reconnect for a tenant (clear cached connection)
 * Useful when connection becomes stale
 */
export function reconnectTenantJobDatabase(
  tenantId: string,
): Database.Database {
  // Close existing connection if any
  if (tenantDatabases.has(tenantId)) {
    try {
      const oldDb = tenantDatabases.get(tenantId);
      oldDb?.close();
    } catch (e) {
      // Ignore close errors
    }
    tenantDatabases.delete(tenantId);
  }

  // Wait a bit before reconnecting to allow locks to be released
  const waitMs = 500;
  const start = Date.now();
  while (Date.now() - start < waitMs) {
    // Brief wait for file system to settle
  }

  // Create fresh connection with retry
  return withRetry(() => createNewDatabaseConnection(tenantId), {
    maxRetries: 5,
    baseDelay: 500,
    operationName: `Reconnect ${tenantId}`,
  });
}

/**
 * Get database connection for specific tenant
 * Creates new connection if not cached
 * ✅ Per-tenant isolation
 * ✅ Connection reuse (no recreation)
 * ✅ Automatic initialization
 *
 * ⚠️ THROWS ERROR for 'system' tenant - it uses GlobalConfig only, no jobs database!
 */
export function getTenantJobDatabase(tenantId?: string): Database.Database {
  if (!tenantId) {
    tenantId = tenantContext.getTenantId(); // Default from config
  }

  // ⛔ STRICT: system tenant does NOT have a jobs database per AGENTS.MD
  if (tenantId === "system" || tenantId === "default") {
    throw new Error(
      `❌ INVALID: '${tenantId}' is not a real tenant and cannot have a jobs database. ` +
        `Use getTenantsWithJobDatabases() to get valid tenants.`,
    );
  }

  // Return cached connection if exists and healthy
  if (tenantDatabases.has(tenantId)) {
    const cachedDb = tenantDatabases.get(tenantId)!;
    // Quick health check - use lightweight query instead of integrity_check
    try {
      cachedDb.prepare("SELECT 1").get();
      return cachedDb;
    } catch (e) {
      logger.warn(`⚠️ Stale connection for ${tenantId}, will reconnect...`);
      tenantDatabases.delete(tenantId);
      try {
        cachedDb.close();
      } catch (_) {
        // Ignore close errors
      }
      // Wait before trying to reconnect
      const waitMs = 300;
      const start = Date.now();
      while (Date.now() - start < waitMs) {
        // Brief wait
      }
    }
  }

  // Create new connection with retry for this tenant
  return withRetry(() => createNewDatabaseConnection(tenantId!), {
    maxRetries: 5,
    baseDelay: 500,
    operationName: `Connect ${tenantId}`,
  });
}

/**
 * Internal: Create a fresh database connection (no caching logic)
 * Optimized for Docker/WSL environments with better WAL handling
 */
function createNewDatabaseConnection(tenantId: string): Database.Database {
  const dbPath = path.join(dbDir, `${tenantId}_jobs.db`);
  const shmPath = `${dbPath}-shm`;
  const walPath = `${dbPath}-wal`;

  // Clean up stale WAL files before opening (prevents SQLITE_PROTOCOL errors)
  // This only helps if the files are from a crashed/killed process
  try {
    if (fs.existsSync(shmPath) && fs.existsSync(walPath)) {
      const walStats = fs.statSync(walPath);
      // If WAL is 0 bytes but SHM exists, it's likely stale
      if (walStats.size === 0) {
        logger.warn(`🧹 Cleaning stale WAL files for ${tenantId}`);
        fs.unlinkSync(shmPath);
        fs.unlinkSync(walPath);
      }
    }
  } catch (e) {
    // Ignore cleanup errors - the main open will fail if truly corrupted
  }

  const db = new Database(dbPath);

  // Optimize for Docker/WSL environment with file locking issues
  // Using DELETE mode instead of WAL to avoid cross-process locking issues
  // WAL mode requires shared memory which can be problematic in Docker with
  // bind-mounted volumes on Windows/WSL
  db.pragma("journal_mode = DELETE"); // DELETE mode is more reliable in Docker
  db.pragma("foreign_keys = ON");
  db.pragma("synchronous = NORMAL");
  db.pragma("cache_size = -16000"); // 16MB cache (smaller = less memory pressure)
  db.pragma("busy_timeout = 30000"); // Wait 30s on locks instead of failing
  db.pragma("temp_store = MEMORY"); // Keep temp tables in memory

  logger.info(
    `✅ Initialized job database for tenant: ${tenantId} (${dbPath})`,
  );

  // Initialize schema
  initializeJobDatabaseSchema(db);

  // Cache the connection
  tenantDatabases.set(tenantId, db);

  return db;
}

/**
 * Initialize database schema (used by per-tenant database)
 */
function initializeJobDatabaseSchema(db: Database.Database): void {
  try {
    // Create jobs table (NO tenant_id column needed - database per tenant!)
    db.exec(`
      CREATE TABLE IF NOT EXISTS jobs (
        id TEXT PRIMARY KEY,
        type TEXT NOT NULL,
        status TEXT NOT NULL CHECK(status IN ('pending', 'running', 'completed', 'failed', 'cancelled')),
        priority TEXT DEFAULT 'normal' CHECK(priority IN ('low', 'normal', 'high')),
        data TEXT NOT NULL,
        error_message TEXT,
        started_at DATETIME,
        completed_at DATETIME,
        created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
        updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
      );
    `);

    // Create job_history table (NO tenant_id needed - per-tenant database)
    db.exec(`
      CREATE TABLE IF NOT EXISTS job_history (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        job_id TEXT NOT NULL,
        job_type TEXT,
        status TEXT NOT NULL,
        error_message TEXT,
        duration_ms INTEGER,
        started_at DATETIME,
        completed_at DATETIME,
        created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
        FOREIGN KEY (job_id) REFERENCES jobs(id) ON DELETE CASCADE
      );
    `);

    // Create auto_functions_config table (per-tenant, no tenant_id column)
    db.exec(`
      CREATE TABLE IF NOT EXISTS auto_functions_config (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        name TEXT NOT NULL UNIQUE,
        enabled BOOLEAN DEFAULT 0,
        interval_minutes INTEGER DEFAULT 30,
        start_time TEXT,
        end_time TEXT,
        last_executed DATETIME,
        next_scheduled_execution DATETIME,
        created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
        updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
      );
    `);

    // Create auto_functions_history table (per-tenant, no tenant_id)
    db.exec(`
      CREATE TABLE IF NOT EXISTS auto_functions_history (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        function_name TEXT NOT NULL,
        executed_at DATETIME DEFAULT CURRENT_TIMESTAMP,
        status TEXT NOT NULL CHECK(status IN ('success', 'failed')),
        error_message TEXT,
        duration_ms INTEGER
      );
    `);

    // Create route_execution_config table (per-tenant, no tenant_id)
    db.exec(`
      CREATE TABLE IF NOT EXISTS route_execution_config (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        route_key TEXT NOT NULL UNIQUE,
        route_name TEXT NOT NULL,
        description TEXT,
        execution_mode TEXT NOT NULL DEFAULT 'direct' CHECK(execution_mode IN ('queue', 'direct')),
        priority TEXT DEFAULT 'normal' CHECK(priority IN ('low', 'normal', 'high')),
        enabled INTEGER DEFAULT 1,
        icon TEXT DEFAULT '📋',
        category TEXT DEFAULT 'general',
        created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
        updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
      );
    `);

    // Create indexes for performance
    db.exec(`
      CREATE INDEX IF NOT EXISTS idx_jobs_status ON jobs(status);
      CREATE INDEX IF NOT EXISTS idx_jobs_created_at ON jobs(created_at);
      CREATE INDEX IF NOT EXISTS idx_job_history_job_id ON job_history(job_id);
      CREATE INDEX IF NOT EXISTS idx_auto_functions_enabled ON auto_functions_config(enabled);
      CREATE INDEX IF NOT EXISTS idx_auto_functions_name ON auto_functions_config(name);
      CREATE INDEX IF NOT EXISTS idx_route_execution_config_key ON route_execution_config(route_key);
      CREATE INDEX IF NOT EXISTS idx_route_execution_config_category ON route_execution_config(category);
    `);

    logger.debug("✅ Job database schema initialized");
  } catch (error) {
    logger.error(`❌ Failed to initialize job database schema: ${error}`);
    throw error;
  }
}

/**
 * @deprecated Use getTenantJobDatabase(tenantId) instead
 * Kept for backward compatibility only
 */
export function getJobDatabase(): Database.Database {
  logger.warn(
    "⚠️ DEPRECATED: getJobDatabase() - use getTenantJobDatabase(tenantId) instead",
  );
  return getTenantJobDatabase(tenantContext.getTenantId());
}

/**
 * @deprecated Use getTenantJobDatabase(tenantId) instead
 */
export function initJobDatabase(): Database.Database {
  logger.warn(
    "⚠️ DEPRECATED: initJobDatabase() - use getTenantJobDatabase(tenantId) instead",
  );
  return getTenantJobDatabase(tenantContext.getTenantId());
}

/**
 * Close all tenant database connections
 * Call on server shutdown - ensures clean checkpoint for WAL mode databases
 */
export function closeAllTenantJobDatabases(): void {
  for (const [tenantId, db] of tenantDatabases.entries()) {
    try {
      // Ensure any pending WAL data is checkpointed before closing
      try {
        db.pragma("wal_checkpoint(TRUNCATE)");
      } catch (e) {
        // Ignore checkpoint errors - may not be in WAL mode
      }
      db.close();
      tenantDatabases.delete(tenantId);
      logger.info(`✅ Closed job database for tenant: ${tenantId}`);
    } catch (error) {
      logger.error(`❌ Failed to close database for ${tenantId}: ${error}`);
    }
  }
}

/**
 * @deprecated Use closeAllTenantJobDatabases() instead
 */
export function closeJobDatabase(): void {
  logger.warn(
    "⚠️ DEPRECATED: closeJobDatabase() - use closeAllTenantJobDatabases() instead",
  );
  closeAllTenantJobDatabases();
}
