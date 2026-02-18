import { useState, useEffect } from "react";
import { Modal, Steps } from "antd";
import {
  useShippingFiles,
  useProcessShippingFile,
  useExportShippingToSheets,
} from "@/hooks/useShipping";
import { ShippingOptionSelect } from "./ShippingOptionSelect";
import { ShippingFileList } from "./ShippingFileList";
import { ShippingProcessing } from "./ShippingProcessing";
import { ShippingResult } from "./ShippingResult";

interface DashboardShippingModalProps {
  open: boolean;
  onClose: () => void;
}

const STEPS = [
  { title: "Select Option" },
  { title: "Load Files" },
  { title: "Process" },
  { title: "Result" },
];

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

  const {
    data: filesData,
    isLoading: isLoadingFiles,
    refetch: refetchFiles,
  } = useShippingFiles();
  const { mutate: processFile, error: processError } = useProcessShippingFile();
  const { mutate: exportToSheets, error: exportError } =
    useExportShippingToSheets();

  useEffect(() => {
    if (open) {
      setCurrentStep(0);
      setSelectedOption(null);
      setProcessedResult(null);
    }
  }, [open]);

  const handleExportToSheets = () => {
    exportToSheets(
      {},
      {
        onSuccess: () => setCurrentStep(3),
        onError: () => setCurrentStep(3),
      },
    );
  };

  const handleOptionSelect = (option: "wallet" | "file") => {
    setSelectedOption(option);
    if (option === "file") {
      setCurrentStep(1);
      void refetchFiles();
    } else {
      setCurrentStep(2);
      handleExportToSheets();
    }
  };

  const handleProcessFile = (filename: string) => {
    setCurrentStep(2);
    processFile(filename, {
      onSuccess: (data) => {
        if (data.success && data.data) {
          setProcessedResult(data.data);
          setCurrentStep(3);
        } else {
          setCurrentStep(3);
        }
      },
      onError: () => setCurrentStep(3),
    });
  };

  return (
    <Modal
      title="Shipping Workflow"
      open={open}
      onCancel={onClose}
      destroyOnHidden
      footer={null}
      width={700}
    >
      <Steps
        current={currentStep}
        items={STEPS}
        size="small"
        style={{ marginBottom: 24 }}
      />

      <div style={{ minHeight: 300 }}>
        {currentStep === 0 && (
          <ShippingOptionSelect onSelect={handleOptionSelect} />
        )}
        {currentStep === 1 && (
          <ShippingFileList
            filesData={filesData}
            isLoadingFiles={isLoadingFiles}
            onRefetch={() => {
              void refetchFiles();
            }}
            onProcessFile={handleProcessFile}
          />
        )}
        {currentStep === 2 && (
          <ShippingProcessing selectedOption={selectedOption} />
        )}
        {currentStep === 3 && (
          <ShippingResult
            selectedOption={selectedOption}
            exportError={exportError}
            processError={processError}
            processedResult={processedResult}
            onClose={onClose}
            onRetry={() => setCurrentStep(0)}
            onProcessAnother={() => setCurrentStep(1)}
          />
        )}
      </div>
    </Modal>
  );
}
