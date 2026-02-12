import {
  Card,
  Typography,
  Button,
  Select,
  Row,
  Col,
  Empty,
  Spin,
  Tag,
  theme,
  message,
} from "antd";
import {
  RobotOutlined,
  DownloadOutlined,
  EyeOutlined,
  FileTextOutlined,
  CalendarOutlined,
  DatabaseOutlined,
} from "@ant-design/icons";
import { useState } from "react";
import {
  useReports,
  useGenerateReport,
  useReportHTML,
} from "../../hooks/useAnalyticsIntelligence";
import type { MLReport } from "../../api/analyticsIntelligence";
import { ReportModal } from "@/components/analytics/ml";

const { Title, Text } = Typography;
const { useToken } = theme;

export const AIReportGalleryPage = () => {
  const { token } = useToken();
  const [platform, setPlatform] = useState<"shopee" | "tiktok">("tiktok");
  const [selectedReport, setSelectedReport] = useState<MLReport | null>(null);

  const { data, isLoading, refetch } = useReports(platform);
  const { mutate: generate, isPending: generating } = useGenerateReport();

  // HTML content for preview
  const { data: reportHtml, isLoading: loadingHtml } = useReportHTML(
    selectedReport?.platform || "tiktok",
    selectedReport?.file_name || null,
  );

  const handleGenerate = () => {
    generate(
      { platform, report_type: "full" },
      {
        onSuccess: () => {
          message.success("Report generation started");
          refetch();
        },
        onError: (err: Error) => {
          message.error(`Failed to generate report: ${err.message}`);
        },
      },
    );
  };

  const handleView = (report: MLReport) => {
    setSelectedReport(report);
  };

  const handleDownload = (report: MLReport) => {
    // This assumes the backend serves the file or we can download the HTML
    // For now, we'll assume we can create a blob from the HTML if we fetch it,
    // or trigger a direct download if the API supports it.
    // Given the hook fetches HTML, let's wait for HTML then download.
    // Actually, simpler to just open the HTML in a new window/tab or trigger download
    // Let's implement a simple download function if we have the content,
    // but without content we can't easily download.
    // Ideally the API would provide a download endpoint.
    // For this MVP, let's open the view modal and use a print/save function there
    // or just show a message that download is handled via view.
    setSelectedReport(report);
  };

  const formatDate = (dateString: string) => {
    return new Date(dateString).toLocaleDateString("id-ID", {
      year: "numeric",
      month: "long",
      day: "numeric",
      hour: "2-digit",
      minute: "2-digit",
    });
  };

  const formatFileSize = (bytes: number) => {
    if (bytes === 0) return "0 Bytes";
    const k = 1024;
    const sizes = ["Bytes", "KB", "MB", "GB"];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + " " + sizes[i];
  };

  return (
    <div style={{ padding: 24 }}>
      <div style={{ marginBottom: 24 }}>
        <div
          style={{
            display: "flex",
            justifyContent: "space-between",
            alignItems: "center",
            marginBottom: 8,
          }}
        >
          <Title level={2} style={{ margin: 0 }}>
            <RobotOutlined style={{ marginRight: 8 }} />
            AI Report Gallery
          </Title>
          <Button
            type="primary"
            icon={<RobotOutlined />}
            onClick={handleGenerate}
            loading={generating}
          >
            Generate New Report
          </Button>
        </div>
        <Text type="secondary">
          Browse and view ML-generated analytics reports
        </Text>
      </div>

      <div style={{ marginBottom: 24 }}>
        <Text strong style={{ marginRight: 8 }}>
          Platform:
        </Text>
        <Select
          value={platform}
          onChange={setPlatform}
          style={{ width: 200 }}
          options={[
            { value: "tiktok", label: "TikTok Shop" },
            { value: "shopee", label: "Shopee" },
          ]}
        />
      </div>

      {isLoading ? (
        <div style={{ textAlign: "center", padding: 80 }}>
          <Spin size="large" />
          <div style={{ marginTop: 16 }}>Loading reports...</div>
        </div>
      ) : !data?.reports || data.reports.length === 0 ? (
        <Empty
          image={Empty.PRESENTED_IMAGE_SIMPLE}
          description="No reports found. Generate one to get started!"
        />
      ) : (
        <Row gutter={[16, 16]}>
          {data.reports.map((report: MLReport) => (
            <Col xs={24} sm={12} md={8} lg={6} key={report.id}>
              <Card
                hoverable
                style={{ borderRadius: token.borderRadiusLG }}
                actions={[
                  <Button
                    key="view"
                    type="text"
                    icon={<EyeOutlined />}
                    onClick={() => handleView(report)}
                  >
                    View
                  </Button>,
                  <Button
                    key="download"
                    type="text"
                    icon={<DownloadOutlined />}
                    onClick={() => handleDownload(report)}
                  >
                    Download
                  </Button>,
                ]}
              >
                <Card.Meta
                  avatar={
                    <div
                      style={{
                        background: token.colorFillAlter,
                        padding: 8,
                        borderRadius: "50%",
                      }}
                    >
                      <FileTextOutlined
                        style={{ fontSize: 24, color: token.colorPrimary }}
                      />
                    </div>
                  }
                  title={report.period_label || report.file_name}
                  description={
                    <div
                      style={{
                        display: "flex",
                        flexDirection: "column",
                        gap: 4,
                        marginTop: 8,
                      }}
                    >
                      <Tag
                        color={
                          report.platform === "shopee" ? "orange" : "black"
                        }
                        style={{ width: "fit-content" }}
                      >
                        {report.platform.toUpperCase()}
                      </Tag>
                      <div style={{ fontSize: 12 }}>
                        <CalendarOutlined style={{ marginRight: 4 }} />
                        {formatDate(report.created_at)}
                      </div>
                      <div style={{ fontSize: 12 }}>
                        <DatabaseOutlined style={{ marginRight: 4 }} />
                        {formatFileSize(report.file_size)}
                      </div>
                    </div>
                  }
                />
              </Card>
            </Col>
          ))}
        </Row>
      )}

      {/* Report Viewer Modal */}
      <ReportModal
        report={selectedReport}
        open={!!selectedReport}
        onClose={() => setSelectedReport(null)}
        htmlContent={reportHtml || null}
        loading={loadingHtml}
      />
    </div>
  );
};

export default AIReportGalleryPage;
