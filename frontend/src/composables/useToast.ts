interface ToastInstance {
  add: (
    summary: string,
    detail?: string,
    severity?: string,
    life?: number
  ) => void;
  clear: () => void;
}

// Global toast instance reference
let toastInstance: ToastInstance | null = null;

export const useToast = () => {
  /**
   * Show toast with custom Toast component API
   * The Toast component API: add(summary, detail?, severity?, life?)
   */
  const show = (
    summary: string,
    detail?: string,
    severity: "success" | "info" | "warning" | "error" = "info",
    life: number = 3000
  ) => {
    if (!toastInstance) {
      console.warn(
        "Toast component not found. Make sure Toast component is mounted in App.vue"
      );
      return;
    }

    toastInstance.add(summary, detail, severity, life);
  };

  const success = (summary: string, detail?: string, life: number = 3000) => {
    show(summary, detail, "success", life);
  };

  const error = (summary: string, detail?: string, life: number = 4000) => {
    show(summary, detail, "error", life);
  };

  const warning = (summary: string, detail?: string, life: number = 3000) => {
    show(summary, detail, "warning", life);
  };

  const info = (summary: string, detail?: string, life: number = 3000) => {
    show(summary, detail, "info", life);
  };

  const clear = () => {
    if (toastInstance) {
      toastInstance.clear();
    }
  };

  /**
   * Support for object-style API (for compatibility with code using object syntax)
   */
  const add = (options: {
    severity?: "success" | "info" | "warning" | "error";
    summary?: string;
    detail?: string;
    life?: number;
  }) => {
    const {
      severity = "info",
      summary = "",
      detail = "",
      life = 3000,
    } = options;
    show(summary, detail, severity, life);
  };

  return {
    show,
    success,
    error,
    warning,
    info,
    clear,
    add,
    // For setting the global instance (called from App.vue)
    setInstance: (instance: ToastInstance) => {
      toastInstance = instance;
    },
  };
};

// Create global singleton
export const toast = useToast();
