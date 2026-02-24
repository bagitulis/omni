import { render, screen, fireEvent } from "@testing-library/react";
import { describe, expect, it, vi, beforeEach } from "vitest";
import { OrdersBulkActionsBar } from "./OrdersBulkActionsBar";
import type { BulkResult, ProgressState } from "../hooks/bulkActionTypes";

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

function makeProgress(overrides: Partial<ProgressState> = {}): ProgressState {
  return { current: 0, total: 0, status: "idle", ...overrides };
}

function makeResult(overrides: Partial<BulkResult> = {}): BulkResult {
  return { succeeded: [], failed: [], ...overrides };
}

function makeProps(overrides = {}) {
  return {
    selectedCount: 2,
    onBulkShip: vi.fn(),
    onBulkPrint: vi.fn(),
    onBulkCancel: vi.fn(),
    onRetryFailedPrint: vi.fn(),
    onClearSelection: vi.fn(),
    isShipping: false,
    isPrinting: false,
    isCancelling: false,
    shipProgress: makeProgress(),
    printProgress: makeProgress(),
    cancelProgress: makeProgress(),
    shipResult: makeResult(),
    printResult: makeResult(),
    cancelResult: makeResult(),
    ...overrides,
  };
}

describe("OrdersBulkActionsBar", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("renders nothing when selectedCount is 0", () => {
    const { container } = render(
      <OrdersBulkActionsBar {...makeProps({ selectedCount: 0 })} />,
    );
    expect(container.firstChild).toBeNull();
  });

  it("shows selected count", () => {
    render(<OrdersBulkActionsBar {...makeProps({ selectedCount: 5 })} />);
    expect(screen.getByText("5 orders selected")).toBeInTheDocument();
  });

  it("renders Bulk Ship, Bulk Print Labels, Bulk Cancel buttons", () => {
    render(<OrdersBulkActionsBar {...makeProps()} />);
    expect(
      screen.getByRole("button", { name: /bulk ship/i }),
    ).toBeInTheDocument();
    expect(
      screen.getByRole("button", { name: /bulk print labels/i }),
    ).toBeInTheDocument();
    expect(
      screen.getByRole("button", { name: /bulk cancel/i }),
    ).toBeInTheDocument();
  });

  it("calls onBulkShip when Bulk Ship is clicked", () => {
    const props = makeProps();
    render(<OrdersBulkActionsBar {...props} />);
    fireEvent.click(screen.getByRole("button", { name: /bulk ship/i }));
    expect(props.onBulkShip).toHaveBeenCalled();
  });

  it("calls onBulkPrint when Bulk Print Labels is clicked", () => {
    const props = makeProps();
    render(<OrdersBulkActionsBar {...props} />);
    fireEvent.click(screen.getByRole("button", { name: /bulk print labels/i }));
    expect(props.onBulkPrint).toHaveBeenCalled();
  });

  it("calls onBulkCancel when Bulk Cancel is clicked", () => {
    const props = makeProps();
    render(<OrdersBulkActionsBar {...props} />);
    fireEvent.click(screen.getByRole("button", { name: /bulk cancel/i }));
    expect(props.onBulkCancel).toHaveBeenCalled();
  });

  it("calls onClearSelection when Clear Selection is clicked", () => {
    const props = makeProps();
    render(<OrdersBulkActionsBar {...props} />);
    fireEvent.click(screen.getByRole("button", { name: /clear selection/i }));
    expect(props.onClearSelection).toHaveBeenCalled();
  });

  it("shows ship progress when shipProgress.status is processing", () => {
    render(
      <OrdersBulkActionsBar
        {...makeProps({
          shipProgress: makeProgress({
            status: "processing",
            current: 2,
            total: 5,
          }),
        })}
      />,
    );
    expect(screen.getByText("Shipping 2/5...")).toBeInTheDocument();
  });

  it("shows print progress when printProgress.status is processing", () => {
    render(
      <OrdersBulkActionsBar
        {...makeProps({
          printProgress: makeProgress({
            status: "processing",
            current: 1,
            total: 3,
          }),
        })}
      />,
    );
    expect(screen.getByText("Printing 1/3...")).toBeInTheDocument();
  });

  it("shows cancel progress when cancelProgress.status is processing", () => {
    render(
      <OrdersBulkActionsBar
        {...makeProps({
          cancelProgress: makeProgress({
            status: "processing",
            current: 3,
            total: 4,
          }),
        })}
      />,
    );
    expect(screen.getByText("Cancelling 3/4...")).toBeInTheDocument();
  });

  it("shows print result summary when print is done with successes", () => {
    render(
      <OrdersBulkActionsBar
        {...makeProps({
          printProgress: makeProgress({ status: "done", current: 3, total: 3 }),
          printResult: makeResult({ succeeded: ["A", "B", "C"] }),
        })}
      />,
    );
    expect(screen.getByText(/✓ 3 printed/)).toBeInTheDocument();
  });

  it("shows print result with failure count when some prints fail", () => {
    render(
      <OrdersBulkActionsBar
        {...makeProps({
          printProgress: makeProgress({ status: "done", current: 2, total: 2 }),
          printResult: makeResult({
            succeeded: ["A"],
            failed: [{ order_sn: "B", error: "error" }],
          }),
        })}
      />,
    );
    expect(screen.getByText("1 printed, 1 failed")).toBeInTheDocument();
  });

  it("shows View Print Details button when print done and results exist", () => {
    render(
      <OrdersBulkActionsBar
        {...makeProps({
          printProgress: makeProgress({ status: "done", current: 2, total: 2 }),
          printResult: makeResult({ succeeded: ["A", "B"] }),
        })}
      />,
    );
    expect(
      screen.getByRole("button", { name: /view print details/i }),
    ).toBeInTheDocument();
  });

  it("opens BulkPrintResultModal when View Print Details is clicked", () => {
    render(
      <OrdersBulkActionsBar
        {...makeProps({
          printProgress: makeProgress({ status: "done", current: 2, total: 2 }),
          printResult: makeResult({ succeeded: ["A", "B"] }),
        })}
      />,
    );
    fireEvent.click(
      screen.getByRole("button", { name: /view print details/i }),
    );
    expect(screen.getByText("Bulk Print Result Details")).toBeInTheDocument();
  });

  it("shows ship result summary when ship is done", () => {
    render(
      <OrdersBulkActionsBar
        {...makeProps({
          shipProgress: makeProgress({ status: "done", current: 2, total: 2 }),
          shipResult: makeResult({ succeeded: ["A", "B"] }),
        })}
      />,
    );
    expect(screen.getByText(/✓ 2 shipped/)).toBeInTheDocument();
  });
});
