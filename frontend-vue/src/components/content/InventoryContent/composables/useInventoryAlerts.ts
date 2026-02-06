import { reactive } from 'vue';

interface Alert {
  type: 'success' | 'error' | 'warning' | 'info';
  title: string;
  message: string;
  details?: string | null;
}

export function useInventoryAlerts() {
  const state = reactive<{ syncAlert: Alert | null }>({
    syncAlert: null,
  });

  const showAlertSuccess = (title: string, message: string, details: string | null = null) => {
    state.syncAlert = { type: 'success', title, message, details };
    setTimeout(() => (state.syncAlert = null), 5000);
  };

  const showAlertError = (title: string, message: string) => {
    state.syncAlert = { type: 'error', title, message };
  };

  const showAlertWarning = (title: string, message: string) => {
    state.syncAlert = { type: 'warning', title, message };
    setTimeout(() => (state.syncAlert = null), 4000);
  };

  const showAlertInfo = (title: string, message: string) => {
    state.syncAlert = { type: 'info', title, message };
  };

  const clearAlert = () => {
    state.syncAlert = null;
  };

  return {
    state,
    showAlertSuccess,
    showAlertError,
    showAlertWarning,
    showAlertInfo,
    clearAlert,
  };
}
