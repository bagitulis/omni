import { useQuery } from "@tanstack/react-query";
import { getDashboardData } from "@/api/dashboard";
import { DashboardData } from "@/types/dashboard";

export function useDashboard() {
  return useQuery<DashboardData>({
    queryKey: ["dashboard"],
    queryFn: getDashboardData,
    staleTime: 60 * 1000, // 1 minute
    refetchInterval: 60 * 1000, // Refetch every minute
  });
}
