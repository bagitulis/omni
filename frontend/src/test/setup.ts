import "@testing-library/jest-dom/vitest";
import { vi } from "vitest";

Object.defineProperty(window, "scrollTo", {
  writable: true,
  value: vi.fn(),
});

Object.defineProperty(window, "matchMedia", {
  writable: true,
  value: vi.fn().mockImplementation((query: string) => ({
    matches: false,
    media: query,
    onchange: null,
    addListener: vi.fn(),
    removeListener: vi.fn(),
    addEventListener: vi.fn(),
    removeEventListener: vi.fn(),
    dispatchEvent: vi.fn(),
  })),
});

const antStaticMessage = {
  success: vi.fn(),
  error: vi.fn(),
  warning: vi.fn(),
  info: vi.fn(),
  loading: vi.fn(),
  destroy: vi.fn(),
};

const antStaticNotification = {
  success: vi.fn(),
  error: vi.fn(),
  warning: vi.fn(),
  info: vi.fn(),
  destroy: vi.fn(),
};

// Global mock for the context-aware Ant Design singletons used by hooks/utilities.
vi.mock("@/components/AntStaticApi", () => ({
  message: {
    ...antStaticMessage,
  },
  notification: {
    ...antStaticNotification,
  },
}));

// Component stays mocked so App-level renders do not require AntApp context.
vi.mock("@/components/AntStaticHolder", () => ({
  AntStaticHolder: () => null,
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
