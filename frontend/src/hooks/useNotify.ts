import { message } from "@/components/AntStaticHolder";
import {
  useNotificationStore,
  type NotificationCategory,
} from "@/stores/notificationStore";
import { sanitizeForUser } from "@/lib/notificationSecurity";

/**
 * Options accepted by the `notify` tier functions.
 */
interface NotifyOptions {
  category?: NotificationCategory;
  actionUrl?: string;
}

/**
 * Unified notification hook with two tiers:
 *
 *  - **toast**  — ephemeral Ant Design message (disappears after ~3s)
 *  - **notify** — toast *plus* saved to the Notification Center
 *
 * The `notify` tier automatically sanitises error details before
 * persisting them to the store (see `notificationSecurity.ts`).
 *
 * Usage:
 * ```ts
 * const { toast, notify } = useNotify();
 *
 * // Quick feedback (stock updated, settings saved)
 * toast.success("Stock updated");
 *
 * // Important result that should be reviewable later
 * notify.success("Sync Complete", "3 products synced", { category: "sync" });
 * notify.error("Bulk Ship Failed", "2 of 5 orders failed", { category: "order" });
 * ```
 */
export function useNotify() {
  const addNotification = useNotificationStore((s) => s.addNotification);

  return {
    /** Tier 1 — toast only (ephemeral, not saved) */
    toast: {
      success: (content: string) => message.success(content),
      error: (content: string) => message.error(content),
      warning: (content: string) => message.warning(content),
      info: (content: string) => message.info(content),
    },

    /** Tier 2 — toast + saved to Notification Center */
    notify: {
      success: (title: string, detail?: string, opts?: NotifyOptions) => {
        message.success(title);
        addNotification({
          type: "success",
          category: opts?.category ?? "system",
          title,
          message: detail ?? "",
          actionUrl: opts?.actionUrl,
        });
      },
      error: (title: string, detail?: string, opts?: NotifyOptions) => {
        const safeDetail = sanitizeForUser(detail ?? "");
        message.error(title);
        addNotification({
          type: "error",
          category: opts?.category ?? "system",
          title,
          message: safeDetail,
          actionUrl: opts?.actionUrl,
        });
      },
      warning: (title: string, detail?: string, opts?: NotifyOptions) => {
        message.warning(title);
        addNotification({
          type: "warning",
          category: opts?.category ?? "system",
          title,
          message: detail ?? "",
          actionUrl: opts?.actionUrl,
        });
      },
      info: (title: string, detail?: string, opts?: NotifyOptions) => {
        message.info(title);
        addNotification({
          type: "info",
          category: opts?.category ?? "system",
          title,
          message: detail ?? "",
          actionUrl: opts?.actionUrl,
        });
      },
    },
  };
}
