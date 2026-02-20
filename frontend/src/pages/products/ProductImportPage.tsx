import { useState } from "react";
import {
  Button,
  Card,
  Steps,
  Alert,
  Space,
  Statistic,
  Row,
  Col,
  message,
  Modal,
  theme,
} from "antd";
import {
  ArrowLeftOutlined,
  CheckCircleOutlined,
  DownloadOutlined,
} from "@ant-design/icons";
import { useNavigate } from "react-router-dom";
import type { ImportPreviewData } from "@/types/product";
import { ImportUploader } from "@/components/forms/ImportUploader";
import { ImportPreviewTable } from "@/components/tables/ImportPreviewTable";
import {
  useImportPreview,
  useImportProducts,
  useAutoMapSkus,
} from "@/hooks/useProductImport";
import { downloadImportTemplate } from "@/api/products";
import "./ProductImportPage.css";

export default function ProductImportPage() {
  const { token } = theme.useToken();
  const navigate = useNavigate();
  const [currentStep, setCurrentStep] = useState(0);
  const [previewData, setPreviewData] = useState<ImportPreviewData | null>(
    null,
  );

  const previewMutation = useImportPreview();
  const importMutation = useImportProducts();
  const autoMapMutation = useAutoMapSkus();

  const handleFileUpload = async (file: File) => {
    try {
      const data = await previewMutation.mutateAsync(file);
      setPreviewData(data);
      setCurrentStep(1);
    } catch {
      // Error message already shown by mutation's onError handler
    }
  };

  const handleAutoMap = async () => {
    if (!previewData) return;

    const skus = previewData.rows
      .filter((row) => row.valid)
      .map((row) => row.item_sku);

    if (skus.length === 0) {
      message.warning("No valid SKUs to auto-map");
      return;
    }

    try {
      await autoMapMutation.mutateAsync(skus);
    } catch {
      // Error message already shown by mutation's onError handler
    }
  };

  const handleConfirmImport = () => {
    if (!previewData || previewData.valid_rows === 0) {
      message.error("No valid rows to import");
      return;
    }

    Modal.confirm({
      title: "Confirm Import",
      content: (
        <div>
          <p>
            Are you sure you want to import {previewData.valid_rows} products?
          </p>
          <p style={{ color: token.colorTextSecondary, fontSize: "12px" }}>
            {previewData.invalid_rows > 0 && (
              <>
                Note: {previewData.invalid_rows} row(s) with errors will be
                skipped.
              </>
            )}
          </p>
        </div>
      ),
      okText: "Import",
      okType: "primary",
      onOk() {
        performImport();
      },
    });
  };

  const performImport = async () => {
    if (!previewData) return;

    try {
      await importMutation.mutateAsync(previewData.rows);
      setCurrentStep(2);

      // Auto-redirect after 2 seconds
      setTimeout(() => {
        navigate("/products");
      }, 2000);
    } catch {
      // Error message already shown by mutation's onError handler
    }
  };

  const handleReset = () => {
    setPreviewData(null);
    setCurrentStep(0);
  };

  const renderStepContent = () => {
    switch (currentStep) {
      case 0:
        return (
          <div className="step-content">
            <div style={{ marginBottom: 16, textAlign: "right" }}>
              <Button
                icon={<DownloadOutlined />}
                onClick={() => downloadImportTemplate("xlsx")}
              >
                Download Template
              </Button>
            </div>
            <Alert
              message="Prepare your CSV or Excel file"
              description="Your file must contain columns: item_name, item_sku, price, stock"
              type="info"
              showIcon
              style={{ marginBottom: 24 }}
            />
            <ImportUploader
              onFileSelect={handleFileUpload}
              loading={previewMutation.isPending}
            />
          </div>
        );

      case 1:
        return (
          <div className="step-content">
            <Card className="stats-card">
              <Row gutter={16}>
                <Col span={6}>
                  <Statistic
                    title="Total Rows"
                    value={previewData?.total_rows || 0}
                    valueStyle={{ color: token.colorPrimary }}
                  />
                </Col>
                <Col span={6}>
                  <Statistic
                    title="Valid Rows"
                    value={previewData?.valid_rows || 0}
                    valueStyle={{ color: token.colorSuccess }}
                    prefix={<CheckCircleOutlined />}
                  />
                </Col>
                <Col span={6}>
                  <Statistic
                    title="Invalid Rows"
                    value={previewData?.invalid_rows || 0}
                    valueStyle={{
                      color: previewData?.invalid_rows
                        ? token.colorError
                        : token.colorPrimary,
                    }}
                  />
                </Col>
              </Row>
            </Card>

            {previewData && previewData.invalid_rows > 0 && (
              <Alert
                message="Some rows have validation errors"
                description="Only valid rows will be imported. Review errors in the table below."
                type="warning"
                showIcon
                style={{ marginTop: 16, marginBottom: 16 }}
              />
            )}

            {/* Auto-Map SKUs button */}
            {previewData && previewData.valid_rows > 0 && (
              <div style={{ marginTop: 16, marginBottom: 16 }}>
                <Button
                  type="default"
                  onClick={handleAutoMap}
                  loading={autoMapMutation.isPending}
                  disabled={importMutation.isPending}
                >
                  Auto-Map SKUs to Platform Products
                </Button>
                <span
                  style={{
                    marginLeft: 12,
                    fontSize: "12px",
                    color: token.colorTextSecondary,
                  }}
                >
                  Automatically link imported SKUs to existing platform products
                </span>
              </div>
            )}

            {previewData && <ImportPreviewTable data={previewData.rows} />}
          </div>
        );

      case 2:
        return (
          <div className="step-content success-content">
            <Card className="success-card">
              <div className="success-icon">✓</div>
              <h2>Import Complete!</h2>
              <p>Your products have been successfully imported.</p>
              <p style={{ fontSize: "12px", color: token.colorTextSecondary }}>
                Redirecting to products page in 2 seconds...
              </p>
            </Card>
          </div>
        );

      default:
        return null;
    }
  };

  return (
    <div className="product-import-page">
      <div className="page-header">
        <Button
          type="text"
          icon={<ArrowLeftOutlined />}
          onClick={() => navigate("/products")}
          className="back-button"
        >
          Back to Products
        </Button>
        <h1>Import Products</h1>
        <p className="subtitle">Bulk import products via CSV or Excel file</p>
      </div>

      <Card className="import-card">
        <Steps
          current={currentStep}
          items={[
            {
              title: "Upload File",
              description: "Choose CSV or Excel",
            },
            {
              title: "Review & Validate",
              description: "Check your data",
            },
            {
              title: "Complete",
              description: "Import finished",
            },
          ]}
          style={{ marginBottom: 32 }}
        />

        {renderStepContent()}

        <div className="step-actions">
          {currentStep === 1 && (
            <Space>
              <Button onClick={handleReset}>Reset</Button>
              <Button
                type="primary"
                onClick={handleConfirmImport}
                loading={importMutation.isPending}
                disabled={!previewData || previewData.valid_rows === 0}
              >
                Confirm & Import
              </Button>
            </Space>
          )}
        </div>
      </Card>
    </div>
  );
}
