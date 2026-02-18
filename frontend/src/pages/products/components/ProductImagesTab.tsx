import { useState } from "react";
import { Upload, Button, Flex, message } from "antd";
import { PlusOutlined, SaveOutlined, PictureOutlined } from "@ant-design/icons";
import type { UploadFile } from "antd/es/upload/interface";
import { ImageGalleryPicker } from "@/components/shared/ImageGalleryPicker";
import type { GalleryImage } from "@/types/shared";
import { getImageUrl } from "@/types/shared";

const MAX_IMAGES = 8;

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
  const [galleryOpen, setGalleryOpen] = useState(false);

  const handleGalleryConfirm = (images: GalleryImage[]) => {
    const existingUrls = new Set(fileList.map((f) => f.url).filter(Boolean));

    const newFiles: UploadFile[] = images
      .map((img) => {
        const url = getImageUrl(img.local_path, "original");
        return {
          url,
          uid: `gallery-${img.id}`,
          name: img.filename,
          status: "done" as const,
        };
      })
      .filter((f) => !existingUrls.has(f.url));

    const merged = [...fileList, ...newFiles].slice(0, MAX_IMAGES);

    if (fileList.length + newFiles.length > MAX_IMAGES) {
      message.warning(
        `Maximum ${MAX_IMAGES} images allowed. Some selections were trimmed.`,
      );
    }

    setFileList(merged);
  };

  return (
    <div>
      <Upload
        listType="picture-card"
        fileList={fileList}
        onChange={({ fileList: newFileList }) => setFileList(newFileList)}
        maxCount={MAX_IMAGES}
        beforeUpload={() => false}
      >
        {fileList.length < MAX_IMAGES && (
          <div>
            <PlusOutlined />
            <div style={{ marginTop: 8 }}>Upload</div>
          </div>
        )}
      </Upload>
      <Flex justify="space-between" align="center" style={{ marginTop: 16 }}>
        <Button
          icon={<PictureOutlined />}
          onClick={() => setGalleryOpen(true)}
          disabled={fileList.length >= MAX_IMAGES}
        >
          Browse Gallery
        </Button>
        <Button
          type="primary"
          icon={<SaveOutlined />}
          onClick={() => onSave(fileList)}
          loading={loading}
        >
          Save Images
        </Button>
      </Flex>
      <ImageGalleryPicker
        open={galleryOpen}
        onClose={() => setGalleryOpen(false)}
        onConfirm={handleGalleryConfirm}
        maxSelect={MAX_IMAGES - fileList.length}
      />
    </div>
  );
};
