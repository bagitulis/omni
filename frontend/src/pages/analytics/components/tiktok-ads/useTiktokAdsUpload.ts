import { useState } from "react";
import { UploadProps } from "antd";
import { useQueryClient } from "@tanstack/react-query";
import apiClient from "@/api/client";
import { message } from "@/components/AntStaticApi";

interface UploadResult {
  totalRows: number;
  period: {
    start: string;
    end: string;
    label: string;
  };
}

export const useTiktokAdsUpload = () => {
  const [uploading, setUploading] = useState(false);
  const [lastResult, setLastResult] = useState<UploadResult | null>(null);
  const queryClient = useQueryClient();

  const uploadProps: UploadProps = {
    name: "file",
    accept: ".xlsx,.xls",
    multiple: false,
    showUploadList: true,
    customRequest: async ({ file, onSuccess, onError }) => {
      setUploading(true);
      const formData = new FormData();
      formData.append("file", file as Blob);

      try {
        const response = await apiClient.post<UploadResult>(
          "/ads/tiktok/upload",
          formData,
          { headers: { "Content-Type": "multipart/form-data" } },
        );

        if (response.success && response.data) {
          setLastResult(response.data);
          message.success(
            `Upload successful — ${response.data.totalRows} rows processed (${response.data.period?.label || "unknown"})`,
          );
          // Refresh data queries
          queryClient.invalidateQueries({ queryKey: ["tiktok-ads-data"] });
          queryClient.invalidateQueries({
            queryKey: ["tiktok-ads-dashboard"],
          });
          onSuccess?.(response.data);
        } else {
          const errMsg =
            (response as unknown as { error?: string }).error ||
            "Upload failed";
          message.error(errMsg);
          onError?.(new Error(errMsg));
        }
      } catch (err) {
        const errMsg =
          err instanceof Error ? err.message : "Upload failed unexpectedly";
        // Check for duplicate period error
        if (errMsg.includes("already uploaded")) {
          message.warning(errMsg);
        } else {
          message.error(errMsg);
        }
        onError?.(err instanceof Error ? err : new Error(errMsg));
      } finally {
        setUploading(false);
      }
    },
  };

  return { uploading, lastResult, uploadProps };
};
