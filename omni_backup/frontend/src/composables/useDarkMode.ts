import { ref, watch, onMounted } from "vue";

/**
 * useDarkMode Composable
 * Single Responsibility: Manage dark/light theme toggle
 */
export function useDarkMode() {
  const isDark = ref(false);
  const STORAGE_KEY = "omni-theme";

  const applyTheme = (dark: boolean) => {
    const root = document.documentElement;
    if (dark) {
      root.classList.add("dark");
      root.style.setProperty("--bg-primary", "#1a1a2e");
      root.style.setProperty("--bg-secondary", "#16213e");
      root.style.setProperty("--bg-card", "#0f3460");
      root.style.setProperty("--text-primary", "#eaeaea");
      root.style.setProperty("--text-secondary", "#a0a0a0");
      root.style.setProperty("--border-color", "#2a3f5f");
    } else {
      root.classList.remove("dark");
      root.style.setProperty("--bg-primary", "#f5f7fa");
      root.style.setProperty("--bg-secondary", "#ffffff");
      root.style.setProperty("--bg-card", "#ffffff");
      root.style.setProperty("--text-primary", "#333333");
      root.style.setProperty("--text-secondary", "#666666");
      root.style.setProperty("--border-color", "#e0e0e0");
    }
  };

  const toggle = () => {
    isDark.value = !isDark.value;
  };

  const setDark = (value: boolean) => {
    isDark.value = value;
  };

  // Watch for changes and persist
  watch(isDark, (newValue) => {
    localStorage.setItem(STORAGE_KEY, newValue ? "dark" : "light");
    applyTheme(newValue);
  });

  // Initialize from localStorage or system preference
  onMounted(() => {
    const stored = localStorage.getItem(STORAGE_KEY);
    if (stored) {
      isDark.value = stored === "dark";
    } else {
      // Check system preference
      isDark.value = window.matchMedia("(prefers-color-scheme: dark)").matches;
    }
    applyTheme(isDark.value);
  });

  return {
    isDark,
    toggle,
    setDark,
  };
}
