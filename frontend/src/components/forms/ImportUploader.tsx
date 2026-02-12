import { Upload, message, theme } from "antd";
import { InboxOutlined } from "@ant-design/icons";
import "./ImportUploader.css";

interface ImportUploaderProps {
  onFileSelect: (file: File) => void;
  loading?: boolean;
}

const { Dragger } = Upload;

export function ImportUploader({
  onFileSelect,
  loading = false,
}: ImportUploaderProps) {
  const { token } = theme.useToken();
  const beforeUpload = async (file: File) => {
    const isCSV = file.type === "text/csv" || file.name.endsWith(".csv");
    const isExcel =
      file.type ===
        "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet" ||
      file.type === "application/vnd.ms-excel" ||
      file.name.endsWith(".xlsx") ||
      file.name.endsWith(".xls");

    if (!isCSV && !isExcel) {
      message.error("Please upload a CSV or Excel file");
      return false;
    }

    if (file.size > 10 * 1024 * 1024) {
      // 10MB limit
      message.error("File size must be less than 10MB");
      return false;
    }

    // Call parent handler with the file
    onFileSelect(file);

    // Prevent automatic upload
    return false;
  };

  return (
    <div className="import-uploader">
      <Dragger
        accept=".csv,.xlsx,.xls"
        multiple={false}
        beforeUpload={beforeUpload}
        disabled={loading}
        showUploadList={false}
      >
        <p className="ant-upload-drag-icon">
          <InboxOutlined />
        </p>
        <p className="ant-upload-text">Click or drag CSV/Excel file here</p>
        <p className="ant-upload-hint">
          Supported formats: CSV, XLSX, XLS (max 10MB)
        </p>
        {loading && (
          <p style={{ color: token.colorPrimary }}>
            Uploading and parsing file...
          </p>
        )}
      </Dragger>
    </div>
  );
}
