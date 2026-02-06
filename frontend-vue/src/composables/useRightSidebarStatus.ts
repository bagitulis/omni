export function useRightSidebarStatus() {
  const getTokenStatusClass = (statusData: any): string => {
    if (typeof statusData === "string") {
      if (statusData.includes("EXPIRED") || statusData.includes("MISSING")) {
        return "token-expired";
      } else if (statusData.includes("VALID")) {
        return "token-valid";
      }
    } else if (statusData && typeof statusData === "object") {
      if (statusData.isExpired || !statusData.connected) {
        return "token-expired";
      } else if (statusData.connected) {
        return "token-valid";
      }
    }
    return "token-warning";
  };

  const getTokenStatus = (statusData: any): string => {
    if (typeof statusData === "string") {
      if (statusData.includes("EXPIRED") || statusData.includes("MISSING")) {
        return "expired";
      } else if (statusData.includes("VALID")) {
        return "valid";
      }
    } else if (statusData && typeof statusData === "object") {
      if (statusData.isExpired || !statusData.connected) {
        return "expired";
      } else if (statusData.connected) {
        return "valid";
      }
    }
    return "warning";
  };

  const formatStatusText = (statusData: any): string => {
    if (typeof statusData === "string") {
      return statusData
        .replace(/\n/g, "<br>")
        .replace(/✅/g, '<span style="color: #27ae60;">✅</span>')
        .replace(/❌/g, '<span style="color: #e74c3c;">❌</span>')
        .replace(/⚠️/g, '<span style="color: #f39c12;">⚠️</span>');
    } else if (statusData && typeof statusData === "object") {
      const accessStatus = statusData.isExpired ? "❌ EXPIRED" : "✅ VALID";
      const refreshStatus = statusData.refreshTokenExpiresAt
        ? new Date(statusData.refreshTokenExpiresAt).getTime() < Date.now()
          ? "❌ EXPIRED"
          : "✅ VALID"
        : "⚠️ UNKNOWN";

      const accessExpiresAt = new Date(statusData.expiresAt);
      const now = new Date();
      const diffMs = accessExpiresAt.getTime() - now.getTime();
      const accessDays = Math.floor(diffMs / (1000 * 60 * 60 * 24));
      const accessHours = Math.floor(
        (diffMs % (1000 * 60 * 60 * 24)) / (1000 * 60 * 60)
      );
      const accessMinutes = Math.floor(
        (diffMs % (1000 * 60 * 60)) / (1000 * 60)
      );
      const accessTimeRemaining = `${accessDays}d ${accessHours}h ${accessMinutes}m`;

      let refreshTimeRemaining = "Unknown";
      if (statusData.refreshTokenExpiresAt) {
        const refreshExpiresAt = new Date(statusData.refreshTokenExpiresAt);
        const refreshDiffMs =
          refreshExpiresAt.getTime() - now.getTime();
        const refreshDays = Math.floor(
          refreshDiffMs / (1000 * 60 * 60 * 24)
        );
        const refreshHours = Math.floor(
          (refreshDiffMs % (1000 * 60 * 60 * 24)) / (1000 * 60 * 60)
        );
        const refreshMinutes = Math.floor(
          (refreshDiffMs % (1000 * 60 * 60)) / (1000 * 60)
        );
        refreshTimeRemaining = `${refreshDays}d ${refreshHours}h ${refreshMinutes}m`;
      }

      const html =
        `<span style="color: #27ae60;">Access Token:</span> ${accessStatus}<br>` +
        `&nbsp;&nbsp;Expires in: ${accessTimeRemaining}<br>` +
        `<span style="color: #27ae60;">Refresh Token:</span> ${refreshStatus}<br>` +
        `&nbsp;&nbsp;Expires in: ${refreshTimeRemaining}`;

      return html;
    }
    return "Status unavailable";
  };

  return {
    getTokenStatusClass,
    getTokenStatus,
    formatStatusText,
  };
}
