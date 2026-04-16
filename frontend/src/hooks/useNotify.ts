import { message } from "@/components/AntStaticApi";
import { notificationApi } from "@/api/notifications";
import { sanitizeForUser } from "@/lib/notificationSecurity";

/**
 * Options accepted by the `notify` tier functions.
 */
interface NotifyOptions {
  category?: string;
  actionUrl?: string;
}

/**
 * Unified notification hook with two tiers:
 *
 *  - **toast**  — ephemeral Ant Design message (disappears after ~3s)
 *  - **notify** — toast *plus* saved to the persistent Database (via SSE)
 *
 * Tier 2 (notify) now hits the backend API. Real-time feedback is handled
 * by the SSE listener in NotificationContext to avoid duplicate toasts.
 */
export function useNotify() {
  return {
    /** Tier 1 — toast only (ephemeral, not saved) */
    toast: {
      success: (content: string) => message.success(content),
      error: (content: string) => message.error(content),
      warning: (content: string) => message.warning(content),
      info: (content: string) => message.info(content),
    },

    /** Tier 2 — saved to Persistent Notification Center (Database) */
    notify: {
      success: (title: string, detail?: string, opts?: NotifyOptions) => {
        notificationApi.create({
          type: "success",
          category: opts?.category ?? "system",
          title,
          message: detail ?? "",
          action_url: opts?.actionUrl,
        });
      },
      error: (title: string, detail?: string, opts?: NotifyOptions) => {
        const safeDetail = sanitizeForUser(detail ?? "");
        notificationApi.create({
          type: "error",
          category: opts?.category ?? "system",
          title,
          message: safeDetail,
          action_url: opts?.actionUrl,
        });
      },
      warning: (title: string, detail?: string, opts?: NotifyOptions) => {
        notificationApi.create({
          type: "warning",
          category: opts?.category ?? "system",
          title,
          message: detail ?? "",
          action_url: opts?.actionUrl,
        });
      },
      info: (title: string, detail?: string, opts?: NotifyOptions) => {
        notificationApi.create({
          type: "info",
          category: opts?.category ?? "system",
          title,
          message: detail ?? "",
          action_url: opts?.actionUrl,
        });
      },
    },
  };
}
