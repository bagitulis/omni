import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { getWalletData } from "@/api/dashboard";
import apiClient from "@/api/client";
import { message } from "antd";

/**
 * Wallet transaction type - matches backend response
 */
export interface WalletTransaction {
  transaction_id: string;
  type: string;
  amount: number;
  description: string;
  created_at: string;
  status?: string;
}

/**
 * Hook to fetch wallet balance
 * Uses existing getWalletData from @/api/dashboard
 */
export function useWalletBalance() {
  return useQuery({
    queryKey: ["wallet", "shopee"],
    queryFn: () => getWalletData("shopee"),
    staleTime: 60 * 1000, // 1 minute
  });
}

/**
 * Hook to fetch wallet transactions with optional month/year filters
 * Backend route: GET /api/shopee/wallet/transactions
 */
export function useWalletTransactions(params?: {
  month?: number;
  year?: number;
}) {
  return useQuery({
    queryKey: ["wallet-transactions", "shopee", params],
    queryFn: async () => {
      const response = await apiClient.get<{
        transactions: WalletTransaction[];
        total: number;
      }>("/shopee/wallet/transactions", { params });

      if (!response.success) {
        throw new Error(response.error || "Failed to fetch transactions");
      }

      return response.data!;
    },
    staleTime: 60 * 1000, // 1 minute
  });
}

/**
 * Hook to export wallet transactions to Google Sheets
 * Backend route: POST /api/shopee/wallet/export-to-sheets
 */
export function useExportWalletToSheets() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (params?: { month?: number; year?: number }) => {
      const response = await apiClient.post<{ sheet_url: string }>(
        "/shopee/wallet/export-to-sheets",
        params,
      );

      if (!response.success) {
        throw new Error(response.error || "Failed to export wallet data");
      }

      return response.data!;
    },
    onSuccess: (data) => {
      message.success("Wallet data exported to Google Sheets successfully");
      if (data.sheet_url) {
        window.open(data.sheet_url, "_blank");
      }
      // Invalidate transactions query to refresh data
      queryClient.invalidateQueries({ queryKey: ["wallet-transactions"] });
    },
    onError: (error: Error) => {
      message.error(error.message || "Failed to export wallet data");
    },
  });
}
