import { useState } from "react";
import { UploadProps, message } from "antd";
import dayjs from "dayjs";
import { TikTokAdsData } from "./types";

export const useTiktokAdsUpload = () => {
  const [uploadedData, setUploadedData] = useState<TikTokAdsData[]>([]);

  const uploadProps: UploadProps = {
    name: "file",
    accept: ".csv",
    multiple: false,
    showUploadList: false,
    beforeUpload: (file) => {
      message.loading("Processing CSV...", 1);
      const reader = new FileReader();
      reader.onload = (e) => {
        try {
          const text = e.target?.result as string;
          const lines = text.split("\n").filter((l) => l.trim());
          if (lines.length < 2) {
            message.error("CSV file is empty or invalid");
            return;
          }
          // Parse CSV
          const data: TikTokAdsData[] = [];
          for (let i = 1; i < lines.length; i++) {
            const values = lines[i].split(",");
            if (values.length >= 5) {
              const views = parseInt(values[4]) || 0;
              const clicks = parseInt(values[5]) || 0;
              const cost = parseFloat(values[2]) || 0;
              const revenue = parseFloat(values[3]) || 0;
              const videoPlays = Math.floor(views * 0.7);

              data.push({
                creative_id: values[0]?.trim() || `TT${i}`,
                creative_name: values[1]?.trim() || `Creative ${i}`,
                campaign_name:
                  values[8]?.trim() || values[1]?.trim() || `Campaign ${i}`,
                product_id: values[9]?.trim() || `P${i}`,
                creative_type: values[10]?.trim() || "Video",
                cost,
                revenue,
                views,
                clicks,
                ctr: views > 0 ? (clicks / views) * 100 : 0,
                cpc: clicks > 0 ? cost / clicks : 0,
                roi: cost > 0 ? ((revenue - cost) / cost) * 100 : 0,
                conversions: parseInt(values[6]) || 0,
                video_plays: videoPlays,
                engagement_rate:
                  views > 0
                    ? ((clicks + (parseInt(values[6]) || 0)) / views) * 100
                    : 0,
                date: values[7]?.trim() || dayjs().format("YYYY-MM-DD"),
              });
            }
          }
          setUploadedData(data);
          message.success(
            `${file.name} processed - ${data.length} creatives loaded`,
          );
        } catch {
          message.error("Failed to parse CSV file");
        }
      };
      reader.readAsText(file);
      return false;
    },
  };

  return { uploadedData, setUploadedData, uploadProps };
};
