import { Spin } from "antd";

export function PageLoading() {
  return (
    <div
      style={{
        display: "flex",
        alignItems: "center",
        justifyContent: "center",
        minHeight: "calc(100vh - 48px)",
        width: "100%",
      }}
    >
      <Spin size="large" />
    </div>
  );
}
