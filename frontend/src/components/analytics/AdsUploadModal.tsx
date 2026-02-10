import { Modal, Form, Select, Upload, Button, message, Typography } from "antd";
import { InboxOutlined } from "@ant-design/icons";
import { useState } from "react";
import type { UploadFile } from "antd/es/upload/interface";
import api from "../../api/client";

const { Text } = Typography;
const { Dragger } = Upload;

interface Props {
  open: boolean;
  onClose: () => void;
  onSuccess: () => void;
}

export const AdsUploadModal = ({ open, onClose, onSuccess }: Props) => {
  const [form] = Form.useForm();
  const [fileList, setFileList] = useState<UploadFile[]>([]);
  const [uploading, setUploading] = useState(false);

  const handleUpload = async () => {
    try {
      const values = await form.validateFields();
      if (fileList.length === 0) {
        message.error("Please select a file to upload");
        return;
      }

      setUploading(true);
      const formData = new FormData();
      formData.append("file", fileList[0] as any);
      formData.append("platform", values.platform);

      // Determine endpoint based on platform
      const endpoint =
        values.platform === "tiktok"
          ? "/analytics/tiktok/upload"
          : "/analytics/shopee/upload";

      const response = await api.post(endpoint, formData, {
        headers: {
          "Content-Type": "multipart/form-data",
        },
      });

      if (response.data.success) {
        message.success("File uploaded successfully");
        form.resetFields();
        setFileList([]);
        onSuccess();
        onClose();
      } else {
        message.error(response.data.error || "Upload failed");
      }
    } catch (error: any) {
      console.error("Upload error:", error);
      message.error(error.message || "Failed to upload file");
    } finally {
      setUploading(false);
    }
  };

  const props = {
    onRemove: (file: UploadFile) => {
      const index = fileList.indexOf(file);
      const newFileList = fileList.slice();
      newFileList.splice(index, 1);
      setFileList(newFileList);
    },
    beforeUpload: (file: UploadFile) => {
      // Check file type (xlsx or csv)
      const isExcel =
        file.type ===
          "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet" ||
        file.type === "text/csv" ||
        file.name.endsWith(".xlsx") ||
        file.name.endsWith(".csv");

      if (!isExcel) {
        message.error("You can only upload Excel or CSV files!");
        return Upload.LIST_IGNORE;
      }

      setFileList([file]); // Only keep latest file
      return false; // Prevent automatic upload
    },
    fileList,
  };

  return (
    <Modal
      title="Upload Ads Data"
      open={open}
      onCancel={onClose}
      footer={[
        <Button key="cancel" onClick={onClose}>
          Cancel
        </Button>,
        <Button
          key="upload"
          type="primary"
          onClick={handleUpload}
          loading={uploading}
          disabled={fileList.length === 0}
        >
          {uploading ? "Uploading..." : "Start Upload"}
        </Button>,
      ]}
    >
      <Form
        form={form}
        layout="vertical"
        initialValues={{ platform: "tiktok" }}
      >
        <Form.Item
          name="platform"
          label="Platform"
          rules={[{ required: true, message: "Please select a platform" }]}
        >
          <Select>
            <Select.Option value="tiktok">TikTok Ads</Select.Option>
            <Select.Option value="shopee">Shopee Ads</Select.Option>
          </Select>
        </Form.Item>

        <Form.Item label="Data File">
          <Dragger {...props} maxCount={1}>
            <p className="ant-upload-drag-icon">
              <InboxOutlined />
            </p>
            <p className="ant-upload-text">
              Click or drag file to this area to upload
            </p>
            <p className="ant-upload-hint">
              Support for single upload. Strictly prohibit from uploading
              company data or other band files
            </p>
          </Dragger>
        </Form.Item>

        <div style={{ marginTop: 16 }}>
          <Text type="secondary" style={{ fontSize: 12 }}>
            Supported formats: .xlsx, .csv
          </Text>
        </div>
      </Form>
    </Modal>
  );
};
