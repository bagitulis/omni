import { useCallback, useEffect, useState } from "react";
import {
  Card,
  Form,
  Input,
  Button,
  Avatar,
  Typography,
  Divider,
  theme,
  Spin,
} from "antd";
import { UserOutlined, LockOutlined, LoadingOutlined } from "@ant-design/icons";
import apiClient from "@/api/client";
import { updateProfile, changePassword } from "@/api/settings";
import { STORAGE_KEYS } from "@/lib/constants";
import { message } from "@/components/AntStaticHolder";

const { Text, Title } = Typography;
const { useToken } = theme;

interface ProfileForm {
  name: string;
  email: string;
  phone?: string;
}

interface PasswordForm {
  current_password: string;
  new_password: string;
  confirm_password: string;
}

interface UserData {
  id: string;
  username: string;
  email: string;
  role: string;
  phone?: string;
}

export default function AccountTab() {
  const { token } = useToken();
  const [profileForm] = Form.useForm<ProfileForm>();
  const [passwordForm] = Form.useForm<PasswordForm>();
  const [user, setUser] = useState<UserData | null>(null);
  const [loading, setLoading] = useState(true);

  const fetchUserProfile = useCallback(async () => {
    try {
      setLoading(true);
      const response = await apiClient.get<UserData>("/auth/me");
      if (response.success && response.data) {
        setUser(response.data);
        profileForm.setFieldsValue({
          name: response.data.username,
          email: response.data.email,
          phone: response.data.phone,
        });
      }
    } catch {
      // Fallback to sessionStorage
      const storedUser = sessionStorage.getItem(STORAGE_KEYS.AUTH_USER);
      if (storedUser) {
        try {
          const parsed = JSON.parse(storedUser);
          setUser(parsed);
          profileForm.setFieldsValue({
            name: parsed.username,
            email: parsed.email,
          });
        } catch {
          // Ignore parse error
        }
      }
    } finally {
      setLoading(false);
    }
  }, [profileForm]);

  useEffect(() => {
    fetchUserProfile();
  }, [fetchUserProfile]);

  const handleProfileSave = async (values: ProfileForm) => {
    try {
      await updateProfile({
        name: values.name,
        email: values.email,
        phone: values.phone,
      });
      message.success("Profile updated successfully");
    } catch (error) {
      message.error((error as Error).message || "Failed to update profile");
    }
  };

  const handlePasswordChange = async (values: PasswordForm) => {
    try {
      await changePassword({
        current_password: values.current_password,
        new_password: values.new_password,
        confirm_password: values.confirm_password,
      });
      message.success("Password changed successfully");
      passwordForm.resetFields();
    } catch (error) {
      message.error((error as Error).message || "Failed to change password");
    }
  };

  if (loading) {
    return (
      <div
        style={{
          display: "flex",
          justifyContent: "center",
          alignItems: "center",
          minHeight: 200,
        }}
      >
        <Spin indicator={<LoadingOutlined spin />} size="large" />
      </div>
    );
  }

  return (
    <div>
      <div style={{ marginBottom: 16 }}>
        <Title level={5} style={{ margin: 0, fontSize: 14 }}>
          Account Settings
        </Title>
        <Text type="secondary" style={{ fontSize: 12 }}>
          Manage your account profile and security
        </Text>
      </div>

      <Card
        title="Profile Information"
        size="small"
        style={{ marginBottom: 16, borderRadius: token.borderRadius }}
      >
        <div
          style={{
            display: "flex",
            alignItems: "center",
            gap: 16,
            marginBottom: 24,
          }}
        >
          <Avatar
            size={64}
            icon={<UserOutlined />}
            style={{ backgroundColor: token.colorPrimary }}
          />
          <div>
            <Text strong style={{ fontSize: 14, display: "block" }}>
              {user?.username || "User"}
            </Text>
            <Text type="secondary" style={{ fontSize: 12 }}>
              {user?.email || ""}
            </Text>
          </div>
        </div>

        <Form
          form={profileForm}
          layout="vertical"
          onFinish={handleProfileSave}
          style={{ maxWidth: 400 }}
        >
          <Form.Item
            label="Full Name"
            name="name"
            rules={[{ required: true, message: "Name is required" }]}
          >
            <Input prefix={<UserOutlined />} />
          </Form.Item>
          <Form.Item
            label="Email"
            name="email"
            rules={[
              { required: true, message: "Email is required" },
              { type: "email", message: "Invalid email format" },
            ]}
          >
            <Input disabled />
          </Form.Item>
          <Form.Item label="Phone" name="phone">
            <Input placeholder="+62 xxx xxxx xxxx" />
          </Form.Item>
          <Form.Item style={{ marginBottom: 0 }}>
            <Button type="primary" htmlType="submit">
              Update Profile
            </Button>
          </Form.Item>
        </Form>
      </Card>

      <Card
        title="Change Password"
        size="small"
        style={{ borderRadius: token.borderRadius }}
      >
        <Form
          form={passwordForm}
          layout="vertical"
          onFinish={handlePasswordChange}
          style={{ maxWidth: 400 }}
        >
          <Form.Item
            label="Current Password"
            name="current_password"
            rules={[
              { required: true, message: "Current password is required" },
            ]}
          >
            <Input.Password prefix={<LockOutlined />} />
          </Form.Item>
          <Divider style={{ margin: "16px 0" }} />
          <Form.Item
            label="New Password"
            name="new_password"
            rules={[
              { required: true, message: "New password is required" },
              { min: 8, message: "Password must be at least 8 characters" },
            ]}
          >
            <Input.Password prefix={<LockOutlined />} />
          </Form.Item>
          <Form.Item
            label="Confirm New Password"
            name="confirm_password"
            rules={[
              { required: true, message: "Please confirm your password" },
              ({ getFieldValue }) => ({
                validator(_, value) {
                  if (!value || getFieldValue("new_password") === value) {
                    return Promise.resolve();
                  }
                  return Promise.reject(new Error("Passwords do not match"));
                },
              }),
            ]}
          >
            <Input.Password prefix={<LockOutlined />} />
          </Form.Item>
          <Form.Item style={{ marginBottom: 0 }}>
            <Button type="primary" htmlType="submit">
              Change Password
            </Button>
          </Form.Item>
        </Form>
      </Card>
    </div>
  );
}
