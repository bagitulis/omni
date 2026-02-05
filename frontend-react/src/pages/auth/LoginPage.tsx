import React, { useEffect, useState } from "react";
import {
  Form,
  Input,
  Button,
  Card,
  Select,
  Typography,
  Alert,
  Spin,
  Space,
  theme,
} from "antd";
import {
  UserOutlined,
  LockOutlined,
  SafetyOutlined,
  ToolOutlined,
} from "@ant-design/icons";
import { useNavigate, useSearchParams } from "react-router-dom";
import { login, devLogin, LoginPayload } from "@/api/auth";
import { useAuthStore } from "@/stores/authStore";

const { Title, Text } = Typography;
const { Option } = Select;
const { useToken } = theme;

const LoginPage: React.FC = () => {
  const navigate = useNavigate();
  const [searchParams] = useSearchParams();
  const { setAuth, isAuthenticated } = useAuthStore();
  const { token } = useToken();
  const [form] = Form.useForm();

  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [isAutoLogin, setIsAutoLogin] = useState(false);

  // Dev Mode State
  const isLocalhost =
    window.location.hostname === "localhost" ||
    window.location.hostname === "127.0.0.1";
  const [selectedDevTenant, setSelectedDevTenant] =
    useState("yumna_bertigamart");

  const returnUrl = searchParams.get("returnUrl") || "/";

  // Redirect if already authenticated
  useEffect(() => {
    if (isAuthenticated) {
      navigate(returnUrl);
    }
  }, [isAuthenticated, navigate, returnUrl]);

  // Auto Login Logic for Dev
  useEffect(() => {
    if (
      isLocalhost &&
      !isAuthenticated &&
      !sessionStorage.getItem("autoLoginFailed")
    ) {
      handleDevLogin(true);
    }
  }, [isLocalhost, isAuthenticated]);

  const handleLogin = async (values: LoginPayload) => {
    setIsLoading(true);
    setError(null);
    try {
      const response = await login(values);
      setAuth({
        token: response.token,
        user: response.user,
        tenant_id: response.tenant_id,
      });
      navigate(returnUrl);
    } catch (err: any) {
      setError(err.message || "Login failed");
    } finally {
      setIsLoading(false);
    }
  };

  const handleDevLogin = async (auto = false) => {
    if (auto) setIsAutoLogin(true);
    setIsLoading(true);
    setError(null);

    try {
      const response = await devLogin({ tenant_id: selectedDevTenant });
      setAuth({
        token: response.token,
        user: response.user,
        tenant_id: response.tenant_id,
      });
      navigate(returnUrl);
    } catch (err: any) {
      if (auto) {
        console.warn("Auto-login failed:", err.message);
        sessionStorage.setItem("autoLoginFailed", "true");
        setIsAutoLogin(false);
      } else {
        setError(err.message || "Dev login failed");
      }
    } finally {
      setIsLoading(false);
    }
  };

  if (isAutoLogin) {
    return (
      <div
        style={{
          height: "100vh",
          display: "flex",
          justifyContent: "center",
          alignItems: "center",
          flexDirection: "column",
          background: `linear-gradient(135deg, ${token.colorPrimary} 0%, #0f172a 100%)`,
          color: "white",
        }}
      >
        <Spin size="large" />
        <Text style={{ color: "white", marginTop: 16 }}>
          Auto-logging in as tester...
        </Text>
      </div>
    );
  }

  return (
    <div
      style={{
        minHeight: "100vh",
        display: "flex",
        justifyContent: "center",
        alignItems: "center",
        background: `linear-gradient(135deg, ${token.colorPrimary} 0%, #0f172a 100%)`,
        padding: 20,
      }}
    >
      <Card
        style={{
          width: "100%",
          maxWidth: 400,
          boxShadow: "0 10px 25px rgba(0, 0, 0, 0.2)",
        }}
        variant="borderless"
      >
        <div style={{ textAlign: "center", marginBottom: 30 }}>
          <Title level={2} style={{ margin: 0 }}>
            Login
          </Title>
          <Text type="secondary">Welcome back to OMNI</Text>
        </div>

        {error && (
          <Alert
            message={error}
            type="error"
            showIcon
            style={{ marginBottom: 24 }}
          />
        )}

        {isLocalhost && (
          <div
            style={{
              background: "#fffbeb",
              border: "1px solid #fcd34d",
              borderRadius: 6,
              padding: 16,
              marginBottom: 24,
            }}
          >
            <Space direction="vertical" style={{ width: "100%" }}>
              <Space>
                <ToolOutlined style={{ color: "#d97706" }} />
                <Text strong style={{ color: "#92400e" }}>
                  Dev Mode (Localhost)
                </Text>
              </Space>
              <Select
                value={selectedDevTenant}
                onChange={setSelectedDevTenant}
                style={{ width: "100%" }}
              >
                <Option value="yumna_bertigamart">Yumna - Bertigamart</Option>
                <Option value="tika_nusseyba">Tika - Nusseyba</Option>
              </Select>
              <Button
                type="primary"
                style={{ backgroundColor: "#d97706" }}
                block
                onClick={() => handleDevLogin(false)}
                loading={isLoading}
              >
                Quick Dev Login
              </Button>
            </Space>
          </div>
        )}

        <Form
          form={form}
          name="login"
          layout="vertical"
          onFinish={handleLogin}
          size="large"
          disabled={isLoading}
        >
          <Form.Item
            name="username"
            label="Username"
            rules={[{ required: true, message: "Please input your username!" }]}
          >
            <Input prefix={<UserOutlined />} placeholder="Username" />
          </Form.Item>

          <Form.Item
            name="password"
            label="Password"
            rules={[{ required: true, message: "Please input your password!" }]}
          >
            <Input.Password prefix={<LockOutlined />} placeholder="Password" />
          </Form.Item>

          <Form.Item>
            <Button type="primary" htmlType="submit" block loading={isLoading}>
              Login
            </Button>
          </Form.Item>
        </Form>

        <div style={{ textAlign: "center", marginTop: 16 }}>
          <Space direction="vertical" size={4}>
            <Text type="secondary" style={{ fontSize: 12 }}>
              <SafetyOutlined /> Protected by reCAPTCHA
            </Text>
            <Text type="secondary" style={{ fontSize: 10 }}>
              <a
                href="https://policies.google.com/privacy"
                target="_blank"
                rel="noreferrer"
              >
                Privacy
              </a>
              {" & "}
              <a
                href="https://policies.google.com/terms"
                target="_blank"
                rel="noreferrer"
              >
                Terms
              </a>
            </Text>
          </Space>
        </div>
      </Card>
    </div>
  );
};

export default LoginPage;
