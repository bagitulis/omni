import "@testing-library/jest-dom/vitest";
import { vi } from "vitest";

// Global mock for AntStaticHolder — the message/notification singletons
// remain undefined until <AntApp> renders AntStaticHolder, which never
// happens in unit tests. This mock satisfies every hook that imports
// { message } or { notification } from "@/components/AntStaticHolder".
vi.mock("@/components/AntStaticHolder", () => ({
  AntStaticHolder: () => null,
  message: {
    success: vi.fn(),
    error: vi.fn(),
    warning: vi.fn(),
    info: vi.fn(),
    loading: vi.fn(),
    destroy: vi.fn(),
  },
  notification: {
    success: vi.fn(),
    error: vi.fn(),
    warning: vi.fn(),
    info: vi.fn(),
    destroy: vi.fn(),
  },
}));

// Global mock for the centralized logger.
// Tests that spy on console.error for logger output need to mock logger
// rather than console, because logger adds timestamp/prefix formatting.
vi.mock("@/lib/logger", () => ({
  logger: {
    debug: vi.fn(),
    info: vi.fn(),
    warn: vi.fn(),
    error: vi.fn(),
  },
}));

const nativeGetComputedStyle = window.getComputedStyle.bind(window);
window.getComputedStyle = (element: Element, pseudoElement?: string | null) => {
  if (pseudoElement) {
    return nativeGetComputedStyle(element);
  }

  return nativeGetComputedStyle(element, pseudoElement);
};

const nativeConsoleError = console.error.bind(console);
const ignoredJSDOMMessages = [
  "Not implemented: navigation (except hash changes)",
];

console.error = (...args: Parameters<typeof console.error>) => {
  const message = args
    .map((arg) => {
      if (typeof arg === "string") {
        return arg;
      }

      if (arg instanceof Error) {
        return arg.message;
      }

      return "";
    })
    .join(" ");

  if (ignoredJSDOMMessages.some((ignored) => message.includes(ignored))) {
    return;
  }

  nativeConsoleError(...args);
};
