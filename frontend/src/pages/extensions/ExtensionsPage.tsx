import { useState } from "react";
import {
  Alert,
  Button,
  Card,
  Space,
  Typography,
  message,
} from "antd";
import { ReloadOutlined, LinkOutlined } from "@ant-design/icons";
import {
  useExtensions,
  useGeneratePairingCode,
  useUnpairExtension,
} from "@/hooks/useExtensions";
import { ExtensionTable } from "@/components/tables/ExtensionTable";

const { Title, Paragraph, Text } = Typography;

/**
 * Fallback pairing-code lifetime, used only if the backend omits `ttl_seconds`.
 * The real value comes from the response so the two cannot drift apart.
 */
const PAIRING_CODE_TTL_SECONDS = 300;

/** Render a pairing code as plain characters, since it is read aloud or typed. */
function PairingCodePanel({ code, ttlSeconds }: { code: string; ttlSeconds: number }) {
  const [copied, setCopied] = useState(false);

  const copy = async () => {
    try {
      await navigator.clipboard.writeText(code);
      setCopied(true);
      message.success("Pairing code copied");
    } catch {
      // The clipboard is unavailable over plain http and in some browsers; the
      // code is visible on screen either way, so this is not worth an error.
      message.warning("Could not copy automatically — read the code below");
    }
  };

  return (
    <Space direction="vertical" style={{ width: "100%" }} align="center">
      <Text
        strong
        style={{ fontSize: 28, letterSpacing: 4, fontFamily: "monospace" }}
      >
        {code}
      </Text>
      <Space>
        <Button size="small" onClick={copy}>
          {copied ? "Copied" : "Copy"}
        </Button>
        <Text type="secondary">
          Valid for {Math.round(ttlSeconds / 60)} minutes
        </Text>
      </Space>
    </Space>
  );
}

export function ExtensionsPage() {
  const {
    data: extensions,
    isLoading,
    isError,
    error,
    refetch,
    failureNotice,
  } = useExtensions();
  const generateCode = useGeneratePairingCode();
  const unpair = useUnpairExtension();

  const onGenerate = () => {
    generateCode.mutate(undefined, {
      onError: (err: Error) =>
        message.error(err.message || "Failed to generate a pairing code"),
    });
  };

  const onUnpair = (extensionId: string) => {
    unpair.mutate(extensionId, {
      onSuccess: () => message.success("Extension unpaired"),
      onError: (err: Error) => message.error(err.message || "Failed to unpair"),
    });
  };

  return (
    <div style={{ padding: 24 }}>
      <Space
        style={{ width: "100%", justifyContent: "space-between" }}
        align="start"
      >
        <div>
          <Title level={3} style={{ marginBottom: 4 }}>
            Installed Extensions
          </Title>
          <Paragraph type="secondary" style={{ marginBottom: 0 }}>
            Chrome extensions paired to this tenant. Each tenant pairs its own
            browser; an extension can only act within the tenant it was paired
            to.
          </Paragraph>
        </div>
        <Space>
          <Button icon={<ReloadOutlined />} onClick={() => void refetch()}>
            Refresh
          </Button>
          <Button
            type="primary"
            icon={<LinkOutlined />}
            loading={generateCode.isPending}
            onClick={onGenerate}
          >
            Pair New Extension
          </Button>
        </Space>
      </Space>

      {generateCode.data && (
        <Card style={{ marginTop: 16 }} title="Pairing code">
          <PairingCodePanel
            code={generateCode.data.code}
            ttlSeconds={
              generateCode.data.ttl_seconds > 0
                ? generateCode.data.ttl_seconds
                : PAIRING_CODE_TTL_SECONDS
            }
          />
          <Paragraph type="secondary" style={{ marginTop: 16, marginBottom: 0 }}>
            Enter this code in the Omni extension popup. It works once and
            expires shortly, so generate a new one if it lapses.
          </Paragraph>
        </Card>
      )}

      {failureNotice && (
        <Alert
          style={{ marginTop: 16 }}
          type="error"
          showIcon
          message="Live updates are failing"
          description={failureNotice}
        />
      )}

      {isError && !failureNotice && (
        <Alert
          style={{ marginTop: 16 }}
          type="error"
          showIcon
          message="Failed to load extensions"
          description={(error as Error)?.message}
        />
      )}

      <Card style={{ marginTop: 16 }}>
        <ExtensionTable
          extensions={extensions ?? []}
          loading={isLoading}
          onUnpair={onUnpair}
          unpairing={unpair.isPending}
        />
      </Card>
    </div>
  );
}

export default ExtensionsPage;
