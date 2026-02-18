import { Button, Result, Alert } from "antd";

interface ShippingResultProps {
  selectedOption: "wallet" | "file" | null;
  exportError: Error | null;
  processError: Error | null;
  processedResult: { processed: number; errors: string[] } | null;
  onClose: () => void;
  onRetry: () => void;
  onProcessAnother: () => void;
}

export function ShippingResult({
  selectedOption,
  exportError,
  processError,
  processedResult,
  onClose,
  onRetry,
  onProcessAnother,
}: ShippingResultProps) {
  const isError =
    (selectedOption === "wallet" && exportError) ||
    (selectedOption === "file" && (processError || !processedResult));

  const successCount = processedResult?.processed || 0;
  const errorList = processedResult?.errors || [];
  const hasProcessingErrors = errorList.length > 0;

  if (isError) {
    return (
      <Result
        status="error"
        title="Operation Failed"
        subTitle={
          processError?.message ||
          exportError?.message ||
          "An unknown error occurred during processing."
        }
        extra={[
          <Button type="primary" key="close" onClick={onClose}>
            Close
          </Button>,
          <Button key="retry" onClick={onRetry}>
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
          <Button type="primary" key="close" onClick={onClose}>
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
        hasProcessingErrors ? "Completed with Issues" : "Processing Successful"
      }
      subTitle={`Processed ${successCount} records.`}
      extra={[
        <Button type="primary" key="close" onClick={onClose}>
          Done
        </Button>,
        <Button key="another" onClick={onProcessAnother}>
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
}
