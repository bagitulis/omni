import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { render, screen } from "@testing-library/react";
import type { ReactNode } from "react";
import { beforeAll, describe, expect, it, vi } from "vitest";
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
	return ({ children }: { children: ReactNode }) => (
		<QueryClientProvider client={client}>{children}</QueryClientProvider>
	);
};

describe("GoogleSheetsTab", () => {
	beforeAll(() => {
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
	});

	it("hydrates saved links and shows selected sheet info after reload", () => {
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
		const inputs = screen.getAllByPlaceholderText(
			"https://docs.google.com/spreadsheets/d/...",
		) as HTMLInputElement[];

		expect(getByText("Google Sheets Settings")).toBeTruthy();
		expect(inputs[0].value).toBe(
			"https://docs.google.com/spreadsheets/d/inventory",
		);
		expect(screen.getAllByText("InventorySheet").length).toBeGreaterThan(0);
	});
});
