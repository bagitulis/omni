import { describe, it, expect, beforeEach } from "vitest";
import { render, screen } from "@testing-library/react";
import { PermissionGate } from "./PermissionGate";
import { useAuthStore } from "@/stores/authStore";
import type { User } from "@/types/auth";

const adminUser: User = {
  id: "u-admin",
  username: "admin",
  email: "admin@example.com",
  role: "admin",
};

const normalUser: User = {
  id: "u-user",
  username: "user",
  email: "user@example.com",
  role: "user",
};

function setUser(user: User | null) {
  useAuthStore.setState({
    user,
    isAuthenticated: Boolean(user),
    token: user ? "token" : null,
    accessToken: user ? "token" : null,
    tenantId: null,
    expiresAt: null,
  });
}

describe("PermissionGate", () => {
  beforeEach(() => {
    setUser(null);
  });

  it("renders children when no permission and no role are provided", () => {
    render(
      <PermissionGate>
        <span>Visible Content</span>
      </PermissionGate>,
    );

    expect(screen.getByText("Visible Content")).toBeInTheDocument();
  });

  it("renders children when permission check passes", () => {
    setUser(adminUser);

    render(
      <PermissionGate permission="users.update">
        <span>Can Update</span>
      </PermissionGate>,
    );

    expect(screen.getByText("Can Update")).toBeInTheDocument();
  });

  it("renders fallback when permission check fails", () => {
    setUser(normalUser);

    render(
      <PermissionGate
        permission="users.delete"
        fallback={<span>Forbidden</span>}
      >
        <span>Hidden Content</span>
      </PermissionGate>,
    );

    expect(screen.getByText("Forbidden")).toBeInTheDocument();
    expect(screen.queryByText("Hidden Content")).not.toBeInTheDocument();
  });

  it("renders fallback when role check fails", () => {
    setUser(adminUser);
    const roleOnlyProps = { role: "owner" };

    render(
      <PermissionGate {...roleOnlyProps} fallback={<span>Owner Only</span>}>
        <span>Hidden Content</span>
      </PermissionGate>,
    );

    expect(screen.getByText("Owner Only")).toBeInTheDocument();
    expect(screen.queryByText("Hidden Content")).not.toBeInTheDocument();
  });

  it("requires both permission and role when both are provided", () => {
    setUser(adminUser);
    const roleAndPermissionProps = {
      permission: "users.update",
      role: "owner",
    };

    render(
      <PermissionGate
        {...roleAndPermissionProps}
        fallback={<span>Denied</span>}
      >
        <span>Hidden Content</span>
      </PermissionGate>,
    );

    expect(screen.getByText("Denied")).toBeInTheDocument();
    expect(screen.queryByText("Hidden Content")).not.toBeInTheDocument();
  });
});
