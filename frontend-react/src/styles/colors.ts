// src/styles/colors.ts
export const primary = {
  DEFAULT: "#0369a1", // sky-700
  hover: "#0284c7", // sky-600
  active: "#075985", // sky-800
  light: "#e0f2fe", // sky-100
  lighter: "#f0f9ff", // sky-50
} as const;

export const neutral = {
  white: "#ffffff",
  background: "#f8fafc", // slate-50
  surface: "#ffffff", // white
  border: "#e2e8f0", // slate-200
  textMuted: "#64748b", // slate-500
  text: "#334155", // slate-700
  textStrong: "#0f172a", // slate-900
} as const;

export const status = {
  success: { DEFAULT: "#16a34a", light: "#dcfce7", dark: "#15803d" },
  warning: { DEFAULT: "#d97706", light: "#fef3c7", dark: "#b45309" },
  error: { DEFAULT: "#dc2626", light: "#fee2e2", dark: "#b91c1c" },
  info: { DEFAULT: "#0369a1", light: "#e0f2fe", dark: "#075985" },
} as const;

export const platform = {
  shopee: "#ee4d2d",
  tiktok: "#000000",
  lazada: "#0f146d",
} as const;
