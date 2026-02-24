import { beforeEach, describe, expect, it, vi } from "vitest";

const messageMock = {
  success: vi.fn(),
  error: vi.fn(),
  warning: vi.fn(),
  info: vi.fn(),
};
const notificationMock = {
  success: vi.fn(),
  error: vi.fn(),
  warning: vi.fn(),
  info: vi.fn(),
};

vi.mock("antd", () => ({
  App: {
    useApp: vi.fn(() => ({
      message: messageMock,
      notification: notificationMock,
    })),
  },
}));

import { useNotification } from "./useNotification";

describe("useNotification", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("showSuccess calls message.success", () => {
    const { showSuccess } = useNotification();
    showSuccess("Operation successful");
    expect(messageMock.success).toHaveBeenCalledWith("Operation successful");
  });

  it("showError calls message.error", () => {
    const { showError } = useNotification();
    showError("Something went wrong");
    expect(messageMock.error).toHaveBeenCalledWith("Something went wrong");
  });

  it("showWarning calls message.warning", () => {
    const { showWarning } = useNotification();
    showWarning("Please check this");
    expect(messageMock.warning).toHaveBeenCalledWith("Please check this");
  });

  it("showInfo calls message.info", () => {
    const { showInfo } = useNotification();
    showInfo("FYI");
    expect(messageMock.info).toHaveBeenCalledWith("FYI");
  });

  it("notify.success calls notification.success with title and description", () => {
    const { notify } = useNotification();
    notify.success("Title", "Details here");
    expect(notificationMock.success).toHaveBeenCalledWith({
      message: "Title",
      description: "Details here",
    });
  });

  it("notify.error calls notification.error with title", () => {
    const { notify } = useNotification();
    notify.error("Error title");
    expect(notificationMock.error).toHaveBeenCalledWith({
      message: "Error title",
      description: undefined,
    });
  });

  it("notify.warning calls notification.warning", () => {
    const { notify } = useNotification();
    notify.warning("Warn title", "Warn desc");
    expect(notificationMock.warning).toHaveBeenCalledWith({
      message: "Warn title",
      description: "Warn desc",
    });
  });

  it("notify.info calls notification.info", () => {
    const { notify } = useNotification();
    notify.info("Info title");
    expect(notificationMock.info).toHaveBeenCalledWith({
      message: "Info title",
      description: undefined,
    });
  });
});
