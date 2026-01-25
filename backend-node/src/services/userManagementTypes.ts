/**
 * User Management Types
 * SRP: Type definitions for user management
 */

export interface CreateUserRequest {
  username: string;
  email: string;
  password: string;
  role?: string;
  shopName?: string;
}

export interface UpdateUserRequest {
  email?: string;
  password?: string;
  role?: string;
}

export interface UserResponse {
  id: string;
  username: string;
  email: string;
  role: string;
  createdAt: Date;
  updatedAt: Date;
}

export interface UserWithTenant extends UserResponse {
  tenantId: string;
}

/**
 * Role visibility rules:
 * - developer: can see/manage all roles
 * - owner: can see/manage admin, user
 * - admin: can see/manage user only
 * - user: can see/manage self only
 */
export function getVisibleRoles(role: string): string[] {
  switch (role) {
    case "developer":
      return ["developer", "owner", "admin", "user"];
    case "owner":
      return ["admin", "user"];
    case "admin":
      return ["user"];
    default:
      return [];
  }
}
