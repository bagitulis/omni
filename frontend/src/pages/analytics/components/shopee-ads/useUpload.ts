import { useState } from "react";
import { UploadProps } from "antd";
import dayjs from "dayjs";
import { AdsData } from "./types";
import { message } from "@/components/AntStaticApi";

export const useUpload = () => {
  const [uploadedData, setUploadedData] = useState<AdsData[]>([]);

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
          // Parse CSV (simple parser - assumes comma-separated)
          const data: AdsData[] = [];
          for (let i = 1; i < lines.length; i++) {
            const values = lines[i].split(",");
            if (values.length >= 5) {
              data.push({
                product_id: values[0]?.trim() || `PROD${i}`,
                product_name: values[1]?.trim() || `Product ${i}`,
                bidding_mode: values[8]?.trim() || "Manual",
                cost: parseFloat(values[2]) || 0,
                revenue: parseFloat(values[3]) || 0,
                clicks: parseInt(values[4]) || 0,
                impressions: parseInt(values[5]) || 0,
                ctr: 0,
                cpc: 0,
                roas: 0,
                conversions: parseInt(values[6]) || 0,
                date: values[7]?.trim() || dayjs().format("YYYY-MM-DD"),
                period_label: values[9]?.trim() || "Uploaded CSV",
              });
            }
          }
          // Calculate derived metrics
          data.forEach((d) => {
            d.ctr = d.impressions > 0 ? (d.clicks / d.impressions) * 100 : 0;
            d.cpc = d.clicks > 0 ? d.cost / d.clicks : 0;
            d.roas = d.cost > 0 ? d.revenue / d.cost : 0;
          });
          setUploadedData(data);
          message.success(
            `${file.name} processed - ${data.length} products loaded`,
          );
        } catch (err) { console.warn("Operation failed:", err);
          message.error("Failed to parse CSV file");
        }
      };
      reader.readAsText(file);
      return false;
    },
  };

  return { uploadedData, setUploadedData, uploadProps };
};
