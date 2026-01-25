/**
 * User Password Service
 * SRP: Handle password-related operations
 */

import bcrypt from "bcryptjs";
import { getDbManager } from "./dbConnectionManager";
import { getLogger } from "../utils/logger";
import AuditService from "./auditService";

const logger = getLogger("UserPasswordService");

export class UserPasswordService {
  /**
   * Change password for user
   */
  async changePassword(
    tenantId: string,
    userId: string,
    currentPassword: string,
    newPassword: string
  ): Promise<void> {
    try {
      const prisma = getDbManager().getConnection(tenantId);

      const user = await prisma.user.findUnique({ where: { id: userId } });
      if (!user) {
        throw new Error("User not found");
      }

      // Verify current password
      const isValid = await bcrypt.compare(currentPassword, user.password);
      if (!isValid) {
        throw new Error("Current password is incorrect");
      }

      // Hash and update new password
      const hashedPassword = await bcrypt.hash(newPassword, 10);
      await prisma.user.update({
        where: { id: userId },
        data: { password: hashedPassword, updatedAt: new Date() },
      });

      await AuditService.logAction({
        action: "PASSWORD_CHANGED",
        userId: userId,
        targetTenantId: tenantId,
        status: "success",
      });

      logger.info(`Password changed for user ${userId} in tenant ${tenantId}`);
    } catch (error: any) {
      logger.error(`Failed to change password: ${error.message}`);
      await AuditService.logAction({
        action: "PASSWORD_CHANGED",
        userId: userId,
        targetTenantId: tenantId,
        status: "failed",
        errorMessage: error.message,
      });
      throw error;
    }
  }

  /**
   * Reset password for user (admin action)
   */
  async resetPassword(
    tenantId: string,
    userId: string,
    newPassword: string
  ): Promise<void> {
    try {
      if (newPassword.length < 6) {
        throw new Error("Password must be at least 6 characters");
      }

      const prisma = getDbManager().getConnection(tenantId);
      const hashedPassword = await bcrypt.hash(newPassword, 10);

      await prisma.user.update({
        where: { id: userId },
        data: { password: hashedPassword, updatedAt: new Date() },
      });

      await AuditService.logAction({
        action: "PASSWORD_RESET",
        userId: "ADMIN",
        targetUserId: userId,
        targetTenantId: tenantId,
        status: "success",
      });

      logger.info(`Password reset for user ${userId} in tenant ${tenantId}`);
    } catch (error: any) {
      logger.error(`Failed to reset password: ${error.message}`);
      throw error;
    }
  }
}
