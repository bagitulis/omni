import React from "react";
import { describe, expect, it, vi } from "vitest";
import { render } from "@testing-library/react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import GoogleSheetsTab from "./GoogleSheetsTab";

const useGoogleSheetsLinks = vi.fn();
const useGoogleSheetsDetails = vi.fn();
const useSaveLinks = vi.fn();
const useUpdateSettings = vi.fn();
const useValidateLink = vi.fn();

vi.mock("@/hooks/useGoogleSheets", () => ({
  useGoogleSheetsLinks: () => useGoogleSheetsLinks(),
  useGoogleSheetsDetails: () => useGoogleSheetsDetails(),
  useSaveLinks: () => useSaveLinks(),
  useUpdateSettings: () => useUpdateSettings(),
  useValidateLink: () => useValidateLink(),
}));

const createWrapper = () => {
  const client = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  });
  return ({ children }: { children: React.ReactNode }) => (
    <QueryClientProvider client={client}>{children}</QueryClientProvider>
  );
};

describe("GoogleSheetsTab", () => {
  it("hydrates selected sheets from details", () => {
    useGoogleSheetsLinks.mockReturnValue({
      data: {
        inventory_url: "https://docs.google.com/spreadsheets/d/inventory",
        wallet_url: "https://docs.google.com/spreadsheets/d/wallet",
        shipping_url: "https://docs.google.com/spreadsheets/d/shipping",
        order_url: "https://docs.google.com/spreadsheets/d/order",
      },
      isLoading: false,
    });
    useGoogleSheetsDetails.mockReturnValue({
      data: {
        inventory_sheet_name: "InventorySheet",
        wallet_sheet_name: "WalletSheet",
        shipping_sheet_name: "ShippingSheet",
        order_sheet_name: "OrderSheet",
      },
      isLoading: false,
    });
    useValidateLink.mockReturnValue({
      mutate: vi.fn(),
      isPending: false,
    });
    useSaveLinks.mockReturnValue({
      mutate: vi.fn(),
      isPending: false,
    });
    useUpdateSettings.mockReturnValue({
      mutate: vi.fn(),
      isPending: false,
    });

    const { getByText } = render(<GoogleSheetsTab />, {
      wrapper: createWrapper(),
    });

    expect(getByText("Google Sheets Settings")).toBeTruthy();
  });
});
