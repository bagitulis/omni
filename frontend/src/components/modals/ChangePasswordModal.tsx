import { Modal, Form, Input, Button, message } from "antd";
import { useState } from "react";

interface ChangePasswordModalProps {
  open: boolean;
  onClose: () => void;
}

export function ChangePasswordModal({
  open,
  onClose,
}: ChangePasswordModalProps) {
  const [form] = Form.useForm();
  const [loading, setLoading] = useState(false);

  const handleSubmit = async () => {
    try {
      const values = await form.validateFields();
      if (values.newPassword !== values.confirmPassword) {
        form.setFields([
          {
            name: "confirmPassword",
            errors: ["Passwords do not match"],
          },
        ]);
        return;
      }

      setLoading(true);
      // Mock API call
      await new Promise((resolve) => setTimeout(resolve, 1000));
      message.success("Password changed successfully");
      form.resetFields();
      onClose();
    } catch (error) {
      // Validation error or API error
      console.error(error);
    } finally {
      setLoading(false);
    }
  };

  return (
    <Modal
      title="Change Password"
      open={open}
      onCancel={onClose}
      footer={null}
      width={400}
    >
      <Form form={form} layout="vertical" onFinish={handleSubmit}>
        <Form.Item
          name="currentPassword"
          label="Current Password"
          rules={[{ required: true, message: "Please enter current password" }]}
        >
          <Input.Password />
        </Form.Item>
        <Form.Item
          name="newPassword"
          label="New Password"
          rules={[{ required: true, message: "Please enter new password" }]}
        >
          <Input.Password />
        </Form.Item>
        <Form.Item
          name="confirmPassword"
          label="Confirm Password"
          dependencies={["newPassword"]}
          rules={[
            { required: true, message: "Please confirm new password" },
            ({ getFieldValue }) => ({
              validator(_, value) {
                if (!value || getFieldValue("newPassword") === value) {
                  return Promise.resolve();
                }
                return Promise.reject(new Error("Passwords do not match"));
              },
            }),
          ]}
        >
          <Input.Password />
        </Form.Item>
        <div style={{ display: "flex", justifyContent: "flex-end", gap: 8 }}>
          <Button onClick={onClose} disabled={loading}>
            Cancel
          </Button>
          <Button type="primary" htmlType="submit" loading={loading}>
            Change Password
          </Button>
        </div>
      </Form>
    </Modal>
  );
}
