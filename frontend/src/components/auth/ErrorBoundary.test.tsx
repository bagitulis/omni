import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { render, screen } from "@testing-library/react";
import { ErrorBoundary } from "../ErrorBoundary";
import { logger } from "@/lib/logger";

vi.mock("@/lib/logger", () => ({
  logger: {
    error: vi.fn(),
  },
}));

function ThrowOnRender({ shouldThrow = true }: { shouldThrow?: boolean }) {
  if (shouldThrow) {
    throw new Error("Boundary test error");
  }

  return <div>No Error</div>;
}

describe("ErrorBoundary", () => {
  const loggerErrorMock = vi.mocked(logger.error);
  let consoleErrorSpy: ReturnType<typeof vi.spyOn>;

  beforeEach(() => {
    loggerErrorMock.mockReset();
    consoleErrorSpy = vi
      .spyOn(console, "error")
      .mockImplementation(() => undefined);
  });

  afterEach(() => {
    consoleErrorSpy.mockRestore();
  });

  it("renders children when there is no error", () => {
    render(
      <ErrorBoundary>
        <div>Healthy Tree</div>
      </ErrorBoundary>,
    );

    expect(screen.getByText("Healthy Tree")).toBeInTheDocument();
  });

  it("renders fallback UI when child throws", () => {
    render(
      <ErrorBoundary>
        <ThrowOnRender />
      </ErrorBoundary>,
    );

    expect(screen.getByText("Something went wrong")).toBeInTheDocument();

    if (import.meta.env.DEV) {
      expect(loggerErrorMock).toHaveBeenCalledWith(
        "ErrorBoundary caught an error",
        expect.objectContaining({
          error: "Boundary test error",
          componentStack: expect.any(String),
        }),
      );
    }
  });

  it("renders reload action button when fallback UI is shown", () => {
    render(
      <ErrorBoundary>
        <ThrowOnRender />
      </ErrorBoundary>,
    );

    expect(
      screen.getByRole("button", { name: "Reload Page" }),
    ).toBeInTheDocument();
  });
});
