import type { MessageInstance } from "antd/es/message/interface";
import type { NotificationInstance } from "antd/es/notification/interface";

let message!: MessageInstance;
let notification!: NotificationInstance;

export function setAntStaticApi(next: {
  message: MessageInstance;
  notification: NotificationInstance;
}): void {
  message = next.message;
  notification = next.notification;
}

export { message, notification };
