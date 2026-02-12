import { Modal, Button, Empty, Spin, Typography, Tag, theme } from "antd";
import {
  FileTextOutlined,
  CalendarOutlined,
  DatabaseOutlined,
  DownloadOutlined,
} from "@ant-design/icons";
import type { MLReport } from "@/api/analyticsIntelligence";

const { Text } = Typography;
const { useToken } = theme;

interface ReportModalProps {
  report: MLReport | null;
  open: boolean;
  onClose: () => void;
  htmlContent: string | null;
  loading: boolean;
}

export const ReportModal = ({
  report,
  open,
  onClose,
  htmlContent,
  loading,
}: ReportModalProps) => {
  const { token } = useToken();

  const handleDownload = () => {
    // Ideally this would trigger a download.
    // For now, since we display HTML, user can print/save page,
    // or we could implement blob download if needed.
    // Placeholder behavior matching AIReportGalleryPage intention.
    if (report) {
      console.log("Downloading report:", report.file_name);
    }
  };

  const formatDate = (dateString?: string) => {
    if (!dateString) return "";
    return new Date(dateString).toLocaleDateString("id-ID", {
      year: "numeric",
      month: "long",
      day: "numeric",
      hour: "2-digit",
      minute: "2-digit",
    });
  };

  const formatFileSize = (bytes?: number) => {
    if (bytes === undefined || bytes === 0) return "0 Bytes";
    const k = 1024;
    const sizes = ["Bytes", "KB", "MB", "GB"];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return `${parseFloat((bytes / k ** i).toFixed(2))} ${sizes[i]}`;
  };

  return (
    <Modal
      title={
        <div style={{ display: "flex", alignItems: "center", gap: 8 }}>
          <FileTextOutlined style={{ color: token.colorPrimary }} />
          <span>{report?.file_name || "Report Viewer"}</span>
        </div>
      }
      open={open}
      onCancel={onClose}
      width="90%"
      style={{ top: 20 }}
      footer={[
        <Button
          key="download"
          icon={<DownloadOutlined />}
          onClick={handleDownload}
        >
          Download
        </Button>,
        <Button key="close" onClick={onClose}>
          Close
        </Button>,
      ]}
    >
      {report && (
        <div
          style={{
            marginBottom: 16,
            padding: 12,
            background: token.colorFillAlter,
            borderRadius: token.borderRadius,
            display: "flex",
            gap: 16,
            flexWrap: "wrap",
            alignItems: "center",
          }}
        >
          <Tag color={report.platform === "shopee" ? "orange" : "black"}>
            {report.platform.toUpperCase()}
          </Tag>
          <Text type="secondary" style={{ fontSize: 12 }}>
            <CalendarOutlined style={{ marginRight: 4 }} />
            {formatDate(report.created_at)}
          </Text>
          <Text type="secondary" style={{ fontSize: 12 }}>
            <DatabaseOutlined style={{ marginRight: 4 }} />
            {formatFileSize(report.file_size)}
          </Text>
        </div>
      )}

      <div
        style={{
          minHeight: "60vh",
          maxHeight: "75vh",
          overflowY: "auto",
          padding: 16,
          background: token.colorBgContainer,
          border: `1px solid ${token.colorBorderSecondary}`,
          borderRadius: token.borderRadius,
        }}
      >
        {loading ? (
          <div style={{ textAlign: "center", padding: 80 }}>
            <Spin size="large" tip="Loading report content..." />
          </div>
        ) : htmlContent ? (
          <iframe
            title="ML report content"
            srcDoc={htmlContent}
            style={{
              width: "100%",
              minHeight: "58vh",
              border: "none",
              background: token.colorBgContainer,
            }}
          />
        ) : (
          <Empty description="Report content not available" />
        )}
      </div>
    </Modal>
  );
};

export default ReportModal;
