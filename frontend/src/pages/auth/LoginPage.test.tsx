import { render, screen, fireEvent, waitFor } from "@testing-library/react";
import "@testing-library/jest-dom/vitest";
import { describe, it, expect, vi, beforeEach } from "vitest";
import LoginPage from "./LoginPage";

// ---------------------------------------------------------------------------
// Browser API stubs required by Ant Design
// ---------------------------------------------------------------------------
Object.defineProperty(window, "matchMedia", {
  writable: true,
  value: vi.fn().mockImplementation((q: string) => ({
    matches: false,
    media: q,
    onchange: null,
    addListener: vi.fn(),
    removeListener: vi.fn(),
    addEventListener: vi.fn(),
    removeEventListener: vi.fn(),
    dispatchEvent: vi.fn(),
  })),
});

// ---------------------------------------------------------------------------
// Mocks
// ---------------------------------------------------------------------------
const mockNavigate = vi.fn();
const mockSearchParams = new URLSearchParams();

vi.mock("react-router-dom", async () => {
  const actual = await vi.importActual("react-router-dom");
  return {
    ...actual,
    useNavigate: () => mockNavigate,
    useSearchParams: () => [mockSearchParams],
  };
});

const mockLogin = vi.fn();
const mockDevLogin = vi.fn();

vi.mock("@/api/auth", () => ({
  login: (payload: unknown) => mockLogin(payload),
  devLogin: (payload: unknown) => mockDevLogin(payload),
}));

const mockSetAuth = vi.fn();
const mockIsAuthenticated = false;

vi.mock("@/stores/authStore", () => ({
  useAuthStore: () => ({
    setAuth: mockSetAuth,
    isAuthenticated: mockIsAuthenticated,
  }),
}));

vi.mock("@/lib/logger", () => ({
  logger: {
    warn: vi.fn(),
  },
}));

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------
describe("LoginPage", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockSearchParams.delete("returnUrl");
    Object.defineProperty(window, "location", {
      writable: true,
      value: { hostname: "example.com" },
    });
    sessionStorage.clear();
  });
  it("renders login form", () => {
    render(<LoginPage />);
    expect(screen.getByPlaceholderText("Username")).toBeInTheDocument();
    expect(screen.getByPlaceholderText("Password")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: /login/i })).toBeInTheDocument();
  });

  it("renders dev mode login when on localhost", () => {
    // Mock window.location.hostname
    Object.defineProperty(window, "location", {
      writable: true,
      value: { hostname: "localhost" },
    });
    sessionStorage.setItem("autoLoginFailed", "true");

    render(<LoginPage />);
    expect(screen.getByText("Dev Mode (Localhost)")).toBeInTheDocument();
    expect(screen.getByText("Quick Dev Login")).toBeInTheDocument();


  });

  it("handles standard login", async () => {
    mockLogin.mockResolvedValueOnce({
      token: "fake-token",
      access_token: "fake-access-token",
      user: { id: 1, name: "Test User" },
      tenant_id: "test-tenant",
      expires_in: 3600,
    });

    render(<LoginPage />);

    fireEvent.change(screen.getByPlaceholderText("Username"), {
      target: { value: "testuser" },
    });
    fireEvent.change(screen.getByPlaceholderText("Password"), {
      target: { value: "password123" },
    });
    fireEvent.click(screen.getByRole("button", { name: "Login" }));

    await waitFor(() => {
      expect(mockLogin).toHaveBeenCalledWith({
        username: "testuser",
        password: "password123",
      });
      expect(mockSetAuth).toHaveBeenCalled();
      expect(mockNavigate).toHaveBeenCalledWith("/");
    });
  });

  it("handles login failure", async () => {
    mockLogin.mockRejectedValueOnce(new Error("Invalid credentials"));

    render(<LoginPage />);

    fireEvent.change(screen.getByPlaceholderText("Username"), {
      target: { value: "wrong" },
    });
    fireEvent.change(screen.getByPlaceholderText("Password"), {
      target: { value: "wrong" },
    });
    fireEvent.click(screen.getByRole("button", { name: "Login" }));

    await waitFor(() => {
      expect(screen.getByText("Invalid credentials")).toBeInTheDocument();
    });
  });
});
