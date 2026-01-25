import { Response } from "express";
import { AuthRequest } from "../middleware/tenantMiddleware";
import { getUserManagementService } from "../services/userManagementService";
import { getLogger } from "../utils/logger";

const logger = getLogger("UserManagementController");

/**
 * User Management Controller
 * Single Responsibility: Handle user CRUD operations (list, create, edit, delete)
 *
 * Access Control:
 *   - Developer: Can see all users from all tenants
 *   - Owner: Can see users in own tenant (admin, user roles only)
 *   - Admin: Can see users in own tenant (user role only)
 *   - User: Can only see self
 */
export class UserManagementController {
  private userService = getUserManagementService();

  /**
   * GET /api/admin/users
   * List users based on role and tenant context
   */
  async listAllUsers(req: AuthRequest, res: Response): Promise<void> {
    try {
      const userRole = req.userRole || "user";
      const tenantId = req.tenantId || "";
      
      logger.info(`📋 Listing users for ${userRole} in tenant ${tenantId}`);

      const users = await this.userService.listUsersForRole(tenantId, userRole);

      res.json({
        success: true,
        data: {
          users,
          total: users.length,
        },
      });
    } catch (error: any) {
      logger.error(`❌ Failed to list users: ${error.message}`);
      res.status(500).json({
        success: false,
        error: error.message,
      });
    }
  }

  /**
   * POST /api/admin/users
   * Create new user with auto-provisioning
   *
   * Body:
   *   - username (required): tenant ID + username
   *   - email (required): user email
   *   - password (required): min 6 chars
   *   - shopName (optional): shop display name
   */
  async createUser(req: AuthRequest, res: Response): Promise<void> {
    try {
      const { username, email, password, shopName } = req.body;

      logger.info(`👤 Creating new user: ${username}`);

      // Validation
      if (!username || !email || !password) {
        res.status(400).json({
          success: false,
          error: "Username, email, and password are required",
        });
        return;
      }

      if (password.length < 6) {
        res.status(400).json({
          success: false,
          error: "Password must be at least 6 characters",
        });
        return;
      }

      // Create user with auto-provisioning
      const user = await this.userService.createUser({
        username,
        email,
        password,
        shopName,
      });

      logger.info(`✅ User created: ${username}`);
      res.status(201).json({
        success: true,
        data: user,
      });
    } catch (error: any) {
      logger.error(`❌ Failed to create user: ${error.message}`);
      res.status(400).json({
        success: false,
        error: error.message,
      });
    }
  }

  /**
   * PATCH /api/admin/users/:tenantId
   * Update user (password, email, role)
   *
   * Body (any combination):
   *   - email: new email
   *   - password: new password (min 6 chars)
   *   - role: new role (owner, admin, user)
   */
  async updateUser(req: AuthRequest, res: Response): Promise<void> {
    try {
      const { tenantId } = req.params;
      const updateData = req.body;

      // At least one field must be provided
      if (!Object.keys(updateData).length) {
        res.status(400).json({
          success: false,
          error: "At least one field (email, password, role) required",
        });
        return;
      }

      logger.info(`✏️ Updating user: ${tenantId}`);

      const user = await this.userService.updateUser(tenantId, updateData);

      logger.info(`✅ User updated: ${tenantId}`);
      res.json({
        success: true,
        data: user,
      });
    } catch (error: any) {
      logger.error(`❌ Failed to update user: ${error.message}`);
      res.status(400).json({
        success: false,
        error: error.message,
      });
    }
  }

  /**
   * DELETE /api/admin/users/:tenantId
   * Delete user and cleanup all resources
   *
   * Cleanup includes:
   *   - User database records
   *   - Prisma database file
   *   - Job database file
   *   - Config from tenants.json
   */
  async deleteUser(req: AuthRequest, res: Response): Promise<void> {
    try {
      const { tenantId } = req.params;

      logger.warn(`🗑️ Deleting user: ${tenantId}`);

      await this.userService.deleteUser(tenantId);

      logger.info(`✅ User deleted: ${tenantId}`);
      res.json({
        success: true,
        message: `User and all associated data deleted: ${tenantId}`,
      });
    } catch (error: any) {
      logger.error(`❌ Failed to delete user: ${error.message}`);
      res.status(400).json({
        success: false,
        error: error.message,
      });
    }
  }

  /**
   * POST /api/users/:tenantId/change-password
   * Change password for current user
   *
   * Body:
   *   - currentPassword (required): user's current password
   *   - newPassword (required): new password (min 6 chars)
   */
  async changePassword(req: AuthRequest, res: Response): Promise<void> {
    try {
      const { tenantId } = req.params;
      const { currentPassword, newPassword } = req.body;
      const userId = req.userId;

      if (!currentPassword || !newPassword) {
        res.status(400).json({
          success: false,
          error: "Current password and new password are required",
        });
        return;
      }

      if (newPassword.length < 6) {
        res.status(400).json({
          success: false,
          error: "New password must be at least 6 characters",
        });
        return;
      }

      logger.info(`🔐 Changing password for user: ${userId}`);

      await this.userService.changePassword(
        tenantId,
        userId!,
        currentPassword,
        newPassword
      );

      logger.info(`✅ Password changed for user: ${userId}`);
      res.json({
        success: true,
        message: "Password changed successfully",
      });
    } catch (error: any) {
      logger.error(`❌ Failed to change password: ${error.message}`);
      res.status(400).json({
        success: false,
        error: error.message,
      });
    }
  }
}

export function getUserManagementController(): UserManagementController {
  return new UserManagementController();
}
