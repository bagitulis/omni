import { describe, it, expect } from "vitest";
import {
  isValidRole,
  hasPermission,
  hasRole,
  canManageUsers,
  canCreateUser,
  canDeleteUser,
  canViewAudit,
  canSwitchTenant,
  canAccessAdmin,
  getAllPermissions,
  getRoleDisplayName,
  getRoleDescription,
  ROLE_PERMISSIONS,
} from "./permissionService";

describe("isValidRole", () => {
  it("returns true for developer", () => {
    expect(isValidRole("developer")).toBe(true);
  });

  it("returns true for owner", () => {
    expect(isValidRole("owner")).toBe(true);
  });

  it("returns true for admin", () => {
    expect(isValidRole("admin")).toBe(true);
  });

  it("returns true for user", () => {
    expect(isValidRole("user")).toBe(true);
  });

  it("returns false for unknown role", () => {
    expect(isValidRole("superadmin")).toBe(false);
  });

  it("returns false for empty string", () => {
    expect(isValidRole("")).toBe(false);
  });

  it("returns false for uppercase role", () => {
    expect(isValidRole("ADMIN")).toBe(false);
  });
});

describe("hasPermission", () => {
  it("returns false for invalid role", () => {
    expect(hasPermission("invalid", "users.list")).toBe(false);
  });

  it("developer has users.list", () => {
    expect(hasPermission("developer", "users.list")).toBe(true);
  });

  it("developer has tenant.switch", () => {
    expect(hasPermission("developer", "tenant.switch")).toBe(true);
  });

  it("owner does not have tenant.switch", () => {
    expect(hasPermission("owner", "tenant.switch")).toBe(false);
  });

  it("admin does not have users.create", () => {
    expect(hasPermission("admin", "users.create")).toBe(false);
  });

  it("user has users.changePassword", () => {
    expect(hasPermission("user", "users.changePassword")).toBe(true);
  });

  it("user does not have users.list", () => {
    expect(hasPermission("user", "users.list")).toBe(false);
  });

  it("returns false for non-existent permission", () => {
    expect(hasPermission("developer", "nonexistent.perm")).toBe(false);
  });
});

describe("hasRole", () => {
  it("returns true when roles match", () => {
    expect(hasRole("admin", "admin")).toBe(true);
  });

  it("returns false when roles differ", () => {
    expect(hasRole("admin", "owner")).toBe(false);
  });

  it("is case-sensitive", () => {
    expect(hasRole("Admin", "admin")).toBe(false);
  });
});

describe("canManageUsers", () => {
  it("developer can manage users", () => {
    expect(canManageUsers("developer")).toBe(true);
  });

  it("owner can manage users", () => {
    expect(canManageUsers("owner")).toBe(true);
  });

  it("admin can manage users", () => {
    expect(canManageUsers("admin")).toBe(true);
  });

  it("user cannot manage users", () => {
    expect(canManageUsers("user")).toBe(false);
  });

  it("invalid role cannot manage users", () => {
    expect(canManageUsers("ghost")).toBe(false);
  });
});

describe("canCreateUser", () => {
  it("developer can create users", () => {
    expect(canCreateUser("developer")).toBe(true);
  });

  it("owner can create users", () => {
    expect(canCreateUser("owner")).toBe(true);
  });

  it("admin cannot create users", () => {
    expect(canCreateUser("admin")).toBe(false);
  });

  it("user cannot create users", () => {
    expect(canCreateUser("user")).toBe(false);
  });
});

describe("canDeleteUser", () => {
  it("developer can delete users", () => {
    expect(canDeleteUser("developer")).toBe(true);
  });

  it("owner can delete users", () => {
    expect(canDeleteUser("owner")).toBe(true);
  });

  it("admin cannot delete users", () => {
    expect(canDeleteUser("admin")).toBe(false);
  });

  it("user cannot delete users", () => {
    expect(canDeleteUser("user")).toBe(false);
  });
});

describe("canViewAudit", () => {
  it("developer can view audit", () => {
    expect(canViewAudit("developer")).toBe(true);
  });

  it("owner can view audit", () => {
    expect(canViewAudit("owner")).toBe(true);
  });

  it("admin can view audit", () => {
    expect(canViewAudit("admin")).toBe(true);
  });

  it("user cannot view audit", () => {
    expect(canViewAudit("user")).toBe(false);
  });
});

describe("canSwitchTenant", () => {
  it("developer can switch tenants", () => {
    expect(canSwitchTenant("developer")).toBe(true);
  });

  it("owner cannot switch tenants", () => {
    expect(canSwitchTenant("owner")).toBe(false);
  });

  it("admin cannot switch tenants", () => {
    expect(canSwitchTenant("admin")).toBe(false);
  });

  it("user cannot switch tenants", () => {
    expect(canSwitchTenant("user")).toBe(false);
  });
});

describe("canAccessAdmin", () => {
  it("developer can access admin", () => {
    expect(canAccessAdmin("developer")).toBe(true);
  });

  it("owner can access admin", () => {
    expect(canAccessAdmin("owner")).toBe(true);
  });

  it("admin can access admin", () => {
    expect(canAccessAdmin("admin")).toBe(true);
  });

  it("user cannot access admin", () => {
    expect(canAccessAdmin("user")).toBe(false);
  });

  it("undefined role cannot access admin", () => {
    expect(canAccessAdmin(undefined)).toBe(false);
  });

  it("empty string cannot access admin", () => {
    expect(canAccessAdmin("")).toBe(false);
  });
});

describe("getAllPermissions", () => {
  it("returns array of permissions for developer", () => {
    const perms = getAllPermissions("developer");
    expect(Array.isArray(perms)).toBe(true);
    expect(perms.length).toBeGreaterThan(0);
    expect(perms).toContain("tenant.switch");
    expect(perms).toContain("users.list");
  });

  it("returns array for admin", () => {
    const perms = getAllPermissions("admin");
    expect(perms).toContain("users.list");
    expect(perms).not.toContain("tenant.switch");
  });

  it("returns empty array for invalid role", () => {
    expect(getAllPermissions("unknown")).toEqual([]);
  });

  it("user permissions include users.update.self", () => {
    const perms = getAllPermissions("user");
    expect(perms).toContain("users.update.self");
  });
});

describe("getRoleDisplayName", () => {
  it("returns Developer for developer", () => {
    expect(getRoleDisplayName("developer")).toBe("Developer");
  });

  it("returns Owner for owner", () => {
    expect(getRoleDisplayName("owner")).toBe("Owner");
  });

  it("returns Administrator for admin", () => {
    expect(getRoleDisplayName("admin")).toBe("Administrator");
  });

  it("returns User for user", () => {
    expect(getRoleDisplayName("user")).toBe("User");
  });

  it("returns raw string for unknown role", () => {
    expect(getRoleDisplayName("mystery")).toBe("mystery");
  });
});

describe("getRoleDescription", () => {
  it("returns non-empty string for developer", () => {
    expect(getRoleDescription("developer")).toBeTruthy();
    expect(getRoleDescription("developer")).toContain("tenant");
  });

  it("returns non-empty string for owner", () => {
    expect(getRoleDescription("owner")).toBeTruthy();
  });

  it("returns non-empty string for admin", () => {
    expect(getRoleDescription("admin")).toBeTruthy();
  });

  it("returns empty string for unknown role", () => {
    expect(getRoleDescription("unknown")).toBe("");
  });
});

describe("ROLE_PERMISSIONS constant", () => {
  it("has entries for all four roles", () => {
    expect(ROLE_PERMISSIONS.developer).toBeDefined();
    expect(ROLE_PERMISSIONS.owner).toBeDefined();
    expect(ROLE_PERMISSIONS.admin).toBeDefined();
    expect(ROLE_PERMISSIONS.user).toBeDefined();
  });

  it("developer has more permissions than user", () => {
    expect(ROLE_PERMISSIONS.developer.size).toBeGreaterThan(
      ROLE_PERMISSIONS.user.size,
    );
  });
});
