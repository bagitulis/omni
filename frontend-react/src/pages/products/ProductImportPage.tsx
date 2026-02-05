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
} from "antd";
import { ArrowLeftOutlined, CheckCircleOutlined } from "@ant-design/icons";
import { useNavigate } from "react-router-dom";
import type { ImportPreviewData } from "@/types/product";
import { ImportUploader } from "@/components/forms/ImportUploader";
import { ImportPreviewTable } from "@/components/tables/ImportPreviewTable";
import "./ProductImportPage.css";

export default function ProductImportPage() {
  const navigate = useNavigate();
  const [currentStep, setCurrentStep] = useState(0);
  const [previewData, setPreviewData] = useState<ImportPreviewData | null>(
    null,
  );
  const [importing, setImporting] = useState(false);

  const handlePreview = (data: ImportPreviewData) => {
    setPreviewData(data);
    setCurrentStep(1);
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
          <p style={{ color: "#666", fontSize: "12px" }}>
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

    setImporting(true);

    try {
      // Mock import - simulate API call with valid rows only
      const validRows = previewData.rows.filter((row) => row.valid);

      // Simulate API delay
      await new Promise((resolve) => setTimeout(resolve, 1500));

      // Mock success response
      const importedCount = validRows.length;

      message.success(`Successfully imported ${importedCount} products`);
      setCurrentStep(2);

      // Auto-redirect after 2 seconds
      setTimeout(() => {
        navigate("/master-products");
      }, 2000);
    } catch (error) {
      message.error((error as Error).message || "Import failed");
    } finally {
      setImporting(false);
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
            <Alert
              message="Prepare your CSV or Excel file"
              description="Your file must contain columns: item_name, item_sku, price, stock"
              type="info"
              showIcon
              style={{ marginBottom: 24 }}
            />
            <ImportUploader onPreview={handlePreview} loading={importing} />
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
                    valueStyle={{ color: "#0369a1" }}
                  />
                </Col>
                <Col span={6}>
                  <Statistic
                    title="Valid Rows"
                    value={previewData?.valid_rows || 0}
                    valueStyle={{ color: "#16a34a" }}
                    prefix={<CheckCircleOutlined />}
                  />
                </Col>
                <Col span={6}>
                  <Statistic
                    title="Invalid Rows"
                    value={previewData?.invalid_rows || 0}
                    valueStyle={{
                      color: previewData?.invalid_rows ? "#dc2626" : "#0369a1",
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
              <p style={{ fontSize: "12px", color: "#666" }}>
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
          onClick={() => navigate("/master-products")}
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
                loading={importing}
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
