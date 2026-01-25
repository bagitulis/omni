// Theme management utilities
import { ref, computed, watch } from "vue";

// Theme state
const isDarkMode = ref(
  typeof window !== "undefined"
    ? localStorage.getItem("theme") === "dark" ||
        window.matchMedia("(prefers-color-scheme: dark)").matches
    : false
);

// Initialize theme on app load
const applyTheme = (isDark: boolean) => {
  const html = document.documentElement;
  if (isDark) {
    html.classList.add("dark");
    html.setAttribute("data-theme", "dark");
    localStorage.setItem("theme", "dark");
  } else {
    html.classList.remove("dark");
    html.setAttribute("data-theme", "light");
    localStorage.setItem("theme", "light");
  }
};

// Get current theme
export const useTheme = () => {
  const theme = computed({
    get: () => (isDarkMode.value ? "dark" : "light"),
    set: (value: "dark" | "light") => {
      isDarkMode.value = value === "dark";
      applyTheme(value === "dark");
    },
  });

  const isDark = computed(() => isDarkMode.value);

  const toggleTheme = () => {
    isDarkMode.value = !isDarkMode.value;
    applyTheme(isDarkMode.value);
  };

  // Watch for external theme changes
  watch(isDarkMode, (newValue) => {
    applyTheme(newValue);
  });

  // Apply theme on initialization
  if (typeof window !== "undefined") {
    applyTheme(isDarkMode.value);
  }

  return {
    theme,
    isDark,
    isDarkMode,
    toggleTheme,
  };
};

// Common color schemes (Daisy UI compatible)
export const colors = {
  primary: "bg-primary text-primary-content",
  secondary: "bg-secondary text-secondary-content",
  success: "bg-success text-success-content",
  warning: "bg-warning text-warning-content",
  error: "bg-error text-error-content",
  info: "bg-info text-info-content",
};

// Button classes (Daisy UI)
export const buttonClasses = {
  primary: "btn btn-primary",
  secondary: "btn btn-secondary",
  outline: "btn btn-outline",
  ghost: "btn btn-ghost",
  danger: "btn btn-error",
  success: "btn btn-success",
  warning: "btn btn-warning",
};

// Card classes (Daisy UI)
export const cardClasses = "card bg-base-100 shadow-xl";

// Input classes (Daisy UI)
export const inputClasses = "input input-bordered w-full";

// Badge classes (Daisy UI)
export const badgeClasses = {
  primary: "badge badge-primary",
  secondary: "badge badge-secondary",
  success: "badge badge-success",
  warning: "badge badge-warning",
  error: "badge badge-error",
  info: "badge badge-info",
};

// Responsive container (Tailwind)
export const containerClasses = "container mx-auto px-4 sm:px-6 lg:px-8";
