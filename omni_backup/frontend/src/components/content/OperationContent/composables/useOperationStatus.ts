// Composable untuk platform status logic
export function useOperationStatus(props: any) {
  const getPlatformIcon = (platform: string | undefined): string => {
    const icons: Record<string, string> = {
      shopee: "🛍️",
      lazada: "📦",
      tiktok: "🎵",
    };
    return icons[platform || ""] || "🛒";
  };

  const getPlatformStatusClass = () => {
    if (!props.status || !props.status[props.platform]) {
      return "status-unknown";
    }

    const statusText = props.status[props.platform];
    if (typeof statusText === "string") {
      if (statusText.includes("EXPIRED") || statusText.includes("MISSING")) {
        return "status-error";
      } else if (statusText.includes("VALID")) {
        return "status-success";
      }
    }
    return "status-warning";
  };

  const getPlatformStatusText = () => {
    if (!props.status || !props.status[props.platform]) {
      return "Status Unknown";
    }

    const statusText = props.status[props.platform];
    if (typeof statusText === "string") {
      if (statusText.includes("EXPIRED")) return "Token Expired";
      if (statusText.includes("VALID")) return "Token Valid";
      if (statusText.includes("MISSING")) return "Token Missing";
    }
    return "Check Required";
  };

  return {
    getPlatformIcon,
    getPlatformStatusClass,
    getPlatformStatusText,
  };
}
