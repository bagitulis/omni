import { useState } from "react";
import { Upload, message, Progress } from "antd";
import { InboxOutlined } from "@ant-design/icons";
import type { UploadFile, UploadChangeParam } from "antd/es/upload";
import type { ImportPreviewData } from "@/types/product";
import "./ImportUploader.css";

interface ImportUploaderProps {
  onPreview: (data: ImportPreviewData) => void;
  loading?: boolean;
}

const { Dragger } = Upload;

export function ImportUploader({
  onPreview,
  loading = false,
}: ImportUploaderProps) {
  const [fileList, setFileList] = useState<UploadFile[]>([]);
  const [uploading, setUploading] = useState(false);
  const [progress, setProgress] = useState(0);

  const parseCSV = (content: string): ImportPreviewData => {
    const lines = content
      .split("\n")
      .map((line) => line.trim())
      .filter((line) => line.length > 0);

    if (lines.length < 2) {
      throw new Error("CSV must contain at least 1 header row and 1 data row");
    }

    const headers = lines[0].split(",").map((h) => h.trim().toLowerCase());
    const requiredFields = ["item_name", "item_sku", "price", "stock"];
    const missingFields = requiredFields.filter(
      (field) => !headers.includes(field),
    );

    if (missingFields.length > 0) {
      throw new Error(`Missing required columns: ${missingFields.join(", ")}`);
    }

    const rows = lines.slice(1).map((line, idx) => {
      const values = line.split(",").map((v) => v.trim());
      const row = {
        item_name: values[headers.indexOf("item_name")] || "",
        item_sku: values[headers.indexOf("item_sku")] || "",
        price: parseFloat(values[headers.indexOf("price")] || "0"),
        stock: parseInt(values[headers.indexOf("stock")] || "0", 10),
      };

      const errors: string[] = [];

      // Validation
      if (!row.item_name || row.item_name.length === 0) {
        errors.push("Item name is required");
      }
      if (!row.item_sku || row.item_sku.length === 0) {
        errors.push("Item SKU is required");
      }
      if (isNaN(row.price) || row.price < 0) {
        errors.push("Price must be a positive number");
      }
      if (isNaN(row.stock) || row.stock < 0) {
        errors.push("Stock must be a non-negative integer");
      }

      return {
        row_number: idx + 2, // +2 because idx starts at 0 and row 1 is header
        item_name: row.item_name,
        item_sku: row.item_sku,
        price: row.price,
        stock: row.stock,
        valid: errors.length === 0,
        errors,
      };
    });

    const valid_rows = rows.filter((r) => r.valid).length;
    const invalid_rows = rows.filter((r) => !r.valid).length;

    return {
      total_rows: rows.length,
      valid_rows,
      invalid_rows,
      rows,
    };
  };

  const parseExcel = (content: string): ImportPreviewData => {
    // For now, treat Excel as CSV (mock implementation)
    // Real implementation would use xlsx library
    return parseCSV(content);
  };

  const handleChange = async (info: UploadChangeParam) => {
    const { file } = info;

    if (file.status === "done" || file.status === "removed") {
      if (file.status === "removed") {
        setFileList(fileList.filter((f) => f.uid !== file.uid));
        return;
      }
    }
  };

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

    if (file.size > 5 * 1024 * 1024) {
      // 5MB limit
      message.error("File size must be less than 5MB");
      return false;
    }

    setUploading(true);
    setProgress(0);

    try {
      // Read file with real progress tracking
      const reader = new FileReader();

      reader.onprogress = (event) => {
        if (event.lengthComputable) {
          const percent = Math.round((event.loaded / event.total) * 100);
          setProgress(percent > 90 ? 90 : percent); // Cap at 90, reader.onload sets 100
        }
      };

      reader.onload = (e) => {
        setProgress(100);

        try {
          const content = e.target?.result as string;
          const data = isCSV ? parseCSV(content) : parseExcel(content);

          if (data.total_rows === 0) {
            message.error("No data rows found in file");
            return;
          }

          message.success(
            `File parsed successfully: ${data.valid_rows} valid rows`,
          );
          onPreview(data);

          // Reset file list after successful parse
          setTimeout(() => {
            setFileList([]);
            setUploading(false);
            setProgress(0);
          }, 500);
        } catch (error) {
          message.error((error as Error).message || "Failed to parse file");
          setUploading(false);
          setProgress(0);
        }
      };

      reader.readAsText(file);
      return false; // Prevent automatic upload
    } catch (error) {
      message.error("Failed to read file");
      setUploading(false);
      setProgress(0);
      return false;
    }
  };

  return (
    <div className="import-uploader">
      <Dragger
        accept=".csv,.xlsx,.xls"
        multiple={false}
        fileList={fileList}
        onChange={handleChange}
        beforeUpload={beforeUpload}
        disabled={loading || uploading}
      >
        <p className="ant-upload-drag-icon">
          <InboxOutlined />
        </p>
        <p className="ant-upload-text">Click or drag CSV/Excel file here</p>
        <p className="ant-upload-hint">
          Supported formats: CSV, XLSX, XLS (max 5MB)
        </p>
        {uploading && <Progress percent={Math.floor(progress)} size="small" />}
      </Dragger>
    </div>
  );
}
