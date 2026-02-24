import { describe, it, expect, beforeEach } from "vitest";
import { render, screen } from "@testing-library/react";
import { MemoryRouter, Route, Routes, useLocation } from "react-router-dom";
import { ProtectedRoute } from "./ProtectedRoute";
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

function setAuth(user: User | null) {
  useAuthStore.setState({
    user,
    isAuthenticated: Boolean(user),
    token: user ? "token" : null,
    accessToken: user ? "token" : null,
    tenantId: null,
    expiresAt: null,
  });
}

function LoginProbe() {
  const location = useLocation();
  return <div data-testid="login-search">{location.search}</div>;
}

describe("ProtectedRoute", () => {
  beforeEach(() => {
    setAuth(null);
  });

  it("redirects unauthenticated user to login with encoded returnUrl", () => {
    render(
      <MemoryRouter initialEntries={["/protected?tab=ship&filter=open"]}>
        <Routes>
          <Route path="/login" element={<LoginProbe />} />
          <Route
            path="/protected"
            element={
              <ProtectedRoute>
                <div>Secure Content</div>
              </ProtectedRoute>
            }
          />
        </Routes>
      </MemoryRouter>,
    );

    expect(screen.getByTestId("login-search")).toHaveTextContent(
      "?returnUrl=%2Fprotected%3Ftab%3Dship%26filter%3Dopen",
    );
  });

  it("renders children when user is authenticated and no guard is specified", () => {
    setAuth(normalUser);

    render(
      <MemoryRouter initialEntries={["/protected"]}>
        <Routes>
          <Route
            path="/protected"
            element={
              <ProtectedRoute>
                <div>Secure Content</div>
              </ProtectedRoute>
            }
          />
        </Routes>
      </MemoryRouter>,
    );

    expect(screen.getByText("Secure Content")).toBeInTheDocument();
  });

  it("shows Access Denied when permission requirement fails", () => {
    setAuth(normalUser);

    render(
      <MemoryRouter initialEntries={["/protected"]}>
        <Routes>
          <Route
            path="/protected"
            element={
              <ProtectedRoute permission="users.delete">
                <div>Secure Content</div>
              </ProtectedRoute>
            }
          />
        </Routes>
      </MemoryRouter>,
    );

    expect(screen.getByText("Access Denied")).toBeInTheDocument();
    expect(
      screen.getByText("You do not have permission to access this page."),
    ).toBeInTheDocument();
  });

  it("shows Access Denied when role requirement fails", () => {
    setAuth(adminUser);
    const roleOnlyProps = { role: "owner" };

    render(
      <MemoryRouter initialEntries={["/protected"]}>
        <Routes>
          <Route
            path="/protected"
            element={
              <ProtectedRoute {...roleOnlyProps}>
                <div>Secure Content</div>
              </ProtectedRoute>
            }
          />
        </Routes>
      </MemoryRouter>,
    );

    expect(screen.getByText("Access Denied")).toBeInTheDocument();
  });

  it("renders Outlet when children are not provided", () => {
    setAuth(adminUser);

    render(
      <MemoryRouter initialEntries={["/protected"]}>
        <Routes>
          <Route path="/protected" element={<ProtectedRoute />}>
            <Route index element={<div>Outlet Content</div>} />
          </Route>
        </Routes>
      </MemoryRouter>,
    );

    expect(screen.getByText("Outlet Content")).toBeInTheDocument();
  });
});
