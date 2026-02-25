import { render, screen } from "@testing-library/react";
import "@testing-library/jest-dom/vitest";
import { describe, it, expect, vi } from "vitest";
import { BudgetSimulatorPage } from "./BudgetSimulatorPage";

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
vi.mock("@/hooks/useAnalyticsIntelligence", () => ({
  useProductsFromAds: () => ({
    data: [{ product_id: "p1", product_name: "Test Product 1", sku: "SKU-1" }],
    isLoading: false,
  }),
  useBudgetSimulation: () => ({
    mutate: vi.fn(),
    isPending: false,
    data: null,
    error: null,
  }),
}));

vi.mock("./components/budget/SimulationParameters", () => ({
  SimulationParameters: () => <div data-testid="simulation-parameters" />,
}));
vi.mock("./components/budget/SimulationResults", () => ({
  SimulationResults: () => <div data-testid="simulation-results" />,
}));

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------
describe("BudgetSimulatorPage", () => {
  it("renders title and description", () => {
    render(<BudgetSimulatorPage />);
    expect(screen.getByText("Budget Simulator")).toBeInTheDocument();
    expect(screen.getByText(/Predict ROAS and optimize/i)).toBeInTheDocument();
  });

  it("renders parameters and results components", () => {
    render(<BudgetSimulatorPage />);
    expect(screen.getByTestId("simulation-parameters")).toBeInTheDocument();
    expect(screen.getByTestId("simulation-results")).toBeInTheDocument();
  });
});
