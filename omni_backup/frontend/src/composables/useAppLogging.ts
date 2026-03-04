import { ref, Ref } from "vue";

export function useAppLogging() {
  const logs: Ref<string> = ref("");

  const addLog = (message: string): void => {
    const timestamp = new Date().toLocaleTimeString();
    logs.value += `[${timestamp}] ${message}\n`;

    // OPTIMIZED: Use requestAnimationFrame to batch DOM reads/writes
    // This prevents forced reflows by batching layout calculations
    requestAnimationFrame(() => {
      const logsContainer = document.querySelector(".logs-content");
      if (!logsContainer) return;

      // Cache all DOM measurements in one read operation
      const elem = logsContainer as HTMLElement;
      const scrollHeight = elem.scrollHeight;
      const scrollTop = elem.scrollTop;
      const clientHeight = elem.clientHeight;

      // Now perform the comparison with cached values (no reflow)
      if (scrollHeight - scrollTop - clientHeight < 100) {
        elem.scrollTop = scrollHeight;
      }
    });

    const logLines = logs.value.split("\n");
    if (logLines.length > 1000) {
      logs.value = logLines.slice(-500).join("\n");
    }
  };

  const clearLogs = (): void => {
    logs.value = "";
    addLog("📋 Logs cleared");
  };

  return {
    logs,
    addLog,
    clearLogs,
  };
}
