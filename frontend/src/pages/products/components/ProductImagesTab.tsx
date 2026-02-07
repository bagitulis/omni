import { useState } from "react";
import { Upload, Button, Flex } from "antd";
import { PlusOutlined, SaveOutlined } from "@ant-design/icons";
import type { UploadFile } from "antd/es/upload/interface";

interface ProductImagesTabProps {
  initialValues: UploadFile[];
  onSave: (files: UploadFile[]) => void;
  loading: boolean;
}

export const ProductImagesTab = ({
  initialValues,
  onSave,
  loading,
}: ProductImagesTabProps) => {
  const [fileList, setFileList] = useState<UploadFile[]>(initialValues);

  return (
    <div>
      <Upload
        listType="picture-card"
        fileList={fileList}
        onChange={({ fileList: newFileList }) => setFileList(newFileList)}
        maxCount={8}
        beforeUpload={() => false}
      >
        {fileList.length < 8 && (
          <div>
            <PlusOutlined />
            <div style={{ marginTop: 8 }}>Upload</div>
          </div>
        )}
      </Upload>
      <Flex justify="flex-end" style={{ marginTop: 16 }}>
        <Button
          type="primary"
          icon={<SaveOutlined />}
          onClick={() => onSave(fileList)}
          loading={loading}
        >
          Save Images
        </Button>
      </Flex>
    </div>
  );
};
