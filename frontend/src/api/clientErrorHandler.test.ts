import { describe, it, expect, vi, beforeEach } from "vitest";
import type { AxiosError } from "axios";

const { mockClearAuth, mockMessageError, mockLoggerError, mockLoggerInfo } =
  vi.hoisted(() => ({
    mockClearAuth: vi.fn(),
    mockMessageError: vi.fn(),
    mockLoggerError: vi.fn(),
    mockLoggerInfo: vi.fn(),
  }));

vi.mock("@/stores/authStore", () => ({
  useAuthStore: {
    getState: () => ({ clearAuth: mockClearAuth }),
  },
}));

vi.mock("antd", () => ({
  message: { error: mockMessageError },
}));

vi.mock("@/lib/logger", () => ({
  logger: { error: mockLoggerError, info: mockLoggerInfo },
}));

// Setup window mocks
Object.defineProperty(window, "location", {
  value: {
    pathname: "/dashboard",
    hostname: "localhost",
    origin: "http://localhost:5173",
    href: "",
  },
  writable: true,
});

import { handleAuthExpired, handleResponseError } from "./clientErrorHandler";

function makeAxiosError(overrides: {
  code?: string;
  response?: {
    status: number;
    data?: unknown;
  };
  config?: { url?: string };
  message?: string;
}): AxiosError {
  return {
    isAxiosError: true,
    name: "AxiosError",
    message: overrides.message ?? "Network Error",
    code: overrides.code,
    response: overrides.response as AxiosError["response"],
    config: overrides.config as AxiosError["config"],
    toJSON: () => ({}),
  } as AxiosError;
}

describe("handleAuthExpired", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    window.location.pathname = "/dashboard";
    window.location.href = "";
  });

  it("calls clearAuth", () => {
    handleAuthExpired();
    expect(mockClearAuth).toHaveBeenCalledOnce();
  });

  it("redirects to /login with returnUrl when not on root", () => {
    window.location.pathname = "/orders";
    handleAuthExpired();
    expect(window.location.href).toBe(
      `/login?returnUrl=${encodeURIComponent("/orders")}`,
    );
  });

  it("redirects to /login without returnUrl when on root path", () => {
    window.location.pathname = "/";
    handleAuthExpired();
    expect(window.location.href).toBe("/login");
  });
});

describe("handleResponseError - timeout", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    Object.assign(window.location, {
      pathname: "/dashboard",
      hostname: "localhost",
      origin: "http://localhost:5173",
    });
  });
  it("rejects with timeout error for ECONNABORTED", async () => {
    const error = makeAxiosError({ code: "ECONNABORTED" });
    await expect(handleResponseError(error)).rejects.toThrow(
      "Request timeout - server is taking too long to respond",
    );
    expect(mockLoggerError).toHaveBeenCalled();
  });
});
describe("handleResponseError - network error", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    Object.assign(window.location, {
      hostname: "localhost",
      origin: "http://localhost:5173",
    });
  });
  it("rejects with network error when no response and host is localhost", async () => {
    const error = makeAxiosError({});
    await expect(handleResponseError(error)).rejects.toThrow(
      "Network error - cannot connect to http://localhost:3000",
    );
    expect(mockLoggerError).toHaveBeenCalled();
  });
  it("uses origin directly for non-localhost host", async () => {
    Object.assign(window.location, {
      hostname: "myapp.com",
      origin: "https://myapp.com",
    });
    const error = makeAxiosError({});
    await expect(handleResponseError(error)).rejects.toThrow(
      "Network error - cannot connect to https://myapp.com",
    );
  });
});

describe("handleResponseError - 401", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    window.location.pathname = "/dashboard";
    window.location.hostname = "localhost";
    window.location.href = "";
  });

  it("calls handleAuthExpired and rejects with session expired message", async () => {
    const error = makeAxiosError({
      response: { status: 401, data: {} },
      config: { url: "/api/orders" },
    });
    await expect(handleResponseError(error)).rejects.toThrow(
      "Session expired - please login again",
    );
    expect(mockClearAuth).toHaveBeenCalled();
  });

  it("does NOT redirect when already on /login", async () => {
    window.location.pathname = "/login";
    const error = makeAxiosError({
      response: { status: 401, data: {} },
      config: { url: "/api/orders" },
    });
    // Should not throw session expired — falls through to generic
    const result = handleResponseError(error);
    await expect(result).rejects.toBeDefined();
    expect(mockClearAuth).not.toHaveBeenCalled();
  });

  it("does NOT redirect for /auth/ endpoints", async () => {
    window.location.pathname = "/dashboard";
    const error = makeAxiosError({
      response: { status: 401, data: {} },
      config: { url: "/auth/login" },
    });
    await expect(handleResponseError(error)).rejects.toBeDefined();
    expect(mockClearAuth).not.toHaveBeenCalled();
  });
});

describe("handleResponseError - 403", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    window.location.hostname = "localhost";
  });

  it("rejects with permission denied error", async () => {
    const error = makeAxiosError({ response: { status: 403, data: {} } });
    await expect(handleResponseError(error)).rejects.toThrow(
      "Permission denied",
    );
    expect(mockMessageError).toHaveBeenCalledWith("Permission denied");
    expect(mockLoggerError).toHaveBeenCalled();
  });
});

describe("handleResponseError - 500", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    window.location.hostname = "localhost";
  });

  it("rejects with server error and shows toast", async () => {
    const error = makeAxiosError({
      response: { status: 500, data: { error: "Internal DB failure" } },
    });
    await expect(handleResponseError(error)).rejects.toThrow(
      "Server error - Internal DB failure",
    );
    expect(mockMessageError).toHaveBeenCalledWith(
      "Server error - please try again",
    );
  });

  it("uses fallback 'Server error' when backendMsg is absent", async () => {
    const error = makeAxiosError({ response: { status: 500, data: {} } });
    await expect(handleResponseError(error)).rejects.toThrow(
      "Server error - Server error",
    );
  });
});

describe("handleResponseError - generic errors", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    window.location.hostname = "localhost";
  });

  it("rejects with original error for unhandled status codes", async () => {
    const error = makeAxiosError({
      response: { status: 422, data: { error: "Validation failed" } },
    });
    const promise = handleResponseError(error);
    // Should reject with the original AxiosError (not a new Error)
    await expect(promise).rejects.toMatchObject({ isAxiosError: true });
    expect(mockLoggerError).toHaveBeenCalled();
  });

  it("uses error.message as fallback when backendMsg is absent", async () => {
    const error = makeAxiosError({
      response: { status: 404, data: {} },
      message: "Request failed with status 404",
    });
    await expect(handleResponseError(error)).rejects.toBeDefined();
    expect(mockLoggerError).toHaveBeenCalled();
  });
});
