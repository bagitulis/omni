/**
 * Theme constants
 * Separated from ThemeContext for better Fast Refresh support
 */

export const THEME_OPTIONS = {
  LIGHT: "light",
  DARK: "dark",
} as const;

export type ThemeMode = (typeof THEME_OPTIONS)[keyof typeof THEME_OPTIONS];
