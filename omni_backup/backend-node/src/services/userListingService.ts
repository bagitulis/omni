/**
 * User Listing Service
 * SRP: List and query users across tenants
 */

import { getDbManager } from "./dbConnectionManager";
import { TenantManagementService } from "./tenantManagementService";
import { getLogger } from "../utils/logger";
import { UserWithTenant, getVisibleRoles } from "./userManagementTypes";

const logger = getLogger("UserListingService");

// Define safe user select to never fetch password
const SAFE_USER_SELECT = {
  id: true,
  username: true,
  email: true,
  role: true,
  createdAt: true,
  updatedAt: true,
  // NEVER include: password, failedLoginAttempts, accountLockedUntil, lastFailedLogin
};

export class UserListingService {
  private tenantService = new TenantManagementService();

  /**
   * List users based on requester's role and tenant
   */
  async listUsersForRole(
    tenantId: string,
    userRole: string
  ): Promise<UserWithTenant[]> {
    try {
      const visibleRoles = getVisibleRoles(userRole);

      // Developer can see all users from all tenants
      if (userRole === "developer") {
        return this.listAllUsers();
      }

      // For other roles, only show users from their tenant
      const allUsers: UserWithTenant[] = [];

      try {
        const prisma = getDbManager().getConnection(tenantId);
        const users = await prisma.user.findMany({
          where: { role: { in: visibleRoles } },
          select: SAFE_USER_SELECT,
        });

        for (const user of users) {
          allUsers.push({
            id: user.id,
            username: user.username,
            email: user.email,
            role: user.role,
            createdAt: user.createdAt || new Date(),
            updatedAt: user.updatedAt || new Date(),
            tenantId: user.username,
          });
        }
      } catch (error) {
        logger.warn(`Failed to load users for tenant ${tenantId}`);
      }

      return allUsers;
    } catch (error: any) {
      logger.error(`Failed to list users for role: ${error.message}`);
      return [];
    }
  }

  /**
   * List all users across all tenants (for developer only)
   */
  async listAllUsers(): Promise<UserWithTenant[]> {
    try {
      const allUsers: UserWithTenant[] = [];
      const processedDatabases = new Set<string>();
      const tenantConfig = this.tenantService.getAllTenants();

      for (const [tenantId, config] of Object.entries(tenantConfig)) {
        if (processedDatabases.has(config.dbPath)) continue;
        processedDatabases.add(config.dbPath);

        try {
          const prisma = getDbManager().getConnection(tenantId);
          const users = await prisma.user.findMany({
            select: SAFE_USER_SELECT,
          });

          for (const user of users) {
            allUsers.push({
              id: user.id,
              username: user.username,
              email: user.email,
              role: user.role,
              createdAt: user.createdAt || new Date(),
              updatedAt: user.updatedAt || new Date(),
              tenantId: user.username,
            });
          }
        } catch (error) {
          logger.warn(`Failed to load users from tenant ${tenantId}`);
        }
      }

      return allUsers;
    } catch (error: any) {
      logger.error(`Failed to list users: ${error.message}`);
      return [];
    }
  }
}
