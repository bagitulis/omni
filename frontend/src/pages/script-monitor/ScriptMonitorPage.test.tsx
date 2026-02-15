import { render, screen } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { ScriptMonitorPage } from "./ScriptMonitorPage";

const setSearchParamsMock = vi.fn();

let currentSearchParams = new URLSearchParams();

vi.mock("react-router-dom", () => ({
  useSearchParams: () => [currentSearchParams, setSearchParamsMock],
}));

vi.mock("@/hooks/useScriptMonitor", () => ({
  useScriptMonitor: () => ({
    monitorData: {
      current_job: null,
      pending_queue: [],
      recent_history: [],
      total_pending: 0,
      total_completed: 0,
    },
    isLoadingMonitor: false,
    autoFunctions: [],
    isLoadingAutoFunctions: false,
    cancelJob: vi.fn(),
    forceCancelJob: vi.fn(),
    clearHistory: vi.fn(),
    enableAutoFunction: vi.fn(),
    disableAutoFunction: vi.fn(),
    createAutoFunction: vi.fn(),
    updateAutoFunction: vi.fn(),
    deleteAutoFunction: vi.fn(),
    cancelScheduled: vi.fn(),
  }),
}));

describe("ScriptMonitorPage", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    currentSearchParams = new URLSearchParams();
    Object.defineProperty(window, "matchMedia", {
      writable: true,
      value: vi.fn().mockImplementation((query: string) => ({
        matches: false,
        media: query,
        onchange: null,
        addListener: vi.fn(),
        removeListener: vi.fn(),
        addEventListener: vi.fn(),
        removeEventListener: vi.fn(),
        dispatchEvent: vi.fn(),
      })),
    });
    window.getComputedStyle = vi.fn(() => ({
      getPropertyValue: vi.fn(),
    })) as unknown as typeof window.getComputedStyle;
  });

  it("defaults to current tab when query param missing", () => {
    render(<ScriptMonitorPage />);

    expect(screen.getByText("Current Job")).toBeTruthy();
  });

  it("renders tab labels when tab query param provided", () => {
    currentSearchParams = new URLSearchParams("tab=history");

    render(<ScriptMonitorPage />);

    expect(screen.getByText("History")).toBeTruthy();
  });

  it("falls back to current for unsupported tab", () => {
    currentSearchParams = new URLSearchParams("tab=unknown");

    render(<ScriptMonitorPage />);

    expect(screen.getByText("Current Job")).toBeTruthy();
  });

  it("maps auto-functions tab param to config tab", () => {
    currentSearchParams = new URLSearchParams("tab=auto-functions");

    render(<ScriptMonitorPage />);

    expect(screen.getByText("Auto-Functions")).toBeTruthy();
  });
});
