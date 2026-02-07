import { useMemo } from "react";
import { AdsData } from "./types";

export interface Summary {
  totalCost: number;
  totalRevenue: number;
  avgRoas: number;
  avgCtr: number;
  avgCpc: number;
  totalConversions: number;
}

export const useSummary = (adsData: AdsData[]): Summary => {
  const hasData = adsData.length > 0;

  return useMemo(() => {
    if (!hasData) {
      return {
        totalCost: 0,
        totalRevenue: 0,
        avgRoas: 0,
        avgCtr: 0,
        avgCpc: 0,
        totalConversions: 0,
      };
    }
    const totalCost = adsData.reduce((sum, d) => sum + d.cost, 0);
    const totalRevenue = adsData.reduce((sum, d) => sum + d.revenue, 0);
    const totalClicks = adsData.reduce((sum, d) => sum + d.clicks, 0);
    const totalImpressions = adsData.reduce((sum, d) => sum + d.impressions, 0);

    return {
      totalCost,
      totalRevenue,
      avgRoas: totalCost > 0 ? totalRevenue / totalCost : 0,
      avgCtr: totalImpressions > 0 ? (totalClicks / totalImpressions) * 100 : 0,
      avgCpc: totalClicks > 0 ? totalCost / totalClicks : 0,
      totalConversions: adsData.reduce((sum, d) => sum + d.conversions, 0),
    };
  }, [adsData, hasData]);
};
