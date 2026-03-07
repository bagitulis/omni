import { useQuery } from "@tanstack/react-query";
import * as adsApi from "@/api/ads";
import { ReportInfo } from "@/types/ads";

export function useAdsReports(platform: "shopee" | "tiktok") {
  const shopeeQuery = useQuery({
    queryKey: ["shopee-ads-reports"],
    queryFn: () => adsApi.getShopeeAdsReports(),
    enabled: platform === "shopee",
  });

  const tiktokQuery = useQuery({
    queryKey: ["tiktok-ads-reports"],
    queryFn: () => adsApi.getTiktokAdsReports(),
    enabled: platform === "tiktok",
  });

  const query = platform === "shopee" ? shopeeQuery : tiktokQuery;

  const reports = query.data?.success ? query.data.data : [];

  const getReportFileUrl = (filename: string) => {
    // Assuming the backend serves reports at /reports/{platform}/{filename}
    // We need to construct the full URL if it's external, or relative to API base
    // Since this is likely served by the same backend, we can use a relative path
    // OR we can assume the backend returns a full URL if we implemented it that way.
    // Based on Vue code: `${baseUrl}/reports/${platform}/${filename}`
    // We can use the same logic.
    return `/api/analytics/ml/reports/${platform}/${filename}`;
  };

  return {
    reports: reports as ReportInfo[],
    loading: query.isLoading,
    error: query.error,
    refetch: query.refetch,
    getReportFileUrl,
  };
}
