import type { MessageInstance } from "antd/es/message/interface";
import type { ModalStaticFunctions } from "antd/es/modal/confirm";
import type { NotificationInstance } from "antd/es/notification/interface";

let message!: MessageInstance;
let modal!: Omit<ModalStaticFunctions, "warn">;
let notification!: NotificationInstance;

export function setAntStaticApi(next: {
  message: MessageInstance;
  modal: Omit<ModalStaticFunctions, "warn">;
  notification: NotificationInstance;
}): void {
  message = next.message;
  modal = next.modal;
  notification = next.notification;
}

export { message, modal, notification };
