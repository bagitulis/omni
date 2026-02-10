import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import {
  getUnifiedAnalytics,
  getClassifiedProducts,
  getProductsFromAds,
  runSimulation,
  generateReport,
  getReports,
  getReportHTML,
  type SimulationRequest,
  type GenerateReportRequest,
} from "../api/analyticsIntelligence";

// --- Unified Analytics ---

export function useUnifiedAnalytics() {
  return useQuery({
    queryKey: ["analytics", "unified"],
    queryFn: getUnifiedAnalytics,
    staleTime: 5 * 60 * 1000, // 5 minutes
  });
}

export function useClassifiedProducts() {
  return useQuery({
    queryKey: ["analytics", "classification"],
    queryFn: getClassifiedProducts,
    staleTime: 5 * 60 * 1000, // 5 minutes
  });
}

// --- Budget Simulator ---

export function useProductsFromAds() {
  return useQuery({
    queryKey: ["analytics", "products-ads"],
    queryFn: getProductsFromAds,
    staleTime: 10 * 60 * 1000, // 10 minutes
  });
}

export function useBudgetSimulation() {
  return useMutation({
    mutationFn: (request: SimulationRequest) => runSimulation(request),
  });
}

// --- AI Reports ---

export function useGenerateReport() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (request: GenerateReportRequest) => generateReport(request),
    onSuccess: (_, variables) => {
      queryClient.invalidateQueries({
        queryKey: ["analytics", "reports", variables.platform],
      });
    },
  });
}

export function useReports(
  platform: "shopee" | "tiktok",
  page: number = 1,
  limit: number = 20,
) {
  return useQuery({
    queryKey: ["analytics", "reports", platform, page, limit],
    queryFn: () => getReports(platform, page, limit),
    staleTime: 60 * 1000, // 1 minute
  });
}

export function useReportHTML(
  platform: "shopee" | "tiktok",
  filename: string | null,
) {
  return useQuery({
    queryKey: ["analytics", "report-html", platform, filename],
    queryFn: () => getReportHTML(platform, filename!),
    enabled: !!filename,
    staleTime: Infinity, // HTML content is static once generated
  });
}
