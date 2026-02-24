import "@testing-library/jest-dom/vitest";

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
