import React, { useState, useEffect } from "react";
import { Select, Button, Spin, Empty, Card, Alert } from "antd";
import { ReloadOutlined } from "@ant-design/icons";
import { useAdsReports } from "@/hooks/useAdsReports";

interface AdsReportViewerProps {
  platform: "shopee" | "tiktok";
}

export const AdsReportViewer: React.FC<AdsReportViewerProps> = ({
  platform,
}) => {
  const { reports, loading, error, refetch, getReportFileUrl } =
    useAdsReports(platform);

  const [selectedReport, setSelectedReport] = useState<string | null>(null);

  // Auto-select latest report
  useEffect(() => {
    if (reports && reports.length > 0 && !selectedReport) {
      setSelectedReport(reports[0].filename);
    }
  }, [reports, selectedReport]);

  const formatReportName = (filename: string) => {
    const match = filename.match(/(\d{4})(\d{2})(\d{2})/);
    if (match) {
      const date = new Date(`${match[1]}-${match[2]}-${match[3]}`);
      const dateStr = date.toLocaleDateString("en-US", {
        month: "short",
        day: "numeric",
        year: "numeric",
      });
      const name = filename
        .replace(/_\d{8}\.html$/, "")
        .replace(/_/g, " ")
        .replace(/\b\w/g, (c) => c.toUpperCase());
      return `${name} - ${dateStr}`;
    }
    return filename;
  };

  const reportUrl = selectedReport ? getReportFileUrl(selectedReport) : "";

  return (
    <Card
      title="AI Insights Reports"
      extra={
        <div style={{ display: "flex", gap: 8 }}>
          <Select
            style={{ width: 250 }}
            placeholder="Select a report"
            value={selectedReport}
            onChange={setSelectedReport}
            options={reports.map((r) => ({
              label: formatReportName(r.filename),
              value: r.filename,
            }))}
            loading={loading}
          />
          <Button icon={<ReloadOutlined />} onClick={() => refetch()} />
        </div>
      }
      bodyStyle={{ padding: 0, height: 600 }}
    >
      {loading && !reports.length ? (
        <div
          style={{
            height: "100%",
            display: "flex",
            justifyContent: "center",
            alignItems: "center",
          }}
        >
          <Spin size="large" tip="Loading reports..." />
        </div>
      ) : error ? (
        <Alert
          message="Error"
          description={
            error instanceof Error ? error.message : "Failed to load reports"
          }
          type="error"
          showIcon
          style={{ margin: 16 }}
        />
      ) : !reports.length ? (
        <div
          style={{
            height: "100%",
            display: "flex",
            justifyContent: "center",
            alignItems: "center",
          }}
        >
          <Empty description="No AI reports available yet" />
        </div>
      ) : selectedReport ? (
        <iframe
          src={reportUrl}
          style={{ width: "100%", height: "100%", border: "none" }}
          title={`${platform} AI Report`}
          sandbox="allow-same-origin allow-scripts"
        />
      ) : (
        <div
          style={{
            height: "100%",
            display: "flex",
            justifyContent: "center",
            alignItems: "center",
          }}
        >
          <Empty description="Select a report to view" />
        </div>
      )}
    </Card>
  );
};

export default AdsReportViewer;
