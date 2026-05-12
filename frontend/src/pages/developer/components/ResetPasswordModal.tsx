import { useEffect, useState } from "react";
import { App, Form, Input, Modal, Typography } from "antd";
import { developerApi, type CrossTenantUser } from "@/api/developer";

const { Text } = Typography;

interface ResetPasswordModalProps {
  open: boolean;
  user: CrossTenantUser | null;
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

export function ResetPasswordModal({
  open,
  user,
  onClose,
  onSuccess,
}: ResetPasswordModalProps) {
  const { message } = App.useApp();
  const [password, setPassword] = useState("");
  const [confirmPassword, setConfirmPassword] = useState("");
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (!open) {
      setPassword("");
      setConfirmPassword("");
      setError(null);
    }
  }, [open]);

  const handleOk = async () => {
    const validationError = validatePassword(password);
    if (validationError) {
      setError(validationError);
      return;
    }
    if (password !== confirmPassword) {
      setError("Passwords do not match");
      return;
    }
    if (!user) return;

    setLoading(true);
    try {
      const response = await developerApi.resetUserPassword(user.id, password);
      if (response.success) {
        message.success(`Password reset for ${user.username}`);
        onSuccess();
        onClose();
      } else {
        message.error(response.error ?? "Failed to reset password");
      }
    } catch {
      message.error("Failed to reset password");
    } finally {
      setLoading(false);
    }
  };

  const passwordError = password.length > 0 ? validatePassword(password) : null;
  const mismatchError =
    confirmPassword.length > 0 && password !== confirmPassword
      ? "Passwords do not match"
      : null;

  return (
    <Modal
      title="Reset User Password"
      open={open}
      onCancel={onClose}
      onOk={handleOk}
      okText="Reset Password"
      okButtonProps={{
        danger: true,
        loading,
        disabled: !password || !confirmPassword || !!passwordError || !!mismatchError,
      }}
      destroyOnClose
      width={440}
    >
      {user && (
        <div style={{ marginBottom: 16 }}>
          <div style={{ display: "flex", gap: 16, marginBottom: 8 }}>
            <Text type="secondary">Username:</Text>
            <Text strong>{user.username}</Text>
          </div>
          <div style={{ display: "flex", gap: 16, marginBottom: 8 }}>
            <Text type="secondary">Email:</Text>
            <Text>{user.email}</Text>
          </div>
          <div style={{ display: "flex", gap: 16 }}>
            <Text type="secondary">Tenant:</Text>
            <Text>{user.tenant_name}</Text>
          </div>
        </div>
      )}

      <Form layout="vertical">
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
        <Form.Item
          label="Confirm Password"
          validateStatus={mismatchError ? "error" : undefined}
          help={mismatchError}
        >
          <Input.Password
            value={confirmPassword}
            onChange={(e) => {
              setConfirmPassword(e.target.value);
              setError(null);
            }}
            placeholder="Re-enter password"
          />
        </Form.Item>
      </Form>

      <Text type="secondary" style={{ fontSize: 12 }}>
        Requirements: min 8 characters, uppercase, lowercase, and a number.
      </Text>
    </Modal>
  );
}
