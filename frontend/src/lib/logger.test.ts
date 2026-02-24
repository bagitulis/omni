import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";

// logger reads import.meta.env.DEV at construction time (class field in constructor).
// We use vi.stubEnv + vi.resetModules + dynamic import to control the DEV flag per test.

describe("logger — DEV mode (isDev=true)", () => {
  beforeEach(() => {
    vi.resetModules();
    vi.stubEnv("DEV", true);
    vi.spyOn(console, "debug").mockImplementation(() => {});
    vi.spyOn(console, "info").mockImplementation(() => {});
    vi.spyOn(console, "warn").mockImplementation(() => {});
    vi.spyOn(console, "error").mockImplementation(() => {});
  });

  afterEach(() => {
    vi.unstubAllEnvs();
    vi.restoreAllMocks();
  });

  it("debug calls console.debug with [DEBUG] prefix", async () => {
    const { logger } = await import("./logger");
    logger.debug("test debug message");
    expect(console.debug).toHaveBeenCalledTimes(1);
    const call = (console.debug as ReturnType<typeof vi.fn>).mock.calls[0];
    expect(call[0]).toContain("[DEBUG]");
    expect(call[0]).toContain("test debug message");
  });

  it("debug passes context as second arg when provided", async () => {
    const { logger } = await import("./logger");
    const ctx = { key: "value" };
    logger.debug("msg", ctx);
    const call = (console.debug as ReturnType<typeof vi.fn>).mock.calls[0];
    expect(call[1]).toEqual(ctx);
  });

  it("debug passes empty string as second arg when no context", async () => {
    const { logger } = await import("./logger");
    logger.debug("msg");
    const call = (console.debug as ReturnType<typeof vi.fn>).mock.calls[0];
    expect(call[1]).toBe("");
  });

  it("info calls console.info with [INFO] prefix", async () => {
    const { logger } = await import("./logger");
    logger.info("test info message");
    expect(console.info).toHaveBeenCalledTimes(1);
    const call = (console.info as ReturnType<typeof vi.fn>).mock.calls[0];
    expect(call[0]).toContain("[INFO]");
    expect(call[0]).toContain("test info message");
  });

  it("info passes context as second arg when provided", async () => {
    const { logger } = await import("./logger");
    const ctx = { requestId: "abc" };
    logger.info("msg", ctx);
    const call = (console.info as ReturnType<typeof vi.fn>).mock.calls[0];
    expect(call[1]).toEqual(ctx);
  });

  it("warn calls console.warn with [WARN] prefix in DEV", async () => {
    const { logger } = await import("./logger");
    logger.warn("test warn message");
    expect(console.warn).toHaveBeenCalledTimes(1);
    const call = (console.warn as ReturnType<typeof vi.fn>).mock.calls[0];
    expect(call[0]).toContain("[WARN]");
    expect(call[0]).toContain("test warn message");
  });

  it("error calls console.error with [ERROR] prefix in DEV", async () => {
    const { logger } = await import("./logger");
    logger.error("test error message");
    expect(console.error).toHaveBeenCalledTimes(1);
    const call = (console.error as ReturnType<typeof vi.fn>).mock.calls[0];
    expect(call[0]).toContain("[ERROR]");
    expect(call[0]).toContain("test error message");
  });

  it("error passes context as second arg when provided", async () => {
    const { logger } = await import("./logger");
    const ctx = { errorCode: 500 };
    logger.error("msg", ctx);
    const call = (console.error as ReturnType<typeof vi.fn>).mock.calls[0];
    expect(call[1]).toEqual(ctx);
  });

  it("log message includes ISO timestamp", async () => {
    const { logger } = await import("./logger");
    logger.debug("timestamp test");
    const call = (console.debug as ReturnType<typeof vi.fn>).mock.calls[0];
    // ISO 8601 pattern: e.g. 2026-02-24T10:30:00.000Z
    expect(call[0]).toMatch(/\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}/);
  });
});

describe("logger — production mode (isDev=false)", () => {
  beforeEach(() => {
    vi.resetModules();
    vi.stubEnv("DEV", false);
    vi.spyOn(console, "debug").mockImplementation(() => {});
    vi.spyOn(console, "info").mockImplementation(() => {});
    vi.spyOn(console, "warn").mockImplementation(() => {});
    vi.spyOn(console, "error").mockImplementation(() => {});
  });

  afterEach(() => {
    vi.unstubAllEnvs();
    vi.restoreAllMocks();
  });

  it("debug does NOT call console.debug in production", async () => {
    const { logger } = await import("./logger");
    logger.debug("should be suppressed");
    expect(console.debug).not.toHaveBeenCalled();
  });

  it("info does NOT call console.info in production", async () => {
    const { logger } = await import("./logger");
    logger.info("should be suppressed");
    expect(console.info).not.toHaveBeenCalled();
  });

  it("warn DOES call console.warn in production", async () => {
    const { logger } = await import("./logger");
    logger.warn("important warning");
    expect(console.warn).toHaveBeenCalledTimes(1);
    const call = (console.warn as ReturnType<typeof vi.fn>).mock.calls[0];
    expect(call[0]).toContain("[WARN]");
    expect(call[0]).toContain("important warning");
  });

  it("error DOES call console.error in production", async () => {
    const { logger } = await import("./logger");
    logger.error("critical error");
    expect(console.error).toHaveBeenCalledTimes(1);
    const call = (console.error as ReturnType<typeof vi.fn>).mock.calls[0];
    expect(call[0]).toContain("[ERROR]");
    expect(call[0]).toContain("critical error");
  });

  it("warn passes context in production", async () => {
    const { logger } = await import("./logger");
    const ctx = { userId: "u1" };
    logger.warn("msg", ctx);
    const call = (console.warn as ReturnType<typeof vi.fn>).mock.calls[0];
    expect(call[1]).toEqual(ctx);
  });

  it("error passes empty string when no context in production", async () => {
    const { logger } = await import("./logger");
    logger.error("msg");
    const call = (console.error as ReturnType<typeof vi.fn>).mock.calls[0];
    expect(call[1]).toBe("");
  });
});
