import { Router } from "express";
import { AuthRequest} from "../middleware/tenantMiddleware";
import { authMiddleware } from "../middleware/authMiddleware";
// import { roleMiddleware } from "../middleware/roleMiddleware";
import {
  permissionMiddleware,
  canManageRoleMiddleware,
  canUpdateUserMiddleware,
  canDeleteUserMiddleware,
} from "../middleware/permissionMiddleware";
import { getUserManagementController } from "../controllers/userManagementController";
import { getLogger } from "../utils/logger";

getLogger("UserManagementRoutes"); // Initialize logger for module
const controller = getUserManagementController();

/**
 * User Management Routes
 * Single Responsibility: User CRUD endpoints with role-based access control
 *
 * All endpoints require:
 *   - Authentication (Authorization header with token)
 *   - Permission checks (enforced via middleware)
 *
 * Endpoints:
 *   GET /api/admin/users - List all users (users.list permission)
 *   POST /api/admin/users - Create new user (users.create permission)
 *   PATCH /api/admin/users/:tenantId - Update user (users.update permission)
 *   DELETE /api/admin/users/:tenantId - Delete user (permission-based)
 */
export const userManagementRouter = Router();

// Authentication middleware - verify JWT token first
userManagementRouter.use(authMiddleware);

// Extract role from token and verify it
userManagementRouter.use((req: any, _res, next) => {
  if (req.role) {
    (req as AuthRequest).userRole = req.role;
  }
  next();
});

/**
 * GET /api/admin/users
 * List all users (requires users.list permission)
 */
userManagementRouter.get(
  "/",
  permissionMiddleware("users.list"),
  async (req: AuthRequest, res) => {
    await controller.listAllUsers(req, res);
  }
);

/**
 * POST /api/admin/users
 * Create new user (requires users.create permission)
 *
 * Body:
 * {
 *   "username": "newuser",
 *   "email": "user@example.com",
 *   "password": "securepassword",
 *   "shopName": "My Shop" (optional)
 * }
 */
userManagementRouter.post(
  "/",
  permissionMiddleware("users.create"),
  canManageRoleMiddleware,
  async (req: AuthRequest, res) => {
    await controller.createUser(req, res);
  }
);

/**
 * PATCH /api/admin/users/:tenantId
 * Update user (requires users.update permission)
 *
 * Body (any combination):
 * {
 *   "email": "newemail@example.com",
 *   "password": "newpassword",
 *   "role": "owner|admin|user"
 * }
 */
userManagementRouter.patch(
  "/:tenantId",
  permissionMiddleware("users.update"),
  canUpdateUserMiddleware,
  canManageRoleMiddleware,
  async (req: AuthRequest, res) => {
    await controller.updateUser(req, res);
  }
);

/**
 * DELETE /api/admin/users/:tenantId
 * Delete user (permission-based deletion)
 *
 * Cleanup includes:
 *   - User database records
 *   - Prisma database file
 *   - Job database file
 *   - Configuration from tenants.json
 */
userManagementRouter.delete(
  "/:tenantId",
  permissionMiddleware("users.delete"),
  canDeleteUserMiddleware,
  async (req: AuthRequest, res) => {
    await controller.deleteUser(req, res);
  }
);

/**
 * POST /api/users/:tenantId/change-password
 * Change password for current user (authenticated users only)
 *
 * Body:
 * {
 *   "currentPassword": "oldpassword",
 *   "newPassword": "newpassword123"
 * }
 */
userManagementRouter.post(
  "/:tenantId/change-password",
  async (req: AuthRequest, res) => {
    await controller.changePassword(req, res);
  }
);

export default userManagementRouter;
