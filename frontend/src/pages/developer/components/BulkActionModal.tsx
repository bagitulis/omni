import { useState } from "react";
import { App, Form, Input, Modal, Typography } from "antd";
import { TypeToConfirmModal } from "@/components/modals/TypeToConfirmModal";
import { developerApi, type BulkOperationResult } from "@/api/developer";

const { Text } = Typography;

type BulkAction = "reset-passwords" | "disable";

interface BulkActionModalProps {
  open: boolean;
  action: BulkAction;
  users: {user_id: string; tenant_id: string}[];
  onClose: () => void;
  onSuccess: () => void;
}

function validatePassword(password: string): string | null {
  if (password.length < 8) return "Password must be at least 8 characters";
  if (!/[A-Z]/.test(password)) return "Must contain an uppercase letter";
  if (!/[a-z]/.test(password)) return "Must contain a lowercase letter";
  if (!/[0-9]/.test(password)) return "Must contain a number";
  return null;
}

export function BulkActionModal({
  open,
  action,
  users,
  onClose,
  onSuccess,
}: BulkActionModalProps) {
  const { message } = App.useApp();
  const [password, setPassword] = useState("");
  const [loading, setLoading] = useState(false);
  const [result, setResult] = useState<BulkOperationResult | null>(null);
  const [error, setError] = useState<string | null>(null);

  const count = users.length;

  const handleResetPasswords = async () => {
    const validationError = validatePassword(password);
    if (validationError) {
      setError(validationError);
      return;
    }

    setLoading(true);
    try {
      const response = await developerApi.bulkResetPasswords(users, password);
      if (response.success && response.data) {
        setResult(response.data);
        message.success(`Reset complete: ${response.data.success_count} succeeded`);
      } else {
        message.error(response.error ?? "Bulk reset failed");
      }
    } catch {
      message.error("Bulk reset failed");
    } finally {
      setLoading(false);
    }
  };

  const handleDisableUsers = async () => {
    setLoading(true);
    try {
      const response = await developerApi.bulkDisableUsers(users);
      if (response.success && response.data) {
        setResult(response.data);
        message.success(`Disable complete: ${response.data.success_count} succeeded`);
      } else {
        message.error(response.error ?? "Bulk disable failed");
      }
    } catch {
      message.error("Bulk disable failed");
    } finally {
      setLoading(false);
    }
  };

  const handleClose = () => {
    setPassword("");
    setError(null);
    setResult(null);
    onClose();
    if (result) onSuccess();
  };

  // Show result summary
  if (result) {
    return (
      <Modal
        title="Bulk Operation Result"
        open={open}
        onCancel={handleClose}
        onOk={handleClose}
        okText="Done"
        cancelButtonProps={{ style: { display: "none" } }}
        width={440}
      >
        <div style={{ marginTop: 8 }}>
          <div style={{ marginBottom: 8 }}>
            <Text>
              <Text strong style={{ color: "#16a34a" }}>
                {result.success_count}
              </Text>{" "}
              succeeded
            </Text>
          </div>
          {result.failure_count > 0 && (
            <div style={{ marginBottom: 8 }}>
              <Text>
                <Text strong style={{ color: "#dc2626" }}>
                  {result.failure_count}
                </Text>{" "}
                failed
              </Text>
              {result.errors.length > 0 && (
                <ul style={{ marginTop: 8, paddingLeft: 20 }}>
                  {result.errors.slice(0, 5).map((err) => (
                    <li key={err.user_id}>
                      <Text type="secondary" style={{ fontSize: 12 }}>
                        {err.user_id}: {err.error}
                      </Text>
                    </li>
                  ))}
                  {result.errors.length > 5 && (
                    <li>
                      <Text type="secondary" style={{ fontSize: 12 }}>
                        ...and {result.errors.length - 5} more
                      </Text>
                    </li>
                  )}
                </ul>
              )}
            </div>
          )}
        </div>
      </Modal>
    );
  }

  // Disable action uses TypeToConfirmModal
  if (action === "disable") {
    return (
      <TypeToConfirmModal
        open={open}
        title={`Disable ${count} User${count > 1 ? "s" : ""}`}
        description={`This will disable ${count} user${count > 1 ? "s" : ""}. They will no longer be able to log in.`}
        confirmText={`DISABLE ${count} USERS`}
        onConfirm={handleDisableUsers}
        onCancel={handleClose}
        danger
        loading={loading}
      />
    );
  }

  // Reset passwords action
  const passwordError = password.length > 0 ? validatePassword(password) : null;

  return (
    <Modal
      title={`Reset Password for ${count} User${count > 1 ? "s" : ""}`}
      open={open}
      onCancel={handleClose}
      onOk={handleResetPasswords}
      okText="Reset All Passwords"
      okButtonProps={{
        danger: true,
        loading,
        disabled: !password || !!passwordError,
      }}
      destroyOnClose
      width={440}
    >
      <div style={{ marginTop: 8 }}>
        <Text type="secondary">
          Set a new password for {count} selected user{count > 1 ? "s" : ""}.
        </Text>

        <Form layout="vertical" style={{ marginTop: 16 }}>
          <Form.Item
            label="New Password"
            validateStatus={passwordError || error ? "error" : undefined}
            help={passwordError || error}
          >
            <Input.Password
              value={password}
              onChange={(e) => {
                setPassword(e.target.value);
                setError(null);
              }}
              placeholder="Min 8 chars, upper + lower + number"
            />
          </Form.Item>
        </Form>

        <Text type="secondary" style={{ fontSize: 12 }}>
          Requirements: min 8 characters, uppercase, lowercase, and a number.
        </Text>
      </div>
    </Modal>
  );
}
