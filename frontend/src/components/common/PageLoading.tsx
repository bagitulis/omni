import { Spin } from "antd";

export function PageLoading() {
  return (
    <div
      style={{
        display: "flex",
        alignItems: "center",
        justifyContent: "center",
        minHeight: "400px",
        width: "100%",
      }}
    >
      <Spin size="large" />
    </div>
  );
}
