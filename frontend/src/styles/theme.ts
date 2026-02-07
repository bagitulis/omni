import { ThemeConfig } from "antd";

export const antdTheme: ThemeConfig = {
  token: {
    // Colors
    colorPrimary: "#0369a1",
    colorSuccess: "#16a34a",
    colorWarning: "#d97706",
    colorError: "#dc2626",
    colorInfo: "#0369a1",

    // Text
    colorText: "#334155",
    colorTextSecondary: "#64748b",
    colorTextTertiary: "#94a3b8",
    colorTextQuaternary: "#cbd5e1",

    // Backgrounds
    colorBgContainer: "#ffffff",
    colorBgLayout: "#f8fafc",
    colorBgSpotlight: "#f1f5f9",
    colorBgElevated: "#ffffff",

    // Borders
    colorBorder: "#e2e8f0",
    colorBorderSecondary: "#f1f5f9",

    // Typography
    fontFamily:
      "system-ui, -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif",
    fontSize: 12,
    fontSizeSM: 10,
    fontSizeLG: 14,
    fontSizeXL: 16,

    // Sizing
    controlHeight: 32,
    controlHeightSM: 28,
    controlHeightLG: 40,

    // Shape
    borderRadius: 3,
    borderRadiusSM: 2,
    borderRadiusLG: 6,

    // Spacing (Ant uses margin/padding multipliers)
    marginXS: 4,
    marginSM: 8,
    margin: 12,
    marginMD: 16,
    marginLG: 20,
    marginXL: 24,

    paddingXS: 4,
    paddingSM: 8,
    padding: 12,
    paddingMD: 16,
    paddingLG: 20,
    paddingXL: 24,

    // Shadows
    boxShadow: "0 1px 3px rgba(0, 0, 0, 0.1)",
    boxShadowSecondary: "0 4px 6px rgba(0, 0, 0, 0.1)",

    // Motion (Ant Design 5 official tokens)
    motionDurationFast: "0.1s",
    motionDurationMid: "0.2s",
    motionDurationSlow: "0.3s",
    motionEaseInOut: "cubic-bezier(0.645, 0.045, 0.355, 1)",
    motionEaseOut: "cubic-bezier(0.215, 0.61, 0.355, 1)",
  },

  components: {
    Button: {
      controlHeight: 32,
      paddingContentHorizontal: 16,
      fontWeight: 500,
    },
    Input: {
      controlHeight: 32,
      paddingInline: 12,
    },
    Select: {
      controlHeight: 32,
    },
    Table: {
      fontSize: 12,
      cellPaddingBlock: 8,
      cellPaddingInline: 12,
      headerBg: "#f8fafc",
      headerColor: "#334155",
      rowHoverBg: "#f0f9ff",
      borderColor: "#e2e8f0",
    },
    Card: {
      paddingLG: 16,
      borderRadiusLG: 3,
    },
    Menu: {
      itemHeight: 40,
      itemPaddingInline: 16,
      subMenuItemBg: "#f8fafc",
    },
    Modal: {
      borderRadiusLG: 6,
    },
    Drawer: {
      paddingLG: 24,
    },
    Tag: {
      defaultBg: "#f1f5f9",
      defaultColor: "#334155",
    },
    Badge: {
      fontSize: 10,
    },
  },
};

// Dark mode overrides
export const antdDarkTheme: ThemeConfig = {
  ...antdTheme,
  token: {
    ...antdTheme.token,
    colorBgContainer: "#0F172A",
    colorBgLayout: "#0B1220",
    colorBgSpotlight: "#1E293B",
    colorBgElevated: "#111C33",
    colorBorder: "#1E293B",
    colorBorderSecondary: "#334155",
    colorText: "#E2E8F0",
    colorTextSecondary: "#94A3B8",
    colorTextTertiary: "#64748B",
  },
  components: {
    ...antdTheme.components,
    Table: {
      ...antdTheme.components?.Table,
      headerBg: "#1E293B",
      rowHoverBg: "#1E293B",
      borderColor: "#334155",
    },
  },
};
