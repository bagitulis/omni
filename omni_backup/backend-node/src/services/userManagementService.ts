/**
 * User Management Service
 * SRP: Facade for user lifecycle operations
 * Delegates to specialized services for specific operations
 */

import bcrypt from "bcryptjs";
import { getDbManager } from "./dbConnectionManager";
import { AuthService } from "./authService";
import { TenantManagementService } from "./tenantManagementService";
import { getLogger } from "../utils/logger";
import AuditService from "./auditService";
import { UserListingService } from "./userListingService";
import { UserPasswordService } from "./userPasswordService";
import {
  CreateUserRequest,
  UpdateUserRequest,
  UserResponse,
  UserWithTenant,
} from "./userManagementTypes";

const logger = getLogger("UserManagementService");

// Re-export types for backward compatibility
export { CreateUserRequest, UpdateUserRequest, UserResponse };

export class UserManagementService {
  private tenantService = new TenantManagementService();
  private listingService = new UserListingService();
  private passwordService = new UserPasswordService();

  /**
   * Create new user with automatic tenant setup
   */
  async createUser(data: CreateUserRequest, _createdByAdmin = false): Promise<UserResponse> {
    try {
      const tenantId = data.username;
      const shopName = data.shopName || `Shop ${data.username}`;

      logger.info(`👤 Creating user: ${data.username} (${data.email})`);

      if (!data.username || !data.email || !data.password) {
        throw new Error("Username, email, and password are required");
      }
      if (data.password.length < 6) {
        throw new Error("Password must be at least 6 characters");
      }

      // Auto-setup tenant database if doesn't exist
      const exists = await this.tenantService.databaseExists(tenantId);
      if (!exists) {
        logger.info(`📦 Provisioning database for tenant: ${tenantId}`);
        await this.tenantService.setupTenantDatabase(tenantId, shopName);
      }
      this.tenantService.setupJobDatabase(tenantId);

      const prisma = getDbManager().getConnection(tenantId);
      const authService = new AuthService(prisma);

      const existingUser = await prisma.user.findFirst({
        where: { OR: [{ username: data.username }, { email: data.email }] },
      });
      if (existingUser) throw new Error("Username or email already exists");

      const user = await authService.register({
        username: data.username,
        email: data.email,
        password: data.password,
      });

      await AuditService.logAction({
        action: "USER_CREATED",
        userId: "ADMIN",
        targetUserId: user.id,
        targetTenantId: tenantId,
        details: { username: data.username, email: data.email, role: user.role },
        status: "success",
      });

      logger.info(`✅ User created: ${data.username} (tenantId: ${tenantId})`);
      return { id: user.id, username: user.username, email: user.email, role: user.role, createdAt: new Date(), updatedAt: new Date() };
    } catch (error: any) {
      logger.error(`❌ Failed to create user: ${error.message}`);
      await AuditService.logAction({ action: "USER_CREATED", userId: "ADMIN", targetTenantId: data.username, details: { email: data.email }, status: "failed", errorMessage: error.message });
      throw error;
    }
  }

  /**
   * Delete user and cleanup all associated resources
   */
  async deleteUser(tenantId: string): Promise<void> {
    try {
      logger.warn(`🗑️ Deleting user: ${tenantId}`);
      try {
        const prisma = getDbManager().getConnection(tenantId);
        await prisma.user.deleteMany({});
        await prisma.$disconnect();
      } catch { logger.warn(`Could not clean database records`); }

      await this.tenantService.deleteTenant(tenantId);
      await AuditService.logAction({ action: "USER_DELETED", userId: "ADMIN", targetTenantId: tenantId, status: "success" });
      logger.info(`✅ User and tenant deleted: ${tenantId}`);
    } catch (error: any) {
      logger.error(`❌ Failed to delete user: ${error.message}`);
      await AuditService.logAction({ action: "USER_DELETED", userId: "ADMIN", targetTenantId: tenantId, status: "failed", errorMessage: error.message });
      throw error;
    }
  }

  /**
   * Update user information
   */
  async updateUser(tenantId: string, data: UpdateUserRequest): Promise<UserResponse> {
    try {
      logger.info(`✏️ Updating user: ${tenantId}`);
      const prisma = getDbManager().getConnection(tenantId);

      const user = await prisma.user.findFirst();
      if (!user) throw new Error("User not found");

      const updateData: any = {};
      if (data.email) {
        const existingEmail = await prisma.user.findUnique({ where: { email: data.email } });
        if (existingEmail && existingEmail.id !== user.id) throw new Error("Email already in use");
        updateData.email = data.email;
      }
      if (data.password) {
        if (data.password.length < 6) throw new Error("Password must be at least 6 characters");
        updateData.password = await bcrypt.hash(data.password, 10);
      }
      if (data.role) updateData.role = data.role;

      const updatedUser = await prisma.user.update({ where: { id: user.id }, data: updateData });
      await AuditService.logAction({ action: "USER_UPDATED", userId: "ADMIN", targetTenantId: tenantId, targetUserId: user.id, details: updateData, status: "success" });

      logger.info(`✅ User updated: ${tenantId}`);
      return { id: updatedUser.id, username: updatedUser.username, email: updatedUser.email, role: updatedUser.role, createdAt: new Date(), updatedAt: new Date() };
    } catch (error: any) {
      logger.error(`❌ Failed to update user: ${error.message}`);
      await AuditService.logAction({ action: "USER_UPDATED", userId: "ADMIN", targetTenantId: tenantId, status: "failed", errorMessage: error.message });
      throw error;
    }
  }

  /**
   * Get user by tenant ID
   */
  async getUser(tenantId: string): Promise<UserResponse | null> {
    try {
      const prisma = getDbManager().getConnection(tenantId);
      const user = await prisma.user.findFirst({ where: { username: tenantId } });
      if (!user) return null;
      return { id: user.id, username: user.username, email: user.email, role: user.role, createdAt: new Date(), updatedAt: new Date() };
    } catch (error: any) {
      logger.error(`Failed to get user: ${error.message}`);
      return null;
    }
  }

  // Delegate to specialized services
  async listUsersForRole(tenantId: string, userRole: string): Promise<UserWithTenant[]> {
    return this.listingService.listUsersForRole(tenantId, userRole);
  }

  async listAllUsers(): Promise<UserWithTenant[]> {
    return this.listingService.listAllUsers();
  }

  async changePassword(tenantId: string, userId: string, currentPassword: string, newPassword: string): Promise<void> {
    return this.passwordService.changePassword(tenantId, userId, currentPassword, newPassword);
  }
}

export function getUserManagementService(): UserManagementService {
  return new UserManagementService();
}
