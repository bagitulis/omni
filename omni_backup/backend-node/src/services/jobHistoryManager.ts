import { getTenantJobDatabase } from "../utils/jobDb";
import { tenantContext } from "../utils/tenantContext";
import { getLogger } from "../utils/logger";
import { IJobHistory, JobStatus } from "../types/job.types";

const logger = getLogger("JobHistoryManager");

/**
 * Job History Manager - Tenant-aware job history tracking
 * Uses TenantContext for automatic per-tenant isolation
 */
export class JobHistoryManager {
  /**
   * Add history for Direct Mode actions (not from job queue)
   * Creates a dummy job entry first to satisfy foreign key constraint
   */
  addDirectHistory(
    jobId: string,
    jobType: string,
    status: JobStatus,
    errorMessage?: string,
    durationMs?: number
  ): void {
    const tenantId = tenantContext.getTenantId();
    const db = getTenantJobDatabase(tenantId);

    try {
      const now = new Date().toISOString();

      // Insert dummy job to satisfy foreign key
      const insertJobStmt = db.prepare(`
        INSERT OR IGNORE INTO jobs (id, type, status, priority, data, created_at, started_at, completed_at)
        VALUES (?, ?, ?, ?, ?, ?, ?, ?)
      `);
      insertJobStmt.run(
        jobId,
        jobType,
        status,
        "normal",
        JSON.stringify({ direct: true }),
        now,
        now,
        now
      );

      // Now insert history
      const historyStmt = db.prepare(`
        INSERT INTO job_history (job_id, job_type, status, error_message, duration_ms, started_at, completed_at)
        VALUES (?, ?, ?, ?, ?, ?, ?)
      `);
      historyStmt.run(
        jobId,
        jobType,
        status,
        errorMessage || null,
        durationMs || null,
        now,
        now
      );
    } catch (error) {
      logger.error(`❌ Failed to add direct history: ${error}`);
      throw error;
    }
  }

  addJobHistory(
    jobId: string,
    status: JobStatus,
    errorMessage?: string,
    durationMs?: number,
    jobType?: string
  ): void {
    const tenantId = tenantContext.getTenantId();
    const db = getTenantJobDatabase(tenantId);

    try {
      const jobStmt = db.prepare(
        "SELECT type, started_at, completed_at FROM jobs WHERE id = ?"
      );
      const jobRow = jobStmt.get(jobId) as any;

      const typeToSave = jobType || jobRow?.type || null;
      const startedAt = jobRow?.started_at || null;
      const completedAt = jobRow?.completed_at || null;

      const stmt = db.prepare(`
        INSERT INTO job_history (job_id, job_type, status, error_message, duration_ms, started_at, completed_at)
        VALUES (?, ?, ?, ?, ?, ?, ?)
      `);

      stmt.run(
        jobId,
        typeToSave,
        status,
        errorMessage || null,
        durationMs || null,
        startedAt,
        completedAt
      );
    } catch (error) {
      logger.error(`❌ Failed to add job history: ${error}`);
      throw error;
    }
  }

  getJobHistory(jobId?: string, limit: number = 100): IJobHistory[] {
    const tenantId = tenantContext.getTenantId();
    const db = getTenantJobDatabase(tenantId);

    try {
      let query = "SELECT * FROM job_history";
      const params: any[] = [];

      if (jobId) {
        query += " WHERE job_id = ?";
        params.push(jobId);
      }

      query += " ORDER BY created_at DESC LIMIT ?";
      params.push(limit);

      const stmt = db.prepare(query);
      const rows = stmt.all(...params) as any[];

      return rows.map((row) => this.mapRowToHistory(row));
    } catch (error) {
      logger.error(`❌ Failed to get job history: ${error}`);
      return [];
    }
  }

  /**
   * Get paginated job history with filters
   */
  getJobHistoryPaginated(options: {
    page?: number;
    pageSize?: number;
    status?: string;
    jobType?: string;
    search?: string;
    sortBy?: string;
    sortOrder?: "asc" | "desc";
  }): {
    data: IJobHistory[];
    total: number;
    page: number;
    pageSize: number;
    totalPages: number;
  } {
    const tenantId = tenantContext.getTenantId();
    const db = getTenantJobDatabase(tenantId);

    const page = options.page || 1;
    const pageSize = options.pageSize || 20;
    const offset = (page - 1) * pageSize;
    const sortBy = options.sortBy || "created_at";
    const sortOrder = options.sortOrder || "desc";

    // Whitelist valid sort columns
    const validColumns = [
      "id",
      "job_id",
      "job_type",
      "status",
      "duration_ms",
      "created_at",
      "completed_at",
    ];
    const sortColumn = validColumns.includes(sortBy) ? sortBy : "created_at";

    try {
      let whereClause = "";
      const params: any[] = [];

      const conditions: string[] = [];

      if (options.status && options.status !== "all") {
        conditions.push("status = ?");
        params.push(options.status);
      }

      if (options.jobType && options.jobType !== "all") {
        conditions.push("job_type = ?");
        params.push(options.jobType);
      }

      if (options.search) {
        conditions.push(
          "(job_id LIKE ? OR job_type LIKE ? OR error_message LIKE ?)"
        );
        const searchPattern = `%${options.search}%`;
        params.push(searchPattern, searchPattern, searchPattern);
      }

      if (conditions.length > 0) {
        whereClause = " WHERE " + conditions.join(" AND ");
      }

      // Get total count
      const countStmt = db.prepare(
        `SELECT COUNT(*) as total FROM job_history${whereClause}`
      );
      const countResult = countStmt.get(...params) as { total: number };
      const total = countResult.total;

      // Get paginated data
      const dataQuery = `SELECT * FROM job_history${whereClause} ORDER BY ${sortColumn} ${sortOrder.toUpperCase()} LIMIT ? OFFSET ?`;
      const dataParams = [...params, pageSize, offset];
      const stmt = db.prepare(dataQuery);
      const rows = stmt.all(...dataParams) as any[];

      return {
        data: rows.map((row) => this.mapRowToHistory(row)),
        total,
        page,
        pageSize,
        totalPages: Math.ceil(total / pageSize),
      };
    } catch (error) {
      logger.error(`❌ Failed to get paginated job history: ${error}`);
      return { data: [], total: 0, page: 1, pageSize, totalPages: 0 };
    }
  }

  /**
   * Get distinct job types for filter dropdown
   */
  getDistinctJobTypes(): string[] {
    const tenantId = tenantContext.getTenantId();
    const db = getTenantJobDatabase(tenantId);

    try {
      const stmt = db.prepare(
        "SELECT DISTINCT job_type FROM job_history WHERE job_type IS NOT NULL ORDER BY job_type"
      );
      const rows = stmt.all() as { job_type: string }[];
      return rows.map((r) => r.job_type);
    } catch (error) {
      logger.error(`❌ Failed to get distinct job types: ${error}`);
      return [];
    }
  }

  clearJobHistory(): number {
    const tenantId = tenantContext.getTenantId();
    const db = getTenantJobDatabase(tenantId);

    try {
      const countStmt = db.prepare("SELECT COUNT(*) as count FROM job_history");
      const countResult = countStmt.get() as any;
      const count = countResult?.count || 0;

      const deleteStmt = db.prepare("DELETE FROM job_history");
      deleteStmt.run();

      logger.info(`✅ Cleared job history: ${count} records deleted`);
      return count;
    } catch (error) {
      logger.error(`❌ Failed to clear job history: ${error}`);
      throw error;
    }
  }

  private mapRowToHistory(row: any): IJobHistory {
    return {
      id: row.id,
      jobId: row.job_id,
      jobType: row.job_type,
      status: row.status,
      errorMessage: row.error_message,
      durationMs: row.duration_ms,
      startedAt: row.started_at ? new Date(row.started_at) : undefined,
      completedAt: row.completed_at ? new Date(row.completed_at) : undefined,
      createdAt: new Date(row.created_at),
    };
  }
}
