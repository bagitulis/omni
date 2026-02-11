import { App } from "antd";

/**
 * Hook for showing notifications using Ant Design's App context.
 * Must be used within an <App> wrapper component.
 *
 * Provides: showSuccess, showError, showWarning, showInfo
 */
export function useNotification() {
  const { message, notification } = App.useApp();

  return {
    showSuccess: (content: string) => message.success(content),
    showError: (content: string) => message.error(content),
    showWarning: (content: string) => message.warning(content),
    showInfo: (content: string) => message.info(content),
    // Full notification API for richer notifications
    notify: {
      success: (title: string, description?: string) =>
        notification.success({ message: title, description }),
      error: (title: string, description?: string) =>
        notification.error({ message: title, description }),
      warning: (title: string, description?: string) =>
        notification.warning({ message: title, description }),
      info: (title: string, description?: string) =>
        notification.info({ message: title, description }),
    },
  };
}
