import { useQuery } from "@tanstack/react-query";
import { getUnifiedKPI, getUnifiedSummary } from "@/api/analytics";

export function useAnalyticsHub() {
  const kpiQuery = useQuery({
    queryKey: ["analytics", "unified", "kpi"],
    queryFn: getUnifiedKPI,
    staleTime: 5 * 60 * 1000, // 5 minutes
  });

  const summaryQuery = useQuery({
    queryKey: ["analytics", "unified", "summary"],
    queryFn: getUnifiedSummary,
    staleTime: 5 * 60 * 1000, // 5 minutes
  });

  const isLoading = kpiQuery.isLoading || summaryQuery.isLoading;
  const error = kpiQuery.error || summaryQuery.error;
  const refetch = () => {
    kpiQuery.refetch();
    summaryQuery.refetch();
  };

  return {
    kpi: kpiQuery.data,
    summary: summaryQuery.data,
    isLoading,
    error,
    refetch,
  };
}
