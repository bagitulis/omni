import { Response, NextFunction } from "express";
import { AuthRequest } from "./tenantMiddleware";
import PermissionService, { UserRole } from "../services/permissionService";
import { getLogger } from "../utils/logger";

const logger = getLogger("PermissionMiddleware");

/**
 * Middleware to check if user has specific permission
 * Usage: permissionMiddleware("users.create")
 */
export function permissionMiddleware(requiredPermission: string) {
  return (req: AuthRequest, res: Response, next: NextFunction) => {
    const userRole = (req.userRole || "user") as UserRole;

    // Check if user has permission
    if (!PermissionService.hasPermission(userRole, requiredPermission)) {
      logger.warn(
        `Unauthorized access: ${req.username} (${userRole}) tried to access ${requiredPermission}`
      );
      return res.status(403).json({
        success: false,
        error: `Insufficient permissions. Required: ${requiredPermission}`,
      });
    }

    return next();
  };
}

/**
 * Middleware to check if user can manage specific user role
 * Usage: canManageRoleMiddleware
 * Checks: req.body.role or req.params.targetRole
 */
export function canManageRoleMiddleware(
  req: AuthRequest,
  res: Response,
  next: NextFunction
) {
  const userRole = (req.userRole || "user") as UserRole;
  const targetRole = req.body?.role || req.params?.targetRole || "user";

  if (!PermissionService.canManageUser(userRole, targetRole)) {
    logger.warn(
      `Cannot manage role: ${req.username} (${userRole}) tried to manage role ${targetRole}`
    );
    return res.status(403).json({
      success: false,
      error: `Cannot manage users with role: ${targetRole}`,
    });
  }

  return next();
}

/**
 * Middleware to check if user can update specific user
 * Usage: canUpdateUserMiddleware
 * Requires: req.params.tenantId and req.userId (from token)
 */
export function canUpdateUserMiddleware(
  req: AuthRequest,
  res: Response,
  next: NextFunction
) {
  const userRole = (req.userRole || "user") as UserRole;
  const userId = req.userId || "";
  const targetUserId = req.params.tenantId; // tenantId = userId in some cases

  if (!PermissionService.canUpdateUser(userRole, targetUserId, userId)) {
    logger.warn(
      `Cannot update user: ${req.username} (${userRole}) tried to update ${targetUserId}`
    );
    return res.status(403).json({
      success: false,
      error: "You can only update your own profile",
    });
  }

  return next();
}

/**
 * Middleware to check if user can delete specific user
 * Usage: canDeleteUserMiddleware
 * Requires: req.body.targetRole (the role of user being deleted)
 */
export function canDeleteUserMiddleware(
  req: AuthRequest,
  res: Response,
  next: NextFunction
) {
  const userRole = (req.userRole || "user") as UserRole;
  const targetRole = req.body?.targetRole || "user";

  if (!PermissionService.canDeleteUser(userRole, targetRole)) {
    logger.warn(
      `Cannot delete user: ${req.username} (${userRole}) tried to delete user with role ${targetRole}`
    );
    return res.status(403).json({
      success: false,
      error: `Cannot delete users with role: ${targetRole}`,
    });
  }

  return next();
}
