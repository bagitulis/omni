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


