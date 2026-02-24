import { render, screen, fireEvent } from "@testing-library/react";
import { describe, expect, it, vi, beforeEach } from "vitest";
import { BulkPrintResultModal } from "./BulkPrintResultModal";
import type { BulkResult } from "../hooks/bulkActionTypes";

// Ant Design needs matchMedia in jsdom
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

function makeResult(overrides: Partial<BulkResult> = {}): BulkResult {
  return {
    succeeded: [],
    failed: [],
    ...overrides,
  };
}

describe("BulkPrintResultModal", () => {
  const onClose = vi.fn();
  const onRetryFailed = vi.fn();

  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("renders nothing when open is false", () => {
    render(
      <BulkPrintResultModal
        open={false}
        result={makeResult()}
        onClose={onClose}
        onRetryFailed={onRetryFailed}
        isRetrying={false}
      />,
    );
    expect(
      screen.queryByText("Bulk Print Result Details"),
    ).not.toBeInTheDocument();
  });

  it("renders modal title when open", () => {
    render(
      <BulkPrintResultModal
        open={true}
        result={makeResult()}
        onClose={onClose}
        onRetryFailed={onRetryFailed}
        isRetrying={false}
      />,
    );
    expect(screen.getByText("Bulk Print Result Details")).toBeInTheDocument();
  });

  it("shows success alert when there are no failures", () => {
    render(
      <BulkPrintResultModal
        open={true}
        result={makeResult({ succeeded: ["ORD-001", "ORD-002"] })}
        onClose={onClose}
        onRetryFailed={onRetryFailed}
        isRetrying={false}
      />,
    );
    expect(
      screen.getByText("All 2 labels downloaded successfully"),
    ).toBeInTheDocument();
  });

  it("shows warning alert with counts when there are failures", () => {
    render(
      <BulkPrintResultModal
        open={true}
        result={makeResult({
          succeeded: ["ORD-001"],
          failed: [{ order_sn: "ORD-002", error: "Print error" }],
        })}
        onClose={onClose}
        onRetryFailed={onRetryFailed}
        isRetrying={false}
      />,
    );
    expect(
      screen.getByText("1 labels downloaded, 1 failed"),
    ).toBeInTheDocument();
  });

  it("renders succeeded order SNs in Downloaded Labels section", () => {
    render(
      <BulkPrintResultModal
        open={true}
        result={makeResult({ succeeded: ["ORD-001", "ORD-002"] })}
        onClose={onClose}
        onRetryFailed={onRetryFailed}
        isRetrying={false}
      />,
    );
    expect(screen.getByText("ORD-001")).toBeInTheDocument();
    expect(screen.getByText("ORD-002")).toBeInTheDocument();
  });

  it("renders failed order SNs and errors in Failed Labels section", () => {
    render(
      <BulkPrintResultModal
        open={true}
        result={makeResult({
          failed: [{ order_sn: "ORD-FAIL", error: "Platform error: timeout" }],
        })}
        onClose={onClose}
        onRetryFailed={onRetryFailed}
        isRetrying={false}
      />,
    );
    expect(screen.getByText("ORD-FAIL")).toBeInTheDocument();
    expect(screen.getByText("Platform error: timeout")).toBeInTheDocument();
  });

  it("calls onClose when Close button is clicked", () => {
    render(
      <BulkPrintResultModal
        open={true}
        result={makeResult()}
        onClose={onClose}
        onRetryFailed={onRetryFailed}
        isRetrying={false}
      />,
    );
    fireEvent.click(screen.getByRole("button", { name: /close/i }));
    expect(onClose).toHaveBeenCalled();
  });

  it("calls onRetryFailed when Retry Failed button is clicked and failures exist", () => {
    render(
      <BulkPrintResultModal
        open={true}
        result={makeResult({
          failed: [{ order_sn: "ORD-FAIL", error: "error" }],
        })}
        onClose={onClose}
        onRetryFailed={onRetryFailed}
        isRetrying={false}
      />,
    );
    fireEvent.click(screen.getByRole("button", { name: /retry failed/i }));
    expect(onRetryFailed).toHaveBeenCalled();
  });

  it("Retry Failed button is disabled when there are no failures", () => {
    render(
      <BulkPrintResultModal
        open={true}
        result={makeResult({ succeeded: ["ORD-001"] })}
        onClose={onClose}
        onRetryFailed={onRetryFailed}
        isRetrying={false}
      />,
    );
    const retryBtn = screen.getByRole("button", {
      name: /retry failed \(0\)/i,
    });
    expect(retryBtn).toBeDisabled();
  });

  it("shows loading state on Retry Failed button when isRetrying is true", () => {
    render(
      <BulkPrintResultModal
        open={true}
        result={makeResult({
          failed: [{ order_sn: "ORD-FAIL", error: "err" }],
        })}
        onClose={onClose}
        onRetryFailed={onRetryFailed}
        isRetrying={true}
      />,
    );
    // Ant Design renders a loading icon inside the button when loading prop is true
    const retryBtn = screen.getByRole("button", { name: /retry failed/i });
    expect(retryBtn).toBeInTheDocument();
  });

  it("shows empty text for no downloaded labels", () => {
    render(
      <BulkPrintResultModal
        open={true}
        result={makeResult()}
        onClose={onClose}
        onRetryFailed={onRetryFailed}
        isRetrying={false}
      />,
    );
    expect(screen.getByText("No downloaded labels")).toBeInTheDocument();
    expect(screen.getByText("No failed labels")).toBeInTheDocument();
  });
});
