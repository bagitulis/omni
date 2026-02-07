import { useMemo } from "react";
import { TikTokAdsData } from "./types";

export interface TikTokAdsSummary {
  totalCost: number;
  totalRevenue: number;
  avgRoi: number;
  avgCtr: number;
  totalViews: number;
  totalPlays: number;
  totalConversions: number;
}

export const useTiktokAdsSummary = (
  adsData: TikTokAdsData[],
): TikTokAdsSummary => {
  const hasData = adsData.length > 0;

  return useMemo(() => {
    if (!hasData) {
      return {
        totalCost: 0,
        totalRevenue: 0,
        avgRoi: 0,
        avgCtr: 0,
        totalViews: 0,
        totalPlays: 0,
        totalConversions: 0,
      };
    }
    const totalCost = adsData.reduce((sum, d) => sum + d.cost, 0);
    const totalRevenue = adsData.reduce((sum, d) => sum + d.revenue, 0);
    const totalViews = adsData.reduce((sum, d) => sum + d.views, 0);
    const totalClicks = adsData.reduce((sum, d) => sum + d.clicks, 0);
    const totalPlays = adsData.reduce((sum, d) => sum + d.video_plays, 0);

    return {
      totalCost,
      totalRevenue,
      avgRoi:
        totalCost > 0 ? ((totalRevenue - totalCost) / totalCost) * 100 : 0,
      avgCtr: totalViews > 0 ? (totalClicks / totalViews) * 100 : 0,
      totalViews,
      totalPlays,
      totalConversions: adsData.reduce((sum, d) => sum + d.conversions, 0),
    };
  }, [adsData, hasData]);
};
