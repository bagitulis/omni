import { useState, useEffect } from "react";
import {
  Modal,
  Steps,
  Button,
  Card,
  Typography,
  Table,
  Result,
  Alert,
  Spin,
  Row,
  Col,
  Flex,
} from "antd";
import {
  CloudDownloadOutlined,
  FileTextOutlined,
  ReloadOutlined,
} from "@ant-design/icons";
import {
  useShippingFiles,
  useProcessShippingFile,
  useExportShippingToSheets,
} from "@/hooks/useShipping";

const { Title, Text } = Typography;

interface DashboardShippingModalProps {
  open: boolean;
  onClose: () => void;
}

export function DashboardShippingModal({
  open,
  onClose,
}: DashboardShippingModalProps) {
  const [currentStep, setCurrentStep] = useState(0);
  const [selectedOption, setSelectedOption] = useState<
    "wallet" | "file" | null
  >(null);
  const [processedResult, setProcessedResult] = useState<{
    processed: number;
    errors: string[];
  } | null>(null);

  // Hooks
  const {
    data: filesData,
    isLoading: isLoadingFiles,
    refetch: refetchFiles,
  } = useShippingFiles();
  const { mutate: processFile, error: processError } = useProcessShippingFile();
  const { mutate: exportToSheets, error: exportError } =
    useExportShippingToSheets();

  // Reset state when modal opens
  useEffect(() => {
    if (open) {
      setCurrentStep(0);
      setSelectedOption(null);
      setProcessedResult(null);
    }
  }, [open]);

  const handleOptionSelect = (option: "wallet" | "file") => {
    setSelectedOption(option);
    if (option === "file") {
      setCurrentStep(1);
      refetchFiles();
    } else {
      // Wallet option: export shipping to sheets
      setCurrentStep(2);
      handleExportToSheets();
    }
  };

  const handleExportToSheets = () => {
    exportToSheets(
      {},
      {
        onSuccess: () => {
          setCurrentStep(3);
        },
        onError: () => {
          setCurrentStep(3);
        },
      },
    );
  };

  const handleProcessFile = (filename: string) => {
    setCurrentStep(2); // Process step
    processFile(filename, {
      onSuccess: (data) => {
        if (data.success && data.data) {
          setProcessedResult(data.data);
          setCurrentStep(3); // Result
        } else {
          // Handle logical error (success: false)
          setCurrentStep(3);
        }
      },
      onError: () => {
        setCurrentStep(3);
      },
    });
  };

  const handleClose = () => {
    onClose();
  };

  // Step 0: Select Option Content
  const renderOptionSelection = () => (
    <Row gutter={16} style={{ paddingTop: 32, paddingBottom: 32 }}>
      <Col span={12}>
        <Card
          hoverable
          onClick={() => handleOptionSelect("wallet")}
          style={{ textAlign: "center", cursor: "pointer" }}
        >
          <CloudDownloadOutlined
            style={{ fontSize: 36, color: "#0369a1", marginBottom: 16 }}
          />
          <Title level={4}>Get from Wallet</Title>
          <Text type="secondary">
            Export shipping fees directly from platform wallet to Google Sheets
          </Text>
        </Card>
      </Col>

      <Col span={12}>
        <Card
          hoverable
          onClick={() => handleOptionSelect("file")}
          style={{ textAlign: "center", cursor: "pointer" }}
        >
          <FileTextOutlined
            style={{ fontSize: 36, color: "#16a34a", marginBottom: 16 }}
          />
          <Title level={4}>Process Shipping File</Title>
          <Text type="secondary">
            Load and process downloaded shipping files from local storage
          </Text>
        </Card>
      </Col>
    </Row>
  );

  // Step 1: Load Files Content
  const renderFileList = () => {
    const columns = [
      {
        title: "Filename",
        dataIndex: "name",
        key: "name",
        render: (text: string) => <Text strong>{text}</Text>,
      },
      {
        title: "Action",
        key: "action",
        width: 120,
        render: (_: unknown, record: { name: string }) => (
          <Button
            type="primary"
            size="small"
            onClick={() => handleProcessFile(record.name)}
          >
            Process
          </Button>
        ),
      },
    ];

    // Transform string[] to object[] for Table
    const tableData =
      filesData?.data?.files?.map((file) => ({ key: file, name: file })) || [];

    return (
      <div style={{ paddingTop: 16, paddingBottom: 16 }}>
        <Flex
          justify="space-between"
          align="center"
          style={{ marginBottom: 16 }}
        >
          <Title level={5} style={{ margin: 0 }}>
            Available Files
          </Title>
          <Button
            icon={<ReloadOutlined />}
            onClick={() => refetchFiles()}
            loading={isLoadingFiles}
          >
            Refresh
          </Button>
        </Flex>

        {isLoadingFiles ? (
          <div
            style={{ textAlign: "center", paddingTop: 32, paddingBottom: 32 }}
          >
            <Spin tip="Loading files..." />
          </div>
        ) : (
          <Table
            dataSource={tableData}
            columns={columns}
            pagination={{ pageSize: 5 }}
            size="small"
            bordered
            locale={{ emptyText: "No shipping files found" }}
          />
        )}
      </div>
    );
  };

  // Step 2: Processing Content
  const renderProcessing = () => (
    <div style={{ textAlign: "center", paddingTop: 48, paddingBottom: 48 }}>
      <Spin size="large" />
      <div style={{ marginTop: 16 }}>
        <Title level={4}>Processing...</Title>
        <Text type="secondary">
          {selectedOption === "wallet"
            ? "Exporting shipping data to Google Sheets"
            : "Processing shipping file records"}
        </Text>
      </div>
    </div>
  );

  // Step 3: Result Content
  const renderResult = () => {
    const isError =
      (selectedOption === "wallet" && exportError) ||
      (selectedOption === "file" && (processError || !processedResult));

    // For file processing, check specific result data
    const successCount = processedResult?.processed || 0;
    const errorList = processedResult?.errors || [];
    const hasProcessingErrors = errorList.length > 0;

    if (isError) {
      return (
        <Result
          status="error"
          title="Operation Failed"
          subTitle={
            (processError as Error)?.message ||
            (exportError as Error)?.message ||
            "An unknown error occurred during processing."
          }
          extra={[
            <Button type="primary" key="close" onClick={handleClose}>
              Close
            </Button>,
            <Button key="retry" onClick={() => setCurrentStep(0)}>
              Try Again
            </Button>,
          ]}
        />
      );
    }

    if (selectedOption === "wallet") {
      return (
        <Result
          status="success"
          title="Export Successful"
          subTitle="Shipping fees have been successfully exported to Google Sheets."
          extra={[
            <Button type="primary" key="close" onClick={handleClose}>
              Done
            </Button>,
          ]}
        />
      );
    }

    // File processing result
    return (
      <Result
        status={hasProcessingErrors ? "warning" : "success"}
        title={
          hasProcessingErrors
            ? "Completed with Issues"
            : "Processing Successful"
        }
        subTitle={`Processed ${successCount} records.`}
        extra={[
          <Button type="primary" key="close" onClick={handleClose}>
            Done
          </Button>,
          <Button key="another" onClick={() => setCurrentStep(1)}>
            Process Another File
          </Button>,
        ]}
      >
        {hasProcessingErrors && (
          <div style={{ textAlign: "left", marginTop: 16 }}>
            <Alert
              message="Errors encountered"
              description={
                <ul
                  style={{
                    paddingLeft: 16,
                    marginTop: 8,
                    maxHeight: 160,
                    overflowY: "auto",
                  }}
                >
                  {errorList.map((err, idx) => (
                    <li
                      key={`${idx}-${err.substring(0, 10)}`}
                      style={{ fontSize: 12 }}
                    >
                      {err}
                    </li>
                  ))}
                </ul>
              }
              type="error"
              showIcon
            />
          </div>
        )}
      </Result>
    );
  };

  const steps = [
    { title: "Select Option" },
    { title: "Load Files" },
    { title: "Process" },
    { title: "Result" },
  ];

  return (
    <Modal
      title="Shipping Workflow"
      open={open}
      onCancel={handleClose}
      destroyOnHidden
      footer={null}
      width={700}
    >
      <Steps
        current={currentStep}
        items={steps}
        size="small"
        style={{ marginBottom: 24 }}
      />

      <div style={{ minHeight: 300 }}>
        {currentStep === 0 && renderOptionSelection()}
        {currentStep === 1 && renderFileList()}
        {currentStep === 2 && renderProcessing()}
        {currentStep === 3 && renderResult()}
      </div>
    </Modal>
  );
}
