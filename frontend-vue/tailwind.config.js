/** @type {import('tailwindcss').Config} */
import forms from "@tailwindcss/forms";
import typography from "@tailwindcss/typography";

export default {
  content: [
    "./index.html",
    "./src/**/*.{vue,js,ts,jsx,tsx}",
  ],
  theme: {
    screens: {
      xs: "320px",
      sm: "640px",
      md: "768px",
      lg: "1024px",
      xl: "1280px",
      "2xl": "1536px",
    },
    extend: {
      colors: {
        primary: {
          50: "#eff6ff",
          100: "#dbeafe",
          200: "#bfdbfe",
          300: "#93c5fd",
          400: "#60a5fa",
          500: "#3b82f6",
          600: "#2563eb",
          700: "#1d4ed8",
          800: "#1e40af",
          900: "#1e3a8a",
        },
        secondary: "#8b5cf6",
        success: "#10b981",
        warning: "#f59e0b",
        error: "#ef4444",
      },
      fontFamily: {
        sans: ["system-ui", "sans-serif"],
      },
      spacing: {
        safe: "env(safe-area-inset-bottom)",
      },
    },
  },
  // Minimal color palette for light/dark mode
  darkMode: "class",
  plugins: [forms, typography],
  corePlugins: {
    // Disable unused utilities to reduce bundle
    animation: false,
    keyframes: false,
    preflight: true,
    scrollSnapAlign: false,
    scrollSnapStop: false,
    scrollSnapType: false,
    touchAction: false,
  },
  safelist: [
    // Pre-include critical classes that are generated dynamically
    "dark",
  ],
};
