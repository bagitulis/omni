import { fireEvent, render, screen } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import "@testing-library/jest-dom";
import { ColumnManager } from "./ColumnManager";
import type { ColumnConfig } from "@/types/shared";

Object.defineProperty(window, "matchMedia", {
  writable: true,
  value: vi.fn().mockImplementation((query) => ({
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

describe("ColumnManager", () => {
  const columns: ColumnConfig[] = [
    { key: "sku", title: "SKU", visible: true, order: 2 },
    {
      key: "name",
      title: "Product Name",
      visible: true,
      locked: true,
      order: 0,
    },
    { key: "stock", title: "Stock", visible: false, order: 1 },
  ];

  const onChange = vi.fn<(columns: ColumnConfig[]) => void>();
  const onReset = vi.fn();

  const setup = () =>
    render(
      <ColumnManager columns={columns} onChange={onChange} onReset={onReset} />,
    );

  const openPopover = async () => {
    fireEvent.click(screen.getByTestId("column-manager-trigger"));
    await screen.findByTestId("column-manager-content");
  };

  beforeEach(() => {
    onChange.mockReset();
    onReset.mockReset();
  });

  it("opens popover from settings button and sorts rows by order", async () => {
    setup();
    await openPopover();

    const orderedRows = screen
      .getAllByTestId(/column-row-/)
      .map((row) => row.getAttribute("data-testid"));

    expect(orderedRows).toEqual([
      "column-row-name",
      "column-row-stock",
      "column-row-sku",
    ]);
  });

  it("toggles visibility for unlocked columns", async () => {
    setup();
    await openPopover();

    fireEvent.click(screen.getByTestId("column-checkbox-stock"));

    expect(onChange).toHaveBeenCalledTimes(1);
    const nextColumns = onChange.mock.calls[0][0];
    const stockColumn = nextColumns.find((column) => column.key === "stock");

    expect(stockColumn?.visible).toBe(true);
  });

  it("enforces lock constraints: non-draggable, disabled checkbox, locked marker", async () => {
    setup();
    await openPopover();

    const lockedRow = screen.getByTestId("column-row-name");
    const lockedCheckbox = screen.getByRole("checkbox", {
      name: /Product Name/i,
    });

    expect(lockedRow).toHaveAttribute("draggable", "false");
    expect(lockedCheckbox).toBeDisabled();
    expect(screen.getByTestId("column-locked-marker-name")).toBeInTheDocument();

    fireEvent.click(lockedCheckbox);
    expect(onChange).not.toHaveBeenCalled();
  });

  it("reorders non-locked rows with native drag events", async () => {
    setup();
    await openPopover();

    const skuRow = screen.getByTestId("column-row-sku");
    const stockRow = screen.getByTestId("column-row-stock");
    const dataTransfer = {
      effectAllowed: "all",
      dropEffect: "move",
      setData: vi.fn(),
      getData: vi.fn(),
      clearData: vi.fn(),
    } as unknown as DataTransfer;

    fireEvent.dragStart(skuRow, { dataTransfer });
    fireEvent.dragOver(stockRow, { dataTransfer });
    fireEvent.dragEnd(skuRow, { dataTransfer });

    expect(onChange).toHaveBeenCalledTimes(1);
    const reordered = onChange.mock.calls[0][0];

    expect(reordered.map((column) => column.key)).toEqual([
      "name",
      "sku",
      "stock",
    ]);
    expect(reordered.map((column) => column.order)).toEqual([0, 1, 2]);
  });

  it("invokes reset callback from reset button", async () => {
    setup();
    await openPopover();

    fireEvent.click(screen.getByTestId("column-reset-button"));

    expect(onReset).toHaveBeenCalledTimes(1);
  });
});
